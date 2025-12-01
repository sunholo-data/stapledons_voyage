package assets

import (
	"fmt"
	"image"
	"image/color"
	_ "image/png" // Register PNG decoder
	"os"
	"path/filepath"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
)

// SpriteManager handles sprite loading and caching.
type SpriteManager struct {
	sprites     map[int]*ebiten.Image
	animations  map[int]*SpriteAnimDef // Animation definitions per sprite ID
	placeholder *ebiten.Image
}

// SpriteAnimDef defines animation sequences for a sprite.
type SpriteAnimDef struct {
	Animations  map[string]SpriteAnimSeq `json:"animations"`
	FrameWidth  int                      `json:"frameWidth"`
	FrameHeight int                      `json:"frameHeight"`
}

// SpriteAnimSeq defines a single animation sequence.
type SpriteAnimSeq struct {
	StartFrame int     `json:"startFrame"`
	FrameCount int     `json:"frameCount"`
	FPS        float64 `json:"fps"`
}

// NewSpriteManager creates a new sprite manager with an empty cache.
func NewSpriteManager() *SpriteManager {
	// Create a 16x16 magenta placeholder for missing sprites
	placeholder := ebiten.NewImage(16, 16)
	placeholder.Fill(color.RGBA{255, 0, 255, 255})

	return &SpriteManager{
		sprites:     make(map[int]*ebiten.Image),
		animations:  make(map[int]*SpriteAnimDef),
		placeholder: placeholder,
	}
}

// LoadManifest loads sprites defined in the manifest.json file.
func (sm *SpriteManager) LoadManifest(spritePath string) error {
	manifestPath := filepath.Join(spritePath, "manifest.json")

	var manifest SpriteManifest
	if err := loadJSON(manifestPath, &manifest); err != nil {
		return fmt.Errorf("loading manifest: %w", err)
	}

	for idStr, entry := range manifest.Sprites {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			fmt.Printf("Warning: invalid sprite ID %q, skipping\n", idStr)
			continue
		}

		imgPath := filepath.Join(spritePath, entry.File)
		img, err := loadImage(imgPath)
		if err != nil {
			fmt.Printf("Warning: failed to load sprite %d (%s): %v\n", id, entry.File, err)
			continue
		}

		sm.sprites[id] = img

		// Store animation definitions if present
		if len(entry.Animations) > 0 && entry.FrameWidth > 0 && entry.FrameHeight > 0 {
			sm.animations[id] = &SpriteAnimDef{
				Animations:  entry.Animations,
				FrameWidth:  entry.FrameWidth,
				FrameHeight: entry.FrameHeight,
			}
		}
	}

	return nil
}

// Get returns a sprite by ID, or the placeholder if not found.
func (sm *SpriteManager) Get(id int) *ebiten.Image {
	if sprite, ok := sm.sprites[id]; ok {
		return sprite
	}
	return sm.placeholder
}

// Has returns true if a sprite with the given ID is loaded.
func (sm *SpriteManager) Has(id int) bool {
	_, ok := sm.sprites[id]
	return ok
}

// GetAnimation returns the animation definition for a sprite, or nil if not animated.
func (sm *SpriteManager) GetAnimation(id int) *SpriteAnimDef {
	return sm.animations[id]
}

// HasAnimation returns true if the sprite has animation definitions.
func (sm *SpriteManager) HasAnimation(id int) bool {
	_, ok := sm.animations[id]
	return ok
}

// loadImage loads an image file and converts it to an Ebiten image.
func loadImage(path string) (*ebiten.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	return ebiten.NewImageFromImage(img), nil
}
