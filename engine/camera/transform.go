// Package camera handles viewport transformations and culling.
package camera

import "stapledons_voyage/sim_gen"

// Transform holds the computed screen offset and scale for rendering.
type Transform struct {
	OffsetX float64 // Screen X offset for world origin
	OffsetY float64 // Screen Y offset for world origin
	Scale   float64 // Zoom factor
}

// FromOutput creates a Transform from a Camera and screen dimensions.
// The transform centers the camera position on screen.
func FromOutput(cam sim_gen.Camera, screenW, screenH int) Transform {
	// Camera position is world center, so offset = half screen - camera * zoom
	return Transform{
		OffsetX: float64(screenW)/2 - cam.X*cam.Zoom,
		OffsetY: float64(screenH)/2 - cam.Y*cam.Zoom,
		Scale:   cam.Zoom,
	}
}

// WorldToScreen converts world coordinates to screen coordinates.
func (t Transform) WorldToScreen(worldX, worldY float64) (screenX, screenY float64) {
	screenX = worldX*t.Scale + t.OffsetX
	screenY = worldY*t.Scale + t.OffsetY
	return
}

// ScreenToWorld converts screen coordinates to world coordinates.
func (t Transform) ScreenToWorld(screenX, screenY float64) (worldX, worldY float64) {
	worldX = (screenX - t.OffsetX) / t.Scale
	worldY = (screenY - t.OffsetY) / t.Scale
	return
}

// Apply returns the screen position for a world position.
// Convenience method that returns both values.
func (t Transform) Apply(worldX, worldY float64) (float64, float64) {
	return t.WorldToScreen(worldX, worldY)
}
