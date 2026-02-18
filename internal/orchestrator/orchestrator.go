package orchestrator

import (
	"bytes"
	"context"
	"fmt"
	"hey-bobik/internal/logger"
	"strconv"
	"strings"
	"text/template"
	"time"
)

var log = logger.New("orchestrator")

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
	DeleteLastNote() error
}

// TimerService defines the interface for setting timers.
type TimerService interface {
	Start(name string, duration time.Duration)
	CancelAll() int
}

// ClockService defines the interface for time reporting.
type ClockService interface {
	GetCurrentTime() string
}

// TTSService defines the interface for text-to-speech.
type TTSService interface {
	SpeakAsync(ctx context.Context, text string)
}

// ClipboardService defines the interface for clipboard operations.
type ClipboardService interface {
	Read() (string, error)
	Write(content string) error
}

// CalcService defines the interface for calculations.
type CalcService interface {
	Eval(expr string) (float64, error)
	Percentage(percent, value float64) float64
	FormatResult(val float64) string
}

// ScreenService defines the interface for screen capture and analysis.
type ScreenService interface {
	Capture() (base64Image string, filePath string, err error)
	CaptureWindow() (base64Image string, filePath string, err error)
	Cleanup(filePath string) error
}

// VisionLLMClient defines the interface for vision-capable LLM.
type VisionLLMClient interface {
	GenerateWithImages(ctx context.Context, system, prompt string, images []string) (string, error)
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
	VisionLLM     VisionLLMClient // Отдельный клиент для vision модели (может быть nil)
	Obsidian      ObsidianService
	Timer         TimerService
	Clock         ClockService
	TTS           TTSService
	Clipboard     ClipboardService
	Calc          CalcService
	Screen        ScreenService // Инструмент для скриншотов
	Memory        *ContextMemory
	OnStateChange func(State)
}

