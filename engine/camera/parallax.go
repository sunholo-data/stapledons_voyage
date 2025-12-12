// Package camera handles viewport transformations and culling.
package camera

import (
	"stapledons_voyage/engine/depth"
	"stapledons_voyage/engine/view/background"
)

// ParallaxCamera extends the basic camera with layer-aware parallax transformations.
// It provides different effective camera positions for different depth layers,
// creating a parallax effect where distant layers move slower than near ones.
type ParallaxCamera struct {
	// Base camera position (world coordinates)
	X, Y float64
	// Zoom level
	Zoom float64
	// Screen dimensions
	ScreenW, ScreenH int
}

// NewParallaxCamera creates a new parallax camera at the origin.
func NewParallaxCamera(screenW, screenH int) *ParallaxCamera {
	return &ParallaxCamera{
		X:       0,
		Y:       0,
		Zoom:    1.0,
		ScreenW: screenW,
		ScreenH: screenH,
	}
}

// SetPosition updates the camera's world position.
func (c *ParallaxCamera) SetPosition(x, y float64) {
	c.X = x
	c.Y = y
}

// SetZoom updates the camera's zoom level.
func (c *ParallaxCamera) SetZoom(zoom float64) {
	c.Zoom = zoom
}

// ForLayer returns the effective camera position for a specific depth layer.
// Layers with lower parallax factors move less, appearing further away.
func (c *ParallaxCamera) ForLayer(layer depth.Layer) (x, y float64) {
	factor := layer.Parallax()
	// Parallax: lower factor = moves less = appears further away
	return c.X * factor, c.Y * factor
}

// TransformForLayer returns a camera Transform adjusted for the given layer.
// This can be used with existing rendering code that expects a Transform.
func (c *ParallaxCamera) TransformForLayer(layer depth.Layer) Transform {
	x, y := c.ForLayer(layer)
	return Transform{
		OffsetX: float64(c.ScreenW)/2 - x*c.Zoom,
		OffsetY: float64(c.ScreenH)/2 - y*c.Zoom,
		Scale:   c.Zoom,
	}
}

// ForLayerOffset returns a CameraOffset for use with SpaceBackground.
// This allows the existing SpaceBackground to work with our parallax system.
func (c *ParallaxCamera) ForLayerOffset(layer depth.Layer) *background.CameraOffset {
	x, y := c.ForLayer(layer)
	return &background.CameraOffset{
		X:    x,
		Y:    y,
		Zoom: c.Zoom,
	}
}

// BaseTransform returns a Transform without parallax adjustment (factor 1.0).
// Use this for the main scene layer.
func (c *ParallaxCamera) BaseTransform() Transform {
	return Transform{
		OffsetX: float64(c.ScreenW)/2 - c.X*c.Zoom,
		OffsetY: float64(c.ScreenH)/2 - c.Y*c.Zoom,
		Scale:   c.Zoom,
	}
}

// WorldToScreen converts world coordinates to screen coordinates,
// adjusted for a specific layer's parallax.
func (c *ParallaxCamera) WorldToScreen(worldX, worldY float64, layer depth.Layer) (screenX, screenY float64) {
	t := c.TransformForLayer(layer)
	return t.WorldToScreen(worldX, worldY)
}

// ScreenToWorld converts screen coordinates to world coordinates.
// Note: This uses the base camera (no parallax), since screen clicks should
// map to the main scene layer.
func (c *ParallaxCamera) ScreenToWorld(screenX, screenY float64) (worldX, worldY float64) {
	t := c.BaseTransform()
	return t.ScreenToWorld(screenX, screenY)
}

// Resize updates the camera's screen dimensions.
func (c *ParallaxCamera) Resize(screenW, screenH int) {
	c.ScreenW = screenW
	c.ScreenH = screenH
}
