// Package game_views contains game-specific rendering helpers for Stapledon's Voyage.
package game_views

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// TransitionDirection indicates the direction of deck change.
type TransitionDirection int

const (
	TransitionNone TransitionDirection = iota
	TransitionUp                       // Going to higher deck (toward Bridge)
	TransitionDown                     // Going to lower deck (toward Core)
)

// DeckTransition handles smooth animated transitions between decks.
type DeckTransition struct {
	screenW int
	screenH int

	// Transition state
	active        bool
	fromDeck      int
	toDeck        int
	progress      float64 // 0.0 to 1.0
	duration      float64 // Total transition time in seconds
	elapsedTime   float64
	direction     TransitionDirection

	// Visual effects
	slideAmount   float64 // How far decks slide during transition
	fadeOverlay   bool    // Whether to use fade overlay
	overlayColor  color.RGBA

	// Buffers for transition effect
	transitionBuffer *ebiten.Image
}

// NewDeckTransition creates a new deck transition handler.
func NewDeckTransition(screenW, screenH int) *DeckTransition {
	return &DeckTransition{
		screenW:       screenW,
		screenH:       screenH,
		active:        false,
		duration:      0.5, // 0.5 second default
		slideAmount:   100.0,
		fadeOverlay:   true,
		overlayColor:  color.RGBA{R: 0, G: 0, B: 0, A: 128},
		transitionBuffer: ebiten.NewImage(screenW, screenH),
	}
}

// Resize updates dimensions when screen changes.
func (dt *DeckTransition) Resize(w, h int) {
	if w != dt.screenW || h != dt.screenH {
		dt.screenW = w
		dt.screenH = h
		if dt.transitionBuffer != nil {
			dt.transitionBuffer.Dispose()
		}
		dt.transitionBuffer = ebiten.NewImage(w, h)
	}
}

// StartTransition begins a new transition between decks.
func (dt *DeckTransition) StartTransition(fromDeck, toDeck int, duration float64) {
	if dt.active {
		return // Already transitioning
	}

	dt.active = true
	dt.fromDeck = fromDeck
	dt.toDeck = toDeck
	dt.progress = 0.0
	dt.elapsedTime = 0.0

	if duration > 0 {
		dt.duration = duration
	}

	if toDeck > fromDeck {
		dt.direction = TransitionUp
	} else if toDeck < fromDeck {
		dt.direction = TransitionDown
	} else {
		dt.direction = TransitionNone
	}
}

// Update advances the transition animation by delta time (in seconds).
// Returns true if transition is still active.
func (dt *DeckTransition) Update(deltaTime float64) bool {
	if !dt.active {
		return false
	}

	dt.elapsedTime += deltaTime
	dt.progress = dt.elapsedTime / dt.duration

	// Apply easing function (ease-in-out cubic)
	dt.progress = dt.easeInOutCubic(dt.progress)

	if dt.progress >= 1.0 {
		dt.progress = 1.0
		dt.active = false
		return false
	}

	return true
}

// IsActive returns whether a transition is currently in progress.
func (dt *DeckTransition) IsActive() bool {
	return dt.active
}

// GetProgress returns the current transition progress (0.0-1.0).
func (dt *DeckTransition) GetProgress() float64 {
	return dt.progress
}

// GetFromDeck returns the deck being transitioned from.
func (dt *DeckTransition) GetFromDeck() int {
	return dt.fromDeck
}

// GetToDeck returns the deck being transitioned to.
func (dt *DeckTransition) GetToDeck() int {
	return dt.toDeck
}

// GetDirection returns the transition direction.
func (dt *DeckTransition) GetDirection() TransitionDirection {
	return dt.direction
}

