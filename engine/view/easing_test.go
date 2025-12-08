package view

import (
	"math"
	"testing"
)

// tolerance for float comparisons
const epsilon = 0.0001

func TestEasingEndpoints(t *testing.T) {
	// All easing functions should return 0 for t=0 and 1 for t=1
	easings := []struct {
		name string
		fn   EasingFunc
	}{
		{"Linear", Linear},
		{"EaseInQuad", EaseInQuad},
		{"EaseOutQuad", EaseOutQuad},
		{"EaseInOutQuad", EaseInOutQuad},
		{"EaseInCubic", EaseInCubic},
		{"EaseOutCubic", EaseOutCubic},
		{"EaseInOutCubic", EaseInOutCubic},
		{"EaseInExpo", EaseInExpo},
		{"EaseOutExpo", EaseOutExpo},
		{"EaseInOutExpo", EaseInOutExpo},
		{"EaseInSine", EaseInSine},
		{"EaseOutSine", EaseOutSine},
		{"EaseInOutSine", EaseInOutSine},
		{"EaseInBack", EaseInBack},
		{"EaseOutBack", EaseOutBack},
		{"EaseInOutBack", EaseInOutBack},
		{"EaseOutBounce", EaseOutBounce},
		{"EaseInBounce", EaseInBounce},
		{"EaseInOutBounce", EaseInOutBounce},
	}

	for _, e := range easings {
		t.Run(e.name, func(t *testing.T) {
			v0 := e.fn(0)
			if math.Abs(v0) > epsilon {
				t.Errorf("%s(0) = %v, want ~0", e.name, v0)
			}

			v1 := e.fn(1)
			if math.Abs(v1-1) > epsilon {
				t.Errorf("%s(1) = %v, want ~1", e.name, v1)
			}
		})
	}
}

func TestLinear(t *testing.T) {
	tests := []struct {
		in, want float64
	}{
		{0.0, 0.0},
		{0.25, 0.25},
		{0.5, 0.5},
		{0.75, 0.75},
		{1.0, 1.0},
	}

	for _, tt := range tests {
		got := Linear(tt.in)
		if math.Abs(got-tt.want) > epsilon {
			t.Errorf("Linear(%v) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

func TestEaseInOutQuadSymmetry(t *testing.T) {
	// EaseInOutQuad should be symmetric around t=0.5
	v025 := EaseInOutQuad(0.25)
	v075 := EaseInOutQuad(0.75)

	// v(0.25) + v(0.75) should equal 1
	sum := v025 + v075
	if math.Abs(sum-1) > epsilon {
		t.Errorf("EaseInOutQuad(0.25) + EaseInOutQuad(0.75) = %v, want 1", sum)
	}

	// Midpoint should be 0.5
	v05 := EaseInOutQuad(0.5)
	if math.Abs(v05-0.5) > epsilon {
		t.Errorf("EaseInOutQuad(0.5) = %v, want 0.5", v05)
	}
}

func TestClamp(t *testing.T) {
	tests := []struct {
		in, want float64
	}{
		{-0.5, 0.0},
		{0.0, 0.0},
		{0.5, 0.5},
		{1.0, 1.0},
		{1.5, 1.0},
	}

	for _, tt := range tests {
		got := Clamp(tt.in)
		if got != tt.want {
			t.Errorf("Clamp(%v) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

func TestLerp(t *testing.T) {
	tests := []struct {
		a, b, t, want float64
	}{
		{0, 100, 0.0, 0},
		{0, 100, 0.5, 50},
		{0, 100, 1.0, 100},
		{50, 100, 0.5, 75},
	}

	for _, tt := range tests {
		got := Lerp(tt.a, tt.b, tt.t)
		if math.Abs(got-tt.want) > epsilon {
			t.Errorf("Lerp(%v, %v, %v) = %v, want %v", tt.a, tt.b, tt.t, got, tt.want)
		}
	}
}

func TestLerpWithEasing(t *testing.T) {
	// With linear easing, should equal regular Lerp
	got := LerpWithEasing(0, 100, 0.5, Linear)
	if math.Abs(got-50) > epsilon {
		t.Errorf("LerpWithEasing(0, 100, 0.5, Linear) = %v, want 50", got)
	}

	// With EaseInQuad at t=0.5, eased value is 0.25
	// So interpolated value should be 25
	got = LerpWithEasing(0, 100, 0.5, EaseInQuad)
	if math.Abs(got-25) > epsilon {
		t.Errorf("LerpWithEasing(0, 100, 0.5, EaseInQuad) = %v, want 25", got)
	}

	// Should clamp out-of-range values
	got = LerpWithEasing(0, 100, 1.5, Linear)
	if math.Abs(got-100) > epsilon {
		t.Errorf("LerpWithEasing(0, 100, 1.5, Linear) = %v, want 100 (clamped)", got)
	}
}

func TestDefaultTransitionConfig(t *testing.T) {
	config := DefaultTransitionConfig()

	if config.Duration != 0.5 {
		t.Errorf("Duration = %v, want 0.5", config.Duration)
	}
	if config.Effect != TransitionFade {
		t.Errorf("Effect = %v, want TransitionFade", config.Effect)
	}
	if config.Easing == nil {
		t.Error("Easing should not be nil")
	}
}

func TestQuickFade(t *testing.T) {
	config := QuickFade()

	if config.Duration != 0.3 {
		t.Errorf("Duration = %v, want 0.3", config.Duration)
	}
	if config.Effect != TransitionFade {
		t.Errorf("Effect = %v, want TransitionFade", config.Effect)
	}
}

func TestSlowCrossfade(t *testing.T) {
	config := SlowCrossfade()

	if config.Duration != 1.0 {
		t.Errorf("Duration = %v, want 1.0", config.Duration)
	}
	if config.Effect != TransitionCrossfade {
		t.Errorf("Effect = %v, want TransitionCrossfade", config.Effect)
	}
}

func TestEasingMonotonicity(t *testing.T) {
	// Standard easing functions (non-back, non-bounce) should be monotonically increasing
	monotonic := []struct {
		name string
		fn   EasingFunc
	}{
		{"Linear", Linear},
		{"EaseInQuad", EaseInQuad},
		{"EaseOutQuad", EaseOutQuad},
		{"EaseInOutQuad", EaseInOutQuad},
		{"EaseInCubic", EaseInCubic},
		{"EaseOutCubic", EaseOutCubic},
		{"EaseInOutCubic", EaseInOutCubic},
		{"EaseInSine", EaseInSine},
		{"EaseOutSine", EaseOutSine},
		{"EaseInOutSine", EaseInOutSine},
	}

	for _, e := range monotonic {
		t.Run(e.name, func(t *testing.T) {
			prev := e.fn(0)
			for i := 1; i <= 100; i++ {
				x := float64(i) / 100.0
				curr := e.fn(x)
				if curr < prev-epsilon {
					t.Errorf("%s is not monotonic: f(%v) = %v < f(%v) = %v",
						e.name, x, curr, x-0.01, prev)
				}
				prev = curr
			}
		})
	}
}
