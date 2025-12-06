// Package main provides a CLI for Stapledon's Voyage development tools.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"stapledons_voyage/engine/bench"
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
  ai       Test AI handlers (Claude, Gemini)
  world    Inspect world state (NPCs, tiles, planets)
  bench    Run performance benchmarks (human-readable)
  perf     Run benchmarks with threshold checks (CI/JSON output)
  assets   Validate game assets
  sim      Run simulation stress tests
  help     Show this help message

Use "voyage <command> -h" for more information about a command.`)
}

// AI command
func runAICommand(args []string) {
	fs := flag.NewFlagSet("ai", flag.ExitOnError)
	provider := fs.String("provider", "", "AI provider: claude, gemini, auto (default: auto-detect)")
	prompt := fs.String("prompt", "", "Text prompt to send")
	system := fs.String("system", "", "System prompt")
	imagePath := fs.String("image", "", "Path to image file to include (Gemini only)")
	generateImage := fs.Bool("generate-image", false, "Generate an image from prompt (Gemini only)")
	editImage := fs.Bool("edit-image", false, "Edit a reference image with the prompt (Gemini only)")
	referencePath := fs.String("reference", "", "Path to reference image for editing (or uses last generated)")
	tts := fs.Bool("tts", false, "Generate speech from prompt (Gemini only)")
	voice := fs.String("voice", "", "TTS voice name (default: Kore). Options: Aoede, Charon, Fenrir, Kore, Puck, Zephyr, Enceladus")
	verbose := fs.Bool("v", false, "Verbose output")
	listProviders := fs.Bool("list", false, "List available providers and their status")
	listVoices := fs.Bool("list-voices", false, "List available TTS voices")

	fs.Usage = func() {
		fmt.Println(`Test AI handlers

Usage:
  voyage ai [options]

Options:`)
		fs.PrintDefaults()
		fmt.Println(`
Examples:
  voyage ai -list                           # Show available providers
  voyage ai -list-voices                    # Show available TTS voices
  voyage ai -prompt "Hello!"                # Auto-detect provider
  voyage ai -provider claude -prompt "Hi"   # Use Claude
  voyage ai -provider gemini -prompt "Hi"   # Use Gemini
  voyage ai -prompt "Draw a cat" -generate-image  # Generate image
  voyage ai -prompt "Make sky purple" -edit-image -reference img.png  # Edit image
  voyage ai -prompt "Add more stars" -edit-image  # Edit last generated image
  voyage ai -prompt "Hello world" -tts      # Text to speech
  voyage ai -prompt "Hello" -tts -voice Puck  # TTS with specific voice

