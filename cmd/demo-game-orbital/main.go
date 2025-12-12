// cmd/demo-game-orbital/main.go
// AILANG-driven orbital solar system demo.
// Shows planets orbiting the sun with textured rendering and orbit paths.
// All planet data comes from AILANG (sim/celestial.ail).
//
// Usage:
//
//	go build -o bin/demo-game-orbital ./cmd/demo-game-orbital && bin/demo-game-orbital
//	bin/demo-game-orbital --screenshot 60                    # Screenshot after 60 frames
//	bin/demo-game-orbital --time-scale=10                    # 10x faster orbits
//	bin/demo-game-orbital --mode cruise                      # 3D flythrough mode
package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	"log"
	"os"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"stapledons_voyage/engine/assets"
	"stapledons_voyage/engine/display"
	"stapledons_voyage/engine/handlers"
	"stapledons_voyage/engine/render"
	"stapledons_voyage/engine/tetra"
	"stapledons_voyage/engine/view/background"
	"stapledons_voyage/sim_gen"
)

var (
	mode            = flag.String("mode", "orbital", "Demo mode: orbital (top-down) or cruise (3D flythrough)")
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

	// Mode
	demoMode string

	// Cruise mode 3D rendering
	scene3D    *tetra.Scene
	planets3D  []*tetra.Planet
	rings3D    []*tetra.Ring
	sun3D      *tetra.SunLight
	ambient3D  *tetra.AmbientLight
	cameraZ    float64
	cameraY    float64
	cruiseLoop bool
}

