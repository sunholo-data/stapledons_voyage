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

	// Create test NPCs
	testNPCs := []NPC{
		{ID: 1, X: 10, Y: 10, Sprite: 0, Pattern: PatternRandomWalk{Interval: 30}, PatrolIndex: 0, MoveCounter: 30},
		{ID: 2, X: 20, Y: 15, Sprite: 1, Pattern: PatternRandomWalk{Interval: 45}, PatrolIndex: 0, MoveCounter: 45},
		{ID: 3, X: 30, Y: 20, Sprite: 2, Pattern: PatternStatic{}, PatrolIndex: 0, MoveCounter: 0},
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
	}
}

// TileSize is the pixel size of each tile
const TileSize = 8.0

// directionOffset returns the x, y offset for a direction
func directionOffset(dir Direction) (int, int) {
	switch dir {
	case North:
		return 0, -1
	case South:
		return 0, 1
	case East:
		return 1, 0
	case West:
		return -1, 0
	default:
		return 0, 0
	}
}

// isValidPosition checks if position is within world bounds
func isValidPosition(x, y, width, height int) bool {
	return x >= 0 && x < width && y >= 0 && y < height
}

// updateNPC processes a single NPC for one tick
func updateNPC(npc NPC, world World) NPC {
	switch p := npc.Pattern.(type) {
	case PatternStatic:
		return npc
	case PatternRandomWalk:
		return updateRandomWalk(npc, world, p.Interval)
	case PatternPatrol:
		return updatePatrol(npc, world, p.Path)
	default:
		return npc
	}
}

// updateRandomWalk moves NPC every N ticks in pseudo-random direction
func updateRandomWalk(npc NPC, world World, interval int) NPC {
	if npc.MoveCounter <= 0 {
		// Time to move! Pick direction based on tick + id (deterministic "random")
		dirIndex := (world.Tick + npc.ID) % 4
		dx, dy := directionOffset(Direction(dirIndex))
		newX, newY := npc.X+dx, npc.Y+dy

		if isValidPosition(newX, newY, world.Planet.Width, world.Planet.Height) {
			return NPC{
				ID:          npc.ID,
				X:           newX,
				Y:           newY,
				Sprite:      npc.Sprite,
				Pattern:     npc.Pattern,
				PatrolIndex: npc.PatrolIndex,
				MoveCounter: interval,
			}
		}
		// Blocked - reset counter but don't move
		return NPC{
			ID:          npc.ID,
			X:           npc.X,
			Y:           npc.Y,
			Sprite:      npc.Sprite,
			Pattern:     npc.Pattern,
			PatrolIndex: npc.PatrolIndex,
			MoveCounter: interval,
		}
	}
	// Decrement counter
	return NPC{
		ID:          npc.ID,
		X:           npc.X,
		Y:           npc.Y,
		Sprite:      npc.Sprite,
		Pattern:     npc.Pattern,
		PatrolIndex: npc.PatrolIndex,
		MoveCounter: npc.MoveCounter - 1,
	}
}

// updatePatrol follows a fixed path (placeholder for now)
func updatePatrol(npc NPC, world World, path []Direction) NPC {
	// TODO: Implement patrol logic
	return npc
}

// updateAllNPCs processes all NPCs for one tick
func updateAllNPCs(npcs []NPC, world World) []NPC {
	result := make([]NPC, len(npcs))
	for i, npc := range npcs {
		result[i] = updateNPC(npc, world)
	}
	return result
}

