package render

import (
	"image/color"
	"sort"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"stapledons_voyage/sim_gen"
)

// Biome colors for rendering tiles
var biomeColors = []color.RGBA{
	{0, 100, 200, 255},  // 0: Water (blue)
	{34, 139, 34, 255},  // 1: Forest (green)
	{210, 180, 140, 255}, // 2: Desert (tan)
	{139, 90, 43, 255},  // 3: Mountain (brown)
}

// RenderFrame renders the FrameOutput to the Ebiten screen
func RenderFrame(screen *ebiten.Image, out sim_gen.FrameOutput) {
	// Sort draw commands by z-index
	cmds := make([]sim_gen.DrawCmd, len(out.Draw))
	copy(cmds, out.Draw)
	sort.Slice(cmds, func(i, j int) bool {
		return getZ(cmds[i]) < getZ(cmds[j])
	})

	// Render each command
	for _, cmd := range cmds {
		switch c := cmd.(type) {
		case sim_gen.DrawCmdRect:
			col := biomeColors[c.Color%len(biomeColors)]
			ebitenutil.DrawRect(screen, c.X, c.Y, c.W, c.H, col)
		case sim_gen.DrawCmdSprite:
			// TODO: Load and draw sprite by ID
			// For now, draw a placeholder rectangle
			ebitenutil.DrawRect(screen, c.X, c.Y, 16, 16, color.White)
		case sim_gen.DrawCmdText:
			ebitenutil.DebugPrintAt(screen, c.Text, int(c.X), int(c.Y))
		}
	}
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
