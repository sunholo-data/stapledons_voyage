// Package stardata handles importing and processing star catalog data from Gaia DR3.
package stardata

import (
	"encoding/csv"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
)

// GaiaStar represents a star from the Gaia catalog with raw astrometric data.
type GaiaStar struct {
	SourceID     string  // Gaia DR3 source ID
	Name         string  // Common name if known
	RA           float64 // Right ascension (degrees)
	Dec          float64 // Declination (degrees)
	Parallax     float64 // Parallax (milliarcseconds)
	ParallaxErr  float64 // Parallax error (milliarcseconds)
	GMag         float64 // G-band magnitude
	BPMinusRP    float64 // BP-RP color index (for spectral type estimation)
	RadialVel    float64 // Radial velocity (km/s), 0 if unknown
	HasRV        bool    // Whether radial velocity is known
}

// Star represents a processed star in galactocentric cartesian coordinates.
type Star struct {
	ID           int64   // Unique star ID
	Name         string  // Star name
	X            float64 // X position (light-years, Sol-centered)
	Y            float64 // Y position (light-years)
	Z            float64 // Z position (light-years)
	SpectralType string  // Spectral type (O, B, A, F, G, K, M)
	Luminosity   float64 // Luminosity (solar luminosities)
	HasHZPlanet  bool    // Whether star likely has habitable zone planet
}

// Constants for coordinate conversion
const (
	// Parsecs to light-years conversion
	ParsecToLY = 3.26156

	// Sun's position in galactic coordinates (not used for Sol-centered)
	// We use Sol-centered coordinates, so Sol is at origin

	// Minimum parallax to accept (avoid distant stars with huge errors)
	MinParallax = 1.0 // milliarcseconds = 1000 parsecs max distance

	// Maximum parallax error ratio
	MaxParallaxErrorRatio = 0.2 // 20% error max
)

// ToGalactocentric converts equatorial coordinates (RA, Dec, parallax) to
// Sol-centered galactocentric cartesian coordinates in light-years.
//
// Input:
//   - ra: Right ascension in degrees (0-360)
//   - dec: Declination in degrees (-90 to +90)
//   - parallax: Parallax in milliarcseconds
//
// Output:
//   - x, y, z: Position in light-years relative to Sol
//
// The coordinate system:
//   - X: Toward galactic center (roughly toward RA=266°, Dec=-29°)
//   - Y: Direction of galactic rotation
//   - Z: North galactic pole
//
// For simplicity, we use a Sol-centered equatorial-aligned system:
//   - X: Toward vernal equinox (RA=0)
//   - Y: Toward RA=90°
//   - Z: Toward celestial north pole (Dec=+90°)
func ToGalactocentric(ra, dec, parallax float64) (x, y, z float64) {
	// Distance in parsecs = 1000 / parallax(mas)
	distPC := 1000.0 / parallax

	// Convert to light-years
	distLY := distPC * ParsecToLY

	// Convert RA/Dec to radians
	raRad := ra * math.Pi / 180.0
	decRad := dec * math.Pi / 180.0

	// Spherical to cartesian
	// x = d * cos(dec) * cos(ra)
	// y = d * cos(dec) * sin(ra)
	// z = d * sin(dec)
	x = distLY * math.Cos(decRad) * math.Cos(raRad)
	y = distLY * math.Cos(decRad) * math.Sin(raRad)
	z = distLY * math.Sin(decRad)

	return x, y, z
}

// EstimateSpectralType estimates spectral type from BP-RP color index.
// BP-RP ranges roughly:
//   - O/B: < 0.0 (blue)
//   - A: 0.0 - 0.3
//   - F: 0.3 - 0.6
//   - G: 0.6 - 0.9 (Sun is ~0.82)
//   - K: 0.9 - 1.4
//   - M: > 1.4 (red)
func EstimateSpectralType(bpMinusRP float64) string {
	switch {
	case bpMinusRP < 0.0:
		return "B" // Could be O but extremely rare
	case bpMinusRP < 0.3:
		return "A"
	case bpMinusRP < 0.6:
		return "F"
	case bpMinusRP < 0.9:
		return "G"
	case bpMinusRP < 1.4:
		return "K"
	default:
		return "M"
	}
}

// EstimateLuminosity estimates luminosity in solar luminosities from absolute magnitude.
// L/L☉ = 10^((M☉ - M) / 2.5) where M☉ ≈ 4.83
func EstimateLuminosity(absoluteMag float64) float64 {
	const sunAbsMag = 4.83
	return math.Pow(10.0, (sunAbsMag-absoluteMag)/2.5)
}

// AbsoluteMagnitude calculates absolute magnitude from apparent magnitude and parallax.
// M = m - 5 * log10(d/10) where d is in parsecs
// M = m + 5 + 5 * log10(parallax/1000)
func AbsoluteMagnitude(apparentMag, parallax float64) float64 {
	distPC := 1000.0 / parallax
	return apparentMag - 5.0*math.Log10(distPC/10.0)
}

