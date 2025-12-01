package sim_gen

import "fmt"

// =============================================================================
// Mode-Specific Rendering
// =============================================================================

// renderShipExploration generates draw commands for ship exploration mode
func renderShipExploration(world World, mode ModeShipExploration, planet PlanetState, selection Selection, camera Camera, testMode bool) []DrawCmd {
	drawCmds := make([]DrawCmd, 0, len(planet.Tiles)*2+len(world.NPCs)+10)

	// Draw tiles
	for i, tile := range planet.Tiles {
		x := i % planet.Width
		y := i / planet.Width
		cmd := DrawCmdIsoTile{
			Tile:     Coord{X: x, Y: y},
			Height:   0,
			SpriteID: 0,
			Layer:    0,
			Color:    tile.Biome,
		}
		drawCmds = append(drawCmds, cmd)

		if hs, ok := tile.Structure.(HasStructure); ok {
			structCmd := DrawCmdIsoTile{
				Tile:     Coord{X: x, Y: y},
				Height:   1,
				SpriteID: 0,
				Layer:    50,
				Color:    5 + int(hs.Type),
			}
			drawCmds = append(drawCmds, structCmd)
		}
	}

	// Draw NPCs
	for _, npc := range world.NPCs {
		npcCmd := DrawCmdIsoEntity{
			ID:       fmt.Sprintf("npc-%d", npc.ID),
			Tile:     Coord{X: npc.X, Y: npc.Y},
			OffsetX:  npc.VisualOffsetX,
			OffsetY:  npc.VisualOffsetY,
			Height:   0,
			SpriteID: 100 + npc.Sprite,
			Layer:    100,
		}
		drawCmds = append(drawCmds, npcCmd)
	}

	// Selection highlight - slightly elevated to "pop" above the tile
	if sel, ok := selection.(SelectionTile); ok {
		highlightCmd := DrawCmdIsoTile{
			Tile:     Coord{X: sel.X, Y: sel.Y},
			Height:   1, // Elevated for visibility
			SpriteID: 0,
			Layer:    150,
			Color:    4, // White semi-transparent
		}
		drawCmds = append(drawCmds, highlightCmd)
	}

	// UI panels
	if !testMode {
		cameraPanel := DrawCmdUi{
			ID:    "camera-info",
			Kind:  UiKindPanel,
			X:     0.02,
			Y:     0.02,
			W:     0.28,
			H:     0.06,
			Text:  fmt.Sprintf("Cam: (%.0f, %.0f) Zoom: %.2fx", camera.X, camera.Y, camera.Zoom),
			Z:     0,
			Color: 3,
		}
		drawCmds = append(drawCmds, cameraPanel)

		modePanel := DrawCmdUi{
			ID:    "mode-info",
			Kind:  UiKindPanel,
			X:     0.70,
			Y:     0.02,
			W:     0.28,
			H:     0.06,
			Text:  "Ship Exploration | M: Galaxy Map",
			Z:     0,
			Color: 1,
		}
		drawCmds = append(drawCmds, modePanel)

		controlsPanel := DrawCmdUi{
			ID:    "controls-help",
			Kind:  UiKindPanel,
			X:     0.02,
			Y:     0.90,
			W:     0.96,
			H:     0.08,
			Text:  "WASD: Move | Q/E: Zoom | Click: Select | I: Info | B: Build | X: Clear | M: Map",
			Z:     0,
			Color: 0,
		}
		drawCmds = append(drawCmds, controlsPanel)
	}

	return drawCmds
}

