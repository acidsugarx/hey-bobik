# AGENTS.md — hey-bobik

Bobik is a local, privacy-focused Linux voice assistant written in Go. It listens
for the wake word "Эй, Бобик", transcribes voice via Vosk, routes commands through
a local LLM (Ollama), and executes actions (notes, timers, clipboard, calculations,
screen analysis, TTS).

## Build / Run / Test Commands

```bash
# Install dependencies
go mod tidy

# Build the binary
go build ./cmd/bobik

# Run the application
go run ./cmd/bobik -config config.example.json

# Run ALL tests
go test ./...

# Run a SINGLE test by name (regex match)
go test ./internal/orchestrator/ -run TestHandleCommand

# Run tests in a single package
go test ./internal/llm/

# Run tests with verbose output
go test -v ./internal/config/

# Run tests with coverage
go test -coverprofile=coverage.out ./...

# Format all code (mandatory — must pass before commit)
gofmt -w .
# or
go fmt ./...

# Vet (static analysis)
go vet ./...
```

There is no Makefile, no CI/CD pipeline, no `.golangci.yml`. The project uses
standard Go tooling only. Target **>80% code coverage** for all modules.

## Project Structure

```
cmd/bobik/main.go          — Entry point, wiring, signal handling (no business logic)
internal/
  audio/                    — PortAudio microphone capture
  config/                   — JSON config + env overrides (BOBIK_*)
  llm/                      — Ollama HTTP client (text + vision)
  logger/                   — Custom leveled logger
  orchestrator/             — Core coordinator, tool dispatch, context memory
  stt/                      — Vosk speech-to-text + wake word detection
  tools/{calc,clipboard,clock,notifier,obsidian,screen,timer,tts}/
  ui/tray/                  — System tray icon (getlantern/systray)
conductor/                  — Project management docs (PRD, plans, styleguides)
```

## Code Style

### Formatting
- **All code MUST be formatted with `gofmt`**. No exceptions.
- Indentation: tabs (handled by `gofmt`).
- No strict line length limit — let `gofmt` handle wrapping.

### Naming Conventions
- **Files**: `snake_case.go` (e.g., `context_test.go`, `config.go`).
- **Packages**: short, single-word, lowercase (e.g., `calc`, `llm`, `tray`, `stt`).
- **Exported types**: `PascalCase` — `Config`, `Client`, `Orchestrator`, `Engine`.
- **Unexported**: `camelCase` — `parseLLMOutput`, `handleCommand`, `detectBackend`.
- **No `Get` prefix** on getters: use `Owner()` not `GetOwner()`.
- **No `I` prefix** on interfaces: name by behavior (`Recorder`, `Notifier`, `LLMClient`).
- **Constants**: unexported `camelCase` (`defaultTimeout`, `systemPrompt`), exported `PascalCase` (`StateIdle`, `LevelDebug`).
- **Iota enums**: `type State int` with `StateIdle`, `StateListening`, `StateThinking`.

### Imports
Imports are grouped with a blank line separating stdlib from everything else.
Internal and third-party packages are mixed in the second group. No import aliases
except when needed for clarity (e.g., `vosk "github.com/alphacep/vosk-api/go"`).

```go
import (
    "context"
    "fmt"
    "strings"
    "time"

    "hey-bobik/internal/logger"
    "hey-bobik/internal/config"
)
```

Module path is `hey-bobik` — all internal imports use `hey-bobik/internal/...`.

### Error Handling
- **Always wrap errors** with `fmt.Errorf("context: %w", err)`.
- **Early return** on error (guard clause pattern). Never nest happy path.
- **Never panic**. No `panic()` anywhere in the codebase.
- **Never discard errors** with `_`. Check every error explicitly.
- **Fatal errors in main only**: `log.Error(...)` then `os.Exit(1)`.
- **Nil-check optional services** before use and notify the user:
  ```go
  if o.Clipboard == nil {
      o.Notifier.Notify(ctx, "Bobik Error", "Буфер обмена недоступен")
      return
  }
  ```
- **Use `defer` for cleanup**: `defer f.Close()`, `defer resp.Body.Close()`.

### Types and Interfaces
- **Interfaces are defined at the consumer** (in `orchestrator.go`), not at the provider. This is idiomatic Go.
- **Interfaces are small** (1-3 methods each): `Recorder`, `Notifier`, `LLMClient`, etc.
- **Structs for all domain types**. No type aliases.
- **Constructor pattern**: `New()` or `NewXxx()` returning `*Type`:
  ```go
  func New(url, model string) *Client { return &Client{url: url, model: model} }
  ```
- **Dependency injection** via exported struct fields on `Orchestrator`, not constructor params.
- **Time injection** for testability: `Now func() time.Time` field.
- **Exec injection** for testability: `execFunc commandExec` field to mock `os/exec`.

### Testing
- **stdlib `testing` only** — no testify, no gomock, no external frameworks.
- **Every source file has a `_test.go` counterpart** in the same package.
- **Test naming**: `TestXxx`, `TestXxxError` for error paths, `TestXxxIntegration` for integration.
- **Table-driven tests** where applicable:
  ```go
  tests := []struct{ input string; expected Level }{
      {"debug", LevelDebug},
      {"info", LevelInfo},
  }
  for _, tt := range tests {
      result := ParseLevel(tt.input)
      if result != tt.expected { t.Errorf(...) }
  }
  ```
- **Inline mock types** defined in test files:
  ```go
  type mockLLM struct { response string }
  func (m *mockLLM) Generate(...) (string, error) { return m.response, nil }
  ```
- **`httptest.NewServer`** for HTTP mocking (LLM client tests).
- **`t.TempDir()` / `os.MkdirTemp`** for filesystem tests.
- **`t.Skip()`** for tests requiring physical resources (audio, screen).
- **TDD workflow**: write failing tests first (Red), implement to pass (Green), then refactor.

### Concurrency
- `sync.Mutex` / `sync.RWMutex` for shared state (`ContextMemory`, `Timer`, `Logger`).
- Goroutines for: audio loop, orchestrator main loop, TTS async, signal handling.
- Buffered channels: `audioChan chan []int16` (capacity 100).
- Non-blocking sends to avoid deadlocks:
  ```go
  select {
  case audioChan <- samples:
  default: // Drop if full
  }
  ```
- `context.Context` propagated through all operations; `context.WithCancel` for shutdown.

### Comments and Documentation
- **GoDoc comments** on all exported types and functions (English).
- **Inline comments** may be in Russian (the assistant's target language).
- Document *why*, not *what*.

## Commit Convention

Conventional Commits format:
```
<type>(<scope>): <description>
```
- **Types**: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`
- **Scopes**: `ui`, `tools`, `orchestrator`, `stt`, `audio`, `config`, `llm`
- **Special**: `conductor(plan):` and `conductor(checkpoint):` for project management

## Key Dependencies

| Component     | Package                                     |
|---------------|---------------------------------------------|
| Audio capture | `github.com/gordonklaus/portaudio`          |
| STT           | `github.com/alphacep/vosk-api/go`           |
| LLM           | Ollama REST API (`net/http`)                |
| System tray   | `github.com/getlantern/systray`             |
| Notifications | `notify-send` via `os/exec`                 |
| TTS           | `espeak-ng` or `piper` via `os/exec`        |
| Clipboard     | `xclip`/`xsel`/`wl-paste` via `os/exec`    |
| Screenshots   | `gnome-screenshot`/`scrot`/`grim` via exec  |
