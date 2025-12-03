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
	world    *sim_gen.World // Typed world (M-DX16: RecordUpdate preserves struct types)
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

	// Step returns []interface{}{newWorld, output} - RecordUpdate preserves *World type
	result := sim_gen.Step(g.world, input)
	tuple, ok := result.([]interface{})
	if !ok || len(tuple) != 2 {
		return fmt.Errorf("unexpected Step result")
	}
	if w, ok := tuple[0].(*sim_gen.World); ok {
		g.world = w
	}
	if out, ok := tuple[1].(*sim_gen.FrameOutput); ok {
		g.out = *out
	}

	// Play any sounds requested by the simulation
	if g.assets != nil && len(g.out.Sounds) > 0 {
		// Convert []int64 to []int for PlaySounds
		sounds := make([]int, len(g.out.Sounds))
		for i, s := range g.out.Sounds {
			sounds[i] = int(s)
		}
		g.assets.PlaySounds(sounds)
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

	// Note: Star catalog loading removed - not available in AILANG codegen yet
	// TODO: Add star catalog support when available

	// Create renderer with asset manager
	renderer := render.NewRenderer(assetMgr)

	// Initialize world - type assert to *World (M-DX16: struct types preserved)
	worldIface := sim_gen.InitWorld(*seed)
	world, ok := worldIface.(*sim_gen.World)
	if !ok {
		log.Fatal("InitWorld did not return *World")
	}

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
