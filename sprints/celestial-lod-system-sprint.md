# Sprint: Celestial LOD System

**Design Doc:** [design_docs/planned/celestial-lod-system.md](../design_docs/planned/celestial-lod-system.md)

## Goal
Implement a Level of Detail (LOD) system for rendering thousands of celestial objects efficiently by switching between 3D meshes, 2D circles, and points based on distance.

## Summary
| Aspect | Value |
|--------|-------|
| Total Effort | 2-3 sessions |
| AILANG Work | ~5% (minimal - data only) |
| Engine Work | ~95% (new `engine/lod/` package) |
| Risk | Low (pure engine, no AILANG complexity) |

---

## Session 1: Core LOD Manager ✅

### Tasks
- [x] Create `engine/lod/` package directory
- [x] Implement `engine/lod/config.go` - LODConfig and tier constants
- [x] Implement `engine/lod/types.go` - LODObject, LODTier, Vector3
- [x] Implement `engine/lod/manager.go` - LODManager core logic
- [x] Add distance calculation and tier assignment
- [x] Write basic unit tests

### Files Created
```
engine/lod/
├── config.go       # LODTier enum, Config struct, DefaultConfig(), GalaxyConfig(), SystemConfig()
├── types.go        # Vector3, Object, Stats, Camera interface, SimpleCamera
├── manager.go      # Manager struct, Update(), calcTier(), tier accessors
└── manager_test.go # Unit tests for tier assignment and object management
```

### Acceptance Criteria
- [x] `go build ./...` succeeds
- [x] LODManager correctly assigns tiers based on distance
- [x] Unit tests pass for tier calculation (5 tests passing)

---

## Session 2: Rendering Functions ✅

### Tasks
- [x] Implement `engine/lod/points.go` - batch point rendering
- [x] Implement `engine/lod/circles.go` - circle rendering with apparent size
- [x] Implement `engine/lod/billboard.go` - sprite billboard rendering
- [x] Add WorldToScreen projection helper (in types.go - SimpleCamera)
- [ ] Test rendering with simple demo (→ Session 3)

### Files Created
```
engine/lod/
├── points.go     # PointRenderer with RenderPoints(), RenderPointsTriangles(), RenderPointsDirect()
├── circles.go    # CircleRenderer with RenderCircles(), RenderCirclesWithGlow(), RenderCirclesTriangles()
└── billboard.go  # BillboardRenderer with RenderBillboards(), CreateDefaultPlanetSprite()
```

### Acceptance Criteria
- [x] Points render as colored pixels (multiple methods: direct, DrawImage, DrawTriangles)
- [x] Circles scale based on distance (apparent size) with min/max clamping
- [x] Billboard sprites support color tinting and scaling

---

## Session 3: Integration & Demo ✅

### Tasks
- [x] Create `cmd/demo-lod/main.go` - LOD stress test demo
- [ ] Integrate LODManager with Saturn demo (optional - future work)
- [x] Add frustum culling (integrated in manager.go Update())
- [x] Performance testing: 5000, 10000 objects (both render successfully)
- [ ] Update engine-capabilities.md with LOD documentation (future work)

### Files Created
```
cmd/demo-lod/main.go           # LOD stress test demo with WASD controls
out/screenshots/lod-5000.png   # 5000 objects test
out/screenshots/lod-10000.png  # 10000 objects test
```

### Acceptance Criteria
- [x] Demo renders 10000 points successfully
- [x] Demo renders circles with apparent size scaling
- [x] Frustum culling skips off-screen objects (integrated in manager)
- [ ] Saturn demo optionally uses LOD for moons (future work)

---

## Performance Targets

| Scenario | Objects | Target FPS |
|----------|---------|------------|
| Point cloud | 10,000 | 60 |
| Mixed (points + circles) | 5,000 | 60 |
| Full system (all tiers) | 1,000 | 60 |
| Galaxy view | 100,000 | 30 |

---

## Technical Notes

### Tier Thresholds (Configurable)
```go
DefaultLODConfig = LODConfig{
    Full3DDistance:    50,    // < 50 units: 3D mesh
    BillboardDistance: 200,   // < 200 units: sprite
    CircleDistance:    1000,  // < 1000 units: circle
    PointDistance:     10000, // < 10000 units: point
}
```

### Key Functions
```go
// manager.go
func (m *LODManager) Update(cameraPos Vector3)
func (m *LODManager) calcTier(distance float64) LODTier

// points.go
func (m *LODManager) RenderPoints(screen *ebiten.Image, camera *Camera)

// circles.go
func (m *LODManager) RenderCircles(screen *ebiten.Image, camera *Camera)
```

### No AILANG Changes Required
AILANG already provides celestial object data (position, color, radius) via `CelestialPlanet` type. The LOD system only affects HOW objects are rendered, not WHAT data exists.

---

## Dependencies
- Ebiten (existing)
- Tetra3D (existing, for Full3D tier)
- `engine/tetra/` (existing planet/ring code)

## Risks
- **Low**: Pure Go engine work, no AILANG complexity
- **Performance**: May need to optimize point rendering with DrawTriangles batching
- **Visual**: Tier transitions may cause popping (add hysteresis later)

---

## After Sprint
- [ ] Move design doc to `implemented/`
- [ ] Update engine-capabilities.md
- [ ] Consider: Object pooling for 3D tier
- [ ] Consider: Adaptive LOD based on frame time
