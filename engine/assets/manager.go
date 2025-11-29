// Package assets handles loading and caching of game assets (sprites, fonts, sounds).
package assets

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
)

// Manager handles loading and caching of all game assets.
type Manager struct {
	basePath string
	sprites  *SpriteManager
}

// NewManager creates an asset manager rooted at the given base path.
func NewManager(basePath string) (*Manager, error) {
	m := &Manager{
		basePath: basePath,
		sprites:  NewSpriteManager(),
	}

	// Load sprite manifest
	spritePath := filepath.Join(basePath, "sprites")
	if err := m.sprites.LoadManifest(spritePath); err != nil {
		// Log warning but don't fail - assets are optional
		fmt.Printf("Warning: failed to load sprite manifest: %v\n", err)
	}

	return m, nil
}

// GetSprite returns a sprite by ID, or a placeholder if not found.
func (m *Manager) GetSprite(id int) *ebiten.Image {
	return m.sprites.Get(id)
}

// SpriteManifest represents the sprites/manifest.json structure.
type SpriteManifest struct {
	Sprites map[string]SpriteEntry `json:"sprites"`
}

// SpriteEntry defines a single sprite in the manifest.
type SpriteEntry struct {
	File   string `json:"file"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

// loadJSON loads and unmarshals a JSON file.
func loadJSON(path string, v interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}
