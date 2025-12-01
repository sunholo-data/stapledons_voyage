#!/bin/bash
# Run test scenarios with UI stripped (test mode) for golden file comparison
# Usage: run_tests.sh [scenario-name]

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../../.." && pwd)"
SCENARIOS_DIR="$PROJECT_ROOT/scenarios"
OUTPUT_DIR="$PROJECT_ROOT/out/test"

cd "$PROJECT_ROOT"

# Get list of scenarios
if [ -n "$1" ]; then
    SCENARIOS=("$1")
else
    SCENARIOS=($(ls -1 "$SCENARIOS_DIR"/*.json 2>/dev/null | xargs -n1 basename | sed 's/\.json$//'))
fi

if [ ${#SCENARIOS[@]} -eq 0 ]; then
    echo "No scenarios found in $SCENARIOS_DIR"
    exit 1
fi

echo "Running ${#SCENARIOS[@]} test scenario(s) with --test-mode..."
echo ""

PASSED=0
FAILED=0

for scenario in "${SCENARIOS[@]}"; do
    echo "=== Running: $scenario ==="

    OUTPUT="$OUTPUT_DIR/$scenario"
    mkdir -p "$OUTPUT"

    if go run ./cmd/game -scenario "$scenario" -test-mode 2>&1; then
        # Move captured files from out/scenarios to out/test
        if [ -d "out/scenarios/$scenario" ]; then
            mv out/scenarios/$scenario/* "$OUTPUT/" 2>/dev/null || true
            rmdir "out/scenarios/$scenario" 2>/dev/null || true
        fi
        echo "  ✓ Passed"
        ((PASSED++))
    else
        echo "  ✗ Failed"
        ((FAILED++))
    fi
    echo ""
done

echo "=== Summary ==="
echo "Passed: $PASSED"
echo "Failed: $FAILED"
echo "Output: $OUTPUT_DIR/"

if [ $FAILED -gt 0 ]; then
    exit 1
fi
