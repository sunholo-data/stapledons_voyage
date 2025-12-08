// cmd/demo-sr-flyby/main.go
// Demo showing SR effects during planetary flyby at varying velocities.
// Shows Doppler shift (blue approaching, red receding) and aberration.
// Usage:
//   go run ./cmd/demo-sr-flyby                     # Interactive flyby
//   go run ./cmd/demo-sr-flyby --screenshot 300   # Capture full orbit
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
	outputPath      = flag.String("output", "out/sr-flyby.png", "Screenshot output path")
	startVelocity   = flag.Float64("start-v", 0.9, "Starting velocity as fraction of c")
	endVelocity     = flag.Float64("end-v", 0.3, "Ending velocity as fraction of c")
	demoDuration    = flag.Float64("duration", 60.0, "Demo duration in seconds")
	captureFrames   = flag.Bool("capture", false, "Capture frames for video (exits after capture-duration)")
	captureDuration = flag.Float64("capture-duration", 45.0, "Duration to capture in seconds")
	captureInterval = flag.Int("capture-interval", 2, "Capture every N frames (2=30fps output)")
)

// Planet configs for the solar system - with real textures!
// REVERSED ORDER: Neptune closest (seen first), Sun furthest (climactic ending)
// Camera starts at positive Z, flies toward negative Z
// See Neptune first, pass through the system, end approaching the Sun!
var planetConfigs = []struct {
	name    string
	color   color.RGBA // Fallback color if texture missing
	radius  float64
	dist    float64 // distance from camera start (higher = seen later)
	texture string  // texture path (empty = solid color)
}{
	// Ice giants - FIRST we encounter (closest to start)
	{"neptune", color.RGBA{80, 120, 200, 255}, 1.0, 15, "assets/planets/neptune.jpg"},
	{"uranus", color.RGBA{180, 220, 230, 255}, 1.1, 40, "assets/planets/uranus.jpg"},
	// Gas giants
	{"saturn", color.RGBA{210, 190, 150, 255}, 1.8, 75, "assets/planets/saturn.jpg"},
	{"jupiter", color.RGBA{220, 180, 140, 255}, 2.2, 115, "assets/planets/jupiter.jpg"},
	// Inner planets
	{"mars", color.RGBA{200, 100, 80, 255}, 0.5, 145, "assets/planets/mars.jpg"},
	{"earth", color.RGBA{60, 120, 200, 255}, 0.7, 160, "assets/planets/earth_daymap.jpg"},
	{"venus", color.RGBA{230, 200, 150, 255}, 0.6, 175, "assets/planets/venus_atmosphere.jpg"},
	{"mercury", color.RGBA{180, 160, 140, 255}, 0.45, 188, "assets/planets/mercury.jpg"},
	// Sun at the END - the climax!
	{"sun", color.RGBA{255, 220, 100, 255}, 3.5, 210, "assets/planets/sun.jpg"},
}

type FlybyGame struct {
	// View system
	spaceView   *view.SpaceView
	planetLayer *view.PlanetLayer

	// Shader system
	shaderMgr *shader.Manager
	srWarp    *shader.SRWarp

	// Planet references for special handling
	earth *tetra.Planet

	// Orbit state
	orbitAngle   float64 // Current position in orbit (radians)
	orbitRadius  float64 // Distance from center
	orbitSpeed   float64 // Angular velocity
	velocity     float64 // Current velocity as fraction of c
	approaching  bool    // Are we approaching or receding?

	// Animation
	time       float64
	frameCount int

	// Screenshot
	screenshotTaken bool

	// Frame capture for video
	captureDir     string
	capturedFrames int

	// Buffers
	preShaderBuffer *ebiten.Image
}

