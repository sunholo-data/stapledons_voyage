// Package main provides a CLI for Stapledon's Voyage development tools.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"stapledons_voyage/engine/bench"
)

// perfFlags holds all parsed perf command flags
type perfFlags struct {
	iterations      int
	warmup          int
	outputPath      string
	failOnThreshold bool
	stepMax         time.Duration
	initMax         time.Duration
	step100Max      time.Duration
	quiet           bool
}

// parsePerfFlags parses and returns perf command flags
func parsePerfFlags(args []string) perfFlags {
	fs := flag.NewFlagSet("perf", flag.ExitOnError)
	flags := perfFlags{}

	fs.IntVar(&flags.iterations, "n", 1000, "Number of iterations")
	fs.IntVar(&flags.warmup, "warmup", 100, "Warmup iterations")
	fs.StringVar(&flags.outputPath, "o", "", "Output JSON file path (default: stdout)")
	fs.BoolVar(&flags.failOnThreshold, "fail", true, "Exit with code 1 if thresholds exceeded")
	fs.DurationVar(&flags.stepMax, "step-max", 5*time.Millisecond, "Max time for Step()")
	fs.DurationVar(&flags.initMax, "init-max", 100*time.Millisecond, "Max time for InitWorld()")
	fs.DurationVar(&flags.step100Max, "step100-max", 500*time.Millisecond, "Max time for 100 steps")
	fs.BoolVar(&flags.quiet, "q", false, "Quiet mode (only output JSON)")

	fs.Usage = func() {
		fmt.Println(`Run performance benchmarks with threshold checks

Usage:
  voyage perf [options]

Options:`)
		fs.PrintDefaults()
		fmt.Println(`
Examples:
  voyage perf                         # Run with defaults, check thresholds
  voyage perf -o perf.json            # Output to file
  voyage perf -fail=false             # Don't fail on threshold violations
  voyage perf -step-max 10ms          # Custom Step threshold
  voyage perf -n 5000 -q              # More iterations, quiet mode

Thresholds (for 60 FPS):
  Step:      5ms  (leaves 11ms for rendering)
  InitWorld: 100ms (one-time cost)
  Step100:   500ms (5ms average per step)

Exit codes:
  0 - All benchmarks passed thresholds
  1 - One or more benchmarks exceeded thresholds (if -fail=true)`)
	}

	fs.Parse(args)
	return flags
}

// runPerfCommand handles the "perf" subcommand for performance benchmarks with threshold checks.
func runPerfCommand(args []string) {
	flags := parsePerfFlags(args)

	runner := bench.NewRunner(flags.iterations, flags.warmup)
	runner.SetThresholds(bench.Thresholds{
		StepMax:      flags.stepMax,
		InitWorldMax: flags.initMax,
		Step100Max:   flags.step100Max,
		FrameTimeMax: 16 * time.Millisecond,
	})

	if !flags.quiet {
		fmt.Println("Performance Benchmarks with Threshold Checks")
		fmt.Println("=============================================")
		fmt.Printf("Iterations: %d (warmup: %d)\n", flags.iterations, flags.warmup)
		fmt.Printf("Thresholds: Step=%v, Init=%v, Step100=%v\n\n", flags.stepMax, flags.initMax, flags.step100Max)
	}

	report := runner.RunAll()

	if !flags.quiet {
		printPerfResults(report)
	}

	outputPerfReport(report, flags)
}

// printPerfResults prints benchmark results to stdout
func printPerfResults(report bench.PerfReport) {
	for _, r := range report.Results {
		status := "PASS"
		if !r.Passed {
			status = "FAIL"
		}
		fmt.Printf("[%s] %-12s P95=%v (threshold=%v)\n", status, r.Name, r.P95, r.Threshold)
		fmt.Printf("       avg=%v min=%v max=%v p50=%v p99=%v\n", r.Avg, r.Min, r.Max, r.P50, r.P99)
	}
	fmt.Println()

	if report.AllPassed {
		fmt.Println("All benchmarks PASSED threshold checks")
	} else {
		fmt.Println("Some benchmarks FAILED threshold checks")
	}
}

// outputPerfReport handles JSON output and exit code
func outputPerfReport(report bench.PerfReport, flags perfFlags) {
	jsonData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling report: %v\n", err)
		os.Exit(1)
	}

	if flags.outputPath != "" {
		if err := os.WriteFile(flags.outputPath, jsonData, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing output: %v\n", err)
			os.Exit(1)
		}
		if !flags.quiet {
			fmt.Printf("\nReport written to: %s\n", flags.outputPath)
		}
	} else if flags.quiet {
		fmt.Println(string(jsonData))
	}

	if flags.failOnThreshold && !report.AllPassed {
		os.Exit(1)
	}
}
