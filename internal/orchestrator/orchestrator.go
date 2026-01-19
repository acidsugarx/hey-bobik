package orchestrator

import (
	"context"
	"log"
)

// Recorder defines the interface for audio capture.
type Recorder interface {
	Read() ([]int16, error)
}

// STTEngine defines the interface for speech-to-text.
type STTEngine interface {
	ListenForWakeWord(audioChan <-chan []int16, grammar string, wakeWord string) (bool, error)
	Transcribe(audioChan <-chan []int16) (string, error)
}

// Notifier defines the interface for system notifications.
type Notifier interface {
	Notify(ctx context.Context, title, message string) error
}

// LLMClient defines the interface for LLM inference.
type LLMClient interface {
	Generate(ctx context.Context, system, prompt string) (string, error)
}

// ObsidianService defines the interface for note-taking.
type ObsidianService interface {
	AppendToDailyNote(content string) error
}

// Orchestrator coordinates the audio capture, STT, and tool execution.
type Orchestrator struct {
	Recorder Recorder
	STT      STTEngine
	Notifier Notifier
	LLM      LLMClient
	Obsidian ObsidianService
}

const systemPrompt = `Ты помощник Linux по имени Бобик. 
Твоя задача — извлечь содержание заметки из текста, который продиктовал пользователь.
Верни только текст заметки, без лишних слов, кавычек или пояснений.
Если пользователь просит сделать заметку, убери вводные слова вроде "запиши", "сделай заметку".`

// Start begins the main wake word detection loop.
func (o *Orchestrator) Start(ctx context.Context) error {
	log.Println("Bobik is listening for 'Эй, Бобик'...")

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := o.runOnce(ctx); err != nil {
				log.Printf("error in orchestrator loop: %v", err)
			}
		}
	}
}

func (o *Orchestrator) runOnce(ctx context.Context) error {
	audioChan := make(chan []int16, 100)
	
	// Create a sub-context that we can cancel to stop the recorder goroutine
	recCtx, cancelRec := context.WithCancel(ctx)
	defer cancelRec()

	go func() {
		defer close(audioChan)
		for {
			select {
			case <-recCtx.Done():
				return
								default:
									samples, err := o.Recorder.Read()
									if err != nil {
										return
									}
									// Copy the buffer to avoid data corruption by the next Read()
									samplesCopy := make([]int16, len(samples))
									copy(samplesCopy, samples)
									audioChan <- samplesCopy
								}		}
	}()

	// 1. Listen for Wake Word
	detected, err := o.STT.ListenForWakeWord(audioChan, `["эй бобик", "[unk]"]`, "эй бобик")
	if err != nil {
		return err
	}

	if detected {
		log.Println("Wake word detected!")
		o.Notifier.Notify(ctx, "Bobik", "Listening...")

		// 2. Transcribe Command
		// We stop the wake word detection and start transcription on the remaining/incoming audio
		// In a simplified MVP, we use the same audioChan but Vosk recognizers will be swapped
		text, err := o.STT.Transcribe(audioChan)
		if err != nil {
			return err
		}
		log.Printf("Transcribed: %s", text)

		if text == "" {
			return nil
		}

		// 3. Process with LLM
		noteContent, err := o.LLM.Generate(ctx, systemPrompt, text)
		if err != nil {
			o.Notifier.Notify(ctx, "Bobik Error", "LLM failed")
			return err
		}
		log.Printf("Note content: %s", noteContent)

		// 4. Save to Obsidian
		err = o.Obsidian.AppendToDailyNote(noteContent)
		if err != nil {
			o.Notifier.Notify(ctx, "Bobik Error", "Failed to save note")
			return err
		}

		o.Notifier.Notify(ctx, "Bobik", "Note saved to Daily Notes")
	}

	return nil
}