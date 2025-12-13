package shader

import (
	"github.com/hajimehoshi/ebiten/v2"
	"stapledons_voyage/engine/relativity"
)

// GRWarp handles general relativistic visual effects near massive objects.
// Applies gravitational lensing and redshift based on proximity to BH/NS/WD.
type GRWarp struct {
	manager *Manager
	enabled bool
	buffer  *ebiten.Image
	screenW int
	screenH int

	// GR context from simulation
	uniforms relativity.GRShaderUniforms

	// Demo mode parameters (when not using sim context)
	demoMode     bool
	demoPhi      float64
	demoCenter   [2]float32
	demoRs       float32
}

// NewGRWarp creates a new GR warp effect.
func NewGRWarp(manager *Manager) *GRWarp {
	return &GRWarp{
		manager:    manager,
		enabled:    false,
		demoCenter: [2]float32{0.5, 0.5}, // Screen center
		demoRs:     0.05,                  // 5% of screen
		demoPhi:    0.05,                  // Strong GR
	}
}

// SetEnabled enables or disables the GR warp effect.
func (g *GRWarp) SetEnabled(enabled bool) {
	g.enabled = enabled
}

// IsEnabled returns whether GR warp is enabled.
func (g *GRWarp) IsEnabled() bool {
	return g.enabled
}

// Toggle toggles the effect and returns new state.
func (g *GRWarp) Toggle() bool {
	g.enabled = !g.enabled
	return g.enabled
}

// IsDemoMode returns whether the effect is in demo mode.
func (g *GRWarp) IsDemoMode() bool {
	return g.demoMode
}

// SetUniforms sets the GR shader uniforms from simulation context.
func (g *GRWarp) SetUniforms(uniforms relativity.GRShaderUniforms) {
	g.uniforms = uniforms
	g.demoMode = false
}

// SetDemoMode enables demo mode with manual parameters.
func (g *GRWarp) SetDemoMode(centerX, centerY, rs, phi float32) {
	g.demoMode = true
	g.demoCenter = [2]float32{centerX, centerY}
	g.demoRs = rs
	g.demoPhi = float64(phi)
}

// CycleDemoIntensity cycles through demo intensity levels.
// Returns the new danger level string.
func (g *GRWarp) CycleDemoIntensity() string {
	// Cycle: Faint -> Subtle -> Strong -> Extreme -> Faint
	switch {
	case g.demoPhi < 0.0002:
		g.demoPhi = 0.0005 // Subtle
		return "Subtle"
	case g.demoPhi < 0.001:
		g.demoPhi = 0.005 // Strong
		return "Strong"
	case g.demoPhi < 0.01:
		g.demoPhi = 0.05 // Extreme
		return "Extreme"
	default:
		g.demoPhi = 0.0001 // Faint (barely perceptible)
		return "Faint"
	}
}

// GetDemoIntensity returns the current demo intensity as a string.
func (g *GRWarp) GetDemoIntensity() string {
	level := relativity.ClassifyDangerLevel(g.demoPhi)
	return level.String()
}

// SetSize updates render buffer dimensions.
func (g *GRWarp) SetSize(w, h int) {
	if g.screenW == w && g.screenH == h {
		return
	}
	g.screenW = w
	g.screenH = h
	g.buffer = ebiten.NewImage(w, h)
}

// Apply applies the GR lensing effect.
// Returns true if the effect was applied, false if skipped.
func (g *GRWarp) Apply(dst, src *ebiten.Image) bool {
	if !g.enabled {
		return false
	}

	shader, err := g.manager.Get("gr_lensing")
	if err != nil {
		return false
	}

	// Ensure buffer is sized
	w, h := src.Bounds().Dx(), src.Bounds().Dy()
	g.SetSize(w, h)

	// Build uniforms
	var uniforms map[string]any
	if g.demoMode {
		uniforms = g.buildDemoUniforms()
	} else if g.uniforms.Enabled {
		uniforms = g.buildSimUniforms()
	} else {
		return false // GR context not active
	}

	opts := &ebiten.DrawRectShaderOptions{}
	opts.Images[0] = src
	opts.Uniforms = uniforms

	dst.DrawRectShader(g.screenW, g.screenH, shader, opts)
	return true
}

// buildDemoUniforms creates uniforms for demo mode.
func (g *GRWarp) buildDemoUniforms() map[string]any {
	// Compute effect parameters from demo phi
	phi := float32(g.demoPhi)
	lensStrength := phi * 100 // Scale for visibility

	// Max effect radius depends on intensity
	var maxRadius float32
	switch {
	case g.demoPhi >= 0.01:
		maxRadius = 0.4 // 40% of screen for Extreme
	case g.demoPhi >= 0.001:
		maxRadius = 0.25 // 25% for Strong
	case g.demoPhi >= 0.0002:
		maxRadius = 0.15 // 15% for Subtle
	default:
		maxRadius = 0.08 // 8% for Faint (barely noticeable)
	}

	return map[string]any{
		"CenterX":         g.demoCenter[0],
		"CenterY":         g.demoCenter[1],
		"Rs":              g.demoRs,
		"Phi":             phi,
		"LensStrength":    lensStrength,
		"MaxEffectRadius": maxRadius,
		"RedshiftFactor":  float32(1.0 / (1.0 - phi*2)), // Approximate
	}
}

// buildSimUniforms creates uniforms from simulation context.
func (g *GRWarp) buildSimUniforms() map[string]any {
	u := g.uniforms
	return map[string]any{
		"CenterX":         u.ScreenCenter[0],
		"CenterY":         u.ScreenCenter[1],
		"Rs":              u.Rs,
		"Phi":             u.Phi,
		"LensStrength":    u.LensStrength,
		"MaxEffectRadius": u.MaxEffectRadius / float32(g.screenW), // Normalize
		"RedshiftFactor":  u.RedshiftFactor,
	}
}

// ApplyRedshift applies the GR redshift effect separately.
// Call this after lensing for layered effects.
func (g *GRWarp) ApplyRedshift(dst, src *ebiten.Image) bool {
	if !g.enabled {
		return false
	}

	shader, err := g.manager.Get("gr_redshift")
	if err != nil {
		return false
	}

	w, h := src.Bounds().Dx(), src.Bounds().Dy()
	g.SetSize(w, h)

	var uniforms map[string]any
	if g.demoMode {
		uniforms = g.buildDemoUniforms()
	} else if g.uniforms.Enabled {
		uniforms = g.buildSimUniforms()
	} else {
		return false
	}

	opts := &ebiten.DrawRectShaderOptions{}
	opts.Images[0] = src
	opts.Uniforms = uniforms

	dst.DrawRectShader(g.screenW, g.screenH, shader, opts)
	return true
}
