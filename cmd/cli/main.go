// Package main provides a CLI for Stapledon's Voyage development tools.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"stapledons_voyage/engine/handlers"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "ai":
		runAICommand(args)
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
