# Performance Externs (Go-Implemented Kernels)

**Status**: Planned
**Target**: v0.5.2 (tracks AILANG extern support)
**Priority**: P2 - Medium
**Estimated**: 1 week
**Dependencies**: AILANG v0.5.2 with extern functions, AILANG Go codegen

## Game Vision Alignment

| Pillar | Relevance | Score | Notes |
|--------|-----------|-------|-------|
| Time Dilation Consequence | N/A | 0 | Infrastructure - enables fast simulation |
| Civilization Simulation | + | +1 | Needed for galaxy-scale pathfinding & influence |
| Philosophical Depth | N/A | 0 | Infrastructure |
| Ship & Crew Life | N/A | 0 | Infrastructure |
| Legacy Impact | + | +1 | Fast Year 1,000,000 calculation |
| Hard Sci-Fi Authenticity | N/A | 0 | Infrastructure |
| **Net Score** | | **+2** | **Decision: Move forward** |

**Feature type:** Infrastructure (enables simulation performance at scale)

## Problem Statement

Some game computations are performance-critical and cannot be efficiently expressed in pure functional AILANG:

1. **Pathfinding (A*)** - Galaxy navigation with 100k+ stars
2. **Influence maps** - Civ territory calculation across the galaxy
3. **Spatial queries** - "What stars are within X light-years?"
4. **Mass simulation** - Fast-forwarding 1M years for endgame

AILANG's pure functional style is excellent for game logic correctness but O(n) list access and no mutation make these algorithms prohibitively slow.

**Current State:**
- No extern mechanism exists
- Mock sim_gen implements these in Go directly
- When AILANG ships, we need a way to call back into Go

**Impact:**
- Without externs, game will be too slow for galaxy-scale simulation
- 100k stars × 1M years = unacceptable without native performance

## Goals

**Primary Goal:** Enable AILANG to call Go-implemented performance kernels while maintaining determinism and type safety.

**Success Metrics:**
- A* pathfinding handles 100k nodes in <100ms
- Influence map calculation for 1000 civs in <500ms
- Spatial queries return in O(log n) not O(n)
- All externs are deterministic (same input → same output)

## Solution Design

### Overview

AILANG declares `extern` functions with type signatures. The generated Go code calls registered Go implementations via a handler registry.

```ailang
-- Declared in AILANG, implemented in Go
extern pathfind(from: Star, to: Star, galaxy: Galaxy) -> [Star]
extern starsWithinRadius(center: Vec3, radius: float, galaxy: Galaxy) -> [Star]
extern calculateInfluence(civs: [Civ], galaxy: Galaxy) -> InfluenceMap
```

### Architecture

```
┌─────────────────────────────────────────────────────────────┐
│  AILANG Source (sim/*.ail)                                  │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ extern pathfind(from: Star, to: Star, galaxy: Galaxy)│   │
│  │     -> [Star]                                        │   │
│  │                                                      │   │
│  │ func planJourney(start: Star, end: Star, w: World)   │   │
│  │     -> JourneyPlan {                                 │   │
│  │     let route = pathfind(start, end, w.galaxy)       │   │
│  │     -- use route...                                  │   │
│  │ }                                                    │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼ ailang compile --emit-go
┌─────────────────────────────────────────────────────────────┐
│  Generated Go (sim_gen/*.go)                                │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ func Pathfind(from Star, to Star, galaxy Galaxy)     │   │
│  │     []Star {                                         │   │
│  │     return externRegistry.Pathfind(from, to, galaxy) │   │
│  │ }                                                    │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼ Calls registered handler
┌─────────────────────────────────────────────────────────────┐
│  Go Implementation (engine/extern/*.go)                     │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ func init() {                                        │   │
│  │     sim_gen.RegisterPathfind(astarPathfind)          │   │
│  │ }                                                    │   │
│  │                                                      │   │
│  │ func astarPathfind(from, to Star, g Galaxy) []Star { │   │
│  │     // A* implementation using priority queue        │   │
│  │     // Returns deterministic result                  │   │
│  │ }                                                    │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

### Extern Categories

**1. Pathfinding**

```ailang
-- A* pathfinding through star graph
extern pathfind(from: Star, to: Star, galaxy: Galaxy) -> [Star]

-- Multi-target: find path to nearest of several destinations
extern pathfindNearest(from: Star, targets: [Star], galaxy: Galaxy)
    -> (Star, [Star])  -- (chosen target, path to it)
```

Go implementation uses:
- Priority queue (heap)
- Precomputed neighbor graph (stars within jump range)
- Heuristic: Euclidean distance

**2. Spatial Queries**

```ailang
-- Stars within radius (light-years)
extern starsWithinRadius(center: Vec3, radius: float, galaxy: Galaxy) -> [Star]

-- K nearest stars
extern nearestStars(center: Vec3, k: int, galaxy: Galaxy) -> [Star]

