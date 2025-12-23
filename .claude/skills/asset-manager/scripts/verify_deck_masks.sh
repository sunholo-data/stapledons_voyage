#!/bin/bash
# Verify that deck background masks match their source images
# Checks dimensions and provides regeneration commands using AI segmentation

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../../.." && pwd)"

cd "$PROJECT_ROOT"

echo "=== Deck Mask Verification ==="
echo ""

# Build tool if needed
if [[ ! -f "bin/generate-window-mask-ai" ]]; then
    echo "Building generate-window-mask-ai..."
    go build -o bin/generate-window-mask-ai ./cmd/generate-window-mask-ai
    echo ""
fi

check_deck() {
    local NAME="$1"
    local BG_PATH="$2"
    local MASK_PATH="$3"
    local DECK_TYPE="$4"
    local SPRITE_ID="$5"

    echo "$NAME:"

    if [[ ! -f "$BG_PATH" ]]; then
        echo "  ❌ Background not found: $BG_PATH"
        return
    fi

    if [[ ! -f "$MASK_PATH" ]]; then
        echo "  ❌ Mask not found: $MASK_PATH"
        echo "     To generate (AI):"
        echo "       bin/generate-window-mask-ai -deck $DECK_TYPE -overlay -o $MASK_PATH $BG_PATH"
        return
    fi

    # Get dimensions
    BG_INFO=$(file "$BG_PATH")
    MASK_INFO=$(file "$MASK_PATH")

    BG_DIMS=$(echo "$BG_INFO" | grep -o '[0-9]* x [0-9]*' | head -1)
    MASK_DIMS=$(echo "$MASK_INFO" | grep -o '[0-9]* x [0-9]*' | head -1)

    if [[ "$BG_DIMS" == "$MASK_DIMS" ]]; then
        echo "  ✅ Dimensions match: $BG_DIMS"
    else
        echo "  ❌ Dimension mismatch:"
        echo "     Background: $BG_DIMS"
        echo "     Mask:       $MASK_DIMS"
        echo "     To fix (AI):"
        echo "       bin/generate-window-mask-ai -deck $DECK_TYPE -overlay -o $MASK_PATH $BG_PATH"
    fi

    # Check sprite manifest if sprite ID provided
    if [[ -n "$SPRITE_ID" ]] && [[ -f "assets/sprites/manifest.json" ]]; then
        MANIFEST_LINE=$(grep "\"$SPRITE_ID\"" assets/sprites/manifest.json 2>/dev/null || true)
        if [[ -n "$MANIFEST_LINE" ]]; then
            # Extract width and height from manifest
            MANIFEST_W=$(echo "$MANIFEST_LINE" | grep -o '"width": *[0-9]*' | grep -o '[0-9]*')
            MANIFEST_H=$(echo "$MANIFEST_LINE" | grep -o '"height": *[0-9]*' | grep -o '[0-9]*')

            if [[ -n "$MANIFEST_W" ]] && [[ -n "$MANIFEST_H" ]]; then
                MANIFEST_WH="${MANIFEST_W} x ${MANIFEST_H}"
                if [[ "$BG_DIMS" == "$MANIFEST_WH" ]]; then
                    echo "  ✅ Manifest sprite $SPRITE_ID matches: $MANIFEST_WH"
                else
                    echo "  ⚠️  Manifest mismatch:"
                    echo "     Actual:   $BG_DIMS"
                    echo "     Manifest: $MANIFEST_WH"
                    echo "     Update manifest or regenerate image"
                fi
            fi
        fi
    fi
}

# Check all known decks
check_deck "Bridge Deck" \
    "assets/decks/bridge/background.png" \
    "assets/decks/bridge/window_mask_large.png" \
    "bridge" \
    "9000"

echo ""

check_deck "Observation Deck" \
    "assets/decks/observation.png" \
    "assets/decks/observation/window_mask_large.png" \
    "observation" \
    "9001"

echo ""

# Check for any other decks in assets/decks/
echo "Other deck images:"
for IMG in assets/decks/*.png; do
    if [[ -f "$IMG" ]]; then
        BASENAME=$(basename "$IMG" .png)
        # Skip if it's a mask file
        if [[ "$BASENAME" == *"_mask"* ]]; then
            continue
        fi
        # Skip known decks
        if [[ "$BASENAME" == "bridge" ]] || [[ "$BASENAME" == "observation" ]]; then
            continue
        fi

        MASK_PATH="assets/decks/${BASENAME}_mask.png"
        if [[ -f "$MASK_PATH" ]]; then
            echo "  ✅ $BASENAME has mask"
        else
            echo "  ⚠️  $BASENAME - no mask found"
            echo "     To generate: bin/generate-window-mask-ai -deck generic -o $MASK_PATH $IMG"
        fi
    fi
done

echo ""
echo "=== Verification Complete ==="
echo ""
echo "Quick commands:"
echo "  Regenerate all masks:  for f in assets/decks/*.png; do [[ ! \"\$f\" == *_mask* ]] && bin/generate-window-mask-ai -overlay \"\$f\"; done"
echo "  Single mask:           bin/generate-window-mask-ai -deck observation assets/decks/observation.png"
