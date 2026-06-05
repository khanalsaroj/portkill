//go:build !windows

package killer

import (
	"errors"
	"syscall"
)

// killProcess sends SIGTERM (graceful) or SIGKILL (force) to pid. A process
// that has already exited is treated as success.
func killProcess(pid int, force bool) error {
	sig := syscall.SIGTERM
	if force {
		sig = syscall.SIGKILL
	}
	err := syscall.Kill(pid, sig)
	if err == nil || errors.Is(err, syscall.ESRCH) {
		return nil
	}
	return err
}

// processAlive reports whether pid currently exists. Signal 0 performs the
// kernel's permission/existence check without delivering a signal: nil means
// alive, EPERM means alive-but-not-ours, ESRCH means gone.
func processAlive(pid int) bool {
	err := syscall.Kill(pid, 0)
	if err == nil {
		return true
	}
	return errors.Is(err, syscall.EPERM)
}
