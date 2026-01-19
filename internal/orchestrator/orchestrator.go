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

const systemPrompt = `Ты — Бобик, интеллектуальный голосовой помощник для Linux. 

Твоя задача состоит из двух этапов:
1. **Очистка (Refinement):** Исправь ошибки распознавания речи (STT) в полученном тексте. Учти контекст предыдущих команд, если они есть. Исправь грамматику, падежи и опечатки, сохраняя смысл.
2. **Извлечение (Extraction):** Выдели суть команды для записи в заметку. Удали вводные слова ("запиши", "сделай заметку", "эй бобик"). 

Верни ТОЛЬКО очищенный текст заметки. Не пиши "Вот ваша заметка" или "Исправленный текст:". Только содержание.

Контекст последних действий:
{{.Context}}

Текст от пользователя:
{{.Input}}`

const (

	wakeWordGrammar = `["эй бобик", "бобик", "запиши", "сделай", "напомни", "поставь", "[unk]"]`

	wakeWord        = "эй бобик"

)



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

			}

		}

	}()



	// 1. Listen for Wake Word with prioritized grammar

	detected, err := o.STT.ListenForWakeWord(audioChan, wakeWordGrammar, wakeWord)


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

				history := o.Memory.GetHistory()

				var contextStr strings.Builder

				for _, entry := range history {

					contextStr.WriteString(fmt.Sprintf("- Команда: %s, Действие: %s\n", entry.Command, entry.Action))

				}

		

				tmpl, err := template.New("prompt").Parse(systemPrompt)

				if err != nil {

					return fmt.Errorf("failed to parse prompt template: %w", err)

				}

		

				var promptBuf bytes.Buffer

				err = tmpl.Execute(&promptBuf, map[string]string{

					"Context": contextStr.String(),

					"Input":   text,

				})

				if err != nil {

					return fmt.Errorf("failed to execute prompt template: %w", err)

				}

		

				noteContent, err := o.LLM.Generate(ctx, "", promptBuf.String())

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

		

				// 5. Update Memory

				o.Memory.Add(text, "Saved note: "+noteContent)

		

				o.Notifier.Notify(ctx, "Bobik", "Note saved to Daily Notes")

			}

		

			return nil

		}

		