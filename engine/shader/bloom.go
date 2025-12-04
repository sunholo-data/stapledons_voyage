package shader

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// Bloom provides glow/bloom post-processing.
type Bloom struct {
	manager     *Manager
	threshold   float32
	intensity   float32
	blurPasses  int
	blurBuffer1 *ebiten.Image
	blurBuffer2 *ebiten.Image
	extractBuf  *ebiten.Image
	screenW     int
	screenH     int
	enabled     bool
}

// NewBloom creates a new bloom effect.
func NewBloom(manager *Manager) *Bloom {
	return &Bloom{
		manager:    manager,
		threshold:  0.4,  // Lower threshold = more glow
		intensity:  1.2,  // Stronger bloom
		blurPasses: 3,    // More blur passes
		enabled:    false,
	}
}

// SetThreshold sets the brightness threshold (0.0-1.0).
func (b *Bloom) SetThreshold(t float32) {
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	b.threshold = t
}

// GetThreshold returns the current threshold.
func (b *Bloom) GetThreshold() float32 {
	return b.threshold
}

// SetIntensity sets the bloom intensity (0.0-2.0).
func (b *Bloom) SetIntensity(i float32) {
	if i < 0 {
		i = 0
	}
	if i > 2 {
		i = 2
	}
	b.intensity = i
}

// GetIntensity returns the current intensity.
func (b *Bloom) GetIntensity() float32 {
	return b.intensity
}

// SetBlurPasses sets the number of blur passes (1-5).
func (b *Bloom) SetBlurPasses(p int) {
	if p < 1 {
		p = 1
	}
	if p > 5 {
		p = 5
	}
	b.blurPasses = p
}

// SetEnabled enables or disables the bloom effect.
func (b *Bloom) SetEnabled(enabled bool) {
	b.enabled = enabled
}

// IsEnabled returns whether bloom is enabled.
func (b *Bloom) IsEnabled() bool {
	return b.enabled
}

// Toggle toggles bloom on/off and returns new state.
func (b *Bloom) Toggle() bool {
	b.enabled = !b.enabled
	return b.enabled
}

// SetSize allocates/resizes render buffers.
func (b *Bloom) SetSize(w, h int) {
	if b.screenW == w && b.screenH == h {
		return
	}

	b.screenW = w
	b.screenH = h

	// All buffers at full resolution (simpler, works with DrawRectShader)
	b.blurBuffer1 = ebiten.NewImage(w, h)
	b.blurBuffer2 = ebiten.NewImage(w, h)
	b.extractBuf = ebiten.NewImage(w, h)
}

// Apply applies bloom effect to the input image.
// Returns true if bloom was applied, false if disabled or failed.
func (b *Bloom) Apply(dst, src *ebiten.Image) bool {
	if !b.enabled {
		return false
	}

	w, h := b.screenW, b.screenH
	if w < 1 || h < 1 {
		return false
	}

	// Step 1: Extract bright pixels
	extractShader, err := b.manager.Get("bloom_extract")
	if err != nil {
		return false
	}

	b.extractBuf.Clear()
	extractOpts := &ebiten.DrawRectShaderOptions{}
	extractOpts.Images[0] = src
	extractOpts.Uniforms = map[string]any{
		"Threshold": b.threshold,
	}
	b.extractBuf.DrawRectShader(w, h, extractShader, extractOpts)

	// Step 2: Blur (multiple passes, horizontal then vertical)
	blurShader, err := b.manager.Get("blur")
	if err != nil {
		return false
	}

	// Copy extract to blur buffer 1
	b.blurBuffer1.Clear()
	b.blurBuffer1.DrawImage(b.extractBuf, nil)

	for i := 0; i < b.blurPasses; i++ {
		// Horizontal pass
		b.blurBuffer2.Clear()
		hOpts := &ebiten.DrawRectShaderOptions{}
		hOpts.Images[0] = b.blurBuffer1
		hOpts.Uniforms = map[string]any{
			"Radius":    float32(4 + i*2),
			"Direction": [2]float32{1, 0},
		}
		b.blurBuffer2.DrawRectShader(w, h, blurShader, hOpts)

		// Vertical pass
		b.blurBuffer1.Clear()
		vOpts := &ebiten.DrawRectShaderOptions{}
		vOpts.Images[0] = b.blurBuffer2
		vOpts.Uniforms = map[string]any{
			"Radius":    float32(4 + i*2),
			"Direction": [2]float32{0, 1},
		}
		b.blurBuffer1.DrawRectShader(w, h, blurShader, vOpts)
	}

	// Step 3: Combine original + bloom
	combineShader, err := b.manager.Get("bloom_combine")
	if err != nil {
		return false
	}

	combineOpts := &ebiten.DrawRectShaderOptions{}
	combineOpts.Images[0] = src
	combineOpts.Images[1] = b.blurBuffer1
	combineOpts.Uniforms = map[string]any{
		"Intensity": b.intensity,
	}
	dst.DrawRectShader(w, h, combineShader, combineOpts)

	return true
}
