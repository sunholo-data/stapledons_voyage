package tetra

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/solarlune/tetra3d"
)

// Ring represents a planetary ring (like Saturn's rings).
type Ring struct {
	model    *tetra3d.Model
	material *tetra3d.Material
}

// RingBand defines a single band within a ring system (e.g., Saturn's A, B, C rings).
// InnerRadius and OuterRadius are absolute distances from planet center.
type RingBand struct {
	InnerRadius float64    // Inner edge (absolute distance from planet center)
	OuterRadius float64    // Outer edge (absolute distance from planet center)
	Color       color.RGBA // Band color
	Opacity     float64    // 0.0-1.0 (for dust transparency)
	Density     float64    // 0.0-1.0 (affects vertex color variation)
}

// RingBandSpec defines a ring band using planet radii multipliers.
// This is the portable format - convert to RingBand with MakeRingBands().
type RingBandSpec struct {
	InnerMult float64    // Inner edge as planet radii multiplier (1.0 = planet surface)
	OuterMult float64    // Outer edge as planet radii multiplier
	Color     color.RGBA // Band color
	Opacity   float64    // 0.0-1.0 (for dust transparency)
	Density   float64    // 0.0-1.0 (affects vertex color variation)
}

// MakeRingBands converts portable RingBandSpecs to absolute RingBands for a given planet radius.
func MakeRingBands(planetRadius float64, specs []RingBandSpec) []RingBand {
	bands := make([]RingBand, len(specs))
	for i, spec := range specs {
		bands[i] = RingBand{
			InnerRadius: spec.InnerMult * planetRadius,
			OuterRadius: spec.OuterMult * planetRadius,
			Color:       spec.Color,
			Opacity:     spec.Opacity,
			Density:     spec.Density,
		}
	}
	return bands
}

// NewRing creates a ring around a planet.
// innerRadius: distance from planet center to inner edge of ring
// outerRadius: distance from planet center to outer edge of ring
// segments: number of segments around the ring (more = smoother)
func NewRing(name string, innerRadius, outerRadius float64, texture *ebiten.Image) *Ring {
	mesh := NewRingMesh(32, float32(innerRadius), float32(outerRadius))

	mat := tetra3d.NewMaterial("ring_" + name)
	if texture != nil {
		mat.Texture = texture
		// With texture, use alpha clip for ring gaps
		mat.TransparencyMode = tetra3d.TransparencyModeAlphaClip
	} else {
		// Without texture, use opaque beige/tan ring color
		mat.Color = tetra3d.NewColor(0.85, 0.75, 0.6, 1.0)
		mat.TransparencyMode = tetra3d.TransparencyModeOpaque
	}
	mat.Shadeless = true      // Rings don't need lighting
	mat.BackfaceCulling = false // Render both sides of ring

	for _, meshPart := range mesh.MeshParts {
		meshPart.Material = mat
	}

	model := tetra3d.NewModel(name+"_ring", mesh)

	return &Ring{
		model:    model,
		material: mat,
	}
}

// NewRingMesh creates a flat ring mesh with proper UV mapping.
// The ring lies in the XZ plane (horizontal).
func NewRingMesh(segments int, innerRadius, outerRadius float32) *tetra3d.Mesh {
	if segments < 8 {
		segments = 8
	}

	mesh := tetra3d.NewMesh("Ring")

	// Generate vertices: inner and outer circles
	vertices := make([]tetra3d.VertexInfo, 0, (segments+1)*2)

	for i := 0; i <= segments; i++ {
		// u coordinate wraps around the ring
		u := float32(i) / float32(segments)
		theta := float64(u) * 2 * math.Pi

		cosTheta := float32(math.Cos(theta))
		sinTheta := float32(math.Sin(theta))

		// Inner vertex (v = 0)
		vertices = append(vertices, tetra3d.VertexInfo{
			X:       innerRadius * cosTheta,
			Y:       0,
			Z:       innerRadius * sinTheta,
			U:       u,
			V:       0,
			NormalX: 0,
			NormalY: 1,
			NormalZ: 0,
		})

		// Outer vertex (v = 1)
		vertices = append(vertices, tetra3d.VertexInfo{
			X:       outerRadius * cosTheta,
			Y:       0,
			Z:       outerRadius * sinTheta,
			U:       u,
			V:       1,
			NormalX: 0,
			NormalY: 1,
			NormalZ: 0,
		})
	}

	mesh.AddVertices(vertices...)

	// Generate indices for quads (as two triangles each)
	indices := make([]int, 0, segments*6)

	for i := 0; i < segments; i++ {
		// Each segment has 4 vertices: inner0, outer0, inner1, outer1
		inner0 := i * 2
		outer0 := i*2 + 1
		inner1 := (i + 1) * 2
		outer1 := (i+1)*2 + 1

		// First triangle (inner0, outer0, outer1)
		indices = append(indices, inner0, outer0, outer1)

		// Second triangle (inner0, outer1, inner1)
		indices = append(indices, inner0, outer1, inner1)
	}

	mat := tetra3d.NewMaterial("Ring")
	mesh.AddMeshPart(mat, indices...)

	mesh.UpdateBounds()

	return mesh
}

