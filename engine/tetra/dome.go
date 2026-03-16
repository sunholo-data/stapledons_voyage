package tetra

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/solarlune/tetra3d"
)

// Dome represents a hemispherical or spherical dome mesh for observation windows.
// The dome is rendered from the inside, showing space through its surface.
type Dome struct {
	model    *tetra3d.Model
	mesh     *tetra3d.Mesh
	material *tetra3d.Material
	radius   float64
	arcAngle float64 // π = hemisphere, 2π = full sphere
	rings    int
	segments int
}

// NewDome creates a new dome mesh with the given parameters.
// radius: dome radius in meters
// arcAngle: coverage angle (π = hemisphere, 2π = full sphere)
// rings: number of latitude rings (more = smoother)
// segments: number of longitude segments (more = smoother)
func NewDome(name string, radius, arcAngle float64, rings, segments int) *Dome {
	d := &Dome{
		radius:   radius,
		arcAngle: arcAngle,
		rings:    rings,
		segments: segments,
	}

	d.mesh = d.generateMesh(name)
	d.model = tetra3d.NewModel(name+"_model", d.mesh)

	return d
}

// NewHemisphere creates a hemisphere dome (half sphere, π/2 arc angle).
// This creates the upper half of a sphere, suitable for observation decks.
func NewHemisphere(name string, radius float64) *Dome {
	return NewDome(name, radius, math.Pi/2, 16, 32)
}

// NewBubble creates a full sphere dome (π arc angle from top to equator and beyond).
func NewBubble(name string, radius float64) *Dome {
	return NewDome(name, radius, math.Pi, 24, 48)
}

// NewTransparentBubble creates a see-through bubble sphere for external ship view.
// alpha: transparency (0.0 = invisible, 1.0 = opaque, 0.25-0.4 recommended)
// The bubble is rendered with both inside and outside faces visible.
func NewTransparentBubble(name string, radius float64, alpha float64) *Dome {
	d := NewDome(name, radius, math.Pi, 24, 48)
	d.SetTransparency(alpha)
	// For external viewing, we want to see OUTSIDE the sphere (normals outward)
	// But also see through it, so disable backface culling
	d.material.BackfaceCulling = false
	return d
}

// NewExternalBubble creates a bubble for viewing from OUTSIDE (normals point outward).
// Use this for the bubble ship HUD where camera is outside looking at the bubble.
func NewExternalBubble(name string, radius float64, alpha float64) *Dome {
	d := &Dome{
		radius:   radius,
		arcAngle: math.Pi,
		rings:    24,
		segments: 48,
	}
	d.mesh = d.generateExternalMesh(name)
	d.model = tetra3d.NewModel(name+"_model", d.mesh)
	d.SetTransparency(alpha)
	return d
}

// generateExternalMesh creates a sphere with normals pointing OUTWARD (for external viewing).
func (d *Dome) generateExternalMesh(name string) *tetra3d.Mesh {
	mesh := tetra3d.NewMesh(name + "_mesh")

	vertices := make([]tetra3d.VertexInfo, 0, (d.rings+1)*(d.segments+1))

	for ring := 0; ring <= d.rings; ring++ {
		phi := (float64(ring) / float64(d.rings)) * d.arcAngle

		for seg := 0; seg <= d.segments; seg++ {
			theta := (float64(seg) / float64(d.segments)) * 2 * math.Pi

			sinPhi := math.Sin(phi)
			cosPhi := math.Cos(phi)
			sinTheta := math.Sin(theta)
			cosTheta := math.Cos(theta)

			x := float32(d.radius * sinPhi * cosTheta)
			y := float32(d.radius * cosPhi)
			z := float32(d.radius * sinPhi * sinTheta)

			u := float32(seg) / float32(d.segments)
			v := float32(phi / math.Pi)

			// Normals point OUTWARD (we're outside the bubble looking at it)
			nx := x / float32(d.radius)
			ny := y / float32(d.radius)
			nz := z / float32(d.radius)

			vertices = append(vertices, tetra3d.VertexInfo{
				X: x, Y: y, Z: z,
				U: u, V: v,
				NormalX: nx, NormalY: ny, NormalZ: nz,
			})
		}
	}

	mesh.AddVertices(vertices...)

	// Generate indices with standard winding (CCW for external view)
	indices := make([]int, 0, d.rings*d.segments*6)
	vertsPerRing := d.segments + 1

	for ring := 0; ring < d.rings; ring++ {
		for seg := 0; seg < d.segments; seg++ {
			topLeft := ring*vertsPerRing + seg
			topRight := ring*vertsPerRing + seg + 1
			bottomLeft := (ring+1)*vertsPerRing + seg
			bottomRight := (ring+1)*vertsPerRing + seg + 1

			// Standard winding for external view
			indices = append(indices, topLeft, bottomLeft, bottomRight)
			indices = append(indices, topLeft, bottomRight, topRight)
		}
	}

	// Create transparent material
	d.material = tetra3d.NewMaterial(name + "_mat")
	d.material.Shadeless = false // Allow lighting
	d.material.BackfaceCulling = false
	d.material.TransparencyMode = tetra3d.TransparencyModeTransparent
	d.material.Color = tetra3d.NewColor(0.3, 0.5, 0.7, 0.3) // Blue tint, 30% alpha

	mesh.AddMeshPart(d.material, indices...)
	mesh.UpdateBounds()

	return mesh
}

