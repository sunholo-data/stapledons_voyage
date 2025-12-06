# Journey Planning UI

**Status**: Planned
**Target**: v0.4.0
**Priority**: P1 - High
**Estimated**: 1 week
**Dependencies**: Starmap Data Model, Planet State Transitions

## Game Vision Alignment

| Pillar | Relevance | Score | Notes |
|--------|-----------|-------|-------|
| Time Dilation Consequence | + | +1 | Shows exactly what journey costs in subjective/external time |
| Civilization Simulation | + | +1 | Shows predicted civ states at arrival |
| Philosophical Depth | + | +1 | The commitment moment - knowing the cost, choosing anyway |
| Ship & Crew Life | + | +1 | Shows crew age at arrival, who might not survive |
| Legacy Impact | + | +1 | Shows years remaining, what you're giving up |
| Hard Sci-Fi Authenticity | + | +1 | Real relativistic math, no hand-waving |
| **Net Score** | | **+6** | **Decision: Move forward** |

**Feature type:** Gameplay (core interaction loop)

## Problem Statement

The journey planning screen is where Pillar 1 (Choices Are Final) becomes visceral. The player must:
- See all available destinations
- Understand the time costs (subjective and external)
- See probability distributions of what they'll find
- **Commit** knowing they cannot turn back

**Current State:**
- No journey planning exists
- Need to make the commitment feel weighty but not paralyzing

**Impact:**
- This IS the game loop: plan → commit → travel → arrive → consequences
- Poor UX here ruins the entire experience
- Must balance information density with clarity

## Goals

**Primary Goal:** Create a journey planning interface that makes relativistic travel decisions feel consequential, informed, and irreversible.

**Success Metrics:**
- Player understands time costs before committing
- Predictions are visible but uncertainty is clear
- Commitment feels weighty (no accidental clicks)
- Information is scannable (not overwhelming)

## Solution Design

### Overview

Three-panel layout:
1. **Star Map** (left) - 3D navigable galaxy view with targets
2. **Target Details** (right) - Selected destination information
3. **Journey Calculator** (bottom) - Time dilation math and commitment

### Star Map Panel

```
┌──────────────────────────────────────────────────────────────┐
│  ◉ Sol (You are here)                                        │
│                                                              │
│       ○ Kepler-442  (1200 ly)                               │
│         [Biosignature likely]                                │
│                                                              │
│                    ○ HD 40307  (800 ly)                      │
│                      [Technosignature 90%]                   │
│                                                              │
│  ○ Tau Ceti (12 ly)                                         │
│    [No data]                    ○ Gliese 667  (23 ly)       │
│                                   [Biosphere possible]       │
│                                                              │
│         ◎ Trappist-1 (40 ly)                                │
│           [Multiple HZ planets]                              │
│                                                              │
│  [Drag to rotate]  [Scroll to zoom]  [Click to select]      │
└──────────────────────────────────────────────────────────────┘
```

**Visual indicators:**
- Circle size = detection confidence
- Color coding:
  - White: No data / Sterile
  - Green: Biosignature
  - Blue: Technosignature (active)
  - Yellow: Technosignature (old/uncertain)
  - Red: Ruins (confirmed)
  - Purple: Unknown/Ambiguous
- Reachable bubble shows "within N subjective years" radius

### Target Details Panel

```
┌────────────────────────────────────────┐
│  HD 40307g                             │
│  ════════════════════════════════════  │
│                                        │
│  Distance: 800 ly                      │
│  Star type: K2.5V (orange dwarf)       │
│  Planet: Super-Earth, 1.8 M⊕          │
│                                        │
│  ─── Detection History ───             │
│                                        │
│  800 years ago (last light):           │
│  • Narrowband radio detected           │
│  • Industrial waste heat signature     │
│  • Confidence: 90% technological civ   │
│                                        │
│  ─── Predicted State at Arrival ───    │
│                                        │
│  Time since observation: 1600 years    │
│                                        │
│  ┌─────────────────────────────────┐   │
│  │ ████████████████░░░░ 45% Tech   │   │
│  │ ██████████░░░░░░░░░░ 30% Ruins  │   │
│  │ █████░░░░░░░░░░░░░░░ 15% Trans  │   │
│  │ ███░░░░░░░░░░░░░░░░░ 10% Other  │   │
│  └─────────────────────────────────┘   │
│                                        │
│  "Strong industrial signatures 800     │
│   years ago. Civilizations at this     │
│   stage are historically unstable.     │
│   They may have advanced, collapsed,   │
│   or transcended by your arrival."     │
│                                        │
└────────────────────────────────────────┘
```

