package screen

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNew(t *testing.T) {
	tool := New()
	if tool == nil {
		t.Fatal("New() returned nil")
	}
	if tool.TempDir != os.TempDir() {
		t.Errorf("Expected TempDir %s, got %s", os.TempDir(), tool.TempDir)
	}
	if tool.Backend != "auto" {
		t.Errorf("Expected Backend 'auto', got %s", tool.Backend)
	}
}

func TestNewWithConfig(t *testing.T) {
	customDir := "/tmp/custom"
	tool := NewWithConfig(customDir, "scrot")

	if tool.TempDir != customDir {
		t.Errorf("Expected TempDir %s, got %s", customDir, tool.TempDir)
	}
	if tool.Backend != "scrot" {
		t.Errorf("Expected Backend 'scrot', got %s", tool.Backend)
	}
}

func TestNewWithConfigDefaults(t *testing.T) {
	// Пустые значения должны использовать defaults
	tool := NewWithConfig("", "")

	if tool.TempDir != os.TempDir() {
		t.Errorf("Expected TempDir %s, got %s", os.TempDir(), tool.TempDir)
	}
	if tool.Backend != "auto" {
		t.Errorf("Expected Backend 'auto', got %s", tool.Backend)
	}
}

func TestDetectBackend(t *testing.T) {
	tool := New()

	// Просто проверяем, что метод не паникует
	backend := tool.detectBackend()
	t.Logf("Detected backend: %s", backend)

	// IsAvailable должен соответствовать результату detectBackend
	if tool.IsAvailable() && backend == "" {
		t.Error("IsAvailable() returns true but detectBackend() returns empty string")
	}
	if !tool.IsAvailable() && backend != "" {
		t.Error("IsAvailable() returns false but detectBackend() returns non-empty string")
	}
}

func TestDetectBackendWithSpecificBackend(t *testing.T) {
	// Тест с несуществующим бэкендом
	tool := NewWithConfig("", "nonexistent-screenshot-tool")
	backend := tool.detectBackend()

	if backend != "" {
		t.Errorf("Expected empty backend for nonexistent tool, got %s", backend)
	}
}

func TestGetAvailableBackend(t *testing.T) {
	tool := New()
	backend := tool.GetAvailableBackend()

	// Должен соответствовать detectBackend
	if backend != tool.detectBackend() {
		t.Errorf("GetAvailableBackend() doesn't match detectBackend()")
	}
}

func TestCleanup(t *testing.T) {
	tool := New()

	// Создаем временный файл
	tmpFile := filepath.Join(os.TempDir(), "test-screenshot-cleanup.png")
	if err := os.WriteFile(tmpFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Проверяем, что файл существует
	if _, err := os.Stat(tmpFile); os.IsNotExist(err) {
		t.Fatal("Test file was not created")
	}

	// Вызываем Cleanup
	result := &ScreenshotResult{FilePath: tmpFile}
	if err := tool.Cleanup(result); err != nil {
		t.Errorf("Cleanup failed: %v", err)
	}

	// Проверяем, что файл удален
	if _, err := os.Stat(tmpFile); !os.IsNotExist(err) {
		t.Error("File was not deleted by Cleanup")
		os.Remove(tmpFile) // Cleanup вручную
	}
}

func TestCleanupNilResult(t *testing.T) {
	tool := New()

	// Не должен паниковать или возвращать ошибку для nil
	if err := tool.Cleanup(nil); err != nil {
		t.Errorf("Cleanup(nil) returned error: %v", err)
	}
}

func TestCleanupEmptyPath(t *testing.T) {
	tool := New()

	// Не должен паниковать или возвращать ошибку для пустого пути
	result := &ScreenshotResult{FilePath: ""}
	if err := tool.Cleanup(result); err != nil {
		t.Errorf("Cleanup with empty path returned error: %v", err)
	}
}

// Адаптер тесты
func TestAdapter(t *testing.T) {
	tool := New()
	adapter := NewAdapter(tool)

	if adapter == nil {
		t.Fatal("NewAdapter returned nil")
	}

	// IsAvailable должен соответствовать tool.IsAvailable
	if adapter.IsAvailable() != tool.IsAvailable() {
		t.Error("Adapter.IsAvailable() doesn't match Tool.IsAvailable()")
	}

	// GetBackend должен соответствовать tool.GetAvailableBackend
	if adapter.GetBackend() != tool.GetAvailableBackend() {
		t.Error("Adapter.GetBackend() doesn't match Tool.GetAvailableBackend()")
	}
}

func TestAdapterCleanup(t *testing.T) {
	tool := New()
	adapter := NewAdapter(tool)

	// Создаем временный файл
	tmpFile := filepath.Join(os.TempDir(), "test-adapter-cleanup.png")
	if err := os.WriteFile(tmpFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Вызываем Cleanup через адаптер
	if err := adapter.Cleanup(tmpFile); err != nil {
		t.Errorf("Adapter.Cleanup failed: %v", err)
	}

	// Проверяем, что файл удален
	if _, err := os.Stat(tmpFile); !os.IsNotExist(err) {
		t.Error("File was not deleted by Adapter.Cleanup")
		os.Remove(tmpFile)
	}
}

// Интеграционные тесты (пропускаются, если инструменты недоступны)
func TestCaptureIntegration(t *testing.T) {
	tool := New()
	if !tool.IsAvailable() {
		t.Skip("No screenshot tool available, skipping integration test")
	}

	// Примечание: этот тест делает реальный скриншот
	// В CI-окружении без дисплея может не работать
	result, err := tool.Capture()
	if err != nil {
		// Может быть ошибка если нет дисплея (headless environment)
		t.Skipf("Screenshot capture failed (probably no display): %v", err)
	}

	defer tool.Cleanup(result)

	// Проверяем результат
	if result.FilePath == "" {
		t.Error("FilePath is empty")
	}
	if result.Base64 == "" {
		t.Error("Base64 is empty")
	}

	// Проверяем, что файл существует
	if _, err := os.Stat(result.FilePath); os.IsNotExist(err) {
		t.Errorf("Screenshot file does not exist: %s", result.FilePath)
	}

	t.Logf("Screenshot captured: %s (base64 length: %d)", result.FilePath, len(result.Base64))
}
