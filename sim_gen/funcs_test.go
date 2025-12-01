package sim_gen

import "testing"

// TestInspectAction verifies inspect action shows tile info
func TestInspectAction(t *testing.T) {
	world := InitWorld(1234)

	// Select a tile first
	input := FrameInput{
		ClickedThisFrame: true,
		TileMouseX:       2, // Tile (2, 1)
		TileMouseY:       1,
	}
	world, _, _ = Step(world, input)

	// Now inspect
	input = FrameInput{
		ActionRequested: ActionInspect{},
	}
	_, output, err := Step(world, input)
	if err != nil {
		t.Fatalf("Step failed: %v", err)
	}

	if len(output.Debug) == 0 {
		t.Error("Expected debug message from inspect")
	}
}

// TestInspectNoSelection verifies inspect with no selection shows message
func TestInspectNoSelection(t *testing.T) {
	world := InitWorld(1234)

	input := FrameInput{
		ActionRequested: ActionInspect{},
	}
	_, output, _ := Step(world, input)

	if len(output.Debug) == 0 {
		t.Error("Expected 'No tile selected' message")
	}
	if len(output.Debug) > 0 && output.Debug[0] != "No tile selected" {
		t.Errorf("Expected 'No tile selected', got %q", output.Debug[0])
	}
}

// TestBuildOnEmptyTile verifies building on empty tile succeeds
func TestBuildOnEmptyTile(t *testing.T) {
	world := InitWorld(1234)

	// Select tile (5, 5)
	input := FrameInput{
		ClickedThisFrame: true,
		TileMouseX:       5,
		TileMouseY:       5,
	}
	world, _, _ = Step(world, input)

	// Build house
	input = FrameInput{
		ActionRequested: ActionBuild{StructureType: StructureHouse},
	}
	newWorld, output, _ := Step(world, input)

	// Check structure was placed
	idx := 5*64 + 5 // y * width + x
	tile := newWorld.Planet.Tiles[idx]
	if _, ok := tile.Structure.(HasStructure); !ok {
		t.Error("Expected structure to be built")
	}

	if len(output.Debug) == 0 || output.Debug[0] != "Built House at (5,5)" {
		t.Errorf("Expected build success message, got %v", output.Debug)
	}
}

// TestBuildOnOccupiedTile verifies building on occupied tile fails
func TestBuildOnOccupiedTile(t *testing.T) {
	world := InitWorld(1234)

	// Select and build first
	input := FrameInput{
		ClickedThisFrame: true,
		TileMouseX:       5,
		TileMouseY:       5,
	}
	world, _, _ = Step(world, input)

	input = FrameInput{
		ActionRequested: ActionBuild{StructureType: StructureHouse},
	}
	world, _, _ = Step(world, input)

	// Try to build again
	input = FrameInput{
		ActionRequested: ActionBuild{StructureType: StructureHouse},
	}
	_, output, _ := Step(world, input)

	if len(output.Debug) == 0 || output.Debug[0] != "Tile already has a structure" {
		t.Errorf("Expected 'Tile already has a structure', got %v", output.Debug)
	}
}

// TestClearStructure verifies clearing a structure works
func TestClearStructure(t *testing.T) {
	world := InitWorld(1234)

	// Select, build, then clear
	input := FrameInput{
		ClickedThisFrame: true,
		TileMouseX:       5,
		TileMouseY:       5,
	}
	world, _, _ = Step(world, input)

	input = FrameInput{
		ActionRequested: ActionBuild{StructureType: StructureHouse},
	}
	world, _, _ = Step(world, input)

	input = FrameInput{
		ActionRequested: ActionClear{},
	}
	newWorld, output, _ := Step(world, input)

	// Check structure was cleared
	idx := 5*64 + 5
	tile := newWorld.Planet.Tiles[idx]
	if _, ok := tile.Structure.(NoStructure); !ok {
		t.Error("Expected structure to be cleared")
	}

	if len(output.Debug) == 0 || output.Debug[0] != "Cleared structure at (5,5)" {
		t.Errorf("Expected clear success message, got %v", output.Debug)
	}
}

// TestClearEmptyTile verifies clearing empty tile shows message
func TestClearEmptyTile(t *testing.T) {
	world := InitWorld(1234)

	// Select tile
	input := FrameInput{
		ClickedThisFrame: true,
		TileMouseX:       5,
		TileMouseY:       5,
	}
	world, _, _ = Step(world, input)

	// Try to clear (no structure)
	input = FrameInput{
		ActionRequested: ActionClear{},
	}
	_, output, _ := Step(world, input)

	if len(output.Debug) == 0 || output.Debug[0] != "No structure to clear" {
		t.Errorf("Expected 'No structure to clear', got %v", output.Debug)
	}
}

