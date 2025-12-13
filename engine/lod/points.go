package lod

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

// PointRenderer handles efficient batch rendering of many points.
type PointRenderer struct {
	// pixel is a 1x1 white image for drawing colored points
	pixel *ebiten.Image
}

// NewPointRenderer creates a new point renderer.
func NewPointRenderer() *PointRenderer {
	// Create a 1x1 white pixel image
	pixel := ebiten.NewImage(1, 1)
	pixel.Fill(color.White)

	return &PointRenderer{
		pixel: pixel,
	}
}

// RenderPoints draws all point-tier objects as single colored pixels.
// Uses batched DrawImage calls for efficiency.
func (pr *PointRenderer) RenderPoints(screen *ebiten.Image, objects []*Object) {
	if len(objects) == 0 {
		return
	}

	for _, obj := range objects {
		if !obj.Visible {
			continue
		}

		opts := &ebiten.DrawImageOptions{}
		opts.GeoM.Translate(obj.ScreenX, obj.ScreenY)

		// Color the pixel
		r := float64(obj.Color.R) / 255
		g := float64(obj.Color.G) / 255
		b := float64(obj.Color.B) / 255
		a := float64(obj.Color.A) / 255
		opts.ColorScale.Scale(float32(r), float32(g), float32(b), float32(a))

		screen.DrawImage(pr.pixel, opts)
	}
}

// RenderPointsTriangles draws points using DrawTriangles for maximum performance.
// This is more efficient for very large numbers of points (10000+).
func (pr *PointRenderer) RenderPointsTriangles(screen *ebiten.Image, objects []*Object) {
	if len(objects) == 0 {
		return
	}

	// Count visible objects
	visibleCount := 0
	for _, obj := range objects {
		if obj.Visible {
			visibleCount++
		}
	}

	if visibleCount == 0 {
		return
	}

	// Build vertex and index arrays
	// Each point is 2 triangles (quad) = 4 vertices, 6 indices
	vertices := make([]ebiten.Vertex, visibleCount*4)
	indices := make([]uint16, visibleCount*6)

	vi := 0
	ii := 0
	for _, obj := range objects {
		if !obj.Visible {
			continue
		}

		x := float32(obj.ScreenX)
		y := float32(obj.ScreenY)
		r := float32(obj.Color.R) / 255
		g := float32(obj.Color.G) / 255
		b := float32(obj.Color.B) / 255
		a := float32(obj.Color.A) / 255

		// Create a 1x1 quad at the point position
		baseVertex := uint16(vi)

		// Top-left
		vertices[vi] = ebiten.Vertex{
			DstX:   x,
			DstY:   y,
			SrcX:   0,
			SrcY:   0,
			ColorR: r,
			ColorG: g,
			ColorB: b,
			ColorA: a,
		}
		vi++

		// Top-right
		vertices[vi] = ebiten.Vertex{
			DstX:   x + 1,
			DstY:   y,
			SrcX:   1,
			SrcY:   0,
			ColorR: r,
			ColorG: g,
			ColorB: b,
			ColorA: a,
		}
		vi++

		// Bottom-right
		vertices[vi] = ebiten.Vertex{
			DstX:   x + 1,
			DstY:   y + 1,
			SrcX:   1,
			SrcY:   1,
			ColorR: r,
			ColorG: g,
			ColorB: b,
			ColorA: a,
		}
		vi++

		// Bottom-left
		vertices[vi] = ebiten.Vertex{
			DstX:   x,
			DstY:   y + 1,
			SrcX:   0,
			SrcY:   1,
			ColorR: r,
			ColorG: g,
			ColorB: b,
			ColorA: a,
		}
		vi++

		// Two triangles
		indices[ii] = baseVertex
		indices[ii+1] = baseVertex + 1
		indices[ii+2] = baseVertex + 2
		indices[ii+3] = baseVertex
		indices[ii+4] = baseVertex + 2
		indices[ii+5] = baseVertex + 3
		ii += 6
	}

	// Draw all points in a single call
	screen.DrawTriangles(vertices, indices, pr.pixel, &ebiten.DrawTrianglesOptions{
		FillRule: ebiten.NonZero,
	})
}

// RenderPointsLarger draws points as 2x2 or 3x3 pixels for better visibility.
func (pr *PointRenderer) RenderPointsLarger(screen *ebiten.Image, objects []*Object, size int) {
	if len(objects) == 0 || size < 1 {
		return
	}

	bounds := screen.Bounds()
	screenW := bounds.Dx()
	screenH := bounds.Dy()

	for _, obj := range objects {
		if !obj.Visible {
			continue
		}

		x := int(obj.ScreenX)
		y := int(obj.ScreenY)

		// Draw size x size square of pixels
		half := size / 2
		for dy := -half; dy <= half; dy++ {
			for dx := -half; dx <= half; dx++ {
				px := x + dx
				py := y + dy
				if px >= 0 && px < screenW && py >= 0 && py < screenH {
					screen.Set(px, py, obj.Color)
				}
			}
		}
	}
}

