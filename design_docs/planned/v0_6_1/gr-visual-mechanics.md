# GR Visuals & Mechanics Near Massive Objects

**Status**: Planned
**Target**: v0.6.1
**Priority**: P1 - Core to black hole experience
**Estimated**: 5-7 days
**Dependencies**: SR Effects (implemented), Shader System (implemented), Black Holes (black-holes.md)
**AILANG Workarounds**: None required (engine-side rendering + sim-side GR context)

## Related Documents

- [Black Holes](black-holes.md) - Core black hole mechanics, time dilation gameplay
- [SR Effects](../implemented/v0_5_0/sr-effects.md) - Special relativity effects (implemented)
- [Shader System](../implemented/v0_4_5/shader-system.md) - Kage shader pipeline (implemented)
- [Journey System](v0_6_0/journey-system.md) - Travel system (SR effects active during journey)

## Game Vision Alignment

**Score this feature against Stapledon's Voyage core pillars:**

| Pillar | Relevance | Score | Notes |
|--------|-----------|-------|-------|
| Time Dilation Consequence | Critical | +2 | GR time dilation is extreme near compact objects; visual reinforcement |
| Civilization Simulation | Moderate | +1 | Civs may orbit, worship, or exploit GR regions |
| Philosophical Depth | Critical | +2 | "Touching the edge of spacetime"; visible distortion of reality |
| Ship & Crew Life | Strong | +1 | Crew sees universe warp; visceral psychological impact |
| Legacy Impact | Moderate | +1 | Decisions near BH affect legacy; visuals make stakes clear |
| Hard Sci-Fi Authenticity | Critical | +2 | Real GR optics (Schwarzschild lensing, not fake effects) |
| We Are Not Built For This | Strong | +1 | Universe visibly breaking down around crew |
| **Net Score** | | **+10** | **Decision: Move forward - essential for BH experience** |

**Feature type:** Engine/Rendering + AILANG Simulation

**Reference:** See [game-vision.md](../../docs/game-vision.md), [black-holes.md](black-holes.md)

## Goals

**Primary Goals:**
- Add general relativity effects that matter only near compact objects (BH/NS/WD)
- Keep AILANG responsible for physics state; Go/Ebiten responsible for rendering
- Reuse existing SR pipeline (aberration/Doppler) and layer GR as an extra stage
- Make GR effects gameplay-relevant, not just eye candy

**Non-goal:** Full GR ray-tracing. We want approximations that look right and are cheap.

## Scope

GR effects are active only near "massive objects" with strong curvature:
- **Black holes** (stellar and supermassive)
- **Neutron stars**
- **White dwarfs** (subtle effects only)

Effects covered:
1. **Gravitational lensing** of background (stars/skybox)
2. **Gravitational redshift/blueshift**
3. **Gravitational time dilation** (mechanical)
4. **Danger zones** (photon sphere, tidal stress thresholds)

## Key Quantities / Thresholds

For a non-rotating Schwarzschild object:
- **Schwarzschild radius:** r_s = 2GM/c²
- **Dimensionless potential:** Φ = GM/(rc²) = r_s/(2r)

Use Φ as the single heuristic knob:

| Φ Range | GR Level | Effects |
|---------|----------|---------|
| Φ < 1e-4 | Off | Only SR active |
| 1e-4 ≤ Φ < 1e-3 | Subtle | Light shading only |
| 1e-3 ≤ Φ < 1e-2 | Noticeable | Redshift + mild lensing |
| 1e-2 ≤ Φ < 0.1 | Strong | Visible lensing, time dilation |
| Φ ≥ 0.1 | Extreme | Photon sphere, heavy distortion |

**Key radii:**
- **Photon sphere:** r_ph = 1.5 r_s (very strong lensing)
- **Event horizon (BH):** r = r_s
- **Surface (NS/WD):** Given radius; treat r < ~3×radius as "strong field region"

## Solution Design

### Overview

GR effects are computed in AILANG and passed to the engine for rendering:

```
Ship Position + Massive Objects → AILANG computes GRContext → Engine applies GR shader
```

The GR stage is a post-process that runs after SR effects, layering additional distortion and color shifts near compact objects.

### AILANG: Simulation / Data Model

#### New Types

