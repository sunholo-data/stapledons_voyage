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
	"image/png"
	"log"
	"math"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"stapledons_voyage/engine/display"
	"stapledons_voyage/engine/shader"
	"stapledons_voyage/engine/view"
)

var (
	screenshotFrame = flag.Int("screenshot", 0, "Take screenshot after N frames (0=disabled)")
	outputPath      = flag.String("output", "out/sr-flyby.png", "Screenshot output path")
	maxVelocity     = flag.Float64("max-v", 0.8, "Maximum velocity as fraction of c")
)

// Planet colors for the solar system
var planetConfigs = []struct {
	name   string
	color  color.RGBA
	radius float64
	dist   float64 // distance from center
}{
	{"sun", color.RGBA{255, 200, 50, 255}, 1.5, 0},
	{"mercury", color.RGBA{180, 160, 140, 255}, 0.3, 4},
	{"venus", color.RGBA{230, 200, 150, 255}, 0.5, 6},
	{"earth", color.RGBA{60, 120, 200, 255}, 0.5, 8},
	{"mars", color.RGBA{200, 100, 80, 255}, 0.4, 10},
	{"jupiter", color.RGBA{220, 180, 140, 255}, 1.0, 14},
}

type FlybyGame struct {
	// View system
	spaceView   *view.SpaceView
	planetLayer *view.PlanetLayer

	// Shader system
	shaderMgr *shader.Manager
	srWarp    *shader.SRWarp

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

	// Buffers
	preShaderBuffer *ebiten.Image
}

func NewFlybyGame() *FlybyGame {
	g := &FlybyGame{
		orbitRadius: 20.0,
		orbitSpeed:  0.3,
		orbitAngle:  0,
	}

	// Create space view with starfield
	g.spaceView = view.NewSpaceView()
	g.spaceView.Init()

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

func (g *FlybyGame) createSolarSystem() {
	for _, cfg := range planetConfigs {
		planet := g.planetLayer.AddPlanet(cfg.name, cfg.radius, cfg.color)

		// Position planets in a line (we'll fly past them)
		planet.SetPosition(cfg.dist, 0, 0)

		// Vary rotation speeds
		if cfg.name == "sun" {
			planet.SetRotationSpeed(0.1) // Sun rotates slowly
		} else {
			planet.SetRotationSpeed(0.3 + cfg.dist*0.02)
		}
	}
}

func (g *FlybyGame) Update() error {
	dt := 1.0 / 60.0
	g.time += dt
	g.frameCount++

	// Flyby mode: fly past the solar system in a straight line
	// Camera moves along Z axis, looking forward (+Z direction)

	// Position cycles from -25 to +25, repeating
	cycleLength := 50.0
	cycleTime := cycleLength / ((*maxVelocity) * 10) // Time to complete one pass
	phase := math.Mod(g.time, cycleTime*2) / cycleTime // 0-2 range

	var camZ float64
	if phase < 1.0 {
		// Flying forward (approaching planets)
		camZ = -25.0 + phase*cycleLength
		g.approaching = true
		g.velocity = *maxVelocity
	} else {
		// Flying backward (receding from planets)
		camZ = 25.0 - (phase-1.0)*cycleLength
		g.approaching = false
		g.velocity = *maxVelocity
	}

	camY := 2.0 // Slightly above the plane
	camX := 0.0 // Centered

	g.planetLayer.SetCameraPosition(camX, camY, camZ)

	// SR effect: ViewAngle = 0 means looking forward (in direction of motion)
	// When approaching (flying +Z), objects ahead are blue-shifted
	// When receding (flying -Z), objects "ahead" are red-shifted
	g.srWarp.SetForwardVelocity(g.velocity)
	g.srWarp.SetViewAngle(0) // Always looking forward in direction of motion

	// Update velocity for background stars
	g.spaceView.SetVelocity(g.velocity)

	// Update planets
	g.planetLayer.Update(dt)

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
}

func (g *FlybyGame) drawHUD(screen *ebiten.Image) {
	y := 20.0
	lineHeight := 18.0

	ebitenutil.DebugPrintAt(screen, "SR Flyby Demo", 10, int(y))
	y += lineHeight

	fps := ebiten.ActualFPS()
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("FPS: %.1f", fps), 10, int(y))
	y += lineHeight

	// Velocity display with color indicator
	gamma := 1.0 / math.Sqrt(1-g.velocity*g.velocity+0.0001)
	direction := "APPROACHING"
	if !g.approaching {
		direction = "RECEDING"
	}
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Velocity: %.2fc", g.velocity), 10, int(y))
	y += lineHeight
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Gamma: %.2f", gamma), 10, int(y))
	y += lineHeight
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Direction: %s", direction), 10, int(y))
	y += lineHeight

	// Visual velocity bar
	barWidth := 200.0
	barHeight := 10.0
	barX := 10.0
	barY := y + 5

	// Background bar
	ebitenutil.DrawRect(screen, barX, barY, barWidth, barHeight, color.RGBA{50, 50, 50, 255})

	// Velocity fill
	fillWidth := (g.velocity / *maxVelocity) * barWidth
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

	// Orbit phase
	phaseDeg := math.Mod(g.orbitAngle*180/math.Pi, 360)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Orbit: %.0f deg", phaseDeg), 10, int(y))

	// Help at bottom
	y = float64(display.InternalHeight) - 60
	ebitenutil.DebugPrintAt(screen, "SR Effects: Blue=approaching, Red=receding", 10, int(y))
	y += lineHeight
	ebitenutil.DebugPrintAt(screen, "Watch the Doppler shift change as we orbit!", 10, int(y))
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

func (g *FlybyGame) Layout(outsideWidth, outsideHeight int) (int, int) {
	return display.InternalWidth, display.InternalHeight
}

func main() {
	flag.Parse()

	if *maxVelocity <= 0 || *maxVelocity >= 1 {
		log.Fatal("Max velocity must be between 0 and 1")
	}

	fmt.Println("SR Flyby Demo")
	fmt.Println("=============")
	fmt.Printf("Resolution: %dx%d\n", display.InternalWidth, display.InternalHeight)
	fmt.Printf("Max velocity: %.0f%% c\n", *maxVelocity*100)
	fmt.Println()
	fmt.Println("Watch the Doppler shift change as we orbit the solar system!")
	fmt.Println("- Blue shift when APPROACHING (moving toward planets)")
	fmt.Println("- Red shift when RECEDING (moving away from planets)")
	fmt.Println()

	ebiten.SetWindowSize(display.InternalWidth, display.InternalHeight)
	ebiten.SetWindowTitle("Stapledon's Voyage - SR Flyby Demo")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	game := NewFlybyGame()
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
