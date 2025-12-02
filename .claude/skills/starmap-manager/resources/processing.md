# Star Data Processing Pipeline

Technical documentation for converting astronomical data to game-ready format.

## Coordinate Conversion

### Input: ICRS (International Celestial Reference System)

Standard astronomical coordinates:
- **RA** (Right Ascension): 0-360 degrees, measured eastward from vernal equinox
- **Dec** (Declination): -90 to +90 degrees, from celestial equator
- **Parallax**: Angular shift in milliarcseconds (mas), distance = 1000/parallax parsecs

### Output: Galactic Cartesian

Game uses galactic coordinates centered on Sol:
- **X**: Light-years toward galactic center (positive = toward center)
- **Y**: Light-years in direction of galactic rotation (positive = rotation direction)
- **Z**: Light-years toward north galactic pole (positive = north)

### Conversion Formula

Using astropy (Python):

```python
from astropy.coordinates import ICRS, Galactic
import astropy.units as u

# Input: RA, Dec in degrees, parallax in mas
ra_deg = 280.0
dec_deg = -45.0
parallax_mas = 100.0  # 10 parsecs

# Calculate distance
distance_pc = 1000.0 / parallax_mas

# Create ICRS coordinate
icrs = ICRS(ra=ra_deg*u.deg, dec=dec_deg*u.deg, distance=distance_pc*u.pc)

# Transform to Galactic
galactic = icrs.transform_to(Galactic())
cartesian = galactic.cartesian

# Extract as light-years
x_ly = cartesian.x.to(u.lyr).value
y_ly = cartesian.y.to(u.lyr).value
z_ly = cartesian.z.to(u.lyr).value
```

### Pure Go Conversion (for engine)

If needed without astropy:

```go
// Approximate conversion (good to ~1%)
func ICRSToGalactic(raDeg, decDeg, distPc float64) (x, y, z float64) {
    // Convert to radians
    ra := raDeg * math.Pi / 180
    dec := decDeg * math.Pi / 180

    // Galactic pole (J2000)
    const (
        alphaGP = 192.85948 * math.Pi / 180  // RA of galactic pole
        deltaGP = 27.12825 * math.Pi / 180   // Dec of galactic pole
        lNCP    = 122.93192 * math.Pi / 180  // Galactic longitude of NCP
    )

    // Convert to Galactic
    sinB := math.Sin(deltaGP)*math.Sin(dec) +
            math.Cos(deltaGP)*math.Cos(dec)*math.Cos(ra-alphaGP)
    b := math.Asin(sinB)

    cosB := math.Cos(b)
    sinLmLNCP := math.Cos(dec) * math.Sin(ra-alphaGP) / cosB
    cosLmLNCP := (math.Cos(deltaGP)*math.Sin(dec) -
                  math.Sin(deltaGP)*math.Cos(dec)*math.Cos(ra-alphaGP)) / cosB
    l := math.Atan2(sinLmLNCP, cosLmLNCP) + lNCP

    // Convert to Cartesian (light-years)
    distLy := distPc * 3.26156
    x = distLy * math.Cos(b) * math.Cos(l)
    y = distLy * math.Cos(b) * math.Sin(l)
    z = distLy * math.Sin(b)

    return x, y, z
}
```

## Filtering Criteria

### Quick Tier (CNS5)
- All stars in CNS5 catalog (~6k stars)
- No additional filtering needed
- Stars within ~25 pc (~82 ly)

### Medium Tier (Filtered GCNS)

```python
filters = {
    # Good parallax measurement
    'parallax_over_error': lambda row: row['Plx'] / row['e_Plx'] > 10,

    # Not a white dwarf
    'not_white_dwarf': lambda row: row['WDprob'] < 0.5,

    # G/K/M stars (BP-RP color > 0.5)
    'gkm_stars': lambda row: (row['BPmag'] - row['RPmag']) > 0.5,

    # Within 100 parsecs
    'within_100pc': lambda row: 1000/row['Plx'] < 100,
}
```

Expected result: ~50,000 stars

### Large Tier (Full GCNS)
- All stars with valid parallax
- Remove white dwarfs only (WDprob > 0.5)
- Full 331k catalog

## Spectral Type Estimation

From Gaia BP-RP color index:

| BP-RP Range | Spectral Type | Star Color |
|-------------|---------------|------------|
| < -0.3 | O | Blue |
| -0.3 to 0.0 | B | Blue-white |
| 0.0 to 0.3 | A | White |
| 0.3 to 0.5 | F | Yellow-white |
| 0.5 to 0.8 | G | Yellow (Sun-like) |
| 0.8 to 1.4 | K | Orange |
| > 1.4 | M | Red |

