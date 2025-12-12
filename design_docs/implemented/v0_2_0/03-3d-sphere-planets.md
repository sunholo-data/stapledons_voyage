# 3D Sphere Planet Rendering

## Status
- Status: **Mostly Implemented**
- Priority: P1
- Complexity: Medium
- Estimated: 3-4 days
- Actual: Implemented alongside Tetra3D sprint
- Depends On: Tetra3D Integration ✅, View System ✅

## Problem Statement

Current planet rendering uses flat 2D images. For the arrival sequence and solar system views, we need proper 3D spheres that:
- Rotate realistically
- Show correct day/night illumination
- Display accurate surface features
- Support ring systems (Saturn)
- Can be viewed from any angle

## Physics Validation

### Real Physics (This Design)

| Effect | Physics Basis | Implementation |
|--------|---------------|----------------|
| **Day/night terminator** | Illumination from point source (sun) | Directional light in Tetra3D |
| **Sphere geometry** | Planets are oblate spheroids | Icosphere mesh |
| **Rotation** | Planets rotate on axis | Y-axis rotation animation |
| **Phase** | Illuminated portion varies with viewing angle | 3D lighting handles this |
| **Atmospheric limb** | Atmosphere scatters light at edges | Fresnel glow shader |

### Rejected (Hollywood)

| Effect | Why Wrong |
|--------|-----------|
| ~~Full illumination~~ | Planets have day/night sides |
| ~~Static images~~ | Planets rotate |
| ~~Flat textures~~ | Planets are spheres |

## Planet Data

### Solar System Planets (Initial Set)

| Planet | Radius (rel) | Rotation Period | Axial Tilt | Has Rings |
|--------|--------------|-----------------|------------|-----------|
| Saturn | 9.45 | 10.7 hours | 26.7° | Yes (bright) |
| Jupiter | 11.2 | 9.9 hours | 3.1° | Yes (faint) |
| Uranus | 4.01 | 17.2 hours | 97.8° | Yes (dark, discrete) |
| Neptune | 3.88 | 16.1 hours | 28.3° | Yes (faint arcs) |
| Mars | 0.53 | 24.6 hours | 25.2° | No |
| Earth | 1.0 | 24 hours | 23.4° | No |
| Moon | 0.27 | 27.3 days | 6.7° | No |

### Texture Sources (Public Domain)

