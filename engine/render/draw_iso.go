package render

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"stapledons_voyage/sim_gen"
)

// =============================================================================
// Isometric Rendering Functions
// =============================================================================

// drawIsoTile renders an isometric tile.
func (r *Renderer) drawIsoTile(screen *ebiten.Image, c *sim_gen.DrawCmdIsoTile, cam sim_gen.Camera, screenW, screenH int) {
	// Dereference tile coord pointer
	tile := *c.Tile

	// Check if tile is in view
	if !TileInView(tile, int(c.Height), cam, screenW, screenH) {
		return
	}

	// Convert tile to screen coordinates
	sx, sy := TileToScreen(tile, int(c.Height), cam, screenW, screenH)

	// Calculate tile size in screen space
	tileW := TileWidth * cam.Zoom
	tileH := TileHeight * cam.Zoom

	// Draw sprite if available, otherwise colored diamond
	if r.assets != nil && c.SpriteId > 0 {
		sprite := r.assets.GetSprite(int(c.SpriteId))
		if sprite != nil {
			spriteW := float64(sprite.Bounds().Dx())
			spriteH := float64(sprite.Bounds().Dy())

			// Scale sprite to fill the isometric tile footprint exactly
			// Isometric tiles are 2:1 aspect ratio (64x32)
			// Stretch sprite to match tile dimensions for proper tessellation
			scaleX := tileW / spriteW
			scaleY := tileH / spriteH

			op := &ebiten.DrawImageOptions{}
			op.GeoM.Scale(scaleX, scaleY)
			// Center sprite on tile position
			op.GeoM.Translate(sx-tileW/2, sy-tileH/2)
			screen.DrawImage(sprite, op)
			return
		}
	}

	// Fallback: draw colored diamond (isometric tile shape)
	col := biomeColors[int(c.Color)%len(biomeColors)]
	drawIsoDiamond(screen, sx, sy, tileW, tileH, col)
}

// drawIsoTileAlpha renders a transparent isometric tile with alpha blending.
func (r *Renderer) drawIsoTileAlpha(screen *ebiten.Image, c *sim_gen.DrawCmdIsoTileAlpha, cam sim_gen.Camera, screenW, screenH int) {
	// Dereference tile coord pointer
	tile := *c.Tile

	// Check if tile is in view
	if !TileInView(tile, int(c.Height), cam, screenW, screenH) {
		return
	}

	// Convert tile to screen coordinates
	sx, sy := TileToScreen(tile, int(c.Height), cam, screenW, screenH)

	// Calculate tile size in screen space
	tileW := TileWidth * cam.Zoom
	tileH := TileHeight * cam.Zoom

	// Draw sprite if available
	if r.assets != nil && c.SpriteId > 0 {
		sprite := r.assets.GetSprite(int(c.SpriteId))
		if sprite != nil {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Scale(cam.Zoom, cam.Zoom)
			// Center sprite on tile position
			op.GeoM.Translate(sx-tileW/2, sy-tileH/2)

			// Apply alpha
			op.ColorScale.ScaleAlpha(float32(c.Alpha))

			// Apply tint if specified
			if c.TintRgba != 0 {
				tintCol := unpackRGBA(c.TintRgba)
				// Multiply with tint color
				op.ColorScale.ScaleWithColor(tintCol)
			}

			screen.DrawImage(sprite, op)
			return
		}
	}

	// Fallback: draw semi-transparent colored diamond
	baseCol := color.RGBA{100, 150, 200, 255} // Blue-ish default

	// Apply tint if specified
	if c.TintRgba != 0 {
		baseCol = unpackRGBA(c.TintRgba)
	}

	// Apply alpha to the color
	finalCol := color.RGBA{
		R: baseCol.R,
		G: baseCol.G,
		B: baseCol.B,
		A: uint8(float64(baseCol.A) * c.Alpha),
	}
	drawIsoDiamond(screen, sx, sy, tileW, tileH, finalCol)
}

