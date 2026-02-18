package audio

import (
	"fmt"
	"github.com/gordonklaus/portaudio"
)

// Recorder handles audio capture from the microphone.
type Recorder struct {
	SampleRate int
	Channels   int
	stream     *portaudio.Stream
	buffer     []int16
}

// NewRecorder creates a new Recorder instance.
func NewRecorder(sampleRate, channels, bufferSize int) *Recorder {
	return &Recorder{
		SampleRate: sampleRate,
		Channels:   channels,
		buffer:     make([]int16, bufferSize),
	}
}

// Start begins audio capture.
func (r *Recorder) Start() error {
	err := portaudio.Initialize()
	if err != nil {
		return fmt.Errorf("failed to initialize PortAudio: %w", err)
	}

	stream, err := portaudio.OpenDefaultStream(r.Channels, 0, float64(r.SampleRate), len(r.buffer), r.buffer)
	if err != nil {
		portaudio.Terminate()
		return fmt.Errorf("failed to open default stream: %w", err)
	}

	err = stream.Start()
	if err != nil {
		stream.Close()
		portaudio.Terminate()
		return fmt.Errorf("failed to start stream: %w", err)
	}

	r.stream = stream
	return nil
}

// Read captures a chunk of audio into the internal buffer and returns a copy.
func (r *Recorder) Read() ([]int16, error) {
	err := r.stream.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read from stream: %w", err)
	}
	// Return a copy to avoid data corruption when buffer is reused
	result := make([]int16, len(r.buffer))
	copy(result, r.buffer)
	return result, nil
}

// Stop ends audio capture and cleans up resources.
func (r *Recorder) Stop() error {
	if r.stream != nil {
		r.stream.Stop()
		r.stream.Close()
	}
	return portaudio.Terminate()
}