// generateMesh creates the hemisphere/sphere mesh vertices and triangles.
// Vertices are generated from top (pole) down to the cutoff angle.
// UVs are mapped for equirectangular texture sampling.
func (d *Dome) generateMesh(name string) *tetra3d.Mesh {
	mesh := tetra3d.NewMesh(name + "_mesh")

	// Generate vertices
	vertices := make([]tetra3d.VertexInfo, 0, (d.rings+1)*(d.segments+1))

	for ring := 0; ring <= d.rings; ring++ {
		// phi is the angle from the top (Y-axis positive)
		phi := (float64(ring) / float64(d.rings)) * d.arcAngle

		for seg := 0; seg <= d.segments; seg++ {
			// theta is the angle around the Y-axis
			theta := (float64(seg) / float64(d.segments)) * 2 * math.Pi

			// Spherical to Cartesian (Y-up)
			sinPhi := math.Sin(phi)
			cosPhi := math.Cos(phi)
			sinTheta := math.Sin(theta)
			cosTheta := math.Cos(theta)

			x := float32(d.radius * sinPhi * cosTheta)
			y := float32(d.radius * cosPhi)
			z := float32(d.radius * sinPhi * sinTheta)

			// UV for equirectangular mapping
			// U wraps around horizontally (theta)
			// V goes from top to bottom (phi / π)
			u := float32(seg) / float32(d.segments)
			v := float32(phi / math.Pi)

			// Normals point INWARD (we're inside the dome looking out)
			nx := -x / float32(d.radius)
			ny := -y / float32(d.radius)
			nz := -z / float32(d.radius)

			vertices = append(vertices, tetra3d.VertexInfo{
				X:       x,
				Y:       y,
				Z:       z,
				U:       u,
				V:       v,
				NormalX: nx,
				NormalY: ny,
				NormalZ: nz,
			})
		}
	}

	mesh.AddVertices(vertices...)

	// Generate triangle indices
	// For inside viewing, we need correct winding order
	indices := make([]int, 0, d.rings*d.segments*6)
	vertsPerRing := d.segments + 1

	for ring := 0; ring < d.rings; ring++ {
		for seg := 0; seg < d.segments; seg++ {
			// Current quad corners
			topLeft := ring*vertsPerRing + seg
			topRight := ring*vertsPerRing + seg + 1
			bottomLeft := (ring+1)*vertsPerRing + seg
			bottomRight := (ring+1)*vertsPerRing + seg + 1

			// Reversed winding for inside view (CCW when looking from inside)
			// First triangle
			indices = append(indices, topLeft, bottomRight, bottomLeft)
			// Second triangle
			indices = append(indices, topLeft, topRight, bottomRight)
		}
	}

	// Create material - shadeless for star texture
	d.material = tetra3d.NewMaterial(name + "_mat")
	d.material.Shadeless = true
	d.material.BackfaceCulling = false
	d.material.TransparencyMode = tetra3d.TransparencyModeOpaque
	d.material.Color = tetra3d.NewColor(0.1, 0.1, 0.2, 1) // Dark blue default

	mesh.AddMeshPart(d.material, indices...)
	mesh.UpdateBounds()

	return mesh
}

