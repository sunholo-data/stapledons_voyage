package sim_gen

// DrawCmd is the interface for all draw commands (tagged union in AILANG)
type DrawCmd interface {
	isDrawCmd()
}

// DrawCmdRect draws a solid color rectangle
// AILANG: Rect(float, float, float, float, int, int) -- x, y, w, h, color, z
type DrawCmdRect struct {
	X     float64
	Y     float64
	W     float64
	H     float64
	Color int
	Z     int
}

func (DrawCmdRect) isDrawCmd() {}

// DrawCmdSprite draws a sprite by ID
// AILANG: Sprite(int, float, float, int) -- id, x, y, z
type DrawCmdSprite struct {
	ID int
	X  float64
	Y  float64
	Z  int
}

func (DrawCmdSprite) isDrawCmd() {}

// DrawCmdText draws text at a position
// AILANG: Text(string, float, float, int) -- text, x, y, z
type DrawCmdText struct {
	Text string
	X    float64
	Y    float64
	Z    int
}

func (DrawCmdText) isDrawCmd() {}

// =============================================================================
// Isometric Draw Commands
// =============================================================================

// DrawCmdIsoTile draws a tile in isometric view
// Engine converts tile coords to screen coords via isometric projection
type DrawCmdIsoTile struct {
	Tile     Coord // Tile position in grid
	Height   int   // Height level (0 = ground)
	SpriteID int   // Sprite to draw (0 = use colored rect)
	Layer    int   // Draw order hint (0=bg, 100=entities, 200=UI)
	Color    int   // Fallback color if no sprite
}

func (DrawCmdIsoTile) isDrawCmd() {}

// DrawCmdIsoEntity draws an entity in isometric view
// Supports sub-tile positioning for smooth movement
type DrawCmdIsoEntity struct {
	ID       string  // Entity identifier
	Tile     Coord   // Base tile position
	OffsetX  float64 // Sub-tile X offset (-0.5 to 0.5)
	OffsetY  float64 // Sub-tile Y offset (-0.5 to 0.5)
	Height   int     // Height level
	SpriteID int     // Sprite to draw
	Layer    int     // Draw order hint
}

func (DrawCmdIsoEntity) isDrawCmd() {}

// DrawCmdUi draws a UI element in screen space (not affected by camera)
type DrawCmdUi struct {
	ID       string  // Element identifier (for click handling)
	Kind     UiKind  // Type of element (Panel, Button, Label, Portrait)
	X        float64 // Left edge in normalized coords (0.0-1.0)
	Y        float64 // Top edge in normalized coords (0.0-1.0)
	W        float64 // Width in normalized coords
	H        float64 // Height in normalized coords
	Text     string  // Text content (if applicable)
	SpriteID int     // Sprite for icons/portraits
	Z        int     // Z-order within UI layer
	Color    int     // Background color
}

func (DrawCmdUi) isDrawCmd() {}
