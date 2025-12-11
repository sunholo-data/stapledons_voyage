#!/bin/bash
# Audit design docs by checking for corresponding sprint files
# A design doc is considered implemented when:
#   1. It has a corresponding sprint file in sprints/
#   2. The sprint file references the design doc
#
# SOURCE OF TRUTH: JSON sprint files (sprints/*.json)
# Markdown checkboxes in sprint plans are for human readability only.
# Progress is tracked via:
#   - features[].passes = true/false
#   - status = "pending" | "in_progress" | "completed"
#
# This approach is maintainable - no hardcoded checks per doc

set -e
cd "$(git rev-parse --show-toplevel)"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo "=========================================="
echo "Design Doc Audit - Stapledon's Voyage"
echo "=========================================="
echo ""

# Track results
HAS_SPRINT=()
NO_SPRINT=()
ORPHAN_SPRINTS=()

# Find sprint file for a design doc (returns both .md plan and .json tracking file)
find_sprint_for_doc() {
    local doc_path="$1"
    local doc_name=$(basename "$doc_path")

    # Search sprints/ for any file that references this design doc
    local sprint_file=$(grep -rl "$doc_name" sprints/ 2>/dev/null | head -1)

    if [ -n "$sprint_file" ]; then
        echo "$sprint_file"
        return 0
    fi

    return 1
}

