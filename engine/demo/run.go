// Package demo provides a reusable wrapper for Ebiten demos with screenshot support.
package demo

import (
	"flag"
	"fmt"
	"image/png"
	"os"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
	"stapledons_voyage/engine/display"
)

var (
	screenshotFrame  = flag.Int("screenshot", -1, "Take screenshot at frame N and exit")
	screenshotOutput = flag.String("output", "", "Screenshot output path (default: out/screenshots/<title>.png)")
)

// Config holds demo configuration.
type Config struct {
	Title  string // Window title
	Width  int    // Window width (0 = use display.InternalWidth)
	Height int    // Window height (0 = use display.InternalHeight)
}

// Run wraps any ebiten.Game with standard CLI flags and screenshot support.
// Call this instead of ebiten.RunGame for demos.
func Run(game ebiten.Game, cfg Config) error {
	flag.Parse()

	// Set defaults
	width := cfg.Width
	if width == 0 {
		width = display.InternalWidth
	}
	height := cfg.Height
	if height == 0 {
		height = display.InternalHeight
	}

	// Setup window
	ebiten.SetWindowSize(width, height)
	ebiten.SetWindowTitle(cfg.Title)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	// Wrap game with capture support if screenshot requested
	if *screenshotFrame >= 0 {
		output := *screenshotOutput
		if output == "" {
			output = fmt.Sprintf("out/screenshots/%s.png", sanitizeFilename(cfg.Title))
		}
		game = &captureWrapper{
			inner:       game,
			targetFrame: *screenshotFrame,
			outputPath:  output,
		}
	}

	return ebiten.RunGame(game)
}

// captureWrapper wraps an ebiten.Game to capture a screenshot at a specific frame.
type captureWrapper struct {
	inner       ebiten.Game
	frame       int
	targetFrame int
	outputPath  string
}

func (w *captureWrapper) Update() error {
	err := w.inner.Update()
	w.frame++
	return err
}

func (w *captureWrapper) Draw(screen *ebiten.Image) {
	w.inner.Draw(screen)

	if w.frame >= w.targetFrame {
		if err := w.saveScreenshot(screen); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to save screenshot: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Screenshot saved: %s\n", w.outputPath)
		os.Exit(0)
	}
}

func (w *captureWrapper) Layout(outsideWidth, outsideHeight int) (int, int) {
	return w.inner.Layout(outsideWidth, outsideHeight)
}

func (w *captureWrapper) saveScreenshot(screen *ebiten.Image) error {
	// Ensure output directory exists
	dir := filepath.Dir(w.outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	f, err := os.Create(w.outputPath)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer f.Close()

	if err := png.Encode(f, screen); err != nil {
		return fmt.Errorf("encode PNG: %w", err)
	}

	return nil
}

// sanitizeFilename makes a string safe for use as a filename.
func sanitizeFilename(s string) string {
	result := make([]byte, 0, len(s))
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z':
			result = append(result, byte(r))
		case r >= 'A' && r <= 'Z':
			result = append(result, byte(r-'A'+'a'))
		case r >= '0' && r <= '9':
			result = append(result, byte(r))
		case r == ' ' || r == '-' || r == '_':
			result = append(result, '-')
		}
	}
	if len(result) == 0 {
		return "demo"
	}
	return string(result)
}
