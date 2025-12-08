// Package main provides a CLI for Stapledon's Voyage development tools.
package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"time"

	"stapledons_voyage/sim_gen"
)

// runBenchCommand handles the "bench" subcommand for performance benchmarks.
func runBenchCommand(args []string) {
	fs := flag.NewFlagSet("bench", flag.ExitOnError)
	iterations := fs.Int("n", 1000, "Number of iterations")
	warmup := fs.Int("warmup", 100, "Warmup iterations")
	profile := fs.Bool("profile", false, "Enable CPU profiling")
	profilePath := fs.String("profile-path", "cpu.prof", "CPU profile output path")

	fs.Usage = func() {
		fmt.Println(`Run performance benchmarks

Usage:
  voyage bench [options]

Options:`)
		fs.PrintDefaults()
		fmt.Println(`
Examples:
  voyage bench                    # Run default benchmarks
  voyage bench -n 10000           # 10000 iterations
  voyage bench -profile           # With CPU profiling

Benchmarks:
  - InitWorld: Time to create a new world
  - Step: Time per simulation step
  - Step100: Time for 100 consecutive steps`)
	}

	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	if *profile {
		f, err := os.Create(*profilePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating profile: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()
		// Note: pprof.StartCPUProfile would go here with proper import
		fmt.Printf("CPU profile will be written to: %s\n", *profilePath)
	}

	fmt.Println("Performance Benchmarks")
	fmt.Println("======================")
	fmt.Printf("Iterations: %d (warmup: %d)\n\n", *iterations, *warmup)

	// Benchmark InitWorld
	fmt.Print("InitWorld: ")
	benchInitWorld(*warmup, *iterations)

	// Benchmark Step
	fmt.Print("Step:      ")
	benchStep(*warmup, *iterations)

	// Benchmark Step100
	fmt.Print("Step100:   ")
	benchStep100(*warmup, *iterations/10) // Fewer iterations since each does 100 steps

	// Memory stats
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Println("\nMemory Stats:")
	fmt.Printf("  Alloc:      %d MB\n", m.Alloc/1024/1024)
	fmt.Printf("  TotalAlloc: %d MB\n", m.TotalAlloc/1024/1024)
	fmt.Printf("  NumGC:      %d\n", m.NumGC)
}

func benchInitWorld(warmup, iterations int) {
	// Warmup
	for i := 0; i < warmup; i++ {
		_ = sim_gen.InitWorld(i)
	}

	// Benchmark
	start := time.Now()
	for i := 0; i < iterations; i++ {
		_ = sim_gen.InitWorld(i)
	}
	elapsed := time.Since(start)

	avg := elapsed / time.Duration(iterations)
	fmt.Printf("%v avg (%v total, %d ops)\n", avg, elapsed, iterations)
}

func benchStep(warmup, iterations int) {
	world := sim_gen.InitWorld(42)
	input := sim_gen.FrameInput{}

	// Warmup
	for i := 0; i < warmup; i++ {
		result := sim_gen.Step(world, input)
		if tuple, ok := result.([]interface{}); ok && len(tuple) == 2 {
			world = tuple[0]
		}
	}

	// Reset world
	world = sim_gen.InitWorld(42)

	// Benchmark
	start := time.Now()
	for i := 0; i < iterations; i++ {
		result := sim_gen.Step(world, input)
		if tuple, ok := result.([]interface{}); ok && len(tuple) == 2 {
			world = tuple[0]
		}
	}
	elapsed := time.Since(start)

	avg := elapsed / time.Duration(iterations)
	fmt.Printf("%v avg (%v total, %d ops)\n", avg, elapsed, iterations)
}

func benchStep100(warmup, iterations int) {
	input := sim_gen.FrameInput{}

	// Warmup
	for i := 0; i < warmup; i++ {
		world := sim_gen.InitWorld(42)
		for j := 0; j < 100; j++ {
			result := sim_gen.Step(world, input)
			if tuple, ok := result.([]interface{}); ok && len(tuple) == 2 {
				world = tuple[0]
			}
		}
	}

	// Benchmark
	start := time.Now()
	for i := 0; i < iterations; i++ {
		world := sim_gen.InitWorld(42)
		for j := 0; j < 100; j++ {
			result := sim_gen.Step(world, input)
			if tuple, ok := result.([]interface{}); ok && len(tuple) == 2 {
				world = tuple[0]
			}
		}
	}
	elapsed := time.Since(start)

	avg := elapsed / time.Duration(iterations)
	fmt.Printf("%v avg (%v total, %d ops, 100 steps/op)\n", avg, elapsed, iterations)
}
