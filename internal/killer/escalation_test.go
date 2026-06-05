package killer

import (
	"errors"
	"testing"
	"time"
)

// fakeProc models a process for testing the escalation state machine. It dies
// (stops being "alive") after a configurable number of signals of the right
// kind, and can be made to reject graceful or all kills.
type fakeProc struct {
	killCalls     []bool // records the force flag of each kill request
	dieAfterKills int    // becomes dead once this many qualifying kills are sent
	requireForce  bool   // graceful (force=false) kills error out and do nothing
	gracefulNoop  bool   // graceful kills are accepted (nil) but never effective
	immortal      bool   // never dies, every kill errors
	qualifying    int
}

func (f *fakeProc) alive(int) bool {
	if f.immortal {
		return true
	}
	return f.qualifying < f.dieAfterKills
}

func (f *fakeProc) kill(_ int, force bool) error {
	f.killCalls = append(f.killCalls, force)
	if f.immortal {
		return errors.New("cannot kill")
	}
	if !force {
		if f.requireForce {
			return errors.New("can only be terminated forcefully (/F)")
		}
		if f.gracefulNoop {
			return nil // accepted, but the process ignores it
		}
	}
	f.qualifying++
	return nil
}

func withFake(t *testing.T, f *fakeProc) {
	t.Helper()
	origKill, origAlive, origPoll := killProcessFn, processAliveFn, pollInterval
	killProcessFn, processAliveFn = f.kill, f.alive
	pollInterval = time.Millisecond
	t.Cleanup(func() {
		killProcessFn, processAliveFn, pollInterval = origKill, origAlive, origPoll
	})
}

func countForce(calls []bool, force bool) int {
	n := 0
	for _, c := range calls {
		if c == force {
			n++
		}
	}
	return n
}

func TestKillAndWaitAlreadyDead(t *testing.T) {
	f := &fakeProc{immortal: false, dieAfterKills: 0} // dead from the start
	withFake(t, f)
	if err := killAndWait(123, Options{Timeout: 50 * time.Millisecond}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(f.killCalls) != 0 {
		t.Errorf("expected no kill calls for a dead process, got %v", f.killCalls)
	}
}

func TestKillAndWaitGracefulSucceeds(t *testing.T) {
	f := &fakeProc{dieAfterKills: 1} // one graceful signal is enough
	withFake(t, f)
	if err := killAndWait(1, Options{Timeout: 50 * time.Millisecond}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if countForce(f.killCalls, true) != 0 {
		t.Errorf("should not have escalated to force, calls=%v", f.killCalls)
	}
	if countForce(f.killCalls, false) == 0 {
		t.Errorf("expected at least one graceful kill, calls=%v", f.killCalls)
	}
}

func TestKillAndWaitEscalatesAfterGrace(t *testing.T) {
	// Graceful signals are accepted but never kill; only a forced kill does.
	f := &fakeProc{dieAfterKills: 1, requireForce: true}
	withFake(t, f)
	if err := killAndWait(1, Options{Timeout: 5 * time.Millisecond}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if countForce(f.killCalls, true) == 0 {
		t.Errorf("expected escalation to a forced kill, calls=%v", f.killCalls)
	}
}

func TestKillAndWaitGracefulIgnoredThenForced(t *testing.T) {
	// Graceful is accepted (no error) but ineffective; the deadline loop must
	// time out and escalate to a forced kill.
	f := &fakeProc{dieAfterKills: 1, gracefulNoop: true}
	withFake(t, f)
	if err := killAndWait(1, Options{Timeout: 3 * time.Millisecond}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if countForce(f.killCalls, false) == 0 {
		t.Errorf("expected an initial graceful attempt, calls=%v", f.killCalls)
	}
	if countForce(f.killCalls, true) == 0 {
		t.Errorf("expected escalation to force after grace period, calls=%v", f.killCalls)
	}
}

func TestKillAndWaitForceImmediate(t *testing.T) {
	f := &fakeProc{dieAfterKills: 1}
	withFake(t, f)
	if err := killAndWait(1, Options{Force: true, Timeout: 50 * time.Millisecond}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if countForce(f.killCalls, false) != 0 {
		t.Errorf("force mode must not send graceful signals, calls=%v", f.killCalls)
	}
	if countForce(f.killCalls, true) == 0 {
		t.Errorf("expected a forced kill, calls=%v", f.killCalls)
	}
}

func TestKillAndWaitImmortalFails(t *testing.T) {
	f := &fakeProc{immortal: true}
	withFake(t, f)
	if err := killAndWait(1, Options{Timeout: 5 * time.Millisecond}); err == nil {
		t.Fatalf("expected an error for an unkillable process")
	}
}
