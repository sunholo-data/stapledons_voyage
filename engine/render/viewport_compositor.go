// Package render provides viewport compositing with layer integration.
package render

import (
	"sort"

	"stapledons_voyage/engine/depth"
	"stapledons_voyage/engine/shader"

	"github.com/hajimehoshi/ebiten/v2"
)

// ViewportCompositor manages multiple viewports and composites them
// into the depth layer system.
type ViewportCompositor struct {
	renderer     *ViewportRenderer
	layerManager *DepthLayerManager
	viewports    []ViewportConfig
	screenW      int
	screenH      int
}

// NewViewportCompositor creates a new compositor.
func NewViewportCompositor(shaderMgr *shader.Manager, screenW, screenH int) *ViewportCompositor {
	return &ViewportCompositor{
		renderer:     NewViewportRenderer(shaderMgr, screenW, screenH),
		layerManager: NewDepthLayerManager(screenW, screenH),
		viewports:    make([]ViewportConfig, 0),
		screenW:      screenW,
		screenH:      screenH,
	}
}

// SetSRWarp sets the SR warp effect for space view content.
func (vc *ViewportCompositor) SetSRWarp(srWarp *shader.SRWarp) {
	vc.renderer.SetSRWarp(srWarp)
}

// SetViewports updates the list of viewports to render.
// Viewports are sorted by layer (z-order) automatically.
func (vc *ViewportCompositor) SetViewports(viewports []ViewportConfig) {
	vc.viewports = make([]ViewportConfig, len(viewports))
	copy(vc.viewports, viewports)

	// Sort by layer (lower layers render first, behind higher layers)
	sort.Slice(vc.viewports, func(i, j int) bool {
		return vc.viewports[i].Layer < vc.viewports[j].Layer
	})
}

// AddViewport adds a single viewport to the compositor.
func (vc *ViewportCompositor) AddViewport(cfg ViewportConfig) {
	vc.viewports = append(vc.viewports, cfg)
	// Re-sort after adding
	sort.Slice(vc.viewports, func(i, j int) bool {
		return vc.viewports[i].Layer < vc.viewports[j].Layer
	})
}

// ClearViewports removes all viewports.
func (vc *ViewportCompositor) ClearViewports() {
	vc.viewports = vc.viewports[:0]
}

// Resize updates buffer sizes.
func (vc *ViewportCompositor) Resize(w, h int) {
	if w != vc.screenW || h != vc.screenH {
		vc.screenW = w
		vc.screenH = h
		vc.renderer.Resize(w, h)
		vc.layerManager.Resize(w, h)
	}
}

// mapLayerToDepth maps a viewport layer (0-100) to a depth layer.
// Layer mapping:
//
//	0-24:   DeepBackground (space, distant stars)
//	25-49:  MidBackground (nebulae, far objects)
//	50-74:  Scene (main game content)
//	75-100: Foreground (UI, overlays)
func mapLayerToDepth(layer int) depth.Layer {
	switch {
	case layer < 25:
		return depth.LayerDeepBackground
	case layer < 50:
		return depth.LayerMidBackground
	case layer < 75:
		return depth.LayerScene
	default:
		return depth.LayerForeground
	}
}

// Composite renders all viewports and composites them to the target screen.
// spaceDrawFunc is called for ContentSpaceView viewports.
func (vc *ViewportCompositor) Composite(screen *ebiten.Image, spaceDrawFunc func(*ebiten.Image)) {
	// Clear all depth layer buffers
	vc.layerManager.Clear()

	// Render each viewport to its assigned depth layer
	for _, cfg := range vc.viewports {
		// Render the viewport content with masking
		result := vc.renderer.RenderViewport(cfg, spaceDrawFunc)
		if result == nil {
			continue
		}

		// Get the appropriate depth layer buffer
		depthLayer := mapLayerToDepth(cfg.Layer)
		layerBuf := vc.layerManager.GetBuffer(depthLayer)
		if layerBuf == nil {
			continue
		}

		// Draw the viewport result to the layer buffer at its screen position
		opts := &ebiten.DrawImageOptions{}
		opts.GeoM.Translate(cfg.ScreenX, cfg.ScreenY)
		opts.ColorScale.ScaleAlpha(float32(cfg.Opacity))
		layerBuf.DrawImage(result, opts)
	}

	// Composite all depth layers to the screen (back to front)
	vc.layerManager.Composite(screen)
}

// CompositeToLayers renders viewports to a provided DepthLayerManager.
// This allows integration with an existing layer system.
func (vc *ViewportCompositor) CompositeToLayers(layers *DepthLayerManager, spaceDrawFunc func(*ebiten.Image)) {
	// Render each viewport to its assigned depth layer
	for _, cfg := range vc.viewports {
		// Render the viewport content with masking
		result := vc.renderer.RenderViewport(cfg, spaceDrawFunc)
		if result == nil {
			continue
		}

		// Get the appropriate depth layer buffer
		depthLayer := mapLayerToDepth(cfg.Layer)
		layerBuf := layers.GetBuffer(depthLayer)
		if layerBuf == nil {
			continue
		}

		// Draw the viewport result to the layer buffer at its screen position
		opts := &ebiten.DrawImageOptions{}
		opts.GeoM.Translate(cfg.ScreenX, cfg.ScreenY)
		opts.ColorScale.ScaleAlpha(float32(cfg.Opacity))
		layerBuf.DrawImage(result, opts)
	}
}

// GetViewportAt returns the topmost viewport at the given screen coordinates.
// Returns nil if no viewport contains the point.
func (vc *ViewportCompositor) GetViewportAt(x, y float64) *ViewportConfig {
	// Check viewports in reverse order (highest layer first)
	for i := len(vc.viewports) - 1; i >= 0; i-- {
		cfg := &vc.viewports[i]
		// Translate to viewport-local coordinates
		localX := x - cfg.ScreenX
		localY := y - cfg.ScreenY
		if cfg.Shape.Contains(localX, localY) {
			return cfg
		}
	}
	return nil
}

// GetLayerManager returns the internal depth layer manager.
func (vc *ViewportCompositor) GetLayerManager() *DepthLayerManager {
	return vc.layerManager
}
