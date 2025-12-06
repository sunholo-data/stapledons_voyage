package render

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font"
	"stapledons_voyage/sim_gen"
)

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
