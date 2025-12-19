// Package main provides a demo of the bubble ship dome observation system.
// This demonstrates a hemispherical dome with a starfield texture, showing
// what observation decks and transparent bubble ships will look like.
//
// Now includes real star catalog and LOD (Level of Detail) system for
// nearby celestial objects that appear as 3D spheres when close.
//
// The dome can face different directions, making it suitable for:
//   - Ceiling domes (up) - classic planetarium view
//   - Floor domes (down) - observation lounges looking "down" at planets
//   - Wall domes (north/south/east/west) - side-facing observation windows
//
// Controls:
//
//	Mouse: Look around inside dome
//	WASD: Move camera position
//	Q/E: Up/Down
//	+/-: Adjust dome radius
//	D: Cycle dome direction (up, north, south, east, west, down)
//	V: Cycle velocity (SR effects - aberration, Doppler)
//	G: Cycle gravity (GR effects - lensing, redshift)
//	F: Toggle dome struts (orientation indicators)
//	P: Toggle nearby planets (LOD demo)
//	R: Reset camera to Sol
//	Shift: Fast movement
//	Esc: Quit
//
// Usage:
//
//	go run ./cmd/demo-engine-dome
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
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/solarlune/tetra3d"
	"stapledons_voyage/engine/lod"
	"stapledons_voyage/engine/render"
	"stapledons_voyage/engine/tetra"
)

const (
	screenWidth  = 1280
	screenHeight = 720
	texWidth     = 2048 // Equirectangular texture width
	texHeight    = 1024 // Height is half of width for equirectangular
)

var (
	screenshotFrame = flag.Int("screenshot", 0, "Take screenshot at frame N and exit")
	screenshotPath  = flag.String("output", "out/screenshots/dome.png", "Screenshot output path")
)

// NearbyObject represents a celestial object rendered via LOD
type NearbyObject struct {
	lodObj *lod.Object
	planet *tetra.Planet
	name   string
}

// Game implements ebiten.Game for the dome demo.
type Game struct {
	scene     *tetra.Scene
	skySphere *tetra.Dome // Large sphere at "infinity" with starfield texture
	spaceView *render.SpaceView
	starTex   *ebiten.Image
	frameCount int

	// LOD system for nearby objects
	lodManager        *lod.Manager
	lodCamera         *lod.SimpleCamera
	nearbyObjects     []*NearbyObject
	showNearbyObjects bool
	sunTexture        *ebiten.Image

	// LOD renderers
	pointRenderer  *lod.PointRenderer
	circleRenderer *lod.CircleRenderer

	// Camera controls
	yaw      float64
	pitch    float64
	lastX    int
	lastY    int
	camX     float64
	camY     float64
	camZ     float64
	camSpeed float64

	// Dome parameters
	domeRadius    float64
	domeDirection string
	directions    []string
	dirIndex      int
	velocity      float64 // Ship velocity for SR effects (0-0.8c)
	grPhi         float64 // Gravitational potential for GR effects (0-0.4)

	// Observation platform and dome structure
	platform   *tetra3d.Model
	struts     []*tetra3d.Model
	showStruts bool

	// Performance
	lastUpdate time.Time
	fps        float64
}

// SpectralInfo contains visual properties for a spectral type
type SpectralInfo struct {
	Color color.RGBA
	Temp  int
}

var spectralTypes = map[string]SpectralInfo{
	"O": {color.RGBA{155, 176, 255, 255}, 33000},
	"B": {color.RGBA{170, 191, 255, 255}, 15000},
	"A": {color.RGBA{202, 215, 255, 255}, 8500},
	"F": {color.RGBA{248, 247, 255, 255}, 6500},
	"G": {color.RGBA{255, 244, 234, 255}, 5800},
	"K": {color.RGBA{255, 210, 161, 255}, 4500},
	"M": {color.RGBA{255, 180, 100, 255}, 3200},
}

// StarData represents a star from the JSON catalog
type StarData struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
	Z        float64 `json:"z"`
	DistLY   float64 `json:"dist_ly"`
	VMag     float64 `json:"vmag"`
	Spectral string  `json:"spectral"`
}

