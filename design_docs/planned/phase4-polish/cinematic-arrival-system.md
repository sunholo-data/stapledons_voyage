# Cinematic Arrival System

## Status
- Status: Planned
- Priority: P1 (blocks Phase 2 of arrival-sequence sprint)
- Complexity: Medium-High
- Estimated Effort: 2-3 weeks

## Narrative Context

**The AI Pilot's Tour**: When arriving at a star system after decades of relativistic travel, the ship's AI (Archive) deliberately decelerates and pauses at each major body. This serves multiple purposes:

1. **Crew welfare** - After years in transit, the crew needs to see their destination
2. **Navigation verification** - Visual confirmation of orbital positions
3. **Scientific observation** - Recording planetary conditions after centuries of change
4. **Player experience** - The sightseeing tour IS the arrival sequence

This narrative justifies the lower velocities (0.1c-0.5c range) that make SR effects visible without overwhelming the visuals.

## Physics-Accurate Design Principles

### What We Render (Real Physics)

| Effect | Physics Basis | Implementation |
|--------|---------------|----------------|
| **SR Doppler Shift** | Light wavelength changes with relative velocity | Existing `sr.go` shader |
| **SR Aberration** | Stars bunch forward at high velocity | Existing `sr.go` shader |
| **GR Lensing** | Light bends around massive objects | Existing `gr.go` shader |
| **GR Redshift** | Light escaping gravity wells shifts red | Existing `gr.go` shader |
| **Time Dilation Display** | γ factor shows proper time vs coordinate time | HUD element |

### What We DON'T Render (Hollywood Conventions)

| Rejected Effect | Why It's Wrong |
|-----------------|----------------|
| ~~Star Streaks~~ | Stars are too distant - you see aberration, not motion blur |
| ~~Radial Motion Blur~~ | SR effects ARE the visual change, blur is redundant |
| ~~Warp Tunnel~~ | No physical basis - just looks "spacey" |
| ~~Engine Glow Behind~~ | No medium to illuminate in vacuum |

## Black Hole Emergence

The ship uses a **gravitational slingshot maneuver** around a stellar-mass black hole to achieve relativistic velocities. The "emergence" sequence depicts:

1. **Close approach** - Ship at ~3 Schwarzschild radii (photon sphere region)
2. **Maximum GR effects** - Extreme lensing, gravitational redshift
3. **Slingshot acceleration** - Using the black hole's gravity well
4. **Escape trajectory** - GR effects diminish as distance increases

This is physically plausible - spacecraft using black hole gravity assists is theoretically sound, though the engineering challenges are immense.

## Arrival Sequence Flow

```
[Black Hole Emergence]
    ↓ Ship escapes photon sphere region
    ↓ GR lensing diminishes over ~10 seconds
    ↓ Camera stabilizes from tumble

[High-Speed Transit] @ 0.5c
    ↓ Strong SR aberration + Doppler
    ↓ Stars compressed forward
    ↓ Blue-shifted view ahead

[Saturn Approach] - AI slows to 0.3c
    ↓ "Crew requested visual confirmation"
    ↓ Saturn grows in view
    ↓ Moderate SR effects visible
    ↓ Pause for observation

[Jupiter Flyby] @ 0.2c
    ↓ Mild SR effects
    ↓ Jupiter dominates view

[Mars Pass] @ 0.1c
    ↓ Subtle SR effects
    ↓ Red planet visible

[Earth Arrival] @ 0.0c
    ↓ No SR effects
    ↓ Home in view
    ↓ "Welcome home. [X] years have passed."
```

## Current Engine Capabilities

### Working (Use These)

| Component | Status | Notes |
|-----------|--------|-------|
| SR Doppler/Aberration shader | Working | Physically accurate |
| GR Lensing shader | Working | Physically accurate |
| GR Redshift shader | Working | Physically accurate |
| Bloom effect | Working | Good for star glow |
| RectRGBA/CircleRGBA | Working | Screen-space rendering |
| Camera transform | Working | Pan, zoom |

### Needed (Build These)

| Component | Purpose | Effort |
|-----------|---------|--------|
| **Smooth Camera Paths** | Easing functions for transitions | 1-2 days |
| **Camera Tumble/Stabilization** | Black hole emergence drama | 1 day |
| **Planet Scaling Animation** | Smooth approach effect | 1 day |
| **Parallax Star Layers** | Depth during transit | 2 days |
| **Raymarched Planet Spheres** | True 3D planets (optional) | 3-5 days |

## Technical Implementation

