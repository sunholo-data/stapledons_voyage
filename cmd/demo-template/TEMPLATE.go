// cmd/demo-YOURNAME/main.go
// DESCRIPTION: Brief description of what this demo tests/shows.
//
// Usage:
//   go build -o bin/demo-YOURNAME ./cmd/demo-YOURNAME && bin/demo-YOURNAME
//   bin/demo-YOURNAME --screenshot 60          # Screenshot after 60 frames
//   bin/demo-YOURNAME --output path/to/out.png # Custom output path
//
// Copy this template to create a new demo:
//   cp -r cmd/demo-template cmd/demo-YOURNAME
//   mv cmd/demo-YOURNAME/TEMPLATE.go cmd/demo-YOURNAME/main.go
//   # Edit main.go - replace YOURNAME, DESCRIPTION, and implement your demo
package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"stapledons_voyage/engine/assets"
	"stapledons_voyage/engine/display"
	"stapledons_voyage/engine/handlers"
	"stapledons_voyage/engine/render"
	"stapledons_voyage/engine/view/background"
	"stapledons_voyage/sim_gen"
)

// =============================================================================
// Standard Demo Flags (DO NOT REMOVE - required for CI/testing)
// =============================================================================

var (
	screenshotFrame = flag.Int("screenshot", 0, "Take screenshot after N frames (0=disabled)")
	outputPath      = flag.String("output", "out/screenshots/demo-YOURNAME.png", "Screenshot output path")
	// Add your custom flags here:
	// customFlag = flag.String("custom", "default", "Description")
)

// =============================================================================
// Demo Game Struct
// =============================================================================

// DemoGame is the main game struct for this demo.
type DemoGame struct {
	// Rendering
	renderer        *render.Renderer
	spaceBackground *background.SpaceBackground

	// Animation state
	time       float64
	frameCount int

	// Screenshot tracking (DO NOT REMOVE)
	screenshotTaken bool

	// Add your demo-specific state here:
	// myState *MyState
}

// =============================================================================
// Initialization
// =============================================================================

func NewDemoGame() *DemoGame {
	screenW := display.InternalWidth
	screenH := display.InternalHeight

	g := &DemoGame{}

	// Initialize AILANG handlers (required for any AILANG calls)
	sim_gen.Init(sim_gen.Handlers{
		Debug: sim_gen.NewDebugContext(),
		Rand:  &handlers.DefaultRandHandler{},
		Clock: &handlers.EbitenClockHandler{},
		AI:    handlers.NewStubAIHandler(),
	})

	// Create renderer
	assetMgr, err := assets.NewManager("assets")
	if err != nil {
		log.Printf("Warning: could not load assets: %v", err)
	}
	g.renderer = render.NewRenderer(assetMgr)

	// Create space background (optional - remove if not needed)
	g.spaceBackground = background.NewSpaceBackground(screenW, screenH)

	// Initialize your demo-specific state here:
	// g.myState = InitMyState()

	return g
}

// =============================================================================
// Game Loop
// =============================================================================

func (g *DemoGame) Update() error {
	dt := 1.0 / 60.0
	g.time += dt
	g.frameCount++

	// Update your demo state here:
	// g.myState = UpdateMyState(g.myState, dt)

	return nil
}

func (g *DemoGame) Draw(screen *ebiten.Image) {
	// Layer 1: Background
	if g.spaceBackground != nil {
		g.spaceBackground.Draw(screen, nil)
	} else {
		screen.Fill(color.RGBA{10, 10, 20, 255})
	}

	// Layer 2: Your demo content
	// Option A: Use AILANG DrawCmds via renderer
	// drawCmds := sim_gen.YourRenderFunction(g.myState)
	// frameOut := sim_gen.FrameOutput{
	// 	Draw:   drawCmds,
	// 	Camera: &sim_gen.Camera{X: 0, Y: 0, Zoom: 1.0},
	// }
	// g.renderer.RenderFrame(screen, frameOut)

	// Option B: Direct Ebiten drawing
	// ebitenutil.DrawRect(screen, 100, 100, 50, 50, color.White)

	// Layer 3: HUD (debug info)
	g.drawHUD(screen)

	// Screenshot handling (DO NOT REMOVE)
	if *screenshotFrame > 0 && g.frameCount >= *screenshotFrame && !g.screenshotTaken {
		g.takeScreenshot(screen)
		g.screenshotTaken = true
	}
}

// =============================================================================
// HUD Drawing
// =============================================================================

func (g *DemoGame) drawHUD(screen *ebiten.Image) {
	y := 20.0
	lineHeight := 18.0

	// Demo title
	ebitenutil.DebugPrintAt(screen, "Demo: YOURNAME", 10, int(y))
	y += lineHeight

	// Standard stats
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Frame: %d", g.frameCount), 10, int(y))
	y += lineHeight

	fps := ebiten.ActualFPS()
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("FPS: %.1f", fps), 10, int(y))
	y += lineHeight * 2

	// Your demo-specific HUD here:
	// ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Custom: %v", g.myValue), 10, int(y))

	// Bottom help text
	y = float64(display.InternalHeight) - 40
	ebitenutil.DebugPrintAt(screen, "DESCRIPTION of what this demo shows", 10, int(y))
	y += lineHeight
	if *screenshotFrame > 0 {
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Screenshot at frame %d -> %s", *screenshotFrame, *outputPath), 10, int(y))
	}
}

// =============================================================================
// Screenshot (DO NOT MODIFY)
// =============================================================================

func (g *DemoGame) takeScreenshot(screen *ebiten.Image) {
	if err := os.MkdirAll("out/screenshots", 0755); err != nil {
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

// =============================================================================
// Layout
// =============================================================================

func (g *DemoGame) Layout(outsideWidth, outsideHeight int) (int, int) {
	return display.InternalWidth, display.InternalHeight
}

// =============================================================================
// Main
// =============================================================================

func main() {
	flag.Parse()

	// Print demo info
	fmt.Println("Demo: YOURNAME")
	fmt.Println("==============")
	fmt.Println("DESCRIPTION")
	fmt.Println()
	fmt.Printf("Resolution: %dx%d\n", display.InternalWidth, display.InternalHeight)
	if *screenshotFrame > 0 {
		fmt.Printf("Screenshot at frame: %d -> %s\n", *screenshotFrame, *outputPath)
	}
	fmt.Println()

	// Window setup
	ebiten.SetWindowSize(display.InternalWidth, display.InternalHeight)
	ebiten.SetWindowTitle("Stapledon's Voyage - Demo: YOURNAME")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	// Run
	game := NewDemoGame()
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
