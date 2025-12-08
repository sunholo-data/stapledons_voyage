// cmd/demo-tetra/main.go
// Demo command for testing Tetra3D 3D rendering.
// Usage:
//   go run ./cmd/demo-tetra                     # Basic sphere demo
//   go run ./cmd/demo-tetra --screenshot 60     # Take screenshot after N frames
package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"stapledons_voyage/engine/display"
	"stapledons_voyage/engine/tetra"
)

var (
	screenshotFrame = flag.Int("screenshot", 0, "Take screenshot after N frames (0=disabled)")
	outputPath      = flag.String("output", "out/demo-tetra.png", "Screenshot output path")
)

type DemoGame struct {
	scene   *tetra.Scene
	planet  *tetra.Planet
	sun     *tetra.SunLight
	ambient *tetra.AmbientLight

	// Animation state
	time       float64
	frameCount int

	// Screenshot state
	screenshotTaken bool
}

func NewDemoGame() *DemoGame {
	g := &DemoGame{}

	// Create 3D scene
	g.scene = tetra.NewScene(display.InternalWidth, display.InternalHeight)

	// Add lighting
	g.sun = tetra.NewSunLight()
	g.sun.SetPosition(5, 3, 5) // Upper-right-front
	g.sun.AddToScene(g.scene)

	g.ambient = tetra.NewAmbientLight(0.2, 0.2, 0.3, 0.5) // Dim blue ambient
	g.ambient.AddToScene(g.scene)

	// Create a blue planet (Earth-like)
	g.planet = tetra.NewPlanet("earth", 1.0, color.RGBA{60, 120, 200, 255})
	g.planet.AddToScene(g.scene)
	g.planet.SetPosition(0, 0, 0)
	g.planet.SetRotationSpeed(0.3) // Slow rotation

	// Position camera to see the planet
	g.scene.SetCameraPosition(0, 0, 4)

	return g
}

func (g *DemoGame) Update() error {
	dt := 1.0 / 60.0
	g.time += dt
	g.frameCount++

	// Update planet rotation
	g.planet.Update(dt)

	return nil
}

func (g *DemoGame) Draw(screen *ebiten.Image) {
	// Clear with black
	screen.Fill(color.Black)

	// Render 3D scene
	img3d := g.scene.Render()
	screen.DrawImage(img3d, nil)

	// Draw HUD
	g.drawHUD(screen)

	// Take screenshot if requested
	if *screenshotFrame > 0 && g.frameCount >= *screenshotFrame && !g.screenshotTaken {
		g.takeScreenshot(screen)
		g.screenshotTaken = true
	}
}

func (g *DemoGame) drawHUD(screen *ebiten.Image) {
	y := 20.0
	lineHeight := 18.0

	ebitenutil.DebugPrintAt(screen, "Tetra3D Demo", 10, int(y))
	y += lineHeight

	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Frame: %d", g.frameCount), 10, int(y))
	y += lineHeight

	fps := ebiten.ActualFPS()
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("FPS: %.1f", fps), 10, int(y))
	y += lineHeight

	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Rotation: %.2f rad", g.planet.Rotation()), 10, int(y))
	y += lineHeight

	// Help at bottom
	y = float64(display.InternalHeight) - 40
	ebitenutil.DebugPrintAt(screen, "Tetra3D Integration Test", 10, int(y))
	y += lineHeight
	if *screenshotFrame > 0 {
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Screenshot at frame %d", *screenshotFrame), 10, int(y))
	}
}

func (g *DemoGame) takeScreenshot(screen *ebiten.Image) {
	// Create output directory if needed
	dir := "out"
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Printf("Failed to create output dir: %v", err)
		return
	}

	// Get image from screen
	bounds := screen.Bounds()
	img := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			img.Set(x, y, screen.At(x, y))
		}
	}

	// Save to file
	f, err := os.Create(*outputPath)
	if err != nil {
		log.Printf("Failed to create screenshot file: %v", err)
		return
	}
	defer f.Close()

	if err := png.Encode(f, img); err != nil {
		log.Printf("Failed to encode PNG: %v", err)
		return
	}

	log.Printf("Screenshot saved to %s (frame %d)", *outputPath, g.frameCount)
}

func (g *DemoGame) Layout(outsideWidth, outsideHeight int) (int, int) {
	return display.InternalWidth, display.InternalHeight
}

func main() {
	flag.Parse()

	// Print info
	fmt.Println("Tetra3D Demo")
	fmt.Println("============")
	fmt.Printf("Resolution: %dx%d\n", display.InternalWidth, display.InternalHeight)
	if *screenshotFrame > 0 {
		fmt.Printf("Screenshot at frame: %d -> %s\n", *screenshotFrame, *outputPath)
	}
	fmt.Println()

	// Set up window
	ebiten.SetWindowSize(display.InternalWidth, display.InternalHeight)
	ebiten.SetWindowTitle("Stapledon's Voyage - Tetra3D Demo")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	// Run game
	game := NewDemoGame()
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
