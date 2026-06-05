package killer

import (
	"encoding/csv"
	"strconv"
	"strings"
)

// endpointPort extracts the numeric port from a "host:port" endpoint, coping
// with IPv6 forms such as "[::]:135" and "[::1]:8080" by taking the text after
// the final colon. It returns false when no numeric port is present.
func endpointPort(addr string) (int, bool) {
	i := strings.LastIndex(addr, ":")
	if i < 0 || i == len(addr)-1 {
		return 0, false
	}
	port, err := strconv.Atoi(addr[i+1:])
	if err != nil || port < 1 || port > 65535 {
		return 0, false
	}
	return port, true
}

// parseNetstat extracts processes bound to port from `netstat -ano` output.
// It matches TCP sockets in the LISTENING state and all UDP sockets, returning
// one Process per matching row (names are resolved separately).
func parseNetstat(output string, port int) []Process {
	var procs []Process

	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(strings.TrimSpace(line))
		if len(fields) < 4 {
			continue
		}

		var proto string
		switch strings.ToUpper(fields[0]) {
		case "TCP":
			proto = ProtoTCP
		case "UDP":
			proto = ProtoUDP
		default:
			continue
		}

		local := fields[1]
		p, ok := endpointPort(local)
		if !ok || p != port {
			continue
		}

		// TCP rows carry a state and a trailing PID (5 columns); only
		// listening sockets own the port. UDP rows have no state (4 columns).
		var pidField string
		if proto == ProtoTCP {
			if len(fields) < 5 || !strings.EqualFold(fields[3], "LISTENING") {
				continue
			}
			pidField = fields[4]
		} else {
			pidField = fields[len(fields)-1]
		}

		pid, err := strconv.Atoi(pidField)
		if err != nil {
			continue
		}
		procs = append(procs, Process{
			PID:      pid,
			Port:     port,
			Protocol: proto,
			Address:  local,
		})
	}

	return procs
}

// parseTasklist builds a PID->image-name map from `tasklist /FO CSV /NH`
// output. Malformed rows are skipped.
func parseTasklist(output string) map[int]string {
	names := make(map[int]string)

	r := csv.NewReader(strings.NewReader(output))
	r.FieldsPerRecord = -1 // rows may vary; we only need the first two columns
	records, err := r.ReadAll()
	if err != nil {
		return names
	}

	for _, rec := range records {
		if len(rec) < 2 {
			continue
		}
		pid, err := strconv.Atoi(strings.TrimSpace(rec[1]))
		if err != nil {
			continue
		}
		names[pid] = strings.TrimSuffix(strings.TrimSpace(rec[0]), ".exe")
	}

	return names
}
