#!/usr/bin/env bash
set -eo pipefail

# generate_feature_summary.sh - Generate markdown summary of feature status
#
# Reads design docs and generates a summary table for README or CHANGELOG
#
# Usage: ./scripts/generate_feature_summary.sh [--format table|list]

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
DESIGN_DOCS_DIR="$PROJECT_ROOT/design_docs"

FORMAT="${1:-table}"

# Extract version and title from a doc
extract_info() {
    local file="$1"
    local title=""
    local version=""
    local priority=""

    # Get title from first H1
    title=$(grep -m1 "^# " "$file" 2>/dev/null | sed 's/^# //' || echo "Untitled")

    # Get version
    version=$(grep -i "Version\|Target" "$file" 2>/dev/null | grep -E "^\*\*Version|^\*\*Target|^Version|^Target" | head -1 | sed 's/.*[:\*][[:space:]]*//' | sed 's/\*\*$//' | tr -d '\r')

    # Get priority
    priority=$(grep -i "Priority" "$file" 2>/dev/null | grep -E "^\*\*Priority|^Priority" | head -1 | sed 's/.*[:\*][[:space:]]*//' | sed 's/\*\*$//' | tr -d '\r')

    echo "$title|$version|$priority"
}

# Generate table format
generate_table() {
    echo "## Feature Status Summary"
    echo ""
    echo "### Implemented Features"
    echo ""
    echo "| Feature | Version | Design Doc |"
    echo "|---------|---------|------------|"

    # Find implemented docs
    find "$DESIGN_DOCS_DIR/implemented" -name "*.md" -type f | while read -r file; do
        [[ "$(basename "$file")" == "README.md" ]] && continue
        [[ "$(basename "$file")" == ".gitkeep" ]] && continue

        info=$(extract_info "$file")
        IFS='|' read -r title version priority <<< "$info"
        rel_path="${file#$PROJECT_ROOT/}"

        printf "| %s | %s | [doc](%s) |\n" "$title" "${version:-?}" "$rel_path"
    done | sort

    echo ""
    echo "### Planned Features (by version)"
    echo ""

    # Group planned docs by version
    local current_version=""
    find "$DESIGN_DOCS_DIR/planned" -name "*.md" -type f | while read -r file; do
        [[ "$(basename "$file")" == "README.md" ]] && continue
        [[ "$(basename "$file")" == ".gitkeep" ]] && continue

        info=$(extract_info "$file")
        IFS='|' read -r title version priority <<< "$info"

        echo "$version|$title|$file"
    done | sort | while IFS='|' read -r version title file; do
        rel_path="${file#$PROJECT_ROOT/}"
        printf "| %s | %s | [doc](%s) |\n" "${version:-unversioned}" "$title" "$rel_path"
    done
}

# Generate list format
generate_list() {
    echo "## Feature Status"
    echo ""
    echo "### Implemented"
    echo ""

    find "$DESIGN_DOCS_DIR/implemented" -name "*.md" -type f | while read -r file; do
        [[ "$(basename "$file")" == "README.md" ]] && continue

        info=$(extract_info "$file")
        IFS='|' read -r title version priority <<< "$info"
        rel_path="${file#$PROJECT_ROOT/}"

        echo "- **$title** (v${version:-?}) - [$rel_path]($rel_path)"
    done | sort

    echo ""
    echo "### Planned"
    echo ""

    find "$DESIGN_DOCS_DIR/planned" -name "*.md" -type f | while read -r file; do
        [[ "$(basename "$file")" == "README.md" ]] && continue

        info=$(extract_info "$file")
        IFS='|' read -r title version priority <<< "$info"
        rel_path="${file#$PROJECT_ROOT/}"

        echo "- **$title** (${version:-?}) - [$rel_path]($rel_path)"
    done | sort
}

# Generate stats
generate_stats() {
    local impl_count=$(find "$DESIGN_DOCS_DIR/implemented" -name "*.md" -type f | grep -v README | wc -l | tr -d ' ')
    local plan_count=$(find "$DESIGN_DOCS_DIR/planned" -name "*.md" -type f | grep -v README | wc -l | tr -d ' ')

    echo ""
    echo "---"
    echo ""
    echo "*$impl_count features implemented, $plan_count planned*"
    echo ""
    echo "See [CHANGELOG.md](../CHANGELOG.md) for release history."
}

# Main
case "$FORMAT" in
    table)
        generate_table
        generate_stats
        ;;
    list)
        generate_list
        generate_stats
        ;;
    *)
        echo "Usage: $0 [table|list]"
        exit 1
        ;;
esac
