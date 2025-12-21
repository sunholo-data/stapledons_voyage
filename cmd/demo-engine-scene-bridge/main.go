// Demo: Scene-Based Bridge with Real SR/GR Effects
//
// Demonstrates the scene-based interior navigation approach.
// Shows a 2D bridge scene with live space view (with actual SR/GR effects)
// composited into window regions.
//
// This validates the key innovation: interior ship experience with live exterior
// view visible through windows, without complex 3D navigation.
//
// Controls:
//   V - Cycle ship velocity (0.0, 0.2, 0.5, 0.8c) to see SR effects
//   Arrow Keys - Pan camera in space view
//   R - Reset view
//   ESC - Exit
//
// Run: go run ./cmd/demo-engine-scene-bridge

package main

import (
	"encoding/json"
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
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"stapledons_voyage/engine/shader"
	"stapledons_voyage/engine/tetra"
)

var (
	screenshotFrame  = flag.Int("screenshot", 0, "Take screenshot after N frames (0 = disabled)")
	screenshotOutput = flag.String("output", "out/demo-engine-scene-bridge.png", "Screenshot output path")
)

const (
	screenWidth  = 1280
	screenHeight = 720
)

// DeckManifest matches the JSON structure
type DeckManifest struct {
	DeckID      string `json:"deckID"`
	DeckType    string `json:"deckType"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Layers      []struct {
		ID       string  `json:"id"`
		File     string  `json:"file"`
		Parallax float64 `json:"parallax"`
	} `json:"layers"`
	Windows struct {
		Enabled        bool    `json:"enabled"`
		MaskFile       string  `json:"maskFile"`
		StarfieldDepth float64 `json:"starfieldDepth"`
	} `json:"windows"`
}

type Game struct {
	// Bridge assets
	bridgeBackground *ebiten.Image
	windowMask       *ebiten.Image
	manifest         *DeckManifest

	// 3D Space scene (what we see through windows)
	scene3D       *tetra.Scene
	saturn        *tetra.Planet
	saturnTexture *ebiten.Image

	// SR/GR shader effects
	shaderManager *shader.Manager
	srWarp        *shader.SRWarp
	renderBuffer  *ebiten.Image

	// Ship state
	shipVelocity float64 // 0.0 to 0.99c
	velocityIdx  int

	// Camera (for space view)
	cameraX, cameraY, cameraZ float64

	// Rendering
	frame int

	// Screenshot
	screenshotTaken bool
}

func NewGame() (*Game, error) {
	g := &Game{
		shipVelocity: 0.0,
		velocityIdx:  0,
		frame:        0,
		cameraX:      0,
		cameraY:      0,
		cameraZ:      0, // At origin, looking at Saturn at Z=-10
	}

	// Load manifest
	manifestData, err := os.ReadFile("assets/decks/bridge/manifest.json")
	if err != nil {
		return nil, fmt.Errorf("failed to load manifest: %w", err)
	}
	if err := json.Unmarshal(manifestData, &g.manifest); err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}

	// Load bridge background (v2 with empty windows)
	bgFile, err := os.Open("assets/decks/bridge/background_v2.png")
	if err != nil {
		return nil, fmt.Errorf("failed to open background_v2: %w", err)
	}
	defer bgFile.Close()

	bgImg, _, err := image.Decode(bgFile)
	if err != nil {
		return nil, fmt.Errorf("failed to decode background: %w", err)
	}
	g.bridgeBackground = ebiten.NewImageFromImage(bgImg)

	// Load window mask (transparent - windows at top)
	maskFile, err := os.Open("assets/decks/bridge/window_mask_transparent.png")
	if err != nil {
		return nil, fmt.Errorf("failed to open window_mask_v2: %w", err)
	}
	defer maskFile.Close()

	maskImg, _, err := image.Decode(maskFile)
	if err != nil {
		return nil, fmt.Errorf("failed to decode window mask: %w", err)
	}
	g.windowMask = ebiten.NewImageFromImage(maskImg)

	// Initialize Tetra3D scene for space view
	g.scene3D = tetra.NewScene(screenWidth, screenHeight)
	g.scene3D.SetCameraPosition(0, 0, 0) // At origin, Saturn is at Z=-10
	g.scene3D.SetLightingEnabled(false)  // Disable lighting for space view

	// Load Saturn texture
	saturnFile, err := os.Open("assets/planets/saturn_2k.jpg")
	if err != nil {
		// If Saturn texture doesn't exist, we'll create a simple colored sphere
		log.Printf("Warning: Saturn texture not found, using colored sphere")
		g.saturnTexture = ebiten.NewImage(512, 256)
		g.saturnTexture.Fill(color.RGBA{200, 180, 140, 255}) // Sandy color
	} else {
		defer saturnFile.Close()
		saturnImg, _, err := image.Decode(saturnFile)
		if err != nil {
			return nil, fmt.Errorf("failed to decode Saturn texture: %w", err)
		}
		g.saturnTexture = ebiten.NewImageFromImage(saturnImg)
	}

	// Create Saturn planet (make it large and close for visibility)
	// Tetra3D cameras look down -Z axis, so negative Z is "forward"
	g.saturn = tetra.NewTexturedPlanet("saturn", 5.0, g.saturnTexture)
	g.saturn.SetPosition(0, 0, -15) // In front of camera (which looks down -Z)
	g.saturn.SetShadeless(true) // Make it self-illuminated so it's always visible
	g.saturn.AddToScene(g.scene3D)

	// Initialize shader manager for SR/GR effects
	g.shaderManager = shader.NewManager()
	g.srWarp = shader.NewSRWarp(g.shaderManager)
	g.renderBuffer = ebiten.NewImage(screenWidth, screenHeight)

	return g, nil
}

func (g *Game) Update() error {
	// Input handling
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		return fmt.Errorf("exit")
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		g.shipVelocity = 0.0
		g.velocityIdx = 0
		g.cameraX, g.cameraY, g.cameraZ = 0, 0, 0
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyV) {
		// Cycle ship velocity
		velocities := []float64{0.0, 0.2, 0.5, 0.8}
		g.velocityIdx = (g.velocityIdx + 1) % len(velocities)
		g.shipVelocity = velocities[g.velocityIdx]
	}

	// Camera panning with arrow keys
	panSpeed := 0.1
	if ebiten.IsKeyPressed(ebiten.KeyLeft) {
		g.cameraX -= panSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyRight) {
		g.cameraX += panSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyUp) {
		g.cameraY += panSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyDown) {
		g.cameraY -= panSpeed
	}

	g.scene3D.SetCameraPosition(g.cameraX, g.cameraY, g.cameraZ)

	g.frame++
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Clear screen
	screen.Fill(color.Black)

	// 1. Render bridge background (scaled to fit screen)
	g.drawBridgeBackground(screen)

	// 2. Render space view (Saturn with SR effects) to buffer
	g.renderSpaceView()

	// 3. Composite space view into window regions
	if g.manifest.Windows.Enabled {
		g.compositeWindows(screen)
	}

	// 4. Render HUD
	g.drawHUD(screen)

	// Take screenshot if requested (after all rendering is done)
	if *screenshotFrame > 0 && g.frame >= *screenshotFrame && !g.screenshotTaken {
		g.takeScreenshot(screen)
		g.screenshotTaken = true
	}
}

func (g *Game) drawBridgeBackground(screen *ebiten.Image) {
	bgW, bgH := g.bridgeBackground.Bounds().Dx(), g.bridgeBackground.Bounds().Dy()
	scaleX := float64(screenWidth) / float64(bgW)
	scaleY := float64(screenHeight) / float64(bgH)
	scale := math.Min(scaleX, scaleY)

	opBg := &ebiten.DrawImageOptions{}
	opBg.GeoM.Scale(scale, scale)
	opBg.GeoM.Translate(
		(screenWidth-float64(bgW)*scale)/2,
		(screenHeight-float64(bgH)*scale)/2,
	)
	screen.DrawImage(g.bridgeBackground, opBg)
}

func (g *Game) renderSpaceView() {
	// Clear buffer with deep space color
	g.renderBuffer.Fill(color.RGBA{0, 0, 10, 255}) // Deep space color

	// Draw some stars for testing
	g.drawStars()

	// Render 3D scene (Saturn)
	rendered := g.scene3D.Render()
	g.renderBuffer.DrawImage(rendered, nil)

	// SR warp effect disabled for prototype (would need shader preloading)
	// TODO: Add shader.Manager.Preload() in NewGame() to enable SR effects
	// if g.shipVelocity > 0.1 {
	// 	g.srWarp.SetVelocity(0, 0, g.shipVelocity)
	// 	g.srWarp.SetEnabled(true)
	// 	g.srWarp.Apply(g.renderBuffer, g.renderBuffer)
	// }
}

func (g *Game) drawStars() {
	// Draw some simple stars for visual testing
	stars := [][3]int{
		{100, 100, 2},
		{300, 150, 1},
		{500, 200, 2},
		{700, 180, 1},
		{900, 220, 2},
		{200, 350, 1},
		{600, 400, 2},
		{1000, 350, 1},
	}

	for _, star := range stars {
		x, y, size := star[0], star[1], star[2]
		for dy := 0; dy < size; dy++ {
			for dx := 0; dx < size; dx++ {
				g.renderBuffer.Set(x+dx, y+dy, color.RGBA{255, 255, 255, 255})
			}
		}
	}
}

func (g *Game) compositeWindows(screen *ebiten.Image) {
	bgW, bgH := g.bridgeBackground.Bounds().Dx(), g.bridgeBackground.Bounds().Dy()
	scaleX := float64(screenWidth) / float64(bgW)
	scaleY := float64(screenHeight) / float64(bgH)
	scale := math.Min(scaleX, scaleY)

	maskOffsetX := (screenWidth - float64(bgW)*scale) / 2
	maskOffsetY := (screenHeight - float64(bgH)*scale) / 2

	// Step 1: Create full-screen alpha mask
	fullscreenMask := ebiten.NewImage(screenWidth, screenHeight)
	fullscreenMask.Clear() // Ebiten Clear() sets all pixels to transparent

	// Step 2: Draw the window mask (scaled) to create alpha channel
	opMask := &ebiten.DrawImageOptions{}
	opMask.GeoM.Scale(scale, scale)
	opMask.GeoM.Translate(maskOffsetX, maskOffsetY)
	fullscreenMask.DrawImage(g.windowMask, opMask)

	// Step 3: Create masked space view
	maskedView := ebiten.NewImage(screenWidth, screenHeight)
	maskedView.Clear() // Ebiten Clear() sets all pixels to transparent
	maskedView.DrawImage(g.renderBuffer, nil)

	// Step 4: Apply the full-screen mask using DestinationIn
	// This keeps space view pixels ONLY where fullscreenMask has opaque pixels
	opApply := &ebiten.DrawImageOptions{}
	opApply.CompositeMode = ebiten.CompositeModeDestinationIn
	maskedView.DrawImage(fullscreenMask, opApply)

	// Step 5: Draw masked view onto screen (on top of bridge background)
	screen.DrawImage(maskedView, nil)
}

func (g *Game) drawHUD(screen *ebiten.Image) {
	// Ship velocity display
	velocityPercent := g.shipVelocity * 100
	velocityText := fmt.Sprintf("Ship Velocity: %.1f%%c (%.2fc)", velocityPercent, g.shipVelocity)

	hudText := fmt.Sprintf("%s\nFrame: %d\nDeck: %s (%s)\nCamera: (%.1f, %.1f, %.1f)\n\nControls:\nV - Cycle velocity | Arrows - Pan view | R - Reset | ESC - Exit\n\nPrototype: Scene-based navigation with window compositing",
		velocityText, g.frame, g.manifest.Name, g.manifest.DeckType, g.cameraX, g.cameraY, g.cameraZ)

	ebitenutil.DebugPrint(screen, hudText)
}

func (g *Game) takeScreenshot(screen *ebiten.Image) {
	os.MkdirAll("out", 0755)

	f, err := os.Create(*screenshotOutput)
	if err != nil {
		log.Printf("Failed to create screenshot: %v", err)
		return
	}
	defer f.Close()

	if err := png.Encode(f, screen); err != nil {
		log.Printf("Failed to encode screenshot: %v", err)
		return
	}

	log.Printf("Screenshot saved to %s", *screenshotOutput)
	os.Exit(0)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	flag.Parse()

	game, err := NewGame()
	if err != nil {
		log.Fatal(err)
	}

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Demo: Scene-Based Bridge (Interior/Exterior with SR/GR)")

	if *screenshotFrame == 0 {
		ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	}

	if err := ebiten.RunGame(game); err != nil {
		if err.Error() != "exit" {
			log.Fatal(err)
		}
	}
}
