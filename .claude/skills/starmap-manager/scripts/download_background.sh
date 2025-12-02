#!/bin/bash
# Download galactic background imagery from ESA
# Usage: download_background.sh [resolution]
#   resolution: 2k, 4k (default), 8k

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../../.." && pwd)"
OUTPUT_DIR="$PROJECT_ROOT/assets/data/starmap/background"

RESOLUTION="${1:-4k}"

mkdir -p "$OUTPUT_DIR"

echo "=== ESA Gaia All-Sky Map Downloader ==="
echo "Source: ESA/Gaia/DPAC"
echo "License: CC BY-SA 3.0 IGO"
echo "Credit: ESA/Gaia/DPAC"
echo ""

case "$RESOLUTION" in
    2k)
        echo "Downloading 2K resolution (~1024x768)..."
        echo "  Size: ~800 KB"
        # ESO CDN - Gaia's view of the Milky Way (eso1908e)
        URL="https://cdn.eso.org/images/wallpaper2/eso1908e.jpg"
        OUTPUT="$OUTPUT_DIR/galaxy_2k.jpg"
        ;;
    4k)
        echo "Downloading 4K resolution (publication quality)..."
        echo "  Size: ~2.6 MB"
        URL="https://cdn.eso.org/images/publicationjpg/eso1908e.jpg"
        OUTPUT="$OUTPUT_DIR/galaxy_4k.jpg"
        ;;
    8k)
        echo "Downloading 8K resolution (large, ~15MB)..."
        echo "  Size: ~15 MB"
        URL="https://cdn.eso.org/images/large/eso1908e.jpg"
        OUTPUT="$OUTPUT_DIR/galaxy_8k.jpg"
        ;;
    *)
        echo "ERROR: Unknown resolution '$RESOLUTION'"
        echo "Usage: $0 [2k|4k|8k]"
        echo ""
        echo "Resolutions:"
        echo "  2k - ~1024x768, ~800 KB (low-end devices)"
        echo "  4k - Publication quality, ~2.6 MB (default, recommended)"
        echo "  8k - Large, ~15 MB (HD/4K displays)"
        exit 1
        ;;
esac

echo ""

# Try primary URL
if curl -L --fail -o "$OUTPUT" "$URL" 2>/dev/null; then
    echo "  Downloaded: $(basename "$OUTPUT")"
    echo "  Size: $(du -h "$OUTPUT" | cut -f1)"
else
    echo "  Primary URL failed, trying alternative sources..."

    # Alternative: ESO archive
    ALT_URL="https://cdn.eso.org/images/original/ESA_Gaia_DR2_AllSky_Brightness_Colour_black_bg_${RESOLUTION}.png"
    if curl -L --fail -o "$OUTPUT" "$ALT_URL" 2>/dev/null; then
        echo "  Downloaded from ESO: $(basename "$OUTPUT")"
    else
        # Fallback: Direct sci.esa.int
        case "$RESOLUTION" in
            4k)
                FALLBACK="https://sci.esa.int/documents/33565/0/Gaia_EDR3_flux_equirect_4096x2048.png"
                ;;
            8k)
                FALLBACK="https://sci.esa.int/documents/33565/0/Gaia_EDR3_flux_equirect_8192x4096.png"
                ;;
            *)
                FALLBACK=""
                ;;
        esac

        if [ -n "$FALLBACK" ] && curl -L --fail -o "$OUTPUT" "$FALLBACK" 2>/dev/null; then
            echo "  Downloaded from sci.esa.int: $(basename "$OUTPUT")"
        else
            echo "  ERROR: Could not download galactic background"
            echo ""
            echo "  Manual download options:"
            echo "  1. ESA Gaia Archive: https://www.cosmos.esa.int/web/gaia/edr3-gcns"
            echo "  2. ESO Image Archive: https://www.eso.org/public/images/eso1908e/"
            echo "  3. Sci.esa.int: https://sci.esa.int/web/gaia/-/60196-gaia-s-sky-in-colour"
            exit 1
        fi
    fi
fi

echo ""
echo "Background download complete!"
echo ""
echo "Files in $OUTPUT_DIR:"
ls -lh "$OUTPUT_DIR"/ 2>/dev/null || echo "  (empty)"
echo ""
echo "IMPORTANT: Add credit to game:"
echo '  "Galaxy imagery: ESA/Gaia/DPAC (CC BY-SA 3.0 IGO)"'
