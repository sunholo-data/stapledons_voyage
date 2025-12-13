// Package main provides a stress test demo for the LOD (Level of Detail) system.
// It renders celestial objects with automatic LOD tier switching including actual 3D planets.
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
	"math/rand"
	"os"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

	"stapledons_voyage/engine/lod"
	"stapledons_voyage/engine/tetra"
)

const (
	screenWidth  = 1280
	screenHeight = 720
)

// Planet3D represents a planet that can be rendered in 3D
type Planet3D struct {
	lodObj    *lod.Object
	planet    *tetra.Planet
	texture   *ebiten.Image
	billboard *ebiten.Image // Billboard sprite generated from texture
}

// Game implements ebiten.Game interface for the LOD demo.
type Game struct {
	lodManager        *lod.Manager
	camera            *lod.SimpleCamera
	pointRenderer     *lod.PointRenderer
	circleRenderer    *lod.CircleRenderer
	billboardRenderer *lod.BillboardRenderer

	// Tetra3D scene for Full3D rendering
	scene3D  *tetra.Scene
	planets  []*Planet3D
	testMode bool

	// Default sprite for billboards
	defaultSprite *ebiten.Image
	// Planet-specific billboard sprites (keyed by LOD object ID)
	planetSprites map[string]*ebiten.Image

	// Camera movement
	cameraSpeed float64

	// Screenshot handling
	screenshotFrame int
	screenshotPath  string
	frameCount      int

	// Performance tracking
	lastUpdate time.Time
	fps        float64

	// Lazy initialization flag for billboards
	billboardsInitialized bool
}

// NewGame creates a new LOD demo with the specified number of objects.
func NewGame(objectCount int, screenshotFrame int, screenshotPath string, testMode bool) *Game {
	// Create LOD manager
	config := lod.DefaultConfig()
	config.Max3DObjects = 20
	manager := lod.NewManager(config)

	// Create camera
	camera := lod.NewSimpleCamera(screenWidth, screenHeight)
	camera.Fov = 60
	camera.Far = 20000

	// Create Tetra3D scene
	scene3D := tetra.NewScene(screenWidth, screenHeight)

	var planets []*Planet3D

	if testMode {
		// Test mode: 4 planets at specific distances
		// Start far away so you can see the full LOD progression
		camera.Pos = lod.Vector3{X: 0, Y: 50, Z: 500}
		camera.LookAt = lod.Vector3{X: 0, Y: 0, Z: 0}
		planets = createTestPlanets(manager, scene3D)
	} else {
		// Normal mode: random star field
		camera.Pos = lod.Vector3{X: 0, Y: 0, Z: 500}
		camera.LookAt = lod.Vector3{X: 0, Y: 0, Z: 0}
		rng := rand.New(rand.NewSource(42))
		generateObjects(manager, rng, objectCount)
	}

	// Create billboard renderer with larger default sprite (128x128)
	billboardRenderer := lod.NewBillboardRenderer()
	defaultSprite := lod.CreateDefaultPlanetSprite(128, color.RGBA{255, 255, 255, 255})
	billboardRenderer.SetDefaultSprite(defaultSprite)

	// Add lighting to scene
	sun := tetra.NewSunLight()
	sun.AddToScene(scene3D)
	ambient := tetra.NewAmbientLight(0.3, 0.3, 0.3, 1.0)
	ambient.AddToScene(scene3D)

	// Build sprite map from planet billboards
	planetSprites := make(map[string]*ebiten.Image)
	for _, p := range planets {
		if p.billboard != nil {
			planetSprites[p.lodObj.ID] = p.billboard
		}
	}

	return &Game{
		lodManager:        manager,
		camera:            camera,
		pointRenderer:     lod.NewPointRenderer(),
		circleRenderer:    lod.NewCircleRenderer(),
		billboardRenderer: billboardRenderer,
		scene3D:           scene3D,
		planets:           planets,
		testMode:          testMode,
		defaultSprite:     defaultSprite,
		planetSprites:     planetSprites,
		cameraSpeed:       50.0,
		screenshotFrame:   screenshotFrame,
		screenshotPath:    screenshotPath,
		lastUpdate:        time.Now(),
	}
}

