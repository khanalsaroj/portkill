package killer

import (
	"encoding/hex"
	"fmt"
	"net"
	"strconv"
	"strings"
)

// procSocket is one matching row from /proc/net/{tcp,udp}{,6}.
type procSocket struct {
	Inode    string
	Protocol string
	Address  string // decoded "ip:port", best effort
}

// TCP state code for LISTEN in /proc/net/tcp (hex).
const procStateListen = "0A"

// parseProcNet extracts sockets bound to port from the contents of a
// /proc/net/{tcp,udp}{,6} file. proto must be ProtoTCP or ProtoUDP. TCP rows
// are limited to the LISTEN state; UDP rows (which have no listen concept) are
// matched on port alone.
func parseProcNet(output, proto string, port int) []procSocket {
	var sockets []procSocket

	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 10 || fields[0] == "sl" {
			continue // header or short line
		}

		if proto == ProtoTCP && !strings.EqualFold(fields[3], procStateListen) {
			continue
		}

		local := fields[1]
		colon := strings.LastIndex(local, ":")
		if colon < 0 {
			continue
		}
		p, err := hexPort(local[colon+1:])
		if err != nil || p != port {
			continue
		}

		sockets = append(sockets, procSocket{
			Inode:    fields[9],
			Protocol: proto,
			Address:  decodeProcAddr(local),
		})
	}

	return sockets
}

// hexPort parses a 16-bit port written as upper-case hex (as in /proc/net/*).
func hexPort(s string) (int, error) {
	v, err := strconv.ParseUint(s, 16, 16)
	if err != nil {
		return 0, err
	}
	return int(v), nil
}

// decodeProcAddr turns a "HEXIP:HEXPORT" /proc address into a readable
// "ip:port" string. IPv4 (8 hex chars) and IPv6 (32 hex chars) are supported;
// anything unexpected is returned unchanged.
func decodeProcAddr(addr string) string {
	colon := strings.LastIndex(addr, ":")
	if colon < 0 {
		return addr
	}
	ipHex, portHex := addr[:colon], addr[colon+1:]

	port, err := hexPort(portHex)
	if err != nil {
		return addr
	}

	raw, err := hex.DecodeString(ipHex)
	if err != nil {
		return addr
	}

	// Each little-endian 32-bit word must be byte-reversed to recover the IP.
	for i := 0; i+4 <= len(raw); i += 4 {
		raw[i], raw[i+3] = raw[i+3], raw[i]
		raw[i+1], raw[i+2] = raw[i+2], raw[i+1]
	}

	switch len(raw) {
	case 4:
		return fmt.Sprintf("%s:%d", net.IP(raw), port)
	case 16:
		return fmt.Sprintf("[%s]:%d", net.IP(raw), port)
	default:
		return addr
	}
}
