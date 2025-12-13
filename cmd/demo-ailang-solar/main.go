// Package main provides an AILANG-controlled solar system demo.
// This validates the AILANG-first architecture by having AILANG control all celestial data
// while the Go engine only handles rendering with LOD support.
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

	"stapledons_voyage/engine/lod"
	"stapledons_voyage/engine/shader"
	"stapledons_voyage/engine/tetra"
	"stapledons_voyage/sim_gen"
)

const (
	screenWidth  = 1280
	screenHeight = 720
)

// Planet3D represents a planet that can be rendered with LOD
type Planet3D struct {
	lodObj    *lod.Object
	planet    *tetra.Planet
	texture   *ebiten.Image
	billboard *ebiten.Image
	rings     *tetra.RingSystem
}

// Game implements ebiten.Game interface for the AILANG solar demo.
type Game struct {
	// AILANG state
	state *sim_gen.SolarDemoState

	// LOD system
	lodManager        *lod.Manager
	lodCamera         *lod.SimpleCamera
	pointRenderer     *lod.PointRenderer
	circleRenderer    *lod.CircleRenderer
	billboardRenderer *lod.BillboardRenderer

	// Tetra3D scene for 3D rendering
	scene3D *tetra.Scene

	// SR/GR shader effects
	shaderManager *shader.Manager
	srWarp        *shader.SRWarp
	grWarp        *shader.GRWarp
	renderBuffer  *ebiten.Image

	// Light sources
	sunLight     *tetra.StarLight
	ambientLight *tetra.AmbientLight

	// Planets with LOD tracking
	planets       []*Planet3D
	planetSprites map[string]*ebiten.Image

	// Camera movement
	cameraSpeed float64
	lastUpdate  time.Time

	// Cruise mode
	cruiseMode   bool
	cruiseTarget int // Index of target planet

	// Lighting controls
	lightMultiplier float64
	ambientLevel    float64

	// Velocity (fraction of c)
	velocity float64

	// Frame counter
	frameCount      int
	screenshotFrame int
	screenshotPath  string
	screenshotTaken bool

	// Simulation time for orbits
	simTime float64

	// Billboard initialization flag
	billboardsInitialized bool

	// FPS tracking
	fpsLastTime   time.Time
	fpsFrameCount int
	fpsCurrent    float64
}

