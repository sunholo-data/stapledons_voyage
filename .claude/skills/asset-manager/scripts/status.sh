#!/bin/bash
# Asset Manager - Status Script
# Shows current asset inventory and identifies gaps

set -e

# Find project root (where assets/ directory is)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../../.." && pwd)"
ASSETS_DIR="$PROJECT_ROOT/assets"
SPRITES_DIR="$ASSETS_DIR/sprites"
MANIFEST="$SPRITES_DIR/manifest.json"

echo "=== Stapledon's Voyage Asset Status ==="
echo "Project root: $PROJECT_ROOT"
echo ""

# Check if manifest exists
if [[ ! -f "$MANIFEST" ]]; then
    echo "ERROR: Manifest not found at $MANIFEST"
    exit 1
fi

# Count sprites in manifest
SPRITE_COUNT=$(jq '.sprites | length' "$MANIFEST" 2>/dev/null || echo "0")
echo "Sprites in manifest: $SPRITE_COUNT"
echo ""

# Isometric Tiles
echo "=== Isometric Tiles (ID 1-99) ==="
if [[ -d "$SPRITES_DIR/iso_tiles" ]]; then
    TILE_COUNT=$(ls -1 "$SPRITES_DIR/iso_tiles"/*.png 2>/dev/null | wc -l | tr -d ' ')
    echo "Files: $TILE_COUNT"
    ls -1 "$SPRITES_DIR/iso_tiles"/*.png 2>/dev/null | xargs -I {} basename {} | sed 's/^/  /'
else
    echo "Directory missing: iso_tiles/"
fi
echo ""

# Entity Sprites
echo "=== Entity Sprites (ID 100-199) ==="
if [[ -d "$SPRITES_DIR/iso_entities" ]]; then
    ENTITY_COUNT=$(ls -1 "$SPRITES_DIR/iso_entities"/*.png 2>/dev/null | wc -l | tr -d ' ')
    echo "Files: $ENTITY_COUNT"
    ls -1 "$SPRITES_DIR/iso_entities"/*.png 2>/dev/null | xargs -I {} basename {} | sed 's/^/  /'
else
    echo "Directory missing: iso_entities/"
fi
echo ""

# Star Sprites
echo "=== Star Sprites (ID 200-299) ==="
if [[ -d "$SPRITES_DIR/stars" ]]; then
    STAR_COUNT=$(ls -1 "$SPRITES_DIR/stars"/*.png 2>/dev/null | wc -l | tr -d ' ')
    echo "Files: $STAR_COUNT"
    ls -1 "$SPRITES_DIR/stars"/*.png 2>/dev/null | xargs -I {} basename {} | sed 's/^/  /'
else
    echo "Directory missing: stars/"
fi
echo ""

# UI Elements
echo "=== UI Elements (ID 300-399) ==="
if [[ -d "$SPRITES_DIR/ui" ]]; then
    UI_COUNT=$(ls -1 "$SPRITES_DIR/ui"/*.png 2>/dev/null | wc -l | tr -d ' ')
    echo "Files: $UI_COUNT"
    ls -1 "$SPRITES_DIR/ui"/*.png 2>/dev/null | xargs -I {} basename {} | sed 's/^/  /'
else
    echo "Directory not created yet: ui/"
fi
echo ""

# Planets
echo "=== Planet Sprites (ID 400-499) ==="
if [[ -d "$SPRITES_DIR/planets" ]]; then
    PLANET_COUNT=$(ls -1 "$SPRITES_DIR/planets"/*.png 2>/dev/null | wc -l | tr -d ' ')
    echo "Files: $PLANET_COUNT"
    ls -1 "$SPRITES_DIR/planets"/*.png 2>/dev/null | xargs -I {} basename {} | sed 's/^/  /'
else
    echo "Directory not created yet: planets/"
fi
echo ""

# Portraits
echo "=== Portraits (ID 600-699) ==="
if [[ -d "$SPRITES_DIR/portraits" ]]; then
    PORTRAIT_COUNT=$(ls -1 "$SPRITES_DIR/portraits"/*.png 2>/dev/null | wc -l | tr -d ' ')
    echo "Files: $PORTRAIT_COUNT"
    ls -1 "$SPRITES_DIR/portraits"/*.png 2>/dev/null | xargs -I {} basename {} | sed 's/^/  /'
else
    echo "Directory not created yet: portraits/"
fi
echo ""

# Backgrounds
echo "=== Backgrounds ==="
BG_DIR="$ASSETS_DIR/data/starmap/background"
if [[ -d "$BG_DIR" ]]; then
    BG_COUNT=$(ls -1 "$BG_DIR"/*.{jpg,png} 2>/dev/null | wc -l | tr -d ' ')
    echo "Files: $BG_COUNT"
    ls -1 "$BG_DIR"/*.{jpg,png} 2>/dev/null | xargs -I {} basename {} | sed 's/^/  /'
else
    echo "Directory missing: data/starmap/background/"
fi
echo ""

# Generated (staging)
echo "=== Generated (Staging) ==="
GEN_DIR="$ASSETS_DIR/generated"
if [[ -d "$GEN_DIR" ]]; then
    GEN_COUNT=$(ls -1 "$GEN_DIR"/*.png 2>/dev/null | wc -l | tr -d ' ')
    echo "Pending images: $GEN_COUNT"
    if [[ "$GEN_COUNT" -gt 0 ]]; then
        ls -1 "$GEN_DIR"/*.png 2>/dev/null | xargs -I {} basename {} | sed 's/^/  /'
    fi
else
    echo "No generated images pending"
fi
echo ""

# Summary
echo "=== Summary ==="
echo "Run 'update_manifest.sh' to sync manifest with files"
echo "Run 'organize_asset.sh <source> <type> <name>' to install generated images"
