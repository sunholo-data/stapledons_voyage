// Package screenshot provides headless screenshot capture for automated testing.
package screenshot

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"stapledons_voyage/engine/assets"
	"stapledons_voyage/engine/display"
	"stapledons_voyage/engine/handlers"
	"stapledons_voyage/engine/render"
	"stapledons_voyage/engine/shader"
	"stapledons_voyage/sim_gen"
)

// Config holds screenshot capture configuration.
type Config struct {
	Frames      int     // Number of frames to run before capture
	OutputPath  string  // Path to save PNG
	Seed        int64   // World seed for determinism
	CameraX     float64 // Initial camera X
	CameraY     float64 // Initial camera Y
	CameraZoom  float64 // Initial camera zoom
	TestMode    bool    // Strip UI for golden file testing
	Effects     string  // Comma-separated effects: "bloom,vignette,crt,aberration,sr_warp,all"
	DemoScene   bool    // Use shader demo scene instead of simulation
	Velocity    float64 // Ship velocity as fraction of c (0.0-0.99) for SR effects
	ViewAngle   float64 // View direction: 0=front, 1.57=side, 3.14=back (radians)
	ArrivalMode bool    // Use arrival sequence instead of normal game
}

// DefaultConfig returns a config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		Frames:     1,
		OutputPath: "out/screenshot.png",
		Seed:       1234,
		CameraX:    0,
		CameraY:    0,
		CameraZoom: 1.0,
	}
}

// screenshotGame implements ebiten.Game for screenshot capture
type screenshotGame struct {
	config       Config
	world        *sim_gen.World // Pointer type in v0.5.8+
	out          sim_gen.FrameOutput
	renderer     *render.Renderer
	effects      *shader.Effects
	renderBuffer *ebiten.Image
	currentFrame int
	captured     bool
	capturedImg  *image.RGBA
	err          error
}

func (g *screenshotGame) Update() error {
	if g.currentFrame >= g.config.Frames {
		// Signal to stop
		return errors.New("screenshot complete")
	}

	// Empty input (no keys, no clicks)
	input := &sim_gen.FrameInput{
		Mouse:            &sim_gen.MouseState{},
		Keys:             []*sim_gen.KeyEvent{},
		ClickedThisFrame: false,
		ActionRequested:  sim_gen.NewPlayerActionActionNone(),
		TestMode:         g.config.TestMode,
	}

	// Step returns []interface{}{*World, *FrameOutput} in v0.5.8+
	result := sim_gen.Step(g.world, input)
	tuple, ok := result.([]interface{})
	if !ok || len(tuple) != 2 {
		g.err = fmt.Errorf("simulation error at frame %d: unexpected Step result", g.currentFrame)
		return g.err
	}
	if w, ok := tuple[0].(*sim_gen.World); ok {
		g.world = w
	}
	if out, ok := tuple[1].(*sim_gen.FrameOutput); ok {
		g.out = *out
	}

	g.currentFrame++
	return nil
}

