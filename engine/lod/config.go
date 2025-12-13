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

// Config defines LOD tier thresholds.
// Uses apparent size (pixels on screen) rather than distance, which automatically
// handles object radius, camera distance, FOV, and resolution.
type Config struct {
	// Apparent size thresholds (in screen pixels)
	// Objects upgrade to a tier when apparent radius exceeds these values
	Full3DPixels    float64 // Minimum pixels for Full3D (e.g., 40)
	BillboardPixels float64 // Minimum pixels for Billboard (e.g., 10)
	CirclePixels    float64 // Minimum pixels for Circle (e.g., 3)
	PointPixels     float64 // Minimum pixels for Point (e.g., 0.5)

	// Legacy distance-based thresholds (used as fallback/override)
	Full3DDistance    float64
	BillboardDistance float64
	CircleDistance    float64
	PointDistance     float64

	// Max3DObjects limits how many objects can be in Full3D tier
	Max3DObjects int

	// Hysteresis prevents flickering at tier boundaries (0.0 to 0.5)
	// A value of 0.2 means downgrade threshold is 20% lower than upgrade
	Hysteresis float64

	// TransitionTime is how long tier transitions take (seconds)
	// Set to 0 for instant transitions
	TransitionTime float64

	// UseApparentSize enables pixel-based thresholds (recommended)
	// When false, uses legacy distance-based thresholds
	UseApparentSize bool
}

// DefaultConfig returns reasonable default LOD thresholds.
// Uses apparent size (pixels) by default for automatic scaling.
func DefaultConfig() Config {
	return Config{
		// Apparent size thresholds (recommended)
		// These are apparent RADIUS in pixels
		// Transitions should be smooth: point → circle → billboard → 3D
		// Key: show texture (billboard) early so planets look good from distance
		Full3DPixels:    12,  // Show 3D when object is 12+ px radius (24px diameter)
		BillboardPixels: 6,   // Show billboard when 6+ px radius (12px diameter) - texture visible early!
		CirclePixels:    3,   // Show circle when 3+ px radius (6px diameter)
		PointPixels:     1,   // Show point when 1+ px (below this = culled)

		// Legacy distance thresholds (fallback)
		Full3DDistance:    50,
		BillboardDistance: 200,
		CircleDistance:    1000,
		PointDistance:     10000,

		Max3DObjects:    30,
		Hysteresis:      0.2, // 20% hysteresis to prevent flickering
		TransitionTime:  0.3, // 300ms smooth transitions
		UseApparentSize: true,
	}
}

// GalaxyConfig returns LOD thresholds optimized for galaxy-scale views.
// Lower pixel thresholds since objects are typically very small at galactic scale.
func GalaxyConfig() Config {
	return Config{
		Full3DPixels:      60,  // 3D at 60px radius
		BillboardPixels:   20,  // Billboard at 20px radius
		CirclePixels:      6,   // Circle at 6px radius
		PointPixels:       1.5, // Point at 1.5px
		Full3DDistance:    20,
		BillboardDistance: 100,
		CircleDistance:    500,
		PointDistance:     50000,
		Max3DObjects:      10,
		Hysteresis:        0.2,
		TransitionTime:    0.3,
		UseApparentSize:   true,
	}
}

// SystemConfig returns LOD thresholds optimized for star system views.
// Higher pixel thresholds for more detail at closer range.
func SystemConfig() Config {
	return Config{
		Full3DPixels:      100, // 3D at 100px radius (200px diameter)
		BillboardPixels:   30,  // Billboard at 30px radius
		CirclePixels:      10,  // Circle at 10px radius
		PointPixels:       2,   // Point at 2px
		Full3DDistance:    100,
		BillboardDistance: 500,
		CircleDistance:    2000,
		PointDistance:     20000,
		Max3DObjects:      50,
		Hysteresis:        0.2,
		TransitionTime:    0.3,
		UseApparentSize:   true,
	}
}
