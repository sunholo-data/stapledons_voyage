// cmd/demo-arrival/main.go
// Demo for arrival sequence: approach at SR speed, decelerate, orbit.
// Shows SR effects with controllable view direction (front/side/back).
// Usage:
//   go run ./cmd/demo-arrival                     # Interactive arrival
//   go run ./cmd/demo-arrival --screenshot 300   # Capture sequence
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
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"stapledons_voyage/engine/display"
	"stapledons_voyage/engine/shader"
	"stapledons_voyage/engine/tetra"
	"stapledons_voyage/engine/view"
)

var (
	screenshotFrame = flag.Int("screenshot", 0, "Take screenshot after N frames (0=disabled)")
	outputPath      = flag.String("output", "out/arrival.png", "Screenshot output path")
	maxVelocity     = flag.Float64("max-v", 0.9, "Starting velocity as fraction of c")
)

// ViewDirection represents where the camera is looking relative to velocity
type ViewDirection int

const (
	ViewForward ViewDirection = iota // Looking in direction of motion
	ViewLeft                         // Looking perpendicular left
	ViewRight                        // Looking perpendicular right
	ViewBehind                       // Looking opposite to motion
)

func (v ViewDirection) String() string {
	switch v {
	case ViewForward:
		return "FORWARD"
	case ViewLeft:
		return "LEFT"
	case ViewRight:
		return "RIGHT"
	case ViewBehind:
		return "BEHIND"
	}
	return "UNKNOWN"
}

func (v ViewDirection) Angle() float64 {
	// ViewAngle is the angle between our view direction and velocity direction
	// 0 = looking in direction of motion (forward)
	// π/2 = looking perpendicular to motion (left or right)
	// π = looking opposite to motion (behind)
	switch v {
	case ViewForward:
		return 0
	case ViewLeft:
		return math.Pi / 2
	case ViewRight:
		return math.Pi / 2 // Same physics as left (perpendicular to velocity)
	case ViewBehind:
		return math.Pi
	}
	return 0
}

// ArrivalPhase represents the current phase of arrival
type ArrivalPhase int

const (
	PhaseApproach ArrivalPhase = iota // High-speed approach
	PhaseDecel                        // Decelerating
	PhaseOrbit                        // Orbiting destination
)

func (p ArrivalPhase) String() string {
	switch p {
	case PhaseApproach:
		return "APPROACH"
	case PhaseDecel:
		return "DECELERATE"
	case PhaseOrbit:
		return "ORBIT"
	}
	return "UNKNOWN"
}

type ArrivalGame struct {
	// View system
	spaceView   *view.SpaceView
	planetLayer *view.PlanetLayer

	// Shader system
	shaderMgr *shader.Manager
	srWarp    *shader.SRWarp

	// Position and velocity
	position  float64 // Distance from destination (0 = at destination)
	velocity  float64 // Current velocity as fraction of c
	phase     ArrivalPhase
	viewDir   ViewDirection
	orbitAngle float64

	// Animation
	time       float64
	frameCount int

	// Screenshot
	screenshotTaken bool

	// Buffers
	preShaderBuffer *ebiten.Image

	// Planet reference
	destPlanet *tetra.Planet
}

