#!/bin/bash
# Iteratively improve window mask using AI segmentation
# Uses Gemini vision for detection, validates quality, refines prompt if needed
#
# Usage: iterate_mask.sh <image.png> [deck_type] [max_iterations]
#
# UPDATED: Now uses AI segmentation instead of threshold-based detection

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../../.." && pwd)"

cd "$PROJECT_ROOT"

if [[ $# -lt 1 ]]; then
    echo "Usage: $0 <image.png> [deck_type] [max_iterations]"
    echo ""
    echo "Iteratively generate and validate window mask using AI segmentation"
    echo ""
    echo "Arguments:"
    echo "  image.png       Input image"
    echo "  deck_type       observation, bridge, or generic (default: auto-detect)"
    echo "  max_iterations  Maximum refinement attempts (default: 3)"
    echo ""
    echo "Examples:"
    echo "  $0 assets/decks/observation.png"
    echo "  $0 assets/decks/bridge.png bridge 5"
    echo ""
    echo "NOTE: Old threshold-based syntax is deprecated."
    echo "      Use generate_mask_ai.sh for single-shot generation."
    exit 1
fi

IMAGE_PATH="$1"

# Handle legacy threshold argument (ignore it, show warning)
if [[ "$2" =~ ^[0-9]+$ ]]; then
    echo "WARNING: Threshold argument ignored - now using AI segmentation"
    DECK_TYPE="${3:-}"
    MAX_ITERATIONS="${4:-3}"
else
    DECK_TYPE="${2:-}"
    MAX_ITERATIONS="${3:-3}"
fi

# Auto-detect deck type
if [[ -z "$DECK_TYPE" ]]; then
    BASENAME=$(basename "$IMAGE_PATH" | tr '[:upper:]' '[:lower:]')
    if [[ "$BASENAME" == *"observation"* ]] || [[ "$BASENAME" == *"dome"* ]]; then
        DECK_TYPE="observation"
    elif [[ "$BASENAME" == *"bridge"* ]] || [[ "$BASENAME" == *"command"* ]]; then
        DECK_TYPE="bridge"
    else
        DECK_TYPE="generic"
    fi
fi

echo "=== AI Mask Generation ==="
echo "Image: $IMAGE_PATH"
echo "Deck type: $DECK_TYPE"
echo "Max iterations: $MAX_ITERATIONS"
echo ""

# Build tools if needed
if [[ ! -f "bin/generate-window-mask-ai" ]]; then
    echo "Building generate-window-mask-ai..."
    go build -o bin/generate-window-mask-ai ./cmd/generate-window-mask-ai
fi

ITERATION=1
BEST_SCORE=0
BEST_MASK=""
TEMP_DIR="/tmp/mask_iterate_$$"
mkdir -p "$TEMP_DIR"

# Cleanup on exit
trap "rm -rf $TEMP_DIR" EXIT

# Refinement prompts for each iteration
declare -a PROMPTS
PROMPTS[0]="" # Use default deck prompt
PROMPTS[1]="Focus on the LARGEST transparent areas only. Ignore small lights, screens, or reflections. The main window should be one large region."
PROMPTS[2]="Detect the main window or dome area where SPACE is visible. This should be ONE or TWO large polygons covering the sky/space areas."
PROMPTS[3]="Look for where stars would shine through the hull. Ignore interior elements. Return only the transparent viewport regions."

while [[ $ITERATION -le $MAX_ITERATIONS ]]; do
    echo "--- Iteration $ITERATION ---"

    TEMP_MASK="$TEMP_DIR/mask_$ITERATION.png"

    # Build command with optional custom prompt
    CMD="bin/generate-window-mask-ai -deck $DECK_TYPE -overlay -o $TEMP_MASK"

    PROMPT_IDX=$((ITERATION - 1))
    if [[ $PROMPT_IDX -lt ${#PROMPTS[@]} ]] && [[ -n "${PROMPTS[$PROMPT_IDX]}" ]]; then
        echo "Refined prompt: ${PROMPTS[$PROMPT_IDX]:0:60}..."
        CMD="$CMD -prompt \"${PROMPTS[$PROMPT_IDX]}\""
    fi

    CMD="$CMD $IMAGE_PATH"

    echo "Generating mask..."
    OUTPUT=$(eval $CMD 2>&1) || true
    echo "$OUTPUT" | grep -E "(Detected|Mask:|coverage)" || true

    if [[ ! -f "$TEMP_MASK" ]]; then
        echo "Error: Failed to generate mask"
        ITERATION=$((ITERATION + 1))
        continue
    fi

    # Check coverage from output
    COVERAGE=$(echo "$OUTPUT" | grep -o "[0-9.]*% coverage" | grep -o "[0-9.]*" || echo "0")
    REGIONS=$(echo "$OUTPUT" | grep -o "Detected [0-9]* region" | grep -o "[0-9]*" || echo "0")

    echo "Coverage: ${COVERAGE}%, Regions: $REGIONS"

    # Simple scoring: good coverage (10-60%) with few regions (1-3)
    SCORE=5
    if (( $(echo "$COVERAGE > 5" | bc -l) )) && (( $(echo "$COVERAGE < 70" | bc -l) )); then
        SCORE=$((SCORE + 3))
    fi
    if [[ $REGIONS -ge 1 ]] && [[ $REGIONS -le 5 ]]; then
        SCORE=$((SCORE + 2))
    fi

    echo "Score: $SCORE/10"

    # Track best
    if [[ $SCORE -gt $BEST_SCORE ]]; then
        BEST_SCORE=$SCORE
        BEST_MASK="$TEMP_MASK"
        cp "$TEMP_MASK" "$TEMP_DIR/best_mask.png"
        OVERLAY_FILE="${TEMP_MASK%.png}_overlay.png"
        if [[ -f "$OVERLAY_FILE" ]]; then
            cp "$OVERLAY_FILE" "$TEMP_DIR/best_overlay.png"
        fi
        echo "New best! (score: $SCORE)"
    fi

    # Stop if good enough
    if [[ $SCORE -ge 8 ]]; then
        echo ""
        echo "Good quality achieved! Stopping iteration."
        break
    fi

    ITERATION=$((ITERATION + 1))
    echo ""
done

echo ""
echo "=== Results ==="

if [[ -f "$TEMP_DIR/best_mask.png" ]]; then
    # Determine final output path
    BASENAME=$(basename "$IMAGE_PATH" .png)
    DIRNAME=$(dirname "$IMAGE_PATH")

    # Put mask next to source or in assets/decks for generated images
    if [[ "$DIRNAME" == *"generated"* ]]; then
        FINAL_DIR="assets/decks"
    else
        FINAL_DIR="$DIRNAME"
    fi

    FINAL_MASK="$FINAL_DIR/${BASENAME}_mask.png"

    echo "Best score: $BEST_SCORE/10"
    echo ""
    echo "Files:"
    echo "  Mask:    $TEMP_DIR/best_mask.png"
    if [[ -f "$TEMP_DIR/best_overlay.png" ]]; then
        echo "  Overlay: $TEMP_DIR/best_overlay.png"
    fi
    echo ""
    echo "To use this mask:"
    echo "  cp $TEMP_DIR/best_mask.png $FINAL_MASK"
    echo ""

    # Interactive mode - ask to copy
    if [[ -t 0 ]]; then
        read -p "Copy mask to $FINAL_MASK? [y/N] " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            mkdir -p "$FINAL_DIR"
            cp "$TEMP_DIR/best_mask.png" "$FINAL_MASK"
            echo "Copied to: $FINAL_MASK"
        fi
    fi
else
    echo "No mask was generated successfully."
    exit 1
fi
