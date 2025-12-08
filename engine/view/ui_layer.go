package view

import (
	"sort"

	"github.com/hajimehoshi/ebiten/v2"
)

// BasicUILayer is a simple implementation of UILayer.
// It manages a collection of panels and an optional dialogue overlay.
type BasicUILayer struct {
	panels    map[string]*UIPanel
	dialogue  *Dialogue
	screenW   int
	screenH   int
}

// NewBasicUILayer creates a new UI layer.
func NewBasicUILayer(screenW, screenH int) *BasicUILayer {
	return &BasicUILayer{
		panels:  make(map[string]*UIPanel),
		screenW: screenW,
		screenH: screenH,
	}
}

// AddPanel adds a UI panel to the layer.
func (u *BasicUILayer) AddPanel(panel *UIPanel) {
	if panel != nil && panel.ID != "" {
		u.panels[panel.ID] = panel
	}
}

// RemovePanel removes a panel by ID.
func (u *BasicUILayer) RemovePanel(id string) {
	delete(u.panels, id)
}

// GetPanel returns a panel by ID, or nil if not found.
func (u *BasicUILayer) GetPanel(id string) *UIPanel {
	return u.panels[id]
}

// ShowDialogue displays a dialogue overlay.
func (u *BasicUILayer) ShowDialogue(dialogue *Dialogue) {
	u.dialogue = dialogue
}

// HideDialogue hides the current dialogue.
func (u *BasicUILayer) HideDialogue() {
	u.dialogue = nil
}

// HasDialogue returns true if a dialogue is being displayed.
func (u *BasicUILayer) HasDialogue() bool {
	return u.dialogue != nil
}

// GetDialogue returns the current dialogue, or nil if none.
func (u *BasicUILayer) GetDialogue() *Dialogue {
	return u.dialogue
}

// Draw renders all visible panels and the dialogue overlay.
func (u *BasicUILayer) Draw(screen *ebiten.Image) {
	screenW := float64(u.screenW)
	screenH := float64(u.screenH)

	// Get visible panels sorted by Z-index
	var visiblePanels []*UIPanel
	for _, panel := range u.panels {
		if panel.Visible {
			visiblePanels = append(visiblePanels, panel)
		}
	}
	sort.Slice(visiblePanels, func(i, j int) bool {
		return visiblePanels[i].Z < visiblePanels[j].Z
	})

	// Draw panels
	for _, panel := range visiblePanels {
		bounds := ComputePanelBounds(panel, screenW, screenH)
		if panel.DrawFunc != nil {
			panel.DrawFunc(screen, bounds)
		}
	}

	// Draw dialogue on top if present
	if u.dialogue != nil {
		u.drawDialogue(screen, screenW, screenH)
	}
}

// drawDialogue renders the dialogue overlay.
func (u *BasicUILayer) drawDialogue(screen *ebiten.Image, screenW, screenH float64) {
	// Default dialogue box: centered at bottom
	boxW := screenW * 0.8
	boxH := 120.0
	boxX := (screenW - boxW) / 2
	boxY := screenH - boxH - 20

	// Draw semi-transparent background
	for y := int(boxY); y < int(boxY+boxH); y++ {
		for x := int(boxX); x < int(boxX+boxW); x++ {
			screen.Set(x, y, colorWithAlpha(20, 20, 40, 220))
		}
	}

	// Draw border
	for x := int(boxX); x < int(boxX+boxW); x++ {
		screen.Set(x, int(boxY), colorWithAlpha(100, 100, 150, 255))
		screen.Set(x, int(boxY+boxH-1), colorWithAlpha(100, 100, 150, 255))
	}
	for y := int(boxY); y < int(boxY+boxH); y++ {
		screen.Set(int(boxX), y, colorWithAlpha(100, 100, 150, 255))
		screen.Set(int(boxX+boxW-1), y, colorWithAlpha(100, 100, 150, 255))
	}

	// Note: Text rendering would require a font system
	// For now, panels should provide their own DrawFunc
}

// HandleInput processes input events.
// Returns true if the input was consumed.
func (u *BasicUILayer) HandleInput(input *Input) bool {
	if input == nil {
		return false
	}

	// If dialogue is showing, it captures all input
	if u.dialogue != nil {
		// Check for dialogue option selection
		// This is a simplified implementation
		return true
	}

	// Check panels in reverse Z order (top to bottom)
	var visiblePanels []*UIPanel
	for _, panel := range u.panels {
		if panel.Visible {
			visiblePanels = append(visiblePanels, panel)
		}
	}
	sort.Slice(visiblePanels, func(i, j int) bool {
		return visiblePanels[i].Z > visiblePanels[j].Z // Reverse order
	})

	screenW := float64(u.screenW)
	screenH := float64(u.screenH)

	for _, panel := range visiblePanels {
		bounds := ComputePanelBounds(panel, screenW, screenH)
		if bounds.Contains(input.MouseX, input.MouseY) {
			if input.LeftClick {
				// Panel was clicked
				return true
			}
		}
	}

	return false
}

// Resize updates the layer's screen dimensions.
func (u *BasicUILayer) Resize(screenW, screenH int) {
	u.screenW = screenW
	u.screenH = screenH
}

// colorWithAlpha creates a color.Color with the given RGBA values.
type uiColor struct {
	R, G, B, A uint8
}

func colorWithAlpha(r, g, b, a uint8) uiColor {
	return uiColor{r, g, b, a}
}

func (c uiColor) RGBA() (r, g, b, a uint32) {
	r = uint32(c.R) * 0x101
	g = uint32(c.G) * 0x101
	b = uint32(c.B) * 0x101
	a = uint32(c.A) * 0x101
	return
}
