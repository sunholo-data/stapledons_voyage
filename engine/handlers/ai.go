// Package handlers provides effect handler implementations for the game.
package handlers

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ContentType represents the type of content in a message
type ContentType string

const (
	ContentTypeText  ContentType = "text"
	ContentTypeImage ContentType = "image"
	ContentTypeAudio ContentType = "audio"
	ContentTypeVideo ContentType = "video"
)

// ContentBlock represents a piece of content (text, image, audio, video)
type ContentBlock struct {
	Type     ContentType `json:"type"`
	Text     string      `json:"text,omitempty"`
	ImageRef string      `json:"image_ref,omitempty"` // Asset path or URL
	AudioRef string      `json:"audio_ref,omitempty"` // Asset path or URL
	VideoRef string      `json:"video_ref,omitempty"` // Asset path or URL
	MimeType string      `json:"mime_type,omitempty"`
	AltText  string      `json:"alt_text,omitempty"` // Accessibility text for images
}

// AIRequest represents a multimodal AI request
type AIRequest struct {
	Messages []ContentBlock         `json:"messages"`
	Context  map[string]interface{} `json:"context,omitempty"`
	System   string                 `json:"system,omitempty"` // System prompt
}

// AIResponse represents a multimodal AI response
type AIResponse struct {
	Content []ContentBlock `json:"content"`
	Error   string         `json:"error,omitempty"`
}

// StubAIHandler provides a stub implementation for testing.
// Returns canned responses based on input patterns.
type StubAIHandler struct {
	// Responses maps input patterns to canned responses
	Responses map[string]AIResponse
	// DefaultResponse is returned when no pattern matches
	DefaultResponse AIResponse
	// LogCalls enables logging of AI calls for debugging
	LogCalls bool
}

// NewStubAIHandler creates a stub AI handler with default responses.
func NewStubAIHandler() *StubAIHandler {
	return &StubAIHandler{
		Responses: make(map[string]AIResponse),
		DefaultResponse: AIResponse{
			Content: []ContentBlock{
				{Type: ContentTypeText, Text: "I am a stub AI. I cannot provide real responses."},
			},
		},
		LogCalls: true,
	}
}

// Call implements the AIHandler interface.
// Input is JSON-encoded AIRequest, output is JSON-encoded AIResponse.
func (h *StubAIHandler) Call(input string) (string, error) {
	if h.LogCalls {
		fmt.Printf("[AI STUB] Received: %s\n", truncate(input, 200))
	}

	// Parse input as AIRequest
	var req AIRequest
	if err := json.Unmarshal([]byte(input), &req); err != nil {
		// If not valid JSON, treat as plain text
		req = AIRequest{
			Messages: []ContentBlock{{Type: ContentTypeText, Text: input}},
		}
	}

	// Find matching response
	resp := h.findResponse(req)

	// Encode response
	output, err := json.Marshal(resp)
	if err != nil {
		return "", fmt.Errorf("encoding AI response: %w", err)
	}

	if h.LogCalls {
		fmt.Printf("[AI STUB] Responding: %s\n", truncate(string(output), 200))
	}

	return string(output), nil
}

// findResponse looks for a matching canned response.
func (h *StubAIHandler) findResponse(req AIRequest) AIResponse {
	// Extract text from messages for pattern matching
	var texts []string
	for _, msg := range req.Messages {
		if msg.Type == ContentTypeText && msg.Text != "" {
			texts = append(texts, msg.Text)
		}
	}
	combined := strings.ToLower(strings.Join(texts, " "))

	// Check registered patterns
	for pattern, resp := range h.Responses {
		if strings.Contains(combined, strings.ToLower(pattern)) {
			return resp
		}
	}

	// Check for common intents
	if strings.Contains(combined, "decide") || strings.Contains(combined, "decision") {
		return AIResponse{
			Content: []ContentBlock{
				{Type: ContentTypeText, Text: `{"action": "wait", "reason": "I need more information to decide."}`},
			},
		}
	}

	if strings.Contains(combined, "emotion") || strings.Contains(combined, "feel") {
		return AIResponse{
			Content: []ContentBlock{
				{Type: ContentTypeText, Text: `{"emotion": "contemplative", "dialogue": "The vastness of space makes me feel small, yet purposeful."}`},
			},
		}
	}

	if strings.Contains(combined, "describe") || strings.Contains(combined, "what do you see") {
		return AIResponse{
			Content: []ContentBlock{
				{Type: ContentTypeText, Text: "I see stars scattered like diamonds across the void."},
				{Type: ContentTypeImage, ImageRef: "assets/sprites/starfield.png", AltText: "A field of distant stars"},
			},
		}
	}

	return h.DefaultResponse
}

// RegisterResponse adds a canned response for a pattern.
func (h *StubAIHandler) RegisterResponse(pattern string, resp AIResponse) {
	h.Responses[pattern] = resp
}

// RegisterTextResponse adds a simple text response for a pattern.
func (h *StubAIHandler) RegisterTextResponse(pattern, text string) {
	h.Responses[pattern] = AIResponse{
		Content: []ContentBlock{{Type: ContentTypeText, Text: text}},
	}
}

// RegisterMultimodalResponse adds a response with multiple content types.
func (h *StubAIHandler) RegisterMultimodalResponse(pattern string, blocks ...ContentBlock) {
	h.Responses[pattern] = AIResponse{Content: blocks}
}

// truncate shortens a string for logging
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// TextBlock is a helper to create a text content block
func TextBlock(text string) ContentBlock {
	return ContentBlock{Type: ContentTypeText, Text: text}
}

// ImageBlock is a helper to create an image content block
func ImageBlock(ref, altText string) ContentBlock {
	return ContentBlock{Type: ContentTypeImage, ImageRef: ref, AltText: altText}
}

// AudioBlock is a helper to create an audio content block
func AudioBlock(ref string) ContentBlock {
	return ContentBlock{Type: ContentTypeAudio, AudioRef: ref}
}

// VideoBlock is a helper to create a video content block
func VideoBlock(ref string) ContentBlock {
	return ContentBlock{Type: ContentTypeVideo, VideoRef: ref}
}
