#!/bin/bash
# Generate a bug report design doc for visual regression
# Usage: report_bug.sh <scenario-name> "<description>"

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../../.." && pwd)"
BUGS_DIR="$PROJECT_ROOT/design_docs/bugs"

if [ $# -lt 2 ]; then
    echo "Usage: report_bug.sh <scenario-name> \"<description>\""
    echo ""
    echo "Example:"
    echo "  report_bug.sh camera-zoom \"Zoom out produces artifacts at tile edges\""
    exit 1
fi

SCENARIO="$1"
DESCRIPTION="$2"
DATE=$(date +%Y-%m-%d)
SLUG=$(echo "$DESCRIPTION" | tr '[:upper:]' '[:lower:]' | sed 's/[^a-z0-9]/-/g' | sed 's/--*/-/g' | sed 's/^-//' | sed 's/-$//' | cut -c1-40)
FILENAME="$BUGS_DIR/${DATE}-${SLUG}.md"

# Create bugs directory
mkdir -p "$BUGS_DIR"

# Generate bug report
cat > "$FILENAME" << EOF
# Bug: $DESCRIPTION

**Date:** $DATE
**Scenario:** $SCENARIO
**Status:** Open

## Description

$DESCRIPTION

## Reproduction

1. Run the test scenario:
   \`\`\`bash
   go run ./cmd/game -scenario $SCENARIO -test-mode
   \`\`\`

2. Compare against golden files:
   \`\`\`bash
   .claude/skills/test-manager/scripts/compare_golden.sh $SCENARIO
   \`\`\`

## Expected Behavior

[Describe what should happen based on golden files]

## Actual Behavior

[Describe what actually happens]

## Screenshots

**Golden (expected):**
![Golden](../../golden/$SCENARIO/[filename].png)

**Actual:**
![Actual](../../out/test/$SCENARIO/[filename].png)

**Diff:**
![Diff](../../out/test/$SCENARIO/diff/[filename].png)

## Analysis

[Root cause analysis]

## Fix

[Proposed fix]

## Verification

- [ ] Fix implemented
- [ ] Tests pass
- [ ] Golden files updated (if behavior change is intentional)
- [ ] Code reviewed
EOF

echo "Bug report created: $FILENAME"
echo ""
echo "Next steps:"
echo "  1. Edit the bug report to fill in details"
echo "  2. Investigate the root cause"
echo "  3. Implement a fix"
echo "  4. Re-run tests to verify"
