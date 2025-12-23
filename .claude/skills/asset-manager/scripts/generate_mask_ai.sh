#!/bin/bash
# Generate window mask using AI segmentation
# Replaces threshold-based detection with Gemini vision
#
# Usage: generate_mask_ai.sh <image.png> [output_mask.png] [deck_type]
#
# Deck types:
#   observation - Large dome/panoramic windows
#   bridge      - Command center viewports
#   generic     - Any spaceship interior (default)

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../../.." && pwd)"

cd "$PROJECT_ROOT"

if [[ $# -lt 1 ]]; then
    echo "Usage: $0 <image.png> [output_mask.png] [deck_type]"
    echo ""
    echo "Generate window mask using AI segmentation (Gemini 2.5)"
    echo ""
    echo "Arguments:"
    echo "  image.png        Input image to generate mask for"
    echo "  output_mask.png  Output mask path (default: <input>_mask.png)"
    echo "  deck_type        observation, bridge, or generic (default: auto-detect)"
    echo ""
    echo "Examples:"
    echo "  $0 assets/decks/observation.png"
    echo "  $0 assets/decks/bridge.png assets/decks/bridge_mask.png bridge"
    echo "  $0 assets/generated/response_123.png out/mask.png"
    exit 1
fi

IMAGE_PATH="$1"
OUTPUT_PATH="${2:-}"
DECK_TYPE="${3:-}"

# Auto-detect deck type from filename if not specified
if [[ -z "$DECK_TYPE" ]]; then
    BASENAME=$(basename "$IMAGE_PATH" | tr '[:upper:]' '[:lower:]')
    if [[ "$BASENAME" == *"observation"* ]] || [[ "$BASENAME" == *"dome"* ]]; then
        DECK_TYPE="observation"
    elif [[ "$BASENAME" == *"bridge"* ]] || [[ "$BASENAME" == *"command"* ]]; then
        DECK_TYPE="bridge"
    else
        DECK_TYPE="generic"
    fi
    echo "Auto-detected deck type: $DECK_TYPE"
fi

# Build the tool if needed
if [[ ! -f "bin/generate-window-mask-ai" ]]; then
    echo "Building generate-window-mask-ai..."
    go build -o bin/generate-window-mask-ai ./cmd/generate-window-mask-ai
fi

# Build command
CMD="bin/generate-window-mask-ai -deck $DECK_TYPE -overlay -v"

if [[ -n "$OUTPUT_PATH" ]]; then
    CMD="$CMD -o $OUTPUT_PATH"
fi

CMD="$CMD $IMAGE_PATH"

echo "Running: $CMD"
echo ""
eval $CMD

echo ""
echo "Done! Check the output files."
