// Package handlers provides effect handler implementations for the game.
package handlers

import (
	"context"
	"fmt"
	"os"
)

// AIProvider identifies which AI backend to use.
type AIProvider string

const (
	AIProviderStub   AIProvider = "stub"
	AIProviderClaude AIProvider = "claude"
	AIProviderGemini AIProvider = "gemini"
)

// AIConfig holds configuration for AI handler creation.
type AIConfig struct {
	Provider AIProvider // Which provider to use (stub, claude, gemini)

	// Claude-specific config
	ClaudeAPIKey string // If empty, uses ANTHROPIC_API_KEY env var
	ClaudeModel  string // Default: claude-haiku-4-5-20251001

	// Gemini-specific config (Vertex AI preferred over API key)
	GeminiProject     string // GCP project for Vertex AI (uses ADC)
	GeminiLocation    string // GCP region for Vertex AI (default: us-central1)
	GeminiAPIKey      string // Fallback: GOOGLE_API_KEY env var
	GeminiModel       string // Default: gemini-2.5-flash
	GeminiImagenModel string // Default: gemini-2.5-flash-image
	GeminiTTSModel    string // Default: gemini-2.5-flash-tts (may require allowlisting)
	GeminiTTSVoice    string // Default: Kore (options: Aoede, Charon, Fenrir, Kore, Puck, Zephyr)

	// Shared config
	AssetsDir string // Where to save generated media (default: assets/generated)
}

// AIHandler is the interface that all AI handlers must implement.
// This matches the interface in sim_gen/handlers.go.
type AIHandler interface {
	Call(input string) (string, error)
}

// NewAIHandler creates an AI handler based on configuration.
// Falls back to stub if the configured provider fails to initialize.
func NewAIHandler(ctx context.Context, cfg AIConfig) (AIHandler, error) {
	switch cfg.Provider {
	case AIProviderClaude:
		handler, err := NewClaudeAIHandler(ClaudeConfig{
			APIKey: cfg.ClaudeAPIKey,
			Model:  cfg.ClaudeModel,
		})
		if err != nil {
			fmt.Printf("[AI] Claude init failed: %v, falling back to stub\n", err)
			return NewStubAIHandler(), nil
		}
		fmt.Println("[AI] Using Claude provider")
		return handler, nil

	case AIProviderGemini:
		handler, err := NewGeminiAIHandler(ctx, GeminiConfig{
			Project:     cfg.GeminiProject,
			Location:    cfg.GeminiLocation,
			APIKey:      cfg.GeminiAPIKey,
			Model:       cfg.GeminiModel,
			ImagenModel: cfg.GeminiImagenModel,
			TTSModel:    cfg.GeminiTTSModel,
			TTSVoice:    cfg.GeminiTTSVoice,
			AssetsDir:   cfg.AssetsDir,
		})
		if err != nil {
			fmt.Printf("[AI] Gemini init failed: %v, falling back to stub\n", err)
			return NewStubAIHandler(), nil
		}
		fmt.Println("[AI] Using Gemini provider (text + images + TTS)")
		return handler, nil

	case AIProviderStub:
		fmt.Println("[AI] Using stub provider")
		return NewStubAIHandler(), nil

	default:
		// Auto-detect based on available API keys
		return autoDetectProvider(ctx, cfg)
	}
}

// autoDetectProvider tries to initialize providers based on available env vars.
// Priority: Vertex AI (GOOGLE_CLOUD_PROJECT) > Gemini API Key > Claude API Key > Stub
func autoDetectProvider(ctx context.Context, cfg AIConfig) (AIHandler, error) {
	// Check for Vertex AI config (GOOGLE_CLOUD_PROJECT)
	hasVertexConfig := cfg.GeminiProject != "" ||
		os.Getenv("GOOGLE_CLOUD_PROJECT") != "" ||
		os.Getenv("GCLOUD_PROJECT") != ""

	// Check for Gemini API key
	hasGeminiKey := cfg.GeminiAPIKey != "" || os.Getenv("GOOGLE_API_KEY") != ""

	// Try Gemini (Vertex AI or API key)
	if hasVertexConfig || hasGeminiKey {
		handler, err := NewGeminiAIHandler(ctx, GeminiConfig{
			Project:     cfg.GeminiProject,
			Location:    cfg.GeminiLocation,
			APIKey:      cfg.GeminiAPIKey,
			Model:       cfg.GeminiModel,
			ImagenModel: cfg.GeminiImagenModel,
			TTSModel:    cfg.GeminiTTSModel,
			TTSVoice:    cfg.GeminiTTSVoice,
			AssetsDir:   cfg.AssetsDir,
		})
		if err == nil {
			fmt.Println("[AI] Auto-detected Gemini provider")
			return handler, nil
		}
		fmt.Printf("[AI] Gemini auto-detect failed: %v\n", err)
	}

	// Try Claude (API key only)
	if os.Getenv("ANTHROPIC_API_KEY") != "" || cfg.ClaudeAPIKey != "" {
		handler, err := NewClaudeAIHandler(ClaudeConfig{
			APIKey: cfg.ClaudeAPIKey,
			Model:  cfg.ClaudeModel,
		})
		if err == nil {
			fmt.Println("[AI] Auto-detected Claude provider")
			return handler, nil
		}
	}

	// Fall back to stub
	fmt.Println("[AI] No API keys found, using stub provider")
	return NewStubAIHandler(), nil
}

// NewAIHandlerFromEnv creates an AI handler using environment variables.
// AI_PROVIDER env var selects the provider (claude, gemini, stub).
// If not set, auto-detects based on available config:
//   - Vertex AI: GOOGLE_CLOUD_PROJECT (uses ADC for auth)
//   - Gemini API: GOOGLE_API_KEY
//   - Claude: ANTHROPIC_API_KEY
func NewAIHandlerFromEnv(ctx context.Context) (AIHandler, error) {
	provider := AIProvider(os.Getenv("AI_PROVIDER"))

	return NewAIHandler(ctx, AIConfig{
		Provider:  provider,
		AssetsDir: os.Getenv("AI_ASSETS_DIR"),

		// Claude config
		ClaudeModel: os.Getenv("CLAUDE_MODEL"),

		// Gemini config (Vertex AI takes priority over API key)
		GeminiProject:     os.Getenv("GOOGLE_CLOUD_PROJECT"),
		GeminiLocation:    os.Getenv("GOOGLE_CLOUD_LOCATION"),
		GeminiModel:       os.Getenv("GEMINI_MODEL"),
		GeminiImagenModel: os.Getenv("GEMINI_IMAGEN_MODEL"),
		GeminiTTSModel:    os.Getenv("GEMINI_TTS_MODEL"),
		GeminiTTSVoice:    os.Getenv("GEMINI_TTS_VOICE"),
	})
}

// ChainAIHandler wraps multiple handlers with fallback behavior.
type ChainAIHandler struct {
	handlers []AIHandler
}

// NewChainAIHandler creates a handler that tries each provider in order.
func NewChainAIHandler(handlers ...AIHandler) *ChainAIHandler {
	return &ChainAIHandler{handlers: handlers}
}

// Call tries each handler until one succeeds.
func (c *ChainAIHandler) Call(input string) (string, error) {
	var lastErr error
	for _, h := range c.handlers {
		result, err := h.Call(input)
		if err == nil {
			return result, nil
		}
		lastErr = err
	}
	if lastErr != nil {
		return "", lastErr
	}
	return "", fmt.Errorf("no handlers available")
}
