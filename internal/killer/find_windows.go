//go:build windows

package killer

import (
	"fmt"
	"strings"
)

// findProcesses lists processes bound to port using `netstat -ano`, enriching
// each with its image name via a single `tasklist` call (best effort).
func findProcesses(port int) ([]Process, error) {
	out, errStr, err := runCommand("netstat", "-ano")
	if err != nil {
		return nil, fmt.Errorf("netstat failed: %v: %s", err, strings.TrimSpace(errStr))
	}

	procs := parseNetstat(out, port)
	if len(procs) == 0 {
		return nil, nil
	}

	if names := processNames(); len(names) > 0 {
		for i := range procs {
			if name, ok := names[procs[i].PID]; ok {
				procs[i].Name = name
			}
		}
	}
	return procs, nil
}

// processNames returns a PID->image-name map from tasklist. Best effort: an
// error yields an empty map and detection proceeds without names.
func processNames() map[int]string {
	out, _, err := runCommand("tasklist", "/FO", "CSV", "/NH")
	if err != nil {
		return nil
	}
	return parseTasklist(out)
}