// TestActionWithNoSelection verifies actions fail gracefully without selection
func TestActionWithNoSelection(t *testing.T) {
	world := InitWorld(1234)

	tests := []struct {
		name   string
		action PlayerAction
	}{
		{"Build", ActionBuild{StructureType: StructureHouse}},
		{"Clear", ActionClear{}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			input := FrameInput{
				ActionRequested: tc.action,
			}
			_, output, _ := Step(world, input)

			if len(output.Debug) == 0 || output.Debug[0] != "No tile selected" {
				t.Errorf("Expected 'No tile selected', got %v", output.Debug)
			}
		})
	}
}

// TestNPCInitialization verifies NPCs are created in InitWorld
func TestNPCInitialization(t *testing.T) {
	world := InitWorld(1234)

	if len(world.NPCs) != 4 {
		t.Errorf("Expected 4 NPCs, got %d", len(world.NPCs))
	}

	// Check first NPC
	if world.NPCs[0].ID != 1 || world.NPCs[0].X != 10 || world.NPCs[0].Y != 10 {
		t.Errorf("NPC 1 has wrong position: got (%d,%d)", world.NPCs[0].X, world.NPCs[0].Y)
	}

	// Check patrol NPC exists
	if _, ok := world.NPCs[3].Pattern.(PatternPatrol); !ok {
		t.Error("Expected NPC 4 to have PatternPatrol")
	}
}

// TestStaticNPCDoesntMove verifies static NPCs stay in place
func TestStaticNPCDoesntMove(t *testing.T) {
	world := InitWorld(1234)

	// NPC 3 is static at (30, 20)
	staticNPC := world.NPCs[2]
	if _, ok := staticNPC.Pattern.(PatternStatic); !ok {
		t.Fatal("Expected NPC 3 to have PatternStatic")
	}

	// Run 100 steps
	for i := 0; i < 100; i++ {
		world, _, _ = Step(world, FrameInput{})
	}

	// Check position unchanged
	if world.NPCs[2].X != 30 || world.NPCs[2].Y != 20 {
		t.Errorf("Static NPC moved from (30,20) to (%d,%d)", world.NPCs[2].X, world.NPCs[2].Y)
	}
}

// TestRandomWalkMovement verifies random walk NPCs move over time
func TestRandomWalkMovement(t *testing.T) {
	world := InitWorld(1234)

	// NPC 1 at (10, 10) with interval 30
	initialX, initialY := world.NPCs[0].X, world.NPCs[0].Y

	// Run 60 steps (enough for at least 1 move)
	for i := 0; i < 60; i++ {
		world, _, _ = Step(world, FrameInput{})
	}

	// NPC should have moved
	if world.NPCs[0].X == initialX && world.NPCs[0].Y == initialY {
		t.Error("RandomWalk NPC didn't move after 60 ticks")
	}
}

// TestBoundaryCollision verifies NPCs can't move out of bounds
func TestBoundaryCollision(t *testing.T) {
	// Create NPC at corner
	npc := NPC{
		ID:          1,
		X:           0,
		Y:           0,
		Sprite:      0,
		Pattern:     PatternRandomWalk{Interval: 1},
		PatrolIndex: 0,
		MoveCounter: 0,
	}

	world := World{
		Tick: 0,
		Planet: PlanetState{
			Width:  64,
			Height: 64,
			Tiles:  make([]Tile, 64*64),
		},
		NPCs:      []NPC{npc},
		Selection: SelectionNone{},
	}

	// Run 100 steps
	for i := 0; i < 100; i++ {
		world, _, _ = Step(world, FrameInput{})
	}

	// NPC should still be in bounds
	if world.NPCs[0].X < 0 || world.NPCs[0].X >= 64 ||
		world.NPCs[0].Y < 0 || world.NPCs[0].Y >= 64 {
		t.Errorf("NPC moved out of bounds to (%d,%d)", world.NPCs[0].X, world.NPCs[0].Y)
	}
}

// TestMultipleNPCsIndependent verifies NPCs move independently
func TestMultipleNPCsIndependent(t *testing.T) {
	world := InitWorld(1234)

	// Record initial positions
	initial := make([][2]int, len(world.NPCs))
	for i, npc := range world.NPCs {
		initial[i] = [2]int{npc.X, npc.Y}
	}

	// Run 100 steps
	for i := 0; i < 100; i++ {
		world, _, _ = Step(world, FrameInput{})
	}

	// NPCs should have different movement patterns
	// Static NPC shouldn't move, others should
	if world.NPCs[2].X != initial[2][0] || world.NPCs[2].Y != initial[2][1] {
		t.Error("Static NPC moved when it shouldn't")
	}

	// At least one random walker should have moved
	moved := false
	for i := 0; i < 2; i++ { // Check first two (random walkers)
		if world.NPCs[i].X != initial[i][0] || world.NPCs[i].Y != initial[i][1] {
			moved = true
			break
		}
	}
	if !moved {
		t.Error("No random walk NPCs moved after 100 ticks")
	}
}

