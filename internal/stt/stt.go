package stt

import (
	"encoding/json"
	"fmt"
	"time"

	vosk "github.com/alphacep/vosk-api/go"
)

// Engine handles speech-to-text and wake word detection using Vosk.
type Engine struct {
	ModelPath string
	model     *vosk.VoskModel
}

// NewEngine creates a new Vosk engine.
func NewEngine(modelPath string) (*Engine, error) {
	vosk.SetLogLevel(-1)
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
	// If grammar is empty, use a very broad one or a default
	if grammar == "" {
		grammar = `["эй бобик", "[unk]"]`
	}
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

// Transcribe records audio for a short duration and returns the combined text.
func (e *Engine) Transcribe(audioChan <-chan []int16) (string, error) {
	rec, err := vosk.NewRecognizer(e.model, 16000.0)
	if err != nil {
		return "", fmt.Errorf("failed to create recognizer: %w", err)
	}
	defer rec.Free()

	var fullText string
	// Listen for a max of 7 seconds or until we get some results
	timeout := time.After(7 * time.Second)

	var silenceTimer *time.Timer
	silenceDelay := 1 * time.Second

	for {
		select {
		case <-timeout:
			var res RecognitionResult
			if err := json.Unmarshal([]byte(rec.FinalResult()), &res); err == nil {
				if res.Text != "" {
					fullText += res.Text
				}
			}
			return fullText, nil
		case <-func() <-chan time.Time {
			if silenceTimer != nil {
				return silenceTimer.C
			}
			return nil
		}():
			// Silence duration reached, return what we have
			var res RecognitionResult
			if err := json.Unmarshal([]byte(rec.FinalResult()), &res); err == nil {
				if res.Text != "" {
					fullText += res.Text
				}
			}
			return fullText, nil
		case samples, ok := <-audioChan:
			if !ok {
				return fullText, nil
			}
			byteBuf := make([]byte, len(samples)*2)
			for i, s := range samples {
				byteBuf[i*2] = byte(s & 0xff)
				byteBuf[i*2+1] = byte(s >> 8)
			}

			if rec.AcceptWaveform(byteBuf) == 1 {
				// Silence detected by Vosk, start/reset the silence timer
				if silenceTimer == nil {
					silenceTimer = time.NewTimer(silenceDelay)
				} else {
					if !silenceTimer.Stop() {
						select {
						case <-silenceTimer.C:
						default:
						}
					}
					silenceTimer.Reset(silenceDelay)
				}
			} else {
				// User is speaking, stop the timer if it exists
				if silenceTimer != nil {
					silenceTimer.Stop()
					silenceTimer = nil
				}
				
				// Optional: pull intermediate results to show progress (for future UI)
				var res RecognitionResult
				if err := json.Unmarshal([]byte(rec.PartialResult()), &res); err == nil {
					// We could append partial results here if needed
				}
			}
		}
	}
}

// Close releases Vosk resources.
func (e *Engine) Close() {
	if e.model != nil {
		e.model.Free()
	}
}
