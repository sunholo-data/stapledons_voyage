# Starmap Data Sources

Complete documentation of astronomical data sources used in Stapledon's Voyage.

## Star Catalogs

### 1. CNS5 - Catalogue of Nearby Stars (Quick Tier)

**Source**: Gaia Sky / German Virtual Observatory
**URL**: https://gaiasky.space/resources/datasets/
**Size**: 1.2 MB
**Stars**: 5,930 nearest stars

The CNS5 (Fifth Catalogue of Nearby Stars) contains the closest stars to Sol, all within ~25 parsecs (~82 light-years). Perfect for rapid prototyping.

**Columns**:
- Source ID, RA, Dec, Parallax
- Proper motion (RA/Dec)
- G/BP/RP magnitudes
- Radial velocity (if available)

**Download**:
```bash
# Direct from Gaia Sky datasets
curl -L "https://gaia.ari.uni-heidelberg.de/gaiasky/files/catalogs/dr3/cns5-dr3.json.gz" \
  -o assets/data/raw/cns5.json.gz
```

### 2. GCNS - Gaia Catalogue of Nearby Stars (Medium/Large Tiers)

**Source**: VizieR / CDS Strasbourg
**URL**: https://cdsarc.cds.unistra.fr/ftp/J/A+A/649/A6/
**Paper**: Gaia Collaboration 2021, A&A 649, A6
**License**: CC BY-SA 3.0

| File | Size | Contents |
|------|------|----------|
| table1c.csv | 164 MB | Full catalog (331,312 stars) |
| table1c.dat.gz | 72 MB | Compressed format |

**Columns** (table1c):
- `GaiaEDR3`: Gaia EDR3 source ID
- `RAdeg`, `DEdeg`: Position (ICRS, epoch 2016.0)
- `Plx`, `e_Plx`: Parallax and error (mas)
- `pmRA`, `pmDE`: Proper motion (mas/yr)
- `Gmag`, `BPmag`, `RPmag`: Gaia photometry
- `Dist50`: Median distance (pc)
- `xcoord50`, `ycoord50`, `zcoord50`: Galactic XYZ (pc)
- `WDprob`: White dwarf probability

**Download**:
```bash
# Large tier: Full catalog
curl "https://cdsarc.cds.unistra.fr/ftp/J/A+A/649/A6/table1c.dat.gz" \
  -o assets/data/raw/gcns_full.dat.gz

# With CSV (uncompressed)
curl "https://cdsarc.cds.unistra.fr/ftp/J/A+A/649/A6/table1c.csv" \
  -o assets/data/raw/gcns_full.csv
```

**Medium Tier Filtering**:
```python
# Filter to ~50k G/K/M dwarfs with good parallax
criteria = {
    "parallax_error_ratio": "< 0.1",  # Plx/e_Plx > 10
    "white_dwarf_prob": "< 0.5",      # Not a white dwarf
    "bp_rp_color": "> 0.5",           # G/K/M types (redder)
    "distance": "< 100 pc",           # Within local bubble
}
```

### 3. Hipparcos (Alternative)

**Source**: ESA / Gaia Sky
**URL**: https://gaiasky.space/resources/datasets/
**Size**: 7.7 MB
**Stars**: 117,950 bright stars

Good coverage of bright stars visible from Earth. Less accurate than Gaia but historically significant.

## Exoplanet Data

### NASA Exoplanet Archive

**Source**: NASA/IPAC/Caltech
**URL**: https://exoplanetarchive.ipac.caltech.edu/
**API**: TAP (Table Access Protocol)
**License**: Public Domain

**Current Count**: ~6,042 confirmed exoplanets (as of Oct 2025)

**Key Tables**:
- `pscomppars`: Planetary Systems Composite Parameters (recommended)
- `ps`: Planetary Systems (all published parameters)

**Download via TAP**:
```bash
# All confirmed planets with composite parameters
curl "https://exoplanetarchive.ipac.caltech.edu/TAP/sync?query=select+*+from+pscomppars&format=csv" \
  -o assets/data/raw/exoplanets.csv
```

**Key Columns**:
- `pl_name`: Planet name
- `hostname`: Host star name
- `ra`, `dec`: Host star position
- `sy_dist`: System distance (pc)
- `pl_orbper`: Orbital period (days)
- `pl_rade`: Planet radius (Earth radii)
- `pl_masse`: Planet mass (Earth masses)
- `pl_eqt`: Equilibrium temperature (K)
- `pl_insol`: Insolation flux (Earth flux)
- `st_spectype`: Host star spectral type
- `st_teff`: Host star temperature (K)

