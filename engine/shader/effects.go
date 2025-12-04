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
	renderBuffer *ebiten.Image
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
	e.pipeline.SetSize(w, h)
	e.bloom.SetSize(w, h)
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

	// Apply bloom first (if enabled)
	if e.bloom.IsEnabled() {
		e.renderBuffer.Clear()
		if e.bloom.Apply(e.renderBuffer, current) {
			current = e.renderBuffer
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
		fmt.Sprintf("F5: Bloom         [%s]", boolToOnOff(e.bloom.IsEnabled())),
		fmt.Sprintf("F6: Vignette      [%s]", boolToOnOff(e.pipeline.IsEnabled("vignette"))),
		fmt.Sprintf("F7: CRT           [%s]", boolToOnOff(e.pipeline.IsEnabled("crt"))),
		fmt.Sprintf("F8: Aberration    [%s]", boolToOnOff(e.pipeline.IsEnabled("aberration"))),
		"",
		"F9: Toggle this overlay",
		"Shift+F5: Cycle bloom intensity",
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
