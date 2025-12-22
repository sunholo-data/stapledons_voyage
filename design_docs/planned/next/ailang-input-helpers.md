# AILANG Input Helpers

**Status**: Planned
**Target**: v0.2.0
**Priority**: P1 - High
**Estimated**: 1 day
**Dependencies**: None (engine input capture already complete)

## Game Vision Alignment

**Score this feature against Stapledon's Voyage core pillars:**

| Pillar | Relevance | Score | Notes |
|--------|-----------|-------|-------|
| Time Dilation Consequence | N/A | 0 | Infrastructure feature |
| Civilization Simulation | N/A | 0 | Infrastructure feature |
| Philosophical Depth | N/A | 0 | Infrastructure feature |
| Ship & Crew Life | N/A | 0 | Infrastructure feature |
| Legacy Impact | N/A | 0 | Infrastructure feature |
| Hard Sci-Fi Authenticity | N/A | 0 | Infrastructure feature |
| **Net Score** | | **0** | **Decision: Move forward** |

**Feature type:** Infrastructure
- Input helpers are enabling tech that allows game features to be built in AILANG
- No negative scores; enables proper AILANG-first development

**Reference:** See [game-vision.md](../../docs/game-vision.md)

## Problem Statement

Demos and game code are handling input inconsistently, with many bypassing AILANG and handling keys directly in Go. This violates the core architecture ("all game logic in AILANG") and causes wasted time reinventing input handling.

**Current State:**
- Engine correctly captures ALL keys and passes them in `FrameInput.keys` as `[KeyEvent]`
- Each `KeyEvent` has `key: int` (Ebiten key code) and `kind: string` ("press", "down", "up")
- **BUT**: No AILANG helper functions exist to easily query this list
- Result: Demos handle keys in Go (see `demo-engine-scene-bridge` V key handling)

**Examples of wasted effort:**
- `demo-engine-scene-bridge`: V key handling in Go (lines 190-195)
- `demo-engine-dome`: V key handling duplicated in Go
- Multiple demos each reinventing arrow key handling
- Tab key mentioned in design but never implemented

**Root cause:** AILANG has the data (`FrameInput.keys`) but no ergonomic way to use it.

**Impact:**
- Every new demo wastes time figuring out input handling
- Inconsistent patterns across demos
- Violates AILANG-first architecture
- Harder to add input rebinding later

## Goals

**Primary Goal:** Create standard AILANG helper functions that make it trivial to check key states from `FrameInput.keys`.

**Success Metrics:**
- New demos can check any key with a one-liner: `is_key_pressed(input.keys, KEY_V)`
- V key handling moved from Go to AILANG in existing demos
- No new Go-side key handling added for game logic
- Clear documentation of key code constants

## Solution Design

### Overview

Create `sim/input.ail` with:
1. **Key code constants** - Named constants for all Ebiten key codes
2. **Helper functions** - Query functions for common input patterns
3. **Documentation** - Clear usage examples in the file

### Architecture

The engine already does the hard work:
```
Ebiten → CaptureInputWithCamera() → FrameInput.keys: [KeyEvent] → AILANG
```

AILANG just needs helper functions to query this list:
```ailang
-- Check if V was just pressed this frame
if is_key_just_pressed(input.keys, KEY_V) then ...

-- Check if Left Arrow is held
if is_key_held(input.keys, KEY_LEFT) then ...
```

**Components:**
1. **Key Constants**: Named constants for Ebiten key codes (KEY_V, KEY_TAB, KEY_ESCAPE, etc.)
2. **Query Functions**: `is_key_just_pressed`, `is_key_held`, `is_key_just_released`
3. **Convenience Functions**: `get_movement_vector` from FlightInput

### Implementation Plan

**Phase 1: Core Helpers** (~2 hours)
- [ ] Create `sim/input.ail` module
- [ ] Define key code constants (match Ebiten key codes)
- [ ] Implement `is_key_just_pressed(keys: [KeyEvent], keyCode: int) -> bool`
- [ ] Implement `is_key_held(keys: [KeyEvent], keyCode: int) -> bool`
- [ ] Implement `is_key_just_released(keys: [KeyEvent], keyCode: int) -> bool`

**Phase 2: Convenience Helpers** (~1 hour)
- [ ] Implement `get_key_state(keys: [KeyEvent], keyCode: int) -> KeyState`
- [ ] Implement `any_key_just_pressed(keys: [KeyEvent]) -> Option<int>`
- [ ] Implement movement helpers using FlightInput

**Phase 3: Migration & Documentation** (~1 hour)
- [ ] Add usage examples to `sim/input.ail` header
- [ ] Document key codes available
- [ ] Update CLAUDE.md with input helper usage

### Files to Modify/Create

**New files:**
- `sim/input.ail` - Input helper module (~100 LOC)

**Modified files:**
- `CLAUDE.md` - Add input helper documentation (~20 LOC)

## Examples

### Example 1: V Key for Velocity Cycling

**Before (Go - wrong):**
```go
// In demo-engine-scene-bridge/main.go
func (g *Game) Update() error {
    if inpututil.IsKeyJustPressed(ebiten.KeyV) {
        velocities := []float64{0.0, 0.2, 0.5, 0.8}
        g.velocityIdx = (g.velocityIdx + 1) % len(velocities)
        g.shipVelocity = velocities[g.velocityIdx]
    }
}
```

