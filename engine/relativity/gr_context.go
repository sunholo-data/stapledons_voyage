// Package relativity - General Relativity context for gravitational effects
// near massive objects (black holes, neutron stars, white dwarfs).

package relativity

import "math"

// MassiveObjectKind represents types of compact objects with strong gravity.
type MassiveObjectKind int

const (
	BlackHole MassiveObjectKind = iota
	NeutronStar
	WhiteDwarf
)

func (k MassiveObjectKind) String() string {
	switch k {
	case BlackHole:
		return "BlackHole"
	case NeutronStar:
		return "NeutronStar"
	case WhiteDwarf:
		return "WhiteDwarf"
	default:
		return "Unknown"
	}
}

// GRDangerLevel represents severity of GR effects.
type GRDangerLevel int

const (
	GR_None GRDangerLevel = iota
	GR_Subtle
	GR_Strong
	GR_Extreme
)

func (l GRDangerLevel) String() string {
	switch l {
	case GR_None:
		return "None"
	case GR_Subtle:
		return "Subtle"
	case GR_Strong:
		return "Strong"
	case GR_Extreme:
		return "Extreme"
	default:
		return "Unknown"
	}
}

// MassiveObject represents a compact object with strong gravitational field.
type MassiveObject struct {
	Kind                 MassiveObjectKind
	MassSolar            float64 // Mass in solar masses
	SchwarzschildRadius  float64 // r_s = 2GM/c^2 in game units
	Position             Vec3    // Position in galaxy frame
	SurfaceRadius        float64 // For NS/WD: actual surface radius
}

// GRContext holds computed GR quantities for the current frame.
type GRContext struct {
	Active            bool
	ObjectKind        MassiveObjectKind
	Distance          float64       // Ship distance from object center
	Phi               float64       // Dimensionless potential: GM/(rc^2) = r_s/(2r)
	TimeDilation      float64       // dτ/dt = sqrt(1 - r_s/r)
	RedshiftFactor    float64       // z_gr ≈ 1/sqrt(1 - r_s/r)
	TidalSeverity     float64       // 0..1 heuristic
	DangerLevel       GRDangerLevel
	CanHoverSafely    bool
	NearPhotonSphere  bool
	Rs                float64 // Schwarzschild radius for reference
}

// GRShaderUniforms contains data passed to GR shaders.
type GRShaderUniforms struct {
	Enabled         bool
	ObjectKind      int       // 0=BH, 1=NS, 2=WD
	Rs              float32   // Schwarzschild radius (screen units)
	Distance        float32   // Ship→object distance
	Phi             float32   // Dimensionless potential
	TimeDilation    float32   // dτ/dt
	RedshiftFactor  float32   // Gravitational z
	ScreenCenter    [2]float32 // Screen position of object center
	MaxEffectRadius float32   // Effect bounds in screen units
	LensStrength    float32   // Tunable lensing parameter
}

// NewMassiveObject creates a MassiveObject with computed Schwarzschild radius.
// massSolar is in solar masses.
func NewMassiveObject(kind MassiveObjectKind, massSolar float64, pos Vec3) MassiveObject {
	// r_s = 2GM/c^2 ≈ 2.95 km per solar mass
	// We use game units where 1 unit ≈ 1000 km
	rsKm := 2.95 * massSolar
	rs := rsKm / 1000.0 // Convert to game units

	return MassiveObject{
		Kind:                kind,
		MassSolar:           massSolar,
		SchwarzschildRadius: rs,
		Position:            pos,
	}
}

// NewBlackHole creates a black hole with given solar mass.
func NewBlackHole(massSolar float64, pos Vec3) MassiveObject {
	return NewMassiveObject(BlackHole, massSolar, pos)
}

// NewNeutronStar creates a neutron star. Surface radius typically ~10km.
func NewNeutronStar(massSolar float64, radiusKm float64, pos Vec3) MassiveObject {
	obj := NewMassiveObject(NeutronStar, massSolar, pos)
	obj.SurfaceRadius = radiusKm / 1000.0
	return obj
}

