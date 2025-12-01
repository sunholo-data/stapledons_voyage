#!/usr/bin/env bash
# Check that layer boundaries are respected
# - engine/ must not contain game logic
# - sim_gen/ must not import ebiten directly
# Usage: .claude/skills/game-architect/scripts/check_layer_boundaries.sh

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../../.." && pwd)"

cd "$PROJECT_ROOT"

VIOLATIONS=0

echo "Checking layer boundaries..."
echo ""

# Check 1: engine/ should not manipulate World directly (beyond storing)
echo "1. Checking engine/ for game logic..."

# Patterns that suggest game logic in engine
GAME_LOGIC_PATTERNS=(
    "World\{"           # Creating World structs
    "\.NPCs"            # Accessing NPC lists
    "\.Buildings"       # Accessing building lists
    "npc\.Pattern"      # NPC behavior logic
    "entity\.Move"      # Entity movement logic
    "BuildAction"       # Build actions
)

for pattern in "${GAME_LOGIC_PATTERNS[@]}"; do
    matches=$(grep -rn "$pattern" engine/ 2>/dev/null | grep -v "_test.go" | grep -v "// allowed:" || true)
    if [ -n "$matches" ]; then
        echo "  ✗ Found game logic pattern '$pattern' in engine/:"
        echo "$matches" | head -5 | sed 's/^/      /'
        VIOLATIONS=$((VIOLATIONS + 1))
    fi
done

# Check 2: sim_gen/ should not import ebiten
echo ""
echo "2. Checking sim_gen/ for rendering imports..."

RENDER_IMPORTS=(
    "github.com/hajimehoshi/ebiten"
    "image/color"
    "image/png"
)

for import in "${RENDER_IMPORTS[@]}"; do
    matches=$(grep -rn "\"$import" sim_gen/ 2>/dev/null || true)
    if [ -n "$matches" ]; then
        echo "  ✗ Found rendering import '$import' in sim_gen/:"
        echo "$matches" | head -3 | sed 's/^/      /'
        VIOLATIONS=$((VIOLATIONS + 1))
    fi
done

# Check 3: cmd/ should only wire things together
echo ""
echo "3. Checking cmd/ for inline logic..."

# cmd/ files should be small - wiring only
for file in cmd/*/*.go; do
    if [ -f "$file" ]; then
        lines=$(wc -l < "$file" | tr -d ' ')
        if [ "$lines" -gt 200 ]; then
            echo "  ! WARN: $file is $lines lines (cmd/ should be minimal wiring)"
        fi
    fi
done

# Check 4: Ensure sim_gen exports correct API
echo ""
echo "4. Checking sim_gen/ exports..."

required_exports=("InitWorld" "Step" "World" "FrameInput" "FrameOutput" "DrawCmd")
for export in "${required_exports[@]}"; do
    if ! grep -q "type $export\|func $export" sim_gen/*.go 2>/dev/null; then
        if ! grep -q "^$export" sim_gen/*.go 2>/dev/null; then
            echo "  ! WARN: '$export' may not be exported from sim_gen/"
        fi
    fi
done

echo ""
echo "----------------------------------------"
echo "Layer Boundary Summary:"
echo "  Violations: $VIOLATIONS"

if [ $VIOLATIONS -gt 0 ]; then
    echo ""
    echo "Layer rules:"
    echo "  - engine/: IO only (input, render, assets)"
    echo "  - sim_gen/: Game logic (types, Step function)"
    echo "  - cmd/: Wiring only"
    exit 1
fi

echo "  ✓ All layer boundaries respected"
exit 0
