//go:build windows

package killer

import (
	"fmt"
	"strconv"
	"strings"
	"syscall"
)

const (
	// processQueryLimitedInformation is enough to read a process exit code.
	processQueryLimitedInformation = 0x1000
	// stillActive (STILL_ACTIVE) is the exit code reported for live processes.
	stillActive = 259
)

// killProcess terminates a process tree with taskkill. Without force it asks
// politely (/T); with force it adds /F. The orchestration escalates as needed.
func killProcess(pid int, force bool) error {
	args := []string{"/PID", strconv.Itoa(pid), "/T"}
	if force {
		args = append(args, "/F")
	}
	_, errStr, err := runCommand("taskkill", args...)
	if err != nil {
		return fmt.Errorf("%v: %s", err, strings.TrimSpace(errStr))
	}
	return nil
}

// processAlive reports whether pid exists by opening it and checking its exit
// code. A PID that can no longer be opened is considered gone.
func processAlive(pid int) bool {
	handle, err := syscall.OpenProcess(processQueryLimitedInformation, false, uint32(pid))
	if err != nil {
		return false
	}
	defer syscall.CloseHandle(handle)

	var code uint32
	if err := syscall.GetExitCodeProcess(handle, &code); err != nil {
		return false
	}
	return code == stillActive
}
