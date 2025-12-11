package stardata

import (
	"fmt"
	"math"
	"sort"
)

// Catalog holds all loaded stars and provides query methods.
type Catalog struct {
	Stars    []Star
	byID     map[int64]*Star
	solIndex int // Index of Sol in Stars slice (-1 if not present)
}

// NewCatalog creates an empty catalog.
func NewCatalog() *Catalog {
	return &Catalog{
		Stars:    make([]Star, 0),
		byID:     make(map[int64]*Star),
		solIndex: -1,
	}
}

// AddStar adds a star to the catalog.
func (c *Catalog) AddStar(s Star) {
	c.Stars = append(c.Stars, s)
	c.byID[s.ID] = &c.Stars[len(c.Stars)-1]
}

// AddSol adds Sol (our Sun) at the origin.
func (c *Catalog) AddSol() {
	sol := Star{
		ID:           0,
		Name:         "Sol",
		X:            0,
		Y:            0,
		Z:            0,
		SpectralType: "G",
		Luminosity:   1.0,
		HasHZPlanet:  true,
	}
	c.solIndex = len(c.Stars)
	c.AddStar(sol)
}

// GetByID retrieves a star by its ID.
func (c *Catalog) GetByID(id int64) (*Star, bool) {
	s, ok := c.byID[id]
	return s, ok
}

// Sol returns Sol if present in the catalog.
func (c *Catalog) Sol() (*Star, bool) {
	if c.solIndex >= 0 && c.solIndex < len(c.Stars) {
		return &c.Stars[c.solIndex], true
	}
	return c.GetByID(0)
}

// Count returns the number of stars in the catalog.
func (c *Catalog) Count() int {
	return len(c.Stars)
}

// Distance calculates the Euclidean distance between two points.
func Distance(x1, y1, z1, x2, y2, z2 float64) float64 {
	dx := x1 - x2
	dy := y1 - y2
	dz := z1 - z2
	return math.Sqrt(dx*dx + dy*dy + dz*dz)
}

// DistanceToStar calculates distance from a point to a star.
func DistanceToStar(x, y, z float64, s *Star) float64 {
	return Distance(x, y, z, s.X, s.Y, s.Z)
}

// StarsWithinRadius returns all stars within radius light-years of the given point.
func (c *Catalog) StarsWithinRadius(x, y, z, radius float64) []Star {
	radiusSq := radius * radius
	var result []Star

	for i := range c.Stars {
		s := &c.Stars[i]
		dx := s.X - x
		dy := s.Y - y
		dz := s.Z - z
		distSq := dx*dx + dy*dy + dz*dz
		if distSq <= radiusSq {
			result = append(result, *s)
		}
	}

	return result
}

// NearestStar returns the star nearest to the given point.
func (c *Catalog) NearestStar(x, y, z float64) (*Star, float64) {
	if len(c.Stars) == 0 {
		return nil, 0
	}

	var nearest *Star
	minDistSq := math.MaxFloat64

	for i := range c.Stars {
		s := &c.Stars[i]
		dx := s.X - x
		dy := s.Y - y
		dz := s.Z - z
		distSq := dx*dx + dy*dy + dz*dz
		if distSq < minDistSq {
			minDistSq = distSq
			nearest = s
		}
	}

	return nearest, math.Sqrt(minDistSq)
}

// NearestNStars returns the N nearest stars to the given point.
func (c *Catalog) NearestNStars(x, y, z float64, n int) []Star {
	if n <= 0 || len(c.Stars) == 0 {
		return nil
	}

	// Calculate distances for all stars
	type starDist struct {
		star *Star
		dist float64
	}
	distances := make([]starDist, len(c.Stars))

	for i := range c.Stars {
		s := &c.Stars[i]
		dx := s.X - x
		dy := s.Y - y
		dz := s.Z - z
		distances[i] = starDist{star: s, dist: math.Sqrt(dx*dx + dy*dy + dz*dz)}
	}

	// Sort by distance
	sort.Slice(distances, func(i, j int) bool {
		return distances[i].dist < distances[j].dist
	})

	// Return top N
	if n > len(distances) {
		n = len(distances)
	}

	result := make([]Star, n)
	for i := 0; i < n; i++ {
		result[i] = *distances[i].star
	}

	return result
}

