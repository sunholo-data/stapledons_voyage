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
		// Patrol NPC: walks in a 4x4 square pattern (clockwise)
		{ID: 4, X: 40, Y: 30, Sprite: 3, Pattern: PatternPatrol{
			Path:     []Direction{South, South, South, East, East, East, North, North, North, West, West, West},
			Interval: 20,
		}, PatrolIndex: 0, MoveCounter: 20},
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
	}
}

// TileSize is the pixel size of each tile
const TileSize = 8.0

// CameraSpeed is how fast the camera moves per frame (in world units)
const CameraSpeed = 4.0

// Ebiten key codes for camera movement (from ebiten.Key* constants)
const (
	KeyArrowUp    = 31
	KeyArrowDown  = 28
	KeyArrowLeft  = 29
	KeyArrowRight = 30
	KeyW          = 22
	KeyA          = 0
	KeyS          = 18
	KeyD          = 3
	KeyQ          = 16 // Zoom out
	KeyE          = 4  // Zoom in
)

// isKeyDown checks if a key is currently pressed
func isKeyDown(keys []KeyEvent, keyCode int) bool {
	for _, k := range keys {
		if k.Key == keyCode && k.Kind == "down" {
			return true
		}
	}
	return false
}

// updateCamera processes camera movement from input
func updateCamera(cam Camera, keys []KeyEvent) Camera {
	newX, newY := cam.X, cam.Y
	newZoom := cam.Zoom

	// Arrow keys or WASD for movement
	if isKeyDown(keys, KeyArrowUp) || isKeyDown(keys, KeyW) {
		newY -= CameraSpeed / cam.Zoom
	}
	if isKeyDown(keys, KeyArrowDown) || isKeyDown(keys, KeyS) {
		newY += CameraSpeed / cam.Zoom
	}
	if isKeyDown(keys, KeyArrowLeft) || isKeyDown(keys, KeyA) {
		newX -= CameraSpeed / cam.Zoom
	}
	if isKeyDown(keys, KeyArrowRight) || isKeyDown(keys, KeyD) {
		newX += CameraSpeed / cam.Zoom
	}

	// Q/E for zoom
	if isKeyDown(keys, KeyQ) {
		newZoom = cam.Zoom * 0.98 // Zoom out
		if newZoom < 0.25 {
			newZoom = 0.25
		}
	}
	if isKeyDown(keys, KeyE) {
		newZoom = cam.Zoom * 1.02 // Zoom in
		if newZoom > 4.0 {
			newZoom = 4.0
		}
	}

	return Camera{X: newX, Y: newY, Zoom: newZoom}
}

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
		return updatePatrol(npc, world, p.Path, p.Interval)
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

