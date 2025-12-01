// Package render provides animation management for sprite-based entities.
package render

// AnimationDef defines animation sequences for a sprite.
type AnimationDef struct {
	Animations  map[string]AnimationSeq // Named animation sequences
	FrameWidth  int                     // Width of each frame in pixels
	FrameHeight int                     // Height of each frame in pixels
}

// AnimationSeq defines a single animation sequence.
type AnimationSeq struct {
	StartFrame int     // First frame index in sprite sheet
	FrameCount int     // Number of frames in this animation
	FPS        float64 // Frames per second (0 = static)
}

// AnimationState tracks the current animation state for an entity.
type AnimationState struct {
	SpriteID     int
	CurrentAnim  string
	CurrentFrame int
	FrameTime    float64 // Seconds since last frame change
	Playing      bool
}

// AnimationManager handles animation state and frame advancement for entities.
type AnimationManager struct {
	states map[string]*AnimationState // Keyed by entity ID
	defs   map[int]*AnimationDef      // Keyed by sprite ID
}

// NewAnimationManager creates a new animation manager.
func NewAnimationManager() *AnimationManager {
	return &AnimationManager{
		states: make(map[string]*AnimationState),
		defs:   make(map[int]*AnimationDef),
	}
}

// RegisterSprite registers animation definitions for a sprite ID.
func (am *AnimationManager) RegisterSprite(spriteID int, def *AnimationDef) {
	am.defs[spriteID] = def
}

// HasAnimations returns true if the sprite has animation definitions.
func (am *AnimationManager) HasAnimations(spriteID int) bool {
	def, ok := am.defs[spriteID]
	return ok && def != nil && len(def.Animations) > 0
}

// GetFrameDimensions returns the frame width and height for a sprite, or 0,0 if not animated.
func (am *AnimationManager) GetFrameDimensions(spriteID int) (width, height int) {
	if def, ok := am.defs[spriteID]; ok {
		return def.FrameWidth, def.FrameHeight
	}
	return 0, 0
}

// Update advances all animation states by the given delta time.
func (am *AnimationManager) Update(dt float64) {
	for _, state := range am.states {
		if !state.Playing {
			continue
		}

		def, ok := am.defs[state.SpriteID]
		if !ok {
			continue
		}

		anim, ok := def.Animations[state.CurrentAnim]
		if !ok || anim.FPS == 0 {
			continue // Static animation or not found
		}

		state.FrameTime += dt
		frameDuration := 1.0 / anim.FPS

		// Advance frames if enough time has passed
		for state.FrameTime >= frameDuration {
			state.FrameTime -= frameDuration
			state.CurrentFrame++
			if state.CurrentFrame >= anim.FrameCount {
				state.CurrentFrame = 0 // Loop
			}
		}
	}
}

// GetFrame returns the current frame index for an entity's animation.
// Returns 0 if the entity has no animation state or no animation definitions.
func (am *AnimationManager) GetFrame(entityID string, spriteID int, animName string) int {
	// Get or create animation state for this entity
	state, ok := am.states[entityID]
	if !ok {
		// Create new state
		state = &AnimationState{
			SpriteID:     spriteID,
			CurrentAnim:  animName,
			CurrentFrame: 0,
			FrameTime:    0,
			Playing:      true,
		}
		am.states[entityID] = state
	}

	// Check if animation changed
	if state.CurrentAnim != animName || state.SpriteID != spriteID {
		state.SpriteID = spriteID
		state.CurrentAnim = animName
		state.CurrentFrame = 0
		state.FrameTime = 0
		state.Playing = true
	}

	// Get animation definition
	def, ok := am.defs[spriteID]
	if !ok {
		return 0
	}

	anim, ok := def.Animations[animName]
	if !ok {
		// Try "idle" as fallback
		anim, ok = def.Animations["idle"]
		if !ok {
			return 0
		}
	}

	// Return absolute frame index in sprite sheet
	return anim.StartFrame + state.CurrentFrame
}

// SetAnimation sets the current animation for an entity.
func (am *AnimationManager) SetAnimation(entityID string, animName string) {
	if state, ok := am.states[entityID]; ok {
		if state.CurrentAnim != animName {
			state.CurrentAnim = animName
			state.CurrentFrame = 0
			state.FrameTime = 0
		}
	}
}

// RemoveEntity removes animation state for an entity (when entity is destroyed).
func (am *AnimationManager) RemoveEntity(entityID string) {
	delete(am.states, entityID)
}

// Clear removes all animation states.
func (am *AnimationManager) Clear() {
	am.states = make(map[string]*AnimationState)
}
