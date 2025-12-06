# Sprint Plan: Arrival Sequence & Bridge View

**Sprint ID:** `arrival-sequence`
**Target Version:** 0.2.0
**Duration:** 2 sprints (8-10 working days)
**Design Doc:** [arrival-sequence.md](next/arrival-sequence.md)

## Executive Summary

Implement the first playable experience: player emerges spinning in the solar system, gains control, tours planets (Saturn → Jupiter → Mars → Earth) with SR effects, and arrives at the Bridge View—the main gameplay screen.

## Current State Assessment

### What We Have
| Component | Status | Location |
|-----------|--------|----------|
| SR Shader | ✅ Implemented | `engine/shader/sr.go` |
| Star rendering | ✅ Basic | `engine/render/draw_stars.go` |
| Galaxy background | ✅ 4K image | `assets/data/starmap/background/galaxy_4k.jpg` |
| Input system | ✅ Working | `engine/input/input.go` |
| Display manager | ✅ Working | `engine/display/manager.go` |
| Game loop | ✅ Working | `cmd/game/main.go` |

### What We Need
| Component | Priority | Effort |
|-----------|----------|--------|
| Planet textures | P0 | Download + process |
| Arrival state machine | P0 | New package |
| Camera tumble system | P0 | New |
| Planet renderer | P0 | New |
| Bridge view layout | P1 | New |
| HUD elements | P1 | New |
| Audio integration | P2 | Extend existing |

---

## Sprint 1: Core Sequence (5 days)

**Goal:** Playable emergence → Saturn arrival with SR effects

### Day 1: Asset Acquisition & Project Setup

**Morning: Planet Textures**
- [ ] Download NASA planetary images (public domain)
  - Saturn + rings: NASA Cassini
  - Jupiter: NASA Juno
  - Mars: NASA composite
  - Earth: Blue Marble
  - Moon: LRO
- [ ] Process to 2K resolution (4K optional later)
- [ ] Create `assets/planets/` directory structure

**Afternoon: Package Scaffold**
- [ ] Create `engine/arrival/` package
  - `sequence.go` - State machine
  - `camera.go` - Tumble camera
  - `planets.go` - Planet rendering
  - `hud.go` - Arrival HUD
- [ ] Add arrival phase enum to `sim_gen/` (mock mode)

**Scripts:**
```bash
# Planet download script (NASA public domain)
mkdir -p assets/planets
# Saturn - Cassini
curl -o assets/planets/saturn.jpg "https://photojournal.jpl.nasa.gov/jpeg/PIA17172.jpg"
# Jupiter - Juno
curl -o assets/planets/jupiter.jpg "https://photojournal.jpl.nasa.gov/jpeg/PIA21974.jpg"
# Mars - Viking
curl -o assets/planets/mars.jpg "https://photojournal.jpl.nasa.gov/jpeg/PIA00407.jpg"
# Earth - Blue Marble
curl -o assets/planets/earth.jpg "https://eoimages.gsfc.nasa.gov/images/imagerecords/57000/57723/land_ocean_ice_2048.jpg"
```

**Verification:**
- [ ] All planet images load correctly
- [ ] Package compiles with `go build ./...`

---

### Day 2: Camera Tumble & Stabilization

**Morning: Tumble Camera**
- [ ] Implement `ArrivalCamera` with pitch/yaw/roll rates
- [ ] Add tumble damping based on input
- [ ] Create view matrix from Euler angles

```go
// engine/arrival/camera.go
type ArrivalCamera struct {
    PitchRate, YawRate, RollRate float64  // deg/sec
    Pitch, Yaw, Roll             float64  // current angles
}

func (c *ArrivalCamera) ApplyTumble(dt float64)
func (c *ArrivalCamera) ApplyDamping(input InputDamping, dt float64)
func (c *ArrivalCamera) GetViewMatrix() mgl64.Mat4
```

**Afternoon: Input Integration**
- [ ] Extend `engine/input/` for stabilization controls
- [ ] WASD reduces rotation rates
- [ ] Visual feedback (rotation indicators)

**Verification:**
- [ ] Camera tumbles convincingly at game start
- [ ] WASD input dampens rotation
- [ ] Rotation stops when stabilized (< 5 deg/sec all axes)

---

### Day 3: State Machine & SR Integration

