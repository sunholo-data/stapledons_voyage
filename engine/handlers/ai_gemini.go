// Package handlers provides effect handler implementations for the game.
package handlers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"google.golang.org/genai"
)

// GeminiAIHandler provides a real Gemini API implementation with multimodal support.
// Supports text, images (input/output/editing), and audio (input/output via TTS).
type GeminiAIHandler struct {
	client             *genai.Client
	model              string
	imagenModel        string // For image generation
	ttsModel           string // For text-to-speech
	ttsVoice           string // TTS voice name
	assetsDir          string // Where to save generated media
	lastGeneratedImage string // Track last generated image for iterative editing
}

// GeminiConfig holds configuration for the Gemini handler.
type GeminiConfig struct {
	// Vertex AI config (preferred - uses ADC)
	Project  string // GCP project ID for Vertex AI
	Location string // GCP region (e.g., "us-central1")

	// API Key config (fallback)
	APIKey string // If empty, uses GOOGLE_API_KEY env var

	// Model config
	Model       string // Default: gemini-2.5-flash
	ImagenModel string // Default: gemini-2.5-flash-image
	TTSModel    string // Default: gemini-2.5-flash-tts (may require allowlisting)
	TTSVoice    string // Default: Kore (options: Aoede, Charon, Fenrir, Kore, Puck, Zephyr, etc.)
	AssetsDir   string // Where to save generated media (default: assets/generated)
}

// NewGeminiAIHandler creates a Gemini AI handler with multimodal support.
// Authentication priority:
// 1. Vertex AI (Project + Location, uses Application Default Credentials)
// 2. API Key (GOOGLE_API_KEY env var)
// 3. Fail if neither configured
func NewGeminiAIHandler(ctx context.Context, cfg GeminiConfig) (*GeminiAIHandler, error) {
	model := cfg.Model
	if model == "" {
		model = "gemini-2.5-flash"
	}

	imagenModel := cfg.ImagenModel
	if imagenModel == "" {
		imagenModel = "gemini-2.5-flash-image" // "nano banana" - native image generation
	}

	ttsModel := cfg.TTSModel
	if ttsModel == "" {
		// Note: TTS may require allowlisting on Vertex AI
		// Gemini API uses: gemini-2.5-flash-preview-tts
		// Vertex AI uses: gemini-2.5-flash-tts
		ttsModel = "gemini-2.5-flash-tts"
	}

	ttsVoice := cfg.TTSVoice
	if ttsVoice == "" {
		ttsVoice = "Kore" // Default voice
	}

	assetsDir := cfg.AssetsDir
	if assetsDir == "" {
		assetsDir = "assets/generated"
	}

	// Try Vertex AI first (uses Application Default Credentials)
	project := cfg.Project
	if project == "" {
		project = os.Getenv("GOOGLE_CLOUD_PROJECT")
		if project == "" {
			project = os.Getenv("GCLOUD_PROJECT")
		}
	}

	location := cfg.Location
	if location == "" {
		location = os.Getenv("GOOGLE_CLOUD_LOCATION")
		if location == "" {
			location = "us-central1" // Default region
		}
	}

	var client *genai.Client
	var err error

	if project != "" {
		// Try Vertex AI backend (uses ADC)
		client, err = genai.NewClient(ctx, &genai.ClientConfig{
			Project:  project,
			Location: location,
			Backend:  genai.BackendVertexAI,
		})
		if err == nil {
			fmt.Printf("[Gemini] Using Vertex AI (project=%s, location=%s)\n", project, location)
			return &GeminiAIHandler{
				client:      client,
				model:       model,
				imagenModel: imagenModel,
				ttsModel:    ttsModel,
				ttsVoice:    ttsVoice,
				assetsDir:   assetsDir,
			}, nil
		}
		fmt.Printf("[Gemini] Vertex AI init failed: %v, trying API key...\n", err)
	}

	// Fall back to API Key
	apiKey := cfg.APIKey
	if apiKey == "" {
		apiKey = os.Getenv("GOOGLE_API_KEY")
	}

	if apiKey != "" {
		client, err = genai.NewClient(ctx, &genai.ClientConfig{
			APIKey:  apiKey,
			Backend: genai.BackendGeminiAPI,
		})
		if err == nil {
			fmt.Println("[Gemini] Using API Key backend")
			return &GeminiAIHandler{
				client:      client,
				model:       model,
				imagenModel: imagenModel,
				ttsModel:    ttsModel,
				ttsVoice:    ttsVoice,
				assetsDir:   assetsDir,
			}, nil
		}
		return nil, fmt.Errorf("creating Gemini client with API key: %w", err)
	}

	// Neither Vertex nor API key available
	return nil, fmt.Errorf("no Gemini auth configured: set GOOGLE_CLOUD_PROJECT for Vertex AI or GOOGLE_API_KEY for API access")
}