// ClassifyDangerLevel determines GR severity from dimensionless potential.
func ClassifyDangerLevel(phi float64) GRDangerLevel {
	switch {
	case phi < 1e-4:
		return GR_None
	case phi < 1e-3:
		return GR_Subtle
	case phi < 1e-2:
		return GR_Strong
	default:
		return GR_Extreme
	}
}

// ComputeGRContext calculates GR quantities for ship near massive object.
func ComputeGRContext(shipPos Vec3, obj MassiveObject) GRContext {
	// Distance from ship to object center
	r := shipPos.Sub(obj.Position).Length()
	rs := obj.SchwarzschildRadius

	// Avoid division by zero / inside horizon
	if r < rs*1.01 {
		r = rs * 1.01
	}

	// Dimensionless potential: Φ = r_s / (2r)
	phi := rs / (2.0 * r)

	// Time dilation: dτ/dt = sqrt(1 - r_s/r)
	// Clamp argument for numerical stability
	tdArg := clamp(1.0-rs/r, 0.001, 1.0)
	timeDilation := math.Sqrt(tdArg)

	// Redshift factor: z ≈ 1/sqrt(1 - r_s/r)
	redshiftFactor := 1.0 / timeDilation

	// Tidal severity heuristic: stronger for smaller mass at same Φ
	// Tidal acceleration ~ GM/r^3 ~ r_s/r^3
	tidalSeverity := clamp(rs/(r*r)*1e6, 0.0, 1.0)

	// Danger level
	dangerLevel := ClassifyDangerLevel(phi)

	// Near photon sphere (only meaningful for black holes)
	nearPhotonSphere := false
	if obj.Kind == BlackHole {
		// Photon sphere at r = 1.5 r_s
		nearPhotonSphere = r >= 1.3*rs && r <= 2.0*rs
	}

	return GRContext{
		Active:           phi >= 1e-4,
		ObjectKind:       obj.Kind,
		Distance:         r,
		Phi:              phi,
		TimeDilation:     timeDilation,
		RedshiftFactor:   redshiftFactor,
		TidalSeverity:    tidalSeverity,
		DangerLevel:      dangerLevel,
		CanHoverSafely:   tidalSeverity < 0.5,
		NearPhotonSphere: nearPhotonSphere,
		Rs:               rs,
	}
}

// ToShaderUniforms converts GRContext to shader-compatible uniforms.
// screenCenterX, screenCenterY: object position in screen coordinates (0-1 range).
// screenRs: Schwarzschild radius in screen units.
func (ctx *GRContext) ToShaderUniforms(screenCenterX, screenCenterY, screenRs float32) GRShaderUniforms {
	if !ctx.Active {
		return GRShaderUniforms{Enabled: false}
	}

	// Max effect radius scales with danger level
	var maxRadius float32
	switch ctx.DangerLevel {
	case GR_Subtle:
		maxRadius = screenRs * 20
	case GR_Strong:
		maxRadius = screenRs * 50
	case GR_Extreme:
		maxRadius = screenRs * 100
	default:
		maxRadius = screenRs * 10
	}

	// Lens strength scales with Φ
	lensStrength := float32(ctx.Phi * 100)

	return GRShaderUniforms{
		Enabled:         true,
		ObjectKind:      int(ctx.ObjectKind),
		Rs:              screenRs,
		Distance:        float32(ctx.Distance),
		Phi:             float32(ctx.Phi),
		TimeDilation:    float32(ctx.TimeDilation),
		RedshiftFactor:  float32(ctx.RedshiftFactor),
		ScreenCenter:    [2]float32{screenCenterX, screenCenterY},
		MaxEffectRadius: maxRadius,
		LensStrength:    lensStrength,
	}
}

// DisabledGRContext returns a GRContext with Active=false.
func DisabledGRContext() GRContext {
	return GRContext{Active: false}
}