func NewArrivalGame() *ArrivalGame {
	g := &ArrivalGame{
		position: 30.0,            // Start 30 units away
		velocity: *maxVelocity,    // Start at max velocity
		phase:    PhaseApproach,
		viewDir:  ViewForward,
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

	// Add destination system
	g.createDestinationSystem()

	return g
}

func (g *ArrivalGame) createDestinationSystem() {
	// Try to load Earth texture, fall back to solid color
	var planet *tetra.Planet

	tex, err := loadTexture("assets/planets/earth.jpg")
	if err != nil {
		log.Printf("Using solid color for destination planet: %v", err)
		planet = g.planetLayer.AddPlanet("destination", 1.5, color.RGBA{60, 130, 220, 255})
	} else {
		planet = g.planetLayer.AddTexturedPlanet("destination", 1.5, tex)
	}

	planet.SetPosition(0, 0, 0)
	planet.SetRotationSpeed(0.1)
	g.destPlanet = planet

	// Add a moon
	moon := g.planetLayer.AddPlanet("moon", 0.3, color.RGBA{180, 180, 180, 255})
	moon.SetPosition(3, 0, 0)
	moon.SetRotationSpeed(0.2)

	// Add a sun/star (behind the planet system)
	sun := g.planetLayer.AddPlanet("sun", 2.0, color.RGBA{255, 220, 100, 255})
	sun.SetPosition(-10, 0, -5)
	sun.SetRotationSpeed(0.02)
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

func (g *ArrivalGame) Update() error {
	dt := 1.0 / 60.0
	g.time += dt
	g.frameCount++

	// Handle view direction controls
	if inpututil.IsKeyJustPressed(ebiten.KeyW) || inpututil.IsKeyJustPressed(ebiten.KeyUp) {
		g.viewDir = ViewForward
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyA) || inpututil.IsKeyJustPressed(ebiten.KeyLeft) {
		g.viewDir = ViewLeft
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyD) || inpututil.IsKeyJustPressed(ebiten.KeyRight) {
		g.viewDir = ViewRight
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyS) || inpututil.IsKeyJustPressed(ebiten.KeyDown) {
		g.viewDir = ViewBehind
	}

	// Handle phase transitions and movement
	switch g.phase {
	case PhaseApproach:
		// High-speed approach
		g.position -= g.velocity * dt * 5.0
		if g.position < 15 {
			g.phase = PhaseDecel
		}

	case PhaseDecel:
		// Decelerate as we approach
		g.velocity = math.Max(0.05, g.velocity-dt*0.2)
		g.position -= g.velocity * dt * 5.0
		if g.position < 5 {
			g.phase = PhaseOrbit
		}

	case PhaseOrbit:
		// Orbit around destination
		g.velocity = 0.1 // Slow orbit speed
		g.orbitAngle += dt * 0.3
	}

	// Update camera position based on phase and view direction
	var camX, camY, camZ float64

	if g.phase == PhaseOrbit {
		// Orbiting - camera moves in circle around planet
		orbitRadius := 5.0
		camX = math.Cos(g.orbitAngle) * orbitRadius
		camZ = math.Sin(g.orbitAngle) * orbitRadius
		camY = 1.0
	} else {
		// Approaching - camera moves along Z axis toward origin
		camX = 0
		camZ = g.position
		camY = 1.0
	}

	g.planetLayer.SetCameraPosition(camX, camY, camZ)

	// Set SR effect for all view directions
	// The shader handles ViewAngle properly for all directions:
	// - ViewAngle = 0: forward (blueshift)
	// - ViewAngle = π/2: perpendicular (mixed)
	// - ViewAngle = π: behind (redshift)
	if g.velocity > 0.05 {
		g.srWarp.SetEnabled(true)
		g.srWarp.SetForwardVelocity(g.velocity)
		g.srWarp.SetViewAngle(g.viewDir.Angle())
	} else {
		g.srWarp.SetEnabled(false)
	}

	// Update velocity for background stars
	g.spaceView.SetVelocity(g.velocity)

	// Update planets
	g.planetLayer.Update(dt)

	return nil
}

func (g *ArrivalGame) Draw(screen *ebiten.Image) {
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

func (g *ArrivalGame) drawHUD(screen *ebiten.Image) {
	y := 20.0
	lineHeight := 18.0

	ebitenutil.DebugPrintAt(screen, "Arrival Demo", 10, int(y))
	y += lineHeight

	fps := ebiten.ActualFPS()
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("FPS: %.1f", fps), 10, int(y))
	y += lineHeight

	// Phase indicator
	phaseColor := ""
	switch g.phase {
	case PhaseApproach:
		phaseColor = ">> "
	case PhaseDecel:
		phaseColor = "-- "
	case PhaseOrbit:
		phaseColor = "~~ "
	}
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Phase: %s%s", phaseColor, g.phase.String()), 10, int(y))
	y += lineHeight

	// Velocity display
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Velocity: %.2fc", g.velocity), 10, int(y))
	y += lineHeight

	// Gamma
	gamma := 1.0 / math.Sqrt(1-g.velocity*g.velocity+0.0001)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Gamma: %.2f", gamma), 10, int(y))
	y += lineHeight

	// Distance
	if g.phase != PhaseOrbit {
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Distance: %.1f ly", g.position), 10, int(y))
		y += lineHeight
	}

	y += 10

	// VIEW DIRECTION indicator (prominent)
	viewBox := fmt.Sprintf("[ VIEW: %s ]", g.viewDir.String())
	var viewColor color.RGBA
	switch g.viewDir {
	case ViewForward:
		viewColor = color.RGBA{100, 150, 255, 255} // Blue (approaching)
	case ViewBehind:
		viewColor = color.RGBA{255, 100, 100, 255} // Red (receding)
	default:
		viewColor = color.RGBA{200, 200, 100, 255} // Yellow (side)
	}
	ebitenutil.DrawRect(screen, 8, y-2, float64(len(viewBox)*7+8), lineHeight+4, viewColor)
	ebitenutil.DebugPrintAt(screen, viewBox, 12, int(y))
	y += lineHeight + 15

	// Doppler shift indicator
	var dopplerLabel string
	if g.viewDir == ViewForward {
		dopplerLabel = "BLUESHIFT (approaching)"
	} else if g.viewDir == ViewBehind {
		dopplerLabel = "REDSHIFT (receding)"
	} else {
		dopplerLabel = "MIXED SHIFT (side view)"
	}
	ebitenutil.DebugPrintAt(screen, dopplerLabel, 10, int(y))

	// Draw compass showing planet direction
	g.drawCompass(screen)

	// Controls at bottom
	y = float64(display.InternalHeight) - 80
	ebitenutil.DebugPrintAt(screen, "Controls:", 10, int(y))
	y += lineHeight
	ebitenutil.DebugPrintAt(screen, "  W/UP    = Look FORWARD  (blueshift)", 10, int(y))
	y += lineHeight
	ebitenutil.DebugPrintAt(screen, "  A/LEFT  = Look LEFT     (mixed)", 10, int(y))
	y += lineHeight
	ebitenutil.DebugPrintAt(screen, "  D/RIGHT = Look RIGHT    (mixed)", 10, int(y))
	y += lineHeight
	ebitenutil.DebugPrintAt(screen, "  S/DOWN  = Look BEHIND   (redshift)", 10, int(y))
}

func (g *ArrivalGame) drawCompass(screen *ebiten.Image) {
	// Compass in top-right corner showing where destination planet is
	centerX := float64(display.InternalWidth) - 80
	centerY := 80.0
	radius := 40.0

	// Draw compass background
	ebitenutil.DrawRect(screen, centerX-radius-5, centerY-radius-5, radius*2+10, radius*2+25, color.RGBA{30, 30, 40, 200})

	// Draw compass circle
	for i := 0; i < 36; i++ {
		angle := float64(i) * math.Pi / 18
		x1 := centerX + math.Cos(angle)*(radius-2)
		y1 := centerY + math.Sin(angle)*(radius-2)
		x2 := centerX + math.Cos(angle)*radius
		y2 := centerY + math.Sin(angle)*radius
		ebitenutil.DrawLine(screen, x1, y1, x2, y2, color.RGBA{100, 100, 100, 255})
	}

	// Calculate planet direction relative to view
	// When looking forward: planet is ahead (up on compass)
	// When looking left: planet is to our right (right on compass)
	// When looking behind: planet is behind (down on compass)
	var planetAngle float64
	switch g.viewDir {
	case ViewForward:
		planetAngle = -math.Pi / 2 // Up
	case ViewLeft:
		planetAngle = 0 // Right
	case ViewRight:
		planetAngle = math.Pi // Left
	case ViewBehind:
		planetAngle = math.Pi / 2 // Down
	}

	// Draw planet indicator (arrow)
	arrowLen := radius * 0.7
	arrowX := centerX + math.Cos(planetAngle)*arrowLen
	arrowY := centerY + math.Sin(planetAngle)*arrowLen

	// Arrow line
	ebitenutil.DrawLine(screen, centerX, centerY, arrowX, arrowY, color.RGBA{60, 200, 120, 255})

	// Arrow head
	headLen := 10.0
	headAngle := 0.4
	h1x := arrowX - math.Cos(planetAngle-headAngle)*headLen
	h1y := arrowY - math.Sin(planetAngle-headAngle)*headLen
	h2x := arrowX - math.Cos(planetAngle+headAngle)*headLen
	h2y := arrowY - math.Sin(planetAngle+headAngle)*headLen
	ebitenutil.DrawLine(screen, arrowX, arrowY, h1x, h1y, color.RGBA{60, 200, 120, 255})
	ebitenutil.DrawLine(screen, arrowX, arrowY, h2x, h2y, color.RGBA{60, 200, 120, 255})

	// Planet dot at arrow tip
	ebitenutil.DrawRect(screen, arrowX-3, arrowY-3, 6, 6, color.RGBA{60, 200, 120, 255})

	// "You" indicator at center
	ebitenutil.DrawRect(screen, centerX-2, centerY-2, 4, 4, color.RGBA{255, 255, 255, 255})

	// Direction labels
	ebitenutil.DebugPrintAt(screen, "N", int(centerX-3), int(centerY-radius-15))
	ebitenutil.DebugPrintAt(screen, "DEST", int(centerX-12), int(centerY+radius+8))
}

func (g *ArrivalGame) takeScreenshot(screen *ebiten.Image) {
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

func (g *ArrivalGame) Layout(outsideWidth, outsideHeight int) (int, int) {
	return display.InternalWidth, display.InternalHeight
}

func main() {
	flag.Parse()

	if *maxVelocity <= 0 || *maxVelocity >= 1 {
		log.Fatal("Max velocity must be between 0 and 1")
	}

	fmt.Println("Arrival Demo")
	fmt.Println("============")
	fmt.Printf("Resolution: %dx%d\n", display.InternalWidth, display.InternalHeight)
	fmt.Printf("Starting velocity: %.0f%% c\n", *maxVelocity*100)
	fmt.Println()
	fmt.Println("Simulates arriving at a star system:")
	fmt.Println("  1. APPROACH - High-speed approach")
	fmt.Println("  2. DECELERATE - Slowing down")
	fmt.Println("  3. ORBIT - Circling destination")
	fmt.Println()
	fmt.Println("Use W/A/S/D or arrows to change view direction")
	fmt.Println("  FORWARD (W) = Blue shift (approaching)")
	fmt.Println("  BEHIND (S)  = Red shift (receding)")
	fmt.Println("  SIDES (A/D) = Mixed effects")
	fmt.Println()

	ebiten.SetWindowSize(display.InternalWidth, display.InternalHeight)
	ebiten.SetWindowTitle("Stapledon's Voyage - Arrival Demo")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	game := NewArrivalGame()
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