// updatePatrol follows a fixed path, looping back to start when complete
func updatePatrol(npc NPC, world World, path []Direction, interval int) NPC {
	// Empty path means static
	if len(path) == 0 {
		return npc
	}

	if npc.MoveCounter <= 0 {
		// Time to move! Get current direction from path
		dir := path[npc.PatrolIndex%len(path)]
		dx, dy := directionOffset(dir)
		newX, newY := npc.X+dx, npc.Y+dy

		// Advance to next path index (wrap around)
		nextIndex := (npc.PatrolIndex + 1) % len(path)

		if isValidPosition(newX, newY, world.Planet.Width, world.Planet.Height) {
			return NPC{
				ID:          npc.ID,
				X:           newX,
				Y:           newY,
				Sprite:      npc.Sprite,
				Pattern:     npc.Pattern,
				PatrolIndex: nextIndex,
				MoveCounter: interval,
			}
		}
		// Blocked - still advance index so patrol continues, reset counter
		return NPC{
			ID:          npc.ID,
			X:           npc.X,
			Y:           npc.Y,
			Sprite:      npc.Sprite,
			Pattern:     npc.Pattern,
			PatrolIndex: nextIndex,
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
	// Process camera movement
	newCamera := updateCamera(world.Camera, input.Keys)

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
	var sounds []int                // Sound IDs to play this frame
	newPlanet := world.Planet       // May be modified by build/clear

	// Sound IDs: 1=click, 2=build, 3=error, 4=select
	const (
		SoundClick  = 1
		SoundBuild  = 2
		SoundError  = 3
		SoundSelect = 4
	)

	// Play select sound when selection changes
	if input.ClickedThisFrame {
		if _, ok := newSelection.(SelectionTile); ok {
			sounds = append(sounds, SoundSelect)
		}
	}

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
					sounds = append(sounds, SoundClick)
				}
			} else {
				debugMessages = append(debugMessages, "No tile selected")
				sounds = append(sounds, SoundError)
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
	}

	// Increment tick and update NPCs
	newWorld := World{
		Tick:      world.Tick + 1,
		Planet:    newPlanet,
		NPCs:      world.NPCs,
		Selection: newSelection,
		Camera:    newCamera,
	}
	// Update all NPCs (needs newWorld for tick)
	newWorld.NPCs = updateAllNPCs(world.NPCs, newWorld)

	// Generate draw commands for tiles using isometric rendering
	drawCmds := make([]DrawCmd, 0, len(newPlanet.Tiles)*2+len(newWorld.NPCs)+1)

	for i, tile := range newPlanet.Tiles {
		x := i % newPlanet.Width
		y := i / newPlanet.Width
		// Draw base tile as isometric diamond
		cmd := DrawCmdIsoTile{
			Tile:     Coord{X: x, Y: y},
			Height:   0,
			SpriteID: 0, // Use colored fallback
			Layer:    0, // Ground layer
			Color:    tile.Biome,
		}
		drawCmds = append(drawCmds, cmd)

		// Draw structure on top if present
		if hs, ok := tile.Structure.(HasStructure); ok {
			structCmd := DrawCmdIsoTile{
				Tile:     Coord{X: x, Y: y},
				Height:   1, // Slightly elevated
				SpriteID: 0,
				Layer:    50, // Structure layer
				Color:    5 + int(hs.Type), // 5=House, 6=Farm, 7=Road
			}
			drawCmds = append(drawCmds, structCmd)
		}
	}

	// Draw NPCs as isometric entities
	// NPC.Sprite field (0-3) maps to SpriteID 100-103
	for _, npc := range newWorld.NPCs {
		npcCmd := DrawCmdIsoEntity{
			ID:       fmt.Sprintf("npc-%d", npc.ID),
			Tile:     Coord{X: npc.X, Y: npc.Y},
			OffsetX:  0,
			OffsetY:  0,
			Height:   0,
			SpriteID: 100 + npc.Sprite, // Map NPC sprite index to sprite ID 100+
			Layer:    100,              // Entity layer (above tiles)
		}
		drawCmds = append(drawCmds, npcCmd)
	}

	// Add selection highlight
	if sel, ok := newSelection.(SelectionTile); ok {
		highlightCmd := DrawCmdIsoTile{
			Tile:     Coord{X: sel.X, Y: sel.Y},
			Height:   0,
			SpriteID: 0,
			Layer:    150, // Highlight layer (above entities)
			Color:    4, // Yellow highlight
		}
		drawCmds = append(drawCmds, highlightCmd)
	}

	// Add UI panels only when not in test mode (test mode strips UI for golden files)
	if !input.TestMode {
		// Add UI panel showing camera position and controls
		cameraPanel := DrawCmdUi{
			ID:       "camera-info",
			Kind:     UiKindPanel,
			X:        0.02,
			Y:        0.02,
			W:        0.28,
			H:        0.06,
			Text:     fmt.Sprintf("Cam: (%.0f, %.0f) Zoom: %.2fx", newCamera.X, newCamera.Y, newCamera.Zoom),
			SpriteID: 0,
			Z:        0,
			Color:    3,
		}
		drawCmds = append(drawCmds, cameraPanel)

		// Controls help panel
		controlsPanel := DrawCmdUi{
			ID:       "controls-help",
			Kind:     UiKindPanel,
			X:        0.02,
			Y:        0.88,
			W:        0.45,
			H:        0.10,
			Text:     "WASD/Arrows: Move | Q/E: Zoom | Click: Select | I: Inspect | B: Build | X: Clear",
			SpriteID: 0,
			Z:        0,
			Color:    0, // Water blue
		}
		drawCmds = append(drawCmds, controlsPanel)
	}

	output := FrameOutput{
		Draw:   drawCmds,
		Sounds: sounds,
		Debug:  debugMessages,
		Camera: newCamera,
	}

	return newWorld, output, nil
}

