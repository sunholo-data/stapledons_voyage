#!/bin/bash
# Process downloaded star catalogs into game-ready JSON format
# Usage: process_stars.sh

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../../.." && pwd)"
RAW_DIR="$PROJECT_ROOT/assets/data/raw"
OUTPUT_DIR="$PROJECT_ROOT/assets/data/starmap"

mkdir -p "$OUTPUT_DIR"

echo "=== Star Data Processor ==="
echo "Input: $RAW_DIR"
echo "Output: $OUTPUT_DIR"
echo ""

# Check what data we have
HAS_CNS5=false
HAS_GCNS=false
HAS_EXOPLANETS=false

[ -f "$RAW_DIR/cns5.vot" ] && HAS_CNS5=true
[ -f "$RAW_DIR/gcns_full.dat" ] && HAS_GCNS=true
[ -f "$RAW_DIR/exoplanets_all.csv" ] && HAS_EXOPLANETS=true

echo "Available data:"
echo "  CNS5 (nearby stars): $HAS_CNS5"
echo "  GCNS (full catalog): $HAS_GCNS"
echo "  Exoplanets: $HAS_EXOPLANETS"
echo ""

if [ "$HAS_CNS5" = false ] && [ "$HAS_GCNS" = false ]; then
    echo "ERROR: No star data found!"
    echo "Run download_stars.sh first."
    exit 1
fi

# Check for Python (needed for coordinate conversion)
if ! command -v python3 &> /dev/null; then
    echo "WARNING: python3 not found"
    echo "Coordinate conversion requires Python + astropy"
    echo "Install with: pip3 install astropy"
    echo ""
    echo "Creating placeholder output..."

    # Create minimal placeholder
    cat > "$OUTPUT_DIR/stars.json" << 'EOF'
{
  "version": "1.0",
  "source": "placeholder",
  "count": 0,
  "stars": [],
  "note": "Run process_stars.sh with Python+astropy installed for real data"
}
EOF
    echo "Created placeholder: $OUTPUT_DIR/stars.json"
    exit 0
fi

# Create Python processing script
PROCESSOR="$SCRIPT_DIR/.star_processor.py"
cat > "$PROCESSOR" << 'PYTHON'
#!/usr/bin/env python3
"""
Convert star catalog data to game JSON format.
Handles CNS5 (VOTable) and GCNS (fixed-width) formats.
"""
import sys
import json
import os

try:
    from astropy.io.votable import parse as parse_votable
    from astropy.coordinates import ICRS, Galactic
    import astropy.units as u
    HAS_ASTROPY = True
except ImportError:
    HAS_ASTROPY = False
    print("WARNING: astropy not installed, using basic parsing")

def icrs_to_galactic_xyz(ra_deg, dec_deg, dist_pc):
    """
    Convert ICRS (RA, Dec) to Galactic Cartesian coordinates.
    Returns (x, y, z) in light-years.
    """
    import math

    # Convert to radians
    ra = math.radians(ra_deg)
    dec = math.radians(dec_deg)

    # Galactic pole position (J2000, IAU 1958)
    alpha_gp = math.radians(192.85948)  # RA of North Galactic Pole
    delta_gp = math.radians(27.12825)   # Dec of North Galactic Pole
    l_ncp = math.radians(122.93192)     # Galactic longitude of North Celestial Pole

    # Transform to galactic coordinates
    sin_b = (math.sin(delta_gp) * math.sin(dec) +
             math.cos(delta_gp) * math.cos(dec) * math.cos(ra - alpha_gp))
    b = math.asin(max(-1, min(1, sin_b)))  # Clamp for numerical stability

    cos_b = math.cos(b)
    if abs(cos_b) < 1e-10:
        l = 0
    else:
        sin_l_minus = math.cos(dec) * math.sin(ra - alpha_gp) / cos_b
        cos_l_minus = (math.cos(delta_gp) * math.sin(dec) -
                       math.sin(delta_gp) * math.cos(dec) * math.cos(ra - alpha_gp)) / cos_b
        l = math.atan2(sin_l_minus, cos_l_minus) + l_ncp

    # Convert to Cartesian (light-years)
    dist_ly = dist_pc * 3.26156
    x = dist_ly * math.cos(b) * math.cos(l)
    y = dist_ly * math.cos(b) * math.sin(l)
    z = dist_ly * math.sin(b)

    return x, y, z


