// Package handlers provides effect handler implementations for the game.
package handlers

// EbitenClockHandler provides frame timing for the game.
// Updated each frame by the game loop before calling Step.
type EbitenClockHandler struct {
	deltaTime  float64 // Seconds since last frame
	totalTime  float64 // Total game time in seconds
	frameCount int64   // Current frame number
}

// NewEbitenClockHandler creates a new clock handler.
func NewEbitenClockHandler() *EbitenClockHandler {
	return &EbitenClockHandler{}
}

// Update advances the clock by one frame.
// Called by the game loop before sim_gen.Step().
// dt is the time since the last frame in seconds.
func (h *EbitenClockHandler) Update(dt float64) {
	h.deltaTime = dt
	h.totalTime += dt
	h.frameCount++
}

// DeltaTime returns seconds since last frame.
func (h *EbitenClockHandler) DeltaTime() float64 {
	return h.deltaTime
}

// TotalTime returns total game time in seconds.
func (h *EbitenClockHandler) TotalTime() float64 {
	return h.totalTime
}

// FrameCount returns the current frame number.
func (h *EbitenClockHandler) FrameCount() int64 {
	return h.frameCount
}

// Reset clears all timing values (for testing or restart).
func (h *EbitenClockHandler) Reset() {
	h.deltaTime = 0
	h.totalTime = 0
	h.frameCount = 0
}
