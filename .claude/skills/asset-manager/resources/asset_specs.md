# Asset Specifications

Detailed technical specifications for all game asset types.

## Directory Structure

```
assets/
├── sprites/
│   ├── manifest.json           # Sprite registry
│   ├── iso_tiles/              # Isometric terrain tiles
│   │   ├── water.png
│   │   ├── forest.png
│   │   └── ...
│   ├── iso_entities/           # Characters and objects
│   │   ├── player.png
│   │   ├── npc_red.png
│   │   └── ...
│   ├── stars/                  # Star sprites by spectral class
│   │   ├── star_blue.png
│   │   ├── star_yellow.png
│   │   └── ...
│   ├── planets/                # Planet renders
│   │   └── ...
│   ├── ui/                     # UI elements
│   │   └── ...
│   └── portraits/              # Character portraits
│       └── ...
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
      "height": 32,
      "type": "tile|entity|star|ui|planet|portrait",
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

### Tiles (1-99)
| ID | Name | File | Status |
|----|------|------|--------|
| 1 | Water | `iso_tiles/water.png` | Exists |
| 2 | Forest | `iso_tiles/forest.png` | Exists |
| 3 | Desert | `iso_tiles/desert.png` | Exists |
| 4 | Mountain | `iso_tiles/mountain.png` | Exists |
| 5-9 | Reserved | - | Planned |
| 10-19 | Alien biomes | - | Future |
| 20-29 | Space station | - | Future |
| 30-39 | Ship interior | - | Future |

### Entities (100-199)
| ID | Name | File | Status |
|----|------|------|--------|
| 100 | NPC Red | `iso_entities/npc_red.png` | Exists |
| 101 | NPC Green | `iso_entities/npc_green.png` | Exists |
| 102 | NPC Blue | `iso_entities/npc_blue.png` | Exists |
| 103 | NPC Yellow | `iso_entities/npc_yellow.png` | Exists |
| 104 | NPC Purple | `iso_entities/npc_purple.png` | Exists |
| 105 | Player | `iso_entities/player.png` | Exists |
| 106-119 | Crew members | - | Planned |
| 120-139 | Aliens | - | Future |
| 140-159 | Robots/Drones | - | Future |

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

### Planets (400-499)
| ID | Name | File | Status |
|----|------|------|--------|
| 400-409 | Rocky planets | - | Planned |
| 410-419 | Gas giants | - | Planned |
| 420-429 | Ice worlds | - | Planned |
| 430-439 | Habitable | - | Planned |

### Ships (500-599)
| ID | Name | File | Status |
|----|------|------|--------|
| 500 | Player ship | - | Planned |
| 501-509 | Ship variants | - | Future |
| 510-519 | Alien vessels | - | Future |

### Portraits (600-699)
| ID | Name | File | Status |
|----|------|------|--------|
| 600-609 | Crew portraits | - | Planned |
| 610-619 | Alien portraits | - | Future |
| 620-629 | Historical figures | - | Future |

## Dimension Specifications

### Isometric Tiles
- **Dimensions**: 64x32 pixels
- **Format**: PNG with alpha channel
- **Shape**: Diamond (isometric rhombus)
- **Anchor**: Center-bottom of diamond

```
Pixel layout:
        0         32        64
   0    ........XX........
        ......XXXXXX......
        ....XXXXXXXXXX....
        ..XXXXXXXXXXXXXX..
  16    XXXXXXXXXXXXXXXXXX
        ..XXXXXXXXXXXXXX..
        ....XXXXXXXXXX....
        ......XXXXXX......
  32    ........XX........
```

### Entity Sprites
- **Single frame**: 32x48 pixels
- **Sprite sheet**: 128x48 pixels (4 frames)
- **Format**: PNG with alpha channel
- **Anchor**: Center-bottom (feet position)

```
Frame layout (128x48 sprite sheet):
+--------+--------+--------+--------+
| Frame0 | Frame1 | Frame2 | Frame3 |
| 32x48  | 32x48  | 32x48  | 32x48  |
+--------+--------+--------+--------+
  idle     walk1    walk2    walk3
```

### Star Sprites
- **Dimensions**: 16x16 pixels
- **Format**: PNG with alpha channel
- **Style**: Soft glow, brightest at center

### Planet Sprites
- **Dimensions**: 256x256 pixels
- **Format**: PNG with alpha channel
- **Style**: Spherical, with atmosphere rim

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
iso_tiles/water.png
iso_tiles/grass_1.png
iso_tiles/grass_2.png
iso_entities/npc_red.png
iso_entities/alien_trader.png
stars/star_blue.png
stars/star_yellow_giant.png
planets/rocky_desert.png
portraits/captain_chen.png
```

### Rules
1. All lowercase
2. Underscores for spaces
3. No special characters
4. Descriptive but concise
5. Variants numbered (grass_1, grass_2)

## Animation Specifications

### Standard Animations
| Name | Frames | FPS | Loop |
|------|--------|-----|------|
| idle | 1-2 | 2 | Yes |
| walk | 4 | 6 | Yes |
| action | 3-4 | 8 | No |

### Frame Order
- Walk: right-foot-forward, passing, left-foot-forward, passing
- All frames face same direction (flip for opposite)

## Color Depth

- **Sprites**: 32-bit RGBA (8 bits per channel)
- **Backgrounds**: 24-bit RGB (JPG compatible)
- **Working palette**: Limit to 16-32 colors per sprite for pixel art look