```python
def estimate_spectral_type(bp_rp):
    if bp_rp < -0.3: return 'O'
    if bp_rp < 0.0: return 'B'
    if bp_rp < 0.3: return 'A'
    if bp_rp < 0.5: return 'F'
    if bp_rp < 0.8: return 'G'
    if bp_rp < 1.4: return 'K'
    return 'M'
```

## Output JSON Schema

### stars.json

```json
{
  "version": "1.0",
  "source": "gcns_filtered",
  "count": 50000,
  "units": {
    "x": "light-years (toward galactic center)",
    "y": "light-years (direction of rotation)",
    "z": "light-years (north galactic pole)",
    "dist_ly": "light-years from Sol",
    "gmag": "Gaia G-band magnitude"
  },
  "stars": [
    {
      "id": "4472832130942575872",
      "x": 4.37,
      "y": 0.12,
      "z": 1.23,
      "dist_ly": 4.37,
      "gmag": -0.01,
      "spectral": "G"
    }
  ]
}
```

### exoplanets.json

```json
{
  "version": "1.0",
  "source": "nasa_exoplanet_archive",
  "count": 6000,
  "planets": [
    {
      "name": "Proxima Centauri b",
      "host": "Proxima Centauri",
      "ra": 217.43,
      "dec": -62.68,
      "dist_pc": 1.30,
      "period_days": 11.19,
      "radius_earth": 1.08,
      "mass_earth": 1.07,
      "eq_temp_k": 234,
      "insolation": 0.65,
      "host_spectype": "M5.5V"
    }
  ]
}
```

### habitable.json

Same schema as exoplanets.json, filtered to:
- Insolation: 0.25 - 2.0 Earth flux (conservative habitable zone)
- Radius: < 2.0 Earth radii (likely rocky)

## Habitable Zone Calculation

The habitable zone depends on stellar properties:

```python
def habitable_zone_bounds(star_luminosity_solar, star_temp_k):
    """
    Calculate conservative habitable zone inner/outer bounds.
    Returns distances in AU.
    """
    # Kopparapu et al. (2013) coefficients
    # Conservative limits: runaway greenhouse to maximum greenhouse

    # Inner edge (runaway greenhouse)
    S_in = 1.0140 + (8.1774e-5)*star_temp_k + (1.7063e-9)*star_temp_k**2
    d_in = (star_luminosity_solar / S_in) ** 0.5

    # Outer edge (maximum greenhouse)
    S_out = 0.3438 + (5.8942e-5)*star_temp_k + (1.6558e-9)*star_temp_k**2
    d_out = (star_luminosity_solar / S_out) ** 0.5

    return d_in, d_out
```

For game purposes, we use insolation flux directly:
- **0.25 < insolation < 2.0** = conservative habitable zone
- **0.5 < insolation < 1.5** = optimistic habitable zone

## Game Integration

### Loading in Go

```go
type StarCatalog struct {
    Version string `json:"version"`
    Source  string `json:"source"`
    Count   int    `json:"count"`
    Stars   []Star `json:"stars"`
}

type Star struct {
    ID       string  `json:"id"`
    X        float64 `json:"x"`
    Y        float64 `json:"y"`
    Z        float64 `json:"z"`
    DistLY   float64 `json:"dist_ly"`
    GMag     float64 `json:"gmag"`
    Spectral string  `json:"spectral"`
}

func LoadStarCatalog(path string) (*StarCatalog, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }
    var catalog StarCatalog
    if err := json.Unmarshal(data, &catalog); err != nil {
        return nil, err
    }
    return &catalog, nil
}
```

### Distance Queries

For "stars within N light-years", use simple distance check:

```go
func (c *StarCatalog) WithinRadius(cx, cy, cz, radius float64) []Star {
    var result []Star
    r2 := radius * radius
    for _, s := range c.Stars {
        dx := s.X - cx
        dy := s.Y - cy
        dz := s.Z - cz
        if dx*dx + dy*dy + dz*dz <= r2 {
            result = append(result, s)
        }
    }
    return result
}
```

For large catalogs, consider octree indexing (see starmap-data-model.md).

## Performance Considerations

| Catalog Size | Load Time | Memory | Query (brute) |
|--------------|-----------|--------|---------------|
| Quick (6k) | ~10ms | ~1 MB | <1ms |
| Medium (50k) | ~50ms | ~8 MB | ~5ms |
| Large (331k) | ~300ms | ~50 MB | ~30ms |

Recommendations:
- Quick/Medium: Brute-force queries are fast enough
- Large: Consider octree for repeated spatial queries
- All tiers: Load on game start, keep in memory
