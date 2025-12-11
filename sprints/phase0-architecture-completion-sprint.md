# Sprint: Phase 0 Architecture Completion

**Status:** Planned
**Duration:** 3-4 days
**Priority:** P0 (BLOCKS ALL OTHER PHASES)
**Design Docs:**
- [view-layer-ailang-migration.md](../design_docs/planned/phase0-architecture/view-layer-ailang-migration.md)
- [view-types-cleanup.md](../design_docs/planned/phase0-architecture/view-types-cleanup.md)

## Goal

Complete the AILANG-first architecture migration. After this sprint:
- ALL game state lives in AILANG (`sim/*.ail`)
- Engine is "dumb" (just renders DrawCmds)
- `engine/view/` reduced from **3,166 lines → < 200 lines**
- No duplicate types (use `sim_gen.*` everywhere)

## Current State

| Metric | Before | Target |
|--------|--------|--------|
| `engine/view/` LOC | 3,166 | < 200 |
| Duplicate types in `layer.go` | 4 (Camera, Input, Dialogue, DialogueOption) | 0 |
| Go files with game state | 5+ | 0 |
| Dome migration | 60% | 100% |

### Files to Remove/Reduce

| File | Lines | Action | Reason |
|------|-------|--------|--------|
| `layer.go` | 170 | DELETE | Duplicate types |
| `manager.go` | 180 | DELETE | View coordination → AILANG |
| `transition.go` | 175 | DELETE | Transitions → AILANG |
| `ui_layer.go` | 172 | DELETE | UI state → AILANG |
| `bubble_arc.go` | 440 | DELETE | Particles → AILANG |
| `space_view.go` | 156 | DELETE | Space background → AILANG |
| `dome_renderer.go` | 450 | REDUCE → 100 | Keep 3D rendering helpers only |
| `bridge_view.go` | 530 | REDUCE → 50 | Keep render dispatch only |
| `planet_layer.go` | 165 | REDUCE → 30 | Keep Tetra3D helpers only |
| `easing.go` | 180 | KEEP | Pure math utilities |
| `view.go` | 120 | REDUCE → 20 | Keep ViewType enum only |

## Sprint Tasks

### Day 1: Finish Dome Migration ✅ (from existing sprint)

#### Task 1.1: Complete day4-integration
- [ ] Verify `stepBridge` calls `stepDome` with delta time
- [ ] Ensure Go `dome_renderer` no longer owns cruise state
- [ ] All dome animation driven by AILANG `DomeState`

**Test:**
```bash
make run
# Verify planets fly by correctly
# Verify cruise animation loops
```

#### Task 1.2: Complete day5-polish
- [ ] Struts parallax working from AILANG `cameraZ`
- [ ] Star field parallax via GalaxyBg DrawCmd
- [ ] 60 FPS maintained

**Files:** `sim/bridge.ail` (already has most of this)

---

### Day 2: Delete Duplicate Types

#### Task 2.1: Audit Type Usage
- [ ] Run: `grep -r "view\.Camera\|view\.Input\|view\.Dialogue" engine/`
- [ ] List all files using duplicate types
- [ ] Map each usage to `sim_gen.*` equivalent

#### Task 2.2: Replace Camera References
- [ ] Replace `view.Camera` → `sim_gen.Camera` in all files
- [ ] Update function signatures
- [ ] Verify compilation: `go build ./...`

**Before:**
```go
func (v *BridgeView) Render(cam view.Camera)
```

**After:**
```go
func (v *BridgeView) Render(cam *sim_gen.Camera)
```

#### Task 2.3: Replace Input References
- [ ] Replace `view.Input` → `sim_gen.FrameInput`
- [ ] Remove any `HandleInput()` methods (input via AILANG step)
- [ ] Verify compilation

#### Task 2.4: Remove Dialogue Types (Not Used Yet)
- [ ] Delete `Dialogue`, `DialogueOption` from `layer.go`
- [ ] Will add back when AILANG dialogue implemented

#### Task 2.5: Delete layer.go
- [ ] Verify no remaining imports of `engine/view.Camera`, etc.
- [ ] Delete `engine/view/layer.go`
- [ ] Run tests: `go test ./engine/...`

**Files:**
- `engine/view/layer.go` → DELETE
- `engine/view/bridge_view.go` → modify imports
- `engine/view/dome_renderer.go` → modify imports

---

### Day 3: Remove Go View State

#### Task 3.1: Delete View Manager
- [ ] Audit what `manager.go` does
- [ ] View transitions should come from AILANG `ViewState`
- [ ] Move any needed logic to AILANG
- [ ] Delete `engine/view/manager.go`

#### Task 3.2: Delete Transition System
- [ ] Transitions controlled by AILANG `ViewTransition` type
- [ ] Delete `engine/view/transition.go`

#### Task 3.3: Delete UI Layer
- [ ] UI panels defined in AILANG, rendered via DrawCmds
- [ ] Delete `engine/view/ui_layer.go`

#### Task 3.4: Delete Bubble Arc (Particles)
- [ ] Particles already in AILANG (`sim/bridge.ail` debris)
- [ ] Delete `engine/view/bubble_arc.go`

#### Task 3.5: Delete Space View
- [ ] Space background via `GalaxyBg` DrawCmd
- [ ] Delete `engine/view/space_view.go`

