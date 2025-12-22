// Package main provides a demo combining 3D interior with 3D solar system objects visible through windows.
// This demonstrates Tetra3D objects rendered to window textures.
//
// Controls:
//   WASD: Move | Mouse: Look | Shift: Run
//   1-6: Select body to view (1=Sun, 2=Mercury, 3=Venus, 4=Mars, 5=Jupiter, 6=Saturn)
//   +/-: Move closer/further from object
//   Esc: Quit
//
// Usage:
//
//	go run ./cmd/demo-interior-solar
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
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"stapledons_voyage/engine/assets"
	"stapledons_voyage/engine/handlers"
	"stapledons_voyage/engine/render"
	"stapledons_voyage/engine/tetra"
	"stapledons_voyage/sim_gen"
)

const (
	screenWidth    = 1280
	screenHeight   = 720
	windowTexWidth  = 2048 // Window texture width (matches 4:2 aspect ratio)
	windowTexHeight = 1024 // Window texture height
)

var (
	screenshotFrame = flag.Int("screenshot", 0, "Take screenshot at frame N and exit")
	screenshotPath  = flag.String("output", "out/screenshots/interior-solar.png", "Screenshot output path")
)

// PlanetView holds info about a viewable planet
type PlanetView struct {
	Name        string
	TexturePath string
	Radius      float64
	HasRings    bool
	Color       color.RGBA
}

// Game implements ebiten.Game for the interior + solar demo.
type Game struct {
	renderer       *render.Renderer
	interior       *sim_gen.InteriorState
	shipNavigation *sim_gen.ShipNavigation
	frameCount     int

	// 3D space scene (for rendering planets to window textures)
	spaceScene   *tetra.Scene
	planets      []*tetra.Planet
	rings        []*tetra.RingSystem
	currentPlanet int
	planetDist    float64 // Distance from planet

	// Window texture from 3D scene
	windowTex     *ebiten.Image
	needsTexUpdate bool

	// Planet definitions
	planetViews []PlanetView
}

// NewGame creates a new interior + solar demo game.
func NewGame() *Game {
	// Initialize AILANG handlers
	sim_gen.Init(sim_gen.Handlers{
		Debug: sim_gen.NewDebugContext(),
		Rand:  &handlers.DefaultRandHandler{},
		Clock: &handlers.EbitenClockHandler{},
		AI:    handlers.NewStubAIHandler(),
	})

	// Initialize interior state from AILANG
	interior := sim_gen.InitInterior()
	interior.Player.Yaw = 3.14159 // Face north toward window
	interior.Player.Pitch = 0

	// Initialize ship navigation
	shipNav := sim_gen.InitNavigation()

	// Create renderer
	assetMgr, err := assets.NewManager("assets")
	if err != nil {
		log.Printf("Warning: could not load assets: %v", err)
		assetMgr = nil
	}

	// Define viewable celestial bodies (1=Sun, 2-6=Planets)
	planetViews := []PlanetView{
		{"Sun", "assets/planets/sun.jpg", 4.0, false, color.RGBA{255, 220, 100, 255}},
		{"Mercury", "assets/planets/mercury.jpg", 0.8, false, color.RGBA{180, 160, 140, 255}},
		{"Venus", "assets/planets/venus.jpg", 1.2, false, color.RGBA{230, 200, 150, 255}},
		{"Mars", "assets/planets/mars.jpg", 0.9, false, color.RGBA{200, 100, 80, 255}},
		{"Jupiter", "assets/planets/jupiter.jpg", 3.0, false, color.RGBA{220, 180, 140, 255}},
		{"Saturn", "assets/planets/saturn.jpg", 2.5, true, color.RGBA{210, 190, 150, 255}},
	}

	g := &Game{
		renderer:       render.NewRenderer(assetMgr),
		interior:       interior,
		shipNavigation: shipNav,
		planetViews:    planetViews,
		currentPlanet:  5, // Start with Saturn (has rings!)
		planetDist:     8.0,
		needsTexUpdate: true,
	}

	// Create 3D space scene for planet rendering
	g.initSpaceScene()

	return g
}

