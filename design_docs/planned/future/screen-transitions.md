# Screen Transitions

> **REVIEWED 2025-12-10** - This approach is acceptable.
>
> Screen transitions are purely visual polish. AILANG changes the game mode,
> engine detects the change and applies a visual transition (fade, wipe).
> The transition animation doesn't affect gameplay state.
>
> **Key distinction:** AILANG owns the MODE (what state we're in),
> engine owns the TRANSITION ANIMATION (how we visually move between modes).

**Version:** 0.5.0
**Status:** Planned
**Priority:** P2 (Polish)
**Complexity:** Low
**Dependencies:** None
**AILANG Impact:** Minimal - AILANG changes mode, engine animates transition

## Problem Statement

**Current State:**
- Mode changes are instant (jarring)
- No visual feedback during transitions
- No loading screens

**What's Needed:**
- Smooth fade transitions between modes
- Optional slide/wipe effects
- Loading indicator for slow operations

**AILANG Interface:**
```
// AILANG just changes mode - engine detects change and applies transition
World{mode: ModeGalaxyMap}  →  World{mode: ModeShipExploration}
                            ↑
                    Engine sees mode changed, triggers fade
```

## Design

### Transition Types

| Type | Effect | Duration | Use Case |
|------|--------|----------|----------|
| `None` | Instant | 0ms | Debug, fast switching |
| `Fade` | Fade to black, fade in | 500ms | Default for most transitions |
| `FadeWhite` | Fade to white, fade in | 500ms | Dream sequences, time jumps |
| `SlideLeft` | Old slides left, new slides in | 300ms | Menu navigation |
| `SlideUp` | Old slides up, new slides in | 300ms | Overlay panels |

### Transition Manager

```go
package transition

type Type int

const (
    None Type = iota
    Fade
    FadeWhite
    SlideLeft
    SlideUp
)

type Manager struct {
    active       bool
    transType    Type
    progress     float64  // 0.0 to 1.0
    duration     float64  // seconds
    phase        Phase    // FadingOut, FadingIn
    oldScreen    *ebiten.Image
    onMidpoint   func()   // called at transition midpoint
}

func (m *Manager) Start(t Type, duration float64, onMidpoint func())
func (m *Manager) Update(dt float64) bool  // returns true when complete
func (m *Manager) Draw(screen *ebiten.Image, currentFrame *ebiten.Image)
func (m *Manager) IsActive() bool
```

### Transition Phases

```
Phase 1: FadingOut (0.0 → 0.5)
┌──────────┐     ┌──────────┐     ┌──────────┐
│ Old Mode │  →  │ 50% Dark │  →  │ Black    │
└──────────┘     └──────────┘     └──────────┘

                     ↓ onMidpoint() called - mode actually changes

Phase 2: FadingIn (0.5 → 1.0)
┌──────────┐     ┌──────────┐     ┌──────────┐
│ Black    │  →  │ 50% Dark │  →  │ New Mode │
└──────────┘     └──────────┘     └──────────┘
```

### Fade Implementation

```go
func (m *Manager) Draw(screen *ebiten.Image, currentFrame *ebiten.Image) {
    if !m.active {
        return
    }

    // Calculate alpha based on progress
    var alpha float64
    if m.progress < 0.5 {
        // Fading out: 0 → 1
        alpha = m.progress * 2
    } else {
        // Fading in: 1 → 0
        alpha = (1.0 - m.progress) * 2
    }

    // Draw overlay
    overlay := ebiten.NewImage(screen.Bounds().Dx(), screen.Bounds().Dy())
    if m.transType == FadeWhite {
        overlay.Fill(color.White)
    } else {
        overlay.Fill(color.Black)
    }

    op := &ebiten.DrawImageOptions{}
    op.ColorScale.ScaleAlpha(float32(alpha))
    screen.DrawImage(overlay, op)
}
```

### Mode Change Detection

```go
func (g *Game) Update() error {
    oldMode := g.world.Mode

    // Run simulation
    g.world, g.out, _ = sim_gen.Step(g.world, input)

    // Detect mode change
    if g.world.Mode != oldMode && !g.transition.IsActive() {
        g.transition.Start(transition.Fade, 0.5, func() {
            // Mode already changed in world state
            // This callback can trigger additional setup
        })
    }

    g.transition.Update(1.0/60.0)
    return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
    g.renderer.RenderFrame(screen, g.out)
    g.transition.Draw(screen, nil)
}
```

### Transition Configuration

**Per-mode transition rules:**
```go
var transitionRules = map[ModeTransition]transition.Type{
    {From: ModeShip, To: ModeGalaxy}:     transition.Fade,
    {From: ModeGalaxy, To: ModeShip}:     transition.Fade,
    {From: ModeShip, To: ModeDialogue}:   transition.SlideUp,
    {From: ModeDialogue, To: ModeShip}:   transition.SlideUp,
    {From: Any, To: ModeLegacy}:          transition.FadeWhite,
}
```

## Implementation Plan

### Files to Create
| File | Purpose |
|------|---------|
| `engine/transition/manager.go` | Transition state and rendering |
| `engine/transition/effects.go` | Individual effect implementations |

### Files to Modify
| File | Change |
|------|--------|
| `cmd/game/main.go` | Add transition manager, detect mode changes |

## Testing Strategy

### Visual Test
```bash
make run-mock
# Press M to switch modes - should see fade transition
```

### Unit Tests
```go
func TestFadeProgress(t *testing.T)
func TestMidpointCallback(t *testing.T)
func TestTransitionComplete(t *testing.T)
```

## Success Criteria

- [ ] Mode changes trigger transitions
- [ ] Fade effect renders smoothly
- [ ] Transition completes in expected duration
- [ ] No input accepted during transition
- [ ] Midpoint callback fires correctly
- [ ] Multiple transition types work

---

**Created:** 2025-12-01
