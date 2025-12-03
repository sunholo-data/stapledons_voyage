package render

import (
	"image"
	"image/color"
	_ "image/jpeg" // Register JPEG decoder
	"os"
	"sort"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font"
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
	{255, 255, 255, 180}, // 4: Selection highlight (white, semi-transparent)
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
	lastTick  uint64         // Track simulation tick for animation updates
	galaxyBg  *ebiten.Image  // Galaxy background image (loaded lazily)
	galaxyBgLoaded bool      // Whether we've attempted to load the background
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

	// Render each command using Kind-based dispatch (discriminator struct pattern)
	for _, s := range sortables {
		cmd := s.cmd
		switch cmd.Kind {
		case sim_gen.DrawCmdKindRect:
			c := cmd.Rect
			// Cull if outside viewport
			if !viewport.ContainsRect(c.X, c.Y, c.W, c.H) {
				continue
			}
			// Transform to screen coordinates
			sx, sy := transform.WorldToScreen(c.X, c.Y)
			sw := c.W * transform.Scale
			sh := c.H * transform.Scale
			col := biomeColors[int(c.Color)%len(biomeColors)]
			ebitenutil.DrawRect(screen, sx, sy, sw, sh, col)

		case sim_gen.DrawCmdKindSprite:
			c := cmd.Sprite
			// Cull if outside viewport (assuming 16x16 sprite)
			if !viewport.ContainsRect(c.X, c.Y, 16, 16) {
				continue
			}
			r.drawSprite(screen, c, transform)

		case sim_gen.DrawCmdKindText:
			c := cmd.Text
			// Screen-space coordinates (no transform)
			r.drawText(screen, c, int(c.X), int(c.Y))

		case sim_gen.DrawCmdKindIsoTile:
			r.drawIsoTile(screen, cmd.IsoTile, out.Camera, screenW, screenH)

		case sim_gen.DrawCmdKindIsoEntity:
			r.drawIsoEntity(screen, cmd.IsoEntity, out.Camera, screenW, screenH)

		case sim_gen.DrawCmdKindUi:
			r.drawUiElement(screen, cmd.Ui, screenW, screenH)

		case sim_gen.DrawCmdKindLine:
			r.drawLine(screen, cmd.Line)

		case sim_gen.DrawCmdKindTextWrapped:
			r.drawTextWrapped(screen, cmd.TextWrapped, screenW, screenH)

		case sim_gen.DrawCmdKindCircle:
			r.drawCircle(screen, cmd.Circle)

		case sim_gen.DrawCmdKindRectScreen:
			c := cmd.RectScreen
			// Screen-space rectangle (no camera transform)
			col := biomeColors[int(c.Color)%len(biomeColors)]
			ebitenutil.DrawRect(screen, c.X, c.Y, c.W, c.H, col)

		case sim_gen.DrawCmdKindGalaxyBg:
			c := cmd.GalaxyBg
			r.drawGalaxyBackground(screen, c.Opacity, screenW, screenH, c.SkyViewMode, c.ViewLon, c.ViewLat, c.Fov)

		case sim_gen.DrawCmdKindStar:
			r.drawStar(screen, cmd.Star)
		}
	}

	// Render debug messages below UI panels (UI layer, not transformed)
	// Start at y=50 to avoid overlapping with camera panel
	for i, msg := range out.Debug {
		ebitenutil.DebugPrintAt(screen, msg, 10, 50+i*16)
	}
}

