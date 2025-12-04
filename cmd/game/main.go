package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"stapledons_voyage/engine/assets"
	"stapledons_voyage/engine/display"
	"stapledons_voyage/engine/handlers"
	"stapledons_voyage/engine/render"
	"stapledons_voyage/engine/save"
	"stapledons_voyage/engine/scenario"
	"stapledons_voyage/engine/screenshot"
	"stapledons_voyage/engine/shader"
	"stapledons_voyage/sim_gen"
)

type Game struct {
	world        *sim_gen.World               // Typed world (M-DX16: RecordUpdate preserves struct types)
	out          sim_gen.FrameOutput
	renderer     *render.Renderer
	display      *display.Manager
	assets       *assets.Manager
	clock        *handlers.EbitenClockHandler // For frame timing
	save         *save.Manager                // Auto-save manager (Pillar 1: single save file)
	effects      *shader.Effects              // Post-processing effects
	renderBuffer *ebiten.Image                // Buffer for effects processing
	statusMsg    string                       // Temporary status message
	statusTimer  float64                      // Status message timer
}

func (g *Game) Update() error {
	// Update clock handler (1/60 second per frame at 60 FPS)
	dt := 1.0 / 60.0
	g.clock.Update(dt)

	// Update status message timer
	if g.statusTimer > 0 {
		g.statusTimer -= dt
		if g.statusTimer <= 0 {
			g.statusMsg = ""
		}
	}

	// Track play time for save file
	if g.save != nil {
		g.save.UpdatePlayTime(dt)
	}

	// Handle display input (F11 for fullscreen)
	g.display.HandleInput()

	// Handle effects demo input (F5-F9)
	if g.effects != nil {
		if msgs := g.effects.HandleInput(); len(msgs) > 0 {
			g.statusMsg = msgs[len(msgs)-1]
			g.statusTimer = 2.0
		}
	}

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

	// Auto-save check (Pillar 1: automatic, not player-controlled)
	if g.save != nil && g.save.ShouldAutoSave() {
		if err := g.save.SaveGame(g.world); err != nil {
			log.Printf("Auto-save failed: %v", err)
		}
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Check if we need to apply effects
	if g.effects != nil && g.effects.Bloom().IsEnabled() || g.effects.Pipeline().EnabledEffects() != nil && len(g.effects.Pipeline().EnabledEffects()) > 0 {
		// Render to buffer first
		w, h := screen.Bounds().Dx(), screen.Bounds().Dy()
		if g.renderBuffer == nil || g.renderBuffer.Bounds().Dx() != w {
			g.renderBuffer = ebiten.NewImage(w, h)
		}
		g.renderBuffer.Clear()
		g.renderer.RenderFrame(g.renderBuffer, g.out)

		// Apply effects
		g.effects.Apply(screen, g.renderBuffer)
	} else {
		// Direct render without effects
		g.renderer.RenderFrame(screen, g.out)
	}

	// Draw effects overlay if enabled
	if g.effects != nil && g.effects.ShowOverlay() {
		g.drawEffectsOverlay(screen)
	}

	// Draw status message
	if g.statusMsg != "" {
		g.drawStatusMessage(screen)
	}
}

func (g *Game) drawEffectsOverlay(screen *ebiten.Image) {
	lines := g.effects.OverlayText()
	if len(lines) == 0 {
		return
	}

	// Draw semi-transparent background
	w := screen.Bounds().Dx()
	overlayW, overlayH := 280, 200
	x, y := w-overlayW-10, 10

	// Draw background
	for dy := 0; dy < overlayH; dy++ {
		for dx := 0; dx < overlayW; dx++ {
			screen.Set(x+dx, y+dy, colorWithAlpha(0, 0, 0, 180))
		}
	}

	// Draw text using ebitenutil debug print
	text := ""
	for _, line := range lines {
		text += line + "\n"
	}
	ebitenutil.DebugPrintAt(screen, text, x+10, y+10)
}

func (g *Game) drawStatusMessage(screen *ebiten.Image) {
	w := screen.Bounds().Dx()

	// Center at top
	textW := len(g.statusMsg) * 6
	x := (w - textW) / 2
	y := 30

	// Draw background
	for dy := -2; dy < 14; dy++ {
		for dx := -5; dx < textW+5; dx++ {
			screen.Set(x+dx, y+dy, colorWithAlpha(0, 0, 0, 200))
		}
	}

	ebitenutil.DebugPrintAt(screen, g.statusMsg, x, y)
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
	demoMode := flag.Bool("demo", false, "Enable shader effects demo mode (F5-F9 keys)")
	effectsStr := flag.String("effects", "", "Enable effects for screenshot: bloom,vignette,crt,aberration,all")
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
		cfg.Effects = *effectsStr

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

	// Initialize effect handlers BEFORE any sim_gen calls
	// CRITICAL: Handlers must be set up before InitWorld or Step
	clockHandler := handlers.NewEbitenClockHandler()

	// Initialize AI handler - auto-detects provider from env vars
	// Set AI_PROVIDER=claude, AI_PROVIDER=gemini, or let it auto-detect
	ctx := context.Background()
	aiHandler, err := handlers.NewAIHandlerFromEnv(ctx)
	if err != nil {
		log.Printf("Warning: AI handler init failed: %v, using stub", err)
		aiHandler = handlers.NewStubAIHandler()
	}

	sim_gen.Init(sim_gen.Handlers{
		Debug: sim_gen.NewDebugContext(),
		Rand:  handlers.NewSeededRandHandler(*seed),
		Clock: clockHandler,
		AI:    aiHandler,
	})

	// Initialize save manager (Pillar 1: single save file, auto-save only)
	saveMgr := save.NewManager()

	// Try to load existing save
	var world *sim_gen.World
	savedWorld, err := saveMgr.LoadGame()
	if err != nil {
		log.Printf("Warning: failed to load save: %v", err)
	}

	if savedWorld != nil {
		// Continue from saved game
		world = savedWorld
		log.Printf("Loaded save with %.1f minutes play time", saveMgr.PlayTime()/60)
	} else {
		// New game - initialize fresh world
		worldIface := sim_gen.InitWorld(*seed)
		var ok bool
		world, ok = worldIface.(*sim_gen.World)
		if !ok {
			log.Fatal("InitWorld did not return *World")
		}
	}

	// Initialize shader effects
	effects := shader.NewEffects()
	if err := effects.Preload(); err != nil {
		log.Printf("Warning: shader preload failed: %v", err)
	}
	effects.SetDemoMode(*demoMode)

	// Show demo mode hint if enabled
	if *demoMode {
		log.Println("Demo mode enabled: F5=Bloom, F6=Vignette, F7=CRT, F8=Aberration, F9=Overlay")
	}

	game := &Game{
		world:    world,
		renderer: renderer,
		display:  displayMgr,
		assets:   assetMgr,
		clock:    clockHandler,
		save:     saveMgr,
		effects:  effects,
	}

	// Set up graceful shutdown handler to save on exit
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("Saving game before exit...")
		if err := saveMgr.SaveGame(game.world); err != nil {
			log.Printf("Failed to save on exit: %v", err)
		}
		os.Exit(0)
	}()

	ebiten.SetWindowTitle("Stapledons Voyage")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	if err := ebiten.RunGame(game); err != nil {
		// Save on error exit too
		if saveErr := saveMgr.SaveGame(game.world); saveErr != nil {
			log.Printf("Failed to save on exit: %v", saveErr)
		}
		log.Fatal(err)
	}

	// Save on normal exit
	if err := saveMgr.SaveGame(game.world); err != nil {
		log.Printf("Failed to save on exit: %v", err)
	}
}
