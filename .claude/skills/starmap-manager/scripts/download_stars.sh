#!/bin/bash
# Download star catalog data for Stapledon's Voyage
# Usage: download_stars.sh <tier>
#   tier: quick (CNS5, ~1MB), medium (filtered GCNS, ~15MB), large (full GCNS, ~72MB)

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../../.." && pwd)"
RAW_DIR="$PROJECT_ROOT/assets/data/raw"
OUTPUT_DIR="$PROJECT_ROOT/assets/data/starmap"

TIER="${1:-quick}"

# Create directories
mkdir -p "$RAW_DIR"
mkdir -p "$OUTPUT_DIR"

echo "=== Starmap Data Downloader ==="
echo "Tier: $TIER"
echo "Output: $RAW_DIR"
echo ""

case "$TIER" in
    quick)
        echo "Downloading CNS5 (Catalogue of Nearby Stars)..."
        echo "  Source: Gaia Sky / German Virtual Observatory"
        echo "  Stars: ~5,930 nearest stars"
        echo "  Size: ~1.2 MB"
        echo ""

        # CNS5 from Gaia Sky - using their hosted version
        # Note: This is a pre-processed nearby star catalog
        CNS5_URL="https://gaia.ari.uni-heidelberg.de/gaiasky/files/catalogs/dr3/cns5-dr3.vot.gz"

        if curl -L --fail -o "$RAW_DIR/cns5.vot.gz" "$CNS5_URL" 2>/dev/null; then
            echo "  Downloaded: cns5.vot.gz"
            gunzip -f "$RAW_DIR/cns5.vot.gz" 2>/dev/null || true
        else
            echo "  Primary URL failed, trying alternative..."
            # Alternative: Query VizieR for CNS5
            # This gets the 5th Catalogue of Nearby Stars
            ALT_URL="https://vizier.cds.unistra.fr/viz-bin/votable?-source=V/70A&-out.max=10000"
            curl -L -o "$RAW_DIR/cns5.vot" "$ALT_URL" 2>/dev/null || {
                echo "  ERROR: Could not download CNS5 data"
                echo "  Try manually from: https://gaiasky.space/resources/datasets/"
                exit 1
            }
        fi

        echo ""
        echo "Quick tier complete!"
        ;;

    medium)
        echo "Downloading filtered GCNS (Gaia Catalogue of Nearby Stars)..."
        echo "  Source: VizieR / CDS Strasbourg"
        echo "  Stars: ~50,000 G/K/M dwarfs within 100pc"
        echo "  Size: ~10-15 MB"
        echo ""

        # GCNS from CDS - using TAP query for filtered subset
        # Filter: G/K/M stars (BP-RP > 0.5), good parallax, not white dwarfs
        GCNS_URL="https://cdsarc.cds.unistra.fr/ftp/J/A+A/649/A6/table1c.dat.gz"

        echo "  Downloading full GCNS (will filter locally)..."
        if curl -L --fail -o "$RAW_DIR/gcns_full.dat.gz" "$GCNS_URL"; then
            echo "  Downloaded: gcns_full.dat.gz ($(du -h "$RAW_DIR/gcns_full.dat.gz" | cut -f1))"
            gunzip -kf "$RAW_DIR/gcns_full.dat.gz"
            echo "  Extracted: gcns_full.dat"
        else
            echo "  ERROR: Could not download GCNS data"
            echo "  Try manually from: https://cdsarc.cds.unistra.fr/viz-bin/cat/J/A+A/649/A6"
            exit 1
        fi

        echo ""
        echo "Medium tier complete!"
        echo "Run process_stars.sh to filter to ~50k G/K/M dwarfs"
        ;;

    large)
        echo "Downloading full GCNS (Gaia Catalogue of Nearby Stars)..."
        echo "  Source: VizieR / CDS Strasbourg"
        echo "  Stars: 331,312 within 100 parsecs"
        echo "  Size: ~72 MB compressed, ~164 MB uncompressed"
        echo ""

        GCNS_URL="https://cdsarc.cds.unistra.fr/ftp/J/A+A/649/A6/table1c.dat.gz"

        if curl -L --fail -o "$RAW_DIR/gcns_full.dat.gz" "$GCNS_URL"; then
            echo "  Downloaded: gcns_full.dat.gz ($(du -h "$RAW_DIR/gcns_full.dat.gz" | cut -f1))"
            gunzip -kf "$RAW_DIR/gcns_full.dat.gz"
            echo "  Extracted: gcns_full.dat ($(du -h "$RAW_DIR/gcns_full.dat" | cut -f1))"
        else
            echo "  ERROR: Could not download GCNS data"
            exit 1
        fi

        # Also download the ReadMe for column definitions
        curl -L -o "$RAW_DIR/gcns_readme.txt" \
            "https://cdsarc.cds.unistra.fr/ftp/J/A+A/649/A6/ReadMe" 2>/dev/null || true

        echo ""
        echo "Large tier complete!"
        ;;

    *)
        echo "ERROR: Unknown tier '$TIER'"
        echo "Usage: $0 <quick|medium|large>"
        echo ""
        echo "Tiers:"
        echo "  quick  - CNS5, ~5,930 stars, ~1 MB"
        echo "  medium - Filtered GCNS, ~50,000 stars, ~15 MB"
        echo "  large  - Full GCNS, 331,312 stars, ~72 MB"
        exit 1
        ;;
esac

echo ""
echo "Files in $RAW_DIR:"
ls -lh "$RAW_DIR"/ 2>/dev/null || echo "  (empty)"
