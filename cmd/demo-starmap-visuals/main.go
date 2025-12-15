// Package main provides a 3D starmap visualization demo using real star data.
// It renders 3,802 real stars from the CNS5 catalog with proper spectral colors,
// 3D textured spheres, and dynamic lighting to test performance limits.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	"log"
	"math"
	"os"
	"sort"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"stapledons_voyage/engine/lod"
	"stapledons_voyage/engine/tetra"
)

const (
	screenWidth  = 1280
	screenHeight = 720
)

// StarData represents a star from the JSON catalog
type StarData struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	X        float64 `json:"x"`        // light-years toward galactic center
	Y        float64 `json:"y"`        // light-years direction of rotation
	Z        float64 `json:"z"`        // light-years north galactic pole
	DistLY   float64 `json:"dist_ly"`  // light-years from Sol
	VMag     float64 `json:"vmag"`     // visual magnitude
	Spectral string  `json:"spectral"` // spectral type (O,B,A,F,G,K,M)
}

// StarCatalog represents the JSON file structure
type StarCatalog struct {
	Version string     `json:"version"`
	Source  string     `json:"source"`
	Count   int        `json:"count"`
	Stars   []StarData `json:"stars"`
}

// Star3D represents a star that can be rendered in 3D
type Star3D struct {
	lodObj  *lod.Object
	planet  *tetra.Planet
	texture *ebiten.Image
	data    *StarData
}

// SpectralInfo contains visual properties for a spectral type
type SpectralInfo struct {
	Color      color.RGBA // Surface color
	LightColor color.RGBA // Light emission color
	Luminosity float64    // Relative luminosity (Sun = 1.0)
	Radius     float64    // Relative radius (Sun = 1.0)
	Temp       int        // Surface temperature (K)
}

// Spectral type visual properties
// Colors are more saturated for better visual distinction when applied as texture modulation
var spectralTypes = map[string]SpectralInfo{
	// Hot blue stars - intense blue tint
	"O": {color.RGBA{120, 160, 255, 255}, color.RGBA{155, 176, 255, 255}, 30000.0, 8.0, 33000},
	"B": {color.RGBA{140, 180, 255, 255}, color.RGBA{170, 191, 255, 255}, 1000.0, 4.0, 15000},
	// White/blue-white stars
	"A": {color.RGBA{200, 220, 255, 255}, color.RGBA{202, 215, 255, 255}, 20.0, 2.0, 8500},
	// Yellow-white stars
	"F": {color.RGBA{255, 250, 240, 255}, color.RGBA{248, 247, 255, 255}, 3.0, 1.4, 6500},
	// Yellow stars (like Sun) - neutral, let texture show through
	"G": {color.RGBA{255, 250, 230, 255}, color.RGBA{255, 244, 234, 255}, 1.0, 1.0, 5800},
	// Orange stars - warm tint
	"K": {color.RGBA{255, 200, 140, 255}, color.RGBA{255, 210, 161, 255}, 0.4, 0.8, 4500},
	// Red dwarf stars - deep red/orange
	"M": {color.RGBA{255, 160, 80, 255}, color.RGBA{255, 180, 100, 255}, 0.04, 0.3, 3200},
}

// Game implements ebiten.Game interface for the starmap demo.
type Game struct {
	lodManager        *lod.Manager
	camera            *lod.SimpleCamera
	pointRenderer     *lod.PointRenderer
	circleRenderer    *lod.CircleRenderer
	billboardRenderer *lod.BillboardRenderer

	// Tetra3D scene for Full3D rendering
	scene3D *tetra.Scene
	stars   []*Star3D

	// Sun texture for all stars
	sunTexture *ebiten.Image

	// Default sprite for billboards
	defaultSprite *ebiten.Image
	// Star-specific billboard sprites (keyed by LOD object ID)
	starSprites map[string]*ebiten.Image

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

	// Light sources for runtime adjustment
	starLights      []*tetra.StarLight
	ambientLight    *tetra.AmbientLight
	lightMultiplier float64
	ambientLevel    float64

	// Star count for display
	totalStars   int
	visibleStars int

	// Scale factor (light-years to world units)
	scale float64

	// Camera target (for navigation)
	targetStar *Star3D

	// Grid visibility toggle
	showGrid bool
}