// Step advances the simulation by one frame.
// Returns new world state and frame output with draw commands.
func Step(world World, input FrameInput) (World, FrameOutput, error) {
	// Process selection on click
	newSelection := world.Selection
	if input.ClickedThisFrame {
		// Convert world coords to tile coords
		tileX := int(input.WorldMouseX / TileSize)
		tileY := int(input.WorldMouseY / TileSize)

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
	newPlanet := world.Planet // May be modified by build/clear

	if input.ActionRequested != nil {
		switch action := input.ActionRequested.(type) {
		case ActionInspect:
			if sel, ok := newSelection.(SelectionTile); ok {
				idx := sel.Y*world.Planet.Width + sel.X
				if idx >= 0 && idx < len(world.Planet.Tiles) {
					tile := world.Planet.Tiles[idx]
					structInfo := "Empty"
					if hs, ok := tile.Structure.(HasStructure); ok {
						structInfo = getStructureName(hs.Type)
					}
					msg := fmt.Sprintf("Tile (%d,%d): %s - %s", sel.X, sel.Y, getBiomeName(tile.Biome), structInfo)
					debugMessages = append(debugMessages, msg)
				}
			} else {
				debugMessages = append(debugMessages, "No tile selected")
			}

		case ActionBuild:
			if sel, ok := newSelection.(SelectionTile); ok {
				idx := sel.Y*world.Planet.Width + sel.X
				if idx >= 0 && idx < len(world.Planet.Tiles) {
					tile := world.Planet.Tiles[idx]
					if _, isEmpty := tile.Structure.(NoStructure); isEmpty {
						// Build on empty tile
						newTiles := make([]Tile, len(world.Planet.Tiles))
						copy(newTiles, world.Planet.Tiles)
						newTiles[idx] = Tile{
							Biome:     tile.Biome,
							Structure: HasStructure{Type: action.StructureType},
						}
						newPlanet = PlanetState{
							Width:  world.Planet.Width,
							Height: world.Planet.Height,
							Tiles:  newTiles,
						}
						debugMessages = append(debugMessages, fmt.Sprintf("Built %s at (%d,%d)", getStructureName(action.StructureType), sel.X, sel.Y))
					} else {
						debugMessages = append(debugMessages, "Tile already has a structure")
					}
				}
			} else {
				debugMessages = append(debugMessages, "No tile selected")
			}

		case ActionClear:
			if sel, ok := newSelection.(SelectionTile); ok {
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
						newPlanet = PlanetState{
							Width:  world.Planet.Width,
							Height: world.Planet.Height,
							Tiles:  newTiles,
						}
						debugMessages = append(debugMessages, fmt.Sprintf("Cleared structure at (%d,%d)", sel.X, sel.Y))
					} else {
						debugMessages = append(debugMessages, "No structure to clear")
					}
				}
			} else {
				debugMessages = append(debugMessages, "No tile selected")
			}
		}
	}

	// Increment tick and update NPCs
	newWorld := World{
		Tick:      world.Tick + 1,
		Planet:    newPlanet,
		NPCs:      world.NPCs,
		Selection: newSelection,
	}
	// Update all NPCs (needs newWorld for tick)
	newWorld.NPCs = updateAllNPCs(world.NPCs, newWorld)

	// Generate draw commands for tiles
	drawCmds := make([]DrawCmd, 0, len(newPlanet.Tiles)*2+1) // tiles + structures + possible highlight

	for i, tile := range newPlanet.Tiles {
		x := i % newPlanet.Width
		y := i / newPlanet.Width
		// Draw base tile
		cmd := DrawCmdRect{
			X:     float64(x) * TileSize,
			Y:     float64(y) * TileSize,
			W:     TileSize,
			H:     TileSize,
			Color: tile.Biome,
			Z:     0,
		}
		drawCmds = append(drawCmds, cmd)

		// Draw structure on top if present (colors 5-7)
		if hs, ok := tile.Structure.(HasStructure); ok {
			structCmd := DrawCmdRect{
				X:     float64(x)*TileSize + 1, // Slightly inset
				Y:     float64(y)*TileSize + 1,
				W:     TileSize - 2,
				H:     TileSize - 2,
				Color: 5 + int(hs.Type), // 5=House, 6=Farm, 7=Road
				Z:     1,                // On top of terrain
			}
			drawCmds = append(drawCmds, structCmd)
		}
	}

	// Draw NPCs on top of terrain and structures (Z=2)
	for _, npc := range newWorld.NPCs {
		npcCmd := DrawCmdRect{
			X:     float64(npc.X) * TileSize,
			Y:     float64(npc.Y) * TileSize,
			W:     TileSize,
			H:     TileSize,
			Color: 10 + npc.Sprite, // NPC colors start at index 10
			Z:     2,
		}
		drawCmds = append(drawCmds, npcCmd)
	}

	// Add selection highlight on top (Z=3, above NPCs)
	if sel, ok := newSelection.(SelectionTile); ok {
		highlightCmd := DrawCmdRect{
			X:     float64(sel.X) * TileSize,
			Y:     float64(sel.Y) * TileSize,
			W:     TileSize,
			H:     TileSize,
			Color: 4, // Yellow highlight (index 4 in biomeColors)
			Z:     3, // Draw on top of tiles, structures, and NPCs
		}
		drawCmds = append(drawCmds, highlightCmd)
	}

	// Calculate camera centered on world
	worldPixelW := float64(world.Planet.Width) * TileSize
	worldPixelH := float64(world.Planet.Height) * TileSize
	camera := Camera{
		X:    worldPixelW / 2,
		Y:    worldPixelH / 2,
		Zoom: 1.0,
	}

	output := FrameOutput{
		Draw:   drawCmds,
		Sounds: []int{},
		Debug:  debugMessages,
		Camera: camera,
	}

	return newWorld, output, nil
}
