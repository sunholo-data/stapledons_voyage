#!/usr/bin/env bash
# Detect import cycles between packages
# Usage: .claude/skills/game-architect/scripts/check_import_cycles.sh

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../../.." && pwd)"

cd "$PROJECT_ROOT"

echo "Checking for import cycles..."
echo ""

# Get module name
MODULE=$(go list -m 2>/dev/null || echo "stapledons_voyage")

# Go compiler itself catches import cycles, so let's verify
echo "1. Compiler cycle check:"
echo "----------------------------------------"

if go build ./... 2>&1 | grep -q "import cycle"; then
    echo "  ✗ Import cycle detected by compiler:"
    go build ./... 2>&1 | grep "import cycle" | sed 's/^/      /'
    exit 1
else
    echo "  ✓ No import cycles (build passes)"
fi

echo ""
echo "2. Layer dependency analysis:"
echo "----------------------------------------"

# Build dependency graph
declare -A DEPS

for pkg in $(go list ./... 2>/dev/null); do
    rel=${pkg#$MODULE/}
    if [ "$rel" = "$pkg" ]; then continue; fi

    imports=$(go list -f '{{range .Imports}}{{.}} {{end}}' "$pkg" 2>/dev/null | tr ' ' '\n' | grep "^$MODULE" | sed "s|^$MODULE/||" || true)
    DEPS[$rel]="$imports"
done

# Check for potential architectural cycles
echo "Checking layer dependencies..."

# Allowed: cmd → engine → sim_gen
# Allowed: cmd → sim_gen
# Not allowed: sim_gen → engine, engine → cmd, sim_gen → cmd

VIOLATIONS=0

# Check sim_gen dependencies
simgen_deps=$(go list -f '{{range .Imports}}{{.}} {{end}}' ./sim_gen/... 2>/dev/null | tr ' ' '\n' | grep "^$MODULE" | sed "s|^$MODULE/||" || true)
for dep in $simgen_deps; do
    if [[ "$dep" == engine* ]] || [[ "$dep" == cmd* ]]; then
        echo "  ✗ sim_gen/ imports $dep (layer violation)"
        VIOLATIONS=$((VIOLATIONS + 1))
    fi
done

# Check engine dependencies
engine_deps=$(go list -f '{{range .Imports}}{{.}} {{end}}' ./engine/... 2>/dev/null | tr ' ' '\n' | grep "^$MODULE" | sed "s|^$MODULE/||" || true)
for dep in $engine_deps; do
    if [[ "$dep" == cmd* ]]; then
        echo "  ✗ engine/ imports $dep (layer violation)"
        VIOLATIONS=$((VIOLATIONS + 1))
    fi
done

if [ $VIOLATIONS -eq 0 ]; then
    echo "  ✓ No layer violations"
fi

echo ""
echo "3. Dependency depth:"
echo "----------------------------------------"

# Calculate max dependency depth
calculate_depth() {
    local pkg=$1
    local visited=$2

    if [[ "$visited" == *":$pkg:"* ]]; then
        echo 0
        return
    fi

    local max_depth=0
    local imports=$(go list -f '{{range .Imports}}{{.}} {{end}}' "./$pkg" 2>/dev/null | tr ' ' '\n' | grep "^$MODULE" | sed "s|^$MODULE/||" || true)

    for imp in $imports; do
        if [ -n "$imp" ]; then
            local depth=$(calculate_depth "$imp" "$visited:$pkg:")
            if [ "$depth" -gt "$max_depth" ]; then
                max_depth=$depth
            fi
        fi
    done

    echo $((max_depth + 1))
}

# Show dependency depths for main packages
for pkg in cmd/game engine/render sim_gen; do
    if [ -d "$pkg" ]; then
        depth=$(go list -f '{{range .Imports}}{{.}} {{end}}' "./$pkg" 2>/dev/null | tr ' ' '\n' | grep "^$MODULE" | wc -l | tr -d ' ')
        echo "  $pkg: $depth internal imports"
    fi
done

echo ""
echo "4. Circular reference check (within packages):"
echo "----------------------------------------"

# Check for files that import each other within same package (unusual but possible via build tags)
for dir in sim_gen engine/render engine/assets cmd/game; do
    if [ -d "$dir" ]; then
        files=$(find "$dir" -maxdepth 1 -name "*.go" -type f 2>/dev/null | wc -l | tr -d ' ')
        echo "  $dir/: $files files"
    fi
done

echo ""
echo "----------------------------------------"
if [ $VIOLATIONS -gt 0 ]; then
    echo "✗ Found $VIOLATIONS layer violations"
    exit 1
fi

echo "✓ No import cycles or layer violations"
exit 0
