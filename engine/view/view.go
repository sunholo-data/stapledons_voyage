// Package view provides a composable view system for the game.
// Views are composed of three layers: Background, Content, and UI.
// The ViewManager handles transitions between views.
package view

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// ViewType identifies the type of view.
type ViewType int

const (
	ViewNone ViewType = iota
	ViewSpace         // Exterior space with planets
	ViewBridge        // Bridge interior (isometric)
	ViewShip          // Ship exploration (isometric)
	ViewGalaxyMap     // Star navigation
	ViewPlanetSurface // Ground exploration (isometric)
	ViewArrival       // Arrival sequence
)

// String returns a human-readable name for the view type.
func (v ViewType) String() string {
	names := []string{
		"None",
		"Space",
		"Bridge",
		"Ship",
		"GalaxyMap",
		"PlanetSurface",
		"Arrival",
	}
	if int(v) < len(names) {
		return names[v]
	}
	return "Unknown"
}

// View is the interface that all game views must implement.
// Views handle their own lifecycle and rendering through three layers.
type View interface {
	// Type returns the ViewType of this view.
	Type() ViewType

	// Init initializes the view. Called once when the view is first created.
	Init() error

	// Enter is called when transitioning into this view.
	// The from parameter indicates which view we're coming from.
	Enter(from ViewType)

	// Exit is called when transitioning out of this view.
	// The to parameter indicates which view we're going to.
	Exit(to ViewType)

	// Update updates the view state.
	// dt is delta time in seconds.
	// Returns a ViewTransition if the view wants to transition to another view.
	Update(dt float64) *ViewTransition

	// Draw renders the view to the screen.
	// Views should draw in layer order: Background, Content, UI.
	Draw(screen *ebiten.Image)

	// Layers returns the view's layer components for external access.
	Layers() ViewLayers
}

// ViewLayers holds references to a view's layer components.
// Note: These are interface{} because the concrete layer types vary by view.
// Most views return nil for layers they don't use.
type ViewLayers struct {
	Background interface{}
	Content    interface{}
	UI         interface{}
}

// ViewTransition describes a requested transition to another view.
type ViewTransition struct {
	To       ViewType
	Duration float64          // Transition duration in seconds
	Effect   TransitionEffect // Type of transition effect
}

// NewTransition creates a transition to another view with the given effect.
func NewTransition(to ViewType, duration float64, effect TransitionEffect) *ViewTransition {
	return &ViewTransition{
		To:       to,
		Duration: duration,
		Effect:   effect,
	}
}

// TransitionEffect defines the visual effect used when transitioning between views.
type TransitionEffect int

const (
	TransitionNone TransitionEffect = iota
	TransitionFade                  // Fade to black and back
	TransitionCrossfade             // Blend between views
	TransitionWipe                  // Directional wipe
	TransitionZoom                  // Zoom in/out
)

// String returns a human-readable name for the transition effect.
func (t TransitionEffect) String() string {
	names := []string{
		"None",
		"Fade",
		"Crossfade",
		"Wipe",
		"Zoom",
	}
	if int(t) < len(names) {
		return names[t]
	}
	return "Unknown"
}