```ailang
type MassiveObjectKind =
    | BlackHole
    | NeutronStar
    | WhiteDwarf

type MassiveObject = {
    id: MassiveId,
    kind: MassiveObjectKind,
    mass_solar: float,              -- in M_sun
    schwarzschild_radius: float,    -- meters or game units
    position: Vec3                  -- galaxy-frame
}

type GRDangerLevel =
    | GR_None
    | GR_Subtle
    | GR_Strong
    | GR_Extreme

type GRContext = {
    active: bool,
    object_id: MassiveId,
    kind: MassiveObjectKind,
    r: float,                       -- distance ship→object center
    phi: float,                     -- GM/(r c^2)
    time_dilation: float,           -- dτ/dt = sqrt(1 - 2GM/(rc^2)) clamped
    redshift_factor: float,         -- z_gr ≈ 1/sqrt(1-2GM/rc^2)
    tidal_severity: float,          -- 0..1 heuristic
    danger_level: GRDangerLevel,
    can_hover_safely: bool,         -- tidal stress below threshold
    near_photon_sphere: bool        -- r in [1.3 r_s, 2.0 r_s] for BH
}
```

#### Extension to FrameOutput

```ailang
type FrameOutput = {
    -- existing fields...
    camera: CameraState,
    draw_cmds: [DrawCmd],

    -- new GR context
    gr: GRContext
}
```

#### GR Context Computation

The sim updates `GRContext` when:
- Ship is within `GR_RANGE` of a massive object (e.g., r < 100 r_s)
- Otherwise `active = false`

```ailang
pure func classify_gr(phi: float) -> GRDangerLevel {
    match true {
        phi < 0.0001 => GR_None,
        phi < 0.001  => GR_Subtle,
        phi < 0.01   => GR_Strong,
        _            => GR_Extreme
    }
}

pure func compute_gr_context(ship_pos: Vec3, obj: MassiveObject) -> GRContext {
    let r = distance(ship_pos, obj.position)
    let r_s = obj.schwarzschild_radius
    let phi = r_s / (2.0 * r)

    -- Time dilation: dτ/dt = sqrt(1 - r_s/r), clamped for stability
    let td_arg = clamp(1.0 - r_s / r, 0.001, 1.0)
    let time_dilation = sqrt(td_arg)

    -- Redshift factor: z ≈ 1/sqrt(1 - r_s/r)
    let redshift_factor = 1.0 / time_dilation

    -- Tidal severity heuristic (higher for smaller r and smaller mass)
    let tidal_severity = clamp(r_s / (r * r) * 1e6, 0.0, 1.0)

    let danger = classify_gr(phi)
    let near_photon = match obj.kind {
        BlackHole => r >= 1.3 * r_s && r <= 2.0 * r_s,
        _ => false
    }

    {
        active: phi >= 0.0001,
        object_id: obj.id,
        kind: obj.kind,
        r: r,
        phi: phi,
        time_dilation: time_dilation,
        redshift_factor: redshift_factor,
        tidal_severity: tidal_severity,
        danger_level: danger,
        can_hover_safely: tidal_severity < 0.5,
        near_photon_sphere: near_photon
    }
}
```

#### Time Dilation (Mechanical)

AILANG uses `time_dilation` for gameplay:
- **Local proper time step:** Δτ = time_dilation × Δt_external
- **External galaxy simulation** uses Δt_external
- **Ship/crew evolve** using Δτ

Effects:
- Near BH/NS, crew experiences shorter subjective intervals per external century
- Used in: crew aging, event timing, "years passed outside" displays

### Engine: Rendering Architecture

#### Pipeline Integration

Current pipeline:
1. AILANG → FrameOutput
2. Engine: SR stage (aberration + Doppler to starfield/skybox)
3. Draw starfield/background
4. Draw in-world objects

**New with GR:**
1. AILANG → FrameOutput.gr
2. Engine:
   - Compute GR parameters from GRContext
   - **SR stage:** as before
   - **GR stage (post-process):**
     - If `gr.active` and `danger_level ≥ GR_Subtle`:
       - Apply GR distortion shader over background layer
       - Apply GR redshift/blueshift near the object
   - Draw foreground objects last (ship UI, HUD)

#### Shader Uniforms

```go
type GRShaderUniforms struct {
    Enabled         bool
    ObjectKind      int       // 0=BH, 1=NS, 2=WD
    Rs              float32   // Schwarzschild radius (screen units)
    Distance        float32   // ship→object distance
    Phi             float32   // dimensionless potential
    TimeDilation    float32   // dτ/dt
    RedshiftFactor  float32   // gravitational z
    ScreenCenter    Vec2      // screen position of object center
    MaxEffectRadius float32   // in screen-space units
}
```

