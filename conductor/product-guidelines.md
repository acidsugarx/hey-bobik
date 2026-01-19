# Product Guidelines: Bobik - Linux Local Voice Agent

## Prose Style & Tone
- **Technical & Precise:** All system communications, logs, and documentation should be technically accurate and concise. Avoid unnecessary fluff or overly conversational language.
- **Clarity over Personality:** The agent should feel like a reliable system utility, not a social companion.

## User Feedback & Notifications
- **Status Updates:** Use `notify-send` for critical lifecycle events (Wake Word detected, Command started, Task completed).
- **Error Reporting:** Failures should be reported via system notifications with enough technical detail for the user to understand the point of failure (e.g., "LLM Inference Failed: Connection Timeout").
- **Visual Only:** Avoid audio feedback (TTS or chimes) unless explicitly requested, to maintain a quiet work environment.

## Integration Guidelines (Obsidian/Second Brain)
- **Format:** All notes must be saved as Markdown files (`.md`).
- **Metadata:** Include a YAML frontmatter at the top of each note for compatibility with Obsidian (e.g., timestamp, source: "voice-agent").
- **Storage:** Maintain a flat or logical structure within `~/SECOND_BRAIN/SECOND_BRAIN` as defined by the user's workflow.

## Resource Management
- **Idle Efficiency:** The Wake Word engine (Vosk) must be optimized for low CPU usage while in standby mode.
- **GPU Optimization:** Ensure LLM inference via Ollama is optimized for the NVIDIA RTX 4060, balancing response speed and VRAM usage.

## Development Principles
- **Modularity:** Tools and orchestrator components must be decoupled. Adding a new capability (e.g., "Set Timer") should involve adding a discrete Go package or module.
- **Security:** Never execute arbitrary shell commands directly from LLM output. Use predefined "tools" with strict argument parsing.
- **Test-Driven Development:** Core logic, especially audio processing and tool execution, should be covered by unit tests.
