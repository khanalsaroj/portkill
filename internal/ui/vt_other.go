//go:build !windows

package ui

import "os"

// enableVirtualTerminal is a no-op on Unix; ANSI is natively understood.
func enableVirtualTerminal(*os.File) bool { return true }
