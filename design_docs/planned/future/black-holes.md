# Black Holes

**Status**: Planned
**Target**: v0.6.0 (Major Feature)
**Priority**: P1 - High
**Estimated**: Complex feature, multiple sprints
**Dependencies**: Galaxy Map, Time Dilation System, Crew Psychology, New Game+ Infrastructure

## Game Vision Alignment

**Score this feature against Stapledon's Voyage core pillars:**

| Pillar | Relevance | Score | Notes |
|--------|-----------|-------|-------|
| Time Dilation Consequence | Critical | +2 | Maximum time dilation mechanic; ultimate irreversibility |
| Civilization Simulation | Strong | +1 | Civs can worship, study, colonize, or fall into BHs |
| Philosophical Depth | Critical | +2 | Endings, cycles, Last Question resonance, existence beyond time |
| Ship & Crew Life | Strong | +1 | Peak psychological stress; mutiny trigger; OCEAN drift accelerant |
| Legacy Impact | Critical | +2 | New Game+ mechanism; universe-hopping; ultimate archive |
| Hard Sci-Fi Authenticity | Strong | +1 | GR-accurate; Higgs bubble enables but doesn't break physics |
| We Are Not Built For This | Critical | +2 | Human frailty at maximum; madness, refusal, breakdown |
| **Net Score** | | **+11** | **Decision: Move forward - cornerstone feature** |

**Feature type:** Gameplay (Meta-Game + Endgame + Navigation)

**Reference:** See [game-vision.md](../../docs/game-vision.md), [core-pillars.md](../../docs/vision/core-pillars.md)

## Problem Statement

**What problem does this solve?**

Black holes are:
1. **The ultimate time dilation node** - Players need extreme time-skip options beyond normal travel
2. **The replayability mechanism** - Game needs a New Game+ system that feels earned and mysterious
3. **The narrative frame** - Game starts and ends with black holes; they define the cycle
4. **A Fermi answer** - Time dilation explains why contact is so rare

**Current State:**
- No black hole mechanics exist
- No New Game+ system
- Game start/end narrative undefined
- Extreme time skips not possible

**Impact:**
- Without BHs, the game lacks its meta-structure (cycle between universes)
- Without extreme time skips, players can't access "deep future" scenarios
- Without the mystery of origin, the narrative lacks its philosophical core

## Goals

**Primary Goal:** Black holes serve as time-acceleration nodes, endgame choices, and the New Game+ mechanism that ties the entire experience together.

**Success Metrics:**
- Players can navigate to and orbit black holes at chosen radii
- Time dilation correctly calculated based on orbital radius
- Crew psychology responds to BH proximity (stress, opinions, potential mutiny)
- BH entry triggers end-of-run and seeds next universe
- Visual representation conveys the cosmic significance
- Player archives persist meaningfully across universe cycles

## Solution Design

### Overview

Black holes are special navigation destinations with three distinct functions:

1. **Time-Skip Nodes**: Park at chosen radius to skip millennia
2. **Endgame Choice**: Enter horizon to end playthrough and begin new cycle
3. **Narrative Frame**: Game starts from BH emergence; ending connects to beginning

### Core Mechanics

#### 1. Black Hole Types

| Type | Mass | Tidal Forces | Gameplay Role |
|------|------|--------------|---------------|
| **Stellar-mass** | 5-30 M☉ | Extreme | High-risk time skip; can kill careless ships |
| **Intermediate** | ~1000 M☉ | Moderate | Safe deliberate deep-future skips; civ megastructures |
| **Supermassive** | 10⁶-10⁹ M☉ | Negligible | Endgame hubs; pilgrimage destinations; safe horizon approach |

#### 2. Time Dilation Formula

At orbital radius `r` from Schwarzschild radius `r_s`:

```
t_external = t_ship / sqrt(1 - r_s/r)
```

Where `r_s = 2GM/c²`

