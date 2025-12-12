// Demo-lookat tests camera and light LookAt functionality.
// This diagnostic tool helps identify issues with the LookAt implementation.
package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	"log"
	"math"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"stapledons_voyage/engine/tetra"
	"stapledons_voyage/engine/view"
	"stapledons_voyage/engine/view/background"
)

const (
	screenWidth  = 1280
	screenHeight = 960
)

var (
	demoMode        string
	screenshotFrame int
	outputPath      string
)

func init() {
	flag.StringVar(&demoMode, "mode", "sun-position", "Demo mode: sun-position, sun-lookat, camera-track, dome-replica, compare")
	flag.IntVar(&screenshotFrame, "screenshot", 0, "Frame to capture screenshot (0 = no screenshot)")
	flag.StringVar(&outputPath, "output", "", "Screenshot output path")
}

// DemoGame is the main game struct.
type DemoGame struct {
	mode       string
	frameCount int
	time       float64

	// Scene for testing
	scene   *tetra.Scene
	planet  *tetra.Planet
	sun     *tetra.SunLight
	ambient *tetra.AmbientLight

	// For dome-replica mode
	planetLayer     *view.PlanetLayer
	spaceBackground *background.SpaceBackground
	planets         []*tetra.Planet

	// For camera-track mode
	orbitAngle float64
}

