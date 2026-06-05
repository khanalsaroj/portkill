package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/khanalsaroj/portkill/internal/killer"
	"github.com/khanalsaroj/portkill/internal/ui"
)

func TestDescribe(t *testing.T) {
	procs := []killer.Process{
		{PID: 100, Name: "node"},
		{PID: 200}, // no name -> "process"
	}
	if got, want := describe(procs), "node (PID 100), process (PID 200)"; got != want {
		t.Errorf("describe() = %q, want %q", got, want)
	}
}

// plain returns stylers with color disabled so output is deterministic text.
func plain(out, errs *bytes.Buffer) (*ui.Styler, *ui.Styler) {
	return ui.New(out, true), ui.New(errs, true)
}

func TestPrintKill(t *testing.T) {
	results := []killer.Result{
		{Port: 8080, Status: killer.StatusKilled, Processes: []killer.Process{{PID: 1, Name: "node"}}},
		{Port: 3000, Status: killer.StatusNotInUse},
		{Port: 5432, Status: killer.StatusDryRun, Processes: []killer.Process{{PID: 2, Name: "pg"}}},
	}
	var out, errs bytes.Buffer
	so, se := plain(&out, &errs)
	printKill(&out, &errs, so, se, results, false)

	stdout := out.String()
	for _, want := range []string{
		"scanning port 8080",
		"port 8080 freed — killed node (PID 1)",
		"port 3000 is not in use",
		"port 5432 in use by pg (PID 2) — dry run",
	} {
		if !strings.Contains(stdout, want) {
			t.Errorf("stdout missing %q\n---\n%s", want, stdout)
		}
	}
	if errs.Len() != 0 {
		t.Errorf("expected empty stderr, got %q", errs.String())
	}
}

func TestPrintKillFailureGoesToStderr(t *testing.T) {
	results := []killer.Result{
		{Port: 22, Status: killer.StatusFailed, Err: errString("permission denied")},
	}
	var out, errs bytes.Buffer
	so, se := plain(&out, &errs)
	printKill(&out, &errs, so, se, results, false)

	if !strings.Contains(errs.String(), "port 22 — permission denied") {
		t.Errorf("stderr = %q, want failure line", errs.String())
	}
}

func TestPrintKillQuiet(t *testing.T) {
	results := []killer.Result{
		{Port: 8080, Status: killer.StatusKilled, Processes: []killer.Process{{PID: 1, Name: "node"}}},
		{Port: 22, Status: killer.StatusFailed, Err: errString("nope")},
	}
	var out, errs bytes.Buffer
	so, se := plain(&out, &errs)
	printKill(&out, &errs, so, se, results, true)

	if out.Len() != 0 {
		t.Errorf("quiet mode should print nothing to stdout, got %q", out.String())
	}
	if !strings.Contains(errs.String(), "nope") {
		t.Errorf("quiet mode should still report failures, stderr=%q", errs.String())
	}
}

func TestPrintList(t *testing.T) {
	rows := []listRow{
		{port: 8080, procs: []killer.Process{{PID: 1, Name: "node", Protocol: "tcp", Address: "0.0.0.0:8080", Port: 8080}}},
		{port: 3000}, // not in use
	}
	var out, errs bytes.Buffer
	so, se := plain(&out, &errs)
	printList(&out, &errs, so, se, rows)

	stdout := out.String()
	for _, want := range []string{"PORT", "PROTO", "PID", "PROCESS", "ADDRESS", "node", "8080", "(not in use)"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("list output missing %q\n---\n%s", want, stdout)
		}
	}
}

func TestWriteKillJSON(t *testing.T) {
	results := []killer.Result{
		{Port: 8080, Status: killer.StatusKilled, Processes: []killer.Process{{PID: 1, Name: "node", Port: 8080, Protocol: "tcp"}}},
		{Port: 22, Status: killer.StatusFailed, Err: errString("denied")},
		{Port: 3000, Status: killer.StatusNotInUse},
	}
	var buf bytes.Buffer
	writeKillJSON(&buf, results)

	var got []jsonKillResult
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
	}
	if len(got) != 3 {
		t.Fatalf("got %d results, want 3", len(got))
	}
	if got[0].Status != "killed" || len(got[0].Processes) != 1 || got[0].Processes[0].Name != "node" {
		t.Errorf("unexpected first result: %+v", got[0])
	}
	if got[1].Status != "failed" || got[1].Error != "denied" {
		t.Errorf("unexpected failed result: %+v", got[1])
	}
	// not_in_use must serialize processes as an empty array, never null.
	if got[2].Processes == nil {
		t.Errorf("processes should be [] not null for not_in_use")
	}
}

// errString is a tiny error helper for tests.
type errString string

func (e errString) Error() string { return string(e) }
