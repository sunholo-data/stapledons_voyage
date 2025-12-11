# Sprint: Tetra3D Planet Rendering with Solar System Demo

**Status:** Active
**Duration:** 2-3 days
**Design Doc:** [tetra3d-planet-rendering.md](../design_docs/planned/next/tetra3d-planet-rendering.md)
**Priority:** P2 (visual polish)

## Goal

Enable 3D textured planet rendering using Tetra3D for the dome view. Create a dedicated demo binary (`cmd/demo-solar-system`) to isolate, debug, and showcase the solar system flyby. The demo will be screenshot-driven to iterate on visuals before integrating with the main game.

## Investigation Summary

### What Works
- **Tetra3D rendering in isolation** - `demo-tetra` produces visible 3D sphere with lighting
- **AILANG CircleRGBA fallback** - 2D planet circles render correctly
- **Galaxy background** - GalaxyBg DrawCmd works through AILANG
- **DomeState migration** - AILANG controls camera position (cameraZ, velocity)

### What's Broken
- **Tetra3D compositing over starfield** - `planetLayer.Draw(screen)` produces no visible output
- **Disabled in dome_renderer.go:294-299** - Code commented out pending fix

### Root Cause Hypothesis
Based on investigation, likely issues:
1. **Transparent clear** - `camera.ClearWithColor(0,0,0,0)` may cause compositing issues
2. **Camera orientation** - Camera in PlanetLayer differs from working demo-tetra
3. **Lighting position** - Sun position for dome view differs from isolated test
4. **Scene graph hierarchy** - Planets may not be properly attached

## Sprint Tasks

### Day 1: Create Solar System Demo Binary

#### Task 1.1: Create cmd/demo-solar-system/main.go
- [ ] Copy structure from cmd/demo-tetra/main.go
- [ ] Add all 5 planets (Neptune, Saturn, Jupiter, Mars, Earth)
- [ ] Add Saturn's rings
- [ ] Include starfield background layer
- [ ] Support --screenshot flag for automated testing

**Key differences from demo-tetra:**
- Multiple planets at varying distances
- Background starfield (not just black)
- Camera movement (cruise simulation)
- Layer compositing test

#### Task 1.2: Test Basic Rendering
- [ ] Run demo, capture screenshot at frame 30
- [ ] Verify single planet renders on black background
- [ ] Verify multiple planets render at different depths
- [ ] Log scene graph state for debugging

#### Task 1.3: Add Starfield Background
- [ ] Create simple starfield (can be static)
- [ ] Composite Tetra3D planets OVER starfield
- [ ] Screenshot to verify compositing works

### Day 2: Debug Compositing Issues

#### Task 2.1: Investigate Alpha Channel
The transparent clear may cause issues. Test:
- [ ] Clear with solid black (0,0,0,255) instead of transparent
- [ ] Check if planets appear when background is opaque
- [ ] If yes, issue is alpha blending

#### Task 2.2: Test Camera Configurations
The PlanetLayer uses different camera setup than demo-tetra:
- [ ] Demo-tetra: camera at (0,0,4), planet at origin
- [ ] PlanetLayer: camera at (0,-3,cameraZ), planets at negative Z
- [ ] Test with demo-tetra camera settings
- [ ] Test explicit LookAt vs implicit orientation

#### Task 2.3: Test Lighting Configurations
- [ ] Verify light reaches planets in dome configuration
- [ ] Test with very bright ambient (1.0)
- [ ] Test sun position relative to camera/planets
- [ ] Screenshot each configuration

#### Task 2.4: Debug Scene Graph
- [ ] Add logging to verify planets added to scene root
- [ ] Log camera ColorTexture() dimensions
- [ ] Verify ColorTexture has non-zero pixel data

### Day 3: Integration + Polish

#### Task 3.1: Fix Identified Issue
Based on Day 2 findings:
- [ ] Apply fix to engine/tetra/scene.go or engine/view/planet_layer.go
- [ ] Verify fix in demo-solar-system
- [ ] Screenshot before/after

#### Task 3.2: Re-enable in Dome Renderer
- [ ] Uncomment planetLayer.Draw() in dome_renderer.go:294-299
- [ ] Test with main game
- [ ] Screenshot dome view with 3D planets

#### Task 3.3: Add Planet Textures
- [ ] Load equirectangular textures from assets/planets/
- [ ] Apply to UV sphere mesh
- [ ] Test texture orientation (flip if upside down)
- [ ] Saturn rings with transparency

