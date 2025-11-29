# Sprint 002: Camera & Viewport

**Status:** Completed
**Goal:** Implement camera system to fix black edge issue and enable larger world rendering
**Estimated Effort:** 2-3 sessions
**AILANG Dependency:** Low (mock types only)
**Design Doc:** [camera-viewport.md](../design_docs/planned/v0_3_0/camera-viewport.md)

## Context

Sprint 001 revealed a "black edge" issue - the 64x64 tile world (512x512 pixels) doesn't fill the 640x480 internal resolution. A camera system will:
1. Center the view on the world
2. Enable scrolling for larger worlds
3. Provide foundation for player following

## Success Criteria

- [x] World centered in viewport (no black edges for current size)
- [x] Camera type added to mock sim_gen
- [x] World-to-screen coordinate transforms work correctly
- [x] Screen-to-world transforms work (for future mouse interaction)
- [x] Viewport culling skips off-screen tiles (performance)
- [x] `make eval-mock` still passes

## Tasks

### Phase 1: Camera Types in Mock (P0)

Add Camera type to sim_gen mock to match design doc.

| Task | File | Description |
|------|------|-------------|
| 1.1 | `sim_gen/protocol.go` | Add Camera struct |
| 1.2 | `sim_gen/protocol.go` | Add Camera to FrameOutput |
| 1.3 | `sim_gen/funcs.go` | InitWorld returns Camera in output |
| 1.4 | `sim_gen/funcs.go` | Step returns Camera centered on world |

**Camera struct:**
```go
type Camera struct {
    X, Y float64  // World position (center of view)
    Zoom float64  // 1.0 = normal
}
```

**Estimated:** 0.5 session

### Phase 2: Camera Transform Package (P0)

Create engine/camera package for coordinate transforms.

| Task | File | Description |
|------|------|-------------|
| 2.1 | `engine/camera/transform.go` | Transform struct, WorldToScreen, ScreenToWorld |
| 2.2 | `engine/camera/viewport.go` | Viewport struct, Contains(), CalculateViewport() |
| 2.3 | Tests | Unit tests for transforms |

**Key functions:**
```go
func FromOutput(cam sim_gen.Camera, screenW, screenH int) Transform
func (t Transform) WorldToScreen(worldX, worldY float64) (float64, float64)
func (t Transform) ScreenToWorld(screenX, screenY float64) (float64, float64)
```

**Estimated:** 1 session

### Phase 3: Renderer Integration (P0)

Update renderer to apply camera transform.

| Task | File | Description |
|------|------|-------------|
| 3.1 | `engine/render/draw.go` | Import camera package |
| 3.2 | `engine/render/draw.go` | Apply transform to all DrawCmds |
| 3.3 | `engine/render/draw.go` | Cull off-screen DrawCmds |
| 3.4 | Test | Visual verification with `make run-mock` |

**Estimated:** 0.5 session

### Phase 4: Camera Centering Logic (P1)

Center camera on world for current size, prepare for future scrolling.

| Task | File | Description |
|------|------|-------------|
| 4.1 | `sim_gen/funcs.go` | Calculate world center |
| 4.2 | `sim_gen/funcs.go` | Set Camera.X, Camera.Y to center |
| 4.3 | Test | Black edges should disappear |

**Estimated:** 0.5 session

### Phase 5: Verification & Cleanup (P1)

| Task | Description |
|------|-------------|
| 5.1 | Run `make eval-mock`, verify scenarios pass |
| 5.2 | Visual test - world should be centered |
| 5.3 | Test zoom (set Camera.Zoom to 2.0, verify scaling) |
| 5.4 | Update sprint progress JSON |

**Estimated:** 0.5 session

## Technical Details

### Coordinate Systems

```
Screen Space (pixels)         World Space (tiles * tileSize)
┌─────────────────┐           ┌─────────────────┐
│ (0,0)           │           │                 │
│     ┌─────┐     │           │ Camera.X,Y      │
│     │View │     │  ←─────→  │    (center)     │
│     └─────┘     │           │                 │
│         (640,480)           │                 │
└─────────────────┘           └─────────────────┘
```

### Transform Math

```go
// World to Screen
screenX = (worldX - camera.X) * camera.Zoom + screenWidth/2
screenY = (worldY - camera.Y) * camera.Zoom + screenHeight/2

// Screen to World
worldX = (screenX - screenWidth/2) / camera.Zoom + camera.X
worldY = (screenY - screenHeight/2) / camera.Zoom + camera.Y
```

### Default Camera Position

For a 64x64 world at 8px/tile = 512x512 world pixels:
- Center X: 256
- Center Y: 256
- Zoom: 1.0

This centers the 512x512 world in the 640x480 viewport.

## Dependencies

| Dependency | Status | Notes |
|------------|--------|-------|
| Sprint 001 | Complete | Mock sim_gen exists |
| Display Manager | Complete | Provides screen dimensions |
| Asset Manager | Complete | Not directly needed |

## Risks & Mitigations

| Risk | Mitigation |
|------|------------|
| Transform math errors | Write unit tests first |
| Performance with culling | Benchmark before/after |
| Mouse coords broken | Implement ScreenToWorld early |

## Future Work (not this sprint)

- Smooth camera following (lerp)
- Camera shake
- Zoom controls
- Camera bounds (don't scroll past world edge)

## Notes

- This is pure Go work - no AILANG changes needed
- Camera position comes from mock, will eventually come from AILANG
- Keep camera logic simple - complexity should be in AILANG later
- Test with different world sizes to verify centering
