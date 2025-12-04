# UI Layout Engine

**Version:** 0.4.5
**Status:** Planned
**Priority:** P1 (Critical)
**Complexity:** High
**Package:** `engine/ui`
**AILANG Impact:** None - layout is engine-side; AILANG emits logical UI commands

## Related Documents

- [UI Modes](../v0_5_0/ui-modes.md) - Game UI state machine
- [AILANG and Engine UI](../v0_5_0/ailang-and-engine-ui.md) - Boundary definition
- [Display Config](../v0_2_0/display-config.md) - Screen resolution handling

## Problem Statement

**Current State:**
- UI elements positioned with hardcoded pixel coordinates
- No responsive layout for different resolutions
- No anchoring system (UI breaks on window resize)
- No layout containers (vertical lists, grids, panels)

**What's Needed:**
- Resolution-independent UI positioning
- Anchor system (top-left, center, bottom-right, etc.)
- Layout containers (VBox, HBox, Grid)
- Margin/padding support
- AILANG-to-layout translation layer

**Design Principle:** AILANG outputs *logical* UI (what to show), engine handles *physical* layout (where to show it).

## AILANG / Engine Boundary

### AILANG Outputs (sim layer)

```ailang
type UIElement =
    | Label(string, UIStyle)
    | Button(int, string, UIStyle)        -- id, text
    | Panel(UILayout, [UIElement])        -- layout, children
    | Spacer(float)                       -- relative size

type UILayout = Vertical | Horizontal | Grid(int, int)  -- cols, rows

type UIStyle = {
    anchor: UIAnchor,       -- Where to position
    size: UISize,           -- How to size
    priority: int           -- Z-order (higher = front)
}

type UIAnchor = TopLeft | Top | TopRight | Left | Center | Right | BottomLeft | Bottom | BottomRight

type UISize =
    | Fixed(float, float)          -- pixels
    | Relative(float, float)       -- 0.0-1.0 of parent
    | FitContent                   -- Size to children
    | Fill                         -- Fill remaining space

type FrameOutput = {
    draw: [DrawCmd],
    ui: [UIElement],               -- NEW: UI tree for this frame
    -- ...
}
```

AILANG describes the UI tree. It knows nothing about actual pixel positions.

### Engine Outputs (Go/Ebiten layer)

The engine:
1. Receives `[UIElement]` from FrameOutput
2. Runs layout algorithm to compute screen positions
3. Renders UI elements at computed positions
4. Handles input (click detection, hover states)

## Core Types (engine/ui/types.go)

```go
package ui

// Anchor defines where a widget attaches to its parent
type Anchor int

const (
    AnchorTopLeft Anchor = iota
    AnchorTop
    AnchorTopRight
    AnchorLeft
    AnchorCenter
    AnchorRight
    AnchorBottomLeft
    AnchorBottom
    AnchorBottomRight
)

// SizeMode defines how a widget calculates its size
type SizeMode int

const (
    SizeModeFixed    SizeMode = iota // Exact pixel size
    SizeModeRelative                  // Percentage of parent
    SizeModeFit                       // Fit to content
    SizeModeFill                      // Fill available space
)

// Size represents widget dimensions
type Size struct {
    Mode   SizeMode
    Width  float64 // Pixels or ratio (0-1)
    Height float64
}

// Rect is a positioned rectangle
type Rect struct {
    X, Y          float64
    Width, Height float64
}

// Widget is the base interface for all UI elements
type Widget interface {
    // Layout computes position and size given parent bounds
    Layout(parent Rect) Rect

    // Draw renders the widget to screen
    Draw(screen *ebiten.Image, bounds Rect)

    // Children returns child widgets (for containers)
    Children() []Widget

    // HandleInput processes mouse/keyboard events
    HandleInput(x, y float64, clicked bool) bool
}

// Constraints passed down during layout
type Constraints struct {
    MinWidth, MaxWidth   float64
    MinHeight, MaxHeight float64
}
```

## Layout Algorithm (engine/ui/layout.go)

