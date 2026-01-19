# Implementation Plan: System Tray Indicator

## Phase 1: Tray Infrastructure [checkpoint: ffb63a0]
Integrate the tray library and set up the main thread loop.

- [x] Task: Set up Tray Library & Icons (45ce8f4)
    - [x] `go get github.com/getlantern/systray`
    - [x] Add placeholder icon assets (embedded bytes)
    - [x] Implement `internal/ui/tray` package
- [x] Task: Refactor Main for Tray Compatibility (c7d4169)
    - [x] Move orchestrator logic to a goroutine to free up the main thread for the tray
- [ ] Task: Conductor - User Manual Verification 'Phase 1' (Protocol in workflow.md)

## Phase 2: State Integration
Connect the orchestrator states to the tray icon.

- [ ] Task: Implement State Notification
    - [ ] Define `State` type and a channel/callback for updates
    - [ ] Update orchestrator to emit state changes (Idle -> Listening -> Thinking -> Idle)
- [ ] Task: Update Tray Visuals
    - [ ] Implement icon switching logic in `internal/ui/tray`
- [ ] Task: Conductor - User Manual Verification 'Phase 2' (Protocol in workflow.md)
