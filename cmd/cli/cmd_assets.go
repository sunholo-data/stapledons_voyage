// Package main provides a CLI for Stapledon's Voyage development tools.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// assetsFlags holds all parsed assets command flags
type assetsFlags struct {
	dir     string
	verbose bool
	fix     bool
}

// parseAssetsFlags parses and returns assets command flags
func parseAssetsFlags(args []string) assetsFlags {
	fs := flag.NewFlagSet("assets", flag.ExitOnError)
	flags := assetsFlags{}

	fs.StringVar(&flags.dir, "dir", "assets", "Assets directory to validate")
	fs.BoolVar(&flags.verbose, "v", false, "Verbose output")
	fs.BoolVar(&flags.fix, "fix", false, "Create missing directories")

	fs.Usage = func() {
		fmt.Println(`Validate game assets

Usage:
  voyage assets [options]

Options:`)
		fs.PrintDefaults()
		fmt.Println(`
Examples:
  voyage assets                   # Validate assets/
  voyage assets -dir ./my-assets  # Custom directory
  voyage assets -v                # Verbose output
  voyage assets -fix              # Create missing directories`)
	}

	fs.Parse(args)
	return flags
}

// countFilesInDir counts files with matching extensions in a directory
func countFilesInDir(path, baseDir string, extensions []string, verbose bool) int {
	count := 0
	filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if len(extensions) > 0 {
			ext := strings.ToLower(filepath.Ext(p))
			for _, expected := range extensions {
				if ext == expected {
					count++
					if verbose {
						relPath, _ := filepath.Rel(baseDir, p)
						fmt.Printf("   - %s\n", relPath)
					}
					break
				}
			}
		} else {
			count++
			if verbose {
				relPath, _ := filepath.Rel(baseDir, p)
				fmt.Printf("   - %s\n", relPath)
			}
		}
		return nil
	})
	return count
}

// validateAssetDir validates a single asset directory
func validateAssetDir(dir, baseDir string, extensions []string, flags assetsFlags) bool {
	path := filepath.Join(baseDir, dir)
	info, err := os.Stat(path)

	if os.IsNotExist(err) {
		if flags.fix {
			if err := os.MkdirAll(path, 0755); err != nil {
				fmt.Printf("❌ %s: failed to create: %v\n", dir, err)
				return false
			}
			fmt.Printf("✅ %s: created\n", dir)
			return true
		}
		fmt.Printf("❌ %s: missing\n", dir)
		return false
	}
	if err != nil {
		fmt.Printf("❌ %s: error: %v\n", dir, err)
		return false
	}
	if !info.IsDir() {
		fmt.Printf("❌ %s: not a directory\n", dir)
		return false
	}

	count := countFilesInDir(path, baseDir, extensions, flags.verbose)
	fmt.Printf("✅ %s: %d files\n", dir, count)
	return true
}

// validateManifest checks if manifest.json exists and reports its contents
func validateManifest(assetsDir string, verbose bool) {
	manifestPath := filepath.Join(assetsDir, "manifest.json")
	if _, err := os.Stat(manifestPath); err == nil {
		fmt.Printf("✅ manifest.json: found\n")
		if verbose {
			data, err := os.ReadFile(manifestPath)
			if err == nil {
				var manifest map[string]interface{}
				if json.Unmarshal(data, &manifest) == nil {
					fmt.Printf("   Entries: %d\n", len(manifest))
				}
			}
		}
	} else {
		fmt.Printf("⚠️  manifest.json: not found (optional)\n")
	}
}

// runAssetsCommand handles the "assets" subcommand for asset validation.
func runAssetsCommand(args []string) {
	flags := parseAssetsFlags(args)

	fmt.Println("Asset Validation")
	fmt.Println("================")
	fmt.Printf("Directory: %s\n\n", flags.dir)

	expectedDirs := []string{"sprites", "fonts", "sounds", "generated", "starmap"}
	expectedFiles := map[string][]string{
		"sprites": {".png"},
		"fonts":   {".ttf", ".otf"},
		"sounds":  {".wav", ".ogg", ".mp3"},
	}

	hasErrors := false
	for _, dir := range expectedDirs {
		if !validateAssetDir(dir, flags.dir, expectedFiles[dir], flags) {
			hasErrors = true
		}
	}

	validateManifest(flags.dir, flags.verbose)

	fmt.Println()
	if hasErrors {
		fmt.Println("Some issues found. Use -fix to create missing directories.")
		os.Exit(1)
	}
	fmt.Println("All checks passed!")
}
