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

// BackgroundLayer renders behind everything.
// Typically starfields, nebulae, or gradients.
type BackgroundLayer interface {
	// SetParallax sets the parallax depth (0=static, 1=full camera motion).
	SetParallax(depth float64)

	// SetVelocity sets the ship velocity for SR aberration effects.
	SetVelocity(v float64)

	// SetGRIntensity sets the GR intensity for lensing effects.
	SetGRIntensity(intensity float64)

	// Draw renders the background to the screen.
	Draw(screen *ebiten.Image, camera *Camera)
}

// ContentLayer renders the main interactive content.
// This includes 3D planets, isometric tiles, and entities.
type ContentLayer interface {
	// Draw renders the content to the screen.
	Draw(screen *ebiten.Image, camera *Camera)

	// HandleInput processes input events.
	// Returns true if the input was consumed.
	HandleInput(input *Input) bool
}

// UILayer renders HUD elements, panels, and dialogue.
// Always rendered on top, not affected by camera transforms.
type UILayer interface {
	// AddPanel adds a UI panel to the layer.
	AddPanel(panel *UIPanel)

	// RemovePanel removes a panel by ID.
	RemovePanel(id string)

	// GetPanel returns a panel by ID, or nil if not found.
	GetPanel(id string) *UIPanel

	// ShowDialogue displays a dialogue overlay.
	ShowDialogue(dialogue *Dialogue)

	// HideDialogue hides the current dialogue.
	HideDialogue()

	// Draw renders the UI to the screen.
	Draw(screen *ebiten.Image)

	// HandleInput processes input events.
	// Returns true if the input was consumed.
	HandleInput(input *Input) bool
}

// Camera provides viewport transformation for layers.
type Camera struct {
	X, Y float64 // World position
	Zoom float64 // Zoom level (1.0 = normal)
}

// NewCamera creates a camera at the origin with default zoom.
func NewCamera() *Camera {
	return &Camera{
		X:    0,
		Y:    0,
		Zoom: 1.0,
	}
}

// Input represents input state for layers to process.
type Input struct {
	MouseX, MouseY  float64 // Screen coordinates
	WorldX, WorldY  float64 // World coordinates (transformed by camera)
	LeftClick       bool    // Left mouse button clicked this frame
	RightClick      bool    // Right mouse button clicked this frame
	KeysPressed     []ebiten.Key
	KeysJustPressed []ebiten.Key
}

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
	Z        int // Z-index within UI layer
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
