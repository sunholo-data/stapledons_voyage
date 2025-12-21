#!/bin/bash
# Generate scene images with transparent windows
# Pipeline: AI generation -> transparency conversion
#
# Usage: generate_transparent_scene.sh <name> <description> [--mode black|checker]

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../../.." && pwd)"

# Default mode
MODE="black"
NAME=""
DESCRIPTION=""

# Parse arguments
while [[ $# -gt 0 ]]; do
    case "$1" in
        --mode)
            MODE="$2"
            shift 2
            ;;
        --help|-h)
            echo "Generate scene images with transparent windows"
            echo ""
            echo "Usage: $0 <name> <description> [--mode black|checker]"
            echo ""
            echo "Modes:"
            echo "  black   - Request black windows, convert to transparent (default)"
            echo "  checker - Request transparent windows (AI renders checkerboard)"
            echo ""
            echo "Examples:"
            echo "  $0 observation_deck 'luxury observation lounge with curved windows'"
            echo "  $0 bridge_frame 'spaceship bridge window frame only' --mode checker"
            echo ""
            echo "The script will:"
            echo "  1. Generate image with AI (prompting for appropriate window style)"
            echo "  2. Convert detected regions to true alpha transparency"
            echo "  3. Save to assets/decks/<name>.png"
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
    echo "Usage: $0 <name> <description> [--mode black|checker]"
    echo "Run with --help for more information"
    exit 1
fi

# Build prompts based on mode
case "$MODE" in
    black)
        PROMPT="Sci-fi spaceship interior scene: ${DESCRIPTION}.
CRITICAL: All windows showing space must be PURE BLACK (#000000).
The window glass areas should be solid black with no stars, no reflections, no gradients.
Interior should have warm ambient lighting with detailed sci-fi aesthetic.
The black window areas will be used as a mask for compositing a dynamic starfield.
Art style: Detailed illustration, warm orange/amber interior lighting, art-deco sci-fi."
        ;;
    checker)
        PROMPT="PNG with alpha transparency. Spaceship interior frame: ${DESCRIPTION}.
CRITICAL: Window openings must be COMPLETELY TRANSPARENT (alpha=0, rendered as transparency).
Only render the solid interior surfaces and frames.
Leave window/viewport areas as transparent cutouts for compositing.
This is for layering over a starfield background.
Art style: Detailed illustration, warm orange/amber interior lighting, art-deco sci-fi."
        ;;
    *)
        echo "Unknown mode: $MODE (use 'black' or 'checker')"
        exit 1
        ;;
esac

echo "=== Transparent Scene Generator ==="
echo "Name: $NAME"
echo "Mode: $MODE"
echo "Description: $DESCRIPTION"
echo ""

# Step 1: Generate with AI
echo "[1/3] Generating image with AI..."
OUTPUT=$(go run "$PROJECT_ROOT/cmd/voyage" ai -generate-image -prompt "$PROMPT" 2>&1)
echo "$OUTPUT"

# Extract generated file path
GENERATED=$(echo "$OUTPUT" | grep -o 'assets/generated/response_[0-9]*.png' | head -1)
if [[ -z "$GENERATED" ]]; then
    echo "ERROR: Could not find generated file"
    exit 1
fi

echo ""
echo "[2/3] Converting to transparent PNG..."

# Step 2: Convert to transparent
TEMP_OUTPUT="/tmp/${NAME}_transparent.png"
go run "$PROJECT_ROOT/cmd/generate-window-mask" -mode="$MODE" "$PROJECT_ROOT/$GENERATED" "$TEMP_OUTPUT"

# Step 3: Move to final location
FINAL_DIR="$PROJECT_ROOT/assets/decks"
mkdir -p "$FINAL_DIR"
FINAL_PATH="$FINAL_DIR/${NAME}.png"

echo ""
echo "[3/3] Saving to $FINAL_PATH..."
cp "$TEMP_OUTPUT" "$FINAL_PATH"

# Verify alpha
echo ""
echo "=== Result ==="
sips -g hasAlpha -g pixelWidth -g pixelHeight "$FINAL_PATH" 2>/dev/null | grep -E "(hasAlpha|pixel)"
echo ""
echo "Generated: $GENERATED (original)"
echo "Output: $FINAL_PATH (with transparency)"
echo ""
echo "To preview: open $FINAL_PATH"
