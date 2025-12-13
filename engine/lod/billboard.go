package lod

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

// BillboardRenderer handles rendering of camera-facing 2D sprites.
type BillboardRenderer struct {
	// DefaultSprite is used when an object doesn't have a specific sprite
	DefaultSprite *ebiten.Image

	// MinScale prevents billboards from being too small
	MinScale float64
	// MaxScale prevents billboards from being too large
	MaxScale float64
}

// NewBillboardRenderer creates a new billboard renderer.
func NewBillboardRenderer() *BillboardRenderer {
	return &BillboardRenderer{
		MinScale: 0.1,
		MaxScale: 10.0,
	}
}

// SetDefaultSprite sets the fallback sprite for objects without specific sprites.
func (br *BillboardRenderer) SetDefaultSprite(sprite *ebiten.Image) {
	br.DefaultSprite = sprite
}

// BillboardObject extends Object with sprite data for billboard rendering.
type BillboardObject struct {
	*Object
	Sprite *ebiten.Image
}

// RenderBillboards draws all billboard-tier objects as camera-facing sprites.
// Sprites are scaled based on apparent radius and centered on screen position.
func (br *BillboardRenderer) RenderBillboards(screen *ebiten.Image, objects []*Object, sprites map[string]*ebiten.Image) {
	if len(objects) == 0 {
		return
	}

	for _, obj := range objects {
		if !obj.Visible {
			continue
		}

		// Get sprite for this object (or use default)
		sprite := br.DefaultSprite
		if sprites != nil {
			if s, ok := sprites[obj.ID]; ok && s != nil {
				sprite = s
			}
		}

		if sprite == nil {
			// No sprite available, skip
			continue
		}

		// Calculate scale based on apparent radius
		// Sprite should appear at approximately the same size as a circle would
		spriteW := float64(sprite.Bounds().Dx())
		spriteH := float64(sprite.Bounds().Dy())
		spriteSize := max(spriteW, spriteH)

		targetSize := obj.ApparentRadius * 2 // Diameter
		scale := targetSize / spriteSize

		// Clamp scale
		if scale < br.MinScale {
			scale = br.MinScale
		}
		if scale > br.MaxScale {
			scale = br.MaxScale
		}

		// Draw sprite centered on screen position
		opts := &ebiten.DrawImageOptions{}

		// Center the sprite
		opts.GeoM.Translate(-spriteW/2, -spriteH/2)

		// Scale to apparent size
		opts.GeoM.Scale(scale, scale)

		// Position at screen coordinates
		opts.GeoM.Translate(obj.ScreenX, obj.ScreenY)

		// Apply color tint if object has non-white color
		if obj.Color.R != 255 || obj.Color.G != 255 || obj.Color.B != 255 {
			r := float64(obj.Color.R) / 255
			g := float64(obj.Color.G) / 255
			b := float64(obj.Color.B) / 255
			a := float64(obj.Color.A) / 255
			opts.ColorScale.Scale(float32(r), float32(g), float32(b), float32(a))
		}

		screen.DrawImage(sprite, opts)
	}
}

// RenderBillboardsWithBillboardObjects draws billboard objects that have embedded sprites.
func (br *BillboardRenderer) RenderBillboardsWithBillboardObjects(screen *ebiten.Image, objects []*BillboardObject) {
	if len(objects) == 0 {
		return
	}

	for _, obj := range objects {
		if !obj.Visible {
			continue
		}

		sprite := obj.Sprite
		if sprite == nil {
			sprite = br.DefaultSprite
		}
		if sprite == nil {
			continue
		}

		spriteW := float64(sprite.Bounds().Dx())
		spriteH := float64(sprite.Bounds().Dy())
		spriteSize := max(spriteW, spriteH)

		targetSize := obj.ApparentRadius * 2
		scale := targetSize / spriteSize

		if scale < br.MinScale {
			scale = br.MinScale
		}
		if scale > br.MaxScale {
			scale = br.MaxScale
		}

		opts := &ebiten.DrawImageOptions{}
		opts.GeoM.Translate(-spriteW/2, -spriteH/2)
		opts.GeoM.Scale(scale, scale)
		opts.GeoM.Translate(obj.ScreenX, obj.ScreenY)

		if obj.Color.R != 255 || obj.Color.G != 255 || obj.Color.B != 255 {
			r := float64(obj.Color.R) / 255
			g := float64(obj.Color.G) / 255
			b := float64(obj.Color.B) / 255
			a := float64(obj.Color.A) / 255
			opts.ColorScale.Scale(float32(r), float32(g), float32(b), float32(a))
		}

		screen.DrawImage(sprite, opts)
	}
}

