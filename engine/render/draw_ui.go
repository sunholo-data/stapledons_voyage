package render

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"stapledons_voyage/sim_gen"
)

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
