package clipboard

import (
	"fmt"
	"os/exec"
	"strings"
)

// Service handles clipboard operations using xclip/xsel.
type Service struct {
	readCmd  string
	writeCmd string
}

// New creates a new Clipboard service.
func New() *Service {
	// Try to find available clipboard tool
	readCmd := ""
	writeCmd := ""

	if _, err := exec.LookPath("xclip"); err == nil {
		readCmd = "xclip"
		writeCmd = "xclip"
	} else if _, err := exec.LookPath("xsel"); err == nil {
		readCmd = "xsel"
		writeCmd = "xsel"
	} else if _, err := exec.LookPath("wl-paste"); err == nil {
		// Wayland
		readCmd = "wl-paste"
		writeCmd = "wl-copy"
	}

	return &Service{
		readCmd:  readCmd,
		writeCmd: writeCmd,
	}
}

// IsAvailable checks if clipboard tools are installed.
func (s *Service) IsAvailable() bool {
	return s.readCmd != ""
}

// Read returns the current clipboard content.
func (s *Service) Read() (string, error) {
	if s.readCmd == "" {
		return "", fmt.Errorf("no clipboard tool available (install xclip or xsel)")
	}

	var cmd *exec.Cmd
	switch s.readCmd {
	case "xclip":
		cmd = exec.Command("xclip", "-selection", "clipboard", "-o")
	case "xsel":
		cmd = exec.Command("xsel", "--clipboard", "--output")
	case "wl-paste":
		cmd = exec.Command("wl-paste")
	default:
		return "", fmt.Errorf("unknown clipboard command: %s", s.readCmd)
	}

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to read clipboard: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// Write sets the clipboard content.
func (s *Service) Write(content string) error {
	if s.writeCmd == "" {
		return fmt.Errorf("no clipboard tool available (install xclip or xsel)")
	}

	var cmd *exec.Cmd
	switch s.writeCmd {
	case "xclip":
		cmd = exec.Command("xclip", "-selection", "clipboard")
	case "xsel":
		cmd = exec.Command("xsel", "--clipboard", "--input")
	case "wl-copy":
		cmd = exec.Command("wl-copy")
	default:
		return fmt.Errorf("unknown clipboard command: %s", s.writeCmd)
	}

	cmd.Stdin = strings.NewReader(content)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to write to clipboard: %w", err)
	}

	return nil
}
