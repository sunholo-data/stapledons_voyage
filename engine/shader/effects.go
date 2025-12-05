package shader

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// Effects manages all post-processing effects with demo mode support.
type Effects struct {
	manager      *Manager
	pipeline     *Pipeline
	bloom        *Bloom
	srWarp       *SRWarp
	grWarp       *GRWarp
	renderBuffer *ebiten.Image
	grBuffer     *ebiten.Image // Secondary buffer for GR effects
	screenW      int
	screenH      int
	demoMode     bool
	showOverlay  bool
}

// NewEffects creates a new effects controller.
func NewEffects() *Effects {
	manager := NewManager()

	e := &Effects{
		manager:  manager,
		pipeline: NewPipeline(manager),
		bloom:    NewBloom(manager),
		srWarp:   NewSRWarp(manager),
		grWarp:   NewGRWarp(manager),
	}

	// Setup default effects (all disabled initially)
	// Note: Higher values make effects more visible for testing
	e.pipeline.AddEffect("vignette", "vignette", map[string]any{
		"Intensity": float32(0.8),  // Strong edge darkening
		"Softness":  float32(0.2),  // Sharper falloff
	})
	e.pipeline.SetEnabled("vignette", false)

	e.pipeline.AddEffect("crt", "crt", map[string]any{
		"ScanlineIntensity": float32(0.3),  // More visible scanlines
		"Curvature":         float32(0.1),  // More barrel distortion
		"VignetteAmount":    float32(0.4),
	})
	e.pipeline.SetEnabled("crt", false)

	e.pipeline.AddEffect("aberration", "aberration", map[string]any{
		"Amount": float32(6.0),  // More visible RGB separation
	})
	e.pipeline.SetEnabled("aberration", false)

	return e
}

// Manager returns the underlying shader manager.
func (e *Effects) Manager() *Manager {
	return e.manager
}

// Pipeline returns the post-processing pipeline.
func (e *Effects) Pipeline() *Pipeline {
	return e.pipeline
}

// Bloom returns the bloom effect.
func (e *Effects) Bloom() *Bloom {
	return e.bloom
}

// SRWarp returns the special relativity warp effect.
func (e *Effects) SRWarp() *SRWarp {
	return e.srWarp
}

// GRWarp returns the general relativity warp effect.
func (e *Effects) GRWarp() *GRWarp {
	return e.grWarp
}

// Preload compiles all shaders at startup.
func (e *Effects) Preload() error {
	return e.manager.Preload()
}

// SetSize updates screen dimensions for all effects.
func (e *Effects) SetSize(w, h int) {
	if e.screenW == w && e.screenH == h {
		return
	}

	e.screenW = w
	e.screenH = h

	e.renderBuffer = ebiten.NewImage(w, h)
	e.grBuffer = ebiten.NewImage(w, h)
	e.pipeline.SetSize(w, h)
	e.bloom.SetSize(w, h)
	e.grWarp.SetSize(w, h)
}

// SetDemoMode enables or disables demo mode (F-key controls).
func (e *Effects) SetDemoMode(enabled bool) {
	e.demoMode = enabled
}

// IsDemoMode returns whether demo mode is enabled.
func (e *Effects) IsDemoMode() bool {
	return e.demoMode
}

