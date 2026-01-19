package obsidian

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestAppendToDailyNote(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "obsidian_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	service := &Service{
		VaultPath: tempDir,
		Now: func() time.Time {
			return time.Date(2026, 1, 19, 12, 0, 0, 0, time.UTC)
		},
	}

	noteContent := "Test note content"
	err = service.AppendToDailyNote(noteContent)
	if err != nil {
		t.Fatalf("AppendToDailyNote failed: %v", err)
	}

	expectedFileName := "2026-01-19.md"
	filePath := filepath.Join(tempDir, expectedFileName)
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read created note: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "---") {
		t.Error("expected YAML frontmatter, not found")
	}
	if !strings.Contains(contentStr, "source: Bobik") {
		t.Error("expected source: Bobik in frontmatter")
	}
	if !strings.Contains(contentStr, noteContent) {
		t.Errorf("expected %q in content, not found", noteContent)
	}

	// Test appending to existing file
	secondNote := "Second note content"
	err = service.AppendToDailyNote(secondNote)
	if err != nil {
		t.Fatalf("Second AppendToDailyNote failed: %v", err)
	}

	content, err = os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read note after second append: %v", err)
	}
	contentStr = string(content)
	if !strings.Contains(contentStr, noteContent) {
		t.Error("first note content lost")
	}
	if !strings.Contains(contentStr, secondNote) {
		t.Error("second note content not found")
	}
}

func TestAppendToDailyNoteError(t *testing.T) {
	service := &Service{
		VaultPath: "/non-existent-path-abc-123",
		Now:       time.Now,
	}

	err := service.AppendToDailyNote("test")
	if err == nil {
		t.Error("expected error for non-existent path, got nil")
	}
}
