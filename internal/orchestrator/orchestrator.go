package orchestrator

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strings"
	"text/template"
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
	RewriteLastNote(content string) error
}

// Orchestrator coordinates the audio capture, STT, and tool execution.
type Orchestrator struct {
	Recorder Recorder
	STT      STTEngine
	Notifier Notifier
	LLM      LLMClient
	Obsidian ObsidianService
	Memory   *ContextMemory
}

const systemPrompt = `Ты — Бобик. Твоя задача: вернуть очищенный текст заметки без лишних слов.
НЕ используй жирный шрифт. НЕ пиши слова "Суть", "Текст", "Исправлено". 
Если пользователь просит "исправить" или "изменить" последнюю запись, ОБЯЗАТЕЛЬНО начни ответ с "UPDATE:".
В остальных случаях пиши ТОЛЬКО текст.

Примеры:
Ввод: "запиши купить хлеб"
Ответ: Купить хлеб

Ввод: "исправь на батон"
Ответ: UPDATE: Купить батон

Ввод: "запиши напомни позвонить маме вечером"
Ответ: Позвонить маме вечером

Контекст:
{{.Context}}

Ввод: {{.Input}}
Ответ:`

const (
	wakeWordGrammar = `["эй бобик", "бобик", "запиши", "сделай", "напомни", "поставь", "[unk]"]`
	wakeWord        = "эй бобик"
)

// Start begins the main wake word detection loop.
func (o *Orchestrator) Start(ctx context.Context) error {
	log.Println("Bobik is listening for 'Эй, Бобик'...")

	// Global audio channel to keep the stream drained and avoid ALSA XRUNs
	audioChan := make(chan []int16, 100)

	// Single persistent recorder goroutine
	go func() {
		for {
			select {
			case <-ctx.Done():
				close(audioChan)
				return
			default:
				samples, err := o.Recorder.Read()
				if err != nil {
					continue
				}
				// Copy the buffer to avoid data corruption
			samplesCopy := make([]int16, len(samples))
			copy(samplesCopy, samples)

				// Non-blocking send to avoid blocking the recorder if the consumer is slow
				select {
				case audioChan <- samplesCopy:
				default:
					// Drop samples if buffer is full (overflow)
				}
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// 1. Listen for Wake Word
			detected, err := o.STT.ListenForWakeWord(audioChan, wakeWordGrammar, wakeWord)
			if err != nil {
				log.Printf("wake word error: %v", err)
				continue
			}

			if detected {
				o.handleCommand(ctx, audioChan)
			}
		}
	}
}

func (o *Orchestrator) handleCommand(ctx context.Context, audioChan <-chan []int16) {
	log.Println("Wake word detected!")
	o.Notifier.Notify(ctx, "Bobik", "Listening...")

	// 2. Transcribe Command
	text, err := o.STT.Transcribe(audioChan)
	if err != nil {
		log.Printf("transcription error: %v", err)
		return
	}
	text = strings.TrimSpace(text)
	log.Printf("Transcribed: %s", text)

	if text == "" {
		return
	}

	// 3. Process with LLM
	history := o.Memory.GetHistory()
	var contextStr strings.Builder
	for _, entry := range history {
		contextStr.WriteString(fmt.Sprintf("- Команда: %s, Действие: %s\n", entry.Command, entry.Action))
	}

	tmpl, _ := template.New("prompt").Parse(systemPrompt)
	var promptBuf bytes.Buffer
	tmpl.Execute(&promptBuf, map[string]string{
		"Context": contextStr.String(),
		"Input":   text,
	})

	noteContent, err := o.LLM.Generate(ctx, "", promptBuf.String())
	if err != nil {
		o.Notifier.Notify(ctx, "Bobik Error", "LLM failed")
		return
	}
	log.Printf("LLM Raw output: %s", noteContent)

	// 4. Determine Action and Save
	isUpdate := false
	if strings.HasPrefix(noteContent, "UPDATE:") {
		isUpdate = true
		noteContent = strings.TrimPrefix(noteContent, "UPDATE:")
		noteContent = strings.TrimSpace(noteContent)
	}

	if isUpdate {
		err = o.Obsidian.RewriteLastNote(noteContent)
	} else {
		err = o.Obsidian.AppendToDailyNote(noteContent)
	}

	if err != nil {
		log.Printf("Save error: %v", err)
		o.Notifier.Notify(ctx, "Bobik Error", "Failed to save note")
		return
	}

	// 5. Update Memory
	action := "Saved note"
	if isUpdate {
		action = "Updated last note"
	}
	o.Memory.Add(text, fmt.Sprintf("%s: %s", action, noteContent))

	o.Notifier.Notify(ctx, "Bobik", "Note saved to Daily Notes")

	// Drain any leftover audio from the channel to avoid "ghost" commands
	for len(audioChan) > 0 {
		<-audioChan
	}
}