**Morning: Arrival State Machine**
- [ ] Implement phase transitions
  - `Emergence` (tumbling, high velocity)
  - `Stabilizing` (player gaining control)
  - `ApproachingPlanet` (deceleration, SR visible)
  - `BridgeEstablished` (sequence complete)
- [ ] Velocity curve (0.99c → 0.95c → 0.9c → 0.5c → 0)

```go
// engine/arrival/sequence.go
type ArrivalPhase int
const (
    PhaseEmergence ArrivalPhase = iota
    PhaseStabilizing
    PhaseSaturn
    PhaseJupiter
    PhaseMars
    PhaseEarth
    PhaseBridge
)
```

**Afternoon: SR Shader Connection**
- [ ] Connect arrival velocity to existing SR shader
- [ ] Test velocity transitions visually
- [ ] Ensure smooth aberration/Doppler changes

**Verification:**
- [ ] Phase transitions work correctly
- [ ] SR effects match velocity (0.99c = extreme, 0.5c = subtle)
- [ ] Smooth interpolation between velocities

---

### Day 4: Planet Rendering

**Morning: Basic Planet Display**
- [ ] Load planet textures via asset manager
- [ ] Render planet as scaled sprite/billboard
- [ ] Distance-based sizing (point → disk → full)

```go
// engine/arrival/planets.go
type Planet struct {
    Name     string
    Texture  *ebiten.Image
    Position Vec3  // Solar system coords
    Radius   float64
}

func (p *Planet) Draw(screen *ebiten.Image, camera *ArrivalCamera, distance float64)
```

**Afternoon: Saturn with Rings**
- [ ] Implement ring rendering (separate texture with alpha)
- [ ] Ring tilt relative to camera
- [ ] Saturn as first visual landmark

**Verification:**
- [ ] Saturn visible during approach
- [ ] Rings render correctly with transparency
- [ ] Planet grows as distance decreases

---

### Day 5: Integration & Sprint 1 Demo

**Morning: Full Sequence Integration**
- [ ] Wire arrival sequence into game loop
- [ ] Entry point: game starts in arrival mode
- [ ] Exit point: transition to existing game state

**Afternoon: Polish & Testing**
- [ ] Tune timing (emergence duration, planet approach)
- [ ] Add screen shake during chaos phase
- [ ] Test complete flow: start → Saturn
- [ ] Create golden screenshot at Saturn approach

**Verification:**
- [ ] Complete sequence plays from start
- [ ] SR effects are visually impressive
- [ ] Saturn landmark is recognizable
- [ ] No crashes or visual glitches

**Sprint 1 Deliverable:**
```bash
make run  # Game starts with arrival sequence
# Player experiences: tumble → stabilize → approach Saturn
# SR effects visible and working
```

---

## Sprint 2: Polish & Bridge View (5 days)

**Goal:** Complete planetary tour, establish Bridge View as main gameplay screen

### Day 6: Remaining Planets

**Morning: Jupiter**
- [ ] Jupiter texture and rendering
- [ ] Great Red Spot positioning
- [ ] Transition from Saturn phase

**Afternoon: Mars & Earth**
- [ ] Mars rendering (simpler, smaller)
- [ ] Earth as climax (Blue Marble beauty)
- [ ] Moon visible near Earth
- [ ] Emotional beat: "Home" arrival

**Verification:**
- [ ] All 4 planets render correctly
- [ ] Visual progression satisfying
- [ ] Earth arrival feels significant

---

### Day 7: Bridge View Layout

**Morning: Screen Layout**
- [ ] Define viewport split (60% space, 40% bridge)
- [ ] Space view renders above
- [ ] Bridge interior placeholder below

```go
// engine/arrival/bridge.go
type BridgeView struct {
    SpaceViewport  image.Rectangle  // Top 60%
    BridgeViewport image.Rectangle  // Bottom 40%
}
```

**Afternoon: HUD Elements**
- [ ] Ship time / Galaxy time displays
- [ ] Velocity indicator (β, γ)
- [ ] Phase indicator during sequence

**Verification:**
- [ ] Screen layout looks correct
- [ ] HUD readable and informative
- [ ] Time displays update correctly

---

### Day 8: Bridge Interior (Basic)

