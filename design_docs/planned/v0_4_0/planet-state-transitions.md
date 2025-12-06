# Planet State Transitions

**Status**: Planned
**Target**: v0.4.0
**Priority**: P1 - High
**Estimated**: 1 week
**Dependencies**: Starmap Data Model, World Gen Settings

## Game Vision Alignment

| Pillar | Relevance | Score | Notes |
|--------|-----------|-------|-------|
| Time Dilation Consequence | + | +1 | Core mechanic: what you saw vs what you find |
| Civilization Simulation | + | +1 | How civs evolve over millennia |
| Philosophical Depth | + | +1 | Rise and fall of civilizations prompts reflection |
| Ship & Crew Life | + | +1 | Crew reacts to finding ruins where signals were |
| Legacy Impact | + | +1 | Your choices affect transition probabilities |
| Hard Sci-Fi Authenticity | + | +1 | Timescales grounded in real evolutionary/civilizational estimates |
| **Net Score** | | **+6** | **Decision: Move forward** |

**Feature type:** Gameplay (core to the "epistemic gap" experience)

## Problem Statement

When you observe a planet from afar, you see old light. When you arrive, centuries or millennia have passed. The planet's state may have changed dramatically:

- Biosignature → Civilization (life evolved intelligence)
- Civilization → Ruins (they collapsed)
- Ruins → New civilization (someone else arose)
- Biosignature → Ecocide (runaway climate)
- Sterile → Biosignature (abiogenesis)

This is the emotional core of the "epistemic gap" - you commit to a journey based on uncertain, outdated information.

**Current State:**
- No planet evolution model exists
- Need to define states and transition probabilities

**Impact:**
- Every arrival should feel like opening a surprise
- Player learns to read probabilities, not certainties
- Deep time becomes tangible through state changes

## Goals

**Primary Goal:** Define planet states and transitions that create meaningful surprises on arrival while remaining scientifically plausible.

