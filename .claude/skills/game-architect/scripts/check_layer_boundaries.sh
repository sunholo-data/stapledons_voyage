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

# Check 1b: engine/ should not contain game-specific concepts (genericization)
echo ""
echo "1b. Checking engine/ for game-specific concepts (genericization)..."

# These patterns indicate game-specific content that should be in game_views/
GAME_SPECIFIC_PATTERNS=(
    "DeckType"          # Ship deck types (game concept)
    "DeckCore"          # Specific deck name
    "DeckBridge"        # Specific deck name
    "DeckEngineering"   # Specific deck name
    "DomeViewState"     # Game-specific view state
    "ArrivalState"      # Game-specific state
    "GetArrivalPlanet"  # Game-specific function
)

WARNINGS=0
for pattern in "${GAME_SPECIFIC_PATTERNS[@]}"; do
    matches=$(grep -rn "sim_gen\.$pattern\|sim_gen\.New$pattern" engine/ 2>/dev/null | grep -v "_test.go" | grep -v "// allowed:" || true)
    if [ -n "$matches" ]; then
        echo "  ! WARN: Found game-specific type '$pattern' in engine/ (should be in game_views/):"
        echo "$matches" | head -3 | sed 's/^/      /'
        WARNINGS=$((WARNINGS + 1))
    fi
done

if [ $WARNINGS -gt 0 ]; then
    echo "  → $WARNINGS game-specific patterns found in engine/"
    echo "  → These should move to game_views/ for engine reusability"
    echo "  → See design_docs/planned/engine-genericization.md"
fi

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
    echo "  - sim/*.ail: Game logic (AILANG source)"
    echo "  - sim_gen/: Generated code (OK to have game types)"
    echo "  - game_views/: Game-specific rendering helpers"
    echo "  - engine/: Generic rendering (reusable for ANY game)"
    echo "  - cmd/: Wiring only"
    exit 1
fi

echo "  ✓ All layer boundaries respected"
exit 0
