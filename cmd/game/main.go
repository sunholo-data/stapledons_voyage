package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"stapledons_voyage/engine/assets"
	"stapledons_voyage/engine/display"
	"stapledons_voyage/engine/render"
	"stapledons_voyage/engine/scenario"
	"stapledons_voyage/engine/screenshot"
	"stapledons_voyage/sim_gen"
)

type Game struct {
	world    sim_gen.World
	out      sim_gen.FrameOutput
	renderer *render.Renderer
	display  *display.Manager
	assets   *assets.Manager
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

	// Play any sounds requested by the simulation
	if g.assets != nil && len(out.Sounds) > 0 {
		g.assets.PlaySounds(out.Sounds)
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.renderer.RenderFrame(screen, g.out)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return g.display.Layout(outsideWidth, outsideHeight)
}

func main() {
	// Parse command line flags
	screenshotFrames := flag.Int("screenshot", 0, "Take screenshot after N frames and exit")
	screenshotOutput := flag.String("output", "out/screenshot.png", "Screenshot output path")
	seed := flag.Int64("seed", 1234, "World seed for determinism")
	cameraStr := flag.String("camera", "", "Initial camera position: x,y,zoom")
	scenarioName := flag.String("scenario", "", "Run a test scenario by name")
	testMode := flag.Bool("test-mode", false, "Strip UI for golden file testing")
	flag.Parse()

	// Handle scenario mode
	if *scenarioName != "" {
		scenarioPath, err := scenario.FindScenario(*scenarioName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Scenario error: %v\n", err)
			os.Exit(1)
		}

		outputDir := filepath.Join("out", "scenarios", *scenarioName)
		if err := scenario.RunVisualScenarioWithOptions(scenarioPath, outputDir, *testMode, *testMode); err != nil {
			fmt.Fprintf(os.Stderr, "Scenario failed: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	// Handle screenshot mode
	if *screenshotFrames > 0 {
		cfg := screenshot.DefaultConfig()
		cfg.Frames = *screenshotFrames
		cfg.OutputPath = *screenshotOutput
		cfg.Seed = *seed
		cfg.TestMode = *testMode

		// Parse camera position if provided
		if *cameraStr != "" {
			parts := strings.Split(*cameraStr, ",")
			if len(parts) == 3 {
				if x, err := strconv.ParseFloat(parts[0], 64); err == nil {
					cfg.CameraX = x
				}
				if y, err := strconv.ParseFloat(parts[1], 64); err == nil {
					cfg.CameraY = y
				}
				if z, err := strconv.ParseFloat(parts[2], 64); err == nil {
					cfg.CameraZoom = z
				}
			}
		}

		if err := screenshot.Capture(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Screenshot failed: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	// Normal game mode
	// Initialize display manager (loads config from file)
	displayMgr := display.NewManager("config.json")

	// Initialize asset manager
	assetMgr, err := assets.NewManager("assets")
	if err != nil {
		log.Printf("Warning: failed to initialize assets: %v", err)
	}

	// Scale fonts for internal resolution
	if assetMgr != nil {
		assetMgr.SetFontScale(display.InternalHeight)
	}

	// Load star catalog for galaxy map
	starCatalogPath := "assets/data/starmap/stars.json"
	if _, err := sim_gen.LoadStarCatalog(starCatalogPath); err != nil {
		log.Printf("Warning: failed to load star catalog: %v", err)
	} else {
		catalog := sim_gen.GetStarCatalog()
		if catalog != nil {
			log.Printf("Loaded %d stars from %s", catalog.Count, starCatalogPath)
		}
	}

	// Create renderer with asset manager
	renderer := render.NewRenderer(assetMgr)

	// Initialize world
	world := sim_gen.InitWorld(*seed)

	game := &Game{
		world:    world,
		renderer: renderer,
		display:  displayMgr,
		assets:   assetMgr,
	}

	ebiten.SetWindowTitle("Stapledons Voyage")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
