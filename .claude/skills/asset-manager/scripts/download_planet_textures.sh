#!/bin/bash
# Asset Manager - Download Equirectangular Planet Textures
# Downloads proper UV-mappable textures for 3D sphere rendering
# Source: Solar System Scope (CC BY 4.0) - https://www.solarsystemscope.com/textures/

set -e

usage() {
    echo "Usage: $0 [resolution]"
    echo ""
    echo "Resolutions:"
    echo "  2k    - 2048x1024 (default, good for game use)"
    echo "  4k    - 4096x2048 (high quality)"
    echo "  8k    - 8192x4096 (very high quality, large files)"
    echo ""
    echo "Downloads equirectangular planet textures to assets/planets/"
    echo "These are 2:1 aspect ratio maps suitable for sphere UV mapping."
    exit 1
}

RES="${1:-2k}"

# Find project root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../../.." && pwd)"
DEST_DIR="$PROJECT_ROOT/assets/planets"

mkdir -p "$DEST_DIR"

# Base URL for Solar System Scope textures
# License: CC BY 4.0 - Attribution required
BASE_URL="https://www.solarsystemscope.com/textures/download"

echo "=== Downloading Equirectangular Planet Textures ==="
echo "Resolution: $RES"
echo "Destination: $DEST_DIR/"
echo "Source: Solar System Scope (CC BY 4.0)"
echo ""

download_texture() {
    local name="$1"
    local filename="$2"
    local desc="$3"

    echo "Downloading $name ($desc)..."

    # Solar System Scope URL format
    local url="${BASE_URL}/${RES}_${filename}.jpg"
    local dest="$DEST_DIR/${name}.jpg"

    if curl -s -L -o "$dest" "$url" 2>/dev/null; then
        # Verify it's actually an image
        if file "$dest" | grep -q "JPEG"; then
            local size=$(du -h "$dest" | cut -f1)
            local dims=$(file "$dest" | grep -oE '[0-9]+x[0-9]+' | head -1)
            echo "  ✓ $name: $dims, $size"
        else
            echo "  ✗ $name: Download failed (not a valid image)"
            rm -f "$dest"
        fi
    else
        echo "  ✗ $name: Network error"
    fi
}

# Solar System planets - using Solar System Scope's file naming
download_texture "mercury" "mercury" "rocky planet, cratered"
download_texture "venus_surface" "venus_surface" "surface map (radar)"
download_texture "venus_atmosphere" "venus_atmosphere" "cloud layer"
download_texture "earth_daymap" "earth_daymap" "day side"
download_texture "earth_nightmap" "earth_nightmap" "city lights"
download_texture "earth_clouds" "earth_clouds" "cloud layer"
download_texture "moon" "moon" "lunar surface"
download_texture "mars" "mars" "red planet"
download_texture "jupiter" "jupiter" "gas giant"
download_texture "saturn" "saturn" "gas giant (no rings)"
download_texture "saturn_ring" "saturn_ring_alpha" "ring system"
download_texture "uranus" "uranus" "ice giant"
download_texture "neptune" "neptune" "ice giant"
download_texture "pluto" "pluto" "dwarf planet"

echo ""
echo "=== Download Complete ==="
echo ""
echo "These textures use equirectangular projection (2:1 aspect ratio)"
echo "and are suitable for UV mapping onto 3D spheres in Tetra3D."
echo ""
echo "License: CC BY 4.0 - Credit 'Solar System Scope' in game credits."
echo "         https://www.solarsystemscope.com/textures/"
echo ""
echo "For Saturn, use saturn.jpg for the planet body and saturn_ring.jpg"
echo "for the ring system (alpha channel defines ring density)."
