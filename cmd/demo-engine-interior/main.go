// Package main provides a 3D interior rendering engine demo.
// This demonstrates that the Tetra3D engine can render interior rooms
// with high-quality AI-generated textures.
//
// Usage:
//
//	go run ./cmd/demo-engine-interior
//	go run ./cmd/demo-engine-interior -floor assets/textures/floor.png -wall assets/textures/wall.png
//	go run ./cmd/demo-engine-interior -floor path/to/floor.png -wall path/to/wall.png -ceiling path/to/ceiling.png
package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	_ "image/png"
	"log"
	"math"
	"os"

	"stapledons_voyage/engine/tetra"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/solarlune/tetra3d"
)

const (
	screenWidth  = 1280
	screenHeight = 720
)

var (
	// Screenshot flags
	screenshotFrame = flag.Int("screenshot", 0, "Take screenshot at frame N and exit")
	screenshotPath  = flag.String("output", "out/screenshots/engine-interior.png", "Screenshot output path")

	// Texture flags
	floorTexture   = flag.String("floor", "", "Path to floor texture image")
	wallTexture    = flag.String("wall", "", "Path to wall texture image")
	ceilingTexture = flag.String("ceiling", "", "Path to ceiling texture image")
	consoleTexture = flag.String("console", "", "Path to console texture image")

	// Room dimensions
	roomWidth  = flag.Float64("width", 8, "Room width in meters")
	roomDepth  = flag.Float64("depth", 6, "Room depth in meters")
	roomHeight = flag.Float64("height", 3, "Room height in meters")

	// Rendering options
	useLighting = flag.Bool("lighting", false, "Enable dynamic lighting (default: shadeless)")
)

type Game struct {
	scene       *tetra.Scene
	room        *tetra.Room
	frameCount  int
	initialized bool

	// Loaded textures
	floorTex   *ebiten.Image
	wallTex    *ebiten.Image
	ceilingTex *ebiten.Image
	consoleTex *ebiten.Image

	// Camera controls
	cameraYaw   float64 // Horizontal rotation
	cameraPitch float64 // Vertical rotation
	cameraX     float64 // Position
	cameraY     float64
	cameraZ     float64
}

func NewGame() *Game {
	return &Game{
		cameraY:     1.7,  // Eye level
		cameraZ:     1.0,  // Start in center of room
		cameraPitch: -0.3, // Look slightly down to see floor/consoles
		cameraYaw:   0,
	}
}

// loadTexture loads an image file as an ebiten.Image
func loadTexture(path string) (*ebiten.Image, error) {
	if path == "" {
		return nil, nil
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening %s: %w", path, err)
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("decoding %s: %w", path, err)
	}

	return ebiten.NewImageFromImage(img), nil
}

func (g *Game) loadTextures() error {
	var err error

	if *floorTexture != "" {
		g.floorTex, err = loadTexture(*floorTexture)
		if err != nil {
			return fmt.Errorf("floor texture: %w", err)
		}
		fmt.Printf("Loaded floor texture: %s\n", *floorTexture)
	}

	if *wallTexture != "" {
		g.wallTex, err = loadTexture(*wallTexture)
		if err != nil {
			return fmt.Errorf("wall texture: %w", err)
		}
		fmt.Printf("Loaded wall texture: %s\n", *wallTexture)
	}

	if *ceilingTexture != "" {
		g.ceilingTex, err = loadTexture(*ceilingTexture)
		if err != nil {
			return fmt.Errorf("ceiling texture: %w", err)
		}
		fmt.Printf("Loaded ceiling texture: %s\n", *ceilingTexture)
	}

	if *consoleTexture != "" {
		g.consoleTex, err = loadTexture(*consoleTexture)
		if err != nil {
			return fmt.Errorf("console texture: %w", err)
		}
		fmt.Printf("Loaded console texture: %s\n", *consoleTexture)
	}

	return nil
}

