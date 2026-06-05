// Package ui handles terminal presentation: status symbols and optional ANSI
// color. Color is enabled only when writing to a real terminal, honoring the
// NO_COLOR convention (https://no-color.org) and an explicit opt-out.
package ui

import (
	"io"
	"os"
)

// ANSI SGR codes used for styling. Empty when color is disabled.
type palette struct {
	reset, bold, dim, red, green, yellow, blue, cyan string
}

// Styler renders status lines. Construct one with New.
type Styler struct {
	p palette
}

// New returns a Styler. Color is on when w is a terminal, NO_COLOR is unset,
// and forceOff is false. On Windows it also enables virtual-terminal output.
func New(w io.Writer, forceOff bool) *Styler {
	if colorEnabled(w, forceOff) {
		return &Styler{p: palette{
			reset:  "\033[0m",
			bold:   "\033[1m",
			dim:    "\033[2m",
			red:    "\033[31m",
			green:  "\033[32m",
			yellow: "\033[33m",
			blue:   "\033[34m",
			cyan:   "\033[36m",
		}}
	}
	return &Styler{} // zero palette => no escapes
}

func colorEnabled(w io.Writer, forceOff bool) bool {
	if forceOff {
		return false
	}
	if _, ok := os.LookupEnv("NO_COLOR"); ok {
		return false
	}
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	if !isTerminal(f) {
		return false
	}
	return enableVirtualTerminal(f)
}

func isTerminal(f *os.File) bool {
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

func (s *Styler) wrap(code, text string) string {
	if code == "" {
		return text
	}
	return code + text + s.p.reset
}

// Color helpers. Each is a no-op passthrough when color is disabled.
func (s *Styler) Bold(t string) string   { return s.wrap(s.p.bold, t) }
func (s *Styler) Dim(t string) string    { return s.wrap(s.p.dim, t) }
func (s *Styler) Red(t string) string    { return s.wrap(s.p.red, t) }
func (s *Styler) Green(t string) string  { return s.wrap(s.p.green, t) }
func (s *Styler) Yellow(t string) string { return s.wrap(s.p.yellow, t) }
func (s *Styler) Blue(t string) string   { return s.wrap(s.p.blue, t) }
func (s *Styler) Cyan(t string) string   { return s.wrap(s.p.cyan, t) }

// Status symbols, colorized when possible.
func (s *Styler) Scan() string    { return s.Blue("→") }
func (s *Styler) Found() string   { return s.Yellow("⚡") }
func (s *Styler) Success() string { return s.Green("✓") }
func (s *Styler) Info() string    { return s.Cyan("ℹ") }
func (s *Styler) Fail() string    { return s.Red("✗") }
func (s *Styler) Bullet() string  { return s.Dim("•") }