// createTestPlanets creates 4 test planets with real textures at known positions
func createTestPlanets(manager *lod.Manager, scene3D *tetra.Scene) []*Planet3D {
	planets := make([]*Planet3D, 0, 4)

	// Planet definitions: name, position, radius, texture path, fallback color
	defs := []struct {
		name     string
		pos      lod.Vector3
		radius   float64
		texPath  string
		color    color.RGBA
	}{
		{"Sun", lod.Vector3{X: 0, Y: 0, Z: 0}, 15.0, "assets/planets/sun.jpg", color.RGBA{255, 200, 50, 255}},
		{"Earth", lod.Vector3{X: 60, Y: 0, Z: 20}, 8.0, "assets/planets/earth.jpg", color.RGBA{50, 100, 200, 255}},
		{"Jupiter", lod.Vector3{X: -80, Y: 10, Z: -30}, 12.0, "assets/planets/jupiter.jpg", color.RGBA{200, 150, 100, 255}},
		{"Neptune", lod.Vector3{X: 0, Y: -20, Z: -100}, 6.0, "assets/planets/neptune.jpg", color.RGBA{50, 100, 200, 255}},
	}

	for _, def := range defs {
		// Create LOD object
		lodObj := lod.NewObject(def.name, def.pos, def.radius, def.color)
		manager.Add(lodObj)

		// Load texture
		tex := loadTexture(def.texPath)

		// Create 3D planet
		var planet *tetra.Planet
		if tex != nil {
			planet = tetra.NewTexturedPlanet(def.name, def.radius, tex)
			log.Printf("Created textured %s", def.name)
		} else {
			planet = tetra.NewPlanet(def.name, def.radius, def.color)
			log.Printf("Created solid %s (no texture)", def.name)
		}
		planet.AddToScene(scene3D)
		planet.SetPosition(def.pos.X, def.pos.Y, def.pos.Z)

		// Note: Billboard sprites are created lazily in first Update() call
		// because texture.At() requires the game loop to be running
		planets = append(planets, &Planet3D{
			lodObj:    lodObj,
			planet:    planet,
			texture:   tex,
			billboard: nil, // Created lazily
		})
	}

	return planets
}

// initializeBillboards creates billboard sprites from planet textures.
// This must be called after the game loop starts because texture.At()
// internally calls ReadPixels which requires the game loop to be running.
func (g *Game) initializeBillboards() {
	for _, p := range g.planets {
		if p.texture != nil && p.billboard == nil {
			// Extract average color from texture for circle/point rendering
			avgColor := lod.ExtractAverageColor(p.texture)
			p.lodObj.Color = avgColor
			log.Printf("Extracted color for %s: R=%d G=%d B=%d", p.lodObj.ID, avgColor.R, avgColor.G, avgColor.B)

			// Create billboard from the planet's texture
			p.billboard = lod.CreateBillboardFromTexture(p.texture, 128)
			g.planetSprites[p.lodObj.ID] = p.billboard
			log.Printf("Created billboard for %s from texture", p.lodObj.ID)
		} else if p.billboard == nil {
			// No texture, use a colored procedural billboard
			p.billboard = lod.CreateDefaultPlanetSprite(128, p.lodObj.Color)
			g.planetSprites[p.lodObj.ID] = p.billboard
			log.Printf("Created procedural billboard for %s", p.lodObj.ID)
		}
	}
}

// loadTexture loads an image file as an ebiten image
func loadTexture(path string) *ebiten.Image {
	f, err := os.Open(path)
	if err != nil {
		log.Printf("Warning: Could not load texture %s: %v", path, err)
		return nil
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		log.Printf("Warning: Could not decode texture %s: %v", path, err)
		return nil
	}

	return ebiten.NewImageFromImage(img)
}

