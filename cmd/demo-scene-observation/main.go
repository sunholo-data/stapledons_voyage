// Demo: Scene Compositing with Observation Deck
//
// Demonstrates the window mask compositing system with the observation-v2 deck.
// Shows how deck backgrounds are composited with 3D space views through transparent windows.
//
// Controls:
//   - Arrow keys: Pan camera in space view
//   - F11: Fullscreen
//   - ESC: Exit
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"stapledons_voyage/engine/assets"
	"stapledons_voyage/engine/display"
	"stapledons_voyage/engine/handlers"
	"stapledons_voyage/engine/render"
	"stapledons_voyage/engine/screenshot"
	"stapledons_voyage/sim_gen"
)

type Demo struct {
	world    *sim_gen.World
	out      sim_gen.FrameOutput
	renderer *render.Renderer
	display  *display.Manager
	assets   *assets.Manager
	clock    *handlers.EbitenClockHandler

	statusMsg   string
	statusTimer float64
}

func (d *Demo) Update() error {
	dt := 1.0 / 60.0
	d.clock.Update(dt)

	// Update status message timer
	if d.statusTimer > 0 {
		d.statusTimer -= dt
		if d.statusTimer <= 0 {
			d.statusMsg = ""
		}
	}

	// Handle display input (F11 for fullscreen)
	d.display.HandleInput()

	// Capture input and step simulation
	input := render.CaptureInputWithCamera(d.out.Camera, display.InternalWidth, display.InternalHeight)

	result := sim_gen.Step(d.world, &input)
	tuple, ok := result.([]interface{})
	if !ok || len(tuple) != 2 {
		return fmt.Errorf("unexpected Step result")
	}
	if w, ok := tuple[0].(*sim_gen.World); ok {
		d.world = w
	}
	if out, ok := tuple[1].(*sim_gen.FrameOutput); ok {
		d.out = *out
	}

	return nil
}

func (d *Demo) Draw(screen *ebiten.Image) {
	d.renderer.RenderFrame(screen, d.out)

	// Draw instructions overlay
	instructions := []string{
		"Scene Compositing Demo",
		"Observation Deck",
		"",
		"Arrows: Pan space view",
		"F11: Fullscreen",
		"ESC: Exit",
	}

	// Draw semi-transparent background
	overlayW, overlayH := 280, 140
	x, y := 10, 10

	for dy := 0; dy < overlayH; dy++ {
		for dx := 0; dx < overlayW; dx++ {
			screen.Set(x+dx, y+dy, colorWithAlpha(0, 0, 0, 180))
		}
	}

	// Draw instructions
	text := ""
	for _, line := range instructions {
		text += line + "\n"
	}
	ebitenutil.DebugPrintAt(screen, text, x+10, y+10)

	// Draw status message (deck name)
	if d.statusMsg != "" {
		w := screen.Bounds().Dx()
		textW := len(d.statusMsg) * 6
		sx := (w - textW) / 2
		sy := 30

		// Background
		for dy := -2; dy < 14; dy++ {
			for dx := -5; dx < textW+5; dx++ {
				screen.Set(sx+dx, sy+dy, colorWithAlpha(0, 0, 0, 200))
			}
		}

		ebitenutil.DebugPrintAt(screen, d.statusMsg, sx, sy)
	}
}

func (d *Demo) Layout(outsideWidth, outsideHeight int) (int, int) {
	return d.display.Layout(outsideWidth, outsideHeight)
}

func colorWithAlpha(r, g, b, a uint8) colorRGBA {
	return colorRGBA{r, g, b, a}
}

type colorRGBA struct {
	R, G, B, A uint8
}

func (c colorRGBA) RGBA() (r, g, b, a uint32) {
	r = uint32(c.R) * 0x101
	g = uint32(c.G) * 0x101
	b = uint32(c.B) * 0x101
	a = uint32(c.A) * 0x101
	return
}

func main() {
	screenshotFrames := flag.Int("screenshot", 0, "Take screenshot at frame N and exit")
	outputPath := flag.String("output", "out/screenshots/demo-scene-observation.png", "Screenshot output path")
	seed := flag.Int64("seed", 1234, "World seed for determinism")
	flag.Parse()

	// Screenshot mode
	if *screenshotFrames > 0 {
		cfg := screenshot.DefaultConfig()
		cfg.Frames = *screenshotFrames
		cfg.OutputPath = *outputPath
		cfg.Seed = *seed

		if err := screenshot.Capture(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Screenshot failed: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	// Initialize display and assets
	displayMgr := display.NewManager("config.json")

	assetMgr, err := assets.NewManager("assets")
	if err != nil {
		log.Printf("Warning: failed to initialize assets: %v", err)
	}
	if assetMgr != nil {
		assetMgr.SetFontScale(display.InternalHeight)
	}

	// Initialize handlers
	clockHandler := handlers.NewEbitenClockHandler()
	aiHandler := handlers.NewStubAIHandler()

	sim_gen.Init(sim_gen.Handlers{
		Debug: sim_gen.NewDebugContext(),
		Rand:  handlers.NewSeededRandHandler(*seed),
		Clock: clockHandler,
		AI:    aiHandler,
	})

	// Initialize world
	world := sim_gen.InitWorld(*seed)

	renderer := render.NewRenderer(assetMgr)

	demo := &Demo{
		world:    world,
		renderer: renderer,
		display:  displayMgr,
		assets:   assetMgr,
		clock:    clockHandler,
		out: sim_gen.FrameOutput{
			Camera: sim_gen.Camera{X: 0, Y: 0, Zoom: 1.0},
		},
	}

	ebiten.SetWindowTitle("Scene Compositing Demo - Observation Deck")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	if err := ebiten.RunGame(demo); err != nil {
		log.Fatal(err)
	}
}