// initSpaceScene creates the Tetra3D scene for rendering planets
func (g *Game) initSpaceScene() {
	g.spaceScene = tetra.NewScene(windowTexWidth, windowTexHeight)
	g.spaceScene.SetLightingEnabled(true)

	// Add sun light (coming from behind camera)
	sunLight := tetra.NewSunLight()
	sunLight.SetPosition(0, 2, 10)
	sunLight.AddToScene(g.spaceScene)

	// Add ambient light
	ambient := tetra.NewAmbientLight(0.3, 0.3, 0.35, 0.5)
	ambient.AddToScene(g.spaceScene)

	// Create all planets
	for _, pv := range g.planetViews {
		var planet *tetra.Planet
		tex := loadTexture(pv.TexturePath)
		if tex != nil {
			planet = tetra.NewTexturedPlanet(pv.Name, pv.Radius, tex)
			log.Printf("Loaded texture for %s", pv.Name)
		} else {
			planet = tetra.NewPlanet(pv.Name, pv.Radius, pv.Color)
			log.Printf("Using solid color for %s", pv.Name)
		}
		planet.AddToScene(g.spaceScene)
		planet.SetPosition(0, 0, 0)
		planet.SetRotationSpeed(0.2)
		planet.Model().SetVisible(false, true) // Start hidden

		// Sun is self-illuminated (shadeless)
		if pv.Name == "Sun" {
			planet.SetShadeless(true)
		}

		g.planets = append(g.planets, planet)

		// Add rings for Saturn
		if pv.HasRings {
			bands := tetra.SaturnRingBands(pv.Radius)
			rings := tetra.NewRingSystem(pv.Name+"_rings", bands)
			rings.AddToScene(g.spaceScene)
			rings.SetPosition(0, 0, 0)
			rings.SetTilt(0.47)
			rings.SetVisible(false)
			g.rings = append(g.rings, rings)
			log.Printf("Added rings for %s", pv.Name)
		} else {
			g.rings = append(g.rings, nil)
		}
	}

	// Show the current planet
	g.showPlanet(g.currentPlanet)

	// Create window texture with correct aspect ratio (2:1 for bridge window)
	g.windowTex = ebiten.NewImage(windowTexWidth, windowTexHeight)
}

// showPlanet makes only the specified planet visible
func (g *Game) showPlanet(index int) {
	for i, p := range g.planets {
		visible := (i == index)
		p.Model().SetVisible(visible, true)
		if g.rings[i] != nil {
			g.rings[i].SetVisible(visible)
		}
	}
	g.needsTexUpdate = true
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

func captureInput() *sim_gen.FrameInput {
	x, y := ebiten.CursorPosition()

	var buttons []int64
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		buttons = append(buttons, 0)
	}

	var keys []sim_gen.KeyEvent
	for _, k := range inpututil.AppendJustPressedKeys(nil) {
		keys = append(keys, sim_gen.KeyEvent{
			Key:  int64(k),
			Kind: "pressed",
		})
	}

	flight := &sim_gen.FlightInput{
		W:     ebiten.IsKeyPressed(ebiten.KeyW),
		A:     ebiten.IsKeyPressed(ebiten.KeyA),
		S:     ebiten.IsKeyPressed(ebiten.KeyS),
		D:     ebiten.IsKeyPressed(ebiten.KeyD),
		Shift: ebiten.IsKeyPressed(ebiten.KeyShift),
	}

	return &sim_gen.FrameInput{
		Mouse: sim_gen.MouseState{
			X:       float64(x),
			Y:       float64(y),
			Buttons: buttons,
		},
		Keys:   keys,
		Flight: flight,
	}
}

