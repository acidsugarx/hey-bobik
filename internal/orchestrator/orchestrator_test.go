package orchestrator

import (
	"testing"
)

func TestOrchestrator(t *testing.T) {
	// We'll implement this properly once we have the interface-based design
	// for audio and stt to allow mocking.
	o := &Orchestrator{}
	if o == nil {
		t.Fatal("orchestrator is nil")
	}
}
