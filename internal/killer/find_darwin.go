//go:build darwin

package killer

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// exitCode returns the process exit code for err, 0 for nil, or -1 when err is
// not an *exec.ExitError (e.g. the binary could not be started).
func exitCode(err error) int {
	if err == nil {
		return 0
	}
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		return ee.ExitCode()
	}
	return -1
}

// findProcesses lists processes bound to port on macOS via lsof, which ships
// with the OS. lsof exits 1 when nothing matches — that is not an error.
func findProcesses(port int) ([]Process, error) {
	if _, err := exec.LookPath("lsof"); err != nil {
		return nil, fmt.Errorf("lsof not found in PATH (required on macOS)")
	}

	out, errStr, err := runCommand("lsof", "-nP", "-i", fmt.Sprintf(":%d", port))
	procs := parseLsof(out, port)
	if len(procs) == 0 && exitCode(err) > 1 {
		return nil, fmt.Errorf("lsof failed: %s", strings.TrimSpace(errStr))
	}
	return procs, nil
}
