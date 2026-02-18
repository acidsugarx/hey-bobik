package clipboard

import (
	"testing"
)

func TestNew(t *testing.T) {
	s := New()
	if s == nil {
		t.Fatal("New() returned nil")
	}
}

func TestIsAvailable(t *testing.T) {
	s := New()
	// This test will pass or fail depending on the system
	// Just check it doesn't panic
	_ = s.IsAvailable()
}

func TestReadWriteIntegration(t *testing.T) {
	s := New()
	if !s.IsAvailable() {
		t.Skip("clipboard tools not available")
	}

	// Write test content
	testContent := "Bobik clipboard test 12345"
	err := s.Write(testContent)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// Read it back
	content, err := s.Read()
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	if content != testContent {
		t.Errorf("expected %q, got %q", testContent, content)
	}
}

func TestReadNoTool(t *testing.T) {
	s := &Service{readCmd: "", writeCmd: ""}

	_, err := s.Read()
	if err == nil {
		t.Error("expected error when no tool available")
	}
}

func TestWriteNoTool(t *testing.T) {
	s := &Service{readCmd: "", writeCmd: ""}

	err := s.Write("test")
	if err == nil {
		t.Error("expected error when no tool available")
	}
}
