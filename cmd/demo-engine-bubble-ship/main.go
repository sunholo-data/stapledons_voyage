// Package main provides a demo of the transparent bubble ship external view.
// This tests rendering a semi-transparent sphere with a ship silhouette inside,
// viewed from outside - for use as a 3rd-person HUD element.
//
// Controls:
//
//	Mouse: Orbit camera around bubble
//	Scroll: Zoom in/out
//	A/D: Adjust bubble alpha transparency
//	C: Toggle ship silhouette color
//	S: Toggle stars/space background
//	R: Reset view
//	Esc: Quit
//
// Usage:
//
//	go run ./cmd/demo-engine-bubble-ship
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
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/solarlune/tetra3d"
	"stapledons_voyage/engine/tetra"
)

const (
	screenWidth  = 1280
	screenHeight = 720
)

var (
	screenshotFrame = flag.Int("screenshot", 0, "Take screenshot at frame N and exit")
	screenshotPath  = flag.String("output", "out/screenshots/bubble-ship.png", "Screenshot output path")
)

// Game implements ebiten.Game for the bubble ship demo.
type Game struct {
	scene      *tetra.Scene
	bubble     *tetra.Dome
	spire      *tetra3d.Model
	levels     []*tetra3d.Model
	starSphere *tetra.Dome
	frameCount int

	// Camera orbit
	orbitYaw    float64
	orbitPitch  float64
	orbitDist   float64
	lastX       int
	lastY       int
	dragging    bool

	// Bubble parameters
	bubbleAlpha float64
	bubbleR     float64
	bubbleG     float64
	bubbleB     float64

	// Toggles
	showStars bool
	spireColor int // 0=white, 1=gold, 2=blue

	// Performance
	lastUpdate time.Time
	fps        float64
}

// loadTexture loads an image file as an ebiten image
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

// NewGame creates a new bubble ship demo game.
func NewGame() *Game {
	g := &Game{
		orbitYaw:    0.5,
		orbitPitch:  0.3,
		orbitDist:   300,
		bubbleAlpha: 0.25,
		bubbleR:     0.3,
		bubbleG:     0.5,
		bubbleB:     0.7,
		showStars:   true,
		spireColor:  0,
		lastUpdate:  time.Now(),
	}

	// Create 3D scene
	g.scene = tetra.NewScene(screenWidth, screenHeight)
	g.scene.SetLightingEnabled(true)

	// Create star sphere (background) - large radius
	g.starSphere = tetra.NewBubble("star_sphere", 800)
	g.starSphere.AddToScene(g.scene)
	g.starSphere.SetPosition(0, 0, 0)
	// Load star texture if available
	if starTex := loadTexture("assets/textures/milkyway_2020_8k.jpg"); starTex != nil {
		g.starSphere.SetTexture(starTex)
	} else if starTex := loadTexture("assets/textures/starfield.png"); starTex != nil {
		g.starSphere.SetTexture(starTex)
	} else {
		// Dark blue default
		g.starSphere.SetColor(0.02, 0.02, 0.08, 1.0)
	}

	// Create transparent bubble (100m radius = the ship's Higgs bubble)
	const bubbleRadius = 100.0
	g.bubble = tetra.NewExternalBubble("ship_bubble", bubbleRadius, g.bubbleAlpha)
	g.bubble.AddToScene(g.scene)
	g.bubble.SetPosition(0, 0, 0)
	g.bubble.SetColor(g.bubbleR, g.bubbleG, g.bubbleB, g.bubbleAlpha)
	g.bubble.SetShadeless(false) // Allow some lighting on bubble surface

	// Create ship interior silhouette (spire + levels)
	g.createShipSilhouette(bubbleRadius)

	// Add sun light (simulating distant star)
	sun := tetra.NewSunLight()
	sun.AddToScene(g.scene)

	// Add ambient light
	ambient := tetra.NewAmbientLight(0.15, 0.15, 0.2, 0.8)
	ambient.AddToScene(g.scene)

	return g
}

