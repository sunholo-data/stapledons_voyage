# Tetra3D Integration

## Status
- Status: Planned
- Priority: P0 (Foundation for 3D)
- Complexity: Medium
- Estimated: 2-3 days
- Blocks: 3D planet rendering, solar system view

## Problem Statement

We need to render 3D spheres with textures for planets. Ebiten is a 2D engine, but [Tetra3D](https://github.com/solarlune/Tetra3D) provides 3D capabilities built on top of Ebitengine.

## Physics Validation

**This is infrastructure**, not a visual effect. No physics validation required.

Tetra3D will enable physics-accurate rendering:
- Proper sphere geometry for planets
- Correct lighting (terminator line)
- Accurate day/night sides based on sun position

## Tetra3D Capabilities

Based on [documentation](https://pkg.go.dev/github.com/solarlune/tetra3d):

| Feature | Support | Use Case |
|---------|---------|----------|
| Mesh rendering | Yes | Planet spheres |
| Icosphere generation | `NewIcosphereMesh()` | Smooth spheres |
| Textures | Yes | NASA planet images |
| Materials | Yes | Surface properties |
| Lighting | Ambient, Directional, Point | Sun illumination |
| Camera | Perspective + Orthographic | Space views |
| GLTF loading | Yes | Pre-made models |
| Animation | Yes | Future use |
| Render to Image | Yes | Composite with 2D |

**Aesthetic**: PS1/N64 era polygon rendering with high-resolution textures (see below)

## Texture Resolution Support

Tetra3D uses Ebitengine's image system, inheriting its texture capabilities:

| Resolution | Support | Notes |
|------------|---------|-------|
| 2K (2048×1024) | ✅ Optimal | Fits in Ebitengine texture atlas (4096×4096) |
| 4K (4096×2048) | ✅ Works | At atlas limit, separate allocation |
| 8K+ (7680×4320) | ⚠️ Risky | Device-dependent, may fail on some GPUs |

**Key insight**: The "N64 aesthetic" comes from **polygon rendering** (low-poly meshes, vertex lighting), NOT texture resolution.

### Visual Combination

| Component | Aesthetic | Result |
|-----------|-----------|--------|
| Geometry | N64/PS1 (low-poly icosphere, 3-4 subdivisions) | Stylized, retro feel |
| Textures | High-res (4K NASA images) | Beautiful, detailed surfaces |
| Combined | Lo-fi geometry + hi-fi textures | Unique artistic style |

This gives **crisp NASA imagery wrapped on stylized 3D geometry** - a distinctive look that's both retro and beautiful.

## SR/GR Shader Compatibility

**Critical**: Our existing SR/GR shaders work perfectly with Tetra3D content.

### Why It Works

Tetra3D renders to `*ebiten.Image` → Our shaders process `*ebiten.Image` → Shaders don't care about pixel source.

```
Rendering Pipeline:
┌─────────────────────────────────────────────────────────┐
│ 1. TETRA3D RENDER                                       │
│    └─→ Renders 3D planets to *ebiten.Image buffer       │
│                                                         │
│ 2. COMPOSITE                                            │
│    └─→ Draw Tetra3D buffer onto main scene              │
│                                                         │
│ 3. SR/GR SHADERS (post-processing)                      │
│    └─→ Process entire composed *ebiten.Image            │
│    └─→ Apply Doppler shift, aberration, lensing         │
│                                                         │
│ 4. FINAL OUTPUT                                         │
│    └─→ Physics-accurate relativistic rendering          │
└─────────────────────────────────────────────────────────┘
```

### Rendering Code Pattern

```go
func (v *SpaceView) Draw(screen *ebiten.Image) {
    // 1. Draw starfield background
    v.starfield.Draw(v.buffer)

    // 2. Render 3D planet (Tetra3D) - HIGH-RES TEXTURE
    planetImg := v.tetraScene.Render()
    v.buffer.DrawImage(planetImg, &opts)

    // 3. Apply SR effects (if moving)
    if v.velocity > 0.01 {
        v.srWarp.SetForwardVelocity(v.velocity)
        v.srWarp.Apply(v.srBuffer, v.buffer)
        v.buffer = v.srBuffer
    }

    // 4. Apply GR effects (if near massive object)
    if v.grIntensity > 0 {
        v.grWarp.Apply(screen, v.buffer)
    } else {
        screen.DrawImage(v.buffer, nil)
    }
}
```

### Effect Examples

| Scenario | Tetra3D Renders | Shader Applied | Result |
|----------|-----------------|----------------|--------|
| Approaching Saturn at 0.3c | Saturn sphere with rings | SR Doppler shift | Blue-tinted Saturn |
| Near black hole | Distant stars + planets | GR lensing | Bent light around BH |
| Earth arrival at 0.0c | Earth sphere | None | Normal Earth view |

## Architecture

### Integration Pattern

Tetra3D renders to an `*ebiten.Image`, which we composite into our 2D view:

```
┌──────────────────────────────────────────┐
│            MAIN EBITEN LOOP               │
│                                          │
│  ┌────────────────────────────────────┐  │
│  │         VIEW SYSTEM                │  │
│  │                                    │  │
│  │  ┌──────────────────────────────┐  │  │
│  │  │    TETRA3D SCENE             │  │  │
│  │  │  ┌────────────────────────┐  │  │  │
│  │  │  │ Camera → RenderBuffer │  │  │  │
│  │  │  │ Lights                 │  │  │  │
│  │  │  │ Meshes (planets)       │  │  │  │
│  │  │  └────────────────────────┘  │  │  │
│  │  │            │                 │  │  │
│  │  │            ▼                 │  │  │
│  │  │     *ebiten.Image            │  │  │
│  │  └──────────────────────────────┘  │  │
│  │            │                       │  │
│  │            ▼                       │  │
│  │  ┌──────────────────────────────┐  │  │
│  │  │   Composite into View        │  │  │
│  │  │   (behind UI, over starfield)│  │  │
│  │  └──────────────────────────────┘  │  │
│  └────────────────────────────────────┘  │
└──────────────────────────────────────────┘
```

### File Structure

```
engine/
├── tetra/
│   ├── scene.go        # Tetra3D scene wrapper
│   ├── camera.go       # Camera helpers
│   ├── planet.go       # Planet sphere rendering
│   └── lighting.go     # Sun/star lighting
```

## Implementation

### Scene Wrapper

```go
// engine/tetra/scene.go
package tetra

import (
    "github.com/hajimehoshi/ebiten/v2"
    "github.com/solarlune/tetra3d"
)

type Scene struct {
    library *tetra3d.Library
    scene   *tetra3d.Scene
    camera  *tetra3d.Camera
    buffer  *ebiten.Image
    width   int
    height  int
}

func NewScene(width, height int) *Scene {
    s := &Scene{
        library: tetra3d.NewLibrary(),
        width:   width,
        height:  height,
    }

    s.scene = s.library.AddScene("main")
    s.buffer = ebiten.NewImage(width, height)

    // Setup camera
    s.camera = tetra3d.NewCamera(width, height)
    s.camera.SetFieldOfView(60)  // degrees
    s.camera.SetNear(0.1)
    s.camera.SetFar(1000)
    s.scene.Root.AddChildren(s.camera)

    return s
}

func (s *Scene) Render() *ebiten.Image {
    s.buffer.Clear()
    s.camera.Clear()
    s.camera.RenderScene(s.scene)

    // Draw camera's color texture to buffer
    opt := &ebiten.DrawImageOptions{}
    s.buffer.DrawImage(s.camera.ColorTexture(), opt)

    return s.buffer
}
```

### Planet Rendering

**Note**: Tetra3D v0.17+ API uses `float32` for positions and `Rotate()` for rotations.

```go
// engine/tetra/planet.go
package tetra

import (
    "github.com/hajimehoshi/ebiten/v2"
    "github.com/solarlune/tetra3d"
)

type Planet struct {
    model       *tetra3d.Model
    material    *tetra3d.Material
    rotation    float64
    rotationSpd float64
}

func NewPlanet(name string, radius float64, texture *ebiten.Image) *Planet {
    // Create icosphere mesh - NewIcosphereMesh only takes subdivision level
    // Use SetLocalScale for radius
    mesh := tetra3d.NewIcosphereMesh(3)  // 3 subdivisions for N64 aesthetic

    // Create material with texture
    mat := tetra3d.NewMaterial("planet_" + name)
    mat.Texture = texture
    mat.Color = tetra3d.NewColor(1, 1, 1, 1)

    // Apply material to all mesh parts
    for _, meshPart := range mesh.MeshParts {
        meshPart.Material = mat
    }

    // Create model node
    model := tetra3d.NewModel(name, mesh)

    // Scale to desired radius
    r := float32(radius)
    model.SetLocalScale(r, r, r)

    return &Planet{
        model:       model,
        material:    mat,
        rotationSpd: 0.5,
    }
}

func (p *Planet) SetPosition(x, y, z float64) {
    // Tetra3D uses float32 for positions
    p.model.SetLocalPosition(float32(x), float32(y), float32(z))
}

func (p *Planet) Update(dt float64) {
    // Rotate around Y axis incrementally
    delta := p.rotationSpd * dt
    p.rotation += delta
    p.model.Rotate(0, 1, 0, float32(delta))
}

func (p *Planet) AddToScene(scene *Scene) {
    scene.Root().AddChildren(p.model)
}
```

### Lighting (Sun)

**Note**: Tetra3D v0.17+ uses float32 and Rotate() for orientation.

```go
// engine/tetra/lighting.go
package tetra

import "github.com/solarlune/tetra3d"

type SunLight struct {
    light *tetra3d.DirectionalLight
}

func NewSunLight() *SunLight {
    // NewDirectionalLight(name, r, g, b, energy)
    light := tetra3d.NewDirectionalLight("sun", 1, 1, 1, 1)  // White, full energy

    return &SunLight{light: light}
}

func (s *SunLight) SetDirection(x, y, z float64) {
    // Position the light and rotate to point at origin
    s.light.SetLocalPosition(float32(x*100), float32(y*100), float32(z*100))
    // Rotate to face origin (use LookAt if available, or calculate rotation)
}

func (s *SunLight) AddToScene(scene *Scene) {
    scene.Root().AddChildren(s.light)
}
```

## Technical Spike Plan

Before full implementation, run a spike to validate:

### Day 1: Basic Setup

1. Add Tetra3D dependency:
   ```bash
   go get github.com/solarlune/tetra3d
   ```

2. Create minimal `cmd/tetra-demo/main.go`:
   - Initialize Tetra3D scene
   - Add icosphere with solid color
   - Render to screen
   - Verify it displays

### Day 2: Textures & Lighting

3. Load NASA Earth texture
4. Apply to icosphere
5. Add directional light (sun)
6. Verify terminator line (day/night boundary)
7. Add rotation animation

### Day 3: Integration

8. Composite 3D render into view system
9. Test with starfield background
10. Performance benchmarking

## Performance Considerations

| Concern | Mitigation |
|---------|------------|
| Tetra3D is software rendering | Only render visible planets (1-2 at a time) |
| Large textures | 4K for hero planets (Earth), 2K for others |
| Multiple planets | LOD: distant planets as 2D sprites |
| Frame rate | Budget 5ms for 3D render per frame |
| Texture memory | ~32MB for 4K, ~8MB for 2K per planet |

### Texture Resolution Recommendations

| Planet | Recommended | Reason |
|--------|-------------|--------|
| Earth | 4K (4096×2048) | Hero moment, emotional impact |
| Saturn | 2K body + 2K rings | Rings need detail |
| Jupiter | 2K (2048×1024) | Bands visible at 2K |
| Mars | 2K (2048×1024) | Surface features |
| Moon | 1K (1024×512) | Smaller, less detail needed |

**Total VRAM for all planets**: ~60MB (acceptable)

## Success Criteria

- [ ] Tetra3D compiles and runs with Ebitengine
- [ ] Can render textured icosphere (planet)
- [ ] 4K textures load and display correctly
- [ ] Directional lighting creates terminator line
- [ ] Sphere rotates smoothly
- [ ] 3D render composites correctly with 2D view
- [ ] **SR shader applies to Tetra3D output** (Doppler shift visible)
- [ ] **GR shader applies to Tetra3D output** (lensing works)
- [ ] 60fps maintained with one planet + shaders

## Demo Command

After implementation, add demo command:

```bash
./bin/game --demo-3d          # Show rotating Earth sphere
./bin/game --demo-3d --planet saturn  # Show Saturn with rings
```

## Dependencies

- **Requires**: View system (for compositing)
- **Enables**: 3D planet rendering, solar system view

## Rejected Alternatives

| Alternative | Why Rejected |
|-------------|--------------|
| Custom raymarched shader | More work, harder to add features |
| Pre-rendered sprites | No true 3D rotation, limited angles |
| OpenGL direct | Breaks Ebitengine abstraction |

## References

- [Tetra3D GitHub](https://github.com/solarlune/Tetra3D)
- [Tetra3D Examples](https://github.com/solarlune/Tetra3D/tree/main/examples)
- [Tetra3D Wiki](https://github.com/solarlune/Tetra3D/wiki)
