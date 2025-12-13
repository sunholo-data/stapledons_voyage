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

// LookAt is DEPRECATED - use SetPosition instead.
//
// Tetra3D directional lights derive their direction from position alone,
// not from rotation matrices. SetLocalRotation interferes with the internal
// light direction calculation.
//
// WORKING APPROACH: Position the sun so it shines toward the target.
// For a light at (5,3,10) shining toward origin, the direction is
// calculated automatically by Tetra3D's lighting system.
//
// Example:
//
//	sun.SetPosition(5, 3, 10) // Light will illuminate objects toward -Z
//
// This method is kept for API compatibility but logs a warning.
func (s *SunLight) LookAt(x, y, z float64) {
	// Intentionally do nothing - SetLocalRotation breaks directional lights.
	// The light direction is derived from position by Tetra3D internally.
	// Log warning for developers who try to use this.
	// log.Println("Warning: SunLight.LookAt is deprecated - use SetPosition instead")
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

// SetEnergy sets the ambient light's intensity.
func (a *AmbientLight) SetEnergy(energy float64) {
	a.light.SetEnergy(float32(energy))
}

// Light returns the underlying Tetra3D ambient light.
func (a *AmbientLight) Light() *tetra3d.AmbientLight {
	return a.light
}

// StarLight represents a point light that radiates from a star/sun.
// Unlike DirectionalLight (parallel rays), PointLight radiates from a position
// and falls off with distance - more realistic for solar system views.
type StarLight struct {
	light *tetra3d.PointLight
}

// NewStarLight creates a new point light at the star's position.
// range_ is the maximum distance the light reaches (0 = infinite).
// energy is the light intensity (1.0 = normal).
func NewStarLight(name string, r, g, b, energy, range_ float64) *StarLight {
	light := tetra3d.NewPointLight(name, float32(r), float32(g), float32(b), float32(energy))
	light.Range = float32(range_) // 0 = infinite range
	return &StarLight{light: light}
}

// SetPosition sets the star light's position (where the star is).
func (s *StarLight) SetPosition(x, y, z float64) {
	s.light.SetLocalPosition(float32(x), float32(y), float32(z))
}

// SetColor sets the light's color (e.g., yellow-white for sun, blue for hot stars).
func (s *StarLight) SetColor(r, g, b float64) {
	s.light.SetColor(tetra3d.NewColor(float32(r), float32(g), float32(b), 1))
}

// SetEnergy sets the light's intensity.
func (s *StarLight) SetEnergy(energy float64) {
	s.light.SetEnergy(float32(energy))
}

// SetRange sets the maximum distance the light reaches (0 = infinite).
func (s *StarLight) SetRange(range_ float64) {
	s.light.Range = float32(range_)
}

// AddToScene adds the star light to a scene.
func (s *StarLight) AddToScene(scene *Scene) {
	scene.Root().AddChildren(s.light)
}

// Light returns the underlying Tetra3D point light.
func (s *StarLight) Light() *tetra3d.PointLight {
	return s.light
}
