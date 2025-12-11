// cmd/demo-solar-system/main.go
// Demo for testing Tetra3D 3D planet rendering composited over starfield.
// This demo isolates the compositing issue identified in the dome renderer.
//
// Usage:
//   go run ./cmd/demo-solar-system                          # Interactive demo
//   go run ./cmd/demo-solar-system --screenshot 60          # Screenshot after N frames
//   go run ./cmd/demo-solar-system --mode opaque            # Test opaque background
//   go run ./cmd/demo-solar-system --mode simple            # Single planet test
//   go run ./cmd/demo-solar-system --mode cruise            # Animated cruise through solar system
package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"stapledons_voyage/engine/display"
	"stapledons_voyage/engine/tetra"
	"stapledons_voyage/engine/view/background"
)

var (
	screenshotFrame = flag.Int("screenshot", 0, "Take screenshot after N frames (0=disabled)")
	outputPath      = flag.String("output", "out/screenshots/demo-solar-system.png", "Screenshot output path")
	mode            = flag.String("mode", "full", "Demo mode: simple|opaque|full|cruise")
	debugCamera     = flag.Bool("debug-camera", false, "Show camera debug info")
)

// PlanetConfig defines a planet for the solar system
type PlanetConfig struct {
	Name        string
	Color       color.RGBA
	Radius      float64
	Distance    float64 // Distance along Z axis (negative = in front of camera)
	YOffset     float64 // Height above cruise path
	TexturePath string
	HasRings    bool
	RingInner   float64
	RingOuter   float64
}

// DemoGame is the main game struct
type DemoGame struct {
	// Background
	spaceBackground *background.SpaceBackground

	// 3D scene
	scene   *tetra.Scene
	planets []*tetra.Planet
	rings   []*tetra.Ring

	// Lighting
	sun     *tetra.SunLight
	ambient *tetra.AmbientLight

	// Camera state
	cameraZ  float64
	cameraY  float64
	velocity float64

	// Animation
	time       float64
	frameCount int

	// Screenshot
	screenshotTaken bool

	// Mode
	demoMode string

	// For opaque mode test
	useOpaqueBackground bool
}

func NewDemoGame(demoMode string) *DemoGame {
	g := &DemoGame{
		demoMode: demoMode,
		cameraZ:  10.0,
		cameraY:  0.0,
		velocity: 0.15,
	}

	screenW := display.InternalWidth
	screenH := display.InternalHeight

	// Create space background (except in opaque mode)
	if demoMode != "opaque" {
		g.spaceBackground = background.NewSpaceBackground(screenW, screenH)
	}

	// Create 3D scene
	g.scene = tetra.NewScene(screenW, screenH)

	// Add lighting - position sun to illuminate planets
	g.sun = tetra.NewSunLight()
	g.sun.SetPosition(5, 3, 15) // Behind camera, shining forward
	g.sun.AddToScene(g.scene)

	// Bright ambient for visibility
	g.ambient = tetra.NewAmbientLight(0.5, 0.5, 0.6, 0.7)
	g.ambient.AddToScene(g.scene)

	// Set up planets based on mode
	switch demoMode {
	case "simple":
		g.setupSimplePlanet()
	case "opaque":
		g.useOpaqueBackground = true
		g.setupSimplePlanet()
	case "cruise":
		g.setupSolarSystem()
		g.cameraZ = 20.0 // Start further back
	default: // "full"
		g.setupSolarSystem()
	}

	// Initial camera position
	g.scene.SetCameraPosition(0, g.cameraY, g.cameraZ)

	return g
}

// setupSimplePlanet creates a single planet for basic testing
func (g *DemoGame) setupSimplePlanet() {
	// Single Earth-like planet at origin - same setup as working demo-tetra
	planet := tetra.NewPlanet("earth", 1.5, color.RGBA{60, 120, 200, 255})
	planet.AddToScene(g.scene)
	planet.SetPosition(0, 0, 0)
	planet.SetRotationSpeed(0.3)
	g.planets = append(g.planets, planet)

	// Camera at Z=4 looking at origin (matches working demo-tetra)
	g.cameraZ = 4.0
	g.cameraY = 0.0
}

// setupSolarSystem creates the full solar system
func (g *DemoGame) setupSolarSystem() {
	configs := []PlanetConfig{
		{"Neptune", color.RGBA{80, 120, 200, 255}, 1.0, -15, 2.25, "assets/planets/neptune.jpg", false, 0, 0},
		{"Saturn", color.RGBA{210, 190, 150, 255}, 1.8, -50, 7.5, "assets/planets/saturn.jpg", true, 2.2, 4.2},
		{"Jupiter", color.RGBA{220, 180, 140, 255}, 2.2, -90, 13.5, "assets/planets/jupiter.jpg", false, 0, 0},
		{"Mars", color.RGBA{200, 100, 80, 255}, 0.5, -130, 19.5, "assets/planets/mars.jpg", false, 0, 0},
		{"Earth", color.RGBA{60, 120, 200, 255}, 0.7, -150, 22.5, "assets/planets/earth_daymap.jpg", false, 0, 0},
	}

	for _, cfg := range configs {
		var planet *tetra.Planet

		// Try to load texture
		tex := loadTexture(cfg.TexturePath)
		if tex != nil {
			planet = tetra.NewTexturedPlanet(cfg.Name, cfg.Radius, tex)
			log.Printf("Loaded texture for %s", cfg.Name)
		} else {
			planet = tetra.NewPlanet(cfg.Name, cfg.Radius, cfg.Color)
			log.Printf("Using solid color for %s (texture not found: %s)", cfg.Name, cfg.TexturePath)
		}

		planet.AddToScene(g.scene)
		planet.SetPosition(0, cfg.YOffset, cfg.Distance)
		planet.SetRotationSpeed(0.1)
		g.planets = append(g.planets, planet)

		// Add rings if configured
		if cfg.HasRings {
			ringTex := loadTexture("assets/planets/saturn_ring.png")
			ring := tetra.NewRing(cfg.Name+"_ring", cfg.RingInner, cfg.RingOuter, ringTex)
			ring.AddToScene(g.scene)
			ring.SetPosition(0, cfg.YOffset, cfg.Distance)
			ring.SetTilt(0.47) // Saturn's tilt
			g.rings = append(g.rings, ring)
			log.Printf("Added rings for %s", cfg.Name)
		}
	}
}

