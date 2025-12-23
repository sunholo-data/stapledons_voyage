#!/bin/bash
# Generate scene images with transparent windows using AI segmentation
# Pipeline: AI generation -> AI segmentation (Gemini vision) -> mask + transparent PNG
#
# Usage: generate_transparent_scene.sh [--aspect RATIO] [--deck TYPE] <name> <description>

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../../.." && pwd)"

cd "$PROJECT_ROOT"

# Gemini 2.5 Flash Image supported aspect ratios
VALID_RATIOS=("21:9" "16:9" "4:3" "3:2" "1:1" "2:3" "3:4" "9:16" "5:4" "4:5")

NAME=""
DESCRIPTION=""
ASPECT_RATIO="16:9"  # Default for game backgrounds
DECK_TYPE=""         # Auto-detect from name

# Parse arguments
while [[ $# -gt 0 ]]; do
    case "$1" in
        --aspect)
            ASPECT_RATIO="$2"
            shift 2
            ;;
        --deck)
            DECK_TYPE="$2"
            shift 2
            ;;
        --help|-h)
            echo "Generate scene images with transparent windows using AI segmentation"
            echo ""
            echo "Usage: $0 [--aspect RATIO] [--deck TYPE] <name> <description>"
            echo ""
            echo "Options:"
            echo "  --aspect RATIO     Aspect ratio (default: 16:9)"
            echo "                     Valid: ${VALID_RATIOS[*]}"
            echo "  --deck TYPE        Deck type: observation, bridge, generic (auto-detect)"
            echo ""
            echo "Native Resolutions (no upscaling):"
            echo "  21:9 = 1792x768   (ultra-wide panorama)"
            echo "  16:9 = 1344x768   (widescreen, recommended)"
            echo "  4:3  = 1024x768   (classic display)"
            echo "  3:2  = 1216x810   (photo standard)"
            echo "  1:1  = 1024x1024  (square)"
            echo ""
            echo "Examples:"
            echo "  $0 bridge 'spaceship bridge with three panoramic viewports'"
            echo "  $0 --aspect 21:9 observation 'dome observation deck'"
            echo "  $0 --deck bridge command 'starship command center'"
            echo ""
            echo "The script will:"
            echo "  1. Generate image with Gemini (native resolution)"
            echo "  2. Use AI segmentation to detect window regions"
            echo "  3. Create mask and transparent PNG"
            echo "  4. Save to assets/decks/<name>.png and <name>_mask.png"
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

# Validate inputs
if [[ -z "$NAME" ]] || [[ -z "$DESCRIPTION" ]]; then
    echo "ERROR: Missing required arguments"
    echo ""
    echo "Usage: $0 [--aspect RATIO] [--deck TYPE] <name> <description>"
    echo "Run with --help for more information"
    exit 1
fi

# Validate aspect ratio
VALID=0
for ratio in "${VALID_RATIOS[@]}"; do
    if [[ "$ASPECT_RATIO" == "$ratio" ]]; then
        VALID=1
        break
    fi
done

if [[ $VALID -eq 0 ]]; then
    echo "ERROR: Invalid aspect ratio '$ASPECT_RATIO'"
    echo "Supported: ${VALID_RATIOS[*]}"
    exit 1
fi

# Auto-detect deck type from name
if [[ -z "$DECK_TYPE" ]]; then
    NAME_LOWER=$(echo "$NAME" | tr '[:upper:]' '[:lower:]')
    if [[ "$NAME_LOWER" == *"observation"* ]] || [[ "$NAME_LOWER" == *"dome"* ]]; then
        DECK_TYPE="observation"
    elif [[ "$NAME_LOWER" == *"bridge"* ]] || [[ "$NAME_LOWER" == *"command"* ]]; then
        DECK_TYPE="bridge"
    else
        DECK_TYPE="generic"
    fi
fi

# Build prompt for scene with clearly visible windows
PROMPT="Spaceship interior scene: ${DESCRIPTION}.
Windows/viewports should show bright sky or space - make them clearly visible as transparent viewing areas.
Interior structure should provide good contrast with window areas.
Art style: Detailed illustration, sci-fi aesthetic."

