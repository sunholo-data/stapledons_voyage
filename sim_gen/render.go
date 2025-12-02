package sim_gen

import (
	"fmt"
	"math"
)

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
	// Branch between plane view and sky view
	if mode.SkyViewMode {
		return renderGalaxySkyView(world, mode, testMode)
	}
	return renderGalaxyPlaneView(world, mode, testMode)
}

// renderGalaxySkyView renders the galaxy map as a planetarium view from Sol
// The camera is always at Sol, looking in the direction specified by ViewLon/ViewLat
// Stars are projected based on their angular position from the view direction
func renderGalaxySkyView(world World, mode ModeGalaxyMap, testMode bool) []DrawCmd {
	drawCmds := make([]DrawCmd, 0, 500)

	// Screen dimensions from constants
	screenW := float64(ScreenWidth)
	screenH := float64(ScreenHeight)
	screenCenterX := screenW / 2
	screenCenterY := screenH / 2

	// Draw galaxy background with scrolling based on view direction
	galaxyBg := DrawCmdGalaxyBg{
		Opacity:     0.5, // Brighter in sky view
		Z:           0,
		SkyViewMode: true,
		ViewLon:     mode.ViewLon,
		ViewLat:     mode.ViewLat,
		FOV:         mode.FOV,
	}
	drawCmds = append(drawCmds, galaxyBg)

	// Get star catalog
	catalog := GetStarCatalog()
	if catalog == nil || len(catalog.Stars) == 0 {
		return drawCmds
	}

	// Calculate pixels per degree based on FOV
	// If FOV is 90°, it should span the screen width
	pixelsPerDegree := screenW / mode.FOV

	// LOD: Calculate magnitude cutoff based on FOV (inverse of zoom)
	// Wide FOV (zoomed out): show fewer stars
	// Narrow FOV (zoomed in): show more stars
	// FOV 150° (max): mag < 5 (~500 stars)
	// FOV 90° (default): mag < 8 (~2000 stars)
	// FOV 20° (min): show all stars
	magCutoff := 12.0 - mode.FOV*0.05
	if magCutoff < 4.0 {
		magCutoff = 4.0
	}
	if magCutoff > 15.0 {
		magCutoff = 15.0
	}

	// Render stars visible within the FOV
	starsRendered := 0

	for _, star := range catalog.Stars {
		// LOD: Skip dim stars when FOV is wide
		if star.VMag > magCutoff {
			continue
		}

		// Get star's galactic coordinates
		starLon, starLat := star.GalacticLonLat()

		// Calculate angular distance from view center
		angDist := AngularDistance(mode.ViewLon, mode.ViewLat, starLon, starLat)

		// Skip stars outside FOV (with some margin)
		if angDist > mode.FOV*0.7 {
			continue
		}

		// Project star position to screen using gnomonic projection
		// This projects the celestial sphere onto a tangent plane
		sx, sy := projectToScreen(mode.ViewLon, mode.ViewLat, starLon, starLat, pixelsPerDegree, screenCenterX, screenCenterY)

		// Skip if off screen
		if sx < -20 || sx > screenW+20 || sy < -20 || sy > screenH+20 {
			continue
		}

		// Calculate scale based on magnitude
		scale := StarScale(star.VMag)
		if scale < 0.15 {
			scale = 0.15
		}
		if scale > 1.2 {
			scale = 1.2
		}

		// Get sprite ID based on spectral type
		spriteID := SpectralSpriteID(star.Spectral)

		starCmd := DrawCmdStar{
			X:        sx,
			Y:        sy,
			SpriteID: spriteID,
			Scale:    scale,
			Z:        10,
		}
		drawCmds = append(drawCmds, starCmd)

		// Show names for bright/close stars
		showName := false
		if star.DistLY < 15 {
			showName = true
		} else if star.VMag < 3 {
			showName = true
		}

		if showName {
			labelCmd := DrawCmdText{
				Text:     star.Name,
				X:        sx + 10,
				Y:        sy - 2,
				FontSize: 0,
				Color:    4, // White
				Z:        11,
			}
			drawCmds = append(drawCmds, labelCmd)
		}

		starsRendered++
	}

	// Draw crosshair at center (view direction)
	drawCmds = append(drawCmds, DrawCmdLine{
		X1: screenCenterX - 15, Y1: screenCenterY,
		X2: screenCenterX + 15, Y2: screenCenterY,
		Color: 1, Width: 1, Z: 20,
	})
	drawCmds = append(drawCmds, DrawCmdLine{
		X1: screenCenterX, Y1: screenCenterY - 15,
		X2: screenCenterX, Y2: screenCenterY + 15,
		Color: 1, Width: 1, Z: 20,
	})

	// UI panels
	if !testMode {
		// Info panel showing view direction
		viewInfo := fmt.Sprintf("Sky View | l=%.0f° b=%.0f° | FOV %.0f° | %d stars",
			mode.ViewLon, mode.ViewLat, mode.FOV, starsRendered)

		modePanel := DrawCmdUi{
			ID:    "mode-info",
			Kind:  UiKindPanel,
			X:     0.02,
			Y:     0.02,
			W:     0.50,
			H:     0.07,
			Text:  viewInfo,
			Z:     100,
			Color: 0,
		}
		drawCmds = append(drawCmds, modePanel)

		controlsPanel := DrawCmdUi{
			ID:    "controls-help",
			Kind:  UiKindPanel,
			X:     0.02,
			Y:     0.91,
			W:     0.70,
			H:     0.07,
			Text:  "WASD: Look | Q/E: Zoom | V: Plane View | M: Exit",
			Z:     100,
			Color: 0,
		}
		drawCmds = append(drawCmds, controlsPanel)

		// Compass showing galactic directions
		compassPanel := DrawCmdUi{
			ID:    "compass",
			Kind:  UiKindPanel,
			X:     0.80,
			Y:     0.02,
			W:     0.18,
			H:     0.12,
			Text:  "l=0: Center\nl=180: Anti-C\nb+: N Pole\nb-: S Pole",
			Z:     100,
			Color: 3,
		}
		drawCmds = append(drawCmds, compassPanel)
	}

	return drawCmds
}

