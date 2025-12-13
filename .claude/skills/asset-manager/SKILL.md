---
name: Asset Manager
description: Generate and manage game image assets using AI. Use when user asks to create sprites, tiles, backgrounds, portraits, or other visual assets for the game. Handles isometric formatting, style consistency, and manifest updates. (project)
---

# Asset Manager

Generate, organize, and manage visual assets for Stapledon's Voyage using AI image generation. Ensures consistent pixel art styling, proper isometric formatting, and automatic manifest updates.

## Quick Start

**Most common usage:**
```bash
# Generate a new asset (uses voyage ai CLI under the hood)
.claude/skills/asset-manager/scripts/generate_asset.sh tile "alien_crystal" "purple crystal formations, sci-fi"

# Generate bridge-specific assets
.claude/skills/asset-manager/scripts/generate_asset.sh console "weapons" "torpedo targeting display, red warning lights"
.claude/skills/asset-manager/scripts/generate_asset.sh crew "pilot" "blue uniform pilot, 4 poses in 2x2 grid"

# Preview prompt without generating
.claude/skills/asset-manager/scripts/generate_asset.sh console "helm" "pilot controls" --dry-run

# Organize generated asset with auto-optimization
.claude/skills/asset-manager/scripts/organize_asset.sh assets/generated/response_xxx.png console helm --optimize

# Check what assets exist
.claude/skills/asset-manager/scripts/status.sh

# Update manifest after adding assets
.claude/skills/asset-manager/scripts/update_manifest.sh
```

## When to Use This Skill

Invoke this skill when:
- User asks to "create a sprite", "generate an image", or "make game art"
- User wants new tiles, entities, backgrounds, or portraits
- User asks about art style or asset specifications
- User wants to organize or catalog existing assets
- User mentions needing visuals for a new feature

## Asset Types

### 2D Sprites (for UI, maps, pixel art)

| Type | Dimensions | Location | Format |
|------|------------|----------|--------|
| **Isometric Tile** | 64x32 px | `assets/sprites/iso_tiles/` | PNG, transparent bg |
| **Entity Sprite** | 128x48 px (4 frames) | `assets/sprites/iso_entities/` | PNG, sprite sheet |
| **Star Sprite** | 16x16 px | `assets/sprites/stars/` | PNG, glow effect |
| **UI Element** | Varies | `assets/sprites/ui/` | PNG, transparent |
| **Portrait** | 128x128 px | `assets/sprites/portraits/` | PNG |
| **Planet Sprite** | 256x256 px | `assets/sprites/planets/` | PNG, transparent |

### Ship Interior Assets

Interior sprites for any ship area (bridge, engineering, cargo, quarters, etc.):

| Type | Dimensions | Location | Format | Notes |
|------|------------|----------|--------|-------|
| **Interior Tile** | 64x32 px | `assets/sprites/{area}/tiles/` | PNG | Metallic floors, glow accents |
| **Console/Station** | 128x96 px | `assets/sprites/{area}/consoles/` | PNG | Isometric workstations |
| **Crew/Character** | 128x128 px | `assets/sprites/{area}/crew/` | PNG | 4-frame pose sheet |
| **Furniture** | 64-128 px | `assets/sprites/{area}/props/` | PNG | Chairs, tables, equipment |

**Sprite Sheet Layouts for Animation:**
- **Horizontal (4x1)**: 4 frames in a row - ideal for walk cycles, total 128x48 px
- **Grid (2x2)**: 4 frames in 2x2 - ideal for directional poses, each ~256x256 in 1024x1024 from AI

**Prompt Tips for Better Integration:**
- Specify "4 frames in HORIZONTAL row" or "4 frames in 2x2 GRID" explicitly
- Request specific pixel dimensions, even if AI generates larger (we resize)
- Use "transparent background" for all sprites
- Add "isometric perspective" for consistent angle

### 3D Textures (for Tetra3D sphere rendering)

| Type | Dimensions | Location | Format | Notes |
|------|------------|----------|--------|-------|
| **Planet Texture** | 2048x1024+ | `assets/planets/` | JPG | **Equirectangular (2:1 ratio)** |
| **Background** | 1920x1080+ | `assets/data/starmap/background/` | JPG/PNG | Panoramic or tiled |
| **Ring Texture** | 1024x64+ | `assets/planets/` | PNG with alpha | For Saturn-like rings |

**Important:** Planet textures for 3D rendering MUST use equirectangular projection (2:1 aspect ratio) to wrap correctly on spheres. Square photos will distort at the poles.

## Image Optimization

AI image generation (via Gemini/Imagen) produces high-resolution 1024x1024 images. These need to be resized for game use.

### Optimization Workflow

```
AI Generation (1024x1024) → Review → Organize → Resize → Manifest → Test
```