// Call implements the AIHandler interface.
// Input is JSON-encoded AIRequest, output is JSON-encoded AIResponse.
// Supports text, images (input/generation/editing), and audio (input and TTS output).
func (h *GeminiAIHandler) Call(input string) (string, error) {
	// Parse input as AIRequest
	var req AIRequest
	if err := json.Unmarshal([]byte(input), &req); err != nil {
		// If not valid JSON, treat as plain text
		req = AIRequest{
			Messages: []ContentBlock{{Type: ContentTypeText, Text: input}},
		}
	}

	// Check if this is an image edit request (must check before generate)
	if h.isImageEditRequest(req) {
		return h.editImage(req)
	}

	// Check if this is an image generation request
	if h.isImageGenRequest(req) {
		return h.generateImage(req)
	}

	// Check if TTS is requested
	if h.isTTSRequest(req) {
		return h.textToSpeech(req)
	}

	// Otherwise, do standard multimodal chat
	return h.chat(req)
}

// isImageGenRequest checks if the request is asking for image generation.
func (h *GeminiAIHandler) isImageGenRequest(req AIRequest) bool {
	for _, msg := range req.Messages {
		if msg.Type == ContentTypeText {
			lower := strings.ToLower(msg.Text)
			if strings.Contains(lower, "generate image") ||
				strings.Contains(lower, "create image") ||
				strings.Contains(lower, "draw") ||
				strings.Contains(lower, "imagen:") {
				return true
			}
		}
	}
	// Also check context for explicit image generation flag
	if req.Context != nil {
		if genImg, ok := req.Context["generate_image"].(bool); ok && genImg {
			return true
		}
	}
	return false
}

// isTTSRequest checks if the request is asking for text-to-speech.
func (h *GeminiAIHandler) isTTSRequest(req AIRequest) bool {
	for _, msg := range req.Messages {
		if msg.Type == ContentTypeText {
			lower := strings.ToLower(msg.Text)
			if strings.Contains(lower, "speak:") ||
				strings.Contains(lower, "say:") ||
				strings.Contains(lower, "tts:") ||
				strings.Contains(lower, "voice:") {
				return true
			}
		}
	}
	// Also check context for explicit TTS flag
	if req.Context != nil {
		if tts, ok := req.Context["tts"].(bool); ok && tts {
			return true
		}
	}
	return false
}

