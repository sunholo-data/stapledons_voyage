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

// GameMode tracks whether we're in arrival sequence or main game
type GameMode int

const (
	ModeArrival GameMode = iota // Black hole emergence sequence
	ModePlaying                 // Normal gameplay
)

type Game struct {
	mode             GameMode                  // Current game mode (Arrival or Playing)
	world            *sim_gen.World            // Pointer type in v0.5.8+
	arrivalState     *sim_gen.ArrivalState     // Arrival sequence state (when mode == ModeArrival)
	arrivalInitiated bool                      // Whether arrival state has been initialized
	out              sim_gen.FrameOutput
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

	// Handle effects demo input (F5-F9) - only when not in arrival mode
	if g.effects != nil && g.mode != ModeArrival {
		if msgs := g.effects.HandleInput(); len(msgs) > 0 {
			g.statusMsg = msgs[len(msgs)-1]
			g.statusTimer = 2.0
		}
	}

	// Mode-specific update
	switch g.mode {
	case ModeArrival:
		return g.updateArrival(dt)
	case ModePlaying:
		return g.updatePlaying()
	}
	return nil
}

// updateArrival handles the black hole emergence sequence
func (g *Game) updateArrival(dt float64) error {
	if !g.arrivalInitiated {
		g.arrivalState = sim_gen.InitArrival() // returns *ArrivalState
		g.arrivalInitiated = true
	}

	// Step the arrival simulation
	input := &sim_gen.ArrivalInput{Dt: dt}
	g.arrivalState = sim_gen.StepArrival(g.arrivalState, input) // takes and returns *ArrivalState

	// Wire arrival state to shader effects
	g.updateArrivalEffects()

	// Generate visual output for arrival sequence
	g.generateArrivalOutput()

	// Check for arrival completion
	if sim_gen.IsArrivalComplete(g.arrivalState) {
		g.transitionToPlaying()
	}

	return nil
}

// generateArrivalOutput creates DrawCmds for the arrival sequence
func (g *Game) generateArrivalOutput() {
	if !g.arrivalInitiated {
		return
	}

	// Build draw commands for arrival scene
	var cmds []*sim_gen.DrawCmd

	// Use internal resolution for screen-space drawing
	screenW := float64(display.InternalWidth)
	screenH := float64(display.InternalHeight)

	// Dark space background using RectRGBA (screen-space, packed 0xRRGGBBAA)
	// Color: very dark blue (R=5, G=5, B=16, A=255) = 0x050510FF
	cmds = append(cmds, sim_gen.NewDrawCmdRectRGBA(0, 0, screenW, screenH, 0x050510FF, 0))

	// Add stars using CircleRGBA (screen-space with RGBA colors)
	// Use deterministic positions based on a simple pattern
	for i := 0; i < 200; i++ { // More stars for larger screen
		x := float64((i*127+53)%int(screenW)) + float64(i%7)   // Pseudo-random x
		y := float64((i*89+37)%int(screenH)) + float64(i%5)    // Pseudo-random y
		size := float64(1 + (i % 3))                            // 1-3 pixel stars
		// Vary brightness - pack as 0xRRGGBBAA
		brightness := int64(0x80 + (i%4)*0x20)
		rgba := (brightness << 24) | (brightness << 16) | (brightness << 8) | 0xFF
		cmds = append(cmds, sim_gen.NewDrawCmdCircleRGBA(x, y, size, rgba, true, 1))
	}

	// Show current phase as debug text
	phaseName := sim_gen.GetArrivalPhaseName(g.arrivalState)
	velocity := sim_gen.GetArrivalVelocity(g.arrivalState)
	grIntensity := sim_gen.GetGRIntensity(g.arrivalState)

	// Phase indicator (fontSize=12, z=10 for UI)
	phaseText := fmt.Sprintf("Phase: %s", phaseName)
	cmds = append(cmds, sim_gen.NewDrawCmdText(phaseText, 10, 20, 12, 0xFFFFFFFF, 10))

	// Velocity indicator
	velText := fmt.Sprintf("Velocity: %.2fc", velocity)
	cmds = append(cmds, sim_gen.NewDrawCmdText(velText, 10, 40, 12, 0xAAFFAAFF, 10))

	// GR intensity (only during black hole phase)
	if grIntensity > 0.001 {
		grText := fmt.Sprintf("GR Intensity: %.1f%%", grIntensity*100)
		cmds = append(cmds, sim_gen.NewDrawCmdText(grText, 10, 60, 12, 0xFFAAAAFF, 10))
	}

	// Planet name if approaching one (centered at bottom)
	planetName := sim_gen.GetArrivalPlanetName(g.arrivalState)
	if planetName != "" {
		planetText := fmt.Sprintf("Approaching: %s", strings.ToUpper(planetName))
		// Position centered at 40% from left, 85% down
		textX := screenW * 0.35
		textY := screenH * 0.85
		cmds = append(cmds, sim_gen.NewDrawCmdText(planetText, textX, textY, 16, 0xFFFF00FF, 10))
	}

	// Set frame output
	g.out = sim_gen.FrameOutput{
		Draw:   cmds,
		Sounds: nil,
		Camera: &sim_gen.Camera{X: 0, Y: 0, Zoom: 1.0},
	}
}