func (g *screenshotGame) Draw(screen *ebiten.Image) {
	// Check if we need to apply effects
	if g.effects != nil && g.hasEnabledEffects() {
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

	// Capture on final frame
	if g.currentFrame >= g.config.Frames && !g.captured {
		g.captured = true
		// Copy pixels to regular Go image
		bounds := screen.Bounds()
		g.capturedImg = image.NewRGBA(bounds)
		screen.ReadPixels(g.capturedImg.Pix)
	}
}

func (g *screenshotGame) hasEnabledEffects() bool {
	if g.effects == nil {
		return false
	}
	bloomEnabled := g.effects.Bloom().IsEnabled()
	pipelineEffects := g.effects.Pipeline().EnabledEffects()
	if bloomEnabled {
		return true
	}
	return len(pipelineEffects) > 0
}

func (g *screenshotGame) Layout(outsideWidth, outsideHeight int) (int, int) {
	return display.InternalWidth, display.InternalHeight
}

// Capture runs the game for N frames and saves a screenshot.
// Uses Ebiten's game loop for proper GPU command flushing.
func Capture(cfg Config) error {
	// Use demo scene for shader effects testing
	if cfg.DemoScene {
		return CaptureDemo(cfg)
	}

	// Use arrival mode for black hole emergence sequence
	if cfg.ArrivalMode {
		return CaptureArrival(cfg)
	}

	// Initialize asset manager (may fail, that's ok)
	assetMgr, _ := assets.NewManager("assets")

	// Create renderer
	renderer := render.NewRenderer(assetMgr)

	// Initialize effect handlers BEFORE any sim_gen calls
	sim_gen.Init(sim_gen.Handlers{
		Debug: sim_gen.NewDebugContext(),
		Rand:  handlers.NewSeededRandHandler(cfg.Seed),
		Clock: handlers.NewEbitenClockHandler(),
		AI:    handlers.NewStubAIHandler(),
	})

	// Initialize world with seed - returns *World in v0.5.8+
	world := sim_gen.InitWorld(cfg.Seed)

	// Initialize shader effects if requested
	var effects *shader.Effects
	if cfg.Effects != "" || cfg.Velocity > 0 {
		effects = shader.NewEffects()
		if err := effects.Preload(); err != nil {
			return fmt.Errorf("failed to preload shaders: %w", err)
		}
		if cfg.Effects != "" {
			enableEffects(effects, cfg.Effects)
		}
		if cfg.Velocity > 0 {
			enableSRWarpWithVelocity(effects, cfg.Velocity, cfg.ViewAngle)
		}
	}

	game := &screenshotGame{
		config:   cfg,
		world:    world,
		renderer: renderer,
		effects:  effects,
	}

	// Run with hidden window
	ebiten.SetWindowSize(display.InternalWidth, display.InternalHeight)
	ebiten.SetWindowTitle("Screenshot Mode")
	ebiten.SetScreenClearedEveryFrame(true)

	// Run game loop until screenshot is captured
	// The error "screenshot complete" is expected and signals success
	err := ebiten.RunGame(game)
	if err != nil && err.Error() != "screenshot complete" {
		return err
	}

	if game.err != nil {
		return game.err
	}

	if game.capturedImg == nil {
		return fmt.Errorf("failed to capture screenshot")
	}

	// Ensure output directory exists
	dir := filepath.Dir(cfg.OutputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Save as PNG
	f, err := os.Create(cfg.OutputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer f.Close()

	if err := png.Encode(f, game.capturedImg); err != nil {
		return fmt.Errorf("failed to encode PNG: %w", err)
	}

	return nil
}

// enableEffects parses the effects string and enables the specified effects.
// Supports: "bloom", "vignette", "crt", "aberration", "sr_warp", "gr_warp", "all"
func enableEffects(effects *shader.Effects, effectStr string) {
	parts := strings.Split(strings.ToLower(effectStr), ",")
	for _, part := range parts {
		effect := strings.TrimSpace(part)
		switch effect {
		case "all":
			effects.EnableAll()
			return
		case "bloom":
			effects.Bloom().SetEnabled(true)
		case "vignette":
			effects.Pipeline().SetEnabled("vignette", true)
		case "crt":
			effects.Pipeline().SetEnabled("crt", true)
		case "aberration":
			effects.Pipeline().SetEnabled("aberration", true)
		case "sr_warp", "sr", "relativity":
			effects.SRWarp().SetEnabled(true)
		case "gr_warp", "gr", "gravity", "lensing":
			// Enable GR warp with demo mode (centered black hole)
			effects.GRWarp().SetDemoMode(0.5, 0.5, 0.05, 0.01)
			effects.GRWarp().SetEnabled(true)
		case "gr_subtle":
			effects.GRWarp().SetDemoMode(0.5, 0.5, 0.03, 0.0005)
			effects.GRWarp().SetEnabled(true)
		case "gr_strong":
			effects.GRWarp().SetDemoMode(0.5, 0.5, 0.05, 0.005)
			effects.GRWarp().SetEnabled(true)
		case "gr_extreme":
			effects.GRWarp().SetDemoMode(0.5, 0.5, 0.08, 0.05)
			effects.GRWarp().SetEnabled(true)
		}
	}
}

// enableSRWarpWithVelocity enables SR warp with a specific velocity and view angle.
func enableSRWarpWithVelocity(effects *shader.Effects, velocity, viewAngle float64) {
	if velocity > 0 {
		effects.SRWarp().SetForwardVelocity(velocity)
		effects.SRWarp().SetViewAngle(viewAngle)
		effects.SRWarp().SetEnabled(true)
	}
}

// arrivalGame implements ebiten.Game for arrival sequence screenshot capture
type arrivalGame struct {
	config       Config
	arrivalState *sim_gen.ArrivalState // Pointer type in v0.5.8+
	effects      *shader.Effects
	renderBuffer *ebiten.Image
	planetImages map[string]*ebiten.Image // Loaded planet images
	currentFrame int
	captured     bool
	capturedImg  *image.RGBA
	err          error
}

func (g *arrivalGame) Update() error {
	if g.currentFrame >= g.config.Frames {
		return errors.New("screenshot complete")
	}

	// Step arrival simulation
	dt := 1.0 / 60.0
	input := &sim_gen.ArrivalInput{Dt: dt}
	g.arrivalState = sim_gen.StepArrival(g.arrivalState, input)

	// Wire effects based on arrival state
	g.updateEffects()

	g.currentFrame++
	return nil
}

func (g *arrivalGame) updateEffects() {
	if g.effects == nil {
		return
	}

	// Wire GR intensity
	grIntensity := sim_gen.GetGRIntensity(g.arrivalState) // takes *ArrivalState
	if grIntensity > 0.001 {
		phi := float32(grIntensity * 0.05)
		g.effects.GRWarp().SetEnabled(true)
		g.effects.GRWarp().SetDemoMode(0.5, 0.5, 0.08, phi)
	} else {
		g.effects.GRWarp().SetEnabled(false)
	}

	// Wire SR velocity
	velocity := sim_gen.GetArrivalVelocity(g.arrivalState) // takes *ArrivalState
	if velocity > 0.1 {
		g.effects.SRWarp().SetEnabled(true)
		g.effects.SRWarp().SetForwardVelocity(velocity)
	} else {
		g.effects.SRWarp().SetEnabled(false)
	}
}

func (g *arrivalGame) Draw(screen *ebiten.Image) {
	w, h := screen.Bounds().Dx(), screen.Bounds().Dy()

	// Ensure render buffer exists
	if g.renderBuffer == nil || g.renderBuffer.Bounds().Dx() != w {
		g.renderBuffer = ebiten.NewImage(w, h)
	}
	g.renderBuffer.Clear()

	// Draw scene directly to render buffer using Ebiten primitives
	g.drawScene(g.renderBuffer)

	// Apply effects
	hasEffects := g.effects != nil && (g.effects.GRWarp().IsEnabled() || g.effects.SRWarp().IsEnabled())
	if hasEffects {
		g.effects.Apply(screen, g.renderBuffer)
	} else {
		screen.DrawImage(g.renderBuffer, nil)
	}

	// Capture on final frame
	if g.currentFrame >= g.config.Frames && !g.captured {
		g.captured = true
		bounds := screen.Bounds()
		g.capturedImg = image.NewRGBA(bounds)
		screen.ReadPixels(g.capturedImg.Pix)
	}
}

func (g *arrivalGame) drawScene(screen *ebiten.Image) {
	// Dark space background
	screen.Fill(color.RGBA{5, 5, 16, 255})

	// Draw stars
	for i := 0; i < 150; i++ {
		x := float64((i*127 + 53) % 640)
		y := float64((i*89 + 37) % 480)
		size := float64(1 + (i % 3))
		brightness := uint8(0x60 + (i%4)*0x30)
		starColor := color.RGBA{brightness, brightness, brightness, 255}
		ebitenutil.DrawRect(screen, x, y, size, size, starColor)
	}

	// Get state info
	phaseName := sim_gen.GetArrivalPhaseName(g.arrivalState)
	velocity := sim_gen.GetArrivalVelocity(g.arrivalState)
	grIntensity := sim_gen.GetGRIntensity(g.arrivalState)
	planetName := sim_gen.GetArrivalPlanetName(g.arrivalState)

	// Log state at capture frame
	if g.currentFrame == g.config.Frames-1 {
		fmt.Printf("Frame %d state: phase=%s, velocity=%.2fc, gr=%.2f, planet=%q\n",
			g.currentFrame, phaseName, velocity, grIntensity, planetName)
		fmt.Printf("Planet images available: %d\n", len(g.planetImages))
	}

	// Draw planet if approaching one
	if planetName != "" {
		g.drawPlanet(screen, planetName)
	}

	// Draw HUD text
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Phase: %s", phaseName), 10, 10)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Velocity: %.2fc", velocity), 10, 26)
	if grIntensity > 0.001 {
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("GR Intensity: %.0f%%", grIntensity*100), 10, 42)
	}
	if planetName != "" {
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Approaching: %s", strings.ToUpper(planetName)), 250, 420)
	}
}

