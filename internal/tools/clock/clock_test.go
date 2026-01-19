package clock

import (
	"testing"
	"time"
)

func TestGetCurrentTime(t *testing.T) {
	mockTime := time.Date(2026, 1, 19, 21, 30, 0, 0, time.Local)
	c := &Clock{
		Now: func() time.Time { return mockTime },
	}

	result := c.GetCurrentTime()
	expected := "21:30"
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestClockNew(t *testing.T) {
	c := New()
	if c == nil {
		t.Fatal("expected New to return a Clock, got nil")
	}
	if c.Now == nil {
		t.Error("expected Now function to be initialized")
	}
}
