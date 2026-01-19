# Specification: Timers & Time Info

## Overview
Add specialized tools to Bobik beyond note-taking. The agent should be able to set timers and report the current system time when asked.

## Functional Requirements
- **Timer Tool:**
    - Parse duration from LLM output (e.g., "5 minutes" -> 300 seconds).
    - Run a background process that notifies the user when the time is up.
    - Message should be like: "Timer finished: [Name of timer or 'Time is up']".
- **Time Info Tool:**
    - Report the current system time via a visual notification.
- **Action Routing (Orchestration):**
    - The LLM must choose between the following actions:
        - `NOTE`: Create/update Obsidian note.
        - `TIMER`: Start a countdown.
        - `TIME`: Show current time.
    - Use a strict response format for tool calling (e.g., `ACTION: TIMER | ARG: 300`).

## Technical Constraints
- **Timer Logic:** Use Go goroutines and `time.AfterFunc` to avoid blocking.
- **Notifications:** Reuse the existing `notifier` package.
- **Parsing:** LLM prompt must ensure precise argument extraction for the timer (seconds).

## Success Criteria
- [ ] Saying "поставь таймер на 1 минуту" triggers a timer.
- [ ] After 60 seconds, a "Timer finished" notification appears.
- [ ] Saying "сколько времени" results in a notification with the current HH:MM.
- [ ] Note-taking functionality remains unaffected and accessible.
