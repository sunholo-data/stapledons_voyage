// cmd/demo-planet-view/main.go
// Demo command for testing 3D planet composited over starfield.
// Usage:
//   go run ./cmd/demo-planet-view                     # Interactive demo
//   go run ./cmd/demo-planet-view --screenshot 60    # Take screenshot after N frames
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
	"stapledons_voyage/engine/view"
)

var (
	screenshotFrame = flag.Int("screenshot", 0, "Take screenshot after N frames (0=disabled)")
	outputPath      = flag.String("output", "out/demo-planet-view.png", "Screenshot output path")
)

type DemoGame struct {
	spaceView   *view.SpaceView
	planetLayer *view.PlanetLayer

	// Animation state
	time       float64
	frameCount int

	// Screenshot state
	screenshotTaken bool
}

func NewDemoGame() *DemoGame {
	g := &DemoGame{}

	// Create space view with starfield background
	g.spaceView = view.NewSpaceView()
	g.spaceView.Init()

	// Create planet layer
	g.planetLayer = view.NewPlanetLayer(display.InternalWidth, display.InternalHeight)

	// Add an Earth-like planet
	planet := g.planetLayer.AddPlanet("earth", 1.0, color.RGBA{60, 120, 200, 255})
	planet.SetPosition(0, 0, 0)
	planet.SetRotationSpeed(0.3)

	// Position camera a bit further back
	g.planetLayer.SetCameraPosition(0, 0, 4)

	return g
}

func (g *DemoGame) Update() error {
	dt := 1.0 / 60.0
	g.time += dt
	g.frameCount++

	// Update planet rotation
	g.planetLayer.Update(dt)

	// Slowly move camera for visual interest
	camZ := 4.0 + 0.5*float64(g.frameCount%360)/60.0
	g.planetLayer.SetCameraPosition(0, 0, camZ)

	return nil
}

func (g *DemoGame) Draw(screen *ebiten.Image) {
	// Layer 1: Background (starfield)
	g.spaceView.Draw(screen)

	// Layer 2: 3D content (planets)
	g.planetLayer.Draw(screen)

	// Layer 3: HUD (on top)
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

	ebitenutil.DebugPrintAt(screen, "Planet View Demo", 10, int(y))
	y += lineHeight

	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Frame: %d", g.frameCount), 10, int(y))
	y += lineHeight

	fps := ebiten.ActualFPS()
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("FPS: %.1f", fps), 10, int(y))
	y += lineHeight

	ebitenutil.DebugPrintAt(screen, "Starfield + 3D Planet composite", 10, int(y))
	y += lineHeight

	// Help at bottom
	y = float64(display.InternalHeight) - 40
	ebitenutil.DebugPrintAt(screen, "Tetra3D + View System Integration", 10, int(y))
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
	fmt.Println("Planet View Demo")
	fmt.Println("================")
	fmt.Printf("Resolution: %dx%d\n", display.InternalWidth, display.InternalHeight)
	if *screenshotFrame > 0 {
		fmt.Printf("Screenshot at frame: %d -> %s\n", *screenshotFrame, *outputPath)
	}
	fmt.Println()

	// Set up window
	ebiten.SetWindowSize(display.InternalWidth, display.InternalHeight)
	ebiten.SetWindowTitle("Stapledon's Voyage - Planet View Demo")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	// Run game
	game := NewDemoGame()
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
