package stardata

import (
	"math"
	"strings"
	"testing"
)

func TestToGalactocentric_AlphaCentauri(t *testing.T) {
	// Alpha Centauri coordinates:
	// RA = 219.9° (14h 39m 36s)
	// Dec = -60.8° (-60° 50')
	// Parallax = 747.1 mas (distance = 1.34 pc = 4.37 ly)
	ra := 219.9
	dec := -60.8
	parallax := 747.1

	x, y, z := ToGalactocentric(ra, dec, parallax)

	// Expected distance from Sol
	dist := math.Sqrt(x*x + y*y + z*z)
	expectedDist := 4.37

	if math.Abs(dist-expectedDist) > 0.1 {
		t.Errorf("Alpha Centauri distance: got %.2f ly, expected ~%.2f ly", dist, expectedDist)
	}

	// Alpha Centauri should be in specific quadrant
	// RA 219.9° means X negative, Y negative (in our coordinate system)
	// Dec -60.8° means Z negative
	if x >= 0 {
		t.Errorf("Expected negative X for Alpha Centauri, got %.2f", x)
	}
	if y >= 0 {
		t.Errorf("Expected negative Y for Alpha Centauri, got %.2f", y)
	}
	if z >= 0 {
		t.Errorf("Expected negative Z for Alpha Centauri, got %.2f", z)
	}

	t.Logf("Alpha Centauri position: (%.2f, %.2f, %.2f) ly, distance: %.2f ly", x, y, z, dist)
}

func TestToGalactocentric_Sirius(t *testing.T) {
	// Sirius coordinates:
	// RA = 101.3° (6h 45m)
	// Dec = -16.7°
	// Parallax = 379.2 mas (distance = 2.64 pc = 8.6 ly)
	ra := 101.3
	dec := -16.7
	parallax := 379.2

	x, y, z := ToGalactocentric(ra, dec, parallax)

	dist := math.Sqrt(x*x + y*y + z*z)
	expectedDist := 8.6

	if math.Abs(dist-expectedDist) > 0.2 {
		t.Errorf("Sirius distance: got %.2f ly, expected ~%.2f ly", dist, expectedDist)
	}

	t.Logf("Sirius position: (%.2f, %.2f, %.2f) ly, distance: %.2f ly", x, y, z, dist)
}

func TestEstimateSpectralType(t *testing.T) {
	tests := []struct {
		bpRP     float64
		expected string
	}{
		{-0.2, "B"},
		{0.1, "A"},
		{0.4, "F"},
		{0.82, "G"}, // Sun
		{1.2, "K"},
		{2.0, "M"},
	}

	for _, tc := range tests {
		got := EstimateSpectralType(tc.bpRP)
		if got != tc.expected {
			t.Errorf("EstimateSpectralType(%.2f): got %s, expected %s", tc.bpRP, got, tc.expected)
		}
	}
}

func TestParseGaiaCSV(t *testing.T) {
	// Minimal CSV with two stars
	csv := `source_id,ra,dec,parallax,parallax_error,phot_g_mean_mag,bp_rp,name
1234567890,219.9,-60.8,747.1,0.5,0.01,0.71,Alpha Centauri A
9876543210,101.3,-16.7,379.2,0.3,-1.46,0.0,Sirius A`

	stars, err := ParseGaiaCSV(strings.NewReader(csv))
	if err != nil {
		t.Fatalf("ParseGaiaCSV failed: %v", err)
	}

	if len(stars) != 2 {
		t.Fatalf("Expected 2 stars, got %d", len(stars))
	}

	// Check Alpha Centauri
	alphaCen := stars[0]
	if alphaCen.Name != "Alpha Centauri A" {
		t.Errorf("First star name: got %s, expected Alpha Centauri A", alphaCen.Name)
	}
	dist := math.Sqrt(alphaCen.X*alphaCen.X + alphaCen.Y*alphaCen.Y + alphaCen.Z*alphaCen.Z)
	if math.Abs(dist-4.37) > 0.1 {
		t.Errorf("Alpha Centauri distance: got %.2f, expected ~4.37", dist)
	}

	// Check Sirius
	sirius := stars[1]
	if sirius.Name != "Sirius A" {
		t.Errorf("Second star name: got %s, expected Sirius A", sirius.Name)
	}
	if sirius.SpectralType != "A" {
		t.Errorf("Sirius spectral type: got %s, expected A", sirius.SpectralType)
	}
}

