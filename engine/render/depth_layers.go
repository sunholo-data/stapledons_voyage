// Package render provides rendering utilities for the game.
package render

import (
	"stapledons_voyage/engine/depth"

	"github.com/hajimehoshi/ebiten/v2"
)

// Re-export layer types for convenience
type DepthLayer = depth.Layer

const (
	LayerDeepBackground = depth.LayerDeepBackground
	LayerMidBackground  = depth.LayerMidBackground
	LayerScene          = depth.LayerScene
	LayerForeground     = depth.LayerForeground
	LayerCount          = depth.LayerCount
)

// DepthLayerManager manages multiple rendering layers for parallax and transparency effects.
// Each layer is an off-screen buffer that can be composited together.
type DepthLayerManager struct {
	buffers [LayerCount]*ebiten.Image
	width   int
	height  int
}

// NewDepthLayerManager creates a new layer manager with buffers of the given size.
func NewDepthLayerManager(width, height int) *DepthLayerManager {
	m := &DepthLayerManager{
		width:  width,
		height: height,
	}
	for i := 0; i < int(LayerCount); i++ {
		m.buffers[i] = ebiten.NewImage(width, height)
	}
	return m
}

// Clear clears all layer buffers to transparent.
func (m *DepthLayerManager) Clear() {
	for _, buf := range m.buffers {
		buf.Clear()
	}
}

// GetBuffer returns the buffer for a specific layer.
// Draw operations should target this buffer.
func (m *DepthLayerManager) GetBuffer(layer DepthLayer) *ebiten.Image {
	if layer >= 0 && layer < LayerCount {
		return m.buffers[layer]
	}
	return nil
}

// Composite renders all layers to the target screen, back to front.
// Uses alpha blending so transparent areas show through to layers behind.
func (m *DepthLayerManager) Composite(screen *ebiten.Image) {
	for i := 0; i < int(LayerCount); i++ {
		opts := &ebiten.DrawImageOptions{}
		// Default blend mode is alpha blending, which is what we want
		screen.DrawImage(m.buffers[i], opts)
	}
}

// CompositeWithBlend renders all layers with custom blend options per layer.
func (m *DepthLayerManager) CompositeWithBlend(screen *ebiten.Image, layerAlpha [LayerCount]float64) {
	for i := 0; i < int(LayerCount); i++ {
		opts := &ebiten.DrawImageOptions{}
		opts.ColorScale.ScaleAlpha(float32(layerAlpha[i]))
		screen.DrawImage(m.buffers[i], opts)
	}
}

// Resize recreates all buffers at a new size.
func (m *DepthLayerManager) Resize(width, height int) {
	if m.width == width && m.height == height {
		return
	}
	m.width = width
	m.height = height

	// Dispose old buffers
	for i := range m.buffers {
		if m.buffers[i] != nil {
			m.buffers[i].Dispose()
		}
		m.buffers[i] = ebiten.NewImage(width, height)
	}
}

// Width returns the buffer width.
func (m *DepthLayerManager) Width() int {
	return m.width
}

// Height returns the buffer height.
func (m *DepthLayerManager) Height() int {
	return m.height
}

// IsEmpty returns true if a layer's buffer is completely transparent.
// Useful for skipping empty layers during compositing.
func (m *DepthLayerManager) IsEmpty(layer DepthLayer) bool {
	// For now, assume layers are non-empty if they exist
	// Could implement pixel checking if needed for optimization
	return false
}
