package render

import (
	"image"
	"image/color"
	"sort"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"stapledons_voyage/engine/assets"
	"stapledons_voyage/engine/camera"
	"stapledons_voyage/sim_gen"
)

// Biome and structure colors for rendering (fallback when not using sprites)
var biomeColors = []color.RGBA{
	{0, 100, 200, 255},   // 0: Water (blue)
	{34, 139, 34, 255},   // 1: Forest (green)
	{210, 180, 140, 255}, // 2: Desert (tan)
	{139, 90, 43, 255},   // 3: Mountain (brown)
	{255, 255, 0, 128},   // 4: Selection highlight (yellow, semi-transparent)
	{139, 69, 19, 255},   // 5: House (saddle brown)
	{50, 205, 50, 255},   // 6: Farm (lime green)
	{128, 128, 128, 255}, // 7: Road (gray)
	{0, 0, 0, 255},       // 8: Reserved
	{0, 0, 0, 255},       // 9: Reserved
	{255, 0, 0, 255},     // 10: NPC 0 (red)
	{0, 255, 0, 255},     // 11: NPC 1 (green)
	{0, 0, 255, 255},     // 12: NPC 2 (blue)
	{255, 255, 0, 255},   // 13: NPC 3 (yellow)
	{255, 0, 255, 255},   // 14: NPC 4 (magenta)
}

// Renderer handles drawing FrameOutput to the screen.
type Renderer struct {
	assets    *assets.Manager
	anims     *AnimationManager
	lastTick  uint64 // Track simulation tick for animation updates
}

// NewRenderer creates a renderer with the given asset manager.
func NewRenderer(assets *assets.Manager) *Renderer {
	r := &Renderer{
		assets: assets,
		anims:  NewAnimationManager(),
	}
	// Register animation definitions from asset manager
	if assets != nil {
		r.registerAnimations()
	}
	return r
}

// registerAnimations copies animation definitions from assets to the animation manager.
func (r *Renderer) registerAnimations() {
	// Get animation definitions from sprite manager
	// We need to check each sprite ID that might have animations
	for spriteID := 100; spriteID <= 105; spriteID++ {
		animDef := r.assets.Sprites().GetAnimation(spriteID)
		if animDef != nil {
			r.anims.RegisterSprite(spriteID, &AnimationDef{
				Animations:  convertAnimations(animDef.Animations),
				FrameWidth:  animDef.FrameWidth,
				FrameHeight: animDef.FrameHeight,
			})
		}
	}
}

// convertAnimations converts asset animation sequences to render animation sequences.
func convertAnimations(src map[string]assets.SpriteAnimSeq) map[string]AnimationSeq {
	dst := make(map[string]AnimationSeq)
	for name, seq := range src {
		dst[name] = AnimationSeq{
			StartFrame: seq.StartFrame,
			FrameCount: seq.FrameCount,
			FPS:        seq.FPS,
		}
	}
	return dst
}

// RenderFrame renders the FrameOutput to the Ebiten screen.
func (r *Renderer) RenderFrame(screen *ebiten.Image, out sim_gen.FrameOutput) {
	// Update animations (assume 60 FPS, ~16.67ms per frame)
	const dt = 1.0 / 60.0
	if r.anims != nil {
		r.anims.Update(dt)
	}

	// Get screen dimensions
	screenW, screenH := screen.Bounds().Dx(), screen.Bounds().Dy()

	// Create camera transform
	transform := camera.FromOutput(out.Camera, screenW, screenH)

	// Calculate viewport for culling
	viewport := camera.CalculateViewport(out.Camera, screenW, screenH)

	// Sort draw commands using isometric depth sorting
	// This handles both legacy Z-sorting and iso (layer, screenY) sorting
	sortables := make([]isoSortable, len(out.Draw))
	for i, cmd := range out.Draw {
		sortables[i] = isoSortable{
			cmd:     cmd,
			sortKey: getIsoSortKey(cmd, out.Camera, screenW, screenH),
		}
	}
	sort.Slice(sortables, func(i, j int) bool {
		return sortables[i].sortKey < sortables[j].sortKey
	})

	// Render each command
	for _, s := range sortables {
		switch c := s.cmd.(type) {
		case sim_gen.DrawCmdRect:
			// Cull if outside viewport
			if !viewport.ContainsRect(c.X, c.Y, c.W, c.H) {
				continue
			}
			// Transform to screen coordinates
			sx, sy := transform.WorldToScreen(c.X, c.Y)
			sw := c.W * transform.Scale
			sh := c.H * transform.Scale
			col := biomeColors[c.Color%len(biomeColors)]
			ebitenutil.DrawRect(screen, sx, sy, sw, sh, col)

		case sim_gen.DrawCmdSprite:
			// Cull if outside viewport (assuming 16x16 sprite)
			if !viewport.ContainsRect(c.X, c.Y, 16, 16) {
				continue
			}
			r.drawSprite(screen, c, transform)

		case sim_gen.DrawCmdText:
			// Transform text position
			sx, sy := transform.WorldToScreen(c.X, c.Y)
			ebitenutil.DebugPrintAt(screen, c.Text, int(sx), int(sy))

		case sim_gen.DrawCmdIsoTile:
			r.drawIsoTile(screen, c, out.Camera, screenW, screenH)

		case sim_gen.DrawCmdIsoEntity:
			r.drawIsoEntity(screen, c, out.Camera, screenW, screenH)

		case sim_gen.DrawCmdUi:
			r.drawUiElement(screen, c, screenW, screenH)
		}
	}

	// Render debug messages at top-left of screen (UI layer, not transformed)
	for i, msg := range out.Debug {
		ebitenutil.DebugPrintAt(screen, msg, 10, 10+i*16)
	}
}