// isImageEditRequest checks if the request is asking for image editing.
// Image editing requires either:
// - An image in the messages + edit keywords
// - Context flag "edit_image": true
// - Keywords like "edit:", "modify:", "change:", "refine:", "fix:"
// - Reference to "last image" or "previous image"
func (h *GeminiAIHandler) isImageEditRequest(req AIRequest) bool {
	hasImage := false
	hasEditKeyword := false
	usesLastImage := false

	for _, msg := range req.Messages {
		if msg.Type == ContentTypeImage && msg.ImageRef != "" {
			hasImage = true
		}
		if msg.Type == ContentTypeText {
			lower := strings.ToLower(msg.Text)
			// Check for edit keywords
			if strings.Contains(lower, "edit:") ||
				strings.Contains(lower, "edit image") ||
				strings.Contains(lower, "modify:") ||
				strings.Contains(lower, "modify image") ||
				strings.Contains(lower, "change:") ||
				strings.Contains(lower, "refine:") ||
				strings.Contains(lower, "fix:") ||
				strings.Contains(lower, "adjust:") ||
				strings.Contains(lower, "update image") {
				hasEditKeyword = true
			}
			// Check for reference to last/previous image
			if strings.Contains(lower, "last image") ||
				strings.Contains(lower, "previous image") ||
				strings.Contains(lower, "that image") ||
				strings.Contains(lower, "the image") {
				usesLastImage = true
			}
		}
	}

	// Check context for explicit edit flag or reference image
	if req.Context != nil {
		if editImg, ok := req.Context["edit_image"].(bool); ok && editImg {
			hasEditKeyword = true
		}
		if refImg, ok := req.Context["reference_image"].(string); ok && refImg != "" {
			hasImage = true
		}
		if useLast, ok := req.Context["use_last_image"].(bool); ok && useLast {
			usesLastImage = true
		}
	}

	// Need edit keyword AND (explicit image OR reference to last image with existing last image)
	return hasEditKeyword && (hasImage || (usesLastImage && h.lastGeneratedImage != ""))
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

// loadImagePart loads an image from file path and creates a Gemini part.
func (h *GeminiAIHandler) loadImagePart(ref string, mimeType string) (*genai.Part, error) {
	data, err := os.ReadFile(ref)
	if err != nil {
		return nil, fmt.Errorf("reading image file: %w", err)
	}

	if mimeType == "" {
		mimeType = mimeTypeFromExt(filepath.Ext(ref))
	}

	return &genai.Part{
		InlineData: &genai.Blob{
			MIMEType: mimeType,
			Data:     data,
		},
	}, nil
}

// loadAudioPart loads an audio file and creates a Gemini part.
func (h *GeminiAIHandler) loadAudioPart(ref string, mimeType string) (*genai.Part, error) {
	data, err := os.ReadFile(ref)
	if err != nil {
		return nil, fmt.Errorf("reading audio file: %w", err)
	}

	if mimeType == "" {
		ext := strings.ToLower(filepath.Ext(ref))
		switch ext {
		case ".wav":
			mimeType = "audio/wav"
		case ".mp3":
			mimeType = "audio/mp3"
		case ".ogg":
			mimeType = "audio/ogg"
		case ".flac":
			mimeType = "audio/flac"
		default:
			mimeType = "audio/wav"
		}
	}

	return &genai.Part{
		InlineData: &genai.Blob{
			MIMEType: mimeType,
			Data:     data,
		},
	}, nil
}

// saveInlineMedia saves inline media data (image or audio) to a file.
func (h *GeminiAIHandler) saveInlineMedia(data *genai.Blob) (string, ContentType, error) {
	if err := os.MkdirAll(h.assetsDir, 0755); err != nil {
		return "", "", err
	}

	// Determine extension and content type from MIME type
	var ext string
	var contentType ContentType

	if strings.HasPrefix(data.MIMEType, "image/") {
		contentType = ContentTypeImage
		switch data.MIMEType {
		case "image/jpeg":
			ext = ".jpg"
		case "image/gif":
			ext = ".gif"
		case "image/webp":
			ext = ".webp"
		default:
			ext = ".png"
		}
	} else if strings.HasPrefix(data.MIMEType, "audio/") {
		contentType = ContentTypeAudio
		switch data.MIMEType {
		case "audio/mp3", "audio/mpeg":
			ext = ".mp3"
		case "audio/ogg":
			ext = ".ogg"
		case "audio/flac":
			ext = ".flac"
		default:
			ext = ".wav"
		}
	} else {
		ext = ".bin"
		contentType = ContentTypeText // fallback
	}

	// Generate unique filename
	filename := fmt.Sprintf("response_%d%s", time.Now().UnixNano(), ext)
	mediaPath := filepath.Join(h.assetsDir, filename)

	if err := os.WriteFile(mediaPath, data.Data, 0644); err != nil {
		return "", "", err
	}

	return mediaPath, contentType, nil
}

// mimeTypeFromExt returns MIME type from file extension.
func mimeTypeFromExt(ext string) string {
	switch strings.ToLower(ext) {
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	default:
		return "image/png"
	}
}

// errorResponse creates a JSON error response.
func (h *GeminiAIHandler) errorResponse(msg string) (string, error) {
	resp := AIResponse{Error: msg}
	output, _ := json.Marshal(resp)
	return string(output), fmt.Errorf("%s", msg)
}

// LastGeneratedImage returns the path to the most recently generated image.
// Returns empty string if no image has been generated yet.
func (h *GeminiAIHandler) LastGeneratedImage() string {
	return h.lastGeneratedImage
}

// ClearLastGeneratedImage clears the reference to the last generated image.
func (h *GeminiAIHandler) ClearLastGeneratedImage() {
	h.lastGeneratedImage = ""
}

// SetLastGeneratedImage allows setting a reference image manually.
// Useful for loading a previous session's image for continued editing.
func (h *GeminiAIHandler) SetLastGeneratedImage(path string) {
	h.lastGeneratedImage = path
}

// EncodeImageBase64 is a helper to encode an image file as base64.
func EncodeImageBase64(path string) (string, string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", "", err
	}

	mimeType := mimeTypeFromExt(filepath.Ext(path))
	return base64.StdEncoding.EncodeToString(data), mimeType, nil
}

// EncodeAudioBase64 is a helper to encode an audio file as base64.
func EncodeAudioBase64(path string) (string, string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", "", err
	}

	ext := strings.ToLower(filepath.Ext(path))
	mimeType := "audio/wav"
	switch ext {
	case ".mp3":
		mimeType = "audio/mp3"
	case ".ogg":
		mimeType = "audio/ogg"
	case ".flac":
		mimeType = "audio/flac"
	}

	return base64.StdEncoding.EncodeToString(data), mimeType, nil
}
