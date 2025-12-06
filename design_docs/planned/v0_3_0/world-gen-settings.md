# World Generation Settings (Drake Parameters UI)

**Status**: Planned
**Target**: v0.3.0
**Priority**: P1 - High
**Estimated**: 3 days
**Dependencies**: Starmap Data Model

## Game Vision Alignment

| Pillar | Relevance | Score | Notes |
|--------|-----------|-------|-------|
| Time Dilation Consequence | + | +1 | Civ density determines how far you must go to find them |
| Civilization Simulation | + | +1 | Core to how many civs exist and where |
| Philosophical Depth | + | +1 | Player chooses what kind of universe they believe in |
| Ship & Crew Life | 0 | 0 | Indirect - affects what crew will encounter |
| Legacy Impact | + | +1 | Sparse vs dense galaxy changes what legacy is possible |
| Hard Sci-Fi Authenticity | + | +1 | Parameters grounded in real Drake equation / astrobiology |
| **Net Score** | | **+5** | **Decision: Move forward** |

**Feature type:** Gameplay (player makes meaningful choice about universe)

## Problem Statement

The game needs a world generation screen where players configure:
- How dense civilizations are (Anthropic Luck factor)
- What the galaxy "feels like" (lonely vs crowded)
- Implicit difficulty (sparse = harder to find civs, less network effects)

**Current State:**
- No world gen exists
- Need to balance scientific grounding with accessibility

**Impact:**
- Sets the entire experience for a playthrough
- Choice is permanent (Pillar 1: Choices Are Final)
- Different settings create different "stories"

## Goals

**Primary Goal:** Create a world generation UI that lets players configure galaxy parameters in a scientifically grounded but accessible way.

**Success Metrics:**
- Player understands what each setting does
- Settings map to clear gameplay differences
- Default "Stapledon's Voyage" preset feels right for core experience
- Hardcore players can tune for realism or challenge

## Solution Design

### Overview

Three-tier settings approach:

1. **Presets** - Named configurations for quick start
2. **Simple Sliders** - Player-friendly abstractions
3. **Advanced** - Raw Drake-ish parameters for nerds

### Presets

| Preset | Description | Anthropic L | Civs in 1000 ly |
|--------|-------------|-------------|-----------------|
| **Lonely Universe** | Rare Earth hypothesis. You may be alone. | 0.1 | 0-2 |
| **Scattered Few** | Drake pessimistic. Civs exist but far apart. | 0.3 | 2-5 |
| **Stapledon's Voyage** (default) | Rich local cluster. Core experience. | 0.7 | 8-15 |
| **Teeming Galaxy** | Optimistic Drake. Civs everywhere. | 1.0 | 20-40 |

### Simple Sliders

**Civilization Density**
```
Lonely ─────────●───────── Crowded
        "How many other minds exist?"
```
Maps to: Anthropic Luck factor L (0.1 to 1.0)

**Life Abundance**
```
Rare ─────────●───────── Common
     "How often does life emerge?"
```
Maps to: f_life parameter (fraction of HZ planets with life)

**Technology Emergence**
```
Difficult ─────────●───────── Easy
          "How often does life become technological?"
```
Maps to: f_tech parameter (fraction of life that develops tech)

### Advanced Parameters (collapsible)

For players who want precise control:

| Parameter | Range | Default | Description |
|-----------|-------|---------|-------------|
| `anthropic_luck` | 0.0-1.0 | 0.7 | Observer selection bias strength |
| `f_life` | 0.0-1.0 | 0.3 | Fraction of HZ planets with any life |
| `f_complex` | 0.0-1.0 | 0.1 | Fraction of life that becomes complex |
| `f_tech` | 0.0-1.0 | 0.05 | Fraction of complex life that develops tech |
| `civ_lifetime_mean` | 100-100000 | 10000 | Mean civ technological lifetime (years) |
| `gamma_max` | 5-100 | 20 | Player ship max Lorentz factor |
| `local_cluster_radius` | 200-2000 | 800 | Dense region radius (ly) |
| `local_cluster_boost` | 1.0-10.0 | 5.0 | Density multiplier in local cluster |

### AILANG Types

```ailang
type WorldGenParams = {
    seed: int,
    anthropicLuck: float,
    fLife: float,
    fComplex: float,
    fTech: float,
    civLifetimeMean: int,
    gammaMax: float,
    localClusterRadius: float,
    localClusterBoost: float
}

type Preset = Lonely | ScatteredFew | Stapledon | Teeming | Custom(WorldGenParams)

pure func paramsFromPreset(p: Preset) -> WorldGenParams
pure func generateGalaxy(params: WorldGenParams) -> Galaxy
```

### UI Layout

