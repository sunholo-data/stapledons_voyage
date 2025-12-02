#!/bin/bash
# Download exoplanet data from NASA Exoplanet Archive
# Usage: download_exoplanets.sh

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../../.." && pwd)"
RAW_DIR="$PROJECT_ROOT/assets/data/raw"
OUTPUT_DIR="$PROJECT_ROOT/assets/data/starmap"

mkdir -p "$RAW_DIR"
mkdir -p "$OUTPUT_DIR"

echo "=== NASA Exoplanet Archive Downloader ==="
echo "Source: NASA/IPAC/Caltech"
echo "URL: https://exoplanetarchive.ipac.caltech.edu/"
echo ""

# Download all confirmed planets with composite parameters
# This is the recommended table for most use cases
echo "Downloading confirmed exoplanets (pscomppars table)..."
echo "  Includes: ~6,000 confirmed planets"
echo "  Size: ~3-5 MB"
echo ""

EXOPLANET_URL="https://exoplanetarchive.ipac.caltech.edu/TAP/sync?query=select+*+from+pscomppars&format=csv"

if curl -L --fail -o "$RAW_DIR/exoplanets_all.csv" "$EXOPLANET_URL"; then
    LINES=$(wc -l < "$RAW_DIR/exoplanets_all.csv")
    SIZE=$(du -h "$RAW_DIR/exoplanets_all.csv" | cut -f1)
    echo "  Downloaded: exoplanets_all.csv"
    echo "  Size: $SIZE"
    echo "  Planets: $((LINES - 1))"  # Subtract header
else
    echo "  ERROR: Could not download exoplanet data"
    echo "  Try manually from: https://exoplanetarchive.ipac.caltech.edu/"
    exit 1
fi

echo ""

# Also download habitable zone candidates separately
echo "Downloading habitable zone candidates..."
# Conservative HZ: insolation 0.25-2.0 Earth flux, radius < 2 Earth radii
HZ_QUERY="select+pl_name,hostname,ra,dec,sy_dist,pl_orbper,pl_rade,pl_masse,pl_eqt,pl_insol,st_spectype,st_teff+from+pscomppars+where+pl_insol+between+0.25+and+2.0+and+pl_rade+<+2.0"
HZ_URL="https://exoplanetarchive.ipac.caltech.edu/TAP/sync?query=$HZ_QUERY&format=csv"

if curl -L --fail -o "$RAW_DIR/exoplanets_hz.csv" "$HZ_URL" 2>/dev/null; then
    LINES=$(wc -l < "$RAW_DIR/exoplanets_hz.csv")
    echo "  Downloaded: exoplanets_hz.csv"
    echo "  Habitable zone candidates: $((LINES - 1))"
else
    echo "  Warning: Could not download HZ subset (using full dataset)"
fi

echo ""
echo "Exoplanet download complete!"
echo ""
echo "Files in $RAW_DIR:"
ls -lh "$RAW_DIR"/exoplanet*.csv 2>/dev/null || echo "  (no exoplanet files)"
