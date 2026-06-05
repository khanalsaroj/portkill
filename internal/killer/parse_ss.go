package killer

import (
	"regexp"
	"strconv"
	"strings"
)

// ssProcess matches each ("name",pid=NNN,...) tuple inside the users:(...)
// column emitted by `ss -p`. A socket may list several such tuples.
var ssProcess = regexp.MustCompile(`\("([^"]*)",pid=(\d+)`)

// parseSS extracts processes bound to port from `ss -tulnpH` output. Columns:
// Netid State Recv-Q Send-Q LocalAddress:Port PeerAddress:Port [Process].
// TCP sockets must be in LISTEN state; UDP sockets are accepted as-is.
func parseSS(output string, port int) []Process {
	var procs []Process

	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}

		var proto string
		switch strings.ToLower(fields[0]) {
		case "tcp":
			proto = ProtoTCP
		case "udp":
			proto = ProtoUDP
		default:
			continue
		}
		if proto == ProtoTCP && !strings.EqualFold(fields[1], "LISTEN") {
			continue
		}

		p, ok := endpointPort(fields[4])
		if !ok || p != port {
			continue
		}

		for _, m := range ssProcess.FindAllStringSubmatch(line, -1) {
			pid, err := strconv.Atoi(m[2])
			if err != nil {
				continue
			}
			procs = append(procs, Process{
				PID:      pid,
				Name:     m[1],
				Port:     port,
				Protocol: proto,
				Address:  fields[4],
			})
		}
	}

	return procs
}
