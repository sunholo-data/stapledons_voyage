// Package texgen provides dimension-aware texture generation for interiors.
// Uses AILANG TextureSpec to build prompts and the AI handler to generate textures.
package texgen

import (
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"stapledons_voyage/engine/handlers"
	"stapledons_voyage/sim_gen"
)

// Generator handles texture generation from TextureSpec.
type Generator struct {
	aiHandler AIHandler
	cacheDir  string
	cache     *Cache // Optional cache for generated textures
}

// AIHandler defines the interface for AI operations.
// This allows swapping between real Gemini and stub handlers.
type AIHandler interface {
	Call(input string) (string, error)
}

// NewGenerator creates a new texture generator.
func NewGenerator(aiHandler AIHandler, cacheDir string) *Generator {
	return &Generator{
		aiHandler: aiHandler,
		cacheDir:  cacheDir,
	}
}

// SetCache sets the texture cache for the generator.
func (g *Generator) SetCache(cache *Cache) {
	g.cache = cache
}

// Generate creates a texture from a TextureSpec.
// Returns the path to the generated texture file.
func (g *Generator) Generate(spec *sim_gen.TextureSpec) (string, error) {
	// Get cache key from AILANG
	cacheKey := sim_gen.SpecCacheKey(spec)

	// Check cache first
	if g.cache != nil {
		if path, ok := g.cache.Get(cacheKey); ok {
			log.Printf("[texgen] Cache hit: %s", cacheKey)
			return path, nil
		}
	}

	log.Printf("[texgen] Cache miss, generating: %s", cacheKey)

	// Build prompt from AILANG
	prompt := sim_gen.BuildPrompt(spec)

	// Create AI request
	req := handlers.AIRequest{
		Messages: []handlers.ContentBlock{
			{Type: handlers.ContentTypeText, Text: "generate image: " + prompt},
		},
		Context: map[string]interface{}{
			"generate_image": true,
		},
	}

	reqJSON, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("encoding AI request: %w", err)
	}

	// Call AI handler
	respJSON, err := g.aiHandler.Call(string(reqJSON))
	if err != nil {
		return "", fmt.Errorf("AI generation failed: %w", err)
	}

	// Parse response
	var resp handlers.AIResponse
	if err := json.Unmarshal([]byte(respJSON), &resp); err != nil {
		return "", fmt.Errorf("parsing AI response: %w", err)
	}

	// Find generated image
	for _, block := range resp.Content {
		if block.Type == handlers.ContentTypeImage && block.ImageRef != "" {
			// Store in cache
			if g.cache != nil {
				g.cache.Put(cacheKey, block.ImageRef)
			}
			return block.ImageRef, nil
		}
	}

	return "", fmt.Errorf("no image in AI response")
}

// GenerateToImage generates a texture and loads it as an Ebiten image.
func (g *Generator) GenerateToImage(spec *sim_gen.TextureSpec) (*ebiten.Image, error) {
	path, err := g.Generate(spec)
	if err != nil {
		return nil, err
	}

	return LoadTextureFromPath(path)
}

// LoadTextureFromPath loads an image file as an Ebiten image.
func LoadTextureFromPath(path string) (*ebiten.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening texture %s: %w", path, err)
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("decoding texture %s: %w", path, err)
	}

	return ebiten.NewImageFromImage(img), nil
}

// GenerateRoomTextures generates all textures for a room.
// Returns a map of surface type to texture path.
func (g *Generator) GenerateRoomTextures(roomID string, width, depth, height float64, theme *sim_gen.RoomTheme, ppm int64, seed int64) (map[string]string, error) {
	// Build specs using AILANG
	specs := sim_gen.BuildRoomSpecs(
		roomID,
		width, depth, height,
		theme,
		ppm, seed,
	)

	result := make(map[string]string)
	for _, spec := range specs {
		path, err := g.Generate(spec)
		if err != nil {
			log.Printf("[texgen] Warning: failed to generate %s: %v", spec.SurfaceType, err)
			continue
		}
		result[spec.SurfaceType] = path
	}

	return result, nil
}

// GenerateRoomTexturesAsync generates all textures for a room asynchronously.
// Returns immediately with a channel that receives results as they complete.
func (g *Generator) GenerateRoomTexturesAsync(roomID string, width, depth, height float64, theme *sim_gen.RoomTheme, ppm int64, seed int64) <-chan TextureResult {
	results := make(chan TextureResult, 6) // 6 surfaces per room

	go func() {
		defer close(results)

		specs := sim_gen.BuildRoomSpecs(
			roomID,
			width, depth, height,
			theme,
			ppm, seed,
		)

		for _, spec := range specs {
			path, err := g.Generate(spec)
			results <- TextureResult{
				SurfaceType: spec.SurfaceType,
				Path:        path,
				Error:       err,
			}
		}
	}()

	return results
}

// TextureResult represents the result of a texture generation.
type TextureResult struct {
	SurfaceType string
	Path        string
	Error       error
}
