# Bubble Ship External View HUD

**Status**: Planned
**Priority**: P2 - Visual Polish / Navigation Aid
**Feature Type**: Rendering/UI
**Dependencies**: Tetra3D dome system (implemented), first-person navigation (implemented)

## Game Vision Alignment

Scored against [core-pillars.md](../../docs/vision/core-pillars.md):

| Pillar | Score | Notes |
|--------|-------|-------|
| Choices Are Final | N/A | UI element, no gameplay impact |
| The Game Doesn't Judge | N/A | Pure visual feature |
| Time Has Emotional Weight | **+2** | Seeing your tiny ship bubble against the vastness of space reinforces cosmic isolation |
| The Ship Is Home | **+2** | Exterior view shows the bubble as a self-contained world - your only home floating in the void |
| Grounded Strangeness | **+1** | The Higgs bubble's visual appearance (transparent, glowing boundary) is grounded in the game's physics |
| We Are Not Built For This | **+2** | The scale contrast (tiny bubble vs infinite space) emphasizes human fragility |

**Net Score: +7** - Strong alignment with core pillars.

## Problem Statement

**Current State:**
- First-person interior view provides immersive ship experience
- Window compositing shows space through deck windows with SR/GR effects
- No way to see the ship from outside during travel
- Players lose spatial awareness of their ship's relationship to surrounding space

**User Need:**
- Visual reference showing ship position/orientation during space traversal
- Sense of scale - see the bubble as a tiny speck in the cosmos
- Navigation aid showing ship's heading relative to stars/destinations

**Impact:**
- Without external view, the "bubble ship" concept is only experienced from inside
- Missing the iconic visual of "a tiny transparent sphere drifting through infinite space"
- Lost opportunity for emotional reinforcement of isolation/scale themes

## Solution Overview

Add a **Bubble Ship HUD** - a small 3D viewport in the corner of the screen showing the player's ship from an external third-person perspective during space navigation.

### Key Visual Elements

1. **Transparent Bubble Sphere**
   - Semi-transparent sphere using `TransparencyModeTransparent`
   - Alpha 0.2-0.4 so stars are visible through it
   - Optional subtle fresnel effect (edges more opaque than center)
   - Boundary glow when moving (ISM particle impacts)

2. **Ship Interior Silhouette**
   - Simplified 3D model of the spire and deck structure visible inside the bubble
   - Low-poly "tower" shape representing the internal ship
   - Lit from direction of nearest star

3. **Space Background**
   - Same starfield/SR effects as main view
   - Destination markers visible (if any)
   - Nearby planets/objects in miniature

4. **HUD Position**
   - Bottom-right or bottom-center of screen
   - Size: approximately 200x200 pixels (configurable)
   - Semi-transparent background or none

## Physics Basis

### Bubble Appearance (from design-decisions.md)

Per existing decisions:
- Bubble boundary is **optically permeable but refractive** - light passes through with slight distortion
- **Boundary glow** when ISM particles impact (mass rejection creates visible light)
- From outside: appears as a "shimmering gravitational lens effect" - not fully transparent

### Visual Implementation

| Effect | Physics Basis | Implementation |
|--------|---------------|----------------|
| Transparency | Light passes through Higgs boundary | `TransparencyModeTransparent`, alpha 0.3 |
| Boundary glow | ISM particle kinetic energy → light | Emissive shader on sphere edges, scaled by velocity |
| Fresnel effect | Grazing angles show more refraction | Shader: edge alpha higher than center |
| Interior visibility | Light passes both ways | Ship silhouette rendered inside bubble |

### What We're NOT Doing (Hollywood Conventions)

| Rejected | Why | Alternative |
|----------|-----|-------------|
| Opaque metallic hull | Bubble is transparent by physics | Transparent sphere with interior visible |
| Engine trails | No medium in vacuum | Point-source engine glow only |
| Motion blur streaks | Stars too distant | SR aberration on starfield instead |

## 3D Assets Required

### 1. Ship Interior Silhouette Model

A simplified 3D model representing the internal structure visible from outside:

```
Requirements:
- Low-poly (~500-1000 triangles)
- Central spire (cylinder, tapered)
- 3-4 horizontal disc levels around spire
- Subtle detail for scale reference
- Single material (can be tinted/lit)

Approximate dimensions:
- Spire height: ~150m (fills most of bubble diameter)
- Level discs: ~60-80m diameter
- Model should fit within 100m radius sphere
```

**Asset path**: `assets/models/ship_silhouette.gltf` (or `.obj`)

### 2. Bubble Sphere (Procedural)

No asset needed - use existing `NewBubble()` from `engine/tetra/dome.go` with modified material:

```go
// Modified for transparency
d.material.TransparencyMode = tetra3d.TransparencyModeTransparent
d.material.Color = tetra3d.NewColor(0.3, 0.5, 0.7, 0.25) // Blue tint, 25% alpha
d.material.Shadeless = false // Allow lighting on surface
```

### 3. Optional: Boundary Glow Texture

For ISM impact glow effect:
- Procedural or simple gradient texture
- Applied as emissive layer on bubble exterior
- Intensity scales with ship velocity

**Asset path**: `assets/textures/bubble_glow.png` (if needed)

## Architecture

### New Components

```
engine/
├── tetra/
│   └── bubble_ship.go      # BubbleShip composite (bubble + interior model)
└── hud/
    └── ship_view.go        # HUD viewport rendering bubble ship externally
```

