#!/bin/bash
# Asset Manager - Update Manifest Script
# Scan assets directory and update manifest.json with new sprites

set -e

# Find project root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../../.." && pwd)"
SPRITES_DIR="$PROJECT_ROOT/assets/sprites"
MANIFEST="$SPRITES_DIR/manifest.json"

echo "=== Updating Sprite Manifest ==="
echo "Manifest: $MANIFEST"
echo ""

# Check if manifest exists
if [[ ! -f "$MANIFEST" ]]; then
    echo "Creating new manifest..."
    echo '{"sprites": {}}' > "$MANIFEST"
fi

# Read current manifest
CURRENT=$(cat "$MANIFEST")

# Function to get next available ID in range
get_next_id() {
    local min=$1
    local max=$2
    for ((id=min; id<=max; id++)); do
        if ! echo "$CURRENT" | jq -e ".sprites[\"$id\"]" > /dev/null 2>&1; then
            echo $id
            return
        fi
    done
    echo "-1"
}

# Function to check if file is in manifest
file_in_manifest() {
    local file=$1
    echo "$CURRENT" | jq -e ".sprites[] | select(.file == \"$file\")" > /dev/null 2>&1
}

# Track changes
ADDED=0
SKIPPED=0

# Scan iso_tiles (ID 1-99)
echo "Scanning iso_tiles/..."
for file in "$SPRITES_DIR/iso_tiles"/*.png; do
    [[ -f "$file" ]] || continue
    BASENAME=$(basename "$file")
    REL_PATH="iso_tiles/$BASENAME"

    if file_in_manifest "$REL_PATH"; then
        ((SKIPPED++))
        continue
    fi

    ID=$(get_next_id 1 99)
    if [[ "$ID" == "-1" ]]; then
        echo "  WARNING: No available IDs for tiles (1-99 full)"
        continue
    fi

    # Get image dimensions
    DIMS=$(file "$file" | grep -oE '[0-9]+ x [0-9]+' | head -1)
    WIDTH=$(echo "$DIMS" | cut -d' ' -f1)
    HEIGHT=$(echo "$DIMS" | cut -d' ' -f3)

    CURRENT=$(echo "$CURRENT" | jq ".sprites[\"$ID\"] = {\"file\": \"$REL_PATH\", \"width\": ${WIDTH:-64}, \"height\": ${HEIGHT:-32}, \"type\": \"tile\"}")
    echo "  Added: $REL_PATH (ID: $ID)"
    ((ADDED++))
done

# Scan iso_entities (ID 100-199)
echo "Scanning iso_entities/..."
for file in "$SPRITES_DIR/iso_entities"/*.png; do
    [[ -f "$file" ]] || continue
    BASENAME=$(basename "$file")
    REL_PATH="iso_entities/$BASENAME"

    if file_in_manifest "$REL_PATH"; then
        ((SKIPPED++))
        continue
    fi

    ID=$(get_next_id 100 199)
    if [[ "$ID" == "-1" ]]; then
        echo "  WARNING: No available IDs for entities (100-199 full)"
        continue
    fi

    # Entity sprites are typically 128x48 (4 frames of 32x48)
    CURRENT=$(echo "$CURRENT" | jq ".sprites[\"$ID\"] = {
        \"file\": \"$REL_PATH\",
        \"width\": 128,
        \"height\": 48,
        \"type\": \"entity\",
        \"frameWidth\": 32,
        \"frameHeight\": 48,
        \"animations\": {
            \"idle\": {\"startFrame\": 0, \"frameCount\": 1, \"fps\": 0},
            \"walk\": {\"startFrame\": 0, \"frameCount\": 4, \"fps\": 6}
        }
    }")
    echo "  Added: $REL_PATH (ID: $ID)"
    ((ADDED++))
done

# Scan stars (ID 200-299)
echo "Scanning stars/..."
for file in "$SPRITES_DIR/stars"/*.png; do
    [[ -f "$file" ]] || continue
    BASENAME=$(basename "$file")
    REL_PATH="stars/$BASENAME"

    if file_in_manifest "$REL_PATH"; then
        ((SKIPPED++))
        continue
    fi

    ID=$(get_next_id 200 299)
    if [[ "$ID" == "-1" ]]; then
        echo "  WARNING: No available IDs for stars (200-299 full)"
        continue
    fi

    CURRENT=$(echo "$CURRENT" | jq ".sprites[\"$ID\"] = {\"file\": \"$REL_PATH\", \"width\": 16, \"height\": 16, \"type\": \"star\"}")
    echo "  Added: $REL_PATH (ID: $ID)"
    ((ADDED++))
done

# Scan ui (ID 300-399)
if [[ -d "$SPRITES_DIR/ui" ]]; then
    echo "Scanning ui/..."
    for file in "$SPRITES_DIR/ui"/*.png; do
        [[ -f "$file" ]] || continue
        BASENAME=$(basename "$file")
        REL_PATH="ui/$BASENAME"

        if file_in_manifest "$REL_PATH"; then
            ((SKIPPED++))
            continue
        fi

        ID=$(get_next_id 300 399)
        if [[ "$ID" == "-1" ]]; then
            echo "  WARNING: No available IDs for UI (300-399 full)"
            continue
        fi

        DIMS=$(file "$file" | grep -oE '[0-9]+ x [0-9]+' | head -1)
        WIDTH=$(echo "$DIMS" | cut -d' ' -f1)
        HEIGHT=$(echo "$DIMS" | cut -d' ' -f3)

        CURRENT=$(echo "$CURRENT" | jq ".sprites[\"$ID\"] = {\"file\": \"$REL_PATH\", \"width\": ${WIDTH:-64}, \"height\": ${HEIGHT:-64}, \"type\": \"ui\"}")
        echo "  Added: $REL_PATH (ID: $ID)"
        ((ADDED++))
    done
fi

# Scan planets (ID 400-499)
if [[ -d "$SPRITES_DIR/planets" ]]; then
    echo "Scanning planets/..."
    for file in "$SPRITES_DIR/planets"/*.png; do
        [[ -f "$file" ]] || continue
        BASENAME=$(basename "$file")
        REL_PATH="planets/$BASENAME"

        if file_in_manifest "$REL_PATH"; then
            ((SKIPPED++))
            continue
        fi

        ID=$(get_next_id 400 499)
        if [[ "$ID" == "-1" ]]; then
            echo "  WARNING: No available IDs for planets (400-499 full)"
            continue
        fi

        CURRENT=$(echo "$CURRENT" | jq ".sprites[\"$ID\"] = {\"file\": \"$REL_PATH\", \"width\": 256, \"height\": 256, \"type\": \"planet\"}")
        echo "  Added: $REL_PATH (ID: $ID)"
        ((ADDED++))
    done
fi

# Scan portraits (ID 600-699)
if [[ -d "$SPRITES_DIR/portraits" ]]; then
    echo "Scanning portraits/..."
    for file in "$SPRITES_DIR/portraits"/*.png; do
        [[ -f "$file" ]] || continue
        BASENAME=$(basename "$file")
        REL_PATH="portraits/$BASENAME"

        if file_in_manifest "$REL_PATH"; then
            ((SKIPPED++))
            continue
        fi

        ID=$(get_next_id 600 699)
        if [[ "$ID" == "-1" ]]; then
            echo "  WARNING: No available IDs for portraits (600-699 full)"
            continue
        fi

        CURRENT=$(echo "$CURRENT" | jq ".sprites[\"$ID\"] = {\"file\": \"$REL_PATH\", \"width\": 128, \"height\": 128, \"type\": \"portrait\"}")
        echo "  Added: $REL_PATH (ID: $ID)"
        ((ADDED++))
    done
fi

# Write updated manifest
echo "$CURRENT" | jq '.' > "$MANIFEST"

echo ""
echo "=== Summary ==="
echo "Added: $ADDED new sprites"
echo "Skipped: $SKIPPED existing sprites"
echo "Manifest updated: $MANIFEST"
