//go:build linux || darwin

package killer

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

func findPID(port int) (int, error) {
	cmd := exec.Command("lsof", "-t", "-i", fmt.Sprintf("tcp:%d", port))

	var out bytes.Buffer
	var errBuf bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errBuf

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return 0, fmt.Errorf("%w %d", ErrPortNotInUse, port)
		}
		return 0, fmt.Errorf("lsof error: %w — %s", err, errBuf.String())
	}

	pidStr := strings.TrimSpace(strings.SplitN(out.String(), "\n", 2)[0])
	if pidStr == "" {
		return 0, fmt.Errorf("%w %d", ErrPortNotInUse, port)
	}

	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return 0, fmt.Errorf("could not parse PID %q: %w", pidStr, err)
	}

	return pid, nil
}

func killPID(pid int) error {
	if err := syscall.Kill(pid, syscall.SIGTERM); err != nil {
		return err
	}
	return nil
}
