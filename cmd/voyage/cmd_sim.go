// Package main provides a CLI for Stapledon's Voyage development tools.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"stapledons_voyage/sim_gen"
)

// runSimCommand handles the "sim" subcommand for simulation stress tests.
func runSimCommand(args []string) {
	fs := flag.NewFlagSet("sim", flag.ExitOnError)
	steps := fs.Int("steps", 10000, "Number of steps to simulate")
	seed := fs.Int("seed", 42, "World seed")
	checkInterval := fs.Int("check", 1000, "Interval for progress output")
	validateState := fs.Bool("validate", false, "Validate state after each step")

	fs.Usage = func() {
		fmt.Println(`Run simulation stress tests

Usage:
  voyage sim [options]

Options:`)
		fs.PrintDefaults()
		fmt.Println(`
Examples:
  voyage sim                      # Run 10000 steps
  voyage sim -steps 100000        # Longer stress test
  voyage sim -validate            # Validate state each step
  voyage sim -seed 123            # Specific seed`)
	}

	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	fmt.Println("Simulation Stress Test")
	fmt.Println("======================")
	fmt.Printf("Steps: %d, Seed: %d\n\n", *steps, *seed)

	world := sim_gen.InitWorld(int64(*seed))
	input := &sim_gen.FrameInput{}

	start := time.Now()
	lastCheck := start
	errors := 0

	for i := 1; i <= *steps; i++ {
		result := sim_gen.Step(world, input)
		tuple, ok := result.([]interface{})
		if !ok || len(tuple) != 2 {
			fmt.Printf("Step %d: unexpected result type: %T\n", i, result)
			errors++
			if errors > 10 {
				fmt.Println("Too many errors, stopping.")
				os.Exit(1)
			}
			continue
		}
		if w, ok := tuple[0].(*sim_gen.World); ok {
			world = w
		}

		if *validateState {
			if err := validateWorld(world); err != nil {
				fmt.Printf("Step %d: validation error: %v\n", i, err)
				errors++
			}
		}

		if i%*checkInterval == 0 {
			elapsed := time.Since(lastCheck)
			stepsPerSec := float64(*checkInterval) / elapsed.Seconds()
			fmt.Printf("Step %d: %.0f steps/sec\n", i, stepsPerSec)
			lastCheck = time.Now()
		}
	}

	elapsed := time.Since(start)
	stepsPerSec := float64(*steps) / elapsed.Seconds()

	fmt.Println()
	fmt.Println("Results:")
	fmt.Printf("  Total time: %v\n", elapsed)
	fmt.Printf("  Steps/sec:  %.0f\n", stepsPerSec)
	fmt.Printf("  Errors:     %d\n", errors)

	// Final state summary
	fmt.Println()
	fmt.Println("Final State:")
	printWorldSummary(world)
}

func validateWorld(world interface{}) error {
	// Basic validation - check it's not nil and can be serialized
	if world == nil {
		return fmt.Errorf("world is nil")
	}

	// Try to marshal to verify structure
	_, err := json.Marshal(world)
	if err != nil {
		return fmt.Errorf("world not serializable: %v", err)
	}

	return nil
}
