// Package assets handles loading and caching of game assets (sprites, fonts, sounds).
package assets

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
	"golang.org/x/image/font"
)

// Manager handles loading and caching of all game assets.
type Manager struct {
	basePath string
	sprites  *SpriteManager
	audio    *AudioManager
	fonts    *FontManager
}

// NewManager creates an asset manager rooted at the given base path.
func NewManager(basePath string) (*Manager, error) {
	m := &Manager{
		basePath: basePath,
		sprites:  NewSpriteManager(),
		audio:    NewAudioManager(),
		fonts:    NewFontManager(),
	}

	// Load sprite manifest
	spritePath := filepath.Join(basePath, "sprites")
	if err := m.sprites.LoadManifest(spritePath); err != nil {
		// Log warning but don't fail - assets are optional
		fmt.Printf("Warning: failed to load sprite manifest: %v\n", err)
	}

	// Load audio manifest
	soundPath := filepath.Join(basePath, "sounds")
	if err := m.audio.LoadManifest(soundPath); err != nil {
		// Log warning but don't fail - audio is optional
		fmt.Printf("Warning: failed to load audio manifest: %v\n", err)
	}

	// Load font manifest
	fontPath := filepath.Join(basePath, "fonts")
	if err := m.fonts.LoadManifest(fontPath); err != nil {
		// Log warning but don't fail - fonts are optional (falls back to debug font)
		fmt.Printf("Warning: failed to load font manifest: %v\n", err)
	}

	return m, nil
}

// GetSprite returns a sprite by ID, or a placeholder if not found.
func (m *Manager) GetSprite(id int) *ebiten.Image {
	return m.sprites.Get(id)
}

// PlaySound plays a sound effect by ID.
func (m *Manager) PlaySound(id int) {
	m.audio.PlaySound(id)
}

// PlaySounds plays multiple sounds from FrameOutput.Sounds.
func (m *Manager) PlaySounds(soundIDs []int) {
	m.audio.PlaySounds(soundIDs)
}

// Audio returns the audio manager for advanced audio control.
func (m *Manager) Audio() *AudioManager {
	return m.audio
}

// Fonts returns the font manager.
func (m *Manager) Fonts() *FontManager {
	return m.fonts
}

// Sprites returns the sprite manager.
func (m *Manager) Sprites() *SpriteManager {
	return m.sprites
}

// GetFont returns a font face by name, or the default if not found.
func (m *Manager) GetFont(name string) font.Face {
	return m.fonts.Get(name)
}

// GetDefaultFont returns the default font face, or nil if no fonts loaded.
func (m *Manager) GetDefaultFont() font.Face {
	return m.fonts.GetDefault()
}

// GetFontBySize returns a font face at the specified size index.
// Size: 0=small(12pt), 1=normal(16pt), 2=large(22pt), 3=title(30pt)
// Actual point sizes are scaled based on screen resolution.
func (m *Manager) GetFontBySize(size int) font.Face {
	return m.fonts.GetBySize(size)
}

// SetFontScale adjusts font sizes based on screen height.
// Call this after creating the manager to scale fonts for the target resolution.
func (m *Manager) SetFontScale(screenHeight int) {
	m.fonts.SetScale(screenHeight)
}

// SpriteManifest represents the sprites/manifest.json structure.
type SpriteManifest struct {
	Sprites map[string]SpriteEntry `json:"sprites"`
}

// SpriteEntry defines a single sprite in the manifest.
type SpriteEntry struct {
	File        string                   `json:"file"`
	Width       int                      `json:"width"`
	Height      int                      `json:"height"`
	Type        string                   `json:"type,omitempty"`        // "tile" or "entity"
	Animations  map[string]SpriteAnimSeq `json:"animations,omitempty"`  // Animation sequences
	FrameWidth  int                      `json:"frameWidth,omitempty"`  // Width of each frame
	FrameHeight int                      `json:"frameHeight,omitempty"` // Height of each frame
}

// SpriteAnimSeq is re-exported from sprites.go for manifest parsing.
// Already defined in sprites.go, but we need it here for JSON parsing.

// loadJSON loads and unmarshals a JSON file.
func loadJSON(path string, v interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}