### AILANG Integration

The HUD is purely visual - no game logic changes needed. Engine reads existing ship state:

```ailang
-- Already available in FrameOutput:
type ShipState3D = {
    pos: Vec3,           -- Ship position (for camera offset)
    velocity: Vec3,      -- For boundary glow intensity
    heading: Vec3,       -- Ship orientation
    gamma: float         -- For SR effects
}
```

### Rendering Pipeline

```
1. Main first-person view renders normally
2. Bubble HUD renders to separate small viewport:
   a. Position camera behind/above ship
   b. Render starfield (same SR transform as main view)
   c. Render transparent bubble sphere
   d. Render ship interior silhouette inside bubble
   e. Apply boundary glow based on velocity
3. Composite HUD viewport onto main screen
```

### Camera for External View

```go
// Camera positioned behind and slightly above ship
cameraOffset := Vec3{0, 20, -50}  // 50m behind, 20m above
cameraTarget := ship.pos          // Look at ship center
fov := 45                         // Narrow FOV for clean view
```

## Implementation Plan

### Phase 1: Transparent Bubble (~2 hours)
- [ ] Add `SetTransparency(alpha float)` method to `Dome`
- [ ] Add `SetTransparencyMode(mode)` method to `Dome`
- [ ] Create `NewTransparentBubble()` convenience constructor
- [ ] Test: transparent sphere rendering with stars visible through

### Phase 2: Ship Silhouette Model (~4 hours)
- [ ] Create low-poly ship silhouette model (spire + levels)
- [ ] Import model into Tetra3D scene
- [ ] Position inside bubble, attach to bubble transforms
- [ ] Test: ship visible inside transparent bubble

### Phase 3: HUD Viewport (~3 hours)
- [ ] Create `engine/hud/ship_view.go`
- [ ] Implement separate render target for HUD
- [ ] External camera positioning (third-person behind ship)
- [ ] Composite onto main screen
- [ ] Test: HUD displays in corner during navigation

### Phase 4: Boundary Glow (~2 hours)
- [ ] Create emissive/additive material for bubble edge
- [ ] Scale glow intensity by `ship.velocity` magnitude
- [ ] Optional: Fresnel shader for edge-brighter effect
- [ ] Test: glow increases with speed

### Phase 5: Integration & Polish (~2 hours)
- [ ] Add HUD toggle key (H or similar)
- [ ] Configurable HUD position/size
- [ ] Match SR effects between main view and HUD
- [ ] Performance optimization (render every N frames if needed)

## Image Generation Guidance

For AI-generated ship silhouette model or reference images:

### Ship Silhouette Prompt

```
A simplified 3D model of a vertical spaceship interior structure:
- Central cylindrical spire running top to bottom (the Higgs generator)
- 3-4 horizontal disc-shaped levels/platforms around the spire
- Spire is taller than it is wide, approximately 150m tall
- Levels are approximately 60-80m diameter
- Minimalist low-poly style
- No external hull - this is just the internal structure
- Inspired by 70s French sci-fi comics (Moebius/Druillet)
- Color: neutral grey/white for texturing later
- Viewed from 45 degrees above and behind
```

### Bubble Ship External View Prompt

```
A tiny transparent sphere floating in space containing a miniature city/ship:
- The sphere is semi-transparent with a subtle blue tint
- Inside the sphere: a vertical tower structure with horizontal platforms
- The sphere's edge has a faint luminous glow
- Background: deep space with stars, some showing blue/red shift
- Scale: the sphere should look tiny against the vastness of space
- Style: hard science fiction, Moebius/Métal Hurlant aesthetic
- Mood: isolation, fragility, wonder
```

### What NOT to Generate

- No solid metal spaceship hulls
- No engine trails or streaks
- No "Star Trek" style saucer shapes
- No aggressive/militaristic designs
- No warp effects or motion blur

## Success Criteria

- [ ] Transparent bubble renders with stars visible through
- [ ] Ship silhouette visible inside bubble
- [ ] HUD displays in configurable screen position
- [ ] Boundary glow intensity scales with velocity
- [ ] 60 FPS maintained with HUD active
- [ ] HUD can be toggled on/off
- [ ] External view uses same SR effects as main view

## Testing Strategy

**Visual Tests:**
- Screenshot comparison at various velocities
- HUD visibility across different background brightness
- Transparency correct in various lighting conditions

**Performance Tests:**
- FPS with HUD enabled vs disabled
- Memory usage for additional render target

## References

- [bubble-ship-design.md](../input/bubble-ship-design.md) - Full bubble ship physics/layout
- [dome.go](../../engine/tetra/dome.go) - Existing dome/bubble implementation
- [ring.go](../../engine/tetra/ring.go) - TransparencyModeTransparent usage example
- [design-decisions.md](../../docs/vision/design-decisions.md) - Bubble transparency, boundary glow decisions
- [core-pillars.md](../../docs/vision/core-pillars.md) - Game vision alignment

## Future Work

- **Dynamic bubble distortion** - Subtle wobble/refraction effects
- **Crew silhouettes** - Tiny figures visible on deck levels
- **Destination indicator** - Marker showing target star direction
- **Zoom control** - Ability to zoom HUD in/out
- **Alternative view angles** - Front, top, side views

---

**Document created**: 2025-01-09
**Last updated**: 2025-01-09