// TestPatrolMovement verifies patrol NPC follows its path
func TestPatrolMovement(t *testing.T) {
	// Create an NPC with a simple patrol path: East, East, West, West
	// This should result in the NPC moving right 2 tiles, then back to start
	npc := NPC{
		ID:          1,
		X:           10,
		Y:           10,
		Sprite:      0,
		Pattern:     PatternPatrol{Path: []Direction{East, East, West, West}, Interval: 1},
		PatrolIndex: 0,
		MoveCounter: 0, // Ready to move immediately
	}

	world := World{
		Tick: 0,
		Planet: PlanetState{
			Width:  64,
			Height: 64,
			Tiles:  make([]Tile, 64*64),
		},
		NPCs:      []NPC{npc},
		Selection: SelectionNone{},
	}

	// First move: East (10,10) -> (11,10)
	world, _, _ = Step(world, FrameInput{})
	if world.NPCs[0].X != 11 || world.NPCs[0].Y != 10 {
		t.Errorf("After 1st move: expected (11,10), got (%d,%d)", world.NPCs[0].X, world.NPCs[0].Y)
	}

	// Wait for interval (MoveCounter resets to 1)
	world, _, _ = Step(world, FrameInput{})

	// Second move: East (11,10) -> (12,10)
	world, _, _ = Step(world, FrameInput{})
	if world.NPCs[0].X != 12 || world.NPCs[0].Y != 10 {
		t.Errorf("After 2nd move: expected (12,10), got (%d,%d)", world.NPCs[0].X, world.NPCs[0].Y)
	}

	// Wait + Third move: West (12,10) -> (11,10)
	world, _, _ = Step(world, FrameInput{})
	world, _, _ = Step(world, FrameInput{})
	if world.NPCs[0].X != 11 || world.NPCs[0].Y != 10 {
		t.Errorf("After 3rd move: expected (11,10), got (%d,%d)", world.NPCs[0].X, world.NPCs[0].Y)
	}

	// Wait + Fourth move: West (11,10) -> (10,10) - back to start!
	world, _, _ = Step(world, FrameInput{})
	world, _, _ = Step(world, FrameInput{})
	if world.NPCs[0].X != 10 || world.NPCs[0].Y != 10 {
		t.Errorf("After 4th move: expected (10,10), got (%d,%d)", world.NPCs[0].X, world.NPCs[0].Y)
	}

	// Verify patrol loops - should continue with East again
	world, _, _ = Step(world, FrameInput{})
	world, _, _ = Step(world, FrameInput{})
	if world.NPCs[0].X != 11 || world.NPCs[0].Y != 10 {
		t.Errorf("After patrol loop: expected (11,10), got (%d,%d)", world.NPCs[0].X, world.NPCs[0].Y)
	}
}

// TestModeSwitching verifies M key toggles between Ship and Galaxy modes
func TestModeSwitching(t *testing.T) {
	world := InitWorld(1234)

	// Should start in Ship Exploration mode
	if _, ok := world.Mode.(ModeShipExploration); !ok {
		t.Fatalf("Expected ModeShipExploration, got %T", world.Mode)
	}

	// Press M to switch to galaxy map
	input := FrameInput{
		Keys:     []KeyEvent{{Key: KeyM, Kind: "pressed"}},
		TestMode: true,
	}
	world, _, _ = Step(world, input)

	if _, ok := world.Mode.(ModeGalaxyMap); !ok {
		t.Fatalf("Expected ModeGalaxyMap after M press, got %T", world.Mode)
	}

	// Press M again to return to ship
	world, _, _ = Step(world, input)
	if _, ok := world.Mode.(ModeShipExploration); !ok {
		t.Fatalf("Expected ModeShipExploration after second M press, got %T", world.Mode)
	}

	// Switch to galaxy, then press ESC to return
	world, _, _ = Step(world, input) // M -> galaxy
	input = FrameInput{
		Keys:     []KeyEvent{{Key: KeyEscape, Kind: "pressed"}},
		TestMode: true,
	}
	world, _, _ = Step(world, input)
	if _, ok := world.Mode.(ModeShipExploration); !ok {
		t.Fatalf("Expected ModeShipExploration after ESC, got %T", world.Mode)
	}
}

