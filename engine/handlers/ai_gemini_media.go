// Package handlers provides effect handler implementations for the game.
package handlers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"google.golang.org/genai"
)

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