// FilterBySpectralType returns stars of the given spectral type.
func (c *Catalog) FilterBySpectralType(spectralType string) []Star {
	var result []Star
	for i := range c.Stars {
		if c.Stars[i].SpectralType == spectralType {
			result = append(result, c.Stars[i])
		}
	}
	return result
}

// FilterHZPlanets returns stars that have habitable zone planets.
func (c *Catalog) FilterHZPlanets() []Star {
	var result []Star
	for i := range c.Stars {
		if c.Stars[i].HasHZPlanet {
			result = append(result, c.Stars[i])
		}
	}
	return result
}

// BuildCatalog creates a catalog from Gaia data plus manual entries.
// If gaiaFile is empty, returns catalog with only manual stars.
func BuildCatalog(gaiaFile string, manualStars []Star) (*Catalog, error) {
	c := NewCatalog()

	// Add Sol first
	c.AddSol()

	// Load Gaia data if file provided
	if gaiaFile != "" {
		gaiaStars, err := LoadGaiaCatalog(gaiaFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load Gaia catalog: %w", err)
		}

		// Assign IDs starting after manual stars
		startID := int64(len(manualStars) + 1)
		for i := range gaiaStars {
			gaiaStars[i].ID = startID + int64(i)
			c.AddStar(gaiaStars[i])
		}
	}

	// Add manual stars (these may override Gaia entries by position)
	for _, s := range manualStars {
		// Check for duplicates by position (within 0.1 ly)
		isDupe := false
		for j := range c.Stars {
			if DistanceToStar(s.X, s.Y, s.Z, &c.Stars[j]) < 0.1 {
				isDupe = true
				break
			}
		}
		if !isDupe {
			c.AddStar(s)
		}
	}

	return c, nil
}

// Statistics returns summary statistics about the catalog.
type CatalogStats struct {
	TotalStars   int
	SpectralDist map[string]int
	HZPlanetPct  float64
	MaxDistance  float64
	MeanDistance float64
}

// GetStatistics calculates catalog statistics.
func (c *Catalog) GetStatistics() CatalogStats {
	stats := CatalogStats{
		TotalStars:   len(c.Stars),
		SpectralDist: make(map[string]int),
	}

	if len(c.Stars) == 0 {
		return stats
	}

	var hzCount int
	var totalDist float64

	for i := range c.Stars {
		s := &c.Stars[i]

		// Spectral distribution
		stats.SpectralDist[s.SpectralType]++

		// HZ planet count
		if s.HasHZPlanet {
			hzCount++
		}

		// Distance from Sol
		dist := math.Sqrt(s.X*s.X + s.Y*s.Y + s.Z*s.Z)
		totalDist += dist
		if dist > stats.MaxDistance {
			stats.MaxDistance = dist
		}
	}

	stats.HZPlanetPct = float64(hzCount) / float64(len(c.Stars)) * 100
	stats.MeanDistance = totalDist / float64(len(c.Stars))

	return stats
}

// PrintStatistics prints catalog statistics to stdout.
func (c *Catalog) PrintStatistics() {
	stats := c.GetStatistics()
	fmt.Printf("Star Catalog Statistics:\n")
	fmt.Printf("  Total stars: %d\n", stats.TotalStars)
	fmt.Printf("  HZ planets: %.1f%%\n", stats.HZPlanetPct)
	fmt.Printf("  Max distance: %.1f ly\n", stats.MaxDistance)
	fmt.Printf("  Mean distance: %.1f ly\n", stats.MeanDistance)
	fmt.Printf("  Spectral distribution:\n")
	for _, st := range []string{"O", "B", "A", "F", "G", "K", "M"} {
		if count, ok := stats.SpectralDist[st]; ok {
			pct := float64(count) / float64(stats.TotalStars) * 100
			fmt.Printf("    %s: %d (%.1f%%)\n", st, count, pct)
		}
	}
}
