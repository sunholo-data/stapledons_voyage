# Debug & Evaluation System

**Status**: Planned
**Target**: v0.5.1 (tracks AILANG Debug effect)
**Priority**: P1 - High
**Estimated**: 4 days
**Dependencies**: AILANG v0.5.1 with Debug effect, RNG Determinism

## Game Vision Alignment

| Pillar | Relevance | Score | Notes |
|--------|-----------|-------|-------|
| Time Dilation Consequence | 0 | 0 | Infrastructure |
| Civilization Simulation | + | +1 | Validates sim correctness |
| Philosophical Depth | 0 | 0 | Infrastructure |
| Ship & Crew Life | 0 | 0 | Infrastructure |
| Legacy Impact | + | +1 | Ensures legacy calculations are correct |
| Hard Sci-Fi Authenticity | + | +1 | Verifies physics/math accuracy |
| **Net Score** | | **+3** | **Decision: Move forward (critical for quality)** |

**Feature type:** Infrastructure (enables AI-driven development, testing, AILANG feedback)

## Problem Statement

As the primary integration test for AILANG, we need:
1. **Debug output** - Assertions and logs from AILANG code
2. **Eval harness** - Automated benchmarks for AILANG improvements
3. **Scenario testing** - Predefined test cases for regression detection
4. **Visual verification** - Screenshot comparison for rendering bugs

**Current State:**
- Basic `make eval` produces `out/report.json`
- No structured debug output from simulation
- Manual testing only

**Impact:**
- Foundation for AI-driven development workflow
- Required for AILANG feedback loop
- Enables continuous integration

## Goals

**Primary Goal:** Create a comprehensive debug and evaluation system that enables automated testing, AI-driven development, and AILANG integration validation.

**Success Metrics:**
- Debug assertions caught and reported, not thrown
- Eval benchmark produces reproducible metrics
- Screenshot comparison catches visual regressions
- Report format consumable by AI for AILANG improvements

## Solution Design

### Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                    Debug & Eval Pipeline                        │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  AILANG Code                   Go Harness                       │
│  ───────────                   ──────────                       │
│                                                                 │
│  Debug.assert(...)  ──────►   DebugOutput.Assertions            │
│  Debug.log(...)     ──────►   DebugOutput.Logs                  │
│  Debug.collect()    ──────►   FrameOutput.Debug                 │
│                                     │                           │
│                                     ▼                           │
│                              Report Generator                   │
│                                     │                           │
│                                     ▼                           │
│                              out/report.json                    │
│                                     │                           │
│                                     ▼                           │
│                           AILANG Feedback Loop                  │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### Debug Effect Usage

**AILANG Code:**
```ailang
func updateCiv(civ: Civ, dt: float) -> Civ ! {Debug} {
    -- Assertions (collected, not thrown)
    Debug.assert(civ.population >= 0, "population must be non-negative")
    Debug.assert(civ.stability >= 0.0 && civ.stability <= 1.0,
                 "stability must be in [0, 1]")

    -- Logging
    Debug.log("updating civ " ++ show(civ.id) ++ " pop=" ++ show(civ.population))

    -- Actual update logic
    let newPop = civ.population + calculateGrowth(civ, dt)
    { civ | population: newPop }
}

func step(world: World, input: FrameInput) -> (World, FrameOutput) ! {RNG, Debug} {
    let newWorld = updateWorld(world, input)

    -- Collect all debug output from this tick
    let debugData = Debug.collect()

    (newWorld, { drawCmds: render(newWorld), debug: debugData })
}
```

### Debug Output Types

```ailang
type DebugOutput = {
    logs: [LogEntry],
    assertions: [AssertionResult],
    metrics: [Metric]
}

type LogEntry = {
    message: string,
    location: string,    -- e.g., "sim/civ.ail:42"
    timestamp: int       -- Tick number
}

type AssertionResult = {
    passed: bool,
    message: string,
    location: string,
    timestamp: int
}

type Metric = {
    name: string,
    value: float,
    unit: string        -- e.g., "ms", "count", "bytes"
}
```

### Eval Harness

