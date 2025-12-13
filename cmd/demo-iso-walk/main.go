// cmd/demo-iso-walk/main.go
// Simple isometric walk demo - tests proper 64x32 tiles with screen-aligned WASD.
//
// Usage:
//   go build -o bin/demo-iso-walk ./cmd/demo-iso-walk && bin/demo-iso-walk
//   bin/demo-iso-walk --screenshot 60          # Screenshot after 60 frames
//   bin/demo-iso-walk --output path/to/out.png # Custom output path
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
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"stapledons_voyage/engine/assets"
	"stapledons_voyage/engine/display"
	"stapledons_voyage/engine/handlers"
	"stapledons_voyage/engine/render"
	"stapledons_voyage/sim_gen"
)

var (
	screenshotFrame = flag.Int("screenshot", 0, "Take screenshot after N frames (0=disabled)")
	outputPath      = flag.String("output", "out/screenshots/demo-iso-walk.png", "Screenshot output path")
	debug           = flag.Bool("debug", false, "Show debug information")
)

type DemoGame struct {
	renderer   *render.Renderer
	state      *sim_gen.IsoWalkState
	frameCount int

	// Screenshot tracking
	screenshotTaken bool
}

func NewDemoGame() *DemoGame {
	g := &DemoGame{}

	// Initialize AILANG handlers
	sim_gen.Init(sim_gen.Handlers{
		Debug: sim_gen.NewDebugContext(),
		Rand:  &handlers.DefaultRandHandler{},
		Clock: &handlers.EbitenClockHandler{},
		AI:    handlers.NewStubAIHandler(),
	})

	// Create renderer with assets
	assetMgr, err := assets.NewManager("assets")
	if err != nil {
		log.Printf("Warning: could not load assets: %v", err)
	}
	g.renderer = render.NewRenderer(assetMgr)

	// Enable layered rendering for parallax support
	g.renderer.EnableLayers(display.InternalWidth, display.InternalHeight)

	// Initialize AILANG state
	g.state = sim_gen.InitIsoDemo()

	return g
}

func (g *DemoGame) Update() error {
	g.frameCount++

	// Capture input for AILANG
	input := g.captureInput()

	// Step AILANG simulation
	g.state = sim_gen.StepIsoDemo(g.state, input)

	return nil
}

// captureInput builds FrameInput from keyboard state
func (g *DemoGame) captureInput() *sim_gen.FrameInput {
	var keys []*sim_gen.KeyEvent

	// Check movement keys (WASD and arrows)
	movementKeys := []ebiten.Key{
		ebiten.KeyW, ebiten.KeyA, ebiten.KeyS, ebiten.KeyD,
		ebiten.KeyUp, ebiten.KeyDown, ebiten.KeyLeft, ebiten.KeyRight,
	}

	for _, key := range movementKeys {
		if inpututil.IsKeyJustPressed(key) {
			keys = append(keys, &sim_gen.KeyEvent{
				Key:  int64(key),
				Kind: "pressed",
			})
		}
	}

	return &sim_gen.FrameInput{
		Mouse:            &sim_gen.MouseState{X: 0, Y: 0, Buttons: []int64{}},
		Keys:             keys,
		ClickedThisFrame: false,
		WorldMouseX:      0,
		WorldMouseY:      0,
		TileMouseX:       0,
		TileMouseY:       0,
		ActionRequested:  &sim_gen.PlayerAction{Kind: sim_gen.PlayerActionKindActionNone},
		TestMode:         false,
	}
}

func (g *DemoGame) Draw(screen *ebiten.Image) {
	// Dark background
	screen.Fill(color.RGBA{15, 20, 30, 255})

	// Get DrawCmds from AILANG
	cmds := sim_gen.RenderIsoDemo(g.state)

	// Get camera position from AILANG (smooth follow)
	cam := sim_gen.GetIsoCamera(g.state)

	// Render via engine (uses layered rendering with parallax)
	out := sim_gen.FrameOutput{
		Draw:   cmds,
		Camera: cam,
	}
	g.renderer.RenderFrame(screen, out)

	// HUD
	g.drawHUD(screen)

	// Screenshot handling
	if *screenshotFrame > 0 && g.frameCount >= *screenshotFrame && !g.screenshotTaken {
		g.takeScreenshot(screen)
		g.screenshotTaken = true
	}
}

func (g *DemoGame) drawHUD(screen *ebiten.Image) {
	y := 20.0
	lineHeight := 18.0

	// Demo title
	ebitenutil.DebugPrintAt(screen, "Demo: Isometric Walk", 10, int(y))
	y += lineHeight

	// FPS
	fps := ebiten.ActualFPS()
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("FPS: %.1f", fps), 10, int(y))
	y += lineHeight

	// Frame count
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Frame: %d", g.frameCount), 10, int(y))
	y += lineHeight

	// Player position
	if g.state != nil {
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Player: (%d, %d)", g.state.PlayerX, g.state.PlayerY), 10, int(y))
		y += lineHeight
	}

	if *debug {
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Grid: %dx%d", g.state.GridWidth, g.state.GridHeight), 10, int(y))
		y += lineHeight
	}

	// Controls at bottom
	y = float64(display.InternalHeight) - 60
	ebitenutil.DebugPrintAt(screen, "Controls: W=Up  S=Down  A=Left  D=Right", 10, int(y))
	y += lineHeight
	ebitenutil.DebugPrintAt(screen, "Testing: Proper 64x32 isometric tiles with screen-aligned movement", 10, int(y))
	y += lineHeight
	if *screenshotFrame > 0 {
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Screenshot at frame %d -> %s", *screenshotFrame, *outputPath), 10, int(y))
	}
}

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

func (g *DemoGame) Layout(outsideWidth, outsideHeight int) (int, int) {
	return display.InternalWidth, display.InternalHeight
}

func main() {
	flag.Parse()

	fmt.Println("Demo: Isometric Walk")
	fmt.Println("====================")
	fmt.Println("Tests proper 64x32 isometric tiles with screen-aligned WASD movement")
	fmt.Println()
	fmt.Printf("Resolution: %dx%d\n", display.InternalWidth, display.InternalHeight)
	if *screenshotFrame > 0 {
		fmt.Printf("Screenshot at frame: %d -> %s\n", *screenshotFrame, *outputPath)
	}
	fmt.Println()

	// Window setup
	ebiten.SetWindowSize(display.InternalWidth, display.InternalHeight)
	ebiten.SetWindowTitle("Stapledon's Voyage - Isometric Walk Demo")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	// Run
	game := NewDemoGame()
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