func (g *Game) init() error {
	// Load any textures first
	if err := g.loadTextures(); err != nil {
		return err
	}

	// Create 3D scene
	g.scene = tetra.NewScene(screenWidth, screenHeight)
	g.scene.SetFieldOfView(70)
	g.scene.SetNear(0.1)
	g.scene.SetFar(100)

	// Create room with specified dimensions
	g.room = tetra.NewRoom(float32(*roomWidth), float32(*roomDepth), float32(*roomHeight))

	// Create materials - use textures if provided, otherwise solid colors
	shadeless := !*useLighting

	floorMat := tetra3d.NewMaterial("floor")
	if g.floorTex != nil {
		floorMat.Texture = g.floorTex
		floorMat.Color = tetra3d.NewColor(1, 1, 1, 1) // Full brightness for texture
	} else {
		floorMat.Color = tetra3d.NewColor(0.3, 0.35, 0.4, 1) // Blue-gray metallic
	}
	floorMat.Shadeless = shadeless

	ceilingMat := tetra3d.NewMaterial("ceiling")
	if g.ceilingTex != nil {
		ceilingMat.Texture = g.ceilingTex
		ceilingMat.Color = tetra3d.NewColor(1, 1, 1, 1)
	} else {
		ceilingMat.Color = tetra3d.NewColor(0.2, 0.2, 0.25, 1) // Darker ceiling
	}
	ceilingMat.Shadeless = shadeless

	wallMat := tetra3d.NewMaterial("wall")
	if g.wallTex != nil {
		wallMat.Texture = g.wallTex
		wallMat.Color = tetra3d.NewColor(1, 1, 1, 1)
	} else {
		wallMat.Color = tetra3d.NewColor(0.35, 0.38, 0.45, 1) // Lighter walls
	}
	wallMat.Shadeless = shadeless

	g.room.SetFloorMaterial(floorMat)
	g.room.SetCeilingMaterial(ceilingMat)
	g.room.SetWallMaterial(wallMat)

	// Add room to scene
	g.room.AddToScene(g.scene)

	// Add some props (simple cubes for now)
	g.addConsoleCubes(shadeless)

	// Set up lighting
	g.scene.SetLightingEnabled(*useLighting)
	if *useLighting {
		g.setupLighting()
	}

	// Position camera
	g.updateCamera()

	g.initialized = true
	return nil
}

func (g *Game) addConsoleCubes(shadeless bool) {
	// Console material
	consoleMat := tetra3d.NewMaterial("console")
	if g.consoleTex != nil {
		consoleMat.Texture = g.consoleTex
		consoleMat.Color = tetra3d.NewColor(1, 1, 1, 1)
	} else {
		consoleMat.Color = tetra3d.NewColor(0.1, 0.5, 0.6, 1) // Cyan console
	}
	consoleMat.Shadeless = shadeless

	// Front console (helm station)
	helmMesh := tetra.NewCubeMesh("helm", 1.5)
	helm := tetra3d.NewModel("helm_model", helmMesh)
	helm.SetLocalPosition(0, 0.5, float32(-*roomDepth/2+1))
	helm.SetLocalScale(1.5, 0.6, 0.5)
	if len(helm.Mesh.MeshParts) > 0 {
		helm.Mesh.MeshParts[0].Material = consoleMat
	}
	g.scene.Root().AddChildren(helm)

	// Side consoles
	leftConsoleMesh := tetra.NewCubeMesh("left_console", 1)
	leftConsole := tetra3d.NewModel("left_console_model", leftConsoleMesh)
	leftConsole.SetLocalPosition(float32(-*roomWidth/2+1), 0.5, 0)
	leftConsole.SetLocalScale(0.5, 0.6, 1.2)
	if len(leftConsole.Mesh.MeshParts) > 0 {
		leftConsole.Mesh.MeshParts[0].Material = consoleMat
	}
	g.scene.Root().AddChildren(leftConsole)

	rightConsoleMesh := tetra.NewCubeMesh("right_console", 1)
	rightConsole := tetra3d.NewModel("right_console_model", rightConsoleMesh)
	rightConsole.SetLocalPosition(float32(*roomWidth/2-1), 0.5, 0)
	rightConsole.SetLocalScale(0.5, 0.6, 1.2)
	if len(rightConsole.Mesh.MeshParts) > 0 {
		rightConsole.Mesh.MeshParts[0].Material = consoleMat
	}
	g.scene.Root().AddChildren(rightConsole)

	// Captain's chair
	chairMat := tetra3d.NewMaterial("chair")
	chairMat.Color = tetra3d.NewColor(0.5, 0.4, 0.3, 1) // Brown-ish
	chairMat.Shadeless = shadeless

	chairMesh := tetra.NewCubeMesh("chair", 0.6)
	chair := tetra3d.NewModel("chair_model", chairMesh)
	chair.SetLocalPosition(0, 0.4, 1)
	chair.SetLocalScale(0.8, 0.8, 0.8)
	if len(chair.Mesh.MeshParts) > 0 {
		chair.Mesh.MeshParts[0].Material = chairMat
	}
	g.scene.Root().AddChildren(chair)
}

func (g *Game) setupLighting() {
	// Main overhead light
	mainLight := tetra.NewSunLight()
	mainLight.SetPosition(0, 5, 0)
	mainLight.SetColor(0.9, 0.95, 1.0) // Slightly cool white
	mainLight.SetEnergy(1.2)
	mainLight.AddToScene(g.scene)

	// Ambient fill light
	ambient := tetra.NewAmbientLight(0.3, 0.35, 0.4, 0.4) // Blue-ish ambient
	ambient.AddToScene(g.scene)
}

