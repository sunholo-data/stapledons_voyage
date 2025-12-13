// Package main provides an AILANG-controlled solar system demo.
// This validates the AILANG-first architecture by having AILANG control all celestial data
// while the Go engine only handles rendering.
package main

import (
	"flag"
	"fmt"
	"image/color"
	"image/png"
	"log"
	"math"
	"os"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"stapledons_voyage/engine/shader"
	"stapledons_voyage/engine/tetra"
	"stapledons_voyage/sim_gen"
)

const (
	screenWidth  = 1280
	screenHeight = 720
)

// Game implements ebiten.Game interface for the AILANG solar demo.
type Game struct {
	// AILANG state
	state *sim_gen.SolarDemoState

	// Tetra3D scene for 3D rendering
	scene3D *tetra.Scene

	// SR/GR shader effects
	shaderManager *shader.Manager
	srWarp        *shader.SRWarp
	grWarp        *shader.GRWarp
	renderBuffer  *ebiten.Image

	// Light sources
	sunLight     *tetra.StarLight
	ambientLight *tetra.AmbientLight

	// Planets in 3D scene
	planets []*tetra.Planet

	// Saturn's ring system
	saturnRings *tetra.RingSystem

	// Camera movement
	cameraSpeed float64
	lastUpdate  time.Time

	// Lighting controls
	lightMultiplier float64
	ambientLevel    float64

	// Velocity (fraction of c)
	velocity float64

	// Frame counter
	frameCount      int
	screenshotFrame int
	screenshotPath  string
	screenshotTaken bool
}

// NewGame creates a new AILANG solar demo.
func NewGame(screenshotFrame int, screenshotPath string) *Game {
	// Initialize AILANG state
	state := sim_gen.InitSolarDemo()

	// Create Tetra3D scene
	scene3D := tetra.NewScene(screenWidth, screenHeight)
	scene3D.SetLightingEnabled(true)

	// Set camera from AILANG state
	scene3D.SetCameraPosition(state.CameraX, state.CameraY, state.CameraZ)
	scene3D.LookAt(state.LookAtX, state.LookAtY, state.LookAtZ)

	// Create 3D planets from AILANG data
	// GitHub issue #47 FIXED - list literal codegen now returns typed slices
	ailangPlanets := sim_gen.GetSolarDemoPlanets()
	var planets []*tetra.Planet
	var saturnRings *tetra.RingSystem
	for _, p := range ailangPlanets {
		col := rgbaFromInt(p.ColorRgba)
		planet := tetra.NewPlanet(p.Name, p.Radius, col)
		planet.SetPosition(p.PosX, p.PosY, p.PosZ)
		planet.AddToScene(scene3D)

		// Sun should be self-illuminated (shadeless)
		if p.Name == "Sun" {
			planet.SetShadeless(true)
		}

		// Add rings to Saturn
		if p.Name == "Saturn" {
			bands := tetra.SaturnRingBands(p.Radius)
			saturnRings = tetra.NewRingSystem("saturn", bands)
			saturnRings.AddToScene(scene3D)
			saturnRings.SetPosition(p.PosX, p.PosY, p.PosZ)
			saturnRings.SetTilt(0.47) // Saturn's axial tilt
			log.Printf("Added Saturn ring system with %d bands", len(bands))
		}

		planets = append(planets, planet)
	}

	// Create sun light from AILANG data
	sunLight := tetra.NewStarLight("sun_light", 1.0, 0.95, 0.85, state.SunEnergy, 0)
	sunLight.SetPosition(0, 0, 0)
	sunLight.AddToScene(scene3D)

	// Create ambient light
	ambientLight := tetra.NewAmbientLight(0.08, 0.08, 0.1, state.AmbientLevel)
	ambientLight.AddToScene(scene3D)

	// Create shader manager for SR/GR effects
	shaderManager := shader.NewManager()

	// Create SR warp shader
	srWarp := shader.NewSRWarp(shaderManager)

	// Create GR warp shader
	grWarp := shader.NewGRWarp(shaderManager)

	// Pre-configure GR for demo mode (centered on screen)
	grWarp.SetDemoMode(0.5, 0.5, 0.08, 0.01)

	// Create off-screen render buffer for post-processing
	renderBuffer := ebiten.NewImage(screenWidth, screenHeight)

	log.Printf("AILANG Solar Demo: Loaded %d planets from AILANG", len(ailangPlanets))
	log.Printf("  Sun Energy: %.0f, Ambient: %.2f", state.SunEnergy, state.AmbientLevel)

	return &Game{
		state:           state,
		scene3D:         scene3D,
		shaderManager:   shaderManager,
		srWarp:          srWarp,
		grWarp:          grWarp,
		renderBuffer:    renderBuffer,
		sunLight:        sunLight,
		ambientLight:    ambientLight,
		planets:         planets,
		saturnRings:     saturnRings,
		cameraSpeed:     50.0,
		lastUpdate:      time.Now(),
		lightMultiplier: 1.0,
		ambientLevel:    1.0,
		velocity:        0.0,
		screenshotFrame: screenshotFrame,
		screenshotPath:  screenshotPath,
	}
}

