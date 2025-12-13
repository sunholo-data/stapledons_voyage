// Package lod provides Level of Detail management for rendering many celestial objects.
package lod

// LODTier represents the detail level for rendering an object.
type LODTier int

const (
	// TierFull3D renders as full Tetra3D mesh (planets, rings)
	TierFull3D LODTier = iota
	// TierBillboard renders as 2D sprite facing camera
	TierBillboard
	// TierCircle renders as filled 2D circle
	TierCircle
	// TierPoint renders as single pixel
	TierPoint
	// TierCulled is not rendered (too far or off-screen)
	TierCulled
)

// String returns the tier name.
func (t LODTier) String() string {
	switch t {
	case TierFull3D:
		return "Full3D"
	case TierBillboard:
		return "Billboard"
	case TierCircle:
		return "Circle"
	case TierPoint:
		return "Point"
	case TierCulled:
		return "Culled"
	default:
		return "Unknown"
	}
}

// Config defines distance thresholds for LOD tier transitions.
type Config struct {
	// Full3DDistance - objects closer than this use full 3D mesh
	Full3DDistance float64
	// BillboardDistance - objects closer than this use billboard sprites
	BillboardDistance float64
	// CircleDistance - objects closer than this use 2D circles
	CircleDistance float64
	// PointDistance - objects closer than this use single points
	// Objects beyond this are culled
	PointDistance float64
	// Max3DObjects limits how many objects can be in Full3D tier
	Max3DObjects int
}

// DefaultConfig returns reasonable default LOD thresholds.
func DefaultConfig() Config {
	return Config{
		Full3DDistance:    50,
		BillboardDistance: 200,
		CircleDistance:    1000,
		PointDistance:     10000,
		Max3DObjects:      30,
	}
}

// GalaxyConfig returns LOD thresholds optimized for galaxy-scale views.
func GalaxyConfig() Config {
	return Config{
		Full3DDistance:    20,
		BillboardDistance: 100,
		CircleDistance:    500,
		PointDistance:     50000,
		Max3DObjects:      10,
	}
}

// SystemConfig returns LOD thresholds optimized for star system views.
func SystemConfig() Config {
	return Config{
		Full3DDistance:    100,
		BillboardDistance: 500,
		CircleDistance:    2000,
		PointDistance:     20000,
		Max3DObjects:      50,
	}
}
