package main

import (
	"context"
	"flag"
	"hey-bobik/internal/audio"
	"hey-bobik/internal/config"
	"hey-bobik/internal/llm"
	"hey-bobik/internal/logger"
	"hey-bobik/internal/orchestrator"
	"hey-bobik/internal/stt"
	"hey-bobik/internal/tools/calc"
	"hey-bobik/internal/tools/clipboard"
	"hey-bobik/internal/tools/clock"
	"hey-bobik/internal/tools/notifier"
	"hey-bobik/internal/tools/obsidian"
	"hey-bobik/internal/tools/screen"
	"hey-bobik/internal/tools/timer"
	"hey-bobik/internal/tools/tts"
	"hey-bobik/internal/ui/tray"
	"os"
	"os/signal"
	"syscall"
)

var log = logger.New("main")

func main() {
	configPath := flag.String("config", "", "Path to config file (optional)")

	// Legacy flags for backwards compatibility (override config)
	modelPath := flag.String("model", "", "Path to Vosk model directory")
	vaultPath := flag.String("vault", "", "Path to Obsidian vault")
	prefix := flag.String("prefix", "", "Prefix for daily note filenames")
	ollamaURL := flag.String("ollama", "", "Ollama API URL")
	ollamaModel := flag.String("llm", "", "Ollama model name")

	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Error("Failed to load config: %v", err)
		os.Exit(1)
	}

	// Set log level from config
	logger.SetLevel(logger.ParseLevel(cfg.LogLevel))

	// Apply flag overrides
	if *modelPath != "" {
		cfg.ModelPath = *modelPath
	}
	if *vaultPath != "" {
		cfg.VaultPath = *vaultPath
	}
	if *prefix != "" {
		cfg.NotePrefix = *prefix
	}
	if *ollamaURL != "" {
		cfg.OllamaURL = *ollamaURL
	}
	if *ollamaModel != "" {
		cfg.OllamaModel = *ollamaModel
	}

	log.Info("Starting Bobik with model: %s, LLM: %s", cfg.ModelPath, cfg.OllamaModel)

	// 1. Initialize Tools
	n := notifier.New()
	oService := obsidian.New(cfg.VaultPath, cfg.NotePrefix)
	lClient := llm.New(cfg.OllamaURL, cfg.OllamaModel)

	cService := clock.New()
	tService := timer.New(func(name string) {
		n.Notify(context.Background(), "Бобик", "Время вышло: "+name)
	})

	// Initialize TTS
	ttsService := tts.New(cfg.TTSEnabled, cfg.TTSCommand)
	if cfg.TTSEnabled && !ttsService.IsAvailable() {
		log.Warn("TTS enabled but command '%s' not found", cfg.TTSCommand)
	}

	// Initialize Clipboard
	clipboardService := clipboard.New()
	if !clipboardService.IsAvailable() {
		log.Warn("Clipboard tools not found (install xclip or xsel)")
	}

	// Initialize Calculator
	calcService := calc.New()

	// Initialize Screen capture and Vision LLM (if enabled)
	var screenService *screen.Adapter
	var visionClient *llm.Client

	if cfg.VisionEnabled {
		// Инициализируем инструмент скриншотов
		screenTool := screen.New()
		if screenTool.IsAvailable() {
			screenService = screen.NewAdapter(screenTool)
			log.Info("Screen capture available using: %s", screenTool.GetAvailableBackend())

			// Инициализируем отдельный LLM клиент для vision модели
			visionClient = llm.New(cfg.OllamaURL, cfg.VisionModel)
			log.Info("Vision model configured: %s", cfg.VisionModel)
		} else {
			log.Warn("Vision enabled but no screenshot tool found (install gnome-screenshot, scrot, or grim)")
		}
	}

	// 2. Initialize STT Engine
	engine, err := stt.NewEngine(cfg.ModelPath)
	if err != nil {
		log.Error("Failed to initialize STT engine: %v", err)
		os.Exit(1)
	}
	defer engine.Close()

	// 3. Initialize Audio Recorder
	recorder := audio.NewRecorder(cfg.SampleRate, cfg.Channels, cfg.BufferSize)
	err = recorder.Start()
	if err != nil {
		log.Error("Failed to start audio recorder: %v", err)
		os.Exit(1)
	}
	defer recorder.Stop()

	// Handle Graceful Shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 5. Initialize Tray UI
	trayManager := tray.New(func() {
		log.Info("Tray exited, shutting down...")
		cancel()
	})

	// 4. Initialize Orchestrator
	o := &orchestrator.Orchestrator{
		Recorder:  recorder,
		STT:       engine,
		Notifier:  n,
		LLM:       lClient,
		VisionLLM: visionClient,
		Obsidian:  oService,
		Timer:     tService,
		Clock:     cService,
		TTS:       ttsService,
		Clipboard: clipboardService,
		Calc:      calcService,
		Screen:    screenService,
		Memory:    orchestrator.NewContextMemory(10),
		OnStateChange: func(s orchestrator.State) {
			switch s {
			case orchestrator.StateIdle:
				trayManager.SetState(tray.StateIdle)
			case orchestrator.StateListening:
				trayManager.SetState(tray.StateListening)
			case orchestrator.StateThinking:
				trayManager.SetState(tray.StateThinking)
			}
		},
	}

	// Start Orchestrator in a goroutine
	go func() {
		if err := o.Start(ctx); err != nil && err != context.Canceled {
			log.Error("Orchestrator stopped with error: %v", err)
		}
		os.Exit(0)
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Info("Shutting down Bobik...")
		cancel()
	}()

	// Run Tray on the main thread
	trayManager.Run()
}
