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
func processMode(mode WorldMode, keys []KeyEvent) WorldMode {
	switch m := mode.(type) {
	case ModeShipExploration:
		// M key opens galaxy map
		if wasKeyPressed(keys, KeyM) {
			return ModeGalaxyMap{
				CameraX:      0,
				CameraY:      0,
				ZoomLevel:    1.0,
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
		// Update galaxy map camera with keys
		newX, newY := m.CameraX, m.CameraY
		newZoom := m.ZoomLevel
		if isKeyDown(keys, KeyW) || isKeyDown(keys, KeyArrowUp) {
			newY -= 5.0
		}
		if isKeyDown(keys, KeyS) || isKeyDown(keys, KeyArrowDown) {
			newY += 5.0
		}
		if isKeyDown(keys, KeyA) || isKeyDown(keys, KeyArrowLeft) {
			newX -= 5.0
		}
		if isKeyDown(keys, KeyD) || isKeyDown(keys, KeyArrowRight) {
			newX += 5.0
		}
		if isKeyDown(keys, KeyQ) {
			newZoom *= 0.98
			if newZoom < 0.25 {
				newZoom = 0.25
			}
		}
		if isKeyDown(keys, KeyE) {
			newZoom *= 1.02
			if newZoom > 4.0 {
				newZoom = 4.0
			}
		}
		return ModeGalaxyMap{
			CameraX:      newX,
			CameraY:      newY,
			ZoomLevel:    newZoom,
			SelectedStar: m.SelectedStar,
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
