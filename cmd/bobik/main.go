package main

import (
	"context"
	"flag"
	"hey-bobik/internal/audio"
	"hey-bobik/internal/llm"
	"hey-bobik/internal/orchestrator"
	"hey-bobik/internal/stt"
	"hey-bobik/internal/tools/notifier"
	"hey-bobik/internal/tools/obsidian"
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

	// 4. Initialize Orchestrator
	o := &orchestrator.Orchestrator{
		Recorder: recorder,
		STT:      engine,
		Notifier: n,
		LLM:      lClient,
		Obsidian: oService,
	}

	// Handle Graceful Shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down Bobik...")
		cancel()
	}()

	// Start Orchestrator
	if err := o.Start(ctx); err != nil && err != context.Canceled {
		log.Fatalf("Orchestrator stopped with error: %v", err)
	}
}
