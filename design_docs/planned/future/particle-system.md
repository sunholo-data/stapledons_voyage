# Particle System

**Version:** 0.5.0
**Status:** Planned
**Priority:** P2 (Polish)
**Complexity:** Medium
**Dependencies:** None
**AILANG Impact:** None - AILANG spawns particles by ID, engine handles physics/rendering

## Problem Statement

**Current State:**
- No visual effects for events
- Actions feel flat (no feedback)
- No environmental effects

**What's Needed:**
- Spawn particle effects at positions
- Engine manages particle lifetime and physics
- Render particles with proper sorting

**AILANG Interface:**
```
DrawCmdParticle{EffectID: 1, X: 5.5, Y: 3.2, Seed: 12345}
                ↑                              ↑
        Engine looks up effect definition    For deterministic randomness
```

## Design

### Particle Effect Types

| EffectID | Name | Use Case | Particle Count |
|----------|------|----------|----------------|
| 1 | `Dust` | NPC movement | 3-5 |
| 2 | `Spark` | Building/crafting | 8-12 |
| 3 | `Smoke` | Fire, engines | 10-20 |
| 4 | `Star` | Selection, highlight | 5-8 |
| 5 | `Rain` | Weather | 50-100 |
| 6 | `Snow` | Weather | 30-50 |

### Effect Definition

**In manifest or code:**
```go
type ParticleEffectDef struct {
    ID           int
    Name         string
    ParticleCount Range        // min-max particles to spawn
    Lifetime     Range         // seconds each particle lives
    Velocity     VelocityDef   // initial velocity pattern
    Gravity      float64       // pixels/sec^2 downward
    Fade         bool          // alpha fades over lifetime
    Shrink       bool          // scale shrinks over lifetime
    Color        ColorDef      // color or color range
    SpriteID     int           // 0 = use colored circle
}

type Range struct {
    Min, Max float64
}

type VelocityDef struct {
    Pattern string  // "radial", "upward", "directional"
    Speed   Range   // pixels per second
    Angle   Range   // degrees (for directional)
}

type ColorDef struct {
    Start color.RGBA
    End   color.RGBA  // interpolate over lifetime
}
```

### Particle Manager

```go
package particle

type Particle struct {
    X, Y       float64
    VelX, VelY float64
    Life       float64  // remaining life in seconds
    MaxLife    float64  // initial life (for fade calculation)
    Scale      float64
    Color      color.RGBA
    SpriteID   int
}

type Emitter struct {
    EffectID   int
    X, Y       float64
    Seed       int64
    Particles  []Particle
    Active     bool
}

type Manager struct {
    emitters  []*Emitter
    defs      map[int]ParticleEffectDef
    rng       *rand.Rand  // seeded per-emitter for determinism
}

func NewManager() *Manager
func (m *Manager) Spawn(effectID int, x, y float64, seed int64)
func (m *Manager) Update(dt float64)
func (m *Manager) Draw(screen *ebiten.Image, cam sim_gen.Camera, screenW, screenH int)
func (m *Manager) Clear()
```

### Particle Physics

```go
func (m *Manager) Update(dt float64) {
    for _, e := range m.emitters {
        if !e.Active {
            continue
        }

        def := m.defs[e.EffectID]
        allDead := true

        for i := range e.Particles {
            p := &e.Particles[i]
            if p.Life <= 0 {
                continue
            }
            allDead = false

            // Apply velocity
            p.X += p.VelX * dt
            p.Y += p.VelY * dt

            // Apply gravity
            p.VelY += def.Gravity * dt

            // Reduce lifetime
            p.Life -= dt

            // Apply fade/shrink
            lifeRatio := p.Life / p.MaxLife
            if def.Fade {
                p.Color.A = uint8(255 * lifeRatio)
            }
            if def.Shrink {
                p.Scale = lifeRatio
            }
        }

        // Deactivate emitter when all particles dead
        if allDead {
            e.Active = false
        }
    }

    // Remove inactive emitters
    m.cleanup()
}
```

### Rendering

```go
func (m *Manager) Draw(screen *ebiten.Image, cam sim_gen.Camera, screenW, screenH int) {
    for _, e := range m.emitters {
        if !e.Active {
            continue
        }

        for _, p := range e.Particles {
            if p.Life <= 0 {
                continue
            }

            // Convert world to screen coordinates
            sx := (p.X-cam.X)*cam.Zoom + float64(screenW)/2
            sy := (p.Y-cam.Y)*cam.Zoom + float64(screenH)/2

            // Skip if off screen
            if sx < -10 || sx > float64(screenW)+10 ||
               sy < -10 || sy > float64(screenH)+10 {
                continue
            }

            // Draw particle
            size := 4.0 * p.Scale * cam.Zoom
            if p.SpriteID > 0 {
                // Draw sprite
            } else {
                // Draw colored circle
                ebitenutil.DrawRect(screen, sx-size/2, sy-size/2, size, size, p.Color)
            }
        }
    }
}
```

### Integration with Draw Commands

**New DrawCmd type:**
```go
// In sim_gen/draw_cmd.go
type DrawCmdParticle struct {
    EffectID int
    X, Y     float64  // world position
    Height   int      // for isometric depth sorting
    Seed     int64    // deterministic randomness
}
func (DrawCmdParticle) isDrawCmd() {}
```

**In renderer:**
```go
func (r *Renderer) RenderFrame(screen *ebiten.Image, out sim_gen.FrameOutput) {
    // ... existing rendering ...

    // Process particle commands - spawn new emitters
    for _, cmd := range out.Draw {
        if p, ok := cmd.(sim_gen.DrawCmdParticle); ok {
            r.particles.Spawn(p.EffectID, p.X, p.Y, p.Seed)
        }
    }

    // Update and draw particles (after tiles, before UI)
    r.particles.Update(1.0/60.0)
    r.particles.Draw(screen, out.Camera, screenW, screenH)
}
```

### Determinism

**Seeded RNG per emitter:**
```go
func (m *Manager) Spawn(effectID int, x, y float64, seed int64) {
    def := m.defs[effectID]
    rng := rand.New(rand.NewSource(seed))

    count := def.ParticleCount.Min + rng.Float64()*(def.ParticleCount.Max-def.ParticleCount.Min)

    particles := make([]Particle, int(count))
    for i := range particles {
        particles[i] = m.createParticle(def, rng)
    }

    // ... add emitter
}
```

This ensures same seed + same effect = same visual result (important for replays/testing).

## Implementation Plan

### Files to Create
| File | Purpose |
|------|---------|
| `engine/particle/manager.go` | Particle system core |
| `engine/particle/effects.go` | Effect definitions |

### Files to Modify
| File | Change |
|------|--------|
| `sim_gen/draw_cmd.go` | Add DrawCmdParticle type |
| `engine/render/draw.go` | Integrate particle manager |

## Testing Strategy

### Visual Test
```bash
make run-mock
# Trigger particle effects via debug keys
# F1 = dust, F2 = spark, F3 = smoke, etc.
```

### Unit Tests
```go
func TestParticleLifetime(t *testing.T)
func TestDeterministicSpawn(t *testing.T)
func TestParticleCleanup(t *testing.T)
```

## Success Criteria

- [ ] Particles spawn at correct position
- [ ] Particles follow physics (velocity, gravity)
- [ ] Particles fade/shrink over lifetime
- [ ] Same seed produces identical effects
- [ ] Performance: 1000 particles at 60fps
- [ ] Particles cull when off-screen
- [ ] Multiple effect types work

---

**Created:** 2025-12-01