// RenderBillboardWithAlpha draws a single billboard with custom alpha.
// Used for smooth transitions between LOD tiers.
func (br *BillboardRenderer) RenderBillboardWithAlpha(screen *ebiten.Image, obj *Object, alpha float64, sprites map[string]*ebiten.Image) {
	if !obj.Visible || alpha <= 0 {
		return
	}

	// Get sprite for this object (or use default)
	sprite := br.DefaultSprite
	if sprites != nil {
		if s, ok := sprites[obj.ID]; ok && s != nil {
			sprite = s
		}
	}

	if sprite == nil {
		return
	}

	spriteW := float64(sprite.Bounds().Dx())
	spriteH := float64(sprite.Bounds().Dy())
	spriteSize := max(spriteW, spriteH)

	targetSize := obj.ApparentRadius * 2
	scale := targetSize / spriteSize

	if scale < br.MinScale {
		scale = br.MinScale
	}
	if scale > br.MaxScale {
		scale = br.MaxScale
	}

	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(-spriteW/2, -spriteH/2)
	opts.GeoM.Scale(scale, scale)
	opts.GeoM.Translate(obj.ScreenX, obj.ScreenY)

	// Apply color with alpha
	r := float64(obj.Color.R) / 255
	g := float64(obj.Color.G) / 255
	b := float64(obj.Color.B) / 255
	a := (float64(obj.Color.A) / 255) * alpha
	opts.ColorScale.Scale(float32(r), float32(g), float32(b), float32(a))

	screen.DrawImage(sprite, opts)
}

// CreateDefaultPlanetSprite creates a simple circular sprite for planets.
// Can be used as a placeholder billboard.
func CreateDefaultPlanetSprite(size int, col color.RGBA) *ebiten.Image {
	img := ebiten.NewImage(size, size)
	center := float64(size) / 2
	radiusSq := center * center

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := float64(x) - center + 0.5
			dy := float64(y) - center + 0.5
			distSq := dx*dx + dy*dy

			if distSq < radiusSq {
				// Inside circle - use gradient for 3D-ish look
				dist := float64(1) - (distSq / radiusSq)
				brightness := 0.5 + 0.5*dist

				img.Set(x, y, color.RGBA{
					R: uint8(float64(col.R) * brightness),
					G: uint8(float64(col.G) * brightness),
					B: uint8(float64(col.B) * brightness),
					A: col.A,
				})
			}
		}
	}

	return img
}

// ExtractAverageColor samples an equirectangular texture and returns the average color.
// This is useful for deriving circle/point colors that match the texture appearance.
func ExtractAverageColor(texture *ebiten.Image) color.RGBA {
	if texture == nil {
		return color.RGBA{200, 200, 200, 255}
	}

	bounds := texture.Bounds()
	texW := bounds.Dx()
	texH := bounds.Dy()

	// Sample a grid of pixels for efficiency (not every pixel)
	sampleStep := 8
	if texW < 64 {
		sampleStep = 1
	}

	var totalR, totalG, totalB float64
	var count float64

	for y := 0; y < texH; y += sampleStep {
		for x := 0; x < texW; x += sampleStep {
			c := texture.At(x, y)
			r, g, b, a := c.RGBA()
			if a > 0 {
				// Weight by alpha
				alpha := float64(a) / 65535.0
				totalR += float64(r>>8) * alpha
				totalG += float64(g>>8) * alpha
				totalB += float64(b>>8) * alpha
				count += alpha
			}
		}
	}

	if count == 0 {
		return color.RGBA{200, 200, 200, 255}
	}

	return color.RGBA{
		R: uint8(totalR / count),
		G: uint8(totalG / count),
		B: uint8(totalB / count),
		A: 255,
	}
}

