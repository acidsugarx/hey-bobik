package clock

import (
	"time"
)

// Clock handles time reporting.
type Clock struct {
	Now func() time.Time
}

// New creates a new Clock.
func New() *Clock {
	return &Clock{
		Now: time.Now,
	}
}

// GetCurrentTime returns the current time formatted as HH:MM.
func (c *Clock) GetCurrentTime() string {
	return c.Now().Format("15:04")
}
