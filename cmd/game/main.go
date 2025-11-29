package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"stapledons_voyage/engine/assets"
	"stapledons_voyage/engine/display"
	"stapledons_voyage/engine/render"
	"stapledons_voyage/sim_gen"
)

type Game struct {
	world    sim_gen.World
	out      sim_gen.FrameOutput
	renderer *render.Renderer
	display  *display.Manager
}

func (g *Game) Update() error {
	// Handle display input (F11 for fullscreen)
	g.display.HandleInput()

	// Capture game input with camera for screen-to-world conversion
	// Uses internal resolution (640x480) for coordinate conversion
	input := render.CaptureInputWithCamera(g.out.Camera, display.InternalWidth, display.InternalHeight)
	w2, out, err := sim_gen.Step(g.world, input)
	if err != nil {
		return err
	}
	g.world = w2
	g.out = out
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.renderer.RenderFrame(screen, g.out)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return g.display.Layout(outsideWidth, outsideHeight)
}

func main() {
	// Initialize display manager (loads config from file)
	displayMgr := display.NewManager("config.json")

	// Initialize asset manager
	assetMgr, err := assets.NewManager("assets")
	if err != nil {
		log.Printf("Warning: failed to initialize assets: %v", err)
	}

	// Create renderer with asset manager
	renderer := render.NewRenderer(assetMgr)

	// Initialize world
	world := sim_gen.InitWorld(1234)

	game := &Game{
		world:    world,
		renderer: renderer,
		display:  displayMgr,
	}

	ebiten.SetWindowTitle("Stapledons Voyage")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