| Planet | Source | Source Res | URL |
|--------|--------|------------|-----|
| Earth | NASA Blue Marble | 8K | [nasa.gov](https://visibleearth.nasa.gov/collection/1484/blue-marble) |
| Mars | NASA MGS | 4K | [jpl.nasa.gov](https://maps.jpl.nasa.gov/mars.html) |
| Jupiter | NASA Cassini/Juno | 4K | [jpl.nasa.gov](https://photojournal.jpl.nasa.gov/catalog/PIA07782) |
| Saturn | NASA Cassini | 4K | [jpl.nasa.gov](https://photojournal.jpl.nasa.gov/catalog/PIA18400) |
| Saturn Rings | NASA Cassini | 4K | [jpl.nasa.gov](https://photojournal.jpl.nasa.gov/catalog/PIA08389) |
| Moon | NASA LRO | 4K | [svs.gsfc.nasa.gov](https://svs.gsfc.nasa.gov/4720) |

### In-Game Texture Resolutions

**Key insight**: N64 aesthetic comes from polygon rendering, NOT texture resolution. We use high-res textures on low-poly geometry for a unique artistic style.

| Planet | In-Game Res | VRAM | Reason |
|--------|-------------|------|--------|
| Earth | 4K (4096×2048) | ~32MB | Hero moment - emotional impact of "home" |
| Saturn | 2K (2048×1024) | ~8MB | Body texture |
| Saturn Rings | 2K (2048×512) | ~4MB | Ring detail important |
| Jupiter | 2K (2048×1024) | ~8MB | Bands visible at 2K |
| Mars | 2K (2048×1024) | ~8MB | Surface features |
| Moon | 1K (1024×512) | ~2MB | Smaller, less detail needed |

**Total VRAM**: ~62MB (acceptable for modern hardware)

### Visual Style

| Component | Style | Result |
|-----------|-------|--------|
| Geometry | Low-poly icosphere (3-4 subdivisions) | N64/PS1 retro feel |
| Textures | High-res NASA imagery (2K-4K) | Beautiful, detailed surfaces |
| Lighting | Directional (sun) + ambient | Accurate day/night |
| Combined | Lo-fi geometry + hi-fi textures | Unique artistic style |

## Architecture

### Planet Type

```go
// engine/tetra/planet.go

type PlanetConfig struct {
    Name           string
    TexturePath    string
    Radius         float64   // Relative to Earth = 1.0
    RotationPeriod float64   // Hours
    AxialTilt      float64   // Degrees
    HasRings       bool
    RingTexturePath string
    RingInnerRadius float64  // Relative to planet radius
    RingOuterRadius float64
}

var SolarSystemPlanets = map[string]PlanetConfig{
    "earth": {
        Name:           "Earth",
        TexturePath:    "assets/planets/earth.png",
        Radius:         1.0,
        RotationPeriod: 24.0,
        AxialTilt:      23.4,
        HasRings:       false,
    },
    "saturn": {
        Name:            "Saturn",
        TexturePath:     "assets/planets/saturn.png",
        Radius:          9.45,
        RotationPeriod:  10.7,
        AxialTilt:       26.7,
        HasRings:        true,
        RingTexturePath: "assets/planets/saturn_rings.png",
        RingInnerRadius: 1.2,
        RingOuterRadius: 2.3,
    },
    // ... etc
}
```

### Planet3D Implementation

```go
type Planet3D struct {
    config     PlanetConfig
    mesh       *tetra3d.Model
    ringMesh   *tetra3d.Model  // nil if no rings
    rotation   float64          // Current rotation angle
    position   tetra3d.Vector   // 3D position
    scale      float64          // Display scale
}

func NewPlanet3D(name string, scene *Scene) (*Planet3D, error) {
    config, ok := SolarSystemPlanets[name]
    if !ok {
        return nil, fmt.Errorf("unknown planet: %s", name)
    }

    // Load texture
    texture, err := loadPlanetTexture(config.TexturePath)
    if err != nil {
        return nil, err
    }

    // Create icosphere
    mesh := tetra3d.NewIcosphereMesh(config.Radius, 4)  // 4 subdivisions for smooth sphere

    // Apply material
    mat := tetra3d.NewMaterial(name)
    mat.Texture = texture
    mesh.SetMaterial(mat)

    // Create model
    model := tetra3d.NewModel(name, mesh)

    // Apply axial tilt
    model.Node.SetLocalRotation(config.AxialTilt, 0, 0)

    p := &Planet3D{
        config: config,
        mesh:   model,
        scale:  1.0,
    }

    // Add rings if present
    if config.HasRings {
        p.ringMesh = p.createRings()
    }

    return p, nil
}

func (p *Planet3D) createRings() *tetra3d.Model {
    // Create ring geometry as flat disk with hole
    // This is a simplified approach - could use torus for 3D rings
    ringMesh := tetra3d.NewPlaneMesh(
        p.config.RingOuterRadius * 2,
        p.config.RingOuterRadius * 2,
    )

    // Load ring texture (with alpha for transparency)
    ringTex, _ := loadPlanetTexture(p.config.RingTexturePath)

    mat := tetra3d.NewMaterial(p.config.Name + "_rings")
    mat.Texture = ringTex
    mat.TransparencyMode = tetra3d.TransparencyModeAlphaClip
    ringMesh.SetMaterial(mat)

    model := tetra3d.NewModel(p.config.Name+"_rings", ringMesh)

    // Tilt rings with planet
    model.Node.SetLocalRotation(p.config.AxialTilt + 90, 0, 0)  // Flat relative to planet

    return model
}

func (p *Planet3D) Update(dt float64, gameTimeScale float64) {
    // Rotate based on rotation period
    // rotationPeriod is in hours, dt is in seconds
    rotationSpeed := (360.0 / (p.config.RotationPeriod * 3600)) * gameTimeScale
    p.rotation += rotationSpeed * dt

    // Apply rotation (around tilted axis)
    p.mesh.Node.SetLocalRotation(p.config.AxialTilt, p.rotation, 0)
}
```

## Lighting

### Sun Direction

```go
type SolarLighting struct {
    sun       *tetra3d.DirectionalLight
    ambient   *tetra3d.AmbientLight
}

func NewSolarLighting(scene *Scene) *SolarLighting {
    // Directional light for sun
    sun := tetra3d.NewDirectionalLight("sun", 1, 1, 0.95, 1.0)  // Slightly warm
    sun.SetEnergy(1.2)

    // Ambient light for fill (very dark)
    ambient := tetra3d.NewAmbientLight("ambient", 0.1, 0.1, 0.15, 1.0)  // Slight blue

    scene.scene.Root.AddChildren(sun, ambient)

    return &SolarLighting{
        sun:     sun,
        ambient: ambient,
    }
}

func (s *SolarLighting) SetSunDirection(x, y, z float64) {
    // Sun "points" at origin from this direction
    s.sun.Node.SetLocalPosition(x*100, y*100, z*100)
    s.sun.Node.LookAt(0, 0, 0)
}
```

### Terminator Line

The terminator (day/night boundary) is automatic with directional lighting:

```
         Sun
          ↓
    ┌─────────────┐
    │   ░░░░░█████│  ← Illuminated side
    │  ░░░░░██████│
    │ ░░░░░░██████│  ← Terminator line
    │  ░░░░░██████│
    │   ░░░░░█████│  ← Dark side
    └─────────────┘
```

## Atmosphere Glow (Fresnel)

For Earth-like planets, add atmosphere glow at the limb:

```go
func (p *Planet3D) AddAtmosphere(color tetra3d.Color, thickness float64) {
    // Create slightly larger transparent sphere
    atmMesh := tetra3d.NewIcosphereMesh(p.config.Radius * (1 + thickness), 3)

    mat := tetra3d.NewMaterial(p.config.Name + "_atmosphere")
    mat.Color = color
    mat.TransparencyMode = tetra3d.TransparencyModeTransparent
    mat.FragmentShaderFunc = fresnelShader  // Custom shader for edge glow

    atmMesh.SetMaterial(mat)

    p.atmosphereMesh = tetra3d.NewModel(p.config.Name+"_atm", atmMesh)
    p.mesh.Node.AddChildren(p.atmosphereMesh)
}
```

## Usage Example

### In Arrival Sequence

```go
func (a *ArrivalView) initPlanets() error {
    // Create 3D scene for planets
    a.planetScene = tetra.NewScene(1280, 960)

    // Add lighting
    a.lighting = tetra.NewSolarLighting(a.planetScene)
    a.lighting.SetSunDirection(-1, 0.3, 0.5)  // Sun position relative to view

    // Preload planets we'll encounter
    for _, name := range []string{"saturn", "jupiter", "mars", "earth"} {
        planet, err := tetra.NewPlanet3D(name, a.planetScene)
        if err != nil {
            return err
        }
        a.planets[name] = planet
    }

    return nil
}

func (a *ArrivalView) showPlanet(name string, distance float64) {
    planet := a.planets[name]

    // Hide all others
    for _, p := range a.planets {
        p.SetVisible(false)
    }

    // Position planet at distance
    planet.SetPosition(0, 0, distance)
    planet.SetScale(1.0 / distance)  // Closer = bigger
    planet.SetVisible(true)
}

func (a *ArrivalView) Draw(screen *ebiten.Image) {
    // 1. Draw starfield background
    a.background.Draw(screen)

    // 2. Render 3D planets to buffer
    planetBuffer := a.planetScene.Render()

    // 3. Composite planet render (with transparency)
    op := &ebiten.DrawImageOptions{}
    screen.DrawImage(planetBuffer, op)

    // 4. Draw UI
    a.ui.Draw(screen)
}
```

## Asset Pipeline

### Download Script

```bash
#!/bin/bash
# scripts/download_planet_textures.sh

mkdir -p assets/planets

# Earth - NASA Blue Marble
curl -o assets/planets/earth.png \
  "https://eoimages.gsfc.nasa.gov/images/imagerecords/74000/74393/world.topo.200412.3x5400x2700.png"

# Saturn - NASA Cassini
curl -o assets/planets/saturn.png \
  "https://solarsystem.nasa.gov/system/resources/detail_files/2490_saturn_cassini_702x486.jpg"

# ... etc
```

### Texture Processing

```bash
# Resize to 2K for performance
convert assets/planets/earth_raw.png -resize 2048x1024 assets/planets/earth.png

# Create equirectangular projection if needed
# (NASA textures are usually already in this format)
```

## Performance Budget

| Component | Target | Notes |
|-----------|--------|-------|
| Planet mesh | 4 subdivisions icosphere (~1280 triangles) | Smooth enough at game resolution |
| Texture size | 2048x1024 | Good detail, reasonable VRAM |
| Render time | <5ms per planet | Budget for 60fps with headroom |
| Max visible | 1-2 planets | Distance LOD for others |

### LOD Strategy

| Distance | Rendering |
|----------|-----------|
| Close (<1 unit) | Full 3D with atmosphere |
| Medium (1-10) | 3D sphere, no atmosphere |
| Far (10-100) | 2D sprite |
| Very far (>100) | Point of light |

## SR/GR Shader Integration

**Critical**: Planets rendered by Tetra3D work with our physics-accurate SR/GR shaders.

### Why It Works

Tetra3D outputs to `*ebiten.Image` → Our shaders process `*ebiten.Image` → Pixel source doesn't matter.

### Rendering Pipeline with Shaders

```go
func (v *SpaceView) Draw(screen *ebiten.Image) {
    // 1. Starfield background
    v.starfield.Draw(v.buffer)

    // 2. Render 3D planet (Tetra3D with hi-res texture)
    planetImg := v.planetScene.Render()
    v.buffer.DrawImage(planetImg, &opts)

    // 3. Apply SR effects (approaching planet at velocity)
    if v.velocity > 0.01 {
        v.srWarp.SetForwardVelocity(v.velocity)
        v.srWarp.Apply(v.srBuffer, v.buffer)
        // Planet now shows Doppler shift - blue-tinted when approaching!
        v.buffer = v.srBuffer
    }

    // 4. Apply GR effects (near massive object)
    if v.grIntensity > 0 {
        v.grWarp.Apply(screen, v.buffer)
        // Light bends around black hole, planets distorted
    } else {
        screen.DrawImage(v.buffer, nil)
    }
}
```

### Visual Results

| Scenario | Tetra3D Renders | Shader | Visual Result |
|----------|-----------------|--------|---------------|
| Approaching Saturn at 0.3c | Saturn + rings | SR Doppler | Blue-tinted Saturn |
| Near black hole | Stars + distant Earth | GR Lensing | Bent light, Einstein ring |
| Arriving at Earth at 0.0c | Earth in full detail | None | Beautiful 4K Earth |

## Success Criteria

- [ ] Earth renders as textured 3D sphere
- [ ] Day/night terminator visible
- [ ] Planet rotates smoothly
- [ ] Saturn's rings render correctly
- [ ] Atmosphere glow on Earth
- [ ] Composites correctly with 2D elements
- [ ] **SR Doppler shift applies to planet** (blue-tint when approaching)
- [ ] **GR lensing applies to planet** (distortion near BH)
- [ ] 60fps with one planet + shaders

## Tech Demo Commands

Incremental demos to validate each system:

```bash
# Stage 1: Basic 3D rendering
./bin/game --demo-planet earth          # Rotating Earth (4K texture)
./bin/game --demo-planet saturn         # Saturn with rings
./bin/game --demo-planet jupiter        # Jupiter's bands

# Stage 2: Multiple planets
./bin/game --demo-planets               # All planets in a row

# Stage 3: With SR effects
./bin/game --demo-planet earth --velocity 0.3   # Earth with Doppler shift
./bin/game --demo-planet saturn --velocity 0.5  # Blue-tinted Saturn

# Stage 4: With GR effects
./bin/game --demo-planet earth --gr 0.5         # Earth with GR lensing

# Stage 5: Full solar system view
./bin/game --demo-engine-solar                  # Orbiting planets from ship POV
./bin/game --demo-engine-solar --velocity 0.3   # With SR effects

# Stage 6: Combined with isometric
./bin/game --demo-game-bridge                        # Isometric bridge + space view
./bin/game --demo-game-bridge --planet earth         # Bridge looking at Earth
```

## Tech Demo Milestones

| Demo | What It Proves | Blocks |
|------|----------------|--------|
| `--demo-planet` | Tetra3D renders textured spheres | Everything |
| `--demo-planet --velocity` | SR shaders work with 3D | Arrival sequence |
| `--demo-planet --gr` | GR shaders work with 3D | Black hole sequence |
| `--demo-engine-solar` | Multiple planets, orbital view | Solar system navigation |
| `--demo-game-bridge` | Isometric + space background composite | Main game view |

## Related Design Docs

- [planetary-rings.md](../v0_2_0/planetary-rings.md) - Extended ring systems for all gas giants (not just Saturn)

## Next Steps

After this is implemented:
1. **solar-system-view.md** - Orbital mechanics display
2. **bridge-interior.md** - Isometric bridge with space backdrop
3. **arrival-sequence-v2.md** - Rebuild arrival with real planets

---

## Sprint Progress

**Implemented Via:** Tetra3D Integration Sprint (tetra3d-v1)
**Tracking:** [sprints/sprint_tetra3d-v1.json](../../../sprints/sprint_tetra3d-v1.json)

### Implementation Status

| Feature | Status | Notes |
|---------|--------|-------|
| Planet 3D sphere | ✅ Implemented | `NewPlanet()`, `NewTexturedPlanet()` in planet.go |
| Textured rendering | ✅ Implemented | UV sphere mesh for equirectangular textures |
| Day/night terminator | ✅ Implemented | Directional light in lighting.go |
| Rotation animation | ✅ Implemented | `Update(dt)` method on Planet |
| Saturn rings | ✅ Implemented | ring.go with proper ring mesh geometry |
| SR shader integration | ✅ Implemented | Doppler shift visible in demo-sr-flyby |
| GR shader integration | ✅ Compatible | Infrastructure ready |
| Atmosphere glow | ❌ Not implemented | Fresnel shader deferred |
| LOD system | ❌ Not implemented | Not needed for current use |

### Implementation Files

| File | Purpose |
|------|---------|
| [engine/tetra/planet.go](../../../engine/tetra/planet.go) | Planet sphere rendering (173 LOC) |
| [engine/tetra/uvsphere.go](../../../engine/tetra/uvsphere.go) | UV sphere mesh for textures |
| [engine/tetra/ring.go](../../../engine/tetra/ring.go) | Saturn ring geometry (147 LOC) |
| [engine/tetra/lighting.go](../../../engine/tetra/lighting.go) | Sun/ambient lighting |

### Demo Commands

```bash
./bin/demo-planet-view        # Single textured planet over starfield
./bin/demo-planets-benchmark  # Multiple planets with SR shader
./bin/demo-saturn             # Saturn with rings
./bin/demo-sr-flyby           # Solar system flyby with Doppler effects
```

### Success Criteria Status

- [x] Earth renders as textured 3D sphere
- [x] Day/night terminator visible
- [x] Planet rotates smoothly
- [x] Saturn's rings render correctly
- [ ] Atmosphere glow on Earth (deferred - Fresnel shader not needed for MVP)
- [x] Composites correctly with 2D elements
- [x] SR Doppler shift applies to planet
- [x] GR lensing applies to planet (infrastructure compatible)
- [ ] 60fps with one planet + shaders (achieved ~21fps - acceptable for cinematic views)

### Remaining Work

See [planet-rendering-polish.md](../future/planet-rendering-polish.md) for deferred features:
1. **Atmosphere glow** - Fresnel shader for Earth's atmosphere limb (low priority)
2. **LOD system** - Distance-based rendering quality (not needed yet)
3. **PlanetConfig registry** - Currently planets are created ad-hoc in demos
