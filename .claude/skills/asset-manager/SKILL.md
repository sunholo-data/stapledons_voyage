---
name: Asset Manager
description: Generate and manage game image assets using AI. Use when user asks to create sprites, textures, backgrounds, portraits, or other visual assets for the game. Handles 3D textures, style consistency, and manifest updates. (project)
---

# Asset Manager

Generate, organize, and manage visual assets for Stapledon's Voyage using AI image generation. Ensures consistent styling, proper 3D texture formatting, and automatic manifest updates.

## Quick Start

**Most common usage:**
```bash
# Generate 3D interior textures (via voyage ai CLI directly)
bin/voyage ai -generate-image -prompt "Seamless tileable metallic spaceship floor texture..."

# Generate UI/portrait assets
.claude/skills/asset-manager/scripts/generate_asset.sh portrait "captain" "human ship captain, weathered"

# Check what assets exist
.claude/skills/asset-manager/scripts/status.sh

# Update manifest after adding assets
.claude/skills/asset-manager/scripts/update_manifest.sh
```

## When to Use This Skill

Invoke this skill when:
- User asks to "create a texture", "generate an image", or "make game art"
- User wants new textures for 3D interiors (floor, wall, ceiling)
- User wants backgrounds, portraits, or UI elements
- User asks about art style or asset specifications
- User wants to organize or catalog existing assets

## Asset Types

### 3D Interior Textures (for Tetra3D room rendering)

The game uses first-person 3D interiors rendered with Tetra3D. Textures must be seamless and tileable.

| Type | Dimensions | Location | Format | Notes |
|------|------------|----------|--------|-------|
| **Floor Texture** | 512x512+ | `assets/textures/interior/` | PNG | Top-down, seamless tile |
| **Wall Texture** | 512x512+ | `assets/textures/interior/` | PNG | Vertical orientation, seamless |
| **Ceiling Texture** | 512x512+ | `assets/textures/interior/` | PNG | Top-down with lights |
| **Console Texture** | 256x256+ | `assets/textures/interior/` | PNG | Equipment surfaces |

**Tested Prompts for Interior Textures:**

```bash
# Floor texture
bin/voyage ai -generate-image -prompt "Seamless tileable metallic spaceship floor texture, industrial grating with subtle blue LED strip lights embedded between panels, sci-fi starship interior deck plating, dark gunmetal gray steel with cyan glow accents in the seams, top-down orthographic view for 3D texture mapping, 512x512 seamless repeating tile, no perspective distortion"

# Wall texture
bin/voyage ai -generate-image -prompt "Seamless tileable spaceship interior wall panel texture, sci-fi bulkhead with recessed panel sections and subtle orange warning stripe, dark gray brushed metal with rivets and access hatches, vertical orientation for wall mapping, industrial starship corridor aesthetic, 512x512 seamless repeating tile"

# Ceiling texture
bin/voyage ai -generate-image -prompt "Seamless tileable spaceship ceiling texture, dark industrial ceiling panels with recessed fluorescent light strips glowing white, exposed conduits and pipes, ventilation grilles, sci-fi starship bridge overhead, very dark charcoal gray metal with bright white light panels, top-down view for ceiling mapping, 512x512 seamless repeating tile"
```

**Testing textures:**
```bash
go run ./cmd/demo-engine-interior \
  -floor assets/textures/interior/bridge_floor.png \
  -wall assets/textures/interior/bridge_wall.png \
  -ceiling assets/textures/interior/bridge_ceiling.png
```

### 3D Planet Textures (for Tetra3D sphere rendering)

| Type | Dimensions | Location | Format | Notes |
|------|------------|----------|--------|-------|
| **Planet Texture** | 2048x1024+ | `assets/planets/` | JPG | **Equirectangular (2:1 ratio)** |
| **Ring Texture** | 1024x64+ | `assets/planets/` | PNG with alpha | For Saturn-like rings |

**Important:** Planet textures MUST use equirectangular projection (2:1 aspect ratio) to wrap correctly on spheres.

### 2D Sprites (for UI, maps)

| Type | Dimensions | Location | Format |
|------|------------|----------|--------|
| **Star Sprite** | 16x16 px | `assets/sprites/stars/` | PNG, glow effect |
| **UI Element** | Varies | `assets/sprites/ui/` | PNG, transparent |
| **Portrait** | 128x128 px | `assets/sprites/portraits/` | PNG |
| **Background** | 1920x1080+ | `assets/data/starmap/background/` | JPG/PNG |

**Sprite Sheet Layouts for Animation:**
- **Horizontal (4x1)**: 4 frames in a row - ideal for walk cycles
- **Grid (2x2)**: 4 frames in 2x2 - ideal for directional poses

## Available Scripts

### `scripts/generate_asset.sh <type> <name> <prompt> [--dry-run]`
Generate a 2D asset using AI with proper styling. Uses `voyage ai` CLI.

```bash
# Generate portrait
.claude/skills/asset-manager/scripts/generate_asset.sh portrait captain "human ship captain, weathered, wise"

# Generate star sprite
.claude/skills/asset-manager/scripts/generate_asset.sh star blue "bright blue O-type star"

# Generate background
.claude/skills/asset-manager/scripts/generate_asset.sh background nebula "purple nebula with stars"
```

### Direct AI Generation (for 3D textures)

For 3D textures, use `voyage ai` directly with detailed prompts:

