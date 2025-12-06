// Package handlers provides effect handler implementations for the game.
package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"google.golang.org/genai"
)

// textToSpeech generates audio from text using Gemini's native TTS capability.
// Tries dedicated TTS model first, falls back to regular model with audio output.
// Available voices: Aoede, Charon, Fenrir, Kore, Puck, Zephyr, Enceladus, etc.
func (h *GeminiAIHandler) textToSpeech(req AIRequest) (string, error) {
	ctx := context.Background()

	// Extract the text to speak and optional voice override
	var textToSpeak string
	voice := h.ttsVoice // Use configured default
	for _, msg := range req.Messages {
		if msg.Type == ContentTypeText {
			text := msg.Text
			// Strip TTS prefixes
			text = strings.TrimPrefix(text, "speak:")
			text = strings.TrimPrefix(text, "Speak:")
			text = strings.TrimPrefix(text, "say:")
			text = strings.TrimPrefix(text, "Say:")
			text = strings.TrimPrefix(text, "tts:")
			text = strings.TrimPrefix(text, "TTS:")
			text = strings.TrimPrefix(text, "voice:")
			text = strings.TrimPrefix(text, "Voice:")
			textToSpeak = strings.TrimSpace(text)
			if textToSpeak != "" {
				break
			}
		}
	}

	// Check context for voice override
	if req.Context != nil {
		if v, ok := req.Context["voice"].(string); ok && v != "" {
			voice = v
		}
	}

	if textToSpeak == "" {
		return h.errorResponse("no text for TTS")
	}

	// Config for TTS with AUDIO response modality
	config := &genai.GenerateContentConfig{
		ResponseModalities: []string{"AUDIO"},
		SpeechConfig: &genai.SpeechConfig{
			VoiceConfig: &genai.VoiceConfig{
				PrebuiltVoiceConfig: &genai.PrebuiltVoiceConfig{
					VoiceName: voice,
				},
			},
		},
	}

	// For TTS, we just send the text directly - the model speaks it
	contents := []*genai.Content{
		genai.NewContentFromText(textToSpeak, genai.RoleUser),
	}

	// Try dedicated TTS model first
	resp, err := h.client.Models.GenerateContent(ctx, h.ttsModel, contents, config)
	if err != nil {
		// Fall back to regular model with audio output
		fmt.Printf("[TTS] Dedicated model failed: %v, trying regular model...\n", err)
		resp, err = h.client.Models.GenerateContent(ctx, h.model, contents, config)
		if err != nil {
			return h.errorResponse(fmt.Sprintf("Gemini TTS error: %v", err))
		}
	}

	// Extract audio from response
	var responseBlocks []ContentBlock
	if resp != nil && len(resp.Candidates) > 0 && resp.Candidates[0].Content != nil {
		for _, part := range resp.Candidates[0].Content.Parts {
			if part.InlineData != nil && strings.HasPrefix(part.InlineData.MIMEType, "audio/") {
				// Save audio file
				audioPath, _, err := h.saveInlineMedia(part.InlineData)
				if err == nil {
					responseBlocks = append(responseBlocks, ContentBlock{
						Type:     ContentTypeAudio,
						AudioRef: audioPath,
						MimeType: part.InlineData.MIMEType,
					})
				}
			}
			if part.Text != "" {
				responseBlocks = append(responseBlocks, ContentBlock{
					Type: ContentTypeText,
					Text: part.Text,
				})
			}
		}
	}

	if len(responseBlocks) == 0 {
		// Fallback - return the text at least
		responseBlocks = append(responseBlocks, ContentBlock{
			Type: ContentTypeText,
			Text: fmt.Sprintf("(TTS generation returned no audio for: %s)", textToSpeak),
		})
	}

	aiResp := AIResponse{Content: responseBlocks}
	output, err := json.Marshal(aiResp)
	if err != nil {
		return h.errorResponse(fmt.Sprintf("encoding response: %v", err))
	}

	return string(output), nil
}
