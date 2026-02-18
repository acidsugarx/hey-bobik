package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// Config holds all application configuration.
type Config struct {
	// Audio settings
	SampleRate int `json:"sample_rate"`
	Channels   int `json:"channels"`
	BufferSize int `json:"buffer_size"`

	// Vosk STT settings
	ModelPath     string        `json:"model_path"`
	WakeWord      string        `json:"wake_word"`
	WakeGrammar   string        `json:"wake_grammar"`
	SilenceDelay  time.Duration `json:"silence_delay"`
	MaxListenTime time.Duration `json:"max_listen_time"`

	// LLM settings
	OllamaURL     string        `json:"ollama_url"`
	OllamaModel   string        `json:"ollama_model"`
	OllamaTimeout time.Duration `json:"ollama_timeout"`

	// Vision model settings (для анализа скриншотов)
	VisionModel   string `json:"vision_model"`   // e.g., "llava", "llava:13b", "bakllava"
	VisionEnabled bool   `json:"vision_enabled"` // включить возможность анализа экрана

	// Obsidian settings
	VaultPath  string `json:"vault_path"`
	NotePrefix string `json:"note_prefix"`

	// TTS settings
	TTSEnabled bool   `json:"tts_enabled"`
	TTSCommand string `json:"tts_command"` // e.g., "espeak-ng" or "piper"

	// Logging
	LogLevel string `json:"log_level"` // debug, info, warn, error
}

// Default returns the default configuration.
func Default() *Config {
	home, _ := os.UserHomeDir()
	return &Config{
		// Audio
		SampleRate: 16000,
		Channels:   1,
		BufferSize: 4000,

		// STT
		ModelPath:     "models/vosk-model-small-ru-0.22",
		WakeWord:      "эй бобик",
		WakeGrammar:   `["эй бобик", "бобик", "запиши", "сделай", "напомни", "поставь", "[unk]"]`,
		SilenceDelay:  1 * time.Second,
		MaxListenTime: 7 * time.Second,

		// LLM
		OllamaURL:     "http://localhost:11434",
		OllamaModel:   "qwen3:8b",
		OllamaTimeout: 60 * time.Second,

		// Vision
		VisionModel:   "llava",
		VisionEnabled: false,

		// Obsidian
		VaultPath:  filepath.Join(home, "SECOND_BRAIN", "SECOND_BRAIN"),
		NotePrefix: "",

		// TTS
		TTSEnabled: false,
		TTSCommand: "espeak-ng",

		// Logging
		LogLevel: "info",
	}
}

// Load loads configuration from file, applying env overrides.
func Load(path string) (*Config, error) {
	cfg := Default()

	// Try to load from file
	if path != "" {
		if data, err := os.ReadFile(path); err == nil {
			if err := json.Unmarshal(data, cfg); err != nil {
				return nil, err
			}
		}
	} else {
		// Try default locations
		home, _ := os.UserHomeDir()
		defaultPaths := []string{
			filepath.Join(home, ".config", "bobik", "config.json"),
			filepath.Join(home, ".bobik.json"),
			"bobik.json",
		}
		for _, p := range defaultPaths {
			if data, err := os.ReadFile(p); err == nil {
				if err := json.Unmarshal(data, cfg); err != nil {
					return nil, err
				}
				break
			}
		}
	}

	// Apply environment variable overrides
	cfg.applyEnvOverrides()

	return cfg, nil
}

func (c *Config) applyEnvOverrides() {
	if v := os.Getenv("BOBIK_MODEL_PATH"); v != "" {
		c.ModelPath = v
	}
	if v := os.Getenv("BOBIK_OLLAMA_URL"); v != "" {
		c.OllamaURL = v
	}
	if v := os.Getenv("BOBIK_OLLAMA_MODEL"); v != "" {
		c.OllamaModel = v
	}
	if v := os.Getenv("BOBIK_VAULT_PATH"); v != "" {
		c.VaultPath = v
	}
	if v := os.Getenv("BOBIK_NOTE_PREFIX"); v != "" {
		c.NotePrefix = v
	}
	if v := os.Getenv("BOBIK_TTS_ENABLED"); v == "true" || v == "1" {
		c.TTSEnabled = true
	}
	if v := os.Getenv("BOBIK_TTS_COMMAND"); v != "" {
		c.TTSCommand = v
	}
	if v := os.Getenv("BOBIK_LOG_LEVEL"); v != "" {
		c.LogLevel = v
	}
	if v := os.Getenv("BOBIK_WAKE_WORD"); v != "" {
		c.WakeWord = v
	}
	if v := os.Getenv("BOBIK_VISION_MODEL"); v != "" {
		c.VisionModel = v
	}
	if v := os.Getenv("BOBIK_VISION_ENABLED"); v == "true" || v == "1" {
		c.VisionEnabled = true
	}
}

// Save writes the configuration to a file.
func (c *Config) Save(path string) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
