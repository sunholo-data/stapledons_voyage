package sim_gen

import "fmt"

// biomeNames maps biome index to human-readable name
var biomeNames = []string{"Water", "Forest", "Desert", "Mountain"}

// getBiomeName returns the name of a biome by index
func getBiomeName(biome int) string {
	if biome >= 0 && biome < len(biomeNames) {
		return biomeNames[biome]
	}
	return "Unknown"
}

// structureNames maps structure type to human-readable name
var structureNames = []string{"House", "Farm", "Road"}

// getStructureName returns the name of a structure type
func getStructureName(st StructureType) string {
	if int(st) >= 0 && int(st) < len(structureNames) {
		return structureNames[st]
	}
	return "Unknown"
}

// InitWorld creates a new world with the given seed.
// Mock implementation creates a 64x64 world with varied biomes.
func InitWorld(seed int64) World {
	width := 64
	height := 64
	tiles := make([]Tile, width*height)

	// Simple pseudo-random biome assignment based on seed and position
	for i := range tiles {
		x := i % width
		y := i / width
		// Simple hash for deterministic "randomness"
		hash := (x*31 + y*17 + int(seed)) % 4
		tiles[i] = Tile{Biome: hash, Structure: NoStructure{}}
	}

	// Create test NPCs with continuous movement (interval = interpolation frames)
	testNPCs := []NPC{
		{ID: 1, X: 10, Y: 10, Sprite: 0, Pattern: PatternRandomWalk{Interval: 10}, PatrolIndex: 0, MoveCounter: 10},
		{ID: 2, X: 20, Y: 15, Sprite: 1, Pattern: PatternRandomWalk{Interval: 12}, PatrolIndex: 0, MoveCounter: 12},
		{ID: 3, X: 30, Y: 20, Sprite: 2, Pattern: PatternStatic{}, PatrolIndex: 0, MoveCounter: 0},
		// Patrol NPC: walks in a 4x4 square pattern (clockwise)
		{ID: 4, X: 40, Y: 30, Sprite: 3, Pattern: PatternPatrol{
			Path:     []Direction{South, South, South, East, East, East, North, North, North, West, West, West},
			Interval: 10,
		}, PatrolIndex: 0, MoveCounter: 10},
	}

	return World{
		Tick: 0,
		Planet: PlanetState{
			Width:  width,
			Height: height,
			Tiles:  tiles,
		},
		NPCs:      testNPCs,
		Selection: SelectionNone{},
		Camera:    Camera{X: 0, Y: 0, Zoom: 1.0}, // Start centered on origin
		Mode:      ModeShipExploration{PlayerPos: Coord{X: 32, Y: 32}, CurrentDeck: 0},
	}
}

// Step advances the simulation by one frame.
// Returns new world state and frame output with draw commands.
func Step(world World, input FrameInput) (World, FrameOutput, error) {
	// Process mode switching first (also handles star selection clicks)
	newMode := processMode(world.Mode, input)

	// Process camera movement (only in ship exploration mode)
	newCamera := world.Camera
	if _, isShip := newMode.(ModeShipExploration); isShip {
		newCamera = updateCamera(world.Camera, input.Keys)
	}

	// Process selection on click
	newSelection := world.Selection
	if input.ClickedThisFrame {
		// Use pre-computed tile coords from isometric projection
		tileX := input.TileMouseX
		tileY := input.TileMouseY

		// Check if within world bounds
		if tileX >= 0 && tileX < world.Planet.Width &&
			tileY >= 0 && tileY < world.Planet.Height {
			newSelection = SelectionTile{X: tileX, Y: tileY}
		} else {
			newSelection = SelectionNone{}
		}
	}

	// Process action on current selection
	var debugMessages []string
	var sounds []int          // Sound IDs to play this frame
	newPlanet := world.Planet // May be modified by build/clear

	// Sound IDs
	const SoundSelect = 4

	// Play select sound when selection changes
	if input.ClickedThisFrame {
		if _, ok := newSelection.(SelectionTile); ok {
			sounds = append(sounds, SoundSelect)
		}
	}

	if input.ActionRequested != nil {
		newPlanet, debugMessages, sounds = processAction(
			input.ActionRequested, newSelection, world, newPlanet, debugMessages, sounds,
		)
	}

	// Increment tick and update NPCs
	newWorld := World{
		Tick:      world.Tick + 1,
		Planet:    newPlanet,
		NPCs:      world.NPCs,
		Selection: newSelection,
		Camera:    newCamera,
		Mode:      newMode,
	}
	// Update all NPCs (needs newWorld for tick)
	newWorld.NPCs = updateAllNPCs(world.NPCs, newWorld)

	// Generate draw commands based on current mode
	var drawCmds []DrawCmd

	switch m := newMode.(type) {
	case ModeShipExploration:
		drawCmds = renderShipExploration(newWorld, m, newPlanet, newSelection, newCamera, input.TestMode)
	case ModeGalaxyMap:
		drawCmds = renderGalaxyMap(newWorld, m, input.TestMode)
	default:
		// Fallback to ship exploration rendering
		drawCmds = renderShipExploration(newWorld, ModeShipExploration{}, newPlanet, newSelection, newCamera, input.TestMode)
	}

	output := FrameOutput{
		Draw:   drawCmds,
		Sounds: sounds,
		Debug:  debugMessages,
		Camera: newCamera,
	}

	return newWorld, output, nil
}