func NewOrbitalGame() *OrbitalGame {
	screenW := display.InternalWidth
	screenH := display.InternalHeight

	g := &OrbitalGame{
		timeScale: *timeScale,
		demoMode:  *mode,
		cameraZ:   80.0, // Start far out
		cameraY:   5.0,  // Slightly above plane
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

	// Create renderer (for orbital mode)
	assetMgr, err := assets.NewManager("assets")
	if err != nil {
		log.Printf("Warning: could not load assets: %v", err)
	}
	g.renderer = render.NewRenderer(assetMgr)

	// Create space background
	g.spaceBackground = background.NewSpaceBackground(screenW, screenH)

	// Setup cruise mode 3D scene if needed
	if g.demoMode == "cruise" {
		g.setupCruiseScene(screenW, screenH, assetMgr)
	}

	return g
}

// setupCruiseScene creates the 3D scene for cruise mode using AILANG planet data
func (g *OrbitalGame) setupCruiseScene(screenW, screenH int, assetMgr *assets.Manager) {
	g.scene3D = tetra.NewScene(screenW, screenH)

	// Add lighting
	g.sun3D = tetra.NewSunLight()
	g.sun3D.SetPosition(0, 10, 100) // Sun at origin, light from above/behind camera
	g.sun3D.AddToScene(g.scene3D)

	g.ambient3D = tetra.NewAmbientLight(0.4, 0.4, 0.5, 0.6)
	g.ambient3D.AddToScene(g.scene3D)

	// Add sun at origin with texture
	sunTex := loadTexture("assets/planets/sun.jpg")
	var sunPlanet *tetra.Planet
	if sunTex != nil {
		sunPlanet = tetra.NewTexturedPlanet("sun", 3.0, sunTex)
		log.Printf("Cruise mode: Added textured sun")
	} else {
		sunPlanet = tetra.NewPlanet("sun", 3.0, color.RGBA{255, 220, 100, 255})
		log.Printf("Cruise mode: Added solid sun (no texture)")
	}
	sunPlanet.AddToScene(g.scene3D)
	sunPlanet.SetPosition(0, 0, 0)
	sunPlanet.SetRotationSpeed(0.05) // Slow rotation for the sun
	g.planets3D = append(g.planets3D, sunPlanet)

	// Planet textures
	planetTextures := map[string]string{
		"mercury": "assets/planets/mercury.jpg",
		"venus":   "assets/planets/venus_atmosphere.jpg",
		"earth":   "assets/planets/earth_daymap.jpg",
		"mars":    "assets/planets/mars.jpg",
		"jupiter": "assets/planets/jupiter.jpg",
		"saturn":  "assets/planets/saturn.jpg",
		"uranus":  "assets/planets/uranus.jpg",
		"neptune": "assets/planets/neptune.jpg",
	}

	// Create planets from AILANG data
	// Convert AU to scene units: 1 AU = 5 scene units (compressed for visibility)
	auToScene := 5.0
	for _, p := range g.starSystem.Planets {
		name := strings.ToLower(p.Name) // Normalize to lowercase for texture lookup
		dist := p.OrbitDistance * auToScene

		// Planet size based on actual relative sizes (scaled up for visibility)
		radius := getPlanetRadius(name)

		// Load texture
		tex := loadTexture(planetTextures[name])
		var planet *tetra.Planet
		if tex != nil {
			planet = tetra.NewTexturedPlanet(name, radius, tex)
			log.Printf("Cruise mode: Added textured %s at Z=%.1f", name, -dist)
		} else {
			planet = tetra.NewPlanet(name, radius, getPlanetColor(name))
			log.Printf("Cruise mode: Added solid %s at Z=%.1f", name, -dist)
		}

		planet.AddToScene(g.scene3D)
		// Position along negative Z axis (away from camera)
		// Y offset varies slightly for visual interest
		yOffset := float64(len(g.planets3D)%3-1) * 0.5
		planet.SetPosition(0, yOffset, -dist)
		planet.SetRotationSpeed(0.2)
		g.planets3D = append(g.planets3D, planet)

		// Add Saturn's rings
		if name == "saturn" {
			ringTex := loadTexture("assets/planets/saturn_ring.png")
			ringInner := radius * 1.3
			ringOuter := radius * 2.5
			ring := tetra.NewRing("saturn_ring", ringInner, ringOuter, ringTex)
			ring.AddToScene(g.scene3D)
			ring.SetPosition(0, yOffset, -dist)
			ring.SetTilt(0.47) // Saturn's axial tilt ~27 degrees
			g.rings3D = append(g.rings3D, ring)
			log.Printf("Cruise mode: Added Saturn rings at Z=%.1f", -dist)
		}
	}

	// Initial camera position
	g.scene3D.SetCameraPosition(0, g.cameraY, g.cameraZ)
}

func getPlanetRadius(name string) float64 {
	// More realistic relative planet sizes
	// Based on actual Earth radii, with some compression for gas giants
	// to keep them visible but proportional
	// Real: Mercury=0.38, Venus=0.95, Earth=1.0, Mars=0.53,
	//       Jupiter=11.2, Saturn=9.45, Uranus=4.0, Neptune=3.88
	// We use sqrt scaling for gas giants to keep them visible but show hierarchy
	sizes := map[string]float64{
		"mercury": 0.35,  // Tiny
		"venus":   0.9,   // Almost Earth-sized
		"earth":   1.0,   // Reference
		"mars":    0.5,   // Half Earth
		"jupiter": 3.3,   // sqrt(11.2) - dominant but not overwhelming
		"saturn":  3.0,   // sqrt(9.45) - similar to Jupiter
		"uranus":  2.0,   // sqrt(4.0)
		"neptune": 1.95,  // sqrt(3.88)
	}
	if r, ok := sizes[name]; ok {
		return r
	}
	return 1.0
}

func getPlanetColor(name string) color.RGBA {
	colors := map[string]color.RGBA{
		"mercury": {180, 180, 180, 255},
		"venus":   {230, 200, 150, 255},
		"earth":   {60, 120, 200, 255},
		"mars":    {200, 100, 80, 255},
		"jupiter": {220, 180, 140, 255},
		"saturn":  {210, 190, 150, 255},
		"uranus":  {180, 220, 230, 255},
		"neptune": {80, 120, 200, 255},
	}
	if c, ok := colors[name]; ok {
		return c
	}
	return color.RGBA{200, 200, 200, 255}
}

func loadTexture(path string) *ebiten.Image {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return nil
	}

	return ebiten.NewImageFromImage(img)
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

	// Update cruise mode camera
	if g.demoMode == "cruise" {
		g.updateCruise(dt)
	}

	return nil
}

