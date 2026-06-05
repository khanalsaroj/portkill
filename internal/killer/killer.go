// Package killer locates and terminates the processes bound to a TCP or UDP
// port. Detection and termination are implemented per-OS (see find_*.go and
// terminate_*.go); the parsing of external-tool output lives in pure,
// unit-tested helpers (parse_*.go). The orchestration here is OS-agnostic and
// returns structured results so the caller decides how to present them.
package killer

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
)

// Protocol identifiers used in Process.Protocol and JSON output.
const (
	ProtoTCP = "tcp"
	ProtoUDP = "udp"
)

// Process describes a single process holding a port.
type Process struct {
	PID      int    `json:"pid"`
	Name     string `json:"name,omitempty"`
	Port     int    `json:"port"`
	Protocol string `json:"protocol,omitempty"`
	Address  string `json:"address,omitempty"`
}

// Status is the outcome of a Kill for one port.
type Status string

const (
	StatusKilled   Status = "killed"     // every process on the port was terminated
	StatusNotInUse Status = "not_in_use" // nothing was listening
	StatusDryRun   Status = "would_kill" // --dry-run: processes found but left alone
	StatusFailed   Status = "failed"     // detection failed or a process survived
)

// Result is the outcome of Kill for a single port.
type Result struct {
	Port      int       `json:"port"`
	Status    Status    `json:"status"`
	Processes []Process `json:"processes,omitempty"`
	Err       error     `json:"-"`
}

// Options tunes Kill behavior.
type Options struct {
	// Force skips graceful termination and hard-kills immediately
	// (SIGKILL on Unix, taskkill /F on Windows).
	Force bool
	// DryRun reports what would be killed without signaling anything.
	DryRun bool
	// Timeout bounds how long to wait for a process to exit after each
	// signal before escalating (or giving up). Defaults to DefaultTimeout.
	Timeout time.Duration
}

// DefaultTimeout is the grace period applied when Options.Timeout is zero.
const DefaultTimeout = 3 * time.Second

// pollInterval is how often killAndWait re-checks liveness. It is a var (not a
// const) only so tests can shrink it.
var pollInterval = 30 * time.Millisecond

// Indirection seams over the per-OS primitives so the escalation state machine
// in killAndWait can be exercised in tests without real processes.
var (
	killProcessFn  = killProcess
	processAliveFn = processAlive
)

// List returns every process bound to port, sorted by PID. The slice is empty
// (with a nil error) when the port is free.
func List(port int) ([]Process, error) {
	procs, err := findProcesses(port)
	if err != nil {
		return nil, err
	}
	return tidy(procs), nil
}

// Kill terminates every process bound to port and reports the outcome. It never
// returns an error directly; failures are captured in Result.Status/Result.Err.
func Kill(port int, opts Options) Result {
	if opts.Timeout <= 0 {
		opts.Timeout = DefaultTimeout
	}

	procs, err := findProcesses(port)
	if err != nil {
		return Result{Port: port, Status: StatusFailed, Err: err}
	}
	procs = tidy(procs)
	if len(procs) == 0 {
		return Result{Port: port, Status: StatusNotInUse}
	}
	if opts.DryRun {
		return Result{Port: port, Status: StatusDryRun, Processes: procs}
	}

	var failures []string
	for _, p := range procs {
		if err := killAndWait(p.PID, opts); err != nil {
			failures = append(failures, fmt.Sprintf("%s (PID %d): %v", displayName(p), p.PID, err))
		}
	}
	if len(failures) > 0 {
		return Result{
			Port:      port,
			Status:    StatusFailed,
			Processes: procs,
			Err:       errors.New(strings.Join(failures, "; ")),
		}
	}
	return Result{Port: port, Status: StatusKilled, Processes: procs}
}

// killAndWait terminates pid and waits for it to actually exit, escalating from
// a graceful signal to a hard kill if the process refuses to die within the
// grace period. It is OS-agnostic: killProcess and processAlive are per-OS.
func killAndWait(pid int, opts Options) error {
	if !processAliveFn(pid) {
		return nil // won the race; already gone
	}

	hard := opts.Force
	if err := killProcessFn(pid, hard); err != nil {
		if !processAliveFn(pid) {
			return nil // it died despite the error
		}
		if hard {
			return err
		}
		// A graceful request that errors out (e.g. a Windows console app that
		// can only be killed with /F) won't ever succeed — escalate at once
		// rather than waiting out the grace period for nothing.
		hard = true
		if err := killProcessFn(pid, true); err != nil && processAliveFn(pid) {
			return err
		}
	}

	deadline := time.Now().Add(opts.Timeout)
	for {
		if !processAliveFn(pid) {
			return nil
		}
		if time.Now().After(deadline) {
			if hard {
				return errors.New("still running after force kill (try elevated privileges)")
			}
			hard = true // grace expired; escalate to a hard kill once
			if err := killProcessFn(pid, true); err != nil && processAliveFn(pid) {
				return err
			}
			deadline = time.Now().Add(opts.Timeout)
		}
		time.Sleep(pollInterval)
	}
}

// tidy removes processes we must never signal (PID 0/negative and portkill
// itself), de-duplicates by PID keeping the richest record, and sorts by PID
// for stable, predictable output.
func tidy(procs []Process) []Process {
	self := os.Getpid()
	byPID := make(map[int]Process, len(procs))
	for _, p := range procs {
		if p.PID <= 0 || p.PID == self {
			continue
		}
		if existing, ok := byPID[p.PID]; ok {
			byPID[p.PID] = merge(existing, p)
			continue
		}
		byPID[p.PID] = p
	}

	out := make([]Process, 0, len(byPID))
	for _, p := range byPID {
		out = append(out, p)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].PID < out[j].PID })
	return out
}

// merge fills in missing fields of a so it carries the most information from a
// duplicate b (e.g. one record knew the name, the other the protocol).
func merge(a, b Process) Process {
	if a.Name == "" {
		a.Name = b.Name
	}
	if a.Protocol == "" {
		a.Protocol = b.Protocol
	}
	if a.Address == "" {
		a.Address = b.Address
	}
	return a
}

func displayName(p Process) string {
	if p.Name != "" {
		return p.Name
	}
	return "process"
}
