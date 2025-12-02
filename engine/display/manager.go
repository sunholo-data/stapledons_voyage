package display

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
	// Internal resolution (game logic coordinates)
	// Higher resolution for sharper rendering on Retina/HiDPI displays
	InternalWidth  = 1280
	InternalHeight = 960
)

// Manager handles display settings and window management.
type Manager struct {
	config     Config
	configPath string
}

// NewManager creates a display manager, loading config from the given path.
func NewManager(configPath string) *Manager {
	cfg := LoadConfig(configPath)

	m := &Manager{
		config:     cfg,
		configPath: configPath,
	}

	// Apply initial settings
	m.applySettings()

	return m
}

// applySettings applies the current config to Ebiten.
func (m *Manager) applySettings() {
	ebiten.SetWindowSize(m.config.Width, m.config.Height)
	ebiten.SetFullscreen(m.config.Fullscreen)
	ebiten.SetVsyncEnabled(m.config.VSync)
}

// Layout returns the internal (logical) screen dimensions.
// Ebiten handles scaling to the window size.
func (m *Manager) Layout(outsideWidth, outsideHeight int) (int, int) {
	return InternalWidth, InternalHeight
}

// ToggleFullscreen toggles fullscreen mode and saves the setting.
func (m *Manager) ToggleFullscreen() {
	m.config.Fullscreen = !m.config.Fullscreen
	ebiten.SetFullscreen(m.config.Fullscreen)
	m.Save()
}

// SetResolution changes the window resolution (non-fullscreen).
func (m *Manager) SetResolution(width, height int) {
	m.config.Width = width
	m.config.Height = height
	if !m.config.Fullscreen {
		ebiten.SetWindowSize(width, height)
	}
	m.Save()
}

// IsFullscreen returns the current fullscreen state.
func (m *Manager) IsFullscreen() bool {
	return m.config.Fullscreen
}

// Config returns a copy of the current configuration.
func (m *Manager) Config() Config {
	return m.config
}

// Save persists the current configuration to disk.
func (m *Manager) Save() error {
	return SaveConfig(m.configPath, m.config)
}

// HandleInput checks for display-related input (F11 for fullscreen).
// Call this in the game's Update() method.
func (m *Manager) HandleInput() {
	if inpututil.IsKeyJustPressed(ebiten.KeyF11) {
		m.ToggleFullscreen()
	}
}
