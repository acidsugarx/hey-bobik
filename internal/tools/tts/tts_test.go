package tts

import (
	"context"
	"testing"
)

func TestSpeakerDisabled(t *testing.T) {
	s := New(false, "echo")

	err := s.Speak(context.Background(), "test")
	if err != nil {
		t.Errorf("disabled speaker should return nil, got %v", err)
	}
}

func TestSpeakerNoCommand(t *testing.T) {
	s := New(true, "")

	err := s.Speak(context.Background(), "test")
	if err == nil {
		t.Error("expected error for empty command")
	}
}

func TestIsAvailable(t *testing.T) {
	// Test with a command that definitely exists
	s := New(true, "echo")
	if !s.IsAvailable() {
		t.Error("echo should be available")
	}

	// Test with a command that doesn't exist
	s2 := New(true, "nonexistent_command_12345")
	if s2.IsAvailable() {
		t.Error("nonexistent command should not be available")
	}
}

func TestSpeakWithEcho(t *testing.T) {
	// Use echo as a simple test (it won't actually speak but will succeed)
	s := &Speaker{
		Enabled: true,
		Command: "echo",
		Args:    []string{},
	}

	err := s.Speak(context.Background(), "test message")
	if err != nil {
		t.Errorf("speak with echo failed: %v", err)
	}
}

func TestDefaultArgs(t *testing.T) {
	// Test espeak-ng default args
	s := New(true, "espeak-ng")
	if len(s.Args) == 0 || s.Args[0] != "-v" {
		t.Error("espeak-ng should have -v ru args")
	}

	// Test piper default args
	s2 := New(true, "piper")
	if len(s2.Args) == 0 || s2.Args[0] != "--model" {
		t.Error("piper should have --model args")
	}
}
