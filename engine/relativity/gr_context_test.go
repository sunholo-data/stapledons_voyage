package relativity

import (
	"math"
	"testing"
)

func TestClassifyDangerLevel(t *testing.T) {
	tests := []struct {
		phi      float64
		expected GRDangerLevel
	}{
		{0.00001, GR_None},    // Below threshold
		{0.0001, GR_Subtle},   // At Subtle boundary
		{0.0005, GR_Subtle},   // In Subtle range
		{0.001, GR_Strong},    // At Strong boundary
		{0.005, GR_Strong},    // In Strong range
		{0.01, GR_Extreme},    // At Extreme boundary
		{0.1, GR_Extreme},     // Deep Extreme
	}

	for _, tt := range tests {
		result := ClassifyDangerLevel(tt.phi)
		if result != tt.expected {
			t.Errorf("ClassifyDangerLevel(%v) = %v, want %v", tt.phi, result, tt.expected)
		}
	}
}

func TestNewBlackHole(t *testing.T) {
	bh := NewBlackHole(10.0, Vec3{0, 0, 0})

	if bh.Kind != BlackHole {
		t.Errorf("Kind = %v, want BlackHole", bh.Kind)
	}

	if bh.MassSolar != 10.0 {
		t.Errorf("MassSolar = %v, want 10.0", bh.MassSolar)
	}

	// r_s = 2.95 km * 10 = 29.5 km = 0.0295 game units
	expectedRs := 2.95 * 10 / 1000.0
	if math.Abs(bh.SchwarzschildRadius-expectedRs) > 0.0001 {
		t.Errorf("SchwarzschildRadius = %v, want ~%v", bh.SchwarzschildRadius, expectedRs)
	}
}

func TestComputeGRContext_FarAway(t *testing.T) {
	bh := NewBlackHole(10.0, Vec3{0, 0, 0})
	shipPos := Vec3{1000, 0, 0} // Very far away

	ctx := ComputeGRContext(shipPos, bh)

	if ctx.Active {
		t.Error("Expected Active=false for far distance")
	}

	if ctx.DangerLevel != GR_None {
		t.Errorf("DangerLevel = %v, want GR_None", ctx.DangerLevel)
	}

	// Time dilation should be very close to 1.0
	if ctx.TimeDilation < 0.999 {
		t.Errorf("TimeDilation = %v, expected ~1.0 at far distance", ctx.TimeDilation)
	}
}

func TestComputeGRContext_CloseApproach(t *testing.T) {
	// Supermassive black hole: 4 million solar masses (like Sgr A*)
	bh := NewBlackHole(4e6, Vec3{0, 0, 0})

	// Ship at 10 Schwarzschild radii
	rs := bh.SchwarzschildRadius
	shipPos := Vec3{rs * 10, 0, 0}

	ctx := ComputeGRContext(shipPos, bh)

	if !ctx.Active {
		t.Error("Expected Active=true at 10 r_s")
	}

	// Phi = r_s / (2r) = 1 / 20 = 0.05
	expectedPhi := 0.05
	if math.Abs(ctx.Phi-expectedPhi) > 0.001 {
		t.Errorf("Phi = %v, want ~%v", ctx.Phi, expectedPhi)
	}

	if ctx.DangerLevel != GR_Extreme {
		t.Errorf("DangerLevel = %v, want GR_Extreme at Phi=0.05", ctx.DangerLevel)
	}

	// Time dilation at r = 10 r_s: sqrt(1 - 0.1) ≈ 0.949
	expectedTD := math.Sqrt(0.9)
	if math.Abs(ctx.TimeDilation-expectedTD) > 0.01 {
		t.Errorf("TimeDilation = %v, want ~%v", ctx.TimeDilation, expectedTD)
	}

	// Redshift factor ≈ 1.054
	expectedRS := 1.0 / expectedTD
	if math.Abs(ctx.RedshiftFactor-expectedRS) > 0.01 {
		t.Errorf("RedshiftFactor = %v, want ~%v", ctx.RedshiftFactor, expectedRS)
	}
}

func TestComputeGRContext_PhotonSphere(t *testing.T) {
	bh := NewBlackHole(10.0, Vec3{0, 0, 0})
	rs := bh.SchwarzschildRadius

	// At 1.5 r_s (photon sphere)
	shipPos := Vec3{rs * 1.5, 0, 0}
	ctx := ComputeGRContext(shipPos, bh)

	if !ctx.NearPhotonSphere {
		t.Error("Expected NearPhotonSphere=true at 1.5 r_s")
	}

	// At 3 r_s (outside photon sphere region)
	shipPos = Vec3{rs * 3, 0, 0}
	ctx = ComputeGRContext(shipPos, bh)

	if ctx.NearPhotonSphere {
		t.Error("Expected NearPhotonSphere=false at 3 r_s")
	}
}

func TestComputeGRContext_NeutronStar(t *testing.T) {
	// 1.4 solar mass neutron star
	ns := NewNeutronStar(1.4, 10.0, Vec3{0, 0, 0})

	if ns.Kind != NeutronStar {
		t.Errorf("Kind = %v, want NeutronStar", ns.Kind)
	}

	// Close approach to surface
	shipPos := Vec3{0.03, 0, 0} // 30 km
	ctx := ComputeGRContext(shipPos, ns)

	// NS shouldn't trigger photon sphere
	if ctx.NearPhotonSphere {
		t.Error("NeutronStar shouldn't have NearPhotonSphere=true")
	}
}

func TestToShaderUniforms(t *testing.T) {
	bh := NewBlackHole(4e6, Vec3{0, 0, 0})
	rs := bh.SchwarzschildRadius
	shipPos := Vec3{rs * 10, 0, 0}

	ctx := ComputeGRContext(shipPos, bh)
	uniforms := ctx.ToShaderUniforms(0.5, 0.5, 0.1)

	if !uniforms.Enabled {
		t.Error("Expected Enabled=true")
	}

	if uniforms.ScreenCenter[0] != 0.5 || uniforms.ScreenCenter[1] != 0.5 {
		t.Errorf("ScreenCenter = %v, want [0.5, 0.5]", uniforms.ScreenCenter)
	}

	if uniforms.MaxEffectRadius <= 0 {
		t.Error("MaxEffectRadius should be positive")
	}
}

func TestDisabledGRContext(t *testing.T) {
	ctx := DisabledGRContext()

	if ctx.Active {
		t.Error("DisabledGRContext should have Active=false")
	}

	uniforms := ctx.ToShaderUniforms(0.5, 0.5, 0.1)
	if uniforms.Enabled {
		t.Error("Disabled context should produce disabled uniforms")
	}
}

func TestMassiveObjectKind_String(t *testing.T) {
	if BlackHole.String() != "BlackHole" {
		t.Errorf("BlackHole.String() = %v", BlackHole.String())
	}
	if NeutronStar.String() != "NeutronStar" {
		t.Errorf("NeutronStar.String() = %v", NeutronStar.String())
	}
	if WhiteDwarf.String() != "WhiteDwarf" {
		t.Errorf("WhiteDwarf.String() = %v", WhiteDwarf.String())
	}
}

func TestGRDangerLevel_String(t *testing.T) {
	levels := []struct {
		level GRDangerLevel
		str   string
	}{
		{GR_None, "None"},
		{GR_Subtle, "Subtle"},
		{GR_Strong, "Strong"},
		{GR_Extreme, "Extreme"},
	}

	for _, tt := range levels {
		if tt.level.String() != tt.str {
			t.Errorf("%v.String() = %v, want %v", tt.level, tt.level.String(), tt.str)
		}
	}
}