func loadTexture(path string) *ebiten.Image {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return nil
	}

	return ebiten.NewImageFromImage(img)
}

func (g *DemoGame) Update() error {
	dt := 1.0 / 60.0
	g.time += dt
	g.frameCount++

	// Update planet rotations
	for _, p := range g.planets {
		p.Update(dt)
	}
	for _, r := range g.rings {
		r.Update(dt)
	}

	// Cruise mode: animate camera through solar system
	if g.demoMode == "cruise" {
		// Move camera forward (negative Z)
		g.cameraZ -= g.velocity * dt * 20.0

		// Loop back when past all planets
		if g.cameraZ < -160 {
			g.cameraZ = 20.0
		}

		g.scene.SetCameraPosition(0, g.cameraY, g.cameraZ)
	}

	return nil
}

func (g *DemoGame) Draw(screen *ebiten.Image) {
	// Layer 1: Background
	if g.useOpaqueBackground {
		// Opaque black for debugging
		screen.Fill(color.RGBA{0, 0, 0, 255})
	} else if g.spaceBackground != nil {
		// Starfield
		g.spaceBackground.Draw(screen, nil)
	} else {
		// Fallback: dark blue
		screen.Fill(color.RGBA{5, 10, 20, 255})
	}

	// Layer 2: 3D planets (Tetra3D)
	// This is the compositing we're testing!
	img3d := g.scene.Render()
	screen.DrawImage(img3d, nil)

	// Layer 3: HUD
	g.drawHUD(screen)

	// Screenshot
	if *screenshotFrame > 0 && g.frameCount >= *screenshotFrame && !g.screenshotTaken {
		g.takeScreenshot(screen)
		g.screenshotTaken = true
	}
}

func (g *DemoGame) drawHUD(screen *ebiten.Image) {
	y := 20.0
	lineHeight := 18.0

	ebitenutil.DebugPrintAt(screen, "Solar System Demo (Tetra3D)", 10, int(y))
	y += lineHeight

	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Mode: %s", g.demoMode), 10, int(y))
	y += lineHeight

	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Frame: %d", g.frameCount), 10, int(y))
	y += lineHeight

	fps := ebiten.ActualFPS()
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("FPS: %.1f", fps), 10, int(y))
	y += lineHeight

	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Planets: %d", len(g.planets)), 10, int(y))
	y += lineHeight

	if *debugCamera || g.demoMode == "cruise" {
		y += lineHeight
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Camera Z: %.2f", g.cameraZ), 10, int(y))
		y += lineHeight
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Camera Y: %.2f", g.cameraY), 10, int(y))
	}

	// Bottom help
	y = float64(display.InternalHeight) - 60
	ebitenutil.DebugPrintAt(screen, "Testing Tetra3D compositing over starfield", 10, int(y))
	y += lineHeight
	ebitenutil.DebugPrintAt(screen, "If planets are invisible, compositing issue!", 10, int(y))
	y += lineHeight
	if *screenshotFrame > 0 {
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Screenshot at frame %d -> %s", *screenshotFrame, *outputPath), 10, int(y))
	}
}

func (g *DemoGame) takeScreenshot(screen *ebiten.Image) {
	// Create output directory
	if err := os.MkdirAll("out/screenshots", 0755); err != nil {
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

	// Info
	fmt.Println("Solar System Demo (Tetra3D)")
	fmt.Println("===========================")
	fmt.Printf("Resolution: %dx%d\n", display.InternalWidth, display.InternalHeight)
	fmt.Printf("Mode: %s\n", *mode)
	if *screenshotFrame > 0 {
		fmt.Printf("Screenshot at frame: %d -> %s\n", *screenshotFrame, *outputPath)
	}
	fmt.Println()
	fmt.Println("Modes:")
	fmt.Println("  simple  - Single planet on starfield (tests basic compositing)")
	fmt.Println("  opaque  - Single planet on opaque black (baseline)")
	fmt.Println("  full    - All planets, static camera")
	fmt.Println("  cruise  - All planets, animated cruise")
	fmt.Println()

	// Window
	ebiten.SetWindowSize(display.InternalWidth, display.InternalHeight)
	ebiten.SetWindowTitle("Stapledon's Voyage - Solar System Demo")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	// Run
	game := NewDemoGame(*mode)
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