// generateObjects creates random celestial objects distributed in 3D space.
func generateObjects(manager *lod.Manager, rng *rand.Rand, count int) {
	colors := []color.RGBA{
		{255, 255, 255, 255},
		{255, 255, 200, 255},
		{255, 200, 200, 255},
		{200, 200, 255, 255},
		{255, 230, 180, 255},
		{180, 255, 180, 255},
	}

	for i := 0; i < count; i++ {
		var pos lod.Vector3
		var radius float64
		var col color.RGBA

		distFactor := rng.Float64()
		var distance float64
		if distFactor < 0.1 {
			distance = rng.Float64() * 200
		} else if distFactor < 0.6 {
			distance = 200 + rng.Float64()*2000
		} else {
			distance = 2000 + rng.Float64()*8000
		}

		theta := rng.Float64() * 2 * math.Pi
		phi := rng.Float64()*math.Pi - math.Pi/2

		pos.X = distance * math.Cos(phi) * math.Cos(theta)
		pos.Y = distance * math.Sin(phi)
		pos.Z = distance * math.Cos(phi) * math.Sin(theta)

		objType := rng.Float64()
		if objType < 0.1 {
			radius = 5 + rng.Float64()*15
			col = colors[rng.Intn(len(colors))]
		} else if objType < 0.3 {
			radius = 2 + rng.Float64()*5
			col = colors[rng.Intn(3)]
		} else {
			radius = 0.5 + rng.Float64()*2
			starColors := []color.RGBA{
				{255, 255, 255, 255},
				{255, 255, 220, 255},
				{220, 220, 255, 255},
			}
			col = starColors[rng.Intn(len(starColors))]
		}

		obj := lod.NewObject(fmt.Sprintf("obj_%d", i), pos, radius, col)
		manager.Add(obj)
	}
}

// Update handles input and updates game state.
func (g *Game) Update() error {
	now := time.Now()
	dt := now.Sub(g.lastUpdate).Seconds()
	g.lastUpdate = now

	if dt > 0 {
		g.fps = g.fps*0.95 + (1/dt)*0.05
	}

	// Lazy initialize billboard sprites (must happen after game loop starts)
	// because texture.At() internally calls ReadPixels which requires the game loop
	if !g.billboardsInitialized && g.testMode {
		g.initializeBillboards()
		g.billboardsInitialized = true
	}

	// Camera movement
	moveSpeed := g.cameraSpeed * dt
	if ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyUp) {
		g.camera.Pos.Z -= moveSpeed * 2
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyDown) {
		g.camera.Pos.Z += moveSpeed * 2
	}
	if ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyLeft) {
		g.camera.Pos.X -= moveSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyRight) {
		g.camera.Pos.X += moveSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyQ) {
		g.camera.Pos.Y += moveSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyE) {
		g.camera.Pos.Y -= moveSpeed
	}

	if ebiten.IsKeyPressed(ebiten.KeyShift) {
		g.cameraSpeed = 150.0
	} else {
		g.cameraSpeed = 50.0
	}

	if ebiten.IsKeyPressed(ebiten.KeyR) {
		if g.testMode {
			g.camera.Pos = lod.Vector3{X: 0, Y: 20, Z: 150}
		} else {
			g.camera.Pos = lod.Vector3{X: 0, Y: 0, Z: 500}
		}
	}

	// Update LOD manager with explicit delta time for smooth transitions
	g.lodManager.UpdateWithDT(g.camera, dt)

	// Update 3D scene camera to match LOD camera
	if g.scene3D != nil {
		g.scene3D.SetCameraPosition(
			g.camera.Pos.X,
			g.camera.Pos.Y,
			g.camera.Pos.Z,
		)
		g.scene3D.LookAt(
			g.camera.LookAt.X,
			g.camera.LookAt.Y,
			g.camera.LookAt.Z,
		)
	}

	// Update planet visibility based on LOD tier
	if g.testMode {
		for _, p := range g.planets {
			// Only show 3D model when in Full3D tier
			if p.lodObj.CurrentTier == lod.TierFull3D {
				p.planet.Model().SetVisible(true, true)
			} else {
				p.planet.Model().SetVisible(false, true)
			}
			// Rotate planets
			p.planet.Update(dt)
		}
	}

	g.frameCount++
	return nil
}

