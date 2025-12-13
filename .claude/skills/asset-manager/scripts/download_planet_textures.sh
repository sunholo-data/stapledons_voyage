#!/bin/bash
# Download equirectangular planet textures from Solar System Scope
# License: CC BY 4.0 - https://www.solarsystemscope.com/textures/
#
# Usage: ./download_planet_textures.sh [resolution]
#   resolution: 2k (default), 4k, 8k

set -e
cd "$(dirname "$0")/../../../../assets/planets"

RESOLUTION="${1:-2k}"

echo "=== Downloading planet textures (${RESOLUTION}) ==="

# Base URL for Solar System Scope textures
BASE_URL="https://www.solarsystemscope.com/textures/download"

# Function to download texture with retry
download_texture() {
    local name="$1"
    local url="$2"
    local output="$3"

    if [ -f "$output" ] && [ $(stat -f%z "$output" 2>/dev/null || stat -c%s "$output" 2>/dev/null) -gt 10000 ]; then
        echo "  [SKIP] $name already exists"
        return 0
    fi

    echo "  [DOWNLOAD] $name..."
    # Use browser-like headers to avoid 403
    curl -L -f -o "$output" \
        -H "User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36" \
        -H "Accept: image/webp,image/apng,image/*,*/*;q=0.8" \
        -H "Referer: https://www.solarsystemscope.com/textures/" \
        "$url" 2>/dev/null

    # Verify it's actually an image
    if file "$output" | grep -q "image\|JPEG\|PNG"; then
        local size=$(ls -lh "$output" | awk '{print $5}')
        echo "  [OK] $name ($size)"
        return 0
    else
        echo "  [FAIL] $name (not an image)"
        rm -f "$output"
        return 1
    fi
}

# Planets (most should already exist)
echo ""
echo "--- Planets ---"
download_texture "Sun" "${BASE_URL}/${RESOLUTION}_sun.jpg" "sun.jpg" || true
download_texture "Mercury" "${BASE_URL}/${RESOLUTION}_mercury.jpg" "mercury.jpg" || true
download_texture "Venus Surface" "${BASE_URL}/${RESOLUTION}_venus_surface.jpg" "venus_surface.jpg" || true
download_texture "Earth" "${BASE_URL}/${RESOLUTION}_earth_daymap.jpg" "earth_daymap.jpg" || true
download_texture "Moon" "${BASE_URL}/${RESOLUTION}_moon.jpg" "moon.jpg" || true
download_texture "Mars" "${BASE_URL}/${RESOLUTION}_mars.jpg" "mars.jpg" || true
download_texture "Jupiter" "${BASE_URL}/${RESOLUTION}_jupiter.jpg" "jupiter.jpg" || true
download_texture "Saturn" "${BASE_URL}/${RESOLUTION}_saturn.jpg" "saturn.jpg" || true
download_texture "Uranus" "${BASE_URL}/${RESOLUTION}_uranus.jpg" "uranus.jpg" || true
download_texture "Neptune" "${BASE_URL}/${RESOLUTION}_neptune.jpg" "neptune.jpg" || true

# Dwarf planets (fictional textures from SSS)
echo ""
echo "--- Dwarf Planets ---"
download_texture "Ceres" "${BASE_URL}/${RESOLUTION}_ceres_fictional.jpg" "ceres.jpg" || true
download_texture "Eris" "${BASE_URL}/${RESOLUTION}_eris_fictional.jpg" "eris.jpg" || true
download_texture "Haumea" "${BASE_URL}/${RESOLUTION}_haumea_fictional.jpg" "haumea.jpg" || true
download_texture "Makemake" "${BASE_URL}/${RESOLUTION}_makemake_fictional.jpg" "makemake.jpg" || true

echo ""
echo "=== Download complete ==="
ls -la *.jpg *.png 2>/dev/null | head -30
