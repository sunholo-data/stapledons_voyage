#!/bin/bash
#
# Simple Sprint Progress Update Script
#
# Updates the sprint JSON file with task/phase completion status.
# Designed to be called by Claude during sprint execution.
#
# Usage:
#   update_progress.sh <sprint_file> task <task_id> <status>
#   update_progress.sh <sprint_file> phase <phase_id> <status>
#   update_progress.sh <sprint_file> feature <feature_id> <status> [actual_loc]
#   update_progress.sh <sprint_file> sprint <status>
#   update_progress.sh <sprint_file> show
#
# Status values: pending | in_progress | completed | blocked
#
# Examples:
#   update_progress.sh sprints/003-player-interaction.json task 1.1 completed
#   update_progress.sh sprints/003-player-interaction.json phase phase-1 completed
#   update_progress.sh sprints/003-player-interaction.json feature core-interfaces completed 604
#   update_progress.sh sprints/003-player-interaction.json sprint in_progress
#   update_progress.sh sprints/003-player-interaction.json show

set -e

if [ $# -lt 2 ]; then
    echo "Usage: $0 <sprint_file> <command> [args...]"
    echo ""
    echo "Commands:"
    echo "  task <task_id> <status>       Update task status"
    echo "  phase <phase_id> <status>     Update phase status"
    echo "  feature <id> <status> [loc]   Update feature status (sets timestamps)"
    echo "  sprint <status>               Update overall sprint status"
    echo "  show                          Show current progress summary"
    echo ""
    echo "Status values: pending | in_progress | completed | blocked"
    exit 1
fi

SPRINT_FILE="$1"
COMMAND="$2"

if [ ! -f "$SPRINT_FILE" ]; then
    echo "Error: Sprint file not found: $SPRINT_FILE"
    exit 1
fi

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

case "$COMMAND" in
    task)
        TASK_ID="$3"
        STATUS="$4"
        if [ -z "$TASK_ID" ] || [ -z "$STATUS" ]; then
            echo "Usage: $0 $SPRINT_FILE task <task_id> <status>"
            exit 1
        fi

        # Update task status using jq
        UPDATED=$(jq --arg tid "$TASK_ID" --arg status "$STATUS" '
            .phases |= map(
                .tasks |= map(
                    if .id == $tid then .status = $status else . end
                )
            )
        ' "$SPRINT_FILE")

        echo "$UPDATED" > "$SPRINT_FILE"
        echo -e "${GREEN}✓${NC} Task $TASK_ID → $STATUS"
        ;;

    phase)
        PHASE_ID="$3"
        STATUS="$4"
        if [ -z "$PHASE_ID" ] || [ -z "$STATUS" ]; then
            echo "Usage: $0 $SPRINT_FILE phase <phase_id> <status>"
            exit 1
        fi

        # Update phase status
        UPDATED=$(jq --arg pid "$PHASE_ID" --arg status "$STATUS" '
            .phases |= map(
                if .id == $pid then .status = $status else . end
            )
        ' "$SPRINT_FILE")

        echo "$UPDATED" > "$SPRINT_FILE"
        echo -e "${GREEN}✓${NC} Phase $PHASE_ID → $STATUS"
        ;;

    feature)
        FEATURE_ID="$3"
        STATUS="$4"
        ACTUAL_LOC="$5"
        if [ -z "$FEATURE_ID" ] || [ -z "$STATUS" ]; then
            echo "Usage: $0 $SPRINT_FILE feature <feature_id> <status> [actual_loc]"
            exit 1
        fi

        NOW=$(date -u +%Y-%m-%dT%H:%M:%SZ)

        case "$STATUS" in
            in_progress)
                # Set started timestamp if not already set
                UPDATED=$(jq --arg fid "$FEATURE_ID" --arg ts "$NOW" '
                    .features |= map(
                        if .id == $fid then
                            .started = (if .started == null then $ts else .started end) |
                            .passes = null
                        else . end
                    )
                ' "$SPRINT_FILE")
                ;;
            completed)
                # Set completed timestamp and passes=true
                if [ -n "$ACTUAL_LOC" ]; then
                    UPDATED=$(jq --arg fid "$FEATURE_ID" --arg ts "$NOW" --argjson loc "$ACTUAL_LOC" '
                        .features |= map(
                            if .id == $fid then
                                .completed = $ts |
                                .passes = true |
                                .actual_loc = $loc
                            else . end
                        )
                    ' "$SPRINT_FILE")
                else
                    UPDATED=$(jq --arg fid "$FEATURE_ID" --arg ts "$NOW" '
                        .features |= map(
                            if .id == $fid then
                                .completed = $ts |
                                .passes = true
                            else . end
                        )
                    ' "$SPRINT_FILE")
                fi
                ;;
            blocked)
                # Set passes=false
                UPDATED=$(jq --arg fid "$FEATURE_ID" '
                    .features |= map(
                        if .id == $fid then
                            .passes = false
                        else . end
                    )
                ' "$SPRINT_FILE")
                ;;
            pending)
                # Reset to pending state
                UPDATED=$(jq --arg fid "$FEATURE_ID" '
                    .features |= map(
                        if .id == $fid then
                            .started = null |
                            .completed = null |
                            .passes = null |
                            .actual_loc = null
                        else . end
                    )
                ' "$SPRINT_FILE")
                ;;
            *)
                echo "Unknown status: $STATUS"
                echo "Valid statuses: pending | in_progress | completed | blocked"
                exit 1
                ;;
        esac

        echo "$UPDATED" > "$SPRINT_FILE"
        echo -e "${GREEN}✓${NC} Feature $FEATURE_ID → $STATUS"
        ;;

    sprint)
        STATUS="$3"
        if [ -z "$STATUS" ]; then
            echo "Usage: $0 $SPRINT_FILE sprint <status>"
            exit 1
        fi

        # Update sprint status and timestamp
        UPDATED=$(jq --arg status "$STATUS" --arg ts "$(date -u +%Y-%m-%dT%H:%M:%SZ)" '
            .status = $status |
            .last_updated = $ts
        ' "$SPRINT_FILE")

        echo "$UPDATED" > "$SPRINT_FILE"
        echo -e "${GREEN}✓${NC} Sprint status → $STATUS"
        ;;

    show)
        echo "═══════════════════════════════════════════════════════════════"
        echo " Sprint Progress: $(jq -r '.sprint_id' "$SPRINT_FILE")"
        echo "═══════════════════════════════════════════════════════════════"
        echo ""

        # Overall status
        STATUS=$(jq -r '.status' "$SPRINT_FILE")
        case "$STATUS" in
            completed) echo -e "Status: ${GREEN}$STATUS${NC}" ;;
            in_progress) echo -e "Status: ${YELLOW}$STATUS${NC}" ;;
            *) echo "Status: $STATUS" ;;
        esac
        echo ""

        # Check if this is a features-based or phases-based sprint
        HAS_FEATURES=$(jq 'has("features")' "$SPRINT_FILE")
        HAS_PHASES=$(jq 'has("phases")' "$SPRINT_FILE")

        if [ "$HAS_FEATURES" = "true" ]; then
            # Feature-based sprint display
            echo "Features:"
            jq -r '.features[] | "  [\(if .passes == true then "✓" elif .started != null and .completed == null then "→" else " " end)] \(.id): \(.description | .[0:50])\(if (.description | length) > 50 then "..." else "" end)"' "$SPRINT_FILE"
            echo ""

            # Feature counts
            TOTAL=$(jq '[.features[]] | length' "$SPRINT_FILE")
            COMPLETED=$(jq '[.features[] | select(.passes == true)] | length' "$SPRINT_FILE")
            IN_PROGRESS=$(jq '[.features[] | select(.started != null and .completed == null)] | length' "$SPRINT_FILE")

            echo "Features: $COMPLETED/$TOTAL completed"
            if [ "$IN_PROGRESS" -gt 0 ]; then
                echo -e "${YELLOW}In progress: $IN_PROGRESS${NC}"
            fi
            echo ""

            # Show in-progress features
            if [ "$IN_PROGRESS" -gt 0 ]; then
                echo "Currently working on:"
                jq -r '.features[] | select(.started != null and .completed == null) | "  → \(.id)"' "$SPRINT_FILE"
                echo ""
            fi

            # Show next pending feature
            NEXT=$(jq -r '[.features[] | select(.started == null)][0] | "  \(.id): \(.description | .[0:60])"' "$SPRINT_FILE" 2>/dev/null || echo "")
            if [ -n "$NEXT" ] && [ "$NEXT" != "  null: null" ]; then
                echo "Next up:"
                echo "$NEXT"
            fi

        elif [ "$HAS_PHASES" = "true" ]; then
            # Phase-based sprint display (legacy)
            echo "Phases:"
            jq -r '.phases[] | "  [\(.status | if . == "completed" then "✓" elif . == "in_progress" then "→" else " " end)] \(.name)"' "$SPRINT_FILE"
            echo ""

            # Task counts
            TOTAL=$(jq '[.phases[].tasks[]] | length' "$SPRINT_FILE")
            COMPLETED=$(jq '[.phases[].tasks[] | select(.status == "completed")] | length' "$SPRINT_FILE")
            IN_PROGRESS=$(jq '[.phases[].tasks[] | select(.status == "in_progress")] | length' "$SPRINT_FILE")

            echo "Tasks: $COMPLETED/$TOTAL completed"
            if [ "$IN_PROGRESS" -gt 0 ]; then
                echo -e "${YELLOW}In progress: $IN_PROGRESS${NC}"
            fi
            echo ""

            # Show in-progress tasks
            if [ "$IN_PROGRESS" -gt 0 ]; then
                echo "Currently working on:"
                jq -r '.phases[].tasks[] | select(.status == "in_progress") | "  → \(.id): \(.desc)"' "$SPRINT_FILE"
                echo ""
            fi

            # Show next pending task
            NEXT=$(jq -r '[.phases[].tasks[] | select(.status == "pending")][0] | "  \(.id): \(.desc)"' "$SPRINT_FILE" 2>/dev/null || echo "")
            if [ -n "$NEXT" ] && [ "$NEXT" != "  null: null" ]; then
                echo "Next up:"
                echo "$NEXT"
            fi
        else
            echo "No phases or features found in sprint file."
        fi
        ;;

    *)
        echo "Unknown command: $COMMAND"
        echo "Valid commands: task, phase, feature, sprint, show"
        exit 1
        ;;
esac
