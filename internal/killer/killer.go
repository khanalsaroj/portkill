package killer

import (
	"errors"
	"fmt"
)

var ErrPortNotInUse = errors.New("no process found on port")

func KillPort(port int) error {
	fmt.Printf("  →  scanning port %d ...\n", port)

	pid, err := findPID(port)
	if err != nil {
		if errors.Is(err, ErrPortNotInUse) {
			fmt.Printf("  ℹ  port %d is not in use\n", port)
			return nil
		}
		return fmt.Errorf("could not find PID: %w", err)
	}

	fmt.Printf("  ⚡  found PID %d on port %d — terminating ...\n", pid, port)

	if err := killPID(pid); err != nil {
		return fmt.Errorf("could not kill PID %d: %w (try running with elevated permissions)", pid, err)
	}

	fmt.Printf("  ✓  PID %d on port %d terminated successfully\n", pid, port)
	return nil
}
