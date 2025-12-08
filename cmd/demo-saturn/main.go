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
	// Saturn rotation period is ~10.7 hours (0.000163 rad/s real-time)
	// For demo visibility, use 0.02 rad/s (~5 min per rotation)
	g.saturn.SetRotationSpeed(0.02)

	// Add Saturn's rings (physically accurate per design doc)
	// Saturn axial tilt: 26.7°, Ring inner: 1.2×, Ring outer: 2.3×
	innerR := saturnRadius * 1.2
	outerR := saturnRadius * 2.3

	// Try to load ring texture (optional)
	// Prefer generated texture which matches our UV mapping
	var ringTex *ebiten.Image
	if tex, err := loadTexture("assets/planets/saturn_ring_gen.png"); err == nil {
		ringTex = tex
		log.Println("Loaded generated ring texture")
	} else if tex, err := loadTexture("assets/planets/saturn_ring.png"); err == nil {
		ringTex = tex
		log.Println("Loaded ring texture")
	}

	g.ring = g.planetLayer.AddRing("saturn_ring", innerR, outerR, ringTex)
	g.ring.SetPosition(0, 0, 0)
	// Saturn's axial tilt is 26.7° = 0.466 radians
	// Positive tilt tilts the north pole toward the viewer
	g.ring.SetTilt(0.466)

	log.Printf("Created Saturn (radius=%.1f) with rings (inner=%.1f, outer=%.1f)", saturnRadius, innerR, outerR)
}

func (g *SaturnGame) Update() error {
	dt := 1.0 / 60.0
	g.time += dt
	g.frameCount++

	// Orbit Saturn around camera (same visual effect as camera orbiting Saturn)
	// Camera stays at origin looking -Z, Saturn orbits in front
	g.orbitAngle += *orbitSpeed * dt

	// Saturn orbits in the XZ plane at negative Z (in front of camera)
	// Camera is elevated to see rings from above
	saturnX := math.Sin(g.orbitAngle) * *distance
	saturnZ := -math.Cos(g.orbitAngle) * *distance // negative Z = in front of camera
	saturnY := 0.0

	// Camera at elevated position, looking toward negative Z
	camY := 2.0
	g.planetLayer.SetCameraPosition(0, camY, 0)

	// Debug logging every 30 frames
	if g.frameCount%30 == 0 {
		log.Printf("Frame %d: Camera(0, %.2f, 0) Saturn(%.2f, %.2f, %.2f) Orbit=%.1f°",
			g.frameCount, camY, saturnX, saturnY, saturnZ, g.orbitAngle*180/math.Pi)
	}

	// Saturn and ring orbit around camera
	g.saturn.SetPosition(saturnX, saturnY, saturnZ)
	g.ring.SetPosition(saturnX, saturnY, saturnZ)

	// Sun from above and behind camera
	g.planetLayer.SetSunPosition(0, 5, 3)
	g.planetLayer.SetSunTarget(saturnX, saturnY, saturnZ)

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

	// Ring info (physically accurate per design doc)
	y += lineHeight
	ebitenutil.DebugPrintAt(screen, "Ring: inner=1.2x, outer=2.3x planet radius", 10, int(y))
	y += lineHeight
	ebitenutil.DebugPrintAt(screen, "Saturn axial tilt: 26.7° (0.466 rad)", 10, int(y))

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