def parse_cns5_votable(filepath):
    """Parse CNS5 VOTable format (works without astropy)."""
    import xml.etree.ElementTree as ET
    import math

    stars = []

    try:
        tree = ET.parse(filepath)
        root = tree.getroot()

        # Find TABLEDATA - handle namespace
        tabledata = None
        for elem in root.iter():
            if 'TABLEDATA' in elem.tag:
                tabledata = elem
                break

        if tabledata is None:
            print("ERROR: Could not find TABLEDATA in VOTable")
            return []

        row_count = 0
        for tr in tabledata:
            try:
                cells = [td.text or '' for td in tr]
                if len(cells) < 13:
                    continue

                # Based on VOTable structure:
                # 0: recno, 1: Name, 2: RAB1950, 3: DEB1950, 4: pm, 5: pmPA,
                # 6: RV, 7: Sp, 8: Vmag, 9: B-V, 10: plx, 11: _RA.icrs, 12: _DE.icrs

                name = cells[1].strip()
                ra_str = cells[11].strip()  # _RA.icrs (J2000)
                dec_str = cells[12].strip()  # _DE.icrs (J2000)
                spectral = cells[7].strip()[:1] if cells[7].strip() else 'K'
                vmag_str = cells[8].strip()
                plx_str = cells[10].strip()

                if not ra_str or not dec_str or not plx_str:
                    continue

                # Parse RA (HH MM SS.s format)
                ra_parts = ra_str.split()
                if len(ra_parts) >= 3:
                    ra_deg = (float(ra_parts[0]) + float(ra_parts[1])/60 + float(ra_parts[2])/3600) * 15
                else:
                    continue

                # Parse Dec (+DD MM SS format)
                dec_clean = dec_str.replace('−', '-')
                is_negative = '-' in dec_clean[:3]
                dec_parts = dec_clean.replace('+', '').replace('-', '').split()
                if len(dec_parts) >= 3:
                    dec_deg = float(dec_parts[0]) + float(dec_parts[1])/60 + float(dec_parts[2])/3600
                    if is_negative:
                        dec_deg = -dec_deg
                else:
                    continue

                plx = float(plx_str)
                if plx <= 0:
                    continue

                vmag = float(vmag_str) if vmag_str else 10.0
                dist_pc = 1000.0 / plx

                # Convert ICRS to Galactic Cartesian
                x_ly, y_ly, z_ly = icrs_to_galactic_xyz(ra_deg, dec_deg, dist_pc)
                dist_ly = dist_pc * 3.26156

                stars.append({
                    'id': name,
                    'name': name,
                    'x': round(x_ly, 2),
                    'y': round(y_ly, 2),
                    'z': round(z_ly, 2),
                    'dist_ly': round(dist_ly, 2),
                    'vmag': round(vmag, 2),
                    'spectral': spectral if spectral in 'OBAFGKM' else 'K',
                })
                row_count += 1

            except (ValueError, IndexError) as e:
                continue

        print(f"  Parsed {row_count} stars from VOTable")

    except ET.ParseError as e:
        print(f"ERROR: XML parse error: {e}")
        return []

    return stars

def parse_gcns_dat(filepath, max_stars=None, filter_gkm=False):
    """Parse GCNS fixed-width format."""
    stars = []

    # GCNS column positions (from ReadMe)
    # GaiaEDR3: 1-19, RAdeg: 21-35, DEdeg: 37-51, Plx: 84-91, ...
    # xcoord50: 268-278, ycoord50: 280-290, zcoord50: 292-302
    # Gmag: 93-100, BPmag: 119-126, RPmag: 136-143, WDprob: 162-167

    with open(filepath, 'r') as f:
        for line in f:
            if len(line) < 300:
                continue

            try:
                gaia_id = line[0:19].strip()
                plx = float(line[83:91].strip() or '0')
                x_pc = float(line[267:278].strip() or '0')
                y_pc = float(line[279:290].strip() or '0')
                z_pc = float(line[291:302].strip() or '0')
                gmag = float(line[92:100].strip() or '0')
                bp_mag = float(line[118:126].strip() or '0')
                rp_mag = float(line[135:143].strip() or '0')
                wd_prob = float(line[161:167].strip() or '0')

                if plx <= 0:
                    continue

                # Filter white dwarfs
                if wd_prob > 0.5:
                    continue

                bp_rp = bp_mag - rp_mag if bp_mag > 0 and rp_mag > 0 else 0
                spectral = estimate_spectral_type(bp_rp)

                # Filter to G/K/M if requested
                if filter_gkm and spectral not in ['G', 'K', 'M']:
                    continue

                # Convert pc to ly
                x_ly = x_pc * 3.26156
                y_ly = y_pc * 3.26156
                z_ly = z_pc * 3.26156
                dist_ly = (x_ly**2 + y_ly**2 + z_ly**2) ** 0.5

                stars.append({
                    'id': gaia_id,
                    'x': round(x_ly, 2),
                    'y': round(y_ly, 2),
                    'z': round(z_ly, 2),
                    'dist_ly': round(dist_ly, 2),
                    'gmag': round(gmag, 2),
                    'spectral': spectral,
                })

                if max_stars and len(stars) >= max_stars:
                    break

            except (ValueError, IndexError):
                continue

    return stars

def estimate_spectral_type(bp_rp):
    """Estimate spectral type from Gaia BP-RP color."""
    if bp_rp < -0.3: return 'O'
    if bp_rp < 0.0: return 'B'
    if bp_rp < 0.3: return 'A'
    if bp_rp < 0.5: return 'F'
    if bp_rp < 0.8: return 'G'
    if bp_rp < 1.4: return 'K'
    return 'M'

