// Package main provides a CLI for Stapledon's Voyage development tools.
package main

import (
	"fmt"
	"os"

	"stapledons_voyage/engine/handlers"
	"stapledons_voyage/sim_gen"
)

// initSimGenHandlers initializes the effect handlers required by sim_gen.
// AILANG codegen requires handlers to be initialized before calling any
// functions that use effects (Rand, Debug, Clock, etc).
func initSimGenHandlers() {
	sim_gen.Init(sim_gen.Handlers{
		Debug: sim_gen.NewDebugContext(),
		Rand:  handlers.NewSeededRandHandler(42), // Default seed for CLI
	})
}

func main() {
	// Initialize sim_gen handlers for CLI tools
	// This is required because AILANG codegen uses effect handlers
	initSimGenHandlers()

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "demo":
		runDemoCommand(args)
	case "watch":
		runWatchCommand(args)
	case "screenshot":
		runScreenshotCommand(args)
	case "manifest":
		runManifestCommand(args)
	case "ai":
		runAICommand(args)
	case "world":
		runWorldCommand(args)
	case "bench":
		runBenchCommand(args)
	case "perf":
		runPerfCommand(args)
	case "assets":
		runAssetsCommand(args)
	case "sim":
		runSimCommand(args)
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`Stapledon's Voyage CLI - Development Tools

Usage:
  voyage <command> [options]

Commands:
  demo        Run demos interactively or by name
  watch       Watch sim/*.ail and auto-rebuild on changes
  screenshot  Capture screenshots from demos
  manifest    Validate asset manifests

  ai          Test AI handlers (Claude, Gemini)
  world       Inspect world state (NPCs, tiles, planets)
  bench       Run performance benchmarks (human-readable)
  perf        Run benchmarks with threshold checks (CI/JSON output)
  assets      Validate game assets
  sim         Run simulation stress tests
  help        Show this help message

Use "voyage <command> -h" for more information about a command.`)
}
