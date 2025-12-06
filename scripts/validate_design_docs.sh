#!/usr/bin/env bash
set -eo pipefail

# validate_design_docs.sh - Validate and report on design doc organization
#
# This script:
# 1. Scans all design docs for required metadata
# 2. Reports misplaced docs (status doesn't match location)
# 3. Lists orphan docs (not in a version folder)
# 4. Generates a JSON summary for other tools
#
# Usage: ./scripts/validate_design_docs.sh [--fix] [--json]
#   --fix    Offer to move misplaced docs
#   --json   Output JSON summary only (for other tools)

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
DESIGN_DOCS_DIR="$PROJECT_ROOT/design_docs"

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

# Parse arguments
FIX_MODE=false
JSON_MODE=false
for arg in "$@"; do
    case $arg in
        --fix) FIX_MODE=true ;;
        --json) JSON_MODE=true ;;
    esac
done

# Arrays to track docs by category
declare -a IMPLEMENTED_DOCS=()
declare -a PLANNED_DOCS=()
declare -a REFERENCE_DOCS=()
declare -a INPUT_DOCS=()
declare -a ORPHAN_DOCS=()
declare -a MISPLACED_DOCS=()
declare -a MISSING_METADATA=()

# Extract metadata from a design doc
extract_metadata() {
    local file="$1"
    local status=""
    local version=""
    local priority=""
    local target=""

    # Extract status - handle various markdown formats:
    # **Status:** Value, **Status**: Value, Status: Value
    status=$(grep -i "Status" "$file" 2>/dev/null | grep -E "^\*\*Status|^Status" | head -1 | sed 's/.*Status[:\*]*[[:space:]]*//' | sed 's/\*\*$//' | tr -d '\r')

    # Extract version (from **Version:** or **Target:** or folder path)
    version=$(grep -i "Version" "$file" 2>/dev/null | grep -E "^\*\*Version|^Version" | head -1 | sed 's/.*Version[:\*]*[[:space:]]*//' | sed 's/\*\*$//' | tr -d '\r')
    if [ -z "$version" ]; then
        target=$(grep -i "Target" "$file" 2>/dev/null | grep -E "^\*\*Target|^Target" | head -1 | sed 's/.*Target[:\*]*[[:space:]]*//' | sed 's/\*\*$//' | tr -d '\r')
        version="$target"
    fi

    # If no explicit version, try to extract from path
    if [ -z "$version" ]; then
        if [[ "$file" =~ v[0-9]+_[0-9]+_[0-9]+ ]]; then
            version=$(echo "$file" | grep -oE 'v[0-9]+_[0-9]+_[0-9]+' | head -1 | tr '_' '.')
        fi
    fi

    # Extract priority
    priority=$(grep -i "Priority" "$file" 2>/dev/null | grep -E "^\*\*Priority|^Priority" | head -1 | sed 's/.*Priority[:\*]*[[:space:]]*//' | sed 's/\*\*$//' | tr -d '\r')

    echo "$status|$version|$priority"
}

