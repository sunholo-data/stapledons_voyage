// gensprites generates placeholder PNG sprites for testing.
package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
)

var biomeSprites = map[string]color.RGBA{
	"tile_water.png":    {0, 100, 200, 255},   // Blue
	"tile_forest.png":   {34, 139, 34, 255},   // Green
	"tile_desert.png":   {210, 180, 140, 255}, // Tan
	"tile_mountain.png": {139, 90, 43, 255},   // Brown
}

var entitySprites = map[string]color.RGBA{
	"player.png": {255, 215, 0, 255},  // Gold
	"npc.png":    {147, 112, 219, 255}, // Purple
}

func main() {
	outDir := "assets/sprites"
	if len(os.Args) > 1 {
		outDir = os.Args[1]
	}

	// Generate 16x16 biome tiles
	for name, col := range biomeSprites {
		if err := generateSprite(filepath.Join(outDir, name), 16, 16, col); err != nil {
			fmt.Printf("Error generating %s: %v\n", name, err)
			continue
		}
		fmt.Printf("Generated %s\n", name)
	}

	// Generate 32x32 entity sprites
	for name, col := range entitySprites {
		if err := generateSprite(filepath.Join(outDir, name), 32, 32, col); err != nil {
			fmt.Printf("Error generating %s: %v\n", name, err)
			continue
		}
		fmt.Printf("Generated %s\n", name)
	}

	fmt.Println("Done!")
}

func generateSprite(path string, width, height int, col color.RGBA) error {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Fill with solid color
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, col)
		}
	}

	// Add a simple border (darker version of color)
	border := color.RGBA{
		R: uint8(float64(col.R) * 0.7),
		G: uint8(float64(col.G) * 0.7),
		B: uint8(float64(col.B) * 0.7),
		A: 255,
	}
	for x := 0; x < width; x++ {
		img.Set(x, 0, border)
		img.Set(x, height-1, border)
	}
	for y := 0; y < height; y++ {
		img.Set(0, y, border)
		img.Set(width-1, y, border)
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	return png.Encode(file, img)
}
