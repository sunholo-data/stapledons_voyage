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
	// Check if tile is in view
	if !TileInView(c.Tile, int(c.Height), cam, screenW, screenH) {
		return
	}

	// Convert tile to screen coordinates
	sx, sy := TileToScreen(c.Tile, int(c.Height), cam, screenW, screenH)

	// Calculate tile size in screen space
	tileW := TileWidth * cam.Zoom
	tileH := TileHeight * cam.Zoom

	// Draw sprite if available, otherwise colored diamond
	if r.assets != nil && c.SpriteId > 0 {
		sprite := r.assets.GetSprite(int(c.SpriteId))
		if sprite != nil {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Scale(cam.Zoom, cam.Zoom)
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

// drawIsoEntity renders an isometric entity with sub-tile positioning.
func (r *Renderer) drawIsoEntity(screen *ebiten.Image, c *sim_gen.DrawCmdIsoEntity, cam sim_gen.Camera, screenW, screenH int) {
	// Convert tile + offset to screen coordinates
	sx, sy := TileToScreenWithOffset(c.Tile, c.OffsetX, c.OffsetY, int(c.Height), cam, screenW, screenH)

	// Draw sprite if available
	if r.assets != nil && c.SpriteId > 0 {
		sprite := r.assets.GetSprite(int(c.SpriteId))
		if sprite != nil {
			// Check if this sprite is animated
			if r.anims != nil && r.anims.HasAnimations(int(c.SpriteId)) {
				r.drawAnimatedEntity(screen, sprite, c, sx, sy, cam, screenW, screenH)
				return
			}

			// Non-animated sprite: use full image
			spriteW := float64(sprite.Bounds().Dx()) * cam.Zoom
			spriteH := float64(sprite.Bounds().Dy()) * cam.Zoom

			// Check if on screen (simple bounds check)
			if sx+spriteW/2 < 0 || sx-spriteW/2 > float64(screenW) ||
				sy < 0 || sy-spriteH > float64(screenH) {
				return
			}

			op := &ebiten.DrawImageOptions{}
			op.GeoM.Scale(cam.Zoom, cam.Zoom)
			// Anchor at bottom-center (feet at tile center)
			op.GeoM.Translate(sx-spriteW/2, sy-spriteH)
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
	frameX := frameIdx * frameW
	subImg := sprite.SubImage(image.Rect(frameX, 0, frameX+frameW, frameH)).(*ebiten.Image)

	// Calculate scaled dimensions
	scaledW := float64(frameW) * cam.Zoom
	scaledH := float64(frameH) * cam.Zoom

	// Check if on screen
	if sx+scaledW/2 < 0 || sx-scaledW/2 > float64(screenW) ||
		sy < 0 || sy-scaledH > float64(screenH) {
		return
	}

	// Draw the frame
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(cam.Zoom, cam.Zoom)
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
		_, sy := TileToScreen(c.Tile, int(c.Height), cam, screenW, screenH)
		return IsoDepth(int(c.Layer), sy)

	case sim_gen.DrawCmdKindIsoEntity:
		c := cmd.IsoEntity
		_, sy := TileToScreenWithOffset(c.Tile, c.OffsetX, c.OffsetY, int(c.Height), cam, screenW, screenH)
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
