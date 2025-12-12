// cmd/demo-parallax/main.go
// Demo for the depth layer parallax system.
// Usage:
//   go run ./cmd/demo-parallax
//   go run ./cmd/demo-parallax --screenshot 30 --output out/screenshots/parallax.png
//   Arrow keys: Pan camera
//   +/-: Zoom
package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"stapledons_voyage/engine/demo"
	"stapledons_voyage/engine/display"
	"stapledons_voyage/engine/render"
	"stapledons_voyage/sim_gen"
)

var (
	startCamX = flag.Float64("camx", 0, "Starting camera X position")
	startCamY = flag.Float64("camy", 0, "Starting camera Y position")
)

const (
	screenW = 1280
	screenH = 960
)

type DemoGame struct {
	renderer *render.Renderer
	camX     float64
	camY     float64
	zoom     float64
	frame    int
}

func NewDemoGame() *DemoGame {
	r := render.NewRenderer(nil)
	r.EnableLayers(screenW, screenH)

	return &DemoGame{
		renderer: r,
		camX:     *startCamX,
		camY:     *startCamY,
		zoom:     1.0,
	}
}

func (g *DemoGame) Update() error {
	g.frame++

	// Camera controls
	panSpeed := 5.0 / g.zoom
	if ebiten.IsKeyPressed(ebiten.KeyLeft) || ebiten.IsKeyPressed(ebiten.KeyA) {
		g.camX -= panSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyRight) || ebiten.IsKeyPressed(ebiten.KeyD) {
		g.camX += panSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyUp) || ebiten.IsKeyPressed(ebiten.KeyW) {
		g.camY -= panSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyDown) || ebiten.IsKeyPressed(ebiten.KeyS) {
		g.camY += panSpeed
	}

	// Zoom controls
	if inpututil.IsKeyJustPressed(ebiten.KeyEqual) || inpututil.IsKeyJustPressed(ebiten.KeyKPAdd) {
		g.zoom *= 1.2
		if g.zoom > 4.0 {
			g.zoom = 4.0
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyMinus) || inpututil.IsKeyJustPressed(ebiten.KeyKPSubtract) {
		g.zoom /= 1.2
		if g.zoom < 0.25 {
			g.zoom = 0.25
		}
	}

	return nil
}

func (g *DemoGame) Draw(screen *ebiten.Image) {
	// Build FrameOutput with all layer types
	output := g.buildFrameOutput()

	// Render using the layer system
	g.renderer.RenderFrame(screen, output)

	// Draw HUD on top
	g.drawHUD(screen)
}

func (g *DemoGame) buildFrameOutput() sim_gen.FrameOutput {
	var cmds []*sim_gen.DrawCmd

	// Layer 0: Deep Background - Galaxy/Space (0.0x parallax - FIXED)
	cmds = append(cmds, sim_gen.NewDrawCmdGalaxyBg(0.8, 0, false, 0, 0, 90))

	// Layer 6: Mid Background - Spire (0.3x parallax)
	cmds = append(cmds, sim_gen.NewDrawCmdSpireBg(0))

	// Add depth markers at layers 5, 10, and 15 to visualize parallax
	// These are colored bars that move at different rates

	// Layer 5 (0.25x parallax): Far deck - RED marker
	cmds = append(cmds, sim_gen.NewDrawCmdMarker(
		100, 200, // x, y (screen position)
		80, 400,  // w, h
		0xFF0000C0, // red with alpha
		5,          // parallaxLayer
		0,          // z
	))

	// Layer 10 (0.70x parallax): Adjacent deck - GREEN marker
	cmds = append(cmds, sim_gen.NewDrawCmdMarker(
		200, 200,
		80, 400,
		0x00FF00C0, // green with alpha
		10,         // parallaxLayer
		0,
	))

	// Layer 15 (0.95x parallax): Current deck background - BLUE marker
	cmds = append(cmds, sim_gen.NewDrawCmdMarker(
		300, 200,
		80, 400,
		0x0080FFC0, // blue with alpha
		15,         // parallaxLayer
		0,
	))

	// Layer 16: Scene - Isometric floor tiles (1.0x parallax)
	// Create a smaller grid to make room for depth markers
	for tx := int64(-3); tx <= 3; tx++ {
		for ty := int64(-3); ty <= 3; ty++ {
			// Checkerboard pattern: some tiles transparent
			if (tx+ty)%2 == 0 {
				// Solid tile
				cmds = append(cmds, sim_gen.NewDrawCmdIsoTile(
					&sim_gen.Coord{X: tx, Y: ty},
					0,    // height
					1000, // spriteId (floor)
					0,    // layer (iso depth layer, not parallax)
					0,    // color
				))
			} else {
				// Transparent glass tile - shows deep background through it
				cmds = append(cmds, sim_gen.NewDrawCmdIsoTileAlpha(
					&sim_gen.Coord{X: tx, Y: ty},
					0,          // height
					1004,       // spriteId (glass)
					0,          // layer
					0.3,        // alpha
					0x4080C0A0, // blue tint
				))
			}
		}
	}

	// Add player entity on scene layer
	cmds = append(cmds, sim_gen.NewDrawCmdIsoEntity(
		"player",
		&sim_gen.Coord{X: 0, Y: 0},
		0, 0, // offset
		0,    // height
		1205, // player sprite
		1,    // layer
	))

	return sim_gen.FrameOutput{
		Draw:   cmds,
		Debug:  []string{},
		Sounds: []int64{},
		Camera: &sim_gen.Camera{
			X:    g.camX,
			Y:    g.camY,
			Zoom: g.zoom,
		},
	}
}

func (g *DemoGame) drawHUD(screen *ebiten.Image) {
	y := 10
	ebitenutil.DebugPrintAt(screen, "20-LAYER PARALLAX DEMO", 10, y)
	y += 16
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Camera: (%.1f, %.1f) Zoom: %.2f", g.camX, g.camY, g.zoom), 10, y)
	y += 16
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("FPS: %.1f", ebiten.ActualFPS()), 10, y)
	y += 24

	// Layer info with markers
	ebitenutil.DebugPrintAt(screen, "DEPTH MARKERS (colored bars):", 10, y)
	y += 16
	ebitenutil.DebugPrintAt(screen, "  L0  (0.00x) - Galaxy (fixed)", 10, y)
	y += 14
	ebitenutil.DebugPrintAt(screen, "  L5  (0.25x) - RED bar", 10, y)
	y += 14
	ebitenutil.DebugPrintAt(screen, "  L6  (0.30x) - Spire", 10, y)
	y += 14
	ebitenutil.DebugPrintAt(screen, "  L10 (0.70x) - GREEN bar", 10, y)
	y += 14
	ebitenutil.DebugPrintAt(screen, "  L15 (0.95x) - BLUE bar", 10, y)
	y += 14
	ebitenutil.DebugPrintAt(screen, "  L16 (1.00x) - Tiles (scene)", 10, y)
	y += 24

	// Controls
	ebitenutil.DebugPrintAt(screen, "Arrow/WASD: Pan | +/-: Zoom", 10, y)
	y += 24

	// Parallax visualization
	ebitenutil.DebugPrintAt(screen, "Pan to see parallax rates!", 10, y)
}

func (g *DemoGame) Layout(outsideWidth, outsideHeight int) (int, int) {
	return display.InternalWidth, display.InternalHeight
}

func main() {
	// Parse flags before demo.Run so we can use camx/camy
	flag.Parse()

	fmt.Println("20-Layer Parallax Demo")
	fmt.Println("======================")
	fmt.Println("Demonstrating selectable depth layers:")
	fmt.Println("  L0  (0.00x): Galaxy - fixed at infinity")
	fmt.Println("  L5  (0.25x): RED marker - far deck")
	fmt.Println("  L6  (0.30x): Spire - mid distance")
	fmt.Println("  L10 (0.70x): GREEN marker - adjacent deck")
	fmt.Println("  L15 (0.95x): BLUE marker - near")
	fmt.Println("  L16 (1.00x): Tiles - main scene")
	fmt.Println("  - Foreground (1.0x): UI overlay")
	fmt.Println()
	fmt.Println("Controls: Arrow keys to pan, +/- to zoom")
	fmt.Println()

	game := NewDemoGame()
	if err := demo.Run(game, demo.Config{Title: "Stapledon's Voyage - Parallax Demo"}); err != nil {
		log.Fatal(err)
	}
}