```go
package ui

// LayoutEngine manages UI positioning
type LayoutEngine struct {
    screenW, screenH float64
    root             Widget
    focused          Widget
    hovered          Widget
}

func NewLayoutEngine(screenW, screenH int) *LayoutEngine {
    return &LayoutEngine{
        screenW: float64(screenW),
        screenH: float64(screenH),
    }
}

// SetRoot updates the UI tree (call each frame with AILANG output)
func (e *LayoutEngine) SetRoot(root Widget) {
    e.root = root
}

// OnResize handles window resize
func (e *LayoutEngine) OnResize(w, h int) {
    e.screenW = float64(w)
    e.screenH = float64(h)
}

// Layout computes positions for entire tree
func (e *LayoutEngine) Layout() {
    if e.root == nil {
        return
    }

    screenRect := Rect{0, 0, e.screenW, e.screenH}
    e.layoutRecursive(e.root, screenRect)
}

func (e *LayoutEngine) layoutRecursive(w Widget, parent Rect) {
    bounds := w.Layout(parent)

    for _, child := range w.Children() {
        e.layoutRecursive(child, bounds)
    }
}

// Draw renders entire UI tree
func (e *LayoutEngine) Draw(screen *ebiten.Image) {
    if e.root == nil {
        return
    }

    screenRect := Rect{0, 0, e.screenW, e.screenH}
    e.drawRecursive(screen, e.root, screenRect)
}

func (e *LayoutEngine) drawRecursive(screen *ebiten.Image, w Widget, parent Rect) {
    bounds := w.Layout(parent)
    w.Draw(screen, bounds)

    for _, child := range w.Children() {
        e.drawRecursive(screen, child, bounds)
    }
}
```

## Anchor Positioning (engine/ui/anchor.go)

```go
package ui

// ApplyAnchor positions a child rect within a parent using anchor
func ApplyAnchor(parent Rect, child Size, anchor Anchor, margin float64) Rect {
    // Compute actual size
    w, h := resolveSize(child, parent)

    var x, y float64

    switch anchor {
    case AnchorTopLeft:
        x = parent.X + margin
        y = parent.Y + margin
    case AnchorTop:
        x = parent.X + (parent.Width-w)/2
        y = parent.Y + margin
    case AnchorTopRight:
        x = parent.X + parent.Width - w - margin
        y = parent.Y + margin
    case AnchorLeft:
        x = parent.X + margin
        y = parent.Y + (parent.Height-h)/2
    case AnchorCenter:
        x = parent.X + (parent.Width-w)/2
        y = parent.Y + (parent.Height-h)/2
    case AnchorRight:
        x = parent.X + parent.Width - w - margin
        y = parent.Y + (parent.Height-h)/2
    case AnchorBottomLeft:
        x = parent.X + margin
        y = parent.Y + parent.Height - h - margin
    case AnchorBottom:
        x = parent.X + (parent.Width-w)/2
        y = parent.Y + parent.Height - h - margin
    case AnchorBottomRight:
        x = parent.X + parent.Width - w - margin
        y = parent.Y + parent.Height - h - margin
    }

    return Rect{x, y, w, h}
}

func resolveSize(s Size, parent Rect) (float64, float64) {
    switch s.Mode {
    case SizeModeFixed:
        return s.Width, s.Height
    case SizeModeRelative:
        return parent.Width * s.Width, parent.Height * s.Height
    case SizeModeFill:
        return parent.Width, parent.Height
    default:
        return s.Width, s.Height
    }
}
```

## Layout Containers (engine/ui/containers.go)

### VBox (Vertical List)

```go
package ui

// VBox arranges children vertically
type VBox struct {
    children []Widget
    spacing  float64
    anchor   Anchor
    size     Size
    padding  float64
}

func NewVBox(spacing, padding float64) *VBox {
    return &VBox{
        spacing: spacing,
        padding: padding,
        anchor:  AnchorTopLeft,
        size:    Size{Mode: SizeModeFit},
    }
}

func (v *VBox) Add(w Widget) {
    v.children = append(v.children, w)
}

func (v *VBox) Layout(parent Rect) Rect {
    bounds := ApplyAnchor(parent, v.size, v.anchor, 0)

    // Compute total height needed
    y := bounds.Y + v.padding
    maxWidth := 0.0

    for _, child := range v.children {
        childBounds := child.Layout(Rect{
            bounds.X + v.padding,
            y,
            bounds.Width - v.padding*2,
            0, // Height computed by child
        })

        y += childBounds.Height + v.spacing
        if childBounds.Width > maxWidth {
            maxWidth = childBounds.Width
        }
    }

    // Update bounds if FitContent
    if v.size.Mode == SizeModeFit {
        bounds.Width = maxWidth + v.padding*2
        bounds.Height = y - bounds.Y + v.padding
    }

    return bounds
}

func (v *VBox) Children() []Widget { return v.children }
func (v *VBox) Draw(screen *ebiten.Image, bounds Rect) {
    // Optional: draw background
}
func (v *VBox) HandleInput(x, y float64, clicked bool) bool { return false }
```

### HBox (Horizontal List)

