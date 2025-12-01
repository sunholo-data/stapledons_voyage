# Engine Extensions for UI Modes

**Version:** 0.4.5
**Status:** Planned
**Priority:** P0 (Prerequisite for v0.5.x)
**Complexity:** Low
**Location:** Pure engine code (engine/ and sim_gen/draw_cmd.go)
**Depends On:** None

## Problem Statement

Before implementing UI modes (v0.5.0+), the engine needs additional DrawCmd types that the design docs assume exist but don't yet.

## New DrawCmd Types Needed

### 1. DrawCmdLine (High Priority)

**Needed for:** Network edges, graph connections, timeline connectors

```go
// sim_gen/draw_cmd.go
type DrawCmdLine struct {
    X1, Y1 float64  // Start point
    X2, Y2 float64  // End point
    Color  int
    Width  float64  // Line thickness (1.0 = 1 pixel)
    Z      int
}
```

**Engine implementation:** Use `vector.StrokeLine` from Ebiten.

### 2. DrawCmdTextWrapped (High Priority)

**Needed for:** Dialogue text, descriptions, log entries

```go
type DrawCmdTextWrapped struct {
    Text      string
    X, Y      float64
    MaxWidth  float64  // Wrap at this width
    FontSize  int      // 0=small, 1=normal, 2=large
    Color     int
    Z         int
}
```

**Engine implementation:** Word-wrap algorithm + multiple font sizes.

### 3. DrawCmdCircle (Medium Priority)

**Needed for:** Star nodes, sociogram nodes

```go
type DrawCmdCircle struct {
    X, Y   float64
    Radius float64
    Color  int
    Filled bool    // true=filled, false=outline
    Z      int
}
```

**Engine implementation:** Use `vector.DrawFilledCircle` or path.

### 4. Extended DrawCmdText (Medium Priority)

Add font size and color to existing text:

```go
type DrawCmdText struct {
    Text     string
    X, Y     float64
    FontSize int  // 0=small, 1=normal, 2=large, 3=title
    Color    int  // Text color index
    Z        int
}
```

### 5. DrawCmdUi Extensions

Add more UiKind values:

```go
const (
    UiKindPanel UiKind = iota
    UiKindButton
    UiKindLabel
    UiKindPortrait
    // New:
    UiKindSlider      // For velocity selection
    UiKindProgressBar // For journey progress
    UiKindScrollbar   // For lists
)
```

## Font System Enhancement

Currently only one font. Need:

```go
// engine/assets/fonts.go
type FontSet struct {
    Small  font.Face  // 10pt - labels
    Normal font.Face  // 14pt - body
    Large  font.Face  // 18pt - headers
    Title  font.Face  // 24pt - screen titles
}

func (m *Manager) GetFont(size int) font.Face
```

## Implementation Plan

### Phase 1: Line Drawing

| Task | File | Description |
|------|------|-------------|
| 1.1 | `sim_gen/draw_cmd.go` | Add DrawCmdLine type |
| 1.2 | `engine/render/draw.go` | Implement line rendering |
| 1.3 | Test | Draw line on screen |

### Phase 2: Text Wrapping

| Task | File | Description |
|------|------|-------------|
| 2.1 | `sim_gen/draw_cmd.go` | Add DrawCmdTextWrapped type |
| 2.2 | `engine/render/text.go` | Word-wrap algorithm |
| 2.3 | `engine/assets/fonts.go` | Multiple font sizes |
| 2.4 | Test | Wrapped text displays correctly |

### Phase 3: Circles

| Task | File | Description |
|------|------|-------------|
| 3.1 | `sim_gen/draw_cmd.go` | Add DrawCmdCircle type |
| 3.2 | `engine/render/draw.go` | Circle rendering |
| 3.3 | Test | Circles render filled and outline |

### Phase 4: UI Extensions

| Task | File | Description |
|------|------|-------------|
| 4.1 | `sim_gen/protocol.go` | Add new UiKind values |
| 4.2 | `engine/render/draw.go` | Slider, progress bar rendering |
| 4.3 | Test | New UI elements work |

## Success Criteria

- [ ] Can draw lines between arbitrary points
- [ ] Text wraps at specified width
- [ ] Multiple font sizes available
- [ ] Circles render correctly
- [ ] New UI elements (slider, progress bar) work

## Notes

This is **engine-only work** - no game logic changes. These are pure rendering primitives that the UI modes will use.

The engine's job is to interpret DrawCmds faithfully. All decisions about *what* to draw remain in sim_gen (AILANG).
