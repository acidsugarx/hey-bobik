# Tech Stack: Bobik - Linux Local Voice Agent

## Core Language & Runtime
- **Language:** Go (Golang)
- **Version:** 1.25.6 (Latest stable)
- **Reason:** Superior concurrency support for handling multiple audio/processing streams and low-overhead as a system daemon.

## Audio Processing (The Ear)
- **Audio Capture:** [PortAudio](http://www.portaudio.com/) (Go bindings: `gopkg.in/hraban/portaudio.v1`)
- **Wake Word & STT:** [Vosk](https://alphacephei.com/vosk/) (Go bindings: `github.com/alphacep/vosk-api/go`)
- **Reason:** PortAudio provides stable microphone access. Vosk allows for efficient, offline, grammar-based recognition which is critical for a high-accuracy, low-CPU wake word engine.

## Large Language Model (The Brain)
- **Backend:** [Ollama](https://ollama.com/)
- **Models:** `llama3.1:8b` (Default) or `mistral-nemo:12b`
- **Interface:** REST API via `net/http`
- **Reason:** Ollama simplifies local model management and inference, specifically optimized for NVIDIA GPUs (RTX 4060).

## System Integration (The Hands)
- **Notifications:** `notify-send` (via `os/exec`)
- **File System:** Standard Go `os` and `path/filepath` packages for interacting with `~/SECOND_BRAIN/SECOND_BRAIN`.
- **Reason:** Minimalist and reliable integration with standard Linux environments.

## Development & Build Tools
- **Build System:** `go build`
- **Dependency Management:** Go Modules (`go.mod`)
- **Testing:** Standard Go `testing` package
- **Version Control:** Git