#### Lensing Approximation (Fragment Shader)

For each pixel with screen coord `p`:
1. Compute vector from lens center: `d = p - ScreenCenter`
2. Compute impact parameter: `b = |d|` in angular units
3. If `b > MaxEffectRadius`: pass through
4. Else approximate deflection angle:
   - Weak-field approximation: `α(b) ≈ 4GM/(bc²) = 2r_s/b`
   - Adjust sampling direction:

```kage
// Pseudo-code for fragment shader
func Fragment(position vec4, texCoord vec2, color vec4) vec4 {
    d := position.xy - ScreenCenter
    b := length(d)

    if b > MaxEffectRadius {
        return texture(backgroundTex, texCoord)
    }

    dir := normalize(d)
    eps := 0.001

    // Deflection angle (tuned empirically)
    alpha := LensStrength * Rs / max(b, eps)

    // Radial stretching
    warpedB := b + alpha * b
    sampleCoord := ScreenCenter + dir * warpedB

    // Sample warped coordinate
    baseColor := texture(backgroundTex, screenToUV(sampleCoord))

    // Apply gravitational redshift (radial falloff)
    redshiftBlend := clamp(1.0 - b / MaxEffectRadius, 0.0, 1.0)
    shiftedColor := applyRedshift(baseColor, RedshiftFactor * redshiftBlend)

    return shiftedColor
}
```

Visual effects from this approximation:
- Arcs around compact object
- Background wrapping
- Einstein-ring-like patterns at right distances

#### Gravitational Redshift

Two applications:
1. **General background shift** near massive object:
   - Multiply color by function of `redshift_factor` and radial distance from center
   - Closer to center → more red (less blue)

2. **Accretion disk rendering** (if present):
   - Inner disk: increase brightness, shift to blue
   - Outer region: shift to red

Combine GR redshift with SR Doppler:
- Use SR for angle-dependent shift
- Use GR as radial multiplicative factor near compact object

### Architecture Diagram

```
┌──────────────────────────────────────────────────────────────────────┐
│                          AILANG (sim/)                               │
│                                                                      │
│   Ship State + MassiveObjects → compute_gr_context()                 │
│                                                                      │
│   Output: FrameOutput.gr = {                                         │
│       active: true,                                                  │
│       phi: 0.05,                                                     │
│       time_dilation: 0.32,                                           │
│       redshift_factor: 3.16,                                         │
│       danger_level: GR_Strong,                                       │
│       ...                                                            │
│   }                                                                  │
└─────────────────────────────┬────────────────────────────────────────┘
                              │
                              ▼
┌──────────────────────────────────────────────────────────────────────┐
│                     Engine (Go/Ebiten)                               │
│                                                                      │
│   1. SR Stage (existing):                                            │
│      - Aberration transform                                          │
│      - Doppler shift                                                 │
│      - Beaming                                                       │
│                                                                      │
│   2. GR Stage (new):                                                 │
│      - if gr.active && gr.danger_level >= GR_Subtle:                 │
│          - Convert GRContext → GRShaderUniforms                      │
│          - Apply gr_lensing.kage to background                       │
│          - Apply gr_redshift.kage overlay                            │
│                                                                      │
│   3. Foreground:                                                     │
│      - Draw objects, UI, HUD                                         │
└──────────────────────────────────────────────────────────────────────┘
```

### Implementation Plan

**Phase 1: AILANG GR Types** (~4 hours)
- [ ] Add `MassiveObject`, `MassiveObjectKind` types to `sim/world.ail`
- [ ] Add `GRContext`, `GRDangerLevel` types to `sim/protocol.ail`
- [ ] Add `gr: GRContext` field to `FrameOutput`
- [ ] Implement `compute_gr_context()` function
- [ ] Implement `classify_gr()` helper

**Phase 2: Engine GR Data** (~4 hours)
- [ ] Add `GRContext` struct to `sim_gen/` (manual or generated)
- [ ] Add `GRShaderUniforms` struct to `engine/relativity/`
- [ ] Implement conversion from `GRContext` to shader uniforms
- [ ] Project massive object position to screen coordinates

**Phase 3: GR Lensing Shader** (~8 hours)
- [ ] Create `engine/shader/shaders/gr_lensing.kage`
- [ ] Implement radial distortion based on Schwarzschild approximation
- [ ] Add Einstein ring effect at appropriate radii
- [ ] Handle edge cases (object off-screen, very close approach)

