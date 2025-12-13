#!/bin/bash
# Asset Manager - Organize Asset Script
# Move generated images to the correct location with proper naming
# Optionally auto-optimize to game-ready dimensions

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../../.." && pwd)"
ASSETS_DIR="$PROJECT_ROOT/assets"
SPRITES_DIR="$ASSETS_DIR/sprites"

usage() {
    echo "Usage: $0 <source_file> <type> <name> [--optimize]"
    echo ""
    echo "Arguments:"
    echo "  source_file  Path to the generated image"
    echo "  type         Asset type (see below)"
    echo "  name         Asset name (e.g., alien_grass, npc_merchant)"
    echo "  --optimize   Auto-resize to game-ready dimensions"
    echo ""
    echo "Asset Types:"
    echo "  tile         -> assets/sprites/iso_tiles/"
    echo "  entity       -> assets/sprites/iso_entities/"
    echo "  star         -> assets/sprites/stars/"
    echo "  planet       -> assets/sprites/planets/"
    echo "  portrait     -> assets/sprites/portraits/"
    echo "  background   -> assets/data/starmap/background/"
    echo "  ui           -> assets/sprites/ui/"
    echo ""
    echo "Ship interior types:"
    echo "  interior_tile -> assets/sprites/interior/tiles/"
    echo "  console       -> assets/sprites/interior/consoles/"
    echo "  crew          -> assets/sprites/interior/crew/"
    echo "  furniture     -> assets/sprites/interior/props/"
    echo ""
    echo "Examples:"
    echo "  $0 assets/generated/response_123.png tile alien_crystal"
    echo "  $0 assets/generated/response_456.png console helm --optimize"
    echo "  $0 assets/generated/response_789.png crew engineer"
    echo "  $0 assets/generated/response_000.png furniture medical_bed"
    exit 1
}

OPTIMIZE=""
if [[ "$*" == *"--optimize"* ]]; then
    OPTIMIZE="true"
fi

if [[ $# -lt 3 ]]; then
    usage
fi

SOURCE="$1"
TYPE="$2"
NAME="$3"

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
        TARGET_SIZE="64x32"
        ;;
    interior_tile)
        DEST_DIR="$SPRITES_DIR/interior/tiles"
        DEST_FILE="$NAME.png"
        TARGET_SIZE="64x32"
        ;;
    entity)
        DEST_DIR="$SPRITES_DIR/iso_entities"
        DEST_FILE="$NAME.png"
        TARGET_SIZE="128x48"
        ;;
    console)
        DEST_DIR="$SPRITES_DIR/interior/consoles"
        DEST_FILE="$NAME.png"
        TARGET_SIZE="128x96"
        ;;
    crew)
        DEST_DIR="$SPRITES_DIR/interior/crew"
        DEST_FILE="$NAME.png"
        TARGET_SIZE="128x128"
        ;;
    furniture)
        DEST_DIR="$SPRITES_DIR/interior/props"
        DEST_FILE="$NAME.png"
        TARGET_SIZE="128x128"
        ;;
    star)
        DEST_DIR="$SPRITES_DIR/stars"
        DEST_FILE="$NAME.png"
        TARGET_SIZE="16x16"
        ;;
    planet)
        DEST_DIR="$SPRITES_DIR/planets"
        DEST_FILE="$NAME.png"
        TARGET_SIZE="256x256"
        ;;
    portrait)
        DEST_DIR="$SPRITES_DIR/portraits"
        DEST_FILE="$NAME.png"
        TARGET_SIZE="128x128"
        ;;
    ui)
        DEST_DIR="$SPRITES_DIR/ui"
        DEST_FILE="$NAME.png"
        TARGET_SIZE=""  # No auto size for UI
        ;;
    background)
        DEST_DIR="$ASSETS_DIR/data/starmap/background"
        EXT="${SOURCE##*.}"
        DEST_FILE="$NAME.$EXT"
        TARGET_SIZE=""  # Keep large for backgrounds
        ;;
    *)
        echo "ERROR: Unknown asset type: $TYPE"
        usage
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

echo "=== Asset Installed ==="
echo "Type: $TYPE"
echo "Name: $NAME"
echo "From: $SOURCE"
echo "To:   $DEST_PATH"

# Get current dimensions
CURRENT_DIMS=$(file "$DEST_PATH" | grep -oE '[0-9]+ x [0-9]+')
echo "Size: $CURRENT_DIMS"

# Optimize if requested and target size defined
if [[ -n "$OPTIMIZE" && -n "$TARGET_SIZE" ]]; then
    echo ""
    echo "Optimizing to game size: $TARGET_SIZE"
    W=$(echo "$TARGET_SIZE" | cut -dx -f1)
    H=$(echo "$TARGET_SIZE" | cut -dx -f2)
    "$SCRIPT_DIR/optimize_asset.sh" "$DEST_PATH" "$W" "$H"
fi

echo ""
echo "=== Next Steps ==="
if [[ -z "$OPTIMIZE" && -n "$TARGET_SIZE" && "$CURRENT_DIMS" != "${TARGET_SIZE/ /}" ]]; then
    echo "  - Consider resizing: $SCRIPT_DIR/optimize_asset.sh $DEST_PATH"
fi
echo "  - Update manifest: $SCRIPT_DIR/update_manifest.sh"
echo "  - Test in-game: make run"
