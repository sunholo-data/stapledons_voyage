// Package relativity implements Special Relativity optical transformations
// for rendering relativistic visual effects (aberration, Doppler, beaming).
package relativity

import "math"

// Vec3 represents a 3D vector for relativistic calculations.
type Vec3 struct {
	X, Y, Z float64
}

// Add returns v + other.
func (v Vec3) Add(other Vec3) Vec3 {
	return Vec3{v.X + other.X, v.Y + other.Y, v.Z + other.Z}
}

// Sub returns v - other.
func (v Vec3) Sub(other Vec3) Vec3 {
	return Vec3{v.X - other.X, v.Y - other.Y, v.Z - other.Z}
}

// Scale returns v * scalar.
func (v Vec3) Scale(s float64) Vec3 {
	return Vec3{v.X * s, v.Y * s, v.Z * s}
}

// Dot returns the dot product v . other.
func (v Vec3) Dot(other Vec3) float64 {
	return v.X*other.X + v.Y*other.Y + v.Z*other.Z
}

// Length returns |v|.
func (v Vec3) Length() float64 {
	return math.Sqrt(v.Dot(v))
}

// Normalize returns v / |v|. Returns zero vector if length is near zero.
func (v Vec3) Normalize() Vec3 {
	l := v.Length()
	if l < 1e-10 {
		return Vec3{}
	}
	return v.Scale(1.0 / l)
}

// Gamma computes the Lorentz factor from velocity magnitude beta (in units of c).
// gamma = 1 / sqrt(1 - beta^2)
func Gamma(beta float64) float64 {
	if beta >= 1.0 {
		return 100.0 // Clamp to avoid infinity
	}
	if beta <= 0.0 {
		return 1.0
	}
	return 1.0 / math.Sqrt(1.0-beta*beta)
}

// GammaFromVec computes gamma from a velocity vector beta (in units of c).
func GammaFromVec(beta Vec3) float64 {
	return Gamma(beta.Length())
}

// DopplerFactor computes the relativistic Doppler factor for light arriving
// from direction n (unit vector pointing FROM source TO observer) when
// the observer moves with velocity beta (in units of c).
//
// D = gamma * (1 - beta . n)
//
// D > 1 means blueshift (approaching), D < 1 means redshift (receding).
func DopplerFactor(beta Vec3, n Vec3, gamma float64) float64 {
	// n points from source to observer
	// beta . n > 0 when moving toward source (blueshift)
	d := gamma * (1.0 - beta.Dot(n))
	// Clamp to reasonable range to avoid singularities
	if d < 0.01 {
		return 0.01
	}
	if d > 100.0 {
		return 100.0
	}
	return d
}

// TransformDirection applies relativistic aberration to transform a photon
// direction from the galaxy (rest) frame to the ship (moving) frame.
//
// n: unit vector pointing from observer to light source in galaxy frame
// beta: observer velocity vector in units of c
// gamma: Lorentz factor
//
// Returns: unit vector in ship frame (where photon appears to come from)
func TransformDirection(n Vec3, beta Vec3, gamma float64) Vec3 {
	betaMag := beta.Length()
	if betaMag < 1e-10 {
		return n // No transform at rest
	}

	// Unit vector in direction of motion
	betaHat := beta.Normalize()

	// Decompose n into parallel and perpendicular components
	nDotBetaHat := n.Dot(betaHat)
	nParallel := betaHat.Scale(nDotBetaHat)
	nPerp := n.Sub(nParallel)

	// Denominator: 1 - beta . n
	denom := 1.0 - beta.Dot(n)
	if math.Abs(denom) < 1e-10 {
		denom = 1e-10 // Avoid division by zero
	}

	// Transform components
	// n'_parallel = (n_parallel - beta) / (1 - beta . n)
	// n'_perp = n_perp / (gamma * (1 - beta . n))
	nPrimeParallel := nParallel.Sub(beta).Scale(1.0 / denom)
	nPrimePerp := nPerp.Scale(1.0 / (gamma * denom))

	// Combine and normalize
	return nPrimeParallel.Add(nPrimePerp).Normalize()
}

// InverseTransformDirection applies inverse aberration - transforms a direction
// from ship frame back to galaxy frame. Used for sampling skybox textures.
//
// nPrime: direction in ship frame (where we're looking)
// beta: observer velocity in units of c
// gamma: Lorentz factor
//
// Returns: direction in galaxy frame (where to sample skybox)
func InverseTransformDirection(nPrime Vec3, beta Vec3, gamma float64) Vec3 {
	// The inverse transform uses -beta
	return TransformDirection(nPrime, beta.Scale(-1), gamma)
}

// ScreenToDirection converts a screen pixel position to a unit direction vector
// in the ship's view frame. Assumes camera looking along +Z axis.
//
// x, y: screen coordinates (0-1 range, centered at 0.5, 0.5)
// fov: horizontal field of view in degrees
// aspect: screen width / height
func ScreenToDirection(x, y float64, fov float64, aspect float64) Vec3 {
	// Convert FOV to radians and compute half-angles
	fovRad := fov * math.Pi / 180.0
	halfW := math.Tan(fovRad / 2.0)
	halfH := halfW / aspect

	// Map screen coords to view plane
	// x=0.5, y=0.5 -> center -> looking straight ahead (+Z)
	vx := (x - 0.5) * 2.0 * halfW
	vy := (0.5 - y) * 2.0 * halfH // Y is flipped (screen Y goes down)
	vz := 1.0

	return Vec3{vx, vy, vz}.Normalize()
}

// DirectionToEquirectangular converts a 3D direction to equirectangular UV coords.
// Used for sampling galaxy background textures.
//
// Returns (u, v) in range [0, 1] where:
// - u=0 is longitude -180, u=1 is longitude +180
// - v=0 is latitude +90 (north pole), v=1 is latitude -90 (south pole)
func DirectionToEquirectangular(dir Vec3) (u, v float64) {
	// Compute longitude (azimuth) and latitude (elevation)
	// Assuming Y is up, X is right, Z is forward
	lon := math.Atan2(dir.X, dir.Z)           // -PI to PI
	lat := math.Asin(clamp(dir.Y, -1.0, 1.0)) // -PI/2 to PI/2

	// Convert to UV
	u = (lon + math.Pi) / (2.0 * math.Pi)  // 0 to 1
	v = 0.5 - lat/math.Pi                   // 0 to 1 (north at top)

	return u, v
}

func clamp(x, min, max float64) float64 {
	if x < min {
		return min
	}
	if x > max {
		return max
	}
	return x
}
