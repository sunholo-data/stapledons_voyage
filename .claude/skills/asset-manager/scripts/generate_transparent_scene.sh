#!/bin/bash
# Generate scene images with transparent windows using AI vision
# Pipeline: AI generation (bright windows) -> AI vision count -> transparency conversion
#
# Usage: generate_transparent_scene.sh <name> <description>

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../../.." && pwd)"

NAME=""
DESCRIPTION=""

# Parse arguments
while [[ $# -gt 0 ]]; do
    case "$1" in
        --help|-h)
            echo "Generate scene images with transparent windows using AI vision"
            echo ""
            echo "Usage: $0 <name> <description>"
            echo ""
            echo "Examples:"
            echo "  $0 observation_deck 'luxury observation lounge with curved windows'"
            echo "  $0 bridge 'spaceship bridge with three large panoramic windows'"
            echo ""
            echo "The script will:"
            echo "  1. Generate 16:9 2K image with bright windows (AI)"
            echo "  2. Use AI vision to detect number of distinct window regions"
            echo "  3. Extract mask and create transparent PNG"
            echo "  4. Save both to assets/decks/<name>.png and <name>_mask.png"
            exit 0
            ;;
        *)
            if [[ -z "$NAME" ]]; then
                NAME="$1"
            else
                DESCRIPTION="${DESCRIPTION:+$DESCRIPTION }$1"
            fi
            shift
            ;;
    esac
done

if [[ -z "$NAME" ]] || [[ -z "$DESCRIPTION" ]]; then
    echo "Usage: $0 <name> <description>"
    echo "Run with --help for more information"
    exit 1
fi

# Build prompt for bright windows (best for AI vision detection)
PROMPT="Spaceship interior scene: ${DESCRIPTION}.
Windows showing pure bright white overexposed light (no space details visible).
Dark metallic interior with detailed sci-fi panels, consoles, and ambient lighting.
Art style: Detailed illustration, warm orange/amber accent lighting, art-deco sci-fi aesthetic.
The bright window areas will be detected by AI vision for masking."

echo "=== Transparent Scene Generator (AI Vision) ==="
echo "Name: $NAME"
echo "Description: $DESCRIPTION"
echo ""

# Step 1: Generate with AI (16:9, 2K resolution for deck backgrounds)
echo "[1/3] Generating image with AI (16:9, 2K)..."
OUTPUT=$(go run "$PROJECT_ROOT/cmd/voyage" ai -generate-image -aspect 16:9 -size 2K -prompt "$PROMPT" 2>&1)
echo "$OUTPUT"

# Extract generated file path
GENERATED=$(echo "$OUTPUT" | grep -o 'assets/generated/response_[0-9]*.png' | head -1)
if [[ -z "$GENERATED" ]]; then
    echo "ERROR: Could not find generated file"
    exit 1
fi

echo ""
echo "[2/3] Generating transparent PNG and mask with AI vision..."

# Step 2: Convert to transparent using AI vision auto-detection
TEMP_TRANSPARENT="/tmp/${NAME}_transparent.png"
TEMP_MASK="/tmp/${NAME}_mask.png"

# Use AI auto-detection for bright mode (default)
echo "Using AI vision to auto-detect window count..."
go run "$PROJECT_ROOT/cmd/generate-window-mask" -mode=bright -threshold=180 -auto-detect "$PROJECT_ROOT/$GENERATED" "$TEMP_TRANSPARENT"
go run "$PROJECT_ROOT/cmd/generate-window-mask" -mode=bright -threshold=180 -auto-detect -mask "$PROJECT_ROOT/$GENERATED" "$TEMP_MASK"

# Step 3: Move to final location
FINAL_DIR="$PROJECT_ROOT/assets/decks"
mkdir -p "$FINAL_DIR"
FINAL_TRANSPARENT="$FINAL_DIR/${NAME}.png"
FINAL_MASK="$FINAL_DIR/${NAME}_mask.png"

echo ""
echo "[3/3] Saving to $FINAL_DIR..."
cp "$TEMP_TRANSPARENT" "$FINAL_TRANSPARENT"
cp "$TEMP_MASK" "$FINAL_MASK"

# Verify alpha
echo ""
echo "=== Result ==="
sips -g hasAlpha -g pixelWidth -g pixelHeight "$FINAL_TRANSPARENT" 2>/dev/null | grep -E "(hasAlpha|pixel)"
echo ""
echo "Generated: $GENERATED (original)"
echo "Transparent: $FINAL_TRANSPARENT (windows are transparent)"
echo "Mask: $FINAL_MASK (white windows, transparent elsewhere)"
echo ""
echo "To preview: open $FINAL_TRANSPARENT"
