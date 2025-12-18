# Asset Specifications

Detailed technical specifications for all game asset types.

## Directory Structure

```
assets/
├── sprites/
│   ├── manifest.json           # Sprite registry
│   ├── stars/                  # Star sprites by spectral class
│   │   ├── star_blue.png
│   │   ├── star_yellow.png
│   │   └── ...
│   ├── ui/                     # UI elements
│   │   └── ...
│   ├── portraits/              # Character portraits
│   │   └── ...
│   └── billboards/             # 3D billboard sprites
│       └── ...
├── planets/                    # 3D planet textures (equirectangular)
│   ├── earth_daymap.jpg
│   ├── jupiter.jpg
│   └── ...
├── textures/                   # 3D interior textures
│   ├── floor_metal.png
│   ├── wall_panel.png
│   └── ...
├── sounds/
│   └── manifest.json
├── fonts/
│   └── manifest.json
├── data/
│   └── starmap/
│       ├── background/         # Galaxy backgrounds
│       │   └── galaxy_4k.jpg
│       ├── stars.json
│       └── exoplanets.json
└── generated/                  # AI-generated staging area
    └── response_*.png
```

## Sprite Manifest Schema

The `manifest.json` file registers all sprites for the engine:

```json
{
  "sprites": {
    "<id>": {
      "file": "relative/path.png",
      "width": 64,
      "height": 64,
      "type": "star|ui|portrait|billboard",
      "frameWidth": 32,       // Optional: for animated sprites
      "frameHeight": 48,      // Optional: for animated sprites
      "animations": {         // Optional: animation definitions
        "idle": {"startFrame": 0, "frameCount": 1, "fps": 0},
        "walk": {"startFrame": 0, "frameCount": 4, "fps": 6}
      }
    }
  }
}
```

## Sprite ID Allocation

### Stars (200-299)
| ID | Name | File | Status |
|----|------|------|--------|
| 200 | Star Blue (O/B) | `stars/star_blue.png` | Exists |
| 201 | Star White (A/F) | `stars/star_white.png` | Exists |
| 202 | Star Yellow (G) | `stars/star_yellow.png` | Exists |
| 203 | Star Orange (K) | `stars/star_orange.png` | Exists |
| 204 | Star Red (M) | `stars/star_red.png` | Exists |
| 205-209 | Star size variants | - | Planned |
| 210-219 | Binary systems | - | Future |
| 220-229 | Exotic (neutron, etc) | - | Future |

### UI Elements (300-399)
| ID | Name | File | Status |
|----|------|------|--------|
| 300-309 | Buttons | - | Planned |
| 310-319 | Panels | - | Planned |
| 320-329 | Icons | - | Planned |
| 330-339 | Cursors | - | Planned |

### Billboards (400-499)
| ID | Name | File | Status |
|----|------|------|--------|
| 400-449 | Character billboards | - | Planned |
| 450-499 | Object billboards | - | Future |

### Portraits (600-699)
| ID | Name | File | Status |
|----|------|------|--------|
| 600-609 | Crew portraits | - | Planned |
| 610-619 | Alien portraits | - | Future |
| 620-629 | Historical figures | - | Future |

## Dimension Specifications

### Star Sprites
- **Dimensions**: 16x16 pixels
- **Format**: PNG with alpha channel
- **Style**: Soft glow, brightest at center

### Billboard Sprites
- **Dimensions**: 256x256 pixels (or 128x128 for smaller objects)
- **Format**: PNG with alpha channel
- **Style**: Character/object visible from any angle
- **Anchor**: Center-bottom for characters

### Portraits
- **Dimensions**: 128x128 pixels
- **Format**: PNG (transparent or solid bg)
- **Style**: Face-focused, readable at 64x64

### Backgrounds
- **Minimum**: 1920x1080 pixels
- **Preferred**: 3840x2160 (4K)
- **Format**: JPG (lossy OK for backgrounds)

## File Naming Conventions

### Pattern
`<type>_<name>[_<variant>].png`

### Examples
```
stars/star_blue.png
stars/star_yellow_giant.png
billboards/crew_pilot.png
portraits/captain_chen.png
```

### Rules
1. All lowercase
2. Underscores for spaces
3. No special characters
4. Descriptive but concise
5. Variants numbered (variant_1, variant_2)

## Color Depth

- **Sprites**: 32-bit RGBA (8 bits per channel)
- **Backgrounds**: 24-bit RGB (JPG compatible)

## 3D Planet Textures

### Directory Structure
```
assets/planets/           # 3D planet textures (JPG/PNG)
├── earth.jpg            # 2K equirectangular
├── mars.jpg
├── jupiter.jpg
├── saturn.jpg
├── saturn_ring.png      # Ring texture with alpha
├── uranus.jpg
├── neptune.jpg
└── ...
```

### Texture Requirements
- **Format**: Equirectangular projection (2:1 aspect ratio)
- **Resolution**: 2048x1024 minimum (2K), 4096x2048 preferred (4K)
- **File type**: JPG for solid planets, PNG for rings (need alpha)

### Texture Sources (Creative Commons / Public Domain)

**Recommended Sources:**
1. **Solar System Scope** (CC BY 4.0)
   - URL: https://www.solarsystemscope.com/textures/
   - Quality: 2K-8K, good for planets

2. **NASA Visible Earth** (Public Domain)
   - URL: https://visibleearth.nasa.gov/
   - Quality: Very high resolution, authentic

3. **NASA JPL Photojournal** (Public Domain)
   - URL: https://photojournal.jpl.nasa.gov/
   - Quality: Mission imagery, authentic

## 3D Interior Textures

### Directory Structure
```
assets/textures/          # 3D interior textures
├── interior/            # Ship interior textures
│   ├── bridge_floor.png
│   ├── bridge_wall.png
│   └── bridge_ceiling.png
├── floor_metal.png      # Metallic floor panels
├── wall_panel.png       # Wall panels
├── ceiling_lights.png   # Ceiling with lights
├── console_screen.png   # Console displays
└── ...
```

### Texture Requirements
- **Format**: Square, power-of-2 dimensions
- **Resolution**: 256x256, 512x512, or 1024x1024
- **File type**: PNG (for alpha) or JPG (for solid)

### ⚠️ CRITICAL: Tiling Requirements

Interior floor and ceiling textures **MUST be seamless/tileable** because:
- The engine uses a 4x4 grid of tiles for floors and ceilings (16 tiles each)
- This is required to work around Tetra3D's lack of triangle clipping
- Non-seamless textures will show visible seams at tile boundaries

**Seamless Texture Checklist:**
1. ✅ Edges must wrap seamlessly (left→right, top→bottom)
2. ✅ No obvious focal points that repeat visibly
3. ✅ Pattern density should be consistent across texture
4. ✅ Test at 1:1 UV scale (1 texture = 1 meter in-game)

**UV Scale:** Default is 1.0 (one texture tile per meter). Adjust via `uvScale` in AILANG:
```ailang
-- More tiles per meter (smaller pattern, more repeats)
let room = makeRoomTexturedScale(8.0, 6.0, 3.0, floor, wall, ceil, 2.0)
```

**Wall textures** are less critical since walls use single planes, but seamless is still recommended for consistency.
