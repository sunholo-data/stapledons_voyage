package lod

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// CircleRenderer handles efficient rendering of filled circles.
type CircleRenderer struct {
	// MinRadius is the minimum rendered radius (prevents invisible circles)
	MinRadius float64
	// MaxRadius is the maximum rendered radius (prevents giant circles)
	MaxRadius float64
	// AntiAlias enables anti-aliasing for smoother circles
	AntiAlias bool
}

// NewCircleRenderer creates a new circle renderer with default settings.
func NewCircleRenderer() *CircleRenderer {
	return &CircleRenderer{
		MinRadius: 1.0,  // Allow small circles for smoother point→circle transition
		MaxRadius: 200.0,
		AntiAlias: true,
	}
}

// RenderCircles draws all circle-tier objects as filled circles.
// Circle size is based on apparent radius (object radius / distance * fov scale).
func (cr *CircleRenderer) RenderCircles(screen *ebiten.Image, objects []*Object) {
	if len(objects) == 0 {
		return
	}

	for _, obj := range objects {
		if !obj.Visible {
			continue
		}

		// Clamp apparent radius to reasonable bounds
		radius := obj.ApparentRadius
		if radius < cr.MinRadius {
			radius = cr.MinRadius
		}
		if radius > cr.MaxRadius {
			radius = cr.MaxRadius
		}

		// Draw filled circle
		vector.DrawFilledCircle(
			screen,
			float32(obj.ScreenX),
			float32(obj.ScreenY),
			float32(radius),
			obj.Color,
			cr.AntiAlias,
		)
	}
}

// RenderCirclesWithGlow draws circles with an outer glow effect.
// Useful for stars that should appear to have corona/glow.
func (cr *CircleRenderer) RenderCirclesWithGlow(screen *ebiten.Image, objects []*Object, glowScale float64) {
	if len(objects) == 0 {
		return
	}

	for _, obj := range objects {
		if !obj.Visible {
			continue
		}

		radius := obj.ApparentRadius
		if radius < cr.MinRadius {
			radius = cr.MinRadius
		}
		if radius > cr.MaxRadius {
			radius = cr.MaxRadius
		}

		// Draw outer glow (larger, semi-transparent)
		glowRadius := radius * glowScale
		glowColor := color.RGBA{
			R: obj.Color.R,
			G: obj.Color.G,
			B: obj.Color.B,
			A: uint8(float64(obj.Color.A) * 0.3), // 30% opacity
		}
		vector.DrawFilledCircle(
			screen,
			float32(obj.ScreenX),
			float32(obj.ScreenY),
			float32(glowRadius),
			glowColor,
			true,
		)

		// Draw inner core (full opacity)
		vector.DrawFilledCircle(
			screen,
			float32(obj.ScreenX),
			float32(obj.ScreenY),
			float32(radius),
			obj.Color,
			cr.AntiAlias,
		)
	}
}

