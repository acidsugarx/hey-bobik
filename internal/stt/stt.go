package stt

import (
	"encoding/json"
	"fmt"
	vosk "github.com/alphacep/vosk-api/go"
)

// Engine handles speech-to-text and wake word detection using Vosk.
type Engine struct {
	ModelPath string
	model     *vosk.VoskModel
}

// NewEngine creates a new Vosk engine.
func NewEngine(modelPath string) (*Engine, error) {
	model, err := vosk.NewModel(modelPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load vosk model: %w", err)
	}
	return &Engine{
		ModelPath: modelPath,
		model:     model,
	}, nil
}

// RecognitionResult represents the JSON output from Vosk.
type RecognitionResult struct {
	Text string `json:"text"`
}

// PartialResult represents the JSON partial output from Vosk.
type PartialResult struct {
	Partial string `json:"partial"`
}

// ListenForWakeWord listens to an audio stream and returns true when the wake word is detected.
// grammar is a JSON string of allowed words, e.g., `["эй бобик", "бобик", "[unk]"]`
func (e *Engine) ListenForWakeWord(audioChan <-chan []int16, grammar string, wakeWord string) (bool, error) {
	rec, err := vosk.NewRecognizerGrm(e.model, 16000.0, grammar)
	if err != nil {
		return false, fmt.Errorf("failed to create recognizer: %w", err)
	}
	defer rec.Free()

	for samples := range audioChan {
		// Convert int16 to byte buffer (Vosk expects bytes)
		byteBuf := make([]byte, len(samples)*2)
		for i, s := range samples {
			byteBuf[i*2] = byte(s & 0xff)
			byteBuf[i*2+1] = byte(s >> 8)
		}

		if rec.AcceptWaveform(byteBuf) != 0 {
			var res RecognitionResult
			if err := json.Unmarshal([]byte(rec.Result()), &res); err != nil {
				continue
			}
			if fmt.Sprintf("%s", res.Text) == wakeWord {
				return true, nil
			}
		}
	}
	return false, nil
}

// Transcribe records audio until silence and returns the text.
func (e *Engine) Transcribe(audioChan <-chan []int16) (string, error) {
	rec, err := vosk.NewRecognizer(e.model, 16000.0)
	if err != nil {
		return "", fmt.Errorf("failed to create recognizer: %w", err)
	}
	defer rec.Free()

	var fullText string
	for samples := range audioChan {
		byteBuf := make([]byte, len(samples)*2)
		for i, s := range samples {
			byteBuf[i*2] = byte(s & 0xff)
			byteBuf[i*2+1] = byte(s >> 8)
		}

		if rec.AcceptWaveform(byteBuf) != 0 {
			var res RecognitionResult
			if err := json.Unmarshal([]byte(rec.Result()), &res); err != nil {
				continue
			}
			if res.Text != "" {
				fullText += res.Text + " "
			}
			// In a real implementation, we would detect silence to stop.
			// For this MVP, we might stop after the first result or use a timeout.
			return fullText, nil 
		}
	}
	return fullText, nil
}

// Close releases Vosk resources.
func (e *Engine) Close() {
	if e.model != nil {
		e.model.Free()
	}
}