// drawIsoEntity renders an isometric entity with sub-tile positioning.
func (r *Renderer) drawIsoEntity(screen *ebiten.Image, c *sim_gen.DrawCmdIsoEntity, cam sim_gen.Camera, screenW, screenH int) {
	// Dereference tile coord pointer
	tile := *c.Tile

	// Convert tile + offset to screen coordinates
	sx, sy := TileToScreenWithOffset(tile, c.OffsetX, c.OffsetY, int(c.Height), cam, screenW, screenH)

	// Draw sprite if available
	if r.assets != nil && c.SpriteId > 0 {
		sprite := r.assets.GetSprite(int(c.SpriteId))
		if sprite != nil {
			// Check if this sprite is animated
			if r.anims != nil && r.anims.HasAnimations(int(c.SpriteId)) {
				r.drawAnimatedEntity(screen, sprite, c, sx, sy, cam, screenW, screenH)
				return
			}

			// Non-animated sprite: scale to reasonable size for isometric grid
			// Entities should be ~2-3 tiles tall/wide
			spriteW := float64(sprite.Bounds().Dx())
			spriteH := float64(sprite.Bounds().Dy())

			// Target size: entities span ~2 tiles (128 pixels at zoom 1.0)
			targetSize := TileWidth * 2.0 * cam.Zoom
			scale := targetSize / spriteW
			if spriteH*scale > targetSize*1.5 {
				// If sprite is very tall, constrain by height
				scale = (targetSize * 1.5) / spriteH
			}

			scaledW := spriteW * scale
			scaledH := spriteH * scale

			// Check if on screen (simple bounds check)
			if sx+scaledW/2 < 0 || sx-scaledW/2 > float64(screenW) ||
				sy < 0 || sy-scaledH > float64(screenH) {
				return
			}

			op := &ebiten.DrawImageOptions{}
			op.GeoM.Scale(scale, scale)
			// Anchor at bottom-center (feet at tile center)
			op.GeoM.Translate(sx-scaledW/2, sy-scaledH)
			screen.DrawImage(sprite, op)
			return
		}
	}

	// Fallback: draw colored rectangle (16x16 default)
	fallbackSize := 16.0 * cam.Zoom
	if sx+fallbackSize < 0 || sx-fallbackSize > float64(screenW) ||
		sy+fallbackSize < 0 || sy-fallbackSize > float64(screenH) {
		return
	}
	ebitenutil.DrawRect(screen, sx-fallbackSize/2, sy-fallbackSize, fallbackSize, fallbackSize, color.RGBA{255, 100, 100, 255})
}

// drawAnimatedEntity renders an animated entity using the current animation frame.
func (r *Renderer) drawAnimatedEntity(screen *ebiten.Image, sprite *ebiten.Image, c *sim_gen.DrawCmdIsoEntity, sx, sy float64, cam sim_gen.Camera, screenW, screenH int) {
	// Get animation name from entity (default to "walk" for moving entities, "idle" otherwise)
	animName := "idle"
	if c.OffsetX != 0 || c.OffsetY != 0 {
		animName = "walk"
	}

	// Get current frame index
	entityID := c.Id
	frameIdx := r.anims.GetFrame(entityID, int(c.SpriteId), animName)

	// Get frame dimensions
	frameW, frameH := r.anims.GetFrameDimensions(int(c.SpriteId))
	if frameW == 0 || frameH == 0 {
		frameW, frameH = 32, 48 // Default frame size
	}

	// Calculate sub-rectangle for this frame
	// Handle 2D sprite sheets (rows and columns)
	spriteW := sprite.Bounds().Dx()
	framesPerRow := 1
	if frameW > 0 && spriteW >= frameW {
		framesPerRow = spriteW / frameW
	}
	col := frameIdx % framesPerRow
	row := frameIdx / framesPerRow
	frameX := col * frameW
	frameY := row * frameH
	subImg := sprite.SubImage(image.Rect(frameX, frameY, frameX+frameW, frameY+frameH)).(*ebiten.Image)

	// Scale to fit reasonably in the isometric grid
	// Target: entity should be ~1.5 tiles tall (96 pixels at zoom 1.0)
	targetH := TileHeight * 3.0 * cam.Zoom
	scale := targetH / float64(frameH)
	scaledW := float64(frameW) * scale
	scaledH := float64(frameH) * scale

	// Check if on screen
	if sx+scaledW/2 < 0 || sx-scaledW/2 > float64(screenW) ||
		sy < 0 || sy-scaledH > float64(screenH) {
		return
	}

	// Draw the frame
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(scale, scale)
	// Anchor at bottom-center (feet at tile center)
	op.GeoM.Translate(sx-scaledW/2, sy-scaledH)
	screen.DrawImage(subImg, op)
}

