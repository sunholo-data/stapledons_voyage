// cmd/demo-saturn/main.go
// Demo focused on Saturn and its rings for debugging ring rendering.
// Usage:
//   go run ./cmd/demo-saturn                  # Interactive view
//   go run ./cmd/demo-saturn --screenshot 60  # Capture screenshot
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
	"stapledons_voyage/engine/tetra"
	"stapledons_voyage/engine/view"
)

var (
	screenshotFrame = flag.Int("screenshot", 0, "Take screenshot after N frames (0=disabled)")
	outputPath      = flag.String("output", "out/saturn.png", "Screenshot output path")
	orbitSpeed      = flag.Float64("orbit-speed", 0.3, "Camera orbit speed")
	distance        = flag.Float64("distance", 8.0, "Camera distance from Saturn")
)

type SaturnGame struct {
	// View system
	spaceView   *view.SpaceView
	planetLayer *view.PlanetLayer

	// Saturn and ring
	saturn *tetra.Planet
	ring   *tetra.Ring

	// Camera orbit
	orbitAngle float64
	time       float64
	frameCount int

	// Screenshot
	screenshotTaken bool
}

func NewSaturnGame() *SaturnGame {
	g := &SaturnGame{}

	// Create space view with starfield
	g.spaceView = view.NewSpaceView()
	g.spaceView.Init()

	// Load galaxy background if available
	if galaxyImg, err := loadTexture("assets/data/starmap/background/galaxy_4k.jpg"); err == nil {
		g.spaceView.SetGalaxyImage(galaxyImg)
		log.Println("Loaded galaxy background")
	}

	// Create planet layer
	g.planetLayer = view.NewPlanetLayer(display.InternalWidth, display.InternalHeight)

	// Create Saturn
	g.createSaturn()

	return g
}

// loadTexture loads an image file as an Ebiten image
func loadTexture(path string) (*ebiten.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("decode %s: %w", path, err)
	}

	return ebiten.NewImageFromImage(img), nil
}

func (g *SaturnGame) createSaturn() {
	saturnRadius := 2.0

	// Try to load Saturn texture
	tex, err := loadTexture("assets/planets/saturn.jpg")
	if err != nil {
		log.Printf("Warning: couldn't load saturn texture: %v (using fallback color)", err)
		g.saturn = g.planetLayer.AddPlanet("saturn", saturnRadius, color.RGBA{210, 190, 150, 255})
	} else {
		g.saturn = g.planetLayer.AddTexturedPlanet("saturn", saturnRadius, tex)
		log.Println("Loaded Saturn texture")
	}

	// Position Saturn at origin
	g.saturn.SetPosition(0, 0, 0)
	g.saturn.SetRotationSpeed(0.3)

	// Add Saturn's rings
	// Inner ring at 1.2× radius, outer at 2.5× radius
	innerR := saturnRadius * 1.2
	outerR := saturnRadius * 2.5

	// Try to load ring texture (optional)
	var ringTex *ebiten.Image
	if tex, err := loadTexture("assets/planets/saturn_ring.png"); err == nil {
		ringTex = tex
		log.Println("Loaded ring texture")
	}

	g.ring = g.planetLayer.AddRing("saturn_ring", innerR, outerR, ringTex)
	g.ring.SetPosition(0, 0, 0)
	g.ring.SetTilt(0.47) // ~27 degrees tilt

	log.Printf("Created Saturn (radius=%.1f) with rings (inner=%.1f, outer=%.1f)", saturnRadius, innerR, outerR)
}

func (g *SaturnGame) Update() error {
	dt := 1.0 / 60.0
	g.time += dt
	g.frameCount++

	// Keep camera stationary, looking down -Z axis
	// Position camera slightly elevated for better ring view angle
	g.planetLayer.SetCameraPosition(0, 2, *distance)

	// Position sun to illuminate the camera-facing side
	// Sun in front-right of Saturn for nice side lighting
	g.planetLayer.SetSunPosition(5, 5, 5)

	// Rotate Saturn slowly (planet already rotates via Update)
	// But also slowly rotate the whole view angle by moving Saturn in a circle
	g.orbitAngle += *orbitSpeed * dt * 0.1

	// Position Saturn at origin (it rotates on its own axis)
	g.saturn.SetPosition(0, 0, 0)
	g.ring.SetPosition(0, 0, 0)

	// Update planets
	g.planetLayer.Update(dt)

	return nil
}

func (g *SaturnGame) Draw(screen *ebiten.Image) {
	// Layer 1: Starfield background
	g.spaceView.Draw(screen)

	// Layer 2: Saturn and rings
	g.planetLayer.Draw(screen)

	// Layer 3: HUD
	g.drawHUD(screen)

	// Screenshot
	if *screenshotFrame > 0 && g.frameCount >= *screenshotFrame && !g.screenshotTaken {
		g.takeScreenshot(screen)
		g.screenshotTaken = true
	}
}

func (g *SaturnGame) drawHUD(screen *ebiten.Image) {
	y := 20.0
	lineHeight := 18.0

	ebitenutil.DebugPrintAt(screen, "Saturn Ring Demo", 10, int(y))
	y += lineHeight

	fps := ebiten.ActualFPS()
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("FPS: %.1f", fps), 10, int(y))
	y += lineHeight

	// Orbit info
	orbitDeg := math.Mod(g.orbitAngle*180/math.Pi, 360)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Camera orbit: %.0f°", orbitDeg), 10, int(y))
	y += lineHeight

	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Distance: %.1f", *distance), 10, int(y))
	y += lineHeight

	// Ring info
	y += lineHeight
	ebitenutil.DebugPrintAt(screen, "Ring: inner=2.4, outer=5.0", 10, int(y))
	y += lineHeight
	ebitenutil.DebugPrintAt(screen, "Ring tilt: 27° (~0.47 rad)", 10, int(y))

	// Help at bottom
	y = float64(display.InternalHeight) - 40
	ebitenutil.DebugPrintAt(screen, "Camera orbits Saturn to view rings from all angles", 10, int(y))
	y += lineHeight
	if *screenshotFrame > 0 {
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Screenshot at frame %d", *screenshotFrame), 10, int(y))
	}
}

func (g *SaturnGame) takeScreenshot(screen *ebiten.Image) {
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

func (g *SaturnGame) Layout(outsideWidth, outsideHeight int) (int, int) {
	return display.InternalWidth, display.InternalHeight
}

func main() {
	flag.Parse()

	fmt.Println("Saturn Ring Demo")
	fmt.Println("================")
	fmt.Printf("Resolution: %dx%d\n", display.InternalWidth, display.InternalHeight)
	fmt.Printf("Camera distance: %.1f\n", *distance)
	fmt.Printf("Orbit speed: %.2f\n", *orbitSpeed)
	fmt.Println()
	fmt.Println("Camera orbits around Saturn to view rings from all angles.")
	fmt.Println()

	ebiten.SetWindowSize(display.InternalWidth, display.InternalHeight)
	ebiten.SetWindowTitle("Stapledon's Voyage - Saturn Ring Demo")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	game := NewSaturnGame()
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
