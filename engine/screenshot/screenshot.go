// Package screenshot provides headless screenshot capture for automated testing.
package screenshot

import (
	"errors"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"stapledons_voyage/engine/assets"
	"stapledons_voyage/engine/display"
	"stapledons_voyage/engine/render"
	"stapledons_voyage/engine/shader"
	"stapledons_voyage/sim_gen"
)

// Config holds screenshot capture configuration.
type Config struct {
	Frames     int     // Number of frames to run before capture
	OutputPath string  // Path to save PNG
	Seed       int64   // World seed for determinism
	CameraX    float64 // Initial camera X
	CameraY    float64 // Initial camera Y
	CameraZoom float64 // Initial camera zoom
	TestMode   bool    // Strip UI for golden file testing
	Effects    string  // Comma-separated effects: "bloom,vignette,crt,aberration,sr_warp,all"
	DemoScene  bool    // Use shader demo scene instead of simulation
	Velocity   float64 // Ship velocity as fraction of c (0.0-0.99) for SR effects
	ViewAngle  float64 // View direction: 0=front, 1.57=side, 3.14=back (radians)
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
	world        *sim_gen.World // Typed world (M-DX16: RecordUpdate preserves struct types)
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
	input := sim_gen.FrameInput{
		Mouse:            sim_gen.MouseState{},
		Keys:             []*sim_gen.KeyEvent{},
		ClickedThisFrame: false,
		ActionRequested:  *sim_gen.NewPlayerActionActionNone(),
		TestMode:         g.config.TestMode,
	}

	// Step returns []interface{}{newWorld, output} - RecordUpdate preserves *World type
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

	// Initialize asset manager (may fail, that's ok)
	assetMgr, _ := assets.NewManager("assets")

	// Create renderer
	renderer := render.NewRenderer(assetMgr)

	// Initialize world with seed - type assert to *World (M-DX16: struct types preserved)
	worldIface := sim_gen.InitWorld(cfg.Seed)
	world, ok := worldIface.(*sim_gen.World)
	if !ok {
		return fmt.Errorf("InitWorld did not return *World")
	}

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
// Supports: "bloom", "vignette", "crt", "aberration", "sr_warp", "all"
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