// loadStarCatalog loads stars from the JSON file
func loadStarCatalog(path string) (*StarCatalog, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read star catalog: %w", err)
	}

	var catalog StarCatalog
	if err := json.Unmarshal(data, &catalog); err != nil {
		return nil, fmt.Errorf("failed to parse star catalog: %w", err)
	}

	return &catalog, nil
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

// getSpectralInfo returns visual properties for a spectral type
func getSpectralInfo(spectral string) SpectralInfo {
	if info, ok := spectralTypes[spectral]; ok {
		return info
	}
	// Default to K-type if unknown
	return spectralTypes["K"]
}

// magnitudeToLuminosity converts visual magnitude to relative luminosity
// Brighter stars have lower (even negative) magnitudes
// Sun is ~-26.7, Sirius is -1.46, faintest visible is ~6
func magnitudeToLuminosity(vmag float64) float64 {
	// Luminosity ratio = 10^((M_ref - M) / 2.5)
	// Using Sun's absolute magnitude (4.83) as reference
	absM := vmag // This is apparent, not absolute, but gives relative brightness
	return math.Pow(10, (4.83-absM)/2.5)
}

// NewGame creates a new starmap demo
func NewGame(screenshotFrame int, screenshotPath string, maxStars int, maxLights int, scale float64) *Game {
	// Load star catalog
	catalog, err := loadStarCatalog("assets/data/starmap/stars.json")
	if err != nil {
		log.Fatalf("Failed to load star catalog: %v", err)
	}
	log.Printf("Loaded %d stars from %s catalog", catalog.Count, catalog.Source)

	// Limit stars if requested
	stars := catalog.Stars
	if maxStars > 0 && len(stars) > maxStars {
		stars = stars[:maxStars]
		log.Printf("Limited to %d stars", maxStars)
	}

	// Create LOD manager - progressive detail as you approach stars
	// Immediate neighbors get full 3D, then billboards, then quickly to points
	config := lod.DefaultConfig()
	config.Max3DObjects = 30     // Limit 3D for performance
	config.TransitionTime = 0
	config.Hysteresis = 0.25
	// Progressive thresholds:
	// - 3D for immediate neighbors (within ~1-2 ly)
	// - Billboard for nearby/medium stars (kick in early for nice visuals)
	// - Circle briefly before points
	// - Points for distant stars
	config.Full3DPixels = 20     // 3D when star subtends 20+ pixels
	config.BillboardPixels = 3   // Billboard at 3-20 pixels (kicks in much sooner)
	config.CirclePixels = 1      // Circle at 1-3 pixels (brief transition)
	config.PointPixels = 0.2     // Point for smaller
	manager := lod.NewManager(config)

	// Create camera - start at Sol (origin)
	camera := lod.NewSimpleCamera(screenWidth, screenHeight)
	camera.Fov = 60
	camera.Far = 10000 // Far enough to see all stars
	camera.Pos = lod.Vector3{X: 0, Y: 0, Z: 5 * scale} // Start 5 ly away

	// Create Tetra3D scene
	scene3D := tetra.NewScene(screenWidth, screenHeight)
	scene3D.SetFar(10000)

	// Load sun texture - prefer 8K for quality, fallback to 2K, then legacy
	sunTex := loadTexture("assets/stars/sun_8k.jpg")
	if sunTex == nil {
		sunTex = loadTexture("assets/stars/sun_2k.jpg")
	}
	if sunTex == nil {
		sunTex = loadTexture("assets/planets/sun.jpg") // Legacy fallback
	}
	if sunTex == nil {
		log.Println("Warning: Could not load sun texture, using procedural textures")
	} else {
		log.Println("Loaded star texture for spectral rendering")
	}

	// Create stars - add Sol first (not in catalog since catalog is "from Sol")
	star3Ds := make([]*Star3D, 0, len(stars)+1)

	// Add Sol at origin
	solData := StarData{
		ID:       "Sol",
		Name:     "Sol",
		X:        0,
		Y:        0,
		Z:        0,
		DistLY:   0,
		VMag:     -26.74, // Sun's apparent magnitude from Earth
		Spectral: "G",
	}
	stars = append([]StarData{solData}, stars...)

	for i := range stars {
		sd := &stars[i]

		// Get spectral info
		info := getSpectralInfo(sd.Spectral)

		// Position scaled to world units
		pos := lod.Vector3{
			X: sd.X * scale,
			Y: sd.Y * scale,
			Z: sd.Z * scale,
		}

		// Calculate visual radius - exaggerated for artistic effect
		// Real stars would be sub-pixel, but we want nice 3D spheres when close
		// Base radius: 0.2 ly - gives good 3D detail for immediate neighbors
		baseRadius := 0.2 * scale

		// Scale by spectral type (larger stars = bigger radius)
		visualRadius := baseRadius * info.Radius // O=8x, B=4x, G=1x, M=0.3x

		// Also boost by apparent brightness (brighter = more prominent)
		brightness := magnitudeToLuminosity(sd.VMag)
		if brightness > 1000 {
			visualRadius *= 1.8 // Supergiants (Sirius, etc)
		} else if brightness > 100 {
			visualRadius *= 1.4 // Bright stars
		} else if brightness > 10 {
			visualRadius *= 1.1 // Sun-like
		}
		// Dim stars (M dwarfs) stay at spectral radius

		// Create LOD object
		lodObj := lod.NewObject(sd.ID, pos, visualRadius, info.Color)

		// Set luminosity for light-emitting stars
		// Only bright stars emit significant light
		if sd.VMag < 6 { // Visible to naked eye
			lodObj.Luminosity = brightness * 100 // Scale for visibility
			lodObj.LightColor = info.LightColor
		}

		manager.Add(lodObj)

		// Create 3D planet for the star
		var planet *tetra.Planet
		if sunTex != nil {
			planet = tetra.NewTexturedPlanet(sd.ID, visualRadius, sunTex)
		} else {
			planet = tetra.NewPlanet(sd.ID, visualRadius, info.Color)
		}
		planet.AddToScene(scene3D)
		planet.SetPosition(pos.X, pos.Y, pos.Z)

		// Color modulation to show spectral type even with sun texture
		planet.SetColorModulation(
			float64(info.Color.R)/255.0,
			float64(info.Color.G)/255.0,
			float64(info.Color.B)/255.0,
		)

		star3Ds = append(star3Ds, &Star3D{
			lodObj:  lodObj,
			planet:  planet,
			texture: sunTex,
			data:    sd,
		})
	}

	// Create billboard renderer
	billboardRenderer := lod.NewBillboardRenderer()
	defaultSprite := lod.CreateDefaultPlanetSprite(64, color.RGBA{255, 255, 255, 255})
	billboardRenderer.SetDefaultSprite(defaultSprite)

	// Enable scene lighting
	scene3D.SetLightingEnabled(true)

	// Create lights from brightest stars
	var starLights []*tetra.StarLight
	lightCount := 0

	// Sort stars by brightness (lowest vmag = brightest)
	sortedStars := make([]*Star3D, len(star3Ds))
	copy(sortedStars, star3Ds)
	sort.Slice(sortedStars, func(i, j int) bool {
		return sortedStars[i].data.VMag < sortedStars[j].data.VMag
	})

	for _, s := range sortedStars {
		if maxLights > 0 && lightCount >= maxLights {
			break
		}
		if !s.lodObj.IsLightSource() {
			continue
		}

		info := getSpectralInfo(s.data.Spectral)
		r := float64(info.LightColor.R) / 255.0
		g := float64(info.LightColor.G) / 255.0
		b := float64(info.LightColor.B) / 255.0

		starLight := tetra.NewStarLight(
			s.lodObj.ID+"_light",
			r, g, b,
			s.lodObj.Luminosity,
			0, // infinite range
		)
		starLight.SetPosition(s.lodObj.Position.X, s.lodObj.Position.Y, s.lodObj.Position.Z)
		starLight.AddToScene(scene3D)
		starLights = append(starLights, starLight)
		lightCount++

		// Make light sources shadeless
		s.planet.SetShadeless(true)

		log.Printf("Created light for %s (vmag %.2f, spectral %s)",
			s.data.Name, s.data.VMag, s.data.Spectral)
	}
	log.Printf("Created %d light sources", lightCount)

	// Ambient light for deep space
	ambient := tetra.NewAmbientLight(0.05, 0.05, 0.08, 1.0)
	ambient.AddToScene(scene3D)

	return &Game{
		lodManager:        manager,
		camera:            camera,
		pointRenderer:     lod.NewPointRenderer(),
		circleRenderer:    lod.NewCircleRenderer(),
		billboardRenderer: billboardRenderer,
		scene3D:           scene3D,
		stars:             star3Ds,
		sunTexture:        sunTex,
		defaultSprite:     defaultSprite,
		starSprites:       make(map[string]*ebiten.Image),
		cameraSpeed:       5.0 * scale, // Scale camera speed with world
		screenshotFrame:   screenshotFrame,
		screenshotPath:    screenshotPath,
		lastUpdate:        time.Now(),
		starLights:        starLights,
		ambientLight:      ambient,
		lightMultiplier:   1.0,
		ambientLevel:      1.0,
		totalStars:        len(star3Ds),
		scale:             scale,
		showGrid:          true, // Grid visible by default
	}
}