// RenderCirclesTriangles draws circles using DrawTriangles for better performance.
// Each circle is approximated as a polygon with the specified number of segments.
func (cr *CircleRenderer) RenderCirclesTriangles(screen *ebiten.Image, objects []*Object, segments int) {
	if len(objects) == 0 || segments < 3 {
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

	// Each circle needs: 1 center vertex + segments edge vertices
	// And segments triangles (each 3 indices)
	verticesPerCircle := segments + 1
	indicesPerCircle := segments * 3

	vertices := make([]ebiten.Vertex, visibleCount*verticesPerCircle)
	indices := make([]uint16, visibleCount*indicesPerCircle)

	// Create a white pixel for source image
	pixel := ebiten.NewImage(1, 1)
	pixel.Fill(color.White)

	vi := 0
	ii := 0
	for _, obj := range objects {
		if !obj.Visible {
			continue
		}

		radius := obj.ApparentRadius
		if radius < cr.MinRadius {
			radius = cr.MinRadius
		}
		if radius > cr.MaxRadius {
			radius = cr.MaxRadius
		}

		cx := float32(obj.ScreenX)
		cy := float32(obj.ScreenY)
		r := float32(obj.Color.R) / 255
		g := float32(obj.Color.G) / 255
		b := float32(obj.Color.B) / 255
		a := float32(obj.Color.A) / 255

		baseVertex := uint16(vi)

		// Center vertex
		vertices[vi] = ebiten.Vertex{
			DstX:   cx,
			DstY:   cy,
			SrcX:   0.5,
			SrcY:   0.5,
			ColorR: r,
			ColorG: g,
			ColorB: b,
			ColorA: a,
		}
		vi++

		// Edge vertices
		for s := 0; s < segments; s++ {
			angle := float64(s) * 2 * math.Pi / float64(segments)
			px := cx + float32(radius*math.Cos(angle))
			py := cy + float32(radius*math.Sin(angle))

			vertices[vi] = ebiten.Vertex{
				DstX:   px,
				DstY:   py,
				SrcX:   0.5,
				SrcY:   0.5,
				ColorR: r,
				ColorG: g,
				ColorB: b,
				ColorA: a,
			}
			vi++
		}

		// Create triangles from center to each edge segment
		for s := 0; s < segments; s++ {
			next := (s + 1) % segments
			indices[ii] = baseVertex             // Center
			indices[ii+1] = baseVertex + uint16(s+1)   // Current edge
			indices[ii+2] = baseVertex + uint16(next+1) // Next edge
			ii += 3
		}
	}

	// Draw all circles in a single call
	screen.DrawTriangles(vertices, indices, pixel, &ebiten.DrawTrianglesOptions{
		FillRule: ebiten.NonZero,
	})
}

// RenderRings draws objects as ring outlines instead of filled circles.
// Useful for orbital paths or selection indicators.
func (cr *CircleRenderer) RenderRings(screen *ebiten.Image, objects []*Object, strokeWidth float32) {
	if len(objects) == 0 {
		return
	}

	for _, obj := range objects {
		if !obj.Visible {
			continue
		}

		radius := obj.ApparentRadius
		if radius < cr.MinRadius {
			radius = cr.MinRadius
		}
		if radius > cr.MaxRadius {
			radius = cr.MaxRadius
		}

		vector.StrokeCircle(
			screen,
			float32(obj.ScreenX),
			float32(obj.ScreenY),
			float32(radius),
			strokeWidth,
			obj.Color,
			cr.AntiAlias,
		)
	}
}

// RenderCircleWithAlpha draws a single object with a custom alpha multiplier.
// Used for smooth transitions between LOD tiers.
func (cr *CircleRenderer) RenderCircleWithAlpha(screen *ebiten.Image, obj *Object, alpha float64) {
	if !obj.Visible || alpha <= 0 {
		return
	}

	radius := obj.ApparentRadius
	if radius < cr.MinRadius {
		radius = cr.MinRadius
	}
	if radius > cr.MaxRadius {
		radius = cr.MaxRadius
	}

	// Apply alpha to color
	col := color.RGBA{
		R: obj.Color.R,
		G: obj.Color.G,
		B: obj.Color.B,
		A: uint8(float64(obj.Color.A) * alpha),
	}

	vector.DrawFilledCircle(
		screen,
		float32(obj.ScreenX),
		float32(obj.ScreenY),
		float32(radius),
		col,
		cr.AntiAlias,
	)
}

// CalcApparentRadius calculates the screen-space radius for an object.
// fovScale is typically: screenHeight / (2 * tan(fov/2))
func CalcApparentRadius(worldRadius, distance, fovScale float64) float64 {
	if distance <= 0 {
		return fovScale // Max size when at zero distance
	}
	return (worldRadius / distance) * fovScale
}

// BrightnessFromDistance calculates a brightness factor based on distance.
// Uses inverse square falloff for realistic light dimming.
// Returns a value between 0.0 and 1.0.
func BrightnessFromDistance(distance, minDist, maxDist float64) float64 {
	if distance <= minDist {
		return 1.0
	}
	if distance >= maxDist {
		return 0.0
	}

	// Inverse square falloff
	normalized := (distance - minDist) / (maxDist - minDist)
	return 1.0 / (1.0 + normalized*normalized*4)
}

// ColorWithBrightness returns a color with adjusted brightness.
func ColorWithBrightness(c color.RGBA, brightness float64) color.RGBA {
	if brightness >= 1.0 {
		return c
	}
	if brightness <= 0.0 {
		return color.RGBA{0, 0, 0, c.A}
	}

	return color.RGBA{
		R: uint8(float64(c.R) * brightness),
		G: uint8(float64(c.G) * brightness),
		B: uint8(float64(c.B) * brightness),
		A: c.A,
	}
}
