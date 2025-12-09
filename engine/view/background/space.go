// Package background provides background layer implementations for views.
package background

import (
	"image/color"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
)

// CameraOffset provides camera position for parallax calculations.
// This is a simple struct to avoid import cycles with the view package.
type CameraOffset struct {
	X, Y float64
	Zoom float64
}

// NewCameraOffset creates a camera at the origin with default zoom.
func NewCameraOffset() *CameraOffset {
	return &CameraOffset{
		X:    0,
		Y:    0,
		Zoom: 1.0,
	}
}

// SpaceBackground renders a parallax starfield with support for SR/GR effects.
// Implements the BackgroundLayer interface from the view package.
type SpaceBackground struct {
	starLayers  []*StarLayer
	camera      *CameraOffset
	buffer      *ebiten.Image
	screenW     int
	screenH     int

	// Optional galaxy background image
	galaxyImage *ebiten.Image

	// Velocity for SR effects (fraction of c)
	velocity float64

	// GR intensity for lensing effects
	grIntensity float64

	// Parallax multiplier (0=none, 1=full)
	parallaxDepth float64
}

// NewSpaceBackground creates a space background with default star layers.
func NewSpaceBackground(screenW, screenH int) *SpaceBackground {
	bg := &SpaceBackground{
		screenW:       screenW,
		screenH:       screenH,
		parallaxDepth: 1.0,
		camera:        NewCameraOffset(),
		buffer:        ebiten.NewImage(screenW, screenH),
	}

	// Create default star layers (far, mid, near)
	bg.initDefaultLayers()

	return bg
}

// initDefaultLayers creates the standard 3-layer starfield.
func (bg *SpaceBackground) initDefaultLayers() {
	// Far layer: many small, dim stars - no parallax
	far := NewStarLayer(StarLayerConfig{
		Count:        500,
		MinBrightness: 0.3,
		MaxBrightness: 0.6,
		MinSize:      1.0,
		MaxSize:      2.0,
		Parallax:     0.0, // Fixed background
		Seed:         42,
	}, bg.screenW, bg.screenH)

	// Mid layer: medium stars - slight parallax
	mid := NewStarLayer(StarLayerConfig{
		Count:        300,
		MinBrightness: 0.5,
		MaxBrightness: 0.8,
		MinSize:      1.5,
		MaxSize:      3.0,
		Parallax:     0.3,
		Seed:         123,
	}, bg.screenW, bg.screenH)

	// Near layer: few bright stars - more parallax
	near := NewStarLayer(StarLayerConfig{
		Count:        100,
		MinBrightness: 0.7,
		MaxBrightness: 1.0,
		MinSize:      2.0,
		MaxSize:      4.0,
		Parallax:     0.7,
		Seed:         456,
	}, bg.screenW, bg.screenH)

	bg.starLayers = []*StarLayer{far, mid, near}
}

// SetParallax sets the parallax depth multiplier.
func (bg *SpaceBackground) SetParallax(depth float64) {
	bg.parallaxDepth = depth
}

// SetGalaxyImage sets an optional galaxy background image.
// The image is drawn behind the procedural stars at low opacity.
func (bg *SpaceBackground) SetGalaxyImage(img *ebiten.Image) {
	bg.galaxyImage = img
}

// SetVelocity sets the ship velocity for SR effects.
func (bg *SpaceBackground) SetVelocity(v float64) {
	bg.velocity = v
}

// SetGRIntensity sets the GR intensity for lensing effects.
func (bg *SpaceBackground) SetGRIntensity(intensity float64) {
	bg.grIntensity = intensity
}

// GetVelocity returns the current velocity setting.
func (bg *SpaceBackground) GetVelocity() float64 {
	return bg.velocity
}

// GetGRIntensity returns the current GR intensity setting.
func (bg *SpaceBackground) GetGRIntensity() float64 {
	return bg.grIntensity
}

// Draw renders the background to the screen.
// The camera parameter can be nil to use the internal camera.
func (bg *SpaceBackground) Draw(screen *ebiten.Image, camera *CameraOffset) {
	// Clear with black
	screen.Fill(color.Black)

	// Use provided camera or fallback to internal
	cam := camera
	if cam == nil {
		cam = bg.camera
	}

	// Draw galaxy background if set (behind stars)
	if bg.galaxyImage != nil {
		op := &ebiten.DrawImageOptions{}
		// Scale to fit screen
		imgW := float64(bg.galaxyImage.Bounds().Dx())
		imgH := float64(bg.galaxyImage.Bounds().Dy())
		scaleX := float64(bg.screenW) / imgW
		scaleY := float64(bg.screenH) / imgH
		scale := scaleX
		if scaleY > scale {
			scale = scaleY
		}
		op.GeoM.Scale(scale, scale)
		// Center the image
		scaledW := imgW * scale
		scaledH := imgH * scale
		op.GeoM.Translate((float64(bg.screenW)-scaledW)/2, (float64(bg.screenH)-scaledH)/2)
		// Very slow parallax for distant galaxy
		op.GeoM.Translate(-cam.X*0.02, -cam.Y*0.02)
		// Dim the galaxy so stars are visible on top
		// Apply velocity-based color shift to galaxy (slight blue shift at high speed)
		r, g, b := 0.3, 0.3, 0.35
		if bg.velocity > 0.01 {
			// Blue shift forward (exaggerated for visual effect)
			blueShift := bg.velocity * 0.3
			r -= blueShift * 0.2
			b += blueShift * 0.15
		}
		op.ColorScale.Scale(float32(r), float32(g), float32(b), 1.0)
		screen.DrawImage(bg.galaxyImage, op)
	}

	// Draw each star layer with parallax and SR effects
	for _, layer := range bg.starLayers {
		// Calculate parallax offset based on camera position
		parallax := layer.config.Parallax * bg.parallaxDepth
		offsetX := -cam.X * parallax
		offsetY := -cam.Y * parallax

		layer.DrawWithSR(screen, offsetX, offsetY, bg.velocity, bg.screenW, bg.screenH)
	}
}

