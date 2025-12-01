#!/usr/bin/env bash
# Check cyclomatic complexity of functions
# Usage: .claude/skills/game-architect/scripts/check_complexity.sh

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../../.." && pwd)"

cd "$PROJECT_ROOT"

echo "Checking code complexity..."
echo ""

# Thresholds
MAX_FUNC_LINES=100
MAX_SWITCH_CASES=15
MAX_NESTING=4

ISSUES=0
WARNINGS=0

echo "1. Large functions (>${MAX_FUNC_LINES} lines):"
echo "----------------------------------------"

# Find large functions
# This is a heuristic - counts lines between "func" and next "func" or EOF
find_large_funcs() {
    for file in $(find . -name "*.go" -type f | grep -v "/vendor/" | grep -v "_test.go"); do
        awk -v max="$MAX_FUNC_LINES" -v file="$file" '
        /^func / {
            if (func_name != "" && line_count > max) {
                printf "  ✗ %s:%s %s() - %d lines\n", file, func_start, func_name, line_count
            }
            func_name = $2
            gsub(/\(.*/, "", func_name)
            func_start = NR
            line_count = 0
        }
        { line_count++ }
        END {
            if (func_name != "" && line_count > max) {
                printf "  ✗ %s:%s %s() - %d lines\n", file, func_start, func_name, line_count
            }
        }
        ' "$file"
    done
}

LARGE_FUNCS=$(find_large_funcs)
if [ -n "$LARGE_FUNCS" ]; then
    echo "$LARGE_FUNCS"
    ISSUES=$((ISSUES + $(echo "$LARGE_FUNCS" | wc -l)))
else
    echo "  ✓ No functions over $MAX_FUNC_LINES lines"
fi

echo ""
echo "2. Complex switch statements (>${MAX_SWITCH_CASES} cases):"
echo "----------------------------------------"

# Find switches with many cases
find_complex_switches() {
    for file in $(find . -name "*.go" -type f | grep -v "/vendor/"); do
        awk -v max="$MAX_SWITCH_CASES" -v file="$file" '
        /switch/ { in_switch = 1; switch_line = NR; case_count = 0 }
        in_switch && /case / { case_count++ }
        in_switch && /^[[:space:]]*}[[:space:]]*$/ {
            if (case_count > max) {
                printf "  ! %s:%d - switch with %d cases\n", file, switch_line, case_count
            }
            in_switch = 0
        }
        ' "$file"
    done
}

COMPLEX_SWITCHES=$(find_complex_switches)
if [ -n "$COMPLEX_SWITCHES" ]; then
    echo "$COMPLEX_SWITCHES"
    WARNINGS=$((WARNINGS + $(echo "$COMPLEX_SWITCHES" | wc -l)))
else
    echo "  ✓ No overly complex switches"
fi

echo ""
echo "3. Deep nesting (>${MAX_NESTING} levels):"
echo "----------------------------------------"

# Find deeply nested code (heuristic based on indentation)
find_deep_nesting() {
    for file in $(find . -name "*.go" -type f | grep -v "/vendor/" | grep -v "_test.go"); do
        awk -v max="$MAX_NESTING" -v file="$file" '
        {
            # Count leading tabs
            match($0, /^[\t]+/)
            indent = RLENGTH
            if (indent < 0) indent = 0

            # Each tab is roughly one nesting level in Go
            if (indent > max && !/^[\t]*\/\// && !/^[\t]*\*/) {
                if (!reported[file]) {
                    printf "  ! %s:%d - nesting level %d\n", file, NR, indent
                    reported[file] = 1
                }
            }
        }
        ' "$file"
    done
}

DEEP_NESTING=$(find_deep_nesting)
if [ -n "$DEEP_NESTING" ]; then
    echo "$DEEP_NESTING" | head -10
    WARNINGS=$((WARNINGS + $(echo "$DEEP_NESTING" | wc -l)))
else
    echo "  ✓ No deeply nested code"
fi

echo ""
echo "4. Function count per file:"
echo "----------------------------------------"

# Count functions per file
FUNC_WARNINGS=0
while IFS= read -r file; do
    count=$(grep -c "^func " "$file" 2>/dev/null || echo "0")
    count=$((count + 0))  # Ensure it's a number
    if [ "$count" -gt 20 ]; then
        echo "  ! $file - $count functions (consider splitting)"
        FUNC_WARNINGS=$((FUNC_WARNINGS + 1))
    fi
done < <(find . -name "*.go" -type f | grep -v "/vendor/" | grep -v "_test.go" | head -100)

WARNINGS=$((WARNINGS + FUNC_WARNINGS))
if [ $FUNC_WARNINGS -eq 0 ]; then
    echo "  ✓ All files have reasonable function counts"
fi

echo ""
echo "5. Parameter count:"
echo "----------------------------------------"

# Find functions with many parameters
find_many_params() {
    for file in $(find . -name "*.go" -type f | grep -v "/vendor/"); do
        grep -n "^func " "$file" | while read -r line; do
            linenum=$(echo "$line" | cut -d: -f1)
            # Count commas in parameter list (rough estimate)
            params=$(echo "$line" | sed 's/.*(//' | sed 's/).*//' | tr -cd ',' | wc -c)
            params=$((params + 1))
            if [ "$params" -gt 5 ]; then
                funcname=$(echo "$line" | sed 's/.*func //' | sed 's/(.*//')
                echo "  ! $file:$linenum $funcname() - $params parameters"
            fi
        done
    done
}

MANY_PARAMS=$(find_many_params)
if [ -n "$MANY_PARAMS" ]; then
    echo "$MANY_PARAMS" | head -10
else
    echo "  ✓ All functions have reasonable parameter counts"
fi

echo ""
echo "----------------------------------------"
echo "Complexity Summary:"
echo "  Errors (must fix):   $ISSUES"
echo "  Warnings (review):   $WARNINGS"

if [ $ISSUES -gt 0 ]; then
    exit 1
fi

echo ""
echo "✓ Complexity within acceptable limits"
exit 0
