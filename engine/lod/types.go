package lod

import (
	"image/color"
	"math"
)

// Vector3 represents a 3D position.
type Vector3 struct {
	X, Y, Z float64
}

// Distance returns the Euclidean distance between two vectors.
func (v Vector3) Distance(other Vector3) float64 {
	dx := v.X - other.X
	dy := v.Y - other.Y
	dz := v.Z - other.Z
	return math.Sqrt(dx*dx + dy*dy + dz*dz)
}

// Object represents a celestial object that can be rendered at different LOD levels.
type Object struct {
	// ID is a unique identifier for this object
	ID string

	// Position in 3D world space
	Position Vector3

	// Radius of the object (used for apparent size calculation)
	Radius float64

	// Color for rendering as circle or point
	Color color.RGBA

	// CurrentTier is the currently assigned LOD tier
	CurrentTier LODTier

	// Distance to camera (cached after Update)
	Distance float64

	// ScreenX, ScreenY are projected screen coordinates (set by projection)
	ScreenX, ScreenY float64

	// ApparentRadius is the screen-space size based on distance
	ApparentRadius float64

	// Visible indicates if the object is within the view frustum
	Visible bool
}

// NewObject creates a new LOD object with the given parameters.
func NewObject(id string, pos Vector3, radius float64, col color.RGBA) *Object {
	return &Object{
		ID:          id,
		Position:    pos,
		Radius:      radius,
		Color:       col,
		CurrentTier: TierCulled,
	}
}

// Stats tracks LOD rendering statistics.
type Stats struct {
	// Total objects managed
	TotalObjects int

	// Count per tier
	Full3DCount    int
	BillboardCount int
	CircleCount    int
	PointCount     int
	CulledCount    int

	// Objects visible (not culled by frustum or distance)
	VisibleCount int
}

// Reset clears all stats.
func (s *Stats) Reset() {
	s.TotalObjects = 0
	s.Full3DCount = 0
	s.BillboardCount = 0
	s.CircleCount = 0
	s.PointCount = 0
	s.CulledCount = 0
	s.VisibleCount = 0
}

// Camera provides the projection interface needed by the LOD manager.
type Camera interface {
	// Position returns the camera's world position
	Position() Vector3
	// WorldToScreen projects a 3D world position to 2D screen coordinates
	// Returns x, y screen coordinates and whether the point is in front of camera
	WorldToScreen(world Vector3) (x, y float64, visible bool)
	// FOVScale returns a factor for calculating apparent size
	FOVScale() float64
	// ScreenWidth returns the screen width in pixels
	ScreenWidth() int
	// ScreenHeight returns the screen height in pixels
	ScreenHeight() int
}

// SimpleCamera is a basic camera implementation for testing.
type SimpleCamera struct {
	Pos          Vector3
	LookAt       Vector3
	Fov          float64
	Width        int
	Height       int
	Near, Far    float64
}

// NewSimpleCamera creates a camera with reasonable defaults.
func NewSimpleCamera(width, height int) *SimpleCamera {
	return &SimpleCamera{
		Pos:    Vector3{0, 0, 10},
		LookAt: Vector3{0, 0, 0},
		Fov:    60,
		Width:  width,
		Height: height,
		Near:   0.1,
		Far:    10000,
	}
}

// Position returns the camera's world position.
func (c *SimpleCamera) Position() Vector3 {
	return c.Pos
}

// WorldToScreen projects a 3D point to screen space.
// This is a simplified perspective projection.
func (c *SimpleCamera) WorldToScreen(world Vector3) (x, y float64, visible bool) {
	// Vector from camera to object
	dx := world.X - c.Pos.X
	dy := world.Y - c.Pos.Y
	dz := world.Z - c.Pos.Z

	// Simple forward direction (camera looks toward LookAt)
	forwardX := c.LookAt.X - c.Pos.X
	forwardY := c.LookAt.Y - c.Pos.Y
	forwardZ := c.LookAt.Z - c.Pos.Z
	forwardLen := math.Sqrt(forwardX*forwardX + forwardY*forwardY + forwardZ*forwardZ)
	if forwardLen > 0 {
		forwardX /= forwardLen
		forwardY /= forwardLen
		forwardZ /= forwardLen
	}

	// Dot product gives depth
	depth := dx*forwardX + dy*forwardY + dz*forwardZ

	// Behind camera
	if depth < c.Near {
		return 0, 0, false
	}

	// Beyond far plane
	if depth > c.Far {
		return 0, 0, false
	}

	// Simple perspective: project onto plane at distance 1
	// This is a rough approximation
	fovRad := c.Fov * math.Pi / 180
	scale := float64(c.Height) / (2 * math.Tan(fovRad/2) * depth)

	// Right vector (cross product of forward and up)
	upX, upY, upZ := 0.0, 1.0, 0.0
	rightX := forwardY*upZ - forwardZ*upY
	rightY := forwardZ*upX - forwardX*upZ
	rightZ := forwardX*upY - forwardY*upX

	// Up vector (cross product of right and forward)
	actualUpX := rightY*forwardZ - rightZ*forwardY
	actualUpY := rightZ*forwardX - rightX*forwardZ
	actualUpZ := rightX*forwardY - rightY*forwardX

	// Project onto right and up
	rightDot := dx*rightX + dy*rightY + dz*rightZ
	upDot := dx*actualUpX + dy*actualUpY + dz*actualUpZ

	x = float64(c.Width)/2 + rightDot*scale
	y = float64(c.Height)/2 - upDot*scale

	// Check if on screen (with some margin)
	margin := 100.0
	visible = x >= -margin && x < float64(c.Width)+margin &&
		y >= -margin && y < float64(c.Height)+margin

	return x, y, visible
}

// FOVScale returns a factor for apparent size calculation.
func (c *SimpleCamera) FOVScale() float64 {
	fovRad := c.Fov * math.Pi / 180
	return float64(c.Height) / (2 * math.Tan(fovRad/2))
}

// ScreenWidth returns the screen width.
func (c *SimpleCamera) ScreenWidth() int {
	return c.Width
}

// ScreenHeight returns the screen height.
func (c *SimpleCamera) ScreenHeight() int {
	return c.Height
}
