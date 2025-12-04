// Package handlers provides effect handler implementations for the game.
package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// ClaudeAIHandler provides a real Claude API implementation.
type ClaudeAIHandler struct {
	client *anthropic.Client
	model  anthropic.Model
}

// ClaudeConfig holds configuration for the Claude handler.
type ClaudeConfig struct {
	APIKey string // If empty, uses ANTHROPIC_API_KEY env var
	Model  string // Default: claude-sonnet-4-20250514
}

// NewClaudeAIHandler creates a Claude AI handler.
func NewClaudeAIHandler(cfg ClaudeConfig) (*ClaudeAIHandler, error) {
	apiKey := cfg.APIKey
	if apiKey == "" {
		apiKey = os.Getenv("ANTHROPIC_API_KEY")
	}
	if apiKey == "" {
		return nil, fmt.Errorf("ANTHROPIC_API_KEY not set")
	}

	model := cfg.Model
	if model == "" {
		model = "claude-haiku-4-5-20251001"
	}

	client := anthropic.NewClient(option.WithAPIKey(apiKey))

	return &ClaudeAIHandler{
		client: &client,
		model:  anthropic.Model(model),
	}, nil
}

// Call implements the AIHandler interface.
// Input is JSON-encoded AIRequest, output is JSON-encoded AIResponse.
func (h *ClaudeAIHandler) Call(input string) (string, error) {
	// Parse input as AIRequest
	var req AIRequest
	if err := json.Unmarshal([]byte(input), &req); err != nil {
		// If not valid JSON, treat as plain text
		req = AIRequest{
			Messages: []ContentBlock{{Type: ContentTypeText, Text: input}},
		}
	}

	// Build Claude messages from our request
	var contentBlocks []anthropic.ContentBlockParamUnion
	for _, msg := range req.Messages {
		if msg.Type == ContentTypeText && msg.Text != "" {
			contentBlocks = append(contentBlocks, anthropic.NewTextBlock(msg.Text))
		}
		// Note: Claude also supports images via anthropic.NewImageBlockBase64
		// but we're keeping this handler text-only per requirements
	}

	if len(contentBlocks) == 0 {
		return h.errorResponse("no text content in request")
	}

	// Build the request
	params := anthropic.MessageNewParams{
		Model:     h.model,
		MaxTokens: 1024,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(contentBlocks...),
		},
	}

	// Add system prompt if provided
	if req.System != "" {
		params.System = []anthropic.TextBlockParam{
			{Text: req.System},
		}
	}

	// Call the API
	ctx := context.Background()
	message, err := h.client.Messages.New(ctx, params)
	if err != nil {
		return h.errorResponse(fmt.Sprintf("Claude API error: %v", err))
	}

	// Convert response to our format
	var responseBlocks []ContentBlock
	for _, block := range message.Content {
		if block.Type == "text" {
			responseBlocks = append(responseBlocks, ContentBlock{
				Type: ContentTypeText,
				Text: block.Text,
			})
		}
	}

	resp := AIResponse{Content: responseBlocks}
	output, err := json.Marshal(resp)
	if err != nil {
		return h.errorResponse(fmt.Sprintf("encoding response: %v", err))
	}

	return string(output), nil
}

// errorResponse creates a JSON error response.
func (h *ClaudeAIHandler) errorResponse(msg string) (string, error) {
	resp := AIResponse{Error: msg}
	output, _ := json.Marshal(resp)
	return string(output), fmt.Errorf("%s", msg)
}
