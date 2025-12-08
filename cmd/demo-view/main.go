// cmd/demo-view/main.go
// Demo command for testing the view system.
// Usage:
//   go run ./cmd/demo-view                     # Basic space view demo
//   go run ./cmd/demo-view --velocity 0.3      # With SR effect preparation
//   go run ./cmd/demo-view --parallax          # Camera movement demo
//   go run ./cmd/demo-view --transition        # Transition demo (press T/F/W/Z)
//   go run ./cmd/demo-view --screenshot 60     # Take screenshot after N frames
package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"stapledons_voyage/engine/display"
	"stapledons_voyage/engine/view"
)

var (
	velocity        = flag.Float64("velocity", 0.0, "Ship velocity as fraction of c (for SR effects)")
	grIntensity     = flag.Float64("gr", 0.0, "GR intensity (0-1)")
	parallaxDemo    = flag.Bool("parallax", false, "Enable parallax camera movement demo")
	transitionDemo  = flag.Bool("transition", false, "Enable transition demo (press T/F/W/Z)")
	autoTransition  = flag.Int("auto-transition", 0, "Auto-trigger transition at frame N (0=disabled)")
	autoEffect      = flag.String("effect", "fade", "Transition effect: fade, crossfade, wipe, zoom")
	screenshotFrame = flag.Int("screenshot", 0, "Take screenshot after N frames (0=disabled)")
	outputPath      = flag.String("output", "out/demo-view.png", "Screenshot output path")
)

// ColorView is a simple solid-color view for transition testing
type ColorView struct {
	color       color.RGBA
	initialized bool
}

func NewColorView(c color.RGBA) *ColorView {
	return &ColorView{color: c}
}

func (v *ColorView) Type() view.ViewType { return view.ViewBridge } // Use different type
func (v *ColorView) Init() error         { v.initialized = true; return nil }
func (v *ColorView) Enter(from view.ViewType) {}
func (v *ColorView) Exit(to view.ViewType)    {}
func (v *ColorView) Update(dt float64) *view.ViewTransition { return nil }
func (v *ColorView) Draw(screen *ebiten.Image) {
	screen.Fill(v.color)
	// Draw some text to identify this view
	ebitenutil.DebugPrintAt(screen, "COLOR VIEW (Transition Target)", 100, 100)
	ebitenutil.DebugPrintAt(screen, "This view is blue to show transition worked", 100, 120)
}
func (v *ColorView) Layers() view.ViewLayers { return view.ViewLayers{} }

type DemoGame struct {
	manager   *view.Manager
	spaceView *view.SpaceView
	colorView *ColorView // Different view for transitions

	// Animation state
	time          float64
	cameraOffsetX float64
	cameraOffsetY float64
	frameCount    int

	// Screenshot state
	screenshotTaken bool

	// Transition state
	onSpaceView bool
}

func NewDemoGame() *DemoGame {
	g := &DemoGame{
		onSpaceView: true,
	}

	// Create view manager
	g.manager = view.NewManager(display.InternalWidth, display.InternalHeight)

	// Create primary space view
	g.spaceView = view.NewSpaceView()
	g.manager.Register(g.spaceView)

	// Create color view for transition demo (needed for both manual and auto)
	if *transitionDemo || *autoTransition > 0 {
		g.colorView = NewColorView(color.RGBA{30, 30, 80, 255}) // Dark blue
		g.manager.Register(g.colorView)
	}

	// Set initial view
	if err := g.manager.SetCurrent(view.ViewSpace); err != nil {
		log.Printf("Failed to set initial view: %v", err)
	}

	// Apply velocity/GR settings
	g.spaceView.SetVelocity(*velocity)
	g.spaceView.SetGRIntensity(*grIntensity)

	return g
}

func (g *DemoGame) Update() error {
	dt := 1.0 / 60.0
	g.time += dt
	g.frameCount++

	// Handle input
	g.handleInput()

	// Auto-transition at specified frame
	if *autoTransition > 0 && g.frameCount == *autoTransition && !g.manager.IsTransitioning() {
		effect := view.TransitionFade
		switch *autoEffect {
		case "crossfade":
			effect = view.TransitionCrossfade
		case "wipe":
			effect = view.TransitionWipe
		case "zoom":
			effect = view.TransitionZoom
		}
		config := view.TransitionConfig{
			Duration: 1.0,
			Effect:   effect,
			Easing:   view.EaseInOutQuad,
		}
		if err := g.manager.TransitionWithConfig(view.ViewBridge, config); err == nil {
			g.onSpaceView = false
			log.Printf("Auto-triggered %s transition at frame %d", *autoEffect, g.frameCount)
		}
	}

	// Parallax demo: move camera in a figure-8 pattern
	if *parallaxDemo && !g.manager.IsTransitioning() {
		g.cameraOffsetX = math.Sin(g.time*0.5) * 200
		g.cameraOffsetY = math.Sin(g.time*0.3) * math.Cos(g.time*0.5) * 150
		g.spaceView.SetCamera(g.cameraOffsetX, g.cameraOffsetY, 1.0)
	}

	// Update view manager
	return g.manager.Update(dt)
}

