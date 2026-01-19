# Implementation Plan: Timers & Time Info

## Phase 1: New Tools
Implement the standalone timer and time reporting logic.

- [x] Task: Implement Timer Tool (621dc61)
    - [x] Write Tests: Verify timer firing after delay
    - [x] Implement `internal/tools/timer` package
- [ ] Task: Implement Time Reporting Tool
    - [ ] Write Tests: Verify time formatting
    - [ ] Implement `internal/tools/clock` package
- [ ] Task: Conductor - User Manual Verification 'Phase 1' (Protocol in workflow.md)

## Phase 2: Orchestration & Routing
Update the orchestrator to handle multiple tools.

- [ ] Task: Update System Prompt for Tool Routing
    - [ ] Refine Qwen 3 prompt to support `ACTION` and `ARG` format
- [ ] Task: Update Orchestrator Logic
    - [ ] Write Tests: Verify routing between NOTE, TIMER, and TIME
    - [ ] Implement tool dispatcher in `internal/orchestrator`
- [ ] Task: Conductor - User Manual Verification 'Phase 2' (Protocol in workflow.md)
