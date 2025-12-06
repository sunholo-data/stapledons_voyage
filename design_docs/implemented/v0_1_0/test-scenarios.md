# Test Scenario Harness

**Status**: Implemented
**Target**: v0.5.0 (Engine-side, independent of AILANG)
**Priority**: P1 - Enables comprehensive AI self-testing
**Estimated**: 2 implementation sessions
**Dependencies**: screenshot-mode.md

## Game Vision Alignment

**Feature type:** Development Infrastructure

| Pillar | Relevance | Score | Notes |
|--------|-----------|-------|-------|
| Time Dilation Consequence | N/A | 0 | Development tooling |
| Civilization Simulation | N/A | 0 | Development tooling |
| Philosophical Depth | N/A | 0 | Development tooling |
| Ship & Crew Life | N/A | 0 | Development tooling |
| Legacy Impact | N/A | 0 | Development tooling |
| Hard Sci-Fi Authenticity | N/A | 0 | Development tooling |
| AI-Assisted Development | +++ | +3 | Core enabler for automated scenarios |
| **Net Score** | | **+3** | **Decision: Move forward** |

**Rationale:** Extends screenshot mode with scripted input sequences. Enables AI to test interactive features (camera movement, clicking, building) without a human.

## Problem Statement

**Current State:**
- Screenshot mode captures static frames
- No way to simulate user input (keypresses, clicks)
- Cannot test interactive features automatically
- AI cannot verify "if I press W, camera should move up"

**Impact:**
- Interactive features require human testing
- Cannot catch input handling regressions
- AI development limited to static rendering

## Goals

**Primary Goal:** Create a scenario system that simulates user input and captures screenshots at key points.

**Success Metrics:**
- `./bin/game --scenario camera-pan` runs predefined inputs
- Scenario outputs multiple screenshots showing state progression
- AI can verify camera moved correctly by comparing screenshots

## Solution Design

### Overview

Define test scenarios as JSON/YAML files containing:
- Initial world setup (seed, camera position)
- Sequence of timed input events (key presses, clicks)
- Capture points (which frames to screenshot)
- Expected outputs (optional golden images)

### Architecture

```
scenarios/
  camera-pan.json
  tile-selection.json
  building.json

JSON Structure:
{
  "name": "camera-pan",
  "seed": 1234,
  "camera": {"x": 0, "y": 0, "zoom": 1.0},
  "events": [
    {"frame": 0, "capture": "start.png"},
    {"frame": 1, "key": "W", "action": "down"},
    {"frame": 30, "key": "W", "action": "up"},
    {"frame": 30, "capture": "after-pan.png"}
  ]
}

Execution:
    ↓
Load scenario JSON
    ↓
Initialize world with seed
    ↓
For each frame:
    - Inject scheduled inputs
    - Step simulation
    - Capture if scheduled
    ↓
Output screenshots to out/scenarios/<name>/
```

### Implementation Plan

**Phase 1: Scenario Definition**
- [ ] Define `Scenario` struct with events, captures
- [ ] Create `scenarios/` directory
- [ ] Write 3 example scenario JSON files

**Phase 2: Scenario Runner**
- [ ] Create `engine/scenario/runner.go`
- [ ] Parse scenario JSON
- [ ] Inject inputs at scheduled frames
- [ ] Capture screenshots at capture points

**Phase 3: CLI Integration**
- [ ] Add `--scenario <name>` flag to game
- [ ] Load scenario from `scenarios/<name>.json`
- [ ] Output to `out/scenarios/<name>/`

**Phase 4: Built-in Scenarios**
- [ ] `camera-pan` - Test WASD camera movement
- [ ] `camera-zoom` - Test Q/E zoom
- [ ] `tile-select` - Test click-to-select
- [ ] `build-house` - Test B key building
- [ ] `npc-movement` - Capture NPC motion over time

**Phase 5: Golden Image Comparison (Optional)**
- [ ] Store expected screenshots as `golden/`
- [ ] Compare output to golden with tolerance
- [ ] Report pass/fail

