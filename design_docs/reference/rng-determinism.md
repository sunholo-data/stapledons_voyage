# RNG & Determinism System

**Status**: Planned
**Target**: v0.5.1 (tracks AILANG RNG effect)
**Priority**: P1 - High
**Estimated**: 3 days
**Dependencies**: AILANG v0.5.1 with RNG effect

## Game Vision Alignment

| Pillar | Relevance | Score | Notes |
|--------|-----------|-------|-------|
| Time Dilation Consequence | + | +1 | Same seed = reproducible universe history |
| Civilization Simulation | + | +1 | Procedural civ generation, state transitions |
| Philosophical Depth | 0 | 0 | Infrastructure |
| Ship & Crew Life | + | +1 | Crew events, random encounters |
| Legacy Impact | + | +1 | Replay with same seed to see different choices |
| Hard Sci-Fi Authenticity | + | +1 | Deterministic simulation = scientifically reproducible |
| **Net Score** | | **+5** | **Decision: Move forward** |

**Feature type:** Infrastructure + Gameplay (enables replay, debugging, sharing seeds)

## Problem Statement

The game needs random number generation that is:
1. **Deterministic** - Same seed produces same universe every time
2. **Reproducible** - Players can share seeds, compare runs
3. **Debuggable** - Replay bugs with exact same RNG sequence
4. **Streamable** - Multiple independent RNG streams for different systems

**Current State:**
- Mock uses Go's `math/rand` with global state
- Not properly seeded or isolated

**Impact:**
- Foundation for all procedural generation
- Required for eval system (reproducible benchmarks)
- Enables "seed sharing" social feature

## Goals

**Primary Goal:** Implement deterministic RNG using AILANG's RNG effect that produces identical results across runs, platforms, and versions.

**Success Metrics:**
- Same seed + same inputs = byte-identical game state
- 1000 runs with same seed produce same galaxy
- RNG sequences verified against reference implementation
- Replay system can reproduce any game session

## Solution Design

### Overview

AILANG v0.5.1 provides the RNG effect:

```ailang
effect RNG {
    rand_float() -> float    -- Returns [0.0, 1.0)
    rand_int(max: int) -> int -- Returns [0, max)
}
```

We use this for all randomness in the game.

### RNG Streams Architecture

Different game systems get independent RNG streams to prevent coupling:

```
Galaxy Seed (from world gen)
         │
         ├──► Star Generation Stream
         │         └─► Star positions, types, planets
         │
         ├──► Civilization Stream
         │         └─► Civ placement, philosophies, initial states
         │
         ├──► Event Stream
         │         └─► Random events during gameplay
         │
         ├──► NPC Stream
         │         └─► NPC decisions, dialogue variations
         │
         └──► Crew Stream
                   └─► Crew events, relationships, births/deaths
```

### Stream Isolation

Each stream is derived from the master seed:

```ailang
type RNGStreams = {
    starSeed: int,
    civSeed: int,
    eventSeed: int,
    npcSeed: int,
    crewSeed: int
}

pure func deriveStreams(masterSeed: int) -> RNGStreams {
    -- Use hash mixing to derive independent seeds
    {
        starSeed: mixSeed(masterSeed, 0x5TAR),
        civSeed: mixSeed(masterSeed, 0xC1V5),
        eventSeed: mixSeed(masterSeed, 0xEVNT),
        npcSeed: mixSeed(masterSeed, 0x0NPC),
        crewSeed: mixSeed(masterSeed, 0xCREW)
    }
}

pure func mixSeed(seed: int, salt: int) -> int {
    -- FNV-1a style mixing
    let h = seed * 16777619
    h ^ salt
}
```

### Usage Patterns

**World Generation:**
```ailang
func generateGalaxy(params: WorldGenParams) -> Galaxy ! {RNG} {
    let streams = deriveStreams(params.seed)

    -- Generate stars using star stream
    let stars = generateStars(streams.starSeed, params)

    -- Place civilizations using civ stream
    let civs = placeCivilizations(streams.civSeed, stars, params)

    { stars: stars, civs: civs, ... }
}

func generateStars(seed: int, params: WorldGenParams) -> [Star] ! {RNG} {
    -- RNG.rand_float() uses the current stream
    let count = RNG.rand_int(params.maxStars - params.minStars) + params.minStars
    generateStarList(count, [])
}
```

**Per-Tick Events:**
```ailang
func processRandomEvents(world: World) -> World ! {RNG} {
    -- Roll for random events this tick
    let roll = RNG.rand_float()
    if roll < world.eventProbability then
        let eventType = RNG.rand_int(length(possibleEvents))
        applyEvent(world, possibleEvents !! eventType)
    else
        world
}
```

### Determinism Guarantees

**AILANG Contract (v0.5.1+):**
```bash
# Setting seed via environment
AILANG_SEED=42 ./game

# All RNG.rand_*() calls will use this seed
# Identical on all platforms (x64, ARM, etc.)
```

