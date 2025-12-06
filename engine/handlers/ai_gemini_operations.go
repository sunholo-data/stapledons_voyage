// Package handlers provides effect handler implementations for the game.
package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"google.golang.org/genai"
)

// chat handles standard multimodal conversation.
func (h *GeminiAIHandler) chat(req AIRequest) (string, error) {
	ctx := context.Background()

	// Build parts from our request
	var parts []*genai.Part
	for _, msg := range req.Messages {
		switch msg.Type {
		case ContentTypeText:
			if msg.Text != "" {
				parts = append(parts, genai.NewPartFromText(msg.Text))
			}
		case ContentTypeImage:
			// Load image from file
			if msg.ImageRef != "" {
				imgPart, err := h.loadImagePart(msg.ImageRef, msg.MimeType)
				if err != nil {
					fmt.Printf("[Gemini] Warning: failed to load image %s: %v\n", msg.ImageRef, err)
					continue
				}
				parts = append(parts, imgPart)
			}
		case ContentTypeAudio:
			// Load audio from file
			if msg.AudioRef != "" {
				audioPart, err := h.loadAudioPart(msg.AudioRef, msg.MimeType)
				if err != nil {
					fmt.Printf("[Gemini] Warning: failed to load audio %s: %v\n", msg.AudioRef, err)
					continue
				}
				parts = append(parts, audioPart)
			}
		}
	}

	if len(parts) == 0 {
		return h.errorResponse("no content in request")
	}

	// Build config
	config := &genai.GenerateContentConfig{
		MaxOutputTokens: 1024,
	}

	// Add system prompt if provided
	if req.System != "" {
		config.SystemInstruction = genai.NewContentFromText(req.System, genai.RoleUser)
	}

	// Create content and call API
	contents := []*genai.Content{
		genai.NewContentFromParts(parts, genai.RoleUser),
	}
	resp, err := h.client.Models.GenerateContent(ctx, h.model, contents, config)
	if err != nil {
		return h.errorResponse(fmt.Sprintf("Gemini API error: %v", err))
	}

	// Convert response to our format
	var responseBlocks []ContentBlock
	if resp != nil && len(resp.Candidates) > 0 && resp.Candidates[0].Content != nil {
		for _, part := range resp.Candidates[0].Content.Parts {
			if part.Text != "" {
				responseBlocks = append(responseBlocks, ContentBlock{
					Type: ContentTypeText,
					Text: part.Text,
				})
			}
			// Handle inline image data if returned
			if part.InlineData != nil {
				mediaPath, mediaType, err := h.saveInlineMedia(part.InlineData)
				if err == nil {
					responseBlocks = append(responseBlocks, ContentBlock{
						Type:     mediaType,
						ImageRef: mediaPath,
						AudioRef: mediaPath,
						MimeType: part.InlineData.MIMEType,
					})
				}
			}
		}
	}

	if len(responseBlocks) == 0 {
		responseBlocks = append(responseBlocks, ContentBlock{
			Type: ContentTypeText,
			Text: "(no response)",
		})
	}

	aiResp := AIResponse{Content: responseBlocks}
	output, err := json.Marshal(aiResp)
	if err != nil {
		return h.errorResponse(fmt.Sprintf("encoding response: %v", err))
	}

	return string(output), nil
}

// generateImage uses Gemini native image generation (nano banana).
func (h *GeminiAIHandler) generateImage(req AIRequest) (string, error) {
	ctx := context.Background()

	// Extract the prompt
	var prompt string
	for _, msg := range req.Messages {
		if msg.Type == ContentTypeText {
			// Strip prefixes like "imagen:" or "generate image:"
			text := msg.Text
			text = strings.TrimPrefix(text, "imagen:")
			text = strings.TrimPrefix(text, "Imagen:")
			text = strings.TrimPrefix(text, "generate image:")
			text = strings.TrimPrefix(text, "Generate image:")
			text = strings.TrimPrefix(text, "create image:")
			text = strings.TrimPrefix(text, "Create image:")
			text = strings.TrimPrefix(text, "draw:")
			text = strings.TrimPrefix(text, "Draw:")
			prompt = strings.TrimSpace(text)
			if prompt != "" {
				break
			}
		}
	}

	if prompt == "" {
		return h.errorResponse("no prompt for image generation")
	}

	// Use Gemini native image generation with IMAGE response modality
	config := &genai.GenerateContentConfig{
		ResponseModalities: []string{"IMAGE", "TEXT"},
	}

	contents := []*genai.Content{
		genai.NewContentFromText(prompt, genai.RoleUser),
	}

	resp, err := h.client.Models.GenerateContent(ctx, h.imagenModel, contents, config)
	if err != nil {
		return h.errorResponse(fmt.Sprintf("Image generation error: %v", err))
	}

	// Extract images from response
	var responseBlocks []ContentBlock
	if resp != nil && len(resp.Candidates) > 0 && resp.Candidates[0].Content != nil {
		for _, part := range resp.Candidates[0].Content.Parts {
			// Handle inline image data
			if part.InlineData != nil && strings.HasPrefix(part.InlineData.MIMEType, "image/") {
				// Save the image
				mediaPath, _, err := h.saveInlineMedia(part.InlineData)
				if err == nil {
					// Track as last generated image for iterative editing
					h.lastGeneratedImage = mediaPath
					responseBlocks = append(responseBlocks, ContentBlock{
						Type:     ContentTypeImage,
						ImageRef: mediaPath,
						MimeType: part.InlineData.MIMEType,
						AltText:  prompt,
					})
				}
			}
			// Also capture any text response
			if part.Text != "" {
				responseBlocks = append(responseBlocks, ContentBlock{
					Type: ContentTypeText,
					Text: part.Text,
				})
			}
		}
	}

	if len(responseBlocks) == 0 {
		return h.errorResponse("no images generated")
	}

	aiResp := AIResponse{Content: responseBlocks}
	output, err := json.Marshal(aiResp)
	if err != nil {
		return h.errorResponse(fmt.Sprintf("encoding response: %v", err))
	}

	return string(output), nil
}