// initializeBillboards creates circular billboard sprites using the engine's LOD system
func (g *Game) initializeBillboards() {
	// Use the engine's built-in circular sprite generator
	// Sun texture is for 3D spheres; billboards use procedural circles
	for _, s := range g.stars {
		info := getSpectralInfo(s.data.Spectral)
		sprite := lod.CreateDefaultPlanetSprite(64, info.Color)
		g.starSprites[s.lodObj.ID] = sprite
	}
	log.Printf("Created %d billboard sprites", len(g.starSprites))
}

// findNearestStar finds the star closest to the camera
func (g *Game) findNearestStar() *Star3D {
	var nearest *Star3D
	minDist := math.MaxFloat64

	for _, s := range g.stars {
		dx := s.lodObj.Position.X - g.camera.Pos.X
		dy := s.lodObj.Position.Y - g.camera.Pos.Y
		dz := s.lodObj.Position.Z - g.camera.Pos.Z
		dist := math.Sqrt(dx*dx + dy*dy + dz*dz)
		if dist < minDist {
			minDist = dist
			nearest = s
		}
	}
	return nearest
}

// Update handles input and updates game state.
func (g *Game) Update() error {
	now := time.Now()
	dt := now.Sub(g.lastUpdate).Seconds()
	g.lastUpdate = now

	if dt > 0 {
		g.fps = g.fps*0.95 + (1/dt)*0.05
	}

	// Lazy initialize billboards
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
		g.cameraSpeed = 20.0 * g.scale // Fast mode
	} else {
		g.cameraSpeed = 5.0 * g.scale
	}

	// Reset to Sol (origin)
	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		g.camera.Pos = lod.Vector3{X: 0, Y: 0, Z: 5 * g.scale}
		g.camera.LookAt = lod.Vector3{X: 0, Y: 0, Z: 0}
		log.Printf("Reset camera to Sol")
	}

	// Find nearest star (for info display)
	if inpututil.IsKeyJustPressed(ebiten.KeyN) {
		g.targetStar = g.findNearestStar()
		if g.targetStar != nil {
			log.Printf("Nearest star: %s (%s) at %.2f ly",
				g.targetStar.data.Name, g.targetStar.data.Spectral, g.targetStar.data.DistLY)
		}
	}

	// Toggle grid visibility
	if inpututil.IsKeyJustPressed(ebiten.KeyG) {
		g.showGrid = !g.showGrid
		if g.showGrid {
			log.Println("Grid enabled")
		} else {
			log.Println("Grid disabled")
		}
	}

	// Lighting controls
	if ebiten.IsKeyPressed(ebiten.KeyLeftBracket) {
		g.lightMultiplier -= 0.02
		if g.lightMultiplier < 0.1 {
			g.lightMultiplier = 0.1
		}
		g.updateStarLightIntensities()
	}
	if ebiten.IsKeyPressed(ebiten.KeyRightBracket) {
		g.lightMultiplier += 0.02
		if g.lightMultiplier > 5.0 {
			g.lightMultiplier = 5.0
		}
		g.updateStarLightIntensities()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyL) {
		g.lightMultiplier = 1.0
		g.ambientLevel = 1.0
		g.updateStarLightIntensities()
		g.updateAmbientLight()
	}
	if ebiten.IsKeyPressed(ebiten.KeySemicolon) {
		g.ambientLevel -= 0.02
		if g.ambientLevel < 0.0 {
			g.ambientLevel = 0.0
		}
		g.updateAmbientLight()
	}
	if ebiten.IsKeyPressed(ebiten.KeyApostrophe) {
		g.ambientLevel += 0.02
		if g.ambientLevel > 2.0 {
			g.ambientLevel = 2.0
		}
		g.updateAmbientLight()
	}

	// Update LOD manager
	g.lodManager.UpdateWithDT(g.camera, dt)

	// Update 3D scene camera
	if g.scene3D != nil {
		g.scene3D.SetCameraPosition(g.camera.Pos.X, g.camera.Pos.Y, g.camera.Pos.Z)
		g.scene3D.LookAt(g.camera.LookAt.X, g.camera.LookAt.Y, g.camera.LookAt.Z)
	}

	// Update star visibility based on LOD tier
	g.visibleStars = 0
	for _, s := range g.stars {
		if s.lodObj.CurrentTier == lod.TierFull3D {
			s.planet.Model().SetVisible(true, true)
		} else {
			s.planet.Model().SetVisible(false, true)
		}
		if s.lodObj.CurrentTier != lod.TierCulled {
			g.visibleStars++
		}
		// Rotate stars slowly
		s.planet.Update(dt * 0.1) // Slow rotation
	}

	g.frameCount++
	return nil
}

