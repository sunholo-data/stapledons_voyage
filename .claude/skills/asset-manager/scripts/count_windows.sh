#!/bin/bash
# Use AI vision to count window regions in a generated deck image
#
# Usage: count_windows.sh <image_path>

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../../.." && pwd)"

if [[ $# -lt 1 ]]; then
    echo "Usage: $0 <image_path>"
    echo ""
    echo "Uses AI vision to count distinct window regions in the image."
    echo "Returns a single integer: the number of windows to extract."
    exit 1
fi

IMAGE_PATH="$1"

if [[ ! -f "$IMAGE_PATH" ]]; then
    echo "Error: Image not found: $IMAGE_PATH"
    exit 1
fi

# Use voyage CLI with AI vision to analyze the image
PROMPT="Look at this spaceship interior scene. Count how many distinct window regions show bright white light (these are the windows showing space outside).

IMPORTANT: Only count separate, distinct window regions. If there are many small geometric panels that form one large dome window, count that as ONE window region, not multiple.

Examples:
- One large panoramic window = 1
- Three separate viewport windows = 3
- A geodesic dome made of many triangular panels = 1 (it's one dome)
- Two curved observation windows on the sides = 2

Return ONLY a single integer (1-5) representing the number of distinct window regions to extract for masking."

# Call AI with image analysis (Gemini vision)
RESULT=$(go run "$PROJECT_ROOT/cmd/voyage" ai -image "$IMAGE_PATH" -prompt "$PROMPT" -provider gemini 2>&1)

# Extract the AI response (skip log lines)
AI_RESPONSE=$(echo "$RESULT" | grep -v "Warning:" | grep -v "\[Gemini\]" | grep -v "Using Vertex" | tail -1)

# Extract just the number from the AI's response (should be 1-10)
COUNT=$(echo "$AI_RESPONSE" | grep -oE '\b[1-9]\b|\b10\b' | head -1)

if [[ -z "$COUNT" ]] || [[ "$COUNT" -lt 1 ]] || [[ "$COUNT" -gt 10 ]]; then
    echo "Warning: AI returned invalid count ($COUNT), defaulting to 3" >&2
    COUNT=3
fi

# Output just the number (for script chaining)
echo "$COUNT"