// NewGame creates a new AILANG solar demo.
func NewGame(screenshotFrame int, screenshotPath string) *Game {
	// Initialize AILANG state
	state := sim_gen.InitSolarDemo()

	// Create LOD manager - handle 60+ objects, most rendered as points/billboards
	config := lod.DefaultConfig()
	config.Max3DObjects = 15 // Limit full 3D rendering to nearby objects
	lodManager := lod.NewManager(config)

	// Create LOD camera
	lodCamera := lod.NewSimpleCamera(screenWidth, screenHeight)
	lodCamera.Fov = 60
	lodCamera.Far = 20000
	lodCamera.Pos = lod.Vector3{X: state.CameraX, Y: state.CameraY, Z: state.CameraZ}
	lodCamera.LookAt = lod.Vector3{X: state.LookAtX, Y: state.LookAtY, Z: state.LookAtZ}

	// Create Tetra3D scene
	scene3D := tetra.NewScene(screenWidth, screenHeight)
	scene3D.SetLightingEnabled(true)

	// Set camera from AILANG state
	scene3D.SetCameraPosition(state.CameraX, state.CameraY, state.CameraZ)
	scene3D.LookAt(state.LookAtX, state.LookAtY, state.LookAtZ)

	// Create 3D bodies from AILANG data (60+ objects: planets, moons, dwarf planets, asteroids)
	ailangPlanets := sim_gen.GetAllSolarSystemBodies()
	var planets []*Planet3D
	planetSprites := make(map[string]*ebiten.Image)

	// Build initial position map for hierarchical orbits
	initialPositions := make(map[string][2]float64) // name -> (x, z)

	// First pass: calculate positions for primary bodies (planets, no parent)
	for _, p := range ailangPlanets {
		if p.ParentName == "" {
			posX := p.OrbitRadius * math.Cos(p.OrbitPhase)
			posZ := p.OrbitRadius * math.Sin(p.OrbitPhase)
			if p.OrbitRadius == 0 {
				posX, posZ = 0, 0
			}
			initialPositions[p.Name] = [2]float64{posX, posZ}
		}
	}

	// Second pass: calculate positions for moons (relative to parents)
	for _, p := range ailangPlanets {
		if p.ParentName != "" {
			parentPos, ok := initialPositions[p.ParentName]
			if ok {
				offsetX := p.OrbitRadius * math.Cos(p.OrbitPhase)
				offsetZ := p.OrbitRadius * math.Sin(p.OrbitPhase)
				posX := parentPos[0] + offsetX
				posZ := parentPos[1] + offsetZ
				initialPositions[p.Name] = [2]float64{posX, posZ}
			} else {
				// Fallback to Sun orbit if parent not found
				posX := p.OrbitRadius * math.Cos(p.OrbitPhase)
				posZ := p.OrbitRadius * math.Sin(p.OrbitPhase)
				initialPositions[p.Name] = [2]float64{posX, posZ}
			}
		}
	}

	for _, p := range ailangPlanets {
		col := rgbaFromInt(p.ColorRgba)

		// Get initial position from hierarchical calculation
		pos := initialPositions[p.Name]
		posX, posZ := pos[0], pos[1]

		// Create LOD object
		lodObj := lod.NewObject(p.Name, lod.Vector3{X: posX, Y: 0, Z: posZ}, p.Radius, col)
		if p.Name == "Sun" {
			lodObj.Luminosity = state.SunEnergy
			lodObj.LightColor = color.RGBA{255, 243, 217, 255} // G-type star color
		}
		lodManager.Add(lodObj)

		// Load texture
		texPath := fmt.Sprintf("assets/planets/%s.jpg", p.TextureName)
		tex := loadTexture(texPath)

		// Create 3D planet (textured if available)
		var planet *tetra.Planet
		if tex != nil {
			planet = tetra.NewTexturedPlanet(p.Name, p.Radius, tex)
			log.Printf("Loaded texture for %s from %s", p.Name, texPath)
		} else {
			planet = tetra.NewPlanet(p.Name, p.Radius, col)
			log.Printf("Using solid color for %s (texture not found: %s)", p.Name, texPath)
		}
		planet.SetPosition(posX, 0, posZ)
		planet.AddToScene(scene3D)

		// Sun should be self-illuminated (shadeless)
		if p.Name == "Sun" {
			planet.SetShadeless(true)
		}

		// Add rings to ringed planets (Saturn, Uranus) with physically accurate specs
		var rings *tetra.RingSystem
		if p.HasRings {
			var bands []tetra.RingBand
			var tilt float64
			switch p.Name {
			case "Saturn":
				bands = tetra.SaturnRingBands(p.Radius)
				tilt = 0.47 // Saturn's axial tilt ~27°
			case "Uranus":
				bands = tetra.UranusRingBands(p.Radius)
				tilt = 1.71 // Uranus's extreme axial tilt ~98°
			default:
				bands = tetra.SaturnRingBands(p.Radius) // fallback
				tilt = 0.47
			}
			rings = tetra.NewRingSystem(p.Name, bands)
			rings.AddToScene(scene3D)
			rings.SetPosition(posX, 0, posZ)
			rings.SetTilt(tilt)
			log.Printf("Added ring system to %s with %d bands (tilt=%.2f rad)", p.Name, len(bands), tilt)
		}

		planets = append(planets, &Planet3D{
			lodObj:    lodObj,
			planet:    planet,
			texture:   tex,
			billboard: nil, // Created lazily
			rings:     rings,
		})
	}

	// Create sun light from AILANG data
	sunLight := tetra.NewStarLight("sun_light", 1.0, 0.95, 0.85, state.SunEnergy, 0)
	sunLight.SetPosition(0, 0, 0)
	sunLight.AddToScene(scene3D)

	// Create ambient light
	ambientLight := tetra.NewAmbientLight(0.08, 0.08, 0.1, state.AmbientLevel)
	ambientLight.AddToScene(scene3D)

	// Create LOD renderers
	billboardRenderer := lod.NewBillboardRenderer()
	defaultSprite := lod.CreateDefaultPlanetSprite(128, color.RGBA{255, 255, 255, 255})
	billboardRenderer.SetDefaultSprite(defaultSprite)

	// Create shader manager for SR/GR effects
	shaderManager := shader.NewManager()
	srWarp := shader.NewSRWarp(shaderManager)
	grWarp := shader.NewGRWarp(shaderManager)
	grWarp.SetDemoMode(0.5, 0.5, 0.08, 0.01)

	// Create off-screen render buffer for post-processing
	renderBuffer := ebiten.NewImage(screenWidth, screenHeight)

	log.Printf("AILANG Solar Demo: Loaded %d planets from AILANG with LOD support", len(ailangPlanets))
	log.Printf("  Sun Energy: %.0f, Ambient: %.2f", state.SunEnergy, state.AmbientLevel)

	return &Game{
		state:             state,
		lodManager:        lodManager,
		lodCamera:         lodCamera,
		pointRenderer:     lod.NewPointRenderer(),
		circleRenderer:    lod.NewCircleRenderer(),
		billboardRenderer: billboardRenderer,
		scene3D:           scene3D,
		shaderManager:     shaderManager,
		srWarp:            srWarp,
		grWarp:            grWarp,
		renderBuffer:      renderBuffer,
		sunLight:          sunLight,
		ambientLight:      ambientLight,
		planets:           planets,
		planetSprites:     planetSprites,
		cameraSpeed:       50.0,
		lastUpdate:        time.Now(),
		lightMultiplier:   1.0,
		ambientLevel:      1.0,
		velocity:          0.0,
		screenshotFrame:   screenshotFrame,
		screenshotPath:    screenshotPath,
	}
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

// initializeBillboards creates billboard sprites from planet textures.
func (g *Game) initializeBillboards() {
	for _, p := range g.planets {
		if p.texture != nil && p.billboard == nil {
			avgColor := lod.ExtractAverageColor(p.texture)
			p.lodObj.Color = avgColor
			p.billboard = lod.CreateBillboardFromTexture(p.texture, 128)
			g.planetSprites[p.lodObj.ID] = p.billboard
			log.Printf("Created billboard for %s from texture", p.lodObj.ID)
		} else if p.billboard == nil {
			p.billboard = lod.CreateDefaultPlanetSprite(128, p.lodObj.Color)
			g.planetSprites[p.lodObj.ID] = p.billboard
		}
	}
}

// Update implements ebiten.Game interface.
func (g *Game) Update() error {
	g.frameCount++

	// Calculate delta time
	now := time.Now()
	dt := now.Sub(g.lastUpdate).Seconds()
	g.lastUpdate = now

	// FPS calculation (update every second)
	g.fpsFrameCount++
	if g.fpsLastTime.IsZero() {
		g.fpsLastTime = now
	} else if now.Sub(g.fpsLastTime).Seconds() >= 1.0 {
		g.fpsCurrent = float64(g.fpsFrameCount) / now.Sub(g.fpsLastTime).Seconds()
		g.fpsFrameCount = 0
		g.fpsLastTime = now
	}

	// Initialize billboards lazily
	if !g.billboardsInitialized {
		g.initializeBillboards()
		g.billboardsInitialized = true
	}

	// Update simulation time for orbits
	g.simTime += dt

	// Camera movement with WASD
	moveSpeed := g.cameraSpeed * dt
	if ebiten.IsKeyPressed(ebiten.KeyShift) {
		moveSpeed *= 3.0
	}

	// C: Toggle cruise mode
	if inpututil.IsKeyJustPressed(ebiten.KeyC) {
		g.cruiseMode = !g.cruiseMode
		if g.cruiseMode {
			log.Printf("Cruise mode ENABLED - flying toward Sun")
		} else {
			log.Printf("Cruise mode DISABLED")
		}
	}

	if g.cruiseMode {
		// Auto-fly toward center (sun)
		dx := 0.0 - g.state.CameraX
		dy := 50.0 - g.state.CameraY
		dz := 0.0 - g.state.CameraZ
		dist := math.Sqrt(dx*dx + dy*dy + dz*dz)
		if dist > 50 {
			speed := 30.0 * dt
			g.state.CameraX += dx / dist * speed
			g.state.CameraY += dy / dist * speed
			g.state.CameraZ += dz / dist * speed
		}
	} else {
		if ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyUp) {
			g.state.CameraZ -= moveSpeed * 2
		}
		if ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyDown) {
			g.state.CameraZ += moveSpeed * 2
		}
		if ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyLeft) {
			g.state.CameraX -= moveSpeed
		}
		if ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyRight) {
			g.state.CameraX += moveSpeed
		}
		if ebiten.IsKeyPressed(ebiten.KeyQ) {
			g.state.CameraY += moveSpeed
		}
		if ebiten.IsKeyPressed(ebiten.KeyE) {
			g.state.CameraY -= moveSpeed
		}
	}

	// R: Reset camera position
	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		g.state.CameraX = 300.0
		g.state.CameraY = 100.0
		g.state.CameraZ = 200.0
		g.cruiseMode = false
		log.Printf("Camera reset to initial position")
	}

	// SR/GR effect controls
	if inpututil.IsKeyJustPressed(ebiten.Key1) {
		g.srWarp.Toggle()
		if g.srWarp.IsEnabled() {
			log.Printf("SR Warp ENABLED")
		} else {
			log.Printf("SR Warp DISABLED")
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.Key2) {
		g.grWarp.Toggle()
		if g.grWarp.IsEnabled() {
			log.Printf("GR Warp ENABLED")
		} else {
			log.Printf("GR Warp DISABLED")
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.Key3) && g.grWarp.IsEnabled() {
		intensity := g.grWarp.CycleDemoIntensity()
		log.Printf("GR intensity: %s", intensity)
	}

	// Velocity controls
	if ebiten.IsKeyPressed(ebiten.KeyEqual) || ebiten.IsKeyPressed(ebiten.KeyKPAdd) {
		g.velocity += 0.005
		if g.velocity > 0.99 {
			g.velocity = 0.99
		}
		g.srWarp.SetForwardVelocity(g.velocity)
		g.state.ShipVelocity = g.velocity
	}

	if ebiten.IsKeyPressed(ebiten.KeyMinus) || ebiten.IsKeyPressed(ebiten.KeyKPSubtract) {
		g.velocity -= 0.005
		if g.velocity < 0 {
			g.velocity = 0
		}
		g.srWarp.SetForwardVelocity(g.velocity)
		g.state.ShipVelocity = g.velocity
	}

	if inpututil.IsKeyJustPressed(ebiten.Key0) {
		g.velocity = 0
		g.srWarp.SetForwardVelocity(0)
		g.state.ShipVelocity = 0
	}

	// Lighting controls
	if ebiten.IsKeyPressed(ebiten.KeyLeftBracket) {
		g.lightMultiplier -= 0.02
		if g.lightMultiplier < 0.1 {
			g.lightMultiplier = 0.1
		}
		g.sunLight.SetEnergy(g.state.SunEnergy * g.lightMultiplier)
	}

	if ebiten.IsKeyPressed(ebiten.KeyRightBracket) {
		g.lightMultiplier += 0.02
		if g.lightMultiplier > 3.0 {
			g.lightMultiplier = 3.0
		}
		g.sunLight.SetEnergy(g.state.SunEnergy * g.lightMultiplier)
	}

	if ebiten.IsKeyPressed(ebiten.KeySemicolon) {
		g.ambientLevel -= 0.02
		if g.ambientLevel < 0.0 {
			g.ambientLevel = 0.0
		}
		g.ambientLight.SetEnergy(g.ambientLevel)
	}

	if ebiten.IsKeyPressed(ebiten.KeyApostrophe) {
		g.ambientLevel += 0.02
		if g.ambientLevel > 2.0 {
			g.ambientLevel = 2.0
		}
		g.ambientLight.SetEnergy(g.ambientLevel)
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyL) {
		g.lightMultiplier = 1.0
		g.ambientLevel = 1.0
		g.sunLight.SetEnergy(g.state.SunEnergy)
		g.ambientLight.SetEnergy(1.0)
	}

	// Update ALL body positions based on AILANG orbital data (hierarchical orbits)
	ailangBodies := sim_gen.GetAllSolarSystemBodies()

	// First pass: calculate positions for all primary bodies (planets, dwarf planets, asteroids)
	// These orbit the Sun (parentName == "")
	bodyPositions := make(map[string]lod.Vector3)
	for i, ap := range ailangBodies {
		if i >= len(g.planets) {
			break
		}
		if ap.ParentName == "" {
			// Primary body - orbits the Sun at origin
			phase := ap.OrbitPhase + g.simTime*ap.OrbitSpeed
			var posX, posZ float64
			if ap.OrbitRadius > 0 {
				posX = ap.OrbitRadius * math.Cos(phase)
				posZ = ap.OrbitRadius * math.Sin(phase)
			}
			bodyPositions[ap.Name] = lod.Vector3{X: posX, Y: 0, Z: posZ}
		}
	}

	// Second pass: calculate positions for moons (hierarchical orbits)
	// Moons orbit their parent body
	for i, ap := range ailangBodies {
		if i >= len(g.planets) {
			break
		}
		if ap.ParentName != "" {
			// Moon - orbits its parent body
			parentPos, ok := bodyPositions[ap.ParentName]
			if !ok {
				// Parent not found, fall back to Sun orbit
				phase := ap.OrbitPhase + g.simTime*ap.OrbitSpeed
				var posX, posZ float64
				if ap.OrbitRadius > 0 {
					posX = ap.OrbitRadius * math.Cos(phase)
					posZ = ap.OrbitRadius * math.Sin(phase)
				}
				bodyPositions[ap.Name] = lod.Vector3{X: posX, Y: 0, Z: posZ}
			} else {
				// Calculate moon position relative to parent
				phase := ap.OrbitPhase + g.simTime*ap.OrbitSpeed
				offsetX := ap.OrbitRadius * math.Cos(phase)
				offsetZ := ap.OrbitRadius * math.Sin(phase)
				posX := parentPos.X + offsetX
				posZ := parentPos.Z + offsetZ
				bodyPositions[ap.Name] = lod.Vector3{X: posX, Y: 0, Z: posZ}
			}
		}
	}

	// Third pass: update visual positions for all bodies
	for i, ap := range ailangBodies {
		if i >= len(g.planets) {
			break
		}
		p := g.planets[i]
		pos := bodyPositions[ap.Name]

		// Update LOD object position
		p.lodObj.Position = pos

		// Update 3D planet position
		p.planet.SetPosition(pos.X, 0, pos.Z)

		// Update rings if present
		if p.rings != nil {
			p.rings.SetPosition(pos.X, 0, pos.Z)
		}
	}

	// Update LOD camera
	g.lodCamera.Pos = lod.Vector3{X: g.state.CameraX, Y: g.state.CameraY, Z: g.state.CameraZ}
	g.lodCamera.LookAt = lod.Vector3{X: g.state.LookAtX, Y: g.state.LookAtY, Z: g.state.LookAtZ}

	// Update LOD manager
	g.lodManager.UpdateWithDT(g.lodCamera, dt)

	// Update planet visibility based on LOD tier
	for _, p := range g.planets {
		if p.lodObj.CurrentTier == lod.TierFull3D {
			p.planet.Model().SetVisible(true, true)
		} else {
			p.planet.Model().SetVisible(false, true)
		}
		p.planet.Update(dt)
	}

	// Update 3D scene camera
	g.scene3D.SetCameraPosition(g.state.CameraX, g.state.CameraY, g.state.CameraZ)
	g.scene3D.LookAt(g.state.LookAtX, g.state.LookAtY, g.state.LookAtZ)

	// Handle screenshot
	if g.screenshotTaken {
		return ebiten.Termination
	}

	return nil
}

// Draw implements ebiten.Game interface.
func (g *Game) Draw(screen *ebiten.Image) {
	// Determine render target
	renderTarget := screen
	useShaders := g.srWarp.IsEnabled() || g.grWarp.IsEnabled()
	if useShaders {
		renderTarget = g.renderBuffer
		g.renderBuffer.Clear()
	}

	// Draw space background
	renderTarget.Fill(color.RGBA{5, 5, 15, 255})

	// Get LOD tier objects
	points := g.lodManager.GetTierPoint()
	circles := g.lodManager.GetTierCircle()
	billboards := g.lodManager.GetTierBillboard()
	full3D := g.lodManager.GetTier3D()
	transitioning := g.lodManager.GetTransitioning()

	// Layer 1: Render points (distant objects)
	config := g.lodManager.Config()
	g.pointRenderer.RenderPointsScaled(renderTarget, points, config.CirclePixels)

	// Layer 2: Render circles
	g.circleRenderer.RenderCircles(renderTarget, circles)

	// Layer 3: Render billboards
	g.billboardRenderer.RenderBillboards(renderTarget, billboards, g.planetSprites)

	// Layer 4: Render 3D scene (Full3D tier)
	if len(full3D) > 0 {
		img3d := g.scene3D.Render()
		renderTarget.DrawImage(img3d, nil)
	}

	// Layer 5: Render transitioning objects
	for _, obj := range transitioning {
		prevAlpha := obj.PreviousAlpha()
		switch obj.PreviousTier {
		case lod.TierPoint:
			g.pointRenderer.RenderPointWithAlpha(renderTarget, obj, prevAlpha)
		case lod.TierCircle:
			g.circleRenderer.RenderCircleWithAlpha(renderTarget, obj, prevAlpha)
		case lod.TierBillboard:
			g.billboardRenderer.RenderBillboardWithAlpha(renderTarget, obj, prevAlpha, g.planetSprites)
		}
	}

	// Apply shader effects
	if useShaders {
		src := g.renderBuffer
		if g.srWarp.IsEnabled() && g.velocity >= 0.05 {
			intermediate := ebiten.NewImage(screenWidth, screenHeight)
			applied := g.srWarp.Apply(intermediate, src)
			if applied {
				src = intermediate
			}
			if g.frameCount%60 == 0 {
				log.Printf("SR Apply: velocity=%.3f, applied=%v", g.velocity, applied)
			}
		}
		if g.grWarp.IsEnabled() {
			intermediate := ebiten.NewImage(screenWidth, screenHeight)
			applied := g.grWarp.Apply(intermediate, src)
			if applied {
				src = intermediate
			}
			if g.frameCount%60 == 0 {
				log.Printf("GR Apply: demoMode=%v, applied=%v", g.grWarp.IsDemoMode(), applied)
			}
		}
		screen.DrawImage(src, nil)
	}

	// Draw overlay UI
	g.drawOverlay(screen)

	// Take screenshot if requested
	if g.screenshotFrame > 0 && g.frameCount >= g.screenshotFrame && !g.screenshotTaken {
		g.saveScreenshot(screen)
		g.screenshotTaken = true
	}
}

// drawOverlay renders the debug overlay UI.
func (g *Game) drawOverlay(screen *ebiten.Image) {
	stats := g.lodManager.Stats()

	// SR/GR status
	srStatus := "OFF"
	grStatus := "OFF"
	if g.srWarp.IsEnabled() {
		srStatus = "ON"
	}
	if g.grWarp.IsEnabled() {
		grStatus = fmt.Sprintf("ON (%s)", g.grWarp.GetDemoIntensity())
	}

	// Cruise status
	cruiseStatus := "OFF"
	if g.cruiseMode {
		cruiseStatus = "ON (toward Sun)"
	}

	// Calculate gamma
	gamma := 1.0
	if g.velocity > 0 {
		gamma = 1.0 / math.Sqrt(1.0-g.velocity*g.velocity)
	}

	// Total objects count
	totalObjects := stats.Full3DCount + stats.BillboardCount + stats.CircleCount + stats.PointCount

	info := fmt.Sprintf(`AILANG Solar Demo (LOD)
FPS: %.1f | Objects: %d
Frame: %d
Camera: (%.0f, %.0f, %.0f)

LOD Tiers:
  Full3D:    %d
  Billboard: %d
  Circle:    %d
  Point:     %d
  Culled:    %d
  Trans:     %d

Lighting:
  Sun: %.0f (x%.1f)
  Ambient: %.2f

Relativistic:
  Velocity: %.1f%% c
  Gamma:    %.2f
  SR: %s | GR: %s

Cruise: %s

Controls:
  WASD: Move | Q/E: Up/Down
  Shift: Fast | R: Reset
  C: Cruise | [ ]: Light
  ; ': Ambient | L: Reset
  1/2: SR/GR | 3: Cycle GR
  +/-/0: Velocity`,
		g.fpsCurrent, totalObjects,
		g.frameCount,
		g.state.CameraX, g.state.CameraY, g.state.CameraZ,
		stats.Full3DCount,
		stats.BillboardCount,
		stats.CircleCount,
		stats.PointCount,
		stats.CulledCount,
		len(g.lodManager.GetTransitioning()),
		g.state.SunEnergy, g.lightMultiplier,
		g.ambientLevel,
		g.velocity*100,
		gamma,
		srStatus, grStatus,
		cruiseStatus,
	)

	ebitenutil.DebugPrint(screen, info)
}

// saveScreenshot saves the current frame to a PNG file.
func (g *Game) saveScreenshot(screen *ebiten.Image) {
	path := g.screenshotPath
	if path == "" {
		path = "out/screenshots/ailang-solar.png"
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
}

// Layout implements ebiten.Game interface.
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

// rgbaFromInt converts a packed RGBA int to color.RGBA.
func rgbaFromInt(rgba int64) color.RGBA {
	return color.RGBA{
		R: uint8((rgba >> 24) & 0xFF),
		G: uint8((rgba >> 16) & 0xFF),
		B: uint8((rgba >> 8) & 0xFF),
		A: uint8(rgba & 0xFF),
	}
}

func main() {
	screenshotFrame := flag.Int("screenshot", 0, "Take screenshot at frame N and exit")
	screenshotPath := flag.String("output", "", "Screenshot output path")
	testGR := flag.Bool("test-gr", false, "Enable GR effect for testing")
	testSR := flag.Bool("test-sr", false, "Enable SR effect for testing (also sets velocity to 0.3c)")
	flag.Parse()

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("AILANG Solar System Demo (LOD)")

	game := NewGame(*screenshotFrame, *screenshotPath)

	// Enable shader effects for testing if requested
	if *testGR {
		game.grWarp.SetEnabled(true)
		log.Printf("GR effect enabled for testing (demoMode=%v)", game.grWarp.IsDemoMode())
	}
	if *testSR {
		game.srWarp.SetEnabled(true)
		game.velocity = 0.3 // 30% speed of light
		game.srWarp.SetForwardVelocity(0.3)
		log.Printf("SR effect enabled for testing (velocity=0.3c)")
	}

	if err := ebiten.RunGame(game); err != nil && err != ebiten.Termination {
		log.Fatal(err)
	}
}
