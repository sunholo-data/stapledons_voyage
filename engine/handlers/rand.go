// Package handlers provides effect handler implementations for the game.
package handlers

import (
	"math/rand"
)

// DefaultRandHandler provides a standard random number generator.
// Uses math/rand with optional seeding for deterministic behavior.
type DefaultRandHandler struct {
	rng *rand.Rand
}

// NewDefaultRandHandler creates a Rand handler with a new source.
func NewDefaultRandHandler() *DefaultRandHandler {
	return &DefaultRandHandler{
		rng: rand.New(rand.NewSource(0)),
	}
}

// NewSeededRandHandler creates a Rand handler with a specific seed.
func NewSeededRandHandler(seed int64) *DefaultRandHandler {
	return &DefaultRandHandler{
		rng: rand.New(rand.NewSource(seed)),
	}
}

// RandInt returns a random integer in [min, max].
func (h *DefaultRandHandler) RandInt(min, max int64) int64 {
	if min >= max {
		return min
	}
	return min + h.rng.Int63n(max-min+1)
}

// RandFloat returns a random float in [min, max).
func (h *DefaultRandHandler) RandFloat(min, max float64) float64 {
	if min >= max {
		return min
	}
	return min + h.rng.Float64()*(max-min)
}

// RandBool returns a random boolean.
func (h *DefaultRandHandler) RandBool() bool {
	return h.rng.Intn(2) == 1
}

// SetSeed sets the random seed for deterministic behavior.
func (h *DefaultRandHandler) SetSeed(seed int64) {
	h.rng = rand.New(rand.NewSource(seed))
}