```go
// HBox arranges children horizontally
type HBox struct {
    children []Widget
    spacing  float64
    anchor   Anchor
    size     Size
    padding  float64
}

func NewHBox(spacing, padding float64) *HBox {
    return &HBox{
        spacing: spacing,
        padding: padding,
        anchor:  AnchorTopLeft,
        size:    Size{Mode: SizeModeFit},
    }
}

func (h *HBox) Layout(parent Rect) Rect {
    bounds := ApplyAnchor(parent, h.size, h.anchor, 0)

    x := bounds.X + h.padding
    maxHeight := 0.0

    for _, child := range h.children {
        childBounds := child.Layout(Rect{
            x,
            bounds.Y + h.padding,
            0,
            bounds.Height - h.padding*2,
        })

        x += childBounds.Width + h.spacing
        if childBounds.Height > maxHeight {
            maxHeight = childBounds.Height
        }
    }

    if h.size.Mode == SizeModeFit {
        bounds.Width = x - bounds.X + h.padding
        bounds.Height = maxHeight + h.padding*2
    }

    return bounds
}

func (h *HBox) Children() []Widget { return h.children }
func (h *HBox) Draw(screen *ebiten.Image, bounds Rect) {}
func (h *HBox) HandleInput(x, y float64, clicked bool) bool { return false }
```

### Panel (Positioned Container)

```go
// Panel is a positioned container with optional background
type Panel struct {
    child    Widget
    anchor   Anchor
    size     Size
    margin   float64
    padding  float64
    bgColor  color.RGBA
    bounds   Rect // Computed during layout
}

func NewPanel(anchor Anchor, size Size) *Panel {
    return &Panel{
        anchor:  anchor,
        size:    size,
        bgColor: color.RGBA{0, 0, 0, 128},
    }
}

func (p *Panel) SetChild(w Widget) {
    p.child = w
}

func (p *Panel) Layout(parent Rect) Rect {
    p.bounds = ApplyAnchor(parent, p.size, p.anchor, p.margin)
    return p.bounds
}

func (p *Panel) Children() []Widget {
    if p.child == nil {
        return nil
    }
    return []Widget{p.child}
}

func (p *Panel) Draw(screen *ebiten.Image, bounds Rect) {
    // Draw background
    ebitenutil.DrawRect(screen, bounds.X, bounds.Y, bounds.Width, bounds.Height, p.bgColor)
}

func (p *Panel) HandleInput(x, y float64, clicked bool) bool {
    return x >= p.bounds.X && x < p.bounds.X+p.bounds.Width &&
           y >= p.bounds.Y && y < p.bounds.Y+p.bounds.Height
}
```

## Basic Widgets (engine/ui/widgets.go)

### Label

```go
// Label displays text
type Label struct {
    text     string
    font     font.Face
    color    color.RGBA
    anchor   Anchor
    bounds   Rect
}

func NewLabel(text string, f font.Face) *Label {
    return &Label{
        text:   text,
        font:   f,
        color:  color.RGBA{255, 255, 255, 255},
        anchor: AnchorTopLeft,
    }
}

func (l *Label) Layout(parent Rect) Rect {
    // Measure text
    w := font.MeasureString(l.font, l.text).Ceil()
    h := l.font.Metrics().Height.Ceil()

    l.bounds = ApplyAnchor(parent, Size{SizeModeFixed, float64(w), float64(h)}, l.anchor, 0)
    return l.bounds
}

func (l *Label) Draw(screen *ebiten.Image, bounds Rect) {
    text.Draw(screen, l.text, l.font, int(bounds.X), int(bounds.Y)+l.font.Metrics().Ascent.Ceil(), l.color)
}

func (l *Label) Children() []Widget { return nil }
func (l *Label) HandleInput(x, y float64, clicked bool) bool { return false }
```

### Button

