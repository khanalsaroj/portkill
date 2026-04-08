//go:build windows

package killer

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

func findPID(port int) (int, error) {
	cmd := exec.Command("netstat", "-ano")

	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return 0, fmt.Errorf("netstat error: %w", err)
	}

	target := fmt.Sprintf(":%d", port)
	for _, line := range strings.Split(out.String(), "\n") {
		line = strings.TrimSpace(line)

		if !strings.HasPrefix(line, "TCP") {
			continue
		}
		if !strings.Contains(line, "LISTENING") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}

		localAddr := fields[1]
		if !strings.HasSuffix(localAddr, target) {
			continue
		}

		pidStr := fields[4]
		pid, err := strconv.Atoi(pidStr)
		if err != nil {
			return 0, fmt.Errorf("could not parse PID %q: %w", pidStr, err)
		}
		return pid, nil
	}

	return 0, fmt.Errorf("%w %d", ErrPortNotInUse, port)
}

func killPID(pid int) error {
	cmd := exec.Command(
		"taskkill",
		"/PID", strconv.Itoa(pid),
		"/F",
		"/T",
	)

	var errBuf bytes.Buffer
	cmd.Stderr = &errBuf

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%w — %s", err, strings.TrimSpace(errBuf.String()))
	}
	return nil
}
