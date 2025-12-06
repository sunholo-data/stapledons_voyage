#!/bin/bash
# Asset Manager - Download Reference Images
# Download real-world reference images from NASA, ESO, and other sources

set -e

usage() {
    echo "Usage: $0 <type>"
    echo ""
    echo "Types:"
    echo "  planets     - Download NASA Solar System planet images"
    echo "  backgrounds - Download ESO/ESA galaxy and nebula backgrounds"
    echo "  all         - Download all reference images"
    echo ""
    echo "Downloaded files go to assets/reference/ for use as generation references."
    exit 1
}

if [[ $# -ne 1 ]]; then
    usage
fi

TYPE="$1"

# Find project root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../../.." && pwd)"
REF_DIR="$PROJECT_ROOT/assets/reference"

mkdir -p "$REF_DIR/planets"
mkdir -p "$REF_DIR/backgrounds"

download_planets() {
    echo "=== Downloading Planet Reference Images ==="
    echo "Destination: $REF_DIR/planets/"
    echo ""

    # Wikimedia Commons images (public domain / CC licensed)
    # Using stable Wikimedia thumbnail URLs that work reliably

    echo "Downloading Mercury..."
    curl -s -L -o "$REF_DIR/planets/mercury.jpg" \
        "https://upload.wikimedia.org/wikipedia/commons/thumb/4/4a/Mercury_in_true_color.jpg/1024px-Mercury_in_true_color.jpg" 2>/dev/null || echo "  (skipped - network error)"

    echo "Downloading Venus..."
    curl -s -L -o "$REF_DIR/planets/venus.jpg" \
        "https://upload.wikimedia.org/wikipedia/commons/thumb/0/08/Venus_from_Mariner_10.jpg/1024px-Venus_from_Mariner_10.jpg" 2>/dev/null || echo "  (skipped - network error)"

    echo "Downloading Earth..."
    curl -s -L -o "$REF_DIR/planets/earth.jpg" \
        "https://upload.wikimedia.org/wikipedia/commons/thumb/c/cb/The_Blue_Marble_%28remastered%29.jpg/1024px-The_Blue_Marble_%28remastered%29.jpg" 2>/dev/null || echo "  (skipped - network error)"

    echo "Downloading Mars..."
    curl -s -L -o "$REF_DIR/planets/mars.jpg" \
        "https://upload.wikimedia.org/wikipedia/commons/thumb/0/02/OSIRIS_Mars_true_color.jpg/1280px-OSIRIS_Mars_true_color.jpg" 2>/dev/null || echo "  (skipped - network error)"

    echo "Downloading Jupiter..."
    curl -s -L -o "$REF_DIR/planets/jupiter.jpg" \
        "https://upload.wikimedia.org/wikipedia/commons/thumb/2/2b/Jupiter_and_its_shrunken_Great_Red_Spot.jpg/1280px-Jupiter_and_its_shrunken_Great_Red_Spot.jpg" 2>/dev/null || echo "  (skipped - network error)"

    echo "Downloading Saturn..."
    curl -s -L -o "$REF_DIR/planets/saturn.jpg" \
        "https://upload.wikimedia.org/wikipedia/commons/thumb/c/c7/Saturn_during_Equinox.jpg/1280px-Saturn_during_Equinox.jpg" 2>/dev/null || echo "  (skipped - network error)"

    echo "Downloading Moon..."
    curl -s -L -o "$REF_DIR/planets/moon.jpg" \
        "https://upload.wikimedia.org/wikipedia/commons/thumb/e/e1/FullMoon2010.jpg/1024px-FullMoon2010.jpg" 2>/dev/null || echo "  (skipped - network error)"

    echo ""
    echo "Planet references downloaded. Use these as visual references for pixel art planets."
    echo "License: NASA/ESA imagery - public domain. Wikimedia images may require attribution."
}

download_backgrounds() {
    echo "=== Downloading Galaxy/Nebula Background References ==="
    echo "Destination: $REF_DIR/backgrounds/"
    echo ""

    # ESA/ESO/Hubble images (CC BY 4.0 or similar)

    echo "Downloading Milky Way panorama reference..."
    # ESA Gaia all-sky view
    curl -s -L -o "$REF_DIR/backgrounds/milky_way_gaia.jpg" \
        "https://www.esa.int/var/esa/storage/images/esa_multimedia/images/2018/04/gaia_s_sky_in_colour2/17446799-3-eng-GB/Gaia_s_sky_in_colour.jpg" 2>/dev/null || echo "  (skipped - network error)"

    echo "Downloading Carina Nebula reference..."
    # JWST Carina Nebula
    curl -s -L -o "$REF_DIR/backgrounds/carina_nebula.jpg" \
        "https://stsci-opo.org/STScI-01G7JJADTH90FR98AKKJFKSS0B.png" 2>/dev/null || echo "  (skipped - network error)"

    echo "Downloading Pillars of Creation reference..."
    # JWST Pillars of Creation
    curl -s -L -o "$REF_DIR/backgrounds/pillars_of_creation.jpg" \
        "https://stsci-opo.org/STScI-01GK0RKT6T3AYSDMJNKQ4AZYQE.png" 2>/dev/null || echo "  (skipped - network error)"

    echo ""
    echo "Background references downloaded. Use these as visual references for space backgrounds."
    echo "License: ESA/ESO/STScI imagery typically CC BY 4.0 - credit required in commercial use."
}

case "$TYPE" in
    planets)
        download_planets
        ;;
    backgrounds)
        download_backgrounds
        ;;
    all)
        download_planets
        echo ""
        download_backgrounds
        ;;
    *)
        echo "ERROR: Unknown type: $TYPE"
        usage
        ;;
esac

echo ""
echo "=== Reference Download Complete ==="
echo "Reference images are in: $REF_DIR/"
echo ""
echo "These are for visual reference when generating pixel art assets."
echo "Do NOT use these directly in the game - create stylized pixel art versions instead."
