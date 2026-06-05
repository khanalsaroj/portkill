package killer

import (
	"reflect"
	"testing"
)

const sockstatSample = `USER     COMMAND    PID   FD PROTO  LOCAL ADDRESS         FOREIGN ADDRESS
root     sshd       1234  4  tcp4   *:22                  *:*
www      nginx      5678  6  tcp4   *:80                  *:*
www      nginx      5678  7  tcp6   *:80                  *:*
bind     named      999   5  udp4   127.0.0.1:53          *:*
`

func TestParseSockstat(t *testing.T) {
	tests := []struct {
		name     string
		port     int
		wantPIDs []int
	}{
		{"nginx ipv4+ipv6", 80, []int{5678, 5678}},
		{"sshd", 22, []int{1234}},
		{"named udp", 53, []int{999}},
		{"unused", 8080, nil},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := pids(parseSockstat(sockstatSample, tc.port)); !reflect.DeepEqual(got, tc.wantPIDs) {
				t.Errorf("parseSockstat(%d) = %v, want %v", tc.port, got, tc.wantPIDs)
			}
		})
	}
}

func TestParseSockstatFields(t *testing.T) {
	procs := parseSockstat(sockstatSample, 53)
	if len(procs) != 1 {
		t.Fatalf("got %d, want 1", len(procs))
	}
	if p := procs[0]; p.Name != "named" || p.Protocol != ProtoUDP || p.PID != 999 {
		t.Errorf("unexpected: %+v", p)
	}
}
