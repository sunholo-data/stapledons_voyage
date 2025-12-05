package shader

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

// SRWarp handles special relativistic visual effects.
// Applies aberration (direction warping), Doppler shift (color change),
// and relativistic beaming (brightness) based on ship velocity.
type SRWarp struct {
	manager *Manager
	enabled bool
	buffer  *ebiten.Image
	screenW int
	screenH int

	// Velocity components (as fraction of c)
	betaX, betaY, betaZ float64

	// Pre-computed values
	gamma float64

	// Field of view in radians
	fov float64

	// View angle: 0=front, π/2=side, π=back
	viewAngle float64
}

// NewSRWarp creates a new SR warp effect.
func NewSRWarp(manager *Manager) *SRWarp {
	return &SRWarp{
		manager: manager,
		enabled: false,
		fov:     math.Pi / 3, // 60 degrees default
	}
}

// SetEnabled enables or disables the SR warp effect.
func (s *SRWarp) SetEnabled(enabled bool) {
	s.enabled = enabled
}

// IsEnabled returns whether SR warp is enabled.
func (s *SRWarp) IsEnabled() bool {
	return s.enabled
}

// Toggle toggles the effect and returns new state.
func (s *SRWarp) Toggle() bool {
	s.enabled = !s.enabled
	return s.enabled
}

// SetVelocity sets the ship velocity as a fraction of c.
// betaZ is the forward velocity (positive = moving forward).
func (s *SRWarp) SetVelocity(betaX, betaY, betaZ float64) {
	s.betaX = betaX
	s.betaY = betaY
	s.betaZ = betaZ

	// Compute gamma (Lorentz factor)
	betaSquared := betaX*betaX + betaY*betaY + betaZ*betaZ
	if betaSquared >= 1.0 {
		betaSquared = 0.999999 // Clamp to just under c
	}
	s.gamma = 1.0 / math.Sqrt(1.0-betaSquared)
}

// SetForwardVelocity is a convenience method for forward-only motion.
func (s *SRWarp) SetForwardVelocity(beta float64) {
	s.SetVelocity(0, 0, beta)
}

// SetFOV sets the field of view in radians.
func (s *SRWarp) SetFOV(fov float64) {
	s.fov = fov
}

// SetViewAngle sets the viewing direction.
// 0 = front (looking in direction of motion)
// π/2 = side (looking perpendicular to motion)
// π = back (looking opposite to motion)
func (s *SRWarp) SetViewAngle(angle float64) {
	s.viewAngle = angle
}

// GetViewAngle returns the current view angle in radians.
func (s *SRWarp) GetViewAngle() float64 {
	return s.viewAngle
}

// GetVelocity returns the current velocity components.
func (s *SRWarp) GetVelocity() (betaX, betaY, betaZ float64) {
	return s.betaX, s.betaY, s.betaZ
}

// GetBeta returns the magnitude of velocity (|beta|).
func (s *SRWarp) GetBeta() float64 {
	return math.Sqrt(s.betaX*s.betaX + s.betaY*s.betaY + s.betaZ*s.betaZ)
}

// GetGamma returns the Lorentz factor.
func (s *SRWarp) GetGamma() float64 {
	return s.gamma
}

// SetSize updates render buffer dimensions.
func (s *SRWarp) SetSize(w, h int) {
	if s.screenW == w && s.screenH == h {
		return
	}
	s.screenW = w
	s.screenH = h
	s.buffer = ebiten.NewImage(w, h)
}

// Apply applies the SR warp effect.
// Returns true if the effect was applied, false if skipped.
func (s *SRWarp) Apply(dst, src *ebiten.Image) bool {
	if !s.enabled {
		return false
	}

	// Skip if velocity is negligible
	beta := s.GetBeta()
	if beta < 0.001 {
		return false
	}

	shader, err := s.manager.Get("sr_warp")
	if err != nil {
		return false
	}

	// Ensure buffer is sized
	w, h := src.Bounds().Dx(), src.Bounds().Dy()
	s.SetSize(w, h)

	opts := &ebiten.DrawRectShaderOptions{}
	opts.Images[0] = src
	opts.Uniforms = map[string]any{
		"BetaX":     float32(s.betaX),
		"BetaY":     float32(s.betaY),
		"BetaZ":     float32(s.betaZ),
		"Gamma":     float32(s.gamma),
		"FOV":       float32(s.fov),
		"ViewAngle": float32(s.viewAngle),
	}

	dst.DrawRectShader(s.screenW, s.screenH, shader, opts)
	return true
}