// EstimateHZPlanet estimates probability of habitable zone planet based on spectral type.
// Based on eta-Earth estimates from Kepler mission data.
func EstimateHZPlanet(spectralType string) bool {
	// Rough eta-Earth by spectral type (very approximate)
	// These are probabilities, we'd need a seed for deterministic results
	// For now, we use conservative estimates for "known" HZ planets
	switch spectralType {
	case "G":
		return true // Sun-like stars most likely to have HZ planets in our data
	case "K":
		return true // K dwarfs also good candidates
	case "M":
		return false // M dwarfs have HZ issues (flares, tidal locking)
	default:
		return false // O, B, A, F less likely for various reasons
	}
}

// LoadGaiaCatalog loads stars from a Gaia CSV file.
// Expected CSV format (flexible, reads header):
//   source_id,ra,dec,parallax,parallax_error,phot_g_mean_mag,bp_rp,[name],[radial_velocity]
//
// Returns processed stars in galactocentric coordinates.
func LoadGaiaCatalog(filename string) ([]Star, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open Gaia catalog: %w", err)
	}
	defer file.Close()

	return ParseGaiaCSV(file)
}

// ParseGaiaCSV parses Gaia data from a CSV reader.
func ParseGaiaCSV(r io.Reader) ([]Star, error) {
	reader := csv.NewReader(r)
	reader.TrimLeadingSpace = true
	reader.LazyQuotes = true

	// Read header
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV header: %w", err)
	}

	// Build column index map
	colIdx := make(map[string]int)
	for i, col := range header {
		colIdx[strings.ToLower(strings.TrimSpace(col))] = i
	}

	// Required columns
	requiredCols := []string{"source_id", "ra", "dec", "parallax"}
	for _, col := range requiredCols {
		if _, ok := colIdx[col]; !ok {
			return nil, fmt.Errorf("missing required column: %s", col)
		}
	}

	var stars []Star
	var id int64 = 1 // Start IDs at 1, reserve 0 for Sol

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			// Skip malformed rows
			continue
		}

		// Parse required fields
		ra, err := parseFloat(record, colIdx, "ra")
		if err != nil {
			continue
		}
		dec, err := parseFloat(record, colIdx, "dec")
		if err != nil {
			continue
		}
		parallax, err := parseFloat(record, colIdx, "parallax")
		if err != nil || parallax <= 0 {
			continue
		}

		// Filter by parallax quality
		if parallax < MinParallax {
			continue // Too distant
		}

		parallaxErr, _ := parseFloat(record, colIdx, "parallax_error")
		if parallaxErr > 0 && parallaxErr/parallax > MaxParallaxErrorRatio {
			continue // Too much error
		}

		// Convert to galactocentric coordinates
		x, y, z := ToGalactocentric(ra, dec, parallax)

		// Get magnitude and color for spectral estimation
		gMag, _ := parseFloat(record, colIdx, "phot_g_mean_mag")
		bpRP, _ := parseFloat(record, colIdx, "bp_rp")

		spectralType := EstimateSpectralType(bpRP)
		absMag := AbsoluteMagnitude(gMag, parallax)
		luminosity := EstimateLuminosity(absMag)
		hasHZ := EstimateHZPlanet(spectralType)

		// Get name if available
		name := ""
		if nameIdx, ok := colIdx["name"]; ok && nameIdx < len(record) {
			name = strings.TrimSpace(record[nameIdx])
		}
		if name == "" {
			// Use source ID as fallback name
			if srcIdx, ok := colIdx["source_id"]; ok && srcIdx < len(record) {
				name = "Gaia " + strings.TrimSpace(record[srcIdx])
			}
		}

		stars = append(stars, Star{
			ID:           id,
			Name:         name,
			X:            x,
			Y:            y,
			Z:            z,
			SpectralType: spectralType,
			Luminosity:   luminosity,
			HasHZPlanet:  hasHZ,
		})
		id++
	}

	return stars, nil
}

// parseFloat safely parses a float from a CSV record using column index map.
func parseFloat(record []string, colIdx map[string]int, colName string) (float64, error) {
	idx, ok := colIdx[colName]
	if !ok || idx >= len(record) {
		return 0, fmt.Errorf("column %s not found", colName)
	}
	val := strings.TrimSpace(record[idx])
	if val == "" {
		return 0, fmt.Errorf("empty value for %s", colName)
	}
	return strconv.ParseFloat(val, 64)
}

// VerifyAlphaCentauri checks if Alpha Centauri is at expected position.
// Alpha Centauri: RA=219.9°, Dec=-60.8°, parallax=747 mas, distance=4.37 ly
// Expected position roughly: x=-1.64, y=-1.36, z=-3.84 (in our coordinate system)
func VerifyAlphaCentauri(stars []Star) (*Star, error) {
	for i := range stars {
		s := &stars[i]
		// Check if name contains Alpha Centauri or is nearby
		if strings.Contains(strings.ToLower(s.Name), "alpha cen") ||
			strings.Contains(strings.ToLower(s.Name), "rigil kent") {
			return s, nil
		}

		// Check by position - should be ~4.37 ly from origin
		dist := math.Sqrt(s.X*s.X + s.Y*s.Y + s.Z*s.Z)
		if dist > 4.0 && dist < 5.0 {
			// Check approximate position
			// Alpha Cen is in southern sky toward RA ~220° (negative X, negative Y in our system)
			if s.X < -1.0 && s.Y < -1.0 && s.Z < -3.0 {
				return s, nil
			}
		}
	}
	return nil, fmt.Errorf("Alpha Centauri not found in catalog")
}
