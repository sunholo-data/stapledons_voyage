package render

import "stapledons_voyage/sim_gen"

// isSceneBasedMode detects if we're in scene-based interior rendering mode
// Returns true if we have both:
// 1. Deck background sprite (UI with sprite ID 9000-9001)
// 2. Space DrawCmds (TexturedPlanet or SpaceBg)
func (r *Renderer) isSceneBasedMode(out *sim_gen.FrameOutput) bool {
	hasDeckBackground := false
	hasSpaceContent := false

	for _, cmd := range out.Draw {
		switch cmd.Kind {
		case sim_gen.DrawCmdKindUi:
			// Check for deck background sprites (9000-9001)
			if cmd.Ui.SpriteId >= 9000 && cmd.Ui.SpriteId < 9010 {
				hasDeckBackground = true
			}
		case sim_gen.DrawCmdKindTexturedPlanet, sim_gen.DrawCmdKindSpaceBg:
			hasSpaceContent = true
		}

		// Early exit if we found both
		if hasDeckBackground && hasSpaceContent {
			return true
		}
	}

	return hasDeckBackground && hasSpaceContent
}
