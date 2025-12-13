# Celestial Level of Detail (LOD) System

## Status
- **Status**: Planned
- **Priority**: P2 (performance optimization)
- **Estimated**: 2-3 days
- **Location**: `engine/lod/`, `engine/tetra/`

## Game Vision Alignment

| Pillar | Alignment | Notes |
|--------|-----------|-------|
| Choices Are Final | N/A | Infrastructure feature |
| The Game Doesn't Judge | N/A | Infrastructure feature |
| Time Has Emotional Weight | N/A | Infrastructure feature |
| The Ship Is Home | N/A | Infrastructure feature |
| Grounded Strangeness | **Supports** | Enables rendering thousands of alien worlds |
| We Are Not Built For This | N/A | Infrastructure feature |
| **Overall** | Enabler | Essential for galactic-scale rendering |

**This is an engine infrastructure feature** - it enables the game to render galaxy views with thousands of celestial objects while maintaining playable framerates.

## Problem Statement

Currently, all celestial objects are rendered the same way regardless of distance:
- A planet 1000 units away uses the same 3D mesh as one 10 units away
- Tetra3D can handle ~20-50 3D objects before performance degrades
- Galaxy views need thousands of stars/planets visible simultaneously
- No frustum culling - objects behind camera still processed

**Target scenario:**
- Galaxy map: ~100,000 stars visible
- Star system: ~20 planets + ~100 moons + ~1000 asteroids
- Close orbit: ~5 detailed 3D objects + many distant points

## Proposed Solution

### LOD Tiers

| Tier | Distance | Representation | Cost | Max Objects |
|------|----------|----------------|------|-------------|
| **Full3D** | < 50 units | Tetra3D mesh (sphere, rings) | High | ~20-50 |
| **Billboard** | 50-200 units | 2D sprite facing camera | Medium | ~200 |
| **Circle** | 200-1000 units | Filled circle (DrawCmd.Circle) | Low | ~2,000 |
| **Point** | 1000-10000 units | Single pixel | Very Low | ~50,000 |
| **Culled** | > 10000 units | Not rendered | Zero | Unlimited |

### Distance Thresholds (Configurable)

```go
type LODConfig struct {
    Full3DDistance   float64 // Below this: full 3D mesh
    BillboardDistance float64 // Below this: billboard sprite
    CircleDistance   float64 // Below this: colored circle
    PointDistance    float64 // Below this: single point
    // Above PointDistance: culled (not rendered)
}

var DefaultLODConfig = LODConfig{
    Full3DDistance:   50,
    BillboardDistance: 200,
    CircleDistance:   1000,
    PointDistance:    10000,
}
```

### Architecture

```
┌─────────────────────────────────────────────────────────────┐
│ CelestialObject (interface)                                 │
│   - Position() Vector3                                      │
│   - Radius() float64                                        │
│   - Color() color.RGBA                                      │
│   - Visual3D() *tetra.Planet (or nil)                       │
│   - Sprite() *ebiten.Image (or nil for billboard)           │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│ LODManager                                                   │
│   - camera: Vector3                                          │
│   - objects: []CelestialObject                               │
│   - config: LODConfig                                        │
│                                                              │
│   Update(cameraPos Vector3)                                  │
│     → Sort objects by distance                               │
│     → Assign LOD tier to each                                │
│     → Activate/deactivate 3D objects                         │
│                                                              │
│   Render3D(scene *tetra.Scene)                               │
│     → Add only Full3D tier objects to scene                  │
│                                                              │
│   Render2D(screen *ebiten.Image)                             │
│     → Draw Billboard/Circle/Point tiers                      │
└─────────────────────────────────────────────────────────────┘
```

## Engine vs AILANG Responsibilities

### Engine (Go) - 95%

| Component | Location | Responsibility |
|-----------|----------|----------------|
| LODManager | `engine/lod/manager.go` | Distance calculation, tier assignment |
| LODConfig | `engine/lod/config.go` | Threshold configuration |
| Frustum Culling | `engine/lod/frustum.go` | Skip objects outside camera view |
| 3D Pool | `engine/lod/pool.go` | Reuse Tetra3D objects (avoid allocation) |
| Point Buffer | `engine/lod/points.go` | Batch render thousands of points |

### AILANG - 5%

AILANG only provides the **data** - positions, colors, sizes of celestial objects:

```ailang
-- AILANG provides object definitions (already exists in celestial.ail)
export type CelestialPlanet = {
    name: string,
    radius: float,        -- Used for apparent size calculation
    color: Color,         -- Used for point/circle color
    orbitRadius: float,   -- Used to calculate position
    -- ... other fields
}

-- AILANG does NOT decide HOW to render (that's engine's job)
-- AILANG does NOT know about LOD tiers
```

**Why minimal AILANG involvement?**
- LOD is purely visual optimization
- Doesn't affect game simulation or logic
- Engine can make real-time decisions based on camera
- AILANG shouldn't know about rendering details

## Implementation Details

### 1. LODManager

```go
// engine/lod/manager.go
type LODManager struct {
    config   LODConfig
    objects  []LODObject
    pool3D   *Object3DPool

    // Per-frame state
    tier3D     []*LODObject  // Full 3D this frame
    tierBillboard []*LODObject
    tierCircle []*LODObject
    tierPoint  []*LODObject
}

type LODObject struct {
    ID       string
    Position Vector3
    Radius   float64
    Color    color.RGBA

    // Optional higher-detail representations
    Model3D  *tetra.Planet  // For Full3D tier
    Sprite   *ebiten.Image  // For Billboard tier

    // Current state
    CurrentTier LODTier
    Distance    float64     // Cached distance to camera
}

func (m *LODManager) Update(cameraPos Vector3) {
    // 1. Calculate distance for each object
    for i := range m.objects {
        m.objects[i].Distance = distance(m.objects[i].Position, cameraPos)
    }

    // 2. Sort by distance (for priority when 3D pool is full)
    sort.Slice(m.objects, func(i, j int) bool {
        return m.objects[i].Distance < m.objects[j].Distance
    })

    // 3. Assign tiers
    m.tier3D = m.tier3D[:0]
    m.tierBillboard = m.tierBillboard[:0]
    m.tierCircle = m.tierCircle[:0]
    m.tierPoint = m.tierPoint[:0]

    for i := range m.objects {
        obj := &m.objects[i]
        tier := m.calcTier(obj.Distance)
        obj.CurrentTier = tier

        switch tier {
        case TierFull3D:
            m.tier3D = append(m.tier3D, obj)
        case TierBillboard:
            m.tierBillboard = append(m.tierBillboard, obj)
        case TierCircle:
            m.tierCircle = append(m.tierCircle, obj)
        case TierPoint:
            m.tierPoint = append(m.tierPoint, obj)
        // TierCulled: don't add anywhere
        }
    }
}
```

### 2. Point Batch Rendering

For thousands of points, use batched rendering:

```go
// engine/lod/points.go
func (m *LODManager) RenderPoints(screen *ebiten.Image, camera *Camera) {
    for _, obj := range m.tierPoint {
        // Project 3D position to screen
        screenX, screenY := camera.WorldToScreen(obj.Position)

        // Skip if off-screen
        if screenX < 0 || screenX > screenW || screenY < 0 || screenY > screenH {
            continue
        }

        // Draw single pixel (or use DrawTriangles for batching)
        screen.Set(int(screenX), int(screenY), obj.Color)
    }
}
```

### 3. Circle Rendering

```go
// engine/lod/circles.go
func (m *LODManager) RenderCircles(screen *ebiten.Image, camera *Camera) {
    for _, obj := range m.tierCircle {
        screenX, screenY := camera.WorldToScreen(obj.Position)

        // Apparent size based on distance
        apparentRadius := (obj.Radius / obj.Distance) * camera.FOVScale
        if apparentRadius < 2 {
            apparentRadius = 2 // Minimum visible size
        }

        // Use existing DrawCmd.Circle or vector.DrawFilledCircle
        vector.DrawFilledCircle(screen, float32(screenX), float32(screenY),
            float32(apparentRadius), obj.Color, true)
    }
}
```

### 4. 3D Object Pooling

```go
// engine/lod/pool.go
type Object3DPool struct {
    planets []*tetra.Planet  // Pre-allocated planet meshes
    inUse   map[string]int   // objectID -> pool index
    maxSize int
}

func (p *Object3DPool) Acquire(obj *LODObject) *tetra.Planet {
    // Check if already has a 3D object
    if idx, ok := p.inUse[obj.ID]; ok {
        return p.planets[idx]
    }

    // Find unused slot
    for i, planet := range p.planets {
        if _, used := p.inUseByIdx[i]; !used {
            p.inUse[obj.ID] = i
            planet.SetRadius(obj.Radius)
            planet.SetColor(obj.Color)
            return planet
        }
    }

    // Pool exhausted - object stays at lower LOD
    return nil
}
```

