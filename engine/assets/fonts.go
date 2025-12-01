// Package assets provides font loading and management.
package assets

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
)

// FontManager handles font loading and caching.
type FontManager struct {
	fonts        map[string]font.Face
	defaultFace  font.Face
	fallbackFace font.Face // Embedded fallback font
	basePath     string
}

// NewFontManager creates a new font manager with an embedded fallback font.
func NewFontManager() *FontManager {
	fm := &FontManager{
		fonts: make(map[string]font.Face),
	}

	// Create embedded fallback font from Go Regular
	if tt, err := opentype.Parse(goregular.TTF); err == nil {
		if face, err := opentype.NewFace(tt, &opentype.FaceOptions{
			Size:    14,
			DPI:     72,
			Hinting: font.HintingFull,
		}); err == nil {
			fm.fallbackFace = face
			fm.defaultFace = face // Use as default until a custom font is loaded
		}
	}

	return fm
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
	for name, entry := range manifest.Fonts {
		face, err := fm.loadFont(filepath.Join(fontPath, entry.File), entry.Size)
		if err != nil {
			fmt.Printf("Warning: failed to load font %s (%s): %v\n", name, entry.File, err)
			continue
		}
		fm.fonts[name] = face

		// Set default font
		if entry.Default || fm.defaultFace == nil {
			fm.defaultFace = face
		}
	}

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
