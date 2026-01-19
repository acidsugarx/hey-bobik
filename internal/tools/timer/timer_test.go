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

	tm := &Timer{
		Callback: callback,
	}

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
	time.Sleep(20*time.Millisecond)
}
