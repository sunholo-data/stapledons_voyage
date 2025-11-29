#!/bin/bash
# Display current core pillars for quick reference

DOCS_DIR="docs/vision"
FILE="$DOCS_DIR/core-pillars.md"

if [ ! -f "$FILE" ]; then
    echo "No core pillars defined yet."
    echo "Run init_vision_docs.sh and then conduct a vision interview."
    exit 1
fi

echo "=== CORE PILLARS ==="
echo ""
cat "$FILE"
