package orchestrator

import (
	"sync"
)

// ContextEntry represents a single interaction in the history.
type ContextEntry struct {
	Command string
	Action  string
}

// ContextMemory stores a rolling history of interactions.
type ContextMemory struct {
	mu      sync.RWMutex
	entries []ContextEntry
	maxSize int
}

// NewContextMemory creates a new ContextMemory with the specified size.
func NewContextMemory(maxSize int) *ContextMemory {
	return &ContextMemory{
		entries: make([]ContextEntry, 0, maxSize),
		maxSize: maxSize,
	}
}

// Add appends a new entry to the history, removing the oldest if necessary.
func (m *ContextMemory) Add(command, action string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.entries) >= m.maxSize {
		m.entries = append(m.entries[1:], ContextEntry{Command: command, Action: action})
	} else {
		m.entries = append(m.entries, ContextEntry{Command: command, Action: action})
	}
}

// GetHistory returns a copy of the current history.
func (m *ContextMemory) GetHistory() []ContextEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()

	history := make([]ContextEntry, len(m.entries))
	copy(history, m.entries)
	return history
}
