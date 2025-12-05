// Package main provides a CLI for capturing GR effect screenshots.
package main

import (
	"fmt"
	"os"

	"stapledons_voyage/engine/screenshot"
)

func main() {
	// Ensure output directory exists
	if err := os.MkdirAll("out/gr-screenshots", 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output dir: %v\n", err)
		os.Exit(1)
	}

	screenshots := []struct {
		name    string
		effects string
	}{
		{"gr-none", ""},
		{"gr-subtle", "gr_subtle"},
		{"gr-strong", "gr_strong"},
		{"gr-extreme", "gr_extreme"},
		{"gr-with-bloom", "gr_strong,bloom"},
		{"gr-with-sr", "gr_strong,sr_warp"},
		{"gr-all-relativity", "gr_extreme,sr_warp,bloom"},
	}

	for _, s := range screenshots {
		cfg := screenshot.Config{
			Frames:     1,
			OutputPath: fmt.Sprintf("out/gr-screenshots/%s.png", s.name),
			DemoScene:  true,
			Effects:    s.effects,
			Velocity:   0.5, // 0.5c for SR effects
		}

		fmt.Printf("Capturing %s... ", s.name)
		if err := screenshot.Capture(cfg); err != nil {
			fmt.Printf("ERROR: %v\n", err)
		} else {
			fmt.Printf("OK -> %s\n", cfg.OutputPath)
		}
	}

	fmt.Println("\nDone! Screenshots saved to out/gr-screenshots/")
}
