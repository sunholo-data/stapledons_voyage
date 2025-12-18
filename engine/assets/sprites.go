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
	sprites      map[int]*ebiten.Image
	maskedTiles  map[int]*ebiten.Image  // Cached diamond-masked versions
	animations   map[int]*SpriteAnimDef // Animation definitions per sprite ID
	placeholder  *ebiten.Image
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
		maskedTiles: make(map[int]*ebiten.Image),
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

// MaskToDiamond takes a rectangular tile and returns a new image with
// pixels outside the isometric diamond shape made transparent.
// Uses pixel-perfect row-by-row calculation for exact tessellation.
// For a 64x32 tile, each row's visible pixels are calculated as:
// - Row 0: center 2 pixels, Row 1: center 4 pixels, ... Row 15: center 32 pixels
// - Row 16: center 32 pixels, Row 17: center 30 pixels, ... Row 31: center 2 pixels
func MaskToDiamond(src *ebiten.Image) *ebiten.Image {
	bounds := src.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()

	// Create a new RGBA image
	dst := image.NewRGBA(image.Rect(0, 0, w, h))

	halfH := h / 2 // 16 for 32-height tile

	for y := 0; y < h; y++ {
		// Calculate how wide the diamond is at this row
		// Diamond width increases by (w/h) pixels per row from top until middle,
		// then decreases symmetrically
		var rowsFromEdge int
		if y < halfH {
			rowsFromEdge = y + 1 // rows 0-15: 1,2,3...16
		} else {
			rowsFromEdge = h - y // rows 16-31: 16,15,14...1
		}

		// Width at this row: 2 * rowsFromEdge * (w/h) = 2 * rowsFromEdge * 2 = 4 * rowsFromEdge for 64x32
		// But we want to match the exact diamond shape
		// For 64x32: at row 0, width=2; at row 15/16, width=64
		pixelsWide := rowsFromEdge * w / halfH // rowsFromEdge * 4 for 64x32

		// Center the visible pixels
		margin := (w - pixelsWide) / 2
		xStart := margin
		xEnd := w - margin

		for x := 0; x < w; x++ {
			if x >= xStart && x < xEnd {
				c := src.At(x, y)
				dst.Set(x, y, c)
			}
		}
	}

	return ebiten.NewImageFromImage(dst)
}

// GetMaskedTile returns a diamond-masked version of the tile.
// Results are cached for efficiency.
func (sm *SpriteManager) GetMaskedTile(id int) *ebiten.Image {
	// Return cached version if available
	if masked, ok := sm.maskedTiles[id]; ok {
		return masked
	}

	// Get original sprite
	sprite, ok := sm.sprites[id]
	if !ok {
		return sm.placeholder
	}

	// Create masked version and cache it
	masked := MaskToDiamond(sprite)
	sm.maskedTiles[id] = masked

	return masked
}
