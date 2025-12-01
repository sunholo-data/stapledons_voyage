#!/usr/bin/env bash
# Check that files don't exceed line limits
# - Error: > 800 lines
# - Warning: > 600 lines
# Usage: .claude/skills/game-architect/scripts/check_file_sizes.sh

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../../.." && pwd)"

cd "$PROJECT_ROOT"

ERROR_LIMIT=800
WARN_LIMIT=600

ERRORS=0
WARNINGS=0

echo "Checking file sizes (error: >$ERROR_LIMIT, warn: >$WARN_LIMIT)..."
echo ""

# Check Go files
echo "Go files:"
while IFS= read -r file; do
    # Skip generated/vendor directories
    if [[ "$file" == *"/vendor/"* ]] || [[ "$file" == *"_test.go" ]]; then
        continue
    fi

    lines=$(wc -l < "$file" | tr -d ' ')

    if [ "$lines" -gt "$ERROR_LIMIT" ]; then
        echo "  ✗ ERROR: $file ($lines lines)"
        ERRORS=$((ERRORS + 1))
    elif [ "$lines" -gt "$WARN_LIMIT" ]; then
        echo "  ! WARN:  $file ($lines lines)"
        WARNINGS=$((WARNINGS + 1))
    fi
done < <(find . -name "*.go" -type f | grep -v "/vendor/" | sort)

echo ""

# Check AILANG files
echo "AILANG files:"
while IFS= read -r file; do
    lines=$(wc -l < "$file" | tr -d ' ')

    if [ "$lines" -gt "$ERROR_LIMIT" ]; then
        echo "  ✗ ERROR: $file ($lines lines)"
        ERRORS=$((ERRORS + 1))
    elif [ "$lines" -gt "$WARN_LIMIT" ]; then
        echo "  ! WARN:  $file ($lines lines)"
        WARNINGS=$((WARNINGS + 1))
    fi
done < <(find . -name "*.ail" -type f 2>/dev/null | sort || true)

echo ""

# Summary
echo "----------------------------------------"
echo "File Size Summary:"
echo "  Errors (>$ERROR_LIMIT lines):   $ERRORS"
echo "  Warnings (>$WARN_LIMIT lines): $WARNINGS"

# Show largest files
echo ""
echo "Largest Go files:"
find . -name "*.go" -type f | grep -v "/vendor/" | while read -r f; do
    wc -l "$f"
done | sort -rn | head -5 | while read -r count file; do
    echo "  $count $file"
done

if [ $ERRORS -gt 0 ]; then
    exit 1
fi

exit 0
