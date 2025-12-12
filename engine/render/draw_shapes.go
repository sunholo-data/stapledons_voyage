package render

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font"
	"stapledons_voyage/engine/camera"
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

// drawSpireBg renders a placeholder spire silhouette for the MidBackground layer.
// The spire is the central structural element running vertically through the bubble ship.
// In the final game, this would be a proper sprite; this placeholder shows the parallax effect.
func (r *Renderer) drawSpireBg(screen *ebiten.Image, screenW, screenH int) {
	// Spire is a tall, narrow vertical structure in the center of the screen
	// It uses MidBackground parallax (0.3x), so it moves slower than the scene

	// Dark silhouette color with slight transparency
	spireColor := color.RGBA{30, 35, 50, 200}
	highlightColor := color.RGBA{50, 60, 80, 180}

	// Center the spire
	centerX := float32(screenW) / 2

	// Main spire body - tall narrow rectangle
	spireWidth := float32(60)
	spireHeight := float32(screenH) * 1.5 // Taller than screen
	spireTop := float32(screenH)/2 - spireHeight/2

	// Draw main spire body
	vector.DrawFilledRect(screen,
		centerX-spireWidth/2, spireTop,
		spireWidth, spireHeight,
		spireColor, true)

	// Add some structural details (horizontal bands)
	bandColor := color.RGBA{40, 45, 60, 220}
	bandHeight := float32(8)
	bandSpacing := float32(80)

	for y := spireTop; y < spireTop+spireHeight; y += bandSpacing {
		vector.DrawFilledRect(screen,
			centerX-spireWidth/2-5, y,
			spireWidth+10, bandHeight,
			bandColor, true)
	}

	// Add slight glow/edge highlight on one side
	vector.DrawFilledRect(screen,
		centerX-spireWidth/2, spireTop,
		3, spireHeight,
		highlightColor, true)
}

// drawSpireBgParallax renders the spire silhouette with parallax offset.
// The transform contains the parallax-adjusted camera offset for this layer.
func (r *Renderer) drawSpireBgParallax(screen *ebiten.Image, screenW, screenH int, transform camera.Transform) {
	// Calculate parallax shift from transform
	// transform.OffsetX = screenW/2 - camera.X * parallaxFactor * zoom
	// So the shift from center is: transform.OffsetX - screenW/2
	parallaxShiftX := float32(transform.OffsetX - float64(screenW)/2)

	// Dark silhouette color with slight transparency
	spireColor := color.RGBA{30, 35, 50, 200}
	highlightColor := color.RGBA{50, 60, 80, 180}

	// Center the spire with parallax offset
	centerX := float32(screenW)/2 + parallaxShiftX

	// Main spire body - tall narrow rectangle
	spireWidth := float32(60)
	spireHeight := float32(screenH) * 1.5 // Taller than screen
	spireTop := float32(screenH)/2 - spireHeight/2

	// Draw main spire body
	vector.DrawFilledRect(screen,
		centerX-spireWidth/2, spireTop,
		spireWidth, spireHeight,
		spireColor, true)

	// Add some structural details (horizontal bands)
	bandColor := color.RGBA{40, 45, 60, 220}
	bandHeight := float32(8)
	bandSpacing := float32(80)

	for y := spireTop; y < spireTop+spireHeight; y += bandSpacing {
		vector.DrawFilledRect(screen,
			centerX-spireWidth/2-5, y,
			spireWidth+10, bandHeight,
			bandColor, true)
	}

	// Add slight glow/edge highlight on one side
	vector.DrawFilledRect(screen,
		centerX-spireWidth/2, spireTop,
		3, spireHeight,
		highlightColor, true)
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
