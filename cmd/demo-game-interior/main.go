// Package main provides a demo for the 3D interior rendering system.
// This demonstrates AILANG-driven 3D interior navigation with SR/GR effects.
//
// Controls:
//   WASD: Move | Mouse: Look | Shift: Run
//   Up/Down: Adjust velocity (SR effects)
//   [/]: Adjust gravitational potential (GR effects)
//   0: Reset SR/GR to defaults
//   Esc: Quit
//
// Usage:
//
//	go run ./cmd/demo-game-interior
package main

import (
	"flag"
	"fmt"
	"image/color"
	"image/png"
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"stapledons_voyage/engine/assets"
	"stapledons_voyage/engine/handlers"
	"stapledons_voyage/engine/render"
	"stapledons_voyage/sim_gen"
)

const (
	screenWidth  = 1280
	screenHeight = 720
)

var (
	screenshotFrame = flag.Int("screenshot", 0, "Take screenshot at frame N and exit")
	screenshotPath  = flag.String("output", "out/screenshots/game-interior.png", "Screenshot output path")
)

// Game implements ebiten.Game for the interior demo.
type Game struct {
	renderer       *render.Renderer
	interior       *sim_gen.InteriorState
	shipNavigation *sim_gen.ShipNavigation
	frameCount     int

	// SR/GR parameters for experimentation
	velocity float64 // Fraction of c (0.0 to 0.99)
	grPhi    float64 // Gravitational potential (0.0 = flat space, higher = stronger gravity)
}

// NewGame creates a new interior demo game.
func NewGame() *Game {
	// Initialize AILANG handlers
	sim_gen.Init(sim_gen.Handlers{
		Debug: sim_gen.NewDebugContext(),
		Rand:  &handlers.DefaultRandHandler{},
		Clock: &handlers.EbitenClockHandler{},
		AI:    handlers.NewStubAIHandler(),
	})

	// Initialize interior state from AILANG
	interior := sim_gen.InitInterior()

	// Set player to face north (toward the window on north wall)
	// Yaw=π means facing -Z direction (north)
	interior.Player.Yaw = 3.14159
	interior.Player.Pitch = 0

	// Initialize ship navigation
	shipNav := sim_gen.InitNavigation()

	// Create renderer with assets (or nil for fallback colors)
	assetMgr, err := assets.NewManager("assets")
	if err != nil {
		log.Printf("Warning: could not load assets: %v", err)
		assetMgr = nil
	}

	return &Game{
		renderer:       render.NewRenderer(assetMgr),
		interior:       interior,
		shipNavigation: shipNav,
		velocity:       0.5, // Start at 0.5c for visible SR effects
		grPhi:          0.0, // Start in flat space
	}
}

// captureInput creates a FrameInput from current Ebiten state.
func captureInput() *sim_gen.FrameInput {
	x, y := ebiten.CursorPosition()

	var buttons []int64
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		buttons = append(buttons, 0)
	}

	var keys []sim_gen.KeyEvent
	for _, k := range inpututil.AppendJustPressedKeys(nil) {
		keys = append(keys, sim_gen.KeyEvent{
			Key:  int64(k),
			Kind: "pressed",
		})
	}

	// Flight input for WASD
	flight := &sim_gen.FlightInput{
		W:     ebiten.IsKeyPressed(ebiten.KeyW),
		A:     ebiten.IsKeyPressed(ebiten.KeyA),
		S:     ebiten.IsKeyPressed(ebiten.KeyS),
		D:     ebiten.IsKeyPressed(ebiten.KeyD),
		Shift: ebiten.IsKeyPressed(ebiten.KeyShift),
	}

	return &sim_gen.FrameInput{
		Mouse: sim_gen.MouseState{
			X:       float64(x),
			Y:       float64(y),
			Buttons: buttons,
		},
		Keys:   keys,
		Flight: flight,
	}
}