func (g *OrbitalGame) updateCruise(dt float64) {
	// Update planet rotations
	for _, p := range g.planets3D {
		p.Update(dt)
	}
	// Update ring rotations
	for _, r := range g.rings3D {
		r.Update(dt)
	}

	// Move camera forward (decreasing Z)
	cruiseSpeed := 15.0 * g.timeScale
	g.cameraZ -= cruiseSpeed * dt

	// Reset when we've passed Neptune (~30 AU * 5 = 150 scene units)
	if g.cameraZ < -160 {
		g.cameraZ = 80.0
	}

	g.scene3D.SetCameraPosition(0, g.cameraY, g.cameraZ)
}

func (g *OrbitalGame) Draw(screen *ebiten.Image) {
	// Layer 1: Space background (dark with stars)
	if g.spaceBackground != nil {
		g.spaceBackground.Draw(screen, nil)
	} else {
		screen.Fill(color.RGBA{5, 10, 20, 255})
	}

	if g.demoMode == "cruise" {
		// Cruise mode: render 3D scene
		img3d := g.scene3D.Render()
		screen.DrawImage(img3d, nil)
	} else {
		// Orbital mode: render via AILANG DrawCmds
		drawCmds := sim_gen.RenderSolarSystemTextured(g.starSystem)
		frameOut := sim_gen.FrameOutput{
			Draw:   drawCmds,
			Sounds: nil,
			Debug:  nil,
			Camera: &sim_gen.Camera{X: 0, Y: 0, Zoom: 1.0},
		}
		g.renderer.RenderFrame(screen, frameOut)
	}

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

	title := "Orbital Solar System (AILANG-driven)"
	if g.demoMode == "cruise" {
		title = "Solar System Cruise (AILANG data)"
	}
	ebitenutil.DebugPrintAt(screen, title, 10, int(y))
	y += lineHeight

	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Mode: %s", g.demoMode), 10, int(y))
	y += lineHeight

	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Time Scale: %.1fx", g.timeScale), 10, int(y))
	y += lineHeight

	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Frame: %d", g.frameCount), 10, int(y))
	y += lineHeight

	fps := ebiten.ActualFPS()
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("FPS: %.1f", fps), 10, int(y))
	y += lineHeight

	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Planets: %d", len(g.starSystem.Planets)), 10, int(y))
	y += lineHeight

	if g.demoMode == "cruise" {
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Camera Z: %.1f", g.cameraZ), 10, int(y))
		y += lineHeight
	}

	y += lineHeight

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
	ebitenutil.DebugPrintAt(screen, "Use --mode cruise for 3D flythrough", 10, int(y))
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
	fmt.Printf("Mode: %s\n", *mode)
	fmt.Printf("Time scale: %.1fx\n", *timeScale)
	if *screenshotFrame > 0 {
		fmt.Printf("Screenshot at frame: %d -> %s\n", *screenshotFrame, *outputPath)
	}
	fmt.Println()
	fmt.Println("This demo shows planets orbiting the sun with:")
	fmt.Println("  - Planet data from AILANG (sim/celestial.ail)")
	fmt.Println("  - Orbital mechanics from stepSystem()")
	fmt.Println("  - Rendering via TexturedPlanet DrawCmd (orbital mode)")
	fmt.Println("  - Or Tetra3D 3D flythrough (cruise mode)")
	fmt.Println()
	fmt.Println("Modes:")
	fmt.Println("  orbital - Top-down view showing orbits")
	fmt.Println("  cruise  - 3D flythrough of the solar system")
	fmt.Println()

	ebiten.SetWindowSize(display.InternalWidth, display.InternalHeight)
	ebiten.SetWindowTitle("Stapledon's Voyage - Orbital Demo (AILANG)")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	game := NewOrbitalGame()
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