// drawSprite draws a sprite using the asset manager with camera transform.
func (r *Renderer) drawSprite(screen *ebiten.Image, c sim_gen.DrawCmdSprite, transform camera.Transform) {
	sx, sy := transform.WorldToScreen(c.X, c.Y)

	if r.assets == nil {
		// Fallback: draw white placeholder
		sw := 16.0 * transform.Scale
		sh := 16.0 * transform.Scale
		ebitenutil.DrawRect(screen, sx, sy, sw, sh, color.White)
		return
	}

	sprite := r.assets.GetSprite(c.ID)
	if sprite == nil {
		sw := 16.0 * transform.Scale
		sh := 16.0 * transform.Scale
		ebitenutil.DrawRect(screen, sx, sy, sw, sh, color.White)
		return
	}

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(transform.Scale, transform.Scale)
	op.GeoM.Translate(sx, sy)
	screen.DrawImage(sprite, op)
}

// RenderFrame is a convenience function for backwards compatibility.
// Prefer using Renderer.RenderFrame for access to sprites.
func RenderFrame(screen *ebiten.Image, out sim_gen.FrameOutput) {
	r := &Renderer{assets: nil}
	r.RenderFrame(screen, out)
}

func getZ(cmd sim_gen.DrawCmd) int {
	switch c := cmd.(type) {
	case sim_gen.DrawCmdRect:
		return c.Z
	case sim_gen.DrawCmdSprite:
		return c.Z
	case sim_gen.DrawCmdText:
		return c.Z
	case sim_gen.DrawCmdIsoTile:
		return c.Layer
	case sim_gen.DrawCmdIsoEntity:
		return c.Layer
	case sim_gen.DrawCmdUi:
		return c.Z + 10000 // UI always on top
	}
	return 0
}

// =============================================================================
// Isometric Rendering Functions
// =============================================================================