### Files to Modify/Create

**New files:**
- `engine/scenario/scenario.go` - Scenario struct and parser (~100 LOC)
- `engine/scenario/runner.go` - Scenario execution (~150 LOC)
- `scenarios/camera-pan.json` - Example scenario
- `scenarios/tile-select.json` - Example scenario
- `scenarios/npc-movement.json` - Example scenario

**Modified files:**
- `cmd/game/main.go` - Add `--scenario` flag (~20 LOC)

## Examples

### Example 1: Camera Pan Scenario

**scenarios/camera-pan.json:**
```json
{
  "name": "camera-pan",
  "description": "Test camera movement with WASD",
  "seed": 1234,
  "camera": {"x": 0, "y": 0, "zoom": 1.0},
  "events": [
    {"frame": 0, "capture": "initial.png"},
    {"frame": 1, "key": "S", "action": "down"},
    {"frame": 60, "key": "S", "action": "up"},
    {"frame": 60, "capture": "after-down.png"},
    {"frame": 61, "key": "D", "action": "down"},
    {"frame": 120, "key": "D", "action": "up"},
    {"frame": 120, "capture": "after-right.png"}
  ]
}
```

**Execution:**
```bash
./bin/game --scenario camera-pan

# Outputs:
# out/scenarios/camera-pan/initial.png
# out/scenarios/camera-pan/after-down.png
# out/scenarios/camera-pan/after-right.png
```

### Example 2: Tile Selection Scenario

**scenarios/tile-select.json:**
```json
{
  "name": "tile-select",
  "description": "Test click to select tiles",
  "seed": 1234,
  "camera": {"x": 0, "y": 0, "zoom": 1.0},
  "events": [
    {"frame": 0, "capture": "no-selection.png"},
    {"frame": 1, "click": {"x": 320, "y": 240}, "button": "left"},
    {"frame": 2, "capture": "selected.png"},
    {"frame": 3, "key": "I", "action": "press"},
    {"frame": 4, "capture": "inspected.png"}
  ]
}
```

### Example 3: AI Verification Workflow

```bash
# Run scenario
./bin/game --scenario camera-pan

# AI reads outputs
Read out/scenarios/camera-pan/initial.png
# -> Sees camera at (0, 0)

Read out/scenarios/camera-pan/after-down.png
# -> Sees camera moved down (y increased)

Read out/scenarios/camera-pan/after-right.png
# -> Sees camera moved right (x increased)

# AI concludes: Camera movement working correctly!
```

## Success Criteria

- [ ] `--scenario camera-pan` produces 3 screenshots
- [ ] Screenshots show camera position changing correctly
- [ ] `--scenario tile-select` shows selection highlight appearing
- [ ] Scenarios are deterministic (same output every run)
- [ ] All scenarios complete in < 5 seconds

## Testing Strategy

**Unit tests:**
- Scenario JSON parsing
- Event scheduling logic
- Input injection

**Integration tests:**
- Full scenario execution
- Output file creation
- Determinism check

**AI verification:**
- Run each scenario
- Read all output screenshots
- Verify expected visual changes

## Non-Goals

- **Interactive scenario editor** - JSON editing only
- **Complex input patterns** - Simple key/click events only
- **Network/multiplayer** - Single-player scenarios only
- **Audio verification** - Visual only

## Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|-----------|
| Timing-sensitive tests | Med | Use frame counts not real time |
| Input injection complexity | Med | Keep event model simple |
| Scenario maintenance | Low | Start with essential scenarios only |

## Future Work

- **Golden image comparison** - Automatic pass/fail with tolerance
- **Scenario recording** - Record human play as scenario
- **Coverage reporting** - Track which features have scenarios
- **CI integration** - Run scenarios on every PR

## References

- [screenshot-mode.md](screenshot-mode.md) - Base screenshot capability
- [game-vision.md](../../../docs/game-vision.md) - AI-Assisted Development section

---

**Document created**: 2025-12-01
**Last updated**: 2025-12-01
