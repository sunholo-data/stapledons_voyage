// Package main provides a demo for the 3D interior rendering system.
// This demonstrates AILANG-driven 3D interior navigation.
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
	renderer   *render.Renderer
	interior   *sim_gen.InteriorState
	frameCount int
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

	// Create renderer with assets (or nil for fallback colors)
	assetMgr, err := assets.NewManager("assets")
	if err != nil {
		log.Printf("Warning: could not load assets: %v", err)
		// Continue with nil - renderer will use fallback colors
		assetMgr = nil
	}

	return &Game{
		renderer: render.NewRenderer(assetMgr),
		interior: interior,
	}
}

// captureInput creates a FrameInput from current Ebiten state.
func captureInput() *sim_gen.FrameInput {
	x, y := ebiten.CursorPosition()

	var buttons []int64
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		buttons = append(buttons, 0)
	}

	var keys []*sim_gen.KeyEvent
	for _, k := range inpututil.AppendJustPressedKeys(nil) {
		keys = append(keys, &sim_gen.KeyEvent{
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
		Mouse: &sim_gen.MouseState{
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
	drawCmds := sim_gen.RenderInterior(g.interior)

	// Create frame output for renderer
	output := sim_gen.FrameOutput{
		Draw:   drawCmds,
		Camera: &sim_gen.Camera{X: 0, Y: 0, Zoom: 1.0},
	}

	// Render
	g.renderer.RenderFrame(screen, output)

	// HUD overlay
	player := g.interior.Player
	hudText := fmt.Sprintf(
		"3D Interior Demo (AILANG)\n"+
			"WASD: Move | Mouse: Look | Shift: Run | Esc: Quit\n"+
			"Position: (%.1f, %.1f, %.1f)\n"+
			"Yaw: %.1f° | Pitch: %.1f°\n"+
			"Frame: %d",
		player.Pos.X, player.Pos.Y, player.Pos.Z,
		player.Yaw*180/3.14159, player.Pitch*180/3.14159,
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
	ebiten.SetWindowTitle("3D Interior Demo (AILANG)")
	ebiten.SetCursorMode(ebiten.CursorModeCaptured)

	game := NewGame()

	if err := ebiten.RunGame(game); err != nil && err != ebiten.Termination {
		log.Fatal(err)
	}
}
