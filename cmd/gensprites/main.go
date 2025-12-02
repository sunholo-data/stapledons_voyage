// gensprites generates placeholder PNG sprites for testing.
// Creates isometric 64x32 diamond tiles and animated 32x48 entity sprite sheets.
package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
)

// Tile sprites: 64x32 isometric diamonds
var tileSprites = map[string]color.RGBA{
	"iso_tiles/water.png":    {0, 100, 200, 255},   // Blue
	"iso_tiles/forest.png":   {34, 139, 34, 255},   // Green
	"iso_tiles/desert.png":   {210, 180, 140, 255}, // Tan
	"iso_tiles/mountain.png": {139, 90, 43, 255},   // Brown
}

// Entity sprites: animated sprite sheets (4 frames × 32x48 = 128x48)
var entitySprites = map[string]color.RGBA{
	"iso_entities/npc_red.png":    {255, 100, 100, 255}, // Red
	"iso_entities/npc_green.png":  {100, 255, 100, 255}, // Green
	"iso_entities/npc_blue.png":   {100, 100, 255, 255}, // Blue
	"iso_entities/npc_yellow.png": {255, 255, 100, 255}, // Yellow
	"iso_entities/npc_purple.png": {200, 100, 255, 255}, // Purple
	"iso_entities/player.png":     {255, 215, 0, 255},   // Gold
}

// Star sprites by spectral type - used for galaxy map
// Each star sprite is a soft glowing circle with color gradient
var starSprites = map[string]color.RGBA{
	"stars/star_blue.png":   {155, 176, 255, 255}, // O/B type - hot blue
	"stars/star_white.png":  {255, 255, 255, 255}, // A/F type - white
	"stars/star_yellow.png": {255, 244, 214, 255}, // G type - yellow (like Sun)
	"stars/star_orange.png": {255, 210, 161, 255}, // K type - orange
	"stars/star_red.png":    {255, 189, 189, 255}, // M type - red dwarf
}

const (
	frameWidth  = 32
	frameHeight = 48
	frameCount  = 4 // 4-frame walk cycle
)

const starSize = 16 // Star sprite size (16x16 pixels)

func main() {
	outDir := "assets/sprites"
	if len(os.Args) > 1 {
		outDir = os.Args[1]
	}

	// Create subdirectories
	os.MkdirAll(filepath.Join(outDir, "iso_tiles"), 0755)
	os.MkdirAll(filepath.Join(outDir, "iso_entities"), 0755)
	os.MkdirAll(filepath.Join(outDir, "stars"), 0755)

	// Generate 64x32 isometric tile sprites (diamond shape)
	for name, col := range tileSprites {
		if err := generateIsoDiamond(filepath.Join(outDir, name), 64, 32, col); err != nil {
			fmt.Printf("Error generating %s: %v\n", name, err)
			continue
		}
		fmt.Printf("Generated %s (64x32 diamond)\n", name)
	}

	// Generate animated entity sprite sheets (4 frames × 32x48 = 128x48)
	for name, col := range entitySprites {
		if err := generateEntitySheet(filepath.Join(outDir, name), col); err != nil {
			fmt.Printf("Error generating %s: %v\n", name, err)
			continue
		}
		fmt.Printf("Generated %s (%dx%d sprite sheet, %d frames)\n", name, frameWidth*frameCount, frameHeight, frameCount)
	}

	// Generate star sprites (glowing circles for galaxy map)
	for name, col := range starSprites {
		if err := generateStarSprite(filepath.Join(outDir, name), starSize, col); err != nil {
			fmt.Printf("Error generating %s: %v\n", name, err)
			continue
		}
		fmt.Printf("Generated %s (%dx%d star)\n", name, starSize, starSize)
	}

	fmt.Println("Done!")
}

