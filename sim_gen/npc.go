package sim_gen

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

// Visual interpolation speed (0.10 = reaches destination in 10 frames)
// This should match NPC movement intervals for continuous motion
const visualLerpSpeed = 0.10

// lerpToZero interpolates a value toward 0
func lerpToZero(v float64, speed float64) float64 {
	if v > 0 {
		v -= speed
		if v < 0 {
			v = 0
		}
	} else if v < 0 {
		v += speed
		if v > 0 {
			v = 0
		}
	}
	return v
}

// updateRandomWalk moves NPC every N ticks in pseudo-random direction
func updateRandomWalk(npc NPC, world World, interval int) NPC {
	// Always interpolate visual offset toward 0 (smooth movement)
	newVisualX := lerpToZero(npc.VisualOffsetX, visualLerpSpeed)
	newVisualY := lerpToZero(npc.VisualOffsetY, visualLerpSpeed)

	if npc.MoveCounter <= 0 {
		// Time to move! Pick direction based on tick + id (deterministic "random")
		dirIndex := (world.Tick + npc.ID) % 4
		dx, dy := directionOffset(Direction(dirIndex))
		newX, newY := npc.X+dx, npc.Y+dy

		if isValidPosition(newX, newY, world.Planet.Width, world.Planet.Height) {
			return NPC{
				ID:            npc.ID,
				X:             newX,
				Y:             newY,
				Sprite:        npc.Sprite,
				Pattern:       npc.Pattern,
				PatrolIndex:   npc.PatrolIndex,
				MoveCounter:   interval,
				VisualOffsetX: -float64(dx), // Start offset from where we came
				VisualOffsetY: -float64(dy),
			}
		}
		// Blocked - reset counter but don't move
		return NPC{
			ID:            npc.ID,
			X:             npc.X,
			Y:             npc.Y,
			Sprite:        npc.Sprite,
			Pattern:       npc.Pattern,
			PatrolIndex:   npc.PatrolIndex,
			MoveCounter:   interval,
			VisualOffsetX: newVisualX,
			VisualOffsetY: newVisualY,
		}
	}
	// Decrement counter, continue interpolating
	return NPC{
		ID:            npc.ID,
		X:             npc.X,
		Y:             npc.Y,
		Sprite:        npc.Sprite,
		Pattern:       npc.Pattern,
		PatrolIndex:   npc.PatrolIndex,
		MoveCounter:   npc.MoveCounter - 1,
		VisualOffsetX: newVisualX,
		VisualOffsetY: newVisualY,
	}
}

// updatePatrol follows a fixed path, looping back to start when complete
func updatePatrol(npc NPC, world World, path []Direction, interval int) NPC {
	// Empty path means static
	if len(path) == 0 {
		return npc
	}

	// Always interpolate visual offset toward 0 (smooth movement)
	newVisualX := lerpToZero(npc.VisualOffsetX, visualLerpSpeed)
	newVisualY := lerpToZero(npc.VisualOffsetY, visualLerpSpeed)

	if npc.MoveCounter <= 0 {
		// Time to move! Get current direction from path
		dir := path[npc.PatrolIndex%len(path)]
		dx, dy := directionOffset(dir)
		newX, newY := npc.X+dx, npc.Y+dy

		// Advance to next path index (wrap around)
		nextIndex := (npc.PatrolIndex + 1) % len(path)

		if isValidPosition(newX, newY, world.Planet.Width, world.Planet.Height) {
			return NPC{
				ID:            npc.ID,
				X:             newX,
				Y:             newY,
				Sprite:        npc.Sprite,
				Pattern:       npc.Pattern,
				PatrolIndex:   nextIndex,
				MoveCounter:   interval,
				VisualOffsetX: -float64(dx), // Start offset from where we came
				VisualOffsetY: -float64(dy),
			}
		}
		// Blocked - still advance index so patrol continues, reset counter
		return NPC{
			ID:            npc.ID,
			X:             npc.X,
			Y:             npc.Y,
			Sprite:        npc.Sprite,
			Pattern:       npc.Pattern,
			PatrolIndex:   nextIndex,
			MoveCounter:   interval,
			VisualOffsetX: newVisualX,
			VisualOffsetY: newVisualY,
		}
	}
	// Decrement counter, continue interpolating
	return NPC{
		ID:            npc.ID,
		X:             npc.X,
		Y:             npc.Y,
		Sprite:        npc.Sprite,
		Pattern:       npc.Pattern,
		PatrolIndex:   npc.PatrolIndex,
		MoveCounter:   npc.MoveCounter - 1,
		VisualOffsetX: newVisualX,
		VisualOffsetY: newVisualY,
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
