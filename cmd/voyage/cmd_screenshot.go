// Package main provides CLI commands for screenshot capture.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

func runScreenshotCommand(args []string) {
	// Defaults
	frames := 60
	output := "out/screenshots"
	seed := int64(1234)
	demoName := ""
	captureAll := false

	// Parse flags
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-h", "--help":
			printScreenshotHelp()
			return
		case "--frames", "-f":
			if i+1 < len(args) {
				frames, _ = strconv.Atoi(args[i+1])
				i++
			}
		case "--output", "-o":
			if i+1 < len(args) {
				output = args[i+1]
				i++
			}
		case "--seed", "-s":
			if i+1 < len(args) {
				seed, _ = strconv.ParseInt(args[i+1], 10, 64)
				i++
			}
		case "--all":
			captureAll = true
		default:
			if args[i][0] != '-' && demoName == "" {
				demoName = args[i]
			}
		}
	}

	// Ensure output directory exists
	os.MkdirAll(output, 0755)

	if captureAll {
		captureAllDemos(frames, output, seed)
		return
	}

	if demoName == "" {
		// Default to main game
		captureSingle("game", frames, output, seed)
	} else {
		// Find matching demo
		demos, _ := findDemos()
		var matched string
		for _, d := range demos {
			if d == demoName || strings.HasSuffix(d, demoName) || strings.Contains(d, demoName) {
				matched = d
				break
			}
		}
		if matched == "" {
			fmt.Fprintf(os.Stderr, "Demo not found: %s\n", demoName)
			os.Exit(1)
		}
		captureSingle(matched, frames, output, seed)
	}
}

func captureSingle(name string, frames int, outputDir string, seed int64) {
	outputPath := filepath.Join(outputDir, name+".png")
	fmt.Printf("Capturing %s (frame %d) -> %s\n", name, frames, outputPath)

	var cmdPath string
	if name == "game" {
		cmdPath = "./cmd/game"
	} else {
		cmdPath = "./cmd/" + name
	}

	cmd := exec.Command("go", "run", cmdPath,
		"-screenshot", strconv.Itoa(frames),
		"-output", outputPath,
		"-seed", strconv.FormatInt(seed, 10),
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to capture %s: %v\n", name, err)
		return
	}

	fmt.Printf("Saved: %s\n", outputPath)
}

func captureAllDemos(frames int, outputDir string, seed int64) {
	demos, err := findDemos()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error finding demos: %v\n", err)
		os.Exit(1)
	}

	// Also capture main game
	captureSingle("game", frames, outputDir, seed)

	// Capture each demo
	for _, demo := range demos {
		// Skip template
		if strings.Contains(demo, "template") {
			continue
		}
		captureSingle(demo, frames, outputDir, seed)
	}

	fmt.Printf("\nAll screenshots saved to %s/\n", outputDir)
}

func printScreenshotHelp() {
	fmt.Println(`Capture screenshots from demos

Usage:
  voyage screenshot [demo] [flags]

Flags:
  --frames, -f N    Capture after N frames (default: 60)
  --output, -o DIR  Output directory (default: out/screenshots)
  --seed, -s N      Random seed for determinism (default: 1234)
  --all             Capture all demos
  -h, --help        Show this help

Examples:
  voyage screenshot              # Capture main game
  voyage screenshot bridge       # Capture demo-game-bridge
  voyage screenshot --all        # Capture all demos
  voyage screenshot bridge -f 120 -o captures/`)
}
