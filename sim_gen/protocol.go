package sim_gen

// =============================================================================
// Isometric Draw Types
// =============================================================================

// IsoTile represents a tile to be drawn in isometric view
// Uses Coord from types.go
type IsoTile struct {
	Tile     Coord  // Tile position in grid
	Height   int    // Height level (0 = ground, 1+ = raised)
	SpriteID int    // Sprite to draw (0 = use colored rect)
	Layer    int    // Draw order hint (0=background, 100=entities, 200=UI)
}

// IsoEntity represents an entity to be drawn in isometric view
type IsoEntity struct {
	ID       string  // Entity identifier
	Tile     Coord   // Base tile position
	OffsetX  float64 // Sub-tile X offset (-0.5 to 0.5)
	OffsetY  float64 // Sub-tile Y offset (-0.5 to 0.5)
	Height   int     // Height level
	SpriteID int     // Sprite to draw
	Layer    int     // Draw order hint
}

// =============================================================================
// UI Element Types
// =============================================================================

// UiKind identifies the type of UI element
type UiKind int

const (
	UiKindPanel UiKind = iota
	UiKindButton
	UiKindLabel
	UiKindPortrait
	UiKindSlider      // For velocity selection, volume controls
	UiKindProgressBar // For journey progress, loading bars
)

// UiRect defines a rectangle in normalized screen coordinates (0.0-1.0)
type UiRect struct {
	X float64 // Left edge (0.0 = left, 1.0 = right)
	Y float64 // Top edge (0.0 = top, 1.0 = bottom)
	W float64 // Width (0.0-1.0)
	H float64 // Height (0.0-1.0)
}

// UiElement represents a UI element to be drawn
type UiElement struct {
	ID       string   // Element identifier (for click handling)
	Kind     UiKind   // Type of element
	Rect     UiRect   // Position in normalized coords
	Text     string   // Text content (if applicable)
	SpriteID int      // Sprite for icons/portraits
	Z        int      // Z-order within UI layer
	Color    int      // Background color (for panels)
}

// =============================================================================
// Click Event Types (for isometric input)
// =============================================================================

// ClickKind identifies mouse button
type ClickKind int

const (
	ClickLeft ClickKind = iota
	ClickRight
	ClickMiddle
)

// TileClick represents a click on an isometric tile
type TileClick struct {
	Tile  Coord
	Click ClickKind
}

// UiClick represents a click on a UI element
type UiClick struct {
	ElementID string
	Click     ClickKind
}

// =============================================================================
// Input/Output Types
// =============================================================================

// MouseState captures mouse position and button state
type MouseState struct {
	X       float64
	Y       float64
	Buttons []int
}

// KeyEvent represents a keyboard event
type KeyEvent struct {
	Key  int
	Kind string // "down" or "up"
}

// FrameInput is passed from engine to simulation each frame
type FrameInput struct {
	Mouse            MouseState
	Keys             []KeyEvent
	ClickedThisFrame bool         // True if left mouse button was just pressed
	WorldMouseX      float64      // Mouse X in world coordinates (after camera transform)
	WorldMouseY      float64      // Mouse Y in world coordinates (after camera transform)
	TileMouseX       int          // Mouse X in tile coordinates (isometric projection inverted)
	TileMouseY       int          // Mouse Y in tile coordinates (isometric projection inverted)
	ActionRequested  PlayerAction // Action triggered by keyboard (I=inspect, B=build, X=clear)
	TestMode         bool         // When true, strip UI for golden file testing
}

// Camera represents the viewport position and zoom level
type Camera struct {
	X    float64 // World position X (center of view)
	Y    float64 // World position Y (center of view)
	Zoom float64 // Zoom factor (1.0 = normal, 2.0 = zoomed in)
}

// FrameOutput is returned from simulation to engine each frame
type FrameOutput struct {
	Draw   []DrawCmd
	Sounds []int
	Debug  []string
	Camera Camera // Camera state for this frame
}
