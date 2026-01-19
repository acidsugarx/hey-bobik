# Specification: Intelligence Boost

## Overview
Improve Bobik's command recognition and execution accuracy to a "10/10" level by upgrading the LLM backend to Qwen 3, implementing a post-processing stage for STT output, and introducing short-term context memory.

## Functional Requirements
- **LLM Upgrade:** Transition from Llama 3.1 to Qwen3-8B via Ollama.
- **STT Post-processing:** 
    - LLM should receive "raw" text from Vosk.
    - LLM must first "clean up" the text (fix grammar, case, and homophones).
    - Cleaned text is then used for tool extraction.
- **Dynamic Grammar:**
    - Configure Vosk with a prioritized list of action verbs (запиши, сделай, напомни, поставь).
- **Context Memory:**
    - Store the last 5 user commands and system actions in memory.
    - Include this context in the LLM prompt to allow for follow-up commands (e.g., "исправь это").
- **Improved System Prompt:** Refined prompts for Qwen 3 to ensure strict output format and high-quality Russian prose.

## Technical Constraints
- **Model:** `qwen3:8b` (quantized to fit in 8GB VRAM).
- **Architecture:** Update `internal/llm` and `internal/orchestrator`.
- **Memory:** Use a simple in-memory ring buffer for short-term context.

## Success Criteria
- [ ] LLM successfully switches to Qwen 3.
- [ ] Raw, garbled STT output is correctly "rehabilitated" by the LLM.
- [ ] Commands like "исправь последнюю заметку" work using context memory.
- [ ] Vosk's recognition of action verbs is noticeably improved.
