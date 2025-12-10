#!/bin/bash
#
# Game Screenshot Helper
#
# Takes screenshots using the in-game screenshot functionality (NOT macOS screencapture).
# This captures the game's internal render buffer at 1280x960, producing consistent
# images regardless of display resolution.
#
# Usage:
#   .claude/skills/sprint-executor/scripts/take_screenshot.sh [options]
#
# Options:
#   -c, --command CMD    Command to run (default: game)
#                        Options: game, demo-bridge, demo-saturn, demo-arrival, demo-view
#   -f, --frame N        Capture at frame N (default: 30)
#   -o, --output PATH    Output path (default: out/screenshots/<command>-<timestamp>.png)
#   -s, --seed N         World seed (default: 42)
#   --arrival            Use arrival mode (black hole sequence)
#   --effects LIST       Enable effects (bloom,vignette,crt,aberration,sr_warp)
#   --velocity V         Ship velocity 0.0-0.99 for SR effects
#   -h, --help           Show this help
#
# Examples:
#   # Basic game screenshot at frame 30
#   ./take_screenshot.sh
#
#   # Bridge demo at frame 60
#   ./take_screenshot.sh -c demo-bridge -f 60
#
#   # Game with effects
#   ./take_screenshot.sh --effects bloom,sr_warp --velocity 0.5
#
#   # Arrival sequence at frame 120
#   ./take_screenshot.sh --arrival -f 120
#
#   # Custom output path
#   ./take_screenshot.sh -o out/screenshots/my-test.png

set -e

# Defaults
COMMAND="game"
FRAME=30
OUTPUT=""
SEED=42
ARRIVAL=""
EFFECTS=""
VELOCITY=""
EXTRA_ARGS=""

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -c|--command)
            COMMAND="$2"
            shift 2
            ;;
        -f|--frame)
            FRAME="$2"
            shift 2
            ;;
        -o|--output)
            OUTPUT="$2"
            shift 2
            ;;
        -s|--seed)
            SEED="$2"
            shift 2
            ;;
        --arrival)
            ARRIVAL="--arrival"
            shift
            ;;
        --effects)
            EFFECTS="$2"
            shift 2
            ;;
        --velocity)
            VELOCITY="$2"
            shift 2
            ;;
        -h|--help)
            head -40 "$0" | tail -35
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Set default output path if not specified
if [ -z "$OUTPUT" ]; then
    TIMESTAMP=$(date +%Y%m%d-%H%M%S)
    mkdir -p out/screenshots
    OUTPUT="out/screenshots/${COMMAND}-${TIMESTAMP}.png"
fi

# Ensure output directory exists
mkdir -p "$(dirname "$OUTPUT")"

# Build command arguments
ARGS="--screenshot $FRAME --output $OUTPUT --seed $SEED"

if [ -n "$ARRIVAL" ]; then
    ARGS="$ARGS $ARRIVAL"
fi

if [ -n "$EFFECTS" ]; then
    ARGS="$ARGS --effects $EFFECTS"
fi

if [ -n "$VELOCITY" ]; then
    ARGS="$ARGS --velocity $VELOCITY"
fi

# Map command shorthand to full path
case $COMMAND in
    game)
        CMD_PATH="./cmd/game"
        ;;
    demo-bridge|bridge)
        CMD_PATH="./cmd/demo-bridge"
        ;;
    demo-saturn|saturn)
        CMD_PATH="./cmd/demo-saturn"
        ;;
    demo-arrival|arrival)
        CMD_PATH="./cmd/demo-arrival"
        ;;
    demo-view|view)
        CMD_PATH="./cmd/demo-view"
        ;;
    demo-sr-flyby|sr-flyby|flyby)
        CMD_PATH="./cmd/demo-sr-flyby"
        ;;
    *)
        # Assume it's a full path
        CMD_PATH="./cmd/$COMMAND"
        ;;
esac

# Check if command exists
if [ ! -d "$CMD_PATH" ]; then
    echo "Error: Command not found: $CMD_PATH"
    echo ""
    echo "Available commands:"
    ls -1 cmd/ | grep -E "^(game|demo-)" | sed 's/^/  /'
    exit 1
fi

# Run the screenshot
echo "Taking screenshot..."
echo "  Command: go run $CMD_PATH"
echo "  Frame: $FRAME"
echo "  Output: $OUTPUT"
echo ""

# Run the command, capturing output
# Note: The game exits with "screenshot complete" error which is actually success
go run "$CMD_PATH" $ARGS 2>&1 || true

# Check if screenshot was actually created
if [ -f "$OUTPUT" ]; then
    echo ""
    echo "Screenshot saved: $OUTPUT"

    # Show file info
    if command -v file &> /dev/null; then
        echo "File info: $(file "$OUTPUT" | cut -d: -f2)"
    fi

    # Show file size
    SIZE=$(ls -lh "$OUTPUT" | awk '{print $5}')
    echo "Size: $SIZE"
else
    echo ""
    echo "Error: Screenshot file not created"
    exit 1
fi
