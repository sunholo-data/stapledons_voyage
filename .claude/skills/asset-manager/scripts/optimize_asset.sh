#!/bin/bash
# Asset Manager - Optimize Asset Script
# Resizes AI-generated images (1024x1024) to game-ready dimensions
# Uses Go's image package via a small utility (no ImageMagick dependency)

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../../.." && pwd)"

usage() {
    echo "Usage: $0 <image_path> [target_width] [target_height]"
    echo ""
    echo "Resizes an image to game-ready dimensions."
    echo "If dimensions not specified, auto-detects based on asset type from path."
    echo ""
    echo "Auto-detected sizes by path:"
    echo "  iso_tiles/      -> 64x32"
    echo "  bridge/tiles/   -> 64x32"
    echo "  iso_entities/   -> 128x48"
    echo "  bridge/crew/    -> 128x128 (or 256x256 for high-res)"
    echo "  bridge/consoles/ -> 128x96"
    echo "  stars/          -> 16x16"
    echo "  planets/        -> 256x256"
    echo "  portraits/      -> 128x128"
    echo ""
    echo "Examples:"
    echo "  $0 assets/sprites/bridge/tiles/floor.png"
    echo "  $0 assets/sprites/bridge/consoles/helm.png 128 96"
    echo "  $0 assets/generated/response_123.png 64 32"
    echo ""
    exit 1
}

if [[ $# -lt 1 ]]; then
    usage
fi

IMAGE_PATH="$1"
TARGET_WIDTH="${2:-}"
TARGET_HEIGHT="${3:-}"

if [[ ! -f "$IMAGE_PATH" ]]; then
    echo "ERROR: File not found: $IMAGE_PATH"
    exit 1
fi

# Auto-detect target size from path if not specified
if [[ -z "$TARGET_WIDTH" ]]; then
    case "$IMAGE_PATH" in
        *iso_tiles* | *bridge/tiles*)
            TARGET_WIDTH=64
            TARGET_HEIGHT=32
            ;;
        *iso_entities*)
            TARGET_WIDTH=128
            TARGET_HEIGHT=48
            ;;
        *bridge/crew*)
            TARGET_WIDTH=128
            TARGET_HEIGHT=128
            ;;
        *bridge/consoles*)
            TARGET_WIDTH=128
            TARGET_HEIGHT=96
            ;;
        *stars*)
            TARGET_WIDTH=16
            TARGET_HEIGHT=16
            ;;
        *planets*)
            TARGET_WIDTH=256
            TARGET_HEIGHT=256
            ;;
        *portraits*)
            TARGET_WIDTH=128
            TARGET_HEIGHT=128
            ;;
        *)
            echo "ERROR: Cannot auto-detect size for path: $IMAGE_PATH"
            echo "Please specify target dimensions manually."
            usage
            ;;
    esac
fi

echo "=== Image Optimization ==="
echo "Source: $IMAGE_PATH"
echo "Target size: ${TARGET_WIDTH}x${TARGET_HEIGHT}"
echo ""

# Get current dimensions
CURRENT_DIMS=$(file "$IMAGE_PATH" | grep -oE '[0-9]+ x [0-9]+')
echo "Current size: $CURRENT_DIMS"

# Check if resize is needed
if [[ "$CURRENT_DIMS" == "${TARGET_WIDTH} x ${TARGET_HEIGHT}" ]]; then
    echo "Already at target size, no resize needed."
    exit 0
fi

# Create backup
BACKUP_PATH="${IMAGE_PATH%.png}_original.png"
if [[ ! -f "$BACKUP_PATH" ]]; then
    cp "$IMAGE_PATH" "$BACKUP_PATH"
    echo "Backup saved: $BACKUP_PATH"
fi

# Use Go to resize (creates a small inline program)
# This avoids ImageMagick dependency
GO_RESIZE=$(cat <<'GOCODE'
package main

import (
    "fmt"
    "image"
    "image/png"
    "os"
    "strconv"

    "golang.org/x/image/draw"
)

func main() {
    if len(os.Args) != 5 {
        fmt.Fprintln(os.Stderr, "Usage: resize <input> <output> <width> <height>")
        os.Exit(1)
    }

    inputPath := os.Args[1]
    outputPath := os.Args[2]
    width, _ := strconv.Atoi(os.Args[3])
    height, _ := strconv.Atoi(os.Args[4])

    // Open input
    inFile, err := os.Open(inputPath)
    if err != nil {
        fmt.Fprintln(os.Stderr, "Error opening input:", err)
        os.Exit(1)
    }
    defer inFile.Close()

    // Decode
    src, _, err := image.Decode(inFile)
    if err != nil {
        fmt.Fprintln(os.Stderr, "Error decoding:", err)
        os.Exit(1)
    }

    // Create destination
    dst := image.NewRGBA(image.Rect(0, 0, width, height))

    // Scale using high-quality interpolation
    draw.CatmullRom.Scale(dst, dst.Bounds(), src, src.Bounds(), draw.Over, nil)

    // Save output
    outFile, err := os.Create(outputPath)
    if err != nil {
        fmt.Fprintln(os.Stderr, "Error creating output:", err)
        os.Exit(1)
    }
    defer outFile.Close()

    if err := png.Encode(outFile, dst); err != nil {
        fmt.Fprintln(os.Stderr, "Error encoding:", err)
        os.Exit(1)
    }

    fmt.Printf("Resized %s -> %dx%d\n", inputPath, width, height)
}
GOCODE
)

# Check if we have the resize tool built
RESIZE_TOOL="$PROJECT_ROOT/bin/resize-image"

if [[ ! -f "$RESIZE_TOOL" ]]; then
    echo "Building resize tool..."
    TEMP_DIR=$(mktemp -d)
    echo "$GO_RESIZE" > "$TEMP_DIR/resize.go"

    # Initialize module and get dependency
    (cd "$TEMP_DIR" && \
        go mod init resize && \
        go get golang.org/x/image/draw && \
        go build -o "$RESIZE_TOOL" resize.go)

    rm -rf "$TEMP_DIR"
    echo "Resize tool built: $RESIZE_TOOL"
fi

# Perform resize
TEMP_OUTPUT=$(mktemp).png
"$RESIZE_TOOL" "$IMAGE_PATH" "$TEMP_OUTPUT" "$TARGET_WIDTH" "$TARGET_HEIGHT"

# Replace original
mv "$TEMP_OUTPUT" "$IMAGE_PATH"

# Verify new size
NEW_DIMS=$(file "$IMAGE_PATH" | grep -oE '[0-9]+ x [0-9]+')
echo ""
echo "=== Done ==="
echo "New size: $NEW_DIMS"
echo "Original backup: $BACKUP_PATH"