// drawSprite draws a sprite using the asset manager with camera transform.
func (r *Renderer) drawSprite(screen *ebiten.Image, c *sim_gen.DrawCmdSprite, transform camera.Transform) {
	sx, sy := transform.WorldToScreen(c.X, c.Y)

	if r.assets == nil {
		// Fallback: draw white placeholder
		sw := 16.0 * transform.Scale
		sh := 16.0 * transform.Scale
		ebitenutil.DrawRect(screen, sx, sy, sw, sh, color.White)
		return
	}

	sprite := r.assets.GetSprite(int(c.Id))
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

func getZ(cmd *sim_gen.DrawCmd) int {
	switch cmd.Kind {
	case sim_gen.DrawCmdKindRect:
		return int(cmd.Rect.Z)
	case sim_gen.DrawCmdKindSprite:
		return int(cmd.Sprite.Z)
	case sim_gen.DrawCmdKindText:
		return int(cmd.Text.Z)
	case sim_gen.DrawCmdKindLine:
		return int(cmd.Line.Z)
	case sim_gen.DrawCmdKindTextWrapped:
		return int(cmd.TextWrapped.Z)
	case sim_gen.DrawCmdKindCircle:
		return int(cmd.Circle.Z)
	case sim_gen.DrawCmdKindRectScreen:
		return int(cmd.RectScreen.Z)
	case sim_gen.DrawCmdKindGalaxyBg:
		return int(cmd.GalaxyBg.Z)
	case sim_gen.DrawCmdKindStar:
		return int(cmd.Star.Z)
	case sim_gen.DrawCmdKindIsoTile:
		return int(cmd.IsoTile.Layer)
	case sim_gen.DrawCmdKindIsoEntity:
		return int(cmd.IsoEntity.Layer)
	case sim_gen.DrawCmdKindUi:
		return int(cmd.Ui.Z) + 10000 // UI always on top
	}
	return 0
}

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

// drawUiElement renders a UI element in screen space (not affected by camera).
func (r *Renderer) drawUiElement(screen *ebiten.Image, c *sim_gen.DrawCmdUi, screenW, screenH int) {
	// Convert normalized coordinates to screen pixels
	px := c.X * float64(screenW)
	py := c.Y * float64(screenH)
	pw := c.W * float64(screenW)
	ph := c.H * float64(screenH)

	// Get color
	col := biomeColors[int(c.Color)%len(biomeColors)]

	// UiKind is a discriminator struct - switch on Kind.Kind
	switch c.Kind.Kind {
	case sim_gen.UiKindKindUiPanel:
		// Draw panel background
		ebitenutil.DrawRect(screen, px, py, pw, ph, col)

	case sim_gen.UiKindKindUiButton:
		// Draw button with border
		ebitenutil.DrawRect(screen, px, py, pw, ph, col)
		// Simple border effect
		borderCol := color.RGBA{col.R / 2, col.G / 2, col.B / 2, 255}
		ebitenutil.DrawRect(screen, px, py, pw, 2, borderCol)       // top
		ebitenutil.DrawRect(screen, px, py+ph-2, pw, 2, borderCol)  // bottom
		ebitenutil.DrawRect(screen, px, py, 2, ph, borderCol)       // left
		ebitenutil.DrawRect(screen, px+pw-2, py, 2, ph, borderCol)  // right

	case sim_gen.UiKindKindUiLabel:
		// Just draw text (background optional)
		if c.Color > 0 {
			ebitenutil.DrawRect(screen, px, py, pw, ph, col)
		}

	case sim_gen.UiKindKindUiPortrait:
		// Draw sprite or placeholder
		if r.assets != nil && c.SpriteId > 0 {
			sprite := r.assets.GetSprite(int(c.SpriteId))
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

	case sim_gen.UiKindKindUiSlider:
		// Draw slider track (dark background)
		trackCol := color.RGBA{60, 60, 60, 255}
		ebitenutil.DrawRect(screen, px, py+ph/3, pw, ph/3, trackCol)

		// Draw slider fill up to value
		fillWidth := pw * c.Value
		ebitenutil.DrawRect(screen, px, py+ph/3, fillWidth, ph/3, col)

		// Draw slider handle
		handleX := px + fillWidth - 4
		if handleX < px {
			handleX = px
		}
		handleCol := color.RGBA{255, 255, 255, 255}
		ebitenutil.DrawRect(screen, handleX, py, 8, ph, handleCol)

	case sim_gen.UiKindKindUiProgressBar:
		// Draw progress bar background
		bgCol := color.RGBA{40, 40, 40, 255}
		ebitenutil.DrawRect(screen, px, py, pw, ph, bgCol)

		// Draw progress fill
		fillWidth := pw * c.Value
		ebitenutil.DrawRect(screen, px, py, fillWidth, ph, col)

		// Draw border
		borderCol := color.RGBA{100, 100, 100, 255}
		ebitenutil.DrawRect(screen, px, py, pw, 2, borderCol)        // top
		ebitenutil.DrawRect(screen, px, py+ph-2, pw, 2, borderCol)   // bottom
		ebitenutil.DrawRect(screen, px, py, 2, ph, borderCol)        // left
		ebitenutil.DrawRect(screen, px+pw-2, py, 2, ph, borderCol)   // right
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

// drawLine draws a line between two points with specified width and color.
// Coordinates are screen-space pixels (not world coordinates).
func (r *Renderer) drawLine(screen *ebiten.Image, c *sim_gen.DrawCmdLine) {
	// Get color
	col := biomeColors[int(c.Color)%len(biomeColors)]

	width := float32(c.Width)
	if width < 1 {
		width = 1
	}

	// Draw line using vector.StrokeLine (screen-space coordinates)
	vector.StrokeLine(screen, float32(c.X1), float32(c.Y1), float32(c.X2), float32(c.Y2), width, col, true)
}

// drawTextWrapped draws word-wrapped text with a specified font size and color.
// Coordinates are screen-space pixels.
func (r *Renderer) drawTextWrapped(screen *ebiten.Image, c *sim_gen.DrawCmdTextWrapped, screenW, screenH int) {
	// Get color
	col := biomeColors[int(c.Color)%len(biomeColors)]

	// Get font face for the specified size
	var face font.Face
	if r.assets != nil {
		face = r.assets.GetFontBySize(int(c.FontSize))
	}

	// Wrap text and draw (screen-space coordinates)
	if face != nil {
		lines := wrapText(c.Text, face, c.MaxWidth)
		lineHeight := face.Metrics().Height.Ceil()
		for i, line := range lines {
			text.Draw(screen, line, face, int(c.X), int(c.Y)+lineHeight*(i+1), col)
		}
	} else {
		// Fallback to debug text (no wrapping)
		ebitenutil.DebugPrintAt(screen, c.Text, int(c.X), int(c.Y))
	}
}

// wrapText splits text into lines that fit within maxWidth.
func wrapText(s string, face font.Face, maxWidth float64) []string {
	if maxWidth <= 0 || s == "" {
		return []string{s}
	}

	var lines []string
	var currentLine string
	words := splitWords(s)

	for _, word := range words {
		testLine := currentLine
		if testLine != "" {
			testLine += " "
		}
		testLine += word

		// Measure the test line
		bounds, _ := font.BoundString(face, testLine)
		lineWidth := float64((bounds.Max.X - bounds.Min.X).Ceil())

		if lineWidth > maxWidth && currentLine != "" {
			// Start new line
			lines = append(lines, currentLine)
			currentLine = word
		} else {
			currentLine = testLine
		}
	}

	// Add remaining text
	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return lines
}

// splitWords splits a string into words.
func splitWords(s string) []string {
	var words []string
	var current string
	for _, r := range s {
		if r == ' ' || r == '\n' || r == '\t' {
			if current != "" {
				words = append(words, current)
				current = ""
			}
		} else {
			current += string(r)
		}
	}
	if current != "" {
		words = append(words, current)
	}
	return words
}

// drawCircle draws a filled or outline circle.
// Coordinates are screen-space pixels.
func (r *Renderer) drawCircle(screen *ebiten.Image, c *sim_gen.DrawCmdCircle) {
	radius := float32(c.Radius)
	if radius < 1 {
		radius = 1
	}

	// Get color
	col := biomeColors[int(c.Color)%len(biomeColors)]

	if c.Filled {
		// Draw filled circle (screen-space coordinates)
		vector.DrawFilledCircle(screen, float32(c.X), float32(c.Y), radius, col, true)
	} else {
		// Draw circle outline using StrokeCircle
		vector.StrokeCircle(screen, float32(c.X), float32(c.Y), radius, 1, col, true)
	}
}

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

// drawText draws text with specified font size and color.
func (r *Renderer) drawText(screen *ebiten.Image, c *sim_gen.DrawCmdText, sx, sy int) {
	// Get color (0 = white/default)
	var col color.RGBA
	if c.Color == 0 {
		col = color.RGBA{255, 255, 255, 255}
	} else {
		col = biomeColors[int(c.Color)%len(biomeColors)]
	}

	// Get font face for the specified size
	if r.assets != nil {
		face := r.assets.GetFontBySize(int(c.FontSize))
		if face != nil {
			// text.Draw uses baseline Y, so offset down
			lineHeight := face.Metrics().Height.Ceil()
			text.Draw(screen, c.Text, face, sx, sy+lineHeight, col)
			return
		}
	}

	// Fallback to debug text
	ebitenutil.DebugPrintAt(screen, c.Text, sx, sy)
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
