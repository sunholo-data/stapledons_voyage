#!/bin/bash
# Compare current test output against golden files
# Usage: compare_golden.sh [scenario-name]

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

TOTAL_FILES=0
MATCHING=0
DIFFERENT=0
MISSING_GOLDEN=0

echo "Comparing test output against golden files..."
echo ""

for scenario in "${SCENARIOS[@]}"; do
    echo "=== $scenario ==="

    TEST_SCENARIO="$TEST_DIR/$scenario"
    GOLDEN_SCENARIO="$GOLDEN_DIR/$scenario"

    if [ ! -d "$GOLDEN_SCENARIO" ]; then
        echo "  ! No golden files for $scenario"
        echo "    Create with: .claude/skills/test-manager/scripts/update_golden.sh $scenario"
        ((MISSING_GOLDEN++))
        continue
    fi

    for test_file in "$TEST_SCENARIO"/*.png; do
        [ -f "$test_file" ] || continue

        filename=$(basename "$test_file")
        golden_file="$GOLDEN_SCENARIO/$filename"

        ((TOTAL_FILES++))

        if [ ! -f "$golden_file" ]; then
            echo "  ! Missing golden: $filename"
            ((DIFFERENT++))
            continue
        fi

        # Compare files using cmp (byte comparison)
        if cmp -s "$test_file" "$golden_file"; then
            echo "  ✓ $filename"
            ((MATCHING++))
        else
            echo "  ✗ $filename (different)"
            ((DIFFERENT++))

            # Generate diff image if ImageMagick is available
            if command -v compare &> /dev/null; then
                DIFF_DIR="$TEST_SCENARIO/diff"
                mkdir -p "$DIFF_DIR"
                compare "$golden_file" "$test_file" "$DIFF_DIR/$filename" 2>/dev/null || true
                echo "    Diff saved: $DIFF_DIR/$filename"
            fi
        fi
    done
    echo ""
done

echo "=== Summary ==="
echo "Total files: $TOTAL_FILES"
echo "Matching:    $MATCHING"
echo "Different:   $DIFFERENT"
echo "Missing golden dirs: $MISSING_GOLDEN"

if [ $DIFFERENT -gt 0 ] || [ $MISSING_GOLDEN -gt 0 ]; then
    echo ""
    echo "Tests FAILED - differences found"
    exit 1
else
    echo ""
    echo "Tests PASSED - all files match"
    exit 0
fi
