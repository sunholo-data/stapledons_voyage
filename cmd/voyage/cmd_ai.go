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

// aiFlags holds all parsed AI command flags
type aiFlags struct {
	provider      string
	prompt        string
	system        string
	imagePath     string
	generateImage bool
	editImage     bool
	referencePath string
	tts           bool
	voice         string
	verbose       bool
	listProviders bool
	listVoices    bool
}

// parseAIFlags parses and returns AI command flags
func parseAIFlags(args []string) (aiFlags, *flag.FlagSet) {
	fs := flag.NewFlagSet("ai", flag.ExitOnError)
	flags := aiFlags{}

	fs.StringVar(&flags.provider, "provider", "", "AI provider: claude, gemini, auto (default: auto-detect)")
	fs.StringVar(&flags.prompt, "prompt", "", "Text prompt to send")
	fs.StringVar(&flags.system, "system", "", "System prompt")
	fs.StringVar(&flags.imagePath, "image", "", "Path to image file to include (Gemini only)")
	fs.BoolVar(&flags.generateImage, "generate-image", false, "Generate an image from prompt (Gemini only)")
	fs.BoolVar(&flags.editImage, "edit-image", false, "Edit a reference image with the prompt (Gemini only)")
	fs.StringVar(&flags.referencePath, "reference", "", "Path to reference image for editing (or uses last generated)")
	fs.BoolVar(&flags.tts, "tts", false, "Generate speech from prompt (Gemini only)")
	fs.StringVar(&flags.voice, "voice", "", "TTS voice name (default: Kore). Options: Aoede, Charon, Fenrir, Kore, Puck, Zephyr, Enceladus")
	fs.BoolVar(&flags.verbose, "v", false, "Verbose output")
	fs.BoolVar(&flags.listProviders, "list", false, "List available providers and their status")
	fs.BoolVar(&flags.listVoices, "list-voices", false, "List available TTS voices")

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

	fs.Parse(args)
	return flags, fs
}

// createAIHandler creates an AI handler based on provider name
func createAIHandler(ctx context.Context, provider string) (handlers.AIHandler, error) {
	switch provider {
	case "claude":
		return handlers.NewClaudeAIHandler(handlers.ClaudeConfig{})
	case "gemini":
		return handlers.NewGeminiAIHandler(ctx, handlers.GeminiConfig{})
	case "", "auto":
		return handlers.NewAIHandlerFromEnv(ctx)
	default:
		return nil, fmt.Errorf("unknown provider: %s", provider)
	}
}

// buildAIRequest constructs an AIRequest from flags
func buildAIRequest(flags aiFlags) (handlers.AIRequest, error) {
	req := handlers.AIRequest{
		System:   flags.system,
		Messages: []handlers.ContentBlock{},
		Context:  make(map[string]interface{}),
	}

	// Add image if provided
	if flags.imagePath != "" {
		if _, err := os.Stat(flags.imagePath); err != nil {
			return req, fmt.Errorf("image file not found: %s", flags.imagePath)
		}
		req.Messages = append(req.Messages, handlers.ContentBlock{
			Type:     handlers.ContentTypeImage,
			ImageRef: flags.imagePath,
		})
	}

	// Handle special modes
	promptText := flags.prompt
	if flags.generateImage {
		promptText = "imagen: " + promptText
		req.Context["generate_image"] = true
	}
	if flags.tts {
		promptText = "speak: " + promptText
		req.Context["tts"] = true
		if flags.voice != "" {
			req.Context["voice"] = flags.voice
		}
	}
	if flags.editImage {
		promptText = "edit: " + promptText
		req.Context["edit_image"] = true
		if flags.referencePath != "" {
			if _, err := os.Stat(flags.referencePath); err != nil {
				return req, fmt.Errorf("reference image not found: %s", flags.referencePath)
			}
			req.Context["reference_image"] = flags.referencePath
		}
	}

	req.Messages = append(req.Messages, handlers.ContentBlock{
		Type: handlers.ContentTypeText,
		Text: promptText,
	})

	return req, nil
}

// displayAIResponse prints the AI response in human-readable format
func displayAIResponse(aiResp handlers.AIResponse, verbose bool) {
	if verbose {
		respJSON, _ := json.MarshalIndent(aiResp, "", "  ")
		fmt.Printf("Response:\n%s\n", respJSON)
		return
	}

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

// runAICommand handles the "ai" subcommand for testing AI handlers.
func runAICommand(args []string) {
	flags, fs := parseAIFlags(args)
	ctx := context.Background()

	// List providers mode
	if flags.listProviders {
		listAIProviders(ctx, flags.verbose)
		return
	}

	// List voices mode
	if flags.listVoices {
		listTTSVoices()
		return
	}

	// Need a prompt for other operations
	if flags.prompt == "" {
		fmt.Fprintln(os.Stderr, "Error: -prompt is required")
		fs.Usage()
		os.Exit(1)
	}

	// Get handler
	handler, err := createAIHandler(ctx, flags.provider)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create AI handler: %v\n", err)
		os.Exit(1)
	}

	// Build request
	req, err := buildAIRequest(flags)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if flags.verbose {
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

	// Parse response
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

	displayAIResponse(aiResp, flags.verbose)
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
	fmt.Println("Available TTS Voices (Gemini 2.5) - 30 voices")
	fmt.Println("=============================================")
	fmt.Println()
	fmt.Println("Voice Name      | Characteristic")
	fmt.Println("----------------|----------------")
	fmt.Println("Achernar        | Soft")
	fmt.Println("Achird          | Friendly")
	fmt.Println("Algenib         | Gravelly")
	fmt.Println("Algieba         | Smooth")
	fmt.Println("Alnilam         | Firm")
	fmt.Println("Aoede           | Breezy")
	fmt.Println("Autonoe         | Bright")
	fmt.Println("Callirrhoe      | Easy-going")
	fmt.Println("Charon          | Informative")
	fmt.Println("Despina         | Smooth")
	fmt.Println("Enceladus       | Breathy")
	fmt.Println("Erinome         | Clear")
	fmt.Println("Fenrir          | Excitable")
	fmt.Println("Gacrux          | Mature")
	fmt.Println("Iapetus         | Clear")
	fmt.Println("Kore            | Firm (default)")
	fmt.Println("Laomedeia       | Upbeat")
	fmt.Println("Leda            | Youthful")
	fmt.Println("Orus            | Firm")
	fmt.Println("Puck            | Upbeat")
	fmt.Println("Pulcherrima     | Forward")
	fmt.Println("Rasalgethi      | Informative")
	fmt.Println("Sadachbia       | Lively")
	fmt.Println("Sadaltager      | Knowledgeable")
	fmt.Println("Schedar         | Even")
	fmt.Println("Sulafat         | Warm")
	fmt.Println("Umbriel         | Easy-going")
	fmt.Println("Vindemiatrix    | Gentle")
	fmt.Println("Zephyr          | Bright")
	fmt.Println("Zubenelgenubi   | Casual")
	fmt.Println()
	fmt.Println("Usage: voyage ai -tts -voice <name> -prompt \"text to speak\"")
}