// AddStarLayer adds a custom star layer.
func (bg *SpaceBackground) AddStarLayer(layer *StarLayer) {
	bg.starLayers = append(bg.starLayers, layer)
}

// ClearLayers removes all star layers.
func (bg *SpaceBackground) ClearLayers() {
	bg.starLayers = nil
}

// Resize updates the background dimensions.
func (bg *SpaceBackground) Resize(screenW, screenH int) {
	if bg.screenW == screenW && bg.screenH == screenH {
		return
	}

	bg.screenW = screenW
	bg.screenH = screenH

	// Dispose and recreate buffer
	if bg.buffer != nil {
		bg.buffer.Dispose()
	}
	bg.buffer = ebiten.NewImage(screenW, screenH)

	// Regenerate star positions for new dimensions
	for _, layer := range bg.starLayers {
		layer.Regenerate(screenW, screenH)
	}
}

// Star represents a single star with position and properties.
type Star struct {
	X, Y       float64
	Brightness float64
	Size       float64
	Color      color.RGBA
}

// StarLayerConfig configures a star layer.
type StarLayerConfig struct {
	Count         int
	MinBrightness float64
	MaxBrightness float64
	MinSize       float64
	MaxSize       float64
	Parallax      float64 // 0 = fixed, 1 = full camera movement
	Seed          int64
}

// StarLayer manages a layer of stars at a specific depth.
type StarLayer struct {
	config  StarLayerConfig
	stars   []Star
	screenW int
	screenH int
	rng     *rand.Rand
}

// NewStarLayer creates a new star layer with the given configuration.
func NewStarLayer(config StarLayerConfig, screenW, screenH int) *StarLayer {
	layer := &StarLayer{
		config:  config,
		screenW: screenW,
		screenH: screenH,
		rng:     rand.New(rand.NewSource(config.Seed)),
	}
	layer.generateStars()
	return layer
}

// generateStars creates stars based on the configuration.
func (l *StarLayer) generateStars() {
	l.stars = make([]Star, l.config.Count)

	// Extend bounds to allow for parallax scrolling
	padding := 200.0 // Extra padding for parallax

	for i := range l.stars {
		// Random position with padding
		x := l.rng.Float64()*(float64(l.screenW)+padding*2) - padding
		y := l.rng.Float64()*(float64(l.screenH)+padding*2) - padding

		// Random brightness
		brightness := l.config.MinBrightness +
			l.rng.Float64()*(l.config.MaxBrightness-l.config.MinBrightness)

		// Random size
		size := l.config.MinSize +
			l.rng.Float64()*(l.config.MaxSize-l.config.MinSize)

		// Star color (slight variation from pure white)
		r := uint8(200 + l.rng.Intn(56))
		g := uint8(200 + l.rng.Intn(56))
		b := uint8(220 + l.rng.Intn(36))
		a := uint8(brightness * 255)

		l.stars[i] = Star{
			X:          x,
			Y:          y,
			Brightness: brightness,
			Size:       size,
			Color:      color.RGBA{r, g, b, a},
		}
	}
}

// Regenerate recreates stars for new screen dimensions.
func (l *StarLayer) Regenerate(screenW, screenH int) {
	l.screenW = screenW
	l.screenH = screenH
	l.rng = rand.New(rand.NewSource(l.config.Seed)) // Reset seed for consistency
	l.generateStars()
}

// Draw renders the star layer with the given offset.
func (l *StarLayer) Draw(screen *ebiten.Image, offsetX, offsetY float64) {
	for _, star := range l.stars {
		// Apply parallax offset
		x := star.X + offsetX
		y := star.Y + offsetY

		// Wrap around screen edges (seamless scrolling)
		padding := 200.0
		totalW := float64(l.screenW) + padding*2
		totalH := float64(l.screenH) + padding*2

		for x < -padding {
			x += totalW
		}
		for x > float64(l.screenW)+padding {
			x -= totalW
		}
		for y < -padding {
			y += totalH
		}
		for y > float64(l.screenH)+padding {
			y -= totalH
		}

		// Skip if outside visible area
		if x < -star.Size || x > float64(l.screenW)+star.Size ||
			y < -star.Size || y > float64(l.screenH)+star.Size {
			continue
		}

		// Draw the star
		drawStar(screen, x, y, star.Size, star.Color)
	}
}