// HandleInput processes demo mode input.
// Returns list of status messages for any toggles.
func (e *Effects) HandleInput() []string {
	if !e.demoMode {
		return nil
	}

	var messages []string

	// F3 = Toggle GR Warp (gravitational lensing near massive objects)
	if inpututil.IsKeyJustPressed(ebiten.KeyF3) {
		if ebiten.IsKeyPressed(ebiten.KeyShift) {
			// Shift+F3 = Cycle GR intensity in demo mode
			e.grWarp.SetDemoMode(0.5, 0.5, 0.05, 0.05) // Ensure demo mode is on
			level := e.grWarp.CycleDemoIntensity()
			messages = append(messages, fmt.Sprintf("GR Intensity: %s", level))
		} else {
			state := e.grWarp.Toggle()
			if state && !e.grWarp.IsDemoMode() {
				// Enable demo mode when toggling on without sim context
				e.grWarp.SetDemoMode(0.5, 0.5, 0.05, 0.005)
			}
			messages = append(messages, fmt.Sprintf("GR Warp: %v", boolToOnOff(state)))
		}
	}

	// F4 = Toggle SR Warp
	if inpututil.IsKeyJustPressed(ebiten.KeyF4) {
		state := e.srWarp.Toggle()
		messages = append(messages, fmt.Sprintf("SR Warp: %v", boolToOnOff(state)))
	}

	// Shift+F4 = Cycle SR velocity
	if ebiten.IsKeyPressed(ebiten.KeyShift) && inpututil.IsKeyJustPressed(ebiten.KeyF4) {
		beta := e.srWarp.GetBeta()
		// Cycle through 0.5, 0.9, 0.95, 0.99, 0.5...
		switch {
		case beta < 0.6:
			e.srWarp.SetForwardVelocity(0.9)
			messages = append(messages, "SR Velocity: 0.9c")
		case beta < 0.92:
			e.srWarp.SetForwardVelocity(0.95)
			messages = append(messages, "SR Velocity: 0.95c")
		case beta < 0.97:
			e.srWarp.SetForwardVelocity(0.99)
			messages = append(messages, "SR Velocity: 0.99c")
		default:
			e.srWarp.SetForwardVelocity(0.5)
			messages = append(messages, "SR Velocity: 0.5c")
		}
	}

	// F5 = Toggle Bloom
	if inpututil.IsKeyJustPressed(ebiten.KeyF5) {
		state := e.bloom.Toggle()
		messages = append(messages, fmt.Sprintf("Bloom: %v", boolToOnOff(state)))
	}

	// F6 = Toggle Vignette
	if inpututil.IsKeyJustPressed(ebiten.KeyF6) {
		state := e.pipeline.ToggleEffect("vignette")
		messages = append(messages, fmt.Sprintf("Vignette: %v", boolToOnOff(state)))
	}

	// F7 = Toggle CRT
	if inpututil.IsKeyJustPressed(ebiten.KeyF7) {
		state := e.pipeline.ToggleEffect("crt")
		messages = append(messages, fmt.Sprintf("CRT: %v", boolToOnOff(state)))
	}

	// F8 = Toggle Chromatic Aberration
	if inpututil.IsKeyJustPressed(ebiten.KeyF8) {
		state := e.pipeline.ToggleEffect("aberration")
		messages = append(messages, fmt.Sprintf("Chromatic Aberration: %v", boolToOnOff(state)))
	}

	// F9 = Toggle Overlay
	if inpututil.IsKeyJustPressed(ebiten.KeyF9) {
		e.showOverlay = !e.showOverlay
		messages = append(messages, fmt.Sprintf("Effect Overlay: %v", boolToOnOff(e.showOverlay)))
	}

	// Shift+F5 = Cycle bloom intensity
	if ebiten.IsKeyPressed(ebiten.KeyShift) && inpututil.IsKeyJustPressed(ebiten.KeyF5) {
		intensity := e.bloom.GetIntensity() + 0.2
		if intensity > 2.0 {
			intensity = 0.2
		}
		e.bloom.SetIntensity(intensity)
		messages = append(messages, fmt.Sprintf("Bloom Intensity: %.1f", intensity))
	}

	return messages
}