| Orbital Radius | Dilation Factor | 1 Ship Hour = |
|----------------|-----------------|---------------|
| 10 r_s | 1.05x | ~1 hour external |
| 3 r_s | 1.22x | ~1.2 hours external |
| 1.5 r_s | 1.73x | ~1.7 hours external |
| 1.1 r_s | 3.16x | ~3 hours external |
| 1.01 r_s | 10x | ~10 hours external |
| 1.001 r_s | 31.6x | ~32 hours external |
| 1.0001 r_s | 100x | 4 days external |
| Closer... | 1000x+ | Years to millennia |

#### 3. Crew Psychology System

**Stress factors near BH:**
- Proximity to horizon (existential dread)
- Amount of external time passing (losing everything they knew)
- Duration of orbit (prolonged stress compounds)

**Archetype responses:**

| Archetype | Likely Response | Mutiny Risk |
|-----------|-----------------|-------------|
| Skeptic | "This is suicide with extra steps" | High |
| Medic | Concerned about crew mental health | Medium-High |
| Analyst | Calculates consequences coldly | Low |
| Zealot | "We're touching the face of God" | Low (eager) |
| Dreamer | Terror or ecstasy depending on OCEAN | Variable |
| Engineer | Focused on ship safety | Medium |

**Mutiny mechanics:**
- Crew vote triggered at extreme stress thresholds
- Outcomes: Abort, Compromise (shallower orbit), Override (captain authority)
- Successful mutiny = ship changes course; crew relations damaged
- Failed override = permanent crew trust damage

#### 4. New Game+ Mechanism

**Flow:**
```
Current universe → Enter BH horizon → Memory wipe (causality reversal)
    → New universe generated → Emerge from mysterious structure
    → 100 new years → Eventually enter BH → Cycle continues
```

**Universe inheritance (mysterious but beneficial):**
- Prior play influences new universe parameters
- Anthropic Luck (L) may shift based on actions
- γ cap may vary
- Philosophical archives may affect what philosophies emerge
- Exact mechanics hidden; patterns emerge over multiple runs

**What persists:**
- Meta-save tracks cycles completed
- Aggregate statistics (civs met, archives collected, etc.)
- Possible unlock: universe-hopper encounter chance increases

#### 5. Game Start Narrative

Every playthrough begins with emergence from a BH/mysterious structure:
- Player has archives of "Earth" but no proof it's real
- Memory loss justified by "causality reversal" hand-wave
- Stars in wrong positions; constellations unfamiliar
- The mystery of origin is never fully resolved
- Player IS the impossible alien to every civ they meet

### Architecture

**Components:**

1. **BlackHole Entity** (sim_gen/blackhole.go)
   - Type (stellar/intermediate/supermassive)
   - Mass, Schwarzschild radius
   - Position in galaxy
   - Orbiting civs/structures (optional)

2. **BH Navigation UI** (engine/display/bh_orbit.go)
   - Orbital radius selector with dilation preview
   - Time calculator: "X ship hours = Y external years"
   - Crew stress indicator
   - "Approach horizon" option (endgame)

3. **BH Renderer** (engine/render/blackhole.go)
   - Gravitational lensing effect
   - Accretion disk glow
   - Time dilation visualization (external stars streaking)
   - Event horizon as absolute darkness

4. **Crew BH Response** (sim_gen/crew_bh.go)
   - Stress calculation based on proximity/duration
   - Opinion generation per archetype
   - Mutiny vote trigger and resolution
   - OCEAN drift from BH exposure

5. **New Game+ System** (engine/meta/newgame.go)
   - Meta-save persistence
   - Universe parameter generation with inheritance
   - Cycle counter
   - Opening sequence (emergence)

### Implementation Plan

**Phase 1: Core Data & Navigation**
- [ ] Define BlackHole type in sim_gen
- [ ] Add BH generation to galaxy map
- [ ] Implement navigation to BH destinations
- [ ] Time dilation calculation at orbital radii
- [ ] Basic "park at radius" mechanic

