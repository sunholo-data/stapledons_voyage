package tetra

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/solarlune/tetra3d"
)

// Ring represents a planetary ring (like Saturn's rings).
type Ring struct {
	model    *tetra3d.Model
	material *tetra3d.Material
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
