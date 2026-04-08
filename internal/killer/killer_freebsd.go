//go:build freebsd

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
	if pid, err := findPIDViaSockstat(port); err == nil {
		return pid, nil
	}

	if _, err := exec.LookPath("lsof"); err == nil {
		return findPIDViaLsof(port)
	}

	return 0, fmt.Errorf("%w %d", ErrPortNotInUse, port)
}

func findPIDViaSockstat(port int) (int, error) {
	portStr := strconv.Itoa(port)
	for _, proto := range []string{"-4", "-6"} {
		cmd := exec.Command("sockstat", "-l", proto, "-p", portStr)

		var out bytes.Buffer
		var errBuf bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &errBuf

		_ = cmd.Run()

		pid, found := parseSockstatOutput(out.String(), port)
		if found {
			return pid, nil
		}
	}

	return 0, fmt.Errorf("%w %d", ErrPortNotInUse, port)
}

func parseSockstatOutput(output string, port int) (int, bool) {
	target := fmt.Sprintf(":%d", port)

	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)

		// Skip header and blank lines.
		if line == "" || strings.HasPrefix(line, "USER") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 7 {
			continue
		}

		localAddr := fields[5] // e.g. "*:8080" or "127.0.0.1:8080"
		if !strings.HasSuffix(localAddr, target) {
			continue
		}

		pid, err := strconv.Atoi(fields[2])
		if err != nil {
			continue
		}
		return pid, true
	}

	return 0, false
}

func findPIDViaLsof(port int) (int, error) {
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
