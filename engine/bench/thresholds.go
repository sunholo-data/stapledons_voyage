// Package bench provides performance benchmarking and threshold checking.
package bench

import (
	"time"
)

// Thresholds define performance budgets for the game.
// These are checked during CI and perf runs.
type Thresholds struct {
	// StepMax is the maximum time allowed for a single Step() call.
	// For 60 FPS, the total frame budget is 16.67ms. Step should use
	// only a portion of this to leave room for rendering.
	StepMax time.Duration `json:"step_max_ns"`

	// InitWorldMax is the maximum time for world initialization.
	// This is a one-time cost at game start.
	InitWorldMax time.Duration `json:"init_world_max_ns"`

	// Step100Max is the maximum time for 100 consecutive steps.
	// Used to detect performance degradation over time.
	Step100Max time.Duration `json:"step_100_max_ns"`

	// FrameTimeMax is the maximum time for a complete frame
	// (Step + Render). For 60 FPS, this must be under 16.67ms.
	FrameTimeMax time.Duration `json:"frame_time_max_ns"`
}

// DefaultThresholds returns conservative performance budgets.
// These are tuned for 60 FPS gameplay on modest hardware.
func DefaultThresholds() Thresholds {
	return Thresholds{
		StepMax:      5 * time.Millisecond,   // 5ms - leaves 11ms for rendering
		InitWorldMax: 100 * time.Millisecond, // 100ms - one-time cost
		Step100Max:   500 * time.Millisecond, // 500ms for 100 steps (5ms avg)
		FrameTimeMax: 16 * time.Millisecond,  // 16ms for 60 FPS
	}
}

// BenchResult holds the result of a single benchmark run.
type BenchResult struct {
	Name      string        `json:"name"`
	Avg       time.Duration `json:"avg_ns"`
	Min       time.Duration `json:"min_ns"`
	Max       time.Duration `json:"max_ns"`
	P50       time.Duration `json:"p50_ns"`
	P95       time.Duration `json:"p95_ns"`
	P99       time.Duration `json:"p99_ns"`
	Total     time.Duration `json:"total_ns"`
	Ops       int           `json:"ops"`
	Passed    bool          `json:"passed"`
	Threshold time.Duration `json:"threshold_ns"`
	Message   string        `json:"message,omitempty"`
}

// PerfReport holds all benchmark results with threshold checking.
type PerfReport struct {
	Timestamp  string        `json:"timestamp"`
	Thresholds Thresholds    `json:"thresholds"`
	Results    []BenchResult `json:"results"`
	AllPassed  bool          `json:"all_passed"`
	FrameTimes *FrameStats   `json:"frame_times,omitempty"`
}

// FrameStats holds frame timing statistics from visual tests.
type FrameStats struct {
	Count   int           `json:"count"`
	Avg     time.Duration `json:"avg_ns"`
	Min     time.Duration `json:"min_ns"`
	Max     time.Duration `json:"max_ns"`
	P50     time.Duration `json:"p50_ns"`
	P95     time.Duration `json:"p95_ns"`
	P99     time.Duration `json:"p99_ns"`
	Dropped int           `json:"dropped_frames"` // Frames exceeding threshold
}
