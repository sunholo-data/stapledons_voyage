// cmd/demo-game-saturn/main.go
// AILANG-driven Saturn ring demo.
// Shows Saturn with 3D rings and an orbiting camera for ring experimentation.
// Saturn data comes from AILANG (sim/celestial.ail).
//
// Usage:
//
//	go build -o bin/demo-game-saturn ./cmd/demo-game-saturn && bin/demo-game-saturn
//	bin/demo-game-saturn --screenshot 120              # Screenshot after 120 frames
//	bin/demo-game-saturn --orbit-speed=0.5             # Slower orbit
//	bin/demo-game-saturn --ring-inner=1.3              # Adjust ring inner radius
//	bin/demo-game-saturn --ring-outer=2.5              # Adjust ring outer radius
//	bin/demo-game-saturn --tilt=27                     # Ring tilt in degrees
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
	"stapledons_voyage/engine/handlers"
	"stapledons_voyage/engine/tetra"
	"stapledons_voyage/engine/view/background"
	"stapledons_voyage/sim_gen"
)

var (
	screenshotFrame = flag.Int("screenshot", 0, "Take screenshot after N frames (0=disabled)")
	outputPath      = flag.String("output", "out/screenshots/demo-saturn.png", "Screenshot output path")
	orbitSpeed      = flag.Float64("orbit-speed", 0.3, "Camera orbit speed (radians per second)")
	ringInner       = flag.Float64("ring-inner", 1.3, "Ring inner radius (planet radii)")
	ringOuter       = flag.Float64("ring-outer", 2.5, "Ring outer radius (planet radii)")
	ringTilt        = flag.Float64("tilt", 27.0, "Ring tilt in degrees")
	cameraDistance  = flag.Float64("distance", 10.0, "Camera distance from Saturn")
	cameraHeight    = flag.Float64("height", 3.0, "Camera height above ring plane")
)

// SaturnGame is the main game struct
type SaturnGame struct {
	// AILANG state (for saturn data)
	starSystem *sim_gen.StarSystem
	saturnData *sim_gen.CelestialPlanet

	// 3D rendering
	scene3D   *tetra.Scene
	saturn    *tetra.Planet
	ring      *tetra.Ring
	sun       *tetra.SunLight
	ambient   *tetra.AmbientLight

	// Background
	spaceBackground *background.SpaceBackground

	// Camera orbit around Saturn (using fixed LookAt)
	orbitAngle     float64 // Current angle of camera around Saturn (radians)
	cameraDistance float64 // Distance from camera to Saturn
	cameraHeight   float64 // Camera Y offset above the ring plane
	orbitSpeed     float64

	// Animation
	time       float64
	frameCount int

	// Screenshot
	screenshotTaken bool
}

func NewSaturnGame() *SaturnGame {
	screenW := display.InternalWidth
	screenH := display.InternalHeight

	g := &SaturnGame{
		orbitAngle:     0,
		cameraDistance: *cameraDistance,
		cameraHeight:   *cameraHeight,
		orbitSpeed:     *orbitSpeed,
	}

	// Initialize AILANG handlers
	sim_gen.Init(sim_gen.Handlers{
		Debug: sim_gen.NewDebugContext(),
		Rand:  &handlers.DefaultRandHandler{},
		Clock: &handlers.EbitenClockHandler{},
		AI:    handlers.NewStubAIHandler(),
	})

	// Initialize solar system from AILANG to get Saturn data
	g.starSystem = sim_gen.InitSolSystem()
	log.Printf("Initialized solar system from AILANG: %s", g.starSystem.Name)

	// Find Saturn in the planet list
	for _, p := range g.starSystem.Planets {
		if p.Name == "Saturn" {
			g.saturnData = p
			log.Printf("Found Saturn: orbit %.2f AU, period %.2f yr", p.OrbitDistance, p.OrbitalPeriod)
			break
		}
	}
	if g.saturnData == nil {
		log.Printf("Warning: Saturn not found in AILANG data, using defaults")
	}

	// Create space background
	g.spaceBackground = background.NewSpaceBackground(screenW, screenH)

	// Setup 3D scene
	g.setup3DScene(screenW, screenH)

	return g
}

