#!/bin/bash
# Update golden files from current test output
# Usage: update_golden.sh [scenario-name]

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../../.." && pwd)"
GOLDEN_DIR="$PROJECT_ROOT/golden"
TEST_DIR="$PROJECT_ROOT/out/test"

cd "$PROJECT_ROOT"

# Check if test output exists
if [ ! -d "$TEST_DIR" ]; then
    echo "No test output found. Run tests first:"
    echo "  .claude/skills/test-manager/scripts/run_tests.sh"
    exit 1
fi

# Get list of scenarios
if [ -n "$1" ]; then
    SCENARIOS=("$1")
else
    SCENARIOS=($(ls -1 "$TEST_DIR" 2>/dev/null))
fi

if [ ${#SCENARIOS[@]} -eq 0 ]; then
    echo "No test output found in $TEST_DIR"
    exit 1
fi

echo "Updating golden files..."
echo ""

UPDATED=0

for scenario in "${SCENARIOS[@]}"; do
    echo "=== $scenario ==="

    TEST_SCENARIO="$TEST_DIR/$scenario"
    GOLDEN_SCENARIO="$GOLDEN_DIR/$scenario"

    if [ ! -d "$TEST_SCENARIO" ]; then
        echo "  ! No test output for $scenario"
        continue
    fi

    # Create golden directory
    mkdir -p "$GOLDEN_SCENARIO"

    # Copy all PNG files
    for test_file in "$TEST_SCENARIO"/*.png; do
        [ -f "$test_file" ] || continue

        filename=$(basename "$test_file")
        cp "$test_file" "$GOLDEN_SCENARIO/$filename"
        echo "  ✓ Updated: $filename"
        ((UPDATED++))
    done
    echo ""
done

echo "=== Summary ==="
echo "Updated $UPDATED golden file(s)"
echo ""
echo "Don't forget to commit the golden files:"
echo "  git add golden/"
echo "  git commit -m 'Update golden files for visual tests'"