// processAction handles player actions (inspect, build, clear)
func processAction(action PlayerAction, selection Selection, world World, planet PlanetState, debugMessages []string, sounds []int) (PlanetState, []string, []int) {
	const (
		SoundClick = 1
		SoundBuild = 2
		SoundError = 3
	)

	switch a := action.(type) {
	case ActionInspect:
		if sel, ok := selection.(SelectionTile); ok {
			idx := sel.Y*world.Planet.Width + sel.X
			if idx >= 0 && idx < len(world.Planet.Tiles) {
				tile := world.Planet.Tiles[idx]
				structInfo := "Empty"
				if hs, ok := tile.Structure.(HasStructure); ok {
					structInfo = getStructureName(hs.Type)
				}
				msg := fmt.Sprintf("Tile (%d,%d): %s - %s", sel.X, sel.Y, getBiomeName(tile.Biome), structInfo)
				debugMessages = append(debugMessages, msg)
				sounds = append(sounds, SoundClick)
			}
		} else {
			debugMessages = append(debugMessages, "No tile selected")
			sounds = append(sounds, SoundError)
		}

	case ActionBuild:
		if sel, ok := selection.(SelectionTile); ok {
			idx := sel.Y*world.Planet.Width + sel.X
			if idx >= 0 && idx < len(world.Planet.Tiles) {
				tile := world.Planet.Tiles[idx]
				if _, isEmpty := tile.Structure.(NoStructure); isEmpty {
					// Build on empty tile
					newTiles := make([]Tile, len(world.Planet.Tiles))
					copy(newTiles, world.Planet.Tiles)
					newTiles[idx] = Tile{
						Biome:     tile.Biome,
						Structure: HasStructure{Type: a.StructureType},
					}
					planet = PlanetState{
						Width:  world.Planet.Width,
						Height: world.Planet.Height,
						Tiles:  newTiles,
					}
					debugMessages = append(debugMessages, fmt.Sprintf("Built %s at (%d,%d)", getStructureName(a.StructureType), sel.X, sel.Y))
					sounds = append(sounds, SoundBuild)
				} else {
					debugMessages = append(debugMessages, "Tile already has a structure")
					sounds = append(sounds, SoundError)
				}
			}
		} else {
			debugMessages = append(debugMessages, "No tile selected")
			sounds = append(sounds, SoundError)
		}

	case ActionClear:
		if sel, ok := selection.(SelectionTile); ok {
			idx := sel.Y*world.Planet.Width + sel.X
			if idx >= 0 && idx < len(world.Planet.Tiles) {
				tile := world.Planet.Tiles[idx]
				if _, isEmpty := tile.Structure.(NoStructure); !isEmpty {
					// Clear the structure
					newTiles := make([]Tile, len(world.Planet.Tiles))
					copy(newTiles, world.Planet.Tiles)
					newTiles[idx] = Tile{
						Biome:     tile.Biome,
						Structure: NoStructure{},
					}
					planet = PlanetState{
						Width:  world.Planet.Width,
						Height: world.Planet.Height,
						Tiles:  newTiles,
					}
					debugMessages = append(debugMessages, fmt.Sprintf("Cleared structure at (%d,%d)", sel.X, sel.Y))
					sounds = append(sounds, SoundClick)
				} else {
					debugMessages = append(debugMessages, "No structure to clear")
					sounds = append(sounds, SoundError)
				}
			}
		} else {
			debugMessages = append(debugMessages, "No tile selected")
			sounds = append(sounds, SoundError)
		}
	}

	return planet, debugMessages, sounds
}