func (g *Game) updateStarLightIntensities() {
	for i, light := range g.starLights {
		if i < len(g.stars) {
			for _, s := range g.stars {
				if s.lodObj.IsLightSource() && s.lodObj.ID+"_light" == light.Name() {
					light.SetEnergy(s.lodObj.Luminosity * g.lightMultiplier)
					break
				}
			}
		}
	}
}

func (g *Game) updateAmbientLight() {
	if g.ambientLight != nil {
		g.ambientLight.SetEnergy(g.ambientLevel)
	}
}

// drawGrid renders orientation grid around Sol (origin)
// Includes axis lines (X=red, Y=green, Z=blue) and distance rings
func (g *Game) drawGrid(screen *ebiten.Image) {
	// Grid colors
	xAxisColor := color.RGBA{255, 80, 80, 180}   // Red for X (galactic center)
	yAxisColor := color.RGBA{80, 255, 80, 180}   // Green for Y (rotation)
	zAxisColor := color.RGBA{80, 80, 255, 180}   // Blue for Z (north pole)
	ringColor := color.RGBA{100, 100, 120, 100}  // Dim gray for distance rings

	// Grid extent in light-years
	gridExtent := 30.0 * g.scale

	// Draw axis lines through Sol (origin)
	// Each axis is a line from -extent to +extent

	// X axis (red) - toward galactic center
	g.drawLine3D(screen,
		lod.Vector3{X: -gridExtent, Y: 0, Z: 0},
		lod.Vector3{X: gridExtent, Y: 0, Z: 0},
		xAxisColor, 2)

	// Y axis (green) - direction of galactic rotation
	g.drawLine3D(screen,
		lod.Vector3{X: 0, Y: -gridExtent, Z: 0},
		lod.Vector3{X: 0, Y: gridExtent, Z: 0},
		yAxisColor, 2)

	// Z axis (blue) - north galactic pole
	g.drawLine3D(screen,
		lod.Vector3{X: 0, Y: 0, Z: -gridExtent},
		lod.Vector3{X: 0, Y: 0, Z: gridExtent},
		zAxisColor, 2)

	// Draw distance rings on XY plane (galactic plane)
	// Rings at 5, 10, 15, 20, 25, 30 light-years
	distances := []float64{5, 10, 15, 20, 25, 30}
	for _, dist := range distances {
		g.drawRing3D(screen, lod.Vector3{X: 0, Y: 0, Z: 0}, dist*g.scale, ringColor, 1)
	}

	// Draw Sol marker (small cross at origin)
	solColor := color.RGBA{255, 244, 234, 255} // Sun color
	markerSize := 0.3 * g.scale
	g.drawLine3D(screen,
		lod.Vector3{X: -markerSize, Y: 0, Z: 0},
		lod.Vector3{X: markerSize, Y: 0, Z: 0},
		solColor, 3)
	g.drawLine3D(screen,
		lod.Vector3{X: 0, Y: -markerSize, Z: 0},
		lod.Vector3{X: 0, Y: markerSize, Z: 0},
		solColor, 3)
}

