// Package main provides a stress test demo for the LOD (Level of Detail) system.
// It renders celestial objects with automatic LOD tier switching including actual 3D planets.
// Now includes SR (Special Relativity) and GR (General Relativity) visual effects.
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
	"sort"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"stapledons_voyage/engine/lod"
	"stapledons_voyage/engine/shader"
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

	// SR/GR shader effects
	shaderManager *shader.Manager
	srWarp        *shader.SRWarp
	grWarp        *shader.GRWarp
	renderBuffer  *ebiten.Image // Off-screen buffer for shader post-processing

	// Ship velocity (fraction of c, 0.0 to 0.99)
	velocity float64

	// Light sources for runtime adjustment
	starLights      []*tetra.StarLight // Lights from luminous objects
	ambientLight    *tetra.AmbientLight
	lightMultiplier float64 // Star light intensity multiplier (0.1 to 3.0)
	ambientLevel    float64 // Ambient light level (0.0 to 1.0)

	// Loaded planet textures for random assignment
	planetTextures []*ebiten.Image

	// Disable 3D rendering for debugging
	no3D bool

	// Disable transitioning layer for debugging
	noTransitions bool
}

// NewGame creates a new LOD demo with the specified number of objects.
func NewGame(objectCount int, screenshotFrame int, screenshotPath string, testMode bool, maxLights int, no3D bool, noTransitions bool) *Game {
	// Create LOD manager
	config := lod.DefaultConfig()
	config.Max3DObjects = 20
	// Disable transitions to prevent flickering from alpha blending overlapping sprites
	// Transitions cause visual artifacts when multiple transitioning objects overlap
	config.TransitionTime = 0
	// Increase hysteresis to reduce "popping" frequency (objects need 40% size change to switch tiers)
	config.Hysteresis = 0.4
	manager := lod.NewManager(config)

	// Create camera
	camera := lod.NewSimpleCamera(screenWidth, screenHeight)
	camera.Fov = 60
	camera.Far = 20000

	// Create Tetra3D scene
	scene3D := tetra.NewScene(screenWidth, screenHeight)
	scene3D.SetFar(20000) // Match LOD camera far plane

	var planets []*Planet3D

	// Load planet textures for both modes
	textures, sunTex := loadPlanetTextures()

	if testMode {
		// Test mode: 5 planets at specific distances
		// Start with good view of Earth and Sun, can see Saturn and Jupiter
		camera.Pos = lod.Vector3{X: 30, Y: 30, Z: 100}
		camera.LookAt = lod.Vector3{X: 0, Y: 0, Z: 0}
		planets = createTestPlanets(manager, scene3D)
	} else {
		// Normal mode: random star field with textured planets and star lights
		// Start closer to see more Full3D objects
		camera.Pos = lod.Vector3{X: 0, Y: 0, Z: 100}
		camera.LookAt = lod.Vector3{X: 0, Y: 0, Z: 0}
		rng := rand.New(rand.NewSource(42))
		planets = generateObjects(manager, scene3D, rng, objectCount, textures, sunTex)
	}

	// Create billboard renderer with larger default sprite (128x128)
	billboardRenderer := lod.NewBillboardRenderer()
	defaultSprite := lod.CreateDefaultPlanetSprite(128, color.RGBA{255, 255, 255, 255})
	billboardRenderer.SetDefaultSprite(defaultSprite)

	// Enable scene lighting so planets receive light from sources
	scene3D.SetLightingEnabled(true)

	// Create lights dynamically from any object with Luminosity > 0
	// This makes light sources data-driven rather than hardcoded
	// maxLights limits how many point lights are created (0 = unlimited)
	// Multiple point lights can cause flickering due to Tetra3D lighting calculations
	var starLights []*tetra.StarLight
	lightCount := 0
	for _, p := range planets {
		if p.lodObj.IsLightSource() {
			// Check if we've hit the light limit
			if maxLights > 0 && lightCount >= maxLights {
				// Still make it shadeless (self-illuminated) even without a light
				p.planet.SetShadeless(true)
				continue
			}

			// Create a PointLight at the object's position with its luminosity
			// Use EffectiveLightColor for proper spectral light color
			lightCol := p.lodObj.EffectiveLightColor()
			r := float64(lightCol.R) / 255.0
			g := float64(lightCol.G) / 255.0
			b := float64(lightCol.B) / 255.0

			starLight := tetra.NewStarLight(
				p.lodObj.ID+"_light",
				r, g, b,
				p.lodObj.Luminosity,
				0, // infinite range
			)
			starLight.SetPosition(p.lodObj.Position.X, p.lodObj.Position.Y, p.lodObj.Position.Z)
			starLight.AddToScene(scene3D)
			starLights = append(starLights, starLight)
			lightCount++
			log.Printf("Created light for %s: luminosity=%.0f, spectral color=(%.2f,%.2f,%.2f)",
				p.lodObj.ID, p.lodObj.Luminosity, r, g, b)

			// Make light-emitting objects shadeless (self-illuminated)
			p.planet.SetShadeless(true)
			log.Printf("Made %s shadeless (self-illuminated)", p.lodObj.ID)
		}
	}
	if maxLights > 0 {
		log.Printf("Light limit: %d (created %d lights)", maxLights, lightCount)
	}

	// Low ambient so we see clear day/night contrast
	ambient := tetra.NewAmbientLight(0.08, 0.08, 0.1, 1.0)
	ambient.AddToScene(scene3D)

	// Build sprite map from planet billboards
	planetSprites := make(map[string]*ebiten.Image)
	for _, p := range planets {
		if p.billboard != nil {
			planetSprites[p.lodObj.ID] = p.billboard
		}
	}

	// Initialize shader system for SR/GR effects
	shaderMgr := shader.NewManager()
	srWarp := shader.NewSRWarp(shaderMgr)
	grWarp := shader.NewGRWarp(shaderMgr)

	// Pre-configure GR for demo mode (centered on screen)
	grWarp.SetDemoMode(0.5, 0.5, 0.08, 0.01)

	// Create off-screen render buffer for shader post-processing
	renderBuffer := ebiten.NewImage(screenWidth, screenHeight)

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
		shaderManager:     shaderMgr,
		srWarp:            srWarp,
		grWarp:            grWarp,
		renderBuffer:      renderBuffer,
		velocity:          0.0,
		starLights:        starLights,
		ambientLight:      ambient,
		lightMultiplier:   1.0,
		ambientLevel:      1.0,
		no3D:              no3D,
		noTransitions:     noTransitions,
	}
}