// projectToScreen projects a celestial position onto screen coordinates
// Uses gnomonic (tangent plane) projection centered on the view direction
func projectToScreen(viewLon, viewLat, targetLon, targetLat, pixelsPerDeg, centerX, centerY float64) (float64, float64) {
	// Convert to radians
	vLon := viewLon * math.Pi / 180.0
	vLat := viewLat * math.Pi / 180.0
	tLon := targetLon * math.Pi / 180.0
	tLat := targetLat * math.Pi / 180.0

	// Gnomonic projection
	// https://en.wikipedia.org/wiki/Gnomonic_projection
	cosC := math.Sin(vLat)*math.Sin(tLat) + math.Cos(vLat)*math.Cos(tLat)*math.Cos(tLon-vLon)
	if cosC <= 0 {
		// Point is behind the viewer
		return -1000, -1000
	}

	x := math.Cos(tLat) * math.Sin(tLon-vLon) / cosC
	y := (math.Cos(vLat)*math.Sin(tLat) - math.Sin(vLat)*math.Cos(tLat)*math.Cos(tLon-vLon)) / cosC

	// Convert to screen coordinates
	// x is positive to the right (increasing longitude)
	// y is positive upward (increasing latitude), but screen Y is inverted
	sx := centerX + x*pixelsPerDeg*180.0/math.Pi
	sy := centerY - y*pixelsPerDeg*180.0/math.Pi

	return sx, sy
}