// drawLine3D draws a line between two 3D points
func (g *Game) drawLine3D(screen *ebiten.Image, start, end lod.Vector3, c color.RGBA, width float32) {
	// Clip line to front of camera before projecting
	clippedStart, clippedEnd, valid := g.clipLineToFrustum(start, end)
	if !valid {
		return
	}

	// Project clipped points to screen space
	x1, y1, visible1 := g.camera.WorldToScreen(clippedStart)
	x2, y2, visible2 := g.camera.WorldToScreen(clippedEnd)

	// Both should be visible after clipping, but check anyway
	if !visible1 || !visible2 {
		return
	}

	// Skip degenerate lines (both points at same screen location)
	dx := x2 - x1
	dy := y2 - y1
	if dx*dx+dy*dy < 1 {
		return
	}

	vector.StrokeLine(screen, float32(x1), float32(y1), float32(x2), float32(y2), width, c, true)
}

// clipLineToFrustum clips a line segment to the front of the camera
// Returns the clipped endpoints and whether any part is visible
func (g *Game) clipLineToFrustum(start, end lod.Vector3) (lod.Vector3, lod.Vector3, bool) {
	// Calculate forward direction from camera
	forward := lod.Vector3{
		X: g.camera.LookAt.X - g.camera.Pos.X,
		Y: g.camera.LookAt.Y - g.camera.Pos.Y,
		Z: g.camera.LookAt.Z - g.camera.Pos.Z,
	}
	forwardLen := math.Sqrt(forward.X*forward.X + forward.Y*forward.Y + forward.Z*forward.Z)
	if forwardLen > 0 {
		forward.X /= forwardLen
		forward.Y /= forwardLen
		forward.Z /= forwardLen
	}

	// Near plane distance
	nearDist := g.camera.Near
	if nearDist < 0.5 {
		nearDist = 0.5 // Minimum clip distance to avoid projection issues
	}

	// Calculate signed distance from camera along forward direction
	d1 := (start.X-g.camera.Pos.X)*forward.X +
		(start.Y-g.camera.Pos.Y)*forward.Y +
		(start.Z-g.camera.Pos.Z)*forward.Z

	d2 := (end.X-g.camera.Pos.X)*forward.X +
		(end.Y-g.camera.Pos.Y)*forward.Y +
		(end.Z-g.camera.Pos.Z)*forward.Z

	// Both behind camera
	if d1 < nearDist && d2 < nearDist {
		return start, end, false
	}

	// Both in front of camera
	if d1 >= nearDist && d2 >= nearDist {
		return start, end, true
	}

	// One point behind, one in front - clip to near plane
	t := (nearDist - d1) / (d2 - d1)

	clipPoint := lod.Vector3{
		X: start.X + t*(end.X-start.X),
		Y: start.Y + t*(end.Y-start.Y),
		Z: start.Z + t*(end.Z-start.Z),
	}

	if d1 < nearDist {
		return clipPoint, end, true
	}
	return start, clipPoint, true
}

