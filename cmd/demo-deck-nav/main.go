// Demo for multi-level ship deck navigation system.
// Shows the 5-deck structure with Higgs Spire, deck previews, and transitions.
//
// Usage:
//
//	go run ./cmd/demo-deck-nav
//	go run ./cmd/demo-deck-nav --screenshot 60 --output out/screenshots/deck-nav.png
//
// Controls:
//
//	PageUp/W   - Go to deck above
//	PageDown/S - Go to deck below
//	1-5        - Jump directly to deck
//	ESC/Q      - Quit
package main

import (
	"fmt"
	"image/color"
	"log"

	"stapledons_voyage/engine/demo"
	"stapledons_voyage/engine/render"
	"stapledons_voyage/game_views"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	screenWidth  = 1280
	screenHeight = 960
)

// DeckDemo demonstrates the multi-level ship system.
type DeckDemo struct {
	// Renderers
	spire      *render.HiggsSpire
	preview    *game_views.DeckPreview
	transition *game_views.DeckTransition

	// State
	currentDeck int
	targetDeck  int

	// Frame counter
	frameCount int
}

// NewDeckDemo creates a new deck navigation demo.
func NewDeckDemo() *DeckDemo {
	return &DeckDemo{
		spire:       render.NewHiggsSpire(screenWidth, screenHeight),
		preview:     game_views.NewDeckPreview(screenWidth, screenHeight),
		transition:  game_views.NewDeckTransition(screenWidth, screenHeight),
		currentDeck: 4, // Start on Bridge
		targetDeck:  4,
	}
}

// Update handles input and updates state.
func (d *DeckDemo) Update() error {
	d.frameCount++

	// Check for quit
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || inpututil.IsKeyJustPressed(ebiten.KeyQ) {
		return ebiten.Termination
	}

	// Update transition animation
	if d.transition.IsActive() {
		if !d.transition.Update(1.0 / 60.0) {
			// Transition complete
			d.currentDeck = d.targetDeck
		}
	} else {
		// Process navigation input only when not transitioning
		d.handleNavigation()
	}

	return nil
}

// handleNavigation processes deck navigation input.
func (d *DeckDemo) handleNavigation() {
	newDeck := d.currentDeck

	// PageUp/W - Go up
	if inpututil.IsKeyJustPressed(ebiten.KeyPageUp) || inpututil.IsKeyJustPressed(ebiten.KeyW) {
		if d.currentDeck < 4 {
			newDeck = d.currentDeck + 1
		}
	}

	// PageDown/S - Go down
	if inpututil.IsKeyJustPressed(ebiten.KeyPageDown) || inpututil.IsKeyJustPressed(ebiten.KeyS) {
		if d.currentDeck > 0 {
			newDeck = d.currentDeck - 1
		}
	}

	// Direct deck selection (1-5)
	for i := 0; i < 5; i++ {
		key := ebiten.Key(int(ebiten.Key1) + i)
		if inpututil.IsKeyJustPressed(key) {
			newDeck = i
		}
	}

	// Start transition if deck changed
	if newDeck != d.currentDeck {
		d.targetDeck = newDeck
		d.transition.StartTransition(d.currentDeck, newDeck, 0.5)
	}
}

// Draw renders the demo.
func (d *DeckDemo) Draw(screen *ebiten.Image) {
	// Dark background
	screen.Fill(color.RGBA{R: 15, G: 20, B: 30, A: 255})

	// Get transition state
	progress := d.transition.GetProgress()
	activeDeck := d.currentDeck
	if d.transition.IsActive() {
		activeDeck = d.currentDeck
	}

	// Draw deck content areas (simplified colored rectangles for demo)
	d.drawDeckContents(screen, activeDeck, progress)

	// Draw deck previews (adjacent decks at edges)
	d.preview.RenderPreviews(screen, d.currentDeck, progress, d.targetDeck)

	// Draw Higgs Spire
	d.spire.Render(screen, d.currentDeck, progress, d.targetDeck)

	// Apply transition effect (fade overlay)
	d.transition.ApplyTransitionEffect(screen)

	// Draw UI overlay
	d.drawUI(screen)
}