**Phase 4: GR Redshift Shader** (~6 hours)
- [ ] Create `engine/shader/shaders/gr_redshift.kage`
- [ ] Implement radial redshift falloff
- [ ] Integrate with SR Doppler (multiplicative)
- [ ] Optional: Accretion disk color shifts

**Phase 5: Gameplay Integration** (~6 hours)
- [ ] HUD indicators for time dilation, danger level
- [ ] Warnings at danger level transitions
- [ ] Crew psychology hooks for GR proximity
- [ ] Event triggers based on danger_level changes

**Phase 6: Polish & Testing** (~4 hours)
- [ ] Tune shader parameters for visual quality
- [ ] Performance optimization (limit effect radius)
- [ ] Golden image tests for each danger level
- [ ] Quality settings (Off/Basic/Full)

### Files to Modify/Create

**New files:**
- `sim/gr.ail` - GR types and computation functions (~150 LOC)
- `engine/relativity/gr_context.go` - GRContext handling (~100 LOC)
- `engine/shader/shaders/gr_lensing.kage` - Lensing effect (~100 LOC)
- `engine/shader/shaders/gr_redshift.kage` - Redshift overlay (~60 LOC)

**Modified files:**
- `sim/protocol.ail` - Add GRContext to FrameOutput (~20 LOC)
- `sim/world.ail` - Add MassiveObject type (~30 LOC)
- `sim/step.ail` - Call compute_gr_context (~20 LOC)
- `engine/render/draw.go` - Add GR post-process stage (~50 LOC)
- `engine/shader/effects.go` - Register GR shaders (~30 LOC)

## Gameplay Integration

### Player Feedback (HUD)

Use `GRContext` + `GRDangerLevel` to drive:

**HUD Indicators:**
```
┌─────────────────────────────────────────┐
│  ⚫ COMPACT OBJECT: Cygnus X-1 (BH)     │
│                                         │
│  Distance:       12.3 r_s               │
│  Time Dilation:  ×0.32 (local/external) │
│  GR Level:       ████░░ Strong          │
│  Tidal Stress:   ██░░░░ Low             │
└─────────────────────────────────────────┘
```

**Warnings:**
- "Strong curvature region - lensing active"
- "Photon sphere proximity - navigation unstable"
- "Approach horizon = irreversible"

### Event Hooks

When `danger_level` crosses thresholds:
- Trigger crew debates / fear / awe events
- Stress tests on hull / systems (especially for NS / small BH)
- Philosophical prompts ("Enter horizon?", "Skim for data?", "Retreat?")
- Time-skip mechanics: option to "hover here for Δτ hours" letting Δt_external run far

### Endgame / Special Choices

Special conditions:
- `kind == BlackHole AND r <= r_s`:
  - Trigger "cross horizon" endgame branch (The Witness / The Last Star path)
- `kind == SupermassiveBH AND Phi high but tidal_severity low`:
  - Safe-but-extreme time skip zone

The GR visuals make these choices *felt* rather than just text.

## Performance / Quality Controls

**GR shader only enabled when:**
- `gr.active == true` AND `danger_level >= GR_Subtle`

**Quality settings:**

| Setting | Description |
|---------|-------------|
| Off | No GR rendering; only mechanical time dilation |
| Basic | Simple radial distortion + redshift |
| Full | Aggressive lensing (Einstein ring, more iterations) |

**Performance budget:**
- Max cost kept under control by:
  - Limiting `MaxEffectRadius` (small region around the object)
  - Single-pass post-process over background layer
- Target: <3ms GPU time at 1080p for GR effects

## Examples

### Example 1: Approaching a Stellar-Mass Black Hole

**Ship at r = 5 r_s from a 10 M☉ black hole:**

```
GRContext = {
    active: true,
    kind: BlackHole,
    r: 147.7 km,         -- 5 × 29.5 km
    phi: 0.1,
    time_dilation: 0.775,
    redshift_factor: 1.29,
    danger_level: GR_Extreme,
    tidal_severity: 0.8,
    can_hover_safely: false,
    near_photon_sphere: false
}
```

**Visual:** Strong background distortion, visible Einstein ring, stars near edge visibly stretched into arcs. Red tint increases toward center.

### Example 2: Orbiting a Supermassive Black Hole

**Ship at r = 10 r_s from Sagittarius A* (4×10⁶ M☉):**