**Phase 2: Visual & UI**
- [ ] BH visual representation (gravitational lensing)
- [ ] Accretion disk rendering
- [ ] Orbital radius selector UI
- [ ] Time skip preview calculator
- [ ] External time streaking effect during orbit

**Phase 3: Crew Psychology**
- [ ] BH proximity stress calculation
- [ ] Archetype opinion generation
- [ ] Mutiny vote system
- [ ] OCEAN drift acceleration near BH
- [ ] Dialogue for BH approach scenarios

**Phase 4: New Game+ Integration**
- [ ] "Enter horizon" option and confirmation
- [ ] Memory wipe / end sequence
- [ ] Meta-save system
- [ ] Universe parameter inheritance
- [ ] Opening "emergence" sequence
- [ ] Cycle tracking

**Phase 5: Polish & Edge Cases**
- [ ] Universe-hopper encounter (rare)
- [ ] Civ behavior near BHs (worship, colonize, fall in)
- [ ] BH as information horizon (extinct civ light echoes)
- [ ] Achievement/legacy tracking across cycles

### Files to Modify/Create

**New files:**
- `sim_gen/blackhole.go` - BlackHole type and physics calculations (~200 LOC)
- `sim_gen/crew_bh.go` - Crew psychology near BHs (~300 LOC)
- `engine/render/blackhole.go` - BH visual effects (~400 LOC)
- `engine/display/bh_orbit.go` - Orbital selection UI (~250 LOC)
- `engine/meta/newgame.go` - New Game+ system (~350 LOC)
- `engine/meta/metasave.go` - Cross-cycle persistence (~150 LOC)

**Modified files:**
- `sim_gen/world.go` - Add BlackHole to world state
- `sim_gen/galaxy.go` - BH generation in galaxy map
- `engine/scenario/scenario.go` - BH test scenarios
- `cmd/game/main.go` - Meta-save loading, cycle detection

## Examples

### Example 1: Time Skip Decision

**Player arrives at intermediate-mass BH**

UI shows:
```
═══════════════════════════════════════════
  SAGITTARIUS MINOR  (1,200 M☉)
═══════════════════════════════════════════

  Select orbital radius:

  [████████░░] 5.0 r_s  →  1 hour = 1.1 hours
  [██████░░░░] 2.0 r_s  →  1 hour = 1.4 hours
  [████░░░░░░] 1.5 r_s  →  1 hour = 1.7 hours
  [██░░░░░░░░] 1.1 r_s  →  1 hour = 3.2 hours
  [█░░░░░░░░░] 1.01 r_s →  1 hour = 10 hours

  PLANNED ORBIT: 12 ship hours at 1.01 r_s
  RESULT: 120 external hours (5 days)

  Crew Status: ████████░░ (Stable)

  [CONFIRM ORBIT]  [ADJUST]  [ABORT]
```

### Example 2: Crew Mutiny Scenario

**Player attempts extreme orbit near stellar-mass BH**

```
═══════════════════════════════════════════
  ⚠ CREW OBJECTION
═══════════════════════════════════════════

  Dr. Chen (Medic): "Three crew members are
  showing signs of acute dissociative episodes.
  I cannot recommend this course of action."

  Vasquez (Skeptic): "You're asking us to
  watch 10,000 years pass. Everyone we met
  is dead. This is psychological suicide."

  Brother Marcus (Zealot): "Let them object.
  I will stand with you at the threshold."

  ─────────────────────────────────────────
  CREW VOTE TRIGGERED

  In favor:    2  (Zealot, Analyst)
  Opposed:     4  (Skeptic, Medic, Dreamer, Engineer)

  [OVERRIDE (damages trust)]
  [COMPROMISE (orbit at 3 r_s instead)]
  [ABORT]
═══════════════════════════════════════════
```

### Example 3: New Game+ Transition

**Player enters supermassive BH horizon**

