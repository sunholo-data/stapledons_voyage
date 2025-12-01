# Animation System

**Version:** 0.5.0
**Status:** Planned
**Priority:** P1 (Engine Foundation)
**Complexity:** Medium
**Dependencies:** Asset Pipeline (sprite loading)
**AILANG Impact:** None - AILANG sends animation IDs, engine handles timing

## Problem Statement

**Current State:**
- Entities render as static sprites
- No frame-based animation support
- NPCs don't animate when moving

**What's Needed:**
- Load sprite sheets with multiple frames
- Advance frames based on elapsed time
- Support different animation states (idle, walk, etc.)

**AILANG Interface:**
```
DrawCmdIsoEntity{SpriteID: 100, AnimState: "walk", AnimFrame: -1}
                                            ↑ -1 means "engine manages frame"
```

## Design

### Animation Metadata

**Extended Manifest Entry:**
```json
{
  "sprites": {
    "100": {
      "file": "npc_walk.png",
      "width": 128,
      "height": 48,
      "type": "entity",
      "animations": {
        "idle": {"startFrame": 0, "frameCount": 1, "fps": 0},
        "walk": {"startFrame": 0, "frameCount": 4, "fps": 8},
        "action": {"startFrame": 4, "frameCount": 2, "fps": 4}
      },
      "frameWidth": 32,
      "frameHeight": 48
    }
  }
}
```

**Sprite Sheet Layout:**
```
┌────┬────┬────┬────┬────┬────┐
│ 0  │ 1  │ 2  │ 3  │ 4  │ 5  │  ← frames
│idle│walk│walk│walk│act │act │
└────┴────┴────┴────┴────┴────┘
  32px each
```

### Animation Runtime

**AnimationState struct:**
```go
type AnimationState struct {
    SpriteID     int
    CurrentAnim  string
    CurrentFrame int
    FrameTime    float64  // seconds since last frame change
    Playing      bool
}
```

**AnimationManager:**
```go
type AnimationManager struct {
    states map[string]*AnimationState  // keyed by entity ID
    defs   map[int]AnimationDef        // keyed by sprite ID
}

func (am *AnimationManager) Update(dt float64)
func (am *AnimationManager) GetFrame(entityID string, spriteID int, animName string) int
func (am *AnimationManager) SetAnimation(entityID string, animName string)
```

### Frame Advancement

```go
func (am *AnimationManager) Update(dt float64) {
    for _, state := range am.states {
        if !state.Playing {
            continue
        }

        def := am.defs[state.SpriteID]
        anim := def.Animations[state.CurrentAnim]

        if anim.FPS == 0 {
            continue // static animation
        }

        state.FrameTime += dt
        frameDuration := 1.0 / float64(anim.FPS)

        if state.FrameTime >= frameDuration {
            state.FrameTime -= frameDuration
            state.CurrentFrame++
            if state.CurrentFrame >= anim.FrameCount {
                state.CurrentFrame = 0 // loop
            }
        }
    }
}
```

### Rendering Integration

```go
func (r *Renderer) drawIsoEntity(screen *ebiten.Image, c sim_gen.DrawCmdIsoEntity, ...) {
    // Get current animation frame
    frame := r.animations.GetFrame(c.ID, c.SpriteID, c.AnimState)

    // Calculate sub-rectangle in sprite sheet
    def := r.assets.GetSpriteDef(c.SpriteID)
    srcX := frame * def.FrameWidth
    srcRect := image.Rect(srcX, 0, srcX+def.FrameWidth, def.FrameHeight)

    // Draw sub-image
    sprite := r.assets.GetSprite(c.SpriteID)
    subImg := sprite.SubImage(srcRect).(*ebiten.Image)
    screen.DrawImage(subImg, op)
}
```

## AILANG Integration

**Option A: Engine-Managed (Recommended)**
```go
// AILANG sends animation name, engine handles frame timing
DrawCmdIsoEntity{
    ID:        "npc-1",
    SpriteID:  100,
    AnimState: "walk",  // engine advances frames
}
```

**Option B: AILANG-Managed**
```go
// AILANG calculates frame based on tick
DrawCmdIsoEntity{
    ID:        "npc-1",
    SpriteID:  100,
    AnimState: "walk",
    AnimFrame: world.Tick / 8 % 4,  // manual frame calc
}
```

**Recommendation:** Option A - Engine manages timing for smooth 60fps animation independent of simulation tick rate.

## Implementation Plan

### Files to Create
| File | Purpose |
|------|---------|
| `engine/render/animation.go` | AnimationManager and state tracking |

### Files to Modify
| File | Change |
|------|--------|
| `engine/assets/sprites.go` | Parse animation metadata from manifest |
| `engine/render/draw.go` | Use animation frames when rendering entities |
| `sim_gen/draw_cmd.go` | Add AnimState field to DrawCmdIsoEntity |

### Manifest Changes
| File | Change |
|------|--------|
| `assets/sprites/manifest.json` | Add animation definitions |

## Testing Strategy

### Visual Test
```bash
make run-mock
# Watch NPC - should animate when moving
```

### Unit Tests
```go
func TestFrameAdvancement(t *testing.T)
func TestAnimationLoop(t *testing.T)
func TestStaticAnimation(t *testing.T)
```

## Success Criteria

- [ ] Sprite sheets load with frame metadata
- [ ] Animations play at correct FPS
- [ ] Animations loop correctly
- [ ] Different entities can have different animation states
- [ ] Static sprites (no animation) still work
- [ ] No visual stuttering at 60fps

---

**Created:** 2025-12-01
