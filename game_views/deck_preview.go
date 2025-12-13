// Package game_views contains game-specific rendering helpers for Stapledon's Voyage.
package game_views

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// DeckPreview renders preview strips showing adjacent decks at screen edges.
// Deck above shows at top of screen, deck below at bottom.
type DeckPreview struct {
	screenW int
	screenH int

	// Preview dimensions
	previewHeight float32 // Height of preview strip
	fadeHeight    float32 // Height of fade gradient

	// Colors for deck previews
	deckColors [5]color.RGBA // Color scheme per deck
}

// NewDeckPreview creates a new deck preview renderer.
func NewDeckPreview(screenW, screenH int) *DeckPreview {
	dp := &DeckPreview{
		screenW:       screenW,
		screenH:       screenH,
		previewHeight: 80.0,
		fadeHeight:    30.0,
	}

	// Initialize deck color schemes (matching DeckInfo.colorScheme)
	// These are RGBA packed, converted to color.RGBA
	dp.deckColors = [5]color.RGBA{
		{R: 255, G: 51, B: 51, A: 100},   // Core - Red
		{R: 255, G: 136, B: 51, A: 100},  // Engineering - Orange
		{R: 51, G: 255, B: 136, A: 100},  // Culture - Green
		{R: 51, G: 136, B: 255, A: 100},  // Habitat - Blue
		{R: 136, G: 51, B: 255, A: 100},  // Bridge - Purple
	}

	return dp
}

// Resize updates dimensions when screen changes.
func (dp *DeckPreview) Resize(w, h int) {
	dp.screenW = w
	dp.screenH = h
}

// RenderPreviews draws deck previews at screen edges.
//
// currentDeck: Current deck index (0-4)
// transitionProgress: 0.0-1.0 during deck transition
// targetDeck: Target deck during transition
func (dp *DeckPreview) RenderPreviews(screen *ebiten.Image, currentDeck int, transitionProgress float64, targetDeck int) {
	// Calculate effective deck position
	effectiveDeck := float64(currentDeck)
	if transitionProgress > 0 {
		effectiveDeck += (float64(targetDeck) - float64(currentDeck)) * transitionProgress
	}

	// Preview deck above (shown at top of screen)
	deckAbove := currentDeck + 1
	if deckAbove < 5 {
		// Calculate preview opacity based on distance to deck above
		distanceAbove := float64(deckAbove) - effectiveDeck
		opacityAbove := dp.calculatePreviewOpacity(distanceAbove)
		dp.renderTopPreview(screen, deckAbove, opacityAbove)
	}

	// Preview deck below (shown at bottom of screen)
	deckBelow := currentDeck - 1
	if deckBelow >= 0 {
		// Calculate preview opacity based on distance to deck below
		distanceBelow := effectiveDeck - float64(deckBelow)
		opacityBelow := dp.calculatePreviewOpacity(distanceBelow)
		dp.renderBottomPreview(screen, deckBelow, opacityBelow)
	}
}

// calculatePreviewOpacity determines preview visibility based on deck distance.
func (dp *DeckPreview) calculatePreviewOpacity(distance float64) float64 {
	// Full opacity at distance 1, fading as we get closer or further
	if distance < 0.5 {
		return distance * 1.5 // Fade in as we approach
	} else if distance < 1.5 {
		return 1.0 - (distance-1.0)*0.5 // Full at 1.0, fade out beyond
	}
	return 0.3 // Minimum visibility
}

// renderTopPreview draws the deck-above preview at top of screen.
func (dp *DeckPreview) renderTopPreview(screen *ebiten.Image, deckIndex int, opacity float64) {
	if deckIndex < 0 || deckIndex >= 5 {
		return
	}

	baseColor := dp.deckColors[deckIndex]
	previewColor := color.RGBA{
		R: baseColor.R,
		G: baseColor.G,
		B: baseColor.B,
		A: uint8(float64(baseColor.A) * opacity),
	}

	// Draw preview strip at top
	vector.DrawFilledRect(
		screen,
		0,
		0,
		float32(dp.screenW),
		dp.previewHeight,
		previewColor,
		false,
	)

	// Draw fade gradient (darker at edge, transparent toward screen)
	dp.drawTopFadeGradient(screen, dp.previewHeight, opacity, deckIndex)

	// Draw deck label
	dp.drawDeckLabel(screen, deckIndex, 10, int(dp.previewHeight/2), opacity, true)
}