const systemPrompt = `Ты — Бобик, интеллектуальный помощник для Linux. 
Твоя задача: проанализировать ввод пользователя и выбрать одно действие.

Доступные действия:
1. NOTE: Записать или обновить заметку в Obsidian.
2. TIMER: Поставить таймер (нужно указать длительность в секундах).
3. TIME: Сообщить текущее время.
4. CANCEL: Отменить последнее действие (удалить заметку или остановить таймер).
5. CLIPBOARD: Работа с буфером обмена (read - прочитать, write - записать, note - записать буфер в заметку).
6. CALC: Вычислить математическое выражение.
7. SCREEN: Анализ экрана/скриншота (describe - описать что на экране, read - прочитать текст, window - анализ активного окна).

Формат ответа: ACTION: [ACTION_NAME] | ARG: [VALUE]

Правила:
- Если просят "записать" или "заметка" -> ACTION: NOTE | ARG: [Текст заметки]
- Если просят "исправить" или "изменить" последнюю запись -> ACTION: NOTE | ARG: UPDATE: [Новый текст]
- Если просят "таймер" или "напомни через" -> ACTION: TIMER | ARG: [Кол-во секунд]
- Если спрашивают "сколько времени" или "час" -> ACTION: TIME | ARG: none
- Если просят "отменить", "удалить", "отмена" -> ACTION: CANCEL | ARG: [note/timer/all]
- Если просят "скопировать" текст -> ACTION: CLIPBOARD | ARG: write:[текст]
- Если просят "что в буфере" или "прочитай буфер" -> ACTION: CLIPBOARD | ARG: read
- Если просят "вставь из буфера в заметку" -> ACTION: CLIPBOARD | ARG: note
- Если просят "посчитать", "сколько будет", "калькулятор" -> ACTION: CALC | ARG: [выражение или процент:значение]
- Если просят "что на экране", "опиши экран", "прочитай с экрана" -> ACTION: SCREEN | ARG: describe
- Если просят "что в этом окне", "прочитай окно" -> ACTION: SCREEN | ARG: window
- Если просят прочитать текст с экрана -> ACTION: SCREEN | ARG: read

Примеры:
Ввод: "запиши купить хлеб"
Ответ: ACTION: NOTE | ARG: Купить хлеб

Ввод: "поставь таймер на 5 минут"
Ответ: ACTION: TIMER | ARG: 300

Ввод: "сколько времени"
Ответ: ACTION: TIME | ARG: none

Ввод: "отмени последнюю заметку"
Ответ: ACTION: CANCEL | ARG: note

Ввод: "скопируй привет мир"
Ответ: ACTION: CLIPBOARD | ARG: write:привет мир

Ввод: "посчитай 2 плюс 2"
Ответ: ACTION: CALC | ARG: 2+2

Ввод: "сколько будет 15 процентов от 2500"
Ответ: ACTION: CALC | ARG: 15%:2500

Ввод: "посчитай 100 умножить на 5"
Ответ: ACTION: CALC | ARG: 100*5

Ввод: "что на экране"
Ответ: ACTION: SCREEN | ARG: describe

Ввод: "прочитай что написано на экране"
Ответ: ACTION: SCREEN | ARG: read

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
	log.Info("Bobik is listening for 'Эй, Бобик'...")

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

				// Non-blocking send to avoid blocking the recorder if the consumer is slow
				select {
				case audioChan <- samples:
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
				log.Warn("wake word error: %v", err)
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
	log.Info("Wake word detected!")
	if o.OnStateChange != nil {
		o.OnStateChange(StateListening)
	}
	o.Notifier.Notify(ctx, "Bobik", "Listening...")

	// 2. Transcribe Command
	text, err := o.STT.Transcribe(audioChan)
	if err != nil {
		log.Error("transcription error: %v", err)
		return
	}
	text = strings.TrimSpace(text)
	log.Debug("Transcribed: %s", text)

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
		log.Error("LLM error: %v", err)
		o.Notifier.Notify(ctx, "Bobik Error", "LLM failed")
		return
	}
	log.Debug("LLM Raw output: %s", rawOutput)

	// 4. Parse Action and Argument
	action, arg := o.parseLLMOutput(rawOutput)
	log.Info("Parsed Action: %s, Arg: %s", action, arg)

	// 5. Dispatch Tool
	switch action {
	case "NOTE":
		o.handleNoteAction(ctx, text, arg)
	case "TIMER":
		o.handleTimerAction(ctx, text, arg)
	case "TIME":
		o.handleTimeAction(ctx, text)
	case "CANCEL":
		o.handleCancelAction(ctx, text, arg)
	case "CLIPBOARD":
		o.handleClipboardAction(ctx, text, arg)
	case "CALC":
		o.handleCalcAction(ctx, text, arg)
	case "SCREEN":
		o.handleScreenAction(ctx, text, arg)
	default:
		log.Warn("Unknown action: %s", action)
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
		log.Error("Save error: %v", err)
		o.Notifier.Notify(ctx, "Bobik Error", "Failed to save note")
		return
	}

	actionDesc := "Saved note"
	if isUpdate {
		actionDesc = "Updated last note"
	}
	o.Memory.Add(rawInput, fmt.Sprintf("%s: %s", actionDesc, noteContent))
	o.Notifier.Notify(ctx, "Bobik", "Заметка сохранена")
	o.speak(ctx, "Записал")
}

func (o *Orchestrator) handleTimerAction(ctx context.Context, rawInput, arg string) {
	seconds, err := strconv.Atoi(arg)
	if err != nil {
		log.Warn("Invalid timer arg: %s", arg)
		o.Notifier.Notify(ctx, "Bobik Error", "Ошибка времени")
		return
	}

	duration := time.Duration(seconds) * time.Second
	o.Timer.Start("Голосовой таймер", duration)

	o.Memory.Add(rawInput, fmt.Sprintf("Set timer for %d seconds", seconds))
	o.Notifier.Notify(ctx, "Bobik", fmt.Sprintf("Таймер запущен на %d сек", seconds))
	o.speak(ctx, "Таймер запущен")
}

func (o *Orchestrator) handleTimeAction(ctx context.Context, rawInput string) {
	currentTime := o.Clock.GetCurrentTime()
	o.Notifier.Notify(ctx, "Bobik Time", currentTime)
	o.speak(ctx, "Сейчас "+currentTime)
	o.Memory.Add(rawInput, "Reported current time")
}

func (o *Orchestrator) handleCancelAction(ctx context.Context, rawInput, arg string) {
	arg = strings.ToLower(strings.TrimSpace(arg))

	var cancelled []string

	// Cancel note
	if arg == "note" || arg == "all" {
		if err := o.Obsidian.DeleteLastNote(); err == nil {
			cancelled = append(cancelled, "заметка")
		} else {
			log.Debug("No note to cancel: %v", err)
		}
	}

	// Cancel timer(s)
	if arg == "timer" || arg == "all" {
		count := o.Timer.CancelAll()
		if count > 0 {
			cancelled = append(cancelled, fmt.Sprintf("%d таймер(ов)", count))
		}
	}

	if len(cancelled) == 0 {
		o.Notifier.Notify(ctx, "Bobik", "Нечего отменять")
		o.speak(ctx, "Нечего отменять")
		return
	}

	msg := "Отменено: " + strings.Join(cancelled, ", ")
	o.Memory.Add(rawInput, msg)
	o.Notifier.Notify(ctx, "Bobik", msg)
	o.speak(ctx, "Отменено")
}

func (o *Orchestrator) handleClipboardAction(ctx context.Context, rawInput, arg string) {
	if o.Clipboard == nil {
		o.Notifier.Notify(ctx, "Bobik Error", "Буфер обмена недоступен")
		return
	}

	arg = strings.TrimSpace(arg)

	switch {
	case arg == "read":
		content, err := o.Clipboard.Read()
		if err != nil {
			log.Error("Clipboard read error: %v", err)
			o.Notifier.Notify(ctx, "Bobik Error", "Не удалось прочитать буфер")
			return
		}
		// Truncate for notification if too long
		display := content
		if len(display) > 100 {
			display = display[:100] + "..."
		}
		o.Notifier.Notify(ctx, "Буфер обмена", display)
		o.speak(ctx, "В буфере: "+display)
		o.Memory.Add(rawInput, "Read clipboard")

	case arg == "note":
		content, err := o.Clipboard.Read()
		if err != nil {
			log.Error("Clipboard read error: %v", err)
			o.Notifier.Notify(ctx, "Bobik Error", "Не удалось прочитать буфер")
			return
		}
		if content == "" {
			o.Notifier.Notify(ctx, "Bobik", "Буфер пуст")
			return
		}
		if err := o.Obsidian.AppendToDailyNote(content); err != nil {
			log.Error("Save note error: %v", err)
			o.Notifier.Notify(ctx, "Bobik Error", "Не удалось сохранить заметку")
			return
		}
		o.Notifier.Notify(ctx, "Bobik", "Буфер сохранен в заметку")
		o.speak(ctx, "Сохранено")
		o.Memory.Add(rawInput, "Saved clipboard to note")

	case strings.HasPrefix(arg, "write:"):
		content := strings.TrimPrefix(arg, "write:")
		content = strings.TrimSpace(content)
		if err := o.Clipboard.Write(content); err != nil {
			log.Error("Clipboard write error: %v", err)
			o.Notifier.Notify(ctx, "Bobik Error", "Не удалось записать в буфер")
			return
		}
		o.Notifier.Notify(ctx, "Bobik", "Скопировано в буфер")
		o.speak(ctx, "Скопировано")
		o.Memory.Add(rawInput, "Wrote to clipboard: "+content)

	default:
		o.Notifier.Notify(ctx, "Bobik", "Неизвестная операция с буфером")
	}
}

func (o *Orchestrator) handleCalcAction(ctx context.Context, rawInput, arg string) {
	if o.Calc == nil {
		o.Notifier.Notify(ctx, "Bobik Error", "Калькулятор недоступен")
		return
	}

	arg = strings.TrimSpace(arg)

	var result float64
	var err error

	// Check for percentage format: "15%:2500"
	if strings.Contains(arg, "%:") {
		parts := strings.SplitN(arg, "%:", 2)
		if len(parts) == 2 {
			percent, err1 := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
			value, err2 := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
			if err1 == nil && err2 == nil {
				result = o.Calc.Percentage(percent, value)
			} else {
				err = fmt.Errorf("invalid percentage format")
			}
		}
	} else {
		// Regular expression evaluation
		result, err = o.Calc.Eval(arg)
	}

	if err != nil {
		log.Error("Calc error: %v", err)
		o.Notifier.Notify(ctx, "Bobik Error", "Ошибка вычисления")
		o.speak(ctx, "Не могу посчитать")
		return
	}

	formatted := o.Calc.FormatResult(result)
	o.Notifier.Notify(ctx, "Результат", formatted)
	o.speak(ctx, formatted)
	o.Memory.Add(rawInput, fmt.Sprintf("Calculated: %s = %s", arg, formatted))
}

// speak uses TTS if available.
func (o *Orchestrator) speak(ctx context.Context, text string) {
	if o.TTS != nil {
		o.TTS.SpeakAsync(ctx, text)
	}
}

// handleScreenAction обрабатывает команды анализа экрана с использованием vision модели.
func (o *Orchestrator) handleScreenAction(ctx context.Context, rawInput, arg string) {
	// Проверяем доступность компонентов
	if o.Screen == nil {
		o.Notifier.Notify(ctx, "Bobik Error", "Скриншоты недоступны")
		o.speak(ctx, "Скриншоты недоступны")
		return
	}
	if o.VisionLLM == nil {
		o.Notifier.Notify(ctx, "Bobik Error", "Vision модель не настроена")
		o.speak(ctx, "Vision модель не настроена")
		return
	}

	arg = strings.ToLower(strings.TrimSpace(arg))

	o.Notifier.Notify(ctx, "Bobik", "Делаю скриншот...")
	o.speak(ctx, "Секунду")

	// Определяем тип захвата
	var base64Image, filePath string
	var err error

	if arg == "window" {
		base64Image, filePath, err = o.Screen.CaptureWindow()
	} else {
		base64Image, filePath, err = o.Screen.Capture()
	}

	if err != nil {
		log.Error("Screenshot error: %v", err)
		o.Notifier.Notify(ctx, "Bobik Error", "Не удалось сделать скриншот")
		o.speak(ctx, "Не удалось сделать скриншот")
		return
	}

	// Очистим файл после обработки
	defer func() {
		if err := o.Screen.Cleanup(filePath); err != nil {
			log.Debug("Failed to cleanup screenshot: %v", err)
		}
	}()

	// Формируем промпт в зависимости от типа запроса
	var visionPrompt string
	switch arg {
	case "read":
		visionPrompt = "Прочитай весь текст, который ты видишь на этом скриншоте. Выведи только текст, без комментариев."
	case "window":
		visionPrompt = "Опиши содержимое этого окна. Что это за программа? Что на экране?"
	default: // describe
		visionPrompt = "Опиши что ты видишь на этом скриншоте. Кратко, 2-3 предложения."
	}

	o.Notifier.Notify(ctx, "Bobik", "Анализирую изображение...")

	// Отправляем в vision модель
	response, err := o.VisionLLM.GenerateWithImages(ctx, "", visionPrompt, []string{base64Image})
	if err != nil {
		log.Error("Vision LLM error: %v", err)
		o.Notifier.Notify(ctx, "Bobik Error", "Ошибка анализа изображения")
		o.speak(ctx, "Не удалось проанализировать")
		return
	}

	response = strings.TrimSpace(response)

	// Показываем результат
	// Ограничиваем длину для уведомления
	displayText := response
	if len(displayText) > 200 {
		displayText = displayText[:200] + "..."
	}

	o.Notifier.Notify(ctx, "Экран", displayText)

	// Для TTS ограничиваем ещё больше
	speakText := response
	if len(speakText) > 150 {
		speakText = speakText[:150]
	}
	o.speak(ctx, speakText)

	o.Memory.Add(rawInput, fmt.Sprintf("Screen analysis: %s", displayText))
}