def main():
    raw_dir = sys.argv[1] if len(sys.argv) > 1 else 'assets/data/raw'
    output_dir = sys.argv[2] if len(sys.argv) > 2 else 'assets/data/starmap'
    tier = sys.argv[3] if len(sys.argv) > 3 else 'auto'

    stars = []
    source = 'unknown'

    # Process based on what's available
    cns5_path = os.path.join(raw_dir, 'cns5.vot')
    gcns_path = os.path.join(raw_dir, 'gcns_full.dat')

    if tier == 'quick' or (tier == 'auto' and os.path.exists(cns5_path)):
        if os.path.exists(cns5_path):
            print("Processing CNS5 (quick tier)...")
            stars = parse_cns5_votable(cns5_path)
            source = 'cns5'
    elif tier in ['medium', 'large'] or (tier == 'auto' and os.path.exists(gcns_path)):
        if os.path.exists(gcns_path):
            filter_gkm = (tier == 'medium')
            max_stars = 50000 if tier == 'medium' else None
            print(f"Processing GCNS ({tier} tier, filter_gkm={filter_gkm})...")
            stars = parse_gcns_dat(gcns_path, max_stars=max_stars, filter_gkm=filter_gkm)
            source = 'gcns_filtered' if filter_gkm else 'gcns_full'

    if not stars:
        print("ERROR: No stars processed!")
        sys.exit(1)

    # Sort by distance from Sol
    stars.sort(key=lambda s: s['dist_ly'])

    # Write output
    output = {
        'version': '1.0',
        'source': source,
        'count': len(stars),
        'units': {
            'x': 'light-years (toward galactic center)',
            'y': 'light-years (direction of rotation)',
            'z': 'light-years (north galactic pole)',
            'dist_ly': 'light-years from Sol',
            'gmag': 'Gaia G-band magnitude',
        },
        'stars': stars
    }

    output_path = os.path.join(output_dir, 'stars.json')
    os.makedirs(output_dir, exist_ok=True)

    with open(output_path, 'w') as f:
        json.dump(output, f, indent=2)

    print(f"Wrote {len(stars)} stars to {output_path}")

if __name__ == '__main__':
    main()
PYTHON

chmod +x "$PROCESSOR"

# Determine tier based on available data
TIER="auto"
if [ "$HAS_GCNS" = true ]; then
    # Check if we should filter (medium) or use full (large)
    FILE_SIZE=$(du -m "$RAW_DIR/gcns_full.dat" | cut -f1)
    if [ "$FILE_SIZE" -gt 100 ]; then
        echo "Large dataset detected (${FILE_SIZE}MB)"
        TIER="large"
    else
        TIER="medium"
    fi
elif [ "$HAS_CNS5" = true ]; then
    TIER="quick"
fi

echo "Processing tier: $TIER"
echo ""

# Run processor
python3 "$PROCESSOR" "$RAW_DIR" "$OUTPUT_DIR" "$TIER"

# Process exoplanets if available
if [ "$HAS_EXOPLANETS" = true ]; then
    echo ""
    echo "Processing exoplanets..."

    # Simple CSV to JSON conversion
    python3 << PYTHON
import csv
import json
import os

raw_dir = "$RAW_DIR"
output_dir = "$OUTPUT_DIR"

planets = []
with open(os.path.join(raw_dir, 'exoplanets_all.csv'), 'r') as f:
    reader = csv.DictReader(f)
    for row in reader:
        try:
            planet = {
                'name': row.get('pl_name', ''),
                'host': row.get('hostname', ''),
                'ra': float(row.get('ra', 0) or 0),
                'dec': float(row.get('dec', 0) or 0),
                'dist_pc': float(row.get('sy_dist', 0) or 0),
                'period_days': float(row.get('pl_orbper', 0) or 0),
                'radius_earth': float(row.get('pl_rade', 0) or 0),
                'mass_earth': float(row.get('pl_masse', 0) or 0),
                'eq_temp_k': float(row.get('pl_eqt', 0) or 0),
                'insolation': float(row.get('pl_insol', 0) or 0),
                'host_spectype': row.get('st_spectype', ''),
            }
            if planet['name']:
                planets.append(planet)
        except (ValueError, TypeError):
            continue

output = {
    'version': '1.0',
    'source': 'nasa_exoplanet_archive',
    'count': len(planets),
    'planets': planets
}

with open(os.path.join(output_dir, 'exoplanets.json'), 'w') as f:
    json.dump(output, f, indent=2)

print(f"Wrote {len(planets)} exoplanets to exoplanets.json")

# Also create habitable zone subset
hz_planets = [p for p in planets if 0.25 <= p['insolation'] <= 2.0 and 0 < p['radius_earth'] < 2.0]
hz_output = {
    'version': '1.0',
    'source': 'nasa_exoplanet_archive_hz_filtered',
    'count': len(hz_planets),
    'planets': hz_planets
}

with open(os.path.join(output_dir, 'habitable.json'), 'w') as f:
    json.dump(hz_output, f, indent=2)

print(f"Wrote {len(hz_planets)} habitable zone candidates to habitable.json")
PYTHON
fi

# Cleanup temporary processor
rm -f "$PROCESSOR"

echo ""
echo "=== Processing Complete ==="
echo ""
echo "Output files:"
ls -lh "$OUTPUT_DIR"/*.json 2>/dev/null || echo "  (no JSON files)"
