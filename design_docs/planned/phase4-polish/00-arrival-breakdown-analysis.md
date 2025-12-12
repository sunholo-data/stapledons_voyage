# Arrival Sequence Breakdown Analysis

## Status
- Status: Analysis
- Priority: P0
- Created: 2024-12-08

## Problem

The arrival sequence design docs ([arrival-sequence.md](arrival-sequence.md) and [cinematic-arrival-system.md](cinematic-arrival-system.md)) are too ambitious without foundational elements in place.

**Current state:** We jumped into arrival rendering without having:
- A view system (how do different game screens work?)
- 3D sphere rendering capability
- Bridge interior layout
- Isometric gameplay view
- Galaxy map viewer

## 3D Sphere Capability Investigation

### Option 1: Tetra3D (Recommended)

[Tetra3D](https://pkg.go.dev/github.com/solarlune/tetra3d) is a full 3D framework for Ebitengine:

| Feature | Status |
|---------|--------|
| Textured meshes | Yes |
| Icosphere generation | `NewIcosphereMesh()` |
| Materials & textures | Yes |
| Lighting (ambient, directional, point) | Yes |
| Camera with frustum culling | Yes |
| GLTF/GLB model loading | Yes |
| Animation | Yes |
| Ray casting | Yes |

**Rendering quality**: PS1/N64 era (intentionally retro, perfect for our style)

**Integration**: Works within Ebitengine, renders to `*ebiten.Image`

### Option 2: Custom Raymarched Shader

Write a Kage shader that raymarches spheres:

```kage
// Sphere SDF
func sphereSDF(p vec3, center vec3, radius float) float {
    return length(p - center) - radius
}

// UV mapping for texture
func sphereUV(p vec3, center vec3) vec2 {
    d := normalize(p - center)
    u := 0.5 + atan2(d.z, d.x) / (2.0 * 3.14159)
    v := 0.5 - asin(d.y) / 3.14159
    return vec2(u, v)
}
```

**Pros**: No external dependency, full shader control
**Cons**: More complex, need to write from scratch, single-pass limitation

### Recommendation

**Use Tetra3D for 3D planet rendering**:
1. Already handles sphere generation and texturing
2. Proper lighting (terminator line, atmosphere glow)
3. Can composite 3D render onto 2D game via `*ebiten.Image`
4. PS1/N64 aesthetic fits our hard sci-fi + retro style

## Dependency Graph

```
┌─────────────────────────────────────────────────────────────────┐
│                     ARRIVAL SEQUENCE                             │
│                   (what we tried to build)                       │
└───────────────────────────────┬─────────────────────────────────┘
                                │
                                ▼
        ┌───────────────────────┼───────────────────────┐
        │                       │                       │
        ▼                       ▼                       ▼
┌───────────────┐    ┌───────────────────┐    ┌───────────────────┐
│ 3D PLANETS    │    │ BRIDGE VIEW       │    │ CAMERA SYSTEM     │
│ (spheres)     │    │ (interior layout) │    │ (tumble, paths)   │
└───────┬───────┘    └─────────┬─────────┘    └─────────┬─────────┘
        │                      │                        │
        ▼                      ▼                        │
┌───────────────┐    ┌───────────────────┐              │
│ TETRA3D       │    │ ISOMETRIC VIEW    │              │
│ INTEGRATION   │    │ (gameplay area)   │              │
└───────────────┘    └─────────┬─────────┘              │
                               │                        │
                               ▼                        ▼
                    ┌───────────────────────────────────────┐
                    │           VIEW SYSTEM                  │
                    │ (how screens compose & transition)    │
                    └───────────────────────────────────────┘
                                        │
                                        ▼
                    ┌───────────────────────────────────────┐
                    │         SPACE BACKGROUND              │
                    │   (starfield, parallax layers)        │
                    └───────────────────────────────────────┘
```

## Proposed Design Doc Breakdown

### Layer 0: Foundation (Build First)

| Doc | Purpose | Blocks |
|-----|---------|--------|
| **view-system.md** | How game views compose (space bg + content + UI) | Everything |
| **space-background.md** | Starfield rendering, parallax layers | All space views |

### Layer 1: Rendering Capabilities

| Doc | Purpose | Blocks |
|-----|---------|--------|
| **tetra3d-integration.md** | Add Tetra3D for 3D rendering | 3D planets |
| **3d-sphere-planets.md** | Render textured spheres with Tetra3D | Arrival, galaxy map |
| **isometric-view.md** | Isometric tile/entity rendering | Ship interior, surface |

### Layer 2: Game Views

| Doc | Purpose | Blocks |
|-----|---------|--------|
| **bridge-interior.md** | Bridge layout, stations, crew | Arrival Phase 4-5 |
| **galaxy-map.md** | Star system navigation | Destination selection |
| **solar-system-view.md** | Planets orbiting star | Arrival fly-through |

### Layer 3: Arrival (Rebuild)

| Doc | Purpose | Blocks |
|-----|---------|--------|
| **arrival-sequence-v2.md** | Simplified arrival using built components | First experience |

## Recommended Build Order

```
Week 1-2: Foundation
├── 1. view-system.md           # Define how views compose
├── 2. space-background.md      # Starfield with parallax
└── 3. tetra3d-integration.md   # Add 3D capability

Week 3-4: Core Views
├── 4. 3d-sphere-planets.md     # Render actual planets
├── 5. isometric-view.md        # Ship interior basics
└── 6. solar-system-view.md     # Orbital mechanics display

Week 5-6: Polish & Combine
├── 7. bridge-interior.md       # Full bridge with stations
├── 8. galaxy-map.md            # Navigation UI
└── 9. arrival-sequence-v2.md   # Assemble final sequence
```

## Tech Demo Milestones

Each design doc includes demo commands to validate incrementally:

| Stage | Demo Command | What It Proves |
|-------|--------------|----------------|
| 1 | `--demo-planet earth` | Tetra3D renders 4K textured sphere |
| 2 | `--demo-planet saturn` | Rings work |
| 3 | `--demo-planet earth --velocity 0.3` | SR shaders + Tetra3D |
| 4 | `--demo-planet earth --gr 0.5` | GR shaders + Tetra3D |
| 5 | `--demo-engine-solar` | Multiple planets, orbital view |
| 6 | `--demo-game-bridge` | Isometric + space background |
| 7 | `--demo-game-bridge --planet earth` | Full composition |
| 8 | `--demo-arrival` | Complete arrival sequence |

**Key validation**: Each demo proves the next layer works before building on top.

## Immediate Next Steps

1. **Create `view-system.md`** - Define how views layer:
   - Background layer (space/starfield)
   - Content layer (3D planets, isometric tiles, etc.)
   - UI layer (HUD, panels, dialogue)
   - Transition system between views

2. **Create `tetra3d-integration.md`** - Technical spike:
   - Add Tetra3D dependency
   - Demo rendering a textured sphere
   - Composite 3D render onto 2D game
   - Performance testing

3. **Create `3d-sphere-planets.md`** - Planet rendering:
   - Use Tetra3D icospheres
   - Apply NASA textures
   - Add lighting (sun direction)
   - Atmosphere glow (Fresnel)
   - Ring systems (Saturn)

## Questions for User

1. **Tetra3D vs custom shader?** - Tetra3D is faster to implement but adds dependency
2. **PS1/N64 aesthetic acceptable?** - Tetra3D has intentional retro look
3. **Which view first?** - Bridge interior or solar system?
4. **Galaxy map scope?** - Full 3D starmap or 2D projection?

## Conclusion

The arrival sequence was too ambitious because it requires:
- 3D planet rendering (we don't have)
- Bridge view layout (we don't have)
- Isometric interior (we don't have)
- View composition system (we don't have)

**Recommendation**: Build foundation docs first, then layer up to arrival.