// createTestPlanets creates test planets with real textures at known positions
func createTestPlanets(manager *lod.Manager, scene3D *tetra.Scene) []*Planet3D {
	planets := make([]*Planet3D, 0, 5)

	// Planet definitions: name, position, radius, texture path, fallback color, luminosity, light color
	// Luminosity > 0 means the object emits light (e.g., stars).
	// Due to inverse square falloff: intensity = luminosity / distance².
	// At distance 60, luminosity 8000 gives intensity ~2.2, which is visible.
	// Light colors based on stellar spectral classification:
	//   G-type (Sun): Yellow-white (255, 243, 217) - ~5778K
	//
	// Layout (top-down view, +Z is "up" screen):
	//                   Neptune (0, -20, -120)
	//                        |
	//   Jupiter (-100, 10, -40)    Saturn (80, 5, -60)
	//                        |
	//                   Sun (0, 0, 0)
	//                        |
	//                   Earth (50, 0, 40)
	defs := []struct {
		name       string
		pos        lod.Vector3
		radius     float64
		texPath    string
		color      color.RGBA
		luminosity float64    // 0 = not a light source, >0 = emits light
		lightColor color.RGBA // spectral light color (0,0,0 = use object color)
	}{
		// Sun: G-type star, yellow-white light (center)
		{"Sun", lod.Vector3{X: 0, Y: 0, Z: 0}, 15.0, "assets/planets/sun.jpg", color.RGBA{255, 200, 50, 255}, 8000.0, color.RGBA{255, 243, 217, 255}},
		// Earth: Close, in front-right (good for initial view)
		{"Earth", lod.Vector3{X: 50, Y: 0, Z: 40}, 8.0, "assets/planets/earth.jpg", color.RGBA{50, 100, 200, 255}, 0, color.RGBA{}},
		// Jupiter: Large gas giant, left side
		{"Jupiter", lod.Vector3{X: -100, Y: 10, Z: -40}, 12.0, "assets/planets/jupiter.jpg", color.RGBA{200, 150, 100, 255}, 0, color.RGBA{}},
		// Saturn: Ringed gas giant, right side (between Earth and Neptune)
		{"Saturn", lod.Vector3{X: 80, Y: 5, Z: -60}, 10.0, "assets/planets/saturn.jpg", color.RGBA{220, 190, 140, 255}, 0, color.RGBA{}},
		// Neptune: Far ice giant, deep background
		{"Neptune", lod.Vector3{X: 0, Y: -20, Z: -120}, 6.0, "assets/planets/neptune.jpg", color.RGBA{50, 100, 200, 255}, 0, color.RGBA{}},
	}

	for _, def := range defs {
		// Create LOD object with luminosity and light color
		lodObj := lod.NewObject(def.name, def.pos, def.radius, def.color)
		lodObj.Luminosity = def.luminosity
		lodObj.LightColor = def.lightColor
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

// updateStarLightIntensities updates star light sources based on the light multiplier.
func (g *Game) updateStarLightIntensities() {
	// Update star lights (scale their base energy by the multiplier)
	for i, light := range g.starLights {
		if i < len(g.planets) {
			// Find the matching planet to get base luminosity
			for _, p := range g.planets {
				if p.lodObj.IsLightSource() {
					light.SetEnergy(p.lodObj.Luminosity * g.lightMultiplier)
					break
				}
			}
		}
	}
}

// updateAmbientLight updates ambient light energy based on ambientLevel.
func (g *Game) updateAmbientLight() {
	if g.ambientLight != nil {
		g.ambientLight.SetEnergy(g.ambientLevel)
	}
}

// filterOverlappingBillboards removes billboards that overlap with Full3D objects or each other.
// When billboards overlap, only the closest one is kept to prevent alpha blending artifacts.
// This is more aggressive but eliminates flickering from depth-sorting transparent sprites.
func (g *Game) filterOverlappingBillboards(billboards []*lod.Object, full3D []*lod.Object) []*lod.Object {
	if len(billboards) == 0 {
		return billboards
	}

	// First, sort billboards by distance (closest first) so we keep the closer one when overlapping
	// Use stable sort with bucket-based ordering to prevent flickering
	sorted := make([]*lod.Object, len(billboards))
	copy(sorted, billboards)
	bucketSize := lod.DistanceBucketSize()
	sort.SliceStable(sorted, func(i, j int) bool {
		bucketI := int(sorted[i].Distance / bucketSize)
		bucketJ := int(sorted[j].Distance / bucketSize)
		if bucketI != bucketJ {
			return bucketI < bucketJ // Closer first
		}
		return sorted[i].ID < sorted[j].ID // Stable tiebreaker
	})

	// Track which billboards to keep (not overlapping with anything closer)
	keep := make([]bool, len(sorted))
	for i := range keep {
		keep[i] = true
	}

	// Check each billboard against 3D objects
	for i, bb := range sorted {
		for _, obj3D := range full3D {
			dx := bb.ScreenX - obj3D.ScreenX
			dy := bb.ScreenY - obj3D.ScreenY
			dist := math.Sqrt(dx*dx + dy*dy)
			// Use 80% of combined radii for overlap (allows some edge overlap)
			minDist := (bb.ApparentRadius + obj3D.ApparentRadius) * 0.8

			if dist < minDist {
				keep[i] = false
				break
			}
		}
	}

	// Check each billboard against closer billboards (already processed)
	for i := 1; i < len(sorted); i++ {
		if !keep[i] {
			continue
		}
		bb := sorted[i]

		// Check against all closer billboards (indices 0 to i-1)
		for j := 0; j < i; j++ {
			if !keep[j] {
				continue
			}
			closer := sorted[j]

			dx := bb.ScreenX - closer.ScreenX
			dy := bb.ScreenY - closer.ScreenY
			dist := math.Sqrt(dx*dx + dy*dy)
			// Use 60% of combined radii - more aggressive overlap detection
			minDist := (bb.ApparentRadius + closer.ApparentRadius) * 0.6

			if dist < minDist {
				// This billboard overlaps with a closer one - skip it
				keep[i] = false
				break
			}
		}
	}

	// Build result with only non-overlapping billboards
	filtered := make([]*lod.Object, 0, len(sorted))
	for i, bb := range sorted {
		if keep[i] {
			filtered = append(filtered, bb)
		}
	}
	return filtered
}

// initializeBillboards creates billboard sprites from planet textures.
// Uses texture caching - billboards are shared by planets with the same texture.
// This prevents creating thousands of textures which would exhaust GPU memory.
func (g *Game) initializeBillboards() {
	// Cache billboards by texture pointer - only create one billboard per unique texture
	// With ~14 planet textures + sun, this creates at most 15 billboards instead of 5000
	textureBillboards := make(map[*ebiten.Image]*ebiten.Image)
	textureColors := make(map[*ebiten.Image]color.RGBA)

	// First pass: create one billboard per unique texture
	for _, p := range g.planets {
		if p.texture == nil {
			continue
		}

		// Check if we already processed this texture
		if _, exists := textureBillboards[p.texture]; exists {
			continue
		}

		// Extract average color from texture for circle/point rendering
		avgColor := lod.ExtractAverageColor(p.texture)
		textureColors[p.texture] = avgColor

		// Create billboard from the texture
		billboard := lod.CreateBillboardFromTexture(p.texture, 128)
		textureBillboards[p.texture] = billboard
	}

	log.Printf("Created %d cached billboard textures (for %d objects)", len(textureBillboards), len(g.planets))

	// Create one procedural billboard for objects without textures
	var proceduralBillboard *ebiten.Image

	// Second pass: assign shared billboards to all planets
	for _, p := range g.planets {
		if p.texture != nil {
			// Use cached billboard and color
			p.billboard = textureBillboards[p.texture]
			p.lodObj.Color = textureColors[p.texture]
		} else {
			// Use shared procedural billboard
			if proceduralBillboard == nil {
				proceduralBillboard = lod.CreateDefaultPlanetSprite(128, p.lodObj.Color)
			}
			p.billboard = proceduralBillboard
		}
		g.planetSprites[p.lodObj.ID] = p.billboard
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

// loadPlanetTextures loads all available planet textures from assets/planets/
// Returns (planetTextures, sunTexture)
func loadPlanetTextures() ([]*ebiten.Image, *ebiten.Image) {
	texturePaths := []string{
		"assets/planets/earth.jpg",
		"assets/planets/mars.jpg",
		"assets/planets/jupiter.jpg",
		"assets/planets/saturn.jpg",
		"assets/planets/neptune.jpg",
		"assets/planets/uranus.jpg",
		"assets/planets/venus.jpg",
		"assets/planets/mercury.jpg",
		"assets/planets/moon.jpg",
		"assets/planets/pluto_globe.jpg",
		"assets/planets/ceres.jpg",
		"assets/planets/eris.jpg",
		"assets/planets/makemake.jpg",
		"assets/planets/haumea.jpg",
	}

	var textures []*ebiten.Image
	for _, path := range texturePaths {
		tex := loadTexture(path)
		if tex != nil {
			textures = append(textures, tex)
		}
	}

	// Load sun texture separately for stars
	sunTex := loadTexture("assets/planets/sun.jpg")
	log.Printf("Loaded %d planet textures + sun texture for random assignment", len(textures))
	return textures, sunTex
}

// starType represents a stellar spectral classification with associated colors and luminosity
type starType struct {
	col        color.RGBA
	lightCol   color.RGBA
	luminosity float64
}

// placedObject tracks position and radius for collision detection
type placedObject struct {
	pos    lod.Vector3
	radius float64
}

// checkOverlap returns true if a new object at pos with radius would overlap any existing objects
func checkOverlap(placed []placedObject, pos lod.Vector3, radius float64, minGap float64) bool {
	for _, p := range placed {
		dx := pos.X - p.pos.X
		dy := pos.Y - p.pos.Y
		dz := pos.Z - p.pos.Z
		dist := math.Sqrt(dx*dx + dy*dy + dz*dz)
		minDist := radius + p.radius + minGap
		if dist < minDist {
			return true // overlaps
		}
	}
	return false
}

// generateObjects creates random celestial objects distributed in 3D space.
// Returns Planet3D objects with textures and LOD objects.
// All planets get random planet textures, all stars get the sun texture.
// Objects are placed with collision detection to prevent overlapping.
func generateObjects(manager *lod.Manager, scene3D *tetra.Scene, rng *rand.Rand, count int, textures []*ebiten.Image, sunTex *ebiten.Image) []*Planet3D {
	// Star colors based on spectral type
	starColors := []starType{
		// O-type: Blue-white, very hot (30,000-50,000K)
		{color.RGBA{155, 176, 255, 255}, color.RGBA{155, 176, 255, 255}, 5000.0},
		// B-type: Blue-white (10,000-30,000K)
		{color.RGBA{170, 191, 255, 255}, color.RGBA{170, 191, 255, 255}, 3000.0},
		// A-type: White (7,500-10,000K)
		{color.RGBA{202, 215, 255, 255}, color.RGBA{202, 215, 255, 255}, 2000.0},
		// F-type: Yellow-white (6,000-7,500K)
		{color.RGBA{248, 247, 255, 255}, color.RGBA{248, 247, 255, 255}, 1500.0},
		// G-type: Yellow (like Sun) (5,200-6,000K)
		{color.RGBA{255, 244, 234, 255}, color.RGBA{255, 244, 234, 255}, 1000.0},
		// K-type: Orange (3,700-5,200K)
		{color.RGBA{255, 210, 161, 255}, color.RGBA{255, 210, 161, 255}, 600.0},
		// M-type: Red (2,400-3,700K)
		{color.RGBA{255, 204, 111, 255}, color.RGBA{255, 180, 100, 255}, 300.0},
	}

	// Planet colors (non-luminous)
	planetColors := []color.RGBA{
		{200, 150, 100, 255}, // Rocky tan
		{150, 150, 180, 255}, // Rocky gray-blue
		{180, 120, 90, 255},  // Mars-like red
		{100, 130, 200, 255}, // Neptune-like blue
		{220, 190, 140, 255}, // Gas giant tan
		{150, 200, 200, 255}, // Ice giant cyan
		{180, 180, 180, 255}, // Gray rock
		{200, 180, 160, 255}, // Pale tan
	}

	planets := make([]*Planet3D, 0, count)
	placed := make([]placedObject, 0, count) // Track placed objects for collision detection
	const minGap = 2.0                        // Minimum gap between objects to prevent z-fighting

	for i := 0; i < count; i++ {
		var pos lod.Vector3
		var radius float64
		var col color.RGBA
		var luminosity float64
		var lightColor color.RGBA
		var isStar bool
		var textureIdx int = -1

		// Determine object type and radius first (needed for collision check)
		objType := rng.Float64()
		if objType < 0.02 {
			// 2% are large stars (like supergiants)
			radius = 8 + rng.Float64()*12
			st := starColors[rng.Intn(3)] // Prefer hot stars for large ones
			col = st.col
			lightColor = st.lightCol
			luminosity = st.luminosity * (2.0 + rng.Float64()*3.0) // 2-5x base
			isStar = true
		} else if objType < 0.06 {
			// 4% are medium stars (main sequence)
			radius = 3 + rng.Float64()*5
			st := starColors[rng.Intn(len(starColors))]
			col = st.col
			lightColor = st.lightCol
			luminosity = st.luminosity
			isStar = true
		} else if objType < 0.10 {
			// 4% are small stars (red dwarfs) - total 10% stars
			radius = 1 + rng.Float64()*2
			st := starColors[5+rng.Intn(2)] // K or M type
			col = st.col
			lightColor = st.lightCol
			luminosity = st.luminosity * 0.5
			isStar = true
		} else if objType < 0.30 {
			// 20% are large planets/gas giants - use textures
			radius = 4 + rng.Float64()*8
			col = planetColors[rng.Intn(len(planetColors))]
			if len(textures) > 0 {
				textureIdx = rng.Intn(len(textures))
			}
		} else if objType < 0.55 {
			// 25% are medium planets - use textures
			radius = 2 + rng.Float64()*3
			col = planetColors[rng.Intn(len(planetColors))]
			if len(textures) > 0 {
				textureIdx = rng.Intn(len(textures))
			}
		} else {
			// 45% are small bodies (asteroids, moons)
			radius = 0.3 + rng.Float64()*1.5
			col = color.RGBA{
				uint8(100 + rng.Intn(100)),
				uint8(100 + rng.Intn(100)),
				uint8(100 + rng.Intn(100)),
				255,
			}
		}

		// Try to find a non-overlapping position (max 10 attempts)
		var foundPosition bool
		for attempt := 0; attempt < 10; attempt++ {
			// Distance distribution: more close objects for better Full3D testing
			distFactor := rng.Float64()
			var distance float64
			if distFactor < 0.25 {
				// 25% very close (within Full3D range)
				distance = 20 + rng.Float64()*80
			} else if distFactor < 0.50 {
				// 25% close (billboard range)
				distance = 100 + rng.Float64()*200
			} else if distFactor < 0.80 {
				// 30% medium distance
				distance = 300 + rng.Float64()*1000
			} else {
				// 20% far
				distance = 1000 + rng.Float64()*5000
			}

			theta := rng.Float64() * 2 * math.Pi
			phi := rng.Float64()*math.Pi - math.Pi/2

			pos.X = distance * math.Cos(phi) * math.Cos(theta)
			pos.Y = distance * math.Sin(phi)
			pos.Z = distance * math.Cos(phi) * math.Sin(theta)

			// Check for overlap with existing objects
			if !checkOverlap(placed, pos, radius, minGap) {
				foundPosition = true
				break
			}
		}

		// If couldn't find non-overlapping position, skip this object
		if !foundPosition {
			continue
		}

		// Track this placement
		placed = append(placed, placedObject{pos: pos, radius: radius})

		// Create LOD object
		obj := lod.NewObject(fmt.Sprintf("obj_%d", i), pos, radius, col)
		if luminosity > 0 {
			obj.Luminosity = luminosity
			obj.LightColor = lightColor
		}
		manager.Add(obj)

		// Create 3D planet - ALL objects get textures
		var planet *tetra.Planet
		var tex *ebiten.Image
		if isStar && sunTex != nil {
			// Stars use sun texture
			tex = sunTex
			planet = tetra.NewTexturedPlanet(fmt.Sprintf("obj_%d", i), radius, tex)
		} else if len(textures) > 0 {
			// All planets (including small bodies) get random planet textures
			if textureIdx < 0 {
				textureIdx = rng.Intn(len(textures))
			}
			tex = textures[textureIdx]
			planet = tetra.NewTexturedPlanet(fmt.Sprintf("obj_%d", i), radius, tex)
		} else {
			// Fallback to solid color only if no textures available
			planet = tetra.NewPlanet(fmt.Sprintf("obj_%d", i), radius, col)
		}
		planet.AddToScene(scene3D)
		planet.SetPosition(pos.X, pos.Y, pos.Z)

		planets = append(planets, &Planet3D{
			lodObj:    obj,
			planet:    planet,
			texture:   tex,
			billboard: nil, // Created lazily
		})
	}

	log.Printf("Generated %d objects with textures", count)
	return planets
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
	if !g.billboardsInitialized {
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
			g.camera.Pos = lod.Vector3{X: 30, Y: 30, Z: 100}
		} else {
			g.camera.Pos = lod.Vector3{X: 0, Y: 0, Z: 500}
		}
	}

	// SR/GR effect controls
	// 1: Toggle SR warp effect
	if inpututil.IsKeyJustPressed(ebiten.Key1) {
		g.srWarp.Toggle()
		if g.srWarp.IsEnabled() {
			log.Printf("SR Warp ENABLED (velocity: %.1f%%c)", g.velocity*100)
		} else {
			log.Printf("SR Warp DISABLED")
		}
	}

	// 2: Toggle GR warp effect
	if inpututil.IsKeyJustPressed(ebiten.Key2) {
		g.grWarp.Toggle()
		if g.grWarp.IsEnabled() {
			log.Printf("GR Warp ENABLED (intensity: %s)", g.grWarp.GetDemoIntensity())
		} else {
			log.Printf("GR Warp DISABLED")
		}
	}

	// 3: Cycle GR intensity (Subtle → Strong → Extreme)
	if inpututil.IsKeyJustPressed(ebiten.Key3) && g.grWarp.IsEnabled() {
		intensity := g.grWarp.CycleDemoIntensity()
		log.Printf("GR intensity: %s", intensity)
	}

	// +/= : Increase velocity (accelerate toward c)
	if ebiten.IsKeyPressed(ebiten.KeyEqual) || ebiten.IsKeyPressed(ebiten.KeyKPAdd) {
		g.velocity += 0.005 // Increase by 0.5% c per frame
		if g.velocity > 0.99 {
			g.velocity = 0.99
		}
		g.srWarp.SetForwardVelocity(g.velocity)
	}

	// -: Decrease velocity (decelerate)
	if ebiten.IsKeyPressed(ebiten.KeyMinus) || ebiten.IsKeyPressed(ebiten.KeyKPSubtract) {
		g.velocity -= 0.005
		if g.velocity < 0 {
			g.velocity = 0
		}
		g.srWarp.SetForwardVelocity(g.velocity)
	}

	// 0: Reset velocity to zero
	if inpututil.IsKeyJustPressed(ebiten.Key0) {
		g.velocity = 0
		g.srWarp.SetForwardVelocity(0)
		log.Printf("Velocity reset to 0")
	}

	// [: Decrease star light intensity
	if ebiten.IsKeyPressed(ebiten.KeyLeftBracket) {
		g.lightMultiplier -= 0.02
		if g.lightMultiplier < 0.1 {
			g.lightMultiplier = 0.1
		}
		g.updateStarLightIntensities()
	}

	// ]: Increase star light intensity
	if ebiten.IsKeyPressed(ebiten.KeyRightBracket) {
		g.lightMultiplier += 0.02
		if g.lightMultiplier > 3.0 {
			g.lightMultiplier = 3.0
		}
		g.updateStarLightIntensities()
	}

	// L: Reset light intensity to default
	if inpututil.IsKeyJustPressed(ebiten.KeyL) {
		g.lightMultiplier = 1.0
		g.ambientLevel = 1.0
		g.updateStarLightIntensities()
		g.updateAmbientLight()
	}

	// ;: Decrease ambient light
	if ebiten.IsKeyPressed(ebiten.KeySemicolon) {
		g.ambientLevel -= 0.02
		if g.ambientLevel < 0.0 {
			g.ambientLevel = 0.0
		}
		g.updateAmbientLight()
	}

	// ': Increase ambient light
	if ebiten.IsKeyPressed(ebiten.KeyApostrophe) {
		g.ambientLevel += 0.02
		if g.ambientLevel > 2.0 {
			g.ambientLevel = 2.0
		}
		g.updateAmbientLight()
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

	// Update planet visibility based on LOD tier (all modes)
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

	g.frameCount++
	return nil
}

// Draw renders the game.
func (g *Game) Draw(screen *ebiten.Image) {
	// Determine render target: buffer if using shaders, screen if not
	renderTarget := screen
	useShaders := g.srWarp.IsEnabled() || g.grWarp.IsEnabled()
	if useShaders {
		renderTarget = g.renderBuffer
		g.renderBuffer.Clear()
	}

	// Clear to dark space color
	renderTarget.Fill(color.RGBA{5, 5, 15, 255})

	// Get objects by tier
	points := g.lodManager.GetTierPoint()
	circles := g.lodManager.GetTierCircle()
	billboards := g.lodManager.GetTierBillboard()
	full3D := g.lodManager.GetTier3D()
	transitioning := g.lodManager.GetTransitioning()

	// Layer 1: Render distant objects (points)
	// Use scaled points that grow as they approach circle threshold for smoother transition
	config := g.lodManager.Config()
	g.pointRenderer.RenderPointsScaled(renderTarget, points, config.CirclePixels)

	// Layer 2: Render medium-distance objects (circles)
	g.circleRenderer.RenderCircles(renderTarget, circles)

	// Layer 3: Render billboard tier (non-Full3D close objects)
	// NOTE: Overlap filtering removed - root cause of flickering was transitions, not overlaps
	if g.no3D {
		// --no-3d mode: render ALL objects as billboards (no 3D layer)
		allBillboards := append(billboards, full3D...)
		g.billboardRenderer.RenderBillboards(renderTarget, allBillboards, g.planetSprites)
	} else {
		g.billboardRenderer.RenderBillboards(renderTarget, billboards, g.planetSprites)

		// Layer 4: Render 3D scene (Full3D tier) - render actual 3D planets in all modes
		if len(full3D) > 0 {
			img3d := g.scene3D.Render()
			renderTarget.DrawImage(img3d, nil)
		}
	}

	// Layer 5: Render transitioning objects with blending
	// Objects transitioning between tiers are rendered with alpha to smoothly
	// fade between their old and new representations.
	// The slice from GetTransitioning is now pre-sorted back-to-front by the engine.
	if !g.noTransitions && len(transitioning) > 0 {
		for _, obj := range transitioning {
			// Skip objects already rendered in their target tier above
			// We only need to render the "fading out" previous tier representation

			prevAlpha := obj.PreviousAlpha() // 1.0 → 0.0 as transition progresses

			// Render previous tier with fading alpha
			switch obj.PreviousTier {
			case lod.TierPoint:
				g.pointRenderer.RenderPointWithAlpha(renderTarget, obj, prevAlpha)
			case lod.TierCircle:
				g.circleRenderer.RenderCircleWithAlpha(renderTarget, obj, prevAlpha)
			case lod.TierBillboard:
				g.billboardRenderer.RenderBillboardWithAlpha(renderTarget, obj, prevAlpha, g.planetSprites)
			case lod.TierFull3D:
				// For 3D→lower tier transitions, render a fading circle as fallback
				g.circleRenderer.RenderCircleWithAlpha(renderTarget, obj, prevAlpha)
			}
		}
	}

	// Apply shader post-processing effects
	if useShaders {
		// Chain shaders: renderBuffer → intermediate → screen
		// For now, apply SR first, then GR
		src := g.renderBuffer

		// Apply SR warp if enabled
		if g.srWarp.IsEnabled() && g.velocity >= 0.05 {
			// Create intermediate buffer for chaining
			intermediate := ebiten.NewImage(screenWidth, screenHeight)
			if g.srWarp.Apply(intermediate, src) {
				src = intermediate
			}
		}

		// Apply GR warp if enabled
		if g.grWarp.IsEnabled() {
			intermediate := ebiten.NewImage(screenWidth, screenHeight)
			if g.grWarp.Apply(intermediate, src) {
				src = intermediate
			}
		}

		// Draw final result to screen
		screen.DrawImage(src, nil)
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

	// SR/GR status
	srStatus := "OFF"
	grStatus := "OFF"
	if g.srWarp.IsEnabled() {
		srStatus = "ON"
	}
	if g.grWarp.IsEnabled() {
		grStatus = fmt.Sprintf("ON (%s)", g.grWarp.GetDemoIntensity())
	}

	// Calculate gamma (Lorentz factor)
	gamma := 1.0
	if g.velocity > 0 {
		gamma = 1.0 / math.Sqrt(1.0-g.velocity*g.velocity)
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
			"Lighting:\n"+
			"  Star Light: %.1fx\n"+
			"  Ambient:    %.1fx\n"+
			"  Sources:    %d\n"+
			"\n"+
			"Relativistic Effects:\n"+
			"  Velocity: %.1f%% c\n"+
			"  Gamma:    %.2f\n"+
			"  SR Warp:  %s\n"+
			"  GR Warp:  %s\n"+
			"\n"+
			"Controls:\n"+
			"  WASD/Arrows: Move | Q/E: Up/Down\n"+
			"  Shift: Fast | R: Reset position\n"+
			"  [ / ]: Star light | ; / ': Ambient\n"+
			"  L: Reset lights | 1/2: SR/GR warp\n"+
			"  3: Cycle GR | +/-/0: Velocity",
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
		g.lightMultiplier,
		g.ambientLevel,
		len(g.starLights),
		g.velocity*100,
		gamma,
		srStatus,
		grStatus,
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
	testMode := flag.Bool("test", false, "Test mode: 5 textured planets at fixed positions")
	maxLights := flag.Int("max-lights", 0, "Maximum number of point lights (0 = unlimited)")
	no3D := flag.Bool("no-3d", false, "Disable 3D rendering (billboard only) to isolate flickering cause")
	noTransitions := flag.Bool("no-transitions", false, "Disable transitioning layer (Layer 5) to isolate flickering cause")
	flag.Parse()

	if *testMode {
		fmt.Println("LOD Demo: Test mode - 5 textured planets")
		fmt.Println("  Sun (0,0,0): Yellow star (light source)")
		fmt.Println("  Earth (50,0,40): Blue planet")
		fmt.Println("  Jupiter (-100,10,-40): Brown gas giant")
		fmt.Println("  Saturn (80,5,-60): Ringed gas giant")
		fmt.Println("  Neptune (0,-20,-120): Blue ice giant")
		fmt.Println("\nMove toward planets to see 3D textures (Full3D tier < 50 units)")
	} else {
		fmt.Printf("LOD Demo: Rendering %d objects\n", *objectCount)
	}

	game := NewGame(*objectCount, *screenshotFrame, *screenshotPath, *testMode, *maxLights, *no3D, *noTransitions)

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("LOD System Demo")
	ebiten.SetVsyncEnabled(true)

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
