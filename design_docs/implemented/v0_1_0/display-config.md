# Display Configuration

**Version:** 0.2.0
**Status:** Implemented
**Priority:** P1 (Medium)
**Complexity:** Simple
**Package:** `engine/display`

## Related Documents

- [Engine Layer Design](../../implemented/v0_1_0/engine-layer.md) - Current fixed 640x480

## Problem Statement

The game is locked to 640x480 resolution with no configuration options. Players need display flexibility.

**Current State:**
- Fixed 640x480 resolution hardcoded in `Layout()`
- No fullscreen support
- No configuration file

**What's Needed:**
- Configurable window resolution
- Fullscreen toggle (F11 or menu)
- Persist settings between sessions
- Maintain aspect ratio when scaling

## Design

### Configuration File

**Location:** `~/.config/stapledons_voyage/config.json` (or `config.json` in game directory)

```json
{
  "display": {
    "width": 1280,
    "height": 720,
    "fullscreen": false,
    "vsync": true,
    "scale": 1.0
  }
}
```

### Resolution Options

| Preset | Resolution | Aspect Ratio |
|--------|------------|--------------|
| Small | 640x480 | 4:3 |
| Medium | 1280x720 | 16:9 |
| Large | 1920x1080 | 16:9 |
| Native | Monitor resolution | Varies |

### Internal vs Display Resolution

```
Internal resolution: 640x480 (game logic, AILANG coordinates)
Display resolution: 1280x720 (window size, scaled rendering)
Scale factor: 2.0x
```

The game always runs at internal resolution; display is scaled.

## Go Implementation

### Display Manager

```go
package display

type Config struct {
    Width      int     `json:"width"`
    Height     int     `json:"height"`
    Fullscreen bool    `json:"fullscreen"`
    VSync      bool    `json:"vsync"`
    Scale      float64 `json:"scale"`
}

type Manager struct {
    config   Config
    internal struct{ W, H int }  // 640x480
}

func NewManager(configPath string) (*Manager, error)
func (m *Manager) Layout(outsideWidth, outsideHeight int) (int, int)
func (m *Manager) ToggleFullscreen()
func (m *Manager) SetResolution(w, h int)
func (m *Manager) Save() error
```

### Layout Implementation

```go
func (m *Manager) Layout(outsideWidth, outsideHeight int) (int, int) {
    // Always return internal resolution
    // Ebiten handles scaling to window size
    return m.internal.W, m.internal.H
}
```

### Fullscreen Toggle

```go
func (m *Manager) ToggleFullscreen() {
    m.config.Fullscreen = !m.config.Fullscreen
    ebiten.SetFullscreen(m.config.Fullscreen)
    m.Save()
}
```

### Input Handling

```go
// In Update()
if inpututil.IsKeyJustPressed(ebiten.KeyF11) {
    g.display.ToggleFullscreen()
}
```

## Aspect Ratio Handling

### Letterboxing/Pillarboxing

When window aspect ratio differs from internal:

```go
func (m *Manager) CalculateViewport(windowW, windowH int) image.Rectangle {
    internalAspect := float64(m.internal.W) / float64(m.internal.H)
    windowAspect := float64(windowW) / float64(windowH)

    if windowAspect > internalAspect {
        // Pillarbox (black bars on sides)
        scale := float64(windowH) / float64(m.internal.H)
        newW := int(float64(m.internal.W) * scale)
        offsetX := (windowW - newW) / 2
        return image.Rect(offsetX, 0, offsetX+newW, windowH)
    } else {
        // Letterbox (black bars top/bottom)
        scale := float64(windowW) / float64(m.internal.W)
        newH := int(float64(m.internal.H) * scale)
        offsetY := (windowH - newH) / 2
        return image.Rect(0, offsetY, windowW, offsetY+newH)
    }
}
```

## Implementation Plan

### Files to Create

| File | Purpose |
|------|---------|
| `engine/display/config.go` | Config struct and loading |
| `engine/display/manager.go` | Display management |

### Changes to Existing Files

| File | Change |
|------|--------|
| `cmd/game/main.go` | Use display.Manager for Layout |
| `cmd/game/main.go` | Handle F11 for fullscreen |
| `engine/render/input.go` | Expose display toggle input |

## Testing Strategy

### Manual Testing

```bash
make run
# Press F11 → should toggle fullscreen
# Edit config.json → restart → should use new resolution
```

### Automated Testing

```go
func TestConfigLoad(t *testing.T)
func TestConfigSave(t *testing.T)
func TestAspectRatioCalculation(t *testing.T)
```

### Edge Cases

- [ ] Config file missing → use defaults
- [ ] Invalid resolution → clamp to valid range
- [ ] Monitor doesn't support resolution → fall back
- [ ] Config file corrupted → reset to defaults

## Success Criteria

### Configuration
- [ ] Config file loads at startup
- [ ] Config file saves on changes
- [ ] Default config created if missing

### Display
- [ ] Window resizes to configured resolution
- [ ] Fullscreen toggles with F11
- [ ] Aspect ratio maintained (letterbox/pillarbox)
- [ ] VSync setting respected

### Persistence
- [ ] Settings persist between sessions
- [ ] Fullscreen state remembered
- [ ] Resolution changes saved
