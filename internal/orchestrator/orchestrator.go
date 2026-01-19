package orchestrator

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"text/template"
	"time"
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

// TimerService defines the interface for setting timers.
type TimerService interface {
	Start(name string, duration time.Duration)
}

// ClockService defines the interface for time reporting.
type ClockService interface {
	GetCurrentTime() string
}

// State represents the current internal state of the orchestrator.
type State int

const (
	StateIdle State = iota
	StateListening
	StateThinking
)

// Orchestrator coordinates the audio capture, STT, and tool execution.
type Orchestrator struct {
	Recorder      Recorder
	STT           STTEngine
	Notifier      Notifier
	LLM           LLMClient
	Obsidian      ObsidianService
	Timer         TimerService
	Clock         ClockService
	Memory        *ContextMemory
	OnStateChange func(State)
}

const systemPrompt = `Ты — Бобик, интеллектуальный помощник для Linux. 
Твоя задача: проанализировать ввод пользователя и выбрать одно действие.

Доступные действия:
1. NOTE: Записать или обновить заметку в Obsidian.
2. TIMER: Поставить таймер (нужно указать длительность в секундах).
3. TIME: Сообщить текущее время.

Формат ответа: ACTION: [ACTION_NAME] | ARG: [VALUE]

Правила:
- Если просят "записать" или "заметка" -> ACTION: NOTE | ARG: [Текст заметки]
- Если просят "исправить" или "изменить" последнюю запись -> ACTION: NOTE | ARG: UPDATE: [Новый текст]
- Если просят "таймер" или "напомни через" -> ACTION: TIMER | ARG: [Кол-во секунд]
- Если спрашивают "сколько времени" или "час" -> ACTION: TIME | ARG: none

Примеры:
Ввод: "запиши купить хлеб"
Ответ: ACTION: NOTE | ARG: Купить хлеб

Ввод: "исправь на батон"
Ответ: ACTION: NOTE | ARG: UPDATE: Купить батон

Ввод: "поставь таймер на 5 минут"
Ответ: ACTION: TIMER | ARG: 300

Ввод: "сколько времени"
Ответ: ACTION: TIME | ARG: none

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

	if o.OnStateChange != nil {
		o.OnStateChange(StateIdle)
	}

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
				if o.OnStateChange != nil {
					o.OnStateChange(StateIdle)
				}
			}
		}
	}
}

func (o *Orchestrator) handleCommand(ctx context.Context, audioChan <-chan []int16) {
	log.Println("Wake word detected!")
	if o.OnStateChange != nil {
		o.OnStateChange(StateListening)
	}
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

	if o.OnStateChange != nil {
		o.OnStateChange(StateThinking)
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

	rawOutput, err := o.LLM.Generate(ctx, "", promptBuf.String())
	if err != nil {
		o.Notifier.Notify(ctx, "Bobik Error", "LLM failed")
		return
	}
	log.Printf("LLM Raw output: %s", rawOutput)

	// 4. Parse Action and Argument
	action, arg := o.parseLLMOutput(rawOutput)
	log.Printf("Parsed Action: %s, Arg: %s", action, arg)

	// 5. Dispatch Tool
	switch action {
	case "NOTE":
		o.handleNoteAction(ctx, text, arg)
	case "TIMER":
		o.handleTimerAction(ctx, text, arg)
	case "TIME":
		o.handleTimeAction(ctx, text)
	default:
		log.Printf("Unknown action: %s", action)
		o.Notifier.Notify(ctx, "Bobik", "Не понял команду")
	}

	// Drain any leftover audio from the channel to avoid "ghost" commands
	for len(audioChan) > 0 {
		<-audioChan
	}
}

func (o *Orchestrator) parseLLMOutput(output string) (string, string) {
	// Format: ACTION: [ACTION_NAME] | ARG: [VALUE]
	parts := strings.Split(output, "|")
	action := ""
	arg := ""

	for _, part := range parts {
		subParts := strings.SplitN(strings.TrimSpace(part), ":", 2)
		if len(subParts) < 2 {
			continue
		}
		key := strings.TrimSpace(subParts[0])
		val := strings.TrimSpace(subParts[1])

		if key == "ACTION" {
			action = val
		} else if key == "ARG" {
			arg = val
		}
	}
	return action, arg
}

func (o *Orchestrator) handleNoteAction(ctx context.Context, rawInput, arg string) {
	isUpdate := false
	noteContent := arg
	if strings.HasPrefix(arg, "UPDATE:") {
		isUpdate = true
		noteContent = strings.TrimPrefix(arg, "UPDATE:")
		noteContent = strings.TrimSpace(noteContent)
	}

	var err error
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

	actionDesc := "Saved note"
	if isUpdate {
		actionDesc = "Updated last note"
	}
	o.Memory.Add(rawInput, fmt.Sprintf("%s: %s", actionDesc, noteContent))
	o.Notifier.Notify(ctx, "Bobik", "Заметка сохранена")
}

func (o *Orchestrator) handleTimerAction(ctx context.Context, rawInput, arg string) {
	seconds, err := strconv.Atoi(arg)
	if err != nil {
		log.Printf("Invalid timer arg: %s", arg)
		o.Notifier.Notify(ctx, "Bobik Error", "Ошибка времени")
		return
	}

	duration := time.Duration(seconds) * time.Second
	o.Timer.Start("Голосовой таймер", duration)
	
o.Memory.Add(rawInput, fmt.Sprintf("Set timer for %d seconds", seconds))
	o.Notifier.Notify(ctx, "Bobik", fmt.Sprintf("Таймер запущен на %d сек", seconds))
}

func (o *Orchestrator) handleTimeAction(ctx context.Context, rawInput string) {
	currentTime := o.Clock.GetCurrentTime()
	o.Notifier.Notify(ctx, "Bobik Time", currentTime)
	o.Memory.Add(rawInput, "Reported current time")
}