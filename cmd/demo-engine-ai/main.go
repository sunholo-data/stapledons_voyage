// cmd/demo-engine-ai/main.go
// DESCRIPTION: Interactive CLI demo of Gemini AI capabilities (text, image, TTS).
//
// Usage:
//   go run ./cmd/demo-engine-ai
//   go build -o bin/demo-engine-ai ./cmd/demo-engine-ai && bin/demo-engine-ai
//
// Commands:
//   text  - Generate text response
//   image - Generate image
//   tts   - Generate speech (TTS)
//   play  - Play last generated audio
//   voice - Change TTS voice
//   quit  - Exit
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"stapledons_voyage/engine/handlers"
)

// All 30 available TTS voices
var voices = []string{
	"Achernar", "Achird", "Algenib", "Algieba", "Alnilam",
	"Aoede", "Autonoe", "Callirrhoe", "Charon", "Despina",
	"Enceladus", "Erinome", "Fenrir", "Gacrux", "Iapetus",
	"Kore", "Laomedeia", "Leda", "Orus", "Puck",
	"Pulcherrima", "Rasalgethi", "Sadachbia", "Sadaltager", "Schedar",
	"Sulafat", "Umbriel", "Vindemiatrix", "Zephyr", "Zubenelgenubi",
}

func main() {
	fmt.Println("=== Gemini AI Demo ===")
	fmt.Println()

	ctx := context.Background()
	handler, err := handlers.NewGeminiAIHandler(ctx, handlers.GeminiConfig{})
	if err != nil {
		fmt.Printf("Failed to create AI handler: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("AI handler ready!")
	fmt.Println()
	fmt.Println("Commands: text, image, tts, play, voice, quit")
	fmt.Println()

	currentVoice := "Kore"
	lastAudioPath := ""

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		parts := strings.Fields(input)
		if len(parts) == 0 {
			continue
		}

		cmd := strings.ToLower(parts[0])

		switch cmd {
		case "quit", "exit", "q":
			fmt.Println("Goodbye!")
			return

		case "text", "t":
			prompt := "In one sentence, describe a mysterious alien artifact."
			if len(parts) > 1 {
				prompt = strings.Join(parts[1:], " ")
			}
			testText(handler, prompt)

		case "image", "i":
			prompt := "A glowing alien artifact floating in space, sci-fi style"
			if len(parts) > 1 {
				prompt = strings.Join(parts[1:], " ")
			}
			testImage(handler, prompt)

		case "tts", "speak", "s":
			text := "Greetings, traveler. The archive awaits your questions."
			if len(parts) > 1 {
				text = strings.Join(parts[1:], " ")
			}
			path := testTTS(handler, text, currentVoice)
			if path != "" {
				lastAudioPath = path
				playAudio(path)
			}

		case "play", "p":
			if lastAudioPath != "" {
				playAudio(lastAudioPath)
			} else {
				fmt.Println("No audio to play. Generate TTS first.")
			}

		case "voice", "v":
			if len(parts) > 1 {
				newVoice := parts[1]
				found := false
				for _, v := range voices {
					if strings.EqualFold(v, newVoice) {
						currentVoice = v
						found = true
						break
					}
				}
				if found {
					fmt.Printf("Voice set to: %s\n", currentVoice)
				} else {
					fmt.Printf("Unknown voice. Available: %v\n", voices)
				}
			} else {
				fmt.Printf("Current voice: %s\n", currentVoice)
				fmt.Printf("Available voices: %v\n", voices)
				fmt.Println("Usage: voice <name>")
			}

		case "help", "h", "?":
			fmt.Println("Commands:")
			fmt.Println("  text [prompt]  - Generate text response")
			fmt.Println("  image [prompt] - Generate image")
			fmt.Println("  tts [text]     - Generate speech (TTS)")
			fmt.Println("  play           - Play last generated audio")
			fmt.Println("  voice [name]   - Change/show TTS voice")
			fmt.Println("  quit           - Exit")

		default:
			fmt.Println("Unknown command. Type 'help' for commands.")
		}
		fmt.Println()
	}
}

func testText(handler handlers.AIHandler, prompt string) {
	fmt.Printf("Generating text for: %s\n", prompt)

	req := handlers.AIRequest{
		Messages: []handlers.ContentBlock{
			{Type: handlers.ContentTypeText, Text: prompt},
		},
	}

	reqJSON, _ := json.Marshal(req)
	resp, err := handler.Call(string(reqJSON))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	var aiResp handlers.AIResponse
	json.Unmarshal([]byte(resp), &aiResp)

	if aiResp.Error != "" {
		fmt.Printf("AI Error: %s\n", aiResp.Error)
		return
	}

	for _, block := range aiResp.Content {
		if block.Type == handlers.ContentTypeText {
			fmt.Printf("Response: %s\n", block.Text)
			return
		}
	}
	fmt.Println("No text in response")
}

func testImage(handler handlers.AIHandler, prompt string) {
	fmt.Printf("Generating image for: %s\n", prompt)

	req := handlers.AIRequest{
		Messages: []handlers.ContentBlock{
			{Type: handlers.ContentTypeText, Text: "imagen: " + prompt},
		},
		Context: map[string]interface{}{"generate_image": true},
	}

	reqJSON, _ := json.Marshal(req)
	resp, err := handler.Call(string(reqJSON))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	var aiResp handlers.AIResponse
	json.Unmarshal([]byte(resp), &aiResp)

	if aiResp.Error != "" {
		fmt.Printf("AI Error: %s\n", aiResp.Error)
		return
	}

	for _, block := range aiResp.Content {
		if block.Type == handlers.ContentTypeImage && block.ImageRef != "" {
			fmt.Printf("Image saved: %s\n", block.ImageRef)
			return
		}
		if block.Type == handlers.ContentTypeText {
			fmt.Printf("Response: %s\n", block.Text)
		}
	}
}

func testTTS(handler handlers.AIHandler, text, voice string) string {
	fmt.Printf("Generating speech (%s): %s\n", voice, text)

	req := handlers.AIRequest{
		Messages: []handlers.ContentBlock{
			{Type: handlers.ContentTypeText, Text: "speak: " + text},
		},
		Context: map[string]interface{}{"tts": true, "voice": voice},
	}

	reqJSON, _ := json.Marshal(req)
	resp, err := handler.Call(string(reqJSON))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return ""
	}

	var aiResp handlers.AIResponse
	json.Unmarshal([]byte(resp), &aiResp)

	if aiResp.Error != "" {
		fmt.Printf("AI Error: %s\n", aiResp.Error)
		return ""
	}

	for _, block := range aiResp.Content {
		if block.Type == handlers.ContentTypeAudio && block.AudioRef != "" {
			fmt.Printf("Audio saved: %s\n", block.AudioRef)
			return block.AudioRef
		}
	}
	fmt.Println("No audio in response")
	return ""
}

func playAudio(path string) {
	fmt.Println("Playing audio...")
	cmd := exec.Command("afplay", path)
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Play error: %v\n", err)
	} else {
		fmt.Println("Audio finished.")
	}
}
