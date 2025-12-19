// Package render provides rendering utilities.
package render

import (
	"encoding/json"
	"image/color"
	"log"
	"math"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"stapledons_voyage/engine/relativity"
)

// SpaceView provides a unified view of space from any position/direction.
// It loads the real star catalog once and renders view-specific textures.
// This is the "single source of truth" for what's outside the ship.
type SpaceView struct {
	stars      []starEntry
	loaded     bool
	viewCache  map[string]*ebiten.Image // Keyed by view params hash
	cacheLimit int
}

// starEntry represents a star from the catalog
type starEntry struct {
	ID       string
	Name     string
	X, Y, Z  float64 // Position in light-years (Sol-centered)
	DistLY   float64 // Distance from Sol
	VMag     float64 // Visual magnitude
	Spectral string  // Spectral type (O,B,A,F,G,K,M)
}

// starCatalogJSON matches the JSON file structure
type starCatalogJSON struct {
	Version string `json:"version"`
	Source  string `json:"source"`
	Count   int    `json:"count"`
	Stars   []struct {
		ID       string  `json:"id"`
		Name     string  `json:"name"`
		X        float64 `json:"x"`
		Y        float64 `json:"y"`
		Z        float64 `json:"z"`
		DistLY   float64 `json:"dist_ly"`
		VMag     float64 `json:"vmag"`
		Spectral string  `json:"spectral"`
	} `json:"stars"`
}

// NewSpaceView creates a new space view manager
func NewSpaceView() *SpaceView {
	return &SpaceView{
		viewCache:  make(map[string]*ebiten.Image),
		cacheLimit: 16, // Cache up to 16 different views
	}
}