```bash
bin/voyage ai -generate-image -prompt "Your detailed prompt here"
# Output: assets/generated/response_XXXXX.png
```

Then organize:
```bash
mkdir -p assets/textures/interior
cp assets/generated/response_XXXXX.png assets/textures/interior/floor.png
```

### `scripts/status.sh`
Display current asset inventory and identify gaps.

### `scripts/download_planet_textures.sh [resolution]`
Download proper equirectangular planet textures for 3D sphere rendering.

```bash
# Download 2K textures (default, good for game use)
.claude/skills/asset-manager/scripts/download_planet_textures.sh

# Download 4K textures (high quality)
.claude/skills/asset-manager/scripts/download_planet_textures.sh 4k
```

Downloads from Solar System Scope (CC BY 4.0).

## Workflow for 3D Interior Textures

### 1. Generate Texture

```bash
bin/voyage ai -generate-image -prompt "Seamless tileable [surface type] texture, [details], sci-fi starship interior, [color palette], [orientation] view for 3D texture mapping, 512x512 seamless repeating tile"
```

Key prompt elements:
- **"Seamless tileable"** - ensures edges match
- **"512x512"** or **"1024x1024"** - specify resolution
- **"top-down view"** (floors/ceilings) or **"vertical orientation"** (walls)
- **"no perspective distortion"** - flat texture mapping
- **Color accents** - cyan LEDs, orange warning stripes for visual interest

### 2. Review Generated Image

Check `assets/generated/response_XXXXX.png`:
- Is it seamless? Check edges
- Right orientation for the surface?
- Consistent with game aesthetic?

### 3. Organize

```bash
mkdir -p assets/textures/interior
cp assets/generated/response_XXXXX.png assets/textures/interior/bridge_floor.png
```

### 4. Test In-Game

```bash
go run ./cmd/demo-engine-interior \
  -floor assets/textures/interior/bridge_floor.png \
  -screenshot 30 -output out/test.png
```

## Real-World Reference Sources

### Planets & Moons
- **NASA Image Gallery**: https://images.nasa.gov/
- **NASA Solar System**: https://solarsystem.nasa.gov/

### Galaxy & Nebula Backgrounds
- **ESA Gaia**: https://www.esa.int/gaia
- **ESO Image Archive**: https://www.eso.org/public/images/
- **Hubble Gallery**: https://hubblesite.org/images/gallery

### Star References
- Use spectral class colors (O=blue, B=blue-white, A=white, F=yellow-white, G=yellow, K=orange, M=red)

## Sprite ID Allocation

| Range | Type | Example |
|-------|------|---------|
| 200-299 | Stars | 200=blue, 201=white, 202=yellow... |
| 300-399 | UI | Reserved |
| 400-499 | Planets | Reserved |
| 500-599 | Ships | Reserved |
| 600-699 | Portraits | Reserved |

## Planet Type Guidelines

| Type | Base Color | Features | Example |
|------|------------|----------|---------|
| **Rocky** | Gray/brown | Craters, mountains | Mercury, Moon |
| **Terrestrial** | Blue/green/tan | Continents, clouds | Earth |
| **Gas Giant** | Orange/tan bands | Cloud bands, storms | Jupiter |
| **Ice Giant** | Cyan/blue | Subtle bands | Uranus, Neptune |
| **Ocean World** | Deep blue | Cloud patterns | Hypothetical |
| **Lava World** | Black + orange cracks | Magma rivers | Hypothetical |

## Deck Scene Backgrounds (with Window Compositing)

The game composites dynamic space views through deck windows. Each deck requires:
1. **background.png** - Deck interior (1344x768)
2. **window_mask_large.png** - Mask identifying window regions (white=window)

### Workflow: AI Segmentation

```bash
# 1. Generate deck scene (widescreen 1344x768)
bin/voyage ai -generate-image -aspect 16:9 -prompt \
  "Spaceship <DECK_TYPE> interior, large windows showing sky/space, sci-fi aesthetic"

# 2. Generate mask using AI segmentation (recommended)
bin/generate-window-mask-ai \
  -deck <observation|bridge|generic> \
  -overlay \
  assets/generated/response_XXX.png

# 3. Copy to final location
cp assets/generated/response_XXX.png assets/decks/<DECK>/background.png
cp assets/generated/response_XXX_mask.png assets/decks/<DECK>/window_mask_large.png

# 4. Test
go run ./cmd/game
```

### One-Command Pipeline

For fully automated generation + masking:

```bash
.claude/skills/asset-manager/scripts/generate_transparent_scene.sh \
  --deck observation \
  myobservation "large panoramic dome with views of space"
```

### AI Segmentation Benefits

- **No threshold tuning** - AI detects windows semantically
- **Accurate boundaries** - Polygon-based masks follow actual shapes
- **Deck-aware prompts** - Optimized for observation, bridge, or generic decks

### Deck Asset Structure

```
assets/decks/<DECK>/
├── background.png           # 1344x768 deck scene
├── window_mask_large.png    # Window mask (white=window, black=interior)
└── manifest.json           # Asset metadata
```

## Notes

- Generated images go to `assets/generated/` first for review
- 3D interior textures should be 512x512 or 1024x1024 for good detail
- Planet textures (3D): Must be equirectangular (2:1 ratio) for proper UV mapping
- Deck backgrounds: Must be 1344x768 (16:9 widescreen)
- Always test new assets in-game before committing
- Real-world reference data should cite sources (NASA, ESO are CC-compatible)
