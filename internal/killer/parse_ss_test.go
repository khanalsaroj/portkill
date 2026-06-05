package killer

import (
	"reflect"
	"testing"
)

const ssSample = `tcp   LISTEN 0      128          0.0.0.0:22         0.0.0.0:*    users:(("sshd",pid=1234,fd=3))
tcp   LISTEN 0      511          127.0.0.1:8080     0.0.0.0:*    users:(("node",pid=5678,fd=20))
tcp   LISTEN 0      128          [::]:22            [::]:*       users:(("sshd",pid=1234,fd=4))
udp   UNCONN 0      0            0.0.0.0:53         0.0.0.0:*    users:(("dnsmasq",pid=999,fd=4))
tcp   ESTAB  0      0            10.0.0.1:8080      10.0.0.2:55  users:(("node",pid=5678,fd=22))
tcp   LISTEN 0      128          0.0.0.0:9090       0.0.0.0:*    users:(("envoy",pid=10,fd=3),("envoy",pid=11,fd=5))
udp   UNCONN 0      0            0.0.0.0:68         0.0.0.0:*
`

func TestParseSS(t *testing.T) {
	tests := []struct {
		name     string
		port     int
		wantPIDs []int
	}{
		{"listening tcp only, not established", 8080, []int{5678}},
		{"two sshd sockets ipv4+ipv6", 22, []int{1234, 1234}},
		{"udp", 53, []int{999}},
		{"socket shared by two pids", 9090, []int{10, 11}},
		{"no process info yields nothing", 68, nil},
		{"unused", 1, nil},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := pids(parseSS(ssSample, tc.port)); !reflect.DeepEqual(got, tc.wantPIDs) {
				t.Errorf("parseSS(%d) PIDs = %v, want %v", tc.port, got, tc.wantPIDs)
			}
		})
	}
}

func TestParseSSFields(t *testing.T) {
	procs := parseSS(ssSample, 53)
	if len(procs) != 1 {
		t.Fatalf("got %d, want 1", len(procs))
	}
	p := procs[0]
	if p.Name != "dnsmasq" || p.Protocol != ProtoUDP || p.Address != "0.0.0.0:53" {
		t.Errorf("unexpected: %+v", p)
	}
}