### Journey Calculator Panel

```
┌──────────────────────────────────────────────────────────────────────────┐
│  JOURNEY TO HD 40307g                                                    │
│  ════════════════════════════════════════════════════════════════════════│
│                                                                          │
│  Speed: [━━━━━━━━━●━━━━] γ = 20 (0.999c)                                │
│                                                                          │
│  ┌────────────────────────────────────────────────────────────────────┐  │
│  │                                                                    │  │
│  │   YOU                           UNIVERSE                           │  │
│  │   ───                           ────────                           │  │
│  │   Travel time:    40 years      Travel time:    800 years         │  │
│  │   Age at arrival: 67 years      Year at arrival: 2825             │  │
│  │   Years remaining: 33 years     Earth elapsed:   800 years        │  │
│  │                                                                    │  │
│  └────────────────────────────────────────────────────────────────────┘  │
│                                                                          │
│  ─── What You're Committing To ───                                       │
│                                                                          │
│  • 40 years of your remaining 73 will be spent in transit                │
│  • Earth will be 800 years older when you arrive                         │
│  • Any civ you uplifted will have had 800 years to spread                │
│  • You CANNOT turn back without spending another 40 years                │
│                                                                          │
│  ─── Crew Impact ───                                                     │
│                                                                          │
│  • Dr. Chen (age 45): Will be 85 at arrival, may not survive journey    │
│  • Navigator Okafor (age 32): Will be 72, likely your last voyage       │
│  • 3 children born en route will know no other home                      │
│                                                                          │
│  ┌─────────────────────────────────────────────────────────────────────┐ │
│  │                                                                     │ │
│  │   [ CANCEL ]                    [ COMMIT TO JOURNEY ]              │ │
│  │                                                                     │ │
│  │              ⚠ This decision cannot be undone                      │ │
│  │                                                                     │ │
│  └─────────────────────────────────────────────────────────────────────┘ │
│                                                                          │
└──────────────────────────────────────────────────────────────────────────┘
```

### Confirmation Dialog

When "COMMIT TO JOURNEY" is clicked:

```
╔════════════════════════════════════════════════════════════╗
║                                                            ║
║                    CONFIRM DEPARTURE                       ║
║                                                            ║
║  You are about to commit to a journey to HD 40307g.        ║
║                                                            ║
║  This will cost:                                           ║
║  • 40 subjective years of your 73 remaining                ║
║  • 800 years of external time                              ║
║                                                            ║
║  You cannot:                                               ║
║  • Turn back                                               ║
║  • Change destination                                      ║
║  • Communicate with anyone outside the ship                ║
║                                                            ║
║  The universe will continue without you.                   ║
║  When you arrive, everything may have changed.             ║
║                                                            ║
║                                                            ║
║   Type "DEPART" to confirm: [___________]                  ║
║                                                            ║
║   [ Cancel ]                                               ║
║                                                            ║
╚════════════════════════════════════════════════════════════╝
```

### AILANG Types

```ailang
type JourneyPlan = {
    origin: Star,
    destination: Star,
    gamma: float,              -- Lorentz factor (speed selection)
    departureYear: int,        -- External year of departure
    arrivalYearExternal: int,  -- External year of arrival
    travelTimeSubjective: int, -- Ship years
    crewAgesAtArrival: [(CrewMember, int)]
}

type JourneyCommitment = {
    plan: JourneyPlan,
    playerAge: int,
    yearsRemaining: int,
    predictions: [(PlanetState, float)]  -- State distribution
}

pure func calculateJourney(
    origin: Star,
    dest: Star,
    gamma: float,
    departureYear: int
) -> JourneyPlan

pure func canAfford(plan: JourneyPlan, playerAge: int) -> bool
```

### Speed Selection

The γ slider is the key player agency:

| γ | Speed | 100 ly trip (subjective) | 100 ly trip (external) |
|---|-------|--------------------------|------------------------|
| 5 | 0.98c | 20 years | 100 years |
| 10 | 0.995c | 10 years | 100 years |
| 20 | 0.999c | 5 years | 100 years |
| 50 | 0.9998c | 2 years | 100 years |

Trade-off: Higher γ = less subjective time, but same external time. You arrive younger, but the universe ages the same amount.

**Why this matters:**
- Lower γ: You age more, but you have more "presence" during the journey (crew interactions, ship management)
- Higher γ: You age less, but the journey feels like a blink

### Reachable Bubble Visualization