-- Stars in bounding box
extern starsInBox(min: Vec3, max: Vec3, galaxy: Galaxy) -> [Star]
```

Go implementation uses:
- Octree spatial index (built once at galaxy gen)
- O(log n) queries instead of O(n) list scan

**3. Influence Maps**

```ailang
type InfluenceMap = {
    cells: [[InfluenceCell]],  -- 2D grid (projected from 3D)
    cellSize: float,
    origin: Vec3
}

type InfluenceCell = {
    dominantCiv: int,        -- Civ ID or -1
    influence: float,        -- 0-1 strength
    contested: bool          -- Multiple civs claiming
}

extern calculateInfluence(
    civs: [Civ],
    galaxy: Galaxy,
    resolution: int
) -> InfluenceMap

extern influenceAt(pos: Vec3, map: InfluenceMap) -> InfluenceCell
```

Go implementation:
- Grid-based influence propagation
- Distance falloff from civ homeworlds
- O(civs × cells) but parallelizable

**4. Fast-Forward Simulation**

```ailang
-- Advance simulation N years without full step-by-step
extern fastForward(world: World, years: int, seed: int) -> World

-- Used for Year 1,000,000 endgame calculation
extern simulateToEndgame(world: World, targetYear: int) -> World
```

Go implementation:
- Batch state transitions
- Statistical sampling instead of per-tick simulation
- Deterministic given seed

### Type Compatibility

Extern functions must use types that cross the AILANG/Go boundary cleanly:

| AILANG Type | Go Type | Notes |
|-------------|---------|-------|
| `int` | `int64` | Always 64-bit |
| `float` | `float64` | Always 64-bit |
| `bool` | `bool` | Direct |
| `string` | `string` | UTF-8 |
| `[T]` | `[]T` | Slices |
| `(T, U)` | `struct{A T; B U}` | Named tuple struct |
| ADT | Discriminator struct | See integration doc |
| Record | Struct | Field names match |

### Determinism Requirements

All extern functions MUST be deterministic:

```go
// GOOD: Deterministic
func astarPathfind(from, to Star, g Galaxy) []Star {
    // Same from, to, g always produces same result
    // Tie-breaking uses stable sort by star ID
}

// BAD: Non-deterministic
func astarPathfind(from, to Star, g Galaxy) []Star {
    // Uses map iteration order (random in Go)
    // Different runs produce different paths
}
```

For externs that need randomness (e.g., `fastForward`), pass seed explicitly:

```ailang
extern fastForward(world: World, years: int, seed: int) -> World
```

### Registration Pattern

```go
// engine/extern/registry.go
package extern

import "stapledons_voyage/sim_gen"

func init() {
    // Register all extern implementations
    sim_gen.RegisterPathfind(pathfindImpl)
    sim_gen.RegisterStarsWithinRadius(spatialQueryImpl)
    sim_gen.RegisterCalculateInfluence(influenceImpl)
    sim_gen.RegisterFastForward(fastForwardImpl)
}

// Called from main to ensure init runs
func Initialize() {
    // No-op, but importing this package triggers init()
}
```

```go
// cmd/game/main.go
package main

import (
    _ "stapledons_voyage/engine/extern" // Register externs
)

