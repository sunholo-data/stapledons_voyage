package sim_gen

import (
	"encoding/json"
	"math"
	"os"
	"sort"
)

// Star represents a single star in the catalog
type Star struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	X        float64 `json:"x"`        // Light-years (galactic coords)
	Y        float64 `json:"y"`        // Light-years
	Z        float64 `json:"z"`        // Light-years
	DistLY   float64 `json:"dist_ly"`  // Distance from Sol
	VMag     float64 `json:"vmag"`     // Visual magnitude
	Spectral string  `json:"spectral"` // O, B, A, F, G, K, M
}

// StarCatalog holds all loaded star data
type StarCatalog struct {
	Version string `json:"version"`
	Source  string `json:"source"`
	Count   int    `json:"count"`
	Stars   []Star `json:"stars"`
}

// Global star catalog (loaded once at startup)
var loadedStarCatalog *StarCatalog

// LoadStarCatalog loads star data from JSON file
func LoadStarCatalog(path string) (*StarCatalog, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var catalog StarCatalog
	if err := json.Unmarshal(data, &catalog); err != nil {
		return nil, err
	}

	loadedStarCatalog = &catalog
	return &catalog, nil
}

// GetStarCatalog returns the loaded catalog (or nil if not loaded)
func GetStarCatalog() *StarCatalog {
	return loadedStarCatalog
}

// StarsWithinRadius returns stars within given radius of a point (in light-years)
func (c *StarCatalog) StarsWithinRadius(cx, cy, cz, radius float64) []Star {
	if c == nil {
		return nil
	}

	var result []Star
	r2 := radius * radius

	for _, s := range c.Stars {
		dx := s.X - cx
		dy := s.Y - cy
		dz := s.Z - cz
		if dx*dx+dy*dy+dz*dz <= r2 {
			result = append(result, s)
		}
	}

	return result
}

// NearestStars returns the N nearest stars to a point
func (c *StarCatalog) NearestStars(cx, cy, cz float64, n int) []Star {
	if c == nil || len(c.Stars) == 0 {
		return nil
	}

	// Calculate distances and sort
	type starDist struct {
		star Star
		dist float64
	}

	distances := make([]starDist, len(c.Stars))
	for i, s := range c.Stars {
		dx := s.X - cx
		dy := s.Y - cy
		dz := s.Z - cz
		distances[i] = starDist{s, dx*dx + dy*dy + dz*dz}
	}

	sort.Slice(distances, func(i, j int) bool {
		return distances[i].dist < distances[j].dist
	})

	if n > len(distances) {
		n = len(distances)
	}

	result := make([]Star, n)
	for i := 0; i < n; i++ {
		result[i] = distances[i].star
	}

	return result
}

// SpectralColor returns a color index for spectral type
// Uses same color indices as biomeColors in renderer
func SpectralColor(spectral string) int {
	switch spectral {
	case "O":
		return 0 // Blue (very hot)
	case "B":
		return 0 // Blue-white
	case "A":
		return 4 // White
	case "F":
		return 13 // Yellow-white
	case "G":
		return 13 // Yellow (Sun-like)
	case "K":
		return 11 // Orange
	case "M":
		return 10 // Red
	default:
		return 4 // White default
	}
}

// StarRadius calculates visual radius based on magnitude
// Brighter stars (lower magnitude) are larger
func StarRadius(vmag float64, baseRadius float64) float64 {
	// Magnitude scale: -1.5 (Sirius) to 15+ (faint)
	// Map to radius: bright = larger
	if vmag < -1 {
		return baseRadius * 3.0
	} else if vmag < 2 {
		return baseRadius * 2.0
	} else if vmag < 5 {
		return baseRadius * 1.5
	} else if vmag < 8 {
		return baseRadius * 1.0
	} else if vmag < 11 {
		return baseRadius * 0.7
	}
	return baseRadius * 0.5
}

