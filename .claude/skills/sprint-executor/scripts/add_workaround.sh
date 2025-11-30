#!/bin/bash
# add_workaround.sh - Add a new AILANG workaround to tracking
# Usage: add_workaround.sh "<problem>" "<error_message>" "<workaround>"
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../../.." && pwd)"
SKILL_MD="$PROJECT_ROOT/.claude/skills/sprint-executor/SKILL.md"
CLAUDE_MD="$PROJECT_ROOT/CLAUDE.md"

usage() {
    echo "Usage: add_workaround.sh <problem> <error_message> <workaround>"
    echo ""
    echo "Adds a workaround to both SKILL.md and CLAUDE.md"
    echo ""
    echo "Arguments:"
    echo "  problem       - Short description (e.g., 'Nested field access')"
    echo "  error_message - The error you see (e.g., 'cannot unify open record')"
    echo "  workaround    - How to fix it (e.g., 'Break into let bindings')"
    echo ""
    echo "Example:"
    echo "  add_workaround.sh 'Nested field access' 'TVar2 error' 'Use let bindings'"
    exit 1
}

if [ $# -lt 3 ]; then
    usage
fi

PROBLEM="$1"
ERROR_MSG="$2"
WORKAROUND="$3"

echo "Adding workaround..."
echo "  Problem:   $PROBLEM"
echo "  Error:     $ERROR_MSG"
echo "  Workaround: $WORKAROUND"
echo ""

# Check files exist
if [ ! -f "$SKILL_MD" ]; then
    echo "ERROR: SKILL.md not found at $SKILL_MD"
    exit 1
fi

if [ ! -f "$CLAUDE_MD" ]; then
    echo "ERROR: CLAUDE.md not found at $CLAUDE_MD"
    exit 1
fi

# Add to SKILL.md workarounds table
# Find the line with "ADT in inline tests" (last row) and add after it
if grep -q "| ADT in inline tests" "$SKILL_MD"; then
    # Use sed to add new row after the ADT line
    sed -i.bak "/| ADT in inline tests.*/a\\
| $PROBLEM | \"$ERROR_MSG\" | $WORKAROUND |" "$SKILL_MD"
    rm -f "$SKILL_MD.bak"
    echo "✓ Added to SKILL.md workarounds table"
else
    echo "⚠ Could not find workarounds table in SKILL.md - add manually"
fi

# Add to CLAUDE.md Known Limitations
# Find the line before "### Design Choices" and add the new limitation
if grep -q "### Design Choices" "$CLAUDE_MD"; then
    sed -i.bak "/### Design Choices/i\\
- **$PROBLEM** - $WORKAROUND" "$CLAUDE_MD"
    rm -f "$CLAUDE_MD.bak"
    echo "✓ Added to CLAUDE.md Known Limitations"
else
    echo "⚠ Could not find Design Choices section in CLAUDE.md - add manually"
fi

echo ""
echo "Done! Remember to:"
echo "1. Report via ailang-feedback if not already done"
echo "2. Commit these documentation changes"
