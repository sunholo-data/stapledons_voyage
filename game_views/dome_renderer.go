// Package game_views contains game-specific rendering helpers for Stapledon's Voyage.
// These reference game concepts (decks, planets, ship structure) and should not be in engine/.
package game_views

import (
	"image"
	"image/color"
	_ "image/jpeg"
	"log"
	"math"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"stapledons_voyage/engine/tetra"
	"stapledons_voyage/engine/view"
	"stapledons_voyage/engine/view/background"
	"stapledons_voyage/sim_gen"
)

// DomeRenderer renders the observation dome viewport on the bridge.
// It composites a space view through an elliptical mask.
type DomeRenderer struct {
	// Space background to render
	spaceBackground *background.SpaceBackground

	// Planet layer for 3D planets
	planetLayer *view.PlanetLayer
	planets     []*tetra.Planet
	rings       []*tetra.Ring

	// Cruise state (slow flyby through solar system)
	cruiseTime     float64 // Time elapsed
	cruiseVelocity float64 // Current velocity (fraction of c)
	cruiseDuration float64 // Total cruise duration

	// Offscreen buffers
	spaceBuffer *ebiten.Image // Full space render
	maskBuffer  *ebiten.Image // Elliptical mask

	// Dome configuration
	config DomeConfig

	// Cached mask image
	maskImage *ebiten.Image
}

// DomeConfig defines the dome viewport dimensions and position.
type DomeConfig struct {
	// Position of dome center in screen coordinates
	CenterX float64
	CenterY float64

	// Dome dimensions (ellipse radii)
	RadiusX float64
	RadiusY float64

	// Width and height of the render buffer
	BufferWidth  int
	BufferHeight int
}

// DefaultDomeConfig returns the default dome configuration for the bridge.
// The dome is FULL SCREEN - this is a bubble ship where the bridge
// floats inside a transparent observation dome looking out at space.
func DefaultDomeConfig() DomeConfig {
	// Full screen dome - space is the background, bridge floats inside
	return DomeConfig{
		CenterX:      640,  // Center of 1280-wide screen
		CenterY:      480,  // Center of 960-high screen
		RadiusX:      640,  // Full width
		RadiusY:      480,  // Full height
		BufferWidth:  1280, // Full screen buffer
		BufferHeight: 960,
	}
}

// NewDomeRenderer creates a new dome renderer with the given configuration.
func NewDomeRenderer(config DomeConfig) *DomeRenderer {
	d := &DomeRenderer{
		config:         config,
		spaceBuffer:    ebiten.NewImage(config.BufferWidth, config.BufferHeight),
		maskBuffer:     ebiten.NewImage(config.BufferWidth, config.BufferHeight),
		cruiseVelocity: 0.15, // Slow cruise at 15% c
		cruiseDuration: 60.0, // 1 minute loop for demo
	}

	// Create space background for the dome
	d.spaceBackground = background.NewSpaceBackground(config.BufferWidth, config.BufferHeight)

	// Load galaxy background if available
	if galaxyImg := loadTexture("assets/data/starmap/background/galaxy_4k.jpg"); galaxyImg != nil {
		d.spaceBackground.SetGalaxyImage(galaxyImg)
		log.Println("Loaded galaxy background for dome")
	}

	// Create planet layer for 3D planet rendering
	d.planetLayer = view.NewPlanetLayer(config.BufferWidth, config.BufferHeight)

	// Create the solar system
	d.createSolarSystem()

	// Generate the elliptical mask
	d.generateMask()

	return d
}

// loadTexture loads an image file as an Ebiten image.
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

