// Package screen предоставляет функциональность для захвата скриншотов.
// Поддерживает несколько бэкендов: gnome-screenshot, scrot, grim (Wayland).
package screen

import (
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// Tool представляет инструмент для работы со скриншотами.
type Tool struct {
	// TempDir - директория для временных файлов (по умолчанию /tmp)
	TempDir string
	// Backend - предпочитаемый бэкенд (auto, gnome-screenshot, scrot, grim)
	Backend string
}

// ScreenshotResult содержит результат захвата экрана.
type ScreenshotResult struct {
	// FilePath - путь к сохраненному файлу
	FilePath string
	// Base64 - изображение в формате base64 для отправки в LLM
	Base64 string
}

// New создает новый инструмент для скриншотов.
func New() *Tool {
	return &Tool{
		TempDir: os.TempDir(),
		Backend: "auto",
	}
}

// NewWithConfig создает инструмент с настройками.
func NewWithConfig(tempDir, backend string) *Tool {
	if tempDir == "" {
		tempDir = os.TempDir()
	}
	if backend == "" {
		backend = "auto"
	}
	return &Tool{
		TempDir: tempDir,
		Backend: backend,
	}
}

// IsAvailable проверяет, доступен ли хотя бы один инструмент для скриншотов.
func (t *Tool) IsAvailable() bool {
	return t.detectBackend() != ""
}

// GetAvailableBackend возвращает название доступного бэкенда.
func (t *Tool) GetAvailableBackend() string {
	return t.detectBackend()
}

// detectBackend определяет доступный бэкенд для скриншотов.
func (t *Tool) detectBackend() string {
	if t.Backend != "auto" {
		if _, err := exec.LookPath(t.Backend); err == nil {
			return t.Backend
		}
		return ""
	}

	// Порядок предпочтения: gnome-screenshot (наиболее распространен), scrot, grim (Wayland)
	backends := []string{"gnome-screenshot", "scrot", "grim", "spectacle", "maim"}
	for _, b := range backends {
		if _, err := exec.LookPath(b); err == nil {
			return b
		}
	}
	return ""
}

// Capture делает скриншот всего экрана и возвращает его как base64.
func (t *Tool) Capture() (*ScreenshotResult, error) {
	backend := t.detectBackend()
	if backend == "" {
		return nil, fmt.Errorf("screenshot tool not found (install gnome-screenshot, scrot, or grim)")
	}

	// Создаем временный файл
	filename := fmt.Sprintf("bobik-screenshot-%d.png", time.Now().UnixNano())
	filePath := filepath.Join(t.TempDir, filename)

	var cmd *exec.Cmd

	// Настраиваем команду в зависимости от бэкенда
	switch backend {
	case "gnome-screenshot":
		cmd = exec.Command("gnome-screenshot", "-f", filePath)
	case "scrot":
		cmd = exec.Command("scrot", filePath)
	case "grim":
		// Grim для Wayland
		cmd = exec.Command("grim", filePath)
	case "spectacle":
		// KDE Spectacle
		cmd = exec.Command("spectacle", "-b", "-n", "-o", filePath)
	case "maim":
		cmd = exec.Command("maim", filePath)
	default:
		return nil, fmt.Errorf("unsupported backend: %s", backend)
	}

	// Выполняем команду
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("screenshot failed with %s: %w", backend, err)
	}

	// Проверяем, что файл создан
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("screenshot file not created")
	}

	// Читаем файл и конвертируем в base64
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read screenshot: %w", err)
	}

	base64Data := base64.StdEncoding.EncodeToString(data)

	return &ScreenshotResult{
		FilePath: filePath,
		Base64:   base64Data,
	}, nil
}

// CaptureArea делает скриншот выделенной области (интерактивный выбор).
// Поддерживается не всеми бэкендами.
func (t *Tool) CaptureArea() (*ScreenshotResult, error) {
	backend := t.detectBackend()
	if backend == "" {
		return nil, fmt.Errorf("screenshot tool not found")
	}

	filename := fmt.Sprintf("bobik-screenshot-%d.png", time.Now().UnixNano())
	filePath := filepath.Join(t.TempDir, filename)

	var cmd *exec.Cmd

	switch backend {
	case "gnome-screenshot":
		cmd = exec.Command("gnome-screenshot", "-a", "-f", filePath)
	case "scrot":
		cmd = exec.Command("scrot", "-s", filePath)
	case "grim":
		// Для Wayland с grim нужен slurp для выбора области
		// grim -g "$(slurp)" output.png
		return nil, fmt.Errorf("area selection not supported with grim directly, use slurp")
	case "spectacle":
		cmd = exec.Command("spectacle", "-b", "-r", "-n", "-o", filePath)
	case "maim":
		cmd = exec.Command("maim", "-s", filePath)
	default:
		return nil, fmt.Errorf("area capture not supported with %s", backend)
	}

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("area screenshot failed: %w", err)
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("screenshot file not created")
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read screenshot: %w", err)
	}

	base64Data := base64.StdEncoding.EncodeToString(data)

	return &ScreenshotResult{
		FilePath: filePath,
		Base64:   base64Data,
	}, nil
}

// CaptureWindow делает скриншот активного окна.
func (t *Tool) CaptureWindow() (*ScreenshotResult, error) {
	backend := t.detectBackend()
	if backend == "" {
		return nil, fmt.Errorf("screenshot tool not found")
	}

	filename := fmt.Sprintf("bobik-screenshot-%d.png", time.Now().UnixNano())
	filePath := filepath.Join(t.TempDir, filename)

	var cmd *exec.Cmd

	switch backend {
	case "gnome-screenshot":
		cmd = exec.Command("gnome-screenshot", "-w", "-f", filePath)
	case "scrot":
		cmd = exec.Command("scrot", "-u", filePath)
	case "spectacle":
		cmd = exec.Command("spectacle", "-b", "-a", "-n", "-o", filePath)
	case "maim":
		// maim требует xdotool для активного окна
		cmd = exec.Command("sh", "-c", fmt.Sprintf("maim -i $(xdotool getactivewindow) %s", filePath))
	default:
		// Fallback к полному экрану
		return t.Capture()
	}

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("window screenshot failed: %w", err)
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("screenshot file not created")
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read screenshot: %w", err)
	}

	base64Data := base64.StdEncoding.EncodeToString(data)

	return &ScreenshotResult{
		FilePath: filePath,
		Base64:   base64Data,
	}, nil
}

// Cleanup удаляет временный файл скриншота.
func (t *Tool) Cleanup(result *ScreenshotResult) error {
	if result == nil || result.FilePath == "" {
		return nil
	}
	return os.Remove(result.FilePath)
}
