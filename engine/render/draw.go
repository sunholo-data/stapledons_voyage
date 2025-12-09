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

	// Create camera transform (dereference pointer)
	cam := *out.Camera
	transform := camera.FromOutput(cam, screenW, screenH)

	// Calculate viewport for culling
	viewport := camera.CalculateViewport(cam, screenW, screenH)

	// Sort draw commands using isometric depth sorting
	// This handles both legacy Z-sorting and iso (layer, screenY) sorting
	sortables := make([]isoSortable, len(out.Draw))
	for i, cmd := range out.Draw {
		sortables[i] = isoSortable{
			cmd:     cmd,
			sortKey: getIsoSortKey(cmd, cam, screenW, screenH),
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
			r.drawIsoTile(screen, cmd.IsoTile, cam, screenW, screenH)

		case sim_gen.DrawCmdKindIsoEntity:
			r.drawIsoEntity(screen, cmd.IsoEntity, cam, screenW, screenH)

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

		case sim_gen.DrawCmdKindRectRGBA:
			c := cmd.RectRGBA
			// Screen-space rectangle with packed RGBA color (0xRRGGBBAA format)
			col := unpackRGBA(c.Rgba)
			ebitenutil.DrawRect(screen, c.X, c.Y, c.W, c.H, col)

		case sim_gen.DrawCmdKindCircleRGBA:
			c := cmd.CircleRGBA
			// Screen-space circle with packed RGBA color
			col := unpackRGBA(c.Rgba)
			r.drawCircleRGBA(screen, c.X, c.Y, c.Radius, col, c.Filled)
		}
	}

	// Render debug messages below UI panels (UI layer, not transformed)
	// Start at y=50 to avoid overlapping with camera panel
	for i, msg := range out.Debug {
		ebitenutil.DebugPrintAt(screen, msg, 10, 50+i*16)
	}
}

// getBridgeSpriteColor returns a fallback color for bridge sprite IDs
func getBridgeSpriteColor(id int64) color.RGBA {
	switch {
	// Bridge tiles (1000-1099) - BRIGHT colors for visibility
	case id == 1000: // tileFloor
		return color.RGBA{80, 90, 110, 255} // Brighter floor
	case id == 1001: // tileFloorGlow
		return color.RGBA{100, 120, 160, 255} // Glowing floor
	case id == 1002: // tileConsoleBase
		return color.RGBA{70, 80, 100, 255} // Console base
	case id == 1003: // tileWalkway
		return color.RGBA{110, 120, 140, 255} // Bright walkway
	case id == 1004: // tileDomeEdge
		return color.RGBA{60, 100, 140, 230} // Blue-tinted dome edge
	case id == 1005: // tileWall
		return color.RGBA{50, 60, 80, 255} // Wall (darker)
	case id == 1006: // tileHatch
		return color.RGBA{120, 100, 80, 255} // Warm hatch
	case id == 1007: // tileCaptainArea
		return color.RGBA{100, 90, 120, 255} // Purple-tinted captain area
	// Console sprites (1100-1149)
	case id >= 1100 && id < 1150:
		return color.RGBA{80, 120, 180, 255} // Blue consoles
	// Crew sprites (1200-1249)
	case id == 1200: // pilot
		return color.RGBA{180, 80, 80, 255} // Red
	case id == 1201: // comms
		return color.RGBA{80, 180, 80, 255} // Green
	case id == 1202: // engineer
		return color.RGBA{180, 180, 80, 255} // Yellow
	case id == 1203: // scientist
		return color.RGBA{80, 180, 180, 255} // Cyan
	case id == 1204: // captain
		return color.RGBA{180, 80, 180, 255} // Magenta
	case id == 1205: // player
		return color.RGBA{255, 255, 255, 255} // White
	case id >= 1200 && id < 1250:
		return color.RGBA{200, 150, 100, 255} // Generic crew tan
	default:
		return color.RGBA{128, 128, 128, 255} // Gray fallback
	}
}

// drawSprite draws a sprite using the asset manager with camera transform.
func (r *Renderer) drawSprite(screen *ebiten.Image, c *sim_gen.DrawCmdSprite, transform camera.Transform) {
	sx, sy := transform.WorldToScreen(c.X, c.Y)
	sw := 16.0 * transform.Scale
	sh := 16.0 * transform.Scale

	if r.assets == nil {
		// Fallback: draw colored placeholder based on sprite ID
		col := getBridgeSpriteColor(c.Id)
		ebitenutil.DrawRect(screen, sx, sy, sw, sh, col)
		return
	}

	sprite := r.assets.GetSprite(int(c.Id))
	if sprite == nil {
		// No sprite loaded - draw colored placeholder
		col := getBridgeSpriteColor(c.Id)
		ebitenutil.DrawRect(screen, sx, sy, sw, sh, col)
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

// unpackRGBA converts a packed RGBA int64 (0xRRGGBBAA format) to color.RGBA
func unpackRGBA(rgba int64) color.RGBA {
	return color.RGBA{
		R: uint8((rgba >> 24) & 0xFF),
		G: uint8((rgba >> 16) & 0xFF),
		B: uint8((rgba >> 8) & 0xFF),
		A: uint8(rgba & 0xFF),
	}
}

// drawCircleRGBA draws a circle with an RGBA color
func (r *Renderer) drawCircleRGBA(screen *ebiten.Image, x, y, radius float64, col color.RGBA, filled bool) {
	if filled {
		// Draw filled circle using pixel-by-pixel approach
		for dy := -radius; dy <= radius; dy++ {
			for dx := -radius; dx <= radius; dx++ {
				if dx*dx+dy*dy <= radius*radius {
					screen.Set(int(x+dx), int(y+dy), col)
				}
			}
		}
	} else {
		// Draw circle outline using midpoint algorithm
		cx, cy := int(x), int(y)
		r := int(radius)
		px, py := 0, r
		d := 1 - r
		for px <= py {
			screen.Set(cx+px, cy+py, col)
			screen.Set(cx-px, cy+py, col)
			screen.Set(cx+px, cy-py, col)
			screen.Set(cx-px, cy-py, col)
			screen.Set(cx+py, cy+px, col)
			screen.Set(cx-py, cy+px, col)
			screen.Set(cx+py, cy-px, col)
			screen.Set(cx-py, cy-px, col)
			if d < 0 {
				d += 2*px + 3
			} else {
				d += 2*(px-py) + 5
				py--
			}
			px++
		}
	}
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
	case sim_gen.DrawCmdKindRectRGBA:
		return int(cmd.RectRGBA.Z)
	case sim_gen.DrawCmdKindCircleRGBA:
		return int(cmd.CircleRGBA.Z)
	}
	return 0
}
