#!/bin/bash
# Generate comprehensive development report
# Usage: full_report.sh [output_file]

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"

cd "$PROJECT_ROOT"

OUTPUT="${1:-out/dev_report.txt}"
mkdir -p "$(dirname "$OUTPUT")"

echo "Generating development report..."
echo ""

# Ensure CLI is built
if [ ! -f bin/voyage ]; then
    echo "Building CLI..."
    make cli 2>/dev/null
fi

{
    echo "=============================================="
    echo "Stapledon's Voyage - Development Report"
    echo "Generated: $(date)"
    echo "=============================================="
    echo ""

    echo "## Asset Validation"
    echo ""
    ./bin/voyage assets 2>&1 || true
    echo ""

    echo "## World State Summary"
    echo ""
    ./bin/voyage world -summary 2>&1 | grep -v "^tick" || true
    echo ""

    echo "## Performance Benchmarks (100 iterations)"
    echo ""
    ./bin/voyage bench -n 100 -warmup 10 2>&1 | grep -v "^tick" || true
    echo ""

    echo "## Simulation Stress Test (1000 steps)"
    echo ""
    ./bin/voyage sim -steps 1000 -check 500 2>&1 | grep -v "^tick" || true
    echo ""

    echo "## Git Status"
    echo ""
    git status --short 2>&1 || true
    echo ""

    echo "## Recent Commits"
    echo ""
    git log --oneline -5 2>&1 || true
    echo ""

    echo "=============================================="
    echo "Report complete"
    echo "=============================================="
} > "$OUTPUT"

echo "Report saved to: $OUTPUT"
echo ""
cat "$OUTPUT"
