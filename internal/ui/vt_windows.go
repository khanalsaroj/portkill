//go:build windows

package ui

import (
	"os"
	"syscall"
	"unsafe"
)

// enableVirtualTerminalProcessing lets the console interpret ANSI escapes.
const enableVirtualTerminalProcessing = 0x0004

var (
	kernel32           = syscall.NewLazyDLL("kernel32.dll")
	procGetConsoleMode = kernel32.NewProc("GetConsoleMode")
	procSetConsoleMode = kernel32.NewProc("SetConsoleMode")
)

// enableVirtualTerminal turns on VT processing for the given console handle so
// that ANSI color escapes render correctly on Windows 10+. Returns false (and
// the caller falls back to plain text) on older consoles that reject it.
func enableVirtualTerminal(f *os.File) bool {
	handle := syscall.Handle(f.Fd())

	var mode uint32
	if r, _, _ := procGetConsoleMode.Call(uintptr(handle), uintptr(unsafe.Pointer(&mode))); r == 0 {
		return false
	}
	if mode&enableVirtualTerminalProcessing != 0 {
		return true // already enabled
	}
	r, _, _ := procSetConsoleMode.Call(uintptr(handle), uintptr(mode|enableVirtualTerminalProcessing))
	return r != 0
}
