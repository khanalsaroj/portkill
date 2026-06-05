package ui

import (
	"bytes"
	"strings"
	"testing"
)

// A bytes.Buffer is not an *os.File, so color is always disabled when writing
// to one — exactly what we want when output is piped or captured.
func TestStylerNoColorOnNonTerminal(t *testing.T) {
	s := New(&bytes.Buffer{}, false)
	if got := s.Bold("hi"); got != "hi" {
		t.Errorf("Bold on non-terminal = %q, want plain %q", got, "hi")
	}
	if got := s.Red("err"); got != "err" {
		t.Errorf("Red on non-terminal = %q, want plain %q", got, "err")
	}
}

func TestStylerForceOff(t *testing.T) {
	s := New(&bytes.Buffer{}, true)
	for name, got := range map[string]string{
		"Green":  s.Green("x"),
		"Yellow": s.Yellow("x"),
		"Cyan":   s.Cyan("x"),
		"Dim":    s.Dim("x"),
		"Blue":   s.Blue("x"),
	} {
		if got != "x" {
			t.Errorf("%s should be plain when disabled, got %q", name, got)
		}
	}
}

func TestStatusSymbolsPlain(t *testing.T) {
	s := New(&bytes.Buffer{}, true)
	symbols := []struct {
		name string
		got  string
		want string
	}{
		{"Scan", s.Scan(), "→"},
		{"Found", s.Found(), "⚡"},
		{"Success", s.Success(), "✓"},
		{"Info", s.Info(), "ℹ"},
		{"Fail", s.Fail(), "✗"},
		{"Bullet", s.Bullet(), "•"},
	}
	for _, sym := range symbols {
		if sym.got != sym.want {
			t.Errorf("%s = %q, want %q", sym.name, sym.got, sym.want)
		}
		// No ANSI escapes should ever leak when color is disabled.
		if strings.Contains(sym.got, "\033") {
			t.Errorf("%s leaked an escape sequence: %q", sym.name, sym.got)
		}
	}
}
