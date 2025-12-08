// cmd/demo-planets-benchmark/main.go
// Benchmark demo for testing textured planets with SR/GR effects.
// Usage:
//   go run ./cmd/demo-planets-benchmark                          # Single textured Earth
//   go run ./cmd/demo-planets-benchmark --planets 5              # Multiple planets
//   go run ./cmd/demo-planets-benchmark --sr 0.5                 # SR effect at 50% c
//   go run ./cmd/demo-planets-benchmark --screenshot 60          # Screenshot after 60 frames
package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	"log"
	"math"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"stapledons_voyage/engine/display"
	"stapledons_voyage/engine/shader"
	"stapledons_voyage/engine/tetra"
	"stapledons_voyage/engine/view"
)

var (
	screenshotFrame = flag.Int("screenshot", 0, "Take screenshot after N frames (0=disabled)")
	outputPath      = flag.String("output", "out/planets-benchmark.png", "Screenshot output path")
	numPlanets      = flag.Int("planets", 1, "Number of planets to render")
	srVelocity      = flag.Float64("sr", 0.0, "SR velocity as fraction of c (0-0.99)")
	grIntensity     = flag.Float64("gr", 0.0, "GR lensing intensity (0-1)")
)

// Planet textures available
var planetTextures = []string{
	"assets/planets/earth.jpg",
	"assets/planets/mars.jpg",
	"assets/planets/jupiter.jpg",
	"assets/planets/saturn.jpg",
	"assets/planets/moon.jpg",
}

// Planet colors for non-textured fallback
var planetColors = []color.RGBA{
	{60, 120, 200, 255},  // Blue (Earth-like)
	{200, 100, 80, 255},  // Red (Mars-like)
	{220, 180, 140, 255}, // Tan (Jupiter-like)
	{230, 210, 170, 255}, // Yellow (Saturn-like)
	{180, 180, 180, 255}, // Gray (Moon-like)
}

type BenchmarkGame struct {
	// View system
	spaceView   *view.SpaceView
	planetLayer *view.PlanetLayer

	// Shader system
	shaderMgr *shader.Manager
	srWarp    *shader.SRWarp

	// State
	time            float64
	frameCount      int
	screenshotTaken bool

	// Buffers for shader effects
	preShaderBuffer *ebiten.Image
}

func NewBenchmarkGame() *BenchmarkGame {
	g := &BenchmarkGame{}

	// Create space view with starfield
	g.spaceView = view.NewSpaceView()
	g.spaceView.Init()

	// Create planet layer
	g.planetLayer = view.NewPlanetLayer(display.InternalWidth, display.InternalHeight)

	// Create shader manager and SR effect
	g.shaderMgr = shader.NewManager()
	g.srWarp = shader.NewSRWarp(g.shaderMgr)

	// Configure SR effect if velocity specified
	if *srVelocity > 0 {
		g.srWarp.SetEnabled(true)
		g.srWarp.SetForwardVelocity(*srVelocity)
		g.spaceView.SetVelocity(*srVelocity)
	}

	// Configure GR effect
	if *grIntensity > 0 {
		g.spaceView.SetGRIntensity(*grIntensity)
	}

	// Create pre-shader buffer
	g.preShaderBuffer = ebiten.NewImage(display.InternalWidth, display.InternalHeight)

	// Add planets
	g.createPlanets(*numPlanets)

	// Position camera
	g.planetLayer.SetCameraPosition(0, 0, 6)

	return g
}

func (g *BenchmarkGame) createPlanets(count int) {
	for i := 0; i < count; i++ {
		// Load texture
		texIdx := i % len(planetTextures)
		texPath := planetTextures[texIdx]

		var planet *tetra.Planet

		// Try to load texture
		tex, err := loadTexture(texPath)
		if err != nil {
			log.Printf("Failed to load texture %s: %v, using solid color", texPath, err)
			colorIdx := i % len(planetColors)
			planet = g.planetLayer.AddPlanet(
				fmt.Sprintf("planet_%d", i),
				0.8,
				planetColors[colorIdx],
			)
		} else {
			// Create textured planet through the layer
			planet = g.planetLayer.AddTexturedPlanet(
				fmt.Sprintf("planet_%d", i),
				0.8,
				tex,
			)
		}

		// Position planets in a grid or circle
		if count == 1 {
			planet.SetPosition(0, 0, 0)
		} else {
			// Arrange in a circle
			angle := float64(i) * 2 * math.Pi / float64(count)
			radius := 2.0 + float64(count)*0.3
			x := math.Cos(angle) * radius
			y := math.Sin(angle) * radius * 0.5 // Slight perspective
			planet.SetPosition(x, y, 0)
		}

		// Vary rotation speeds
		planet.SetRotationSpeed(0.2 + float64(i)*0.1)
	}
}

