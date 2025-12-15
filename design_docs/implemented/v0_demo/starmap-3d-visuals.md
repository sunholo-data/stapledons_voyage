# 3D Starmap Visuals Demo

**Status:** Implemented  
**Location:** `cmd/demo-starmap-visuals/`  
**Date:** 2025-12-15

## Overview

A 3D starmap visualization demo that renders **3,802 real stars** from the CNS5 catalog with proper spectral colors, 3D textured spheres, and dynamic lighting. This demo tests the performance limits of the engine's LOD system and 3D rendering capabilities.

## Implementation

### Star Data
- **Source:** `assets/data/starmap/stars.json` (CNS5 catalog)
- **Count:** 3,802 stars within ~35 light-years of Sol
- **Fields:** id, name, x/y/z position (ly), distance (ly), visual magnitude, spectral type

### Spectral Type Mapping

| Type | Color (RGBA) | Luminosity | Radius | Temperature |
|------|--------------|------------|--------|-------------|
| O | Blue-white (155,176,255) | 30,000x | 8.0x | 33,000K |
| B | Blue-white (170,191,255) | 1,000x | 4.0x | 15,000K |
| A | White (202,215,255) | 20x | 2.0x | 8,500K |
| F | Yellow-white (248,247,255) | 3x | 1.4x | 6,500K |
| G | Yellow (255,244,234) | 1x | 1.0x | 5,800K |
| K | Orange (255,210,161) | 0.4x | 0.8x | 4,500K |
| M | Red (255,180,100) | 0.04x | 0.3x | 3,200K |

### LOD System Integration

Stars are rendered using the 5-tier LOD system from `engine/lod/`:

1. **Full3D** (closest) - Textured 3D spheres using Tetra3D
2. **Billboard** - 2D sprites facing camera
3. **Circle** - Filled circles
4. **Point** - Single pixels
5. **Culled** (furthest) - Not rendered

### Lighting
- Dynamic star lights created for brightest stars (by visual magnitude)
- Configurable max lights (default: 10) for performance
- Spectral-accurate light colors
- Ambient light for deep space (0.05, 0.05, 0.08)

### Key Files

| File | Purpose |
|------|---------|
| `cmd/demo-starmap-visuals/main.go` | Demo entry point |
| `assets/data/starmap/stars.json` | CNS5 star catalog |
| `assets/stars/sun_8k.jpg` | 8K star texture (CC BY 4.0) |
| `assets/stars/SOURCES.md` | Texture source documentation |
| `engine/lod/` | LOD management system |
| `engine/tetra/` | 3D rendering (Planet, StarLight) |

### New API Methods Added

```go
// engine/tetra/planet.go
func (p *Planet) SetColorModulation(r, g, b float64)

// engine/tetra/lighting.go
func (s *StarLight) Name() string
```

## Performance Results

### Expanded LOD (8K texture, 15 lights, 100 max 3D)
- **FPS:** 392 (uncapped)
- **LOD Distribution:** 100 Full3D, 662 Billboard, 94 Circle, 1 Point, 2945 Culled
- **Texture:** 8K sun texture (3.5MB) with spectral color modulation
- **All 3,802 stars visible**

### Default LOD (original settings)
- **FPS:** 60+ (vsync limited)
- **LOD Distribution:** 2 Full3D, 160 Billboard, 397 Circle, 68 Point, 3175 Culled
- **Memory:** Efficient - billboards created lazily

## Star Textures

### Downloaded Textures (`assets/stars/`)
| File | Resolution | Size | Source | License |
|------|------------|------|--------|---------|
| `sun_8k.jpg` | 8192x4096 | 3.5MB | Solar System Scope | CC BY 4.0 |
| `sun_2k.jpg` | 2048x1024 | 803KB | Solar System Scope | CC BY 4.0 |

### Texture Sources (see `assets/stars/SOURCES.md`)
- **[Solar System Scope](https://www.solarsystemscope.com/textures/)** - CC BY 4.0, commercial OK
- **[Space Spheremaps](https://www.spacespheremaps.com/)** - Free commercial use
- **[FarGetaNik DeviantArt](https://www.deviantart.com/fargetanik)** - CC BY-NC-SA 3.0 (prototyping only)
- **[NASA SVS](https://svs.gsfc.nasa.gov/)** - Public domain

## Usage

```bash
# Run with all stars
go run ./cmd/demo-starmap-visuals

# Limit stars for testing
go run ./cmd/demo-starmap-visuals --max-stars 100 --max-lights 5

# Take screenshot
go run ./cmd/demo-starmap-visuals --screenshot 60 --output out/starmap.png

# Adjust scale
go run ./cmd/demo-starmap-visuals --scale 2.0
```

### Controls

| Key | Action |
|-----|--------|
| WASD/Arrows | Move camera |
| Q/E | Up/Down |
| Shift | Fast movement |
| R | Reset to Sol |
| N | Find nearest star |
| G | Toggle orientation grid |
| [ / ] | Adjust star light intensity |
| ; / ' | Adjust ambient light |
| L | Reset lighting |

### Orientation Grid

The demo includes a 3D orientation grid centered on Sol (origin):

- **X axis (red)** - Points toward galactic center
- **Y axis (green)** - Points in direction of galactic rotation
- **Z axis (blue)** - Points toward north galactic pole
- **Distance rings** - Circles at 5, 10, 15, 20, 25, 30 ly on the galactic plane (XY)
- **Sol marker** - Small cross at origin

Press **G** to toggle grid visibility.

## Future Improvements

### High Priority
- [x] **Star textures downloaded** - 8K sun texture from Solar System Scope (CC BY 4.0)
- [ ] **Spectral-specific textures** - Different base textures for O/B/A vs M-type (currently using color modulation)
- [ ] **Star glow/bloom** - Add glow shader effect for bright stars
- [ ] **Realistic star sizes** - Stars appear as points at any realistic scale; need artistic enlargement
- [ ] **Famous star names** - Show proper names (Sirius, Alpha Centauri) instead of Gliese IDs

### Medium Priority
- [ ] **Constellations** - Draw constellation lines when close to Sol
- [ ] **Search/navigate** - Jump to named stars
- [ ] **Habitable zone indicators** - Highlight stars with habitable planets
- [ ] **Distance ruler** - Show light-year scale
- [ ] **Star info popup** - Click star to see details

### Low Priority (Performance Testing)
- [ ] **10,000+ stars** - Test with extended catalog
- [ ] **50+ lights** - Push lighting limits
- [ ] **Shader-based stars** - GPU instancing for massive star counts
- [ ] **Octree spatial queries** - For nearest-star searches

### AILANG Integration
- [ ] **Connect to sim/starmap.ail** - Use AILANG for star queries and navigation logic
- [ ] **Event system** - AILANG handles star selection, journey planning
- [ ] **Time dilation preview** - Show relativistic effects for planned journeys

## Related Files

- `sim/starmap.ail` - AILANG star data types and queries
- `sim/starmap_visuals.ail` - AILANG starmap rendering (2D, uses StarDrawCmd)
- `cmd/demo-lod/main.go` - Reference implementation for LOD system

## Screenshots

![Starmap with 3,802 stars - default LOD](../../../out/screenshots/starmap-visuals-full.png)
![Starmap with expanded LOD - 8K texture](../../../out/screenshots/starmap-8k-texture.png)
