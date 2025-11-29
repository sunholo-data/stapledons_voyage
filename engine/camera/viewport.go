package camera

import "stapledons_voyage/sim_gen"

// Viewport represents the visible area in world coordinates.
type Viewport struct {
	MinX float64 // Left edge in world coordinates
	MaxX float64 // Right edge in world coordinates
	MinY float64 // Top edge in world coordinates
	MaxY float64 // Bottom edge in world coordinates
}

// CalculateViewport computes the visible world area from camera and screen size.
func CalculateViewport(cam sim_gen.Camera, screenW, screenH int) Viewport {
	// Half screen dimensions in world coordinates (accounting for zoom)
	halfW := float64(screenW) / 2 / cam.Zoom
	halfH := float64(screenH) / 2 / cam.Zoom

	return Viewport{
		MinX: cam.X - halfW,
		MaxX: cam.X + halfW,
		MinY: cam.Y - halfH,
		MaxY: cam.Y + halfH,
	}
}

// Contains checks if a point (with margin) is within the viewport.
// Margin is used to include objects partially on-screen.
func (v Viewport) Contains(x, y, margin float64) bool {
	return x >= v.MinX-margin && x <= v.MaxX+margin &&
		y >= v.MinY-margin && y <= v.MaxY+margin
}

// ContainsRect checks if a rectangle overlaps with the viewport.
func (v Viewport) ContainsRect(x, y, w, h float64) bool {
	// Rectangle overlaps if it's not completely outside
	return x+w >= v.MinX && x <= v.MaxX &&
		y+h >= v.MinY && y <= v.MaxY
}

// Width returns the viewport width in world coordinates.
func (v Viewport) Width() float64 {
	return v.MaxX - v.MinX
}

// Height returns the viewport height in world coordinates.
func (v Viewport) Height() float64 {
	return v.MaxY - v.MinY
}
