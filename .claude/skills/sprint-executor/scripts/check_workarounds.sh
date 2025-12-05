#!/bin/bash
# check_workarounds.sh - Check for AILANG fixes and verify workarounds
# Usage: check_workarounds.sh
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../../.." && pwd)"

echo "=== AILANG Workaround Status Check ==="
echo ""

# 1. Check inbox for fix notifications
echo "📬 Checking AILANG messages..."
echo ""
if command -v ailang &> /dev/null; then
    ailang messages list --unread 2>/dev/null || echo "  (No unread messages)"
else
    echo "  ⚠ ailang not installed"
fi
echo ""

# 2. Type-check all AILANG files
echo "🔍 Type-checking AILANG modules..."
echo ""
cd "$PROJECT_ROOT"
ERRORS=0
for f in sim/*.ail; do
    if [ -f "$f" ]; then
        if ailang check "$f" 2>&1 | grep -q "No errors"; then
            echo "  ✓ $f"
        else
            echo "  ✗ $f"
            ERRORS=$((ERRORS + 1))
        fi
    fi
done
echo ""

if [ $ERRORS -eq 0 ]; then
    echo "✅ All modules compile successfully"
else
    echo "⚠ $ERRORS module(s) have errors"
fi
echo ""

# 3. Show current workarounds being tracked
echo "📋 Current workarounds in SKILL.md:"
echo ""
grep -A 20 "| Problem |" "$PROJECT_ROOT/.claude/skills/sprint-executor/SKILL.md" 2>/dev/null | head -10 || echo "  (Could not read workarounds table)"
echo ""

# 4. Suggestions
echo "💡 Next steps:"
echo "   - If inbox has fix notifications, test them and run:"
echo "     mark_fixed.sh '<keyword>' '<version>'"
echo "   - If you found a new issue, run:"
echo "     add_workaround.sh '<problem>' '<error>' '<workaround>'"
echo "   - Report new issues via:"
echo "     ~/.claude/skills/ailang-feedback/scripts/send_feedback.sh bug ..."