```
═══════════════════════════════════════════
  POINT OF NO RETURN
═══════════════════════════════════════════

  You are approaching the event horizon of
  CENTRUM ALPHA (4.2 million M☉).

  Beyond this point, return is impossible.

  Your archives contain:
  • 7 civilization records
  • 3 philosophical frameworks
  • 2 biological samples
  • Memory of Earth (unverified)

  External time will accelerate without limit.
  You will witness the end of stars.

  This universe will continue without you.

  [CROSS THE HORIZON]  [TURN BACK]
═══════════════════════════════════════════

      ▼ Player selects CROSS THE HORIZON ▼

═══════════════════════════════════════════

  [Screen fades to absolute black]

  [Time passes...]

  [Stars fade and die, one by one]

  [The last light goes out]

  [And then—]

═══════════════════════════════════════════

  [New universe generation screen]

  CYCLE 3

  You emerge from darkness.

  Stars burn in unfamiliar patterns.
  Your archives speak of a place called Earth.
  You have no memory of how you came here.

  You have 100 years.

  [BEGIN]
═══════════════════════════════════════════
```

## Success Criteria

- [ ] Three BH types spawn correctly in galaxy generation
- [ ] Time dilation calculates accurately based on orbital radius
- [ ] UI allows orbital radius selection with clear time preview
- [ ] Crew stress increases near BH; opinions vary by archetype
- [ ] Mutiny vote triggers at high stress; all three outcomes work
- [ ] BH entry ends playthrough and triggers new universe
- [ ] New universe parameters influenced by prior play (mysteriously)
- [ ] Opening "emergence" sequence plays for each new cycle
- [ ] Meta-save tracks cycles and aggregate stats
- [ ] BH visual effects convey cosmic significance
- [ ] Golden image tests for BH rendering

## Testing Strategy

**Unit tests:**
- Time dilation formula accuracy
- Crew stress calculation
- Universe parameter inheritance logic
- Mutiny vote resolution

**Integration tests:**
- Full orbit sequence (approach, park, exit)
- Full horizon crossing sequence
- New Game+ transition
- Meta-save persistence across sessions

**Visual tests (golden images):**
- BH rendering at various distances
- Accretion disk glow
- Gravitational lensing effect
- Time dilation streaking

**Manual testing:**
- Emotional impact of horizon crossing
- Crew dialogue quality
- Mystery/discovery pacing
- UI clarity for time skip decisions

## Non-Goals

**Not in this feature:**
- Kerr (rotating) BH ergosphere mechanics - Future expansion
- BH-to-BH wormholes - Violates physics premise
- Rescuing things from inside BH - Impossible by design
- Detailed "inside the BH" simulation - Hand-wave only

## Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|-----------|
| Time dilation math complexity | Medium | Pre-calculate lookup tables; validate against physics references |
| New Game+ feels like "reset" not "cycle" | High | Mysterious inheritance; narrative framing; archives matter |
| BH visuals too demanding | Medium | LOD system; simpler fallback for lower-end systems |
| Crew mutiny feels punishing | Medium | Make compromise viable; mutiny = character moment, not failure |
| Mystery becomes confusing | Medium | Enough hints to theorize; community discussion encouraged |

## References

- [Design Decisions: Black Hole Interview](../../docs/vision/design-decisions.md) (2025-12-02 entries)
- [Open Questions: BH Origin](../../docs/vision/open-questions.md)
- [Core Pillars](../../docs/vision/core-pillars.md) (including Pillar 6)
- [Interview Log: BH Deep Dive](../../docs/vision/interview-log.md)
- Lee Smolin - Cosmological Natural Selection (theoretical basis for BH → new universe)
- Asimov - "The Last Question" (thematic inspiration)

## Future Work

- **Ergosphere mechanics** - Rotating BH energy extraction for advanced civs
- **BH archaeology** - Light echoes from civs that fell in
- **Universe-hopper encounters** - Meeting travelers from other dead universes
- **BH cult civilizations** - Civs that worship or colonize BH regions
- **Accretion disk habitats** - Extreme post-biological life

---

**Document created**: 2025-12-02
**Last updated**: 2025-12-02
**Origin**: Vision interview deep dive on black hole mechanics