# Find JSON tracking file for a sprint plan
# Args: $1 = sprint markdown file, $2 = design doc being audited
find_json_for_sprint() {
    local sprint_md="$1"
    local design_doc="$2"
    local design_doc_name=$(basename "$design_doc")

    # Priority 1: Look for JSON that has matching markdown_plan field
    for json in sprints/*.json; do
        if [ -f "$json" ]; then
            local md_plan=$(jq -r '.markdown_plan // empty' "$json" 2>/dev/null)
            if [ -n "$md_plan" ]; then
                local md_plan_base=$(basename "$md_plan")
                local sprint_md_base=$(basename "$sprint_md")
                if [[ "$md_plan_base" == "$sprint_md_base" ]]; then
                    echo "$json"
                    return 0
                fi
            fi
        fi
    done

    # Priority 2: Look for JSON that references the same design doc
    for json in sprints/*.json; do
        if [ -f "$json" ]; then
            local json_design_doc=$(jq -r '.design_doc // empty' "$json" 2>/dev/null)
            if [ -n "$json_design_doc" ] && [[ "$json_design_doc" == *"$design_doc_name"* ]]; then
                echo "$json"
                return 0
            fi
        fi
    done

    return 1
}

# Get progress from JSON file (source of truth)
get_progress_from_json() {
    local json_file="$1"

    if [ ! -f "$json_file" ]; then
        echo "0:0:pending"
        return
    fi

    # Check if it has features array (modern format)
    local has_features=$(jq 'has("features")' "$json_file" 2>/dev/null)
    local status=$(jq -r '.status // "pending"' "$json_file" 2>/dev/null)

    if [ "$has_features" = "true" ]; then
        local total=$(jq '[.features[]] | length' "$json_file" 2>/dev/null || echo "0")
        local completed=$(jq '[.features[] | select(.passes == true)] | length' "$json_file" 2>/dev/null || echo "0")
        echo "$completed:$total:$status"
    else
        # Fallback to phases/tasks format
        local total=$(jq '[.phases[].tasks[]] | length' "$json_file" 2>/dev/null || echo "0")
        local completed=$(jq '[.phases[].tasks[] | select(.status == "completed")] | length' "$json_file" 2>/dev/null || echo "0")
        echo "$completed:$total:$status"
    fi
}

# Check a design doc
audit_doc() {
    local doc="$1"
    local name=$(basename "$doc" .md)

    echo -e "${BLUE}Checking: $name${NC}"

    # Look for corresponding sprint
    if sprint_file=$(find_sprint_for_doc "$doc"); then
        echo -e "  ${GREEN}✓${NC} Has sprint: $sprint_file"

        # Show what files the sprint says it creates
        if grep -q "Files to Create\|New Files\|Files to create" "$sprint_file"; then
            echo "  Sprint defines implementation files"
        fi

        # Try to find JSON tracking file for progress (SOURCE OF TRUTH)
        local json_file=""
        if json_file=$(find_json_for_sprint "$sprint_file" "$doc"); then
            echo "  Tracking: $json_file"

            # Get progress from JSON
            local progress=$(get_progress_from_json "$json_file")
            local completed=$(echo "$progress" | cut -d: -f1)
            local total=$(echo "$progress" | cut -d: -f2)
            local status=$(echo "$progress" | cut -d: -f3)

            if [ "$total" -gt 0 ]; then
                local pct=$((completed * 100 / total))
                if [ "$status" = "completed" ] || [ "$pct" -eq 100 ]; then
                    echo -e "  ${GREEN}✓ Sprint complete ($completed/$total features) [status: $status]${NC}"
                elif [ "$status" = "in_progress" ]; then
                    echo -e "  ${YELLOW}→ Sprint in progress: $pct% ($completed/$total features)${NC}"
                else
                    echo -e "  ${YELLOW}○ Sprint pending: $pct% ($completed/$total features)${NC}"
                fi
            else
                echo -e "  ${YELLOW}○ Sprint status: $status (no features defined)${NC}"
            fi
        else
            # Fallback: count markdown checkboxes (for sprints without JSON)
            local checked=$(grep -c '\[x\]' "$sprint_file" 2>/dev/null | tr -d '\n' || echo "0")
            local unchecked=$(grep -c '\[ \]' "$sprint_file" 2>/dev/null | tr -d '\n' || echo "0")
            checked=${checked:-0}
            unchecked=${unchecked:-0}
            local total=$((checked + unchecked))

            if [ "$total" -gt 0 ]; then
                local pct=$((checked * 100 / total))
                echo -e "  ${YELLOW}⚠ No JSON tracking file - using markdown checkboxes${NC}"
                if [ "$pct" -eq 100 ]; then
                    echo -e "  ${GREEN}✓ Sprint complete ($checked/$total tasks)${NC}"
                else
                    echo -e "  ${YELLOW}○ Sprint progress: $pct% ($checked/$total tasks)${NC}"
                fi
            fi
        fi

        echo -e "  ${GREEN}Status: HAS SPRINT${NC}"
        HAS_SPRINT+=("$name:$sprint_file")
    else
        echo -e "  ${RED}✗${NC} No sprint file found"
        echo -e "  ${RED}Status: NO SPRINT${NC}"
        NO_SPRINT+=("$name")
    fi
    echo ""
}

# Process all docs in phased folders (in dependency order)
PHASES=(
    "phase0-architecture"
    "phase1-data-models"
    "phase2-core-views"
    "phase3-gameplay"
    "phase4-polish"
)

for phase in "${PHASES[@]}"; do
    phase_dir="design_docs/planned/$phase"
    if [ -d "$phase_dir" ]; then
        echo "=========================================="
        echo "Scanning $phase_dir..."
        echo "=========================================="
        echo ""

        for doc in "$phase_dir"/*.md; do
            if [ -f "$doc" ] && [[ "$(basename "$doc")" != "README.md" ]]; then
                audit_doc "$doc"
            fi
        done
    fi
done

# Check for orphan sprints (sprints without design docs)
echo "Checking for orphan sprints..."
echo ""

for sprint in sprints/*.md sprints/**/*.md; do
    if [ -f "$sprint" ]; then
        sprint_name=$(basename "$sprint" .md)
        # Check if sprint references any design doc
        if ! grep -q "design_docs\|Design Doc:" "$sprint" 2>/dev/null; then
            echo -e "${YELLOW}Warning: Sprint without design doc: $sprint${NC}"
            ORPHAN_SPRINTS+=("$sprint_name")
        fi
    fi
done

echo ""

# Summary
echo "=========================================="
echo "SUMMARY"
echo "=========================================="
echo ""

if [ ${#HAS_SPRINT[@]} -gt 0 ]; then
    echo -e "${GREEN}Design docs WITH sprints (can be implemented):${NC}"
    for item in "${HAS_SPRINT[@]}"; do
        doc=$(echo "$item" | cut -d: -f1)
        sprint=$(echo "$item" | cut -d: -f2-)
        echo "  - $doc.md"
        echo "    └── $sprint"
    done
    echo ""
fi

if [ ${#NO_SPRINT[@]} -gt 0 ]; then
    echo -e "${RED}Design docs WITHOUT sprints (need planning):${NC}"
    for name in "${NO_SPRINT[@]}"; do
        echo "  - $name.md"
    done
    echo ""
    echo "To add a sprint, use: invoke sprint-planner skill"
    echo ""
fi

if [ ${#ORPHAN_SPRINTS[@]} -gt 0 ]; then
    echo -e "${YELLOW}Orphan sprints (no design doc):${NC}"
    for name in "${ORPHAN_SPRINTS[@]}"; do
        echo "  - $name"
    done
    echo ""
fi

echo "=========================================="
echo "Phased Implementation Order:"
echo "  Phase 0: Architecture (MUST DO FIRST) - blocks all other phases"
echo "  Phase 1: Data Models - galaxy, planet, ship structures"
echo "  Phase 2: Core Views - galaxy map, ship exploration, bridge"
echo "  Phase 3: Gameplay - journey system (core mechanic)"
echo "  Phase 4: Polish - arrival cinematics, camera systems"
echo ""
echo "Workflow:"
echo "  1. Work through phases in order (0 → 1 → 2 → 3 → 4)"
echo "  2. Use sprint-planner to create sprint plan for each doc"
echo "  3. Execute sprint (tracks implementation)"
echo "  4. When complete, move doc to design_docs/implemented/vX_Y_Z/"
echo "=========================================="