echo "=== Transparent Scene Generator (AI Segmentation) ==="
echo "Name: $NAME"
echo "Deck type: $DECK_TYPE"
echo "Aspect ratio: $ASPECT_RATIO"
echo ""

# Build tools if needed
if [[ ! -f "bin/generate-window-mask-ai" ]]; then
    echo "Building generate-window-mask-ai..."
    go build -o bin/generate-window-mask-ai ./cmd/generate-window-mask-ai
fi

# Step 1: Generate with AI
echo "[1/3] Generating image with AI ($ASPECT_RATIO)..."
OUTPUT=$(go run "$PROJECT_ROOT/cmd/voyage" ai -generate-image -aspect "$ASPECT_RATIO" -size 2K -prompt "$PROMPT" 2>&1)
echo "$OUTPUT"

# Extract generated file path
GENERATED=$(echo "$OUTPUT" | grep -o 'assets/generated/response_[0-9]*.png' | head -1)
if [[ -z "$GENERATED" ]]; then
    echo "ERROR: Could not find generated file"
    exit 1
fi

echo ""
echo "[2/3] Detecting windows with AI segmentation..."

# Step 2: Use AI segmentation to create mask
TEMP_MASK="/tmp/${NAME}_mask.png"

bin/generate-window-mask-ai \
    -deck "$DECK_TYPE" \
    -overlay \
    -v \
    -o "$TEMP_MASK" \
    "$PROJECT_ROOT/$GENERATED"

if [[ ! -f "$TEMP_MASK" ]]; then
    echo "ERROR: Failed to generate mask"
    exit 1
fi

# Step 3: Create transparent version by applying mask
echo ""
echo "[3/3] Creating transparent PNG..."
TEMP_TRANSPARENT="/tmp/${NAME}_transparent.png"

# Use Go to apply mask to original (make masked areas transparent)
go run - <<'GOSCRIPT' "$PROJECT_ROOT/$GENERATED" "$TEMP_MASK" "$TEMP_TRANSPARENT"
package main

import (
    "image"
    "image/color"
    "image/png"
    "os"
    _ "image/jpeg"
)

func main() {
    srcPath, maskPath, outPath := os.Args[1], os.Args[2], os.Args[3]

    srcFile, _ := os.Open(srcPath)
    defer srcFile.Close()
    src, _, _ := image.Decode(srcFile)

    maskFile, _ := os.Open(maskPath)
    defer maskFile.Close()
    mask, _, _ := image.Decode(maskFile)

    bounds := src.Bounds()
    out := image.NewNRGBA(bounds)

    for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
        for x := bounds.Min.X; x < bounds.Max.X; x++ {
            r, g, b, _ := src.At(x, y).RGBA()
            mr, _, _, _ := mask.At(x, y).RGBA()

            if mr > 32768 { // Mask is white = transparent
                out.Set(x, y, color.NRGBA{0, 0, 0, 0})
            } else {
                out.Set(x, y, color.NRGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), 255})
            }
        }
    }

    outFile, _ := os.Create(outPath)
    defer outFile.Close()
    png.Encode(outFile, out)
}
GOSCRIPT

# Step 4: Move to final location
FINAL_DIR="$PROJECT_ROOT/assets/decks"
mkdir -p "$FINAL_DIR"
FINAL_TRANSPARENT="$FINAL_DIR/${NAME}.png"
FINAL_MASK="$FINAL_DIR/${NAME}_mask.png"

echo ""
echo "Saving to $FINAL_DIR..."
cp "$TEMP_TRANSPARENT" "$FINAL_TRANSPARENT"
cp "$TEMP_MASK" "$FINAL_MASK"

# Verify
echo ""
echo "=== Result ==="
if command -v sips &> /dev/null; then
    sips -g hasAlpha -g pixelWidth -g pixelHeight "$FINAL_TRANSPARENT" 2>/dev/null | grep -E "(hasAlpha|pixel)" || true
fi
echo ""
echo "Generated:   $GENERATED (original)"
echo "Transparent: $FINAL_TRANSPARENT (windows are transparent)"
echo "Mask:        $FINAL_MASK (white = window areas)"
echo "Overlay:     ${TEMP_MASK%.png}_overlay.png (preview)"
echo ""
echo "To preview: open $FINAL_TRANSPARENT"
