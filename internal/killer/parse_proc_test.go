package killer

import "testing"

func TestHexPort(t *testing.T) {
	tests := []struct {
		in      string
		want    int
		wantErr bool
	}{
		{"1538", 5432, false},
		{"0050", 80, false},
		{"FFFF", 65535, false},
		{"0", 0, false},
		{"GG", 0, true},
		{"", 0, true},
	}
	for _, tc := range tests {
		got, err := hexPort(tc.in)
		if (err != nil) != tc.wantErr || got != tc.want {
			t.Errorf("hexPort(%q) = (%d, %v), want (%d, err=%v)", tc.in, got, err, tc.want, tc.wantErr)
		}
	}
}

func TestDecodeProcAddr(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"0100007F:1538", "127.0.0.1:5432"},
		{"00000000:0050", "0.0.0.0:80"},
		{"00000000000000000000000001000000:0050", "[::1]:80"},
		{"garbage", "garbage"}, // no colon -> returned unchanged
		{"ZZ:0050", "ZZ:0050"}, // bad hex IP -> unchanged
	}
	for _, tc := range tests {
		if got := decodeProcAddr(tc.in); got != tc.want {
			t.Errorf("decodeProcAddr(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

const procTCPSample = `  sl  local_address rem_address   st tx_queue rx_queue tr tm->when retrnsmt   uid  timeout inode
   0: 00000000:0050 00000000:0000 0A 00000000:00000000 00:00000000 00000000     0        0 12345 1 0000000000000000 100 0 0 10 0
   1: 0100007F:1538 00000000:0000 0A 00000000:00000000 00:00000000 00000000  1000        0 67890 1 0000000000000000 100 0 0 10 0
   2: 0100007F:0035 0100007F:1234 01 00000000:00000000 00:00000000 00000000  1000        0 11111 1 0000000000000000 100 0 0 10 0
`

const procUDPSample = `  sl  local_address rem_address   st tx_queue rx_queue tr tm->when retrnsmt   uid  timeout inode ref pointer drops
   0: 00000000:0035 00000000:0000 07 00000000:00000000 00:00000000 00000000     0        0 22222 2 0000000000000000 0
`

func TestParseProcNet(t *testing.T) {
	t.Run("tcp listen matched", func(t *testing.T) {
		got := parseProcNet(procTCPSample, ProtoTCP, 80)
		if len(got) != 1 || got[0].Inode != "12345" || got[0].Address != "0.0.0.0:80" {
			t.Fatalf("got %+v", got)
		}
	})
	t.Run("tcp 5432 listen", func(t *testing.T) {
		got := parseProcNet(procTCPSample, ProtoTCP, 5432)
		if len(got) != 1 || got[0].Inode != "67890" {
			t.Fatalf("got %+v", got)
		}
	})
	t.Run("tcp non-listen excluded", func(t *testing.T) {
		// port 53 (0x0035) row is in state 01 (ESTABLISHED), not LISTEN.
		if got := parseProcNet(procTCPSample, ProtoTCP, 53); len(got) != 0 {
			t.Fatalf("expected no match, got %+v", got)
		}
	})
	t.Run("udp ignores state", func(t *testing.T) {
		got := parseProcNet(procUDPSample, ProtoUDP, 53)
		if len(got) != 1 || got[0].Inode != "22222" || got[0].Protocol != ProtoUDP {
			t.Fatalf("got %+v", got)
		}
	})
}
