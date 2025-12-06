# Black Hole Mechanics

## Status
- **Status:** Planned
- **Sprint:** Vision Integration - Sprint 1
- **Priority:** P1 (Core narrative mechanic)
- **Source:** [Interview: Black Hole Deep Dive](../../../docs/vision/interview-log.md#2025-12-02-session-black-hole-feature-deep-dive)

## Game Vision Alignment

| Pillar | Score | Notes |
|--------|-------|-------|
| Choices Are Final | ✅ Strong | BH entry is ultimate permanent choice |
| Game Doesn't Judge | ✅ Strong | BH entry is valid path, not failure |
| Time Has Emotional Weight | ✅ Strong | Millennia pass - everyone you knew is dead |
| Ship Is Home | ⚪ N/A | External mechanic |
| Grounded Strangeness | ✅ Strong | Real physics, extreme consequences |
| We Are Not Built For This | ✅ Strong | Crew psychology breaks near BH |

## Feature Overview

Black holes serve **three distinct functions** in the game:

### 1. Time Weapon
Park near a BH to skip millions of years via gravitational time dilation. The sacrifice IS the emotional weight:
- You lose all connection to the present universe
- Your data becomes archaeological record
- Crew can mutiny to prevent/limit approach
- Some crew go insane watching external time race ahead

### 2. Endgame Choice
BH approach creates the ultimate decision point:
- **Stay:** Witness universe evolution, preserve continuity
- **Enter:** Abandon everything, hope for something beyond

The game presents both as valid. The "Last Star" path (Asimov's "The Last Question") is as legitimate as any other.

### 3. New Game+ Mechanism
BH entry IS the New Game+ system:
- You don't escape consequences - you abandon your universe entirely
- New universe parameters are influenced by what you carried in
- Influence is "mysterious but clearly beneficial" - patterns emerge over many runs
- Avoids optimization while preserving discovery

## Design Decisions

From [design-decisions.md](../../../docs/vision/design-decisions.md):

| Decision | Summary |
|----------|---------|
| BH as Time Weapon | Time skip costs emotional severance, not resources |
| Crew Psychology at BH | Peak stress moment, possible mutiny trigger |
| BH Entry = New Game+ | Abandon universe, seed next with mystery influence |
| Influence Transparency | Mysterious but clearly beneficial |

## Crew Psychology Near Black Holes

This is one of the few moments where human-scale directly conflicts with player agency:

| Archetype | Response | Mutiny Risk |
|-----------|----------|-------------|
| **Skeptic** | Resists approach, wants proof of benefit | High |
| **Medic** | Worried about psychological damage | Medium |
| **Zealot** | Embraces it dangerously, wants closer | Low (different risk) |
| **Navigator** | Calculates risks, may support limited approach | Variable |

**Mutiny Triggers:**
- Approaching too close without crew buy-in
- Forcing deep-time skip against majority will
- Ignoring warnings about psychological damage

## Time Dilation Mechanics

### Near-BH Time Skip

At various distances from a Schwarzschild black hole:

| Distance (rs) | Time Ratio | 1 Year Ship = |
|---------------|------------|---------------|
| 10 rs | 1.05x | 1.05 years external |
| 3 rs | 1.22x | 1.22 years external |
| 1.5 rs | 2x | 2 years external |
| 1.1 rs | ~3x | 3 years external |
| 1.01 rs | ~10x | 10 years external |

**Note:** These are approximations. Real time dilation follows:
```
t_external = t_ship / sqrt(1 - rs/r)
```

For gameplay, we'll use simplified tiers:
- **Safe orbit:** Minor time dilation (educational)
- **Deep dive:** Significant skip (decades to centuries)
- **Event horizon approach:** Extreme skip (millennia+)

### Post-Skip Consequences

After a significant time skip:
- All previous civilizations have evolved/died/ascended
- Your archives become the only record of what was
- Cultural/political landscape completely different
- "You ARE the record" - your data is archaeological treasure

## BH Entry Sequence

### Pre-Entry
1. Approach to safe orbit (crew stress begins)
2. Decision point: enter or retreat
3. If entering: final opportunity for crew mutiny
4. If proceeding: point of no return

### Entry Experience
- Visual: extreme GR lensing, time distortion effects
- Audio: ship systems straining, crew reactions
- Crew dialogue: final thoughts, fears, hopes
- The crossing itself is ambiguous - fade to unknown

### Post-Entry (New Game+)
- Emergence from "mysterious structure" in new universe
- No memory of previous universe explicitly
- Subtle influence on starting conditions
- Archive has strange errors/artifacts (clues)

## Open Questions

From [open-questions.md](../../../docs/vision/open-questions.md):

1. **BH Origin Explicit/Implicit?** - Does player know they emerged from BH?
2. **Universe-Hopper Rarity** - How rare are encounters with others like you?
3. **Influence Transparency** - How much does prior play affect next universe?
4. **GR+SR Visual Interaction** - How do effects combine at high speed near BH?

## AILANG Considerations

### Types

```ailang
type BHApproachPhase =
    | SafeOrbit(float)      -- distance in rs
    | DeepDive(float)       -- danger zone
    | EventHorizon          -- point of no return
    | Entry                 -- crossing threshold

type BHCrewReaction =
    | Supportive
    | Reluctant
    | Resistant
    | Mutinous

type TimeDilationResult = {
    ship_time: float,
    external_time: float,
    crew_stress_delta: float,
    archive_degradation: float
}
```

### Functions

```ailang
-- Calculate time dilation at distance
pure func time_dilation_factor(distance_rs: float) -> float

-- Get crew reaction to BH approach
pure func crew_bh_reaction(crew: Crew, phase: BHApproachPhase) -> BHCrewReaction

-- Check if mutiny triggers
pure func check_bh_mutiny(society: BubbleSociety, phase: BHApproachPhase) -> bool

-- Process time skip consequences
pure func apply_time_skip(world: World, ship_years: float, external_years: float) -> World
```

## Engine Integration

### Visual Effects
- GR lensing shader intensifies with proximity
- Time distortion overlay (external stars streaking)
- Event horizon visualization (darkness, Einstein ring)

### Audio
- Ship stress sounds
- Time-warped external universe audio
- Crew tension/dialogue cues

### UI
- Distance indicator (rs units)
- Time ratio display
- Crew stress meters
- Mutiny warning indicators

## Testing Scenarios

1. **Safe Approach:** Orbit at 10rs, observe minor dilation, retreat safely
2. **Deep Dive:** Approach 1.5rs, skip centuries, crew stress but no mutiny
3. **Mutiny Trigger:** Push to 1.1rs against crew wishes, trigger mutiny
4. **Entry Sequence:** Complete BH entry, verify New Game+ trigger

## Success Criteria

- [ ] BH approach creates meaningful tension
- [ ] Crew psychology responds realistically to proximity
- [ ] Time skip consequences feel impactful
- [ ] Mutiny system integrates with BH decisions
- [ ] New Game+ feels earned, not random
- [ ] Visual/audio create appropriate awe and dread