**Our Verification:**
```ailang
-- Test that verifies determinism
tests [
    (deriveStreams(42), { starSeed: 12345, civSeed: 67890, ... }),
    (deriveStreams(42), { starSeed: 12345, civSeed: 67890, ... })  -- Same!
]
```

### Replay System

**Recording:**
```go
type GameRecording struct {
    Seed      int64
    Inputs    []FrameInput  // All player inputs
    Version   string        // Game version for compatibility
    Timestamp time.Time
}

func RecordGame(seed int64) *GameRecording {
    return &GameRecording{
        Seed:      seed,
        Inputs:    make([]FrameInput, 0),
        Version:   version.String(),
        Timestamp: time.Now(),
    }
}
```

**Playback:**
```go
func ReplayGame(recording *GameRecording) error {
    // Set seed
    os.Setenv("AILANG_SEED", strconv.FormatInt(recording.Seed, 10))

    // Initialize world
    world := sim_gen.InitWorld(recording.Seed)

    // Replay all inputs
    for _, input := range recording.Inputs {
        newWorld, _, err := sim_gen.Step(world, input)
        if err != nil {
            return err
        }
        world = newWorld
    }
    return nil
}
```

### Seed Sharing

Players can share seeds for interesting galaxies:

```
STAPLEDON-42-DENSE-CLUSTER
         │    │       │
         │    │       └─ Description tag
         │    └─ Anthropic luck setting
         └─ Master seed
```

**Decode:**
```ailang
func decodeSeedString(s: string) -> WorldGenParams {
    let parts = split(s, "-")
    {
        seed: parseInt(parts !! 1),
        anthropicLuck: decodeLuck(parts !! 2),
        ...
    }
}
```

### Implementation Plan

**Phase 1: Stream Architecture** (~1 day)
- [ ] Define RNGStreams type
- [ ] Implement seed mixing
- [ ] Wire streams to systems

**Phase 2: Migration** (~1 day)
- [ ] Replace all Go `math/rand` with AILANG RNG
- [ ] Add stream parameter to all random functions
- [ ] Remove global RNG state

**Phase 3: Verification** (~1 day)
- [ ] Add determinism tests
- [ ] Verify cross-platform identical output
- [ ] Implement replay system

### Files to Modify/Create

**AILANG source:**
- `sim/rng.ail` - RNG streams, seed mixing (~100 LOC)

**Go source:**
- `engine/replay/recording.go` - Replay system (~200 LOC)
- `cmd/game/main.go` - Seed handling from env/args

**Tests:**
- `sim/rng_test.ail` - Determinism tests
- `tests/determinism_test.go` - Cross-run verification

## Examples

### Example 1: Identical Galaxies

```bash
# Run 1
AILANG_SEED=42 ./game --headless --ticks=1 --dump-state > state1.json

# Run 2 (same seed)
AILANG_SEED=42 ./game --headless --ticks=1 --dump-state > state2.json

# Verify identical
diff state1.json state2.json  # No output = identical
```

### Example 2: Sharing a Good Seed

```
Player A finds an interesting galaxy:
"Check out seed STAPLEDON-12345-DENSE - there's a cluster of
5 civs within 300 ly, and one has Gift Economy philosophy!"

Player B can recreate exactly the same starting conditions.
```

## Success Criteria

- [ ] Same seed produces identical star positions
- [ ] Same seed produces identical civ placements
- [ ] Replay of 1000 ticks matches original
- [ ] Cross-platform determinism verified (Linux/Mac/Windows)
- [ ] Seed sharing works between players

## Testing Strategy

**Determinism tests:**
- Generate galaxy with seed N, hash state
- Generate again with seed N, compare hash
- Repeat 100 times with different seeds

**Replay tests:**
- Record 100-tick session
- Replay, verify identical final state
- Verify on different machines

**Distribution tests:**
- Verify rand_float() is uniform [0, 1)
- Verify rand_int(N) is uniform [0, N)
- Statistical tests (chi-squared)

## Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|-----------|
| Platform float differences | High | Use integer RNG internally; float conversion spec'd |
| AILANG RNG implementation changes | High | Pin to AILANG version; verify against reference |
| Stream isolation broken | Med | Test streams independently; hash verification |
| Recording files corrupt | Low | Checksum in recording format |

## References

- [consumer-contract-v0.5.md](../../ailang_resources/consumer-contract-v0.5.md) - RNG effect spec
- [world-gen-settings.md](world-gen-settings.md) - Seed usage in world gen
- PCG/xorshift literature for RNG quality

## Future Work

- Visual seed browser (preview galaxies before committing)
- Seed leaderboards (famous seeds for different playstyles)
- Partial replay (start from checkpoint)
- Seed mutation (small changes to explore variations)

---

**Document created**: 2025-12-02
**Last updated**: 2025-12-02
