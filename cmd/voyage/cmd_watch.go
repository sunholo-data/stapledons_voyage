// Package main provides CLI commands for watching and auto-rebuilding.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func runWatchCommand(args []string) {
	// Parse flags
	runDemo := false
	runTest := false
	demoName := ""

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-h", "--help":
			printWatchHelp()
			return
		case "--run":
			runDemo = true
			if i+1 < len(args) && args[i+1][0] != '-' {
				demoName = args[i+1]
				i++
			}
		case "--test":
			runTest = true
		}
	}

	fmt.Println("Watching sim/*.ail for changes...")
	fmt.Println("Press Ctrl+C to stop")
	fmt.Println()

	// Get initial state
	lastMod := getLatestModTime("sim")

	for {
		time.Sleep(500 * time.Millisecond)

		currentMod := getLatestModTime("sim")
		if currentMod.After(lastMod) {
			lastMod = currentMod
			fmt.Printf("\n[%s] Change detected, rebuilding...\n", time.Now().Format("15:04:05"))

			// Run make sim
			cmd := exec.Command("make", "sim")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			err := cmd.Run()

			if err != nil {
				fmt.Println("Build failed!")
				continue
			}

			fmt.Println("Build successful!")

			// Run tests if requested
			if runTest {
				fmt.Println("\nRunning tests...")
				testCmd := exec.Command("ailang", "test", "sim/")
				testCmd.Stdout = os.Stdout
				testCmd.Stderr = os.Stderr
				testCmd.Run()
			}

			// Run demo if requested
			if runDemo && demoName != "" {
				fmt.Printf("\nStarting %s...\n", demoName)
				demoCmd := exec.Command("go", "run", "./cmd/"+demoName)
				demoCmd.Stdout = os.Stdout
				demoCmd.Stderr = os.Stderr
				demoCmd.Stdin = os.Stdin
				demoCmd.Run()
			}
		}
	}
}

func getLatestModTime(dir string) time.Time {
	var latest time.Time

	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && filepath.Ext(path) == ".ail" {
			if info.ModTime().After(latest) {
				latest = info.ModTime()
			}
		}
		return nil
	})

	return latest
}

func printWatchHelp() {
	fmt.Println(`Watch sim/*.ail for changes and auto-rebuild

Usage:
  voyage watch [flags]

Flags:
  --run [demo]    Run specified demo after successful build
  --test          Run ailang test after successful build
  -h, --help      Show this help

Examples:
  voyage watch                    # Watch and rebuild only
  voyage watch --test             # Watch, rebuild, and test
  voyage watch --run bridge       # Watch, rebuild, and run demo-game-bridge`)
}
