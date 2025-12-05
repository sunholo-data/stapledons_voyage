#!/bin/bash
# mark_fixed.sh - Mark an AILANG issue as fixed
# Usage: mark_fixed.sh "<problem_keyword>" "<version>"
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../../.." && pwd)"
SKILL_MD="$PROJECT_ROOT/.claude/skills/sprint-executor/SKILL.md"
CLAUDE_MD="$PROJECT_ROOT/CLAUDE.md"

usage() {
    echo "Usage: mark_fixed.sh <problem_keyword> <version>"
    echo ""
    echo "Moves a workaround from Known Limitations to Fixed section"
    echo ""
    echo "Arguments:"
    echo "  problem_keyword - Unique text from the problem (e.g., 'Nested field')"
    echo "  version         - AILANG version that fixed it (e.g., 'v0.5.0')"
    echo ""
    echo "Example:"
    echo "  mark_fixed.sh 'tuple destructuring' 'v0.5.0'"
    exit 1
}

if [ $# -lt 2 ]; then
    usage
fi

KEYWORD="$1"
VERSION="$2"

echo "Marking as fixed in $VERSION..."
echo "  Searching for: $KEYWORD"
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

# Remove from SKILL.md workarounds table
if grep -qi "$KEYWORD" "$SKILL_MD"; then
    # Remove the line containing the keyword from the table
    sed -i.bak "/$KEYWORD/Id" "$SKILL_MD"
    rm -f "$SKILL_MD.bak"
    echo "✓ Removed from SKILL.md workarounds table"
else
    echo "⚠ '$KEYWORD' not found in SKILL.md table"
fi

# In CLAUDE.md:
# 1. Find and extract the limitation line
# 2. Remove from Known Limitations
# 3. Add to Fixed section

if grep -qi "$KEYWORD" "$CLAUDE_MD"; then
    # Extract the limitation text
    LIMITATION=$(grep -i "$KEYWORD" "$CLAUDE_MD" | grep "^-" | head -1)

    if [ -n "$LIMITATION" ]; then
        # Remove from Known Limitations
        sed -i.bak "/$KEYWORD/Id" "$CLAUDE_MD"

        # Check if Fixed section for this version exists
        if grep -q "### Fixed in $VERSION" "$CLAUDE_MD"; then
            # Add to existing version section
            sed -i.bak "/### Fixed in $VERSION/a\\
$LIMITATION" "$CLAUDE_MD"
            echo "✓ Moved to 'Fixed in $VERSION' section in CLAUDE.md"
        else
            # Check if any Fixed section exists
            if grep -q "### Fixed in v" "$CLAUDE_MD"; then
                # Add new version section before first Fixed section
                sed -i.bak "/### Fixed in v/i\\
### Fixed in $VERSION\\
\\
$LIMITATION\\
" "$CLAUDE_MD"
                echo "✓ Created 'Fixed in $VERSION' section in CLAUDE.md"
            else
                echo "⚠ No Fixed section found - add manually to CLAUDE.md"
            fi
        fi
        rm -f "$CLAUDE_MD.bak"
    fi
else
    echo "⚠ '$KEYWORD' not found in CLAUDE.md"
fi

echo ""
echo "Done! Remember to:"
echo "1. Verify the fix works: ailang check sim/*.ail"
echo "2. Remove workaround code if practical"
echo "3. Acknowledge the message: ailang messages ack <msg-id>"
echo "4. Commit these documentation changes"