// GalacticLonLat converts cartesian galactic coordinates (X,Y,Z in light-years)
// to galactic longitude and latitude (in degrees).
// Returns (longitude, latitude) where:
//   - longitude: 0-360° (0° toward galactic center)
//   - latitude: -90° to +90° (0° on galactic plane)
func GalacticLonLat(x, y, z float64) (lon, lat float64) {
	// Distance from origin (Sol)
	r := math.Sqrt(x*x + y*y + z*z)
	if r < 0.001 {
		return 0, 0 // At origin (Sol)
	}

	// Galactic longitude: angle in X-Y plane from +X axis
	// In galactic coords, +X points toward galactic center
	lon = math.Atan2(y, x) * 180.0 / math.Pi
	if lon < 0 {
		lon += 360.0 // Normalize to 0-360
	}

	// Galactic latitude: angle above/below galactic plane
	lat = math.Asin(z/r) * 180.0 / math.Pi

	return lon, lat
}

// StarGalacticLonLat returns the galactic longitude and latitude for a star
func (s *Star) GalacticLonLat() (lon, lat float64) {
	return GalacticLonLat(s.X, s.Y, s.Z)
}

// AngularDistance returns the angular distance in degrees between two sky positions
func AngularDistance(lon1, lat1, lon2, lat2 float64) float64 {
	// Convert to radians
	lon1r := lon1 * math.Pi / 180.0
	lat1r := lat1 * math.Pi / 180.0
	lon2r := lon2 * math.Pi / 180.0
	lat2r := lat2 * math.Pi / 180.0

	// Haversine formula
	dlon := lon2r - lon1r
	dlat := lat2r - lat1r
	a := math.Sin(dlat/2)*math.Sin(dlat/2) +
		math.Cos(lat1r)*math.Cos(lat2r)*math.Sin(dlon/2)*math.Sin(dlon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return c * 180.0 / math.Pi // Return in degrees
}

// SpectralSpriteID returns the star sprite ID for a spectral type
// Sprite IDs: 200=blue(O/B), 201=white(A/F), 202=yellow(G), 203=orange(K), 204=red(M)
func SpectralSpriteID(spectral string) int {
	switch spectral {
	case "O", "B":
		return 200 // Blue (hot)
	case "A", "F":
		return 201 // White
	case "G":
		return 202 // Yellow (Sun-like)
	case "K":
		return 203 // Orange
	case "M":
		return 204 // Red
	default:
		return 201 // White default
	}
}

// StarScale calculates the sprite scale based on magnitude
// Returns scale factor where 1.0 = 16x16 pixels (base sprite size)
func StarScale(vmag float64) float64 {
	// Magnitude scale: -1.5 (Sirius) to 15+ (faint)
	// Map to scale: bright = larger
	if vmag < -1 {
		return 1.2 // Very bright
	} else if vmag < 2 {
		return 0.8 // Bright
	} else if vmag < 5 {
		return 0.5 // Medium
	} else if vmag < 8 {
		return 0.3 // Dim
	} else if vmag < 11 {
		return 0.2 // Faint
	}
	return 0.15 // Very faint
}

// FindNearestStarToScreen finds the star closest to a screen position
// Returns the star index in the catalog, or -1 if no star is close enough
// clickRadius is the maximum screen distance (in pixels) to consider a hit
func FindNearestStarToScreen(catalog *StarCatalog, screenX, screenY float64, mode ModeGalaxyMap, clickRadius float64) int {
	if catalog == nil || len(catalog.Stars) == 0 {
		return -1
	}

	screenW := float64(ScreenWidth)
	screenH := float64(ScreenHeight)
	screenCenterX := screenW / 2
	screenCenterY := screenH / 2
	pixelsPerLY := screenW / 160.0 * 0.6

	bestIdx := -1
	bestDist := clickRadius * clickRadius // Work with squared distances

	for i, star := range catalog.Stars {
		// Calculate star's screen position (same as render logic)
		starZ := star.Z
		parallaxFactor := 1.0 / (1.0 + math.Abs(starZ)*0.008)

		sx := (star.X-mode.CameraX*parallaxFactor)*pixelsPerLY*mode.ZoomLevel + screenCenterX
		sy := (star.Y-mode.CameraY*parallaxFactor)*pixelsPerLY*mode.ZoomLevel + screenCenterY

		// Calculate squared distance to click point
		dx := sx - screenX
		dy := sy - screenY
		distSq := dx*dx + dy*dy

		if distSq < bestDist {
			bestDist = distSq
			bestIdx = i
		}
	}

	return bestIdx
}
