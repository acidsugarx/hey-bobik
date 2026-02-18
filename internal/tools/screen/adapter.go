// Package screen предоставляет адаптер для интеграции скриншотов с оркестратором.
package screen

// Adapter адаптирует Tool для использования в оркестраторе.
type Adapter struct {
	tool *Tool
}

// NewAdapter создает адаптер для инструмента скриншотов.
func NewAdapter(tool *Tool) *Adapter {
	return &Adapter{tool: tool}
}

// Capture делает скриншот всего экрана.
// Возвращает base64 изображения, путь к файлу и ошибку.
func (a *Adapter) Capture() (base64Image string, filePath string, err error) {
	result, err := a.tool.Capture()
	if err != nil {
		return "", "", err
	}
	return result.Base64, result.FilePath, nil
}

// CaptureWindow делает скриншот активного окна.
func (a *Adapter) CaptureWindow() (base64Image string, filePath string, err error) {
	result, err := a.tool.CaptureWindow()
	if err != nil {
		return "", "", err
	}
	return result.Base64, result.FilePath, nil
}

// Cleanup удаляет временный файл скриншота.
func (a *Adapter) Cleanup(filePath string) error {
	return a.tool.Cleanup(&ScreenshotResult{FilePath: filePath})
}

// IsAvailable проверяет доступность инструмента.
func (a *Adapter) IsAvailable() bool {
	return a.tool.IsAvailable()
}

// GetBackend возвращает используемый бэкенд.
func (a *Adapter) GetBackend() string {
	return a.tool.GetAvailableBackend()
}
