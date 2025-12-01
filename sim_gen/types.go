// Package sim_gen contains mock types matching AILANG protocol.
// This file will be replaced by AILANG-generated code when ailang compile --emit-go ships.
package sim_gen

// Coord represents a 2D coordinate
type Coord struct {
	X int
	Y int
}

// Direction for NPC movement
type Direction int

const (
	North Direction = iota
	South
	East
	West
)

// MovementPattern defines how NPC moves (tagged union)
type MovementPattern interface {
	isMovementPattern()
}

// PatternStatic means NPC stays in place
type PatternStatic struct{}

func (PatternStatic) isMovementPattern() {}

// PatternRandomWalk moves every N ticks in pseudo-random direction
type PatternRandomWalk struct {
	Interval int // ticks between moves
}

func (PatternRandomWalk) isMovementPattern() {}

// PatternPatrol follows a fixed path, moving every Interval ticks
type PatternPatrol struct {
	Path     []Direction
	Interval int // ticks between moves
}

func (PatternPatrol) isMovementPattern() {}

// StructureType identifies what kind of structure to build
type StructureType int

const (
	StructureHouse StructureType = iota
	StructureFarm
	StructureRoad
)

// Structure represents what is built on a tile (tagged union)
type Structure interface {
	isStructure()
}

// NoStructure indicates no building on the tile
type NoStructure struct{}

func (NoStructure) isStructure() {}

// HasStructure indicates a building is present
type HasStructure struct {
	Type StructureType
}

func (HasStructure) isStructure() {}

// Tile represents a single tile with a biome and optional structure
type Tile struct {
	Biome     int
	Structure Structure
}

// PlanetState holds the world grid
type PlanetState struct {
	Width  int
	Height int
	Tiles  []Tile
}

// NPC represents a non-player character
type NPC struct {
	ID            int
	X             int     // Logical tile X position
	Y             int     // Logical tile Y position
	Sprite        int
	Pattern       MovementPattern
	PatrolIndex   int
	MoveCounter   int
	VisualOffsetX float64 // Visual offset from tile center (-1 to 1, interpolates toward 0)
	VisualOffsetY float64 // Visual offset from tile center (-1 to 1, interpolates toward 0)
}

// Selection represents the player's current selection state (tagged union)
type Selection interface {
	isSelection()
}

// SelectionNone indicates no tile is selected
type SelectionNone struct{}

func (SelectionNone) isSelection() {}

// SelectionTile indicates a specific tile is selected
type SelectionTile struct {
	X, Y int
}

func (SelectionTile) isSelection() {}

// =============================================================================
// UI Mode System (v0.5.0)
// =============================================================================

// WorldMode represents the current UI mode (tagged union)
type WorldMode interface {
	isWorldMode()
}

// ModeShipExploration - exploring the ship interior
type ModeShipExploration struct {
	PlayerPos     Coord
	CurrentDeck   int
	HoveredEntity string // Entity ID or empty
}

func (ModeShipExploration) isWorldMode() {}

// ModeGalaxyMap - viewing the galaxy map
type ModeGalaxyMap struct {
	CameraX      float64
	CameraY      float64
	ZoomLevel    float64
	SelectedStar int // -1 = none
	HoveredStar  int // -1 = none
	ShowNetwork  bool
}

func (ModeGalaxyMap) isWorldMode() {}

// ModeDialogue - in conversation
type ModeDialogue struct {
	SpeakerID    string
	Portrait     int
	CurrentText  string
	Choices      []DialogueChoice
	PreviousMode WorldMode // Mode to return to
}

func (ModeDialogue) isWorldMode() {}

// DialogueChoice represents a player dialogue option
type DialogueChoice struct {
	Text      string
	Available bool
	Tooltip   string
}

// ModeJourneyPlan - planning a journey
type ModeJourneyPlan struct {
	Destination     int     // Star ID
	Velocity        float64 // 0.9 to 0.999999
	SubjectiveTime  float64 // Years experienced
	ObjectiveTime   float64 // Years that pass
	Committed       bool
}

func (ModeJourneyPlan) isWorldMode() {}

// ModeCivDetail - viewing civilization details
type ModeCivDetail struct {
	CivID     int
	ActiveTab int // 0=Overview, 1=Philosophy, 2=Timeline, 3=Relationships, 4=Trade
}

func (ModeCivDetail) isWorldMode() {}

// ModeLegacy - endgame legacy visualization
type ModeLegacy struct {
	ActiveSection int // 0=Network, 1=Fates, 2=Philosophy, 3=Lineage, 4=Counterfactuals, 5=Epilogue
}

func (ModeLegacy) isWorldMode() {}

// World is the complete game state
type World struct {
	Tick      int
	Planet    PlanetState
	NPCs      []NPC
	Selection Selection
	Camera    Camera    // Camera position and zoom
	Mode      WorldMode // Current UI mode
}

// PlayerAction represents an action the player wants to perform (tagged union)
type PlayerAction interface {
	isPlayerAction()
}

// ActionNone indicates no action requested
type ActionNone struct{}

func (ActionNone) isPlayerAction() {}

// ActionInspect requests info about the selected tile
type ActionInspect struct{}

func (ActionInspect) isPlayerAction() {}

// ActionBuild requests placing a structure on the selected tile
type ActionBuild struct {
	StructureType StructureType
}

func (ActionBuild) isPlayerAction() {}

// ActionClear requests removing a structure from the selected tile
type ActionClear struct{}

func (ActionClear) isPlayerAction() {}

// ActionResult represents the outcome of an action (tagged union)
type ActionResult interface {
	isActionResult()
}

// ActionSuccess indicates the action succeeded
type ActionSuccess struct {
	Message string
}

func (ActionSuccess) isActionResult() {}

// ActionFailed indicates the action failed
type ActionFailed struct {
	Reason string
}

func (ActionFailed) isActionResult() {}

// ActionNoOp indicates no action was taken
type ActionNoOp struct{}

func (ActionNoOp) isActionResult() {}