func (g *SaturnGame) setup3DScene(screenW, screenH int) {
	g.scene3D = tetra.NewScene(screenW, screenH)

	// Load Saturn texture
	saturnTex := loadTexture("assets/planets/saturn.jpg")

	// Saturn radius for scene - large enough to appreciate the details
	saturnRadius := 3.0

	// Create Saturn at origin
	if saturnTex != nil {
		g.saturn = tetra.NewTexturedPlanet("saturn", saturnRadius, saturnTex)
		log.Printf("Created textured Saturn")
	} else {
		g.saturn = tetra.NewPlanet("saturn", saturnRadius, color.RGBA{210, 190, 150, 255})
		log.Printf("Created solid Saturn (no texture)")
	}
	g.saturn.AddToScene(g.scene3D)
	g.saturn.SetPosition(0, 0, 0) // Saturn at origin
	g.saturn.SetRotationSpeed(0.1) // Slow self-rotation

	// Create rings using command-line parameters (generated mesh, no texture needed)
	ringInnerR := saturnRadius * *ringInner
	ringOuterR := saturnRadius * *ringOuter
	tiltRadians := *ringTilt * math.Pi / 180.0

	// Use nil texture - NewRing generates solid colored ring mesh
	g.ring = tetra.NewRing("saturn_ring", ringInnerR, ringOuterR, nil)
	g.ring.AddToScene(g.scene3D)
	g.ring.SetPosition(0, 0, 0) // Same position as Saturn (origin)
	g.ring.SetTilt(tiltRadians)

	log.Printf("Ring parameters: inner=%.2f, outer=%.2f, tilt=%.1f deg",
		ringInnerR, ringOuterR, *ringTilt)

	// Add lighting - from above/side for good ring visibility
	g.sun = tetra.NewSunLight()
	g.sun.SetPosition(30, 20, 30) // Light from above-right
	g.sun.AddToScene(g.scene3D)

	g.ambient = tetra.NewAmbientLight(0.4, 0.4, 0.45, 0.6)
	g.ambient.AddToScene(g.scene3D)

	// Initial camera position - will be updated in updateCameraOrbit
	g.updateCameraOrbit()

	log.Printf("Saturn at origin, camera orbiting at distance %.1f", g.cameraDistance)
}

