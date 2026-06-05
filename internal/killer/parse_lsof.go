package killer

import (
	"strconv"
	"strings"
)

// parseLsof extracts processes bound to port from `lsof -nP -i:<port>` output.
// Columns are: COMMAND PID USER FD TYPE DEVICE SIZE/OFF NODE NAME [STATE].
// For TCP only listening sockets are returned; all UDP sockets are returned.
func parseLsof(output string, port int) []Process {
	var procs []Process

	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 9 || fields[0] == "COMMAND" {
			continue
		}

		var proto string
		switch strings.ToUpper(fields[7]) {
		case "TCP":
			proto = ProtoTCP
		case "UDP":
			proto = ProtoUDP
		default:
			continue
		}

		// The NAME column holds the endpoint; an established connection looks
		// like "local->remote" — keep the local side only.
		endpoint := fields[8]
		if i := strings.Index(endpoint, "->"); i >= 0 {
			endpoint = endpoint[:i]
		}
		p, ok := endpointPort(endpoint)
		if !ok || p != port {
			continue
		}

		if proto == ProtoTCP {
			state := ""
			if len(fields) > 9 {
				state = fields[9]
			}
			if !strings.Contains(strings.ToUpper(state), "LISTEN") {
				continue
			}
		}

		pid, err := strconv.Atoi(fields[1])
		if err != nil {
			continue
		}
		procs = append(procs, Process{
			PID:      pid,
			Name:     fields[0],
			Port:     port,
			Protocol: proto,
			Address:  endpoint,
		})
	}

	return procs
}
