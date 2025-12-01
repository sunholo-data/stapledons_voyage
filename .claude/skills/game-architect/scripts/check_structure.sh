#!/usr/bin/env bash
# Check directory structure matches architecture
# Usage: .claude/skills/game-architect/scripts/check_structure.sh

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../../.." && pwd)"

cd "$PROJECT_ROOT"

ISSUES=0

echo "Checking directory structure..."
echo ""

# Check 1: Required directories exist
echo "1. Required directories..."
REQUIRED_DIRS=(
    "sim"
    "sim_gen"
    "engine"
    "engine/render"
    "engine/assets"
    "cmd"
    "cmd/game"
    "design_docs"
)

for dir in "${REQUIRED_DIRS[@]}"; do
    if [ -d "$dir" ]; then
        echo "  ✓ $dir/"
    else
        echo "  ✗ $dir/ missing"
        ISSUES=$((ISSUES + 1))
    fi
done

# Check 2: No stray Go files in root
echo ""
echo "2. No stray files in root..."
stray_go=$(find . -maxdepth 1 -name "*.go" -type f 2>/dev/null || true)
if [ -n "$stray_go" ]; then
    echo "  ✗ Found Go files in root (should be in packages):"
    echo "$stray_go" | sed 's/^/      /'
    ISSUES=$((ISSUES + 1))
else
    echo "  ✓ No stray Go files"
fi

# Check 3: engine/ subdirectory organization
echo ""
echo "3. Engine layer organization..."

# render/ should contain input and drawing
if [ -d "engine/render" ]; then
    render_files=$(ls engine/render/*.go 2>/dev/null | wc -l | tr -d ' ')
    echo "  ✓ engine/render/ ($render_files files)"
else
    echo "  ✗ engine/render/ missing"
    ISSUES=$((ISSUES + 1))
fi

# assets/ should exist
if [ -d "engine/assets" ]; then
    echo "  ✓ engine/assets/"
else
    echo "  ✗ engine/assets/ missing"
    ISSUES=$((ISSUES + 1))
fi

# Check 4: sim/ has expected AILANG files
echo ""
echo "4. Simulation source (sim/)..."

EXPECTED_AIL=(
    "protocol.ail"
    "world.ail"
)

for ail in "${EXPECTED_AIL[@]}"; do
    if [ -f "sim/$ail" ]; then
        echo "  ✓ sim/$ail"
    else
        echo "  ! sim/$ail not found (expected for full AILANG integration)"
    fi
done

# Check 5: sim_gen/ has Go files
echo ""
echo "5. Generated simulation (sim_gen/)..."
sim_gen_files=$(find sim_gen -name "*.go" -type f 2>/dev/null | wc -l | tr -d ' ')
if [ "$sim_gen_files" -gt 0 ]; then
    echo "  ✓ $sim_gen_files Go files in sim_gen/"
else
    echo "  ✗ No Go files in sim_gen/"
    ISSUES=$((ISSUES + 1))
fi

# Check 6: Design docs structure
echo ""
echo "6. Design documentation..."
if [ -d "design_docs/planned" ]; then
    planned=$(find design_docs/planned -name "*.md" -type f 2>/dev/null | wc -l | tr -d ' ')
    echo "  ✓ design_docs/planned/ ($planned docs)"
else
    echo "  ! design_docs/planned/ missing"
fi

if [ -d "design_docs/implemented" ]; then
    implemented=$(find design_docs/implemented -name "*.md" -type f 2>/dev/null | wc -l | tr -d ' ')
    echo "  ✓ design_docs/implemented/ ($implemented docs)"
else
    echo "  ! design_docs/implemented/ missing"
fi

echo ""
echo "----------------------------------------"
echo "Structure Summary:"
echo "  Issues: $ISSUES"

if [ $ISSUES -gt 0 ]; then
    exit 1
fi

echo "  ✓ Directory structure valid"
exit 0