**When to Optimize:**
- Always optimize game sprites (tiles, entities, consoles)
- Keep high-res for backgrounds and reference textures
- Optimization happens at organize time with `--optimize` flag or separately

**Auto-Optimization Sizes:**

| Asset Type | Target Size | Notes |
|------------|-------------|-------|
| Isometric Tile | 64x32 | Diamond shape preserved |
| Entity/Creature | 128x48 | 4-frame horizontal sheet |
| Console/Station | 128x96 | Isometric workstation |
| Crew/Character | 128x128 | 4-frame pose sheet |
| Star | 16x16 | Glow preserved |
| Planet | 256x256 | Detail preserved |
| Portrait | 128x128 | Readable at 64x64 |
| Background | Keep original | No resize |

**Resize Quality:**
Uses Catmull-Rom interpolation (high quality, built into Go - no ImageMagick needed).

### Manual Optimization

```bash
# Resize to specific dimensions
.claude/skills/asset-manager/scripts/optimize_asset.sh assets/sprites/bridge/consoles/helm.png 128 96

# Auto-detect size from path
.claude/skills/asset-manager/scripts/optimize_asset.sh assets/sprites/iso_tiles/crystal.png

# Original is backed up as *_original.png
```

## Available Scripts

### `scripts/generate_asset.sh <type> <name> <prompt> [--dry-run]`
Generate a new asset using AI with proper styling and dimensions. Uses `voyage ai` CLI.

```bash
# Generate isometric tile
.claude/skills/asset-manager/scripts/generate_asset.sh tile grass "lush alien grass, bioluminescent"

# Generate entity (4-frame sprite sheet)
.claude/skills/asset-manager/scripts/generate_asset.sh entity robot "maintenance robot, humanoid"

# Generate portrait
.claude/skills/asset-manager/scripts/generate_asset.sh portrait captain "human ship captain, weathered, wise"
```

### `scripts/organize_asset.sh <source> <type> <name> [--optimize]`
Move a generated image to the correct location with proper naming. Optionally resize to game dimensions.

```bash
# Move generated image to assets
.claude/skills/asset-manager/scripts/organize_asset.sh assets/generated/response_123.png tile alien_grass

# Move AND resize to game-ready dimensions
.claude/skills/asset-manager/scripts/organize_asset.sh assets/generated/response_456.png console helm --optimize
```

### `scripts/optimize_asset.sh <image_path> [width] [height]`
Resize an image to game-ready dimensions. Auto-detects target size from path if not specified.

```bash
# Auto-detect size from asset path
.claude/skills/asset-manager/scripts/optimize_asset.sh assets/sprites/iso_tiles/crystal.png

# Specify exact dimensions
.claude/skills/asset-manager/scripts/optimize_asset.sh assets/generated/temp.png 128 96
```

### `scripts/update_manifest.sh`
Scan assets directory and update `manifest.json` with any new sprites.

### `scripts/status.sh`
Display current asset inventory and identify gaps.

### `scripts/download_reference.sh <type>`
Download real-world reference images (NASA planets, ESO backgrounds).

```bash
# Download NASA planet images for reference
.claude/skills/asset-manager/scripts/download_reference.sh planets

# Download ESO/ESA galaxy backgrounds
.claude/skills/asset-manager/scripts/download_reference.sh backgrounds
```

### `scripts/download_planet_textures.sh [resolution]`
Download proper equirectangular planet textures for 3D sphere rendering.

```bash
# Download 2K textures (default, good for game use)
.claude/skills/asset-manager/scripts/download_planet_textures.sh

# Download 4K textures (high quality)
.claude/skills/asset-manager/scripts/download_planet_textures.sh 4k

# Download 8K textures (very high quality, large files)
.claude/skills/asset-manager/scripts/download_planet_textures.sh 8k
```

Downloads from Solar System Scope (CC BY 4.0). Includes all Solar System planets, moons, and ring textures.

## Workflow

### 1. Determine Asset Requirements

Check what assets are needed:
```bash
.claude/skills/asset-manager/scripts/status.sh
```

Review design docs for planned features requiring new art.

### 2. Generate with AI

Use the generate script or direct AI image generation:

**Via script:**
```bash
.claude/skills/asset-manager/scripts/generate_asset.sh tile crystal "purple alien crystals"
```

**Via direct prompt (include style guide):**
```
Generate a 64x32 pixel art isometric tile showing [description].
Style: Retro pixel art, limited palette, clear outlines.
Format: PNG with transparent background.
Perspective: Isometric (2:1 ratio, diamond shape).
```

### 3. Review and Iterate

Check the generated image in `assets/generated/`. If it needs adjustment:
- Request modifications with specific feedback
- Re-generate with refined prompt
- May need multiple iterations for complex assets

### 4. Organize and Install