**Success Metrics:**
- Arrival state differs from prediction 30-60% of the time
- Differences are always explainable (not random)
- Timescales feel right (civs don't appear overnight)
- Player choices can influence transitions

## Solution Design

### Overview

Planets exist in discrete states. States transition over time based on:
1. **Base rates** - How likely is this transition per unit time?
2. **Context** - What's happening nearby? (your influence, other civs)
3. **Random events** - Low-probability catastrophes/breakthroughs

### Planet States

```ailang
type PlanetState =
    | Sterile                           -- No life
    | Prebiotic                         -- Chemistry trending toward life
    | MicrobialLife                     -- Simple life, detectable biosignatures
    | ComplexLife                       -- Multicellular, no intelligence
    | PreTechCiv                        -- Intelligent but pre-industrial
    | TechCiv(CivState)                 -- Technological civilization
    | PostCollapse(CollapseType)        -- Ruins, remnants
    | Ecocide                           -- Runaway climate, dead world
    | Transcended                       -- Post-biological, unclear status

type CivState = {
    techLevel: int,          -- 1-10 scale
    philosophy: Philosophy,
    stability: float,        -- 0-1, affects collapse probability
    expansionDrive: float,   -- 0-1, affects spread to other worlds
    hasHiggsTech: bool       -- Can they travel relativistically?
}

type CollapseType =
    | War
    | Climate
    | Pandemic
    | AIMisalignment
    | ResourceDepletion
    | Unknown

type Philosophy =
    | Expansionist
    | Isolationist
    | GiftEconomy
    | Totalitarian
    | DeathCelebrant
    | ConsensusOnly
    | SacredMortality
    -- ... extensible
```

### State Machine

```
                    ┌─────────────┐
                    │   Sterile   │
                    └──────┬──────┘
                           │ (rare: abiogenesis)
                           ▼
                    ┌─────────────┐
                    │  Prebiotic  │
                    └──────┬──────┘
                           │
                           ▼
                    ┌─────────────┐
              ┌─────│  Microbial  │─────┐
              │     └──────┬──────┘     │
              │            │            │
              ▼            ▼            ▼
        ┌─────────┐  ┌───────────┐  ┌─────────┐
        │ Ecocide │  │  Complex  │  │ Sterile │
        └─────────┘  └─────┬─────┘  └─────────┘
                           │        (extinction)
                           ▼
                    ┌─────────────┐
              ┌─────│  PreTech    │─────┐
              │     └──────┬──────┘     │
              │            │            │
              ▼            ▼            ▼
        ┌─────────┐  ┌───────────┐  ┌─────────┐
        │ Ecocide │  │  TechCiv  │──│  Ruins  │
        └─────────┘  └─────┬─────┘  └─────────┘
                           │              │
                           │              │ (new civ rises)
                           ▼              │
                    ┌─────────────┐       │
                    │ Transcended │◄──────┘
                    └─────────────┘
```

### Transition Timescales

Based on Earth history and Drake-like estimates:

| Transition | Typical Timescale | Notes |
|------------|-------------------|-------|
| Sterile → Prebiotic | 100M - 1B years | Abiogenesis is slow |
| Prebiotic → Microbial | 10M - 500M years | Once chemistry aligns |
| Microbial → Complex | 500M - 2B years | Eukaryotes took a while |
| Complex → PreTech | 100M - 500M years | Intelligence emergence |
| PreTech → TechCiv | 10k - 100k years | Industrial revolution |
| TechCiv → Ruins | 100 - 10k years | Civilizations are fragile |
| TechCiv → Transcended | 1k - 100k years | Post-biological transition |
| Ruins → New PreTech | 10M - 100M years | New intelligence from survivors/evolution |

### Transition Probabilities

Per-year transition rates (example values, tunable):

```ailang
type TransitionRates = {
    -- Per-year probability of transition
    sterileToPrebiotic: float,      -- ~1e-9
    prebioticToMicrobial: float,    -- ~1e-8
    microbialToComplex: float,      -- ~1e-9
    complexToPreTech: float,        -- ~1e-8
    preTechToTech: float,           -- ~1e-5

    -- Collapse rates (per year when TechCiv)
    techToRuinsBase: float,         -- ~1e-4 (1 in 10,000 years)
    techToTranscended: float,       -- ~1e-5

    -- Recovery
    ruinsToPreTech: float,          -- ~1e-8

    -- Catastrophes
    anyToEcocide: float,            -- ~1e-7 (runaway climate)
    anyToSterile: float             -- ~1e-10 (gamma ray burst, etc)
}
```

### Modifiers

Transition rates are modified by:

**1. Player Influence**
```ailang
-- Sharing Higgs tech increases expansion but may destabilize
pure func modifyRatesAfterUplift(rates: TransitionRates, civState: CivState)
    -> TransitionRates {
    let stabilityMod = if civState.stability < 0.3
        then 2.0  -- Unstable civs more likely to collapse
        else 1.0
    { rates | techToRuinsBase: rates.techToRuinsBase * stabilityMod }
}
```

**2. Philosophy**
```ailang
-- Death Celebrants are more stable but don't expand
-- Expansionists spread but risk overextension
pure func philosophyModifier(phil: Philosophy) -> RateModifiers
```

**3. Tech Level**
```ailang
-- Higher tech = more ways to destroy yourself, but also more resilience
pure func techLevelModifier(level: int) -> RateModifiers
```

**4. Network Effects**
```ailang
-- Contact with other civs changes trajectories
-- War, trade, cultural exchange all affect stability/expansion
pure func networkModifier(neighbors: [CivState]) -> RateModifiers
```

### Detection vs Reality

The epistemic gap mechanics:

```ailang
type PlanetBelief = {
    state: PlanetState,           -- What we THINK it is
    confidence: float,            -- 0-1
    lastLightYear: int,           -- When the photons left
    surveyedYear: int,            -- When we looked
    arrivalYear: int              -- When we'll get there (predicted)
}

pure func predictStateAtArrival(
    belief: PlanetBelief,
    rates: TransitionRates,
    yearsElapsed: int
) -> [(PlanetState, float)]  -- State distribution at arrival
```

**Example Prediction:**

> "500 years ago, this world showed technosignatures at early industrial level."
> "In the 2000 years since + until your arrival:"
> - 45% still TechCiv (advanced)
> - 30% Ruins (collapsed)
> - 15% Transcended
> - 10% Other (ecocide, war, etc)

### Arrival Resolution

When you arrive, actual state is determined:

```ailang
pure func resolveArrival(
    belief: PlanetBelief,
    actualState: PlanetState,
    seed: int
) -> ArrivalEvent {
    match (belief.state, actualState) {
        (TechCiv(_), Ruins(collapseType)) =>
            ArrivalEvent.FoundRuins(collapseType, belief.state),
        (MicrobialLife, TechCiv(civ)) =>
            ArrivalEvent.LifeEvolvedToTech(civ),
        (TechCiv(old), TechCiv(new)) =>
            ArrivalEvent.CivChanged(old, new),
        -- ... etc
    }
}
```

### Implementation Plan

**Phase 1: Core State Machine** (~2 days)
- [ ] Define PlanetState ADT in AILANG
- [ ] Implement transition rate calculations
- [ ] Basic state evolution over time

**Phase 2: Modifiers** (~2 days)
- [ ] Philosophy modifiers
- [ ] Tech level modifiers
- [ ] Player influence effects
- [ ] Network effects

**Phase 3: Prediction & Resolution** (~2 days)
- [ ] Belief state tracking
- [ ] Probability distribution calculation
- [ ] Arrival resolution with narrative generation

**Phase 4: Integration** (~1 day)
- [ ] Connect to starmap
- [ ] Wire to UI (journey planning shows predictions)
- [ ] Connect to crew reactions

### Files to Modify/Create

**New files:**
- `sim/planet_state.ail` - State types and transitions (~300 LOC)
- `sim/civ_evolution.ail` - CivState evolution logic (~200 LOC)
- `sim/prediction.ail` - Belief states and predictions (~150 LOC)

**Modified files:**
- `sim/world.ail` - Add PlanetState to Planet type
- `sim/step.ail` - Evolve planet states each tick

## Examples

### Example 1: Biosignature → Civilization

**Observation (year 0):**
```
Planet Kepler-442b
Distance: 1200 ly
Last light: -1200 years
Survey result: Strong O₂/CH₄ disequilibrium
Belief: MicrobialLife (85% confidence)
```

**Journey (γ=20):**
```
Travel time: 60 years subjective
Arrival: year 1200 external
Total elapsed since observation: 2400 years
```

**Arrival (year 1200):**
```
Actual state: TechCiv(level=4, philosophy=GiftEconomy)
"In the 2400 years since the light you observed left this world,
life evolved intelligence, developed technology, and built a
civilization based on reciprocal exchange. They have radio
but no interstellar capability. They are eager to meet you."
```

### Example 2: Technosignature → Ruins

**Observation (year 0):**
```
Planet HD 40307g
Distance: 800 ly
Last light: -800 years
Survey result: Narrowband radio, industrial waste heat
Belief: TechCiv(level=6) (90% confidence)
```

**Journey (γ=20):**
```
Travel time: 40 years subjective
Arrival: year 800 external
Total elapsed since observation: 1600 years
```

**Arrival (year 800):**
```
Actual state: PostCollapse(AIMisalignment)
"The radio signals you detected fell silent 300 years after
they were transmitted. You find a world of silent cities,
maintained by autonomous systems that no longer serve anyone.
Archaeological analysis suggests a rapid transition to machine
intelligence followed by... something. The machines are still
here. They do not seem hostile. They seem... patient."
```

## Success Criteria

- [ ] State transitions follow defined rates
- [ ] Predictions match actual outcomes within expected variance
- [ ] Player influence visibly affects civ trajectories
- [ ] Arrival surprises feel earned, not random
- [ ] Timescales are scientifically defensible

## Testing Strategy

**Unit tests:**
- Transition rate calculations correct
- Modifier stacking works properly
- Probability distributions sum to 1.0

**Integration tests:**
- State evolution over 10k years produces expected distribution
- Beliefs diverge from reality appropriately
- Player uplift affects subsequent states

**Narrative tests:**
- Sample 100 arrival events
- Verify each has coherent narrative explanation
- Check for unexpected/illogical combinations

## Non-Goals

**Not in this feature:**
- Detailed civilization internal simulation (separate system)
- Planet surface/geography (separate system)
- Individual NPC states within civilizations
- Detailed archaeology mechanics

## Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|-----------|
| Transitions feel random | High | Always provide narrative explanation; show probability in UI |
| Timescales too fast/slow | Med | Playtest and tune; expose as advanced param |
| Too many surprise ruins | Med | Tune base collapse rate; player can affect stability |
| Transcendence too common | Med | Make it rare and mysterious; don't explain too much |

## References

- [startmaps.md](startmaps.md) - Epistemic gap discussion
- [resources.md](resources.md) - Alien biosphere science
- Drake equation literature for timescale estimates
- Great Filter hypothesis for collapse modeling

## Future Work

- Detailed archaeology system for ruins
- Transcendence interaction mechanics
- Multi-planet civilizations
- Civ memory of player (reputation across visits)
- Civ-to-civ contact modeling when player introduces them

---

**Document created**: 2025-12-02
**Last updated**: 2025-12-02
