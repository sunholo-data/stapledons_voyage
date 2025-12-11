# View Layer Duplicate Types Cleanup

## Status
- Status: Planned
- Priority: P1 (Architecture)
- Complexity: Low
- Part of: [view-layer-ailang-migration.md](view-layer-ailang-migration.md)
- Estimated: 0.5 days

## Problem Statement

The file `engine/view/layer.go` defines types that duplicate AILANG types in `sim_gen/`:

```go
// engine/view/layer.go (WRONG - duplicates)
type Camera struct {
    X, Y    float64
    Zoom    float64
    TargetX float64
    // ...
}

type Input struct {
    MouseX, MouseY float64
    Clicked        bool
    // ...
}

type Dialogue struct {
    Speaker string
    Text    string
    Options []DialogueOption
}
```

**Why this is wrong:**
- `sim_gen` already exports these types from AILANG
- Duplicate types require manual synchronization
- Type mismatches cause bugs
- Code uses wrong type in wrong place

**What should exist:**
- `sim_gen.Camera` - from AILANG protocol
- `sim_gen.FrameInput` - from AILANG protocol
- `sim_gen.Dialogue` - from AILANG (when implemented)

## Current Duplicates

| `engine/view/layer.go` | `sim_gen/` equivalent | Status |
|------------------------|----------------------|--------|
| `Camera` | `sim_gen.Camera` | EXISTS - use sim_gen |
| `Input` | `sim_gen.FrameInput` | EXISTS - use sim_gen |
| `Dialogue` | Will be `sim_gen.DialogueState` | Pending AILANG impl |
| `DialogueOption` | Will be `sim_gen.DialogueChoice` | Pending AILANG impl |
| `UIPanel` | Will be `sim_gen.UIElement` | Pending AILANG impl |

## Migration Steps

### Phase 1: Camera Migration

**Before:**
```go
// engine/view/some_view.go
import "stapledons_voyage/engine/view"

func (v *SomeView) Update(cam view.Camera) {
    // Uses duplicate Camera type
}
```

**After:**
```go
// engine/view/some_view.go
import "stapledons_voyage/sim_gen"

func (v *SomeView) Render(screen *ebiten.Image, cam sim_gen.Camera) {
    // Uses AILANG Camera type
}
```

**Steps:**
- [ ] Find all uses of `view.Camera` in engine/
- [ ] Replace with `sim_gen.Camera`
- [ ] Update function signatures
- [ ] Delete `Camera` from layer.go

### Phase 2: Input Migration

**Before:**
```go
// engine/view/some_view.go
func (v *SomeView) HandleInput(input view.Input) {
    if input.Clicked {
        // ...
    }
}
```

**After:**
```go
// Input comes from render.CaptureInput() which returns sim_gen.FrameInput
// Views don't handle input - AILANG step() does
```

**Steps:**
- [ ] Remove `HandleInput()` methods from views
- [ ] Input handling moves to AILANG `step()` function
- [ ] Delete `Input` from layer.go

### Phase 3: Dialogue Migration (After AILANG impl)

Once `sim/dialogue.ail` exists (see [dialogue-system.md](../../planned/future/dialogue-system.md)):

- [ ] Delete `Dialogue` type from layer.go
- [ ] Delete `DialogueOption` type from layer.go
- [ ] Use `sim_gen.DialogueState` and `sim_gen.DialogueChoice`

### Phase 4: UIPanel Migration (After AILANG impl)

Once `sim/ui.ail` exists (see [ui-layout-engine.md](../../planned/future/ui-layout-engine.md)):

- [ ] Delete `UIPanel` type from layer.go
- [ ] Use `sim_gen.UIElement`
- [ ] Layout helpers remain in Go (pure math)

### Phase 5: Delete layer.go

Once all types are migrated:

- [ ] Verify no imports of `engine/view.Camera`, etc.
- [ ] Delete `engine/view/layer.go`
- [ ] Keep only rendering helpers in engine/view/

## Affected Files

Files that currently import duplicate types:

```bash
grep -r "view\.Camera\|view\.Input\|view\.Dialogue" engine/
```

Expected hits:
- `engine/view/bridge_view.go`
- `engine/view/space_view.go`
- `engine/view/dome_renderer.go`
- `engine/view/manager.go`

Each needs updating to use `sim_gen.*` types.

## Type Mapping Reference

| Old (layer.go) | New (sim_gen) | Notes |
|----------------|---------------|-------|
| `Camera.X` | `Camera.X` | Same field |
| `Camera.Y` | `Camera.Y` | Same field |
| `Camera.Zoom` | `Camera.Zoom` | Same field |
| `Input.MouseX` | `FrameInput.MouseX` | Renamed struct |
| `Input.Clicked` | `FrameInput.ClickedThisFrame` | Renamed field |
| `Dialogue.Speaker` | `DialogueState.Speaker` | Different structure |

## Success Criteria

- [ ] No types defined in `engine/view/layer.go`
- [ ] All views use `sim_gen.Camera`
- [ ] No `view.Input` usage (input via AILANG)
- [ ] `layer.go` deleted or contains only helpers
- [ ] Code compiles and runs correctly

## Testing

```bash
# After each phase, verify compilation
go build ./...

# Run game to verify no regressions
make run

# Run tests
go test ./engine/...
```

## References

- [view-layer-ailang-migration.md](view-layer-ailang-migration.md) - Parent migration doc
- [dialogue-system.md](../../planned/future/dialogue-system.md) - Dialogue types
- [ui-layout-engine.md](../../planned/future/ui-layout-engine.md) - UI types