// createShipSilhouette creates a simplified ship interior (spire + deck levels).
func (g *Game) createShipSilhouette(bubbleRadius float64) {
	// Central spire (Higgs Generator) - runs most of bubble height
	spireHeight := bubbleRadius * 1.6 // Extends from bottom to near top
	spireRadius := bubbleRadius * 0.08

	spireMesh := tetra3d.NewCylinderMesh(8, float32(spireRadius), float32(spireHeight), true)
	spireMat := tetra3d.NewMaterial("spire_mat")
	spireMat.Color = tetra3d.NewColor(0.9, 0.9, 0.95, 1.0) // Bright white/silver
	for _, part := range spireMesh.MeshParts {
		part.Material = spireMat
	}

	g.spire = tetra3d.NewModel("spire", spireMesh)
	g.spire.SetLocalPosition(0, float32(-bubbleRadius*0.3), 0) // Centered vertically
	g.scene.Root().AddChildren(g.spire)

	// Deck levels (horizontal discs around the spire)
	levelHeights := []float64{-0.4, -0.1, 0.2, 0.5} // Relative to bubble center
	levelRadii := []float64{0.6, 0.5, 0.4, 0.25}    // Decreasing toward top

	for i, height := range levelHeights {
		levelRadius := bubbleRadius * levelRadii[i]
		levelThickness := bubbleRadius * 0.03

		// Create disc mesh
		levelMesh := tetra3d.NewCylinderMesh(16, float32(levelRadius), float32(levelThickness), true)
		levelMat := tetra3d.NewMaterial(fmt.Sprintf("level_%d_mat", i))
		// Slight color variation per level
		brightness := 0.7 + float32(i)*0.05
		levelMat.Color = tetra3d.NewColor(brightness, brightness*0.95, brightness*0.9, 1.0)
		for _, part := range levelMesh.MeshParts {
			part.Material = levelMat
		}

		level := tetra3d.NewModel(fmt.Sprintf("level_%d", i), levelMesh)
		level.SetLocalPosition(0, float32(height*bubbleRadius), 0)
		g.scene.Root().AddChildren(level)
		g.levels = append(g.levels, level)
	}

	log.Printf("Created ship silhouette: spire + %d levels", len(g.levels))
}

// updateBubbleAppearance updates the bubble's visual properties.
func (g *Game) updateBubbleAppearance() {
	g.bubble.SetColor(g.bubbleR, g.bubbleG, g.bubbleB, g.bubbleAlpha)
}

// updateSpireColor cycles through color schemes for the ship interior.
func (g *Game) updateSpireColor() {
	var r, gr, b float32
	switch g.spireColor {
	case 0: // White/silver
		r, gr, b = 0.9, 0.9, 0.95
	case 1: // Gold
		r, gr, b = 0.95, 0.85, 0.5
	case 2: // Blue
		r, gr, b = 0.5, 0.6, 0.95
	}

	// Update spire
	for _, part := range g.spire.Mesh.MeshParts {
		part.Material.Color = tetra3d.NewColor(r, gr, b, 1.0)
	}

	// Update levels with slight variations
	for i, level := range g.levels {
		brightness := 0.7 + float32(i)*0.05
		for _, part := range level.Mesh.MeshParts {
			part.Material.Color = tetra3d.NewColor(r*brightness, gr*brightness, b*brightness, 1.0)
		}
	}
}

