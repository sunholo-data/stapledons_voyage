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
	assets *assets.Manager
}

// NewRenderer creates a renderer with the given asset manager.
func NewRenderer(assets *assets.Manager) *Renderer {
	return &Renderer{assets: assets}
}

// RenderFrame renders the FrameOutput to the Ebiten screen.
func (r *Renderer) RenderFrame(screen *ebiten.Image, out sim_gen.FrameOutput) {
	// Get screen dimensions
	screenW, screenH := screen.Bounds().Dx(), screen.Bounds().Dy()

	// Create camera transform
	transform := camera.FromOutput(out.Camera, screenW, screenH)

	// Calculate viewport for culling
	viewport := camera.CalculateViewport(out.Camera, screenW, screenH)

	// Sort draw commands by z-index
	cmds := make([]sim_gen.DrawCmd, len(out.Draw))
	copy(cmds, out.Draw)
	sort.Slice(cmds, func(i, j int) bool {
		return getZ(cmds[i]) < getZ(cmds[j])
	})

	// Render each command with camera transform
	for _, cmd := range cmds {
		switch c := cmd.(type) {
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
	}
	return 0
}