**Habitable Zone Query**:
```sql
SELECT * FROM pscomppars
WHERE pl_insol BETWEEN 0.25 AND 2.0  -- Conservative HZ
  AND pl_rade < 2.0                   -- Rocky candidates
ORDER BY sy_dist ASC
```

## Galactic Background Images

### ESA Gaia All-Sky Map

**Source**: ESA/Gaia/DPAC
**License**: CC BY-SA 3.0 IGO
**Credit Required**: "ESA/Gaia/DPAC"

**Available Resolutions**:

| Resolution | Dimensions | Size | URL |
|------------|------------|------|-----|
| 2K | 2048x1024 | ~2 MB | See below |
| 4K | 4096x2048 | ~5 MB | See below |
| 8K | 8192x4096 | ~20 MB | See below |
| Full | 40000x20000 | ~800 MB | Archive only |

**Download URLs**:
```bash
# 4K equirectangular (recommended for game)
curl "https://sci.esa.int/documents/33565/0/Gaia_EDR3_flux_equirect_4096x2048.png" \
  -o assets/data/starmap/background/galaxy_4k.png

# 8K for HD mode
curl "https://sci.esa.int/documents/33565/0/Gaia_EDR3_flux_equirect_8192x4096.png" \
  -o assets/data/starmap/background/galaxy_8k.png
```

**Projections Available**:
- Equirectangular (best for skybox/dome)
- Hammer/Aitoff (classic astronomical)
- Mollweide (equal-area)

### ESO Milky Way Panorama (Alternative)

**Source**: ESO/S. Brunier
**URL**: https://www.eso.org/public/images/eso0932a/
**License**: CC BY 4.0
**Credit**: "ESO/S. Brunier"

High-quality photographic panorama, 800 megapixels original. Good alternative with warmer colors.

### Parallax Background Layers

For multi-layer parallax effect, generate procedural layers:

1. **Deep field** (distant galaxies): Procedural noise + galaxy sprites
2. **Dust lanes**: Semi-transparent dark nebulae
3. **Star field**: Procedural points based on galactic density model
4. **Gaia map**: Real star data as brightest layer

## Coordinate Systems

### Input Coordinates (ICRS)
- RA (Right Ascension): 0-360 degrees
- Dec (Declination): -90 to +90 degrees
- Parallax: milliarcseconds (distance = 1000/parallax parsecs)

### Game Coordinates (Galactic Cartesian)
- X: Toward galactic center (positive)
- Y: Direction of galactic rotation (positive)
- Z: North galactic pole (positive)
- Units: Light-years (1 pc = 3.26156 ly)

### Conversion
```python
import astropy.coordinates as coord
import astropy.units as u

# From ICRS to Galactic Cartesian
icrs = coord.ICRS(ra=ra*u.deg, dec=dec*u.deg, distance=(1000/parallax)*u.pc)
galactic = icrs.transform_to(coord.Galactic())
cartesian = galactic.cartesian

x_ly = cartesian.x.to(u.lyr).value
y_ly = cartesian.y.to(u.lyr).value
z_ly = cartesian.z.to(u.lyr).value
```

## Spectral Type Classification

| Type | Color | Temp (K) | % of Stars | Habitable Potential |
|------|-------|----------|------------|---------------------|
| O | Blue | 30,000+ | 0.00003% | Very low (short-lived) |
| B | Blue-white | 10,000-30,000 | 0.13% | Low (UV intense) |
| A | White | 7,500-10,000 | 0.6% | Low (short-lived) |
| F | Yellow-white | 6,000-7,500 | 3% | Medium |
| G | Yellow | 5,200-6,000 | 7.5% | High (Sun-like) |
| K | Orange | 3,700-5,200 | 12% | High (long-lived) |
| M | Red | 2,400-3,700 | 76% | Medium (tidal locking) |

**Game relevance**: G and K stars are prime targets for civilizations.

## Data Update Schedule

| Data | Update Frequency | Notes |
|------|------------------|-------|
| Star positions | Never | Stars don't move perceptibly |
| Proper motions | Never | Only matters over millennia |
| Exoplanets | Quarterly | New discoveries ~monthly |
| Background images | Never | Static imagery |

## Attribution Requirements

Include in game credits:
```
Star data: ESA/Gaia/DPAC (CC BY-SA 3.0 IGO)
Exoplanet data: NASA Exoplanet Archive
Galaxy imagery: ESA/Gaia/DPAC (CC BY-SA 3.0 IGO)
```
