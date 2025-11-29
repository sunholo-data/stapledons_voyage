package sim_gen

import "testing"

// TestInspectAction verifies inspect action shows tile info
func TestInspectAction(t *testing.T) {
	world := InitWorld(1234)

	// Select a tile first
	input := FrameInput{
		ClickedThisFrame: true,
		WorldMouseX:      16, // Tile (2, 0)
		WorldMouseY:      8,  // Tile (2, 1)
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
		WorldMouseX:      44, // 5 * 8 + 4
		WorldMouseY:      44,
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
		WorldMouseX:      44,
		WorldMouseY:      44,
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
		WorldMouseX:      44,
		WorldMouseY:      44,
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
		WorldMouseX:      44,
		WorldMouseY:      44,
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

	if len(world.NPCs) != 3 {
		t.Errorf("Expected 3 NPCs, got %d", len(world.NPCs))
	}

	// Check first NPC
	if world.NPCs[0].ID != 1 || world.NPCs[0].X != 10 || world.NPCs[0].Y != 10 {
		t.Errorf("NPC 1 has wrong position: got (%d,%d)", world.NPCs[0].X, world.NPCs[0].Y)
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
