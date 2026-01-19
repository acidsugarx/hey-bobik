package orchestrator

import (
	"testing"
)

func TestContextMemory(t *testing.T) {
	m := NewContextMemory(3)

	m.Add("command 1", "action 1")
	m.Add("command 2", "action 2")
	m.Add("command 3", "action 3")
	
	history := m.GetHistory()
	if len(history) != 3 {
		t.Errorf("expected history length 3, got %d", len(history))
	}

	if history[0].Command != "command 1" {
		t.Errorf("expected first command to be 'command 1', got %s", history[0].Command)
	}

	// Test overflow
	m.Add("command 4", "action 4")
	history = m.GetHistory()
	if len(history) != 3 {
		t.Errorf("expected history length 3 after overflow, got %d", len(history))
	}
	if history[0].Command != "command 2" {
		t.Errorf("expected first command to be 'command 2' after overflow, got %s", history[0].Command)
	}
	if history[2].Command != "command 4" {
		t.Errorf("expected last command to be 'command 4', got %s", history[2].Command)
	}
}

func TestContextMemoryEmpty(t *testing.T) {
	m := NewContextMemory(5)
	history := m.GetHistory()
	if len(history) != 0 {
		t.Errorf("expected empty history, got %d items", len(history))
	}
}
