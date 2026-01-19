package orchestrator

import (
	"context"
	"testing"
)

type mockRecorder struct {
	samples []int16
}

func (m *mockRecorder) Read() ([]int16, error) {
	return m.samples, nil
}

type mockSTT struct {
	wakeDetected bool
	transcription string
}

func (m *mockSTT) ListenForWakeWord(audioChan <-chan []int16, grammar string, wakeWord string) (bool, error) {
	return m.wakeDetected, nil
}

func (m *mockSTT) Transcribe(audioChan <-chan []int16) (string, error) {
	return m.transcription, nil
}

type mockNotifier struct {
	title   string
	message string
}

func (m *mockNotifier) Notify(ctx context.Context, title, message string) error {
	m.title = title
	m.message = message
	return nil
}

type mockLLM struct {
	response string
}

func (m *mockLLM) Generate(ctx context.Context, system, prompt string) (string, error) {
	return m.response, nil
}

type mockObsidian struct {
	content string
}

func (m *mockObsidian) AppendToDailyNote(content string) error {
	m.content = content
	return nil
}

func (m *mockObsidian) RewriteLastNote(content string) error {
	m.content = "REWRITTEN: " + content
	return nil
}

func TestOrchestratorFlow(t *testing.T) {
	rec := &mockRecorder{samples: make([]int16, 10)}
	stt := &mockSTT{wakeDetected: true, transcription: "сделай заметку тест"}
	notif := &mockNotifier{}
	llm := &mockLLM{response: "тест"}
	obs := &mockObsidian{}

	o := &Orchestrator{
		Recorder: rec,
		STT:      stt,
		Notifier: notif,
		LLM:      llm,
		Obsidian: obs,
		Memory:   NewContextMemory(5),
	}

	audioChan := make(chan []int16, 1)
	o.handleCommand(context.Background(), audioChan)

	if notif.title != "Bobik" || notif.message != "Note saved to Daily Notes" {
		t.Errorf("expected success notification, got %s: %s", notif.title, notif.message)
	}

	if obs.content != "тест" {
		t.Errorf("expected 'тест' in Obsidian, got %s", obs.content)
	}
}

func TestSTTPostProcessing(t *testing.T) {
	rec := &mockRecorder{samples: make([]int16, 10)}
	stt := &mockSTT{wakeDetected: true, transcription: "з опиши ка молоко"}
	notif := &mockNotifier{}
	llm := &mockLLM{response: "купить молоко"}
	obs := &mockObsidian{}

	o := &Orchestrator{
		Recorder: rec,
		STT:      stt,
		Notifier: notif,
		LLM:      llm,
		Obsidian: obs,
		Memory:   NewContextMemory(5),
	}

	audioChan := make(chan []int16, 1)
	o.handleCommand(context.Background(), audioChan)

	if obs.content != "купить молоко" {
		t.Errorf("expected 'купить молоко', got %s", obs.content)
	}
}

func TestOrchestratorUpdateFlow(t *testing.T) {
	rec := &mockRecorder{samples: make([]int16, 10)}
	stt := &mockSTT{wakeDetected: true, transcription: "исправь на кефир"}
	notif := &mockNotifier{}
	llm := &mockLLM{response: "UPDATE: купить кефир"}
	obs := &mockObsidian{}

	o := &Orchestrator{
		Recorder: rec,
		STT:      stt,
		Notifier: notif,
		LLM:      llm,
		Obsidian: obs,
		Memory:   NewContextMemory(5),
	}

	audioChan := make(chan []int16, 1)
	o.handleCommand(context.Background(), audioChan)

	if obs.content != "REWRITTEN: купить кефир" {
		t.Errorf("expected 'REWRITTEN: купить кефир', got %s", obs.content)
	}
}
