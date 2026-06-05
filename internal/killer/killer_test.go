package killer

import (
	"os"
	"reflect"
	"testing"
)

func TestTidy(t *testing.T) {
	self := os.Getpid()
	in := []Process{
		{PID: 0, Name: "idle"},        // dropped: PID 0
		{PID: -1, Name: "bogus"},      // dropped: negative
		{PID: self, Name: "portkill"}, // dropped: self
		{PID: 30, Name: "c"},
		{PID: 10, Protocol: ProtoTCP}, // merged with the name-only dup below
		{PID: 10, Name: "a", Address: "0.0.0.0:80"},
		{PID: 20, Name: "b"},
		{PID: 20, Name: "b"}, // duplicate collapsed
	}
	got := tidy(in)

	wantPIDs := []int{10, 20, 30}
	if g := pids(got); !reflect.DeepEqual(g, wantPIDs) {
		t.Fatalf("tidy PIDs = %v, want %v", g, wantPIDs)
	}

	// PID 10 should carry merged fields from both records.
	var p10 Process
	for _, p := range got {
		if p.PID == 10 {
			p10 = p
		}
	}
	if p10.Name != "a" || p10.Protocol != ProtoTCP || p10.Address != "0.0.0.0:80" {
		t.Errorf("merged PID 10 = %+v, want name=a proto=tcp addr=0.0.0.0:80", p10)
	}
}

func TestMerge(t *testing.T) {
	a := Process{PID: 1}
	b := Process{PID: 1, Name: "x", Protocol: ProtoUDP, Address: "a:1"}
	got := merge(a, b)
	if got.Name != "x" || got.Protocol != ProtoUDP || got.Address != "a:1" {
		t.Errorf("merge filled wrong: %+v", got)
	}

	// Existing values must win over the duplicate's.
	c := Process{PID: 1, Name: "keep", Protocol: ProtoTCP, Address: "keep:1"}
	got = merge(c, b)
	if got.Name != "keep" || got.Protocol != ProtoTCP || got.Address != "keep:1" {
		t.Errorf("merge overwrote existing: %+v", got)
	}
}

func TestDisplayName(t *testing.T) {
	if got := displayName(Process{Name: "node"}); got != "node" {
		t.Errorf("got %q", got)
	}
	if got := displayName(Process{}); got != "process" {
		t.Errorf("empty name = %q, want process", got)
	}
}
