package tts

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// Speaker handles text-to-speech output.
type Speaker struct {
	Enabled bool
	Command string // e.g., "espeak-ng", "piper", "festival"
	Args    []string
}

// New creates a new TTS speaker.
func New(enabled bool, command string) *Speaker {
	args := []string{}

	// Default arguments for common TTS engines
	switch {
	case strings.Contains(command, "espeak"):
		args = []string{"-v", "ru"} // Russian voice
	case strings.Contains(command, "piper"):
		args = []string{"--model", "ru_RU-dmitri-medium"}
	}

	return &Speaker{
		Enabled: enabled,
		Command: command,
		Args:    args,
	}
}

// Speak synthesizes and plays the given text.
func (s *Speaker) Speak(ctx context.Context, text string) error {
	if !s.Enabled {
		return nil
	}

	if s.Command == "" {
		return fmt.Errorf("TTS command not configured")
	}

	args := append(s.Args, text)
	cmd := exec.CommandContext(ctx, s.Command, args...)

	return cmd.Run()
}

// SpeakAsync synthesizes and plays text in the background.
func (s *Speaker) SpeakAsync(ctx context.Context, text string) {
	if !s.Enabled {
		return
	}
	go s.Speak(ctx, text)
}

// IsAvailable checks if the TTS engine is installed.
func (s *Speaker) IsAvailable() bool {
	if s.Command == "" {
		return false
	}
	_, err := exec.LookPath(s.Command)
	return err == nil
}
