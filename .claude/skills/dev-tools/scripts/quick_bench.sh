#!/bin/bash
# Quick benchmark with filtered output (removes AILANG debug noise)
# Usage: quick_bench.sh [iterations]

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"

cd "$PROJECT_ROOT"

ITERATIONS="${1:-1000}"

echo "Running quick benchmark (${ITERATIONS} iterations)..."
echo ""

# Ensure CLI is built
if [ ! -f bin/voyage ]; then
    echo "Building CLI..."
    make cli 2>/dev/null
fi

# Run benchmarks with filtered output
./bin/voyage bench -n "$ITERATIONS" -warmup 50 2>&1 | grep -v "^tick"

echo ""
echo "Done. Use 'voyage bench -profile' for CPU profiling."