// Draw renders the game.
func (g *Game) Draw(screen *ebiten.Image) {
	// Clear to dark space color
	screen.Fill(color.RGBA{5, 5, 15, 255})

	// Get objects by tier
	points := g.lodManager.GetTierPoint()
	circles := g.lodManager.GetTierCircle()
	billboards := g.lodManager.GetTierBillboard()
	full3D := g.lodManager.GetTier3D()
	transitioning := g.lodManager.GetTransitioning()

	// Layer 1: Render distant objects (points)
	// Use scaled points that grow as they approach circle threshold for smoother transition
	config := g.lodManager.Config()
	g.pointRenderer.RenderPointsScaled(screen, points, config.CirclePixels)

	// Layer 2: Render medium-distance objects (circles)
	g.circleRenderer.RenderCircles(screen, circles)

	// Layer 3: Render billboard tier (non-Full3D close objects)
	// Use planetSprites for texture-based billboards in test mode
	g.billboardRenderer.RenderBillboards(screen, billboards, g.planetSprites)

	// Layer 4: Render 3D scene (Full3D tier)
	if g.testMode && len(full3D) > 0 {
		img3d := g.scene3D.Render()
		screen.DrawImage(img3d, nil)
	} else if !g.testMode && len(full3D) > 0 {
		// For non-test mode, render Full3D as glowing circles (no 3D planets)
		g.circleRenderer.RenderCirclesWithGlow(screen, full3D, 1.5)
	}

	// Layer 5: Render transitioning objects with blending
	// Objects transitioning between tiers are rendered with alpha to smoothly
	// fade between their old and new representations
	for _, obj := range transitioning {
		// Skip objects already rendered in their target tier above
		// We only need to render the "fading out" previous tier representation

		prevAlpha := obj.PreviousAlpha() // 1.0 → 0.0 as transition progresses

		// Render previous tier with fading alpha
		switch obj.PreviousTier {
		case lod.TierPoint:
			g.pointRenderer.RenderPointWithAlpha(screen, obj, prevAlpha)
		case lod.TierCircle:
			g.circleRenderer.RenderCircleWithAlpha(screen, obj, prevAlpha)
		case lod.TierBillboard:
			g.billboardRenderer.RenderBillboardWithAlpha(screen, obj, prevAlpha, g.planetSprites)
		case lod.TierFull3D:
			// For 3D, we'd need to fade the mesh - for now just render as glowing circle
			if !g.testMode {
				g.circleRenderer.RenderCircleWithAlpha(screen, obj, prevAlpha)
			}
		}
	}

	// Draw stats overlay
	stats := g.lodManager.Stats()
	// config already declared above for point rendering

	var modeStr string
	if g.testMode {
		modeStr = "TEST MODE - 4 planets with textures"
	} else {
		modeStr = fmt.Sprintf("%d random objects", stats.TotalObjects)
	}

	// Build threshold info based on mode
	var thresholdStr string
	if config.UseApparentSize {
		thresholdStr = fmt.Sprintf(
			"LOD Thresholds (pixels):\n"+
				"  Full3D:    >= %.0f px\n"+
				"  Billboard: >= %.0f px\n"+
				"  Circle:    >= %.0f px\n"+
				"  Point:     >= %.1f px\n"+
				"  Hysteresis: %.0f%%\n"+
				"  Transition: %.1fs",
			config.Full3DPixels,
			config.BillboardPixels,
			config.CirclePixels,
			config.PointPixels,
			config.Hysteresis*100,
			config.TransitionTime,
		)
	} else {
		thresholdStr = fmt.Sprintf(
			"LOD Thresholds (distance):\n"+
				"  Full3D:    < %.0f\n"+
				"  Billboard: < %.0f\n"+
				"  Circle:    < %.0f\n"+
				"  Point:     < %.0f",
			config.Full3DDistance,
			config.BillboardDistance,
			config.CircleDistance,
			config.PointDistance,
		)
	}

	statsText := fmt.Sprintf(
		"LOD Demo - %s\n"+
			"FPS: %.1f\n"+
			"Camera: (%.0f, %.0f, %.0f)\n"+
			"\n"+
			"%s\n"+
			"\n"+
			"Tier Stats:\n"+
			"  Full3D:    %d\n"+
			"  Billboard: %d\n"+
			"  Circle:    %d\n"+
			"  Point:     %d\n"+
			"  Culled:    %d\n"+
			"  Visible:   %d\n"+
			"  Transitioning: %d\n"+
			"\n"+
			"Controls:\n"+
			"  WASD/Arrows: Move camera\n"+
			"  Q/E: Up/Down\n"+
			"  Shift: Fast move\n"+
			"  R: Reset position",
		modeStr,
		g.fps,
		g.camera.Pos.X, g.camera.Pos.Y, g.camera.Pos.Z,
		thresholdStr,
		stats.Full3DCount,
		stats.BillboardCount,
		stats.CircleCount,
		stats.PointCount,
		stats.CulledCount,
		stats.VisibleCount,
		len(transitioning),
	)
	ebitenutil.DebugPrint(screen, statsText)

	// Take screenshot if requested
	if g.screenshotFrame > 0 && g.frameCount == g.screenshotFrame {
		g.saveScreenshot(screen)
	}
}

