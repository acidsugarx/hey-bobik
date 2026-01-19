package obsidian

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Service handles interactions with the Obsidian vault.
type Service struct {
	VaultPath string
	Prefix    string
	Now       func() time.Time
}

// New creates a new Obsidian service.
func New(vaultPath, prefix string) *Service {
	return &Service{
		VaultPath: vaultPath,
		Prefix:    prefix,
		Now:       time.Now,
	}
}

// AppendToDailyNote appends a note to the daily Markdown file.
func (s *Service) AppendToDailyNote(content string) error {
	now := s.Now()
	fileName := fmt.Sprintf("%s%s.md", s.Prefix, now.Format("2006-01-02"))
	filePath := filepath.Join(s.VaultPath, fileName)

	exists := true
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		exists = false
	}

	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open daily note: %w", err)
	}
	defer f.Close()

	if !exists {
		header := fmt.Sprintf("---\ndate: %s\nsource: Bobik\ntags: [voice-note, inbox]\n---\n\n", now.Format(time.RFC3339))
		if _, err := f.WriteString(header); err != nil {
			return fmt.Errorf("failed to write header: %w", err)
		}
	}

	entry := fmt.Sprintf("## %s\n%s\n\n", now.Format("15:04:05"), content)
	if _, err := f.WriteString(entry); err != nil {
		return fmt.Errorf("failed to write entry: %w", err)
	}

	return nil
}
