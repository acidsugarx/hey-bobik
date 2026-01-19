package tray

import (
	"testing"
)

func TestState(t *testing.T) {
	if StateIdle != 0 {
		t.Errorf("expected StateIdle to be 0, got %d", StateIdle)
	}
	if StateListening != 1 {
		t.Errorf("expected StateListening to be 1, got %d", StateListening)
	}
	if StateThinking != 2 {
		t.Errorf("expected StateThinking to be 2, got %d", StateThinking)
	}
}
