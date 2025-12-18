#!/bin/bash
# Asset Manager - Verify Isometric Tile
# Checks if a tile meets tessellation requirements using tile-check tool

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../../.." && pwd)"

usage() {
    echo "Usage: $0 <image_path> [--fix]"
    echo ""
    echo "Verifies an isometric tile meets tessellation requirements."
    echo ""
    echo "Options:"
    echo "  --fix       Apply diamond masking to fix non-compliant tiles"
    echo ""
    echo "Requirements checked:"
    echo "  1. Size: 64x32 pixels (2:1 ratio)"
    echo "  2. Format: PNG with alpha channel"
    echo "  3. Corners: Transparent (alpha=0)"
    echo "  4. Opaque pixels: ~1024 (±10%)"
    echo ""
    echo "Examples:"
    echo "  $0 assets/sprites/iso_tiles/bridge_floor.png"
    echo "  $0 assets/generated/response_123.png --fix"
    exit 1
}

if [[ $# -lt 1 ]]; then
    usage
fi

IMAGE_PATH="$1"
FIX_MODE=""

shift
while [[ $# -gt 0 ]]; do
    case "$1" in
        --fix)
            FIX_MODE="true"
            shift
            ;;
        *)
            echo "Unknown option: $1"
            usage
            ;;
    esac
done

if [[ ! -f "$IMAGE_PATH" ]]; then
    echo "ERROR: File not found: $IMAGE_PATH"
    exit 1
fi

echo "=== Verifying Isometric Tile ==="
echo "File: $IMAGE_PATH"
echo ""

# Use the existing tile-check command but pass specific file
cd "$PROJECT_ROOT"

# Run verification with Go directly
OUTPUT=$(go run ./cmd/tile-check 2>&1 | grep -A 20 "$(basename "$IMAGE_PATH")" || echo "")

if [[ -z "$OUTPUT" ]]; then
    echo "Running single file check..."
    OUTPUT=$(go run ./cmd/tile-check 2>&1)
    # Check if our file is in the results
    if ! echo "$OUTPUT" | grep -q "$(basename "$IMAGE_PATH")"; then
        echo "File not in tile-check inventory. Checking directly..."
    fi
fi

# Extract status for our specific file
TILE_NAME=$(basename "$IMAGE_PATH")
if echo "$OUTPUT" | grep -q "$TILE_NAME: PASS"; then
    echo "$OUTPUT" | grep -A 10 "=== $TILE_NAME ==="
    echo ""
    echo "✓ PASS: Tile meets all tessellation requirements"
    exit 0
elif echo "$OUTPUT" | grep -q "$TILE_NAME: FAIL"; then
    echo "$OUTPUT" | grep -A 15 "=== $TILE_NAME ==="
    echo ""
    echo "✗ FAIL: Tile does not meet requirements"

    if [[ "$FIX_MODE" == "true" ]]; then
        echo ""
        echo "Applying diamond mask fix..."

        # Use the engine's MaskToDiamond via optimize script
        "$SCRIPT_DIR/optimize_asset.sh" "$IMAGE_PATH" 64 32

        echo ""
        echo "Re-verifying..."
        "$0" "$IMAGE_PATH"
    else
        echo ""
        echo "To auto-fix this tile, run:"
        echo "  $0 $IMAGE_PATH --fix"
        echo ""
        echo "Or regenerate with proper requirements:"
        echo "  .claude/skills/asset-manager/scripts/generate_asset.sh tile <name> '<description>'"
    fi
    exit 1
else
    echo "Could not determine status for $TILE_NAME"
    echo "Full output:"
    echo "$OUTPUT"
    exit 1
fi
