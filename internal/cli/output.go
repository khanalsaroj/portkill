package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/khanalsaroj/portkill/internal/killer"
	"github.com/khanalsaroj/portkill/internal/ui"
)

// describe renders a process list as "node (PID 54321), nginx (PID 99)".
func describe(procs []killer.Process) string {
	parts := make([]string, len(procs))
	for i, p := range procs {
		name := p.Name
		if name == "" {
			name = "process"
		}
		parts[i] = fmt.Sprintf("%s (PID %d)", name, p.PID)
	}
	return strings.Join(parts, ", ")
}

// printKill renders kill results. Progress and successes go to stdout;
// failures go to stderr. In quiet mode only failures are printed.
func printKill(stdout, stderr io.Writer, out, errs *ui.Styler, results []killer.Result, quiet bool) {
	for idx, r := range results {
		if !quiet {
			if idx > 0 {
				fmt.Fprintln(stdout)
			}
			fmt.Fprintf(stdout, "  %s  scanning port %s …\n", out.Scan(), out.Bold(strconv.Itoa(r.Port)))
		}

		switch r.Status {
		case killer.StatusKilled:
			if !quiet {
				fmt.Fprintf(stdout, "  %s  port %d freed — killed %s\n", out.Success(), r.Port, describe(r.Processes))
			}
		case killer.StatusNotInUse:
			if !quiet {
				fmt.Fprintf(stdout, "  %s  port %d is not in use\n", out.Info(), r.Port)
			}
		case killer.StatusDryRun:
			if !quiet {
				fmt.Fprintf(stdout, "  %s  port %d in use by %s — dry run, nothing killed\n", out.Bullet(), r.Port, describe(r.Processes))
			}
		case killer.StatusFailed:
			fmt.Fprintf(stderr, "  %s  port %d — %v\n", errs.Fail(), r.Port, r.Err)
		}
	}
}

// listRow is one port's listing outcome.
type listRow struct {
	port  int
	procs []killer.Process
	err   error
}

// printList renders a `list` result as an aligned table on stdout, with any
// detection errors reported on stderr.
func printList(stdout, stderr io.Writer, out, errs *ui.Styler, rows []listRow) {
	tw := tabwriter.NewWriter(stdout, 0, 2, 2, ' ', 0)
	fmt.Fprintln(tw, out.Bold("PORT\tPROTO\tPID\tPROCESS\tADDRESS"))

	for _, row := range rows {
		if row.err != nil {
			fmt.Fprintf(stderr, "  %s  port %d — %v\n", errs.Fail(), row.port, row.err)
			continue
		}
		if len(row.procs) == 0 {
			fmt.Fprintf(tw, "%d\t%s\t%s\t%s\t%s\n", row.port, "—", "—", out.Dim("(not in use)"), "—")
			continue
		}
		for _, p := range row.procs {
			name := p.Name
			if name == "" {
				name = "—"
			}
			proto := p.Protocol
			if proto == "" {
				proto = "—"
			}
			addr := p.Address
			if addr == "" {
				addr = "—"
			}
			fmt.Fprintf(tw, "%d\t%s\t%d\t%s\t%s\n", p.Port, proto, p.PID, name, addr)
		}
	}
	_ = tw.Flush()
}

// ---- JSON output ----

type jsonKillResult struct {
	Port      int              `json:"port"`
	Status    string           `json:"status"`
	Processes []killer.Process `json:"processes"`
	Error     string           `json:"error,omitempty"`
}

type jsonListResult struct {
	Port      int              `json:"port"`
	InUse     bool             `json:"in_use"`
	Processes []killer.Process `json:"processes"`
}

func writeKillJSON(w io.Writer, results []killer.Result) {
	out := make([]jsonKillResult, len(results))
	for i, r := range results {
		jr := jsonKillResult{
			Port:      r.Port,
			Status:    string(r.Status),
			Processes: r.Processes,
		}
		if r.Err != nil {
			jr.Error = r.Err.Error()
		}
		if jr.Processes == nil {
			jr.Processes = []killer.Process{}
		}
		out[i] = jr
	}
	writeJSON(w, out)
}

func writeJSON(w io.Writer, v any) {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}