func (g *Game) Update() error {
	now := time.Now()
	dt := now.Sub(g.lastUpdate).Seconds()
	g.lastUpdate = now

	if dt > 0 {
		g.fps = g.fps*0.95 + (1/dt)*0.05
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}

	// Mouse orbit control
	x, y := ebiten.CursorPosition()

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		g.dragging = true
		g.lastX = x
		g.lastY = y
	}
	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		g.dragging = false
	}

	if g.dragging {
		dx := float64(x - g.lastX)
		dy := float64(y - g.lastY)

		g.orbitYaw += dx * 0.005
		g.orbitPitch += dy * 0.005

		// Clamp pitch
		if g.orbitPitch > math.Pi/2-0.1 {
			g.orbitPitch = math.Pi/2 - 0.1
		}
		if g.orbitPitch < -math.Pi/2+0.1 {
			g.orbitPitch = -math.Pi/2 + 0.1
		}

		g.lastX = x
		g.lastY = y
	}

	// Mouse scroll for zoom
	_, scrollY := ebiten.Wheel()
	if scrollY != 0 {
		g.orbitDist -= scrollY * 20
		if g.orbitDist < 150 {
			g.orbitDist = 150
		}
		if g.orbitDist > 600 {
			g.orbitDist = 600
		}
	}

	// Adjust bubble alpha
	if inpututil.IsKeyJustPressed(ebiten.KeyA) {
		g.bubbleAlpha -= 0.05
		if g.bubbleAlpha < 0.05 {
			g.bubbleAlpha = 0.05
		}
		g.updateBubbleAppearance()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyD) {
		g.bubbleAlpha += 0.05
		if g.bubbleAlpha > 0.8 {
			g.bubbleAlpha = 0.8
		}
		g.updateBubbleAppearance()
	}

	// Toggle ship color
	if inpututil.IsKeyJustPressed(ebiten.KeyC) {
		g.spireColor = (g.spireColor + 1) % 3
		g.updateSpireColor()
	}

	// Toggle stars
	if inpututil.IsKeyJustPressed(ebiten.KeyS) {
		g.showStars = !g.showStars
		g.starSphere.SetVisible(g.showStars)
	}

	// Reset view
	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		g.orbitYaw = 0.5
		g.orbitPitch = 0.3
		g.orbitDist = 300
		g.bubbleAlpha = 0.25
		g.updateBubbleAppearance()
	}

	// Auto-rotate slowly for visual interest
	g.orbitYaw += dt * 0.1

	// Update camera position (orbit around bubble)
	camX := g.orbitDist * math.Cos(g.orbitPitch) * math.Sin(g.orbitYaw)
	camY := g.orbitDist * math.Sin(g.orbitPitch)
	camZ := g.orbitDist * math.Cos(g.orbitPitch) * math.Cos(g.orbitYaw)

	g.scene.SetCameraPosition(camX, camY, camZ)
	g.scene.LookAt(0, 0, 0) // Look at bubble center

	g.frameCount++
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Clear with space black
	screen.Fill(color.RGBA{2, 2, 5, 255})

	// Render the 3D scene
	rendered := g.scene.Render()
	screen.DrawImage(rendered, nil)

	// HUD
	colorNames := []string{"White/Silver", "Gold", "Blue"}
	starsStatus := "ON"
	if !g.showStars {
		starsStatus = "OFF"
	}

	hudText := fmt.Sprintf(
		"Transparent Bubble Ship Demo\n"+
			"═══════════════════════════════════════════════\n"+
			"Mouse Drag: Orbit | Scroll: Zoom\n"+
			"A/D: Adjust Alpha (%.0f%%)\n"+
			"C: Cycle Ship Color (%s)\n"+
			"S: Toggle Stars (%s)\n"+
			"R: Reset | Esc: Quit\n"+
			"═══════════════════════════════════════════════\n"+
			"FPS: %.1f | Frame: %d\n"+
			"Orbit: Yaw %.1f° Pitch %.1f° Dist %.0f\n"+
			"Bubble: RGBA(%.2f, %.2f, %.2f, %.2f)\n"+
			"═══════════════════════════════════════════════\n"+
			"This demonstrates the transparent bubble ship\n"+
			"as it would appear in a 3rd-person HUD view.\n"+
			"The sphere is semi-transparent so stars show\n"+
			"through, with ship silhouette visible inside.",
		g.bubbleAlpha*100,
		colorNames[g.spireColor],
		starsStatus,
		g.fps,
		g.frameCount,
		g.orbitYaw*180/math.Pi,
		g.orbitPitch*180/math.Pi,
		g.orbitDist,
		g.bubbleR, g.bubbleG, g.bubbleB, g.bubbleAlpha,
	)
	ebitenutil.DebugPrint(screen, hudText)

	// Screenshot
	if *screenshotFrame > 0 && g.frameCount >= *screenshotFrame {
		g.saveScreenshot(screen)
	}
}

func (g *Game) saveScreenshot(screen *ebiten.Image) {
	// Ensure output directory exists
	os.MkdirAll("out/screenshots", 0755)

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

	fmt.Println("Transparent Bubble Ship Demo")
	fmt.Println("  Testing semi-transparent sphere with ship interior")
	fmt.Println("  For use as 3rd-person HUD element")

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Transparent Bubble Ship Demo")

	game := NewGame()

	if err := ebiten.RunGame(game); err != nil && err != ebiten.Termination {
		log.Fatal(err)
	}
}
