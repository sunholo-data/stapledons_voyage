# Tetra3D Planet Rendering with AILANG Control

**Status:** Planned
**Priority:** P2 (visual polish)
**Dependencies:** dome-state-migration.md (implemented)

## Goal

Enable 3D textured planet rendering using Tetra3D, with AILANG controlling all state (camera position, cruise timing, velocity). This replaces the current CircleRGBA placeholder planets with proper 3D spheres featuring textures, rotation, lighting, and Saturn's rings.

## Current State

### What Works
- AILANG owns `DomeState` with `cameraZ`, `cruiseVelocity`, `cruiseTime`
- Go receives camera position via `SetCameraFromState(cameraZ, velocity)`
- Tetra3D scene is created with planets, textures load successfully
- CircleRGBA fallback planets render correctly (proves AILANG state flow works)

### What's Broken
- Tetra3D `planetLayer.Draw(screen)` produces no visible output
- Debug sphere at (0, 0, -10) also invisible
- Scene renders to buffer but buffer appears empty/transparent

## Investigation Areas

### 1. Camera Setup
- Default camera at Z=5, looking at -Z
- Planets positioned at negative Z values (-15 to -150)
- Camera should see planets, but nothing appears
- **Check:** Field of view, near/far clip planes, camera orientation

### 2. Tetra3D Scene Rendering
```go
// Current render code in scene.go
func (s *Scene) Render() *ebiten.Image {
    s.buffer.Clear()
    s.camera.ClearWithColor(tetra3d.NewColor(0, 0, 0, 0)) // Transparent
    s.camera.RenderScene(s.scene)
    s.buffer.DrawImage(s.camera.ColorTexture(), opt)
    return s.buffer
}
```
- **Check:** Is `ColorTexture()` returning valid image?
- **Check:** Is the scene graph correct (planets added to Root)?
- **Check:** Are materials/shaders set up correctly?

### 3. Compositing
```go
// In planet_layer.go
func (pl *PlanetLayer) Draw(screen *ebiten.Image) {
    img3d := pl.scene.Render()
    screen.DrawImage(img3d, nil)  // Composite over existing
}
```
- **Check:** Is `img3d` non-nil and has content?
- **Check:** Blend mode for compositing

### 4. Lighting
- Sun light positioned at (0, 5, 20) pointing toward planets
- Ambient light at 0.8 intensity
- **Check:** Are planets lit? Could be rendering black on black?

## Proposed Solution

### Phase 1: Debug Tetra3D Pipeline
1. Add debug logging to verify:
   - Planet models added to scene
   - Camera position updates received
   - Scene.Render() produces non-empty buffer
   - ColorTexture() has pixel data

2. Test with minimal case:
   - Single white sphere at origin
   - Camera at (0, 0, 5) looking at origin
   - Bright ambient light
   - Render to solid color background (not transparent)

### Phase 2: Fix Rendering
Based on investigation, likely fixes:
- Camera orientation (may need explicit LookAt)
- Render buffer format (alpha channel issues?)
- Scene graph hierarchy
- Material setup

### Phase 3: AILANG Integration
Once Tetra3D renders:
1. Remove CircleRGBA fallback from AILANG
2. Keep GalaxyBg background in AILANG
3. Go renders 3D planets on top of AILANG background
4. Layer order: GalaxyBg (AILANG) → Planets (Go/Tetra3D) → Floor (AILANG)

## Architecture

```
AILANG (sim/bridge.ail)              Go (engine/view/)
─────────────────────────────────────────────────────────
DomeState {                          DomeRenderer
  cameraZ: float     ─────────────→  SetCameraFromState()
  cruiseVelocity: float               │
  cruiseTime: float                   ↓
}                                    PlanetLayer
                                      │
stepDome(dt) updates cameraZ          │ SetCameraPosition()
                                      │ Update(dt) - rotation
renderDome() returns GalaxyBg         │
                                      ↓
                                     Tetra3D Scene
                                      │ camera, planets, lighting
                                      │
                                      ↓
                                     Draw() → screen
```

## Planet Configuration

| Planet | Texture | Radius | Distance | Y Offset | Special |
|--------|---------|--------|----------|----------|---------|
| Neptune | neptune.jpg | 1.0 | 15 | 2.25 | - |
| Saturn | saturn.jpg | 1.8 | 50 | 7.5 | Rings |
| Jupiter | jupiter.jpg | 2.2 | 90 | 13.5 | - |
| Mars | mars.jpg | 0.5 | 130 | 19.5 | - |
| Earth | earth_daymap.jpg | 0.7 | 150 | 22.5 | - |

## Success Criteria

- [ ] Tetra3D planets visible on screen
- [ ] Textures display correctly (equirectangular mapping)
- [ ] Planets rotate slowly during cruise
- [ ] Saturn's rings render with tilt
- [ ] Camera position controlled by AILANG cameraZ
- [ ] Smooth cruise animation (60 FPS)
- [ ] Planets scale correctly with perspective

## Files to Modify

### Investigation
- `engine/tetra/scene.go` - Add debug logging to Render()
- `engine/view/planet_layer.go` - Verify Draw() receives content
- `engine/view/dome_renderer.go` - Test minimal rendering case

### Implementation
- `engine/view/dome_renderer.go` - Re-enable planetLayer.Draw()
- `sim/bridge.ail` - Remove CircleRGBA fallback (keep GalaxyBg)

## Rollback Plan

If Tetra3D issues cannot be resolved:
- Keep current CircleRGBA implementation (works, just not textured)
- Consider alternative 3D library
- Or implement 2D sprite-based planets with pre-rendered textures

## Notes

The CircleRGBA fallback demonstrates the AILANG→Go state flow works correctly. The issue is isolated to Tetra3D's rendering pipeline, not the architecture.
