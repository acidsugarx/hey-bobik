# Implementation Plan: Intelligence Boost

## Phase 1: Model & Infrastructure [checkpoint: 333551c]
Upgrade to Qwen 3 and update the LLM client configuration.

- [x] Task: Update Tech Stack & Configuration (6fb52cc)
    - [x] Update `conductor/tech-stack.md` to Qwen 3
    - [x] Update `cmd/bobik/main.go` default flags
- [x] Task: Implement Context Memory (5045106)
    - [x] Write Tests: Verify ring buffer for context storage
    - [x] Implement `internal/orchestrator/context.go`
- [ ] Task: Conductor - User Manual Verification 'Phase 1' (Protocol in workflow.md)

## Phase 2: Intelligence & Logic
Implement STT post-processing and refined prompting.

- [x] Task: Refine System Prompt for Qwen 3 (9b288e4)
    - [x] Update orchestrator with advanced multi-stage prompt (Clean -> Process)
- [ ] Task: Implement Dynamic Grammar in STT
    - [ ] Update `internal/stt` to accept and use a command-focused grammar list
- [ ] Task: Implement STT Post-processing Logic
    - [ ] Write Tests: Verify LLM-based text cleanup of garbled STT input
    - [ ] Update orchestrator flow to include the "Clean" step
- [ ] Task: Conductor - User Manual Verification 'Phase 2' (Protocol in workflow.md)