// renderGalaxyPlaneView renders the top-down galactic plane view
func renderGalaxyPlaneView(world World, mode ModeGalaxyMap, testMode bool) []DrawCmd {
	drawCmds := make([]DrawCmd, 0, 500)

	// Screen dimensions from constants
	screenW := float64(ScreenWidth)
	screenH := float64(ScreenHeight)
	screenCenterX := screenW / 2
	screenCenterY := screenH / 2

	// Visual scale: pixels per light-year (at zoom 1.0)
	// Scale based on screen width - nearby stars (~160 ly diameter) fill ~60% of screen
	pixelsPerLY := screenW / 160.0 * 0.6

	// Black background for plane view (galaxy image only used in sky view)
	drawCmds = append(drawCmds, DrawCmdRectScreen{
		X: 0, Y: 0, W: screenW, H: screenH,
		Color: 8, // Black
		Z:     0,
	})

	// Get real star catalog
	catalog := GetStarCatalog()

	// Get all stars - we'll filter by screen bounds during rendering
	var visibleStars []Star
	if catalog != nil && len(catalog.Stars) > 0 {
		// Use all stars - screen bounds check will filter what's visible
		visibleStars = catalog.Stars
	}

	// LOD: Calculate magnitude cutoff based on zoom level
	// When zoomed out, skip dim stars that would be invisible
	// At zoom 0.3x: only show mag < 5 (brightest ~500 stars)
	// At zoom 1.0x: show mag < 8 (~2000 stars)
	// At zoom 5.0x+: show all stars
	magCutoff := 4.0 + mode.ZoomLevel*3.0
	if magCutoff > 15.0 {
		magCutoff = 15.0 // No limit when zoomed in
	}

	// Render stars from catalog
	starsRendered := 0
	for _, star := range visibleStars {
		// LOD: Skip dim stars when zoomed out
		if star.VMag > magCutoff {
			continue
		}

		// === 3D DEPTH EFFECTS ===
		// Star's Z coordinate (light-years above/below galactic plane)
		starZ := star.Z

		// 1. PARALLAX: Stars at different Z depths move at different speeds
		// Stars above plane (Z>0) appear to move slower (further away)
		// Stars below plane (Z<0) also move slower
		// This creates subtle depth when panning
		parallaxFactor := 1.0 / (1.0 + math.Abs(starZ)*0.008)

		// Transform: world (X,Y in light-years) -> screen with parallax
		// Apply visual scale and zoom, with depth-adjusted panning
		sx := (star.X-mode.CameraX*parallaxFactor)*pixelsPerLY*mode.ZoomLevel + screenCenterX
		sy := (star.Y-mode.CameraY*parallaxFactor)*pixelsPerLY*mode.ZoomLevel + screenCenterY

		// Skip if off screen (with some margin)
		if sx < -20 || sx > screenW+20 || sy < -20 || sy > screenH+20 {
			continue
		}

		// 2. Z-BASED OPACITY: Stars further from galactic plane are more transparent
		// |Z| = 0 -> alpha = 1.0 (on plane, fully visible)
		// |Z| = 50 ly -> alpha = 0.5 (halfway faded)
		// |Z| = 100+ ly -> alpha = 0.3 (very faded)
		zDist := math.Abs(starZ)
		alpha := 1.0 - zDist*0.007
		if alpha < 0.3 {
			alpha = 0.3
		}

		// Calculate scale based on magnitude
		scale := StarScale(star.VMag)

		// 3. DEPTH GLOW: Stars off-plane get slightly larger (soft/unfocused)
		// This makes on-plane stars appear sharper
		if zDist > 10 {
			scale *= 1.0 + zDist*0.003
		}

		// Slight zoom scaling for close-up detail
		if mode.ZoomLevel > 1.5 {
			scale *= 1.0 + (mode.ZoomLevel-1.5)*0.15
		}
		if scale < 0.1 {
			scale = 0.1
		}
		if scale > 1.5 {
			scale = 1.5
		}

		// Get sprite ID based on spectral type
		spriteID := SpectralSpriteID(star.Spectral)

		starCmd := DrawCmdStar{
			X:        sx,
			Y:        sy,
			SpriteID: spriteID,
			Scale:    scale,
			Alpha:    alpha,
			Z:        10,
		}
		drawCmds = append(drawCmds, starCmd)

		// Show names for bright/close stars (zoom-dependent)
		showName := false
		if star.DistLY < 12 { // Always show nearest stars
			showName = true
		} else if star.VMag < 4 && mode.ZoomLevel > 0.8 { // Bright stars when zoomed in
			showName = true
		} else if mode.ZoomLevel > 2.0 && star.DistLY < 25 { // Show more when very zoomed in
			showName = true
		}

		if showName {
			labelCmd := DrawCmdText{
				Text:     star.Name,
				X:        sx + 10,
				Y:        sy - 2,
				FontSize: 0,
				Color:    4, // White
				Z:        11,
			}
			drawCmds = append(drawCmds, labelCmd)
		}

		starsRendered++
	}

	// Draw Sol marker (the Sun at origin)
	solX := (0-mode.CameraX)*pixelsPerLY*mode.ZoomLevel + screenCenterX
	solY := (0-mode.CameraY)*pixelsPerLY*mode.ZoomLevel + screenCenterY
	if solX >= -20 && solX <= screenW+20 && solY >= -20 && solY <= screenH+20 {
		// Sol marker - yellow circle
		drawCmds = append(drawCmds, DrawCmdCircle{
			X: solX, Y: solY, Radius: 6, Color: 13, Filled: true, Z: 15,
		})
		// Sol label
		drawCmds = append(drawCmds, DrawCmdText{
			Text: "Sol", X: solX + 8, Y: solY - 2, FontSize: 0, Color: 13, Z: 16,
		})
	}

	// Draw player position marker (always at center when camera follows player)
	playerCmd := DrawCmdCircle{
		X:      screenCenterX,
		Y:      screenCenterY,
		Radius: 6,
		Color:  1, // Green
		Filled: false,
		Z:      20,
	}
	drawCmds = append(drawCmds, playerCmd)

	// Crosshair at player position
	drawCmds = append(drawCmds, DrawCmdLine{
		X1: screenCenterX - 10, Y1: screenCenterY,
		X2: screenCenterX + 10, Y2: screenCenterY,
		Color: 1, Width: 1, Z: 20,
	})
	drawCmds = append(drawCmds, DrawCmdLine{
		X1: screenCenterX, Y1: screenCenterY - 10,
		X2: screenCenterX, Y2: screenCenterY + 10,
		Color: 1, Width: 1, Z: 20,
	})

	// Draw selected star highlight and info panel
	if mode.SelectedStar >= 0 && catalog != nil && mode.SelectedStar < len(catalog.Stars) {
		star := catalog.Stars[mode.SelectedStar]

		// Calculate selected star's screen position
		starZ := star.Z
		parallaxFactor := 1.0 / (1.0 + math.Abs(starZ)*0.008)
		selX := (star.X-mode.CameraX*parallaxFactor)*pixelsPerLY*mode.ZoomLevel + screenCenterX
		selY := (star.Y-mode.CameraY*parallaxFactor)*pixelsPerLY*mode.ZoomLevel + screenCenterY

		// Selection ring
		drawCmds = append(drawCmds, DrawCmdCircle{
			X: selX, Y: selY, Radius: 15, Color: 1, Filled: false, Z: 25,
		})
		drawCmds = append(drawCmds, DrawCmdCircle{
			X: selX, Y: selY, Radius: 17, Color: 1, Filled: false, Z: 25,
		})

		// Star info panel (bottom right)
		if !testMode {
			infoPanel := DrawCmdUi{
				ID:   "star-info",
				Kind: UiKindPanel,
				X:    0.60,
				Y:    0.70,
				W:    0.38,
				H:    0.28,
				Text: fmt.Sprintf("%s\n\nDist: %.1f ly\nSpectral: %s\nMag: %.1f\nPos: (%.1f, %.1f, %.1f)",
					star.Name, star.DistLY, star.Spectral, star.VMag, star.X, star.Y, star.Z),
				Z:     200,
				Color: 3, // Dark panel
			}
			drawCmds = append(drawCmds, infoPanel)
		}
	}

	// Draw scale indicator
	if !testMode {
		// Scale bar: show 20 light-years at current zoom
		scaleLY := 20.0
		scalePixels := scaleLY * pixelsPerLY * mode.ZoomLevel
		if scalePixels > 150 {
			scaleLY = 10.0
			scalePixels = scaleLY * pixelsPerLY * mode.ZoomLevel
		}
		if scalePixels < 30 {
			scaleLY = 50.0
			scalePixels = scaleLY * pixelsPerLY * mode.ZoomLevel
		}
		scaleY := 440.0
		drawCmds = append(drawCmds, DrawCmdLine{
			X1: 20, Y1: scaleY, X2: 20 + scalePixels, Y2: scaleY,
			Color: 4, Width: 2, Z: 100,
		})
		drawCmds = append(drawCmds, DrawCmdText{
			Text: fmt.Sprintf("%.0f ly", scaleLY),
			X:    20, Y: scaleY + 12,
			FontSize: 0, Color: 4, Z: 100,
		})
	}

	// UI panels
	if !testMode {
		// Info panel - shorter text to avoid overflow
		infoText := fmt.Sprintf("%.0f,%.0f ly | x%.1f | %d stars",
			mode.CameraX, mode.CameraY, mode.ZoomLevel, starsRendered)

		modePanel := DrawCmdUi{
			ID:    "mode-info",
			Kind:  UiKindPanel,
			X:     0.02,
			Y:     0.02,
			W:     0.35,
			H:     0.07,
			Text:  infoText,
			Z:     100,
			Color: 0,
		}
		drawCmds = append(drawCmds, modePanel)

		controlsPanel := DrawCmdUi{
			ID:    "controls-help",
			Kind:  UiKindPanel,
			X:     0.02,
			Y:     0.91,
			W:     0.60,
			H:     0.07,
			Text:  "WASD: Pan | Q/E: Zoom | V: Sky View | M: Exit",
			Z:     100,
			Color: 0,
		}
		drawCmds = append(drawCmds, controlsPanel)

		// Spectral type legend - more compact
		legendPanel := DrawCmdUi{
			ID:    "legend",
			Kind:  UiKindPanel,
			X:     0.75,
			Y:     0.02,
			W:     0.23,
			H:     0.20,
			Text:  "O/B: Blue\nA/F: White\nG: Yellow\nK: Orange\nM: Red",
			Z:     100,
			Color: 3,
		}
		drawCmds = append(drawCmds, legendPanel)
	}

	return drawCmds
}