// generateIsoDiamond creates an isometric diamond-shaped tile sprite.
// The diamond has vertices at top-center, right-center, bottom-center, left-center.
func generateIsoDiamond(path string, width, height int, col color.RGBA) error {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Fill with transparent
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.Transparent)
		}
	}

	// Draw filled diamond
	// Diamond vertices: top=(w/2, 0), right=(w, h/2), bottom=(w/2, h), left=(0, h/2)
	centerX := float64(width) / 2
	centerY := float64(height) / 2

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Check if point is inside diamond using distance from center
			// Diamond equation: |x - centerX| / (width/2) + |y - centerY| / (height/2) <= 1
			dx := abs(float64(x) - centerX) / centerX
			dy := abs(float64(y) - centerY) / centerY

			if dx+dy <= 1.0 {
				img.Set(x, y, col)
			}
		}
	}

	// Add darker border
	border := color.RGBA{
		R: uint8(float64(col.R) * 0.6),
		G: uint8(float64(col.G) * 0.6),
		B: uint8(float64(col.B) * 0.6),
		A: 255,
	}

	// Draw border on diamond edges
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			dx := abs(float64(x) - centerX) / centerX
			dy := abs(float64(y) - centerY) / centerY
			dist := dx + dy

			// Edge detection: close to 1.0 but still inside
			if dist > 0.9 && dist <= 1.0 {
				img.Set(x, y, border)
			}
		}
	}

	// Add slight highlight on top half for 3D effect
	highlight := color.RGBA{
		R: min(255, uint8(float64(col.R)*1.2)),
		G: min(255, uint8(float64(col.G)*1.2)),
		B: min(255, uint8(float64(col.B)*1.2)),
		A: 255,
	}

	for y := 0; y < height/2; y++ {
		for x := 0; x < width; x++ {
			dx := abs(float64(x) - centerX) / centerX
			dy := abs(float64(y) - centerY) / centerY

			if dx+dy <= 0.7 {
				// Blend highlight with base color
				existing := img.RGBAAt(x, y)
				if existing.A > 0 {
					blended := blendColors(existing, highlight, 0.3)
					img.Set(x, y, blended)
				}
			}
		}
	}

	return saveImage(path, img)
}

// generateEntitySheet creates an animated sprite sheet with 4 walk cycle frames.
// Each frame is 32x48, total sheet is 128x48.
// Frames: idle, walk1, walk2, walk1 (symmetric walk cycle)
func generateEntitySheet(path string, col color.RGBA) error {
	sheetWidth := frameWidth * frameCount
	img := image.NewRGBA(image.Rect(0, 0, sheetWidth, frameHeight))

	// Fill with transparent
	for y := 0; y < frameHeight; y++ {
		for x := 0; x < sheetWidth; x++ {
			img.Set(x, y, color.Transparent)
		}
	}

	// Walk cycle offsets: vertical bob for walking animation
	// Frame 0: idle (neutral)
	// Frame 1: step left (bob down slightly, lean left)
	// Frame 2: neutral mid-step
	// Frame 3: step right (bob down slightly, lean right)
	bobOffsets := []int{0, 2, 0, 2}    // Vertical bob
	leanOffsets := []float64{0, -1, 0, 1} // Horizontal lean

	for frame := 0; frame < frameCount; frame++ {
		offsetX := frame * frameWidth
		bob := bobOffsets[frame]
		lean := leanOffsets[frame]

		drawEntityFrame(img, offsetX, bob, lean, col)
	}

	return saveImage(path, img)
}