// Model returns the underlying Tetra3D model.
func (r *Ring) Model() *tetra3d.Model {
	return r.model
}

// SetTilt sets the ring's tilt angle in radians around the X axis.
// Saturn's rings are tilted about 27 degrees.
func (r *Ring) SetTilt(radians float64) {
	// Rotate around X axis (tilt forward/back)
	r.model.Rotate(1, 0, 0, float32(radians))
}

// AddToScene adds the ring to a scene.
func (r *Ring) AddToScene(scene *Scene) {
	scene.Root().AddChildren(r.model)
}

// SetPosition sets the ring's position (should match planet position).
func (r *Ring) SetPosition(x, y, z float64) {
	r.model.SetLocalPosition(float32(x), float32(y), float32(z))
}

// Update rotates the ring (usually synced with planet).
func (r *Ring) Update(dt float64) {
	// Rings typically rotate with the planet
	r.model.Rotate(0, 1, 0, float32(dt*0.5))
}

// NewDustRing creates a translucent ring with dust-like appearance.
// opacity: 0.0-1.0, lower values create more dust-like transparency
func NewDustRing(name string, innerRadius, outerRadius float64, baseColor color.RGBA, opacity float64) *Ring {
	mesh := NewRingMeshWithVertexColors(64, float32(innerRadius), float32(outerRadius), baseColor, float32(opacity))

	mat := tetra3d.NewMaterial("dust_ring_" + name)
	// Use transparent mode for soft blending
	mat.TransparencyMode = tetra3d.TransparencyModeTransparent
	mat.Color = tetra3d.NewColor(
		float32(baseColor.R)/255,
		float32(baseColor.G)/255,
		float32(baseColor.B)/255,
		float32(opacity),
	)
	mat.Shadeless = true
	mat.BackfaceCulling = false

	for _, meshPart := range mesh.MeshParts {
		meshPart.Material = mat
	}

	model := tetra3d.NewModel(name+"_dust_ring", mesh)

	return &Ring{
		model:    model,
		material: mat,
	}
}

// NewRingMeshWithVertexColors creates a ring mesh with per-vertex color/alpha variation.
// This creates a more dust-like, particulate appearance.
func NewRingMeshWithVertexColors(segments int, innerRadius, outerRadius float32, baseColor color.RGBA, opacity float32) *tetra3d.Mesh {
	if segments < 8 {
		segments = 8
	}

	mesh := tetra3d.NewMesh("DustRing")

	// Generate vertices with color variation
	vertices := make([]tetra3d.VertexInfo, 0, (segments+1)*2)

	baseR := float32(baseColor.R) / 255
	baseG := float32(baseColor.G) / 255
	baseB := float32(baseColor.B) / 255

	for i := 0; i <= segments; i++ {
		u := float32(i) / float32(segments)
		theta := float64(u) * 2 * math.Pi

		cosTheta := float32(math.Cos(theta))
		sinTheta := float32(math.Sin(theta))

		// Add noise-based variation to simulate dust clumping
		// Simple pseudo-random based on angle
		noiseI := float32(math.Sin(float64(i)*7.3) * 0.5 + 0.5)
		noiseO := float32(math.Sin(float64(i)*5.7+2.1) * 0.5 + 0.5)

		// Inner vertex - slightly darker, more transparent
		innerAlpha := opacity * (0.6 + 0.4*noiseI) // 60-100% of opacity
		innerColor := tetra3d.NewColor(
			baseR*(0.8+0.2*noiseI),
			baseG*(0.8+0.2*noiseI),
			baseB*(0.7+0.3*noiseI),
			innerAlpha,
		)
		vertices = append(vertices, tetra3d.VertexInfo{
			X:       innerRadius * cosTheta,
			Y:       0,
			Z:       innerRadius * sinTheta,
			U:       u,
			V:       0,
			NormalX: 0,
			NormalY: 1,
			NormalZ: 0,
			Colors:  []tetra3d.Color{innerColor},
		})

		// Outer vertex - brighter, more opaque
		outerAlpha := opacity * (0.7 + 0.3*noiseO)
		outerColor := tetra3d.NewColor(
			baseR*(0.9+0.1*noiseO),
			baseG*(0.9+0.1*noiseO),
			baseB*(0.85+0.15*noiseO),
			outerAlpha,
		)
		vertices = append(vertices, tetra3d.VertexInfo{
			X:       outerRadius * cosTheta,
			Y:       0,
			Z:       outerRadius * sinTheta,
			U:       u,
			V:       1,
			NormalX: 0,
			NormalY: 1,
			NormalZ: 0,
			Colors:  []tetra3d.Color{outerColor},
		})
	}

	mesh.AddVertices(vertices...)

	// Generate indices
	indices := make([]int, 0, segments*6)
	for i := 0; i < segments; i++ {
		inner0 := i * 2
		outer0 := i*2 + 1
		inner1 := (i + 1) * 2
		outer1 := (i+1)*2 + 1

		indices = append(indices, inner0, outer0, outer1)
		indices = append(indices, inner0, outer1, inner1)
	}

	mat := tetra3d.NewMaterial("DustRing")
	mesh.AddMeshPart(mat, indices...)
	mesh.UpdateBounds()

	return mesh
}