// DrawWithSR renders the star layer with Special Relativity effects.
// Applies Doppler shift (blue toward center/forward, red at edges/behind),
// relativistic beaming, and forward motion parallax (stars expand from center).
func (l *StarLayer) DrawWithSR(screen *ebiten.Image, offsetX, offsetY, velocity float64, screenW, screenH int) {
	if velocity < 0.01 {
		// No SR effects at low velocity, use standard draw
		l.Draw(screen, offsetX, offsetY)
		return
	}

	// Direction of travel is upper-center
	// Stars expand outward from this point (forward motion effect)
	centerX := float64(screenW) * 0.5 // Center
	centerY := float64(screenH) * 0.3 // Upper area (direction of travel)

	for _, star := range l.stars {
		// Apply parallax offset
		x := star.X + offsetX
		y := star.Y + offsetY

		// Forward motion parallax: stars expand outward from center
		// This creates the illusion of moving toward the center point
		dx := x - centerX
		dy := y - centerY
		dist := dx*dx + dy*dy
		if dist > 0 {
			// Expand stars outward from center based on velocity and layer parallax
			expansionRate := velocity * l.config.Parallax * 0.3
			expandX := dx * expansionRate
			expandY := dy * expansionRate
			x += expandX
			y += expandY
		}

		// Wrap around screen edges (seamless scrolling)
		padding := 200.0
		totalW := float64(l.screenW) + padding*2
		totalH := float64(l.screenH) + padding*2

		for x < -padding {
			x += totalW
		}
		for x > float64(l.screenW)+padding {
			x -= totalW
		}
		for y < -padding {
			y += totalH
		}
		for y > float64(l.screenH)+padding {
			y -= totalH
		}

		// Skip if outside visible area
		if x < -star.Size || x > float64(screenW)+star.Size ||
			y < -star.Size || y > float64(screenH)+star.Size {
			continue
		}

		// Calculate angle from center (direction of travel)
		dx = x - centerX
		dy = y - centerY
		// Normalize to 0-1 where 0 = center (forward), 1 = edge (perpendicular)
		maxDist := centerX
		if centerY < maxDist {
			maxDist = centerY
		}
		dist = (dx*dx + dy*dy)
		normDist := dist / (maxDist * maxDist)
		if normDist > 1 {
			normDist = 1
		}

		// Apply Doppler shift based on position
		// Forward (center) = blue shift, sides = red shift
		// HEAVILY exaggerated for visual drama
		shiftIntensity := velocity * 8.0 * (1.0 - normDist) // Strong blue at center
		redShift := velocity * 4.0 * normDist              // Red at edges

		// Modify star color - blue shift center, red shift edges
		r := float64(star.Color.R) - shiftIntensity*100 + redShift*80
		g := float64(star.Color.G) + shiftIntensity*30 - redShift*20
		b := float64(star.Color.B) + shiftIntensity*100 - redShift*60

		// Clamp values
		if r < 0 {
			r = 0
		}
		if r > 255 {
			r = 255
		}
		if g < 0 {
			g = 0
		}
		if g > 255 {
			g = 255
		}
		if b < 0 {
			b = 0
		}
		if b > 255 {
			b = 255
		}

		// Relativistic beaming - stars toward center appear brighter
		brightnessBoost := 1.0 + velocity*3.0*(1.0-normDist)
		a := float64(star.Color.A) * brightnessBoost
		if a > 255 {
			a = 255
		}

		srColor := color.RGBA{uint8(r), uint8(g), uint8(b), uint8(a)}

		// Draw the star with SR-modified color
		drawStar(screen, x, y, star.Size, srColor)
	}
}

// drawStar renders a single star at the given position.
func drawStar(screen *ebiten.Image, x, y, size float64, c color.RGBA) {
	// For small stars, draw as a single pixel or small rect
	if size <= 1.5 {
		// Single pixel
		screen.Set(int(x), int(y), c)
	} else if size <= 2.5 {
		// 2x2 block
		ix, iy := int(x), int(y)
		screen.Set(ix, iy, c)
		screen.Set(ix+1, iy, c)
		screen.Set(ix, iy+1, c)
		screen.Set(ix+1, iy+1, c)
	} else {
		// Larger stars with glow
		// Draw core (bright center)
		ix, iy := int(x), int(y)
		screen.Set(ix, iy, c)

		// Draw surrounding pixels with reduced alpha for glow effect
		glowC := color.RGBA{c.R, c.G, c.B, c.A / 2}
		for dx := -1; dx <= 1; dx++ {
			for dy := -1; dy <= 1; dy++ {
				if dx == 0 && dy == 0 {
					continue
				}
				screen.Set(ix+dx, iy+dy, glowC)
			}
		}

		// Extra glow for very large stars
		if size > 3.5 {
			dimC := color.RGBA{c.R, c.G, c.B, c.A / 4}
			for _, offset := range []struct{ x, y int }{{-2, 0}, {2, 0}, {0, -2}, {0, 2}} {
				screen.Set(ix+offset.x, iy+offset.y, dimC)
			}
		}
	}
}
