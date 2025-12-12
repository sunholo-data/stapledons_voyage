// Package render provides viewport shape definitions and mask generation
// for compositing different render sources through shaped regions.
package render

import (
	"image/color"
	"math"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// ViewportShape defines the interface for viewport shapes that can generate masks.
type ViewportShape interface {
	// GenerateMask creates a mask image where white (255) = visible, black (0) = masked.
	// The mask is generated at the specified width and height.
	GenerateMask(w, h int) *ebiten.Image

	// Contains returns true if the point (x, y) is inside the shape.
	// Coordinates are relative to the shape's bounding box origin.
	Contains(x, y float64) bool

	// Bounds returns the bounding box of the shape in local coordinates.
	Bounds() (x, y, w, h float64)
}

// maskCache stores generated masks to avoid regeneration.
var maskCache = struct {
	sync.RWMutex
	masks map[string]*ebiten.Image
}{masks: make(map[string]*ebiten.Image)}

// getMaskCacheKey generates a unique key for caching masks.
func getMaskCacheKey(shapeType string, w, h int, params ...float64) string {
	key := shapeType
	for _, p := range params {
		key += "_" + formatFloat(p)
	}
	key += "_" + formatInt(w) + "x" + formatInt(h)
	return key
}

func formatFloat(f float64) string {
	return string(rune(int(f*100) + '0'))
}

func formatInt(i int) string {
	if i < 10 {
		return string(rune(i + '0'))
	}
	return string(rune(i/10+'0')) + string(rune(i%10+'0'))
}

// getCachedMask retrieves a cached mask or returns nil if not found.
func getCachedMask(key string) *ebiten.Image {
	maskCache.RLock()
	defer maskCache.RUnlock()
	return maskCache.masks[key]
}

// setCachedMask stores a mask in the cache.
func setCachedMask(key string, mask *ebiten.Image) {
	maskCache.Lock()
	defer maskCache.Unlock()
	maskCache.masks[key] = mask
}

// ClearMaskCache clears all cached masks (call on resize).
func ClearMaskCache() {
	maskCache.Lock()
	defer maskCache.Unlock()
	maskCache.masks = make(map[string]*ebiten.Image)
}

// =============================================================================
// EllipseShape - Elliptical viewport
// =============================================================================

// EllipseShape defines an elliptical viewport.
type EllipseShape struct {
	CenterX, CenterY float64 // Center position relative to viewport origin
	RadiusX, RadiusY float64 // Horizontal and vertical radii
}

// GenerateMask creates a mask for the ellipse shape.
func (e *EllipseShape) GenerateMask(w, h int) *ebiten.Image {
	key := getMaskCacheKey("ellipse", w, h, e.CenterX, e.CenterY, e.RadiusX, e.RadiusY)
	if cached := getCachedMask(key); cached != nil {
		return cached
	}

	mask := ebiten.NewImage(w, h)

	// Draw filled ellipse using vector graphics
	var path vector.Path
	path.Arc(float32(e.CenterX), float32(e.CenterY), float32(e.RadiusX), 0, 2*math.Pi, vector.Clockwise)
	path.Close()

	// For non-circular ellipse, we need to scale
	if e.RadiusX != e.RadiusY {
		// Draw at unit scale then transform
		path = vector.Path{}
		steps := 64
		for i := 0; i <= steps; i++ {
			angle := float64(i) * 2 * math.Pi / float64(steps)
			x := e.CenterX + e.RadiusX*math.Cos(angle)
			y := e.CenterY + e.RadiusY*math.Sin(angle)
			if i == 0 {
				path.MoveTo(float32(x), float32(y))
			} else {
				path.LineTo(float32(x), float32(y))
			}
		}
		path.Close()
	}

	// Fill with white (visible)
	vs, is := path.AppendVerticesAndIndicesForFilling(nil, nil)
	for i := range vs {
		vs[i].ColorR = 1
		vs[i].ColorG = 1
		vs[i].ColorB = 1
		vs[i].ColorA = 1
	}
	mask.DrawTriangles(vs, is, emptyImage, nil)

	setCachedMask(key, mask)
	return mask
}

// Contains returns true if the point is inside the ellipse.
func (e *EllipseShape) Contains(x, y float64) bool {
	dx := (x - e.CenterX) / e.RadiusX
	dy := (y - e.CenterY) / e.RadiusY
	return dx*dx+dy*dy <= 1.0
}

// Bounds returns the bounding box of the ellipse.
func (e *EllipseShape) Bounds() (x, y, w, h float64) {
	return e.CenterX - e.RadiusX, e.CenterY - e.RadiusY, e.RadiusX * 2, e.RadiusY * 2
}

// =============================================================================
// CircleShape - Circular viewport (special case of ellipse)
// =============================================================================

// CircleShape defines a circular viewport.
type CircleShape struct {
	CenterX, CenterY float64
	Radius           float64
}

// GenerateMask creates a mask for the circle shape.
func (c *CircleShape) GenerateMask(w, h int) *ebiten.Image {
	key := getMaskCacheKey("circle", w, h, c.CenterX, c.CenterY, c.Radius)
	if cached := getCachedMask(key); cached != nil {
		return cached
	}

	mask := ebiten.NewImage(w, h)

	// Draw filled circle
	vector.DrawFilledCircle(mask, float32(c.CenterX), float32(c.CenterY), float32(c.Radius), color.White, true)

	setCachedMask(key, mask)
	return mask
}

// Contains returns true if the point is inside the circle.
func (c *CircleShape) Contains(x, y float64) bool {
	dx := x - c.CenterX
	dy := y - c.CenterY
	return dx*dx+dy*dy <= c.Radius*c.Radius
}

// Bounds returns the bounding box of the circle.
func (c *CircleShape) Bounds() (x, y, w, h float64) {
	return c.CenterX - c.Radius, c.CenterY - c.Radius, c.Radius * 2, c.Radius * 2
}

// =============================================================================
// RectShape - Rectangular viewport
// =============================================================================

// RectShape defines a rectangular viewport.
type RectShape struct {
	X, Y          float64
	Width, Height float64
}

// GenerateMask creates a mask for the rectangle shape.
func (r *RectShape) GenerateMask(w, h int) *ebiten.Image {
	key := getMaskCacheKey("rect", w, h, r.X, r.Y, r.Width, r.Height)
	if cached := getCachedMask(key); cached != nil {
		return cached
	}

	mask := ebiten.NewImage(w, h)

	// Draw filled rectangle
	vector.DrawFilledRect(mask, float32(r.X), float32(r.Y), float32(r.Width), float32(r.Height), color.White, true)

	setCachedMask(key, mask)
	return mask
}

// Contains returns true if the point is inside the rectangle.
func (r *RectShape) Contains(x, y float64) bool {
	return x >= r.X && x <= r.X+r.Width && y >= r.Y && y <= r.Y+r.Height
}

// Bounds returns the bounding box of the rectangle.
func (r *RectShape) Bounds() (x, y, w, h float64) {
	return r.X, r.Y, r.Width, r.Height
}

// =============================================================================
// DomeShape - Dome viewport (rectangular bottom + elliptical arch top)
// =============================================================================

// DomeShape defines a dome-shaped viewport with a rectangular body and curved arch.
//
//	       ╭─────────────────╮  ← Curved arch (elliptical)
//	      ╱                   ╲
//	     ╱                     ╲
//	    │                       │
//	    │    SPACE CONTENT      │  ← Rectangular body
//	    │                       │
//	    └───────────────────────┘  ← Flat bottom (frame edge)
type DomeShape struct {
	CenterX, CenterY float64 // Center of the dome base
	Width, Height    float64 // Overall dimensions
	ArchHeight       float64 // Height of the curved arch portion
}

// GenerateMask creates a mask for the dome shape.
func (d *DomeShape) GenerateMask(w, h int) *ebiten.Image {
	key := getMaskCacheKey("dome", w, h, d.CenterX, d.CenterY, d.Width, d.Height, d.ArchHeight)
	if cached := getCachedMask(key); cached != nil {
		return cached
	}

	mask := ebiten.NewImage(w, h)

	// Calculate dome geometry
	left := d.CenterX - d.Width/2
	right := d.CenterX + d.Width/2
	bottom := d.CenterY + d.Height/2
	archStart := d.CenterY + d.Height/2 - d.ArchHeight // Where arch begins

	// Build path: start at bottom-left, go up, arch across, down, close
	var path vector.Path

	// Start at bottom-left
	path.MoveTo(float32(left), float32(bottom))

	// Left edge up to arch start
	path.LineTo(float32(left), float32(archStart))

	// Arch curve (using quadratic bezier for smooth dome)
	// Control point is at the top center
	controlX := d.CenterX
	controlY := d.CenterY - d.Height/2 // Top of dome

	// Use multiple segments for smoother curve
	steps := 32
	for i := 1; i <= steps; i++ {
		t := float64(i) / float64(steps)
		// Quadratic bezier: B(t) = (1-t)²P0 + 2(1-t)tP1 + t²P2
		x := (1-t)*(1-t)*left + 2*(1-t)*t*controlX + t*t*right
		y := (1-t)*(1-t)*archStart + 2*(1-t)*t*controlY + t*t*archStart
		path.LineTo(float32(x), float32(y))
	}

	// Right edge down to bottom
	path.LineTo(float32(right), float32(bottom))

	// Close path (back to start)
	path.Close()

	// Fill with white (visible)
	vs, is := path.AppendVerticesAndIndicesForFilling(nil, nil)
	for i := range vs {
		vs[i].ColorR = 1
		vs[i].ColorG = 1
		vs[i].ColorB = 1
		vs[i].ColorA = 1
	}
	mask.DrawTriangles(vs, is, emptyImage, nil)

	setCachedMask(key, mask)
	return mask
}

// Contains returns true if the point is inside the dome.
func (d *DomeShape) Contains(x, y float64) bool {
	// Check rectangular body first
	left := d.CenterX - d.Width/2
	right := d.CenterX + d.Width/2
	bottom := d.CenterY + d.Height/2
	archStart := d.CenterY + d.Height/2 - d.ArchHeight

	// Outside horizontal bounds
	if x < left || x > right {
		return false
	}

	// Below dome (outside)
	if y > bottom {
		return false
	}

	// In rectangular body
	if y >= archStart {
		return true
	}

	// Check if in arch (approximate with parabola)
	// Normalize x to [-1, 1] range
	nx := (x - d.CenterX) / (d.Width / 2)
	// Parabola: y = archStart - archHeight * (1 - x²)
	archTop := archStart - d.ArchHeight*(1-nx*nx)
	return y >= archTop
}

// Bounds returns the bounding box of the dome.
func (d *DomeShape) Bounds() (x, y, w, h float64) {
	return d.CenterX - d.Width/2, d.CenterY - d.Height/2, d.Width, d.Height
}

// =============================================================================
// Helper for drawing
// =============================================================================

// emptyImage is a 1x1 white image used for drawing colored shapes.
var emptyImage = func() *ebiten.Image {
	img := ebiten.NewImage(1, 1)
	img.Fill(color.White)
	return img
}()
