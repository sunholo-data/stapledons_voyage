package handlers

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestStubAIHandler(t *testing.T) {
	handler := NewStubAIHandler()

	// Test plain text input
	resp, err := handler.Call("Hello")
	if err != nil {
		t.Fatalf("stub handler returned error: %v", err)
	}

	var aiResp AIResponse
	if err := json.Unmarshal([]byte(resp), &aiResp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if aiResp.Error != "" {
		t.Errorf("unexpected error in response: %s", aiResp.Error)
	}

	if len(aiResp.Content) == 0 {
		t.Error("expected content in response")
	}
}

func TestStubAIHandlerJSON(t *testing.T) {
	handler := NewStubAIHandler()

	// Test JSON input
	req := AIRequest{
		System: "You are a helpful assistant",
		Messages: []ContentBlock{
			{Type: ContentTypeText, Text: "What is 2+2?"},
		},
	}
	reqJSON, _ := json.Marshal(req)

	resp, err := handler.Call(string(reqJSON))
	if err != nil {
		t.Fatalf("stub handler returned error: %v", err)
	}

	var aiResp AIResponse
	if err := json.Unmarshal([]byte(resp), &aiResp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if len(aiResp.Content) == 0 {
		t.Error("expected content in response")
	}
}

func TestAIRequestTypes(t *testing.T) {
	// Test that content types serialize correctly
	req := AIRequest{
		System: "Test system prompt",
		Messages: []ContentBlock{
			{Type: ContentTypeText, Text: "Hello"},
			{Type: ContentTypeImage, ImageRef: "test.png", MimeType: "image/png"},
			{Type: ContentTypeAudio, AudioRef: "test.wav", MimeType: "audio/wav"},
		},
		Context: map[string]interface{}{
			"generate_image": true,
		},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}

	var parsed AIRequest
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal request: %v", err)
	}

	if parsed.System != req.System {
		t.Errorf("system mismatch: got %q, want %q", parsed.System, req.System)
	}

	if len(parsed.Messages) != 3 {
		t.Errorf("message count mismatch: got %d, want 3", len(parsed.Messages))
	}

	if parsed.Messages[0].Type != ContentTypeText {
		t.Error("first message should be text")
	}

	if parsed.Messages[1].Type != ContentTypeImage {
		t.Error("second message should be image")
	}

	if parsed.Messages[2].Type != ContentTypeAudio {
		t.Error("third message should be audio")
	}
}

func TestAutoDetectProvider_NoKeys(t *testing.T) {
	// Save and clear env vars
	oldAnthropicKey := os.Getenv("ANTHROPIC_API_KEY")
	oldGoogleKey := os.Getenv("GOOGLE_API_KEY")
	oldGoogleProject := os.Getenv("GOOGLE_CLOUD_PROJECT")
	oldGcloudProject := os.Getenv("GCLOUD_PROJECT")

	os.Unsetenv("ANTHROPIC_API_KEY")
	os.Unsetenv("GOOGLE_API_KEY")
	os.Unsetenv("GOOGLE_CLOUD_PROJECT")
	os.Unsetenv("GCLOUD_PROJECT")

	defer func() {
		// Restore env vars
		if oldAnthropicKey != "" {
			os.Setenv("ANTHROPIC_API_KEY", oldAnthropicKey)
		}
		if oldGoogleKey != "" {
			os.Setenv("GOOGLE_API_KEY", oldGoogleKey)
		}
		if oldGoogleProject != "" {
			os.Setenv("GOOGLE_CLOUD_PROJECT", oldGoogleProject)
		}
		if oldGcloudProject != "" {
			os.Setenv("GCLOUD_PROJECT", oldGcloudProject)
		}
	}()

	ctx := context.Background()
	handler, err := NewAIHandlerFromEnv(ctx)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Should return stub handler
	_, isStub := handler.(*StubAIHandler)
	if !isStub {
		t.Error("expected stub handler when no API keys configured")
	}
}

func TestClaudeHandler_NoKey(t *testing.T) {
	_, err := NewClaudeAIHandler(ClaudeConfig{
		APIKey: "", // No key provided
	})

	// Should fail without clearing env var since user might have it set
	// Just test that the function exists and returns appropriate error type
	if err != nil && !strings.Contains(err.Error(), "ANTHROPIC_API_KEY") {
		t.Errorf("expected error about API key, got: %v", err)
	}
}

func TestGeminiHandler_NoAuth(t *testing.T) {
	// Save and clear env vars
	oldGoogleKey := os.Getenv("GOOGLE_API_KEY")
	oldGoogleProject := os.Getenv("GOOGLE_CLOUD_PROJECT")
	oldGcloudProject := os.Getenv("GCLOUD_PROJECT")

	os.Unsetenv("GOOGLE_API_KEY")
	os.Unsetenv("GOOGLE_CLOUD_PROJECT")
	os.Unsetenv("GCLOUD_PROJECT")

	defer func() {
		if oldGoogleKey != "" {
			os.Setenv("GOOGLE_API_KEY", oldGoogleKey)
		}
		if oldGoogleProject != "" {
			os.Setenv("GOOGLE_CLOUD_PROJECT", oldGoogleProject)
		}
		if oldGcloudProject != "" {
			os.Setenv("GCLOUD_PROJECT", oldGcloudProject)
		}
	}()

	ctx := context.Background()
	_, err := NewGeminiAIHandler(ctx, GeminiConfig{})

	if err == nil {
		t.Error("expected error when no auth configured")
	}

	if !strings.Contains(err.Error(), "no Gemini auth configured") {
		t.Errorf("expected auth error, got: %v", err)
	}
}

func TestChainAIHandler(t *testing.T) {
	stub1 := NewStubAIHandler()
	stub2 := NewStubAIHandler()

	chain := NewChainAIHandler(stub1, stub2)

	resp, err := chain.Call("test")
	if err != nil {
		t.Fatalf("chain handler error: %v", err)
	}

	var aiResp AIResponse
	if err := json.Unmarshal([]byte(resp), &aiResp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if len(aiResp.Content) == 0 {
		t.Error("expected content from chain handler")
	}
}

// Integration tests - only run if API keys are available
func TestClaudeHandler_Integration(t *testing.T) {
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		t.Skip("ANTHROPIC_API_KEY not set, skipping integration test")
	}

	handler, err := NewClaudeAIHandler(ClaudeConfig{})
	if err != nil {
		t.Fatalf("failed to create handler: %v", err)
	}

	req := AIRequest{
		Messages: []ContentBlock{
			{Type: ContentTypeText, Text: "Say 'hello' and nothing else."},
		},
	}
	reqJSON, _ := json.Marshal(req)

	resp, err := handler.Call(string(reqJSON))
	if err != nil {
		t.Fatalf("API call failed: %v", err)
	}

	var aiResp AIResponse
	if err := json.Unmarshal([]byte(resp), &aiResp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if aiResp.Error != "" {
		t.Errorf("API returned error: %s", aiResp.Error)
	}

	if len(aiResp.Content) == 0 {
		t.Error("expected content in response")
	}

	t.Logf("Claude response: %+v", aiResp.Content)
}

func TestGeminiHandler_Integration(t *testing.T) {
	hasVertexConfig := os.Getenv("GOOGLE_CLOUD_PROJECT") != "" || os.Getenv("GCLOUD_PROJECT") != ""
	hasAPIKey := os.Getenv("GOOGLE_API_KEY") != ""

	if !hasVertexConfig && !hasAPIKey {
		t.Skip("No Gemini auth configured, skipping integration test")
	}

	ctx := context.Background()
	handler, err := NewGeminiAIHandler(ctx, GeminiConfig{})
	if err != nil {
		t.Fatalf("failed to create handler: %v", err)
	}

	req := AIRequest{
		Messages: []ContentBlock{
			{Type: ContentTypeText, Text: "Say 'hello' and nothing else."},
		},
	}
	reqJSON, _ := json.Marshal(req)

	resp, err := handler.Call(string(reqJSON))
	if err != nil {
		t.Fatalf("API call failed: %v", err)
	}

	var aiResp AIResponse
	if err := json.Unmarshal([]byte(resp), &aiResp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if aiResp.Error != "" {
		t.Errorf("API returned error: %s", aiResp.Error)
	}

	if len(aiResp.Content) == 0 {
		t.Error("expected content in response")
	}

	t.Logf("Gemini response: %+v", aiResp.Content)
}
