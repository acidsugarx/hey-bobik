package tray

import (
	"bytes"
	"github.com/getlantern/systray"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
)

// State represents the current visual state of the tray icon.
type State int

const (
	StateIdle State = iota
	StateListening
	StateThinking
)

// Manager handles the system tray icon and menu.
type Manager struct {
	onExit func()
}

// New creates a new tray manager.
func New(onExit func()) *Manager {
	return &Manager{
		onExit: onExit,
	}
}

// Run initializes and starts the system tray loop.
func (m *Manager) Run() {
	systray.Run(m.onReady, m.onExit)
}

func (m *Manager) onReady() {
	systray.SetTitle("Bobik")
	systray.SetTooltip("Bobik: Linux Voice Agent")

	m.SetState(StateIdle)

	mQuit := systray.AddMenuItem("Quit", "Quit Bobik")

	go func() {
		<-mQuit.ClickedCh
		systray.Quit()
	}()
}

// SetState updates the tray icon based on the provided state.
func (m *Manager) SetState(state State) {
	var c color.Color
	var label string
	switch state {
	case StateIdle:
		c = color.RGBA{128, 128, 128, 255} // Gray
		label = "IDLE"
	case StateListening:
		c = color.RGBA{0, 200, 0, 255}   // Darker Green
		label = "LISTENING"
	case StateThinking:
		c = color.RGBA{0, 100, 255, 255} // Strong Blue
		label = "THINKING"
	}
	
	log.Printf("Tray: Changing state to %s", label)
	systray.SetIcon(createCircleIcon(c))
}

func createCircleIcon(c color.Color) []byte {
	size := 64
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	
	// Transparent background
	draw.Draw(img, img.Bounds(), &image.Uniform{color.Transparent}, image.Point{}, draw.Src)
	
	// Draw a colored circle
	centerX, centerY := size/2, size/2
	radius := size/2 - 4
	innerRadius := 8 // Small black dot in the center
	
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := x - centerX
			dy := y - centerY
			distSq := dx*dx+dy*dy
			if distSq <= radius*radius {
				if distSq <= innerRadius*innerRadius {
					img.Set(x, y, color.Black)
				} else {
					img.Set(x, y, c)
				}
			}
		}
	}

	var buf bytes.Buffer
	png.Encode(&buf, img)
	return buf.Bytes()
}