// createSolarSystem sets up planets for the cruise view.
// DEPRECATED: Planet data is now defined in AILANG (sim/celestial.ail).
// Planets are rendered via CircleRGBA DrawCmds from renderSolarSystem().
// This Tetra3D code is kept for potential future 3D upgrade but is not used.
func (d *DomeRenderer) createSolarSystem() {
	// Planet configurations for cruise view
	// Slower, more scenic - we're cruising, not racing
	planetConfigs := []struct {
		name    string
		color   color.RGBA
		radius  float64
		dist    float64
		texture string
	}{
		// Start with outer planets, cruise inward
		{"neptune", color.RGBA{80, 120, 200, 255}, 1.0, 15, "assets/planets/neptune.jpg"},
		{"saturn", color.RGBA{210, 190, 150, 255}, 1.8, 50, "assets/planets/saturn.jpg"},
		{"jupiter", color.RGBA{220, 180, 140, 255}, 2.2, 90, "assets/planets/jupiter.jpg"},
		{"mars", color.RGBA{200, 100, 80, 255}, 0.5, 130, "assets/planets/mars.jpg"},
		{"earth", color.RGBA{60, 120, 200, 255}, 0.7, 150, "assets/planets/earth_daymap.jpg"},
	}

	// Load ring texture for Saturn
	ringTex := loadTexture("assets/planets/saturn_ring_gen.png")
	if ringTex == nil {
		ringTex = loadTexture("assets/planets/saturn_ring.png")
	}

	for _, cfg := range planetConfigs {
		var planet *tetra.Planet

		// Try to load texture, fall back to solid color
		if cfg.texture != "" {
			tex := loadTexture(cfg.texture)
			if tex != nil {
				planet = d.planetLayer.AddTexturedPlanet(cfg.name, cfg.radius, tex)
				log.Printf("Dome: Loaded texture for %s", cfg.name)
			} else {
				planet = d.planetLayer.AddPlanet(cfg.name, cfg.radius, cfg.color)
			}
		} else {
			planet = d.planetLayer.AddPlanet(cfg.name, cfg.radius, cfg.color)
		}

		// Position planet ABOVE the camera path (Y offset)
		// Planets appear in upper portion of screen as we pass by
		yOffset := cfg.dist * 0.15 // Planets above us
		planet.SetPosition(0, yOffset, -cfg.dist)

		// Slow rotation for scenic cruise
		planet.SetRotationSpeed(0.1)

		d.planets = append(d.planets, planet)

		// Add Saturn's rings
		if cfg.name == "saturn" && ringTex != nil {
			innerR := cfg.radius * 1.2
			outerR := cfg.radius * 2.3
			ring := d.planetLayer.AddRing("saturn_ring", innerR, outerR, ringTex)
			ring.SetPosition(0, yOffset, -cfg.dist)
			ring.SetTilt(0.47)
			d.rings = append(d.rings, ring)
			log.Printf("Dome: Added Saturn's rings")
		}
	}

	// Set initial camera position (ahead of first planet)
	d.planetLayer.SetCameraPosition(0, 2, 10)

	// Sun position - use same as working demo-solar-system
	// NOTE: Don't use SetSunTarget/LookAt - it breaks the directional light
	d.planetLayer.SetSunPosition(5, 3, 15)
}

// Update updates the dome animations (planet rotations, bubble arc).
// Camera position is now controlled by AILANG via SetCameraFromState.
func (d *DomeRenderer) Update(dt float64) {
	// Update planet animations (rotation only)
	d.planetLayer.Update(dt)

	// Update velocity for starfield parallax
	d.spaceBackground.SetVelocity(d.cruiseVelocity)
}

// SetCameraFromState updates the camera position from AILANG's DomeState.
// This allows AILANG to control the cruise animation while Go handles rendering.
func (d *DomeRenderer) SetCameraFromState(cameraZ float64, velocity float64) {
	// Camera at Y=0 to match working demo-solar-system
	// Planets have positive Y offsets so they appear in upper half of screen
	d.planetLayer.SetCameraPosition(0, 0, cameraZ)
	d.cruiseVelocity = velocity
	// Update cruiseTime for HUD (estimate from cameraZ position)
	// Camera goes from Z=10 to Z=-155, so progress = (10 - cameraZ) / 165
	if cameraZ <= 10.0 {
		progress := (10.0 - cameraZ) / 165.0
		d.cruiseTime = progress * d.cruiseDuration
	}
	// Debug: uncomment to verify this is being called
	// log.Printf("SetCameraFromState: cameraZ=%.2f, velocity=%.2f", cameraZ, velocity)
}

// GetCruiseInfo returns current cruise state for HUD display.
func (d *DomeRenderer) GetCruiseInfo() (velocity float64, progress float64) {
	return d.cruiseVelocity, d.cruiseTime / d.cruiseDuration
}

// generateMask creates an elliptical mask image.
func (d *DomeRenderer) generateMask() {
	w := d.config.BufferWidth
	h := d.config.BufferHeight

	d.maskImage = ebiten.NewImage(w, h)

	// Center of the buffer
	cx := float64(w) / 2
	cy := float64(h) / 2

	// Scale to match dome radii within buffer
	rx := float64(w) / 2 * 0.95 // Slight margin
	ry := float64(h) / 2 * 0.90

	// Draw ellipse mask - white inside, transparent outside
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			// Normalized distance from center (ellipse equation)
			dx := (float64(x) - cx) / rx
			dy := (float64(y) - cy) / ry
			dist := dx*dx + dy*dy

			if dist <= 1.0 {
				// Inside ellipse - fade at edges for soft boundary
				alpha := uint8(255)
				if dist > 0.85 {
					// Soft edge fade
					fade := 1.0 - (dist-0.85)/0.15
					alpha = uint8(255 * fade)
				}
				d.maskImage.Set(x, y, color.RGBA{255, 255, 255, alpha})
			}
		}
	}
}