func NewFlybyGame() *FlybyGame {
	g := &FlybyGame{
		orbitRadius: 20.0,
		orbitSpeed:  0.3,
		orbitAngle:  0,
		captureDir:  "out/frames",
	}

	// Set up frame capture directory if enabled
	if *captureFrames {
		if err := os.MkdirAll(g.captureDir, 0755); err != nil {
			log.Printf("Failed to create capture dir: %v", err)
		} else {
			log.Printf("Frame capture enabled: %s (%.0fs at interval %d)", g.captureDir, *captureDuration, *captureInterval)
		}
	}

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

	// Create shader manager and SR effect
	g.shaderMgr = shader.NewManager()
	g.srWarp = shader.NewSRWarp(g.shaderMgr)
	g.srWarp.SetEnabled(true)

	// Create pre-shader buffer
	g.preShaderBuffer = ebiten.NewImage(display.InternalWidth, display.InternalHeight)

	// Add planets
	g.createSolarSystem()

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

func (g *FlybyGame) createSolarSystem() {
	// Load ring texture for Saturn
	var ringTex *ebiten.Image
	if tex, err := loadTexture("assets/planets/saturn_ring_gen.png"); err == nil {
		ringTex = tex
		log.Println("Loaded generated ring texture")
	} else if tex, err := loadTexture("assets/planets/saturn_ring.png"); err == nil {
		ringTex = tex
		log.Println("Loaded ring texture")
	}

	for _, cfg := range planetConfigs {
		var planet *tetra.Planet

		// Try to load texture, fall back to solid color
		if cfg.texture != "" {
			tex, err := loadTexture(cfg.texture)
			if err != nil {
				log.Printf("Warning: couldn't load %s texture: %v (using fallback color)", cfg.name, err)
				planet = g.planetLayer.AddPlanet(cfg.name, cfg.radius, cfg.color)
			} else {
				planet = g.planetLayer.AddTexturedPlanet(cfg.name, cfg.radius, tex)
				log.Printf("Loaded texture for %s", cfg.name)
			}
		} else {
			// No texture specified, use solid color (e.g., sun)
			planet = g.planetLayer.AddPlanet(cfg.name, cfg.radius, cfg.color)
		}

		// Position planets along -Z axis (camera looks toward -Z)
		// Sun at 0, Neptune furthest at -190
		// Camera starts at high +Z and flies toward -Z, seeing Neptune first then Sun last
		planet.SetPosition(0, 0, -cfg.dist)

		// Save Earth reference and flip its texture orientation
		if cfg.name == "earth" {
			g.earth = planet
			planet.FlipX() // Fix upside-down texture
		}

		// Vary rotation speeds - inner planets spin faster
		if cfg.name == "sun" {
			planet.SetRotationSpeed(0.02) // Sun rotates slowly
		} else {
			// Scale rotation speed with distance (outer = slower)
			planet.SetRotationSpeed(0.3 - cfg.dist*0.001)
		}

		// Add Saturn's rings
		if cfg.name == "saturn" {
			// Saturn's rings: inner at 1.2× radius, outer at 2.3× radius
			innerR := cfg.radius * 1.2
			outerR := cfg.radius * 2.3
			ring := g.planetLayer.AddRing("saturn_ring", innerR, outerR, ringTex)
			ring.SetPosition(0, 0, -cfg.dist)
			ring.SetTilt(0.47) // Saturn's rings are tilted ~27 degrees (0.47 radians)
			log.Printf("Added Saturn's rings (inner=%.1f, outer=%.1f) at Z=%.0f", innerR, outerR, -cfg.dist)
		}
	}
}

func (g *FlybyGame) Update() error {
	dt := 1.0 / 60.0
	g.time += dt
	g.frameCount++

	// Flyby mode: fly through the solar system along Z axis
	// Planets are positioned along -Z: Sun at 0, Neptune at -85
	// Camera looks toward -Z by default, so start at +Z and fly toward planets

	// Calculate progress through the demo (0.0 to 1.0)
	progress := g.time / *demoDuration
	if progress > 1.0 {
		progress = 1.0 // Clamp at end
	}

	// Velocity profile: linear deceleration from startVelocity to endVelocity
	// v(t) = v0 - (v0 - v1) * t/T
	g.velocity = *startVelocity - (*startVelocity-*endVelocity)*progress
	g.approaching = true // Always approaching in this demo

	// Camera position: fly from Neptune side toward Earth
	// Planets: Neptune at -15, Earth at -160
	// Camera starts ahead of Neptune, ends hovering above Earth
	startZ := 10.0    // Start ahead of Neptune (at -15)
	endZ := -157.0    // End just above Earth (at -160), hovering to admire home

	// Quadratic easing for deceleration feel:
	// Fast at start (0.9c), slow at end (0.3c)
	// This naturally maps to covering more distance early
	easedProgress := progress * (2 - progress) // Quadratic ease-out
	camZ := startZ - (startZ-endZ)*easedProgress

	camY := 4.0 // Elevated for better ring views
	camX := 0.0 // Centered on planet line

	g.planetLayer.SetCameraPosition(camX, camY, camZ)

	// SR effect: ViewAngle = 0 means looking forward (in direction of motion)
	// When approaching, objects ahead are blue-shifted
	g.srWarp.SetForwardVelocity(g.velocity)
	g.srWarp.SetViewAngle(0) // Always looking forward in direction of motion

	// Update velocity for background stars
	g.spaceView.SetVelocity(g.velocity)

	// Update planets
	g.planetLayer.Update(dt)

	// Exit after capture duration if in capture mode
	if *captureFrames && g.time >= *captureDuration {
		log.Printf("Capture complete: %d frames captured", g.capturedFrames)
		return ebiten.Termination
	}

	return nil
}

func (g *FlybyGame) Draw(screen *ebiten.Image) {
	// Render to pre-shader buffer
	target := g.preShaderBuffer
	target.Clear()

	// Layer 1: Starfield background
	g.spaceView.Draw(target)

	// Layer 2: 3D planets
	g.planetLayer.Draw(target)

	// Apply SR shader
	if g.srWarp.IsEnabled() && g.velocity > 0.01 {
		applied := g.srWarp.Apply(screen, target)
		if !applied {
			screen.DrawImage(target, nil)
		}
	} else {
		screen.DrawImage(target, nil)
	}

	// Layer 3: HUD
	g.drawHUD(screen)

	// Screenshot
	if *screenshotFrame > 0 && g.frameCount >= *screenshotFrame && !g.screenshotTaken {
		g.takeScreenshot(screen)
		g.screenshotTaken = true
	}

	// Frame capture for video
	if *captureFrames && g.frameCount%*captureInterval == 0 {
		g.captureFrame(screen)
	}
}

func (g *FlybyGame) drawHUD(screen *ebiten.Image) {
	y := 20.0
	lineHeight := 18.0

	ebitenutil.DebugPrintAt(screen, "Solar System Flyby", 10, int(y))
	y += lineHeight

	fps := ebiten.ActualFPS()
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("FPS: %.1f", fps), 10, int(y))
	y += lineHeight

	// Time progress
	progress := g.time / *demoDuration
	if progress > 1.0 {
		progress = 1.0
	}
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Time: %.1fs / %.0fs (%.0f%%)", g.time, *demoDuration, progress*100), 10, int(y))
	y += lineHeight

	// Velocity display with color indicator
	gamma := 1.0 / math.Sqrt(1-g.velocity*g.velocity+0.0001)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Velocity: %.2fc (decelerating)", g.velocity), 10, int(y))
	y += lineHeight
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Lorentz γ: %.2f (time dilation)", gamma), 10, int(y))
	y += lineHeight

	// Visual velocity bar
	barWidth := 200.0
	barHeight := 10.0
	barX := 10.0
	barY := y + 5

	// Background bar
	ebitenutil.DrawRect(screen, barX, barY, barWidth, barHeight, color.RGBA{50, 50, 50, 255})

	// Velocity fill (relative to start velocity)
	fillWidth := (g.velocity / *startVelocity) * barWidth
	var barColor color.RGBA
	if g.approaching {
		// Blue shift when approaching
		barColor = color.RGBA{100, 150, 255, 255}
	} else {
		// Red shift when receding
		barColor = color.RGBA{255, 100, 100, 255}
	}
	ebitenutil.DrawRect(screen, barX, barY, fillWidth, barHeight, barColor)

	y += lineHeight + 10

	// Help at bottom
	y = float64(display.InternalHeight) - 60
	ebitenutil.DebugPrintAt(screen, "Neptune -> Uranus -> Saturn -> Jupiter -> Mars -> EARTH", 10, int(y))
	y += lineHeight
	ebitenutil.DebugPrintAt(screen, "Decelerating from 0.9c to 0.3c - coming home!", 10, int(y))
	y += lineHeight
	if *screenshotFrame > 0 {
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Screenshot at frame %d", *screenshotFrame), 10, int(y))
	}
}

