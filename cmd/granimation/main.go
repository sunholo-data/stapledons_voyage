// Package main generates frames for a black hole journey animation.
package main

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
	"stapledons_voyage/engine/display"
	"stapledons_voyage/engine/shader"
	"stapledons_voyage/engine/screenshot"
)

// journeyGame captures frames with varying GR intensity
type journeyGame struct {
	effects     *shader.Effects
	buffer      *ebiten.Image
	frameNum    int
	totalFrames int
	outputDir   string
	captured    []bool
}

func (g *journeyGame) Update() error {
	if g.frameNum >= g.totalFrames {
		return fmt.Errorf("animation complete")
	}
	return nil
}

func (g *journeyGame) Draw(screen *ebiten.Image) {
	if g.frameNum >= g.totalFrames || g.captured[g.frameNum] {
		g.frameNum++
		return
	}

	w, h := screen.Bounds().Dx(), screen.Bounds().Dy()
	if g.buffer == nil {
		g.buffer = ebiten.NewImage(w, h)
	}

	// Calculate journey phase (0 to 1 to 0)
	// First half: approaching (0 -> 1)
	// Second half: retreating (1 -> 0)
	mid := g.totalFrames / 2
	var progress float64
	if g.frameNum <= mid {
		progress = float64(g.frameNum) / float64(mid)
	} else {
		progress = float64(g.totalFrames-g.frameNum) / float64(mid)
	}

	// Map progress to phi (gravitational potential)
	// Start far away (phi ~= 0), get close (phi ~= 0.08), retreat
	phi := float32(progress * progress * 0.08) // Quadratic for more dramatic approach

	// Also vary the Schwarzschild radius based on distance
	// Closer = larger apparent size
	rs := float32(0.02 + progress*0.06)

	// Set GR demo mode with calculated parameters
	g.effects.GRWarp().SetDemoMode(0.5, 0.5, rs, phi)
	g.effects.GRWarp().SetEnabled(true)

	// Draw demo scene to buffer
	g.buffer.Clear()
	screenshot.DrawDemoScenePublic(g.buffer, w, h)

	// Apply GR effects
	g.effects.Apply(screen, g.buffer)

	// Capture frame
	bounds := screen.Bounds()
	img := image.NewRGBA(bounds)
	screen.ReadPixels(img.Pix)

	// Save frame
	filename := filepath.Join(g.outputDir, fmt.Sprintf("frame_%03d.png", g.frameNum))
	f, err := os.Create(filename)
	if err != nil {
		fmt.Printf("Error creating %s: %v\n", filename, err)
	} else {
		png.Encode(f, img)
		f.Close()
		fmt.Printf("Frame %d/%d: phi=%.4f rs=%.3f -> %s\n", g.frameNum+1, g.totalFrames, phi, rs, filename)
	}

	g.captured[g.frameNum] = true
	g.frameNum++
}

func (g *journeyGame) Layout(outsideWidth, outsideHeight int) (int, int) {
	return display.InternalWidth, display.InternalHeight
}

func main() {
	outputDir := "out/gr-animation"
	totalFrames := 60 // 60 frames = 2 seconds at 30fps

	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output dir: %v\n", err)
		os.Exit(1)
	}

	// Initialize effects
	effects := shader.NewEffects()
	if err := effects.Preload(); err != nil {
		fmt.Fprintf(os.Stderr, "Error preloading shaders: %v\n", err)
		os.Exit(1)
	}

	game := &journeyGame{
		effects:     effects,
		totalFrames: totalFrames,
		outputDir:   outputDir,
		captured:    make([]bool, totalFrames),
	}

	ebiten.SetWindowSize(display.InternalWidth, display.InternalHeight)
	ebiten.SetWindowTitle("GR Animation Generator")
	ebiten.SetScreenClearedEveryFrame(true)

	fmt.Printf("Generating %d frames...\n", totalFrames)

	if err := ebiten.RunGame(game); err != nil && err.Error() != "animation complete" {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nFrames saved to %s/\n", outputDir)
	// Use variable to avoid go vet false positive about Printf directive in ffmpeg pattern
	framePattern := "frame_%03d.png"
	fmt.Printf("To create GIF: ffmpeg -framerate 30 -i out/gr-animation/%s -vf 'scale=640:-1:flags=lanczos,split[s0][s1];[s0]palettegen[p];[s1][p]paletteuse' out/gr-journey.gif\n", framePattern)
	fmt.Printf("To create MP4: ffmpeg -framerate 30 -i out/gr-animation/%s -c:v libx264 -pix_fmt yuv420p out/gr-journey.mp4\n", framePattern)
}
