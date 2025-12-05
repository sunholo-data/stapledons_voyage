package screenshot

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"stapledons_voyage/engine/display"
	"stapledons_voyage/engine/shader"
)

// demoGame implements ebiten.Game for effects demo scene
type demoGame struct {
	config      Config
	effects     *shader.Effects
	buffer      *ebiten.Image
	captured    bool
	capturedImg *image.RGBA
	frame       int
}

func (g *demoGame) Update() error {
	g.frame++
	if g.captured {
		return errComplete
	}
	return nil
}

func (g *demoGame) Draw(screen *ebiten.Image) {
	w, h := screen.Bounds().Dx(), screen.Bounds().Dy()

	// Ensure buffer exists
	if g.buffer == nil {
		g.buffer = ebiten.NewImage(w, h)
	}
	g.buffer.Clear()

	// Render demo scene to buffer
	drawDemoScene(g.buffer, w, h)

	// Apply effects if enabled
	if g.effects != nil && hasAnyEffect(g.effects) {
		g.effects.Apply(screen, g.buffer)
	} else {
		screen.DrawImage(g.buffer, nil)
	}

	// Capture after first frame
	if !g.captured {
		g.captured = true
		bounds := screen.Bounds()
		g.capturedImg = image.NewRGBA(bounds)
		screen.ReadPixels(g.capturedImg.Pix)
	}
}

func (g *demoGame) Layout(outsideWidth, outsideHeight int) (int, int) {
	return display.InternalWidth, display.InternalHeight
}

// drawDemoScene renders content designed to showcase shader effects
func drawDemoScene(screen *ebiten.Image, w, h int) {
	// Dark space background (essential for bloom visibility)
	screen.Fill(color.RGBA{5, 5, 15, 255})

	// Draw gradient nebula regions (for vignette visibility)
	drawNebula(screen, w, h)

	// Draw stars of various sizes and colors (for bloom)
	drawStars(screen, w, h)

	// Draw bright geometric shapes (for aberration edge detection)
	drawShapes(screen, w, h)

	// Draw text (for CRT scanlines and aberration)
	drawDemoText(screen, w, h)

	// Draw UI-like elements (high contrast edges)
	drawUIElements(screen, w, h)
}

// drawNebula creates colorful gradient regions
func drawNebula(screen *ebiten.Image, w, h int) {
	// Purple nebula in upper left
	drawGradientCircle(screen, float32(w)*0.2, float32(h)*0.3, 150,
		color.RGBA{80, 20, 120, 100})

	// Blue nebula in center
	drawGradientCircle(screen, float32(w)*0.5, float32(h)*0.5, 200,
		color.RGBA{20, 40, 100, 80})

	// Orange nebula in lower right
	drawGradientCircle(screen, float32(w)*0.8, float32(h)*0.7, 120,
		color.RGBA{120, 60, 20, 90})
}

// drawGradientCircle draws a soft circular gradient
func drawGradientCircle(screen *ebiten.Image, cx, cy, radius float32, col color.RGBA) {
	// Draw multiple concentric circles with decreasing alpha
	for r := radius; r > 0; r -= 5 {
		alpha := uint8(float32(col.A) * (r / radius) * 0.5)
		c := color.RGBA{col.R, col.G, col.B, alpha}
		vector.DrawFilledCircle(screen, cx, cy, r, c, true)
	}
}