// Load loads the star catalog from the given path
func (sv *SpaceView) Load(path string) error {
	if sv.loaded {
		return nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var catalog starCatalogJSON
	if err := json.Unmarshal(data, &catalog); err != nil {
		return err
	}

	sv.stars = make([]starEntry, len(catalog.Stars))
	for i, s := range catalog.Stars {
		sv.stars[i] = starEntry{
			ID:       s.ID,
			Name:     s.Name,
			X:        s.X,
			Y:        s.Y,
			Z:        s.Z,
			DistLY:   s.DistLY,
			VMag:     s.VMag,
			Spectral: s.Spectral,
		}
	}

	sv.loaded = true
	log.Printf("SpaceView: Loaded %d stars from %s", len(sv.stars), catalog.Source)
	return nil
}

// ViewParams describes the view parameters for rendering
type ViewParams struct {
	// Ship position in light-years (galactocentric, Sol at origin)
	ShipX, ShipY, ShipZ float64

	// View direction (normalized)
	ViewDirX, ViewDirY, ViewDirZ float64

	// Up vector (normalized)
	UpX, UpY, UpZ float64

	// Field of view in degrees
	FOV float64

	// Ship velocity (fraction of c, for SR effects)
	Velocity float64

	// Gravitational potential (dimensionless, for GR effects)
	// 0 = flat space, higher = stronger gravity
	// At phi=0.5, you're near a black hole event horizon
	GrPhi float64
}

// RenderView renders a starfield view with the given parameters.
// Returns a texture that can be applied to a window.
func (sv *SpaceView) RenderView(params ViewParams, width, height int) *ebiten.Image {
	if !sv.loaded || len(sv.stars) == 0 {
		// Return a simple procedural starfield if catalog not loaded
		return sv.renderProceduralStarfield(width, height)
	}

	// Create output image
	img := ebiten.NewImage(width, height)
	img.Fill(color.RGBA{5, 8, 15, 255}) // Deep space black

	// Normalize view direction
	viewLen := math.Sqrt(params.ViewDirX*params.ViewDirX +
		params.ViewDirY*params.ViewDirY + params.ViewDirZ*params.ViewDirZ)
	if viewLen < 0.001 {
		params.ViewDirX, params.ViewDirY, params.ViewDirZ = 0, 0, 1
		viewLen = 1
	}
	vx := params.ViewDirX / viewLen
	vy := params.ViewDirY / viewLen
	vz := params.ViewDirZ / viewLen

	// Normalize up vector
	upLen := math.Sqrt(params.UpX*params.UpX + params.UpY*params.UpY + params.UpZ*params.UpZ)
	if upLen < 0.001 {
		params.UpX, params.UpY, params.UpZ = 0, 1, 0
		upLen = 1
	}
	ux := params.UpX / upLen
	uy := params.UpY / upLen
	uz := params.UpZ / upLen

	// Calculate right vector (cross product of view and up)
	rx := vy*uz - vz*uy
	ry := vz*ux - vx*uz
	rz := vx*uy - vy*ux

	// Re-orthogonalize up vector
	ux = ry*vz - rz*vy
	uy = rz*vx - rx*vz
	uz = rx*vy - ry*vx

	// FOV calculations
	fov := params.FOV
	if fov <= 0 {
		fov = 90
	}
	halfFOV := fov * math.Pi / 360 // half angle in radians
	tanHalfFOV := math.Tan(halfFOV)

	centerX := float64(width) / 2
	centerY := float64(height) / 2
	focalLen := centerX / tanHalfFOV

	// Precompute relativistic parameters if moving
	// Beta points in direction of travel - use view direction so we're always
	// looking "forward" through the window (simulates traveling toward what we see)
	beta := relativity.Vec3{X: vx * params.Velocity, Y: vy * params.Velocity, Z: vz * params.Velocity}
	gamma := relativity.Gamma(params.Velocity)

	// For each star, project to screen space
	for _, star := range sv.stars {
		// Vector from ship to star (direction TO star)
		dx := star.X - params.ShipX
		dy := star.Y - params.ShipY
		dz := star.Z - params.ShipZ

		// Distance from ship
		dist := math.Sqrt(dx*dx + dy*dy + dz*dz)
		if dist < 0.001 {
			continue // Skip if at ship position
		}

		// Normalize direction to star
		dx /= dist
		dy /= dist
		dz /= dist

		// Direction vector in relativity package format
		starDir := relativity.Vec3{X: dx, Y: dy, Z: dz}

		// Apply relativistic aberration if moving at significant velocity
		// This causes stars to appear to bunch toward the direction of travel
		if params.Velocity > 0.01 {
			starDir = relativity.TransformDirection(starDir, beta, gamma)
			dx, dy, dz = starDir.X, starDir.Y, starDir.Z
		}

		// Project onto view plane
		// Dot product with view direction (forward distance)
		forward := dx*vx + dy*vy + dz*vz

		// Skip stars behind the camera
		if forward <= 0 {
			continue
		}

		// Dot product with right (x on screen)
		screenRight := dx*rx + dy*ry + dz*rz

		// Dot product with up (y on screen, inverted for screen coords)
		screenUp := dx*ux + dy*uy + dz*uz

		// Perspective projection
		screenX := centerX + (screenRight/forward)*focalLen
		screenY := centerY - (screenUp/forward)*focalLen // Invert Y

		// Apply GR gravitational lensing to star positions
		// Stars get pushed away from the center (where the massive object would be)
		// Deflection angle: α = 2*r_s/b in weak field
		if params.GrPhi > 0.05 {
			// Distance from screen center (impact parameter in screen space)
			offsetX := screenX - centerX
			offsetY := screenY - centerY
			b := math.Sqrt(offsetX*offsetX + offsetY*offsetY)

			if b > 1.0 { // Avoid division by zero at center
				// Schwarzschild radius scaled to screen (in pixels)
				// At phi = 0.5, we're at the event horizon
				// Scale Rs to be visible: Rs ∝ phi * screenScale
				screenScale := centerX / 4.0 // Effect visible in central quarter
				rs := params.GrPhi * screenScale * 2

				// Einstein deflection: α = 2*r_s/b
				deflection := 2.0 * rs / b

				// Clamp deflection to avoid extreme warping
				if deflection > b*0.3 {
					deflection = b * 0.3
				}

				// Push star outward from center
				if b > 0.001 {
					screenX += (offsetX / b) * deflection
					screenY += (offsetY / b) * deflection
				}
			}
		}

		// Skip stars outside the frame
		if screenX < 0 || screenX >= float64(width) || screenY < 0 || screenY >= float64(height) {
			continue
		}

		// Calculate apparent brightness based on magnitude and distance
		brightness := magnitudeToBrightness(star.VMag, dist)

		// Get star's base temperature from spectral type
		baseTemp := 5800.0 // Default to Sun-like
		if temp, ok := relativity.StarTemperature[star.Spectral]; ok {
			baseTemp = temp
		}

		// Apply SR effects using proper relativity package
		var starColor color.RGBA
		if params.Velocity > 0.01 {
			// Compute Doppler factor for this star direction
			// Note: n points FROM source TO observer, so negate starDir
			nToObserver := relativity.Vec3{X: -dx, Y: -dy, Z: -dz}
			dopplerFactor := relativity.DopplerFactor(beta, nToObserver, gamma)

			// Shift color temperature based on Doppler effect
			shiftedTemp := relativity.ShiftColorTemperature(baseTemp, dopplerFactor)
			shiftedRGB := relativity.TemperatureToRGB(shiftedTemp)

			// Apply relativistic beaming (D^3 intensity scaling)
			beamFactor := relativity.BeamBrightness(dopplerFactor)
			// Cap beaming to avoid blowout
			if beamFactor > 8 {
				beamFactor = 8
			}
			brightness *= beamFactor

			starColor = color.RGBA{shiftedRGB.R, shiftedRGB.G, shiftedRGB.B, 255}
		} else {
			// No relativistic effects - use base spectral color
			baseRGB := relativity.TemperatureToRGB(baseTemp)
			starColor = color.RGBA{baseRGB.R, baseRGB.G, baseRGB.B, 255}
		}

		// Apply GR gravitational redshift if in strong gravity field
		// Light climbing out of gravity well loses energy → redshift
		if params.GrPhi > 0.01 {
			// GR redshift factor: z = 1/sqrt(1 - 2*phi) - 1
			// For weak fields, z ≈ phi
			// Clamp to avoid singularity at event horizon (phi = 0.5)
			clampedPhi := params.GrPhi
			if clampedPhi > 0.45 {
				clampedPhi = 0.45
			}

			// Time dilation factor (light is dimmed by this factor^2)
			grTimeDilation := math.Sqrt(1.0 - 2.0*clampedPhi)

			// Redshift the color temperature (lower temp = redder)
			// Current color temp from RGB approximation
			currentTemp := baseTemp // Start with base (SR already applied to shiftedRGB)
			if params.Velocity > 0.01 {
				// SR already shifted, but we need current temp for GR
				// Approximate: use brightness-weighted approach
				currentTemp = baseTemp * (1.0 + 0.5*params.Velocity) // Rough approximation
			}

			// GR redshift factor on temperature
			grRedshiftZ := 1.0/grTimeDilation - 1.0
			grShiftedTemp := currentTemp / (1.0 + grRedshiftZ)

			// Clamp temperature
			if grShiftedTemp < 1000 {
				grShiftedTemp = 1000
			}
			if grShiftedTemp > 40000 {
				grShiftedTemp = 40000
			}

			grRGB := relativity.TemperatureToRGB(grShiftedTemp)
			starColor = color.RGBA{grRGB.R, grRGB.G, grRGB.B, 255}

			// Also dim by time dilation squared (photon rate reduction)
			brightness *= grTimeDilation * grTimeDilation
		}

		// Clamp brightness
		if brightness > 1.0 {
			brightness = 1.0
		}
		if brightness < 0.03 {
			continue // Skip very dim stars
		}

		// Scale color by brightness
		r := uint8(float64(starColor.R) * brightness)
		g := uint8(float64(starColor.G) * brightness)
		b := uint8(float64(starColor.B) * brightness)

		// Draw the star - use size based on distance
		ix, iy := int(screenX), int(screenY)

		// Calculate apparent size based on distance
		// Stars within ~0.5 ly should appear as noticeable disks
		// Typical stellar radius is ~1 solar radius = 0.005 AU = 7.3e-8 ly
		// But we'll use artistic license for visibility
		apparentRadius := 1.0
		if dist < 0.1 {
			// Very close - render as large sphere
			apparentRadius = 30.0 * (0.1 - dist) / 0.1
			if apparentRadius < 5 {
				apparentRadius = 5
			}
		} else if dist < 0.5 {
			// Close - render as medium disk
			apparentRadius = 10.0 * (0.5 - dist) / 0.5
			if apparentRadius < 2 {
				apparentRadius = 2
			}
		} else if dist < 2 {
			// Nearby - slightly bigger point
			apparentRadius = 3.0 * (2 - dist) / 2
			if apparentRadius < 1 {
				apparentRadius = 1
			}
		}

		// If star is a disk (close), render as filled circle with shading
		if apparentRadius > 2 {
			// Render as a shaded sphere
			rad := int(apparentRadius)
			for dy := -rad; dy <= rad; dy++ {
				for dx := -rad; dx <= rad; dx++ {
					d := math.Sqrt(float64(dx*dx + dy*dy))
					if d <= apparentRadius {
						px, py := ix+dx, iy+dy
						if px >= 0 && px < width && py >= 0 && py < height {
							// Limb darkening - edges are dimmer
							limbFactor := 1.0 - (d/apparentRadius)*0.5
							// Normal direction for shading
							nz := math.Sqrt(1 - (d/apparentRadius)*(d/apparentRadius))
							shade := 0.5 + 0.5*nz // Simple shading
							sf := limbFactor * shade * brightness

							sr := uint8(math.Min(255, float64(r)*sf*1.2))
							sg := uint8(math.Min(255, float64(g)*sf*1.2))
							sb := uint8(math.Min(255, float64(b)*sf*1.2))
							img.Set(px, py, color.RGBA{sr, sg, sb, 255})
						}
					}
				}
			}
		} else {
			// Single pixel star
			img.Set(ix, iy, color.RGBA{r, g, b, 255})

			// Add glow for bright stars - more dramatic at high velocities
			glowThreshold := 0.6
			if params.Velocity > 0.5 {
				glowThreshold = 0.4 // More stars get glow at high speed
			}

			if brightness > glowThreshold && ix > 1 && iy > 1 && ix < width-2 && iy < height-2 {
				glowR := r / 2
				glowG := g / 2
				glowB := b / 2

				// Inner glow
				img.Set(ix-1, iy, color.RGBA{glowR, glowG, glowB, 255})
				img.Set(ix+1, iy, color.RGBA{glowR, glowG, glowB, 255})
				img.Set(ix, iy-1, color.RGBA{glowR, glowG, glowB, 255})
				img.Set(ix, iy+1, color.RGBA{glowR, glowG, glowB, 255})

				// Extra bright stars get a larger glow
				if brightness > 0.85 {
					dimR := glowR / 2
					dimG := glowG / 2
					dimB := glowB / 2
					img.Set(ix-2, iy, color.RGBA{dimR, dimG, dimB, 255})
					img.Set(ix+2, iy, color.RGBA{dimR, dimG, dimB, 255})
					img.Set(ix, iy-2, color.RGBA{dimR, dimG, dimB, 255})
					img.Set(ix, iy+2, color.RGBA{dimR, dimG, dimB, 255})
					// Diagonal glow
					img.Set(ix-1, iy-1, color.RGBA{dimR, dimG, dimB, 255})
					img.Set(ix+1, iy-1, color.RGBA{dimR, dimG, dimB, 255})
					img.Set(ix-1, iy+1, color.RGBA{dimR, dimG, dimB, 255})
					img.Set(ix+1, iy+1, color.RGBA{dimR, dimG, dimB, 255})
				}
			}
		}
	}

	// Draw event horizon and photon sphere if grPhi is high
	if params.GrPhi > 0.1 {
		// Calculate event horizon radius in screen space
		// At phi = 0.5, we're at the horizon
		screenScale := centerX / 4.0
		eventHorizonRadius := params.GrPhi * screenScale * 2

		// Photon sphere at 1.5x event horizon
		photonSphereRadius := eventHorizonRadius * 1.5

		// Draw photon sphere ring (subtle orange glow)
		photonColor := color.RGBA{80, 40, 20, 255}
		for angle := 0.0; angle < 2*math.Pi; angle += 0.02 {
			for r := photonSphereRadius - 2; r <= photonSphereRadius+2; r += 1 {
				px := centerX + math.Cos(angle)*r
				py := centerY + math.Sin(angle)*r
				if px >= 0 && px < float64(width) && py >= 0 && py < float64(height) {
					img.Set(int(px), int(py), photonColor)
				}
			}
		}

		// Draw event horizon (black disk)
		cxi := int(centerX)
		cyi := int(centerY)
		maxR := int(eventHorizonRadius) + 1
		for dy := -maxR; dy <= maxR; dy++ {
			for dx := -maxR; dx <= maxR; dx++ {
				dist := math.Sqrt(float64(dx*dx + dy*dy))
				if dist <= eventHorizonRadius {
					px, py := cxi+dx, cyi+dy
					if px >= 0 && px < width && py >= 0 && py < height {
						img.Set(px, py, color.RGBA{0, 0, 0, 255})
					}
				}
			}
		}
	}

	return img
}

// magnitudeToBrightness converts visual magnitude to display brightness (0-1)
func magnitudeToBrightness(vmag, distFromShip float64) float64 {
	// Visual magnitude is apparent, so brighter = lower number
	// Sun is -26.7, Sirius is -1.46, limit of naked eye is ~6
	// We want to show stars down to about mag 8 (telescope range)

	// Adjust for distance from ship vs distance from Sol
	// This is a simplification - real apparent mag depends on absolute mag
	// For now, just use vmag directly with a scaling factor

	// Map magnitude range to brightness
	// -2 to 8 -> 1.0 to 0.1
	minMag := -2.0
	maxMag := 8.0
	if vmag < minMag {
		vmag = minMag
	}
	if vmag > maxMag {
		vmag = maxMag
	}

	// Linear interpolation (could use log scale for more realism)
	brightness := 1.0 - (vmag-minMag)/(maxMag-minMag)*0.9

	// Dim stars that are far from the ship (beyond ~30 ly)
	if distFromShip > 30 {
		brightness *= 30 / distFromShip
	}

	return brightness
}

// renderProceduralStarfield creates a simple procedural starfield
// Used as fallback when the real catalog isn't loaded
func (sv *SpaceView) renderProceduralStarfield(width, height int) *ebiten.Image {
	img := ebiten.NewImage(width, height)
	img.Fill(color.RGBA{5, 8, 15, 255})

	// Simple deterministic stars using position-based hash
	for y := 0; y < height; y += 4 {
		for x := 0; x < width; x += 4 {
			hash := (x*374761393 + y*668265263) & 0xFFFFFF
			if hash%100 < 2 { // 2% chance of star
				brightness := uint8(100 + (hash%156))
				img.Set(x, y, color.RGBA{brightness, brightness, brightness + 20, 255})
			}
		}
	}

	return img
}

// GetStarCount returns the number of loaded stars
func (sv *SpaceView) GetStarCount() int {
	return len(sv.stars)
}

// IsLoaded returns true if the star catalog has been loaded
func (sv *SpaceView) IsLoaded() bool {
	return sv.loaded
}

// RenderEquirectangular renders a full-sky equirectangular projection.
// This maps the entire celestial sphere to a 2:1 aspect ratio texture,
// suitable for dome/hemisphere rendering.
// Width should be 2x height for proper aspect ratio.
func (sv *SpaceView) RenderEquirectangular(params ViewParams, width, height int) *ebiten.Image {
	if !sv.loaded || len(sv.stars) == 0 {
		return sv.renderProceduralEquirectangular(width, height)
	}

	img := ebiten.NewImage(width, height)
	img.Fill(color.RGBA{5, 8, 15, 255})

	// Precompute relativistic parameters
	vx, vy, vz := params.ViewDirX, params.ViewDirY, params.ViewDirZ
	beta := relativity.Vec3{X: vx * params.Velocity, Y: vy * params.Velocity, Z: vz * params.Velocity}
	gamma := relativity.Gamma(params.Velocity)

	for _, star := range sv.stars {
		// Vector from ship to star
		dx := star.X - params.ShipX
		dy := star.Y - params.ShipY
		dz := star.Z - params.ShipZ

		dist := math.Sqrt(dx*dx + dy*dy + dz*dz)
		if dist < 0.001 {
			continue
		}

		// Normalize
		dx /= dist
		dy /= dist
		dz /= dist

		// Apply relativistic aberration
		starDir := relativity.Vec3{X: dx, Y: dy, Z: dz}
		if params.Velocity > 0.01 {
			starDir = relativity.TransformDirection(starDir, beta, gamma)
			dx, dy, dz = starDir.X, starDir.Y, starDir.Z
		}

		// Convert to spherical coordinates
		// theta = azimuth (0 to 2π), phi = elevation (-π/2 to π/2)
		theta := math.Atan2(dz, dx)        // Range: -π to π
		phi := math.Asin(dy)               // Range: -π/2 to π/2

		// Map to equirectangular coordinates
		// U = (theta + π) / 2π -> 0 to 1
		// V = (π/2 - phi) / π -> 0 to 1 (flip so north is top)
		u := (theta + math.Pi) / (2 * math.Pi)
		v := (math.Pi/2 - phi) / math.Pi

		screenX := u * float64(width)
		screenY := v * float64(height)

		// Skip stars outside frame
		if screenX < 0 || screenX >= float64(width) || screenY < 0 || screenY >= float64(height) {
			continue
		}

		// Calculate brightness
		brightness := magnitudeToBrightness(star.VMag, dist)

		// Get star color
		baseTemp := 5800.0
		if temp, ok := relativity.StarTemperature[star.Spectral]; ok {
			baseTemp = temp
		}

		var starColor color.RGBA
		if params.Velocity > 0.01 {
			nToObserver := relativity.Vec3{X: -dx, Y: -dy, Z: -dz}
			dopplerFactor := relativity.DopplerFactor(beta, nToObserver, gamma)
			shiftedTemp := relativity.ShiftColorTemperature(baseTemp, dopplerFactor)
			shiftedRGB := relativity.TemperatureToRGB(shiftedTemp)
			beamFactor := relativity.BeamBrightness(dopplerFactor)
			if beamFactor > 8 {
				beamFactor = 8
			}
			brightness *= beamFactor
			starColor = color.RGBA{shiftedRGB.R, shiftedRGB.G, shiftedRGB.B, 255}
		} else {
			baseRGB := relativity.TemperatureToRGB(baseTemp)
			starColor = color.RGBA{baseRGB.R, baseRGB.G, baseRGB.B, 255}
		}

		// Clamp brightness
		if brightness > 1.0 {
			brightness = 1.0
		}
		if brightness < 0.03 {
			continue
		}

		// Scale color
		r := uint8(float64(starColor.R) * brightness)
		g := uint8(float64(starColor.G) * brightness)
		b := uint8(float64(starColor.B) * brightness)

		ix, iy := int(screenX), int(screenY)
		if ix >= 0 && ix < width && iy >= 0 && iy < height {
			img.Set(ix, iy, color.RGBA{r, g, b, 255})

			// Glow for bright stars
			if brightness > 0.5 && ix > 0 && ix < width-1 && iy > 0 && iy < height-1 {
				glowR, glowG, glowB := r/2, g/2, b/2
				img.Set(ix-1, iy, color.RGBA{glowR, glowG, glowB, 255})
				img.Set(ix+1, iy, color.RGBA{glowR, glowG, glowB, 255})
				img.Set(ix, iy-1, color.RGBA{glowR, glowG, glowB, 255})
				img.Set(ix, iy+1, color.RGBA{glowR, glowG, glowB, 255})
			}
		}
	}

	// Draw GR effects - event horizon and photon sphere
	// Position based on view direction parameter
	if params.GrPhi > 0.1 {
		// Calculate theta/phi for the view direction
		// Normalize view direction first
		vLen := math.Sqrt(params.ViewDirX*params.ViewDirX + params.ViewDirY*params.ViewDirY + params.ViewDirZ*params.ViewDirZ)
		if vLen < 0.001 {
			vLen = 1
		}
		vdx := params.ViewDirX / vLen
		vdy := params.ViewDirY / vLen
		vdz := params.ViewDirZ / vLen

		viewTheta := math.Atan2(vdz, vdx)
		viewPhi := math.Asin(vdy)

		// Map to equirectangular coordinates
		viewU := (viewTheta + math.Pi) / (2 * math.Pi)
		viewV := (math.Pi/2 - viewPhi) / math.Pi

		centerX := viewU * float64(width)
		centerY := viewV * float64(height)

		// Scale effect size relative to image dimensions
		// At phi = 0.5, event horizon covers significant portion of view
		screenScale := float64(height) / 6.0
		eventHorizonRadius := params.GrPhi * screenScale * 2
		photonSphereRadius := eventHorizonRadius * 1.5

		// Draw accretion disk FIRST (behind event horizon)
		// Use finer steps for smoother rendering
		accretionInner := eventHorizonRadius * 1.1
		accretionOuter := eventHorizonRadius * 3.0
		angleStep := 0.003 // Fine angle step for smooth circle
		for angle := 0.0; angle < 2*math.Pi; angle += angleStep {
			cosA := math.Cos(angle)
			sinA := math.Sin(angle)
			for r := accretionInner; r <= accretionOuter; r += 0.3 {
				// Intensity falls off with distance from inner edge
				t := (r - accretionInner) / (accretionOuter - accretionInner)
				intensity := (1.0 - t) * (1.0 - t) // Quadratic falloff

				// Doppler shift simulation - one side brighter (approaching)
				dopplerBoost := 1.0 + 0.6*cosA
				intensity *= dopplerBoost

				// Add some variation for realism
				variation := 0.8 + 0.2*math.Sin(angle*8+r*0.5)
				intensity *= variation

				px := centerX + cosA*r
				py := centerY + sinA*r
				if px >= 0 && px < float64(width) && py >= 0 && py < float64(height) {
					// Hot accretion disk - orange/yellow/white gradient
					cr := uint8(math.Min(255, 255*intensity))
					cg := uint8(math.Min(255, (200-t*100)*intensity))
					cb := uint8(math.Min(255, (100-t*80)*intensity))
					img.Set(int(px), int(py), color.RGBA{cr, cg, cb, 255})
				}
			}
		}

		// Draw photon sphere ring (glowing ring at 1.5x event horizon)
		photonWidth := eventHorizonRadius * 0.15
		for angle := 0.0; angle < 2*math.Pi; angle += angleStep {
			cosA := math.Cos(angle)
			sinA := math.Sin(angle)
			for dr := -photonWidth; dr <= photonWidth; dr += 0.5 {
				r := photonSphereRadius + dr
				// Gaussian falloff from center of ring
				ringIntensity := math.Exp(-dr * dr / (photonWidth * photonWidth * 0.5))
				px := centerX + cosA*r
				py := centerY + sinA*r
				if px >= 0 && px < float64(width) && py >= 0 && py < float64(height) {
					// Orange-red glow
					cr := uint8(math.Min(255, 120*ringIntensity))
					cg := uint8(math.Min(255, 60*ringIntensity))
					cb := uint8(math.Min(255, 20*ringIntensity))
					img.Set(int(px), int(py), color.RGBA{cr, cg, cb, 255})
				}
			}
		}

		// Draw event horizon (black disk with soft edge)
		cxi := int(centerX)
		cyi := int(centerY)
		maxR := int(eventHorizonRadius) + 3
		edgeWidth := 2.0 // Soft edge width
		for dy := -maxR; dy <= maxR; dy++ {
			for dx := -maxR; dx <= maxR; dx++ {
				dist := math.Sqrt(float64(dx*dx + dy*dy))
				px, py := cxi+dx, cyi+dy
				if px >= 0 && px < width && py >= 0 && py < height {
					if dist <= eventHorizonRadius-edgeWidth {
						// Solid black inside
						img.Set(px, py, color.RGBA{0, 0, 0, 255})
					} else if dist <= eventHorizonRadius+edgeWidth {
						// Soft edge with gradient
						t := (dist - (eventHorizonRadius - edgeWidth)) / (edgeWidth * 2)
						alpha := uint8(255 * (1.0 - t))
						img.Set(px, py, color.RGBA{0, 0, 0, alpha})
					}
				}
			}
		}
	}

	return img
}

// renderProceduralEquirectangular creates a procedural starfield in equirectangular projection
func (sv *SpaceView) renderProceduralEquirectangular(width, height int) *ebiten.Image {
	img := ebiten.NewImage(width, height)
	img.Fill(color.RGBA{5, 8, 15, 255})

	// More stars for full-sky view
	for i := 0; i < 2000; i++ {
		h := uint32(i * 2654435761)
		x := int(h % uint32(width))
		h = h * 2654435761
		y := int(h % uint32(height))
		h = h * 2654435761
		brightness := uint8(80 + h%176)

		img.Set(x, y, color.RGBA{brightness, brightness, brightness + 10, 255})

		// Some bright stars get glow
		if brightness > 200 && x > 1 && y > 1 && x < width-2 && y < height-2 {
			dim := brightness / 3
			img.Set(x-1, y, color.RGBA{dim, dim, dim, 255})
			img.Set(x+1, y, color.RGBA{dim, dim, dim, 255})
			img.Set(x, y-1, color.RGBA{dim, dim, dim, 255})
			img.Set(x, y+1, color.RGBA{dim, dim, dim, 255})
		}
	}

	return img
}