func (g *Game) updateCamera() {
	g.scene.SetCameraPosition(g.cameraX, g.cameraY, g.cameraZ)

	// Calculate look direction from yaw and pitch
	lookX := g.cameraX + math.Sin(g.cameraYaw)*math.Cos(g.cameraPitch)
	lookY := g.cameraY + math.Sin(g.cameraPitch)
	lookZ := g.cameraZ - math.Cos(g.cameraYaw)*math.Cos(g.cameraPitch)

	g.scene.LookAt(lookX, lookY, lookZ)
}

func (g *Game) Update() error {
	if !g.initialized {
		if err := g.init(); err != nil {
			return err
		}
	}

	// Mouse look
	dx, dy := ebiten.CursorPosition()
	g.cameraYaw = float64(dx-screenWidth/2) * 0.003
	g.cameraPitch = float64(screenHeight/2-dy) * 0.002
	// Clamp pitch
	if g.cameraPitch > 1.4 {
		g.cameraPitch = 1.4
	}
	if g.cameraPitch < -1.4 {
		g.cameraPitch = -1.4
	}

	// WASD movement
	moveSpeed := 0.05
	if ebiten.IsKeyPressed(ebiten.KeyShift) {
		moveSpeed = 0.1
	}

	// Calculate forward/right vectors from yaw
	forwardX := math.Sin(g.cameraYaw)
	forwardZ := -math.Cos(g.cameraYaw)
	rightX := math.Cos(g.cameraYaw)
	rightZ := math.Sin(g.cameraYaw)

	if ebiten.IsKeyPressed(ebiten.KeyW) {
		g.cameraX += forwardX * moveSpeed
		g.cameraZ += forwardZ * moveSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) {
		g.cameraX -= forwardX * moveSpeed
		g.cameraZ -= forwardZ * moveSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyA) {
		g.cameraX -= rightX * moveSpeed
		g.cameraZ -= rightZ * moveSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) {
		g.cameraX += rightX * moveSpeed
		g.cameraZ += rightZ * moveSpeed
	}

	// Vertical movement (for debugging)
	if ebiten.IsKeyPressed(ebiten.KeySpace) {
		g.cameraY += moveSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyControl) {
		g.cameraY -= moveSpeed
	}

	// Escape to quit
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}

	g.updateCamera()
	g.frameCount++

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	if !g.initialized {
		return
	}

	// Clear with dark color
	screen.Fill(color.RGBA{10, 10, 15, 255})

	// Render 3D scene
	img3d := g.scene.Render()
	screen.DrawImage(img3d, nil)

	// Draw HUD
	g.drawHUD(screen)

	// Screenshot
	if *screenshotFrame > 0 && g.frameCount >= *screenshotFrame {
		g.saveScreenshot(screen)
	}
}

func (g *Game) drawHUD(screen *ebiten.Image) {
	texInfo := ""
	if g.floorTex != nil || g.wallTex != nil || g.ceilingTex != nil {
		texInfo = " [Textured]"
	}
	lightInfo := ""
	if *useLighting {
		lightInfo = " [Lit]"
	}

	hudText := fmt.Sprintf(
		"Engine Interior Demo%s%s\n"+
			"WASD: Move | Mouse: Look | Shift: Fast | Esc: Quit\n"+
			"Room: %.0fx%.0fx%.0fm | Position: (%.1f, %.1f, %.1f)\n"+
			"Frame: %d",
		texInfo, lightInfo,
		*roomWidth, *roomDepth, *roomHeight,
		g.cameraX, g.cameraY, g.cameraZ, g.frameCount)

	ebitenutil.DebugPrint(screen, hudText)
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
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `Engine Interior Demo - 3D room rendering with texture support

Usage:
  demo-engine-interior [flags]

Examples:
  # Basic room with solid colors
  demo-engine-interior

  # Room with AI-generated textures
  demo-engine-interior -floor floor.png -wall wall.png

  # Custom room size with lighting
  demo-engine-interior -width 12 -depth 10 -height 4 -lighting

  # Generate screenshot
  demo-engine-interior -floor floor.png -screenshot 60 -output room.png

Flags:
`)
		flag.PrintDefaults()
	}

	flag.Parse()

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Engine Interior Demo")
	ebiten.SetCursorMode(ebiten.CursorModeCaptured)

	game := NewGame()

	if err := ebiten.RunGame(game); err != nil && err != ebiten.Termination {
		log.Fatal(err)
	}
}
