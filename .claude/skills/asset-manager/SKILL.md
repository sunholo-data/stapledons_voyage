---
name: Asset Manager
description: Generate and manage game image assets using AI. Use when user asks to create sprites, tiles, backgrounds, portraits, or other visual assets for the game. Handles isometric formatting, style consistency, and manifest updates. (project)
---

# Asset Manager

Generate, organize, and manage visual assets for Stapledon's Voyage using AI image generation. Ensures consistent pixel art styling, proper isometric formatting, and automatic manifest updates.

## Quick Start

**Most common usage:**
```bash
# Generate a new isometric tile
.claude/skills/asset-manager/scripts/generate_asset.sh tile "alien_crystal" "purple crystal formations, sci-fi, alien world"

# Generate an entity sprite sheet
.claude/skills/asset-manager/scripts/generate_asset.sh entity "alien_merchant" "alien trader, four-armed, robes"

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

| Type | Dimensions | Location | Format |
|------|------------|----------|--------|
| **Isometric Tile** | 64x32 px | `assets/sprites/iso_tiles/` | PNG, transparent bg |
| **Entity Sprite** | 128x48 px (4 frames) | `assets/sprites/iso_entities/` | PNG, sprite sheet |
| **Star Sprite** | 16x16 px | `assets/sprites/stars/` | PNG, glow effect |
| **UI Element** | Varies | `assets/sprites/ui/` | PNG, transparent |
| **Portrait** | 128x128 px | `assets/sprites/portraits/` | PNG |
| **Background** | 1920x1080+ | `assets/data/starmap/background/` | JPG/PNG |
| **Planet** | 256x256 px | `assets/sprites/planets/` | PNG, transparent |

## Available Scripts

### `scripts/generate_asset.sh <type> <name> <prompt>`
Generate a new asset using AI with proper styling and dimensions.

```bash
# Generate isometric tile
.claude/skills/asset-manager/scripts/generate_asset.sh tile grass "lush alien grass, bioluminescent"

# Generate entity (4-frame sprite sheet)
.claude/skills/asset-manager/scripts/generate_asset.sh entity robot "maintenance robot, humanoid"

# Generate portrait
.claude/skills/asset-manager/scripts/generate_asset.sh portrait captain "human ship captain, weathered, wise"
```

### `scripts/organize_asset.sh <source> <type> <name>`
Move a generated image to the correct location with proper naming.

```bash
# Move generated image to assets
.claude/skills/asset-manager/scripts/organize_asset.sh assets/generated/response_123.png tile alien_grass
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

## Notes

- All game sprites use pixel art style for consistency
- Isometric tiles use 2:1 ratio (64 wide, 32 tall)
- Entity sprites are 4-frame horizontal sheets
- Generated images go to `assets/generated/` first
- Always test new assets in-game before committing
- Real-world reference data should cite sources (NASA, ESO are CC-compatible)
