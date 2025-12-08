package tetra

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/solarlune/tetra3d"
)

// Planet represents a 3D planet sphere.
type Planet struct {
	name        string
	model       *tetra3d.Model
	material    *tetra3d.Material
	rotation    float64
	rotationSpd float64
}

// NewPlanet creates a new planet with a solid color (no texture).
// Radius is used for scaling the unit icosphere.
func NewPlanet(name string, radius float64, c color.RGBA) *Planet {
	// Create icosphere mesh (3 subdivisions for N64 aesthetic)
	// NewIcosphereMesh creates a unit sphere, we scale it
	mesh := tetra3d.NewIcosphereMesh(3)

	// Create material with solid color
	mat := tetra3d.NewMaterial("planet_" + name)
	mat.Color = tetra3d.NewColor(
		float32(c.R)/255.0,
		float32(c.G)/255.0,
		float32(c.B)/255.0,
		float32(c.A)/255.0,
	)

	// Apply material to all mesh parts
	for _, meshPart := range mesh.MeshParts {
		meshPart.Material = mat
	}

	// Create model node
	model := tetra3d.NewModel(name, mesh)

	// Scale to desired radius
	r := float32(radius)
	model.SetLocalScale(r, r, r)

	return &Planet{
		name:        name,
		model:       model,
		material:    mat,
		rotation:    0,
		rotationSpd: 0.5, // Default rotation speed (radians per second)
	}
}

// NewTexturedPlanet creates a planet with a texture.
func NewTexturedPlanet(name string, radius float64, texture *ebiten.Image) *Planet {
	// Create icosphere mesh
	mesh := tetra3d.NewIcosphereMesh(3)

	// Create material with texture
	mat := tetra3d.NewMaterial("planet_" + name)
	mat.Texture = texture
	mat.Color = tetra3d.NewColor(1, 1, 1, 1) // Full brightness

	// Apply material to all mesh parts
	for _, meshPart := range mesh.MeshParts {
		meshPart.Material = mat
	}

	// Create model node
	model := tetra3d.NewModel(name, mesh)

	// Scale to desired radius
	r := float32(radius)
	model.SetLocalScale(r, r, r)

	return &Planet{
		name:        name,
		model:       model,
		material:    mat,
		rotation:    0,
		rotationSpd: 0.5,
	}
}

// Name returns the planet's name.
func (p *Planet) Name() string {
	return p.name
}

// Model returns the underlying Tetra3D model.
func (p *Planet) Model() *tetra3d.Model {
	return p.model
}

// SetPosition sets the planet's position.
func (p *Planet) SetPosition(x, y, z float64) {
	p.model.SetLocalPosition(float32(x), float32(y), float32(z))
}

// SetRotation sets the planet's Y-axis rotation in radians.
// This rotates incrementally from the current rotation.
func (p *Planet) SetRotation(yaw float64) {
	// Calculate delta rotation
	delta := yaw - p.rotation
	p.rotation = yaw

	// Rotate around Y axis
	if delta != 0 {
		p.model.Rotate(0, 1, 0, float32(delta))
	}
}

// SetRotationSpeed sets the rotation speed in radians per second.
func (p *Planet) SetRotationSpeed(speed float64) {
	p.rotationSpd = speed
}

// Update updates the planet's rotation.
func (p *Planet) Update(dt float64) {
	// Rotate directly by the delta amount
	delta := p.rotationSpd * dt
	p.rotation += delta
	p.model.Rotate(0, 1, 0, float32(delta))
}

// AddToScene adds the planet to a scene.
func (p *Planet) AddToScene(scene *Scene) {
	scene.Root().AddChildren(p.model)
}

// RemoveFromScene removes the planet from its parent.
func (p *Planet) RemoveFromScene() {
	if p.model != nil && p.model.Parent() != nil {
		p.model.Unparent()
	}
}

// SetTexture updates the planet's texture.
func (p *Planet) SetTexture(texture *ebiten.Image) {
	p.material.Texture = texture
}

// SetColor sets the planet's color (for untextured planets).
func (p *Planet) SetColor(c color.RGBA) {
	p.material.Color = tetra3d.NewColor(
		float32(c.R)/255.0,
		float32(c.G)/255.0,
		float32(c.B)/255.0,
		float32(c.A)/255.0,
	)
}

// SetScale sets the planet's scale (radius).
func (p *Planet) SetScale(scale float64) {
	s := float32(scale)
	p.model.SetLocalScale(s, s, s)
}

// Rotation returns the current rotation in radians.
func (p *Planet) Rotation() float64 {
	return p.rotation
}
