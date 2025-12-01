#!/usr/bin/env bash
# Check that AILANG source types match sim_gen Go types
# Usage: .claude/skills/game-architect/scripts/check_ailang_sync.sh

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../../.." && pwd)"

cd "$PROJECT_ROOT"

echo "Checking AILANG ↔ Go sync..."
echo ""

# Check if we have AILANG files
AIL_FILES=$(find sim -name "*.ail" -type f 2>/dev/null | wc -l | tr -d ' ')
if [ "$AIL_FILES" -eq 0 ]; then
    echo "  No AILANG files found in sim/"
    echo "  (This is OK if using mock sim_gen)"
    exit 0
fi

echo "Found $AIL_FILES AILANG source files"
echo ""

MISMATCHES=0

# Extract types from AILANG
echo "Types defined in AILANG:"
echo "----------------------------------------"

extract_ail_types() {
    grep -rh "^type " sim/*.ail 2>/dev/null | sed 's/type //' | sed 's/ =.*//' | sed 's/{.*//' | tr -d ' ' | sort -u || true
}

AIL_TYPES=$(extract_ail_types)
echo "$AIL_TYPES" | sed 's/^/  /'

echo ""
echo "Types defined in sim_gen:"
echo "----------------------------------------"

extract_go_types() {
    grep -rh "^type [A-Z]" sim_gen/*.go 2>/dev/null | sed 's/type //' | sed 's/ .*//' | sort -u || true
}

GO_TYPES=$(extract_go_types)
echo "$GO_TYPES" | sed 's/^/  /'

echo ""
echo "Sync check:"
echo "----------------------------------------"

# Check each AILANG type exists in Go
while IFS= read -r typ; do
    if [ -z "$typ" ]; then continue; fi

    if echo "$GO_TYPES" | grep -q "^${typ}$"; then
        echo "  ✓ $typ"
    else
        echo "  ✗ $typ - defined in AILANG but missing in sim_gen"
        MISMATCHES=$((MISMATCHES + 1))
    fi
done <<< "$AIL_TYPES"

echo ""
echo "ADT variant check:"
echo "----------------------------------------"

# Check ADT variants (e.g., DrawCmd variants)
# AILANG: type DrawCmd = Sprite(...) | Rect(...) | Text(...)
# Go: type DrawCmdKind int; const (DrawCmdKindSprite DrawCmdKind = iota; ...)

for ail_file in sim/*.ail; do
    if [ ! -f "$ail_file" ]; then continue; fi

    # Find ADT definitions (type X = A | B | C)
    grep -E "^type .* =" "$ail_file" 2>/dev/null | grep "|" | while read -r line; do
        typename=$(echo "$line" | sed 's/type //' | sed 's/ =.*//')
        variants=$(echo "$line" | sed 's/.*= //' | tr '|' '\n' | sed 's/(.*//' | tr -d ' ')

        echo "  $typename variants:"
        while IFS= read -r variant; do
            if [ -z "$variant" ]; then continue; fi

            # Check if Go has corresponding kind constant
            kind_name="${typename}Kind${variant}"
            if grep -q "$kind_name" sim_gen/*.go 2>/dev/null; then
                echo "    ✓ $variant → $kind_name"
            else
                echo "    ✗ $variant - no $kind_name in Go"
                MISMATCHES=$((MISMATCHES + 1))
            fi
        done <<< "$variants"
    done
done

echo ""
echo "Function check:"
echo "----------------------------------------"

# Check key functions
extract_ail_funcs() {
    grep -rh "^func \|^pure func " sim/*.ail 2>/dev/null | sed 's/pure //' | sed 's/func //' | sed 's/(.*//' | sort -u || true
}

AIL_FUNCS=$(extract_ail_funcs)
GO_FUNCS=$(grep -rh "^func [A-Z]" sim_gen/*.go 2>/dev/null | sed 's/func //' | sed 's/(.*//' | sort -u || true)

# Check critical functions
CRITICAL=("init_world:InitWorld" "step:Step")
for pair in "${CRITICAL[@]}"; do
    ail_name="${pair%%:*}"
    go_name="${pair##*:}"

    ail_exists=$(echo "$AIL_FUNCS" | grep -c "^${ail_name}$" || true)
    go_exists=$(echo "$GO_FUNCS" | grep -c "^${go_name}$" || true)

    if [ "$ail_exists" -gt 0 ] && [ "$go_exists" -gt 0 ]; then
        echo "  ✓ $ail_name → $go_name"
    elif [ "$ail_exists" -gt 0 ] && [ "$go_exists" -eq 0 ]; then
        echo "  ✗ $ail_name defined but $go_name missing"
        MISMATCHES=$((MISMATCHES + 1))
    elif [ "$ail_exists" -eq 0 ] && [ "$go_exists" -gt 0 ]; then
        echo "  ! $go_name exists but $ail_name not in AILANG (mock?)"
    fi
done

echo ""
echo "----------------------------------------"
if [ $MISMATCHES -gt 0 ]; then
    echo "✗ Found $MISMATCHES sync mismatches"
    echo ""
    echo "Options:"
    echo "  1. Run 'ailang compile --emit-go' to regenerate sim_gen/"
    echo "  2. Update sim/*.ail to match sim_gen/ (if mock)"
    exit 1
fi

echo "✓ AILANG and Go types are in sync"
exit 0