### 5. Frustum Culling

```go
// engine/lod/frustum.go
type Frustum struct {
    planes [6]Plane  // Near, far, left, right, top, bottom
}

func (f *Frustum) Contains(pos Vector3, radius float64) bool {
    for _, plane := range f.planes {
        if plane.DistanceTo(pos) < -radius {
            return false  // Completely outside this plane
        }
    }
    return true
}
```

## Integration with Existing Code

### Saturn Demo Integration

```go
// cmd/demo-game-saturn/main.go
type SaturnGame struct {
    lodManager *lod.LODManager
    // ... existing fields
}

func (g *SaturnGame) setupLOD() {
    g.lodManager = lod.NewManager(lod.DefaultLODConfig)

    // Register Saturn (always Full3D since it's the focus)
    g.lodManager.Add(lod.LODObject{
        ID:       "saturn",
        Position: Vector3{0, 0, 0},
        Radius:   saturnRadius,
        Color:    saturnColor,
        Model3D:  g.saturn,
    })

    // Register moons (will transition between tiers)
    for _, moon := range g.moons {
        g.lodManager.Add(lod.LODObject{
            ID:       moon.name,
            Position: moon.position,
            Radius:   moon.radius,
            Color:    moon.color,
            Model3D:  moon.planet,
        })
    }
}

func (g *SaturnGame) Update() error {
    // ... existing update code
    g.lodManager.Update(g.cameraPosition)
}

func (g *SaturnGame) Draw(screen *ebiten.Image) {
    // 1. Background
    g.spaceBackground.Draw(screen, nil)

    // 2. Distant objects (points, circles)
    g.lodManager.RenderPoints(screen, g.camera)
    g.lodManager.RenderCircles(screen, g.camera)

    // 3. 3D scene (only objects in Full3D tier)
    g.lodManager.ConfigureScene(g.scene3D)
    img := g.scene3D.Render()
    screen.DrawImage(img, nil)

    // 4. Billboards (sprites over 3D)
    g.lodManager.RenderBillboards(screen, g.camera)
}
```

## Performance Targets

| Scenario | Objects | Target FPS | Notes |
|----------|---------|------------|-------|
| Saturn orbit | ~10 | 60 | Current baseline |
| Star system | ~100 | 60 | Planets + moons |
| Asteroid field | ~1000 | 60 | Mostly points/circles |
| Galaxy view | ~10000 | 30 | All points |
| Galaxy zoom | ~100000 | 30 | Aggressive culling |

## Files to Create

| File | Purpose |
|------|---------|
| `engine/lod/manager.go` | LODManager and LODObject types |
| `engine/lod/config.go` | LODConfig and tier thresholds |
| `engine/lod/pool.go` | 3D object pooling |
| `engine/lod/points.go` | Batched point rendering |
| `engine/lod/circles.go` | Circle rendering |
| `engine/lod/frustum.go` | Frustum culling |
| `engine/lod/billboard.go` | Billboard sprite rendering |

## Success Criteria

- [ ] LODManager correctly assigns tiers based on distance
- [ ] 3D objects only created for closest N objects
- [ ] Points render efficiently (10000+ at 30fps)
- [ ] Circles render with apparent size based on distance
- [ ] Frustum culling skips off-screen objects
- [ ] Saturn demo works with LOD enabled
- [ ] No visual popping (smooth transitions between tiers)
- [ ] Memory usage stays bounded (object pooling works)

## Future Extensions

1. **Hysteresis**: Add distance buffer to prevent rapid tier switching
2. **Importance Weighting**: Player's target gets LOD priority
3. **Adaptive LOD**: Adjust thresholds based on frame time
4. **GPU Instancing**: Single draw call for identical objects
5. **Impostor Generation**: Pre-render 3D objects as billboards

## References

- [engine-capabilities.md](../reference/engine-capabilities.md) - Current rendering
- [tetra3d docs](https://github.com/SolarLune/tetra3d) - 3D library
- [Ebiten DrawTriangles](https://ebitengine.org/en/documents/performancetips.html) - Batch rendering
