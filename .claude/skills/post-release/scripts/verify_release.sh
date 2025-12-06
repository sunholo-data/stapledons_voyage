#!/usr/bin/env bash
# Verify game is ready for release
# For Stapledons Voyage

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../../.." && pwd)"

VERSION="${1:-}"

echo "Verifying release readiness..."
echo

FAILURES=0
WARNINGS=0

# 1. Check AILANG modules compile
echo "1/6 Checking AILANG modules..."
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
echo "2/6 Testing AILANG runtime..."
if ailang run --entry init_world sim/step.ail > /tmp/ailang_run.log 2>&1; then
    echo "  ✓ init_world runs"
else
    echo "  ⚠ init_world has issues"
    WARNINGS=$((WARNINGS + 1))
fi
echo

# 3. Check game build
echo "3/6 Checking game build..."
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
echo "4/6 Checking AILANG messages..."
if ailang messages list --unread 2>&1 | grep -q "message"; then
    echo "  ⚠ Pending messages - review before release"
    WARNINGS=$((WARNINGS + 1))
else
    echo "  ✓ No pending messages"
fi
echo

# 5. Check design docs organization
echo "5/6 Checking design docs..."
if [[ -f "$PROJECT_ROOT/scripts/validate_design_docs.sh" ]]; then
    # Run validation and capture output
    MISPLACED=$(bash "$PROJECT_ROOT/scripts/validate_design_docs.sh" 2>&1 | grep -c "should be in" || true)
    ORPHANS=$(bash "$PROJECT_ROOT/scripts/validate_design_docs.sh" 2>&1 | grep -c "ORPHAN DOCS" || true)

    if [[ $MISPLACED -gt 0 ]]; then
        echo "  ⚠ $MISPLACED misplaced doc(s) - run: scripts/validate_design_docs.sh"
        WARNINGS=$((WARNINGS + 1))
    else
        echo "  ✓ No misplaced docs"
    fi

    if [[ $ORPHANS -gt 0 ]]; then
        echo "  ⚠ Orphan docs exist (not in version folders)"
        WARNINGS=$((WARNINGS + 1))
    fi
else
    echo "  - Validation script not found"
fi
echo

# 6. Check CHANGELOG
echo "6/6 Checking CHANGELOG..."
if [[ -f "$PROJECT_ROOT/CHANGELOG.md" ]]; then
    if [[ -n "$VERSION" ]]; then
        if grep -q "\[$VERSION\]" "$PROJECT_ROOT/CHANGELOG.md"; then
            echo "  ✓ CHANGELOG has $VERSION entry"
        else
            echo "  ⚠ CHANGELOG missing $VERSION entry"
            WARNINGS=$((WARNINGS + 1))
        fi
    else
        echo "  ✓ CHANGELOG exists"
    fi
else
    echo "  ⚠ CHANGELOG.md not found"
    WARNINGS=$((WARNINGS + 1))
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
