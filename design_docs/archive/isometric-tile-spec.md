# Isometric Tile Specification

This document defines the requirements for isometric tiles to tessellate correctly in Stapledon's Voyage.

## Quick Reference

| Property | Required Value |
|----------|----------------|
| Dimensions | 64x32 pixels (2:1 ratio) |
| Format | PNG with RGBA (alpha channel required) |
| Corners | Must be transparent (alpha=0) |
| Opaque pixels | ~1024 (±10%) in diamond shape |
| Diamond vertices | (32,0), (63,16), (32,31), (0,16) must be opaque |

## Diamond Shape Requirements

For a 64x32 tile, the opaque content must fill an exact isometric diamond:

```
                    ▲ (32,0)
                   /  \
                  /    \
                 /      \
                /        \
(0,16)◄--------/          \--------►(63,16)
               \          /
                \        /
                 \      /
                  \    /
                   \  /
                    ▼ (32,31)
```

### Row-by-Row Pixel Counts

The diamond shape is calculated row-by-row:

| Row | Visible Pixels | Margin Each Side |
|-----|----------------|------------------|
| 0 | 4 | 30 |
| 1 | 8 | 28 |
| ... | ... | ... |
| 15 | 64 | 0 |
| 16 | 64 | 0 |
| ... | ... | ... |
| 30 | 8 | 28 |
| 31 | 4 | 30 |

Formula: `pixels_wide = (rows_from_edge) * 4` where rows_from_edge = min(row+1, 32-row)

## Why These Requirements?

### Corner Transparency

When isometric tiles are placed adjacent to each other, the corners of one tile overlap the body of adjacent tiles. If corners are opaque, you see visible seams:

```
Bad (opaque corners):    Good (transparent corners):
┌──────────┐             ◇◇◇◇◇◇◇◇◇◇
│xxxxxxxx │             ◇xxxxxxxx◇
│xxxxxxxxx│              ◇xxxxxxxx◇
│xxxxxxxxx│               ◇xxxxxxxx◇
└──────────┘                ◇◇◇◇◇◇◇◇◇◇
```

### Exact Diamond Fill

If content doesn't fill the exact diamond, adjacent tiles will have visible gaps or content clipping.

## Verification Tool

Run the tile verification tool to check assets:

```bash
go run ./cmd/tile-check
```

Output shows:
- Size validation
- Corner transparency check
- Opaque pixel count
- Pass/Fail status per tile

### Expected Output (passing tile)

```
=== water.png ===
  Size: 64x32, Format: png, HasAlpha: true
  Corner pixels:
    top-left: OK (transparent)
    top-right: OK (transparent)
    bottom-left: OK (transparent)
    bottom-right: OK (transparent)
  Pixels: 1055 opaque, 993 transparent
  Status: PASS
```

## Current Asset Status

| Asset | Status | Issue |
|-------|--------|-------|
| water.png | PASS | Reference implementation |
| bridge_floor.png | FAIL | Rectangular, no transparency |
| engineering_floor.png | FAIL | Rectangular, no transparency |
| culture_floor.png | FAIL | Rectangular, no transparency |
| habitat_floor.png | FAIL | Rectangular, no transparency |
| core_floor.png | FAIL | Rectangular, no transparency |
| access_point.png | FAIL | Rectangular, no transparency |

## Engine Fallbacks

The engine provides two fallback mechanisms:

### 1. Diamond Masking (`engine/assets/sprites.go`)

```go
// MaskToDiamond masks rectangular tiles to diamond shape
// Uses pixel-perfect row-by-row calculation
masked := sprites.GetMaskedTile(spriteID)
```

This works but loses content in the masked corners.

### 2. Colored Diamond Fallback

When tiles can't be properly masked (content doesn't fill diamond), the renderer falls back to vector-drawn colored diamonds that tessellate perfectly.

## AI Image Generation Prompt

When generating isometric tiles, use this prompt structure:

```
Create a 64x32 pixel isometric floor tile for [DECK_TYPE].

CRITICAL REQUIREMENTS:
1. Exact dimensions: 64 pixels wide, 32 pixels tall
2. Diamond shape ONLY - no rectangular background
3. Transparent background (PNG with alpha)
4. Content must fill the EXACT isometric diamond shape
5. Diamond vertices at: top(32,0), right(63,16), bottom(32,31), left(0,16)
6. Corners (0,0), (63,0), (0,31), (63,31) must be fully transparent
7. Style: [sci-fi/futuristic/metallic] floor plate texture
8. Subtle lighting from top-left
9. No border/outline around the diamond

Visual reference: The tile shape is a perfect rhombus/diamond,
NOT a rectangle with diamond drawn inside it.
```

## See Also

- [cmd/tile-check/main.go](../../cmd/tile-check/main.go) - Verification tool source
- [engine/assets/sprites.go](../../engine/assets/sprites.go) - Masking implementation
- [assets/sprites/iso_tiles/water.png](../../assets/sprites/iso_tiles/water.png) - Reference tile