// drawDeckContents draws simplified deck content for the demo.
func (d *DeckDemo) drawDeckContents(screen *ebiten.Image, currentDeck int, progress float64) {
	// Deck colors
	deckColors := []color.RGBA{
		{R: 80, G: 30, B: 30, A: 255},  // Core - Dark red
		{R: 80, G: 50, B: 30, A: 255},  // Engineering - Dark orange
		{R: 30, G: 80, B: 50, A: 255},  // Culture - Dark green
		{R: 30, G: 50, B: 80, A: 255},  // Habitat - Dark blue
		{R: 50, G: 30, B: 80, A: 255},  // Bridge - Dark purple
	}

	// Deck names
	deckNames := []string{"CORE", "ENGINEERING", "CULTURE", "HABITAT", "BRIDGE"}

	// Calculate effective deck position
	effectiveDeck := float64(currentDeck)
	if d.transition.IsActive() {
		effectiveDeck += (float64(d.targetDeck) - float64(currentDeck)) * progress
	}

	// Draw each deck with parallax
	for i := 0; i < 5; i++ {
		distance := float64(i) - effectiveDeck
		yOffset := float32(distance * 50)

		// Calculate opacity based on distance
		absDistance := distance
		if absDistance < 0 {
			absDistance = -absDistance
		}

		var opacity float32
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

		// Draw deck background
		deckColor := deckColors[i]
		deckColor.A = uint8(float32(deckColor.A) * opacity)

		deckY := float32(screenHeight/2) - 150 + yOffset
		vector.DrawFilledRect(
			screen,
			100,
			deckY,
			float32(screenWidth-250),
			300,
			deckColor,
			false,
		)

		// Draw deck name if visible enough
		if opacity > 0.3 {
			nameColor := color.RGBA{R: 255, G: 255, B: 255, A: uint8(200 * opacity)}
			ebitenutil.DebugPrintAt(screen, deckNames[i], screenWidth/2-50, int(deckY+140))
			_ = nameColor // Would use for proper text rendering
		}

		// Draw some simple "content" rectangles to represent deck features
		d.drawDeckFeatures(screen, i, 100, deckY, opacity)
	}
}

// drawDeckFeatures draws simplified deck feature rectangles.
func (d *DeckDemo) drawDeckFeatures(screen *ebiten.Image, deckIndex int, x, y float32, opacity float32) {
	if opacity < 0.3 {
		return
	}

	// Different features per deck
	featureColors := []color.RGBA{
		{R: 255, G: 100, B: 100, A: uint8(150 * opacity)}, // Core - Red consoles
		{R: 255, G: 180, B: 100, A: uint8(150 * opacity)}, // Engineering - Orange machinery
		{R: 100, G: 255, B: 150, A: uint8(150 * opacity)}, // Culture - Green plants
		{R: 100, G: 150, B: 255, A: uint8(150 * opacity)}, // Habitat - Blue beds
		{R: 180, G: 100, B: 255, A: uint8(150 * opacity)}, // Bridge - Purple consoles
	}

	featureColor := featureColors[deckIndex]

	// Draw 3-5 feature rectangles based on deck
	featureCount := 3 + deckIndex%3
	for j := 0; j < featureCount; j++ {
		fx := x + 50 + float32(j*200)
		fy := y + 80
		fw := float32(80 + (deckIndex*10)%50)
		fh := float32(60)

		vector.DrawFilledRect(screen, fx, fy, fw, fh, featureColor, false)
	}
}

// drawUI draws the control instructions.
func (d *DeckDemo) drawUI(screen *ebiten.Image) {
	// Current deck info
	deckNames := []string{"Core", "Engineering", "Culture", "Habitat", "Bridge"}
	currentName := deckNames[d.currentDeck]

	status := fmt.Sprintf("Current Deck: %s (%d)", currentName, d.currentDeck+1)
	if d.transition.IsActive() {
		targetName := deckNames[d.targetDeck]
		progress := d.transition.GetProgress()
		status = fmt.Sprintf("Transitioning: %s → %s (%.0f%%)", currentName, targetName, progress*100)
	}

	// Draw info panel
	vector.DrawFilledRect(screen, 10, 10, 300, 120, color.RGBA{R: 0, G: 0, B: 0, A: 180}, false)

	ebitenutil.DebugPrintAt(screen, "Multi-Level Ship Demo", 20, 20)
	ebitenutil.DebugPrintAt(screen, status, 20, 40)
	ebitenutil.DebugPrintAt(screen, "", 20, 55)
	ebitenutil.DebugPrintAt(screen, "Controls:", 20, 70)
	ebitenutil.DebugPrintAt(screen, "  W/PageUp: Deck Up", 20, 85)
	ebitenutil.DebugPrintAt(screen, "  S/PageDown: Deck Down", 20, 100)
	ebitenutil.DebugPrintAt(screen, "  1-5: Direct Jump | Q: Quit", 20, 115)
}

// Layout returns the game's logical screen size.
func (d *DeckDemo) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	if err := demo.Run(NewDeckDemo(), demo.Config{
		Title:  "Stapledon's Voyage - Deck Navigation Demo",
		Width:  screenWidth,
		Height: screenHeight,
	}); err != nil && err != ebiten.Termination {
		log.Fatal(err)
	}
}