// ApplyTransitionEffect applies the transition visual effect to the screen.
// This should be called after rendering deck content.
func (dt *DeckTransition) ApplyTransitionEffect(screen *ebiten.Image) {
	if !dt.active || !dt.fadeOverlay {
		return
	}

	// Calculate fade intensity (peak at middle of transition)
	fadeIntensity := dt.calculateFadeIntensity()

	// Draw fade overlay
	overlayAlpha := uint8(float64(dt.overlayColor.A) * fadeIntensity)
	fadeColor := color.RGBA{
		R: dt.overlayColor.R,
		G: dt.overlayColor.G,
		B: dt.overlayColor.B,
		A: overlayAlpha,
	}

	vector.DrawFilledRect(
		screen,
		0,
		0,
		float32(dt.screenW),
		float32(dt.screenH),
		fadeColor,
		false,
	)

	// Draw direction indicator (subtle arrow showing movement)
	dt.drawDirectionIndicator(screen, fadeIntensity)
}

// calculateFadeIntensity returns fade strength based on progress.
// Peaks in the middle of transition, fades at start and end.
func (dt *DeckTransition) calculateFadeIntensity() float64 {
	// Bell curve: peaks at 0.5, zero at 0 and 1
	return math.Sin(dt.progress * math.Pi)
}

// GetSlideOffset returns the Y offset for slide effect based on progress.
func (dt *DeckTransition) GetSlideOffset() float64 {
	if !dt.active {
		return 0
	}

	// Slide direction based on transition direction
	var direction float64 = 1.0
	if dt.direction == TransitionDown {
		direction = -1.0
	}

	// Ease the slide amount
	slideProgress := dt.easeInOutCubic(dt.progress)
	return direction * slideProgress * dt.slideAmount
}

// drawDirectionIndicator draws a subtle indicator showing transition direction.
func (dt *DeckTransition) drawDirectionIndicator(screen *ebiten.Image, intensity float64) {
	if intensity < 0.1 {
		return
	}

	centerX := float32(dt.screenW / 2)
	arrowSize := float32(30)
	alpha := uint8(200 * intensity)

	arrowColor := color.RGBA{R: 255, G: 255, B: 255, A: alpha}

	if dt.direction == TransitionUp {
		// Draw up arrow at top-center
		y := float32(50)
		vector.StrokeLine(screen, centerX, y+arrowSize, centerX-arrowSize/2, y+arrowSize*1.5, 3, arrowColor, false)
		vector.StrokeLine(screen, centerX, y+arrowSize, centerX+arrowSize/2, y+arrowSize*1.5, 3, arrowColor, false)
	} else if dt.direction == TransitionDown {
		// Draw down arrow at bottom-center
		y := float32(dt.screenH - 50)
		vector.StrokeLine(screen, centerX, y-arrowSize, centerX-arrowSize/2, y-arrowSize*1.5, 3, arrowColor, false)
		vector.StrokeLine(screen, centerX, y-arrowSize, centerX+arrowSize/2, y-arrowSize*1.5, 3, arrowColor, false)
	}
}

// easeInOutCubic applies cubic ease-in-out to a value.
func (dt *DeckTransition) easeInOutCubic(t float64) float64 {
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}

	if t < 0.5 {
		return 4 * t * t * t
	}
	return 1 - math.Pow(-2*t+2, 3)/2
}

// SetDuration sets the transition duration in seconds.
func (dt *DeckTransition) SetDuration(duration float64) {
	if duration > 0 {
		dt.duration = duration
	}
}

// SetSlideAmount sets how far decks slide during transition.
func (dt *DeckTransition) SetSlideAmount(amount float64) {
	dt.slideAmount = amount
}

// SetFadeOverlay enables or disables the fade overlay effect.
func (dt *DeckTransition) SetFadeOverlay(enabled bool) {
	dt.fadeOverlay = enabled
}

// SetOverlayColor sets the color used for the fade overlay.
func (dt *DeckTransition) SetOverlayColor(c color.RGBA) {
	dt.overlayColor = c
}

// Cancel cancels any active transition immediately.
func (dt *DeckTransition) Cancel() {
	dt.active = false
	dt.progress = 0
	dt.elapsedTime = 0
}

// Complete immediately completes the transition.
func (dt *DeckTransition) Complete() {
	if dt.active {
		dt.progress = 1.0
		dt.active = false
	}
}
