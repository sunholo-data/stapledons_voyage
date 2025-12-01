#!/usr/bin/env bash
# Pre-release validation checklist
# Usage: .claude/skills/game-architect/scripts/pre_release_check.sh

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../../.." && pwd)"

cd "$PROJECT_ROOT"

ISSUES=0
WARNINGS=0

echo "Running pre-release checks..."
echo ""

# Check 1: Build succeeds
echo "1. Build check..."
if make game-mock >/dev/null 2>&1; then
    echo "  ✓ make game-mock succeeds"
else
    echo "  ✗ make game-mock failed"
    ISSUES=$((ISSUES + 1))
fi

# Check 2: Tests pass (if any)
echo ""
echo "2. Test check..."
if go test ./... >/dev/null 2>&1; then
    echo "  ✓ go test ./... passes"
else
    echo "  ✗ Tests failed"
    ISSUES=$((ISSUES + 1))
fi

# Check 3: No blocking TODOs
echo ""
echo "3. TODO/FIXME check..."

# Count TODOs by category
blocking=$(grep -rn "TODO.*BLOCK\|FIXME.*BLOCK\|XXX" --include="*.go" --include="*.ail" . 2>/dev/null | grep -v "/vendor/" | wc -l | tr -d ' ')
regular=$(grep -rn "TODO\|FIXME" --include="*.go" --include="*.ail" . 2>/dev/null | grep -v "/vendor/" | grep -v "BLOCK" | wc -l | tr -d ' ')

if [ "$blocking" -gt 0 ]; then
    echo "  ✗ Found $blocking blocking TODOs (TODO BLOCK, FIXME BLOCK, XXX):"
    grep -rn "TODO.*BLOCK\|FIXME.*BLOCK\|XXX" --include="*.go" --include="*.ail" . 2>/dev/null | grep -v "/vendor/" | head -5 | sed 's/^/      /'
    ISSUES=$((ISSUES + 1))
else
    echo "  ✓ No blocking TODOs"
fi

if [ "$regular" -gt 0 ]; then
    echo "  ! $regular regular TODOs (review before release)"
    WARNINGS=$((WARNINGS + 1))
fi

# Check 4: No debug code
echo ""
echo "4. Debug code check..."
debug_prints=$(grep -rn "fmt.Print\|log.Print\|println(" --include="*.go" . 2>/dev/null | grep -v "/vendor/" | grep -v "_test.go" | grep -v "// keep:" | wc -l | tr -d ' ')
if [ "$debug_prints" -gt 5 ]; then
    echo "  ! Found $debug_prints debug print statements (review):"
    grep -rn "fmt.Print\|log.Print\|println(" --include="*.go" . 2>/dev/null | grep -v "/vendor/" | grep -v "_test.go" | head -5 | sed 's/^/      /'
    WARNINGS=$((WARNINGS + 1))
else
    echo "  ✓ Debug prints OK ($debug_prints found)"
fi

# Check 5: Version consistency
echo ""
echo "5. Version check..."
if [ -f "version.go" ] || grep -q "Version\s*=" cmd/game/main.go 2>/dev/null; then
    echo "  ✓ Version defined"
else
    echo "  ! No version string found (consider adding)"
    WARNINGS=$((WARNINGS + 1))
fi

# Check 6: go.mod tidy
echo ""
echo "6. Dependencies check..."
if go mod tidy -v 2>&1 | grep -q "unused"; then
    echo "  ! go mod has unused dependencies (run 'go mod tidy')"
    WARNINGS=$((WARNINGS + 1))
else
    echo "  ✓ Dependencies OK"
fi

# Check 7: No commented-out code blocks
echo ""
echo "7. Commented code check..."
commented=$(grep -rn "^[[:space:]]*//" --include="*.go" . 2>/dev/null | grep -v "/vendor/" | grep -E "func |if |for |return " | wc -l | tr -d ' ')
if [ "$commented" -gt 10 ]; then
    echo "  ! Found ~$commented lines of commented-out code (clean up)"
    WARNINGS=$((WARNINGS + 1))
else
    echo "  ✓ Commented code OK"
fi

# Check 8: Design docs exist for version
echo ""
echo "8. Documentation check..."
if [ -d "design_docs/planned" ] && [ "$(find design_docs/planned -name "*.md" 2>/dev/null | wc -l)" -gt 0 ]; then
    echo "  ✓ Design docs present"
else
    echo "  ! No design docs found"
    WARNINGS=$((WARNINGS + 1))
fi

echo ""
echo "=========================================="
echo "Pre-Release Summary:"
echo "  Blocking issues: $ISSUES"
echo "  Warnings:        $WARNINGS"
echo "=========================================="

if [ $ISSUES -gt 0 ]; then
    echo ""
    echo "✗ Fix blocking issues before release"
    exit 1
fi

if [ $WARNINGS -gt 0 ]; then
    echo ""
    echo "! Review warnings before release"
fi

echo ""
echo "✓ Pre-release checks passed"
exit 0
