//go:build freebsd

package killer

import (
	"fmt"
	"os/exec"
	"strconv"
)

// findProcesses lists processes bound to port on FreeBSD, preferring the
// base-system sockstat and falling back to lsof when it is installed.
func findProcesses(port int) ([]Process, error) {
	var procs []Process
	for _, family := range []string{"-4", "-6"} {
		out, _, _ := runCommand("sockstat", "-l", family, "-p", strconv.Itoa(port))
		procs = append(procs, parseSockstat(out, port)...)
	}
	if len(procs) > 0 {
		return procs, nil
	}

	if _, err := exec.LookPath("lsof"); err == nil {
		out, _, _ := runCommand("lsof", "-nP", "-i", fmt.Sprintf(":%d", port))
		return parseLsof(out, port), nil
	}
	return nil, nil
}