func (g *arrivalGame) drawPlanet(screen *ebiten.Image, planetName string) {
	if g.planetImages == nil {
		return
	}

	img, ok := g.planetImages[planetName]
	if !ok || img == nil {
		// Fallback: draw colored circle
		var planetColor color.RGBA
		switch planetName {
		case "saturn":
			planetColor = color.RGBA{210, 180, 140, 255} // Tan
		case "jupiter":
			planetColor = color.RGBA{200, 150, 100, 255} // Orange-brown
		case "mars":
			planetColor = color.RGBA{200, 100, 80, 255} // Reddish
		case "earth":
			planetColor = color.RGBA{50, 100, 200, 255} // Blue
		default:
			planetColor = color.RGBA{150, 150, 150, 255} // Gray
		}
		// Draw a simple circle representation
		cx, cy := 320.0, 240.0
		radius := 80.0
		for dy := -radius; dy <= radius; dy++ {
			for dx := -radius; dx <= radius; dx++ {
				if dx*dx+dy*dy <= radius*radius {
					screen.Set(int(cx+dx), int(cy+dy), planetColor)
				}
			}
		}
		return
	}

	// Draw actual planet image centered
	bounds := img.Bounds()
	opts := &ebiten.DrawImageOptions{}
	// Scale to fit nicely on screen (target ~200px width)
	targetSize := 200.0
	scale := targetSize / float64(bounds.Dx())
	scaledW := float64(bounds.Dx()) * scale
	scaledH := float64(bounds.Dy()) * scale
	opts.GeoM.Scale(scale, scale)
	// Center on screen (640x480)
	screenW, screenH := 640.0, 480.0
	opts.GeoM.Translate((screenW-scaledW)/2, (screenH-scaledH)/2)
	screen.DrawImage(img, opts)
}

