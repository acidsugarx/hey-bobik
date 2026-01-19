# Implementation Plan: Bobik MVP Core

## Phase 1: Project Scaffolding & Basic Integration
Initialize the Go project and implement the most basic system integrations (Notifications and File System).

- [x] Task: Initialize Go module and project structure (2688d91)
    - [x] `go mod init hey-bobik`
    - [x] Create directory structure: `internal/audio`, `internal/stt`, `internal/llm`, `internal/tools`, `cmd/bobik`
- [ ] Task: Implement Notification Tool
    - [ ] Write Tests: Verify `notify-send` command execution
    - [ ] Implement `internal/tools/notifier` package
- [ ] Task: Implement Obsidian Note Tool (Daily Notes)
    - [ ] Write Tests: Verify daily file creation and appending with YAML frontmatter
    - [ ] Implement `internal/tools/obsidian` package
- [ ] Task: Conductor - User Manual Verification 'Phase 1' (Protocol in workflow.md)

## Phase 2: Audio & Wake Word (The Ear)
Implement microphone capture and the Vosk-based wake word detection loop.

- [ ] Task: Implement Audio Capture
    - [ ] Write Tests: Mock PortAudio stream and verify data capture
    - [ ] Implement `internal/audio` package
- [ ] Task: Implement Wake Word Detection (Vosk)
    - [ ] Write Tests: Verify phrase recognition using a sample audio buffer
    - [ ] Implement `internal/stt` package with Vosk integration
- [ ] Task: Implement Wake Word Loop
    - [ ] Integrate audio capture and Vosk into a continuous loop
- [ ] Task: Conductor - User Manual Verification 'Phase 2' (Protocol in workflow.md)

## Phase 3: Brain & Orchestration
Connect the transcription to Ollama and orchestrate the full flow.

- [ ] Task: Implement Ollama Client
    - [ ] Write Tests: Mock Ollama API response
    - [ ] Implement `internal/llm` package (HTTP client)
- [ ] Task: Implement Orchestrator Logic
    - [ ] Write Tests: Verify flow from "Wake" -> "Record" -> "Transcribe" -> "Process" -> "Note"
    - [ ] Implement main orchestrator in `internal/orchestrator`
- [ ] Task: Create Main Entry Point
    - [ ] Implement `cmd/bobik/main.go` to wire all components
- [ ] Task: Conductor - User Manual Verification 'Phase 3' (Protocol in workflow.md)
