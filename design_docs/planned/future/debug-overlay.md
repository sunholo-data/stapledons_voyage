# Debug Overlay

**Version:** 0.5.0
**Status:** Planned
**Priority:** P2 (Polish)
**Complexity:** Low
**Dependencies:** None
**AILANG Impact:** Minimal - AILANG can add to Debug slice, engine renders

## Problem Statement

**Current State:**
- Debug info rendered via FrameOutput.Debug (basic text)
- No performance metrics visible
- No way to toggle debug display
- No visualization of internal state

**What's Needed:**
- Toggle-able debug overlay (F3)
- FPS counter and frame time graph
- Memory usage display
- Entity count and draw call stats
- Tile grid visualization
- Camera bounds visualization

**AILANG Interface:**
```
FrameOutput{
    Debug: ["NPC count: 5", "Camera: (0, 0)"]  // existing - simple strings
}
// Engine adds its own metrics to overlay
```

## Design

### Overlay Panels

| Panel | Contents | Position |
|-------|----------|----------|
| Performance | FPS, frame time, frame time graph | Top-left |
| Memory | Heap alloc, GC count | Top-left (below perf) |
| Render | Draw calls, entities, culled | Top-right |
| World | Tick, mode, camera position | Bottom-left |
| AILANG Debug | FrameOutput.Debug strings | Bottom-left (below world) |

### Debug Manager

```go
package debug

type Level int

const (
    LevelOff Level = iota
    LevelBasic     // FPS only
    LevelFull      // All panels
)

type Manager struct {
    level         Level
    frameTimeRing []float64  // circular buffer for graph
    ringIndex     int
    lastGCCount   uint32
    visible       bool
}

func NewManager() *Manager
func (m *Manager) Toggle()
func (m *Manager) SetLevel(level Level)
func (m *Manager) RecordFrame(dt float64)
func (m *Manager) Draw(screen *ebiten.Image, stats RenderStats, worldDebug []string)
```

### Render Stats

```go
type RenderStats struct {
    DrawCalls    int
    EntitiesTotal int
    EntitiesCulled int
    TilesTotal    int
    TilesCulled   int
    ParticlesActive int
    CameraX, CameraY float64
    CameraZoom    float64
    Tick          int
    Mode          string
}
```

### Performance Panel

```go
func (m *Manager) drawPerformancePanel(screen *ebiten.Image) {
    // Calculate FPS from frame times
    avgFrameTime := m.averageFrameTime()
    fps := 1.0 / avgFrameTime

    // Format strings
    lines := []string{
        fmt.Sprintf("FPS: %.1f", fps),
        fmt.Sprintf("Frame: %.2fms", avgFrameTime*1000),
    }

    // Draw panel background
    panelW, panelH := 150.0, 80.0
    ebitenutil.DrawRect(screen, 5, 5, panelW, panelH, color.RGBA{0, 0, 0, 180})

    // Draw text
    for i, line := range lines {
        ebitenutil.DebugPrintAt(screen, line, 10, 10+i*16)
    }

    // Draw frame time graph
    m.drawFrameTimeGraph(screen, 10, 45, 130, 30)
}

func (m *Manager) drawFrameTimeGraph(screen *ebiten.Image, x, y, w, h float64) {
    // Draw background
    ebitenutil.DrawRect(screen, x, y, w, h, color.RGBA{40, 40, 40, 255})

    // Draw frame times as bars
    barWidth := w / float64(len(m.frameTimeRing))
    targetFrameTime := 1.0 / 60.0  // 16.67ms

    for i, ft := range m.frameTimeRing {
        // Height proportional to frame time (capped at 2x target)
        ratio := ft / (targetFrameTime * 2)
        if ratio > 1 {
            ratio = 1
        }
        barH := ratio * h

        // Color: green if under target, yellow if close, red if over
        var col color.RGBA
        if ft < targetFrameTime {
            col = color.RGBA{0, 200, 0, 255}
        } else if ft < targetFrameTime*1.5 {
            col = color.RGBA{200, 200, 0, 255}
        } else {
            col = color.RGBA{200, 0, 0, 255}
        }

        bx := x + float64(i)*barWidth
        by := y + h - barH
        ebitenutil.DrawRect(screen, bx, by, barWidth-1, barH, col)
    }

    // Draw target line (16.67ms)
    targetY := y + h - (h * 0.5)  // 50% height = target
    ebitenutil.DrawLine(screen, x, targetY, x+w, targetY, color.RGBA{255, 255, 255, 128})
}
```

### Memory Panel

```go
func (m *Manager) drawMemoryPanel(screen *ebiten.Image, startY float64) {
    var memStats runtime.MemStats
    runtime.ReadMemStats(&memStats)

    lines := []string{
        fmt.Sprintf("Heap: %.1f MB", float64(memStats.HeapAlloc)/1024/1024),
        fmt.Sprintf("GCs: %d", memStats.NumGC),
    }

    panelW, panelH := 150.0, 40.0
    ebitenutil.DrawRect(screen, 5, startY, panelW, panelH, color.RGBA{0, 0, 0, 180})

    for i, line := range lines {
        ebitenutil.DebugPrintAt(screen, line, 10, int(startY)+5+i*16)
    }
}
```