Move approved assets to final location:
```bash
.claude/skills/asset-manager/scripts/organize_asset.sh assets/generated/response_xxx.png tile alien_crystal
```

### 5. Update Manifest

Add new sprites to the manifest for engine loading:
```bash
.claude/skills/asset-manager/scripts/update_manifest.sh
```

### 6. Test In-Game

Run the game to verify assets display correctly:
```bash
make run-mock
```

## Real-World Reference Sources

For scientifically accurate assets, use these sources:

### Planets & Moons
- **NASA Image Gallery**: https://images.nasa.gov/
- **NASA Solar System**: https://solarsystem.nasa.gov/
- Good for: Realistic planet textures, moons, asteroids

### Galaxy & Nebula Backgrounds
- **ESA Gaia**: https://www.esa.int/gaia (already in starmap-manager)
- **ESO Image Archive**: https://www.eso.org/public/images/
- **Hubble Gallery**: https://hubblesite.org/images/gallery

### Star References
- Use spectral class colors (O=blue, B=blue-white, A=white, F=yellow-white, G=yellow, K=orange, M=red)

## Resources

### Style Guide
See [`resources/style_guide.md`](resources/style_guide.md) for:
- Color palette specifications
- Pixel art techniques
- Isometric grid rules
- Animation frame standards

### Asset Specifications
See [`resources/asset_specs.md`](resources/asset_specs.md) for:
- Detailed dimensions for each asset type
- File naming conventions
- Manifest.json schema
- Sprite ID allocation

### Prompt Templates
See [`resources/prompt_templates.md`](resources/prompt_templates.md) for:
- Tested prompts for each asset type
- Style consistency phrases
- Common modifications

## Sprite ID Allocation

| Range | Type | Example |
|-------|------|---------|
| 1-99 | Tiles | 1=water, 2=forest, 3=desert, 4=mountain |
| 100-199 | Entities | 100-104=NPCs, 105=player |
| 200-299 | Stars | 200=blue, 201=white, 202=yellow... |
| 300-399 | UI | Reserved |
| 400-499 | Planets | Reserved |
| 500-599 | Ships | Reserved |
| 600-699 | Portraits | Reserved |

## Procedural Exoplanet Textures

For fictional/theoretical planets that players visit, use these approaches:

### Option 1: AI Image Generation (Recommended for unique planets)

Generate equirectangular maps using AI with specific prompts:

```
Generate an equirectangular projection planet texture map, 2:1 aspect ratio (2048x1024).
[Planet type]: [rocky/gas giant/ice giant/ocean world/lava world]
Features: [continents, craters, storms, bands, clouds, etc.]
Color palette: [specific colors matching planet type]
Style: Realistic, NASA-quality, seamless at edges.
No stars or space background - just the planet surface.
```

**Key requirements:**
- Must be 2:1 aspect ratio (equirectangular projection)
- Left and right edges must tile seamlessly
- Top/bottom converge to poles
- No background (solid color or transparent)

### Option 2: Procedural Generation Libraries

For batch generation, consider these tools (not yet integrated):

| Tool | Type | Notes |
|------|------|-------|
| **libnoise** | C++ | Perlin noise for terrain heightmaps |
| **Space Engine** | App | Exports planet textures |
| **Blender + Geometry Nodes** | 3D | Procedural planet shader |

### Option 3: Color Variants

For quick variants, take existing textures and apply color transformations:

```bash
# ImageMagick to recolor Earth for an ocean world
convert earth_daymap.jpg -modulate 100,80,180 ocean_world.jpg

# Make a desert world (warm tones)
convert earth_daymap.jpg -modulate 100,120,30 desert_world.jpg
```

### Planet Type Guidelines

| Type | Base Color | Features | Example |
|------|------------|----------|---------|
| **Rocky** | Gray/brown | Craters, mountains | Mercury, Moon |
| **Terrestrial** | Blue/green/tan | Continents, clouds | Earth, hypothetical |
| **Gas Giant** | Orange/tan bands | Cloud bands, storms | Jupiter |
| **Ice Giant** | Cyan/blue | Subtle bands | Uranus, Neptune |
| **Ocean World** | Deep blue | Cloud patterns | Hypothetical |
| **Lava World** | Black + orange cracks | Magma rivers | Hypothetical |
| **Desert World** | Tan/orange | Dunes, canyons | Mars-like |

## Notes

- All game sprites use pixel art style for consistency
- Isometric tiles use 2:1 ratio (64 wide, 32 tall)
- Entity sprites are 4-frame horizontal sheets
- Generated images go to `assets/generated/` first
- Always test new assets in-game before committing
- Real-world reference data should cite sources (NASA, ESO are CC-compatible)
- **Planet textures (3D)**: Must be equirectangular (2:1 ratio) for proper UV mapping
