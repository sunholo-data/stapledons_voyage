package render

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// AssetManager handles loading and caching of sprites and fonts
type AssetManager struct {
	sprites map[int]*ebiten.Image
}

// NewAssetManager creates a new asset manager
func NewAssetManager() *AssetManager {
	return &AssetManager{
		sprites: make(map[int]*ebiten.Image),
	}
}

// GetSprite returns the sprite image for the given ID
// TODO: Implement sprite loading from assets/
func (am *AssetManager) GetSprite(id int) *ebiten.Image {
	if img, ok := am.sprites[id]; ok {
		return img
	}
	// Placeholder: return nil for now
	return nil
}

// LoadSprite loads a sprite from the assets directory
func (am *AssetManager) LoadSprite(id int, path string) error {
	// TODO: Implement sprite loading
	return nil
}