func (g *FlybyGame) takeScreenshot(screen *ebiten.Image) {
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

func (g *FlybyGame) captureFrame(screen *ebiten.Image) {
	bounds := screen.Bounds()
	img := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			img.Set(x, y, screen.At(x, y))
		}
	}

	// Save frame with sequential numbering
	path := fmt.Sprintf("%s/frame_%05d.png", g.captureDir, g.capturedFrames)
	f, err := os.Create(path)
	if err != nil {
		log.Printf("Failed to create frame file: %v", err)
		return
	}
	defer f.Close()

	if err := png.Encode(f, img); err != nil {
		log.Printf("Failed to encode frame: %v", err)
		return
	}

	g.capturedFrames++
	if g.capturedFrames%100 == 0 {
		log.Printf("Captured %d frames...", g.capturedFrames)
	}
}

func (g *FlybyGame) Layout(outsideWidth, outsideHeight int) (int, int) {
	return display.InternalWidth, display.InternalHeight
}

func main() {
	flag.Parse()

	if *startVelocity <= 0 || *startVelocity >= 1 {
		log.Fatal("Start velocity must be between 0 and 1")
	}
	if *endVelocity <= 0 || *endVelocity >= 1 {
		log.Fatal("End velocity must be between 0 and 1")
	}

	fmt.Println("Solar System Flyby Demo")
	fmt.Println("=======================")
	fmt.Printf("Resolution: %dx%d\n", display.InternalWidth, display.InternalHeight)
	fmt.Printf("Velocity: %.0f%% c → %.0f%% c (deceleration)\n", *startVelocity*100, *endVelocity*100)
	fmt.Printf("Duration: %.0f seconds\n", *demoDuration)
	fmt.Println()
	fmt.Println("Watch the Doppler shift change as velocity decreases!")
	fmt.Println("- Blue shift when traveling fast toward planets")
	fmt.Println("- Less shift as we decelerate to cruise speed")
	fmt.Println()

	ebiten.SetWindowSize(display.InternalWidth, display.InternalHeight)
	ebiten.SetWindowTitle("Stapledon's Voyage - Solar System Flyby")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	game := NewFlybyGame()
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
