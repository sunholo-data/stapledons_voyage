#!/usr/bin/env bash
# Verify game is ready for release
# For Stapledons Voyage

set -euo pipefail

VERSION="${1:-}"

echo "Verifying release readiness..."
echo

FAILURES=0
WARNINGS=0

# 1. Check AILANG modules compile
echo "1/4 Checking AILANG modules..."
AILANG_FAILURES=0
for f in sim/*.ail; do
    if [[ -f "$f" ]]; then
        if ailang check "$f" > /tmp/ailang_check.log 2>&1; then
            echo "  ✓ $f"
        else
            echo "  ✗ $f"
            AILANG_FAILURES=$((AILANG_FAILURES + 1))
        fi
    fi
done
if [[ $AILANG_FAILURES -gt 0 ]]; then
    echo "  ✗ $AILANG_FAILURES module(s) fail"
    FAILURES=$((FAILURES + 1))
fi
echo

# 2. Test AILANG entry function
echo "2/4 Testing AILANG runtime..."
if ailang run --entry init_world sim/step.ail > /tmp/ailang_run.log 2>&1; then
    echo "  ✓ init_world runs"
else
    echo "  ⚠ init_world has issues"
    WARNINGS=$((WARNINGS + 1))
fi
echo

# 3. Check game build
echo "3/4 Checking game build..."
if [[ -f "Makefile" ]]; then
    if make game > /tmp/game_build.log 2>&1; then
        echo "  ✓ Game builds"
    else
        echo "  ✗ Game build failed"
        FAILURES=$((FAILURES + 1))
    fi
else
    echo "  - No Makefile"
fi
echo

# 4. Check AILANG messages
echo "4/4 Checking AILANG messages..."
if ailang messages list --unread 2>&1 | grep -q "message"; then
    echo "  ⚠ Pending messages - review before release"
    WARNINGS=$((WARNINGS + 1))
else
    echo "  ✓ No pending messages"
fi
echo

# Summary
echo "================================"
if [[ $FAILURES -eq 0 ]]; then
    if [[ $WARNINGS -gt 0 ]]; then
        echo "⚠ Release ready with $WARNINGS warning(s)"
    else
        echo "✓ Release verification passed!"
    fi
    echo ""
    if [[ -n "$VERSION" ]]; then
        echo "Ready to tag: git tag -a $VERSION -m 'Release $VERSION'"
    fi
    exit 0
else
    echo "✗ $FAILURES issue(s) - fix before release"
    exit 1
fi