// drawStars places bright points for bloom effect
func drawStars(screen *ebiten.Image, w, h int) {
	// Predefined star positions and properties
	stars := []struct {
		x, y   float32
		radius float32
		col    color.RGBA
	}{
		// Bright white stars (will bloom heavily)
		{float32(w) * 0.15, float32(h) * 0.2, 4, color.RGBA{255, 255, 255, 255}},
		{float32(w) * 0.85, float32(h) * 0.15, 5, color.RGBA{255, 255, 255, 255}},
		{float32(w) * 0.5, float32(h) * 0.1, 3, color.RGBA{255, 255, 255, 255}},

		// Blue-white stars
		{float32(w) * 0.3, float32(h) * 0.4, 3, color.RGBA{200, 220, 255, 255}},
		{float32(w) * 0.7, float32(h) * 0.35, 4, color.RGBA{180, 200, 255, 255}},

		// Yellow stars (like our sun)
		{float32(w) * 0.25, float32(h) * 0.65, 6, color.RGBA{255, 255, 200, 255}},
		{float32(w) * 0.6, float32(h) * 0.7, 3, color.RGBA{255, 250, 180, 255}},

		// Orange giants
		{float32(w) * 0.75, float32(h) * 0.55, 5, color.RGBA{255, 180, 100, 255}},
		{float32(w) * 0.4, float32(h) * 0.85, 4, color.RGBA{255, 160, 80, 255}},

		// Red giants
		{float32(w) * 0.9, float32(h) * 0.8, 7, color.RGBA{255, 120, 100, 255}},
		{float32(w) * 0.1, float32(h) * 0.9, 4, color.RGBA{255, 100, 80, 255}},

		// Dim background stars
		{float32(w) * 0.35, float32(h) * 0.25, 1, color.RGBA{150, 150, 150, 255}},
		{float32(w) * 0.55, float32(h) * 0.45, 1, color.RGBA{130, 130, 130, 255}},
		{float32(w) * 0.45, float32(h) * 0.55, 1, color.RGBA{120, 120, 120, 255}},
		{float32(w) * 0.65, float32(h) * 0.25, 1, color.RGBA{140, 140, 140, 255}},
		{float32(w) * 0.2, float32(h) * 0.75, 1, color.RGBA{130, 130, 130, 255}},
		{float32(w) * 0.8, float32(h) * 0.45, 1, color.RGBA{140, 140, 140, 255}},
	}

	for _, s := range stars {
		// Draw star core
		vector.DrawFilledCircle(screen, s.x, s.y, s.radius, s.col, true)

		// Add glow around bright stars
		if s.radius >= 3 {
			glowCol := color.RGBA{s.col.R, s.col.G, s.col.B, 100}
			vector.DrawFilledCircle(screen, s.x, s.y, s.radius*2, glowCol, true)
		}
	}
}

// drawShapes adds geometric elements with sharp edges
func drawShapes(screen *ebiten.Image, w, h int) {
	// White rectangles (high contrast for aberration)
	ebitenutil.DrawRect(screen, float64(w)*0.05, float64(h)*0.45, 8, 60,
		color.RGBA{255, 255, 255, 255})
	ebitenutil.DrawRect(screen, float64(w)*0.95-8, float64(h)*0.45, 8, 60,
		color.RGBA{255, 255, 255, 255})

	// Colored bars (shows RGB separation in aberration)
	barY := float64(h) * 0.92
	barH := 12.0
	barW := float64(w) * 0.3

	// Red bar
	ebitenutil.DrawRect(screen, float64(w)*0.1, barY, barW, barH,
		color.RGBA{255, 60, 60, 255})
	// Green bar
	ebitenutil.DrawRect(screen, float64(w)*0.35, barY, barW, barH,
		color.RGBA{60, 255, 60, 255})
	// Blue bar
	ebitenutil.DrawRect(screen, float64(w)*0.6, barY, barW, barH,
		color.RGBA{60, 60, 255, 255})

	// White crosshair in center (shows aberration clearly)
	cx, cy := float64(w)/2, float64(h)/2
	// Horizontal line
	ebitenutil.DrawRect(screen, cx-40, cy-1, 80, 2, color.RGBA{255, 255, 255, 200})
	// Vertical line
	ebitenutil.DrawRect(screen, cx-1, cy-40, 2, 80, color.RGBA{255, 255, 255, 200})
}

