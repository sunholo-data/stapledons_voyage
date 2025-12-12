// Package render provides the Higgs Spire renderer - the central visual anchor
// that runs through all decks of the ship.
package render

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// HiggsSpire renders the central spire that connects all ship decks.
// The spire serves as a visual anchor and indicator of current deck position.
type HiggsSpire struct {
	screenW int
	screenH int

	// Spire visual properties
	spireWidth   float32 // Width of spire column
	segmentCount int     // Number of deck segments (5)
	glowIntensity float64 // Overall glow brightness (0.0-1.0)

	// Colors
	spireBaseColor    color.RGBA // Dark metallic base
	spireGlowColor    color.RGBA // Active segment glow
	spireInactiveColor color.RGBA // Inactive segment color
}

// NewHiggsSpire creates a new Higgs Spire renderer.
func NewHiggsSpire(screenW, screenH int) *HiggsSpire {
	return &HiggsSpire{
		screenW:           screenW,
		screenH:           screenH,
		spireWidth:        24.0,
		segmentCount:      5,
		glowIntensity:     0.5,
		spireBaseColor:    color.RGBA{R: 40, G: 50, B: 60, A: 255},
		spireGlowColor:    color.RGBA{R: 100, G: 200, B: 255, A: 200},
		spireInactiveColor: color.RGBA{R: 60, G: 70, B: 80, A: 180},
	}
}

// Resize updates dimensions when screen changes.
func (hs *HiggsSpire) Resize(w, h int) {
	hs.screenW = w
	hs.screenH = h
}

// SetGlowIntensity sets the overall glow brightness.
func (hs *HiggsSpire) SetGlowIntensity(intensity float64) {
	if intensity < 0 {
		intensity = 0
	}
	if intensity > 1 {
		intensity = 1
	}
	hs.glowIntensity = intensity
}

// Render draws the Higgs Spire with current deck indication.
//
// currentDeck: Current deck index (0-4)
// transitionProgress: 0.0 = at currentDeck, 1.0 = at target
// targetDeck: Deck being transitioned to
func (hs *HiggsSpire) Render(screen *ebiten.Image, currentDeck int, transitionProgress float64, targetDeck int) {
	// Spire position (right edge of screen, vertically centered)
	spireX := float32(hs.screenW) - 50.0
	spireTop := float32(100)
	spireBottom := float32(hs.screenH - 100)
	segmentHeight := (spireBottom - spireTop) / float32(hs.segmentCount)

	// Draw spire base column
	vector.DrawFilledRect(
		screen,
		spireX-hs.spireWidth/2,
		spireTop,
		hs.spireWidth,
		spireBottom-spireTop,
		hs.spireBaseColor,
		false,
	)

	// Draw each segment
	for i := 0; i < hs.segmentCount; i++ {
		segmentY := spireTop + float32(i)*segmentHeight
		
		// Calculate segment activation based on current deck and transition
		// Segments are drawn top to bottom, deck indices are bottom to top
		// So segment 0 = Bridge (deck 4), segment 4 = Core (deck 0)
		deckForSegment := hs.segmentCount - 1 - i
		
		// Calculate activation level
		activation := hs.calculateActivation(deckForSegment, currentDeck, transitionProgress, targetDeck)
		
		// Draw segment
		hs.drawSegment(screen, spireX, segmentY, segmentHeight, activation)
	}

	// Draw spire frame/outline
	vector.StrokeRect(
		screen,
		spireX-hs.spireWidth/2-2,
		spireTop-2,
		hs.spireWidth+4,
		spireBottom-spireTop+4,
		2.0,
		color.RGBA{R: 80, G: 100, B: 120, A: 200},
		false,
	)
}

// calculateActivation determines how "lit up" a segment should be.
// Returns 0.0 (inactive) to 1.0 (fully active).
func (hs *HiggsSpire) calculateActivation(deckIndex, currentDeck int, transitionProgress float64, targetDeck int) float64 {
	// Distance from current deck position
	effectiveDeck := float64(currentDeck)
	if transitionProgress > 0 {
		effectiveDeck += (float64(targetDeck) - float64(currentDeck)) * transitionProgress
	}

	distance := math.Abs(float64(deckIndex) - effectiveDeck)

	// Current deck is fully active, adjacent are partially active
	switch {
	case distance < 0.1:
		return 1.0
	case distance < 1.1:
		return 0.6 - (distance-0.1)*0.4
	case distance < 2.1:
		return 0.2 - (distance-1.1)*0.1
	default:
		return 0.1
	}
}

// drawSegment draws a single spire segment with the given activation level.
func (hs *HiggsSpire) drawSegment(screen *ebiten.Image, x, y, height float32, activation float64) {
	// Segment dimensions (slightly smaller than spire width)
	segmentWidth := hs.spireWidth - 6
	padding := float32(3.0)

	// Interpolate between inactive and glow colors based on activation
	segColor := hs.interpolateColor(hs.spireInactiveColor, hs.spireGlowColor, activation*hs.glowIntensity)

	// Draw segment background
	vector.DrawFilledRect(
		screen,
		x-segmentWidth/2,
		y+padding,
		segmentWidth,
		height-padding*2,
		segColor,
		false,
	)

	// Add inner glow effect for active segments
	if activation > 0.5 {
		glowColor := hs.spireGlowColor
		glowColor.A = uint8(float64(100) * (activation - 0.5) * 2 * hs.glowIntensity)
		
		// Inner glow (smaller rectangle)
		innerWidth := segmentWidth - 8
		innerPadding := padding + 4
		vector.DrawFilledRect(
			screen,
			x-innerWidth/2,
			y+innerPadding,
			innerWidth,
			height-innerPadding*2,
			glowColor,
			false,
		)
	}

	// Draw segment border
	vector.StrokeRect(
		screen,
		x-segmentWidth/2,
		y+padding,
		segmentWidth,
		height-padding*2,
		1.0,
		color.RGBA{R: 100, G: 120, B: 140, A: 150},
		false,
	)
}

// interpolateColor linearly interpolates between two colors.
func (hs *HiggsSpire) interpolateColor(c1, c2 color.RGBA, t float64) color.RGBA {
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}

	return color.RGBA{
		R: uint8(float64(c1.R) + (float64(c2.R)-float64(c1.R))*t),
		G: uint8(float64(c1.G) + (float64(c2.G)-float64(c1.G))*t),
		B: uint8(float64(c1.B) + (float64(c2.B)-float64(c1.B))*t),
		A: uint8(float64(c1.A) + (float64(c2.A)-float64(c1.A))*t),
	}
}

// GetDeckLabel returns a display name for a deck index.
func (hs *HiggsSpire) GetDeckLabel(deckIndex int) string {
	switch deckIndex {
	case 0:
		return "CORE"
	case 1:
		return "ENGINEERING"
	case 2:
		return "CULTURE"
	case 3:
		return "HABITAT"
	case 4:
		return "BRIDGE"
	default:
		return "UNKNOWN"
	}
}