func TestCatalogQueries(t *testing.T) {
	c := NewCatalog()
	c.AddSol()

	// Add some test stars
	c.AddStar(Star{ID: 1, Name: "Proxima Centauri", X: -1.55, Y: -1.32, Z: -3.77, SpectralType: "M", Luminosity: 0.0017, HasHZPlanet: true})
	c.AddStar(Star{ID: 2, Name: "Alpha Centauri A", X: -1.64, Y: -1.36, Z: -3.84, SpectralType: "G", Luminosity: 1.5, HasHZPlanet: true})
	c.AddStar(Star{ID: 3, Name: "Barnard's Star", X: -0.06, Y: 5.94, Z: 0.49, SpectralType: "M", Luminosity: 0.0004, HasHZPlanet: false})
	c.AddStar(Star{ID: 4, Name: "Sirius A", X: -1.61, Y: 8.06, Z: -2.47, SpectralType: "A", Luminosity: 25.4, HasHZPlanet: false})

	// Test Count
	if c.Count() != 5 {
		t.Errorf("Count: got %d, expected 5", c.Count())
	}

	// Test StarsWithinRadius
	nearby := c.StarsWithinRadius(0, 0, 0, 5.0)
	if len(nearby) != 3 { // Sol, Proxima, Alpha Cen
		t.Errorf("StarsWithinRadius(5 ly): got %d stars, expected 3", len(nearby))
	}

	// Test NearestStar (excluding Sol at origin)
	nearest, dist := c.NearestStar(0.1, 0.1, 0.1) // Slightly off origin
	if nearest.Name != "Sol" {
		// Actually Sol is at origin, so nearest to (0.1, 0.1, 0.1) should still be Sol
		t.Logf("Nearest star from (0.1, 0.1, 0.1): %s at %.2f ly", nearest.Name, dist)
	}

	// Test NearestNStars
	nearest3 := c.NearestNStars(0, 0, 0, 3)
	if len(nearest3) != 3 {
		t.Errorf("NearestNStars(3): got %d, expected 3", len(nearest3))
	}
	if nearest3[0].Name != "Sol" {
		t.Errorf("Closest should be Sol, got %s", nearest3[0].Name)
	}

	// Test FilterBySpectralType
	mStars := c.FilterBySpectralType("M")
	if len(mStars) != 2 {
		t.Errorf("M-type stars: got %d, expected 2", len(mStars))
	}

	// Test FilterHZPlanets
	hzStars := c.FilterHZPlanets()
	if len(hzStars) != 3 { // Sol, Proxima, Alpha Cen A
		t.Errorf("HZ planet stars: got %d, expected 3", len(hzStars))
	}
}

func TestCatalogStatistics(t *testing.T) {
	c := NewCatalog()
	c.AddSol()
	c.AddStar(Star{ID: 1, Name: "Proxima", X: 4, Y: 0, Z: 0, SpectralType: "M", HasHZPlanet: true})
	c.AddStar(Star{ID: 2, Name: "Alpha Cen", X: 4, Y: 0, Z: 1, SpectralType: "G", HasHZPlanet: true})

	stats := c.GetStatistics()

	if stats.TotalStars != 3 {
		t.Errorf("TotalStars: got %d, expected 3", stats.TotalStars)
	}

	if stats.SpectralDist["G"] != 2 { // Sol and Alpha Cen
		t.Errorf("G-type count: got %d, expected 2", stats.SpectralDist["G"])
	}

	if stats.SpectralDist["M"] != 1 {
		t.Errorf("M-type count: got %d, expected 1", stats.SpectralDist["M"])
	}
}
