package timer

import (
	"testing"
	"time"
)

func TestTimer(t *testing.T) {
	fired := false
	callback := func(name string) {
		fired = true
	}

	tm := New(callback)

	// Start a very short timer
	tm.Start("test timer", 100*time.Millisecond)

	// Wait for it to fire
	time.Sleep(200 * time.Millisecond)

	if !fired {
		t.Error("expected timer to fire, but it didn't")
	}
}

func TestTimerNew(t *testing.T) {
	tm := New(nil)
	if tm == nil {
		t.Fatal("expected New to return a Timer, got nil")
	}
	// Starting with nil callback shouldn't panic
	tm.Start("nil test", 10*time.Millisecond)
	time.Sleep(20 * time.Millisecond)
}

func TestTimerCancel(t *testing.T) {
	fired := false
	callback := func(name string) {
		fired = true
	}

	tm := New(callback)
	tm.Start("cancel test", 100*time.Millisecond)

	// Cancel before it fires
	cancelled := tm.Cancel("cancel test")
	if !cancelled {
		t.Error("expected Cancel to return true")
	}

	// Wait past the original fire time
	time.Sleep(150 * time.Millisecond)

	if fired {
		t.Error("timer should not have fired after cancel")
	}
}

func TestTimerCancelNonExistent(t *testing.T) {
	tm := New(nil)
	cancelled := tm.Cancel("nonexistent")
	if cancelled {
		t.Error("expected Cancel to return false for nonexistent timer")
	}
}

func TestTimerCancelAll(t *testing.T) {
	tm := New(nil)
	tm.Start("timer1", 1*time.Second)
	tm.Start("timer2", 1*time.Second)
	tm.Start("timer3", 1*time.Second)

	if tm.ActiveCount() != 3 {
		t.Errorf("expected 3 active timers, got %d", tm.ActiveCount())
	}

	count := tm.CancelAll()
	if count != 3 {
		t.Errorf("expected CancelAll to return 3, got %d", count)
	}

	if tm.ActiveCount() != 0 {
		t.Errorf("expected 0 active timers after CancelAll, got %d", tm.ActiveCount())
	}
}

func TestTimerReplaceSameName(t *testing.T) {
	fired := 0
	callback := func(name string) {
		fired++
	}

	tm := New(callback)
	tm.Start("same name", 50*time.Millisecond)
	tm.Start("same name", 100*time.Millisecond) // Should replace the first one

	if tm.ActiveCount() != 1 {
		t.Errorf("expected 1 active timer, got %d", tm.ActiveCount())
	}

	time.Sleep(150 * time.Millisecond)

	if fired != 1 {
		t.Errorf("expected callback to fire once, fired %d times", fired)
	}
}
