package render

import (
	"image/color"
	"sort"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
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
	assets         *assets.Manager
	anims          *AnimationManager
	lastTick       uint64        // Track simulation tick for animation updates
	galaxyBg       *ebiten.Image // Galaxy background image (loaded lazily)
	galaxyBgLoaded bool          // Whether we've attempted to load the background
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
