package stardata

import (
	"math"
)

// AABB represents an axis-aligned bounding box in 3D space.
type AABB struct {
	MinX, MinY, MinZ float64
	MaxX, MaxY, MaxZ float64
}

// Center returns the center point of the bounding box.
func (a AABB) Center() (x, y, z float64) {
	return (a.MinX + a.MaxX) / 2, (a.MinY + a.MaxY) / 2, (a.MinZ + a.MaxZ) / 2
}

// Contains checks if a point is inside the bounding box.
func (a AABB) Contains(x, y, z float64) bool {
	return x >= a.MinX && x <= a.MaxX &&
		y >= a.MinY && y <= a.MaxY &&
		z >= a.MinZ && z <= a.MaxZ
}

// IntersectsSphere checks if the AABB intersects with a sphere.
func (a AABB) IntersectsSphere(cx, cy, cz, radius float64) bool {
	// Find the closest point on the AABB to the sphere center
	closestX := math.Max(a.MinX, math.Min(cx, a.MaxX))
	closestY := math.Max(a.MinY, math.Min(cy, a.MaxY))
	closestZ := math.Max(a.MinZ, math.Min(cz, a.MaxZ))

	// Calculate distance from closest point to sphere center
	dx := closestX - cx
	dy := closestY - cy
	dz := closestZ - cz

	return dx*dx+dy*dy+dz*dz <= radius*radius
}

// OctreeNode represents a node in the octree.
type OctreeNode struct {
	Bounds   AABB
	Stars    []*Star     // Leaf data (non-nil only in leaf nodes)
	Children [8]*OctreeNode // Child nodes (nil in leaf nodes)
	IsLeaf   bool
}

// Octree provides efficient spatial queries for stars.
type Octree struct {
	Root     *OctreeNode
	MaxDepth int
	MaxStars int // Max stars per leaf before subdivision
}

// NewOctree creates a new octree with the given bounds.
func NewOctree(bounds AABB, maxDepth, maxStars int) *Octree {
	return &Octree{
		Root: &OctreeNode{
			Bounds: bounds,
			Stars:  make([]*Star, 0),
			IsLeaf: true,
		},
		MaxDepth: maxDepth,
		MaxStars: maxStars,
	}
}

// BuildOctree constructs an octree from a catalog.
func BuildOctree(catalog *Catalog) *Octree {
	if len(catalog.Stars) == 0 {
		return NewOctree(AABB{-1, -1, -1, 1, 1, 1}, 10, 16)
	}

	// Find bounding box for all stars
	minX, minY, minZ := math.MaxFloat64, math.MaxFloat64, math.MaxFloat64
	maxX, maxY, maxZ := -math.MaxFloat64, -math.MaxFloat64, -math.MaxFloat64

	for i := range catalog.Stars {
		s := &catalog.Stars[i]
		if s.X < minX {
			minX = s.X
		}
		if s.Y < minY {
			minY = s.Y
		}
		if s.Z < minZ {
			minZ = s.Z
		}
		if s.X > maxX {
			maxX = s.X
		}
		if s.Y > maxY {
			maxY = s.Y
		}
		if s.Z > maxZ {
			maxZ = s.Z
		}
	}

	// Add small padding
	padding := 1.0
	bounds := AABB{
		MinX: minX - padding,
		MinY: minY - padding,
		MinZ: minZ - padding,
		MaxX: maxX + padding,
		MaxY: maxY + padding,
		MaxZ: maxZ + padding,
	}

	// Create octree with reasonable defaults
	octree := NewOctree(bounds, 12, 16)

	// Insert all stars
	for i := range catalog.Stars {
		octree.Insert(&catalog.Stars[i])
	}

	return octree
}

// Insert adds a star to the octree.
func (o *Octree) Insert(star *Star) {
	o.insertNode(o.Root, star, 0)
}

func (o *Octree) insertNode(node *OctreeNode, star *Star, depth int) {
	// Check if star is within bounds
	if !node.Bounds.Contains(star.X, star.Y, star.Z) {
		return
	}

	if node.IsLeaf {
		// Add to leaf node
		node.Stars = append(node.Stars, star)

		// Subdivide if needed and not at max depth
		if len(node.Stars) > o.MaxStars && depth < o.MaxDepth {
			o.subdivide(node, depth)
		}
	} else {
		// Insert into appropriate child
		childIdx := o.getChildIndex(node, star.X, star.Y, star.Z)
		if node.Children[childIdx] != nil {
			o.insertNode(node.Children[childIdx], star, depth+1)
		}
	}
}

func (o *Octree) subdivide(node *OctreeNode, depth int) {
	cx, cy, cz := node.Bounds.Center()

	// Create 8 children
	for i := 0; i < 8; i++ {
		childBounds := o.getChildBounds(node.Bounds, i, cx, cy, cz)
		node.Children[i] = &OctreeNode{
			Bounds: childBounds,
			Stars:  make([]*Star, 0),
			IsLeaf: true,
		}
	}

	// Redistribute stars to children
	for _, star := range node.Stars {
		childIdx := o.getChildIndex(node, star.X, star.Y, star.Z)
		o.insertNode(node.Children[childIdx], star, depth+1)
	}

	// Clear parent stars and mark as non-leaf
	node.Stars = nil
	node.IsLeaf = false
}

