// cmd/demo-orbital/main.go
// AILANG-driven orbital solar system demo.
// Shows planets orbiting the sun with textured rendering and orbit paths.
// All planet data comes from AILANG (sim/celestial.ail).
//
// Usage:
//   go build -o bin/demo-orbital ./cmd/demo-orbital && bin/demo-orbital
//   bin/demo-orbital --screenshot 60                    # Screenshot after 60 frames
//   bin/demo-orbital --time-scale=10                    # 10x faster orbits
package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"stapledons_voyage/engine/assets"
	"stapledons_voyage/engine/display"
	"stapledons_voyage/engine/handlers"
	"stapledons_voyage/engine/render"
	"stapledons_voyage/engine/view/background"
	"stapledons_voyage/sim_gen"
)

var (
	timeScale       = flag.Float64("time-scale", 1.0, "Time scale multiplier (1.0 = normal)")
	screenshotFrame = flag.Int("screenshot", 0, "Take screenshot after N frames (0=disabled)")
	outputPath      = flag.String("output", "out/screenshots/demo-orbital.png", "Screenshot output path")
)

// OrbitalGame is the main game struct
type OrbitalGame struct {
	// AILANG state
	starSystem *sim_gen.StarSystem

	// Rendering
	renderer        *render.Renderer
	spaceBackground *background.SpaceBackground

	// Animation
	time       float64
	frameCount int

	// Time scaling
	timeScale float64

	// Screenshot
	screenshotTaken bool
}

func NewOrbitalGame() *OrbitalGame {
	screenW := display.InternalWidth
	screenH := display.InternalHeight

	g := &OrbitalGame{
		timeScale: *timeScale,
	}

	// Initialize AILANG handlers
	sim_gen.Init(sim_gen.Handlers{
		Debug: sim_gen.NewDebugContext(),
		Rand:  &handlers.DefaultRandHandler{},
		Clock: &handlers.EbitenClockHandler{},
		AI:    handlers.NewStubAIHandler(),
	})

	// Initialize solar system from AILANG
	g.starSystem = sim_gen.InitSolSystem()
	log.Printf("Initialized solar system: %s with %d planets",
		g.starSystem.Name, len(g.starSystem.Planets))

	// Create renderer
	assetMgr, err := assets.NewManager("assets")
	if err != nil {
		log.Printf("Warning: could not load assets: %v", err)
	}
	g.renderer = render.NewRenderer(assetMgr)

	// Create space background
	g.spaceBackground = background.NewSpaceBackground(screenW, screenH)

	return g
}

func (g *OrbitalGame) Update() error {
	dt := 1.0 / 60.0 * g.timeScale
	g.time += dt
	g.frameCount++

	// Step orbital mechanics via AILANG
	// dt = 0.001 years per frame at normal speed (~8.7 hours per frame)
	// This gives visible orbital motion without being too fast
	orbitalDt := 0.001 * g.timeScale
	g.starSystem = sim_gen.StepSystem(g.starSystem, orbitalDt)

	return nil
}

func (g *OrbitalGame) Draw(screen *ebiten.Image) {
	// Layer 1: Space background (dark with stars)
	if g.spaceBackground != nil {
		g.spaceBackground.Draw(screen, nil)
	} else {
		screen.Fill(color.RGBA{5, 10, 20, 255})
	}

	// Layer 2: Render solar system via AILANG
	// Get DrawCmds from AILANG's renderSolarSystemTextured
	drawCmds := sim_gen.RenderSolarSystemTextured(g.starSystem)

	// Create a FrameOutput to use the renderer
	frameOut := sim_gen.FrameOutput{
		Draw:   drawCmds,
		Sounds: nil,
		Debug:  nil,
		Camera: &sim_gen.Camera{X: 0, Y: 0, Zoom: 1.0},
	}

	// Render the DrawCmds
	g.renderer.RenderFrame(screen, frameOut)

	// Layer 3: HUD
	g.drawHUD(screen)

	// Screenshot
	if *screenshotFrame > 0 && g.frameCount >= *screenshotFrame && !g.screenshotTaken {
		g.takeScreenshot(screen)
		g.screenshotTaken = true
	}
}

func (g *OrbitalGame) drawHUD(screen *ebiten.Image) {
	y := 20.0
	lineHeight := 18.0

	ebitenutil.DebugPrintAt(screen, "Orbital Solar System (AILANG-driven)", 10, int(y))
	y += lineHeight

	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Time Scale: %.1fx", g.timeScale), 10, int(y))
	y += lineHeight

	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Frame: %d", g.frameCount), 10, int(y))
	y += lineHeight

	fps := ebiten.ActualFPS()
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("FPS: %.1f", fps), 10, int(y))
	y += lineHeight

	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Planets: %d", len(g.starSystem.Planets)), 10, int(y))
	y += lineHeight * 2

	// Planet info
	ebitenutil.DebugPrintAt(screen, "Planets (from AILANG):", 10, int(y))
	y += lineHeight
	for _, p := range g.starSystem.Planets {
		info := fmt.Sprintf("  %s: %.2f AU, period %.2f yr", p.Name, p.OrbitDistance, p.OrbitalPeriod)
		ebitenutil.DebugPrintAt(screen, info, 10, int(y))
		y += lineHeight
	}

	// Bottom help
	y = float64(display.InternalHeight) - 60
	ebitenutil.DebugPrintAt(screen, "All data from sim/celestial.ail", 10, int(y))
	y += lineHeight
	ebitenutil.DebugPrintAt(screen, "Use --time-scale=N to speed up orbits", 10, int(y))
	y += lineHeight
	if *screenshotFrame > 0 {
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Screenshot at frame %d -> %s", *screenshotFrame, *outputPath), 10, int(y))
	}
}

func (g *OrbitalGame) takeScreenshot(screen *ebiten.Image) {
	// Create output directory
	if err := os.MkdirAll("out/screenshots", 0755); err != nil {
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

func (g *OrbitalGame) Layout(outsideWidth, outsideHeight int) (int, int) {
	return display.InternalWidth, display.InternalHeight
}

func main() {
	flag.Parse()

	fmt.Println("Orbital Solar System Demo (AILANG-driven)")
	fmt.Println("==========================================")
	fmt.Printf("Resolution: %dx%d\n", display.InternalWidth, display.InternalHeight)
	fmt.Printf("Time scale: %.1fx\n", *timeScale)
	if *screenshotFrame > 0 {
		fmt.Printf("Screenshot at frame: %d -> %s\n", *screenshotFrame, *outputPath)
	}
	fmt.Println()
	fmt.Println("This demo shows planets orbiting the sun with:")
	fmt.Println("  - Planet data from AILANG (sim/celestial.ail)")
	fmt.Println("  - Orbital mechanics from stepSystem()")
	fmt.Println("  - Rendering via TexturedPlanet DrawCmd")
	fmt.Println()

	ebiten.SetWindowSize(display.InternalWidth, display.InternalHeight)
	ebiten.SetWindowTitle("Stapledon's Voyage - Orbital Demo (AILANG)")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	game := NewOrbitalGame()
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