// Update implements ebiten.Game interface.
func (g *Game) Update() error {
	g.frameCount++

	// Calculate delta time
	now := time.Now()
	dt := now.Sub(g.lastUpdate).Seconds()
	g.lastUpdate = now

	// Camera movement with WASD
	moveSpeed := g.cameraSpeed * dt
	if ebiten.IsKeyPressed(ebiten.KeyShift) {
		moveSpeed *= 3.0 // Fast mode
	}

	if ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyUp) {
		g.state.CameraZ -= moveSpeed * 2
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyDown) {
		g.state.CameraZ += moveSpeed * 2
	}
	if ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyLeft) {
		g.state.CameraX -= moveSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyRight) {
		g.state.CameraX += moveSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyQ) {
		g.state.CameraY += moveSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyE) {
		g.state.CameraY -= moveSpeed
	}

	// R: Reset camera position
	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		g.state.CameraX = 300.0
		g.state.CameraY = 100.0
		g.state.CameraZ = 200.0
		log.Printf("Camera reset to initial position")
	}

	// SR/GR effect controls
	// 1: Toggle SR warp effect
	if inpututil.IsKeyJustPressed(ebiten.Key1) {
		g.srWarp.Toggle()
		if g.srWarp.IsEnabled() {
			log.Printf("SR Warp ENABLED (velocity: %.1f%%c)", g.velocity*100)
		} else {
			log.Printf("SR Warp DISABLED")
		}
	}

	// 2: Toggle GR warp effect
	if inpututil.IsKeyJustPressed(ebiten.Key2) {
		g.grWarp.Toggle()
		if g.grWarp.IsEnabled() {
			log.Printf("GR Warp ENABLED (intensity: %s)", g.grWarp.GetDemoIntensity())
		} else {
			log.Printf("GR Warp DISABLED")
		}
	}

	// 3: Cycle GR intensity
	if inpututil.IsKeyJustPressed(ebiten.Key3) && g.grWarp.IsEnabled() {
		intensity := g.grWarp.CycleDemoIntensity()
		log.Printf("GR intensity: %s", intensity)
	}

	// +/=: Increase velocity
	if ebiten.IsKeyPressed(ebiten.KeyEqual) || ebiten.IsKeyPressed(ebiten.KeyKPAdd) {
		g.velocity += 0.005
		if g.velocity > 0.99 {
			g.velocity = 0.99
		}
		g.srWarp.SetForwardVelocity(g.velocity)
		g.state.ShipVelocity = g.velocity
	}

	// -: Decrease velocity
	if ebiten.IsKeyPressed(ebiten.KeyMinus) || ebiten.IsKeyPressed(ebiten.KeyKPSubtract) {
		g.velocity -= 0.005
		if g.velocity < 0 {
			g.velocity = 0
		}
		g.srWarp.SetForwardVelocity(g.velocity)
		g.state.ShipVelocity = g.velocity
	}

	// 0: Reset velocity
	if inpututil.IsKeyJustPressed(ebiten.Key0) {
		g.velocity = 0
		g.srWarp.SetForwardVelocity(0)
		g.state.ShipVelocity = 0
		log.Printf("Velocity reset to 0")
	}

	// [: Decrease star light intensity
	if ebiten.IsKeyPressed(ebiten.KeyLeftBracket) {
		g.lightMultiplier -= 0.02
		if g.lightMultiplier < 0.1 {
			g.lightMultiplier = 0.1
		}
		g.sunLight.SetEnergy(g.state.SunEnergy * g.lightMultiplier)
	}

	// ]: Increase star light intensity
	if ebiten.IsKeyPressed(ebiten.KeyRightBracket) {
		g.lightMultiplier += 0.02
		if g.lightMultiplier > 3.0 {
			g.lightMultiplier = 3.0
		}
		g.sunLight.SetEnergy(g.state.SunEnergy * g.lightMultiplier)
	}

	// ;: Decrease ambient light
	if ebiten.IsKeyPressed(ebiten.KeySemicolon) {
		g.ambientLevel -= 0.02
		if g.ambientLevel < 0.0 {
			g.ambientLevel = 0.0
		}
		g.ambientLight.SetEnergy(g.ambientLevel)
	}

	// ': Increase ambient light
	if ebiten.IsKeyPressed(ebiten.KeyApostrophe) {
		g.ambientLevel += 0.02
		if g.ambientLevel > 2.0 {
			g.ambientLevel = 2.0
		}
		g.ambientLight.SetEnergy(g.ambientLevel)
	}

	// L: Reset light levels
	if inpututil.IsKeyJustPressed(ebiten.KeyL) {
		g.lightMultiplier = 1.0
		g.ambientLevel = 1.0
		g.sunLight.SetEnergy(g.state.SunEnergy)
		g.ambientLight.SetEnergy(1.0)
		log.Printf("Lights reset to defaults")
	}

	// G: Toggle GR enabled in state
	if inpututil.IsKeyJustPressed(ebiten.KeyG) {
		g.state.GrEnabled = !g.state.GrEnabled
		log.Printf("GR state: %v", g.state.GrEnabled)
	}

	// Update camera in scene
	g.scene3D.SetCameraPosition(g.state.CameraX, g.state.CameraY, g.state.CameraZ)
	g.scene3D.LookAt(g.state.LookAtX, g.state.LookAtY, g.state.LookAtZ)

	// Update rings if camera moves close to Saturn
	if g.saturnRings != nil {
		g.saturnRings.Update(dt)
	}

	// Handle screenshot - terminate only AFTER screenshot is taken in Draw()
	if g.screenshotTaken {
		return ebiten.Termination
	}

	return nil
}