// renderGalaxyMap generates draw commands for galaxy map mode
func renderGalaxyMap(world World, mode ModeGalaxyMap, testMode bool) []DrawCmd {
	drawCmds := make([]DrawCmd, 0, 50)

	// Screen center (assume 1280x720 default, will be centered anyway)
	screenCenterX := 640.0
	screenCenterY := 360.0

	// Draw starfield background (dark)
	bgRect := DrawCmdRect{
		X:     0,
		Y:     0,
		W:     1920,
		H:     1080,
		Color: 8, // Black
		Z:     0,
	}
	drawCmds = append(drawCmds, bgRect)

	// Star data in world coordinates (centered around 0,0)
	// In real implementation, these would come from galaxy data
	type starData struct {
		worldX, worldY float64
		radius         float64
		color          int
		name           string
	}
	stars := []starData{
		{-100, -50, 8, 1, "Sol"},       // Green (thriving)
		{100, 50, 6, 13, "Proxima"},    // Yellow (declining) - color 13
		{-150, 100, 5, 7, "Barnard"},   // Gray (extinct)
		{150, -100, 7, 0, "Sirius"},    // Blue (unknown)
		{0, 0, 10, 10, "You Are Here"}, // Red (player position)
	}

	// Transform star positions: world -> screen
	// screen = (world - camera) * zoom + screenCenter
	for _, star := range stars {
		sx := (star.worldX-mode.CameraX)*mode.ZoomLevel + screenCenterX
		sy := (star.worldY-mode.CameraY)*mode.ZoomLevel + screenCenterY
		r := star.radius * mode.ZoomLevel

		starCmd := DrawCmdCircle{
			X:      sx,
			Y:      sy,
			Radius: r,
			Color:  star.color,
			Filled: true,
			Z:      10,
		}
		drawCmds = append(drawCmds, starCmd)

		// Draw label next to star
		labelCmd := DrawCmdText{
			Text:     star.name,
			X:        sx + r + 5,
			Y:        sy - 5,
			FontSize: 0,
			Color:    0,
			Z:        11,
		}
		drawCmds = append(drawCmds, labelCmd)
	}

	// Draw network edges if enabled
	if mode.ShowNetwork {
		// Helper to get screen coords for a star index
		starScreen := func(idx int) (float64, float64) {
			s := stars[idx]
			return (s.worldX-mode.CameraX)*mode.ZoomLevel + screenCenterX,
				(s.worldY-mode.CameraY)*mode.ZoomLevel + screenCenterY
		}

		// Connection: Sol -> You Are Here
		x1, y1 := starScreen(0)
		x2, y2 := starScreen(4)
		drawCmds = append(drawCmds, DrawCmdLine{
			X1: x1, Y1: y1, X2: x2, Y2: y2,
			Color: 7, Width: 2, Z: 5,
		})

		// Connection: You Are Here -> Proxima
		x1, y1 = starScreen(4)
		x2, y2 = starScreen(1)
		drawCmds = append(drawCmds, DrawCmdLine{
			X1: x1, Y1: y1, X2: x2, Y2: y2,
			Color: 7, Width: 2, Z: 5,
		})

		// Connection: Barnard -> You Are Here
		x1, y1 = starScreen(2)
		x2, y2 = starScreen(4)
		drawCmds = append(drawCmds, DrawCmdLine{
			X1: x1, Y1: y1, X2: x2, Y2: y2,
			Color: 7, Width: 1, Z: 5,
		})
	}

	// UI panels
	if !testMode {
		modePanel := DrawCmdUi{
			ID:    "mode-info",
			Kind:  UiKindPanel,
			X:     0.02,
			Y:     0.02,
			W:     0.35,
			H:     0.08,
			Text:  fmt.Sprintf("Galaxy Map | Cam: (%.0f, %.0f) Zoom: %.2fx", mode.CameraX, mode.CameraY, mode.ZoomLevel),
			Z:     100,
			Color: 0,
		}
		drawCmds = append(drawCmds, modePanel)

		controlsPanel := DrawCmdUi{
			ID:    "controls-help",
			Kind:  UiKindPanel,
			X:     0.02,
			Y:     0.90,
			W:     0.96,
			H:     0.08,
			Text:  "WASD: Pan | Q/E: Zoom | M or ESC: Return to Ship",
			Z:     100,
			Color: 0,
		}
		drawCmds = append(drawCmds, controlsPanel)

		// Legend
		legendPanel := DrawCmdUi{
			ID:    "legend",
			Kind:  UiKindPanel,
			X:     0.75,
			Y:     0.02,
			W:     0.23,
			H:     0.15,
			Text:  "Green: Thriving\nYellow: Declining\nGray: Extinct\nBlue: Unknown",
			Z:     100,
			Color: 3,
		}
		drawCmds = append(drawCmds, legendPanel)
	}

	return drawCmds
}
