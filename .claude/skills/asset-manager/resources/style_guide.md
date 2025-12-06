# Stapledon's Voyage Art Style Guide

Consistent visual language for all game assets.

## Visual Identity

**Theme**: Hard sci-fi with retro pixel art aesthetic
**Mood**: Vast, contemplative, cosmic scale with intimate human elements
**Inspiration**: Classic sci-fi book covers, 16-bit era space games, NASA imagery

## Color Palette

### Primary Colors (Deep Space)
| Name | Hex | Use |
|------|-----|-----|
| Void Black | `#1a1a2e` | Space backgrounds |
| Nebula Purple | `#4a4e69` | Distant nebulae |
| Star White | `#edf2f4` | Bright stars |
| Engine Orange | `#f77f00` | Ship engines, highlights |

### Secondary Colors (Planets & Life)
| Name | Hex | Use |
|------|-----|-----|
| Earth Blue | `#4895ef` | Water, ice worlds |
| Forest Green | `#2d6a4f` | Vegetation, habitable |
| Desert Tan | `#d4a373` | Arid worlds |
| Lava Red | `#9d0208` | Volcanic activity |

### UI Colors
| Name | Hex | Use |
|------|-----|-----|
| Panel Dark | `#16213e` | UI backgrounds |
| Panel Light | `#1f4068` | UI highlights |
| Text Primary | `#e8e8e8` | Main text |
| Text Secondary | `#a0a0a0` | Subdued text |
| Accent Cyan | `#00fff5` | Interactive elements |

### Spectral Star Colors
| Class | Hex | Temperature |
|-------|-----|-------------|
| O | `#9bb0ff` | >30,000K (blue) |
| B | `#aabfff` | 10,000-30,000K |
| A | `#cad7ff` | 7,500-10,000K |
| F | `#f8f7ff` | 6,000-7,500K |
| G | `#fff4ea` | 5,200-6,000K (Sun) |
| K | `#ffd2a1` | 3,700-5,200K |
| M | `#ffcc6f` | <3,700K (red dwarf) |

## Pixel Art Guidelines

### Resolution & Scale
- **Base unit**: 1 pixel = 1 unit (no sub-pixel rendering)
- **Tile size**: 64x32 isometric (2:1 ratio)
- **Entity size**: 32x48 per frame (4 frames = 128x48 sheet)
- **Star sprites**: 16x16 with glow
- **Portraits**: 128x128

### Technique
1. **Outlines**: Dark outlines (1px) on sprites for readability
2. **Dithering**: Use sparingly for gradients (2-3 colors max)
3. **Anti-aliasing**: Manual AA only on curves, limited colors
4. **Shading**: Light source from top-left (consistent across all assets)

### Animation
- **Walk cycles**: 4 frames (right foot, passing, left foot, passing)
- **Idle**: 1 frame (or 2 for breathing effect)
- **Frame rate**: 6 FPS for walks, 2 FPS for idle breathing

## Isometric Rules

### Grid System
```
    /\
   /  \      Width: 64px
  /    \     Height: 32px
  \    /     Angle: ~26.57 degrees (arctan 0.5)
   \  /
    \/
```

### Tile Alignment
- Tiles must align perfectly on the isometric grid
- No partial pixels at edges
- Transparent background (alpha channel)

### Depth Ordering
- Objects further "back" (higher Y in world) render first
- Tall objects may need anchor point adjustment

### Entity Placement
- Entities stand ON tiles, not inside them
- Foot placement at tile center
- Height extends upward from tile surface

## Asset-Specific Guidelines

### Isometric Tiles (64x32)
- Diamond-shaped visible area
- Corners must be transparent
- Consider tile connectivity (edges should blend)
- Variations for natural look (grass_1, grass_2, grass_3)

### Entity Sprites (32x48 per frame)
- Character fits within frame with padding
- Consistent shadow direction
- Clear silhouette for readability
- Animation frames left-to-right in sheet

### Star Sprites (16x16)
- Soft glow effect at edges
- Color matches spectral class
- Center brightest, fades outward
- Variations for size (dwarf, giant)

### Planet Sprites (256x256)
- Spherical with atmosphere glow
- Terminator line (day/night boundary)
- Surface detail visible
- Reference NASA imagery for realism

### Portraits (128x128)
- Face-forward or 3/4 view
- Consistent lighting (top-left)
- Background: transparent or solid color
- Expression readable at small sizes

### Backgrounds (1920x1080+)
- Star fields: varying density, no patterns
- Nebulae: soft gradients, not too busy
- Galaxy: spiral structure visible
- Should not distract from foreground

## Prompt Engineering for AI Generation

### Style Keywords (always include)
```
pixel art, retro gaming aesthetic, 16-bit style, limited color palette,
clear outlines, no anti-aliasing, crisp pixels
```

### Isometric Keywords
```
isometric view, 2:1 ratio, diamond tile, top-down angled,
dimetric projection, 26 degree angle
```

### Sci-Fi Keywords
```
hard science fiction, realistic space, cosmic scale,
NASA-inspired, scientific accuracy, deep space
```

### Avoid
- "Smooth gradients" (use dithering instead)
- "Photorealistic" (we want pixel art)
- "3D rendered" (2D pixel aesthetic)
- "High resolution" (we want low-res pixel look)
- Busy, cluttered compositions

## Quality Checklist

Before finalizing any asset:

- [ ] Correct dimensions for asset type
- [ ] Transparent background (where required)
- [ ] Consistent with color palette
- [ ] Light source from top-left
- [ ] Clean pixel edges (no blur)
- [ ] Works at 1x scale (no scaling artifacts)
- [ ] Readable silhouette
- [ ] Fits game's visual language
