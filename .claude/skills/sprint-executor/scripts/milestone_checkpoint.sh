#!/usr/bin/env bash
# Run checkpoint after completing a milestone
# Adapted for Stapledons Voyage (AILANG game project)

set -euo pipefail

MILESTONE_NAME="${1:-Unknown Milestone}"

echo "Running checkpoint for: $MILESTONE_NAME"
echo

FAILURES=0
WARNINGS=0

# 1. Check AILANG modules compile
echo "1/5 Checking AILANG modules..."
AILANG_FAILURES=0
for f in sim/*.ail; do
    if [[ -f "$f" ]]; then
        if ailang check "$f" > /tmp/ailang_check.log 2>&1; then
            echo "  ✓ $f"
        else
            echo "  ✗ $f FAILS"
            cat /tmp/ailang_check.log
            AILANG_FAILURES=$((AILANG_FAILURES + 1))
        fi
    fi
done
if [[ $AILANG_FAILURES -gt 0 ]]; then
    echo "  ✗ FIX BEFORE PROCEEDING!"
    FAILURES=$((FAILURES + 1))
fi
echo

# 2. Test AILANG entry functions
echo "2/5 Testing AILANG entry functions..."
if ailang run --entry init_world sim/step.ail > /tmp/ailang_run.log 2>&1; then
    echo "  ✓ init_world runs successfully"
else
    echo "  ⚠ init_world has runtime issues"
    cat /tmp/ailang_run.log | tail -5
    WARNINGS=$((WARNINGS + 1))
fi
echo

# 3. Show files changed
echo "3/5 Files changed in this milestone..."
if git diff --stat HEAD 2>/dev/null | tail -10; then
    :
else
    echo "No changes yet (or not a git repo)"
fi
echo

# 4. Check for AILANG issues to report
echo "4/5 AILANG feedback check..."
echo "  Did you encounter any issues during this milestone?"
echo "  - Unclear error messages"
echo "  - Missing features"
echo "  - Documentation gaps"
echo "  - Unexpected behavior"
echo ""
echo "  If yes, report with:"
echo "  ~/.claude/skills/ailang-feedback/scripts/send_feedback.sh <type> \"<title>\" \"<description>\" \"stapledons_voyage\""
echo

# 5. Check game build
echo "5/5 Checking game build..."
if [[ -f "Makefile" ]]; then
    if make game > /tmp/game_build.log 2>&1; then
        echo "  ✓ Game builds"
    else
        echo "  ⚠ Game build failed"
        WARNINGS=$((WARNINGS + 1))
    fi
else
    echo "  - No Makefile, skipping"
fi
echo

# Summary
echo "================================"
if [[ $FAILURES -eq 0 ]]; then
    if [[ $WARNINGS -gt 0 ]]; then
        echo "⚠ Milestone checkpoint passed with $WARNINGS warning(s)"
        echo "Can proceed, but review warnings."
    else
        echo "✓ Milestone checkpoint passed!"
        echo "Ready to proceed to next milestone."
    fi
    echo ""
    echo "Next steps:"
    echo "  1. Update sprint plan (mark milestone ✅)"
    echo "  2. Report any AILANG issues encountered"
    echo "  3. Document workarounds used"
    exit 0
else
    echo "✗ $FAILURES check(s) failed"
    echo "Fix AILANG errors before marking milestone complete."
    exit 1
fi
