package timer

import (
	"sync"
	"time"
)

// Timer handles background countdowns.
type Timer struct {
	mu       sync.Mutex
	Callback func(name string)
	active   map[string]*time.Timer
}

// New creates a new Timer.
func New(callback func(name string)) *Timer {
	return &Timer{
		Callback: callback,
		active:   make(map[string]*time.Timer),
	}
}

// Start begins a countdown in the background.
func (t *Timer) Start(name string, duration time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Cancel existing timer with same name if exists
	if existing, ok := t.active[name]; ok {
		existing.Stop()
		delete(t.active, name)
	}

	timer := time.AfterFunc(duration, func() {
		t.mu.Lock()
		delete(t.active, name)
		t.mu.Unlock()

		if t.Callback != nil {
			t.Callback(name)
		}
	})

	t.active[name] = timer
}

// Cancel stops a timer by name. Returns true if timer was found and cancelled.
func (t *Timer) Cancel(name string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	if timer, ok := t.active[name]; ok {
		timer.Stop()
		delete(t.active, name)
		return true
	}
	return false
}

// CancelAll stops all active timers. Returns number of cancelled timers.
func (t *Timer) CancelAll() int {
	t.mu.Lock()
	defer t.mu.Unlock()

	count := len(t.active)
	for name, timer := range t.active {
		timer.Stop()
		delete(t.active, name)
	}
	return count
}

// ActiveCount returns the number of active timers.
func (t *Timer) ActiveCount() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return len(t.active)
}