// Draw implements ebiten.Game interface.
func (g *Game) Draw(screen *ebiten.Image) {
	// Clear render buffer
	g.renderBuffer.Clear()

	// Draw space background
	g.renderBuffer.Fill(color.RGBA{5, 5, 15, 255})

	// Render 3D scene
	img3d := g.scene3D.Render()
	g.renderBuffer.DrawImage(img3d, nil)

	// Apply shader effects
	useShaders := g.state.ShipVelocity > 0.05 || g.state.GrEnabled
	if useShaders {
		src := g.renderBuffer

		// Apply SR warp if ship is moving fast
		if g.srWarp.IsEnabled() && g.state.ShipVelocity >= 0.05 {
			intermediate := ebiten.NewImage(screenWidth, screenHeight)
			if g.srWarp.Apply(intermediate, src) {
				src = intermediate
			}
		}

		// Apply GR warp if enabled
		if g.grWarp.IsEnabled() && g.state.GrEnabled {
			intermediate := ebiten.NewImage(screenWidth, screenHeight)
			if g.grWarp.Apply(intermediate, src) {
				src = intermediate
			}
		}

		screen.DrawImage(src, nil)
	} else {
		// No effects - copy directly
		screen.DrawImage(g.renderBuffer, nil)
	}

	// Draw overlay UI
	g.drawOverlay(screen)

	// Take screenshot if requested
	if g.screenshotFrame > 0 && g.frameCount >= g.screenshotFrame && !g.screenshotTaken {
		g.saveScreenshot(screen)
		g.screenshotTaken = true
	}
}

