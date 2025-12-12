// Package render provides viewport rendering with content compositing and edge blending.
package render

import (
	"stapledons_voyage/engine/shader"

	"github.com/hajimehoshi/ebiten/v2"
)

// ViewportContentType defines what content fills a viewport.
type ViewportContentType int

const (
	ContentNone ViewportContentType = iota
	ContentSpaceView                // Space background with optional SR/GR effects
	ContentStarfield                // Simple parallax starfield
	ContentSolid                    // Solid color fill
)

// ViewportEffectType defines effects that can be applied within a viewport.
type ViewportEffectType int

const (
	EffectNone ViewportEffectType = iota
	EffectSRWarp
	EffectGRLensing
	EffectTint
	EffectBlur
)

// ViewportContent describes what to render inside a viewport.
type ViewportContent struct {
	Type ViewportContentType

	// For ContentSpaceView
	Velocity  float64 // Ship velocity as fraction of c
	ViewAngle float64 // View direction (0 = forward)

	// For ContentStarfield
	Density float64 // Star density (0.0-1.0)
	Scroll  bool    // Whether to animate scrolling

	// For ContentSolid
	Color uint32 // RGBA color
}

// ViewportEffect describes an effect to apply within a viewport.
type ViewportEffect struct {
	Type      ViewportEffectType
	Intensity float64 // Effect strength (0.0-1.0)

	// For EffectSRWarp
	Velocity float64

	// For EffectTint
	TintColor uint32

	// For EffectBlur
	BlurRadius float64
}

// ViewportConfig describes a viewport to render.
type ViewportConfig struct {
	ID        string
	Shape     ViewportShape
	Content   ViewportContent
	Effects   []ViewportEffect
	Layer     int     // Z-order among viewports
	EdgeBlend float64 // 0.0 = hard edge, 1.0 = soft blend
	Opacity   float64 // 0.0-1.0
	ScreenX   float64 // Position on screen
	ScreenY   float64
}

// ViewportRenderer handles rendering content through shaped viewports.
type ViewportRenderer struct {
	shaderManager *shader.Manager
	contentBuffer *ebiten.Image
	maskBuffer    *ebiten.Image
	resultBuffer  *ebiten.Image
	edgeBlend     *ebiten.Shader
	srWarp        *shader.SRWarp
	screenW       int
	screenH       int
}

// NewViewportRenderer creates a new viewport renderer.
func NewViewportRenderer(shaderMgr *shader.Manager, screenW, screenH int) *ViewportRenderer {
	vr := &ViewportRenderer{
		shaderManager: shaderMgr,
		contentBuffer: ebiten.NewImage(screenW, screenH),
		maskBuffer:    ebiten.NewImage(screenW, screenH),
		resultBuffer:  ebiten.NewImage(screenW, screenH),
		screenW:       screenW,
		screenH:       screenH,
	}

	// Load edge blend shader
	if shaderMgr != nil {
		s, err := shaderMgr.Get("edge_blend")
		if err == nil {
			vr.edgeBlend = s
		}
	}

	return vr
}

// SetSRWarp sets the SR warp effect for ContentSpaceView.
func (vr *ViewportRenderer) SetSRWarp(srWarp *shader.SRWarp) {
	vr.srWarp = srWarp
}

// Resize updates buffer sizes.
func (vr *ViewportRenderer) Resize(w, h int) {
	if w != vr.screenW || h != vr.screenH {
		vr.screenW = w
		vr.screenH = h
		vr.contentBuffer = ebiten.NewImage(w, h)
		vr.maskBuffer = ebiten.NewImage(w, h)
		vr.resultBuffer = ebiten.NewImage(w, h)
		ClearMaskCache()
	}
}

// RenderViewport renders a single viewport and returns the composited result.
func (vr *ViewportRenderer) RenderViewport(cfg ViewportConfig, spaceDrawFunc func(*ebiten.Image)) *ebiten.Image {
	// Get viewport bounds
	bx, by, bw, bh := cfg.Shape.Bounds()
	w, h := int(bw), int(bh)
	if w <= 0 || h <= 0 {
		return nil
	}

	// Ensure buffers are large enough
	if w > vr.screenW || h > vr.screenH {
		vr.Resize(w, h)
	}

	// 1. Generate mask for the shape
	mask := cfg.Shape.GenerateMask(w, h)

	// 2. Clear and render content to buffer
	vr.contentBuffer.Clear()
	vr.renderContent(vr.contentBuffer, cfg.Content, bx, by, float64(w), float64(h), spaceDrawFunc)

	// 3. Apply viewport-specific effects
	for _, effect := range cfg.Effects {
		vr.applyEffect(vr.contentBuffer, effect, w, h)
	}

	// 4. Apply mask with edge blending
	result := ebiten.NewImage(w, h)
	vr.applyMaskWithBlend(result, vr.contentBuffer, mask, cfg.EdgeBlend, w, h)

	return result
}

// renderContent renders the viewport content to the buffer.
func (vr *ViewportRenderer) renderContent(dst *ebiten.Image, content ViewportContent, x, y, w, h float64, spaceDrawFunc func(*ebiten.Image)) {
	switch content.Type {
	case ContentSpaceView:
		// Use provided space drawing function
		if spaceDrawFunc != nil {
			spaceDrawFunc(dst)
		} else {
			// Fallback: draw dark space background
			vr.drawSimpleSpace(dst, int(w), int(h))
		}

	case ContentStarfield:
		vr.drawStarfield(dst, int(w), int(h), content.Density, content.Scroll)

	case ContentSolid:
		vr.drawSolid(dst, int(w), int(h), content.Color)

	case ContentNone:
		// Leave transparent
	}
}