// updateArrivalEffects wires arrival state to GR/SR shaders
func (g *Game) updateArrivalEffects() {
	if g.effects == nil || !g.arrivalInitiated {
		return
	}

	// Wire GR intensity (black hole gravitational effects)
	grIntensity := sim_gen.GetGRIntensity(g.arrivalState)
	if grIntensity > 0.001 {
		// GR effects active - set demo mode with intensity
		// phi controls intensity: 0.001=subtle, 0.005=strong, 0.05=extreme
		phi := float32(grIntensity * 0.05) // Scale 0-1 to 0-0.05 (extreme)
		g.effects.GRWarp().SetEnabled(true)
		g.effects.GRWarp().SetDemoMode(0.5, 0.5, 0.08, phi) // Center screen, rs=8%
	} else {
		g.effects.GRWarp().SetEnabled(false)
	}

	// Wire SR velocity (relativistic visual effects)
	velocity := sim_gen.GetArrivalVelocity(g.arrivalState)
	if velocity > 0.1 {
		g.effects.SRWarp().SetEnabled(true)
		g.effects.SRWarp().SetForwardVelocity(velocity)
	} else {
		g.effects.SRWarp().SetEnabled(false)
	}
}

// transitionToPlaying switches from arrival to normal gameplay
func (g *Game) transitionToPlaying() {
	g.mode = ModePlaying
	g.arrivalInitiated = false

	// Disable arrival effects
	if g.effects != nil {
		g.effects.GRWarp().SetEnabled(false)
		g.effects.SRWarp().SetEnabled(false)
	}

	g.statusMsg = "Welcome to the Solar System"
	g.statusTimer = 3.0

	log.Println("Arrival sequence complete - transitioning to gameplay")
}

