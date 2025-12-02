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
// AILANG: Text(string, float, float, int, int, int) -- text, x, y, fontSize, color, z
type DrawCmdText struct {
	Text     string
	X        float64
	Y        float64
	FontSize int // 0=small(10pt), 1=normal(14pt), 2=large(18pt), 3=title(24pt)
	Color    int // Color index (0 = default/white)
	Z        int
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
	Kind     UiKind  // Type of element (Panel, Button, Label, Portrait, Slider, ProgressBar)
	X        float64 // Left edge in normalized coords (0.0-1.0)
	Y        float64 // Top edge in normalized coords (0.0-1.0)
	W        float64 // Width in normalized coords
	H        float64 // Height in normalized coords
	Text     string  // Text content (if applicable)
	SpriteID int     // Sprite for icons/portraits
	Z        int     // Z-order within UI layer
	Color    int     // Background color
	Value    float64 // Value for sliders/progress bars (0.0-1.0)
}

func (DrawCmdUi) isDrawCmd() {}

// =============================================================================
// Extended Draw Commands (v0.4.5 Engine Extensions)
// =============================================================================

// DrawCmdLine draws a line between two points
// AILANG: Line(float, float, float, float, int, float, int) -- x1, y1, x2, y2, color, width, z
type DrawCmdLine struct {
	X1    float64
	Y1    float64
	X2    float64
	Y2    float64
	Color int
	Width float64 // Line thickness (1.0 = 1 pixel)
	Z     int
}

func (DrawCmdLine) isDrawCmd() {}

// DrawCmdTextWrapped draws text with word-wrapping
// AILANG: TextWrapped(string, float, float, float, int, int, int) -- text, x, y, maxWidth, fontSize, color, z
type DrawCmdTextWrapped struct {
	Text     string
	X        float64
	Y        float64
	MaxWidth float64 // Wrap at this width in pixels
	FontSize int     // 0=small(10pt), 1=normal(14pt), 2=large(18pt), 3=title(24pt)
	Color    int
	Z        int
}

func (DrawCmdTextWrapped) isDrawCmd() {}

// DrawCmdCircle draws a circle (filled or outline)
// AILANG: Circle(float, float, float, int, bool, int) -- x, y, radius, color, filled, z
type DrawCmdCircle struct {
	X      float64
	Y      float64
	Radius float64
	Color  int
	Filled bool // true=filled, false=outline
	Z      int
}

func (DrawCmdCircle) isDrawCmd() {}

// DrawCmdRectScreen draws a rectangle in screen space (not affected by camera)
// Used for backgrounds, overlays, etc.
type DrawCmdRectScreen struct {
	X     float64 // Screen X in pixels
	Y     float64 // Screen Y in pixels
	W     float64 // Width in pixels
	H     float64 // Height in pixels
	Color int
	Z     int
}

func (DrawCmdRectScreen) isDrawCmd() {}

// DrawCmdGalaxyBg renders the galaxy background image
// Opacity controls visibility (0.0 = invisible, 1.0 = fully visible)
// For sky view mode, ViewLon/ViewLat/FOV control which part of the image is shown
type DrawCmdGalaxyBg struct {
	Opacity float64 // 0.0 to 1.0
	Z       int
	// Sky view parameters (for equirectangular projection scrolling)
	SkyViewMode bool    // If true, use ViewLon/ViewLat/FOV for scrolling
	ViewLon     float64 // Galactic longitude we're looking at (0-360°)
	ViewLat     float64 // Galactic latitude we're looking at (-90 to +90°)
	FOV         float64 // Field of view in degrees
}

func (DrawCmdGalaxyBg) isDrawCmd() {}

// DrawCmdStar draws a star using a scaled sprite for efficient GPU batching
// Sprite IDs: 200=blue(O/B), 201=white(A/F), 202=yellow(G), 203=orange(K), 204=red(M)
type DrawCmdStar struct {
	X        float64 // Screen X position
	Y        float64 // Screen Y position
	SpriteID int     // Star sprite ID (200-204)
	Scale    float64 // Scale factor (1.0 = 16x16 pixels)
	Alpha    float64 // Opacity (0.0-1.0, default 1.0 if 0)
	Z        int
}

func (DrawCmdStar) isDrawCmd() {}