// StarCatalog represents the JSON file structure
type StarCatalog struct {
	Version string     `json:"version"`
	Source  string     `json:"source"`
	Count   int        `json:"count"`
	Stars   []StarData `json:"stars"`
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

// NewGame creates a new dome demo game.
func NewGame() *Game {
	g := &Game{
		domeRadius:        5.0,
		domeDirection:     "up",
		directions:        []string{"up", "north", "south", "east", "west", "down"},
		dirIndex:          0,
		velocity:          0.0,
		grPhi:             0.0, // No GR by default (G key to cycle)
		yaw:               math.Pi, // Start looking toward Sol (toward -Z)
		pitch:             -0.15,  // Look slightly down to see Sol (which is at Y=0)
		camX:              0,
		camY:              0.3, // Start slightly above the platform
		camZ:              2.0, // Start 2 ly away from Sol
		camSpeed:          0.5,
		showNearbyObjects: true, // Show 3D LOD objects outside the dome
		showStruts:        true,  // Show dome structure by default
		lastUpdate:        time.Now(),
	}

	// Create 3D scene
	g.scene = tetra.NewScene(screenWidth, screenHeight)
	g.scene.SetLightingEnabled(true) // Enable lighting for 3D objects

	// Create SKY SPHERE at "infinity" - this shows the starfield background
	// Large radius so 3D objects appear IN FRONT of it
	const skyRadius = 500.0 // Large enough to be "background"
	g.skySphere = tetra.NewBubble("sky_sphere", skyRadius)
	g.skySphere.AddToScene(g.scene)
	g.skySphere.SetPosition(0, 0, 0) // Centered at world origin

	// The dome structure is visible but transparent (struts only, no texture)
	// We don't create a textured dome - the sky sphere IS the view

	// Create observation platform (floor)
	g.createPlatform()

	// Create dome struts for orientation (visible ship structure)
	g.createStruts()

	// Load space view for star rendering - CORRECT PATH
	g.spaceView = render.NewSpaceView()
	if err := g.spaceView.Load("assets/data/starmap/stars.json"); err != nil {
		log.Printf("Warning: could not load star catalog: %v (using procedural)", err)
	} else {
		log.Printf("Loaded %d stars from CNS5 catalog", g.spaceView.GetStarCount())
	}

	// Load sun texture for 3D stars
	g.sunTexture = loadTexture("assets/stars/sun_8k.jpg")
	if g.sunTexture == nil {
		g.sunTexture = loadTexture("assets/stars/sun_2k.jpg")
	}
	if g.sunTexture == nil {
		g.sunTexture = loadTexture("assets/planets/sun.jpg")
	}

	// Create LOD manager for nearby objects
	config := lod.DefaultConfig()
	config.Max3DObjects = 20
	config.Full3DPixels = 15
	config.BillboardPixels = 3
	config.CirclePixels = 1
	config.PointPixels = 0.3
	config.TransitionTime = 0.2
	g.lodManager = lod.NewManager(config)

	// Create LOD camera
	g.lodCamera = lod.NewSimpleCamera(screenWidth, screenHeight)
	g.lodCamera.Fov = 90
	g.lodCamera.Far = 1000

	// Create LOD renderers
	g.pointRenderer = lod.NewPointRenderer()
	g.circleRenderer = lod.NewCircleRenderer()

	// Load nearby stars as LOD objects
	g.loadNearbyStars()

	// Generate initial star texture
	g.updateStarTexture()

	// Position camera inside the dome
	g.scene.SetCameraPosition(0, 0, 0)

	// Add ambient light
	ambient := tetra.NewAmbientLight(0.2, 0.2, 0.25, 1.0)
	ambient.AddToScene(g.scene)

	return g
}

// loadNearbyStars loads the closest stars as LOD objects
func (g *Game) loadNearbyStars() {
	// Load star catalog
	data, err := os.ReadFile("assets/data/starmap/stars.json")
	if err != nil {
		log.Printf("Could not load star catalog for LOD: %v", err)
		return
	}

	var catalog StarCatalog
	if err := json.Unmarshal(data, &catalog); err != nil {
		log.Printf("Could not parse star catalog: %v", err)
		return
	}

	// Add Sol at origin
	g.addNearbyObject("Sol", 0, 0, 0, "G", 0.3)

	// Add closest stars (within 5 light-years for demo visibility)
	maxDist := 5.0
	count := 0
	for _, star := range catalog.Stars {
		if star.DistLY < maxDist && count < 30 {
			// Scale positions for visibility in dome (1 unit = 1 ly)
			g.addNearbyObject(star.Name, star.X, star.Y, star.Z, star.Spectral, 0.15)
			count++
		}
	}

	log.Printf("Added %d nearby stars as LOD objects", count+1)
}

// addNearbyObject adds a celestial object to the LOD system
func (g *Game) addNearbyObject(name string, x, y, z float64, spectral string, radius float64) {
	info := spectralTypes["G"] // Default
	if s, ok := spectralTypes[spectral]; ok {
		info = s
	}

	pos := lod.Vector3{X: x, Y: y, Z: z}
	lodObj := lod.NewObject(name, pos, radius, info.Color)
	lodObj.Luminosity = 100 // All stars emit light
	lodObj.LightColor = info.Color
	g.lodManager.Add(lodObj)

	// Create 3D planet for Full3D tier
	var planet *tetra.Planet
	if g.sunTexture != nil {
		planet = tetra.NewTexturedPlanet(name, radius, g.sunTexture)
	} else {
		planet = tetra.NewPlanet(name, radius, info.Color)
	}
	planet.AddToScene(g.scene)
	planet.SetPosition(x, y, z)
	planet.SetShadeless(true) // Stars are self-luminous

	// Apply spectral color modulation
	planet.SetColorModulation(
		float64(info.Color.R)/255.0,
		float64(info.Color.G)/255.0,
		float64(info.Color.B)/255.0,
	)

	g.nearbyObjects = append(g.nearbyObjects, &NearbyObject{
		lodObj: lodObj,
		planet: planet,
		name:   name,
	})
}

// createPlatform creates an observation platform/floor for orientation
func (g *Game) createPlatform() {
	// Create a large floor mesh that extends beyond dome base
	platformMesh := tetra.NewPlaneMesh("platform", 15.0, 15.0)

	// Create visible floor material - darker with slight metallic look
	mat := tetra3d.NewMaterial("platform_mat")
	mat.Color = tetra3d.NewColor(0.15, 0.15, 0.18, 1.0) // Dark gray floor
	mat.BackfaceCulling = false // Visible from below too
	if len(platformMesh.MeshParts) > 0 {
		platformMesh.MeshParts[0].Material = mat
	}

	g.platform = tetra3d.NewModel("platform", platformMesh)
	g.platform.SetLocalPosition(0, -2.0, 0) // Well below camera to avoid clipping issues
	g.scene.Root().AddChildren(g.platform)
}

// createStruts creates dome structure indicators (struts/ribs)
// Uses thin ribbon geometry for visibility
func (g *Game) createStruts() {
	strutColor := tetra3d.NewColor(0.5, 0.6, 0.7, 1.0) // Light blue-gray
	strutWidth := float32(0.015)                       // Thin struts

	radius := float32(g.domeRadius)
	innerRadius := radius * 0.98 // Struts slightly inside dome

	// Create vertical struts (meridians) - 8 around the dome
	numMeridians := 8
	for i := 0; i < numMeridians; i++ {
		angle := float64(i) * 2.0 * math.Pi / float64(numMeridians)

		strutMesh := tetra3d.NewMesh(fmt.Sprintf("strut_meridian_%d", i))
		numSegments := 16
		var vertices []tetra3d.VertexInfo

		// Create ribbon with inner and outer edges (radial offset, not perpendicular)
		for j := 0; j <= numSegments; j++ {
			phi := float64(j) * math.Pi / 2.0 / float64(numSegments)

			// Direction from center
			dirX := float32(math.Sin(phi) * math.Cos(angle))
			dirY := float32(math.Cos(phi))
			dirZ := float32(math.Sin(phi) * math.Sin(angle))

			// Inner edge (inside dome)
			vertices = append(vertices, tetra3d.VertexInfo{
				X: dirX * innerRadius, Y: dirY * innerRadius, Z: dirZ * innerRadius,
				NormalX: dirX, NormalY: dirY, NormalZ: dirZ,
				U: 0, V: float32(j) / float32(numSegments),
			})
			// Outer edge (on dome surface)
			outerR := innerRadius + strutWidth
			vertices = append(vertices, tetra3d.VertexInfo{
				X: dirX * outerR, Y: dirY * outerR, Z: dirZ * outerR,
				NormalX: dirX, NormalY: dirY, NormalZ: dirZ,
				U: 1, V: float32(j) / float32(numSegments),
			})
		}
		strutMesh.AddVertices(vertices...)

		// Create quads (two triangles per segment)
		mat := tetra3d.NewMaterial(fmt.Sprintf("strut_mat_%d", i))
		mat.Color = strutColor
		mat.BackfaceCulling = false // Visible from both sides

		var indices []int
		for j := 0; j < numSegments; j++ {
			base := j * 2
			// First triangle
			indices = append(indices, base, base+2, base+1)
			// Second triangle
			indices = append(indices, base+1, base+2, base+3)
		}
		if len(indices) > 0 {
			strutMesh.AddMeshPart(mat, indices...)
		}
		strutMesh.UpdateBounds()

		strut := tetra3d.NewModel(fmt.Sprintf("strut_meridian_%d", i), strutMesh)
		g.scene.Root().AddChildren(strut)
		g.struts = append(g.struts, strut)
	}

	// Create horizontal rings (parallels) - 3 rings at different heights
	// Use vertical offset so rings are visible as bands when looking up
	ringHeights := []float64{0.25, 0.5, 0.75}
	ringWidth := strutWidth * 3 // Make rings wider for visibility
	for ri, heightFrac := range ringHeights {
		phi := math.Acos(heightFrac)

		ringMesh := tetra3d.NewMesh(fmt.Sprintf("strut_ring_%d", ri))
		numRingSegments := 32
		var vertices []tetra3d.VertexInfo

		// Ring position on dome
		ringRadius := float32(math.Sin(phi)) * innerRadius
		ringY := float32(math.Cos(phi)) * innerRadius

		for j := 0; j <= numRingSegments; j++ {
			theta := float64(j) * 2.0 * math.Pi / float64(numRingSegments)
			x := ringRadius * float32(math.Cos(theta))
			z := ringRadius * float32(math.Sin(theta))

			// Normal pointing inward (toward center)
			nx := -float32(math.Cos(theta))
			nz := -float32(math.Sin(theta))

			// Bottom edge of ring band
			vertices = append(vertices, tetra3d.VertexInfo{
				X: x, Y: ringY - ringWidth, Z: z,
				NormalX: nx, NormalY: 0, NormalZ: nz,
				U: float32(j) / float32(numRingSegments), V: 0,
			})
			// Top edge of ring band
			vertices = append(vertices, tetra3d.VertexInfo{
				X: x, Y: ringY + ringWidth, Z: z,
				NormalX: nx, NormalY: 0, NormalZ: nz,
				U: float32(j) / float32(numRingSegments), V: 1,
			})
		}
		ringMesh.AddVertices(vertices...)

		mat := tetra3d.NewMaterial(fmt.Sprintf("ring_mat_%d", ri))
		mat.Color = strutColor
		mat.BackfaceCulling = false

		var indices []int
		for j := 0; j < numRingSegments; j++ {
			base := j * 2
			indices = append(indices, base, base+2, base+1)
			indices = append(indices, base+1, base+2, base+3)
		}
		if len(indices) > 0 {
			ringMesh.AddMeshPart(mat, indices...)
		}
		ringMesh.UpdateBounds()

		ring := tetra3d.NewModel(fmt.Sprintf("strut_ring_%d", ri), ringMesh)
		g.scene.Root().AddChildren(ring)
		g.struts = append(g.struts, ring)
	}

	// Add base ring at equator (slightly thicker)
	baseMesh := tetra3d.NewMesh("strut_base")
	numBaseSegments := 48
	baseThickness := strutWidth * 1.5
	var baseVerts []tetra3d.VertexInfo

	for j := 0; j <= numBaseSegments; j++ {
		theta := float64(j) * 2.0 * math.Pi / float64(numBaseSegments)
		x := innerRadius * float32(math.Cos(theta))
		z := innerRadius * float32(math.Sin(theta))

		// Bottom edge (at Y=0)
		baseVerts = append(baseVerts, tetra3d.VertexInfo{
			X: x, Y: 0, Z: z,
			NormalX: 0, NormalY: -1, NormalZ: 0,
			U: float32(j) / float32(numBaseSegments), V: 0,
		})
		// Top edge
		baseVerts = append(baseVerts, tetra3d.VertexInfo{
			X: x, Y: baseThickness, Z: z,
			NormalX: 0, NormalY: 1, NormalZ: 0,
			U: float32(j) / float32(numBaseSegments), V: 1,
		})
	}
	baseMesh.AddVertices(baseVerts...)

	baseMat := tetra3d.NewMaterial("base_mat")
	baseMat.Color = tetra3d.NewColor(0.6, 0.7, 0.8, 1.0) // Brighter base
	baseMat.BackfaceCulling = false

	var baseIndices []int
	for j := 0; j < numBaseSegments; j++ {
		base := j * 2
		baseIndices = append(baseIndices, base, base+2, base+1)
		baseIndices = append(baseIndices, base+1, base+2, base+3)
	}
	if len(baseIndices) > 0 {
		baseMesh.AddMeshPart(baseMat, baseIndices...)
	}
	baseMesh.UpdateBounds()

	baseRing := tetra3d.NewModel("strut_base", baseMesh)
	g.scene.Root().AddChildren(baseRing)
	g.struts = append(g.struts, baseRing)

	log.Printf("Created %d dome struts (thin ribbon geometry)", len(g.struts))
}

// updateStrutsVisibility shows or hides dome struts
func (g *Game) updateStrutsVisibility() {
	for _, strut := range g.struts {
		strut.SetVisible(g.showStruts, true)
	}
}

// rebuildDomeAndStruts removes and recreates struts with current radius
// Note: The actual dome is just the struts (transparent glass) - sky sphere shows the view
func (g *Game) rebuildDomeAndStruts() {
	// Remove old struts
	for _, strut := range g.struts {
		strut.Unparent()
	}
	g.struts = nil

	// Recreate struts with new radius (dome structure only, no texture)
	g.createStruts()
	g.updateStrutsVisibility()
}

// updateStarTexture regenerates the equirectangular star texture.
func (g *Game) updateStarTexture() {
	// Use camera look direction for SR/GR effects
	// This means we're "traveling" in the direction we're facing
	lookDirX := math.Sin(g.yaw) * math.Cos(g.pitch)
	lookDirY := math.Sin(g.pitch)
	lookDirZ := math.Cos(g.yaw) * math.Cos(g.pitch)

	params := render.ViewParams{
		ShipX:    g.camX,
		ShipY:    g.camY,
		ShipZ:    g.camZ,
		ViewDirX: lookDirX,
		ViewDirY: lookDirY,
		ViewDirZ: lookDirZ,
		UpX:      0,
		UpY:      1,
		UpZ:      0,
		FOV:      90,
		Velocity: g.velocity,
		GrPhi:    g.grPhi,
	}

	g.starTex = g.spaceView.RenderEquirectangular(params, texWidth, texHeight)
	g.skySphere.SetTexture(g.starTex) // Apply to sky sphere (at infinity)
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

	// Mouse look
	x, y := ebiten.CursorPosition()
	if g.frameCount > 0 {
		dx := float64(x - g.lastX)
		dy := float64(y - g.lastY)

		g.yaw -= dx * 0.003
		g.pitch -= dy * 0.003

		// Clamp pitch
		if g.pitch > math.Pi/2-0.1 {
			g.pitch = math.Pi/2 - 0.1
		}
		if g.pitch < -math.Pi/2+0.1 {
			g.pitch = -math.Pi/2 + 0.1
		}
	}
	g.lastX = x
	g.lastY = y

	// Camera movement (WASD)
	moveSpeed := g.camSpeed * dt * 60
	if ebiten.IsKeyPressed(ebiten.KeyW) {
		g.camZ -= moveSpeed * math.Cos(g.yaw)
		g.camX += moveSpeed * math.Sin(g.yaw)
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) {
		g.camZ += moveSpeed * math.Cos(g.yaw)
		g.camX -= moveSpeed * math.Sin(g.yaw)
	}
	if ebiten.IsKeyPressed(ebiten.KeyA) {
		g.camX -= moveSpeed * math.Cos(g.yaw)
		g.camZ -= moveSpeed * math.Sin(g.yaw)
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) && !inpututil.IsKeyJustPressed(ebiten.KeyD) {
		g.camX += moveSpeed * math.Cos(g.yaw)
		g.camZ += moveSpeed * math.Sin(g.yaw)
	}
	if ebiten.IsKeyPressed(ebiten.KeyQ) {
		g.camY += moveSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyE) {
		g.camY -= moveSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyShift) {
		g.camSpeed = 2.0
	} else {
		g.camSpeed = 0.5
	}

	// D key reserved for future dome direction control
	// Currently struts are fixed upward orientation

	// Toggle nearby objects
	if inpututil.IsKeyJustPressed(ebiten.KeyP) {
		g.showNearbyObjects = !g.showNearbyObjects
	}

	// Reset camera to starting position (looking at Sol)
	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		g.camX = 0
		g.camY = 0.3 // Above the platform
		g.camZ = 2.0
		g.yaw = math.Pi // Look toward Sol (toward -Z)
		g.pitch = 0
		g.updateStarTexture()
	}

	// Cycle velocity (SR effects)
	if inpututil.IsKeyJustPressed(ebiten.KeyV) {
		g.velocity += 0.2
		if g.velocity > 0.8 {
			g.velocity = 0
		}
		g.updateStarTexture()
	}

	// Cycle gravity (GR effects)
	if inpututil.IsKeyJustPressed(ebiten.KeyG) {
		g.grPhi += 0.1
		if g.grPhi > 0.4 {
			g.grPhi = 0
		}
		g.updateStarTexture()
	}

	// Toggle dome struts visibility
	if inpututil.IsKeyJustPressed(ebiten.KeyF) {
		g.showStruts = !g.showStruts
		g.updateStrutsVisibility()
	}

	// Adjust dome radius
	if ebiten.IsKeyPressed(ebiten.KeyEqual) || ebiten.IsKeyPressed(ebiten.KeyKPAdd) {
		g.domeRadius += 0.05
		if g.domeRadius > 20 {
			g.domeRadius = 20
		}
		g.rebuildDomeAndStruts()
	}
	if ebiten.IsKeyPressed(ebiten.KeyMinus) || ebiten.IsKeyPressed(ebiten.KeyKPSubtract) {
		g.domeRadius -= 0.05
		if g.domeRadius < 2 {
			g.domeRadius = 2
		}
		g.rebuildDomeAndStruts()
	}

	// Update camera position
	g.scene.SetCameraPosition(g.camX, g.camY, g.camZ)

	// Update sky sphere to follow camera (always surrounds the viewer)
	g.skySphere.SetPosition(g.camX, g.camY, g.camZ)

	// Update platform and struts to follow camera (ship interior)
	if g.platform != nil {
		g.platform.SetLocalPosition(float32(g.camX), float32(g.camY-2.0), float32(g.camZ))
	}
	for _, strut := range g.struts {
		strut.SetLocalPosition(float32(g.camX), float32(g.camY), float32(g.camZ))
	}

	// Update camera based on look direction
	lookX := g.camX + math.Cos(g.pitch)*math.Sin(g.yaw)
	lookY := g.camY + math.Sin(g.pitch)
	lookZ := g.camZ + math.Cos(g.pitch)*math.Cos(g.yaw)
	g.scene.LookAt(lookX, lookY, lookZ)

	// Update LOD camera
	g.lodCamera.Pos = lod.Vector3{X: g.camX, Y: g.camY, Z: g.camZ}
	g.lodCamera.LookAt = lod.Vector3{X: lookX, Y: lookY, Z: lookZ}

	// Update LOD manager
	if g.showNearbyObjects {
		g.lodManager.UpdateWithDT(g.lodCamera, dt)

		// Update planet visibility and position based on LOD tier
		// Position 3D objects in space - close to the sky sphere so they appear as distant objects
		for _, obj := range g.nearbyObjects {
			if obj.lodObj.CurrentTier == lod.TierFull3D {
				// Calculate direction from camera to actual object position
				objPos := obj.lodObj.Position
				dx := objPos.X - g.camX
				dy := objPos.Y - g.camY
				dz := objPos.Z - g.camZ
				dist := math.Sqrt(dx*dx + dy*dy + dz*dz)

				if dist > 0.001 {
					// Normalize direction
					dx /= dist
					dy /= dist
					dz /= dist

					// Don't render objects far below the horizon
					// Allow slight negative (near horizon) but hide anything steeply below
					if dy < -0.5 {
						obj.planet.Model().SetVisible(false, true)
						continue
					}

					obj.planet.Model().SetVisible(true, true)

					// Position just inside the sky sphere (450 units, sky is 500)
					// This places objects at "astronomical distance" visually
					const objectDistance = 450.0
					newX := g.camX + dx*objectDistance
					newY := g.camY + dy*objectDistance
					newZ := g.camZ + dz*objectDistance
					obj.planet.SetPosition(newX, newY, newZ)

					// Scale based on angular size - smaller at distance
					// At 2 ly reference, object should be reasonably visible
					referenceDistance := 2.0
					angularScale := referenceDistance / dist
					// Much smaller scale since objects are at 450 units
					visualScale := angularScale * 15.0
					if visualScale > 30.0 {
						visualScale = 30.0 // Cap maximum
					}
					if visualScale < 3.0 {
						visualScale = 3.0 // Minimum visible
					}
					s := float32(visualScale)
					obj.planet.Model().SetLocalScale(s, s, s)
				} else {
					obj.planet.Model().SetVisible(false, true)
				}
			} else {
				obj.planet.Model().SetVisible(false, true)
			}
		}
	} else {
		// Hide all planets when LOD is disabled
		for _, obj := range g.nearbyObjects {
			obj.planet.Model().SetVisible(false, true)
		}
	}

	g.frameCount++
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Clear with dark background
	screen.Fill(color.RGBA{5, 5, 10, 255})

	// Draw LOD objects (points and circles) BEHIND the dome
	if g.showNearbyObjects {
		points := g.lodManager.GetTierPoint()
		circles := g.lodManager.GetTierCircle()
		g.pointRenderer.RenderPoints(screen, points)
		g.circleRenderer.RenderCircles(screen, circles)
	}

	// Render the 3D scene (dome + Full3D objects)
	rendered := g.scene.Render()
	screen.DrawImage(rendered, nil)

	// HUD
	velocityStr := "Stationary"
	if g.velocity > 0 {
		velocityStr = fmt.Sprintf("%.0f%% c (SR active)", g.velocity*100)
	}

	grStr := "None"
	if g.grPhi > 0 {
		grStr = fmt.Sprintf("φ=%.1f (GR active)", g.grPhi)
	}

	lodStatus := "ON"
	if !g.showNearbyObjects {
		lodStatus = "OFF"
	}

	strutsStatus := "ON"
	if !g.showStruts {
		strutsStatus = "OFF"
	}

	stats := g.lodManager.Stats()

	hudText := fmt.Sprintf(
		"Bubble Ship Dome Demo (with Starmap & LOD)\n"+
			"═══════════════════════════════════════════════\n"+
			"Mouse: Look | WASD: Move | Q/E: Up/Down\n"+
			"+/-: Radius | D: Direction | V: Velocity\n"+
			"G: Gravity | F: Struts | P: Toggle LOD\n"+
			"R: Reset | Shift: Fast | Esc: Quit\n"+
			"═══════════════════════════════════════════════\n"+
			"FPS: %.1f\n"+
			"Dome: %.1f m (%s) | Struts: %s\n"+
			"Ship Velocity: %s\n"+
			"Gravity Well: %s\n"+
			"Stars in Catalog: %d\n"+
			"Camera: (%.2f, %.2f, %.2f) ly\n"+
			"═══════════════════════════════════════════════\n"+
			"LOD System: %s\n"+
			"  Full3D:    %d\n"+
			"  Billboard: %d\n"+
			"  Circle:    %d\n"+
			"  Point:     %d\n"+
			"  Culled:    %d\n"+
			"═══════════════════════════════════════════════\n"+
			"Yaw: %.1f° | Pitch: %.1f° | Frame: %d",
		g.fps,
		g.domeRadius, g.domeDirection, strutsStatus,
		velocityStr,
		grStr,
		g.spaceView.GetStarCount(),
		g.camX, g.camY, g.camZ,
		lodStatus,
		stats.Full3DCount,
		stats.BillboardCount,
		stats.CircleCount,
		stats.PointCount,
		stats.CulledCount,
		g.yaw*180/math.Pi,
		g.pitch*180/math.Pi,
		g.frameCount,
	)
	ebitenutil.DebugPrint(screen, hudText)

	// Screenshot
	if *screenshotFrame > 0 && g.frameCount >= *screenshotFrame {
		g.saveScreenshot(screen)
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

	fmt.Println("Bubble Ship Dome Demo")
	fmt.Println("  Real star data from CNS5 catalog (3,802 stars)")
	fmt.Println("  LOD system for nearby objects")

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Bubble Ship Dome Demo - Starmap & LOD")
	ebiten.SetCursorMode(ebiten.CursorModeCaptured)

	game := NewGame()

	if err := ebiten.RunGame(game); err != nil && err != ebiten.Termination {
		log.Fatal(err)
	}
}