func (o *Octree) getChildIndex(node *OctreeNode, x, y, z float64) int {
	cx, cy, cz := node.Bounds.Center()
	idx := 0
	if x >= cx {
		idx |= 1
	}
	if y >= cy {
		idx |= 2
	}
	if z >= cz {
		idx |= 4
	}
	return idx
}

func (o *Octree) getChildBounds(parent AABB, idx int, cx, cy, cz float64) AABB {
	bounds := AABB{}

	if idx&1 == 0 {
		bounds.MinX = parent.MinX
		bounds.MaxX = cx
	} else {
		bounds.MinX = cx
		bounds.MaxX = parent.MaxX
	}

	if idx&2 == 0 {
		bounds.MinY = parent.MinY
		bounds.MaxY = cy
	} else {
		bounds.MinY = cy
		bounds.MaxY = parent.MaxY
	}

	if idx&4 == 0 {
		bounds.MinZ = parent.MinZ
		bounds.MaxZ = cz
	} else {
		bounds.MinZ = cz
		bounds.MaxZ = parent.MaxZ
	}

	return bounds
}

// Query returns all stars within the given radius of the center point.
func (o *Octree) Query(cx, cy, cz, radius float64) []*Star {
	result := make([]*Star, 0)
	o.queryNode(o.Root, cx, cy, cz, radius, &result)
	return result
}

func (o *Octree) queryNode(node *OctreeNode, cx, cy, cz, radius float64, result *[]*Star) {
	if node == nil {
		return
	}

	// Skip if node's bounds don't intersect the query sphere
	if !node.Bounds.IntersectsSphere(cx, cy, cz, radius) {
		return
	}

	if node.IsLeaf {
		// Check each star in leaf
		radiusSq := radius * radius
		for _, star := range node.Stars {
			dx := star.X - cx
			dy := star.Y - cy
			dz := star.Z - cz
			if dx*dx+dy*dy+dz*dz <= radiusSq {
				*result = append(*result, star)
			}
		}
	} else {
		// Recurse into children
		for _, child := range node.Children {
			o.queryNode(child, cx, cy, cz, radius, result)
		}
	}
}

// QueryNearest returns the N nearest stars to the given point.
func (o *Octree) QueryNearest(cx, cy, cz float64, n int) []*Star {
	if n <= 0 {
		return nil
	}

	// Start with a small radius and expand until we have enough stars
	radius := 10.0
	maxRadius := 100000.0

	for radius <= maxRadius {
		stars := o.Query(cx, cy, cz, radius)
		if len(stars) >= n {
			// Sort by distance and return top N
			sortStarsByDistance(stars, cx, cy, cz)
			if len(stars) > n {
				return stars[:n]
			}
			return stars
		}
		radius *= 2
	}

	// Return whatever we found
	stars := o.Query(cx, cy, cz, maxRadius)
	sortStarsByDistance(stars, cx, cy, cz)
	if len(stars) > n {
		return stars[:n]
	}
	return stars
}

// sortStarsByDistance sorts stars by distance from a point (in-place).
func sortStarsByDistance(stars []*Star, cx, cy, cz float64) {
	// Simple insertion sort for small lists, more efficient for typical queries
	for i := 1; i < len(stars); i++ {
		key := stars[i]
		keyDist := distanceSq(key.X, key.Y, key.Z, cx, cy, cz)
		j := i - 1

		for j >= 0 && distanceSq(stars[j].X, stars[j].Y, stars[j].Z, cx, cy, cz) > keyDist {
			stars[j+1] = stars[j]
			j--
		}
		stars[j+1] = key
	}
}

func distanceSq(x1, y1, z1, x2, y2, z2 float64) float64 {
	dx := x1 - x2
	dy := y1 - y2
	dz := z1 - z2
	return dx*dx + dy*dy + dz*dz
}

// Stats returns statistics about the octree structure.
type OctreeStats struct {
	TotalNodes  int
	LeafNodes   int
	MaxDepth    int
	TotalStars  int
	AvgPerLeaf  float64
}

// GetStats returns statistics about the octree.
func (o *Octree) GetStats() OctreeStats {
	stats := OctreeStats{}
	o.collectStats(o.Root, 0, &stats)
	if stats.LeafNodes > 0 {
		stats.AvgPerLeaf = float64(stats.TotalStars) / float64(stats.LeafNodes)
	}
	return stats
}

func (o *Octree) collectStats(node *OctreeNode, depth int, stats *OctreeStats) {
	if node == nil {
		return
	}

	stats.TotalNodes++
	if depth > stats.MaxDepth {
		stats.MaxDepth = depth
	}

	if node.IsLeaf {
		stats.LeafNodes++
		stats.TotalStars += len(node.Stars)
	} else {
		for _, child := range node.Children {
			o.collectStats(child, depth+1, stats)
		}
	}
}