### Render Stats Panel

```go
func (m *Manager) drawRenderPanel(screen *ebiten.Image, stats RenderStats, screenW int) {
    lines := []string{
        fmt.Sprintf("Draw: %d", stats.DrawCalls),
        fmt.Sprintf("Entities: %d/%d", stats.EntitiesTotal-stats.EntitiesCulled, stats.EntitiesTotal),
        fmt.Sprintf("Tiles: %d/%d", stats.TilesTotal-stats.TilesCulled, stats.TilesTotal),
        fmt.Sprintf("Particles: %d", stats.ParticlesActive),
    }

    panelW, panelH := 150.0, 70.0
    x := float64(screenW) - panelW - 5
    ebitenutil.DrawRect(screen, x, 5, panelW, panelH, color.RGBA{0, 0, 0, 180})

    for i, line := range lines {
        ebitenutil.DebugPrintAt(screen, line, int(x)+5, 10+i*16)
    }
}
```

### World Info Panel

```go
func (m *Manager) drawWorldPanel(screen *ebiten.Image, stats RenderStats, screenH int) {
    lines := []string{
        fmt.Sprintf("Tick: %d", stats.Tick),
        fmt.Sprintf("Mode: %s", stats.Mode),
        fmt.Sprintf("Camera: (%.1f, %.1f)", stats.CameraX, stats.CameraY),
        fmt.Sprintf("Zoom: %.2fx", stats.CameraZoom),
    }

    panelW, panelH := 180.0, 70.0
    y := float64(screenH) - panelH - 5
    ebitenutil.DrawRect(screen, 5, y, panelW, panelH, color.RGBA{0, 0, 0, 180})

    for i, line := range lines {
        ebitenutil.DebugPrintAt(screen, line, 10, int(y)+5+i*16)
    }
}
```

### Grid Visualization

```go
func (m *Manager) drawTileGrid(screen *ebiten.Image, cam sim_gen.Camera, screenW, screenH int) {
    if m.level < LevelFull {
        return
    }

    // Draw tile grid lines
    gridColor := color.RGBA{100, 100, 100, 100}

    // Calculate visible tile range
    minTileX, minTileY := render.ScreenToTile(0, 0, cam, screenW, screenH)
    maxTileX, maxTileY := render.ScreenToTile(float64(screenW), float64(screenH), cam, screenW, screenH)

    // Draw grid for visible tiles
    for tx := int(minTileX) - 1; tx <= int(maxTileX) + 1; tx++ {
        for ty := int(minTileY) - 1; ty <= int(maxTileY) + 1; ty++ {
            // Draw tile outline
            drawTileOutline(screen, sim_gen.Coord{X: tx, Y: ty}, 0, cam, screenW, screenH, gridColor)
        }
    }
}
```

### Integration

**In game loop:**
```go
func (g *Game) Draw(screen *ebiten.Image) {
    // Render game
    r.RenderFrame(screen, g.out)

    // Render debug overlay on top
    if g.debug.IsVisible() {
        stats := RenderStats{
            DrawCalls:    len(g.out.Draw),
            Tick:         g.world.Tick,
            CameraX:      g.out.Camera.X,
            CameraY:      g.out.Camera.Y,
            CameraZoom:   g.out.Camera.Zoom,
            // ... other stats
        }
        g.debug.Draw(screen, stats, g.out.Debug)
    }
}

func (g *Game) Update() error {
    // Toggle with F3
    if inpututil.IsKeyJustPressed(ebiten.KeyF3) {
        g.debug.Toggle()
    }

    // Record frame time
    g.debug.RecordFrame(1.0/60.0)  // or actual delta

    // ...
}
```

## Implementation Plan

### Files to Create
| File | Purpose |
|------|---------|
| `engine/debug/overlay.go` | Debug manager and rendering |
| `engine/debug/stats.go` | RenderStats collection |

### Files to Modify
| File | Change |
|------|--------|
| `cmd/game/main.go` | Initialize debug manager, toggle on F3 |
| `engine/render/draw.go` | Collect render stats during frame |

## Testing Strategy

### Manual Test
```bash
make run-mock
# Press F3 to toggle overlay
# Verify FPS counter accurate
# Verify frame time graph updates
# Verify stats match expectations
```

### Unit Tests
```go
func TestFrameTimeRing(t *testing.T)
func TestLevelToggle(t *testing.T)
```

## Success Criteria

- [ ] F3 toggles debug overlay
- [ ] FPS counter accurate within 5%
- [ ] Frame time graph shows spikes
- [ ] Memory usage displays
- [ ] Render stats accurate
- [ ] Camera position displays
- [ ] AILANG debug strings display
- [ ] Overlay doesn't affect performance significantly

---

**Created:** 2025-12-01
