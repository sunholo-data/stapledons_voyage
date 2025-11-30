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
	ID          int
	X           int
	Y           int
	Sprite      int
	Pattern     MovementPattern
	PatrolIndex int
	MoveCounter int
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

// World is the complete game state
type World struct {
	Tick      int
	Planet    PlanetState
	NPCs      []NPC
	Selection Selection
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
