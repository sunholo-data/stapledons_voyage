#!/usr/bin/env bash
# Check API stability of sim_gen exports
# Maintains a baseline of expected exports and warns on changes
# Usage: .claude/skills/game-architect/scripts/check_api_stability.sh [--update]

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../../.." && pwd)"
BASELINE_FILE="$SCRIPT_DIR/../resources/api_baseline.txt"

cd "$PROJECT_ROOT"

UPDATE_MODE=false
if [ "${1:-}" = "--update" ]; then
    UPDATE_MODE=true
fi

echo "Checking sim_gen API stability..."
echo ""

# Extract current exports from sim_gen
extract_exports() {
    # Types (exported = starts with uppercase)
    echo "# Types"
    grep -rh "^type [A-Z]" sim_gen/*.go 2>/dev/null | sed 's/ struct.*//' | sed 's/ =.*//' | sort -u || true

    echo ""
    echo "# Functions"
    grep -rh "^func [A-Z]" sim_gen/*.go 2>/dev/null | sed 's/(.*$//' | sort -u || true

    echo ""
    echo "# Constants"
    grep -rh "^const [A-Z]\|^[[:space:]]*[A-Z][a-zA-Z]* =" sim_gen/*.go 2>/dev/null | grep -v "func\|type" | head -30 || true
}

CURRENT_EXPORTS=$(extract_exports)

if [ "$UPDATE_MODE" = true ]; then
    echo "$CURRENT_EXPORTS" > "$BASELINE_FILE"
    echo "✓ Updated API baseline at $BASELINE_FILE"
    exit 0
fi

# Check if baseline exists
if [ ! -f "$BASELINE_FILE" ]; then
    echo "No API baseline found. Creating initial baseline..."
    echo "$CURRENT_EXPORTS" > "$BASELINE_FILE"
    echo "✓ Created API baseline at $BASELINE_FILE"
    echo ""
    echo "Current exports:"
    echo "----------------------------------------"
    echo "$CURRENT_EXPORTS"
    exit 0
fi

# Compare with baseline
echo "Comparing with baseline..."
echo ""

BASELINE=$(cat "$BASELINE_FILE")

# Find removed exports (in baseline but not current)
REMOVED=$(diff <(echo "$BASELINE" | grep "^type \|^func " | sort) <(echo "$CURRENT_EXPORTS" | grep "^type \|^func " | sort) 2>/dev/null | grep "^< " | sed 's/^< //' || true)

# Find added exports (in current but not baseline)
ADDED=$(diff <(echo "$BASELINE" | grep "^type \|^func " | sort) <(echo "$CURRENT_EXPORTS" | grep "^type \|^func " | sort) 2>/dev/null | grep "^> " | sed 's/^> //' || true)

ISSUES=0

if [ -n "$REMOVED" ]; then
    echo "⚠ REMOVED exports (breaking change!):"
    echo "$REMOVED" | sed 's/^/  - /'
    echo ""
    ISSUES=$((ISSUES + 1))
fi

if [ -n "$ADDED" ]; then
    echo "+ Added exports:"
    echo "$ADDED" | sed 's/^/  + /'
    echo ""
fi

# Check required exports
echo "Required exports check:"
REQUIRED_TYPES=("World" "FrameInput" "FrameOutput" "DrawCmd" "DrawCmdKind")
REQUIRED_FUNCS=("InitWorld" "Step")

for typ in "${REQUIRED_TYPES[@]}"; do
    if echo "$CURRENT_EXPORTS" | grep -q "^type $typ"; then
        echo "  ✓ type $typ"
    else
        echo "  ✗ type $typ MISSING"
        ISSUES=$((ISSUES + 1))
    fi
done

for fn in "${REQUIRED_FUNCS[@]}"; do
    if echo "$CURRENT_EXPORTS" | grep -q "^func $fn"; then
        echo "  ✓ func $fn"
    else
        echo "  ✗ func $fn MISSING"
        ISSUES=$((ISSUES + 1))
    fi
done

echo ""
echo "----------------------------------------"
if [ $ISSUES -gt 0 ]; then
    echo "✗ API stability issues found"
    echo ""
    echo "If changes are intentional, run with --update to update baseline:"
    echo "  $0 --update"
    exit 1
fi

echo "✓ API stable"
exit 0