// Model returns the Tetra3D model for scene integration.
func (d *Dome) Model() *tetra3d.Model {
	return d.model
}

// SetTexture applies a texture to the dome surface.
func (d *Dome) SetTexture(tex *ebiten.Image) {
	d.material.Texture = tex
	d.material.UseTexture = true
	d.material.Color = tetra3d.NewColor(1, 1, 1, 1)
}

// SetCubemap applies a cubemap texture (as equirectangular) to the dome.
func (d *Dome) SetCubemap(tex *ebiten.Image) {
	d.SetTexture(tex)
}

// SetPosition moves the dome center.
func (d *Dome) SetPosition(x, y, z float64) {
	d.model.SetLocalPosition(float32(x), float32(y), float32(z))
}

// AddToScene adds the dome to a Tetra3D scene.
func (d *Dome) AddToScene(scene *Scene) {
	scene.Root().AddChildren(d.model)
}

// SetVisible shows or hides the dome.
func (d *Dome) SetVisible(visible bool) {
	d.model.SetVisible(visible, true)
}

// Radius returns the dome radius.
func (d *Dome) Radius() float64 {
	return d.radius
}

// ArcAngle returns the dome arc coverage.
func (d *Dome) ArcAngle() float64 {
	return d.arcAngle
}

// SetRotation rotates the dome around the given axis by the specified angle (radians).
// Use this to orient domes to face different directions (side windows, etc).
func (d *Dome) SetRotation(axisX, axisY, axisZ, angle float64) {
	d.model.SetLocalRotation(tetra3d.NewMatrix4Rotate(float32(axisX), float32(axisY), float32(axisZ), float32(angle)))
}

// FaceDirection orients the dome to face the given direction.
// The dome's "open" side will point in this direction.
// direction: "up" (default), "down", "north", "south", "east", "west"
func (d *Dome) FaceDirection(direction string) {
	switch direction {
	case "up":
		// Default - no rotation needed (dome opens upward)
		d.model.SetLocalRotation(tetra3d.NewMatrix4())
	case "down":
		// Flip upside down
		d.SetRotation(1, 0, 0, math.Pi)
	case "north":
		// Face -Z (north wall, opens toward -Z)
		d.SetRotation(1, 0, 0, -math.Pi/2)
	case "south":
		// Face +Z (south wall, opens toward +Z)
		d.SetRotation(1, 0, 0, math.Pi/2)
	case "east":
		// Face +X (east wall)
		d.SetRotation(0, 0, 1, math.Pi/2)
	case "west":
		// Face -X (west wall)
		d.SetRotation(0, 0, 1, -math.Pi/2)
	}
}

// SetTransparency sets the dome's transparency level.
// alpha: 0.0 = fully transparent, 1.0 = fully opaque
// Recommended: 0.25-0.4 for see-through bubble effect
func (d *Dome) SetTransparency(alpha float64) {
	d.material.TransparencyMode = tetra3d.TransparencyModeTransparent
	// Preserve RGB, update alpha
	r := d.material.Color.R
	g := d.material.Color.G
	b := d.material.Color.B
	d.material.Color = tetra3d.NewColor(r, g, b, float32(alpha))
}

// SetColor sets the dome's base color with optional transparency.
// r, g, b: color components (0.0-1.0)
// alpha: transparency (0.0-1.0)
func (d *Dome) SetColor(r, g, b, alpha float64) {
	if alpha < 1.0 {
		d.material.TransparencyMode = tetra3d.TransparencyModeTransparent
	} else {
		d.material.TransparencyMode = tetra3d.TransparencyModeOpaque
	}
	d.material.Color = tetra3d.NewColor(float32(r), float32(g), float32(b), float32(alpha))
}

// SetShadeless controls whether the dome is affected by scene lighting.
// shadeless=true: dome ignores lighting (good for sky domes)
// shadeless=false: dome is lit by scene lights (good for physical objects)
func (d *Dome) SetShadeless(shadeless bool) {
	d.material.Shadeless = shadeless
}
