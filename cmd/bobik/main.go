package main

import (
	"context"
	"flag"
	"hey-bobik/internal/audio"
	"hey-bobik/internal/llm"
	"hey-bobik/internal/orchestrator"
	"hey-bobik/internal/stt"
	"hey-bobik/internal/tools/clock"
	"hey-bobik/internal/tools/notifier"
	"hey-bobik/internal/tools/obsidian"
	"hey-bobik/internal/tools/timer"
	"hey-bobik/internal/ui/tray"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

func main() {
	home, _ := os.UserHomeDir()
	
	modelPath := flag.String("model", "models/vosk-model-small-ru-0.22", "Path to Vosk model directory")
	vaultPath := flag.String("vault", filepath.Join(home, "SECOND_BRAIN", "SECOND_BRAIN"), "Path to Obsidian vault")
	prefix := flag.String("prefix", "", "Prefix for daily note filenames")
	ollamaURL := flag.String("ollama", "http://localhost:11434", "Ollama API URL")
	ollamaModel := flag.String("llm", "qwen3:8b", "Ollama model name")
	
	flag.Parse()

	log.Println("Starting Bobik...")

	// 1. Initialize Tools
	n := notifier.New()
	oService := obsidian.New(*vaultPath, *prefix)
	lClient := llm.New(*ollamaURL, *ollamaModel)
	
	cService := clock.New()
	tService := timer.New(func(name string) {
		n.Notify(context.Background(), "Бобик", "Время вышло: "+name)
	})

	// 2. Initialize STT Engine
	engine, err := stt.NewEngine(*modelPath)
	if err != nil {
		log.Fatalf("Failed to initialize STT engine: %v", err)
	}
	defer engine.Close()

	// 3. Initialize Audio Recorder
	recorder := audio.NewRecorder(16000, 1, 4000)
	err = recorder.Start()
	if err != nil {
		log.Fatalf("Failed to start audio recorder: %v", err)
	}
	defer recorder.Stop()

	// Handle Graceful Shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 4. Initialize Orchestrator
	o := &orchestrator.Orchestrator{
		Recorder: recorder,
		STT:      engine,
		Notifier: n,
		LLM:      lClient,
		Obsidian: oService,
		Timer:    tService,
		Clock:    cService,
		Memory:   orchestrator.NewContextMemory(10),
	}

	// 5. Initialize Tray UI
	trayManager := tray.New(func() {
		log.Println("Tray exited, shutting down...")
		cancel()
	})

	// Start Orchestrator in a goroutine
	go func() {
		if err := o.Start(ctx); err != nil && err != context.Canceled {
			log.Printf("Orchestrator stopped with error: %v", err)
		}
		os.Exit(0)
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down Bobik...")
		cancel()
	}()

	// Run Tray on the main thread
	trayManager.Run()
}
