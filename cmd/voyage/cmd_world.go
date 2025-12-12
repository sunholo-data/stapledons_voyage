// Package main provides a CLI for Stapledon's Voyage development tools.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"stapledons_voyage/sim_gen"
)

// runWorldCommand handles the "world" subcommand for inspecting world state.
func runWorldCommand(args []string) {
	fs := flag.NewFlagSet("world", flag.ExitOnError)
	seed := fs.Int("seed", 42, "World seed for initialization")
	steps := fs.Int("steps", 0, "Run N steps before inspection")
	jsonOutput := fs.Bool("json", false, "Output as JSON")
	summary := fs.Bool("summary", false, "Show summary only")

	fs.Usage = func() {
		fmt.Println(`Inspect world state

Usage:
  voyage world [options]

Options:`)
		fs.PrintDefaults()
		fmt.Println(`
Examples:
  voyage world                    # Inspect world with default seed
  voyage world -seed 123          # Use specific seed
  voyage world -steps 100         # Run 100 steps first
  voyage world -json              # Output as JSON
  voyage world -summary           # Show summary stats only`)
	}

	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	// Initialize world - returns *World in v0.5.8+
	world := sim_gen.InitWorld(int64(*seed))

	// Run steps if requested
	if *steps > 0 {
		fmt.Printf("Running %d steps...\n", *steps)
		input := &sim_gen.FrameInput{}
		for i := 0; i < *steps; i++ {
			result := sim_gen.Step(world, input)
			tuple, ok := result.([]interface{})
			if !ok || len(tuple) != 2 {
				fmt.Fprintln(os.Stderr, "Error: unexpected Step result")
				os.Exit(1)
			}
			if w, ok := tuple[0].(*sim_gen.World); ok {
				world = w
			}
		}
	}

	// Output world state
	if *jsonOutput {
		data, err := json.MarshalIndent(world, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error marshaling world: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(data))
		return
	}

	if *summary {
		printWorldSummary(world)
		return
	}

	// Full inspection
	printWorldDetails(world)
}

func printWorldSummary(world interface{}) {
	fmt.Println("World State Summary")
	fmt.Println("===================")
	fmt.Printf("Type: %T\n", world)

	// Try to extract common fields
	if w, ok := world.(map[string]interface{}); ok {
		if tick, ok := w["Tick"]; ok {
			fmt.Printf("Tick: %v\n", tick)
		}
		if npcs, ok := w["NPCs"].([]interface{}); ok {
			fmt.Printf("NPCs: %d\n", len(npcs))
		}
		if tiles, ok := w["Tiles"].([]interface{}); ok {
			fmt.Printf("Tiles: %d\n", len(tiles))
		}
		if planets, ok := w["Planets"].([]interface{}); ok {
			fmt.Printf("Planets: %d\n", len(planets))
		}
	} else {
		// Generic output for other types
		data, _ := json.MarshalIndent(world, "", "  ")
		// Count lines as a rough size estimate
		lines := strings.Count(string(data), "\n")
		fmt.Printf("State size: ~%d lines\n", lines)
	}
}

func printWorldDetails(world interface{}) {
	fmt.Println("World State Details")
	fmt.Println("===================")
	fmt.Printf("Type: %T\n\n", world)

	data, err := json.MarshalIndent(world, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return
	}

	// Pretty print with truncation for large arrays
	output := string(data)
	if len(output) > 10000 {
		fmt.Println(output[:10000])
		fmt.Printf("\n... (truncated, %d bytes total)\n", len(output))
		fmt.Println("Use -json for full output")
	} else {
		fmt.Println(output)
	}
}