// TestGalaxyMapRendering verifies galaxy map generates correct draw commands
func TestGalaxyMapRendering(t *testing.T) {
	world := InitWorld(1234)

	// Switch to galaxy map
	input := FrameInput{
		Keys:     []KeyEvent{{Key: KeyM, Kind: "pressed"}},
		TestMode: true,
	}
	world, output, _ := Step(world, input)

	// Should have draw commands
	if len(output.Draw) == 0 {
		t.Fatal("Expected draw commands from galaxy map")
	}

	// Count circles (stars) and lines (connections)
	circles := 0
	lines := 0
	for _, cmd := range output.Draw {
		switch cmd.(type) {
		case DrawCmdCircle:
			circles++
		case DrawCmdLine:
			lines++
		}
	}

	if circles != 5 {
		t.Errorf("Expected 5 star circles, got %d", circles)
	}
	if lines < 3 {
		t.Errorf("Expected at least 3 network lines, got %d", lines)
	}
}

// TestGalaxyMapStarPositions verifies stars are centered properly
func TestGalaxyMapStarPositions(t *testing.T) {
	world := InitWorld(1234)

	// Switch to galaxy map
	input := FrameInput{
		Keys:     []KeyEvent{{Key: KeyM, Kind: "pressed"}},
		TestMode: true,
	}
	world, output, _ := Step(world, input)

	// Find the "You Are Here" star (at world 0,0, should be at screen center)
	screenCenterX := 640.0
	screenCenterY := 360.0

	for _, cmd := range output.Draw {
		if c, ok := cmd.(DrawCmdCircle); ok {
			// The red star (color 10) at 0,0 should be at screen center
			if c.Color == 10 {
				if c.X != screenCenterX || c.Y != screenCenterY {
					t.Errorf("Center star should be at (%.0f,%.0f), got (%.0f,%.0f)",
						screenCenterX, screenCenterY, c.X, c.Y)
				}
				return
			}
		}
	}
	t.Error("Did not find center star (color 10)")
}

// TestGalaxyMapLabelsMatchStars verifies labels are positioned next to stars
func TestGalaxyMapLabelsMatchStars(t *testing.T) {
	world := InitWorld(1234)

	// Switch to galaxy map
	input := FrameInput{
		Keys:     []KeyEvent{{Key: KeyM, Kind: "pressed"}},
		TestMode: true,
	}
	_, output, _ := Step(world, input)

	// Collect star positions and label positions
	starsByColor := make(map[int]DrawCmdCircle)
	var labels []DrawCmdText

	for _, cmd := range output.Draw {
		switch c := cmd.(type) {
		case DrawCmdCircle:
			starsByColor[c.Color] = c
		case DrawCmdText:
			labels = append(labels, c)
		}
	}

	// Each label should be near its star (X + radius + small offset)
	for _, label := range labels {
		// Find a star that this label could belong to
		found := false
		for _, star := range starsByColor {
			expectedLabelX := star.X + star.Radius + 5
			expectedLabelY := star.Y - 5
			// Allow small tolerance
			if label.X == expectedLabelX && label.Y == expectedLabelY {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Label %q at (%.0f,%.0f) doesn't match any star position",
				label.Text, label.X, label.Y)
		}
	}
}

// TestPatrolBoundaryHandling verifies patrol NPC handles blocked moves
func TestPatrolBoundaryHandling(t *testing.T) {
	// Create NPC at corner with patrol that tries to go out of bounds
	npc := NPC{
		ID:          1,
		X:           0,
		Y:           0,
		Sprite:      0,
		Pattern:     PatternPatrol{Path: []Direction{North, West, South, East}, Interval: 1},
		PatrolIndex: 0,
		MoveCounter: 0,
	}

	world := World{
		Tick: 0,
		Planet: PlanetState{
			Width:  64,
			Height: 64,
			Tiles:  make([]Tile, 64*64),
		},
		NPCs:      []NPC{npc},
		Selection: SelectionNone{},
	}

	// First move: North blocked (at Y=0), NPC stays but patrol index advances
	world, _, _ = Step(world, FrameInput{})
	if world.NPCs[0].X != 0 || world.NPCs[0].Y != 0 {
		t.Errorf("NPC should stay at (0,0) when blocked, got (%d,%d)", world.NPCs[0].X, world.NPCs[0].Y)
	}
	if world.NPCs[0].PatrolIndex != 1 {
		t.Errorf("Patrol index should advance even when blocked, got %d", world.NPCs[0].PatrolIndex)
	}

	// Wait + Second move: West blocked, index advances to 2
	world, _, _ = Step(world, FrameInput{})
	world, _, _ = Step(world, FrameInput{})
	if world.NPCs[0].PatrolIndex != 2 {
		t.Errorf("Patrol index should be 2, got %d", world.NPCs[0].PatrolIndex)
	}

	// Wait + Third move: South works! (0,0) -> (0,1)
	world, _, _ = Step(world, FrameInput{})
	world, _, _ = Step(world, FrameInput{})
	if world.NPCs[0].X != 0 || world.NPCs[0].Y != 1 {
		t.Errorf("Expected (0,1), got (%d,%d)", world.NPCs[0].X, world.NPCs[0].Y)
	}
}
