// Package handlers provides effect handler implementations for the game.
package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

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

// geminiDefaults returns config values with defaults applied
func geminiDefaults(cfg GeminiConfig) (model, imagenModel, ttsModel, ttsVoice, assetsDir string) {
	model = cfg.Model
	if model == "" {
		model = "gemini-2.5-flash"
	}
	imagenModel = cfg.ImagenModel
	if imagenModel == "" {
		imagenModel = "gemini-2.5-flash-image"
	}
	ttsModel = cfg.TTSModel
	if ttsModel == "" {
		ttsModel = "gemini-2.5-flash-tts"
	}
	ttsVoice = cfg.TTSVoice
	if ttsVoice == "" {
		ttsVoice = "Kore"
	}
	assetsDir = cfg.AssetsDir
	if assetsDir == "" {
		assetsDir = "assets/generated"
	}
	return
}

// newGeminiHandler creates a handler with the given client and config defaults
func newGeminiHandler(client *genai.Client, model, imagenModel, ttsModel, ttsVoice, assetsDir string) *GeminiAIHandler {
	return &GeminiAIHandler{
		client:      client,
		model:       model,
		imagenModel: imagenModel,
		ttsModel:    ttsModel,
		ttsVoice:    ttsVoice,
		assetsDir:   assetsDir,
	}
}

// NewGeminiAIHandler creates a Gemini AI handler with multimodal support.
// Authentication priority: 1. Vertex AI (ADC), 2. API Key, 3. Fail
func NewGeminiAIHandler(ctx context.Context, cfg GeminiConfig) (*GeminiAIHandler, error) {
	model, imagenModel, ttsModel, ttsVoice, assetsDir := geminiDefaults(cfg)

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
			location = "us-central1"
		}
	}

	if project != "" {
		client, err := genai.NewClient(ctx, &genai.ClientConfig{
			Project:  project,
			Location: location,
			Backend:  genai.BackendVertexAI,
		})
		if err == nil {
			fmt.Printf("[Gemini] Using Vertex AI (project=%s, location=%s)\n", project, location)
			return newGeminiHandler(client, model, imagenModel, ttsModel, ttsVoice, assetsDir), nil
		}
		fmt.Printf("[Gemini] Vertex AI init failed: %v, trying API key...\n", err)
	}

	// Fall back to API Key
	apiKey := cfg.APIKey
	if apiKey == "" {
		apiKey = os.Getenv("GOOGLE_API_KEY")
	}

	if apiKey != "" {
		client, err := genai.NewClient(ctx, &genai.ClientConfig{
			APIKey:  apiKey,
			Backend: genai.BackendGeminiAPI,
		})
		if err == nil {
			fmt.Println("[Gemini] Using API Key backend")
			return newGeminiHandler(client, model, imagenModel, ttsModel, ttsVoice, assetsDir), nil
		}
		return nil, fmt.Errorf("creating Gemini client with API key: %w", err)
	}

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