### Phase 1: Smooth Transitions (Required)

#### Easing System
```go
// engine/cinematic/easing.go
type EasingFunc func(t float64) float64

var (
    Linear    EasingFunc = func(t float64) float64 { return t }
    EaseInOut EasingFunc = func(t float64) float64 {
        return t * t * (3 - 2*t)
    }
    EaseOutExpo EasingFunc = func(t float64) float64 {
        if t == 1 { return 1 }
        return 1 - math.Pow(2, -10*t)
    }
)
```

#### Camera Tumble
```go
// engine/cinematic/camera.go
type CameraTumble struct {
    RotationX, RotationY, RotationZ float64
    Damping                          float64 // 0.95 = slow stabilize
    Active                           bool
}

func (c *CameraTumble) Update(dt float64) {
    c.RotationX *= c.Damping
    c.RotationY *= c.Damping
    c.RotationZ *= c.Damping
    if math.Abs(c.RotationX) < 0.001 {
        c.Active = false
    }
}
```

### Phase 2: Planet Rendering (Optional Upgrade)

#### Raymarched Sphere Shader
```glsl
// Render planet as 3D sphere in fragment shader
float sphereSDF(vec3 p, vec3 center, float radius) {
    return length(p - center) - radius;
}

// With texture mapping for surface detail
vec2 sphereUV(vec3 p, vec3 center) {
    vec3 d = normalize(p - center);
    float u = 0.5 + atan(d.z, d.x) / (2.0 * PI);
    float v = 0.5 - asin(d.y) / PI;
    return vec2(u, v);
}
```

This gives true 3D spheres with:
- Proper lighting (terminator line)
- Rotation animation
- Atmosphere glow (Fresnel effect)
- No pre-rendered sprite sheets needed

## Velocity Schedule (Narrative Justified)

| Phase | Velocity | γ (Lorentz) | Visual Effect | Narrative |
|-------|----------|-------------|---------------|-----------|
| Black hole escape | 0.5c | 1.15 | Strong GR→SR transition | Slingshot acceleration |
| High-speed transit | 0.45c | 1.12 | Noticeable aberration | Cruising speed |
| Saturn approach | 0.3c | 1.05 | Moderate blueshift | "Slowing for observation" |
| Jupiter flyby | 0.2c | 1.02 | Mild effects | "Visual confirmation" |
| Mars pass | 0.1c | 1.005 | Subtle shift | "Approaching inner system" |
| Earth arrival | 0.0c | 1.00 | None | "Welcome home" |

**Note**: These velocities are ~10x lower than "realistic" interstellar travel (0.9c+) but show the same physics at visible intensities.

## Success Criteria

- [ ] Black hole emergence feels dramatic (GR lensing, camera tumble)
- [ ] SR effects visible but not overwhelming at each velocity
- [ ] Smooth transitions between phases (no jarring cuts)
- [ ] Planets approach with sense of scale and depth
- [ ] Time dilation displayed accurately (years passed)
- [ ] Archive dialogue explains the sightseeing stops
- [ ] Runs at 60fps throughout

## Game Vision Alignment

| Pillar | Score | Rationale |
|--------|-------|-----------|
| Time Dilation Consequence | +++ | This IS the moment of truth |
| Civilization Simulation | ++ | "What will we find after centuries?" |
| Philosophical Depth | +++ | Wonder, mortality, homecoming |
| Ship & Crew Life | +++ | Crew's first view in years |
| Legacy Impact | ++ | Sets up what's changed |
| **Hard Sci-Fi Authenticity** | **+++** | Physics-accurate effects |

## File Structure

```
engine/
├── cinematic/
│   ├── easing.go          # Easing functions
│   ├── camera_motion.go   # Tumble, stabilization, paths
│   └── sequence.go        # Keyframe-based timing
└── shader/
    └── planet_sphere.kage # Optional: raymarched 3D planets
```

## Open Questions

1. **Arrival skippable?** - After first playthrough only?
2. **Variable length?** - Shorter if returning to visited system?
3. **Audio?** - Archive narration during stops?
4. **Planet detail?** - Sprite sheets vs raymarched spheres?

## Next Steps

1. [x] Remove Hollywood effects from design (star streaks, motion blur)
2. [ ] Implement `engine/cinematic/easing.go`
3. [ ] Implement camera tumble/stabilization
4. [ ] Add parallax star layers for depth
5. [ ] Smooth planet scale transitions
6. [ ] Optional: Raymarched planet shader
7. [ ] Archive dialogue integration