**After Day 3:**
```
engine/view/
├── bridge_view.go     # REDUCED - render dispatch only
├── dome_renderer.go   # REDUCED - 3D helpers only
├── planet_layer.go    # REDUCED - Tetra3D helpers only
├── easing.go          # KEEP - pure math
├── easing_test.go     # KEEP - tests
└── view.go            # REDUCED - ViewType enum only
```

---

### Day 4: Reduce Remaining Files + Final Cleanup

#### Task 4.1: Reduce bridge_view.go
**Current:** 530 lines with state, Update(), Init()

**Target:** ~50 lines
```go
// bridge_view.go - Render dispatch only
package view

func RenderBridgeFrame(screen *ebiten.Image, cmds []*sim_gen.DrawCmd) {
    // Sort by Z, dispatch to renderer
    // No state, no Update(), no Init()
}
```

- [ ] Remove all state fields (`state`, `frameCount`, `domeRenderer`, etc.)
- [ ] Remove `Init()`, `Update()` methods
- [ ] Keep only render dispatch
- [ ] Move 3D rendering to `dome_renderer.go`

#### Task 4.2: Reduce dome_renderer.go
**Current:** 450 lines with cruise state, planet configs

**Target:** ~100 lines
```go
// dome_renderer.go - 3D Tetra3D rendering helpers only
package view

func RenderPlanets3D(screen *ebiten.Image, cameraZ float64) {
    // Use Tetra3D to render planets
    // No state - cameraZ comes from AILANG
}
```

- [ ] Remove `cruiseTime`, `cruiseVelocity`, planet configs
- [ ] Keep Tetra3D scene setup and rendering
- [ ] Camera position comes from AILANG `DomeState.cameraZ`

#### Task 4.3: Reduce planet_layer.go
**Current:** 165 lines

**Target:** ~30 lines
- [ ] Keep only Tetra3D mesh/texture loading helpers
- [ ] Planet positions from AILANG

#### Task 4.4: Reduce view.go
**Current:** 120 lines with View interface, ViewLayers

**Target:** ~20 lines
```go
package view

type ViewType int
const (
    ViewBridge ViewType = iota
    ViewGalaxyMap
    ViewShip
)
```

- [ ] Remove `View` interface (no more views with state)
- [ ] Remove `ViewLayers` (layers via DrawCmd Z values)
- [ ] Keep only `ViewType` enum

#### Task 4.5: Final Verification
- [ ] Count total lines: `wc -l engine/view/*.go`
- [ ] Target: < 200 lines
- [ ] Run full test: `make run`
- [ ] Run architecture check: `.claude/skills/game-architect/scripts/check_layer_boundaries.sh`

---

## Success Criteria

- [ ] `engine/view/` contains < 200 lines total
- [ ] No duplicate types (Camera, Input, Dialogue deleted from layer.go)
- [ ] No Go code owns game state (cruise time, positions, etc.)
- [ ] All state comes from AILANG via `sim_gen.*`
- [ ] Game loop is pure: `input → step → render → draw`
- [ ] 60 FPS maintained
- [ ] Architecture check passes with 0 violations

## Line Count Targets

| File | Current | Target | Status |
|------|---------|--------|--------|
| `layer.go` | 170 | DELETE | [ ] |
| `manager.go` | 180 | DELETE | [ ] |
| `transition.go` | 175 | DELETE | [ ] |
| `ui_layer.go` | 172 | DELETE | [ ] |
| `bubble_arc.go` | 440 | DELETE | [ ] |
| `space_view.go` | 156 | DELETE | [ ] |
| `dome_renderer.go` | 450 | 100 | [ ] |
| `bridge_view.go` | 530 | 50 | [ ] |
| `planet_layer.go` | 165 | 30 | [ ] |
| `view.go` | 120 | 20 | [ ] |
| `easing.go` | 180 | 180 | ✅ KEEP |
| `easing_test.go` | 185 | 185 | ✅ KEEP |
| **TOTAL** | **3,166** | **~565** | [ ] |

*Note: Target is ~200 for view logic, plus ~365 for easing utilities and tests.*

## Testing Strategy

After each day:
```bash
# Verify compilation
go build ./...

# Run game
make run

# Check architecture
.claude/skills/game-architect/scripts/check_layer_boundaries.sh

# Count lines
wc -l engine/view/*.go | tail -1
```

## Rollback Plan

If issues arise:
1. Git revert to last working state
2. File AILANG bugs if needed
3. Keep Go fallback temporarily while fixing

## AILANG Issues to Watch

| Issue | Impact | Mitigation |
|-------|--------|------------|
| Missing `floatToStr` | Can't display HUD numbers | Use DrawCmd Text with formatted ints |
| Recursion limits | Deep particle lists | Limit particle count |
| Record update bugs | Type errors | Use helper functions |

## References

- [view-layer-ailang-migration.md](../design_docs/planned/phase0-architecture/view-layer-ailang-migration.md)
- [view-types-cleanup.md](../design_docs/planned/phase0-architecture/view-types-cleanup.md)
- [CLAUDE.md](../CLAUDE.md) - Architecture mandate
- [bridge-dome-migration-sprint.md](bridge-dome-migration-sprint.md) - Previous sprint (60% complete)

---

**Document created**: 2025-12-11