// renderBottomPreview draws the deck-below preview at bottom of screen.
func (dp *DeckPreview) renderBottomPreview(screen *ebiten.Image, deckIndex int, opacity float64) {
	if deckIndex < 0 || deckIndex >= 5 {
		return
	}

	baseColor := dp.deckColors[deckIndex]
	previewColor := color.RGBA{
		R: baseColor.R,
		G: baseColor.G,
		B: baseColor.B,
		A: uint8(float64(baseColor.A) * opacity),
	}

	// Draw preview strip at bottom
	y := float32(dp.screenH) - dp.previewHeight
	vector.DrawFilledRect(
		screen,
		0,
		y,
		float32(dp.screenW),
		dp.previewHeight,
		previewColor,
		false,
	)

	// Draw fade gradient
	dp.drawBottomFadeGradient(screen, y, opacity, deckIndex)

	// Draw deck label
	dp.drawDeckLabel(screen, deckIndex, 10, dp.screenH-int(dp.previewHeight/2), opacity, false)
}

// drawTopFadeGradient draws a gradient fading from preview into main screen.
func (dp *DeckPreview) drawTopFadeGradient(screen *ebiten.Image, startY float32, opacity float64, deckIndex int) {
	steps := 10
	stepHeight := dp.fadeHeight / float32(steps)
	baseColor := dp.deckColors[deckIndex]

	for i := 0; i < steps; i++ {
		alpha := float64(steps-i) / float64(steps) * opacity * 0.3
		gradColor := color.RGBA{
			R: baseColor.R,
			G: baseColor.G,
			B: baseColor.B,
			A: uint8(float64(baseColor.A) * alpha),
		}

		vector.DrawFilledRect(
			screen,
			0,
			startY+float32(i)*stepHeight,
			float32(dp.screenW),
			stepHeight,
			gradColor,
			false,
		)
	}
}

// drawBottomFadeGradient draws a gradient fading from preview into main screen.
func (dp *DeckPreview) drawBottomFadeGradient(screen *ebiten.Image, startY float32, opacity float64, deckIndex int) {
	steps := 10
	stepHeight := dp.fadeHeight / float32(steps)
	baseColor := dp.deckColors[deckIndex]

	for i := 0; i < steps; i++ {
		alpha := float64(steps-i) / float64(steps) * opacity * 0.3
		gradColor := color.RGBA{
			R: baseColor.R,
			G: baseColor.G,
			B: baseColor.B,
			A: uint8(float64(baseColor.A) * alpha),
		}

		vector.DrawFilledRect(
			screen,
			0,
			startY-dp.fadeHeight+float32(i)*stepHeight,
			float32(dp.screenW),
			stepHeight,
			gradColor,
			false,
		)
	}
}

// drawDeckLabel draws the deck name in the preview strip.
func (dp *DeckPreview) drawDeckLabel(screen *ebiten.Image, deckIndex int, x, y int, opacity float64, isAbove bool) {
	label := dp.getDeckName(deckIndex)
	arrow := "↑"
	if !isAbove {
		arrow = "↓"
	}

	// Draw indicator arrow and label
	// Note: Actual text rendering would need font integration
	// For now, draw a simple indicator rectangle
	labelColor := color.RGBA{
		R: 255,
		G: 255,
		B: 255,
		A: uint8(200 * opacity),
	}

	// Draw arrow indicator
	arrowSize := float32(16)
	if isAbove {
		// Up arrow at left side
		vector.DrawFilledRect(
			screen,
			float32(x),
			float32(y)-arrowSize/2,
			arrowSize,
			arrowSize,
			labelColor,
			false,
		)
	} else {
		// Down arrow at left side
		vector.DrawFilledRect(
			screen,
			float32(x),
			float32(y)-arrowSize/2,
			arrowSize,
			arrowSize,
			labelColor,
			false,
		)
	}

	_ = label // Would be used with font rendering
	_ = arrow
}

// getDeckName returns the display name for a deck index.
func (dp *DeckPreview) getDeckName(deckIndex int) string {
	switch deckIndex {
	case 0:
		return "Core"
	case 1:
		return "Engineering"
	case 2:
		return "Culture"
	case 3:
		return "Habitat"
	case 4:
		return "Bridge"
	default:
		return "Unknown"
	}
}

// SetDeckColor allows customizing the preview color for a deck.
func (dp *DeckPreview) SetDeckColor(deckIndex int, c color.RGBA) {
	if deckIndex >= 0 && deckIndex < 5 {
		dp.deckColors[deckIndex] = c
	}
}