// drawDemoText adds text for CRT scanlines and aberration visibility
func drawDemoText(screen *ebiten.Image, w, h int) {
	// Use basic font (always available)
	face := basicfont.Face7x13

	// Title text
	drawTextCentered(screen, "SHADER EFFECTS DEMO", float64(w)/2, 30, face, color.RGBA{255, 255, 255, 255})

	// Effect labels
	drawTextAt(screen, "BLOOM: Stars glow", 20, float64(h)*0.15, face, color.RGBA{200, 200, 255, 255})
	drawTextAt(screen, "VIGNETTE: Edges darken", 20, float64(h)*0.15+20, face, color.RGBA{200, 200, 255, 255})
	drawTextAt(screen, "CRT: Scanlines + curve", 20, float64(h)*0.15+40, face, color.RGBA{200, 200, 255, 255})
	drawTextAt(screen, "ABERRATION: RGB split", 20, float64(h)*0.15+60, face, color.RGBA{200, 200, 255, 255})

	// Grid of text for scanline visibility
	gridY := float64(h) * 0.78
	for i := 0; i < 5; i++ {
		y := gridY + float64(i)*14
		drawTextAt(screen, "||||||||||||||||||||||||||||||||||||||||", 60, y, face, color.RGBA{100, 100, 100, 255})
	}
}

// drawUIElements adds UI-like panels
func drawUIElements(screen *ebiten.Image, w, h int) {
	// Top bar
	ebitenutil.DrawRect(screen, 0, 0, float64(w), 50, color.RGBA{20, 20, 40, 200})

	// Bottom bar
	ebitenutil.DrawRect(screen, 0, float64(h)-25, float64(w), 25, color.RGBA{20, 20, 40, 200})

	// Side panels (for vignette to show edge darkening)
	panelW := 15.0
	// Left panel gradient
	for i := 0; i < int(panelW); i++ {
		alpha := uint8(200 - i*12)
		ebitenutil.DrawRect(screen, float64(i), 50, 1, float64(h)-75,
			color.RGBA{30, 30, 50, alpha})
	}
	// Right panel gradient
	for i := 0; i < int(panelW); i++ {
		alpha := uint8(200 - i*12)
		ebitenutil.DrawRect(screen, float64(w-1-i), 50, 1, float64(h)-75,
			color.RGBA{30, 30, 50, alpha})
	}
}

func drawTextCentered(screen *ebiten.Image, s string, x, y float64, face font.Face, col color.Color) {
	bounds, _ := font.BoundString(face, s)
	textW := (bounds.Max.X - bounds.Min.X).Ceil()
	text.Draw(screen, s, face, int(x)-textW/2, int(y), col)
}

func drawTextAt(screen *ebiten.Image, s string, x, y float64, face font.Face, col color.Color) {
	text.Draw(screen, s, face, int(x), int(y), col)
}

func hasAnyEffect(effects *shader.Effects) bool {
	if effects.SRWarp().IsEnabled() {
		return true
	}
	if effects.Bloom().IsEnabled() {
		return true
	}
	return len(effects.Pipeline().EnabledEffects()) > 0
}

// CaptureDemo runs the demo scene and captures with effects
func CaptureDemo(cfg Config) error {
	// Initialize shader effects
	effects := shader.NewEffects()
	if err := effects.Preload(); err != nil {
		return err
	}
	enableEffects(effects, cfg.Effects)

	// Apply velocity for SR warp if specified
	if cfg.Velocity > 0 {
		enableSRWarpWithVelocity(effects, cfg.Velocity, cfg.ViewAngle)
	}

	game := &demoGame{
		config:  cfg,
		effects: effects,
	}

	// Set window for headless capture
	ebiten.SetWindowSize(display.InternalWidth, display.InternalHeight)
	ebiten.SetWindowTitle("Effects Demo")
	ebiten.SetScreenClearedEveryFrame(true)

	// Run until captured
	if err := ebiten.RunGame(game); err != nil && err != errComplete {
		return err
	}

	return saveImage(game.capturedImg, cfg.OutputPath)
}

// saveImage writes the captured image to disk
func saveImage(img *image.RGBA, path string) error {
	if img == nil {
		return errNilImage
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	return png.Encode(f, img)
}

// Sentinel errors
var (
	errComplete = errorf("demo complete")
	errNilImage = errorf("no image captured")
)

func errorf(s string) error {
	return &constError{s}
}

type constError struct {
	s string
}

func (e *constError) Error() string {
	return e.s
}

