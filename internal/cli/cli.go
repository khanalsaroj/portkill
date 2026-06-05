// Package cli implements portkill's command-line interface: argument parsing,
// command dispatch, and human/JSON rendering. The actual port work lives in
// internal/killer; this package is the thin, testable shell around it.
package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/khanalsaroj/portkill/internal/killer"
	"github.com/khanalsaroj/portkill/internal/ui"
	"github.com/khanalsaroj/portkill/internal/version"
)

// Exit codes follow the usual convention: 0 ok, 1 runtime failure, 2 misuse.
const (
	exitOK    = 0
	exitFail  = 1
	exitUsage = 2
)

// errHelp is a sentinel signaling that help was requested mid-parse.
var errHelp = errors.New("help requested")

// Run is the entry point. It returns the process exit code.
func Run(args []string) int {
	return run(args, os.Stdout, os.Stderr)
}

// run is the testable core: all output goes to the provided writers.
func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stderr, usage)
		return exitUsage
	}

	switch args[0] {
	case "version", "--version", "-v":
		fmt.Fprintln(stdout, version.String())
		return exitOK
	case "help", "--help", "-h":
		fmt.Fprint(stdout, usage)
		return exitOK
	case "kill":
		return runKill(args[1:], stdout, stderr)
	case "list", "ls":
		return runList(args[1:], stdout, stderr)
	default:
		fmt.Fprintf(stderr, "portkill: unknown command %q\n\n", args[0])
		fmt.Fprint(stderr, usage)
		return exitUsage
	}
}

// config is the result of parsing a kill/list invocation.
type config struct {
	ports   []int
	force   bool
	dryRun  bool
	json    bool
	noColor bool
	quiet   bool
	timeout time.Duration
}

func runKill(args []string, stdout, stderr io.Writer) int {
	cfg, err := parseArgs(args, true)
	if errors.Is(err, errHelp) {
		fmt.Fprint(stdout, usage)
		return exitOK
	}
	if err != nil {
		fmt.Fprintf(stderr, "portkill: %v\n", err)
		return exitUsage
	}
	if len(cfg.ports) == 0 {
		fmt.Fprintln(stderr, "portkill: 'kill' requires at least one port")
		fmt.Fprintln(stderr, "usage: portkill kill [flags] <port> [port...]")
		return exitUsage
	}

	opts := killer.Options{Force: cfg.force, DryRun: cfg.dryRun, Timeout: cfg.timeout}
	results := make([]killer.Result, 0, len(cfg.ports))
	failed := false
	for _, port := range cfg.ports {
		r := killer.Kill(port, opts)
		results = append(results, r)
		if r.Status == killer.StatusFailed {
			failed = true
		}
	}

	if cfg.json {
		writeKillJSON(stdout, results)
	} else {
		printKill(stdout, stderr, ui.New(stdout, cfg.noColor), ui.New(stderr, cfg.noColor), results, cfg.quiet)
	}

	if failed {
		return exitFail
	}
	return exitOK
}

func runList(args []string, stdout, stderr io.Writer) int {
	cfg, err := parseArgs(args, false)
	if errors.Is(err, errHelp) {
		fmt.Fprint(stdout, usage)
		return exitOK
	}
	if err != nil {
		fmt.Fprintf(stderr, "portkill: %v\n", err)
		return exitUsage
	}
	if len(cfg.ports) == 0 {
		fmt.Fprintln(stderr, "portkill: 'list' requires at least one port")
		fmt.Fprintln(stderr, "usage: portkill list [flags] <port> [port...]")
		return exitUsage
	}

	rows := make([]listRow, 0, len(cfg.ports))
	failed := false
	for _, port := range cfg.ports {
		procs, err := killer.List(port)
		if err != nil {
			failed = true
		}
		rows = append(rows, listRow{port: port, procs: procs, err: err})
	}

	if cfg.json {
		out := make([]jsonListResult, len(rows))
		for i, r := range rows {
			out[i] = jsonListResult{Port: r.port, InUse: len(r.procs) > 0, Processes: r.procs}
		}
		writeJSON(stdout, out)
	} else {
		printList(stdout, stderr, ui.New(stdout, cfg.noColor), ui.New(stderr, cfg.noColor), rows)
	}

	if failed {
		return exitFail
	}
	return exitOK
}

// parseArgs splits flags from port arguments. Flags may appear before, after,
// or interleaved with ports. allowKill enables kill-only flags (force/dry-run/
// timeout); they are rejected for the list command.
func parseArgs(args []string, allowKill bool) (config, error) {
	cfg := config{}
	killOnly := func(flag string) error {
		if !allowKill {
			return fmt.Errorf("flag %q is only valid for 'kill'", flag)
		}
		return nil
	}

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if arg == "--" { // everything after is a port
			for _, rest := range args[i+1:] {
				port, err := parsePort(rest)
				if err != nil {
					return cfg, err
				}
				cfg.ports = append(cfg.ports, port)
			}
			break
		}

		if len(arg) > 1 && arg[0] == '-' {
			key, inline, hasInline := arg, "", false
			if strings.HasPrefix(arg, "--") {
				if eq := strings.IndexByte(arg, '='); eq >= 0 {
					key, inline, hasInline = arg[:eq], arg[eq+1:], true
				}
			}

			switch key {
			case "-h", "--help":
				return cfg, errHelp
			case "-f", "--force":
				if err := killOnly(key); err != nil {
					return cfg, err
				}
				cfg.force = true
			case "-n", "--dry-run":
				if err := killOnly(key); err != nil {
					return cfg, err
				}
				cfg.dryRun = true
			case "--json":
				cfg.json = true
			case "--no-color":
				cfg.noColor = true
			case "-q", "--quiet":
				cfg.quiet = true
			case "--timeout":
				if err := killOnly(key); err != nil {
					return cfg, err
				}
				val := inline
				if !hasInline {
					i++
					if i >= len(args) {
						return cfg, errors.New("--timeout requires a value (e.g. 3s)")
					}
					val = args[i]
				}
				d, err := time.ParseDuration(val)
				if err != nil {
					return cfg, fmt.Errorf("invalid --timeout %q: %w", val, err)
				}
				if d < 0 {
					return cfg, fmt.Errorf("invalid --timeout %q: must not be negative", val)
				}
				cfg.timeout = d
			default:
				return cfg, fmt.Errorf("unknown flag %q", arg)
			}
			continue
		}

		port, err := parsePort(arg)
		if err != nil {
			return cfg, err
		}
		cfg.ports = append(cfg.ports, port)
	}

	return cfg, nil
}

func parsePort(s string) (int, error) {
	port, err := strconv.Atoi(s)
	if err != nil || port < 1 || port > 65535 {
		return 0, fmt.Errorf("%q is not a valid port number (1-65535)", s)
	}
	return port, nil
}