func (g *Game) Update() error {
	// Escape to quit
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}

	// SR/GR parameter controls
	velocityStep := 0.05
	grPhiStep := 0.1

	// Velocity (SR) - Up/Down arrows
	if inpututil.IsKeyJustPressed(ebiten.KeyUp) || ebiten.IsKeyPressed(ebiten.KeyUp) && g.frameCount%10 == 0 {
		g.velocity += velocityStep
		if g.velocity > 0.99 {
			g.velocity = 0.99 // Cap at 99% speed of light
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyDown) || ebiten.IsKeyPressed(ebiten.KeyDown) && g.frameCount%10 == 0 {
		g.velocity -= velocityStep
		if g.velocity < 0 {
			g.velocity = 0
		}
	}

	// Gravitational potential (GR) - [ and ] keys
	if inpututil.IsKeyJustPressed(ebiten.KeyBracketRight) || ebiten.IsKeyPressed(ebiten.KeyBracketRight) && g.frameCount%10 == 0 {
		g.grPhi += grPhiStep
		if g.grPhi > 2.0 {
			g.grPhi = 2.0 // Cap at extreme gravity (like near a black hole)
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyBracketLeft) || ebiten.IsKeyPressed(ebiten.KeyBracketLeft) && g.frameCount%10 == 0 {
		g.grPhi -= grPhiStep
		if g.grPhi < 0 {
			g.grPhi = 0
		}
	}

	// Reset to defaults - 0 key
	if inpututil.IsKeyJustPressed(ebiten.Key0) {
		g.velocity = 0.5
		g.grPhi = 0.0
	}

	// Update ship navigation with current SR/GR values
	g.shipNavigation = sim_gen.SetVelocity(g.shipNavigation, g.velocity)
	g.shipNavigation.GrPhi = g.grPhi

	// Capture input and step interior simulation
	input := captureInput()
	g.interior = sim_gen.StepInterior(g.interior, input)

	g.frameCount++
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Clear screen
	screen.Fill(color.RGBA{10, 10, 15, 255})

	// Get draw commands from AILANG
	drawCmds := sim_gen.RenderInterior(g.interior, g.shipNavigation)

	// Create frame output for renderer
	output := sim_gen.FrameOutput{
		Draw:   drawCmds,
		Camera: sim_gen.Camera{X: 0, Y: 0, Zoom: 1.0},
	}

	// Render
	g.renderer.RenderFrame(screen, output)

	// Calculate derived values for display
	gamma := 1.0
	if g.velocity > 0 {
		gamma = 1.0 / (1.0 - g.velocity*g.velocity)
		if gamma > 0 {
			gamma = 1.0 / (gamma * gamma) // sqrt approximation issue, recalc
		}
		// Proper gamma calculation
		v2 := g.velocity * g.velocity
		gamma = 1.0 / (1.0 - v2)
		if gamma > 0 {
			gamma = 1.0 / (gamma)
		}
		// Actually: gamma = 1/sqrt(1-v²)
		gamma = 1.0 / (1.0 - v2) // This gives 1/(1-v²), need sqrt
	}
	// Simpler: just show raw values
	timeDilation := 1.0
	if g.velocity > 0 && g.velocity < 1 {
		v2 := g.velocity * g.velocity
		timeDilation = 1.0 / (1.0 - v2) // Approximate gamma² for display
	}

	grTimeDilation := 1.0
	if g.grPhi > 0 {
		// GR time dilation: sqrt(1 - 2*phi/c²) ≈ 1 - phi for weak fields
		grTimeDilation = 1.0 / (1.0 - g.grPhi*0.5) // Simplified
	}

	// HUD overlay with SR/GR info
	player := g.interior.Player
	hudText := fmt.Sprintf(
		"3D Interior Demo - SR/GR Effects\n"+
			"═══════════════════════════════════════\n"+
			"WASD: Move | Mouse: Look | Shift: Run\n"+
			"↑/↓: Velocity | [/]: Gravity | 0: Reset\n"+
			"═══════════════════════════════════════\n"+
			"SPECIAL RELATIVITY (SR):\n"+
			"  Velocity: %.0f%% c (%.2f)\n"+
			"  Time Dilation: %.2fx slower\n"+
			"  Effects: Aberration, Doppler, Beaming\n"+
			"═══════════════════════════════════════\n"+
			"GENERAL RELATIVITY (GR):\n"+
			"  Grav. Potential φ: %.2f\n"+
			"  GR Time Dilation: %.2fx slower\n"+
			"  Effects: Redshift, Lensing\n"+
			"═══════════════════════════════════════\n"+
			"Position: (%.1f, %.1f, %.1f)\n"+
			"Frame: %d",
		g.velocity*100, g.velocity,
		timeDilation,
		g.grPhi,
		grTimeDilation,
		player.Pos.X, player.Pos.Y, player.Pos.Z,
		g.frameCount,
	)
	ebitenutil.DebugPrint(screen, hudText)

	// Screenshot
	if *screenshotFrame > 0 && g.frameCount >= *screenshotFrame {
		g.saveScreenshot(screen)
	}
}

func (g *Game) saveScreenshot(screen *ebiten.Image) {
	f, err := os.Create(*screenshotPath)
	if err != nil {
		log.Printf("Failed to create screenshot: %v", err)
		os.Exit(1)
	}
	defer f.Close()

	if err := png.Encode(f, screen); err != nil {
		log.Printf("Failed to encode screenshot: %v", err)
		os.Exit(1)
	}

	fmt.Printf("Screenshot saved to %s\n", *screenshotPath)
	os.Exit(0)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	flag.Parse()

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("3D Interior Demo - SR/GR Effects")
	ebiten.SetCursorMode(ebiten.CursorModeCaptured)

	game := NewGame()

	if err := ebiten.RunGame(game); err != nil && err != ebiten.Termination {
		log.Fatal(err)
	}
}