// drawOverlay renders the debug overlay UI.
func (g *Game) drawOverlay(screen *ebiten.Image) {
	planets := sim_gen.GetSolarDemoPlanets()

	// SR/GR status
	srStatus := "OFF"
	grStatus := "OFF"
	if g.srWarp.IsEnabled() {
		srStatus = "ON"
	}
	if g.grWarp.IsEnabled() {
		grStatus = fmt.Sprintf("ON (%s)", g.grWarp.GetDemoIntensity())
	}

	// Calculate gamma (Lorentz factor)
	gamma := 1.0
	if g.velocity > 0 {
		gamma = 1.0 / math.Sqrt(1.0-g.velocity*g.velocity)
	}

	// Overlay info
	info := fmt.Sprintf(`AILANG Solar Demo
Tick: %d
Planets: %d (from AILANG)
Camera: (%.0f, %.0f, %.0f)

Lighting:
  Sun Energy: %.0f (x%.1f)
  Ambient:    %.2f

Relativistic Effects:
  Velocity: %.1f%% c
  Gamma:    %.2f
  SR Warp:  %s
  GR Warp:  %s

Controls:
  WASD/Arrows: Move | Q/E: Up/Down
  Shift: Fast | R: Reset position
  [ / ]: Sun light | ; / ': Ambient
  L: Reset lights
  1/2: SR/GR warp | 3: Cycle GR
  +/-/0: Velocity`,
		g.state.Tick,
		len(planets),
		g.state.CameraX, g.state.CameraY, g.state.CameraZ,
		g.state.SunEnergy, g.lightMultiplier,
		g.ambientLevel,
		g.velocity*100,
		gamma,
		srStatus,
		grStatus,
	)

	ebitenutil.DebugPrint(screen, info)
}

// saveScreenshot saves the current frame to a PNG file.
func (g *Game) saveScreenshot(screen *ebiten.Image) {
	path := g.screenshotPath
	if path == "" {
		path = "out/screenshots/ailang-solar.png"
	}

	f, err := os.Create(path)
	if err != nil {
		log.Printf("Failed to create screenshot: %v", err)
		return
	}
	defer f.Close()

	if err := png.Encode(f, screen); err != nil {
		log.Printf("Failed to encode screenshot: %v", err)
		return
	}

	log.Printf("Screenshot saved to %s", path)
}

// Layout implements ebiten.Game interface.
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

// rgbaFromInt converts a packed RGBA int to color.RGBA.
func rgbaFromInt(rgba int64) color.RGBA {
	return color.RGBA{
		R: uint8((rgba >> 24) & 0xFF),
		G: uint8((rgba >> 16) & 0xFF),
		B: uint8((rgba >> 8) & 0xFF),
		A: uint8(rgba & 0xFF),
	}
}

func main() {
	screenshotFrame := flag.Int("screenshot", 0, "Take screenshot at frame N and exit")
	screenshotPath := flag.String("output", "", "Screenshot output path")
	flag.Parse()

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("AILANG Solar System Demo")

	game := NewGame(*screenshotFrame, *screenshotPath)
	if err := ebiten.RunGame(game); err != nil && err != ebiten.Termination {
		log.Fatal(err)
	}
}
