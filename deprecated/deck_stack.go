// Package game_views contains game-specific rendering helpers for Stapledon's Voyage.
package game_views

import (
	"stapledons_voyage/engine/depth"
	"stapledons_voyage/engine/render"
	"stapledons_voyage/sim_gen"

	"github.com/hajimehoshi/ebiten/v2"
)

// DeckStackRenderer handles rendering multiple ship decks with parallax effect.
// The current deck renders at full detail while adjacent decks are visible
// with offset and opacity based on distance from current.
type DeckStackRenderer struct {
	layerManager *render.DepthLayerManager
	screenW      int
	screenH      int

	// Parallax configuration
	parallaxFactor float64 // How much decks offset per level (pixels)
	deckSpacing    float64 // Vertical spacing between deck levels

	// Pre-allocated buffers for deck rendering
	deckBuffers [5]*ebiten.Image // One buffer per deck (Core to Bridge)
}

// NewDeckStackRenderer creates a new deck stack renderer.
func NewDeckStackRenderer(screenW, screenH int) *DeckStackRenderer {
	dsr := &DeckStackRenderer{
		layerManager:   render.NewDepthLayerManager(screenW, screenH),
		screenW:        screenW,
		screenH:        screenH,
		parallaxFactor: 50.0,  // 50 pixels offset per deck level
		deckSpacing:    200.0, // 200 pixels between deck centers
	}

	// Create buffers for each deck
	for i := 0; i < 5; i++ {
		dsr.deckBuffers[i] = ebiten.NewImage(screenW, screenH)
	}

	return dsr
}

// Resize updates buffer sizes when screen changes.
func (dsr *DeckStackRenderer) Resize(w, h int) {
	if w != dsr.screenW || h != dsr.screenH {
		dsr.screenW = w
		dsr.screenH = h
		dsr.layerManager.Resize(w, h)

		// Recreate deck buffers
		for i := 0; i < 5; i++ {
			if dsr.deckBuffers[i] != nil {
				dsr.deckBuffers[i].Dispose()
			}
			dsr.deckBuffers[i] = ebiten.NewImage(w, h)
		}
	}
}

// GetLayerManager returns the internal depth layer manager.
func (dsr *DeckStackRenderer) GetLayerManager() *render.DepthLayerManager {
	return dsr.layerManager
}

// DeckRenderFunc is a function that renders a specific deck's content.
type DeckRenderFunc func(deckIndex int, buffer *ebiten.Image)

// RenderDeckStack renders all visible decks with parallax based on current deck
// and transition progress. Returns the composited result.
//
// currentDeck: The deck the player is currently on (0-4)
// transitionProgress: 0.0 = at currentDeck, 1.0 = fully at target
// targetDeck: The deck being transitioned to (same as current if not transitioning)
// renderDeck: Function to render each deck's content to its buffer
func (dsr *DeckStackRenderer) RenderDeckStack(
	currentDeck int,
	transitionProgress float64,
	targetDeck int,
	renderDeck DeckRenderFunc,
) *ebiten.Image {
	// Clear all deck buffers and layer manager
	dsr.layerManager.Clear()
	for i := 0; i < 5; i++ {
		dsr.deckBuffers[i].Clear()
	}

	// Calculate effective deck position (interpolated during transition)
	effectiveDeck := float64(currentDeck)
	if transitionProgress > 0 {
		effectiveDeck = float64(currentDeck) + (float64(targetDeck)-float64(currentDeck))*transitionProgress
	}

	// Render each deck with parallax offset
	for i := 0; i < 5; i++ {
		// Render deck content to its buffer
		renderDeck(i, dsr.deckBuffers[i])

		// Calculate distance from effective deck position
		distance := float64(i) - effectiveDeck

		// Calculate parallax offset (decks above move up, below move down)
		yOffset := distance * dsr.parallaxFactor

		// Calculate opacity (current deck = 1.0, adjacent = 0.5, further = 0.25)
		var opacity float64
		absDistance := distance
		if absDistance < 0 {
			absDistance = -absDistance
		}
		switch {
		case absDistance < 0.1:
			opacity = 1.0
		case absDistance < 1.1:
			opacity = 0.6
		case absDistance < 2.1:
			opacity = 0.3
		default:
			opacity = 0.1
		}

		// During transition, blend opacities
		if transitionProgress > 0 {
			// Fade out current, fade in target
			if i == currentDeck {
				opacity *= (1.0 - transitionProgress)
			} else if i == targetDeck {
				opacity = opacity*transitionProgress + 0.6*(1.0-transitionProgress)
			}
		}

		// Map deck to depth layer
		depthLayer := dsr.deckToDepthLayer(i)
		layerBuf := dsr.layerManager.GetBuffer(depthLayer)
		if layerBuf == nil {
			continue
		}

		// Draw deck buffer to layer with offset and opacity
		opts := &ebiten.DrawImageOptions{}
		opts.GeoM.Translate(0, yOffset)
		opts.ColorScale.ScaleAlpha(float32(opacity))
		layerBuf.DrawImage(dsr.deckBuffers[i], opts)
	}

	// Composite all layers to result
	result := ebiten.NewImage(dsr.screenW, dsr.screenH)
	dsr.layerManager.Composite(result)
	return result
}

// deckToDepthLayer maps a deck index to a depth layer.
// Core (0) = DeepBackground, Bridge (4) = Foreground
func (dsr *DeckStackRenderer) deckToDepthLayer(deckIndex int) depth.Layer {
	switch deckIndex {
	case 0: // Core
		return depth.LayerDeepBackground
	case 1: // Engineering
		return depth.LayerMidBackground
	case 2, 3: // Culture, Habitat
		return depth.LayerScene
	case 4: // Bridge
		return depth.LayerForeground
	default:
		return depth.LayerScene
	}
}

// GetDeckInfo returns info for a deck from AILANG.
func GetDeckInfo(deckIndex int) *sim_gen.DeckInfo {
	deckType := indexToDeckType(deckIndex)
	return sim_gen.GetDeckInfo(deckType)
}

// indexToDeckType converts deck index (0-4) to AILANG DeckType.
func indexToDeckType(index int) *sim_gen.DeckType {
	switch index {
	case 0:
		return sim_gen.NewDeckTypeDeckCore()
	case 1:
		return sim_gen.NewDeckTypeDeckEngineering()
	case 2:
		return sim_gen.NewDeckTypeDeckCulture()
	case 3:
		return sim_gen.NewDeckTypeDeckHabitat()
	case 4:
		return sim_gen.NewDeckTypeDeckBridge()
	default:
		return sim_gen.NewDeckTypeDeckBridge()
	}
}

// DeckTypeToIndex converts AILANG DeckType to index (0-4).
func DeckTypeToIndex(dt *sim_gen.DeckType) int {
	return int(sim_gen.DeckIndex(dt))
}