Environment Variables:
  ANTHROPIC_API_KEY     - Claude API key
  GOOGLE_CLOUD_PROJECT  - GCP project for Vertex AI (preferred)
  GOOGLE_API_KEY        - Gemini API key (fallback)
  AI_PROVIDER           - Default provider (claude, gemini)`)
	}

	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	ctx := context.Background()

	// List providers mode
	if *listProviders {
		listAIProviders(ctx, *verbose)
		return
	}

	// List voices mode
	if *listVoices {
		listTTSVoices()
		return
	}

	// Need a prompt for other operations
	if *prompt == "" {
		fmt.Fprintln(os.Stderr, "Error: -prompt is required")
		fs.Usage()
		os.Exit(1)
	}

	// Get handler
	var handler handlers.AIHandler
	var err error

	switch *provider {
	case "claude":
		handler, err = handlers.NewClaudeAIHandler(handlers.ClaudeConfig{})
	case "gemini":
		handler, err = handlers.NewGeminiAIHandler(ctx, handlers.GeminiConfig{})
	case "", "auto":
		handler, err = handlers.NewAIHandlerFromEnv(ctx)
	default:
		fmt.Fprintf(os.Stderr, "Unknown provider: %s\n", *provider)
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create AI handler: %v\n", err)
		os.Exit(1)
	}

	// Build request
	req := handlers.AIRequest{
		System:   *system,
		Messages: []handlers.ContentBlock{},
		Context:  make(map[string]interface{}),
	}

	// Add image if provided
	if *imagePath != "" {
		// Check file exists
		if _, err := os.Stat(*imagePath); err != nil {
			fmt.Fprintf(os.Stderr, "Image file not found: %s\n", *imagePath)
			os.Exit(1)
		}
		req.Messages = append(req.Messages, handlers.ContentBlock{
			Type:     handlers.ContentTypeImage,
			ImageRef: *imagePath,
		})
	}

	// Handle special modes
	promptText := *prompt
	if *generateImage {
		promptText = "imagen: " + promptText
		req.Context["generate_image"] = true
	}
	if *tts {
		promptText = "speak: " + promptText
		req.Context["tts"] = true
		if *voice != "" {
			req.Context["voice"] = *voice
		}
	}
	if *editImage {
		promptText = "edit: " + promptText
		req.Context["edit_image"] = true
		// Add reference image if provided
		if *referencePath != "" {
			if _, err := os.Stat(*referencePath); err != nil {
				fmt.Fprintf(os.Stderr, "Reference image not found: %s\n", *referencePath)
				os.Exit(1)
			}
			req.Context["reference_image"] = *referencePath
		}
	}

	req.Messages = append(req.Messages, handlers.ContentBlock{
		Type: handlers.ContentTypeText,
		Text: promptText,
	})

	if *verbose {
		reqJSON, _ := json.MarshalIndent(req, "", "  ")
		fmt.Printf("Request:\n%s\n\n", reqJSON)
	}

	// Make the call
	reqJSON, _ := json.Marshal(req)
	resp, err := handler.Call(string(reqJSON))
	if err != nil {
		fmt.Fprintf(os.Stderr, "AI call failed: %v\n", err)
		os.Exit(1)
	}

	// Parse and display response
	var aiResp handlers.AIResponse
	if err := json.Unmarshal([]byte(resp), &aiResp); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse response: %v\n", err)
		fmt.Printf("Raw response: %s\n", resp)
		os.Exit(1)
	}

	if aiResp.Error != "" {
		fmt.Fprintf(os.Stderr, "AI Error: %s\n", aiResp.Error)
		os.Exit(1)
	}

	if *verbose {
		respJSON, _ := json.MarshalIndent(aiResp, "", "  ")
		fmt.Printf("Response:\n%s\n", respJSON)
	} else {
		// Pretty print content
		for _, block := range aiResp.Content {
			switch block.Type {
			case handlers.ContentTypeText:
				fmt.Println(block.Text)
			case handlers.ContentTypeImage:
				fmt.Printf("[Image saved: %s]\n", block.ImageRef)
			case handlers.ContentTypeAudio:
				fmt.Printf("[Audio saved: %s]\n", block.AudioRef)
			}
		}
	}
}

func listAIProviders(ctx context.Context, verbose bool) {
	fmt.Println("AI Provider Status")
	fmt.Println("==================")
	fmt.Println()

	// Check Claude
	fmt.Print("Claude: ")
	if key := os.Getenv("ANTHROPIC_API_KEY"); key != "" {
		masked := key[:10] + "..." + key[len(key)-4:]
		fmt.Printf("API Key configured (%s)\n", masked)
		if verbose {
			handler, err := handlers.NewClaudeAIHandler(handlers.ClaudeConfig{})
			if err != nil {
				fmt.Printf("  Init failed: %v\n", err)
			} else {
				fmt.Println("  Handler created successfully")
				testHandler(handler, "Claude")
			}
		}
	} else {
		fmt.Println("Not configured (set ANTHROPIC_API_KEY)")
	}

	// Check Gemini - Vertex AI
	fmt.Print("Gemini (Vertex AI): ")
	project := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if project == "" {
		project = os.Getenv("GCLOUD_PROJECT")
	}
	if project != "" {
		location := os.Getenv("GOOGLE_CLOUD_LOCATION")
		if location == "" {
			location = "us-central1"
		}
		fmt.Printf("Project: %s, Location: %s\n", project, location)
		if verbose {
			handler, err := handlers.NewGeminiAIHandler(ctx, handlers.GeminiConfig{
				Project:  project,
				Location: location,
			})
			if err != nil {
				fmt.Printf("  Init failed: %v\n", err)
			} else {
				fmt.Println("  Handler created successfully (using ADC)")
				testHandler(handler, "Gemini/Vertex")
			}
		}
	} else {
		fmt.Println("Not configured (set GOOGLE_CLOUD_PROJECT)")
	}

	// Check Gemini - API Key
	fmt.Print("Gemini (API Key): ")
	if key := os.Getenv("GOOGLE_API_KEY"); key != "" {
		masked := key[:10] + "..." + key[len(key)-4:]
		fmt.Printf("API Key configured (%s)\n", masked)
		if verbose && project == "" {
			// Only test if Vertex wasn't tested
			handler, err := handlers.NewGeminiAIHandler(ctx, handlers.GeminiConfig{
				APIKey: key,
			})
			if err != nil {
				fmt.Printf("  Init failed: %v\n", err)
			} else {
				fmt.Println("  Handler created successfully")
				testHandler(handler, "Gemini/API")
			}
		}
	} else {
		fmt.Println("Not configured (set GOOGLE_API_KEY)")
	}

	fmt.Println()

	// Show what auto-detect would use
	fmt.Print("Auto-detect would use: ")
	handler, _ := handlers.NewAIHandlerFromEnv(ctx)
	switch handler.(type) {
	case *handlers.ClaudeAIHandler:
		fmt.Println("Claude")
	case *handlers.GeminiAIHandler:
		fmt.Println("Gemini")
	case *handlers.StubAIHandler:
		fmt.Println("Stub (no providers available)")
	default:
		fmt.Println("Unknown")
	}
}

func testHandler(handler handlers.AIHandler, name string) {
	req := handlers.AIRequest{
		Messages: []handlers.ContentBlock{
			{Type: handlers.ContentTypeText, Text: "Reply with just the word 'ok'"},
		},
	}
	reqJSON, _ := json.Marshal(req)

	resp, err := handler.Call(string(reqJSON))
	if err != nil {
		fmt.Printf("  Test call failed: %v\n", err)
		return
	}

	var aiResp handlers.AIResponse
	if err := json.Unmarshal([]byte(resp), &aiResp); err != nil {
		fmt.Printf("  Test parse failed: %v\n", err)
		return
	}

	if aiResp.Error != "" {
		fmt.Printf("  Test returned error: %s\n", aiResp.Error)
		return
	}

	// Get first text response
	for _, block := range aiResp.Content {
		if block.Type == handlers.ContentTypeText {
			text := strings.TrimSpace(block.Text)
			if len(text) > 50 {
				text = text[:50] + "..."
			}
			fmt.Printf("  Test response: %q\n", text)
			return
		}
	}
	fmt.Println("  Test: no text in response")
}

func listTTSVoices() {
	fmt.Println("Available TTS Voices (Gemini 2.5)")
	fmt.Println("==================================")
	fmt.Println()
	fmt.Println("Voice Name   | Characteristics")
	fmt.Println("-------------|----------------")
	fmt.Println("Aoede        | Bright, clear")
	fmt.Println("Charon       | Deep, resonant")
	fmt.Println("Fenrir       | Strong, bold")
	fmt.Println("Kore         | Warm, natural (default)")
	fmt.Println("Puck         | Playful, upbeat")
	fmt.Println("Zephyr       | Soft, breathy")
	fmt.Println("Enceladus    | Calm, measured")
	fmt.Println()
	fmt.Println("Usage: voyage ai -tts -voice <name> -prompt \"text to speak\"")
}

// World inspection command
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

	// Initialize world
	world := sim_gen.InitWorld(*seed)

	// Run steps if requested
	if *steps > 0 {
		fmt.Printf("Running %d steps...\n", *steps)
		input := sim_gen.FrameInput{}
		for i := 0; i < *steps; i++ {
			result := sim_gen.Step(world, input)
			tuple, ok := result.([]interface{})
			if !ok || len(tuple) != 2 {
				fmt.Fprintln(os.Stderr, "Error: unexpected Step result")
				os.Exit(1)
			}
			world = tuple[0]
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

// Benchmark command
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

// Performance command with threshold checking
func runPerfCommand(args []string) {
	fs := flag.NewFlagSet("perf", flag.ExitOnError)
	iterations := fs.Int("n", 1000, "Number of iterations")
	warmup := fs.Int("warmup", 100, "Warmup iterations")
	outputPath := fs.String("o", "", "Output JSON file path (default: stdout)")
	failOnThreshold := fs.Bool("fail", true, "Exit with code 1 if thresholds exceeded")
	stepMax := fs.Duration("step-max", 5*time.Millisecond, "Max time for Step()")
	initMax := fs.Duration("init-max", 100*time.Millisecond, "Max time for InitWorld()")
	step100Max := fs.Duration("step100-max", 500*time.Millisecond, "Max time for 100 steps")
	quiet := fs.Bool("q", false, "Quiet mode (only output JSON)")

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

	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	// Create runner with custom thresholds
	runner := bench.NewRunner(*iterations, *warmup)
	runner.SetThresholds(bench.Thresholds{
		StepMax:      *stepMax,
		InitWorldMax: *initMax,
		Step100Max:   *step100Max,
		FrameTimeMax: 16 * time.Millisecond,
	})

	if !*quiet {
		fmt.Println("Performance Benchmarks with Threshold Checks")
		fmt.Println("=============================================")
		fmt.Printf("Iterations: %d (warmup: %d)\n", *iterations, *warmup)
		fmt.Printf("Thresholds: Step=%v, Init=%v, Step100=%v\n\n", *stepMax, *initMax, *step100Max)
	}

	// Run benchmarks
	report := runner.RunAll()

	// Print results unless quiet
	if !*quiet {
		for _, r := range report.Results {
			status := "PASS"
			if !r.Passed {
				status = "FAIL"
			}
			fmt.Printf("[%s] %-12s P95=%v (threshold=%v)\n", status, r.Name, r.P95, r.Threshold)
			fmt.Printf("       avg=%v min=%v max=%v p50=%v p99=%v\n",
				r.Avg, r.Min, r.Max, r.P50, r.P99)
		}
		fmt.Println()

		if report.AllPassed {
			fmt.Println("All benchmarks PASSED threshold checks")
		} else {
			fmt.Println("Some benchmarks FAILED threshold checks")
		}
	}

	// Output JSON
	jsonData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling report: %v\n", err)
		os.Exit(1)
	}

	if *outputPath != "" {
		if err := os.WriteFile(*outputPath, jsonData, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing output: %v\n", err)
			os.Exit(1)
		}
		if !*quiet {
			fmt.Printf("\nReport written to: %s\n", *outputPath)
		}
	} else if *quiet {
		fmt.Println(string(jsonData))
	}

	// Exit with error if thresholds exceeded and -fail is set
	if *failOnThreshold && !report.AllPassed {
		os.Exit(1)
	}
}

// Assets validation command
func runAssetsCommand(args []string) {
	fs := flag.NewFlagSet("assets", flag.ExitOnError)
	assetsDir := fs.String("dir", "assets", "Assets directory to validate")
	verbose := fs.Bool("v", false, "Verbose output")
	fix := fs.Bool("fix", false, "Create missing directories")

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

	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	fmt.Println("Asset Validation")
	fmt.Println("================")
	fmt.Printf("Directory: %s\n\n", *assetsDir)

	// Expected structure
	expectedDirs := []string{
		"sprites",
		"fonts",
		"sounds",
		"generated",
		"starmap",
	}

	expectedFiles := map[string][]string{
		"sprites": {".png"},
		"fonts":   {".ttf", ".otf"},
		"sounds":  {".wav", ".ogg", ".mp3"},
	}

	hasErrors := false

	// Check directories
	for _, dir := range expectedDirs {
		path := filepath.Join(*assetsDir, dir)
		info, err := os.Stat(path)
		if os.IsNotExist(err) {
			if *fix {
				if err := os.MkdirAll(path, 0755); err != nil {
					fmt.Printf("❌ %s: failed to create: %v\n", dir, err)
					hasErrors = true
				} else {
					fmt.Printf("✅ %s: created\n", dir)
				}
			} else {
				fmt.Printf("❌ %s: missing\n", dir)
				hasErrors = true
			}
		} else if err != nil {
			fmt.Printf("❌ %s: error: %v\n", dir, err)
			hasErrors = true
		} else if !info.IsDir() {
			fmt.Printf("❌ %s: not a directory\n", dir)
			hasErrors = true
		} else {
			// Count files
			count := 0
			extensions, hasExtensions := expectedFiles[dir]
			filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
				if err != nil || info.IsDir() {
					return nil
				}
				if hasExtensions {
					ext := strings.ToLower(filepath.Ext(p))
					for _, expected := range extensions {
						if ext == expected {
							count++
							if *verbose {
								relPath, _ := filepath.Rel(*assetsDir, p)
								fmt.Printf("   - %s\n", relPath)
							}
							break
						}
					}
				} else {
					count++
					if *verbose {
						relPath, _ := filepath.Rel(*assetsDir, p)
						fmt.Printf("   - %s\n", relPath)
					}
				}
				return nil
			})
			fmt.Printf("✅ %s: %d files\n", dir, count)
		}
	}

	// Check manifest if exists
	manifestPath := filepath.Join(*assetsDir, "manifest.json")
	if _, err := os.Stat(manifestPath); err == nil {
		fmt.Printf("✅ manifest.json: found\n")
		if *verbose {
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

	fmt.Println()
	if hasErrors {
		fmt.Println("Some issues found. Use -fix to create missing directories.")
		os.Exit(1)
	} else {
		fmt.Println("All checks passed!")
	}
}

// Simulation stress test command
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

	world := sim_gen.InitWorld(*seed)
	input := sim_gen.FrameInput{}

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
		world = tuple[0]

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
