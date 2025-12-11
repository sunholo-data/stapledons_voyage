// Package view provides UI helper types for rendering.
// This file contains UI layout helpers that are purely rendering-related (no game state).
package view

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// Z-index ranges for each layer type.
// Background: 0-9, Content: 10-99, UI: 100-199
const (
	ZBackground = 0
	ZContent    = 10
	ZUI         = 100
)

// Anchor defines UI element positioning.
type Anchor int

const (
	AnchorTopLeft Anchor = iota
	AnchorTopCenter
	AnchorTopRight
	AnchorCenterLeft
	AnchorCenter
	AnchorCenterRight
	AnchorBottomLeft
	AnchorBottomCenter
	AnchorBottomRight
)

// UIPanel represents a UI panel element.
type UIPanel struct {
	ID       string
	X, Y     float64 // Position (interpretation depends on Anchor)
	W, H     float64 // Size
	Anchor   Anchor  // Positioning anchor
	Visible  bool
	Z        int                                         // Z-index within UI layer
	DrawFunc func(screen *ebiten.Image, bounds Rect) // Custom draw function
}

// Rect represents a rectangle with position and size.
type Rect struct {
	X, Y, W, H float64
}

// Contains returns true if the point (px, py) is inside the rectangle.
func (r Rect) Contains(px, py float64) bool {
	return px >= r.X && px < r.X+r.W && py >= r.Y && py < r.Y+r.H
}

// Dialogue represents a dialogue overlay.
// TODO: This will be replaced by sim_gen.DialogueState when AILANG dialogue is implemented.
type Dialogue struct {
	Speaker string
	Text    string
	Options []DialogueOption
}

// DialogueOption represents a selectable dialogue option.
type DialogueOption struct {
	Text   string
	Action func() // Called when option is selected
}

// ComputePanelBounds calculates the actual screen bounds for a panel.
func ComputePanelBounds(panel *UIPanel, screenW, screenH float64) Rect {
	var x, y float64

	switch panel.Anchor {
	case AnchorTopLeft:
		x, y = panel.X, panel.Y
	case AnchorTopCenter:
		x, y = screenW/2-panel.W/2+panel.X, panel.Y
	case AnchorTopRight:
		x, y = screenW-panel.W-panel.X, panel.Y
	case AnchorCenterLeft:
		x, y = panel.X, screenH/2-panel.H/2+panel.Y
	case AnchorCenter:
		x, y = screenW/2-panel.W/2+panel.X, screenH/2-panel.H/2+panel.Y
	case AnchorCenterRight:
		x, y = screenW-panel.W-panel.X, screenH/2-panel.H/2+panel.Y
	case AnchorBottomLeft:
		x, y = panel.X, screenH-panel.H-panel.Y
	case AnchorBottomCenter:
		x, y = screenW/2-panel.W/2+panel.X, screenH-panel.H-panel.Y
	case AnchorBottomRight:
		x, y = screenW-panel.W-panel.X, screenH-panel.H-panel.Y
	default:
		x, y = panel.X, panel.Y
	}

	return Rect{X: x, Y: y, W: panel.W, H: panel.H}
}
