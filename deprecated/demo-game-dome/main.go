// AILANG Dome Demo - Player Movement with Solar System View
//
// This demo showcases the dual coordinate system architecture:
// - Player movement: Ship-local coordinates in meters (WASD keys)
// - Solar system: Galactic coordinates in AU (reused from solar_demo.ail)
//
// Controls:
//   WASD - Move player around the dome floor
//   ESC  - Exit demo
//
// Implementation:
//   All game logic in sim/dome_demo.ail (AILANG)
//   Go engine only captures input and renders DrawCmds
//
// Build: make demo-game-dome
// Run:   ./bin/demo-game-dome
//       ./bin/demo-game-dome --screenshot 60 --output out/screenshots/dome.png
package main

import (
	"flag"
	"image"
	"image/png"
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"stapledons_voyage/engine/display"
	"stapledons_voyage/engine/handlers"
	"stapledons_voyage/engine/render"
	"stapledons_voyage/sim_gen"
)

var (
	screenshotFrame = flag.Int("screenshot", 0, "Take screenshot after N frames (0=disabled)")
	outputPath      = flag.String("output", "out/screenshots/dome-demo.png", "Screenshot output path")
)

// Game holds the dome demo state
type Game struct {
	state           *sim_gen.DomeState
	frameCount      int
	screenshotTaken bool
}

func main() {
	flag.Parse()

	// Task 3.2: Initialize effect handlers BEFORE calling AILANG code
	sim_gen.Init(sim_gen.Handlers{
		Debug: sim_gen.NewDebugContext(),
		Rand:  handlers.NewDefaultRandHandler(),
		Clock: handlers.NewEbitenClockHandler(),
		AI:    handlers.NewStubAIHandler(),
	})

	// Initialize dome demo state
	state := sim_gen.InitDomeDemo()

	// Create game instance
	game := &Game{state: state}

	// Set up window
	ebiten.SetWindowSize(display.InternalWidth, display.InternalHeight)
	ebiten.SetWindowTitle("AILANG Dome Demo - Player Movement")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	// Run game loop
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

// Task 3.3: Implement ebiten.Game interface
func (g *Game) Update() error {
	// Task 3.4: Capture input
	input := captureInput()
	if input == nil {
		return ebiten.Termination
	}

	// Call AILANG step function
	result := sim_gen.StepDomeDemo(g.state, input)

	// Extract tuple (newState, output)
	tuple, ok := result.([]interface{})
	if !ok || len(tuple) != 2 {
		log.Printf("Unexpected StepDomeDemo result type")
		return nil
	}
	if state, ok := tuple[0].(*sim_gen.DomeState); ok {
		g.state = state
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Call step to get latest output
	input := captureInput()
	if input == nil {
		// Use empty input if captureInput returns nil
		input = &sim_gen.FrameInput{
			Mouse:  &sim_gen.MouseState{X: 0, Y: 0, Buttons: []int64{}},
			Keys:   []*sim_gen.KeyEvent{},
			Flight: &sim_gen.FlightInput{},
		}
	}

	result := sim_gen.StepDomeDemo(g.state, input)
	tuple, ok := result.([]interface{})
	if !ok || len(tuple) != 2 {
		return
	}

	// Extract output from tuple
	if output, ok := tuple[1].(*sim_gen.FrameOutput); ok {
		// Task 3.5: Render DrawCmds
		render.RenderFrame(screen, *output)
	}

	// Screenshot
	g.frameCount++
	if *screenshotFrame > 0 && g.frameCount >= *screenshotFrame && !g.screenshotTaken {
		g.takeScreenshot(screen)
		g.screenshotTaken = true
	}
}

func (g *Game) takeScreenshot(screen *ebiten.Image) {
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

	// Write to file
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

	// Exit after screenshot
	os.Exit(0)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return display.InternalWidth, display.InternalHeight
}

// Task 3.4: Input capture
func captureInput() *sim_gen.FrameInput {
	var keys []*sim_gen.KeyEvent

	// Check WASD keys (W=87, A=65, S=83, D=68, Shift=340)
	if ebiten.IsKeyPressed(ebiten.KeyW) {
		keys = append(keys, &sim_gen.KeyEvent{Key: 87, Kind: "press"})
	}
	if ebiten.IsKeyPressed(ebiten.KeyA) {
		keys = append(keys, &sim_gen.KeyEvent{Key: 65, Kind: "press"})
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) {
		keys = append(keys, &sim_gen.KeyEvent{Key: 83, Kind: "press"})
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) {
		keys = append(keys, &sim_gen.KeyEvent{Key: 68, Kind: "press"})
	}
	if ebiten.IsKeyPressed(ebiten.KeyShift) {
		keys = append(keys, &sim_gen.KeyEvent{Key: 340, Kind: "press"})
	}

	// Exit on Escape
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		log.Println("Escape pressed, exiting")
		return nil
	}

	// Mouse state (not used in this demo but required for FrameInput)
	mx, my := ebiten.CursorPosition()
	var buttons []int64
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		buttons = append(buttons, 0) // Left button
	}
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) {
		buttons = append(buttons, 1) // Right button
	}

	return &sim_gen.FrameInput{
		Mouse: &sim_gen.MouseState{
			X:       float64(mx),
			Y:       float64(my),
			Buttons: buttons,
		},
		Keys:   keys,
		Flight: &sim_gen.FlightInput{},
	}
}
