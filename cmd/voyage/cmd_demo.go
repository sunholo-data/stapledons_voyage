// Package main provides CLI commands for running demos.
package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
)

func runDemoCommand(args []string) {
	// Find all demo directories
	demos, err := findDemos()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error finding demos: %v\n", err)
		os.Exit(1)
	}

	if len(demos) == 0 {
		fmt.Println("No demos found in cmd/demo-*")
		return
	}

	// If a demo name is provided, run it directly
	if len(args) > 0 {
		name := args[0]
		extraArgs := args[1:]

		// Allow partial matching
		var matched string
		for _, d := range demos {
			if d == name || strings.HasSuffix(d, name) || strings.Contains(d, name) {
				matched = d
				break
			}
		}

		if matched == "" {
			fmt.Fprintf(os.Stderr, "Demo not found: %s\n\n", name)
			fmt.Println("Available demos:")
			for _, d := range demos {
				fmt.Printf("  %s\n", d)
			}
			os.Exit(1)
		}

		runDemo(matched, extraArgs)
		return
	}

	// Interactive selection
	fmt.Println("Available demos:")
	fmt.Println()

	// Group demos by category
	engineDemos := []string{}
	gameDemos := []string{}
	otherDemos := []string{}

	for _, d := range demos {
		switch {
		case strings.HasPrefix(d, "demo-engine-"):
			engineDemos = append(engineDemos, d)
		case strings.HasPrefix(d, "demo-game-"):
			gameDemos = append(gameDemos, d)
		default:
			otherDemos = append(otherDemos, d)
		}
	}

	idx := 1
	demoMap := make(map[int]string)

	if len(engineDemos) > 0 {
		fmt.Println("  Engine demos (pure Go/Ebiten):")
		for _, d := range engineDemos {
			shortName := strings.TrimPrefix(d, "demo-engine-")
			fmt.Printf("    [%d] %s\n", idx, shortName)
			demoMap[idx] = d
			idx++
		}
		fmt.Println()
	}

	if len(gameDemos) > 0 {
		fmt.Println("  Game demos (AILANG + engine):")
		for _, d := range gameDemos {
			shortName := strings.TrimPrefix(d, "demo-game-")
			fmt.Printf("    [%d] %s\n", idx, shortName)
			demoMap[idx] = d
			idx++
		}
		fmt.Println()
	}

	if len(otherDemos) > 0 {
		fmt.Println("  Other demos:")
		for _, d := range otherDemos {
			shortName := strings.TrimPrefix(d, "demo-")
			fmt.Printf("    [%d] %s\n", idx, shortName)
			demoMap[idx] = d
			idx++
		}
		fmt.Println()
	}

	fmt.Print("Select demo (number or name, q to quit): ")

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return
	}

	input = strings.TrimSpace(input)
	if input == "q" || input == "quit" || input == "" {
		return
	}

	// Try parsing as number first
	if num, err := strconv.Atoi(input); err == nil {
		if demo, ok := demoMap[num]; ok {
			runDemo(demo, nil)
			return
		}
		fmt.Fprintf(os.Stderr, "Invalid selection: %d\n", num)
		os.Exit(1)
	}

	// Try matching by name
	for _, d := range demos {
		shortName := strings.TrimPrefix(strings.TrimPrefix(strings.TrimPrefix(d, "demo-engine-"), "demo-game-"), "demo-")
		if shortName == input || d == input || strings.Contains(d, input) {
			runDemo(d, nil)
			return
		}
	}

	fmt.Fprintf(os.Stderr, "Demo not found: %s\n", input)
	os.Exit(1)
}

func findDemos() ([]string, error) {
	// Get the project root (assuming CLI is run from project root or cmd/cli)
	entries, err := os.ReadDir("cmd")
	if err != nil {
		// Try from project root
		entries, err = os.ReadDir(".")
		if err != nil {
			return nil, err
		}
	}

	var demos []string
	for _, e := range entries {
		if e.IsDir() && strings.HasPrefix(e.Name(), "demo-") {
			demos = append(demos, e.Name())
		}
	}

	sort.Strings(demos)
	return demos, nil
}

func runDemo(name string, extraArgs []string) {
	fmt.Printf("Running %s...\n\n", name)

	// Build the demo path - must use ./ prefix for go run to treat as local path
	demoPath := "./cmd/" + name

	// Build arguments for go run
	args := []string{"run", demoPath}
	args = append(args, extraArgs...)

	cmd := exec.Command("go", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		fmt.Fprintf(os.Stderr, "Error running demo: %v\n", err)
		os.Exit(1)
	}
}
