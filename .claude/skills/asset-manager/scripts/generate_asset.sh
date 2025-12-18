#!/bin/bash
# Asset Manager - Generate Asset Script
# Uses voyage ai CLI to generate game assets with consistent styling

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../../.." && pwd)"
VOYAGE_CLI="$PROJECT_ROOT/bin/voyage"

usage() {
    echo "Usage: $0 <type> <name> <description> [--dry-run]"
    echo ""
    echo "Arguments:"
    echo "  type         Asset type (see below)"
    echo "  name         Asset name (e.g., bridge_floor, captain)"
    echo "  description  Brief description of what to generate"
    echo "  --dry-run    Print prompt without generating"
    echo ""
    echo "Asset Types:"
    echo "  star         16x16 star sprite"
    echo "  planet       256x256 planet sprite"
    echo "  portrait     128x128 character portrait"
    echo "  background   1920x1080 space background"
    echo ""
    echo "3D Texture types (for Tetra3D interior rendering):"
    echo "  floor        Seamless tileable floor texture"
    echo "  wall         Seamless tileable wall texture"
    echo "  ceiling      Seamless tileable ceiling texture"
    echo ""
    echo "Examples:"
    echo "  $0 portrait captain 'human ship captain, weathered, wise'"
    echo "  $0 star blue 'bright blue O-type star with corona'"
    echo "  $0 floor bridge 'metallic grating with cyan LED strips'"
    echo "  $0 wall corridor 'gray bulkhead panels with orange warning stripe'"
    echo "  $0 ceiling engineering 'dark panels with exposed pipes and lights'"
    echo ""
    echo "For 3D textures, use direct AI commands for more control:"
    echo "  bin/voyage ai -generate-image -prompt 'Seamless tileable...'"
    echo ""
    exit 1
}

# Check if voyage CLI exists
check_voyage_cli() {
    if [[ ! -f "$VOYAGE_CLI" ]]; then
        echo "Building voyage CLI..."
        (cd "$PROJECT_ROOT" && go build -o bin/voyage ./cmd/voyage)
    fi
}

DRY_RUN=""
if [[ "$*" == *"--dry-run"* ]]; then
    DRY_RUN="true"
fi

if [[ $# -lt 3 ]]; then
    usage
fi

TYPE="$1"
NAME="$2"
shift 2
DESCRIPTION="${*//--dry-run/}"
DESCRIPTION="$(echo "$DESCRIPTION" | xargs)"  # Trim whitespace

# Base style for pixel art assets
STYLE_BASE="pixel art style, retro 16-bit aesthetic, limited color palette, crisp pixels, no anti-aliasing, clear dark outlines"

# Sci-fi style for interior textures
STYLE_SCIFI="sci-fi starship interior, dark metallic surfaces, glowing accent lights"

# Generate prompt based on type
case "$TYPE" in
    star)
        PROMPT="Create a 16x16 pixel art star sprite showing ${DESCRIPTION}.
Style: Soft radial glow, brightest at center fading to edges.
Format: PNG with transparent background.
Use appropriate spectral color for star type."
        TARGET_DIR="sprites/stars"
        ;;

    planet)
        PROMPT="Create a 256x256 pixel art planet sprite showing ${DESCRIPTION}.
Style: ${STYLE_BASE} but with more detail allowed at this size.
Lighting: Spherical shading with terminator line (day/night edge).
Atmosphere glow if applicable.
Format: PNG with transparent background.
Reference NASA planetary imagery for realistic features."
        TARGET_DIR="sprites/planets"
        ;;

    portrait)
        PROMPT="Create a 128x128 pixel art portrait showing ${DESCRIPTION}.
Style: ${STYLE_BASE}.
View: Face-forward or 3/4 view. Clear facial features readable at 64x64.
Light source from top-left.
Format: PNG with solid dark (#1a1a2e) or transparent background.
Should convey personality for dialogue UI."
        TARGET_DIR="sprites/portraits"
        ;;

    background)
        PROMPT="Create a 1920x1080 space background showing ${DESCRIPTION}.
Style: Can be more detailed than sprites but cohesive with pixel art foreground.
Should not distract from game UI elements.
Evoke cosmic scale of galaxy-spanning exploration."
        TARGET_DIR="data/starmap/background"
        ;;

    floor)
        PROMPT="Seamless tileable metallic spaceship floor texture, ${DESCRIPTION}, sci-fi starship interior deck plating, industrial grating, top-down orthographic view for 3D texture mapping, 512x512 seamless repeating tile, no perspective distortion"
        TARGET_DIR="textures/interior"
        ;;

    wall)
        PROMPT="Seamless tileable spaceship interior wall panel texture, ${DESCRIPTION}, sci-fi bulkhead with recessed panel sections, dark gray brushed metal with rivets, vertical orientation for wall mapping, 512x512 seamless repeating tile"
        TARGET_DIR="textures/interior"
        ;;

    ceiling)
        PROMPT="Seamless tileable spaceship ceiling texture, ${DESCRIPTION}, dark industrial ceiling panels with lighting elements, sci-fi starship bridge overhead, top-down view for ceiling mapping, 512x512 seamless repeating tile"
        TARGET_DIR="textures/interior"
        ;;

    *)
        echo "ERROR: Unknown asset type: $TYPE"
        usage
        ;;
esac

# Output info
echo "=== Asset Generation ==="
echo "Type: $TYPE"
echo "Name: $NAME"
echo "Target: assets/$TARGET_DIR/${NAME}.png"
echo ""

if [[ -n "$DRY_RUN" ]]; then
    echo "=== PROMPT (dry-run) ==="
    echo "$PROMPT"
    echo ""
    echo "To generate, run without --dry-run"
    exit 0
fi

# Check CLI exists
check_voyage_cli

echo "Generating with AI..."
echo ""

# Generate the asset
OUTPUT=$("$VOYAGE_CLI" ai -generate-image -prompt "$PROMPT" 2>&1)
echo "$OUTPUT"

# Extract the generated file path
GENERATED_FILE=$(echo "$OUTPUT" | grep -o 'assets/generated/response_[0-9]*.png' | head -1)

if [[ -z "$GENERATED_FILE" ]]; then
    echo ""
    echo "ERROR: Could not find generated file in output"
    exit 1
fi

echo ""
echo "=== Generated ==="
echo "File: $GENERATED_FILE"
echo ""
echo "Next steps:"
echo "  1. Review the image: open $GENERATED_FILE"

# Show type-specific next steps
if [[ "$TYPE" == "floor" || "$TYPE" == "wall" || "$TYPE" == "ceiling" ]]; then
    echo "  2. If good, organize it:"
    echo "     mkdir -p assets/$TARGET_DIR"
    echo "     cp $GENERATED_FILE assets/$TARGET_DIR/${NAME}.png"
    echo "  3. Test in-game:"
    echo "     go run ./cmd/demo-engine-interior -${TYPE} assets/$TARGET_DIR/${NAME}.png"
else
    echo "  2. If good, organize it:"
    echo "     mkdir -p assets/$TARGET_DIR"
    echo "     cp $GENERATED_FILE assets/$TARGET_DIR/${NAME}.png"
    echo "  3. Update manifest:"
    echo "     .claude/skills/asset-manager/scripts/update_manifest.sh"
fi
