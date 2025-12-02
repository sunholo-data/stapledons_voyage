package sim_gen

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
	KeyM          = 12 // Toggle galaxy map
	KeyV          = 21 // Toggle view mode (plane/sky)
	KeyEscape     = 6  // Return to previous mode
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

// wasKeyPressed checks if a key was just pressed this frame
func wasKeyPressed(keys []KeyEvent, keyCode int) bool {
	for _, k := range keys {
		if k.Key == keyCode && k.Kind == "pressed" {
			return true
		}
	}
	return false
}

// processMode handles mode-specific input and transitions
func processMode(mode WorldMode, input FrameInput) WorldMode {
	keys := input.Keys
	switch m := mode.(type) {
	case ModeShipExploration:
		// M key opens galaxy map
		if wasKeyPressed(keys, KeyM) {
			return ModeGalaxyMap{
				CameraX:      0,     // Start at Sol (origin)
				CameraY:      0,
				ZoomLevel:    1.5,   // Good starting zoom to see nearby stars
				SkyViewMode:  false, // Start in plane view
				ViewLon:      0,     // Looking toward galactic center
				ViewLat:      0,     // On galactic plane
				FOV:          90,    // 90° field of view
				SelectedStar: -1,
				HoveredStar:  -1,
				ShowNetwork:  true,
			}
		}
		return m

	case ModeGalaxyMap:
		// ESC or M returns to ship exploration
		if wasKeyPressed(keys, KeyEscape) || wasKeyPressed(keys, KeyM) {
			return ModeShipExploration{
				PlayerPos:   Coord{X: 32, Y: 32},
				CurrentDeck: 0,
			}
		}

		// V key toggles between plane and sky view
		newSkyView := m.SkyViewMode
		if wasKeyPressed(keys, KeyV) {
			newSkyView = !m.SkyViewMode
		}

		// Different controls for plane vs sky view
		if newSkyView {
			// Sky view: WASD rotates view direction, Q/E changes FOV
			rotSpeed := 2.0 // Degrees per frame
			newLon := m.ViewLon
			newLat := m.ViewLat
			newFOV := m.FOV

			if isKeyDown(keys, KeyA) || isKeyDown(keys, KeyArrowLeft) {
				newLon -= rotSpeed
				if newLon < 0 {
					newLon += 360
				}
			}
			if isKeyDown(keys, KeyD) || isKeyDown(keys, KeyArrowRight) {
				newLon += rotSpeed
				if newLon >= 360 {
					newLon -= 360
				}
			}
			if isKeyDown(keys, KeyW) || isKeyDown(keys, KeyArrowUp) {
				newLat += rotSpeed
				if newLat > 90 {
					newLat = 90
				}
			}
			if isKeyDown(keys, KeyS) || isKeyDown(keys, KeyArrowDown) {
				newLat -= rotSpeed
				if newLat < -90 {
					newLat = -90
				}
			}
			if isKeyDown(keys, KeyQ) {
				newFOV *= 1.02 // Zoom out = wider FOV
				if newFOV > 150 {
					newFOV = 150
				}
			}
			if isKeyDown(keys, KeyE) {
				newFOV *= 0.98 // Zoom in = narrower FOV
				if newFOV < 20 {
					newFOV = 20
				}
			}

			return ModeGalaxyMap{
				CameraX:      m.CameraX,
				CameraY:      m.CameraY,
				ZoomLevel:    m.ZoomLevel,
				SkyViewMode:  newSkyView,
				ViewLon:      newLon,
				ViewLat:      newLat,
				FOV:          newFOV,
				SelectedStar: m.SelectedStar,
				HoveredStar:  m.HoveredStar,
				ShowNetwork:  m.ShowNetwork,
			}
		}

		// Plane view: WASD pans, Q/E zooms
		// Pan speed inversely proportional to zoom - slower when zoomed in
		// At zoom 1.0: 2.0 ly/frame, at zoom 10.0: 0.2 ly/frame, at zoom 20.0: 0.1 ly/frame
		panSpeed := 2.0 / m.ZoomLevel
		if panSpeed < 0.05 {
			panSpeed = 0.05 // Minimum movement
		}
		if panSpeed > 5.0 {
			panSpeed = 5.0 // Max when very zoomed out
		}

		newX, newY := m.CameraX, m.CameraY
		newZoom := m.ZoomLevel
		if isKeyDown(keys, KeyW) || isKeyDown(keys, KeyArrowUp) {
			newY -= panSpeed
		}
		if isKeyDown(keys, KeyS) || isKeyDown(keys, KeyArrowDown) {
			newY += panSpeed
		}
		if isKeyDown(keys, KeyA) || isKeyDown(keys, KeyArrowLeft) {
			newX -= panSpeed
		}
		if isKeyDown(keys, KeyD) || isKeyDown(keys, KeyArrowRight) {
			newX += panSpeed
		}
		if isKeyDown(keys, KeyQ) {
			newZoom *= 0.97
			if newZoom < 0.3 {
				newZoom = 0.3
			}
		}
		if isKeyDown(keys, KeyE) {
			newZoom *= 1.03
			if newZoom > 20.0 {
				newZoom = 20.0 // Allow much deeper zoom
			}
		}

		// Handle click to select star
		newSelectedStar := m.SelectedStar
		if input.ClickedThisFrame && !m.SkyViewMode {
			// Find nearest star to click position (20 pixel radius)
			catalog := GetStarCatalog()
			// Use the updated mode state for accurate position calculation
			tempMode := ModeGalaxyMap{
				CameraX:   newX,
				CameraY:   newY,
				ZoomLevel: newZoom,
			}
			clickedStar := FindNearestStarToScreen(catalog, input.Mouse.X, input.Mouse.Y, tempMode, 25.0)
			if clickedStar >= 0 {
				// Toggle selection - click again to deselect
				if newSelectedStar == clickedStar {
					newSelectedStar = -1
				} else {
					newSelectedStar = clickedStar
				}
			} else {
				// Clicked empty space - deselect
				newSelectedStar = -1
			}
		}

		return ModeGalaxyMap{
			CameraX:      newX,
			CameraY:      newY,
			ZoomLevel:    newZoom,
			SkyViewMode:  newSkyView,
			ViewLon:      m.ViewLon,
			ViewLat:      m.ViewLat,
			FOV:          m.FOV,
			SelectedStar: newSelectedStar,
			HoveredStar:  m.HoveredStar,
			ShowNetwork:  m.ShowNetwork,
		}

	default:
		return mode
	}
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
