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

// Orchestrator coordinates the audio capture, STT, and tool execution.
type Orchestrator struct {
	Recorder Recorder
	STT      STTEngine
	Notifier Notifier
}

// Start begins the main wake word detection loop.
func (o *Orchestrator) Start(ctx context.Context) error {
	log.Println("Bobik is listening for 'Эй, Бобик'...")

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Continuous loop for wake word
			audioChan := make(chan []int16, 10)
			
			// Start audio capture in a goroutine
			go func() {
				defer close(audioChan)
				for {
					select {
					case <-ctx.Done():
						return
					default:
						samples, err := o.Recorder.Read()
						if err != nil {
							log.Printf("audio read error: %v", err)
							return
						}
						audioChan <- samples
					}
				}
			}()

			// Listen for wake word
			detected, err := o.STT.ListenForWakeWord(audioChan, `["эй бобик", "[unk]"]`, "эй бобик")
			if err != nil {
				log.Printf("wake word detection error: %v", err)
				continue
			}

			if detected {
				log.Println("Wake word detected!")
				o.Notifier.Notify(ctx, "Bobik", "Listening...")
				
				// Capture the command
				// In a real implementation, we'd need a fresh audio channel or reset the stream
				// For now, we'll just log and continue the loop.
				// Phase 3 will implement the command capture and LLM processing.
			}
		}
	}
}