#### Task 3.4: Polish + Performance
- [ ] Verify 60 FPS with all planets
- [ ] Adjust planet rotation speeds
- [ ] Fine-tune lighting for visual appeal
- [ ] Document final configuration

## Technical Details

### Demo Binary Structure

```go
// cmd/demo-solar-system/main.go
type DemoGame struct {
    // Background layer
    spaceBackground *background.SpaceBackground

    // 3D planet layer
    planetLayer *view.PlanetLayer

    // Animation
    cameraZ float64
    time    float64
}

func (g *DemoGame) Draw(screen *ebiten.Image) {
    // Layer 1: Starfield background
    g.spaceBackground.Draw(screen, nil)

    // Layer 2: 3D planets (the part we're debugging)
    g.planetLayer.Draw(screen)

    // Layer 3: Debug HUD
    g.drawHUD(screen)
}
```

### Camera Configuration Matrix

| Setting | demo-tetra (works) | PlanetLayer (broken) |
|---------|-------------------|---------------------|
| Camera Z | 4 | 10 → varies |
| Camera Y | 0 | -3 |
| LookAt | none | none |
| Clear color | transparent | transparent |
| FOV | 60° | 60° |
| Near/Far | 0.1/1000 | 0.1/1000 |

### Planet Configuration

| Planet | Color (fallback) | Radius | Distance | Y Offset | Texture |
|--------|-----------------|--------|----------|----------|---------|
| Neptune | #5078C8 | 1.0 | -15 | 2.25 | neptune.jpg |
| Saturn | #D2BE96 | 1.8 | -50 | 7.5 | saturn.jpg + rings |
| Jupiter | #DCB48C | 2.2 | -90 | 13.5 | jupiter.jpg |
| Mars | #C86450 | 0.5 | -130 | 19.5 | mars.jpg |
| Earth | #3C78C8 | 0.7 | -150 | 22.5 | earth_daymap.jpg |

## Files to Create/Modify

### New Files
- `cmd/demo-solar-system/main.go` - Solar system demo binary

### Files to Modify
- `engine/tetra/scene.go` - Potential fixes to Render()
- `engine/view/planet_layer.go` - Potential camera/lighting fixes
- `engine/view/dome_renderer.go` - Re-enable 3D planets (line 294-299)

### Files to NOT Modify
- `sim/*.ail` - AILANG code unchanged (already has CircleRGBA fallback)
- `sim_gen/*.go` - Generated, never edit

## Screenshot Checkpoints

All screenshots saved to `out/` for iteration:

| Checkpoint | Screenshot | Purpose |
|------------|------------|---------|
| 1.1 | `out/demo-solar-single.png` | Single planet on black |
| 1.2 | `out/demo-solar-multi.png` | All 5 planets |
| 1.3 | `out/demo-solar-starfield.png` | Composited over stars |
| 2.1a | `out/demo-solar-opaque.png` | Opaque clear test |
| 2.2a | `out/demo-solar-camera-test.png` | Camera config test |
| 2.3a | `out/demo-solar-bright.png` | Bright ambient test |
| 3.1 | `out/demo-solar-fixed.png` | After fix applied |
| 3.2 | `out/dome-tetra3d.png` | Integrated in dome |
| 3.3 | `out/demo-solar-textured.png` | With textures |

## Success Criteria

- [ ] `cmd/demo-solar-system` builds and runs
- [ ] Screenshot shows 3D planets on starfield background
- [ ] Planets have visible lighting/shading
- [ ] Saturn's rings render with transparency
- [ ] Camera cruise animation works (Z movement)
- [ ] Textured planets display correctly
- [ ] Integration with dome_renderer works
- [ ] 60 FPS maintained with all planets
- [ ] CircleRGBA fallback can be removed from AILANG

## AILANG Integration Note

This sprint is **Go-only** (engine visual polish). AILANG continues to:
- Control DomeState (cameraZ, velocity, cruiseTime)
- Render GalaxyBg background
- Handle game logic

Once Tetra3D planets work, AILANG's CircleRGBA fallback in `sim/bridge.ail:renderDomePlanets()` can be removed.

## Rollback Plan

If Tetra3D issues cannot be resolved:
1. Keep CircleRGBA fallback (functional, just not 3D)
2. Consider pre-rendered 2D planet sprites with size scaling
3. Document Tetra3D limitations for future reference