**After (AILANG - correct):**
```ailang
import sim/input (is_key_just_pressed, KEY_V)

-- In step function
pure func handle_velocity_input(state: ShipState, keys: [KeyEvent]) -> ShipState {
    if is_key_just_pressed(keys, KEY_V) then
        let velocities = [0.0, 0.2, 0.5, 0.8]
        let newIdx = (state.velocityIdx + 1) % 4
        { state | velocityIdx: newIdx, velocity: velocities[newIdx] }
    else
        state
}
```

### Example 2: Tab Key for Menu

**Usage:**
```ailang
import sim/input (is_key_just_pressed, KEY_TAB)

pure func handle_menu_toggle(state: UIState, keys: [KeyEvent]) -> UIState {
    if is_key_just_pressed(keys, KEY_TAB) then
        { state | menuOpen: not state.menuOpen }
    else
        state
}
```

### Example 3: Escape Key for Exit

**Usage:**
```ailang
import sim/input (is_key_held, KEY_ESCAPE)

pure func should_exit(keys: [KeyEvent]) -> bool {
    is_key_held(keys, KEY_ESCAPE)
}
```

## API Reference

### Key Code Constants

```ailang
-- Common keys (matches Ebiten key codes exactly)
-- Letters (A=0 through Z=25)
export let KEY_A = 0
export let KEY_B = 1
export let KEY_C = 2
export let KEY_D = 3
export let KEY_E = 4
export let KEY_G = 6
export let KEY_I = 8
export let KEY_M = 12
export let KEY_P = 15
export let KEY_Q = 16
export let KEY_R = 17
export let KEY_S = 18
export let KEY_V = 21
export let KEY_W = 22
export let KEY_X = 23

-- Arrow keys
export let KEY_DOWN = 28
export let KEY_LEFT = 29
export let KEY_RIGHT = 30
export let KEY_UP = 31

-- Numbers (0-9 = 43-52)
export let KEY_0 = 43
export let KEY_1 = 44
export let KEY_2 = 45
export let KEY_3 = 46

-- Control keys
export let KEY_ENTER = 54
export let KEY_ESCAPE = 56
export let KEY_F1 = 57
export let KEY_F11 = 67
export let KEY_SPACE = 116
export let KEY_TAB = 117
export let KEY_SHIFT = 120
```

### Helper Functions

```ailang
-- Returns true if key was just pressed this frame (edge detection)
export pure func is_key_just_pressed(keys: [KeyEvent], keyCode: int) -> bool

-- Returns true if key is currently held down
export pure func is_key_held(keys: [KeyEvent], keyCode: int) -> bool

-- Returns true if key was just released this frame
export pure func is_key_just_released(keys: [KeyEvent], keyCode: int) -> bool

-- Get the state of a specific key (for more complex logic)
export type KeyState = KeyUp | KeyJustPressed | KeyHeld | KeyJustReleased
export pure func get_key_state(keys: [KeyEvent], keyCode: int) -> KeyState
```

## Success Criteria

- [ ] `sim/input.ail` compiles with `ailang check`
- [ ] All key code constants match Ebiten values
- [ ] `is_key_just_pressed` returns true only for "press" events
- [ ] `is_key_held` returns true for "down" events
- [ ] `is_key_just_released` returns true only for "up" events
- [ ] All tests passing
- [ ] Documentation in CLAUDE.md updated
- [ ] At least one demo migrated to use AILANG input helpers

## Testing Strategy

**Unit tests:**
- Test each helper function with mock KeyEvent lists
- Verify correct key code matching
- Verify correct event kind filtering

**Inline tests (AILANG):**
```ailang
pure func is_key_just_pressed(keys: [KeyEvent], keyCode: int) -> bool
tests [
    ([], 0, false),                                       -- empty list
    ([{key: 21, kind: "press"}], 21, true),              -- V pressed
    ([{key: 21, kind: "down"}], 21, false),              -- V held, not pressed
    ([{key: 21, kind: "press"}], 0, false),              -- wrong key
]
{ ... }
```

**Manual testing:**
- Run a demo using the new helpers
- Verify V, Tab, Escape keys work as expected
- Verify no lag or missed inputs

## Non-Goals

**Not in this feature:**
- Input rebinding UI - Deferred to future input-rebinding.md design
- Gamepad/controller support - Different input paradigm
- Mouse gesture detection - Out of scope
- Key repeat handling - AILANG can do this in game logic

## Timeline

**Day 1** (4 hours):
- Phase 1: Core helpers implementation
- Phase 2: Convenience helpers
- Phase 3: Documentation and one demo migration

**Total: ~4 hours in 1 day**

## Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|-----------|
| Key code mismatch | Med | Verify against Ebiten constants at compile time |
| List iteration performance | Low | AILANG list ops are O(n) but key lists are tiny (~10 items max) |
| AILANG codegen issue | Med | Test thoroughly before migrating demos |

## References

- [engine/render/input.go](../../engine/render/input.go) - Current input capture
- [sim/protocol.ail](../../sim/protocol.ail) - KeyEvent and FrameInput types
- [design_docs/planned/future/input-rebinding.md](future/input-rebinding.md) - Future rebinding system
- Ebiten key codes: https://pkg.go.dev/github.com/hajimehoshi/ebiten/v2#Key

## Future Work

- **Input Rebinding**: Allow players to customize key bindings (see input-rebinding.md)
- **Input History**: Track input sequences for combo detection
- **Input Buffering**: Handle rapid inputs that span frame boundaries
- **Gamepad Support**: Add gamepad input to FrameInput

---

**Document created**: 2025-12-22
**Last updated**: 2025-12-22