# Check if doc is in a version folder
get_doc_location() {
    local file="$1"
    local rel_path="${file#$DESIGN_DOCS_DIR/}"

    if [[ "$rel_path" == implemented/* ]]; then
        echo "implemented"
    elif [[ "$rel_path" == planned/* ]]; then
        echo "planned"
    elif [[ "$rel_path" == reference/* ]]; then
        echo "reference"
    elif [[ "$rel_path" == input/* ]]; then
        echo "input"
    else
        echo "unknown"
    fi
}

# Check if doc is in a valid subfolder (version, next, or future)
is_in_valid_folder() {
    local file="$1"
    # Accept version folders (v0_1_0) or next/ or future/
    if [[ "$file" =~ v[0-9]+_[0-9]+_[0-9]+ ]] || [[ "$file" =~ /next/ ]] || [[ "$file" =~ /future/ ]]; then
        return 0
    else
        return 1
    fi
}

# Get version folder from path
get_version_from_path() {
    local file="$1"
    if [[ "$file" =~ (v[0-9]+_[0-9]+_[0-9]+) ]]; then
        echo "${BASH_REMATCH[1]}"
    else
        echo ""
    fi
}

# Scan all design docs
scan_docs() {
    local doc
    local metadata
    local status
    local version
    local priority
    local location
    local rel_path

    while IFS= read -r -d '' doc; do
        # Skip README.md and .gitkeep
        [[ "$(basename "$doc")" == "README.md" ]] && continue
        [[ "$(basename "$doc")" == ".gitkeep" ]] && continue

        rel_path="${doc#$DESIGN_DOCS_DIR/}"
        location=$(get_doc_location "$doc")
        metadata=$(extract_metadata "$doc")

        IFS='|' read -r status version priority <<< "$metadata"

        # Normalize status
        status_lower=$(echo "$status" | tr '[:upper:]' '[:lower:]')

        # Check for missing metadata
        if [ -z "$status" ]; then
            MISSING_METADATA+=("$rel_path (missing Status)")
        fi

        # Track docs by location/status
        if [ "$location" == "reference" ]; then
            # Reference docs don't need version folders or status
            REFERENCE_DOCS+=("$rel_path|$version|$priority")
        elif [ "$location" == "input" ]; then
            # Input docs don't need version folders or status
            INPUT_DOCS+=("$rel_path|$version|$priority")
        elif [[ "$status_lower" == *"implemented"* ]]; then
            IMPLEMENTED_DOCS+=("$rel_path|$version|$priority")

            # Check if misplaced (implemented doc in planned/)
            if [ "$location" == "planned" ]; then
                MISPLACED_DOCS+=("$rel_path should be in implemented/")
            fi
        elif [[ "$status_lower" == *"planned"* ]] || [ -z "$status" ]; then
            PLANNED_DOCS+=("$rel_path|$version|$priority")

            # Check if misplaced (planned doc in implemented/)
            if [ "$location" == "implemented" ] && [[ "$status_lower" == *"planned"* ]]; then
                MISPLACED_DOCS+=("$rel_path should be in planned/")
            fi
        fi

        # Check if orphan (not in valid folder) - only for planned/
        if [ "$location" == "planned" ] && ! is_in_valid_folder "$doc"; then
            # Only flag if it's directly in planned/ (not in next/, future/, or version folder)
            if [[ "$rel_path" == "planned/"*.md ]]; then
                ORPHAN_DOCS+=("$rel_path|$version|$priority")
            fi
        fi

    done < <(find "$DESIGN_DOCS_DIR" -name "*.md" -type f -print0)
}

# Print summary
print_summary() {
    local impl_count=${#IMPLEMENTED_DOCS[@]}
    local plan_count=${#PLANNED_DOCS[@]}
    local ref_count=${#REFERENCE_DOCS[@]}
    local input_count=${#INPUT_DOCS[@]}
    local orph_count=${#ORPHAN_DOCS[@]}
    local misp_count=${#MISPLACED_DOCS[@]}
    local miss_count=${#MISSING_METADATA[@]}

    echo ""
    echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}"
    echo -e "${BLUE}                    DESIGN DOC VALIDATION                       ${NC}"
    echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}"
    echo ""

    # Implemented docs
    echo -e "${GREEN}✓ IMPLEMENTED (${impl_count} docs)${NC}"
    echo "─────────────────────────────────────────────────────────────────"
    if [ $impl_count -gt 0 ]; then
        for doc in "${IMPLEMENTED_DOCS[@]}"; do
            IFS='|' read -r path version priority <<< "$doc"
            printf "  %-50s %s\n" "$path" "${version:-unversioned}"
        done
    else
        echo "  (none)"
    fi
    echo ""

    # Planned docs
    echo -e "${YELLOW}○ PLANNED (${plan_count} docs)${NC}"
    echo "─────────────────────────────────────────────────────────────────"
    if [ $plan_count -gt 0 ]; then
        for doc in "${PLANNED_DOCS[@]}"; do
            IFS='|' read -r path version priority <<< "$doc"
            printf "  %-50s %s\n" "$path" "${version:-unversioned}"
        done
    else
        echo "  (none)"
    fi
    echo ""

    # Reference docs
    echo -e "${BLUE}📚 REFERENCE (${ref_count} docs)${NC}"
    echo "─────────────────────────────────────────────────────────────────"
    if [ $ref_count -gt 0 ]; then
        for doc in "${REFERENCE_DOCS[@]}"; do
            IFS='|' read -r path version priority <<< "$doc"
            printf "  %-50s\n" "$path"
        done
    else
        echo "  (none)"
    fi
    echo ""

    # Input docs
    echo -e "${BLUE}📥 INPUT (${input_count} docs)${NC}"
    echo "─────────────────────────────────────────────────────────────────"
    if [ $input_count -gt 0 ]; then
        for doc in "${INPUT_DOCS[@]}"; do
            IFS='|' read -r path version priority <<< "$doc"
            printf "  %-50s\n" "$path"
        done
    else
        echo "  (none)"
    fi
    echo ""

    # Issues
    if [ $orph_count -gt 0 ]; then
        echo -e "${YELLOW}⚠ ORPHAN DOCS (not in version folder) - ${orph_count} docs${NC}"
        echo "─────────────────────────────────────────────────────────────────"
        for doc in "${ORPHAN_DOCS[@]}"; do
            IFS='|' read -r path version priority <<< "$doc"
            echo "  $path"
            if [ -n "$version" ]; then
                echo "    → Target version: $version"
            fi
        done
        echo ""
    fi

    if [ $misp_count -gt 0 ]; then
        echo -e "${RED}✗ MISPLACED DOCS - ${misp_count} docs${NC}"
        echo "─────────────────────────────────────────────────────────────────"
        for issue in "${MISPLACED_DOCS[@]}"; do
            echo "  $issue"
        done
        echo ""
    fi

    if [ $miss_count -gt 0 ]; then
        echo -e "${YELLOW}⚠ MISSING METADATA - ${miss_count} docs${NC}"
        echo "─────────────────────────────────────────────────────────────────"
        for issue in "${MISSING_METADATA[@]}"; do
            echo "  $issue"
        done
        echo ""
    fi

    # Summary stats
    echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}"
    echo "Summary:"
    echo "  Implemented: ${impl_count}"
    echo "  Planned:     ${plan_count}"
    echo "  Reference:   ${ref_count}"
    echo "  Input:       ${input_count}"
    echo "  ---"
    echo "  Orphans:     ${orph_count}"
    echo "  Misplaced:   ${misp_count}"
    echo ""

    if [ $misp_count -gt 0 ] || [ $orph_count -gt 0 ]; then
        if [ "$FIX_MODE" = true ]; then
            echo -e "${YELLOW}Run with --fix to resolve issues${NC}"
        else
            echo -e "${YELLOW}Tip: Run with --fix to organize docs${NC}"
        fi
    else
        echo -e "${GREEN}✓ All docs properly organized${NC}"
    fi
}

# Output JSON summary
output_json() {
    echo "{"
    echo '  "implemented": ['
    local first=true
    for doc in "${IMPLEMENTED_DOCS[@]}"; do
        IFS='|' read -r path version priority <<< "$doc"
        if [ "$first" = true ]; then
            first=false
        else
            echo ","
        fi
        printf '    {"path": "%s", "version": "%s", "priority": "%s"}' "$path" "$version" "$priority"
    done
    echo ""
    echo "  ],"

    echo '  "planned": ['
    first=true
    for doc in "${PLANNED_DOCS[@]}"; do
        IFS='|' read -r path version priority <<< "$doc"
        if [ "$first" = true ]; then
            first=false
        else
            echo ","
        fi
        printf '    {"path": "%s", "version": "%s", "priority": "%s"}' "$path" "$version" "$priority"
    done
    echo ""
    echo "  ],"

    echo '  "orphans": ['
    first=true
    for doc in "${ORPHAN_DOCS[@]}"; do
        IFS='|' read -r path version priority <<< "$doc"
        if [ "$first" = true ]; then
            first=false
        else
            echo ","
        fi
        printf '    {"path": "%s", "version": "%s"}' "$path" "$version"
    done
    echo ""
    echo "  ],"

    echo "  \"stats\": {"
    echo "    \"implemented\": ${#IMPLEMENTED_DOCS[@]},"
    echo "    \"planned\": ${#PLANNED_DOCS[@]},"
    echo "    \"orphans\": ${#ORPHAN_DOCS[@]},"
    echo "    \"misplaced\": ${#MISPLACED_DOCS[@]}"
    echo "  }"
    echo "}"
}

# Main
scan_docs

if [ "$JSON_MODE" = true ]; then
    output_json
else
    print_summary
fi
