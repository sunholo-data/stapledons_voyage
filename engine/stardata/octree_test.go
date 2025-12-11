package stardata

import (
	"math"
	"testing"
)

func TestAABBContains(t *testing.T) {
	box := AABB{MinX: -10, MinY: -10, MinZ: -10, MaxX: 10, MaxY: 10, MaxZ: 10}

	// Point inside
	if !box.Contains(0, 0, 0) {
		t.Error("Origin should be inside box")
	}

	// Point on boundary
	if !box.Contains(10, 0, 0) {
		t.Error("Boundary point should be inside box")
	}

	// Point outside
	if box.Contains(15, 0, 0) {
		t.Error("Point at (15,0,0) should be outside box")
	}
}

func TestAABBIntersectsSphere(t *testing.T) {
	box := AABB{MinX: 0, MinY: 0, MinZ: 0, MaxX: 10, MaxY: 10, MaxZ: 10}

	// Sphere centered in box
	if !box.IntersectsSphere(5, 5, 5, 2) {
		t.Error("Sphere centered in box should intersect")
	}

	// Sphere overlapping corner
	if !box.IntersectsSphere(-1, -1, -1, 3) {
		t.Error("Sphere near corner should intersect")
	}

	// Sphere far away
	if box.IntersectsSphere(100, 100, 100, 1) {
		t.Error("Far sphere should not intersect")
	}
}

func TestOctreeBasic(t *testing.T) {
	// Create catalog with a few stars
	catalog := NewCatalog()
	catalog.AddSol()
	catalog.AddStar(Star{ID: 1, Name: "Alpha Centauri", X: -1.55, Y: -1.32, Z: -3.77})
	catalog.AddStar(Star{ID: 2, Name: "Barnard's Star", X: -0.06, Y: 5.94, Z: 0.49})
	catalog.AddStar(Star{ID: 3, Name: "Far Star", X: 100, Y: 100, Z: 100})

	octree := BuildOctree(catalog)

	// Query near Sol
	stars := octree.Query(0, 0, 0, 10)
	if len(stars) != 3 { // Sol, Alpha Centauri, Barnard's Star
		t.Errorf("Expected 3 stars near Sol within 10 ly, got %d", len(stars))
	}

	// Query that should include far star
	stars = octree.Query(0, 0, 0, 200)
	if len(stars) != 4 {
		t.Errorf("Expected 4 stars within 200 ly, got %d", len(stars))
	}
}

func TestOctreeQueryNearest(t *testing.T) {
	catalog := NewCatalog()
	catalog.AddSol()
	catalog.AddStar(Star{ID: 1, Name: "Proxima", X: -1.55, Y: -1.32, Z: -3.77})
	catalog.AddStar(Star{ID: 2, Name: "Barnard", X: -0.06, Y: 5.94, Z: 0.49})
	catalog.AddStar(Star{ID: 3, Name: "Wolf 359", X: -7.43, Y: 2.11, Z: -0.66})

	octree := BuildOctree(catalog)

	// Get 2 nearest to Sol
	nearest := octree.QueryNearest(0, 0, 0, 2)
	if len(nearest) != 2 {
		t.Errorf("Expected 2 nearest stars, got %d", len(nearest))
	}

	// First should be Sol (distance 0)
	if nearest[0].Name != "Sol" {
		t.Errorf("Nearest to origin should be Sol, got %s", nearest[0].Name)
	}

	// Second should be Proxima (~4.24 ly)
	if nearest[1].Name != "Proxima" {
		t.Errorf("Second nearest should be Proxima, got %s", nearest[1].Name)
	}
}

func TestOctreePerformance(t *testing.T) {
	// Create catalog with many stars centered around origin
	catalog := NewCatalog()

	// Generate 1000 stars in a 200 ly cube centered at origin
	for i := int64(0); i < 1000; i++ {
		// Distribute stars in a grid centered at origin
		x := float64(i%10)*20 - 90
		y := float64((i/10)%10)*20 - 90
		z := float64((i/100)%10)*20 - 90
		catalog.AddStar(Star{ID: i, Name: "Star", X: x, Y: y, Z: z})
	}

	octree := BuildOctree(catalog)
	stats := octree.GetStats()

	t.Logf("Octree stats: nodes=%d, leaves=%d, depth=%d, stars=%d, avg/leaf=%.1f",
		stats.TotalNodes, stats.LeafNodes, stats.MaxDepth, stats.TotalStars, stats.AvgPerLeaf)

	// Query should be much faster than linear scan
	// Just verify it works and returns reasonable results
	stars := octree.Query(0, 0, 0, 100)
	if len(stars) == 0 {
		t.Error("Expected some stars within 100 ly of origin")
	}
	t.Logf("Found %d stars within 100 ly", len(stars))
}

func BenchmarkOctreeQuery(b *testing.B) {
	// Create catalog with 10000 stars centered around origin
	catalog := NewCatalog()
	for i := int64(0); i < 10000; i++ {
		x := float64(i%100)*2 - 100
		y := float64((i/100)%100)*2 - 100
		z := float64((i/10000)%100)*2 - 100
		catalog.AddStar(Star{ID: i, Name: "Star", X: x, Y: y, Z: z})
	}

	octree := BuildOctree(catalog)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		octree.Query(0, 0, 0, 50)
	}
}

func BenchmarkLinearQuery(b *testing.B) {
	// Create catalog with 10000 stars centered around origin
	catalog := NewCatalog()
	for i := int64(0); i < 10000; i++ {
		x := float64(i%100)*2 - 100
		y := float64((i/100)%100)*2 - 100
		z := float64((i/10000)%100)*2 - 100
		catalog.AddStar(Star{ID: i, Name: "Star", X: x, Y: y, Z: z})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		catalog.StarsWithinRadius(0, 0, 0, 50)
	}
}

func TestDistanceAccuracy(t *testing.T) {
	// Verify octree returns same stars as linear search
	catalog := NewCatalog()
	catalog.AddSol()
	catalog.AddStar(Star{ID: 1, X: 4.37, Y: 0, Z: 0})  // Exactly 4.37 ly
	catalog.AddStar(Star{ID: 2, X: 5, Y: 0, Z: 0})      // Exactly 5 ly
	catalog.AddStar(Star{ID: 3, X: 10.1, Y: 0, Z: 0})   // Just over 10 ly

	octree := BuildOctree(catalog)

	// Query with radius 10 should get 3 stars (Sol, 4.37, 5)
	octreeStars := octree.Query(0, 0, 0, 10)
	linearStars := catalog.StarsWithinRadius(0, 0, 0, 10)

	if len(octreeStars) != len(linearStars) {
		t.Errorf("Octree returned %d stars, linear returned %d", len(octreeStars), len(linearStars))
	}

	// Verify boundary case
	octreeStars = octree.Query(0, 0, 0, 4.37)
	count := 0
	for _, s := range octreeStars {
		dist := math.Sqrt(s.X*s.X + s.Y*s.Y + s.Z*s.Z)
		if dist <= 4.37 {
			count++
		}
	}
	if count < 2 { // Sol and the 4.37 star
		t.Errorf("Expected at least 2 stars within 4.37 ly, got %d valid", count)
	}
}