// Apply applies all enabled effects to the input and draws to screen.
func (e *Effects) Apply(screen *ebiten.Image, input *ebiten.Image) {
	// Ensure buffers are sized
	w, h := screen.Bounds().Dx(), screen.Bounds().Dy()
	e.SetSize(w, h)

	// Start with input
	current := input

	// Apply GR effects first (if enabled) - gravitational lensing near massive objects
	// GR happens "in the world" before SR velocity effects
	if e.grWarp.IsEnabled() {
		e.grBuffer.Clear()
		// Apply lensing
		if e.grWarp.Apply(e.grBuffer, current) {
			// Apply redshift on top of lensing
			e.renderBuffer.Clear()
			if e.grWarp.ApplyRedshift(e.renderBuffer, e.grBuffer) {
				current = e.renderBuffer
			} else {
				current = e.grBuffer
			}
		}
	}

	// Apply SR warp (if enabled) - this is the "view from the ship"
	// SR happens after GR (ship's velocity adds to gravitational effects)
	if e.srWarp.IsEnabled() {
		if current == e.renderBuffer {
			e.grBuffer.Clear()
			if e.srWarp.Apply(e.grBuffer, current) {
				current = e.grBuffer
			}
		} else {
			e.renderBuffer.Clear()
			if e.srWarp.Apply(e.renderBuffer, current) {
				current = e.renderBuffer
			}
		}
	}

	// Apply bloom (if enabled)
	if e.bloom.IsEnabled() {
		buf := ebiten.NewImage(w, h)
		if e.bloom.Apply(buf, current) {
			current = buf
		}
	}

	// Apply pipeline effects
	e.pipeline.Apply(screen, current)
}

// OverlayText returns text for the effect overlay.
func (e *Effects) OverlayText() []string {
	if !e.showOverlay {
		return nil
	}

	lines := []string{
		"=== Shader Effects Demo ===",
		"",
		fmt.Sprintf("F3: GR Warp       [%s]", boolToOnOff(e.grWarp.IsEnabled())),
		fmt.Sprintf("F4: SR Warp       [%s]", boolToOnOff(e.srWarp.IsEnabled())),
		fmt.Sprintf("F5: Bloom         [%s]", boolToOnOff(e.bloom.IsEnabled())),
		fmt.Sprintf("F6: Vignette      [%s]", boolToOnOff(e.pipeline.IsEnabled("vignette"))),
		fmt.Sprintf("F7: CRT           [%s]", boolToOnOff(e.pipeline.IsEnabled("crt"))),
		fmt.Sprintf("F8: Aberration    [%s]", boolToOnOff(e.pipeline.IsEnabled("aberration"))),
		"",
		"F9: Toggle this overlay",
		"Shift+F3: Cycle GR intensity",
		"Shift+F4: Cycle SR velocity",
		"Shift+F5: Cycle bloom intensity",
	}

	if e.grWarp.IsEnabled() {
		lines = append(lines, fmt.Sprintf("  GR: %s (demo mode)", e.grWarp.GetDemoIntensity()))
	}

	if e.srWarp.IsEnabled() {
		lines = append(lines, fmt.Sprintf("  SR: v=%.2fc gamma=%.2f", e.srWarp.GetBeta(), e.srWarp.GetGamma()))
	}

	if e.bloom.IsEnabled() {
		lines = append(lines, fmt.Sprintf("  Bloom: thresh=%.2f int=%.2f", e.bloom.GetThreshold(), e.bloom.GetIntensity()))
	}

	return lines
}

// ShowOverlay returns whether the effect overlay should be shown.
func (e *Effects) ShowOverlay() bool {
	return e.showOverlay
}

func boolToOnOff(b bool) string {
	if b {
		return "ON"
	}
	return "OFF"
}

// EnableAll enables all effects (for testing).
func (e *Effects) EnableAll() {
	e.bloom.SetEnabled(true)
	e.pipeline.SetEnabled("vignette", true)
	e.pipeline.SetEnabled("crt", false) // CRT conflicts with vignette
	e.pipeline.SetEnabled("aberration", true)
}

// DisableAll disables all effects.
func (e *Effects) DisableAll() {
	e.bloom.SetEnabled(false)
	e.pipeline.SetEnabled("vignette", false)
	e.pipeline.SetEnabled("crt", false)
	e.pipeline.SetEnabled("aberration", false)
}