// Layout returns the screen dimensions.
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

// saveScreenshot saves the current frame as a PNG and exits.
func (g *Game) saveScreenshot(screen *ebiten.Image) {
	if g.screenshotPath == "" {
		g.screenshotPath = "out/screenshots/lod-demo.png"
	}

	f, err := os.Create(g.screenshotPath)
	if err != nil {
		log.Printf("Failed to create screenshot file: %v", err)
		os.Exit(1)
	}
	defer f.Close()

	if err := png.Encode(f, screen); err != nil {
		log.Printf("Failed to encode screenshot: %v", err)
		os.Exit(1)
	}

	fmt.Printf("Screenshot saved to %s\n", g.screenshotPath)
	os.Exit(0)
}

func main() {
	objectCount := flag.Int("objects", 5000, "Number of objects to render")
	screenshotFrame := flag.Int("screenshot", 0, "Frame to take screenshot (0 = disabled)")
	screenshotPath := flag.String("output", "", "Screenshot output path")
	testMode := flag.Bool("test", false, "Test mode: 4 textured planets at fixed positions")
	flag.Parse()

	if *testMode {
		fmt.Println("LOD Demo: Test mode - 4 textured planets")
		fmt.Println("  Sun (0,0,0): Yellow star")
		fmt.Println("  Earth (60,0,20): Blue planet")
		fmt.Println("  Jupiter (-80,10,-30): Brown gas giant")
		fmt.Println("  Uranus (0,-20,-100): Cyan ice giant")
		fmt.Println("\nMove toward planets to see 3D textures (Full3D tier < 50 units)")
	} else {
		fmt.Printf("LOD Demo: Rendering %d objects\n", *objectCount)
	}

	game := NewGame(*objectCount, *screenshotFrame, *screenshotPath, *testMode)

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("LOD System Demo")
	ebiten.SetVsyncEnabled(true)

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
