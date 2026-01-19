# Specification: System Tray Indicator

## Overview
Provide visual feedback to the user regarding the agent's current state. A system tray icon will change based on whether Bobik is idle, recording audio, or processing a command with the LLM.

## Functional Requirements
- **System Tray Icon:**
    - Display a persistent icon in the Linux system tray.
    - Support three distinct visual states:
        - **Idle:** Waiting for "Эй, Бобик".
        - **Listening:** Active audio capture after wake word detection.
        - **Thinking:** LLM inference or tool execution in progress.
- **Context Menu:**
    - **Quit:** Terminate the application.
    - **Status Info:** Show current model or vault path (optional).
- **State Synchronization:**
    - The orchestrator must notify the tray component of every state change.

## Technical Constraints
- **Library:** `github.com/getlantern/systray`.
- **Graphics:** Use simple, high-contrast PNG or SVG icons.
- **Concurrency:** The tray runs on the main thread (required by GTK/Cocoa/Windows), so Bobik's logic must move to a goroutine.

## Success Criteria
- [ ] An icon appears in the Arch Linux tray (GNOME/KDE/i3/etc.).
- [ ] Saying "Эй, Бобик" changes the icon color/shape.
- [ ] Icon returns to idle after the note is saved.
- [ ] Right-clicking the icon and selecting "Quit" closes the app.
