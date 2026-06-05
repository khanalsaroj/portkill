package killer

import (
	"strconv"
	"strings"
)

// parseSockstat extracts processes bound to port from `sockstat -l` output
// (FreeBSD). Columns: USER COMMAND PID FD PROTO LOCAL-ADDRESS FOREIGN-ADDRESS.
// sockstat -l already restricts output to listening sockets.
func parseSockstat(output string, port int) []Process {
	var procs []Process

	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 7 || fields[0] == "USER" {
			continue
		}

		var proto string
		switch {
		case strings.HasPrefix(fields[4], "tcp"):
			proto = ProtoTCP
		case strings.HasPrefix(fields[4], "udp"):
			proto = ProtoUDP
		default:
			continue
		}

		p, ok := endpointPort(fields[5])
		if !ok || p != port {
			continue
		}

		pid, err := strconv.Atoi(fields[2])
		if err != nil {
			continue
		}
		procs = append(procs, Process{
			PID:      pid,
			Name:     fields[1],
			Port:     port,
			Protocol: proto,
			Address:  fields[5],
		})
	}

	return procs
}
