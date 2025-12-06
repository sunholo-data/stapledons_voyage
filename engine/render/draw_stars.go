package render

import (
	"image"
	"image/color"
	_ "image/jpeg" // Register JPEG decoder
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"stapledons_voyage/sim_gen"
)

// drawStar draws a star sprite with scaling for efficient GPU batching.
// Falls back to colored circle if sprite not available.
func (r *Renderer) drawStar(screen *ebiten.Image, c *sim_gen.DrawCmdStar) {
	// Default alpha to 1.0 if not set
	alpha := c.Alpha
	if alpha <= 0 {
		alpha = 1.0
	}

	if r.assets == nil {
		// Fallback: draw colored circle
		r.drawStarFallback(screen, c, alpha)
		return
	}

	sprite := r.assets.GetSprite(int(c.SpriteId))
	if sprite == nil {
		r.drawStarFallback(screen, c, alpha)
		return
	}

	// Get sprite dimensions
	bounds := sprite.Bounds()
	sw := float64(bounds.Dx())
	sh := float64(bounds.Dy())

	// Center the sprite on the position
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(-sw/2, -sh/2) // Center origin
	op.GeoM.Scale(c.Scale, c.Scale)
	op.GeoM.Translate(c.X, c.Y)

	// Apply alpha for depth-based opacity
	if alpha < 1.0 {
		op.ColorScale.ScaleAlpha(float32(alpha))
	}

	screen.DrawImage(sprite, op)
}

// drawStarFallback draws a colored circle when sprite not available
func (r *Renderer) drawStarFallback(screen *ebiten.Image, c *sim_gen.DrawCmdStar, alpha float64) {
	// Map sprite ID to color
	var col color.RGBA
	switch c.SpriteId {
	case 200: // Blue (O/B)
		col = color.RGBA{155, 176, 255, 255}
	case 201: // White (A/F)
		col = color.RGBA{255, 255, 255, 255}
	case 202: // Yellow (G)
		col = color.RGBA{255, 244, 214, 255}
	case 203: // Orange (K)
		col = color.RGBA{255, 210, 161, 255}
	case 204: // Red (M)
		col = color.RGBA{255, 189, 189, 255}
	default:
		col = color.RGBA{255, 255, 255, 255}
	}

	// Apply alpha to color
	col.A = uint8(255 * alpha)

	// Calculate radius from scale (base sprite is 16x16)
	radius := float32(c.Scale * 8)
	if radius < 1 {
		radius = 1
	}

	vector.DrawFilledCircle(screen, float32(c.X), float32(c.Y), radius, col, true)
}

// drawGalaxyBackground renders the galaxy background image with the given opacity.
// For sky view mode, it scrolls the equirectangular image based on ViewLon/ViewLat/FOV.
func (r *Renderer) drawGalaxyBackground(screen *ebiten.Image, opacity float64, screenW, screenH int, skyViewMode bool, viewLon, viewLat, fov float64) {
	// Lazy-load the galaxy background
	if !r.galaxyBgLoaded {
		r.galaxyBgLoaded = true
		r.loadGalaxyBackground()
	}

	if r.galaxyBg == nil {
		return
	}

	bgW := float64(r.galaxyBg.Bounds().Dx())
	bgH := float64(r.galaxyBg.Bounds().Dy())

	if skyViewMode {
		// Sky view: scroll the equirectangular image based on view direction
		// The image is 360° wide (longitude) and 180° tall (latitude from -90 to +90)

		// Calculate the portion of the image to show based on FOV
		// Horizontal FOV maps to longitude range
		// Vertical FOV is FOV * (screenH/screenW) to maintain aspect ratio
		hFOV := fov
		vFOV := fov * float64(screenH) / float64(screenW)

		// Map view direction to image coordinates
		// Longitude 0° is at image center (x = bgW/2), wraps around
		// Latitude +90° is at top (y = 0), -90° at bottom (y = bgH)

		// Calculate source rectangle in image coordinates
		// Center of view in image coords
		centerX := (viewLon / 360.0) * bgW
		centerY := ((90.0 - viewLat) / 180.0) * bgH

		// Size of source rectangle (how much of image to show)
		srcW := (hFOV / 360.0) * bgW
		srcH := (vFOV / 180.0) * bgH

		// Source rectangle bounds
		srcX := centerX - srcW/2
		srcY := centerY - srcH/2

		// Handle wrapping for longitude (X)
		// For simplicity, if we're near the edge, just clamp
		if srcX < 0 {
			srcX = 0
		}
		if srcX+srcW > bgW {
			srcX = bgW - srcW
		}

		// Clamp latitude (Y) - no wrapping
		if srcY < 0 {
			srcY = 0
		}
		if srcY+srcH > bgH {
			srcY = bgH - srcH
		}

		// Ensure minimum size
		if srcW < 10 {
			srcW = 10
		}
		if srcH < 10 {
			srcH = 10
		}

		// Create sub-image for the visible portion
		subImg := r.galaxyBg.SubImage(image.Rect(
			int(srcX), int(srcY),
			int(srcX+srcW), int(srcY+srcH),
		)).(*ebiten.Image)

		// Scale to fill screen
		scaleX := float64(screenW) / srcW
		scaleY := float64(screenH) / srcH

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(scaleX, scaleY)
		op.ColorScale.Scale(float32(opacity), float32(opacity), float32(opacity), 1.0)

		screen.DrawImage(subImg, op)
	} else {
		// Plane view: show entire image centered and scaled to fit
		scaleX := float64(screenW) / bgW
		scaleY := float64(screenH) / bgH
		scale := scaleX
		if scaleY > scaleX {
			scale = scaleY // Use larger scale to cover screen
		}

		// Center the image
		drawW := bgW * scale
		drawH := bgH * scale
		offsetX := (float64(screenW) - drawW) / 2
		offsetY := (float64(screenH) - drawH) / 2

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(scale, scale)
		op.GeoM.Translate(offsetX, offsetY)

		// Apply opacity (dim the background)
		op.ColorScale.Scale(float32(opacity), float32(opacity), float32(opacity), 1.0)

		screen.DrawImage(r.galaxyBg, op)
	}
}

// loadGalaxyBackground loads the galaxy background image from disk.
func (r *Renderer) loadGalaxyBackground() {
	// Try different paths for the galaxy background
	paths := []string{
		"assets/data/starmap/background/galaxy_4k.jpg",
		"assets/data/starmap/background/galaxy_2k.jpg",
		"assets/data/starmap/background/galaxy_8k.jpg",
	}

	for _, path := range paths {
		f, err := os.Open(path)
		if err != nil {
			continue
		}

		img, _, err := image.Decode(f)
		f.Close()
		if err != nil {
			continue
		}

		r.galaxyBg = ebiten.NewImageFromImage(img)
		return
	}
}