```
GRContext = {
    active: true,
    kind: BlackHole,
    r: 1.2×10⁸ km,       -- 10 × 1.2×10⁷ km
    phi: 0.05,
    time_dilation: 0.316,
    redshift_factor: 3.16,
    danger_level: GR_Strong,
    tidal_severity: 0.001,  -- low (large mass)
    can_hover_safely: true,
    near_photon_sphere: false
}
```

**Visual:** Moderate lensing, clear gravitational redshift gradient. Safe for extended time-skip orbits.

### Example 3: Neutron Star Flyby

**Ship at r = 3 × star radius from a 1.4 M☉ neutron star (R = 10 km):**

```
GRContext = {
    active: true,
    kind: NeutronStar,
    r: 30 km,
    phi: 0.07,
    time_dilation: 0.63,
    redshift_factor: 1.58,
    danger_level: GR_Strong,
    tidal_severity: 0.3,
    can_hover_safely: true,
    near_photon_sphere: false
}
```

**Visual:** Noticeable lensing, moderate redshift. Lower tidal stress than stellar-mass BH at comparable Φ.

## Success Criteria

### Visual
- [ ] At Φ > 0.01: visible background distortion around compact object
- [ ] At Φ > 0.1: dramatic lensing with arc formation
- [ ] Gravitational redshift visible as radial color gradient
- [ ] Einstein ring visible at appropriate viewing angles
- [ ] Smooth transition as ship approaches/retreats
- [ ] GR effects layer correctly over SR effects

### Technical
- [ ] AILANG outputs complete GRContext in FrameOutput
- [ ] Engine correctly interprets all GRContext fields
- [ ] GR lensing shader produces physically plausible distortion
- [ ] Performance: <3ms GPU at 1080p
- [ ] Quality settings work (Off/Basic/Full)

### Gameplay
- [ ] Time dilation affects crew aging correctly
- [ ] Danger level triggers appropriate warnings
- [ ] HUD shows accurate GR information
- [ ] Crew events trigger at danger level thresholds

## Testing Strategy

**Unit tests:**
- `compute_gr_context()` accuracy against known physics
- Danger level classification
- Time dilation formula verification

**Visual tests (golden files):**
- `out/gr/bh_distant.png` - BH at r = 50 r_s
- `out/gr/bh_strong.png` - BH at r = 3 r_s
- `out/gr/bh_extreme.png` - BH at r = 1.5 r_s (photon sphere)
- `out/gr/ns_strong.png` - Neutron star close approach
- `out/gr/combined_sr_gr.png` - Both SR and GR active

**Manual testing:**
- Verify visual progression feels physically correct
- Check performance at different quality levels
- Verify gameplay triggers at correct thresholds

## Non-Goals

**Not in this feature:**
- **Kerr (rotating) BH** - Frame dragging, ergosphere (future expansion)
- **Full geodesic ray-tracing** - Too expensive; approximation sufficient
- **Accretion disk physics** - Separate visual feature
- **Hawking radiation** - Not visually relevant at our scales
- **Relativistic jets** - Separate feature (NS/BH jets)

## Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|-----------|
| Lensing shader too expensive | Medium | LOD system; limit effect radius; quality settings |
| Lensing looks wrong to physics-aware players | High | Follow Schwarzschild formula; document approximations |
| GR + SR interaction confusing | Medium | Clear layering; GR post-processes SR output |
| Edge cases at very close approach | Medium | Clamp values; graceful degradation near horizon |

## References

- [Black Holes Design Doc](black-holes.md) - Core BH mechanics
- [SR Rendering Design Doc](sr-rendering.md) - Special relativity effects
- [Relativistic Visual Effects](relativistic-visual-effects.md) - SR implementation
- Misner, Thorne, Wheeler - *Gravitation* (Schwarzschild geometry)
- [Interstellar VFX](https://www.dneg.com/projects/interstellar/) - GR rendering reference
- [Black Hole Visualization (NASA)](https://svs.gsfc.nasa.gov/13326) - Reference imagery

## Future Work

- **Kerr black holes** - Rotating BH with ergosphere, frame dragging
- **Accretion disk rendering** - Proper disk physics with GR effects
- **Photon orbit visualization** - Light trapped at photon sphere
- **Gravitational waves** - Visual ripples from binary mergers
- **GR audio** - Gravitational redshift on incoming radio signals

---

**Document created**: 2025-12-05
**Last updated**: 2025-12-05
**Origin**: M-GAME-ENGINE epic - GR visual layer for black hole experience