func loadTexture(path string) (*ebiten.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return nil, err
	}

	return ebiten.NewImageFromImage(img), nil
}

func (g *BenchmarkGame) Update() error {
	dt := 1.0 / 60.0
	g.time += dt
	g.frameCount++

	// Update all planets (handled by planetLayer)
	g.planetLayer.Update(dt)

	// Slowly orbit camera for multi-planet view
	if *numPlanets > 1 {
		camDist := 6.0 + float64(*numPlanets)*0.5
		angle := g.time * 0.1
		camX := math.Sin(angle) * camDist * 0.3
		camZ := camDist + math.Cos(angle)*camDist*0.1
		g.planetLayer.SetCameraPosition(camX, 0, camZ)
	}

	return nil
}

func (g *BenchmarkGame) Draw(screen *ebiten.Image) {
	// Render to pre-shader buffer first
	target := g.preShaderBuffer
	target.Clear()

	// Layer 1: Background (starfield)
	g.spaceView.Draw(target)

	// Layer 2: 3D planets
	g.planetLayer.Draw(target)

	// Apply SR shader if enabled
	if g.srWarp.IsEnabled() {
		applied := g.srWarp.Apply(screen, target)
		if !applied {
			// Shader not available, copy directly
			screen.DrawImage(target, nil)
		}
	} else {
		// No shader, copy directly
		screen.DrawImage(target, nil)
	}

	// Layer 3: HUD (always on top, no shader)
	g.drawHUD(screen)

	// Take screenshot if requested
	if *screenshotFrame > 0 && g.frameCount >= *screenshotFrame && !g.screenshotTaken {
		g.takeScreenshot(screen)
		g.screenshotTaken = true
	}
}

func (g *BenchmarkGame) drawHUD(screen *ebiten.Image) {
	y := 20.0
	lineHeight := 18.0

	ebitenutil.DebugPrintAt(screen, "Planets Benchmark", 10, int(y))
	y += lineHeight

	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Planets: %d", *numPlanets), 10, int(y))
	y += lineHeight

	fps := ebiten.ActualFPS()
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("FPS: %.1f", fps), 10, int(y))
	y += lineHeight

	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Frame: %d", g.frameCount), 10, int(y))
	y += lineHeight

	if *srVelocity > 0 {
		gamma := g.srWarp.GetGamma()
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("SR: %.1f%% c (gamma=%.2f)", *srVelocity*100, gamma), 10, int(y))
		y += lineHeight
	}

	if *grIntensity > 0 {
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("GR: %.1f%%", *grIntensity*100), 10, int(y))
		y += lineHeight
	}

	// Help at bottom
	y = float64(display.InternalHeight) - 40
	ebitenutil.DebugPrintAt(screen, "Tetra3D + Shaders Benchmark", 10, int(y))
	y += lineHeight
	if *screenshotFrame > 0 {
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Screenshot at frame %d", *screenshotFrame), 10, int(y))
	}
}

func (g *BenchmarkGame) takeScreenshot(screen *ebiten.Image) {
	dir := "out"
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Printf("Failed to create output dir: %v", err)
		return
	}

	bounds := screen.Bounds()
	img := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			img.Set(x, y, screen.At(x, y))
		}
	}

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

func (g *BenchmarkGame) Layout(outsideWidth, outsideHeight int) (int, int) {
	return display.InternalWidth, display.InternalHeight
}

func main() {
	flag.Parse()

	// Validate inputs
	if *srVelocity < 0 || *srVelocity >= 1 {
		log.Fatal("SR velocity must be between 0 and 0.99")
	}
	if *grIntensity < 0 || *grIntensity > 1 {
		log.Fatal("GR intensity must be between 0 and 1")
	}
	if *numPlanets < 1 || *numPlanets > 50 {
		log.Fatal("Number of planets must be between 1 and 50")
	}

	// Print info
	fmt.Println("Planets Benchmark")
	fmt.Println("=================")
	fmt.Printf("Resolution: %dx%d\n", display.InternalWidth, display.InternalHeight)
	fmt.Printf("Planets: %d\n", *numPlanets)
	if *srVelocity > 0 {
		fmt.Printf("SR Velocity: %.1f%% c\n", *srVelocity*100)
	}
	if *grIntensity > 0 {
		fmt.Printf("GR Intensity: %.1f%%\n", *grIntensity*100)
	}
	if *screenshotFrame > 0 {
		fmt.Printf("Screenshot at frame: %d -> %s\n", *screenshotFrame, *outputPath)
	}
	fmt.Println()

	// Set up window
	ebiten.SetWindowSize(display.InternalWidth, display.InternalHeight)
	ebiten.SetWindowTitle("Stapledon's Voyage - Planets Benchmark")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	// Run game
	game := NewBenchmarkGame()
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
