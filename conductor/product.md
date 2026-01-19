# Initial Concept

Принято! Это важное изменение. Добавление **Wake Word** («Эй, Бобик») меняет логику работы микрофона: теперь мы не просто реагируем на любой шум, а фильтруем аудиопоток, пока не услышим заветную фразу.

Вот обновленный **PRD v1.1**. Изменения выделены.

---

# PRD v1.1: Linux Local Voice Agent ("Bobik")

### 1. Общее видение

Фоновый системный сервис (daemon) для Linux, который активируется голосовой командой **«Эй, Бобик»**, понимает естественный язык через локальную LLM и выполняет действия на компьютере.
**Ключевое отличие:** Режим "Сон" (постоянное прослушивание буфера в ожидании ключевой фразы) -> Режим "Активен" (запись команды) -> Выполнение.

### 2. Целевая система (Hardware Profile)

* **CPU:** Intel Core i5-12400f
* **GPU:** NVIDIA RTX 4060 (8GB VRAM) — *Inference LLM.*
* **OS:** Linux.

---

### 3. Архитектура системы

Модуль "Слух" (The Ear) теперь работает в двух режимах:

1. **Wake Word Detection:** Легковесный поток, анализирующий последние N секунд аудио на наличие фразы "эй бобик".
2. **Command Capture:** Активируется только после срабатывания триггера.

*(Остальная архитектура без изменений: Go Orchestrator -> Ollama -> Go Tools)*

---

### 4. Технологический стек

* **Язык:** Go (Golang) 1.22+.
* **Wake Word Engine & STT:** **Vosk**.
* *Почему:* Мы используем одну библиотеку и для активации, и для распознавания. Vosk позволяет задать грамматику (список слов), чтобы в режиме ожидания он распознавал *только* фразу-триггер, экономя CPU.


* **LLM Backend:** Ollama (`llama3.1:8b` / `mistral-nemo`).
* **Audio:** `portaudio`.

---

### 5. Функциональные требования

#### 5.1 Модуль "Слух" и Активация (Обновлено)

* **Режим ожидания (Hotword Loop):**
* Агент слушает микрофон постоянно.
* Используется "облегченное" распознавание (или проверка скользящего окна текста) на совпадение с фонемной маской **«Эй, Бобик»** (вариации: "бобик", "эй боб", "hey bobik").
* *Важно:* В этом режиме данные никуда не отправляются и не сохраняются.


* **Переход в активный режим:**
* Как только распознано "Эй, Бобик":
1. Воспроизводится короткий звук (Chime) или системное уведомление.
2. Начинается запись основной команды (до наступления тишины).





#### 5.2 Модуль "Мозг" (LLM)

* **Входные данные:** Текст команды *после* ключевой фразы. (Само "Эй, Бобик" вырезается из промпта, чтобы не смущать нейросеть).
* **System Prompt:** "Ты помощник Linux по имени Бобик..."

#### 5.3 Модуль "Действия" (Tools)

* *Без изменений (CreateNote, SetTimer, SystemNotification).*

#### 5.4 Пользовательский интерфейс (UX)

1. Пользователь: **"Эй, Бобик!"**
2. Система: *Звук "Дзынь"* (подтверждение, что слушает).
3. Пользователь: **"Запиши, что нужно купить молоко."**
4. Система: *Молчание (обработка)...* -> *Уведомление "Заметка сохранена"* + (опционально) голос "Сделано".

---

### 6. Реализация Wake Word на Go (Техническая заметка)

Чтобы не подключать сложные Python-библиотеки (типа Porcupine), мы сделаем это на Vosk. В Go это будет выглядеть примерно так:

```go
// Псевдокод логики Wake Word
func listenLoop() {
    // Настраиваем Vosk только на эти слова, чтобы повысить точность и снизить нагрузку
    grammar := `["эй бобик", "бобик", "[unk]"]` 
    rec.SetGrammar(grammar)

    for {
        buffer := readMic()
        if rec.AcceptWaveform(buffer) {
            result := rec.Result() // JSON
            if strings.Contains(result.Text, "бобик") {
                triggerActiveMode()
            }
        }
    }
}
```

# Product Definition: Bobik - Linux Local Voice Agent

## Initial Concept
Bobik is a local, privacy-focused voice assistant for Linux, designed to help users perform tasks without breaking their flow or switching contexts. It uses a "Wake Word" trigger to activate and executes commands via local LLM inference.

## Target Users
- Users who need a hands-free way to interact with their system while maintaining focus on their primary task.
- Privacy-conscious individuals who prefer local processing over cloud services.

## Core Goals
- **Maintain Flow:** Allow users to capture thoughts and perform system actions without leaving their current application or context.
- **Privacy & Security:** All audio processing, speech-to-text (STT), and LLM inference must happen locally on the user's machine.
- **Seamless Integration:** Direct integration with personal workflows, such as a "Second Brain" note-taking system.

## Key Features
- **Wake Word Detection:** High-accuracy activation on the phrase "Эй, Бобик" (and only this phrase) using Vosk.
- **Non-Intrusive Notifications:** Uses system notifications (`notify-send`) to confirm wake word detection and task completion, avoiding disruptive audio feedback.
- **Smart Note-Taking:** Automatically creates formatted notes in `~/SECOND_BRAIN/SECOND_BRAIN` based on user voice input.
- **Local Intelligence:** Leverages Ollama (llama3.1:8b / mistral-nemo) for natural language understanding and tool selection.
- **Sensitivity Control:** Configurable wake word sensitivity to minimize false positives in different environments.

## Success Criteria

- Reliable wake word detection with minimal false activations.

- Fast transition from wake word detection to command capture.

- Successful creation of notes in the specified local directory with correct formatting.



## Future Roadmap

- **Contextual Editing:** Allow the agent to modify any specific note in the current daily file, not just the last one, by identifying it via content or keywords.

- **Multimodal Interactions:** Support for image descriptions or system state analysis (e.g., "What's on my screen?").

- **External API Integrations:** Ability to send notes to other services (Slack, Telegram, etc.) via voice.
