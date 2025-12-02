// Package scenario provides test scenario execution for automated testing.
package scenario

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"stapledons_voyage/sim_gen"
)

// Scenario defines a test scenario with input events and capture points.
type Scenario struct {
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Seed        int64        `json:"seed"`
	Camera      CameraConfig `json:"camera"`
	Events      []Event      `json:"events"`
	TestMode    bool         `json:"test_mode,omitempty"` // Strip UI for golden file testing
}

// CameraConfig defines initial camera position.
type CameraConfig struct {
	X    float64 `json:"x"`
	Y    float64 `json:"y"`
	Zoom float64 `json:"zoom"`
}

// Event represents a timed input event or capture point.
type Event struct {
	Frame   int    `json:"frame"`
	Key     string `json:"key,omitempty"`     // Key name (e.g., "W", "S", "I")
	Action  string `json:"action,omitempty"`  // "down", "up", or "press"
	Click   *Click `json:"click,omitempty"`   // Mouse click
	Capture string `json:"capture,omitempty"` // Screenshot filename to capture
}

// Click represents a mouse click event.
type Click struct {
	X      int    `json:"x"`
	Y      int    `json:"y"`
	Button string `json:"button"` // "left", "right", "middle"
}

// LoadScenario loads a scenario from a JSON file.
func LoadScenario(path string) (*Scenario, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read scenario file: %w", err)
	}

	var s Scenario
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("failed to parse scenario JSON: %w", err)
	}

	// Set defaults
	if s.Camera.Zoom == 0 {
		s.Camera.Zoom = 1.0
	}

	return &s, nil
}

// FindScenario looks for a scenario file in the scenarios directory.
func FindScenario(name string) (string, error) {
	// Try direct path first
	if _, err := os.Stat(name); err == nil {
		return name, nil
	}

	// Try scenarios directory
	path := filepath.Join("scenarios", name+".json")
	if _, err := os.Stat(path); err == nil {
		return path, nil
	}

	return "", fmt.Errorf("scenario not found: %s", name)
}

// KeyNameToCode converts key names to Ebiten key codes.
func KeyNameToCode(name string) int {
	// Ebiten key codes (from ebiten.Key* constants)
	keys := map[string]int{
		"A": 0, "B": 1, "C": 2, "D": 3, "E": 4, "F": 5, "G": 6, "H": 7,
		"I": 8, "J": 9, "K": 10, "L": 11, "M": 12, "N": 13, "O": 14, "P": 15,
		"Q": 16, "R": 17, "S": 18, "T": 19, "U": 20, "V": 21, "W": 22, "X": 23,
		"Y": 24, "Z": 25,
		"ArrowDown": 28, "ArrowLeft": 29, "ArrowRight": 30, "ArrowUp": 31,
		"Down": 28, "Left": 29, "Right": 30, "Up": 31,
	}

	if code, ok := keys[name]; ok {
		return code
	}
	return -1
}

// BuildFrameInput creates a FrameInput from active keys and pending clicks.
// pressedKeys are keys that were just pressed this frame (for mode switching etc).
func BuildFrameInput(activeKeys map[int]bool, pressedKeys map[int]bool, pendingClick *Click, world sim_gen.World, testMode bool) sim_gen.FrameInput {
	var keys []sim_gen.KeyEvent
	// Add "down" events for all active keys
	for code := range activeKeys {
		keys = append(keys, sim_gen.KeyEvent{
			Key:  code,
			Kind: "down",
		})
	}
	// Add "pressed" events for keys just pressed this frame
	for code := range pressedKeys {
		keys = append(keys, sim_gen.KeyEvent{
			Key:  code,
			Kind: "pressed",
		})
	}

	var clicked bool
	var worldX, worldY float64

	if pendingClick != nil && pendingClick.Button == "left" {
		clicked = true
		// Convert screen coords to world coords (simplified)
		// In real use, we'd need proper camera transform
		worldX = float64(pendingClick.X)
		worldY = float64(pendingClick.Y)
	}

	var action sim_gen.PlayerAction = sim_gen.ActionNone{}
	// Check for action keys
	if activeKeys[8] { // I key
		action = sim_gen.ActionInspect{}
	} else if activeKeys[1] { // B key
		action = sim_gen.ActionBuild{StructureType: sim_gen.StructureHouse}
	} else if activeKeys[23] { // X key
		action = sim_gen.ActionClear{}
	}

	return sim_gen.FrameInput{
		Mouse: sim_gen.MouseState{
			X: worldX,
			Y: worldY,
		},
		Keys:             keys,
		ClickedThisFrame: clicked,
		WorldMouseX:      worldX,
		WorldMouseY:      worldY,
		ActionRequested:  action,
		TestMode:         testMode,
	}
}
