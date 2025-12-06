#!/bin/bash
# Asset Manager - Organize Asset Script
# Move generated images to the correct location with proper naming

set -e

usage() {
    echo "Usage: $0 <source_file> <type> <name>"
    echo ""
    echo "Arguments:"
    echo "  source_file  Path to the generated image (e.g., assets/generated/response_123.png)"
    echo "  type         Asset type: tile, entity, star, ui, planet, portrait, background"
    echo "  name         Asset name (e.g., alien_grass, npc_merchant)"
    echo ""
    echo "Examples:"
    echo "  $0 assets/generated/response_123.png tile alien_crystal"
    echo "  $0 assets/generated/response_456.png entity alien_merchant"
    echo "  $0 assets/generated/response_789.png portrait captain_chen"
    exit 1
}

if [[ $# -ne 3 ]]; then
    usage
fi

SOURCE="$1"
TYPE="$2"
NAME="$3"

# Find project root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../../.." && pwd)"
ASSETS_DIR="$PROJECT_ROOT/assets"
SPRITES_DIR="$ASSETS_DIR/sprites"

# Resolve source path
if [[ ! "$SOURCE" = /* ]]; then
    SOURCE="$PROJECT_ROOT/$SOURCE"
fi

# Check source exists
if [[ ! -f "$SOURCE" ]]; then
    echo "ERROR: Source file not found: $SOURCE"
    exit 1
fi

# Determine destination based on type
case "$TYPE" in
    tile)
        DEST_DIR="$SPRITES_DIR/iso_tiles"
        DEST_FILE="$NAME.png"
        ;;
    entity)
        DEST_DIR="$SPRITES_DIR/iso_entities"
        DEST_FILE="$NAME.png"
        ;;
    star)
        DEST_DIR="$SPRITES_DIR/stars"
        DEST_FILE="$NAME.png"
        ;;
    ui)
        DEST_DIR="$SPRITES_DIR/ui"
        DEST_FILE="$NAME.png"
        ;;
    planet)
        DEST_DIR="$SPRITES_DIR/planets"
        DEST_FILE="$NAME.png"
        ;;
    portrait)
        DEST_DIR="$SPRITES_DIR/portraits"
        DEST_FILE="$NAME.png"
        ;;
    background)
        DEST_DIR="$ASSETS_DIR/data/starmap/background"
        # Preserve extension for backgrounds (jpg or png)
        EXT="${SOURCE##*.}"
        DEST_FILE="$NAME.$EXT"
        ;;
    *)
        echo "ERROR: Unknown asset type: $TYPE"
        echo "Valid types: tile, entity, star, ui, planet, portrait, background"
        exit 1
        ;;
esac

# Create destination directory if needed
mkdir -p "$DEST_DIR"

DEST_PATH="$DEST_DIR/$DEST_FILE"

# Check if destination already exists
if [[ -f "$DEST_PATH" ]]; then
    echo "WARNING: Destination already exists: $DEST_PATH"
    read -p "Overwrite? [y/N] " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Aborted."
        exit 1
    fi
fi

# Copy file (preserve original in generated/ for reference)
cp "$SOURCE" "$DEST_PATH"

echo "SUCCESS: Asset installed"
echo "  From: $SOURCE"
echo "  To:   $DEST_PATH"
echo ""
echo "Next steps:"
echo "  1. Run 'update_manifest.sh' to add to manifest"
echo "  2. Test in-game with 'make run-mock'"
echo "  3. Optionally delete source: rm $SOURCE"