// updateCameraOrbit positions the camera in orbit around Saturn and points it at Saturn
func (g *SaturnGame) updateCameraOrbit() {
	// Camera position on orbit circle around Saturn (at origin)
	x := g.cameraDistance * math.Cos(g.orbitAngle)
	z := g.cameraDistance * math.Sin(g.orbitAngle)
	y := g.cameraHeight

	g.scene3D.SetCameraPosition(x, y, z)
	g.scene3D.LookAt(0, 0, 0) // Look at Saturn at origin
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

func (g *SaturnGame) Update() error {
	dt := 1.0 / 60.0
	g.time += dt
	g.frameCount++

	// Update camera orbit angle
	g.orbitAngle += g.orbitSpeed * dt
	if g.orbitAngle > 2*math.Pi {
		g.orbitAngle -= 2 * math.Pi
	}

	// Move camera in orbit around Saturn (now using fixed LookAt!)
	g.updateCameraOrbit()

	// Saturn rotates on its axis
	g.saturn.Update(dt)

	// Ring rotates slowly with Saturn
	g.ring.Update(dt * 0.05)

	return nil
}

func (g *SaturnGame) Draw(screen *ebiten.Image) {
	// Layer 1: Space background
	if g.spaceBackground != nil {
		g.spaceBackground.Draw(screen, nil)
	} else {
		screen.Fill(color.RGBA{5, 10, 20, 255})
	}

	// Layer 2: 3D scene
	img3d := g.scene3D.Render()
	screen.DrawImage(img3d, nil)

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

	ebitenutil.DebugPrintAt(screen, "Saturn Ring Demo (AILANG-driven)", 10, int(y))
	y += lineHeight

	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Frame: %d", g.frameCount), 10, int(y))
	y += lineHeight

	fps := ebiten.ActualFPS()
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("FPS: %.1f", fps), 10, int(y))
	y += lineHeight

	y += lineHeight

	// Ring parameters
	ebitenutil.DebugPrintAt(screen, "Ring Parameters:", 10, int(y))
	y += lineHeight
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("  Inner: %.2f x radius", *ringInner), 10, int(y))
	y += lineHeight
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("  Outer: %.2f x radius", *ringOuter), 10, int(y))
	y += lineHeight
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("  Tilt: %.1f degrees", *ringTilt), 10, int(y))
	y += lineHeight

	y += lineHeight

	// Camera info
	ebitenutil.DebugPrintAt(screen, "Camera:", 10, int(y))
	y += lineHeight
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("  Angle: %.1f deg", g.orbitAngle*180/math.Pi), 10, int(y))
	y += lineHeight
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("  Distance: %.1f", g.cameraDistance), 10, int(y))
	y += lineHeight
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("  Height: %.1f", g.cameraHeight), 10, int(y))
	y += lineHeight
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("  Orbit speed: %.2f rad/s", g.orbitSpeed), 10, int(y))
	y += lineHeight

	// AILANG data (if available)
	if g.saturnData != nil {
		y += lineHeight
		ebitenutil.DebugPrintAt(screen, "Saturn (from AILANG):", 10, int(y))
		y += lineHeight
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("  Orbit: %.2f AU", g.saturnData.OrbitDistance), 10, int(y))
		y += lineHeight
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("  Period: %.2f years", g.saturnData.OrbitalPeriod), 10, int(y))
		y += lineHeight
	}

	// Bottom help
	y = float64(display.InternalHeight) - 80
	ebitenutil.DebugPrintAt(screen, "Flags for experimentation:", 10, int(y))
	y += lineHeight
	ebitenutil.DebugPrintAt(screen, "  --ring-inner, --ring-outer, --tilt", 10, int(y))
	y += lineHeight
	ebitenutil.DebugPrintAt(screen, "  --distance, --height, --orbit-speed", 10, int(y))
	y += lineHeight
	if *screenshotFrame > 0 {
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Screenshot at frame %d -> %s", *screenshotFrame, *outputPath), 10, int(y))
	}
}

func (g *SaturnGame) takeScreenshot(screen *ebiten.Image) {
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

func (g *SaturnGame) Layout(outsideWidth, outsideHeight int) (int, int) {
	return display.InternalWidth, display.InternalHeight
}

func main() {
	flag.Parse()

	fmt.Println("Saturn Ring Demo (AILANG-driven)")
	fmt.Println("=================================")
	fmt.Printf("Resolution: %dx%d\n", display.InternalWidth, display.InternalHeight)
	fmt.Println()
	fmt.Println("Ring parameters:")
	fmt.Printf("  Inner radius: %.2f x Saturn radius\n", *ringInner)
	fmt.Printf("  Outer radius: %.2f x Saturn radius\n", *ringOuter)
	fmt.Printf("  Tilt: %.1f degrees\n", *ringTilt)
	fmt.Println()
	fmt.Println("Camera parameters:")
	fmt.Printf("  Distance: %.1f\n", *cameraDistance)
	fmt.Printf("  Height: %.1f\n", *cameraHeight)
	fmt.Printf("  Orbit speed: %.2f rad/s\n", *orbitSpeed)
	fmt.Println()
	if *screenshotFrame > 0 {
		fmt.Printf("Screenshot at frame: %d -> %s\n", *screenshotFrame, *outputPath)
	}
	fmt.Println()
	fmt.Println("This demo shows Saturn with 3D rings using data from AILANG.")
	fmt.Println("The camera orbits around Saturn for ring visualization.")
	fmt.Println()

	ebiten.SetWindowSize(display.InternalWidth, display.InternalHeight)
	ebiten.SetWindowTitle("Stapledon's Voyage - Saturn Ring Demo (AILANG)")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	game := NewSaturnGame()
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
