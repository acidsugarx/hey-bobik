# Implementation Plan: Timers & Time Info

## Phase 1: New Tools [checkpoint: 7509757]
Implement the standalone timer and time reporting logic.

- [x] Task: Implement Timer Tool (621dc61)
    - [x] Write Tests: Verify timer firing after delay
    - [x] Implement `internal/tools/timer` package
- [x] Task: Implement Time Reporting Tool (4d5a481)
    - [x] Write Tests: Verify time formatting
    - [x] Implement `internal/tools/clock` package
- [ ] Task: Conductor - User Manual Verification 'Phase 1' (Protocol in workflow.md)

## Phase 2: Orchestration & Routing
Update the orchestrator to handle multiple tools.

- [x] Task: Update System Prompt for Tool Routing (3eb05d3)
    - [x] Refine Qwen 3 prompt to support `ACTION` and `ARG` format
- [x] Task: Update Orchestrator Logic (3eb05d3)
    - [x] Write Tests: Verify routing between NOTE, TIMER, and TIME
    - [x] Implement tool dispatcher in `internal/orchestrator`
- [ ] Task: Conductor - User Manual Verification 'Phase 2' (Protocol in workflow.md)