func main() {
    // Externs are now registered, AILANG can call them
}
```

### Implementation Plan

**Phase 1: Spatial Index** (~2 days)
- [ ] Implement octree data structure
- [ ] Build index at galaxy generation
- [ ] Implement `starsWithinRadius` query
- [ ] Implement `nearestStars` query
- [ ] Benchmark: 100k stars, <1ms query time

**Phase 2: Pathfinding** (~2 days)
- [ ] Implement A* with priority queue
- [ ] Build neighbor graph (stars within max γ range)
- [ ] Implement `pathfind` single-target
- [ ] Implement `pathfindNearest` multi-target
- [ ] Benchmark: 100k stars, <100ms path

**Phase 3: Influence Maps** (~2 days)
- [ ] Implement grid-based influence propagation
- [ ] Add distance falloff calculation
- [ ] Implement contested territory detection
- [ ] Benchmark: 1000 civs, <500ms calculation

**Phase 4: Fast-Forward** (~1 day)
- [ ] Implement statistical state transitions
- [ ] Add deterministic seeding
- [ ] Integrate with Year 1,000,000 endgame
- [ ] Benchmark: 1M years in <5 seconds

### Files to Modify/Create

**New files:**
- `engine/extern/registry.go` - Extern registration (~50 LOC)
- `engine/extern/spatial.go` - Octree and spatial queries (~400 LOC)
- `engine/extern/pathfind.go` - A* implementation (~300 LOC)
- `engine/extern/influence.go` - Influence map calculation (~250 LOC)
- `engine/extern/fastforward.go` - Fast-forward simulation (~300 LOC)

**Modified files:**
- `sim_gen/extern_stubs.go` - Generated extern call stubs (auto-generated)
- `cmd/game/main.go` - Import extern package

**AILANG files:**
- `sim/extern_decl.ail` - Extern declarations (~50 LOC)

## Examples

### Example 1: Journey Planning with A*

**AILANG usage:**
```ailang
func planJourney(ship: Ship, destination: Star, world: World) -> JourneyPlan {
    let route = pathfind(ship.location, destination, world.galaxy)
    let totalDistance = sumDistances(route)
    let travelTime = calculateTravelTime(totalDistance, ship.gamma)

    JourneyPlan {
        route: route,
        distance: totalDistance,
        subjectiveYears: travelTime.subjective,
        externalYears: travelTime.external
    }
}
```

**Go implementation:**
```go
func pathfindImpl(from, to sim_gen.Star, g sim_gen.Galaxy) []sim_gen.Star {
    // Build priority queue
    pq := &starHeap{}
    heap.Init(pq)

    // A* search
    gScore := make(map[int64]float64)
    cameFrom := make(map[int64]int64)
    gScore[from.ID] = 0

    heap.Push(pq, &starNode{
        star:     from,
        fScore:   heuristic(from, to),
    })

    for pq.Len() > 0 {
        current := heap.Pop(pq).(*starNode)

        if current.star.ID == to.ID {
            return reconstructPath(cameFrom, to)
        }

        for _, neighbor := range getNeighbors(current.star, g) {
            tentative := gScore[current.star.ID] + distance(current.star, neighbor)
            if old, ok := gScore[neighbor.ID]; !ok || tentative < old {
                cameFrom[neighbor.ID] = current.star.ID
                gScore[neighbor.ID] = tentative
                heap.Push(pq, &starNode{
                    star:   neighbor,
                    fScore: tentative + heuristic(neighbor, to),
                })
            }
        }
    }

    return nil // No path found
}
```

### Example 2: Influence Map for Territory Display

**AILANG usage:**
```ailang
func getGalaxyInfluence(world: World) -> InfluenceMap {
    calculateInfluence(world.civilizations, world.galaxy, 100)
}

func isContested(pos: Vec3, world: World) -> bool {
    let cell = influenceAt(pos, world.influenceMap)
    cell.contested
}
```

**Go implementation:**
```go
func influenceImpl(civs []sim_gen.Civ, g sim_gen.Galaxy, resolution int) sim_gen.InfluenceMap {
    // Create grid
    grid := make([][]sim_gen.InfluenceCell, resolution)
    for i := range grid {
        grid[i] = make([]sim_gen.InfluenceCell, resolution)
    }

    // Calculate influence from each civ
    for _, civ := range civs {
        for _, star := range civ.ControlledStars {
            propagateInfluence(grid, star.Pos, civ.ID, civ.Strength)
        }
    }

    // Detect contested regions
    for i := range grid {
        for j := range grid[i] {
            grid[i][j].Contested = isContested(grid[i][j])
        }
    }

    return sim_gen.InfluenceMap{
        Cells:    grid,
        CellSize: g.Size / float64(resolution),
        Origin:   g.Center,
    }
}
```

## Success Criteria

- [ ] All externs are deterministic (verified by running twice with same input)
- [ ] Pathfinding: 100k stars in <100ms
- [ ] Spatial query: <1ms for radius search
- [ ] Influence map: 1000 civs in <500ms
- [ ] Fast-forward: 1M years in <5 seconds
- [ ] Type compatibility verified (AILANG types ↔ Go types)
- [ ] Registration pattern works with generated code

## Testing Strategy

**Unit tests:**
- Octree insertion/query correctness
- A* finds optimal path on known graphs
- Influence propagation matches expected falloff

**Integration tests:**
- AILANG code calls extern, gets correct result
- Round-trip: AILANG → Go extern → AILANG works

**Performance tests:**
- Benchmark suite in `engine/extern/bench_test.go`
- CI tracks performance regression

**Determinism tests:**
- Run each extern 1000x with same input
- Verify identical output every time

## Non-Goals

**Not in this feature:**
- GPU acceleration - Defer to future
- Network distribution - Single-machine only
- Dynamic extern loading - Compile-time registration only

## Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|-----------|
| Type mismatch between AILANG and Go | High | Generated code includes type assertions; fail loudly |
| Non-determinism breaks replay | High | Determinism test suite; no map iteration |
| Extern panic crashes game | Med | Recover in registration wrapper; return error ADT |
| Performance still insufficient | Med | Profile and optimize hot paths; consider caching |

## References

- [consumer-contract-v0.5.md](../../ailang_resources/consumer-contract-v0.5.md) - Extern spec (section 7)
- [ailang-integration.md](ailang-integration.md) - Overall AILANG/Go architecture
- [rng-determinism.md](rng-determinism.md) - Determinism requirements

## Future Work

- GPU-accelerated influence maps
- Parallel A* for multi-destination queries
- Precomputed distance matrices for common routes
- JIT compilation for hot AILANG functions
- Distributed simulation for very large galaxies

---

**Document created**: 2025-12-02
**Last updated**: 2025-12-02
