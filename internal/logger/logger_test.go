package logger

import (
	"bytes"
	"strings"
	"testing"
)

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected Level
	}{
		{"debug", LevelDebug},
		{"DEBUG", LevelDebug},
		{"info", LevelInfo},
		{"INFO", LevelInfo},
		{"warn", LevelWarn},
		{"warning", LevelWarn},
		{"error", LevelError},
		{"ERROR", LevelError},
		{"unknown", LevelInfo}, // default
	}

	for _, tt := range tests {
		result := ParseLevel(tt.input)
		if result != tt.expected {
			t.Errorf("ParseLevel(%q) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}

func TestLoggerLevels(t *testing.T) {
	var buf bytes.Buffer
	SetOutput(&buf)
	SetLevel(LevelWarn)

	l := New("test")
	l.Debug("debug message")
	l.Info("info message")
	l.Warn("warn message")
	l.Error("error message")

	output := buf.String()

	if strings.Contains(output, "DEBUG") {
		t.Error("debug message should not be logged at Warn level")
	}
	if strings.Contains(output, "[INFO]") {
		t.Error("info message should not be logged at Warn level")
	}
	if !strings.Contains(output, "[WARN]") {
		t.Error("warn message should be logged")
	}
	if !strings.Contains(output, "[ERROR]") {
		t.Error("error message should be logged")
	}
	if !strings.Contains(output, "[test]") {
		t.Error("component name should be included")
	}
}

func TestLevelString(t *testing.T) {
	tests := []struct {
		level    Level
		expected string
	}{
		{LevelDebug, "DEBUG"},
		{LevelInfo, "INFO"},
		{LevelWarn, "WARN"},
		{LevelError, "ERROR"},
		{Level(99), "UNKNOWN"},
	}

	for _, tt := range tests {
		if tt.level.String() != tt.expected {
			t.Errorf("Level(%d).String() = %q, want %q", tt.level, tt.level.String(), tt.expected)
		}
	}
}
