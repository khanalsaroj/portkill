package killer

import (
	"reflect"
	"sort"
	"testing"
)

func TestEndpointPort(t *testing.T) {
	tests := []struct {
		addr     string
		wantPort int
		wantOK   bool
	}{
		{"0.0.0.0:135", 135, true},
		{"127.0.0.1:8080", 8080, true},
		{"[::]:135", 135, true},
		{"[::1]:8080", 8080, true},
		{"192.168.1.5:50123", 50123, true},
		{"0.0.0.0:65535", 65535, true},
		{"*:*", 0, false},
		{"noport", 0, false},
		{"host:", 0, false},
		{"host:0", 0, false},     // 0 is out of range
		{"host:65536", 0, false}, // above range
		{"host:abc", 0, false},   // non-numeric
	}
	for _, tc := range tests {
		gotPort, gotOK := endpointPort(tc.addr)
		if gotPort != tc.wantPort || gotOK != tc.wantOK {
			t.Errorf("endpointPort(%q) = (%d, %v), want (%d, %v)",
				tc.addr, gotPort, gotOK, tc.wantPort, tc.wantOK)
		}
	}
}

const netstatSample = `
Active Connections

  Proto  Local Address          Foreign Address        State           PID
  TCP    0.0.0.0:135            0.0.0.0:0              LISTENING       1980
  TCP    0.0.0.0:445            0.0.0.0:0              LISTENING       4
  TCP    127.0.0.1:8080         0.0.0.0:0              LISTENING       12345
  TCP    [::]:135               [::]:0                 LISTENING       1980
  TCP    192.168.1.5:50123      93.184.216.34:443     ESTABLISHED     6789
  TCP    127.0.0.1:18080        0.0.0.0:0              LISTENING       777
  UDP    0.0.0.0:500            *:*                                    2222
  UDP    [::]:500               *:*                                    2222
`

func pids(procs []Process) []int {
	if len(procs) == 0 {
		return nil
	}
	out := make([]int, len(procs))
	for i, p := range procs {
		out[i] = p.PID
	}
	sort.Ints(out)
	return out
}

func TestParseNetstat(t *testing.T) {
	tests := []struct {
		name     string
		port     int
		wantPIDs []int
	}{
		{"tcp ipv4+ipv6 listening", 135, []int{1980, 1980}},
		{"tcp single listener", 8080, []int{12345}},
		{"udp ipv4+ipv6", 500, []int{2222, 2222}},
		{"established is not a listener", 443, nil},
		{"ephemeral local of established conn ignored", 50123, nil},
		{"port 80 must not match 8080 or 18080", 80, nil},
		{"unused port", 9999, nil},
		// 18080 must be matched exactly, not by the :8080 substring.
		{"exact match 18080", 18080, []int{777}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := pids(parseNetstat(netstatSample, tc.port))
			want := tc.wantPIDs
			sort.Ints(want)
			if !reflect.DeepEqual(got, want) {
				t.Errorf("parseNetstat(port=%d) PIDs = %v, want %v", tc.port, got, want)
			}
		})
	}
}

func TestParseNetstatFields(t *testing.T) {
	procs := parseNetstat(netstatSample, 8080)
	if len(procs) != 1 {
		t.Fatalf("got %d processes, want 1", len(procs))
	}
	p := procs[0]
	if p.PID != 12345 || p.Port != 8080 || p.Protocol != ProtoTCP || p.Address != "127.0.0.1:8080" {
		t.Errorf("unexpected process: %+v", p)
	}
}

func TestParseTasklist(t *testing.T) {
	out := `"svchost.exe","1980","Services","0","12,345 K"
"System","4","Services","0","1,234 K"
"node.exe","12345","Console","1","98,765 K"
"weird"
`
	got := parseTasklist(out)
	want := map[int]string{1980: "svchost", 4: "System", 12345: "node"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("parseTasklist() = %v, want %v", got, want)
	}
}
