# Camera and Viewport System

**Version:** 0.3.0
**Status:** Planned
**Priority:** P0 (High)
**Complexity:** Medium
**Package:** `engine/camera`

## Related Documents

- [Display Configuration](../v0_2_0/display-config.md) - Resolution handling
- [Tilemap Rendering](tilemap-rendering.md) - Efficient rendering of visible tiles
- [Architecture Overview](../../implemented/v0_1_0/architecture.md) - System context

## Problem Statement

The game renders the entire world at once. For larger worlds (64x64+), we need a camera that follows the player and only renders visible content.

**Current State:**
- All DrawCmds rendered regardless of position
- No scrolling or panning
- World size limited by screen size

**What's Needed:**
- Camera position tracks player or target
- Only visible content rendered (culling)
- Smooth camera movement
- Optional zoom support

## Design

### Camera in AILANG vs Go

**Option A: Camera in AILANG** (recommended)
- AILANG step() returns camera position in FrameOutput
- Engine applies camera transform to all DrawCmds
- Game logic controls camera behavior

**Option B: Camera in Go**
- Engine maintains camera state
- Engine decides camera position based on World
- Breaks "thin engine" principle

**Decision:** Option A - Camera position comes from AILANG.

### AILANG Integration

#### Protocol Extension (sim/protocol.ail)

```ailang
type Camera = {
    x: float,           -- World position (center of view)
    y: float,
    zoom: float         -- 1.0 = normal, 2.0 = zoomed in
}

type FrameOutput = {
    draw_cmds: [DrawCmd],
    sounds: [SoundCmd],
    camera: Camera      -- NEW: camera state
}
```

#### Camera Logic (sim/camera.ail)

```ailang
-- Follow player with some smoothing
export pure func update_camera(cam: Camera, target_x: float, target_y: float) -> Camera {
    let lerp_factor = 0.1;
    {
        x: cam.x + (target_x - cam.x) * lerp_factor,
        y: cam.y + (target_y - cam.y) * lerp_factor,
        zoom: cam.zoom
    }
}
```

### Go Implementation

#### Camera Transform

```go
package camera

type Transform struct {
    OffsetX, OffsetY float64  // Screen offset for world origin
    Scale            float64  // Zoom factor
}

func FromOutput(cam sim_gen.Camera, screenW, screenH int) Transform {
    return Transform{
        OffsetX: float64(screenW)/2 - cam.X*cam.Zoom,
        OffsetY: float64(screenH)/2 - cam.Y*cam.Zoom,
        Scale:   cam.Zoom,
    }
}

func (t Transform) WorldToScreen(worldX, worldY float64) (screenX, screenY float64) {
    return worldX*t.Scale + t.OffsetX, worldY*t.Scale + t.OffsetY
}

func (t Transform) ScreenToWorld(screenX, screenY float64) (worldX, worldY float64) {
    return (screenX - t.OffsetX) / t.Scale, (screenY - t.OffsetY) / t.Scale
}
```

#### Rendering with Camera

```go
func RenderFrame(screen *ebiten.Image, out sim_gen.FrameOutput) {
    cam := camera.FromOutput(out.Camera, screenW, screenH)

    for _, cmd := range out.DrawCmds {
        switch cmd.Kind {
        case sim_gen.DrawCmdKindSprite:
            sx, sy := cam.WorldToScreen(cmd.Sprite.X, cmd.Sprite.Y)
            // Cull if off-screen
            if sx < -32 || sx > screenW+32 || sy < -32 || sy > screenH+32 {
                continue
            }
            opts := &ebiten.DrawImageOptions{}
            opts.GeoM.Scale(cam.Scale, cam.Scale)
            opts.GeoM.Translate(sx, sy)
            screen.DrawImage(sprite, opts)
        }
    }
}
```

### Viewport Culling

Only render DrawCmds within the visible viewport:

```go
type Viewport struct {
    MinX, MinY, MaxX, MaxY float64  // World coordinates
}

func (v Viewport) Contains(x, y, margin float64) bool {
    return x >= v.MinX-margin && x <= v.MaxX+margin &&
           y >= v.MinY-margin && y <= v.MaxY+margin
}

func CalculateViewport(cam sim_gen.Camera, screenW, screenH int) Viewport {
    halfW := float64(screenW) / 2 / cam.Zoom
    halfH := float64(screenH) / 2 / cam.Zoom
    return Viewport{
        MinX: cam.X - halfW,
        MaxX: cam.X + halfW,
        MinY: cam.Y - halfH,
        MaxY: cam.Y + halfH,
    }
}
```

## Camera Modes

| Mode | Description | Use Case |
|------|-------------|----------|
| Follow | Track player position | Normal gameplay |
| Fixed | Static position | Cutscenes, menus |
| Pan | Smooth move to target | Transitions |
| Shake | Offset with noise | Damage, explosions |

### Camera Shake (Optional)

```ailang
type Camera = {
    x: float,
    y: float,
    zoom: float,
    shake_intensity: float,  -- 0.0 = none, 1.0 = max
    shake_decay: float       -- How fast shake fades
}
```

## Implementation Plan

### AILANG Files

| File | Change |
|------|--------|
| `sim/protocol.ail` | Add Camera type |
| `sim/protocol.ail` | Add camera field to FrameOutput |
| `sim/world.ail` | Add camera to World state |
| `sim/camera.ail` | Camera update logic (new file) |

### Go Files

| File | Purpose |
|------|---------|
| `engine/camera/transform.go` | World-to-screen transforms |
| `engine/camera/viewport.go` | Culling calculations |

### Changes to Existing Files

| File | Change |
|------|--------|
| `engine/render/draw.go` | Apply camera transform |
| `engine/render/draw.go` | Cull off-screen DrawCmds |
| `engine/render/input.go` | Convert mouse to world coords |

## Testing Strategy

### Manual Testing

```bash
make run
# Move player → camera should follow
# World extends beyond screen → should scroll
```

### Automated Testing

```go
func TestWorldToScreen(t *testing.T)
func TestScreenToWorld(t *testing.T)
func TestViewportCulling(t *testing.T)
```

### Edge Cases

- [ ] Camera at world edge → clamp or allow overshoot
- [ ] Zoom to 0 → clamp minimum zoom
- [ ] Very fast movement → camera lag acceptable
- [ ] Player off-screen → ensure camera catches up

## Success Criteria

### Camera Transform
- [ ] World coordinates transform to screen correctly
- [ ] Screen coordinates (mouse) transform to world
- [ ] Zoom scales all rendering appropriately

### Following
- [ ] Camera follows player smoothly
- [ ] No jitter or oscillation
- [ ] Lerp factor tunable

### Culling
- [ ] Off-screen DrawCmds not rendered
- [ ] Performance improves for large worlds
- [ ] No visual artifacts at edges

### AILANG Integration
- [ ] Camera type compiles
- [ ] step() updates camera state
- [ ] Engine reads camera from FrameOutput