Show a translucent sphere around current position:

```ailang
pure func reachableBubble(
    position: Vec3,
    gamma: float,
    yearsRemaining: int
) -> float {
    -- Maximum reachable distance in light-years
    gamma * intToFloat(yearsRemaining)
}
```

Stars outside this bubble are grayed out with tooltip: "Beyond remaining lifetime at current γ"

### Implementation Plan

**Phase 1: Star Map Rendering** (~2 days)
- [ ] 3D star map with camera controls (rotate, zoom, pan)
- [ ] Star icons with color coding by detection type
- [ ] Selection highlighting
- [ ] Reachable bubble overlay

**Phase 2: Target Details** (~1.5 days)
- [ ] Detection history display
- [ ] Probability distribution bars
- [ ] Narrative prediction text

**Phase 3: Journey Calculator** (~2 days)
- [ ] γ slider with real-time calculation
- [ ] Dual time display (subjective/external)
- [ ] Crew impact predictions
- [ ] "What you're committing to" summary

**Phase 4: Commitment Flow** (~1.5 days)
- [ ] Confirmation dialog
- [ ] Type-to-confirm mechanic
- [ ] Transition to journey state
- [ ] Point of no return enforcement

### Files to Modify/Create

**New files:**
- `engine/ui/starmap_view.go` - 3D star map rendering (~500 LOC)
- `engine/ui/target_panel.go` - Target details panel (~300 LOC)
- `engine/ui/journey_calc.go` - Journey calculator (~400 LOC)
- `engine/ui/commit_dialog.go` - Confirmation flow (~200 LOC)
- `sim/journey.ail` - Journey planning logic (~200 LOC)

**Modified files:**
- `engine/render/camera.go` - 3D camera for star map
- `sim/world.ail` - Add journey state

## Examples

### Example 1: Short Local Trip

```
Destination: Tau Ceti (12 ly)
γ selected: 10

Journey calculation:
- Subjective time: 1.2 years
- External time: 12 years
- Player age: 27 → 28
- Years remaining: 73 → 72

Prediction: No data (first survey)

Commitment text:
"A short hop to our nearest neighbor. The universe
will barely notice your absence. Earth will be 12
years older - your family may still remember you."
```

### Example 2: Deep Galaxy Expedition

```
Destination: Distant technosignature (5000 ly)
γ selected: 50 (maximum)

Journey calculation:
- Subjective time: 100 years (!)
- External time: 5000 years
- Player age: 27 → 127 (impossible)

WARNING: This journey exceeds your remaining lifetime.
You will die en route. Your descendants may complete
the voyage if the ship survives.

Commitment text:
"This is a generational voyage. You will not see
the destination. Your great-grandchildren might.
Earth will be myth by the time anyone arrives."
```

## Success Criteria

- [ ] Player can navigate star map intuitively
- [ ] Time calculations are correct (verified against physics)
- [ ] Predictions display with appropriate uncertainty
- [ ] Commitment feels weighty (no accidental commits)
- [ ] Crew impact is emotionally resonant
- [ ] Journey state properly locks in

## Testing Strategy

**Unit tests:**
- Relativistic time calculations correct
- Journey affordability check works
- Prediction aggregation correct

**Integration tests:**
- Full flow: select → plan → commit → journey state
- Cancellation returns to planning
- Locked journey cannot be cancelled

**UX testing:**
- Players understand what they're committing to
- No accidental commits in 20 test sessions
- Time to understand UI < 2 minutes

## Non-Goals

**Not in this feature:**
- Multi-leg journey planning (one destination at a time)
- Fuel/resource constraints on journey
- Mid-journey events (separate system)
- Return journey planning (handled after arrival)

## Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|-----------|
| Information overload | High | Progressive disclosure; key stats prominent |
| Accidental commits | High | Type-to-confirm; multiple click steps |
| Math feels opaque | Med | Show formula; explain in plain language |
| Paralysis by analysis | Med | Highlight recommended targets; show "interesting" markers |

## References

- [startmaps.md](startmaps.md) - Navigation and time dilation design
- [planet-state-transitions.md](planet-state-transitions.md) - Prediction system
- [design-decisions.md](../../docs/vision/design-decisions.md) - Choices Are Final pillar

## Future Work

- Multi-stop voyage planning (waypoints)
- Comparison mode (side-by-side destinations)
- Historical journey log (where have I been)
- Crew preference weighting (some want to return to Earth)
- "What if" mode for exploring consequences

---

**Document created**: 2025-12-02
**Last updated**: 2025-12-02
