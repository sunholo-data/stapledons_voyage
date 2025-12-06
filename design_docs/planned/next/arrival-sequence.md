# Arrival Sequence & Bridge View

**Version:** 0.2.0
**Status:** Planned
**Priority:** P0 (First Player Experience)
**Complexity:** High
**Depends On:** SR Effects (implemented), GR Effects (implemented), Asset Pipeline, Input System
**Estimated:** 2 sprints

## Related Documents

- [Opening Sequence](../future/opening-sequence.md) - Narrative context (emergence from structure)
- [SR Effects](../../implemented/v0_1_0/sr-effects.md) - Special relativity effects (implemented)
- [GR Effects](../../implemented/v0_1_0/gr-effects.md) - General relativity effects (implemented)
- [Ship Exploration](ship-exploration.md) - Interior navigation (separate mode)
- [Galaxy Map](galaxy-map.md) - Strategic navigation (accessed from bridge)
- [Bubble Ship Design](../../input/bubble-ship-design.md) - Ship layout reference

## Game Vision Alignment

| Pillar | Score | Notes |
|--------|-------|-------|
| Choices Are Final | ⚪ N/A | Tutorial setup, choices come after |
| Game Doesn't Judge | ✅ Strong | Player discovers through experience |
| Time Has Emotional Weight | ✅ Strong | First visceral taste of SR time dilation |
| Ship Is Home | ✅ Strong | First view of bridge, introduces crew/Archive |
| Grounded Strangeness | ✅ Strong | Disorientation → wonder → understanding |
| We Are Not Built For This | ✅ Strong | SR effects are alien, overwhelming at first |

---

## Feature Overview

The **Arrival Sequence** is the player's first experience of the game. It serves multiple purposes:

1. **Spectacle** - Showcase SR effects and planetary visuals
2. **Tutorial** - Teach basic controls through necessity
3. **Worldbuilding** - Establish the ship, Archive, and situation
4. **Tone-setting** - Create sense of disorientation → mastery → wonder

### The Core Experience

> Reality warps around you as you emerge from a structure that bends light itself. Space distorts, twists, and then snaps into clarity as you're ejected at near-lightspeed. You're spinning wildly. Stars streak past in impossible colors. Through the chaos, you glimpse something familiar - Saturn's rings? The Archive's voice cuts through static, guiding you to stabilize. As you gain control, the universe snaps into focus. You're decelerating through our solar system, planets growing from points of light to breathtaking vistas. By the time Earth fills your view, you understand: you've been traveling at near-lightspeed, and nothing will ever be the same.

---

## Detailed Specification

### Phase 0: Black Hole Emergence (8 seconds)

**Visual State:**
- Player emerges from a "mysterious structure" (black hole)
- GR lensing effects at maximum (gravitational distortion)
- Space itself appears warped, light bending around unseen mass
- Accretion disk glow fading as distance increases
- GR intensity fades from 1.0 → 0.0 over 8 seconds

**Audio:**
- Deep rumbling, reality-warping sounds
- Distorted Archive voice: "...transit complete... systems..."
- Gradual transition to SR soundscape

**Player Input:**
- None - purely cinematic
- Player is ejected at 0.99c

**GR Effects:**
- `grIntensity` starts at 1.0 (Extreme level)
- Uses existing GRWarp shader with lensing + redshift
- Center of effect is the "structure" behind the ship
- Fades linearly to 0.0 as player moves away

**Transition:**
- After 8 seconds, GR effects fully fade
- Seamlessly transitions to Phase 1 (SR tumble)

---

### Phase 1: Emergence & Chaos (30-60 seconds)