func main() {
	flag.Parse()

	fmt.Println("LookAt Diagnostic Demo")
	fmt.Println("======================")
	fmt.Printf("Mode: %s\n", demoMode)
	fmt.Printf("Resolution: %dx%d\n", screenWidth, screenHeight)
	if screenshotFrame > 0 {
		fmt.Printf("Screenshot at frame: %d\n", screenshotFrame)
	}
	fmt.Println()

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle(fmt.Sprintf("LookAt Demo - %s", demoMode))

	game := NewDemoGame(demoMode)
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

func NewDemoGame(mode string) *DemoGame {
	g := &DemoGame{
		mode: mode,
	}

	switch mode {
	case "sun-position", "sun-lookat":
		g.setupSunTest()
	case "camera-track":
		g.setupCameraTrack()
	case "dome-replica":
		g.setupDomeReplica()
	case "compare":
		g.setupSunTest()
	default:
		g.setupSunTest()
	}

	return g
}

// setupSunTest creates a simple scene with one planet and configurable sun.
func (g *DemoGame) setupSunTest() {
	g.scene = tetra.NewScene(screenWidth, screenHeight)

	// Add a single planet at origin
	g.planet = tetra.NewPlanet("test", 1.5, color.RGBA{60, 120, 200, 255})
	g.planet.AddToScene(g.scene)
	g.planet.SetPosition(0, 0, 0)
	g.planet.SetRotationSpeed(0.2)

	// Add sun light
	g.sun = tetra.NewSunLight()
	g.sun.AddToScene(g.scene)

	// Position sun based on mode
	switch g.mode {
	case "sun-position":
		// Sun at position, no LookAt - just position determines light direction
		// Directional light shines along its local -Z axis by default
		g.sun.SetPosition(5, 3, 10) // Behind camera, pointing toward planet
		log.Println("Sun: position only at (5, 3, 10)")

	case "sun-lookat":
		// Sun with LookAt (current broken implementation)
		g.sun.SetPosition(5, 3, 10)
		g.sun.LookAt(0, 0, 0) // This uses NewMatrix4LookAt + SetLocalRotation
		log.Println("Sun: position (5, 3, 10) + LookAt(0,0,0) - current impl")

	}

	// Bright ambient to see the sphere even if sun doesn't light it
	g.ambient = tetra.NewAmbientLight(0.5, 0.5, 0.6, 0.5)
	g.ambient.AddToScene(g.scene)

	// Camera positioned to see the planet
	g.scene.SetCameraPosition(0, 0, 5)
}

// setupCameraTrack creates a scene where camera tracks a moving planet.
func (g *DemoGame) setupCameraTrack() {
	g.scene = tetra.NewScene(screenWidth, screenHeight)

	// Planet that will orbit
	g.planet = tetra.NewPlanet("orbiting", 0.8, color.RGBA{200, 150, 100, 255})
	g.planet.AddToScene(g.scene)

	// Sun for lighting
	g.sun = tetra.NewSunLight()
	g.sun.SetPosition(10, 5, 10)
	g.sun.AddToScene(g.scene)

	g.ambient = tetra.NewAmbientLight(0.4, 0.4, 0.5, 0.4)
	g.ambient.AddToScene(g.scene)

	// Camera starts at origin
	g.scene.SetCameraPosition(0, 0, 0)

	log.Println("Camera track: camera at origin, planet orbits around")
}

// setupDomeReplica replicates the exact dome renderer setup.
func (g *DemoGame) setupDomeReplica() {
	// Create space background
	g.spaceBackground = background.NewSpaceBackground(screenWidth, screenHeight)

	// Load galaxy if available
	if galaxyImg := loadTexture("assets/data/starmap/background/galaxy_4k.jpg"); galaxyImg != nil {
		g.spaceBackground.SetGalaxyImage(galaxyImg)
		log.Println("Loaded galaxy background")
	}

	// Create planet layer (same as DomeRenderer does)
	g.planetLayer = view.NewPlanetLayer(screenWidth, screenHeight)

	// Add planets with same config as dome_renderer.go createSolarSystem()
	planetConfigs := []struct {
		name    string
		color   color.RGBA
		radius  float64
		dist    float64
		texture string
	}{
		{"neptune", color.RGBA{80, 120, 200, 255}, 1.0, 15, "assets/planets/neptune.jpg"},
		{"saturn", color.RGBA{210, 190, 150, 255}, 1.8, 50, "assets/planets/saturn.jpg"},
		{"jupiter", color.RGBA{220, 180, 140, 255}, 2.2, 90, "assets/planets/jupiter.jpg"},
		{"mars", color.RGBA{200, 100, 80, 255}, 0.5, 130, "assets/planets/mars.jpg"},
		{"earth", color.RGBA{60, 120, 200, 255}, 0.7, 150, "assets/planets/earth_daymap.jpg"},
	}

	for _, cfg := range planetConfigs {
		var planet *tetra.Planet

		tex := loadTexture(cfg.texture)
		if tex != nil {
			planet = g.planetLayer.AddTexturedPlanet(cfg.name, cfg.radius, tex)
			log.Printf("Loaded texture for %s", cfg.name)
		} else {
			planet = g.planetLayer.AddPlanet(cfg.name, cfg.radius, cfg.color)
		}

		yOffset := cfg.dist * 0.15
		planet.SetPosition(0, yOffset, -cfg.dist)
		planet.SetRotationSpeed(0.1)
		g.planets = append(g.planets, planet)
	}

	// CRITICAL: Use same sun position as working demo-solar-system
	g.planetLayer.SetSunPosition(5, 3, 15)

	// Camera position - matches dome renderer
	g.planetLayer.SetCameraPosition(0, 0, 10)

	log.Printf("Dome replica: %d planets, camera at Z=10", len(g.planets))
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

func (g *DemoGame) Update() error {
	dt := 1.0 / 60.0
	g.time += dt
	g.frameCount++

	switch g.mode {
	case "camera-track":
		// Planet orbits around camera
		g.orbitAngle += dt * 0.5
		radius := 3.0
		x := radius * math.Cos(g.orbitAngle)
		z := radius * math.Sin(g.orbitAngle)
		g.planet.SetPosition(x, 0, z)

		// Camera tracks planet using LookAt
		g.scene.LookAt(x, 0, z)

	case "dome-replica":
		// Update planet rotations
		if g.planetLayer != nil {
			g.planetLayer.Update(dt)
		}

	default:
		// Update planet rotation
		if g.planet != nil {
			g.planet.Update(dt)
		}
	}

	return nil
}

func (g *DemoGame) Draw(screen *ebiten.Image) {
	switch g.mode {
	case "dome-replica":
		g.drawDomeReplica(screen)
	default:
		g.drawStandard(screen)
	}

	// HUD
	g.drawHUD(screen)

	// Screenshot
	if screenshotFrame > 0 && g.frameCount == screenshotFrame {
		g.takeScreenshot(screen)
	}
}

func (g *DemoGame) drawStandard(screen *ebiten.Image) {
	// Dark blue background
	screen.Fill(color.RGBA{5, 10, 20, 255})

	// Render 3D scene
	if g.scene != nil {
		img3d := g.scene.Render()
		screen.DrawImage(img3d, nil)
	}
}

func (g *DemoGame) drawDomeReplica(screen *ebiten.Image) {
	// Layer 1: Starfield background
	if g.spaceBackground != nil {
		g.spaceBackground.Draw(screen, nil)
	}

	// Layer 2: 3D planets (this is what's failing in dome renderer!)
	if g.planetLayer != nil {
		g.planetLayer.Draw(screen)
	}
}

func (g *DemoGame) drawHUD(screen *ebiten.Image) {
	info := fmt.Sprintf("LookAt Demo (%s)\nMode: %s\nFrame: %d\nFPS: %.1f",
		g.mode, g.getModeDescription(), g.frameCount, ebiten.ActualFPS())

	if g.mode == "camera-track" {
		info += fmt.Sprintf("\nOrbit angle: %.2f rad", g.orbitAngle)
	}

	if g.mode == "dome-replica" {
		info += fmt.Sprintf("\nPlanets: %d", len(g.planets))
	}

	ebitenutil.DebugPrint(screen, info)

	// Mode-specific legend at bottom
	legend := g.getModeLegend()
	ebitenutil.DebugPrintAt(screen, legend, 10, screenHeight-60)
}

func (g *DemoGame) getModeDescription() string {
	switch g.mode {
	case "sun-position":
		return "Sun uses position only (no LookAt)"
	case "sun-lookat":
		return "Sun uses LookAt (current impl)"
	case "sun-node-lookat":
		return "Sun uses Node.LookAt (tetra3d)"
	case "camera-track":
		return "Camera tracks orbiting planet"
	case "dome-replica":
		return "Exact dome renderer setup"
	default:
		return "Unknown"
	}
}

func (g *DemoGame) getModeLegend() string {
	switch g.mode {
	case "sun-position":
		return "Testing: Sun at (5,3,10), default direction\nExpect: Planet lit from top-right"
	case "sun-lookat":
		return "Testing: Sun.LookAt(0,0,0) with NewMatrix4LookAt\nIf broken: Planet may be dark or wrongly lit"
	case "sun-node-lookat":
		return "Testing: tetra3d Node.LookAt built-in\nExpect: Should work correctly"
	case "camera-track":
		return "Testing: Camera tracks orbiting planet\nPlanet should stay centered"
	case "dome-replica":
		return "Testing: Exact dome renderer setup\nIf planets visible, issue is elsewhere"
	default:
		return ""
	}
}

func (g *DemoGame) takeScreenshot(screen *ebiten.Image) {
	path := outputPath
	if path == "" {
		path = fmt.Sprintf("out/screenshots/lookat-%s.png", g.mode)
	}

	f, err := os.Create(path)
	if err != nil {
		log.Printf("Failed to create screenshot: %v", err)
		return
	}
	defer f.Close()

	if err := png.Encode(f, screen); err != nil {
		log.Printf("Failed to encode screenshot: %v", err)
		return
	}

	log.Printf("Screenshot saved to %s", path)
	fmt.Printf("Screenshot saved to %s\n", path)
	os.Exit(0)
}

func (g *DemoGame) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}