// drawEntityFrame draws a single entity frame at the given x offset.
func drawEntityFrame(img *image.RGBA, offsetX, bobY int, leanX float64, col color.RGBA) {
	centerX := float64(offsetX) + float64(frameWidth)/2 + leanX

	// Draw body (lower oval, taking bottom 2/3 of sprite)
	bodyTop := frameHeight/3 + bobY
	bodyCenterY := float64(bodyTop+frameHeight) / 2
	bodyRadiusX := float64(frameWidth) / 2.5
	bodyRadiusY := float64(frameHeight-bodyTop) / 2.2

	for y := bodyTop; y < frameHeight; y++ {
		for x := offsetX; x < offsetX+frameWidth; x++ {
			dx := (float64(x) - centerX) / bodyRadiusX
			dy := (float64(y) - bodyCenterY) / bodyRadiusY
			if dx*dx+dy*dy <= 1.0 {
				img.Set(x, y, col)
			}
		}
	}

	// Draw head (upper circle, in top 1/3)
	headCenterY := float64(frameHeight)/5 + float64(bobY)
	headRadius := float64(frameWidth) / 4

	for y := 0; y < bodyTop; y++ {
		for x := offsetX; x < offsetX+frameWidth; x++ {
			dx := float64(x) - centerX
			dy := float64(y) - headCenterY
			if dx*dx+dy*dy <= headRadius*headRadius {
				img.Set(x, y, col)
			}
		}
	}

	// Add darker outline
	outline := color.RGBA{
		R: uint8(float64(col.R) * 0.5),
		G: uint8(float64(col.G) * 0.5),
		B: uint8(float64(col.B) * 0.5),
		A: 255,
	}

	// Simple edge detection for outline
	for y := 1; y < frameHeight-1; y++ {
		for x := offsetX + 1; x < offsetX+frameWidth-1; x++ {
			current := img.RGBAAt(x, y)
			if current.A > 0 {
				// Check if any neighbor is transparent
				neighbors := []image.Point{
					{x - 1, y}, {x + 1, y},
					{x, y - 1}, {x, y + 1},
				}
				for _, n := range neighbors {
					if n.X >= offsetX && n.X < offsetX+frameWidth {
						if img.RGBAAt(n.X, n.Y).A == 0 {
							img.Set(x, y, outline)
							break
						}
					}
				}
			}
		}
	}

	// Add highlight to head
	highlight := color.RGBA{
		R: min(255, uint8(float64(col.R)*1.3)),
		G: min(255, uint8(float64(col.G)*1.3)),
		B: min(255, uint8(float64(col.B)*1.3)),
		A: 255,
	}

	// Small highlight spot on head
	hlX := centerX - headRadius/3
	hlY := headCenterY - headRadius/3
	hlRadius := headRadius / 3

	for y := 0; y < frameHeight/3; y++ {
		for x := offsetX; x < offsetX+frameWidth; x++ {
			dx := float64(x) - hlX
			dy := float64(y) - hlY
			if dx*dx+dy*dy <= hlRadius*hlRadius {
				existing := img.RGBAAt(x, y)
				if existing.A > 0 && existing != outline {
					img.Set(x, y, highlight)
				}
			}
		}
	}
}

// generateStarSprite creates a soft glowing star sprite.
// Uses radial gradient from bright center to transparent edge.
func generateStarSprite(path string, size int, col color.RGBA) error {
	img := image.NewRGBA(image.Rect(0, 0, size, size))

	centerX := float64(size) / 2
	centerY := float64(size) / 2
	maxRadius := float64(size) / 2

	// Create radial gradient
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := float64(x) - centerX + 0.5
			dy := float64(y) - centerY + 0.5
			dist := (dx*dx + dy*dy) / (maxRadius * maxRadius)

			if dist > 1.0 {
				img.Set(x, y, color.Transparent)
				continue
			}

			// Smooth falloff using squared cosine for soft glow
			// intensity = cos²(dist * π/2) gives nice soft edge
			intensity := 1.0 - dist
			intensity = intensity * intensity // Quadratic falloff for soft glow

			// Bright core (inner 30% is near full brightness)
			if dist < 0.09 { // sqrt(0.09) = 0.3
				intensity = 1.0
			}

			// Apply intensity to color
			alpha := uint8(255 * intensity)
			if alpha < 2 {
				img.Set(x, y, color.Transparent)
				continue
			}

			// Blend towards white in center for "hot" look
			blendToWhite := 0.0
			if dist < 0.25 {
				blendToWhite = (0.25 - dist) / 0.25 * 0.5
			}

			r := uint8(float64(col.R) + (255-float64(col.R))*blendToWhite)
			g := uint8(float64(col.G) + (255-float64(col.G))*blendToWhite)
			b := uint8(float64(col.B) + (255-float64(col.B))*blendToWhite)

			img.Set(x, y, color.RGBA{r, g, b, alpha})
		}
	}

	return saveImage(path, img)
}

func saveImage(path string, img *image.RGBA) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	return png.Encode(file, img)
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func min(a, b uint8) uint8 {
	if a < b {
		return a
	}
	return b
}

func blendColors(base, overlay color.RGBA, alpha float64) color.RGBA {
	return color.RGBA{
		R: uint8(float64(base.R)*(1-alpha) + float64(overlay.R)*alpha),
		G: uint8(float64(base.G)*(1-alpha) + float64(overlay.G)*alpha),
		B: uint8(float64(base.B)*(1-alpha) + float64(overlay.B)*alpha),
		A: 255,
	}
}
