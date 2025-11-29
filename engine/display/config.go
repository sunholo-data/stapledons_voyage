// Package display handles window configuration and display settings.
package display

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config holds display settings that can be persisted.
type Config struct {
	Width      int     `json:"width"`
	Height     int     `json:"height"`
	Fullscreen bool    `json:"fullscreen"`
	VSync      bool    `json:"vsync"`
	Scale      float64 `json:"scale"`
}

// DefaultConfig returns sensible default display settings.
func DefaultConfig() Config {
	return Config{
		Width:      1280,
		Height:     720,
		Fullscreen: false,
		VSync:      true,
		Scale:      1.0,
	}
}

// LoadConfig loads display configuration from a JSON file.
// Returns default config if file doesn't exist or is invalid.
func LoadConfig(path string) Config {
	cfg := DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		return cfg
	}

	if err := json.Unmarshal(data, &cfg); err != nil {
		return DefaultConfig()
	}

	// Validate and clamp values
	if cfg.Width < 640 {
		cfg.Width = 640
	}
	if cfg.Height < 480 {
		cfg.Height = 480
	}
	if cfg.Scale < 0.5 {
		cfg.Scale = 0.5
	}
	if cfg.Scale > 4.0 {
		cfg.Scale = 4.0
	}

	return cfg
}

// SaveConfig saves the configuration to a JSON file.
func SaveConfig(path string, cfg Config) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
