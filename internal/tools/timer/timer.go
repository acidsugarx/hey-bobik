package timer

import (
	"time"
)

// Timer handles background countdowns.
type Timer struct {
	Callback func(name string)
}

// New creates a new Timer.
func New(callback func(name string)) *Timer {
	return &Timer{
		Callback: callback,
	}
}

// Start begins a countdown in the background.
func (t *Timer) Start(name string, duration time.Duration) {
	time.AfterFunc(duration, func() {
		if t.Callback != nil {
			t.Callback(name)
		}
	})
}