// updatePlaying handles normal gameplay
func (g *Game) updatePlaying() error {
	// Capture game input with camera for screen-to-world conversion
	// Uses internal resolution (640x480) for coordinate conversion
	input := render.CaptureInputWithCamera(*g.out.Camera, display.InternalWidth, display.InternalHeight)

	// Step returns []interface{}{*World, *FrameOutput} in v0.5.8+
	result := sim_gen.Step(g.world, &input)
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
		if err := g.save.SaveGame(g.world); err != nil { // world is already *World
			log.Printf("Auto-save failed: %v", err)
		}
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Check if we need to apply effects
	hasEffects := g.effects != nil && (g.effects.GRWarp().IsEnabled() || g.effects.SRWarp().IsEnabled() || g.effects.Bloom().IsEnabled() || len(g.effects.Pipeline().EnabledEffects()) > 0)
	if hasEffects {
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

// gameFlags holds all parsed command line flags
type gameFlags struct {
	screenshotFrames int
	screenshotOutput string
	seed             int64
	cameraStr        string
	scenarioName     string
	testMode         bool
	demoMode         bool
	effectsStr       string
	demoScene        bool
	velocity         float64
	viewAngle        float64
	arrivalMode      bool // Start with black hole emergence sequence
}

func parseFlags() gameFlags {
	screenshotFrames := flag.Int("screenshot", 0, "Take screenshot after N frames and exit")
	screenshotOutput := flag.String("output", "out/screenshot.png", "Screenshot output path")
	seed := flag.Int64("seed", 1234, "World seed for determinism")
	cameraStr := flag.String("camera", "", "Initial camera position: x,y,zoom")
	scenarioName := flag.String("scenario", "", "Run a test scenario by name")
	testMode := flag.Bool("test-mode", false, "Strip UI for golden file testing")
	demoMode := flag.Bool("demo", false, "Enable shader effects demo mode (F4-F9 keys)")
	effectsStr := flag.String("effects", "", "Enable effects for screenshot: bloom,vignette,crt,aberration,sr_warp,all")
	demoScene := flag.Bool("demo-scene", false, "Use shader demo scene (dark bg, stars, etc) for effects screenshots")
	velocity := flag.Float64("velocity", 0.0, "Ship velocity as fraction of c (0.0-0.99) for SR visual effects")
	viewAngle := flag.Float64("view-angle", 0.0, "View direction: 0=front, 1.57=side, 3.14=back (radians)")
	arrivalMode := flag.Bool("arrival", false, "Start with black hole emergence arrival sequence")
	flag.Parse()

	return gameFlags{
		screenshotFrames: *screenshotFrames,
		screenshotOutput: *screenshotOutput,
		seed:             *seed,
		cameraStr:        *cameraStr,
		scenarioName:     *scenarioName,
		testMode:         *testMode,
		demoMode:         *demoMode,
		effectsStr:       *effectsStr,
		demoScene:        *demoScene,
		velocity:         *velocity,
		viewAngle:        *viewAngle,
		arrivalMode:      *arrivalMode,
	}
}

func handleScenarioMode(flags gameFlags) {
	scenarioPath, err := scenario.FindScenario(flags.scenarioName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Scenario error: %v\n", err)
		os.Exit(1)
	}

	outputDir := filepath.Join("out", "scenarios", flags.scenarioName)
	if err := scenario.RunVisualScenarioWithOptions(scenarioPath, outputDir, flags.testMode, flags.testMode); err != nil {
		fmt.Fprintf(os.Stderr, "Scenario failed: %v\n", err)
		os.Exit(1)
	}
	os.Exit(0)
}

func handleScreenshotMode(flags gameFlags) {
	cfg := screenshot.DefaultConfig()
	cfg.Frames = flags.screenshotFrames
	cfg.OutputPath = flags.screenshotOutput
	cfg.Seed = flags.seed
	cfg.TestMode = flags.testMode
	cfg.Effects = flags.effectsStr
	cfg.DemoScene = flags.demoScene
	cfg.Velocity = flags.velocity
	cfg.ViewAngle = flags.viewAngle
	cfg.ArrivalMode = flags.arrivalMode

	// Parse camera position if provided
	if flags.cameraStr != "" {
		parts := strings.Split(flags.cameraStr, ",")
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

func initializeHandlers(seed int64) (*handlers.EbitenClockHandler, handlers.AIHandler) {
	clockHandler := handlers.NewEbitenClockHandler()

	ctx := context.Background()
	aiHandler, err := handlers.NewAIHandlerFromEnv(ctx)
	if err != nil {
		log.Printf("Warning: AI handler init failed: %v, using stub", err)
		aiHandler = handlers.NewStubAIHandler()
	}

	sim_gen.Init(sim_gen.Handlers{
		Debug: sim_gen.NewDebugContext(),
		Rand:  handlers.NewSeededRandHandler(seed),
		Clock: clockHandler,
		AI:    aiHandler,
	})

	return clockHandler, aiHandler
}

func loadOrCreateWorld(saveMgr *save.Manager, seed int64) *sim_gen.World {
	savedWorld, err := saveMgr.LoadGame()
	if err != nil {
		log.Printf("Warning: failed to load save: %v", err)
	}

	if savedWorld != nil {
		log.Printf("Loaded save with %.1f minutes play time", saveMgr.PlayTime()/60)
		return savedWorld
	}

	return sim_gen.InitWorld(seed) // returns *World in v0.5.8+
}

func setupShutdownHandler(saveMgr *save.Manager, game *Game) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("Saving game before exit...")
		if err := saveMgr.SaveGame(game.world); err != nil { // world is already *World
			log.Printf("Failed to save on exit: %v", err)
		}
		os.Exit(0)
	}()
}

func main() {
	flags := parseFlags()

	// Handle special modes
	if flags.scenarioName != "" {
		handleScenarioMode(flags)
	}
	if flags.screenshotFrames > 0 {
		handleScreenshotMode(flags)
	}

	// Normal game mode
	displayMgr := display.NewManager("config.json")

	assetMgr, err := assets.NewManager("assets")
	if err != nil {
		log.Printf("Warning: failed to initialize assets: %v", err)
	}
	if assetMgr != nil {
		assetMgr.SetFontScale(display.InternalHeight)
	}

	renderer := render.NewRenderer(assetMgr)
	clockHandler, _ := initializeHandlers(flags.seed)
	saveMgr := save.NewManager()
	world := loadOrCreateWorld(saveMgr, flags.seed)

	effects := shader.NewEffects()
	if err := effects.Preload(); err != nil {
		log.Printf("Warning: shader preload failed: %v", err)
	}
	effects.SetDemoMode(flags.demoMode)
	if flags.demoMode {
		log.Println("Demo mode enabled: F4=SR Warp, F5=Bloom, F6=Vignette, F7=CRT, F8=Aberration, F9=Overlay")
	}

	// Determine starting mode
	startMode := ModePlaying
	if flags.arrivalMode {
		startMode = ModeArrival
		log.Println("Starting in arrival mode - black hole emergence sequence")
	}

	game := &Game{
		mode:     startMode,
		world:    world,
		renderer: renderer,
		display:  displayMgr,
		assets:   assetMgr,
		clock:    clockHandler,
		save:     saveMgr,
		effects:  effects,
		out: sim_gen.FrameOutput{
			Camera: &sim_gen.Camera{X: 0, Y: 0, Zoom: 1.0},
		},
	}

	// Initialize arrival state if starting in arrival mode
	if startMode == ModeArrival {
		game.arrivalState = sim_gen.InitArrival()
		game.arrivalInitiated = true
	}

	setupShutdownHandler(saveMgr, game)

	ebiten.SetWindowTitle("Stapledons Voyage")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	if err := ebiten.RunGame(game); err != nil {
		if saveErr := saveMgr.SaveGame(game.world); saveErr != nil { // world is already *World
			log.Printf("Failed to save on exit: %v", saveErr)
		}
		log.Fatal(err)
	}

	if err := saveMgr.SaveGame(game.world); err != nil { // world is already *World
		log.Printf("Failed to save on exit: %v", err)
	}
}
