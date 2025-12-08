package view

import "math"

// EasingFunc transforms a linear progress value (0-1) into an eased value.
type EasingFunc func(t float64) float64

// Predefined easing functions.
var (
	// Linear provides no easing - constant velocity.
	Linear EasingFunc = func(t float64) float64 {
		return t
	}

	// EaseInQuad starts slow and accelerates.
	EaseInQuad EasingFunc = func(t float64) float64 {
		return t * t
	}

	// EaseOutQuad starts fast and decelerates.
	EaseOutQuad EasingFunc = func(t float64) float64 {
		return 1 - (1-t)*(1-t)
	}

	// EaseInOutQuad accelerates then decelerates.
	EaseInOutQuad EasingFunc = func(t float64) float64 {
		if t < 0.5 {
			return 2 * t * t
		}
		return 1 - math.Pow(-2*t+2, 2)/2
	}

	// EaseInCubic starts slow and accelerates more aggressively.
	EaseInCubic EasingFunc = func(t float64) float64 {
		return t * t * t
	}

	// EaseOutCubic starts fast and decelerates smoothly.
	EaseOutCubic EasingFunc = func(t float64) float64 {
		return 1 - math.Pow(1-t, 3)
	}

	// EaseInOutCubic smooth acceleration and deceleration.
	EaseInOutCubic EasingFunc = func(t float64) float64 {
		if t < 0.5 {
			return 4 * t * t * t
		}
		return 1 - math.Pow(-2*t+2, 3)/2
	}

	// EaseInExpo exponential acceleration.
	EaseInExpo EasingFunc = func(t float64) float64 {
		if t == 0 {
			return 0
		}
		return math.Pow(2, 10*(t-1))
	}

	// EaseOutExpo exponential deceleration.
	EaseOutExpo EasingFunc = func(t float64) float64 {
		if t == 1 {
			return 1
		}
		return 1 - math.Pow(2, -10*t)
	}

	// EaseInOutExpo exponential acceleration and deceleration.
	EaseInOutExpo EasingFunc = func(t float64) float64 {
		if t == 0 {
			return 0
		}
		if t == 1 {
			return 1
		}
		if t < 0.5 {
			return math.Pow(2, 20*t-10) / 2
		}
		return (2 - math.Pow(2, -20*t+10)) / 2
	}

	// EaseInSine gentle sinusoidal acceleration.
	EaseInSine EasingFunc = func(t float64) float64 {
		return 1 - math.Cos(t*math.Pi/2)
	}

	// EaseOutSine gentle sinusoidal deceleration.
	EaseOutSine EasingFunc = func(t float64) float64 {
		return math.Sin(t * math.Pi / 2)
	}

	// EaseInOutSine gentle sinusoidal acceleration and deceleration.
	EaseInOutSine EasingFunc = func(t float64) float64 {
		return -(math.Cos(math.Pi*t) - 1) / 2
	}

	// EaseInBack overshoots slightly then accelerates.
	EaseInBack EasingFunc = func(t float64) float64 {
		const c1 = 1.70158
		const c3 = c1 + 1
		return c3*t*t*t - c1*t*t
	}

	// EaseOutBack decelerates then overshoots slightly.
	EaseOutBack EasingFunc = func(t float64) float64 {
		const c1 = 1.70158
		const c3 = c1 + 1
		return 1 + c3*math.Pow(t-1, 3) + c1*math.Pow(t-1, 2)
	}

	// EaseInOutBack overshoots on both ends.
	EaseInOutBack EasingFunc = func(t float64) float64 {
		const c1 = 1.70158
		const c2 = c1 * 1.525
		if t < 0.5 {
			return (math.Pow(2*t, 2) * ((c2+1)*2*t - c2)) / 2
		}
		return (math.Pow(2*t-2, 2)*((c2+1)*(t*2-2)+c2) + 2) / 2
	}

	// EaseOutBounce bounces at the end.
	EaseOutBounce EasingFunc = func(t float64) float64 {
		const n1 = 7.5625
		const d1 = 2.75

		if t < 1/d1 {
			return n1 * t * t
		} else if t < 2/d1 {
			t -= 1.5 / d1
			return n1*t*t + 0.75
		} else if t < 2.5/d1 {
			t -= 2.25 / d1
			return n1*t*t + 0.9375
		} else {
			t -= 2.625 / d1
			return n1*t*t + 0.984375
		}
	}

	// EaseInBounce bounces at the start.
	EaseInBounce EasingFunc = func(t float64) float64 {
		return 1 - EaseOutBounce(1-t)
	}

	// EaseInOutBounce bounces at both ends.
	EaseInOutBounce EasingFunc = func(t float64) float64 {
		if t < 0.5 {
			return (1 - EaseOutBounce(1-2*t)) / 2
		}
		return (1 + EaseOutBounce(2*t-1)) / 2
	}
)

// Clamp constrains a value to the range [0, 1].
func Clamp(t float64) float64 {
	if t < 0 {
		return 0
	}
	if t > 1 {
		return 1
	}
	return t
}

// Lerp performs linear interpolation between a and b.
func Lerp(a, b, t float64) float64 {
	return a + (b-a)*t
}

// LerpWithEasing performs interpolation with an easing function.
func LerpWithEasing(a, b, t float64, ease EasingFunc) float64 {
	return a + (b-a)*ease(Clamp(t))
}

// TransitionConfig holds configuration for a view transition.
type TransitionConfig struct {
	Duration float64          // Duration in seconds
	Effect   TransitionEffect // Visual effect type
	Easing   EasingFunc       // Easing function to apply
}

// DefaultTransitionConfig returns a default transition configuration.
func DefaultTransitionConfig() TransitionConfig {
	return TransitionConfig{
		Duration: 0.5,
		Effect:   TransitionFade,
		Easing:   EaseInOutQuad,
	}
}

// QuickFade returns a quick fade transition config.
func QuickFade() TransitionConfig {
	return TransitionConfig{
		Duration: 0.3,
		Effect:   TransitionFade,
		Easing:   EaseOutQuad,
	}
}

// SlowCrossfade returns a slow crossfade transition config.
func SlowCrossfade() TransitionConfig {
	return TransitionConfig{
		Duration: 1.0,
		Effect:   TransitionCrossfade,
		Easing:   EaseInOutSine,
	}
}
