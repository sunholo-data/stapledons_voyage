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
    echo "  name         Asset name (e.g., alien_crystal, npc_merchant)"
    echo "  description  Brief description of what to generate"
    echo "  --dry-run    Print prompt without generating"
    echo ""
    echo "Asset Types:"
    echo "  tile         64x32 isometric terrain tile"
    echo "  entity       128x48 sprite sheet (4 frames of 32x48)"
    echo "  star         16x16 star sprite"
    echo "  planet       256x256 planet sprite"
    echo "  portrait     128x128 character portrait"
    echo "  background   1920x1080 space background"
    echo ""
    echo "Ship interior types:"
    echo "  interior_tile   64x32 metallic floor tile with glow accents"
    echo "  console         128x96 isometric console/workstation"
    echo "  crew            4-frame character sprite sheet (specify grid: 4x1 or 2x2)"
    echo "  furniture       Chairs, tables, equipment props"
    echo ""
    echo "Examples:"
    echo "  $0 tile alien_crystal 'purple crystal formations, bioluminescent'"
    echo "  $0 interior_tile floor 'dark metallic plating with cyan circuit lines'"
    echo "  $0 console helm 'pilot station with holographic ship display'"
    echo "  $0 crew engineer 'orange jumpsuit with wrench, 4 poses in 2x2 grid'"
    echo "  $0 furniture medical_bed 'sick bay bed with monitoring equipment'"
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

# Base style for all assets
STYLE_BASE="pixel art style, retro 16-bit aesthetic, limited color palette, crisp pixels, no anti-aliasing, clear dark outlines"

# Sci-fi style for bridge assets
STYLE_SCIFI="sci-fi starship interior, dark metallic surfaces, glowing accent lights, blue and cyan color scheme"

# Generate prompt based on type
case "$TYPE" in
    tile)
        PROMPT="Create a 64x32 pixel art isometric tile showing ${DESCRIPTION}.
Style: ${STYLE_BASE}, isometric view, 2:1 diamond ratio.
Format: PNG with transparent background.
Shape: Diamond/rhombus shaped tile. Light source from top-left.
The tile should work as part of a tileable terrain grid."
        TARGET_DIR="iso_tiles"
        ;;

    interior_tile)
        PROMPT="Create a 64x32 pixel art isometric floor tile for a sci-fi starship interior showing ${DESCRIPTION}.
Style: ${STYLE_BASE}, ${STYLE_SCIFI}.
Format: PNG with transparent background.
Shape: Diamond/rhombus isometric tile (2:1 width-to-height).
Features: Metallic plating with subtle glowing circuit lines or panel seams.
Must tile seamlessly with other floor tiles."
        TARGET_DIR="interior/tiles"
        ;;

    entity)
        PROMPT="Create a 128x48 pixel art sprite sheet showing ${DESCRIPTION}.
CRITICAL LAYOUT: 4 frames arranged HORIZONTALLY in a single row.
Each frame: 32x48 pixels.
Total sheet: 128 pixels wide x 48 pixels tall.
Frame order: idle, walk1, walk2, walk3 (left to right).
Style: ${STYLE_BASE}, isometric-compatible 3/4 view.
Format: PNG with transparent background."
        TARGET_DIR="iso_entities"
        ;;

    crew)
        # Check if description mentions grid format
        if [[ "$DESCRIPTION" == *"2x2"* ]]; then
            GRID_DESC="4 frames arranged in a 2x2 GRID. Each frame approximately 256x256 pixels (512x512 total)."
        else
            GRID_DESC="4 frames arranged HORIZONTALLY in a single row. Each frame approximately 256x256 pixels (1024x256 total)."
        fi
        PROMPT="Create a pixel art sprite sheet of a spaceship crew member: ${DESCRIPTION}.
CRITICAL LAYOUT: ${GRID_DESC}
Frame content: Show 4 different poses/directions (front, back, left-side, right-side OR idle + 3 walk frames).
Style: ${STYLE_BASE}, isometric-compatible view, ${STYLE_SCIFI} uniform colors.
Format: PNG with transparent background.
Character should have clear silhouette and readable at small sizes."
        TARGET_DIR="interior/crew"
        ;;

    console)
        PROMPT="Create an isometric pixel art console/workstation for a sci-fi starship: ${DESCRIPTION}.
Style: ${STYLE_BASE}, ${STYLE_SCIFI}.
View: Isometric perspective (viewed from above at angle).
Features: Holographic displays, control panels, status lights.
Format: PNG with transparent background.
Size: Should work as a 128x96 game asset (will be scaled from AI output)."
        TARGET_DIR="interior/consoles"
        ;;

    furniture)
        PROMPT="Create an isometric pixel art furniture/prop for a sci-fi starship interior: ${DESCRIPTION}.
Style: ${STYLE_BASE}, ${STYLE_SCIFI}.
View: Isometric perspective.
Format: PNG with transparent background.
Size: Should fit within 64x64 to 128x128 pixel area."
        TARGET_DIR="interior/props"
        ;;

    star)
        PROMPT="Create a 16x16 pixel art star sprite showing ${DESCRIPTION}.
Style: Soft radial glow, brightest at center fading to edges.
Format: PNG with transparent background.
Use appropriate spectral color for star type."
        TARGET_DIR="stars"
        ;;

    planet)
        PROMPT="Create a 256x256 pixel art planet sprite showing ${DESCRIPTION}.
Style: ${STYLE_BASE} but with more detail allowed at this size.
Lighting: Spherical shading with terminator line (day/night edge).
Atmosphere glow if applicable.
Format: PNG with transparent background.
Reference NASA planetary imagery for realistic features."
        TARGET_DIR="planets"
        ;;

    portrait)
        PROMPT="Create a 128x128 pixel art portrait showing ${DESCRIPTION}.
Style: ${STYLE_BASE}.
View: Face-forward or 3/4 view. Clear facial features readable at 64x64.
Light source from top-left.
Format: PNG with solid dark (#1a1a2e) or transparent background.
Should convey personality for dialogue UI."
        TARGET_DIR="portraits"
        ;;

    background)
        PROMPT="Create a 1920x1080 space background showing ${DESCRIPTION}.
Style: Can be more detailed than sprites but cohesive with pixel art foreground.
Should not distract from game UI elements.
Evoke cosmic scale of galaxy-spanning exploration."
        TARGET_DIR="backgrounds"
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
echo "Target: assets/sprites/$TARGET_DIR/$NAME.png"
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
echo "  2. If good, organize it:"
echo "     .claude/skills/asset-manager/scripts/organize_asset.sh $GENERATED_FILE $TYPE $NAME"
echo "  3. Optionally optimize/resize:"
echo "     .claude/skills/asset-manager/scripts/optimize_asset.sh assets/sprites/$TARGET_DIR/$NAME.png"
echo "  4. Update manifest:"
echo "     .claude/skills/asset-manager/scripts/update_manifest.sh"
