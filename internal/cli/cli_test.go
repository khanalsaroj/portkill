package cli

import (
	"bytes"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestParseArgsPorts(t *testing.T) {
	tests := []struct {
		name  string
		args  []string
		ports []int
	}{
		{"single", []string{"8080"}, []int{8080}},
		{"multiple", []string{"3000", "8080", "9090"}, []int{3000, 8080, 9090}},
		{"flag before port", []string{"-f", "8080"}, []int{8080}},
		{"flag after port", []string{"8080", "-f"}, []int{8080}},
		{"interleaved", []string{"8080", "-n", "3000"}, []int{8080, 3000}},
		{"double dash terminator", []string{"--", "8080"}, []int{8080}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg, err := parseArgs(tc.args, true)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(cfg.ports, tc.ports) {
				t.Errorf("ports = %v, want %v", cfg.ports, tc.ports)
			}
		})
	}
}

func TestParseArgsFlags(t *testing.T) {
	cfg, err := parseArgs([]string{"-f", "-n", "--json", "--quiet", "--no-color", "--timeout", "5s", "8080"}, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.force || !cfg.dryRun || !cfg.json || !cfg.quiet || !cfg.noColor {
		t.Errorf("boolean flags not all set: %+v", cfg)
	}
	if cfg.timeout != 5*time.Second {
		t.Errorf("timeout = %v, want 5s", cfg.timeout)
	}
}

func TestParseArgsTimeoutInline(t *testing.T) {
	cfg, err := parseArgs([]string{"--timeout=250ms", "80"}, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.timeout != 250*time.Millisecond {
		t.Errorf("timeout = %v, want 250ms", cfg.timeout)
	}
}

func TestParseArgsErrors(t *testing.T) {
	tests := []struct {
		name string
		args []string
		kill bool
	}{
		{"unknown flag", []string{"--bogus", "80"}, true},
		{"invalid port high", []string{"99999"}, true},
		{"invalid port zero", []string{"0"}, true},
		{"invalid port text", []string{"abc"}, true},
		{"bad timeout", []string{"--timeout", "soon", "80"}, true},
		{"negative timeout", []string{"--timeout", "-3s", "80"}, true},
		{"timeout missing value", []string{"--timeout"}, true},
		{"force rejected on list", []string{"-f", "80"}, false},
		{"dry-run rejected on list", []string{"-n", "80"}, false},
		{"timeout rejected on list", []string{"--timeout", "3s", "80"}, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := parseArgs(tc.args, tc.kill); err == nil {
				t.Errorf("expected error for args %v (kill=%v)", tc.args, tc.kill)
			}
		})
	}
}

func TestParseArgsHelp(t *testing.T) {
	if _, err := parseArgs([]string{"-h"}, true); !errors.Is(err, errHelp) {
		t.Errorf("-h should return errHelp, got %v", err)
	}
	if _, err := parseArgs([]string{"8080", "--help"}, true); !errors.Is(err, errHelp) {
		t.Errorf("--help should return errHelp, got %v", err)
	}
}

// runOut is a helper that runs the CLI against buffers and returns code+output.
func runOut(args ...string) (code int, stdout, stderr string) {
	var out, errBuf bytes.Buffer
	code = run(args, &out, &errBuf)
	return code, out.String(), errBuf.String()
}

func TestRunDispatch(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantCode   int
		wantStdout string // substring, "" to skip
		wantStderr string // substring, "" to skip
	}{
		{"version", []string{"version"}, exitOK, "portkill", ""},
		{"version flag", []string{"--version"}, exitOK, "portkill", ""},
		{"version short", []string{"-v"}, exitOK, "portkill", ""},
		{"help", []string{"help"}, exitOK, "USAGE", ""},
		{"help flag", []string{"--help"}, exitOK, "USAGE", ""},
		{"no args", nil, exitUsage, "", "USAGE"},
		{"unknown command", []string{"frob"}, exitUsage, "", "unknown command"},
		{"kill no port", []string{"kill"}, exitUsage, "", "at least one port"},
		{"kill invalid port", []string{"kill", "99999"}, exitUsage, "", "not a valid port"},
		{"kill unknown flag", []string{"kill", "--zz", "80"}, exitUsage, "", "unknown flag"},
		{"list no port", []string{"list"}, exitUsage, "", "at least one port"},
		{"list force rejected", []string{"list", "-f", "80"}, exitUsage, "", "only valid for 'kill'"},
		{"kill help", []string{"kill", "--help"}, exitOK, "USAGE", ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			code, stdout, stderr := runOut(tc.args...)
			if code != tc.wantCode {
				t.Errorf("exit code = %d, want %d", code, tc.wantCode)
			}
			if tc.wantStdout != "" && !strings.Contains(stdout, tc.wantStdout) {
				t.Errorf("stdout = %q, want substring %q", stdout, tc.wantStdout)
			}
			if tc.wantStderr != "" && !strings.Contains(stderr, tc.wantStderr) {
				t.Errorf("stderr = %q, want substring %q", stderr, tc.wantStderr)
			}
		})
	}
}
