# Specification: Bobik MVP Core

## Overview
This track implements the foundational "Ear", "Brain", and "Hands" modules for Bobik on Arch Linux. The goal is to have a background service that listens for "Эй, Бобик", records a note command, processes it via a local Ollama instance, and appends a formatted Markdown note to a daily file in the user's Obsidian `SECOND_BRAIN`.

## Functional Requirements
- **Wake Word Detection:** Continuously monitor the microphone for the phrase "Эй, Бобик" using Vosk with a focused grammar to minimize CPU usage.
- **Command Capture:** Upon wake word detection, trigger a system notification and record the subsequent audio until silence is detected.
- **Speech-to-Text (STT):** Convert the recorded command audio to text using Vosk.
- **LLM Orchestration:** Send the transcribed text to a local Ollama API (llama3.1:8b) to extract the note content.
- **Obsidian Integration:**
    - Target directory: `~/SECOND_BRAIN/SECOND_BRAIN/`.
    - Format: Append to a daily file (e.g., `2026-01-19.md`).
    - Content: Include a YAML frontmatter (if the file is new) and a Markdown entry with a timestamp and "Bobik" source tag.
- **Notifications:** Use `notify-send` for:
    - Wake word detected ("Listening...").
    - Note saved ("Note saved to Daily Notes").
    - Errors (e.g., "Ollama connection failed").

## Technical Constraints
- **Language:** Go 1.25.6.
- **Audio Capture:** PortAudio.
- **STT:** Vosk API.
- **LLM:** Ollama REST API.
- **OS:** Arch Linux.
- **Architecture:** Modular Go packages for `audio`, `stt`, `llm`, and `tools`.

## Success Criteria
- [ ] Service starts and initializes PortAudio/Vosk without errors.
- [ ] Saying "Эй, Бобик" triggers a system notification.
- [ ] A voice command like "сделай заметку купить молоко" results in a new entry in the daily Obsidian file.
- [ ] No audio sounds are played (visual notifications only).
- [ ] CPU usage in standby mode is minimal.
