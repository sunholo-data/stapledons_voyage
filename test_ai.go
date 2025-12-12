//go:build ignore

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"stapledons_voyage/engine/handlers"
)

func main() {
	ctx := context.Background()

	fmt.Println("=== Gemini AI Test Suite ===")
	fmt.Println()

	// Create Gemini handler with defaults
	handler, err := handlers.NewGeminiAIHandler(ctx, handlers.GeminiConfig{})
	if err != nil {
		fmt.Printf("Failed to create handler: %v\n", err)
		os.Exit(1)
	}

	// Test 1: Text
	fmt.Println("1. Testing TEXT...")
	testText(handler)

	// Test 2: Image generation
	fmt.Println("\n2. Testing IMAGE GENERATION...")
	testImageGen(handler)

	// Test 3: TTS
	fmt.Println("\n3. Testing TTS...")
	testTTS(handler)

	fmt.Println("\n=== All tests completed ===")
}

func testText(handler handlers.AIHandler) {
	req := handlers.AIRequest{
		Messages: []handlers.ContentBlock{
			{Type: handlers.ContentTypeText, Text: "Reply with just: 'Text works!'"},
		},
	}
	call(handler, req)
}

func testImageGen(handler handlers.AIHandler) {
	req := handlers.AIRequest{
		Messages: []handlers.ContentBlock{
			{Type: handlers.ContentTypeText, Text: "imagen: A simple red square on white background"},
		},
		Context: map[string]interface{}{"generate_image": true},
	}
	call(handler, req)
}

func testTTS(handler handlers.AIHandler) {
	req := handlers.AIRequest{
		Messages: []handlers.ContentBlock{
			{Type: handlers.ContentTypeText, Text: "speak: Hello, text to speech is working!"},
		},
		Context: map[string]interface{}{"tts": true, "voice": "Kore"},
	}
	call(handler, req)
}

func call(handler handlers.AIHandler, req handlers.AIRequest) {
	reqJSON, _ := json.Marshal(req)
	resp, err := handler.Call(string(reqJSON))
	if err != nil {
		fmt.Printf("   ERROR: %v\n", err)
		return
	}

	var aiResp handlers.AIResponse
	if err := json.Unmarshal([]byte(resp), &aiResp); err != nil {
		fmt.Printf("   Parse error: %v\n", err)
		return
	}

	if aiResp.Error != "" {
		fmt.Printf("   AI Error: %s\n", aiResp.Error)
		return
	}

	for _, block := range aiResp.Content {
		switch block.Type {
		case handlers.ContentTypeText:
			text := block.Text
			if len(text) > 100 {
				text = text[:100] + "..."
			}
			fmt.Printf("   OK: %s\n", text)
		case handlers.ContentTypeImage:
			fmt.Printf("   OK: Image saved to %s\n", block.ImageRef)
		case handlers.ContentTypeAudio:
			fmt.Printf("   OK: Audio saved to %s\n", block.AudioRef)
		}
	}
}
