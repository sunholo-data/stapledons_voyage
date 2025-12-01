#!/usr/bin/env bash
# Analyze and visualize package dependencies
# Usage: .claude/skills/game-architect/scripts/check_dependencies.sh

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../../.." && pwd)"

cd "$PROJECT_ROOT"

echo "Analyzing package dependencies..."
echo ""

# Get module name
MODULE=$(go list -m 2>/dev/null || echo "stapledons_voyage")

# List all packages
echo "Package structure:"
echo "----------------------------------------"
go list ./... 2>/dev/null | while read -r pkg; do
    # Get relative path
    rel=${pkg#$MODULE/}
    if [ "$rel" = "$pkg" ]; then
        rel="(root)"
    fi
    echo "  $rel"
done

echo ""
echo "Dependency graph (internal packages only):"
echo "----------------------------------------"

# For each package, show what it imports
for pkg in $(go list ./... 2>/dev/null); do
    rel=${pkg#$MODULE/}
    if [ "$rel" = "$pkg" ]; then
        rel="main"
    fi

    # Get imports that are internal to this module
    imports=$(go list -f '{{range .Imports}}{{.}} {{end}}' "$pkg" 2>/dev/null | tr ' ' '\n' | grep "^$MODULE" | sed "s|^$MODULE/||" || true)

    if [ -n "$imports" ]; then
        echo ""
        echo "  $rel imports:"
        echo "$imports" | while read -r imp; do
            if [ -n "$imp" ]; then
                echo "    → $imp"
            fi
        done
    fi
done

echo ""
echo "Layer dependency check:"
echo "----------------------------------------"

VIOLATIONS=0

# Check: engine should not import sim (only sim_gen)
engine_imports=$(go list -f '{{range .Imports}}{{.}} {{end}}' ./engine/... 2>/dev/null | tr ' ' '\n' | grep "^$MODULE/sim$" || true)
if [ -n "$engine_imports" ]; then
    echo "  ✗ engine/ imports sim/ directly (should only import sim_gen/)"
    VIOLATIONS=$((VIOLATIONS + 1))
else
    echo "  ✓ engine/ correctly imports sim_gen/ (not sim/)"
fi

# Check: sim_gen should not import engine
simgen_imports=$(go list -f '{{range .Imports}}{{.}} {{end}}' ./sim_gen/... 2>/dev/null | tr ' ' '\n' | grep "^$MODULE/engine" || true)
if [ -n "$simgen_imports" ]; then
    echo "  ✗ sim_gen/ imports engine/ (forbidden)"
    VIOLATIONS=$((VIOLATIONS + 1))
else
    echo "  ✓ sim_gen/ does not import engine/"
fi

# Check: cmd should only import engine and sim_gen
cmd_imports=$(go list -f '{{range .Imports}}{{.}} {{end}}' ./cmd/... 2>/dev/null | tr ' ' '\n' | grep "^$MODULE" | grep -v "engine\|sim_gen" || true)
if [ -n "$cmd_imports" ]; then
    echo "  ! cmd/ imports packages other than engine/ and sim_gen/:"
    echo "$cmd_imports" | sed 's/^/      /'
else
    echo "  ✓ cmd/ only imports engine/ and sim_gen/"
fi

echo ""
echo "External dependencies:"
echo "----------------------------------------"

# List external (non-std, non-module) dependencies
go list -f '{{range .Imports}}{{.}} {{end}}' ./... 2>/dev/null | tr ' ' '\n' | sort -u | grep -v "^$MODULE" | grep "\." | head -20 | while read -r dep; do
    echo "  $dep"
done

echo ""
if [ $VIOLATIONS -gt 0 ]; then
    echo "✗ Found $VIOLATIONS layer dependency violations"
    exit 1
fi

echo "✓ Dependencies OK"
exit 0
