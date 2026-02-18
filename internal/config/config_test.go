package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDefault(t *testing.T) {
	cfg := Default()

	if cfg.SampleRate != 16000 {
		t.Errorf("expected SampleRate 16000, got %d", cfg.SampleRate)
	}
	if cfg.OllamaModel != "qwen3:8b" {
		t.Errorf("expected OllamaModel qwen3:8b, got %s", cfg.OllamaModel)
	}
	if cfg.WakeWord != "эй бобик" {
		t.Errorf("expected WakeWord 'эй бобик', got %s", cfg.WakeWord)
	}
}

func TestLoadFromFile(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.json")

	// Write test config
	content := `{
		"sample_rate": 48000,
		"ollama_model": "test-model",
		"vault_path": "/test/vault"
	}`
	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.SampleRate != 48000 {
		t.Errorf("expected SampleRate 48000, got %d", cfg.SampleRate)
	}
	if cfg.OllamaModel != "test-model" {
		t.Errorf("expected OllamaModel test-model, got %s", cfg.OllamaModel)
	}
	if cfg.VaultPath != "/test/vault" {
		t.Errorf("expected VaultPath /test/vault, got %s", cfg.VaultPath)
	}
	// Check that defaults are preserved for unset values
	if cfg.Channels != 1 {
		t.Errorf("expected Channels 1 (default), got %d", cfg.Channels)
	}
}

func TestEnvOverrides(t *testing.T) {
	os.Setenv("BOBIK_OLLAMA_MODEL", "env-model")
	os.Setenv("BOBIK_TTS_ENABLED", "true")
	defer os.Unsetenv("BOBIK_OLLAMA_MODEL")
	defer os.Unsetenv("BOBIK_TTS_ENABLED")

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.OllamaModel != "env-model" {
		t.Errorf("expected OllamaModel env-model, got %s", cfg.OllamaModel)
	}
	if !cfg.TTSEnabled {
		t.Error("expected TTSEnabled true")
	}
}

func TestSave(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "nested", "config.json")

	cfg := Default()
	cfg.OllamaModel = "saved-model"
	cfg.SilenceDelay = 2 * time.Second

	if err := cfg.Save(cfgPath); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Load it back
	loaded, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.OllamaModel != "saved-model" {
		t.Errorf("expected OllamaModel saved-model, got %s", loaded.OllamaModel)
	}
}