**cmd/eval/main.go:**
```go
func main() {
    scenarios := loadScenarios()
    results := make([]ScenarioResult, 0)

    for _, scenario := range scenarios {
        result := runScenario(scenario)
        results = append(results, result)
    }

    report := generateReport(results)
    writeJSON("out/report.json", report)
}

func runScenario(s Scenario) ScenarioResult {
    // Set deterministic seed
    os.Setenv("AILANG_SEED", strconv.Itoa(s.Seed))

    world := sim_gen.InitWorld(int64(s.Seed))

    var allDebug []sim_gen.DebugOutput
    var allFrameTimes []time.Duration

    for tick := 0; tick < s.Ticks; tick++ {
        start := time.Now()
        newWorld, output, err := sim_gen.Step(world, s.Inputs[tick])
        elapsed := time.Since(start)

        if err != nil {
            return ScenarioResult{
                Name:   s.Name,
                Status: "error",
                Error:  err.Error(),
            }
        }

        allDebug = append(allDebug, output.Debug)
        allFrameTimes = append(allFrameTimes, elapsed)
        world = newWorld
    }

    return ScenarioResult{
        Name:       s.Name,
        Status:     analyzeAssertions(allDebug),
        FrameStats: computeStats(allFrameTimes),
        Debug:      aggregateDebug(allDebug),
    }
}
```

### Scenario Definition

**scenarios/basic_sim.json:**
```json
{
    "name": "basic_sim_100_ticks",
    "description": "Basic simulation for 100 ticks, no input",
    "seed": 42,
    "ticks": 100,
    "inputs": "none",
    "expected": {
        "assertions_passed": true,
        "final_tick": 100,
        "civs_alive": ">= 1"
    }
}
```

### Report Format

**out/report.json:**
```json
{
    "version": "0.1.0",
    "timestamp": "2025-12-02T10:30:00Z",
    "ailang_version": "0.5.1",
    "summary": {
        "scenarios_run": 5,
        "scenarios_passed": 4,
        "scenarios_failed": 1,
        "total_assertions": 1234,
        "assertions_failed": 3,
        "avg_frame_time_ms": 2.3
    },
    "scenarios": [
        {
            "name": "basic_sim_100_ticks",
            "status": "passed",
            "frame_stats": {
                "min_ms": 1.2,
                "max_ms": 5.4,
                "avg_ms": 2.1,
                "p99_ms": 4.8
            }
        }
    ],
    "failed_assertions": [],
    "ailang_feedback": []
}
```

### Implementation Plan

**Phase 1: Debug Output** (~1 day)
- [ ] Define DebugOutput types in protocol.ail
- [ ] Add Debug.assert/log calls to key functions
- [ ] Collect debug in step() function

**Phase 2: Eval Harness** (~1.5 days)
- [ ] Implement scenario loader
- [ ] Implement scenario runner
- [ ] Generate report.json

**Phase 3: Visual Testing** (~1 day)
- [ ] Screenshot capture in headless mode
- [ ] Golden image comparison
- [ ] Diff image generation

**Phase 4: AILANG Integration** (~0.5 days)
- [ ] Feedback generation from report
- [ ] Integration with ailang-feedback skill
- [ ] CI pipeline setup

### Files to Modify/Create

**AILANG source:**
- `sim/protocol.ail` - Add DebugOutput types (~50 LOC)

**Go source:**
- `cmd/eval/main.go` - Eval harness (~400 LOC)
- `engine/testing/screenshot.go` - Visual testing (~200 LOC)

**Config:**
- `scenarios/*.json` - Test scenario definitions
- `golden/*.png` - Golden images

## Success Criteria

- [ ] All assertions collected in DebugOutput
- [ ] Eval harness runs 5+ scenarios
- [ ] Report JSON is valid and complete
- [ ] Visual comparison catches 1px changes

## References

- [consumer-contract-v0.5.md](../../ailang_resources/consumer-contract-v0.5.md) - Debug effect spec
- [game-vision.md](../../docs/game-vision.md) - AI-assisted development section

---

**Document created**: 2025-12-02
**Last updated**: 2025-12-02