// CreateBillboardFromTexture creates a billboard sprite from an equirectangular planet texture.
// This samples the texture with spherical projection and adds lighting for a 3D appearance.
// The result looks much closer to the actual 3D planet than a solid color sprite.
func CreateBillboardFromTexture(texture *ebiten.Image, size int) *ebiten.Image {
	if texture == nil {
		return CreateDefaultPlanetSprite(size, color.RGBA{200, 200, 200, 255})
	}

	img := ebiten.NewImage(size, size)
	texBounds := texture.Bounds()
	texW := float64(texBounds.Dx())
	texH := float64(texBounds.Dy())

	center := float64(size) / 2
	radius := center - 1 // Slight inset to avoid edge artifacts

	// Light direction (from upper-left, like the sun in the 3D scene)
	lightX, lightY, lightZ := 0.5, 0.5, 0.7
	lightLen := math.Sqrt(lightX*lightX + lightY*lightY + lightZ*lightZ)
	lightX, lightY, lightZ = lightX/lightLen, lightY/lightLen, lightZ/lightLen

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			// Normalized coordinates from center (-1 to 1)
			nx := (float64(x) - center) / radius
			ny := (float64(y) - center) / radius

			distSq := nx*nx + ny*ny
			if distSq > 1.0 {
				continue // Outside sphere
			}

			// Calculate Z on sphere surface (pointing toward viewer)
			nz := math.Sqrt(1.0 - distSq)

			// Calculate spherical coordinates for texture lookup
			// phi: longitude (0 to 2π), theta: latitude (0 to π)
			phi := math.Atan2(nx, nz) + math.Pi   // Rotate so front-center is at center of texture
			theta := math.Acos(-ny)                // -ny because Y is inverted in screen space

			// Map to texture coordinates
			u := phi / (2 * math.Pi) // 0 to 1
			v := theta / math.Pi      // 0 to 1

			// Sample texture (with wrapping)
			texX := int(u*texW) % int(texW)
			texY := int(v*texH) % int(texH)
			if texX < 0 {
				texX += int(texW)
			}

			texCol := texture.At(texX, texY)
			r, g, b, a := texCol.RGBA()

			// Calculate lighting (dot product of normal and light direction)
			// Normal is (nx, ny, nz) on the sphere surface
			dot := nx*lightX + (-ny)*lightY + nz*lightZ
			if dot < 0 {
				dot = 0
			}

			// Ambient + diffuse lighting
			ambient := 0.3
			diffuse := 0.7 * dot
			brightness := ambient + diffuse

			// Apply lighting to texture color
			img.Set(x, y, color.RGBA{
				R: uint8(math.Min(float64(r>>8)*brightness, 255)),
				G: uint8(math.Min(float64(g>>8)*brightness, 255)),
				B: uint8(math.Min(float64(b>>8)*brightness, 255)),
				A: uint8(a >> 8),
			})
		}
	}

	return img
}

// CreateDefaultStarSprite creates a simple star sprite with glow effect.
func CreateDefaultStarSprite(size int, col color.RGBA) *ebiten.Image {
	img := ebiten.NewImage(size, size)
	center := float64(size) / 2
	coreRadius := center * 0.3
	glowRadius := center

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := float64(x) - center + 0.5
			dy := float64(y) - center + 0.5
			dist := (dx*dx + dy*dy)

			if dist < coreRadius*coreRadius {
				// Core - bright white/color
				img.Set(x, y, col)
			} else if dist < glowRadius*glowRadius {
				// Glow - fading color
				normalizedDist := (dist - coreRadius*coreRadius) / (glowRadius*glowRadius - coreRadius*coreRadius)
				alpha := uint8(float64(col.A) * (1 - normalizedDist) * 0.5)
				if alpha > 0 {
					img.Set(x, y, color.RGBA{
						R: col.R,
						G: col.G,
						B: col.B,
						A: alpha,
					})
				}
			}
		}
	}

	return img
}

// CreateSpriteAtlas creates an atlas with multiple pre-rendered sprites.
// Returns the atlas image and a map of sprite names to sub-image bounds.
func CreateSpriteAtlas(size int) (*ebiten.Image, map[string]*ebiten.Image) {
	// Create atlas with 4 sprites: white planet, yellow star, red planet, blue planet
	atlasSize := size * 2
	atlas := ebiten.NewImage(atlasSize, atlasSize)

	sprites := make(map[string]*ebiten.Image)

	// White planet (top-left)
	whitePlanet := CreateDefaultPlanetSprite(size, color.RGBA{255, 255, 255, 255})
	opts := &ebiten.DrawImageOptions{}
	atlas.DrawImage(whitePlanet, opts)
	sprites["white_planet"] = whitePlanet

	// Yellow star (top-right)
	yellowStar := CreateDefaultStarSprite(size, color.RGBA{255, 255, 200, 255})
	opts = &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(float64(size), 0)
	atlas.DrawImage(yellowStar, opts)
	sprites["yellow_star"] = yellowStar

	// Red planet (bottom-left)
	redPlanet := CreateDefaultPlanetSprite(size, color.RGBA{200, 100, 80, 255})
	opts = &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(0, float64(size))
	atlas.DrawImage(redPlanet, opts)
	sprites["red_planet"] = redPlanet

	// Blue planet (bottom-right)
	bluePlanet := CreateDefaultPlanetSprite(size, color.RGBA{100, 150, 255, 255})
	opts = &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(float64(size), float64(size))
	atlas.DrawImage(bluePlanet, opts)
	sprites["blue_planet"] = bluePlanet

	return atlas, sprites
}