**Visual State:**
- Camera tumbling rapidly (rotation on multiple axes)
- SR effects at maximum (0.99c equivalent visual distortion)
- Stars compressed into forward cone, violently blue-shifted
- Occasional glimpses of recognizable shapes (Saturn's rings flash past)
- Screen shake, motion blur

**Audio:**
- Ship alarms, system warnings
- Archive voice (fragmented): "...stabilization required... orientation lost..."
- Hull stress sounds, energy discharge

**Player Input:**
- Minimal at first - establishing helplessness
- After 10-15 seconds, prompt appears: "STABILIZE ROTATION - [WASD/Arrow Keys]"
- Player inputs gradually reduce tumble rate
- Success feedback: each axis stabilized reduces chaos

**Mechanics:**
```
Tumble State:
- pitch_rate: -180 to +180 deg/sec (starts random)
- yaw_rate: -180 to +180 deg/sec
- roll_rate: -180 to +180 deg/sec

Player Input Effect:
- W/S: reduce pitch_rate toward 0
- A/D: reduce yaw_rate toward 0
- Q/E (or auto): reduce roll_rate toward 0

Stabilization threshold: all rates < 5 deg/sec
```

**SR Effects During Chaos:**
- Full aberration (stars in tight cone)
- Extreme Doppler (blue forward, red rear)
- D³ beaming (forward blindingly bright)
- Effect intensity tied to apparent velocity

### Phase 2: Gaining Control (30 seconds)

**Trigger:** All rotation rates stabilized

**Visual State:**
- Camera steadies, forward view locked
- SR effects still present but comprehensible
- Saturn becomes visible ahead (first landmark)
- Stars no longer streaking - stable starfield with SR distortion

**Audio:**
- Alarms cease
- Archive voice (clearer): "Rotation stabilized. Initiating deceleration sequence."
- Engine tone changes (thrust reversing)

**Player Input:**
- New prompt: "ENGAGE DECELERATION - [SPACE]"
- Holding SPACE increases deceleration
- Visual feedback: SR effects gradually reduce as velocity drops

**Tutorial Element:**
- HUD elements fade in one by one
- Ship time vs Galaxy time display appears
- Velocity indicator (β, γ) fades in
- Archive explains: "Ship time: 47.3 years. External time: [calculating]..."

### Phase 3: Solar System Tour (2-4 minutes)

**Structure:** Series of planetary approach sequences, each showcasing SR effects at different velocities.

#### 3a. Saturn Approach (0.95c → 0.9c)

**Visual:**
- Saturn grows from point to disk to magnificent ringed giant
- Real NASA/ESA imagery composited as textured sphere
- Ring system rendered with transparency, shadowing
- SR effects: noticeable aberration, blue tint to approaching side
- Moons visible as points of light

**Interactivity:**
- Player can adjust deceleration rate (affects SR intensity)
- Optional: "Look around" with mouse (limited arc)
- Archive provides context: "Saturn. 1.4 billion kilometers from Sol. Or... 4.7 light-hours."

**Duration:** ~45 seconds

#### 3b. Jupiter Approach (0.9c → 0.8c)

**Visual:**
- Jupiter's bands, Great Red Spot
- Real imagery, dynamic cloud patterns
- Galilean moons visible (Io's volcanism?)
- SR effects lessening but still visible

**Interactivity:**
- Same controls
- Archive: "Jupiter. The shepherd. Its gravity shaped this system."

**Tutorial Element:**
- Introduce time dilation readout: "At current velocity, 1 hour ship time = 2.3 hours external"

**Duration:** ~45 seconds

#### 3c. Mars Approach (0.8c → 0.5c)

**Visual:**
- Mars grows, Olympus Mons visible
- Rust-red surface detail
- Phobos/Deimos as tiny dots
- SR effects now subtle (player can compare to earlier)

**Interactivity:**
- Archive notes something: "Unusual readings from the inner system. Continuing approach."

**Duration:** ~30 seconds

#### 3d. Earth Approach (0.5c → orbit)

**Visual:**
- Earth appears - breathtaking blue marble
- Real imagery, cloud patterns
- Moon visible
- SR effects fade to zero as velocity drops
- Final approach: Earth fills significant portion of view

**Emotional Beat:**
- Archive: "Earth. Origin point. Home."
- Pause for impact
- Then: "Detecting broad-spectrum transmissions. Civilization active. Year... [long pause] ...2157 CE."
- Player realization: centuries have passed

**Interactivity:**
- Player prompted to "ENTER ORBIT - [SPACE]"
- Smooth transition to orbital view

**Duration:** ~60 seconds

### Phase 4: Bridge View Established

**Trigger:** Earth orbit achieved

**Visual State:**
- The **Bridge View** - the main gameplay screen - is now established
- Earth visible through the observation dome
- Ship interior UI elements fully visible
- Crew silhouettes at stations (not interactive yet)

**Layout:**
```
┌─────────────────────────────────────────────────────────────────┐
│                                                                 │
│                    OBSERVATION DOME                             │
│                    (Space/Planet View)                          │
│                                                                 │
│         ╭─────────────────────────────────────────╮             │
│         │           Earth / Space View            │             │
│         │        (SR effects applied here)        │             │
│         │                                         │             │
│         │              🌍                         │             │
│         │                                         │             │
│         ╰─────────────────────────────────────────╯             │
│                                                                 │
│  ┌──────────┐                               ┌──────────┐        │
│  │ Ship Time│                               │ Gal Time │        │
│  │  47y 3mo │                               │ 2157 CE  │        │
│  └──────────┘                               └──────────┘        │
│                                                                 │
├─────────────────────────────────────────────────────────────────┤
│                     BRIDGE INTERIOR                             │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │  [Helm]    [Archive]    [Comms]    [Systems]    [Map]   │    │
│  │    👤         📚          📡          ⚙️          🗺️     │    │
│  └─────────────────────────────────────────────────────────┘    │
│                                                                 │
│         Crew silhouettes at stations, ambient activity          │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

**Bridge Elements:**
- **Observation Dome** (top 60%): Space view with SR/GR effects
- **Ship Interior** (bottom 40%): Isometric or front-view of bridge stations
- **HUD Overlays**: Time displays, velocity, status indicators
- **Interactive Hotspots**: Crew members, consoles (click to interact)

**Audio:**
- Ambient ship sounds (ventilation, soft beeps)
- Crew murmurs
- Archive ready for conversation

### Phase 5: First Choices

**After orbital establishment:**

Archive initiates dialogue:
> "Captain. We have arrived. Ship systems nominal. Crew status... [pause] ...I am still compiling crew status."
>
> "Earth is transmitting. Standard identification protocols. They do not recognize our vessel configuration."
>
> "How do you wish to proceed?"

**First Choice:**
1. **"Respond to Earth"** - Initiate contact, begin story
2. **"Check ship status first"** - Enter Ship Exploration mode
3. **"Access the Archive"** - Learn about the journey so far
4. **"View the galaxy map"** - See the larger picture

This branches into the main game loop.

---

## Visual Design

### Real Planetary Imagery

**Sources (Public Domain/CC):**
- NASA Photojournal: https://photojournal.jpl.nasa.gov/
- NASA Image Gallery: https://images.nasa.gov/
- ESA Image Archive: https://www.esa.int/ESA_Multimedia/Images

**Required Assets:**

| Planet | Resolution | Format | Source |
|--------|------------|--------|--------|
| Saturn | 4K+ | PNG with alpha | NASA Cassini |
| Saturn Rings | 4K+ | PNG with alpha | NASA Cassini |
| Jupiter | 4K+ | PNG | NASA Juno/Hubble |
| Mars | 4K+ | PNG | NASA MRO/Viking |
| Earth | 4K+ | PNG | NASA Blue Marble |
| Moon | 2K | PNG | NASA LRO |

**Implementation:**
- Planets rendered as textured spheres (or high-quality billboards)
- Distance-based LOD (point → disk → detailed sphere)
- Rotation animation for gas giants
- Atmospheric glow effect for Earth

### SR Effects Integration

The SR shader system (already implemented) applies to the space view:

| Velocity | Effect |
|----------|--------|
| 0.99c | Extreme: stars in 30° cone, bright blue-white |
| 0.95c | Strong: stars in 60° cone, noticeable blue shift |
| 0.9c | Moderate: visible aberration, color shift |
| 0.5c | Subtle: slight star clustering, hint of shift |
| 0.0c | None: normal starfield |

**Planet Rendering with SR:**
- Planets are "local" objects, not affected by stellar aberration
- But their light IS doppler shifted during approach
- Blue shift when approaching, red shift if receding
- This creates beautiful effect: blue-tinged Saturn as you approach

### Bridge Interior Design

**Style:**
- Semi-realistic sci-fi
- Warm interior lighting contrasting with cold space
- Curved surfaces (inside of bubble observation deck)
- Holographic displays, physical controls mix
- Lived-in feel (personal items, wear marks)

**Perspective:**
- Could be isometric (consistent with ship exploration)
- Or front-facing "looking at bridge from captain's position"
- Recommend: **slight 3/4 view** - shows depth, allows clickable stations

**Crew Visualization:**
- Silhouettes or simple sprites at first
- Not individually interactive during arrival
- After arrival, become interactive NPCs

---

## Engine Integration

### New Components Required

#### 1. `engine/arrival/sequence.go`

```go
type ArrivalSequence struct {
    Phase      ArrivalPhase
    Progress   float64  // 0.0-1.0 within phase

    // Tumble state
    PitchRate  float64
    YawRate    float64
    RollRate   float64

    // Velocity state
    Velocity   float64  // fraction of c
    Position   Vec3     // solar system coords

    // Camera
    Camera     *ArrivalCamera

    // Planets
    Planets    []*PlanetRenderer

    // UI elements
    HUD        *ArrivalHUD
    Archive    *ArchiveDialogue
}

type ArrivalPhase int
const (
    PhaseEmergence ArrivalPhase = iota
    PhaseStabilizing
    PhaseSaturn
    PhaseJupiter
    PhaseMars
    PhaseEarth
    PhaseBridgeEstablished
)
```

#### 2. `engine/arrival/planets.go`

```go
type PlanetRenderer struct {
    Name        string
    Texture     *ebiten.Image
    Position    Vec3
    Radius      float64
    Rotation    float64
    HasRings    bool
    RingTexture *ebiten.Image
}

func (p *PlanetRenderer) Render(screen *ebiten.Image, camera *Camera, srWarp *SRWarp)
func (p *PlanetRenderer) ApproachDistance(shipPos Vec3) float64
```

#### 3. `engine/arrival/bridge.go`

```go
type BridgeView struct {
    SpaceViewport  Rect  // Top portion - space/planets
    BridgeViewport Rect  // Bottom portion - interior

    // Interactive elements
    Stations       []*BridgeStation
    HoveredStation *BridgeStation

    // Ambient
    CrewPositions  []CrewSilhouette
    Lighting       BridgeLighting
}

type BridgeStation struct {
    Name     string
    Position Vec2
    Bounds   Rect
    Action   func()  // What happens on click
}
```

#### 4. Integrate with Existing SR Shader

```go
// In arrival rendering loop
func (a *ArrivalSequence) Draw(screen *ebiten.Image) {
    // 1. Render space background
    a.RenderStarfield(screen)

    // 2. Apply SR warp based on current velocity
    a.SRWarp.SetForwardVelocity(a.Velocity)
    a.SRWarp.Apply(a.spaceBuffer, a.starfieldBuffer)

    // 3. Render planets (after SR, or with individual doppler)
    for _, planet := range a.Planets {
        planet.Render(a.spaceBuffer, a.Camera, a.SRWarp)
    }

    // 4. Composite space view into observation dome area
    screen.DrawImage(a.spaceBuffer, &opts)

    // 5. Render bridge interior
    a.BridgeView.Render(screen)

    // 6. Render HUD overlays
    a.HUD.Render(screen)
}
```

### Camera System

```go
type ArrivalCamera struct {
    // During chaos phase - tumbling
    Pitch, Yaw, Roll float64

    // Target tracking
    LookTarget Vec3  // Planet we're approaching

    // Smooth interpolation
    CurrentDir Vec3
    TargetDir  Vec3
    LerpSpeed  float64
}

func (c *ArrivalCamera) ApplyTumble(dt float64, pitchRate, yawRate, rollRate float64)
func (c *ArrivalCamera) LookAt(target Vec3, dt float64)
func (c *ArrivalCamera) GetViewMatrix() Matrix4
```

---

## AILANG Integration

### Types (sim/arrival.ail)

```ailang
module sim/arrival

type ArrivalPhase =
    | Emergence
    | Stabilizing
    | ApproachingSaturn
    | ApproachingJupiter
    | ApproachingMars
    | ApproachingEarth
    | BridgeEstablished
    | Complete

type ArrivalState = {
    phase: ArrivalPhase,
    progress: float,           -- 0.0 to 1.0 within phase
    velocity: float,           -- fraction of c
    pitch_rate: float,
    yaw_rate: float,
    roll_rate: float,
    ship_time_years: float,
    galaxy_year: int,
    stabilization_complete: bool,
    earth_contact_made: bool
}

type ArrivalInput = {
    stabilize_pitch: int,      -- -1, 0, or 1
    stabilize_yaw: int,
    decelerate: bool,
    interact: Maybe(string)    -- clicked station name
}

type ArrivalOutput = {
    camera_tumble: (float, float, float),
    sr_velocity: float,
    current_planet: Maybe(string),
    hud_elements: [HudElement],
    archive_dialogue: Maybe(string),
    phase_complete: bool,
    transition_to: Maybe(GameMode)
}
```

### Step Function (sim/arrival_step.ail)

```ailang
module sim/arrival_step

import sim/arrival (ArrivalState, ArrivalInput, ArrivalOutput, ArrivalPhase)

pure func step_arrival(state: ArrivalState, input: ArrivalInput, dt: float) -> (ArrivalState, ArrivalOutput) {
    match state.phase {
        Emergence => step_emergence(state, input, dt),
        Stabilizing => step_stabilizing(state, input, dt),
        ApproachingSaturn => step_planet_approach(state, input, dt, "Saturn", 0.95, 0.9),
        ApproachingJupiter => step_planet_approach(state, input, dt, "Jupiter", 0.9, 0.8),
        ApproachingMars => step_planet_approach(state, input, dt, "Mars", 0.8, 0.5),
        ApproachingEarth => step_earth_approach(state, input, dt),
        BridgeEstablished => step_bridge(state, input, dt),
        Complete => (state, default_output())
    }
}

pure func step_stabilizing(state: ArrivalState, input: ArrivalInput, dt: float) -> (ArrivalState, ArrivalOutput) {
    -- Apply player input to reduce rotation rates
    let new_pitch = dampen(state.pitch_rate, intToFloat(input.stabilize_pitch), dt);
    let new_yaw = dampen(state.yaw_rate, intToFloat(input.stabilize_yaw), dt);
    let new_roll = dampen(state.roll_rate, 0.0, dt);  -- Auto-stabilize roll

    let stabilized = abs(new_pitch) < 5.0 && abs(new_yaw) < 5.0 && abs(new_roll) < 5.0;

    let new_phase = if stabilized then ApproachingSaturn else Stabilizing;

    let new_state = { state |
        pitch_rate: new_pitch,
        yaw_rate: new_yaw,
        roll_rate: new_roll,
        phase: new_phase,
        stabilization_complete: stabilized
    };

    let output = {
        camera_tumble: (new_pitch, new_yaw, new_roll),
        sr_velocity: state.velocity,
        current_planet: None,
        hud_elements: stabilization_hud(stabilized),
        archive_dialogue: if stabilized then Some("Rotation stabilized. Initiating deceleration.") else None,
        phase_complete: stabilized,
        transition_to: None
    };

    (new_state, output)
}
```

---

## Asset Pipeline

### Planet Textures

```bash
# Download and process planetary images
assets/
├── planets/
│   ├── saturn.png          # 4096x4096, NASA Cassini
│   ├── saturn_rings.png    # 4096x1024, with alpha
│   ├── jupiter.png         # 4096x4096, NASA Juno
│   ├── mars.png            # 4096x4096, NASA composite
│   ├── earth.png           # 4096x4096, Blue Marble
│   └── moon.png            # 2048x2048
├── bridge/
│   ├── interior.png        # Bridge background
│   ├── stations.png        # Station sprites
│   └── crew_silhouettes.png
└── arrival/
    ├── starfield.png       # High-res starfield base
    └── effects/            # Additional VFX
```

### Manifest Entry

```json
{
  "planets": {
    "saturn": {"path": "planets/saturn.png", "scale": 1.0},
    "saturn_rings": {"path": "planets/saturn_rings.png", "scale": 1.2},
    "jupiter": {"path": "planets/jupiter.png", "scale": 1.0},
    "mars": {"path": "planets/mars.png", "scale": 1.0},
    "earth": {"path": "planets/earth.png", "scale": 1.0},
    "moon": {"path": "planets/moon.png", "scale": 0.27}
  }
}
```

---

## Testing Strategy

### Automated Scenarios

```go
// Scenario: Complete arrival sequence
scenario := Scenario{
    Name: "arrival_complete",
    Steps: []Step{
        // Emergence phase
        {Wait: 5 * time.Second},
        {Assert: PhaseIs(PhaseEmergence)},

        // Stabilization
        {Input: KeyHold(KeyW), Duration: 3 * time.Second},
        {Input: KeyHold(KeyA), Duration: 3 * time.Second},
        {Assert: PhaseIs(PhaseStabilizing)},
        {Wait: 2 * time.Second},
        {Assert: TumbleRatesBelow(5.0)},

        // Planet approaches
        {Input: KeyHold(KeySpace), Duration: 10 * time.Second},
        {Assert: PhaseIs(PhaseSaturn)},
        {Screenshot: "arrival_saturn.png"},

        // Continue through sequence...
        {Wait: 30 * time.Second},
        {Assert: PhaseIs(PhaseBridgeEstablished)},
        {Screenshot: "arrival_bridge.png"},
    },
}
```

### Golden Files

| Screenshot | Description | Velocity |
|------------|-------------|----------|
| `arrival_chaos.png` | Maximum SR distortion, tumbling | 0.99c |
| `arrival_saturn.png` | Saturn approach, clear SR effects | 0.95c |
| `arrival_jupiter.png` | Jupiter approach | 0.9c |
| `arrival_earth.png` | Earth approach, minimal SR | 0.5c |
| `arrival_bridge.png` | Bridge view established | 0.0c |

### Manual Testing Checklist

- [ ] Tumble feels disorienting but not nauseating
- [ ] Stabilization controls are responsive
- [ ] SR effects smoothly transition with velocity
- [ ] Planets look photorealistic
- [ ] Saturn's rings render correctly
- [ ] Earth evokes emotional response
- [ ] Bridge view layout is clear
- [ ] All stations are clickable
- [ ] Archive dialogue timing feels natural
- [ ] Total sequence duration ~4-5 minutes

---

## Performance Considerations

### Rendering Budget

| Component | Target | Notes |
|-----------|--------|-------|
| Starfield | 10k stars | Static texture + few animated |
| SR Shader | Full screen | Already optimized |
| Planet | 1 detailed at a time | LOD for distant |
| Bridge | Static + overlays | Minimal draw calls |

### Memory

- Planet textures: ~64MB (4x 4K textures)
- Starfield: ~16MB
- Bridge assets: ~8MB
- Total: ~100MB for arrival sequence

### Load Strategy

- Preload during title screen
- Stream planets as they approach
- Unload after Earth arrival (keep Earth texture for bridge)

---

## Success Criteria

### Core Experience
- [ ] Player feels genuine disorientation during emergence
- [ ] Gaining control feels earned
- [ ] SR effects are visually stunning
- [ ] Planets are breathtaking (especially Saturn and Earth)
- [ ] Bridge view establishes "home" feeling

### Tutorial Effectiveness
- [ ] Player learns stabilization controls naturally
- [ ] Deceleration mechanic is clear
- [ ] Time dilation concept is introduced
- [ ] Ship/Galaxy time distinction is understood

### Technical
- [ ] 60 FPS throughout sequence
- [ ] No visual glitches in SR shader
- [ ] Smooth velocity transitions
- [ ] Planet rendering without artifacts
- [ ] Audio syncs with visuals

### Emotional
- [ ] "Wow" moment at Saturn's rings
- [ ] Sense of scale at Jupiter
- [ ] Recognition/familiarity at Earth
- [ ] Weight of "centuries have passed" lands
- [ ] Curiosity about what comes next

---

## Future Extensions

| Feature | Description |
|---------|-------------|
| Skip option | For replays, skip to bridge |
| Alternate arrivals | Different solar systems for NG+ |
| Crew reactions | NPCs comment during sequence |
| Customizable velocity | Let player control approach speed more |
| Photo mode | Capture planetary vistas |
| VR support | Full 360° view during sequence |

---

## Implementation Plan

### Sprint 1: Core Framework

| Task | File | Description |
|------|------|-------------|
| 1.1 | `engine/arrival/sequence.go` | Phase state machine |
| 1.2 | `engine/arrival/camera.go` | Tumble + smooth camera |
| 1.3 | `engine/arrival/planets.go` | Basic planet rendering |
| 1.4 | Integration | Connect to existing SR shader |
| 1.5 | Test | Emergence → Stabilization works |

### Sprint 2: Polish & Bridge

| Task | File | Description |
|------|------|-------------|
| 2.1 | Assets | Download/process planet textures |
| 2.2 | `engine/arrival/bridge.go` | Bridge view layout |
| 2.3 | `engine/arrival/hud.go` | Time/velocity displays |
| 2.4 | Audio | Sound effects, Archive dialogue |
| 2.5 | Polish | Transitions, timing, feel |

---

**Created:** 2025-12-06
**Author:** Design Doc Creator Skill
