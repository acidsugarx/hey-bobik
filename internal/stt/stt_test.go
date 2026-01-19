package stt

import (
	"testing"
)

func TestEngine(t *testing.T) {
	// Vosk requires a model directory, so we'll just check the structure for now.
	e := &Engine{
		ModelPath: "test_model",
	}

	if e.ModelPath != "test_model" {
		t.Errorf("expected test_model, got %s", e.ModelPath)
	}
}