// SaturnRingSpecs defines Saturn's ring bands as portable multipliers.
// Saturn's rings (scaled to planet radii):
// D Ring: 1.11-1.24 (very faint) - not included
// C Ring: 1.24-1.53 (dim, inner)
// B Ring: 1.53-1.95 (brightest)
// Cassini Division: 1.95-2.03 (gap)
// A Ring: 2.03-2.27 (bright)
// F Ring: 2.33 (narrow) - not included
var SaturnRingSpecs = []RingBandSpec{
	// C Ring (inner, dim)
	{InnerMult: 1.24, OuterMult: 1.53,
		Color: color.RGBA{180, 160, 130, 255}, Opacity: 0.3, Density: 0.4},
	// B Ring (main, brightest)
	{InnerMult: 1.53, OuterMult: 1.95,
		Color: color.RGBA{220, 205, 170, 255}, Opacity: 0.7, Density: 0.9},
	// A Ring (outer, bright)
	{InnerMult: 2.03, OuterMult: 2.27,
		Color: color.RGBA{210, 190, 150, 255}, Opacity: 0.5, Density: 0.7},
}

// SaturnRingBands returns preset ring bands matching Saturn's main rings.
func SaturnRingBands(planetRadius float64) []RingBand {
	return MakeRingBands(planetRadius, SaturnRingSpecs)
}

// NewMultiBandRing creates a ring system with multiple concentric bands.
// Each band can have different color, opacity, and density.
func NewMultiBandRing(name string, bands []RingBand) []*Ring {
	rings := make([]*Ring, len(bands))

	for i, band := range bands {
		rings[i] = NewDustRing(
			name+"_band_"+string(rune('A'+i)),
			band.InnerRadius,
			band.OuterRadius,
			band.Color,
			band.Opacity,
		)
	}

	return rings
}

// RingSystem represents multiple ring bands rendered together.
type RingSystem struct {
	rings []*Ring
}

// NewRingSystem creates a complete ring system (like Saturn's).
func NewRingSystem(name string, bands []RingBand) *RingSystem {
	return &RingSystem{
		rings: NewMultiBandRing(name, bands),
	}
}

// AddToScene adds all ring bands to the scene.
func (rs *RingSystem) AddToScene(scene *Scene) {
	for _, ring := range rs.rings {
		ring.AddToScene(scene)
	}
}

// SetPosition sets position for all ring bands.
func (rs *RingSystem) SetPosition(x, y, z float64) {
	for _, ring := range rs.rings {
		ring.SetPosition(x, y, z)
	}
}

// SetTilt sets tilt angle for all ring bands.
func (rs *RingSystem) SetTilt(radians float64) {
	for _, ring := range rs.rings {
		ring.SetTilt(radians)
	}
}

// Update updates all ring bands.
func (rs *RingSystem) Update(dt float64) {
	for _, ring := range rs.rings {
		ring.Update(dt)
	}
}
