#!/bin/bash
# Asset Manager - Generate Asset Script
# Outputs an AI prompt for generating game assets with consistent styling

set -e

usage() {
    echo "Usage: $0 <type> <name> <description>"
    echo ""
    echo "Arguments:"
    echo "  type         Asset type: tile, entity, star, planet, portrait, background"
    echo "  name         Asset name (e.g., alien_crystal, npc_merchant)"
    echo "  description  Brief description of what to generate"
    echo ""
    echo "Examples:"
    echo "  $0 tile alien_crystal 'purple crystal formations, bioluminescent'"
    echo "  $0 entity alien_merchant 'four-armed alien trader in flowing robes'"
    echo "  $0 portrait captain 'weathered human ship captain, wise expression'"
    echo ""
    echo "This script outputs a prompt to use with AI image generation."
    exit 1
}

if [[ $# -lt 3 ]]; then
    usage
fi

TYPE="$1"
NAME="$2"
shift 2
DESCRIPTION="$*"

# Base style for all assets
STYLE_BASE="pixel art style, retro 16-bit aesthetic, limited color palette, crisp pixels, no anti-aliasing, clear dark outlines"

# Generate prompt based on type
case "$TYPE" in
    tile)
        DIMENSIONS="64x32"
        PERSPECTIVE="isometric view, 2:1 diamond ratio, top-down angled perspective"
        FORMAT="PNG with transparent background"
        SPECIFIC="Diamond/rhombus shaped tile that tiles seamlessly. Light source from top-left."

        PROMPT="Create a ${DIMENSIONS} pixel art isometric tile showing ${DESCRIPTION}.

Style: ${STYLE_BASE}, ${PERSPECTIVE}.
Format: ${FORMAT}.
Details: ${SPECIFIC}

The tile should work as part of a tileable terrain grid for a sci-fi space exploration game."
        ;;

    entity)
        DIMENSIONS="128x48 (sprite sheet with 4 frames, each 32x48)"
        PERSPECTIVE="isometric-compatible, slight 3/4 view from above"
        FORMAT="PNG with transparent background"
        SPECIFIC="4 frames for walk cycle animation. Frame 1: idle/stand. Frames 2-4: walking animation."

        PROMPT="Create a ${DIMENSIONS} pixel art sprite sheet showing ${DESCRIPTION}.

Layout: 4 frames side-by-side, each 32x48 pixels.
Style: ${STYLE_BASE}, ${PERSPECTIVE}.
Format: ${FORMAT}.
Animation: ${SPECIFIC}

Character should have a clear silhouette and work in an isometric game environment."
        ;;

    star)
        DIMENSIONS="16x16"
        FORMAT="PNG with transparent background"
        SPECIFIC="Soft glow effect, brightest at center fading to edges. Suitable for a starmap display."

        PROMPT="Create a ${DIMENSIONS} pixel art star sprite showing ${DESCRIPTION}.

Style: ${STYLE_BASE}.
Format: ${FORMAT}.
Details: ${SPECIFIC}

Star should convey the appropriate spectral color and work at small sizes on a dark background."
        ;;

    planet)
        DIMENSIONS="256x256"
        FORMAT="PNG with transparent background"
        SPECIFIC="Spherical planet with proper lighting (terminator line between day/night). Atmosphere glow if applicable."

        PROMPT="Create a ${DIMENSIONS} pixel art planet sprite showing ${DESCRIPTION}.

Style: ${STYLE_BASE}, but with more detail allowed at this larger size.
Format: ${FORMAT}.
Details: ${SPECIFIC}

Reference NASA planetary imagery for realistic surface features and coloring."
        ;;

    portrait)
        DIMENSIONS="128x128"
        FORMAT="PNG with transparent or solid dark (#1a1a2e) background"
        SPECIFIC="Face-forward or 3/4 view. Clear facial features readable at 64x64. Light source from top-left."

        PROMPT="Create a ${DIMENSIONS} pixel art portrait showing ${DESCRIPTION}.

Style: ${STYLE_BASE}.
Format: ${FORMAT}.
Details: ${SPECIFIC}

Portrait should convey personality and be suitable for dialogue UI in a sci-fi game."
        ;;

    background)
        DIMENSIONS="1920x1080 or larger"
        FORMAT="PNG or JPG"
        SPECIFIC="Deep space background suitable for starmap view. Should not distract from foreground UI elements."

        PROMPT="Create a ${DIMENSIONS} space background showing ${DESCRIPTION}.

Style: Can be more detailed than pixel sprites, but cohesive with pixel art foreground.
Format: ${FORMAT}.
Details: ${SPECIFIC}

Background should evoke the cosmic scale of a galaxy-spanning civilization game."
        ;;

    *)
        echo "ERROR: Unknown asset type: $TYPE"
        echo "Valid types: tile, entity, star, planet, portrait, background"
        exit 1
        ;;
esac

echo "=== AI Image Generation Prompt ==="
echo "Asset: $NAME ($TYPE)"
echo "=================================="
echo ""
echo "$PROMPT"
echo ""
echo "=== Post-Generation ==="
echo "After generating, the image will be saved to assets/generated/"
echo "Then run:"
echo "  .claude/skills/asset-manager/scripts/organize_asset.sh assets/generated/<file>.png $TYPE $NAME"
echo "  .claude/skills/asset-manager/scripts/update_manifest.sh"
