# Sprint: Engine Discriminator Struct Adaptation

## Sprint ID: `engine-discrim-001`
## Duration: 1 day (focused sprint)
## Priority: P0 (Blocking AILANG integration)

---

## Goal

Update the Go engine layer to work with AILANG-generated discriminator structs instead of interface-based ADTs. This unblocks full AILANG codegen integration.

---

## Pre-Sprint Checklist

- [x] AILANG codegen working (`ailang compile --emit-go`)
- [x] Multi-file compilation working
- [x] Design doc created
- [x] Mock sim_gen available as fallback

---

## Task Breakdown

### Task 1: Camera Type Adaptation (15 min)
**Files:** `engine/camera/transform.go`, `engine/camera/viewport.go`

The Camera struct is a simple record (not ADT). Field names change:
- Mock: `Camera.X`, `Camera.Y`, `Camera.Zoom`
- Generated: `Camera.X`, `Camera.Y`, `Camera.Zoom` (same! Records use named fields)

**Action:** Verify Camera works as-is, may need int64→float64 conversions.

### Task 2: FrameOutput.Draw Type Casting (30 min)
**File:** `engine/render/draw.go`

The generated `FrameOutput.Draw` is `interface{}`. Need to cast to slice.

```go
// Add at start of RenderFrame
var drawCmds []*sim_gen.DrawCmd
switch d := out.Draw.(type) {
case []*sim_gen.DrawCmd:
    drawCmds = d
case []interface{}:
    // Convert each element
    for _, elem := range d {
        if cmd, ok := elem.(*sim_gen.DrawCmd); ok {
            drawCmds = append(drawCmds, cmd)
        }
    }
}
```

### Task 3: Convert Main Render Switch (2 hours)
**File:** `engine/render/draw.go` - `RenderFrame()` function

Convert from:
```go
switch c := s.cmd.(type) {
case sim_gen.DrawCmdRect:
    // use c.X, c.Y, c.W, c.H, c.Color
}
```

To:
```go
switch s.cmd.Kind {
case sim_gen.DrawCmdKindRect:
    c := s.cmd.Rect
    // use c.Value0 (x), c.Value1 (y), etc.
}
```

**Variants to convert:**
- [ ] DrawCmdRect
- [ ] DrawCmdSprite
- [ ] DrawCmdText
- [ ] DrawCmdIsoTile
- [ ] DrawCmdIsoEntity
- [ ] DrawCmdUi
- [ ] DrawCmdLine
- [ ] DrawCmdTextWrapped
- [ ] DrawCmdCircle
- [ ] DrawCmdRectScreen
- [ ] DrawCmdGalaxyBg
- [ ] DrawCmdStar

### Task 4: Convert Helper Functions (1 hour)
**File:** `engine/render/draw.go`

Update helper functions that take DrawCmd:
- [ ] `getZ(cmd sim_gen.DrawCmd) int`
- [ ] `getIsoSortKey(cmd sim_gen.DrawCmd, ...)`
- [ ] `drawSprite(screen, c sim_gen.DrawCmdSprite, ...)`
- [ ] All `draw*` methods

### Task 5: UiKind Enum Handling (30 min)
**File:** `engine/render/draw.go` - `drawUiElement()`

UiKind is also a discriminator struct:
```go
// Before
switch c.Kind {
case sim_gen.UiKindPanel:
}

// After
switch c.Value1.Kind {
case sim_gen.UiKindKindUiPanel:
    // or whatever the generated name is
}
```

Check generated `sim_gen/types.go` for exact enum names.

### Task 6: Build & Test (30 min)

```bash
# 1. Generate AILANG code
make sim

# 2. Build game
make game

# 3. Run and verify visually
./bin/game

# 4. Run visual tests
make test-visual
```

---

## Field Reference Card

Quick reference for positional fields during conversion:

| DrawCmd | Value0 | Value1 | Value2 | Value3 | Value4 | Value5 | Value6 |
|---------|--------|--------|--------|--------|--------|--------|--------|
| **Rect** | x | y | w | h | color | z | - |
| **Sprite** | id | x | y | z | - | - | - |
| **Text** | text | x | y | fontSize | color | z | - |
| **Line** | x1 | y1 | x2 | y2 | color | width | z |
| **Circle** | x | y | radius | color | filled | z | - |
| **Star** | x | y | spriteId | scale | alpha | z | - |

---

## Rollback Plan

If blocked for >1 hour on any task:
1. `git stash` current changes
2. `git checkout sim_gen/` to restore mock
3. Report issue to AILANG via `ailang-feedback` skill
4. Continue with mock until fix available

---

## Definition of Done

- [ ] `make sim` compiles AILANG to Go
- [ ] `make game` builds without errors
- [ ] Game window opens and renders
- [ ] NPCs visible and moving
- [ ] No visual regressions from mock version

---

## AILANG Feedback (post-sprint)

After completion, report:
1. Any codegen issues encountered
2. Suggestions for named fields in ADT payloads
3. Typed list fields vs interface{}
