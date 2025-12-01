#!/usr/bin/env bash
# Runs all architecture validation checks
# Usage: .claude/skills/game-architect/scripts/validate_all.sh [--quick|--full]
#   --quick: Core checks only (file sizes, layer boundaries, structure, pre-release)
#   --full:  All checks including coverage, complexity, API stability (default)

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../../.." && pwd)"

cd "$PROJECT_ROOT"

MODE="${1:---full}"

echo "=========================================="
echo "  Game Architect - Full Validation"
echo "  Mode: $MODE"
echo "=========================================="
echo ""

FAILED=0
WARNINGS=0
CHECK_NUM=0

run_check() {
    local name="$1"
    local script="$2"
    local required="${3:-true}"

    CHECK_NUM=$((CHECK_NUM + 1))
    echo "$CHECK_NUM. $name..."

    if "$SCRIPT_DIR/$script" 2>&1 | sed 's/^/   /'; then
        echo ""
    else
        if [ "$required" = "true" ]; then
            FAILED=$((FAILED + 1))
        else
            WARNINGS=$((WARNINGS + 1))
        fi
        echo ""
    fi
}

# Core checks (always run)
echo "=== CORE CHECKS ==="
echo ""
run_check "File sizes" "check_file_sizes.sh" true
run_check "Layer boundaries" "check_layer_boundaries.sh" true
run_check "Directory structure" "check_structure.sh" true
run_check "Import cycles" "check_import_cycles.sh" true

if [ "$MODE" = "--full" ]; then
    echo ""
    echo "=== EXTENDED CHECKS ==="
    echo ""
    run_check "Code complexity" "check_complexity.sh" false
    run_check "AILANG/Go sync" "check_ailang_sync.sh" false
    run_check "API stability" "check_api_stability.sh" false
    run_check "Test coverage" "check_coverage.sh" false
    run_check "Dependencies" "check_dependencies.sh" false
fi

echo ""
echo "=== PRE-RELEASE ==="
echo ""
run_check "Pre-release checklist" "pre_release_check.sh" true

echo ""
echo "=========================================="
echo "  Summary:"
echo "    Blocking failures: $FAILED"
echo "    Warnings:          $WARNINGS"
echo "=========================================="

if [ $FAILED -gt 0 ]; then
    echo ""
    echo "  ✗ Fix blocking issues before release."
    exit 1
fi

if [ $WARNINGS -gt 0 ]; then
    echo ""
    echo "  ! Review warnings before release."
fi

echo ""
echo "  ✓ Ready for release."
exit 0