func (g *arrivalGame) Layout(outsideWidth, outsideHeight int) (int, int) {
	return display.InternalWidth, display.InternalHeight
}

func loadPlanetImages() map[string]*ebiten.Image {
	planets := map[string]*ebiten.Image{}
	planetNames := []string{"saturn", "jupiter", "mars", "earth"}

	for _, name := range planetNames {
		path := fmt.Sprintf("assets/planets/%s.jpg", name)
		img, _, err := ebitenutil.NewImageFromFile(path)
		if err != nil {
			fmt.Printf("Warning: Failed to load planet image %s: %v\n", path, err)
			continue
		}
		fmt.Printf("Loaded planet image: %s (%dx%d)\n", name, img.Bounds().Dx(), img.Bounds().Dy())
		planets[name] = img
	}

	fmt.Printf("Total planet images loaded: %d\n", len(planets))
	return planets
}

// CaptureArrival captures a screenshot of the arrival sequence.
func CaptureArrival(cfg Config) error {
	sim_gen.Init(sim_gen.Handlers{
		Debug: sim_gen.NewDebugContext(),
		Rand:  handlers.NewSeededRandHandler(cfg.Seed),
		Clock: handlers.NewEbitenClockHandler(),
		AI:    handlers.NewStubAIHandler(),
	})

	// Initialize effects
	effects := shader.NewEffects()
	if err := effects.Preload(); err != nil {
		return fmt.Errorf("failed to preload shaders: %w", err)
	}

	// Load planet images
	planetImages := loadPlanetImages()

	game := &arrivalGame{
		config:       cfg,
		arrivalState: sim_gen.InitArrival(),
		effects:      effects,
		planetImages: planetImages,
	}

	ebiten.SetWindowSize(display.InternalWidth, display.InternalHeight)
	ebiten.SetWindowTitle("Arrival Screenshot")
	ebiten.SetScreenClearedEveryFrame(true)

	err := ebiten.RunGame(game)
	if err != nil && err.Error() != "screenshot complete" {
		return err
	}

	if game.capturedImg == nil {
		return fmt.Errorf("failed to capture screenshot")
	}

	dir := filepath.Dir(cfg.OutputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	f, err := os.Create(cfg.OutputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer f.Close()

	if err := png.Encode(f, game.capturedImg); err != nil {
		return fmt.Errorf("failed to encode PNG: %w", err)
	}

	return nil
}
