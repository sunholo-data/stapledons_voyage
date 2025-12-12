// Package main provides CLI commands for asset manifest validation.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Manifest structures for parsing
type spriteManifest struct {
	Sprites map[string]spriteEntry `json:"sprites"`
}

type spriteEntry struct {
	File string `json:"file"`
}

type soundManifest struct {
	Sounds map[string]soundEntry `json:"sounds"`
	BGM    map[string]soundEntry `json:"bgm"`
}

type soundEntry struct {
	File string `json:"file"`
}

type fontManifest struct {
	Fonts map[string]fontEntry `json:"fonts"`
}

type fontEntry struct {
	File string `json:"file"`
}

func runManifestCommand(args []string) {
	verbose := false
	specificManifest := ""

	// Parse flags
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-h", "--help":
			printManifestHelp()
			return
		case "--verbose", "-v":
			verbose = true
		default:
			if args[i][0] != '-' && specificManifest == "" {
				specificManifest = args[i]
			}
		}
	}

	manifests := []string{"sprites", "sounds", "fonts"}
	if specificManifest != "" {
		manifests = []string{specificManifest}
	}

	totalAssets := 0
	totalMissing := 0
	totalFound := 0

	for _, m := range manifests {
		found, missing, total := checkManifestFiles(m, verbose)
		totalFound += found
		totalMissing += missing
		totalAssets += total
	}

	fmt.Println()
	if totalMissing == 0 {
		fmt.Printf("✓ All %d assets valid\n", totalAssets)
	} else {
		fmt.Printf("✗ %d/%d assets valid, %d missing\n", totalFound, totalAssets, totalMissing)
		os.Exit(1)
	}
}

func checkManifestFiles(name string, verbose bool) (found, missing, total int) {
	var manifestPath string
	var baseDir string

	switch name {
	case "sprites":
		manifestPath = "assets/sprites/manifest.json"
		baseDir = "assets/sprites"
	case "sounds":
		manifestPath = "assets/sounds/manifest.json"
		baseDir = "assets/sounds"
	case "fonts":
		manifestPath = "assets/fonts/manifest.json"
		baseDir = "assets/fonts"
	default:
		fmt.Printf("Unknown manifest: %s\n", name)
		return 0, 0, 0
	}

	data, err := os.ReadFile(manifestPath)
	if err != nil {
		fmt.Printf("✗ %s: failed to read manifest: %v\n", name, err)
		return 0, 1, 1
	}

	fmt.Printf("\n[%s]\n", name)

	switch name {
	case "sprites":
		var m spriteManifest
		if err := json.Unmarshal(data, &m); err != nil {
			fmt.Printf("  ✗ Invalid JSON: %v\n", err)
			return 0, 1, 1
		}
		for id, entry := range m.Sprites {
			total++
			path := filepath.Join(baseDir, entry.File)
			if _, err := os.Stat(path); err != nil {
				fmt.Printf("  ✗ [%s] missing: %s\n", id, entry.File)
				missing++
			} else {
				if verbose {
					fmt.Printf("  ✓ [%s] %s\n", id, entry.File)
				}
				found++
			}
		}

	case "sounds":
		var m soundManifest
		if err := json.Unmarshal(data, &m); err != nil {
			fmt.Printf("  ✗ Invalid JSON: %v\n", err)
			return 0, 1, 1
		}
		for id, entry := range m.Sounds {
			total++
			path := filepath.Join(baseDir, entry.File)
			if _, err := os.Stat(path); err != nil {
				fmt.Printf("  ✗ [%s] missing: %s\n", id, entry.File)
				missing++
			} else {
				if verbose {
					fmt.Printf("  ✓ [%s] %s\n", id, entry.File)
				}
				found++
			}
		}
		for id, entry := range m.BGM {
			total++
			path := filepath.Join(baseDir, entry.File)
			if _, err := os.Stat(path); err != nil {
				fmt.Printf("  ✗ [bgm:%s] missing: %s\n", id, entry.File)
				missing++
			} else {
				if verbose {
					fmt.Printf("  ✓ [bgm:%s] %s\n", id, entry.File)
				}
				found++
			}
		}

	case "fonts":
		var m fontManifest
		if err := json.Unmarshal(data, &m); err != nil {
			fmt.Printf("  ✗ Invalid JSON: %v\n", err)
			return 0, 1, 1
		}
		// Track unique files to avoid double-counting shared fonts
		seen := make(map[string]bool)
		for id, entry := range m.Fonts {
			if seen[entry.File] {
				continue
			}
			seen[entry.File] = true
			total++
			path := filepath.Join(baseDir, entry.File)
			if _, err := os.Stat(path); err != nil {
				fmt.Printf("  ✗ [%s] missing: %s\n", id, entry.File)
				missing++
			} else {
				if verbose {
					fmt.Printf("  ✓ [%s] %s\n", id, entry.File)
				}
				found++
			}
		}
	}

	if missing == 0 {
		fmt.Printf("  ✓ %d assets OK\n", found)
	} else {
		fmt.Printf("  %d found, %d missing\n", found, missing)
	}

	return found, missing, total
}

func printManifestHelp() {
	fmt.Println(`Validate asset manifests

Usage:
  voyage manifest [type] [flags]

Types:
  sprites    Check sprite assets
  sounds     Check sound assets
  fonts      Check font assets
  (none)     Check all manifests

Flags:
  --verbose, -v    Show all assets, not just missing
  -h, --help       Show this help

Examples:
  voyage manifest              # Validate all manifests
  voyage manifest sprites      # Check only sprites
  voyage manifest -v           # Verbose output`)
}
