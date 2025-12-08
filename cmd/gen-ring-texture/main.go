// cmd/gen-ring-texture/main.go
// Generates procedural Saturn ring textures for 3D rendering.
// The output texture maps V to radial distance (inner=0, outer=1).
package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"math/rand"
	"os"
)

var (
	width  = flag.Int("width", 1, "Texture width (only 1 needed for radial)")
	height = flag.Int("height", 512, "Texture height (radial resolution)")
	output = flag.String("output", "assets/planets/saturn_ring_gen.png", "Output file path")
	seed   = flag.Int64("seed", 42, "Random seed for ring variation")
)

func main() {
	flag.Parse()

	rand.Seed(*seed)

	// Create image - width is 1 (uniform around ring), height is radial
	img := image.NewRGBA(image.Rect(0, 0, *width, *height))

	// Generate Saturn-like ring bands
	for y := 0; y < *height; y++ {
		// t goes from 0 (inner) to 1 (outer)
		t := float64(y) / float64(*height-1)

		// Get ring color and alpha at this radial position
		r, g, b, a := getRingColorAt(t)

		c := color.RGBA{r, g, b, a}
		for x := 0; x < *width; x++ {
			img.Set(x, y, c)
		}
	}

	// Save
	f, err := os.Create(*output)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create file: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	if err := png.Encode(f, img); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to encode PNG: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Generated ring texture: %s (%dx%d)\n", *output, *width, *height)
}

// getRingColorAt returns RGBA for a given radial position t (0=inner, 1=outer)
func getRingColorAt(t float64) (r, g, b, a uint8) {
	// Saturn's rings have distinct bands:
	// - D Ring: 0.00-0.10 (faint)
	// - C Ring: 0.10-0.28 (brownish, semi-transparent)
	// - B Ring: 0.28-0.58 (brightest, tan/cream)
	// - Cassini Division: 0.58-0.63 (dark gap)
	// - A Ring: 0.63-0.88 (moderately bright)
	// - Encke Gap: ~0.80 (thin gap)
	// - F Ring: 0.88-1.00 (faint, narrow)

	// Base ring color (tan/cream)
	baseR := 0.85
	baseG := 0.75
	baseB := 0.60

	// Default opacity
	opacity := 0.0

	switch {
	case t < 0.10:
		// D Ring - very faint
		opacity = 0.1 + 0.1*t/0.10
		baseR, baseG, baseB = 0.6, 0.55, 0.45

	case t < 0.28:
		// C Ring - semi-transparent brownish
		tt := (t - 0.10) / 0.18
		opacity = 0.3 + 0.3*tt
		baseR = 0.7 + 0.1*tt
		baseG = 0.6 + 0.1*tt
		baseB = 0.5

	case t < 0.58:
		// B Ring - brightest, with variation
		tt := (t - 0.28) / 0.30
		opacity = 0.85 + 0.15*math.Sin(tt*math.Pi)
		// Add subtle banding
		banding := 0.05 * math.Sin(tt*40*math.Pi)
		baseR = 0.88 + banding
		baseG = 0.78 + banding
		baseB = 0.65 + banding

	case t < 0.63:
		// Cassini Division - dark gap
		opacity = 0.05 + 0.1*rand.Float64()
		baseR, baseG, baseB = 0.3, 0.25, 0.2

	case t < 0.88:
		// A Ring - moderately bright
		tt := (t - 0.63) / 0.25
		opacity = 0.6 + 0.2*math.Sin(tt*math.Pi)

		// Encke Gap at ~0.80 (relative position ~0.68 in A Ring)
		if t > 0.79 && t < 0.81 {
			opacity = 0.1
			baseR, baseG, baseB = 0.3, 0.25, 0.2
		}

		// Add subtle banding
		banding := 0.03 * math.Sin(tt*30*math.Pi)
		baseR = 0.82 + banding
		baseG = 0.72 + banding
		baseB = 0.58 + banding

	default:
		// F Ring - faint and narrow
		tt := (t - 0.88) / 0.12
		opacity = 0.3 * (1 - tt)
		baseR, baseG, baseB = 0.75, 0.65, 0.55
	}

	// Add some noise for texture
	noise := 0.02 * (rand.Float64() - 0.5)
	baseR = clamp(baseR + noise)
	baseG = clamp(baseG + noise)
	baseB = clamp(baseB + noise)

	r = uint8(baseR * 255)
	g = uint8(baseG * 255)
	b = uint8(baseB * 255)
	a = uint8(opacity * 255)

	return
}

func clamp(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