// editImage sends a reference image with an edit prompt to generate a modified version.
// Supports explicit image in request, context["reference_image"], or "last image" reference.
func (h *GeminiAIHandler) editImage(req AIRequest) (string, error) {
	ctx := context.Background()

	// Find the reference image
	var refImagePath string

	// Priority 1: Explicit image in messages
	for _, msg := range req.Messages {
		if msg.Type == ContentTypeImage && msg.ImageRef != "" {
			refImagePath = msg.ImageRef
			break
		}
	}

	// Priority 2: Context reference_image
	if refImagePath == "" && req.Context != nil {
		if refImg, ok := req.Context["reference_image"].(string); ok && refImg != "" {
			refImagePath = refImg
		}
	}

	// Priority 3: Use last generated image
	if refImagePath == "" && h.lastGeneratedImage != "" {
		refImagePath = h.lastGeneratedImage
	}

	if refImagePath == "" {
		return h.errorResponse("no reference image for editing - provide an image or generate one first")
	}

	// Extract the edit prompt
	var editPrompt string
	for _, msg := range req.Messages {
		if msg.Type == ContentTypeText {
			text := msg.Text
			// Strip edit prefixes
			text = strings.TrimPrefix(text, "edit:")
			text = strings.TrimPrefix(text, "Edit:")
			text = strings.TrimPrefix(text, "modify:")
			text = strings.TrimPrefix(text, "Modify:")
			text = strings.TrimPrefix(text, "change:")
			text = strings.TrimPrefix(text, "Change:")
			text = strings.TrimPrefix(text, "refine:")
			text = strings.TrimPrefix(text, "Refine:")
			text = strings.TrimPrefix(text, "fix:")
			text = strings.TrimPrefix(text, "Fix:")
			text = strings.TrimPrefix(text, "adjust:")
			text = strings.TrimPrefix(text, "Adjust:")
			editPrompt = strings.TrimSpace(text)
			if editPrompt != "" {
				break
			}
		}
	}

	if editPrompt == "" {
		return h.errorResponse("no edit prompt provided")
	}

	// Load the reference image
	imgPart, err := h.loadImagePart(refImagePath, "")
	if err != nil {
		return h.errorResponse(fmt.Sprintf("failed to load reference image %s: %v", refImagePath, err))
	}

	// Build multimodal content: image + edit instruction
	parts := []*genai.Part{
		imgPart,
		genai.NewPartFromText(fmt.Sprintf("Edit this image: %s", editPrompt)),
	}

	// Use image generation model with IMAGE response modality
	config := &genai.GenerateContentConfig{
		ResponseModalities: []string{"IMAGE", "TEXT"},
	}

	contents := []*genai.Content{
		genai.NewContentFromParts(parts, genai.RoleUser),
	}

	resp, err := h.client.Models.GenerateContent(ctx, h.imagenModel, contents, config)
	if err != nil {
		return h.errorResponse(fmt.Sprintf("Image editing error: %v", err))
	}

	// Extract images from response
	var responseBlocks []ContentBlock
	if resp != nil && len(resp.Candidates) > 0 && resp.Candidates[0].Content != nil {
		for _, part := range resp.Candidates[0].Content.Parts {
			// Handle inline image data
			if part.InlineData != nil && strings.HasPrefix(part.InlineData.MIMEType, "image/") {
				// Save the image
				mediaPath, _, err := h.saveInlineMedia(part.InlineData)
				if err == nil {
					// Track as last generated image for further edits
					h.lastGeneratedImage = mediaPath
					responseBlocks = append(responseBlocks, ContentBlock{
						Type:     ContentTypeImage,
						ImageRef: mediaPath,
						MimeType: part.InlineData.MIMEType,
						AltText:  fmt.Sprintf("Edited: %s", editPrompt),
					})
				}
			}
			// Also capture any text response
			if part.Text != "" {
				responseBlocks = append(responseBlocks, ContentBlock{
					Type: ContentTypeText,
					Text: part.Text,
				})
			}
		}
	}

	if len(responseBlocks) == 0 {
		return h.errorResponse("no edited image generated")
	}

	aiResp := AIResponse{Content: responseBlocks}
	output, err := json.Marshal(aiResp)
	if err != nil {
		return h.errorResponse(fmt.Sprintf("encoding response: %v", err))
	}

	return string(output), nil
}
