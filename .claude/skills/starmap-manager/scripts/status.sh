#!/bin/bash
# Show current starmap asset status
# Usage: status.sh

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../../.." && pwd)"
RAW_DIR="$PROJECT_ROOT/assets/data/raw"
OUTPUT_DIR="$PROJECT_ROOT/assets/data/starmap"
BG_DIR="$OUTPUT_DIR/background"

echo "=== Starmap Asset Status ==="
echo "Project: $PROJECT_ROOT"
echo ""

# Check raw data
echo "Raw Data ($RAW_DIR):"
if [ -d "$RAW_DIR" ]; then
    if ls "$RAW_DIR"/* &>/dev/null 2>&1; then
        for file in "$RAW_DIR"/*; do
            SIZE=$(du -h "$file" | cut -f1)
            echo "  $(basename "$file"): $SIZE"
        done
    else
        echo "  (empty)"
    fi
else
    echo "  (directory not found)"
fi

echo ""

# Check processed data
echo "Processed Data ($OUTPUT_DIR):"
if [ -d "$OUTPUT_DIR" ]; then
    # Check stars.json
    if [ -f "$OUTPUT_DIR/stars.json" ]; then
        SIZE=$(du -h "$OUTPUT_DIR/stars.json" | cut -f1)
        COUNT=$(python3 -c "import json; print(json.load(open('$OUTPUT_DIR/stars.json'))['count'])" 2>/dev/null || echo "?")
        SOURCE=$(python3 -c "import json; print(json.load(open('$OUTPUT_DIR/stars.json'))['source'])" 2>/dev/null || echo "?")
        echo "  stars.json: $SIZE ($COUNT stars, source: $SOURCE)"
    else
        echo "  stars.json: NOT FOUND"
    fi

    # Check exoplanets.json
    if [ -f "$OUTPUT_DIR/exoplanets.json" ]; then
        SIZE=$(du -h "$OUTPUT_DIR/exoplanets.json" | cut -f1)
        COUNT=$(python3 -c "import json; print(json.load(open('$OUTPUT_DIR/exoplanets.json'))['count'])" 2>/dev/null || echo "?")
        echo "  exoplanets.json: $SIZE ($COUNT planets)"
    else
        echo "  exoplanets.json: NOT FOUND"
    fi

    # Check habitable.json
    if [ -f "$OUTPUT_DIR/habitable.json" ]; then
        SIZE=$(du -h "$OUTPUT_DIR/habitable.json" | cut -f1)
        COUNT=$(python3 -c "import json; print(json.load(open('$OUTPUT_DIR/habitable.json'))['count'])" 2>/dev/null || echo "?")
        echo "  habitable.json: $SIZE ($COUNT HZ candidates)"
    else
        echo "  habitable.json: NOT FOUND"
    fi
else
    echo "  (directory not found)"
fi

echo ""

# Check background images
echo "Background Images ($BG_DIR):"
if [ -d "$BG_DIR" ]; then
    HAS_IMAGES=false
    for ext in png jpg jpeg; do
        if ls "$BG_DIR"/*.$ext &>/dev/null 2>&1; then
            HAS_IMAGES=true
            for file in "$BG_DIR"/*.$ext; do
                SIZE=$(du -h "$file" | cut -f1)
                # Try to get dimensions if 'file' command available
                DIMS=$(file "$file" 2>/dev/null | grep -oE '[0-9]+ ?x ?[0-9]+' | head -1 || echo "")
                if [ -n "$DIMS" ]; then
                    echo "  $(basename "$file"): $SIZE ($DIMS)"
                else
                    echo "  $(basename "$file"): $SIZE"
                fi
            done
        fi
    done
    if [ "$HAS_IMAGES" = false ]; then
        echo "  (no image files)"
    fi
else
    echo "  (directory not found)"
fi

echo ""

# Calculate total size
TOTAL_RAW=0
TOTAL_PROCESSED=0

if [ -d "$RAW_DIR" ]; then
    TOTAL_RAW=$(du -sh "$RAW_DIR" 2>/dev/null | cut -f1 || echo "0")
fi

if [ -d "$OUTPUT_DIR" ]; then
    TOTAL_PROCESSED=$(du -sh "$OUTPUT_DIR" 2>/dev/null | cut -f1 || echo "0")
fi

echo "Total Sizes:"
echo "  Raw data: $TOTAL_RAW"
echo "  Processed: $TOTAL_PROCESSED"

echo ""

# Recommendations
echo "Recommendations:"
if [ ! -f "$OUTPUT_DIR/stars.json" ]; then
    echo "  Run: .claude/skills/starmap-manager/scripts/download_stars.sh quick"
    echo "  Then: .claude/skills/starmap-manager/scripts/process_stars.sh"
elif [ ! -f "$OUTPUT_DIR/exoplanets.json" ]; then
    echo "  Run: .claude/skills/starmap-manager/scripts/download_exoplanets.sh"
    echo "  Then: .claude/skills/starmap-manager/scripts/process_stars.sh"
elif [ ! -d "$BG_DIR" ] || [ ! "$(ls -A "$BG_DIR" 2>/dev/null)" ]; then
    echo "  Run: .claude/skills/starmap-manager/scripts/download_background.sh"
else
    echo "  All assets present!"

    # Check tier
    SOURCE=$(python3 -c "import json; print(json.load(open('$OUTPUT_DIR/stars.json'))['source'])" 2>/dev/null || echo "")
    if [ "$SOURCE" = "cns5" ]; then
        echo "  To upgrade to medium tier: .claude/skills/starmap-manager/scripts/download_stars.sh medium"
    fi
fi