// drawIsoTile renders an isometric tile.
func (r *Renderer) drawIsoTile(screen *ebiten.Image, c sim_gen.DrawCmdIsoTile, cam sim_gen.Camera, screenW, screenH int) {
	// Check if tile is in view
	if !TileInView(c.Tile, c.Height, cam, screenW, screenH) {
		return
	}

	// Convert tile to screen coordinates
	sx, sy := TileToScreen(c.Tile, c.Height, cam, screenW, screenH)

	// Calculate tile size in screen space
	tileW := TileWidth * cam.Zoom
	tileH := TileHeight * cam.Zoom

	// Draw sprite if available, otherwise colored diamond
	if r.assets != nil && c.SpriteID > 0 {
		sprite := r.assets.GetSprite(c.SpriteID)
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
	col := biomeColors[c.Color%len(biomeColors)]
	drawIsoDiamond(screen, sx, sy, tileW, tileH, col)
}

// drawIsoEntity renders an isometric entity with sub-tile positioning.
func (r *Renderer) drawIsoEntity(screen *ebiten.Image, c sim_gen.DrawCmdIsoEntity, cam sim_gen.Camera, screenW, screenH int) {
	// Convert tile + offset to screen coordinates
	sx, sy := TileToScreenWithOffset(c.Tile, c.OffsetX, c.OffsetY, c.Height, cam, screenW, screenH)

	// Draw sprite if available
	if r.assets != nil && c.SpriteID > 0 {
		sprite := r.assets.GetSprite(c.SpriteID)
		if sprite != nil {
			// Check if this sprite is animated
			if r.anims != nil && r.anims.HasAnimations(c.SpriteID) {
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
func (r *Renderer) drawAnimatedEntity(screen *ebiten.Image, sprite *ebiten.Image, c sim_gen.DrawCmdIsoEntity, sx, sy float64, cam sim_gen.Camera, screenW, screenH int) {
	// Get animation name from entity (default to "walk" for moving entities, "idle" otherwise)
	animName := "idle"
	if c.OffsetX != 0 || c.OffsetY != 0 {
		animName = "walk"
	}

	// Get current frame index
	entityID := c.ID
	frameIdx := r.anims.GetFrame(entityID, c.SpriteID, animName)

	// Get frame dimensions
	frameW, frameH := r.anims.GetFrameDimensions(c.SpriteID)
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

// drawUiElement renders a UI element in screen space (not affected by camera).
func (r *Renderer) drawUiElement(screen *ebiten.Image, c sim_gen.DrawCmdUi, screenW, screenH int) {
	// Convert normalized coordinates to screen pixels
	px := c.X * float64(screenW)
	py := c.Y * float64(screenH)
	pw := c.W * float64(screenW)
	ph := c.H * float64(screenH)

	// Get color
	col := biomeColors[c.Color%len(biomeColors)]

	switch c.Kind {
	case sim_gen.UiKindPanel:
		// Draw panel background
		ebitenutil.DrawRect(screen, px, py, pw, ph, col)

	case sim_gen.UiKindButton:
		// Draw button with border
		ebitenutil.DrawRect(screen, px, py, pw, ph, col)
		// Simple border effect
		borderCol := color.RGBA{col.R / 2, col.G / 2, col.B / 2, 255}
		ebitenutil.DrawRect(screen, px, py, pw, 2, borderCol)       // top
		ebitenutil.DrawRect(screen, px, py+ph-2, pw, 2, borderCol)  // bottom
		ebitenutil.DrawRect(screen, px, py, 2, ph, borderCol)       // left
		ebitenutil.DrawRect(screen, px+pw-2, py, 2, ph, borderCol)  // right

	case sim_gen.UiKindLabel:
		// Just draw text (background optional)
		if c.Color > 0 {
			ebitenutil.DrawRect(screen, px, py, pw, ph, col)
		}

	case sim_gen.UiKindPortrait:
		// Draw sprite or placeholder
		if r.assets != nil && c.SpriteID > 0 {
			sprite := r.assets.GetSprite(c.SpriteID)
			if sprite != nil {
				op := &ebiten.DrawImageOptions{}
				// Scale sprite to fit rect
				sw, sh := sprite.Bounds().Dx(), sprite.Bounds().Dy()
				op.GeoM.Scale(pw/float64(sw), ph/float64(sh))
				op.GeoM.Translate(px, py)
				screen.DrawImage(sprite, op)
				return
			}
		}
		// Fallback: draw placeholder
		ebitenutil.DrawRect(screen, px, py, pw, ph, color.RGBA{100, 100, 100, 255})
	}

	// Draw text if present
	if c.Text != "" {
		// Use loaded font if available, otherwise fallback to debug print
		if r.assets != nil {
			face := r.assets.GetDefaultFont()
			if face != nil {
				// text.Draw uses baseline Y, so offset down from top
				text.Draw(screen, c.Text, face, int(px)+4, int(py)+16, color.White)
				return
			}
		}
		// Fallback to debug font
		ebitenutil.DebugPrintAt(screen, c.Text, int(px)+4, int(py)+4)
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

// =============================================================================
// Isometric Sorting Helper
// =============================================================================

// isoSortKey returns a sort key for isometric draw commands.
// Uses (layer, screenY) for proper depth sorting.
type isoSortable struct {
	cmd     sim_gen.DrawCmd
	sortKey float64
}

func getIsoSortKey(cmd sim_gen.DrawCmd, cam sim_gen.Camera, screenW, screenH int) float64 {
	switch c := cmd.(type) {
	case sim_gen.DrawCmdIsoTile:
		_, sy := TileToScreen(c.Tile, c.Height, cam, screenW, screenH)
		return IsoDepth(c.Layer, sy)

	case sim_gen.DrawCmdIsoEntity:
		_, sy := TileToScreenWithOffset(c.Tile, c.OffsetX, c.OffsetY, c.Height, cam, screenW, screenH)
		return IsoDepth(c.Layer, sy)

	case sim_gen.DrawCmdUi:
		// UI is always on top, sorted by Z within UI layer
		return IsoDepth(1000+c.Z, 0)

	default:
		// Legacy commands use simple Z
		return float64(getZ(cmd))
	}
}