```
╔══════════════════════════════════════════════════════════╗
║                    NEW VOYAGE                            ║
╠══════════════════════════════════════════════════════════╣
║                                                          ║
║  Choose Your Universe                                    ║
║  ─────────────────────                                   ║
║                                                          ║
║  [○] Lonely Universe     - Perhaps we are alone          ║
║  [○] Scattered Few       - They exist, but far apart     ║
║  [●] Stapledon's Voyage  - A rich local cluster awaits   ║
║  [○] Teeming Galaxy      - Minds everywhere you look     ║
║  [○] Custom              - Configure parameters          ║
║                                                          ║
║  ─────────────────────────────────────────────────────── ║
║                                                          ║
║  What This Means:                                        ║
║  • ~12 civilizations within 1000 light-years             ║
║  • Many biospheres to discover nearby                    ║
║  • Network effects visible within your lifetime          ║
║  • Round trips possible to nearby civs (2-3 visits)      ║
║                                                          ║
║  ─────────────────────────────────────────────────────── ║
║                                                          ║
║  Galaxy Seed: [_42_______] (or leave blank for random)   ║
║                                                          ║
║           [ ▼ Advanced Parameters ]                      ║
║                                                          ║
║              [ BEGIN VOYAGE ]                            ║
║                                                          ║
╚══════════════════════════════════════════════════════════╝
```

### Lore Integration

Each preset includes a brief lore justification:

**Lonely Universe:**
> "Theoretical estimates suggest technological civilizations may be vanishingly rare. You explore knowing you might never find another mind."

**Stapledon's Voyage (default):**
> "You are in the lucky branch. Of all the possible universes, you exist in one where the conditions for contact align. Perhaps this is anthropic selection. Perhaps it is destiny."

**Teeming Galaxy:**
> "The galaxy hums with activity. Signals cross the void constantly. The question is not whether you will find them, but what you will do when you do."

### Implementation Plan

**Phase 1: Data Model** (~1 day)
- [ ] Define WorldGenParams type in AILANG
- [ ] Implement preset → params conversion
- [ ] Wire params to galaxy generation

**Phase 2: UI Implementation** (~1.5 days)
- [ ] Create world gen screen in engine
- [ ] Implement preset selection radio buttons
- [ ] Add "What This Means" preview text
- [ ] Add seed input field

**Phase 3: Advanced Panel** (~0.5 days)
- [ ] Collapsible advanced parameters
- [ ] Slider controls for each parameter
- [ ] Real-time preview update

### Files to Modify/Create

**New files:**
- `sim/world_gen.ail` - WorldGenParams, presets, generation (~150 LOC)
- `engine/ui/worldgen_screen.go` - UI implementation (~400 LOC)

**Modified files:**
- `cmd/game/main.go` - Add world gen flow before game start

## Examples

### Example 1: Lonely Universe Generation

```ailang
let params = paramsFromPreset(Lonely)
-- params.anthropicLuck = 0.1
-- params.fLife = 0.05
-- params.fTech = 0.01

let galaxy = generateGalaxy(params)
-- galaxy.civs = [] or [1 civ 3000 ly away]
-- galaxy.biospheres = [3 within 500 ly]
```

### Example 2: Custom Hardcore Settings

```ailang
let hardcore = Custom({
    seed: 12345,
    anthropicLuck: 0.05,
    fLife: 0.01,
    fComplex: 0.001,
    fTech: 0.0001,
    civLifetimeMean: 1000,
    gammaMax: 10,
    localClusterRadius: 500,
    localClusterBoost: 1.0
})
-- Nearly impossible to find anyone
-- If you do, they're likely dead by the time you arrive
```

## Success Criteria

- [ ] All 4 presets generate distinct galaxy feels
- [ ] Preview text accurately describes expected experience
- [ ] Same seed + params produces identical galaxy
- [ ] Advanced parameters all function correctly
- [ ] UI is navigable without documentation

## Testing Strategy

**Unit tests:**
- Preset conversion produces expected params
- Galaxy gen with same seed is deterministic
- Civ count falls within expected range for each preset

**Integration tests:**
- Full world gen → game start flow works
- Parameters correctly affect civ placement

**Playtesting:**
- Each preset feels different in first 30 min of play
- Lonely Universe is actually lonely
- Stapledon's Voyage has ~10 reachable civs

## Non-Goals

**Not in this feature:**
- In-game parameter changes (locked at world gen)
- Difficulty rating/scoring differences per preset
- Tutorial integration
- Galaxy preview visualization

## Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|-----------|
| Presets feel too similar | High | Tune parameters with wide spread; playtest each |
| Advanced UI too complex | Med | Default to collapsed; good tooltips |
| "Lonely" preset is boring | Med | Ensure biospheres still common; lone civ is special |

## References

- [startmaps.md](startmaps.md) - Anthropic Luck discussion
- [design-decisions.md](../../docs/vision/design-decisions.md) - "Anthropic Luck: World-Gen Only"
- Drake Equation parameters and real astrobiology estimates

## Future Work

- Galaxy preview visualization before committing
- Import/export custom presets
- Community preset sharing
- Per-preset achievements/challenges

---

**Document created**: 2025-12-02
**Last updated**: 2025-12-02