// RenderPointsDirect draws points using direct pixel setting.
// Fastest for small numbers, avoids image allocation overhead.
func (pr *PointRenderer) RenderPointsDirect(screen *ebiten.Image, objects []*Object) {
	if len(objects) == 0 {
		return
	}

	bounds := screen.Bounds()
	screenW := bounds.Dx()
	screenH := bounds.Dy()

	for _, obj := range objects {
		if !obj.Visible {
			continue
		}

		x := int(obj.ScreenX)
		y := int(obj.ScreenY)

		if x >= 0 && x < screenW && y >= 0 && y < screenH {
			screen.Set(x, y, obj.Color)
		}
	}
}

// RenderPointsScaled draws points with size based on apparent radius.
// Points near the circle threshold are rendered as 2x2 or 3x3 for smoother transition.
// circleThreshold is the pixel radius at which objects become circles.
func (pr *PointRenderer) RenderPointsScaled(screen *ebiten.Image, objects []*Object, circleThreshold float64) {
	if len(objects) == 0 {
		return
	}

	bounds := screen.Bounds()
	screenW := bounds.Dx()
	screenH := bounds.Dy()

	for _, obj := range objects {
		if !obj.Visible {
			continue
		}

		x := int(obj.ScreenX)
		y := int(obj.ScreenY)

		// Determine point size based on how close to circle threshold
		// Scale from 1px to 3px as apparent radius approaches threshold
		var size int
		if circleThreshold > 0 {
			ratio := obj.ApparentRadius / circleThreshold
			if ratio > 0.7 {
				size = 3 // Large point near threshold
			} else if ratio > 0.4 {
				size = 2 // Medium point
			} else {
				size = 1 // Small point
			}
		} else {
			size = 1
		}

		// Draw point of appropriate size
		half := size / 2
		for dy := -half; dy <= half; dy++ {
			for dx := -half; dx <= half; dx++ {
				px := x + dx
				py := y + dy
				if px >= 0 && px < screenW && py >= 0 && py < screenH {
					screen.Set(px, py, obj.Color)
				}
			}
		}
	}
}

// RenderPointWithAlpha draws a single point with custom alpha.
// Used for smooth transitions between LOD tiers.
func (pr *PointRenderer) RenderPointWithAlpha(screen *ebiten.Image, obj *Object, alpha float64) {
	if !obj.Visible || alpha <= 0 {
		return
	}

	bounds := screen.Bounds()
	screenW := bounds.Dx()
	screenH := bounds.Dy()

	x := int(obj.ScreenX)
	y := int(obj.ScreenY)

	if x >= 0 && x < screenW && y >= 0 && y < screenH {
		col := color.RGBA{
			R: obj.Color.R,
			G: obj.Color.G,
			B: obj.Color.B,
			A: uint8(float64(obj.Color.A) * alpha),
		}
		screen.Set(x, y, col)
	}
}

// RenderPointsToImage renders points to a sub-image for compositing.
func (pr *PointRenderer) RenderPointsToImage(img *ebiten.Image, objects []*Object, offsetX, offsetY float64) {
	if len(objects) == 0 {
		return
	}

	bounds := img.Bounds()
	imgW := bounds.Dx()
	imgH := bounds.Dy()

	for _, obj := range objects {
		if !obj.Visible {
			continue
		}

		x := int(obj.ScreenX - offsetX)
		y := int(obj.ScreenY - offsetY)

		if x >= 0 && x < imgW && y >= 0 && y < imgH {
			img.Set(x, y, obj.Color)
		}
	}
}

// CreatePointAtlas creates a small image with pre-rendered point sizes.
// Returns an atlas with 1x1, 2x2, and 3x3 white squares.
func CreatePointAtlas() *ebiten.Image {
	// Atlas layout: 1x1 at (0,0), 2x2 at (2,0), 3x3 at (5,0)
	atlas := ebiten.NewImage(9, 4)
	atlas.Fill(color.Transparent)

	// Draw white squares at each size
	whitePixels := ebiten.NewImage(3, 3)
	whitePixels.Fill(color.White)

	// 1x1 at (0,0)
	atlas.DrawImage(whitePixels.SubImage(image.Rect(0, 0, 1, 1)).(*ebiten.Image), nil)

	// 2x2 at (2,0)
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(2, 0)
	atlas.DrawImage(whitePixels.SubImage(image.Rect(0, 0, 2, 2)).(*ebiten.Image), opts)

	// 3x3 at (5,0)
	opts = &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(5, 0)
	atlas.DrawImage(whitePixels, opts)

	return atlas
}