**Morning: Station Placeholders**
- [ ] Define bridge station positions
  - Navigation console (→ Galaxy Map)
  - Archive terminal
  - Communications
  - Systems status
- [ ] Clickable hotspots

**Afternoon: Visual Polish**
- [ ] Simple bridge background
- [ ] Station icons/sprites
- [ ] Hover highlighting

**Verification:**
- [ ] Stations are clickable
- [ ] Visual feedback on hover
- [ ] Layout is clear and readable

---

### Day 9: Audio & Archive Dialogue

**Morning: Sound Effects**
- [ ] Ship alarm sounds (emergence chaos)
- [ ] Engine hum (ambient)
- [ ] Stabilization success sound
- [ ] System chirps

**Afternoon: Archive Voice**
- [ ] Placeholder text dialogue
- [ ] Timed dialogue triggers
  - "Rotation stabilized"
  - Planet announcements
  - "Earth. Home."
  - Year revelation

**Verification:**
- [ ] Audio plays at correct moments
- [ ] Archive dialogue displays
- [ ] Timing feels natural

---

### Day 10: Final Integration & Testing

**Morning: Complete Flow**
- [ ] Full sequence: emergence → bridge established
- [ ] First choice presentation
- [ ] Transition to game modes (placeholder)

**Afternoon: Testing & Documentation**
- [ ] Create test scenario for arrival
- [ ] Golden screenshots at key moments
- [ ] Update design doc status
- [ ] Performance profiling

**Verification:**
- [ ] Complete experience is compelling
- [ ] 60 FPS throughout
- [ ] No visual glitches
- [ ] Documentation updated

**Sprint 2 Deliverable:**
```bash
make run  # Complete arrival sequence
# Emergence → Stabilize → Saturn → Jupiter → Mars → Earth → Bridge
# SR effects throughout, Archive dialogue, HUD working
# Bridge view established as main game screen
```

---

## AILANG Integration (Future)

For now, arrival sequence is pure Go (engine-side). Future AILANG integration:

```ailang
-- sim/arrival.ail
module sim/arrival

type ArrivalPhase =
    | Emergence | Stabilizing
    | ApproachingSaturn | ApproachingJupiter
    | ApproachingMars | ApproachingEarth
    | BridgeEstablished | Complete

type ArrivalState = {
    phase: ArrivalPhase,
    velocity: float,
    pitch_rate: float,
    yaw_rate: float,
    roll_rate: float
}
```

**Not blocking for v0.2.0** - Can migrate to AILANG later.

---

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Planet textures too large | Medium | Performance | Use 2K, LOD system |
| SR shader + planets conflict | Low | Visual bugs | Test early, separate layers |
| Timing feels wrong | Medium | UX | Iterate with playtesting |
| Camera tumble nauseating | Medium | UX | Tune rates, add comfort options |

---

## Success Criteria

### Sprint 1
- [ ] Arrival sequence starts on game launch
- [ ] Tumble → stabilization feels earned
- [ ] SR effects are visually stunning
- [ ] Saturn approach is recognizable

### Sprint 2
- [ ] All 4 planets render beautifully
- [ ] Earth arrival is emotional highlight
- [ ] Bridge view is clear and functional
- [ ] Archive dialogue enhances experience
- [ ] Ready for main gameplay integration

---

## Files to Create/Modify

### New Files
| File | Purpose |
|------|---------|
| `engine/arrival/sequence.go` | State machine, main logic |
| `engine/arrival/camera.go` | Tumble camera system |
| `engine/arrival/planets.go` | Planet rendering |
| `engine/arrival/bridge.go` | Bridge view layout |
| `engine/arrival/hud.go` | Arrival-specific HUD |
| `assets/planets/*.jpg` | Planetary textures |
| `assets/sounds/arrival/*.wav` | Arrival sound effects |

### Modified Files
| File | Change |
|------|--------|
| `cmd/game/main.go` | Add arrival mode entry |
| `engine/shader/effects.go` | Expose SR velocity control |
| `engine/assets/manager.go` | Load planet textures |

---

## Post-Sprint Review

After completion:
1. Update design doc status to "Implemented"
2. Move to `design_docs/implemented/v0_2_0/`
3. Send AILANG feedback if applicable
4. Create demo video/GIF for documentation

---

**Created:** 2025-12-06
**Sprint Planner Skill**