// drawSimpleSpace draws a simple dark space background with stars.
func (vr *ViewportRenderer) drawSimpleSpace(dst *ebiten.Image, w, h int) {
	// Dark blue-black background
	dst.Fill(colorFromRGBA(0x0a0a1aff))

	// Add some simple stars
	vr.drawStarfield(dst, w, h, 0.5, false)
}

// drawStarfield draws a simple starfield.
func (vr *ViewportRenderer) drawStarfield(dst *ebiten.Image, w, h int, density float64, scroll bool) {
	// Simple procedural stars based on position
	// Use a deterministic pattern that looks random
	starCount := int(float64(w*h) * density * 0.001)
	if starCount > 200 {
		starCount = 200
	}

	for i := 0; i < starCount; i++ {
		// Simple hash for pseudo-random positions
		hash := uint32(i*1103515245 + 12345)
		x := float32(hash%uint32(w))
		hash = hash*1103515245 + 12345
		y := float32(hash % uint32(h))
		hash = hash*1103515245 + 12345

		// Vary brightness
		brightness := uint8(150 + (hash % 106))
		size := float32(1 + (hash%3)/2)

		// Draw star as small rectangle
		opts := &ebiten.DrawImageOptions{}
		opts.GeoM.Scale(float64(size), float64(size))
		opts.GeoM.Translate(float64(x), float64(y))
		opts.ColorScale.Scale(float32(brightness)/255, float32(brightness)/255, float32(brightness)/255, 1)
		dst.DrawImage(emptyImage, opts)
	}
}

// drawSolid fills with a solid color.
func (vr *ViewportRenderer) drawSolid(dst *ebiten.Image, w, h int, rgba uint32) {
	dst.Fill(colorFromRGBA(rgba))
}

// applyEffect applies a single effect to the buffer.
func (vr *ViewportRenderer) applyEffect(dst *ebiten.Image, effect ViewportEffect, w, h int) {
	switch effect.Type {
	case EffectSRWarp:
		// Apply SR warp shader within viewport bounds
		if vr.srWarp != nil && effect.Velocity > 0 {
			// Configure SR warp with the viewport's velocity
			vr.srWarp.SetForwardVelocity(effect.Velocity)
			vr.srWarp.SetEnabled(true)

			// Create temp buffer for warped content
			tempBuf := ebiten.NewImage(w, h)
			tempBuf.DrawImage(dst, nil)

			// Clear dst and apply warp
			dst.Clear()
			if !vr.srWarp.Apply(dst, tempBuf) {
				// If warp failed (velocity too low), just copy back
				dst.DrawImage(tempBuf, nil)
			}
		}

	case EffectTint:
		// Apply color tint overlay
		vr.applyTint(dst, effect.TintColor, effect.Intensity, w, h)

	case EffectBlur:
		// Apply blur (would need blur shader integration)
		// For now, this is a placeholder
	}
}

// applyTint applies a color tint to the buffer.
func (vr *ViewportRenderer) applyTint(dst *ebiten.Image, tintColor uint32, intensity float64, w, h int) {
	if intensity <= 0 {
		return
	}

	// Create tint overlay
	tint := ebiten.NewImage(w, h)
	tint.Fill(colorFromRGBA(tintColor))

	// Blend tint over content
	opts := &ebiten.DrawImageOptions{}
	opts.ColorScale.ScaleAlpha(float32(intensity))
	dst.DrawImage(tint, opts)
}

// applyMaskWithBlend applies the mask with edge blending.
func (vr *ViewportRenderer) applyMaskWithBlend(dst, content, mask *ebiten.Image, blendAmount float64, w, h int) {
	if vr.edgeBlend == nil {
		// Fallback: simple mask without blending
		vr.applyMaskSimple(dst, content, mask, w, h)
		return
	}

	// Use edge blend shader
	opts := &ebiten.DrawRectShaderOptions{}
	opts.Uniforms = map[string]interface{}{
		"BlendAmount": float32(blendAmount),
	}
	opts.Images[0] = content
	opts.Images[1] = mask
	dst.DrawRectShader(w, h, vr.edgeBlend, opts)
}

// applyMaskSimple applies mask without edge blending (fallback).
func (vr *ViewportRenderer) applyMaskSimple(dst, content, mask *ebiten.Image, w, h int) {
	// Draw content
	dst.DrawImage(content, nil)

	// Apply mask using composite mode
	opts := &ebiten.DrawImageOptions{}
	opts.Blend = ebiten.BlendDestinationIn
	dst.DrawImage(mask, opts)
}

// colorFromRGBA converts a packed RGBA uint32 to color.RGBA.
func colorFromRGBA(rgba uint32) *colorRGBA {
	return &colorRGBA{
		R: uint8((rgba >> 24) & 0xff),
		G: uint8((rgba >> 16) & 0xff),
		B: uint8((rgba >> 8) & 0xff),
		A: uint8(rgba & 0xff),
	}
}

// colorRGBA implements color.Color interface.
type colorRGBA struct {
	R, G, B, A uint8
}

func (c *colorRGBA) RGBA() (r, g, b, a uint32) {
	r = uint32(c.R) * 0x101
	g = uint32(c.G) * 0x101
	b = uint32(c.B) * 0x101
	a = uint32(c.A) * 0x101
	return
}
