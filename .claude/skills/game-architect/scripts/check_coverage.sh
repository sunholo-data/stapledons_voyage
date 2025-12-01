#!/usr/bin/env bash
# Check test coverage for critical paths
# Usage: .claude/skills/game-architect/scripts/check_coverage.sh

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../../.." && pwd)"

cd "$PROJECT_ROOT"

echo "Checking test coverage..."
echo ""

# Run tests with coverage
COVERAGE_FILE="/tmp/coverage_$$.out"
go test -coverprofile="$COVERAGE_FILE" ./... >/dev/null 2>&1 || true

if [ ! -f "$COVERAGE_FILE" ]; then
    echo "  ! Could not generate coverage report"
    exit 0
fi

echo "Package coverage:"
echo "----------------------------------------"

# Parse coverage by package
go tool cover -func="$COVERAGE_FILE" 2>/dev/null | grep -E "^[a-z].*total:" | while read -r line; do
    pkg=$(echo "$line" | awk '{print $1}' | sed 's|.*/||' | sed 's/:$//')
    pct=$(echo "$line" | awk '{print $NF}')

    # Color code by coverage level
    pct_num=${pct%\%}
    if (( $(echo "$pct_num < 30" | bc -l) )); then
        status="✗"
    elif (( $(echo "$pct_num < 60" | bc -l) )); then
        status="!"
    else
        status="✓"
    fi

    printf "  %s %-20s %s\n" "$status" "$pkg" "$pct"
done

echo ""
echo "Critical functions coverage:"
echo "----------------------------------------"

# Check coverage of critical functions
CRITICAL_FUNCS=("Step" "InitWorld" "RenderFrame" "CaptureInput")

for func in "${CRITICAL_FUNCS[@]}"; do
    coverage=$(go tool cover -func="$COVERAGE_FILE" 2>/dev/null | grep "$func" | awk '{print $NF}' | head -1 || echo "N/A")
    if [ "$coverage" = "N/A" ] || [ -z "$coverage" ]; then
        echo "  ? $func - not found or not covered"
    else
        pct_num=${coverage%\%}
        if (( $(echo "$pct_num < 50" | bc -l 2>/dev/null || echo 0) )); then
            echo "  ! $func - $coverage (needs more tests)"
        else
            echo "  ✓ $func - $coverage"
        fi
    fi
done

echo ""
echo "Uncovered critical code:"
echo "----------------------------------------"

# Find functions with 0% coverage in sim_gen
go tool cover -func="$COVERAGE_FILE" 2>/dev/null | grep "sim_gen" | grep "0.0%" | head -10 | while read -r line; do
    func=$(echo "$line" | awk '{print $2}')
    echo "  ! sim_gen.$func - 0%"
done

# Overall
echo ""
TOTAL=$(go tool cover -func="$COVERAGE_FILE" 2>/dev/null | grep "total:" | awk '{print $NF}')
echo "Overall coverage: $TOTAL"

# Cleanup
rm -f "$COVERAGE_FILE"

# Check threshold
if [ -n "$TOTAL" ]; then
    pct_num=${TOTAL%\%}
    if (( $(echo "$pct_num < 20" | bc -l 2>/dev/null || echo 0) )); then
        echo ""
        echo "⚠ Coverage below 20% - consider adding tests"
    fi
fi

exit 0