```go
// Button is a clickable label
type Button struct {
    id       int
    text     string
    font     font.Face
    callback func(id int)

    // Style
    bgColor      color.RGBA
    hoverColor   color.RGBA
    textColor    color.RGBA
    padding      float64

    // State
    bounds  Rect
    hovered bool
    pressed bool
}

func NewButton(id int, text string, f font.Face, callback func(int)) *Button {
    return &Button{
        id:         id,
        text:       text,
        font:       f,
        callback:   callback,
        bgColor:    color.RGBA{60, 60, 80, 255},
        hoverColor: color.RGBA{80, 80, 120, 255},
        textColor:  color.RGBA{255, 255, 255, 255},
        padding:    8,
    }
}

func (b *Button) Layout(parent Rect) Rect {
    w := font.MeasureString(b.font, b.text).Ceil()
    h := b.font.Metrics().Height.Ceil()

    b.bounds = Rect{
        parent.X,
        parent.Y,
        float64(w) + b.padding*2,
        float64(h) + b.padding*2,
    }
    return b.bounds
}

func (b *Button) Draw(screen *ebiten.Image, bounds Rect) {
    bg := b.bgColor
    if b.hovered {
        bg = b.hoverColor
    }

    ebitenutil.DrawRect(screen, bounds.X, bounds.Y, bounds.Width, bounds.Height, bg)

    textX := int(bounds.X + b.padding)
    textY := int(bounds.Y + b.padding) + b.font.Metrics().Ascent.Ceil()
    text.Draw(screen, b.text, b.font, textX, textY, b.textColor)
}

func (b *Button) HandleInput(x, y float64, clicked bool) bool {
    inside := x >= b.bounds.X && x < b.bounds.X+b.bounds.Width &&
              y >= b.bounds.Y && y < b.bounds.Y+b.bounds.Height

    b.hovered = inside

    if inside && clicked && b.callback != nil {
        b.callback(b.id)
        return true // Consumed
    }
    return false
}

func (b *Button) Children() []Widget { return nil }
```

## AILANG to Widget Translation (engine/ui/translate.go)

```go
package ui

import "stapledons_voyage/sim_gen"

// Translator converts AILANG UIElement to engine Widget
type Translator struct {
    fonts    map[string]font.Face
    handlers map[int]func(int) // Button callbacks
}

func NewTranslator(fonts map[string]font.Face) *Translator {
    return &Translator{
        fonts:    fonts,
        handlers: make(map[int]func(int)),
    }
}

// RegisterHandler registers a callback for button ID
func (t *Translator) RegisterHandler(id int, handler func(int)) {
    t.handlers[id] = handler
}

// Translate converts AILANG UIElement tree to Widget tree
func (t *Translator) Translate(elem sim_gen.UIElement) Widget {
    switch elem.Kind {
    case sim_gen.UIElementKindLabel:
        return t.translateLabel(elem.Label)
    case sim_gen.UIElementKindButton:
        return t.translateButton(elem.Button)
    case sim_gen.UIElementKindPanel:
        return t.translatePanel(elem.Panel)
    case sim_gen.UIElementKindSpacer:
        return t.translateSpacer(elem.Spacer)
    }
    return nil
}

func (t *Translator) translateLabel(l sim_gen.UILabel) *Label {
    label := NewLabel(l.Text, t.fonts["default"])
    label.anchor = translateAnchor(l.Style.Anchor)
    return label
}

func (t *Translator) translateButton(b sim_gen.UIButton) *Button {
    callback := t.handlers[b.ID]
    btn := NewButton(b.ID, b.Text, t.fonts["default"], callback)
    return btn
}

func (t *Translator) translatePanel(p sim_gen.UIPanel) *Panel {
    panel := NewPanel(translateAnchor(p.Style.Anchor), translateSize(p.Style.Size))

    // Translate layout container
    switch p.Layout.Kind {
    case sim_gen.UILayoutKindVertical:
        vbox := NewVBox(4, 8)
        for _, child := range p.Children {
            vbox.Add(t.Translate(child))
        }
        panel.SetChild(vbox)
    case sim_gen.UILayoutKindHorizontal:
        hbox := NewHBox(4, 8)
        for _, child := range p.Children {
            hbox.Add(t.Translate(child))
        }
        panel.SetChild(hbox)
    }

    return panel
}

func translateAnchor(a sim_gen.UIAnchor) Anchor {
    // Map AILANG anchor to engine anchor
    switch a.Kind {
    case sim_gen.UIAnchorKindTopLeft:
        return AnchorTopLeft
    case sim_gen.UIAnchorKindCenter:
        return AnchorCenter
    // ... etc
    }
    return AnchorTopLeft
}

func translateSize(s sim_gen.UISize) Size {
    switch s.Kind {
    case sim_gen.UISizeKindFixed:
        return Size{SizeModeFixed, s.Fixed.W, s.Fixed.H}
    case sim_gen.UISizeKindRelative:
        return Size{SizeModeRelative, s.Relative.W, s.Relative.H}
    case sim_gen.UISizeKindFill:
        return Size{SizeModeFill, 0, 0}
    }
    return Size{SizeModeFit, 0, 0}
}
```

## Game Loop Integration (cmd/game/main.go)