func (g *Game) Update() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}

	// Celestial body selection (1=Sun, 2-6=Planets)
	if inpututil.IsKeyJustPressed(ebiten.Key1) {
		g.currentPlanet = 0
		g.showPlanet(0)
	}
	if inpututil.IsKeyJustPressed(ebiten.Key2) {
		g.currentPlanet = 1
		g.showPlanet(1)
	}
	if inpututil.IsKeyJustPressed(ebiten.Key3) {
		g.currentPlanet = 2
		g.showPlanet(2)
	}
	if inpututil.IsKeyJustPressed(ebiten.Key4) {
		g.currentPlanet = 3
		g.showPlanet(3)
	}
	if inpututil.IsKeyJustPressed(ebiten.Key5) {
		g.currentPlanet = 4
		g.showPlanet(4)
	}
	if inpututil.IsKeyJustPressed(ebiten.Key6) {
		g.currentPlanet = 5
		g.showPlanet(5)
	}

	// Distance controls
	if ebiten.IsKeyPressed(ebiten.KeyEqual) || ebiten.IsKeyPressed(ebiten.KeyKPAdd) {
		g.planetDist -= 0.1
		if g.planetDist < 3 {
			g.planetDist = 3
		}
		g.needsTexUpdate = true
	}
	if ebiten.IsKeyPressed(ebiten.KeyMinus) || ebiten.IsKeyPressed(ebiten.KeyKPSubtract) {
		g.planetDist += 0.1
		if g.planetDist > 20 {
			g.planetDist = 20
		}
		g.needsTexUpdate = true
	}

	// Update planet rotation
	dt := 1.0 / 60.0
	for _, p := range g.planets {
		p.Update(dt)
	}

	// Update interior
	input := captureInput()
	g.interior = sim_gen.StepInterior(g.interior, input)

	g.frameCount++
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{10, 10, 15, 255})

	// Update window texture if needed
	if g.needsTexUpdate || g.frameCount%2 == 0 { // Update every other frame for rotation
		g.renderWindowTexture()
		g.needsTexUpdate = false
	}

	// Get draw commands from AILANG
	drawCmds := sim_gen.RenderInterior(g.interior, g.shipNavigation)

	// Create frame output
	output := sim_gen.FrameOutput{
		Draw:   drawCmds,
		Camera: sim_gen.Camera{X: 0, Y: 0, Zoom: 1.0},
	}

	// Override window texture with our 3D rendered texture
	g.renderer.SetWindowTexture(g.windowTex)

	// Render interior
	g.renderer.RenderFrame(screen, output)

	// HUD
	pv := g.planetViews[g.currentPlanet]
	hudText := fmt.Sprintf(
		"Interior + 3D Solar System Demo\n"+
			"═══════════════════════════════════════\n"+
			"WASD: Move | Mouse: Look | Shift: Run\n"+
			"1-6: Select Body (1=Sun) | +/-: Distance\n"+
			"═══════════════════════════════════════\n"+
			"Viewing: %s\n"+
			"Distance: %.1f\n"+
			"Window Texture: %dx%d\n"+
			"═══════════════════════════════════════\n"+
			"Position: (%.1f, %.1f, %.1f)\n"+
			"Frame: %d",
		pv.Name,
		g.planetDist,
		windowTexWidth, windowTexHeight,
		g.interior.Player.Pos.X, g.interior.Player.Pos.Y, g.interior.Player.Pos.Z,
		g.frameCount,
	)
	ebitenutil.DebugPrint(screen, hudText)

	// Screenshot
	if *screenshotFrame > 0 && g.frameCount >= *screenshotFrame {
		g.saveScreenshot(screen)
	}
}

// renderWindowTexture renders the 3D planet scene to the window texture
func (g *Game) renderWindowTexture() {
	// Clear with space background
	g.windowTex.Fill(color.RGBA{5, 8, 15, 255})

	// Add procedural stars
	g.drawStars()

	// Set camera position based on distance
	g.spaceScene.SetCameraPosition(0, 0, g.planetDist)
	g.spaceScene.LookAt(0, 0, 0)

	// Render 3D scene
	img3d := g.spaceScene.Render()

	// Composite 3D render onto window texture
	g.windowTex.DrawImage(img3d, nil)
}

// drawStars adds a procedural starfield to the window texture
func (g *Game) drawStars() {
	// Simple deterministic starfield
	for i := 0; i < 500; i++ {
		// Use hash to get deterministic positions
		h := uint32(i * 2654435761)
		x := int(h % uint32(windowTexWidth))
		h = h * 2654435761
		y := int(h % uint32(windowTexHeight))
		h = h * 2654435761
		brightness := uint8(100 + h%156)

		g.windowTex.Set(x, y, color.RGBA{brightness, brightness, brightness, 255})

		// Some stars get glow
		if brightness > 200 && x > 1 && y > 1 && x < windowTexWidth-2 && y < windowTexHeight-2 {
			dim := brightness / 2
			g.windowTex.Set(x-1, y, color.RGBA{dim, dim, dim, 255})
			g.windowTex.Set(x+1, y, color.RGBA{dim, dim, dim, 255})
			g.windowTex.Set(x, y-1, color.RGBA{dim, dim, dim, 255})
			g.windowTex.Set(x, y+1, color.RGBA{dim, dim, dim, 255})
		}
	}
}

func (g *Game) saveScreenshot(screen *ebiten.Image) {
	f, err := os.Create(*screenshotPath)
	if err != nil {
		log.Printf("Failed to create screenshot: %v", err)
		os.Exit(1)
	}
	defer f.Close()

	if err := png.Encode(f, screen); err != nil {
		log.Printf("Failed to encode screenshot: %v", err)
		os.Exit(1)
	}

	fmt.Printf("Screenshot saved to %s\n", *screenshotPath)
	os.Exit(0)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	flag.Parse()

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Interior + 3D Planet Demo")
	ebiten.SetCursorMode(ebiten.CursorModeCaptured)

	// Suppress unused import warning
	_ = math.Pi

	game := NewGame()

	if err := ebiten.RunGame(game); err != nil && err != ebiten.Termination {
		log.Fatal(err)
	}
}
