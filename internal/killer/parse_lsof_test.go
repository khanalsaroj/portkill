package killer

import (
	"reflect"
	"testing"
)

const lsofSample = `COMMAND   PID  USER   FD   TYPE DEVICE SIZE/OFF NODE NAME
node    12345 saroj   23u  IPv4 0xabc      0t0  TCP *:8080 (LISTEN)
node    12345 saroj   24u  IPv6 0xdef      0t0  TCP [::1]:8080 (LISTEN)
chrome   6789 saroj   45u  IPv4 0x123      0t0  TCP 127.0.0.1:54321->127.0.0.1:8080 (ESTABLISHED)
dnsmasq   999 nobody   4u  IPv4 0x456      0t0  UDP *:8080
`

func TestParseLsof(t *testing.T) {
	procs := parseLsof(lsofSample, 8080)
	if got, want := pids(procs), []int{999, 12345, 12345}; !reflect.DeepEqual(got, want) {
		t.Fatalf("parseLsof PIDs = %v, want %v", got, want)
	}

	// Established connection toward :8080 must not be treated as a listener.
	for _, p := range procs {
		if p.PID == 6789 {
			t.Errorf("established connection (chrome) should have been excluded")
		}
	}

	// Names and protocols are captured.
	var sawUDP bool
	for _, p := range procs {
		if p.Protocol == ProtoUDP {
			sawUDP = true
			if p.Name != "dnsmasq" {
				t.Errorf("udp process name = %q, want dnsmasq", p.Name)
			}
		}
		if p.Protocol == ProtoTCP && p.Name != "node" {
			t.Errorf("tcp process name = %q, want node", p.Name)
		}
	}
	if !sawUDP {
		t.Errorf("expected a UDP entry")
	}
}

func TestParseLsofNoMatch(t *testing.T) {
	if got := parseLsof(lsofSample, 1234); len(got) != 0 {
		t.Errorf("parseLsof(1234) = %v, want empty", got)
	}
	if got := parseLsof("", 8080); len(got) != 0 {
		t.Errorf("parseLsof(empty) = %v, want empty", got)
	}
}