```go
type Game struct {
    // ...existing fields...
    ui *ui.LayoutEngine
    translator *ui.Translator
}

func (g *Game) Update() error {
    input := render.CaptureInput()
    w2, out, err := sim_gen.Step(g.world, input)
    g.world = w2
    g.out = out

    // Update UI from AILANG output
    if len(out.UI) > 0 {
        root := g.translator.Translate(out.UI[0])
        g.ui.SetRoot(root)
    }

    // Process UI input
    x, y := ebiten.CursorPosition()
    clicked := inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)
    g.ui.HandleInput(float64(x), float64(y), clicked)

    return err
}

func (g *Game) Draw(screen *ebiten.Image) {
    render.RenderFrame(screen, g.out)
    g.ui.Draw(screen)  // UI drawn last (on top)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
    g.ui.OnResize(outsideWidth, outsideHeight)
    return outsideWidth, outsideHeight
}
```

## Resolution Scaling

```go
// ScaleManager handles resolution-independent sizing
type ScaleManager struct {
    baseWidth, baseHeight float64 // Design resolution (e.g., 1920x1080)
    actualW, actualH      float64
    scale                 float64
}

func NewScaleManager(baseW, baseH int) *ScaleManager {
    return &ScaleManager{
        baseWidth:  float64(baseW),
        baseHeight: float64(baseH),
    }
}

func (s *ScaleManager) OnResize(w, h int) {
    s.actualW = float64(w)
    s.actualH = float64(h)

    // Compute scale to fit base resolution in window
    scaleX := s.actualW / s.baseWidth
    scaleY := s.actualH / s.baseHeight
    s.scale = min(scaleX, scaleY)
}

// ToScreen converts base-resolution coords to actual screen coords
func (s *ScaleManager) ToScreen(x, y float64) (float64, float64) {
    offsetX := (s.actualW - s.baseWidth*s.scale) / 2
    offsetY := (s.actualH - s.baseHeight*s.scale) / 2
    return x*s.scale + offsetX, y*s.scale + offsetY
}

// FromScreen converts screen coords to base-resolution coords
func (s *ScaleManager) FromScreen(sx, sy float64) (float64, float64) {
    offsetX := (s.actualW - s.baseWidth*s.scale) / 2
    offsetY := (s.actualH - s.baseHeight*s.scale) / 2
    return (sx - offsetX) / s.scale, (sy - offsetY) / s.scale
}
```

## Implementation Plan

### Files to Create

| File | Purpose |
|------|---------|
| `engine/ui/types.go` | Core types (Anchor, Size, Widget interface) |
| `engine/ui/layout.go` | LayoutEngine and layout algorithm |
| `engine/ui/anchor.go` | Anchor positioning logic |
| `engine/ui/containers.go` | VBox, HBox, Panel, Grid |
| `engine/ui/widgets.go` | Label, Button, Image, ProgressBar |
| `engine/ui/translate.go` | AILANG → Widget translation |
| `engine/ui/scale.go` | Resolution scaling |

### AILANG Changes

| File | Change |
|------|--------|
| `sim/protocol.ail` | Add UIElement, UILayout, UIStyle types |
| `sim/protocol.ail` | Add ui field to FrameOutput |

### Go Integration

| File | Change |
|------|--------|
| `cmd/game/main.go` | Add LayoutEngine, call in Update/Draw/Layout |
| `engine/render/draw.go` | Optionally integrate with existing DrawCmd |

## Testing Strategy

### Unit Tests

```go
func TestAnchorPositioning(t *testing.T)
func TestVBoxLayout(t *testing.T)
func TestResolutionScaling(t *testing.T)
func TestButtonClick(t *testing.T)
```

### Visual Tests

```bash
make run-mock
# Resize window, verify UI adapts
# Click buttons, verify callbacks fire
# Test at 720p, 1080p, 4K resolutions
```

## Success Criteria

### Core Layout
- [ ] Widgets position correctly with all 9 anchors
- [ ] VBox arranges children vertically with spacing
- [ ] HBox arranges children horizontally with spacing
- [ ] Panel clips children to bounds

### Responsiveness
- [ ] UI scales correctly with window resize
- [ ] Works at 720p, 1080p, 1440p, 4K
- [ ] No pixel overlap or gaps at any resolution

### AILANG Integration
- [ ] UIElement tree translates to Widget tree
- [ ] Button clicks propagate to AILANG as input events
- [ ] UI updates each frame from FrameOutput

### Performance
- [ ] Layout computation < 1ms for 100 widgets
- [ ] Draw calls batched where possible
- [ ] No allocations per frame (reuse widget tree)

---

**Created:** 2025-12-04
