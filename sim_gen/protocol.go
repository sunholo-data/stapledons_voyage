package sim_gen

// MouseState captures mouse position and button state
type MouseState struct {
	X       float64
	Y       float64
	Buttons []int
}

// KeyEvent represents a keyboard event
type KeyEvent struct {
	Key  int
	Kind string // "down" or "up"
}

// FrameInput is passed from engine to simulation each frame
type FrameInput struct {
	Mouse            MouseState
	Keys             []KeyEvent
	ClickedThisFrame bool         // True if left mouse button was just pressed
	WorldMouseX      float64      // Mouse X in world coordinates (after camera transform)
	WorldMouseY      float64      // Mouse Y in world coordinates (after camera transform)
	ActionRequested  PlayerAction // Action triggered by keyboard (I=inspect, B=build, X=clear)
}

// Camera represents the viewport position and zoom level
type Camera struct {
	X    float64 // World position X (center of view)
	Y    float64 // World position Y (center of view)
	Zoom float64 // Zoom factor (1.0 = normal, 2.0 = zoomed in)
}

// FrameOutput is returned from simulation to engine each frame
type FrameOutput struct {
	Draw   []DrawCmd
	Sounds []int
	Debug  []string
	Camera Camera // Camera state for this frame
}
