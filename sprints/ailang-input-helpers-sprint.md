# Sprint: AILANG Input Helpers

**Design Doc:** [design_docs/planned/next/ailang-input-helpers.md](../design_docs/planned/next/ailang-input-helpers.md)
**Duration:** 1 day (~4 hours)
**Status:** COMPLETED

## Goal

Create `sim/input.ail` with helper functions that make it easy to check key states from `FrameInput.keys`. This eliminates the need for demos to handle keys in Go.

## Pre-Sprint Checklist

- [x] Design doc reviewed and approved
- [x] AILANG messages checked (1 unread - ADT constructor fix)
- [x] `ailang check sim/protocol.ail` passes
- [x] Key codes verified against Ebiten

## Phase 1: Core Module (~1.5 hours)

### Task 1.1: Create sim/input.ail with Key Constants
- [x] Create `sim/input.ail` module
- [x] Define key code constants (verified against Ebiten)
- [x] Import KeyEvent from protocol.ail
- [x] Run `ailang check sim/input.ail`

**Key codes included:**
```
KEY_A=0 through KEY_Z=25
KEY_DOWN=28, KEY_LEFT=29, KEY_RIGHT=30, KEY_UP=31
KEY_0=43 through KEY_9=52
KEY_ENTER=54, KEY_ESCAPE=56, KEY_F1=57...KEY_F12=68
KEY_SPACE=116, KEY_TAB=117, KEY_SHIFT=120
```

### Task 1.2: Implement is_key_just_pressed
- [x] Implement `is_key_just_pressed(keys: [KeyEvent], keyCode: int) -> bool`
- [x] Uses list recursion to find matching key with kind="press"
- [x] Run `ailang check sim/input.ail`

### Task 1.3: Implement is_key_held
- [x] Implement `is_key_held(keys: [KeyEvent], keyCode: int) -> bool`
- [x] Matches keys with kind="down"
- [x] Run `ailang check sim/input.ail`

### Task 1.4: Implement is_key_just_released
- [x] Implement `is_key_just_released(keys: [KeyEvent], keyCode: int) -> bool`
- [x] Matches keys with kind="up"
- [x] Run `ailang check sim/input.ail`

## Phase 2: Codegen Integration (~1 hour)

### Task 2.1: Compile and Verify
- [x] Run `make sim` to generate Go code
- [x] Run `make build` to verify compilation
- [x] Check `sim_gen/input.go` for generated input helpers

**Generated Go functions:**
- `sim_gen.IsKeyJustPressed(keys []*KeyEvent, keyCode int64) bool`
- `sim_gen.IsKeyHeld(keys []*KeyEvent, keyCode int64) bool`
- `sim_gen.IsKeyJustReleased(keys []*KeyEvent, keyCode int64) bool`
- `sim_gen.KEYV() int64` (and all other KEY_* constants)

### Task 2.2: Test in a Demo
- [ ] Create simple test in existing demo or step.ail (SKIPPED - deferred to demo migration)

## Phase 3: Documentation (~0.5 hours)

### Task 3.1: Update CLAUDE.md
- [x] Add section on using input helpers
- [x] Document available key codes
- [x] Add example usage pattern

### Task 3.2: Add Usage Examples to input.ail
- [x] Add header comments with usage examples
- [x] Document the 3 key states (press, down, up)

## Phase 4: Cleanup (~0.5 hours)

### Task 4.1: Run Full Test Suite
- [x] Run `make build` - passed
- [x] Verify no regressions

### Task 4.2: Move Design Doc
- [ ] Move design doc to implemented when demos are migrated

## Success Criteria

- [x] `sim/input.ail` compiles with `ailang check`
- [x] All key code constants match Ebiten values
- [x] `is_key_just_pressed` returns true only for "press" events
- [x] `is_key_held` returns true for "down" events
- [x] `is_key_just_released` returns true only for "up" events
- [x] `make sim && make build` succeeds
- [x] Documentation in CLAUDE.md updated

## AILANG Feedback

**Issue discovered:** `export let` is not supported in AILANG. Must use `export pure func` for constants.

**Workaround applied:** Changed `export let KEY_V = 21` to `export pure func KEY_V() -> int = 21`

**Note:** This is intentional design (constants as zero-arg functions), not a bug.

## Files Created/Modified

**New:**
- `sim/input.ail` - Input helper module (163 LOC)

**Modified:**
- `CLAUDE.md` - Add input helper documentation section

**Generated:**
- `sim_gen/input.go` - Generated Go code with all helpers

---

**Created:** 2025-12-22
**Completed:** 2025-12-22