// =============================================================================
// Isometric Sorting Helper
// =============================================================================

// isoSortKey returns a sort key for isometric draw commands.
// Uses (layer, screenY) for proper depth sorting.
type isoSortable struct {
	cmd     *sim_gen.DrawCmd
	sortKey float64
}

func getIsoSortKey(cmd *sim_gen.DrawCmd, cam sim_gen.Camera, screenW, screenH int) float64 {
	switch cmd.Kind {
	case sim_gen.DrawCmdKindIsoTile:
		c := cmd.IsoTile
		_, sy := TileToScreen(*c.Tile, int(c.Height), cam, screenW, screenH)
		return IsoDepth(int(c.Layer), sy)

	case sim_gen.DrawCmdKindIsoTileAlpha:
		c := cmd.IsoTileAlpha
		_, sy := TileToScreen(*c.Tile, int(c.Height), cam, screenW, screenH)
		return IsoDepth(int(c.Layer), sy)

	case sim_gen.DrawCmdKindIsoEntity:
		c := cmd.IsoEntity
		_, sy := TileToScreenWithOffset(*c.Tile, c.OffsetX, c.OffsetY, int(c.Height), cam, screenW, screenH)
		return IsoDepth(int(c.Layer), sy)

	case sim_gen.DrawCmdKindUi:
		c := cmd.Ui
		// UI is always on top, sorted by Z within UI layer
		return IsoDepth(1000+int(c.Z), 0)

	case sim_gen.DrawCmdKindLine:
		// Lines use simple Z for now
		return float64(cmd.Line.Z)

	case sim_gen.DrawCmdKindTextWrapped:
		return float64(cmd.TextWrapped.Z)

	case sim_gen.DrawCmdKindCircle:
		return float64(cmd.Circle.Z)

	default:
		// Legacy commands use simple Z
		return float64(getZ(cmd))
	}
}

// drawIsoDiamond draws a filled diamond shape (isometric tile).
// The diamond has vertices at top, right, bottom, left forming the isometric tile shape.
func drawIsoDiamond(screen *ebiten.Image, cx, cy, w, h float64, col color.RGBA) {
	halfW := float32(w / 2)
	halfH := float32(h / 2)
	fcx := float32(cx)
	fcy := float32(cy)

	// Draw filled diamond using vector path
	var path vector.Path
	path.MoveTo(fcx, fcy-halfH)        // Top vertex
	path.LineTo(fcx+halfW, fcy)        // Right vertex
	path.LineTo(fcx, fcy+halfH)        // Bottom vertex
	path.LineTo(fcx-halfW, fcy)        // Left vertex
	path.Close()                        // Back to top

	// Fill the diamond
	vertices, indices := path.AppendVerticesAndIndicesForFilling(nil, nil)

	// Set color for all vertices
	for i := range vertices {
		vertices[i].SrcX = 1
		vertices[i].SrcY = 1
		vertices[i].ColorR = float32(col.R) / 255
		vertices[i].ColorG = float32(col.G) / 255
		vertices[i].ColorB = float32(col.B) / 255
		vertices[i].ColorA = float32(col.A) / 255
	}

	// Draw using a white pixel as source (color comes from vertices)
	screen.DrawTriangles(vertices, indices, whitePixel, &ebiten.DrawTrianglesOptions{})
}

// whitePixel is a 1x1 white image used as source for colored triangles
var whitePixel = func() *ebiten.Image {
	img := ebiten.NewImage(3, 3)
	img.Fill(color.White)
	return img
}()
