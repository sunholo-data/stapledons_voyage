package tetra

import "github.com/solarlune/tetra3d"

// SunLight represents a directional light simulating the sun.
type SunLight struct {
	light *tetra3d.DirectionalLight
}

// NewSunLight creates a new directional sun light.
func NewSunLight() *SunLight {
	// NewDirectionalLight(name, r, g, b, energy)
	light := tetra3d.NewDirectionalLight("sun", 1, 1, 1, 1) // White, full energy

	// Position the light source
	light.SetLocalPosition(5, 5, 5)

	return &SunLight{light: light}
}

// SetPosition sets the light's position.
func (s *SunLight) SetPosition(x, y, z float64) {
	s.light.SetLocalPosition(float32(x), float32(y), float32(z))
}

// SetColor sets the light's color.
func (s *SunLight) SetColor(r, g, b float64) {
	s.light.SetColor(tetra3d.NewColor(float32(r), float32(g), float32(b), 1))
}

// SetEnergy sets the light's intensity.
func (s *SunLight) SetEnergy(energy float64) {
	s.light.SetEnergy(float32(energy))
}

// AddToScene adds the light to a scene.
func (s *SunLight) AddToScene(scene *Scene) {
	scene.Root().AddChildren(s.light)
}

// Light returns the underlying Tetra3D light.
func (s *SunLight) Light() *tetra3d.DirectionalLight {
	return s.light
}

// AmbientLight represents ambient lighting for the scene.
type AmbientLight struct {
	light *tetra3d.AmbientLight
}

// NewAmbientLight creates a new ambient light.
func NewAmbientLight(r, g, b, energy float64) *AmbientLight {
	light := tetra3d.NewAmbientLight("ambient", float32(r), float32(g), float32(b), float32(energy))
	return &AmbientLight{light: light}
}

// AddToScene adds the ambient light to a scene.
func (a *AmbientLight) AddToScene(scene *Scene) {
	scene.Root().AddChildren(a.light)
}