// drawRing3D draws a circle on the XY plane at the given center and radius
func (g *Game) drawRing3D(screen *ebiten.Image, center lod.Vector3, radius float64, c color.RGBA, width float32) {
	segments := 64

	for i := 0; i < segments; i++ {
		angle1 := float64(i) * 2 * math.Pi / float64(segments)
		angle2 := float64(i+1) * 2 * math.Pi / float64(segments)

		p1 := lod.Vector3{
			X: center.X + radius*math.Cos(angle1),
			Y: center.Y + radius*math.Sin(angle1),
			Z: center.Z,
		}
		p2 := lod.Vector3{
			X: center.X + radius*math.Cos(angle2),
			Y: center.Y + radius*math.Sin(angle2),
			Z: center.Z,
		}

		g.drawLine3D(screen, p1, p2, c, width)
	}
}

// Draw renders the game.
func (g *Game) Draw(screen *ebiten.Image) {
	// Clear to dark space color
	screen.Fill(color.RGBA{2, 2, 8, 255})

	// Draw orientation grid first (behind stars)
	if g.showGrid {
		g.drawGrid(screen)
	}

	// Get objects by tier
	points := g.lodManager.GetTierPoint()
	circles := g.lodManager.GetTierCircle()
	billboards := g.lodManager.GetTierBillboard()
	full3D := g.lodManager.GetTier3D()

	// Layer 1: Render distant objects (points)
	config := g.lodManager.Config()
	g.pointRenderer.RenderPointsScaled(screen, points, config.CirclePixels)

	// Layer 2: Render medium-distance objects (circles)
	g.circleRenderer.RenderCircles(screen, circles)

	// Layer 3: Render billboard tier
	g.billboardRenderer.RenderBillboards(screen, billboards, g.starSprites)

	// Layer 4: Render 3D scene (Full3D tier)
	if len(full3D) > 0 {
		img3d := g.scene3D.Render()
		screen.DrawImage(img3d, nil)
	}

	// Draw stats overlay
	stats := g.lodManager.Stats()

	// Camera position in light-years
	camLY := lod.Vector3{
		X: g.camera.Pos.X / g.scale,
		Y: g.camera.Pos.Y / g.scale,
		Z: g.camera.Pos.Z / g.scale,
	}

	// Nearest star info
	nearestInfo := "Press N to find nearest star"
	if g.targetStar != nil {
		nearestInfo = fmt.Sprintf("Nearest: %s (%s, %.1f vmag, %.2f ly)",
			g.targetStar.data.Name, g.targetStar.data.Spectral,
			g.targetStar.data.VMag, g.targetStar.data.DistLY)
	}

	statsText := fmt.Sprintf(
		"3D Starmap Demo - %d real stars (CNS5 catalog)\n"+
			"FPS: %.1f\n"+
			"Camera: (%.2f, %.2f, %.2f) ly\n"+
			"\n"+
			"LOD Stats:\n"+
			"  Full3D:    %d\n"+
			"  Billboard: %d\n"+
			"  Circle:    %d\n"+
			"  Point:     %d\n"+
			"  Culled:    %d\n"+
			"  Visible:   %d / %d\n"+
			"\n"+
			"Lighting:\n"+
			"  Star Light: %.1fx\n"+
			"  Ambient:    %.1fx\n"+
			"  Sources:    %d\n"+
			"\n"+
			"%s\n"+
			"\n"+
			"Controls:\n"+
			"  WASD/Arrows: Move | Q/E: Up/Down\n"+
			"  Shift: Fast | R: Reset to Sol\n"+
			"  N: Find nearest | G: Toggle grid\n"+
			"  [ / ]: Star light | ; / ': Ambient | L: Reset",
		g.totalStars,
		g.fps,
		camLY.X, camLY.Y, camLY.Z,
		stats.Full3DCount,
		stats.BillboardCount,
		stats.CircleCount,
		stats.PointCount,
		stats.CulledCount,
		g.visibleStars, g.totalStars,
		g.lightMultiplier,
		g.ambientLevel,
		len(g.starLights),
		nearestInfo,
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

func (g *Game) saveScreenshot(screen *ebiten.Image) {
	if g.screenshotPath == "" {
		g.screenshotPath = "out/screenshots/starmap-visuals.png"
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
	screenshotFrame := flag.Int("screenshot", 0, "Frame to take screenshot (0 = disabled)")
	screenshotPath := flag.String("output", "", "Screenshot output path")
	maxStars := flag.Int("max-stars", 0, "Maximum number of stars to render (0 = all)")
	maxLights := flag.Int("max-lights", 10, "Maximum number of point lights")
	scale := flag.Float64("scale", 1.0, "Scale factor for star positions (1.0 = 1 unit per light-year)")
	flag.Parse()

	fmt.Println("3D Starmap Demo")
	fmt.Println("  Real star data from CNS5 catalog")
	fmt.Println("  3,802 stars within ~35 light-years of Sol")
	fmt.Printf("  Scale: %.1f units per light-year\n", *scale)
	if *maxStars > 0 {
		fmt.Printf("  Limited to %d stars\n", *maxStars)
	}
	if *maxLights > 0 {
		fmt.Printf("  Light sources: %d max\n", *maxLights)
	}

	game := NewGame(*screenshotFrame, *screenshotPath, *maxStars, *maxLights, *scale)

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("3D Starmap - Real Star Data")
	ebiten.SetVsyncEnabled(true)

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
