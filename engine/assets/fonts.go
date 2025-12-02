// Package assets provides font loading and management.
package assets

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gomono"
	"golang.org/x/image/font/opentype"
)

// FontSize constants for GetBySize
const (
	FontSizeSmall  = 0 // labels
	FontSizeNormal = 1 // body text
	FontSizeLarge  = 2 // headers
	FontSizeTitle  = 3 // screen titles
)

// Reference resolution for font scaling (720p baseline)
const referenceHeight = 720.0

// baseFontSizes are the point sizes at reference resolution (720p)
// These scale up for higher resolutions
var baseFontSizes = []float64{16, 22, 28, 38}

// fontSizePoints holds the actual scaled point sizes
var fontSizePoints = []float64{16, 22, 28, 38}

// FontManager handles font loading and caching.
type FontManager struct {
	fonts        map[string]font.Face
	defaultFace  font.Face
	fallbackFace font.Face      // Embedded fallback font
	basePath     string
	sizedFaces   [4]font.Face   // Cached faces at standard sizes
	parsedFont   *opentype.Font // Parsed font for creating sized faces
	scale        float64        // Scale factor based on screen resolution
}

// NewFontManager creates a new font manager with an embedded fallback font.
func NewFontManager() *FontManager {
	fm := &FontManager{
		fonts: make(map[string]font.Face),
		scale: 1.0,
	}

	// Create embedded fallback font from Go Mono (clean monospace for sci-fi aesthetic)
	if tt, err := opentype.Parse(gomono.TTF); err == nil {
		fm.parsedFont = tt
		// Create default face at normal size
		if face, err := opentype.NewFace(tt, &opentype.FaceOptions{
			Size:    baseFontSizes[FontSizeNormal],
			DPI:     72,
			Hinting: font.HintingFull,
		}); err == nil {
			fm.fallbackFace = face
			fm.defaultFace = face // Use as default until a custom font is loaded
		}
		// Create sized faces for GetBySize
		fm.initSizedFaces(tt)
	}

	return fm
}

// NewFontManagerWithScale creates a font manager with resolution-based scaling.
func NewFontManagerWithScale(screenHeight int) *FontManager {
	fm := NewFontManager()
	fm.SetScale(screenHeight)
	return fm
}

// SetScale adjusts font sizes based on screen height relative to reference resolution.
func (fm *FontManager) SetScale(screenHeight int) {
	if screenHeight <= 0 {
		screenHeight = int(referenceHeight)
	}
	fm.scale = float64(screenHeight) / referenceHeight
	if fm.scale < 0.5 {
		fm.scale = 0.5
	}
	if fm.scale > 3.0 {
		fm.scale = 3.0
	}

	// Update fontSizePoints with scaled values
	for i, base := range baseFontSizes {
		fontSizePoints[i] = base * fm.scale
	}

	// Recreate sized faces with new scale
	if fm.parsedFont != nil {
		fm.initSizedFaces(fm.parsedFont)
	}
}

// initSizedFaces creates font faces at all standard sizes.
func (fm *FontManager) initSizedFaces(tt *opentype.Font) {
	for i, pts := range fontSizePoints {
		if face, err := opentype.NewFace(tt, &opentype.FaceOptions{
			Size:    pts,
			DPI:     72,
			Hinting: font.HintingFull,
		}); err == nil {
			fm.sizedFaces[i] = face
		}
	}
}

// LoadManifest loads fonts defined in the manifest.json file.
func (fm *FontManager) LoadManifest(fontPath string) error {
	fm.basePath = fontPath
	manifestPath := filepath.Join(fontPath, "manifest.json")

	var manifest FontManifest
	if err := loadJSON(manifestPath, &manifest); err != nil {
		return fmt.Errorf("loading font manifest: %w", err)
	}

	// Load each font
	var defaultFontPath string
	for name, entry := range manifest.Fonts {
		face, err := fm.loadFont(filepath.Join(fontPath, entry.File), entry.Size)
		if err != nil {
			fmt.Printf("Warning: failed to load font %s (%s): %v\n", name, entry.File, err)
			continue
		}
		fm.fonts[name] = face

		// Set default font and remember its path for sized faces
		if entry.Default || fm.defaultFace == nil {
			fm.defaultFace = face
			defaultFontPath = filepath.Join(fontPath, entry.File)
		}
	}

	// Load the default font for sized faces (replaces embedded font)
	if defaultFontPath != "" {
		if err := fm.loadFontForSizes(defaultFontPath); err != nil {
			fmt.Printf("Warning: using fallback font for sizes: %v\n", err)
		}
	}

	return nil
}

// loadFontForSizes loads a font file and creates sized faces from it.
func (fm *FontManager) loadFontForSizes(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading font file: %w", err)
	}

	tt, err := opentype.Parse(data)
	if err != nil {
		return fmt.Errorf("parsing font: %w", err)
	}

	// Store the parsed font and recreate sized faces
	fm.parsedFont = tt
	fm.initSizedFaces(tt)

	return nil
}

// loadFont loads a TTF/OTF font file at the specified size.
func (fm *FontManager) loadFont(path string, size float64) (font.Face, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading font file: %w", err)
	}

	// Parse the font
	tt, err := opentype.Parse(data)
	if err != nil {
		return nil, fmt.Errorf("parsing font: %w", err)
	}

	// Create a face at the specified size
	face, err := opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    size,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		return nil, fmt.Errorf("creating font face: %w", err)
	}

	return face, nil
}

// Get returns a font face by name, or the default/fallback if not found.
func (fm *FontManager) Get(name string) font.Face {
	if face, ok := fm.fonts[name]; ok {
		return face
	}
	if fm.defaultFace != nil {
		return fm.defaultFace
	}
	return fm.fallbackFace
}

// GetDefault returns the default font face, or fallback if no default set.
func (fm *FontManager) GetDefault() font.Face {
	if fm.defaultFace != nil {
		return fm.defaultFace
	}
	return fm.fallbackFace
}

// GetBySize returns a font face at the specified size index.
// Size: 0=small(10pt), 1=normal(14pt), 2=large(18pt), 3=title(24pt)
func (fm *FontManager) GetBySize(size int) font.Face {
	if size < 0 || size >= len(fm.sizedFaces) {
		size = FontSizeNormal // Default to normal
	}
	if face := fm.sizedFaces[size]; face != nil {
		return face
	}
	return fm.fallbackFace
}

// GetSizePoints returns the point size for a given size index.
func (fm *FontManager) GetSizePoints(size int) float64 {
	if size < 0 || size >= len(fontSizePoints) {
		return fontSizePoints[FontSizeNormal]
	}
	return fontSizePoints[size]
}

// Has returns true if a font with the given name is loaded.
func (fm *FontManager) Has(name string) bool {
	_, ok := fm.fonts[name]
	return ok
}

// FontCount returns the number of loaded fonts.
func (fm *FontManager) FontCount() int {
	return len(fm.fonts)
}

// FontManifest represents the fonts/manifest.json structure.
type FontManifest struct {
	Fonts map[string]FontEntry `json:"fonts"`
}

// FontEntry defines a single font in the manifest.
type FontEntry struct {
	File    string  `json:"file"`
	Size    float64 `json:"size"`
	Default bool    `json:"default,omitempty"`
}
