#!/bin/bash
# Log a design decision to design-decisions.md

set -e

if [ $# -lt 3 ]; then
    echo "Usage: log_decision.sh <title> <decision> <rationale>"
    echo ""
    echo "Example:"
    echo "  log_decision.sh \"No FTL Communication\" \"Players cannot send messages faster than light\" \"Reinforces time dilation consequences\""
    exit 1
fi

TITLE="$1"
DECISION="$2"
RATIONALE="$3"
DATE=$(date +%Y-%m-%d)

DOCS_DIR="docs/vision"
FILE="$DOCS_DIR/design-decisions.md"

if [ ! -f "$FILE" ]; then
    echo "Error: $FILE does not exist. Run init_vision_docs.sh first."
    exit 1
fi

# Append the decision
cat >> "$FILE" << EOF

## [$DATE] $TITLE

**Decision:** $DECISION

**Rationale:** $RATIONALE

EOF

echo "✓ Logged decision: $TITLE"
echo "  File: $FILE"