func (g *DemoGame) handleInput() {
	if !*transitionDemo || g.manager.IsTransitioning() {
		return
	}

	// Determine target view (toggle between space and color)
	var targetView view.ViewType
	if g.onSpaceView {
		targetView = view.ViewBridge // ColorView uses Bridge type
	} else {
		targetView = view.ViewSpace
	}

	// T key: Fade transition
	if ebiten.IsKeyPressed(ebiten.KeyT) {
		config := view.TransitionConfig{
			Duration: 1.0,
			Effect:   view.TransitionFade,
			Easing:   view.EaseInOutQuad,
		}
		if err := g.manager.TransitionWithConfig(targetView, config); err == nil {
			g.onSpaceView = !g.onSpaceView
			log.Printf("Starting FADE transition to %v", targetView)
		}
	}

	// F key: Crossfade transition
	if ebiten.IsKeyPressed(ebiten.KeyF) {
		config := view.TransitionConfig{
			Duration: 1.5,
			Effect:   view.TransitionCrossfade,
			Easing:   view.EaseInOutSine,
		}
		if err := g.manager.TransitionWithConfig(targetView, config); err == nil {
			g.onSpaceView = !g.onSpaceView
			log.Printf("Starting CROSSFADE transition to %v", targetView)
		}
	}

	// W key: Wipe transition
	if ebiten.IsKeyPressed(ebiten.KeyW) {
		config := view.TransitionConfig{
			Duration: 0.8,
			Effect:   view.TransitionWipe,
			Easing:   view.EaseOutCubic,
		}
		if err := g.manager.TransitionWithConfig(targetView, config); err == nil {
			g.onSpaceView = !g.onSpaceView
			log.Printf("Starting WIPE transition to %v", targetView)
		}
	}

	// Z key: Zoom transition
	if ebiten.IsKeyPressed(ebiten.KeyZ) {
		config := view.TransitionConfig{
			Duration: 1.2,
			Effect:   view.TransitionZoom,
			Easing:   view.EaseInOutBack,
		}
		if err := g.manager.TransitionWithConfig(targetView, config); err == nil {
			g.onSpaceView = !g.onSpaceView
			log.Printf("Starting ZOOM transition to %v", targetView)
		}
	}
}

func (g *DemoGame) Draw(screen *ebiten.Image) {
	// Draw current view
	g.manager.Draw(screen)

	// Draw HUD
	g.drawHUD(screen)

	// Take screenshot if requested
	if *screenshotFrame > 0 && g.frameCount >= *screenshotFrame && !g.screenshotTaken {
		g.takeScreenshot(screen)
		g.screenshotTaken = true
	}
}

func (g *DemoGame) takeScreenshot(screen *ebiten.Image) {
	// Create output directory if needed
	dir := "out"
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Printf("Failed to create output dir: %v", err)
		return
	}

	// Get image from screen
	bounds := screen.Bounds()
	img := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			img.Set(x, y, screen.At(x, y))
		}
	}

	// Save to file
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

func (g *DemoGame) drawHUD(screen *ebiten.Image) {
	y := 20.0
	lineHeight := 18.0

	// Title
	ebitenutil.DebugPrintAt(screen, "View System Demo", 10, int(y))
	y += lineHeight

	// Frame count
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Frame: %d", g.frameCount), 10, int(y))
	y += lineHeight

	// FPS
	fps := ebiten.ActualFPS()
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("FPS: %.1f", fps), 10, int(y))
	y += lineHeight

	// Current view
	if g.onSpaceView {
		ebitenutil.DebugPrintAt(screen, "View: Space (stars)", 10, int(y))
	} else {
		ebitenutil.DebugPrintAt(screen, "View: Color (blue)", 10, int(y))
	}
	y += lineHeight

	// Current settings
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Velocity: %.2fc", *velocity), 10, int(y))
	y += lineHeight

	if *grIntensity > 0 {
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("GR Intensity: %.1f%%", *grIntensity*100), 10, int(y))
		y += lineHeight
	}

	// Parallax mode
	if *parallaxDemo {
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Camera: (%.0f, %.0f)", g.cameraOffsetX, g.cameraOffsetY), 10, int(y))
		y += lineHeight
	}

	// Transition state
	if *transitionDemo {
		state := "Ready"
		if g.manager.IsTransitioning() {
			state = "Transitioning..."
		}
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Transition: %s", state), 10, int(y))
		y += lineHeight
	}

	// Star count info
	if g.spaceView != nil {
		bg := g.spaceView.GetBackground()
		if bg != nil {
			ebitenutil.DebugPrintAt(screen, "Stars: 900 (3 layers)", 10, int(y))
			y += lineHeight
		}
	}

	// Help at bottom
	y = float64(display.InternalHeight) - 60
	ebitenutil.DebugPrintAt(screen, "Controls:", 10, int(y))
	y += lineHeight
	if *transitionDemo {
		ebitenutil.DebugPrintAt(screen, "  T=Fade  F=Crossfade  W=Wipe  Z=Zoom", 10, int(y))
	} else {
		ebitenutil.DebugPrintAt(screen, "  (Run with --transition to enable)", 10, int(y))
	}
	y += lineHeight
	if *screenshotFrame > 0 {
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("  Screenshot at frame %d", *screenshotFrame), 10, int(y))
	}
}

func (g *DemoGame) Layout(outsideWidth, outsideHeight int) (int, int) {
	return display.InternalWidth, display.InternalHeight
}

func main() {
	flag.Parse()

	// Print info
	fmt.Println("View System Demo")
	fmt.Println("================")
	fmt.Printf("Resolution: %dx%d\n", display.InternalWidth, display.InternalHeight)
	fmt.Printf("Velocity: %.2fc\n", *velocity)
	fmt.Printf("GR Intensity: %.2f\n", *grIntensity)
	fmt.Printf("Parallax Demo: %v\n", *parallaxDemo)
	fmt.Printf("Transition Demo: %v\n", *transitionDemo)
	if *screenshotFrame > 0 {
		fmt.Printf("Screenshot at frame: %d -> %s\n", *screenshotFrame, *outputPath)
	}
	fmt.Println()

	// Set up window
	ebiten.SetWindowSize(display.InternalWidth, display.InternalHeight)
	ebiten.SetWindowTitle("Stapledon's Voyage - View System Demo")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	// Run game
	game := NewDemoGame()
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