// UpdateFromState updates the dome view based on AILANG DomeViewState.
func (d *DomeRenderer) UpdateFromState(state *sim_gen.DomeViewState) {
	if state == nil || d.spaceBackground == nil {
		return
	}

	// Update velocity for SR effects
	d.spaceBackground.SetVelocity(state.ShipVelocity)

	// Could add GR effects based on proximity to massive objects
	// d.spaceBackground.SetGRIntensity(...)
}

// Draw renders the dome viewport to the given screen at the configured position.
// This is the full composite render (background + planets + bubble arc).
// For layered rendering with floor tiles between, use DrawBackground, DrawPlanets, DrawBubbleArc.
func (d *DomeRenderer) Draw(screen *ebiten.Image) {
	d.DrawBackground(screen)
	d.DrawPlanets(screen)
	d.DrawBubbleArc(screen)
}

// DrawBackground renders just the starfield/galaxy background.
// Call this FIRST, before floor tiles.
func (d *DomeRenderer) DrawBackground(screen *ebiten.Image) {
	if d.spaceBackground == nil {
		return
	}
	d.spaceBackground.Draw(screen, nil)
}

// DrawPlanets renders the 3D textured planets.
// Call this AFTER floor tiles so planets appear in front.
func (d *DomeRenderer) DrawPlanets(screen *ebiten.Image) {
	if d.planetLayer != nil {
		d.planetLayer.Draw(screen)
	}
}

// DrawBubbleArc renders the bubble arc edge with plasma effect.
// NOTE: BubbleArc rendering moved to AILANG - this is now a placeholder.
// The actual bubble arc visual can be rendered via AILANG DrawCmds if desired.
func (d *DomeRenderer) DrawBubbleArc(screen *ebiten.Image) {
	// No-op: bubble arc visual rendering is now controlled by AILANG
	// AILANG can emit CircleRGBA or custom DrawCmds for the arc effect
}

// applyMask applies the elliptical mask to the space buffer.
func (d *DomeRenderer) applyMask() {
	// Use destination-in blend to keep only the masked area
	// First, composite the mask using the source alpha

	// Create a temporary buffer for masked result
	w := d.config.BufferWidth
	h := d.config.BufferHeight

	// We need to multiply the space buffer by the mask alpha
	// Using ebiten's composite operations

	// Clear mask buffer
	d.maskBuffer.Clear()

	// Draw space buffer to mask buffer
	d.maskBuffer.DrawImage(d.spaceBuffer, nil)

	// Apply mask - multiply alpha
	op := &ebiten.DrawImageOptions{}
	op.Blend = ebiten.BlendDestinationIn
	d.maskBuffer.DrawImage(d.maskImage, op)

	// Copy back to space buffer
	d.spaceBuffer.Clear()
	d.spaceBuffer.DrawImage(d.maskBuffer, nil)

	_ = w
	_ = h
}

// DrawDomeOutline draws a decorative frame around the dome (optional).
func (d *DomeRenderer) DrawDomeOutline(screen *ebiten.Image) {
	// Draw elliptical frame around dome
	cx := d.config.CenterX
	cy := d.config.CenterY
	rx := d.config.RadiusX
	ry := d.config.RadiusY

	// Draw points along the ellipse
	frameColor := color.RGBA{80, 90, 100, 255} // Dark metallic
	highlightColor := color.RGBA{120, 130, 140, 255}

	steps := 120
	for i := 0; i < steps; i++ {
		angle := float64(i) * 2 * math.Pi / float64(steps)
		x := cx + rx*math.Cos(angle)
		y := cy + ry*math.Sin(angle)

		// Draw outer frame (2 pixel thick)
		screen.Set(int(x), int(y), frameColor)
		screen.Set(int(x+1), int(y), frameColor)
		screen.Set(int(x), int(y+1), frameColor)

		// Highlight on top edge
		if math.Sin(angle) < -0.5 {
			screen.Set(int(x), int(y-1), highlightColor)
		}
	}
}

// SetGalaxyImage sets a galaxy background image for the dome view.
func (d *DomeRenderer) SetGalaxyImage(img *ebiten.Image) {
	if d.spaceBackground != nil {
		d.spaceBackground.SetGalaxyImage(img)
	}
}

// SetVelocity sets the ship velocity for SR visual effects.
func (d *DomeRenderer) SetVelocity(velocity float64) {
	if d.spaceBackground != nil {
		d.spaceBackground.SetVelocity(velocity)
	}
}

// Resize updates the dome for new screen dimensions.
func (d *DomeRenderer) Resize(screenW, screenH int) {
	// Recenter dome based on new screen size
	d.config.CenterX = float64(screenW) / 2
	// Keep Y position relative to top
}
