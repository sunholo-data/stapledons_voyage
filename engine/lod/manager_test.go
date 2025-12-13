package lod

import (
	"image/color"
	"testing"
)

func TestCalcTier(t *testing.T) {
	m := NewManager(DefaultConfig())

	tests := []struct {
		distance float64
		expected LODTier
	}{
		{10, TierFull3D},      // < 50
		{49, TierFull3D},      // < 50
		{51, TierBillboard},   // >= 50, < 200
		{199, TierBillboard},  // < 200
		{201, TierCircle},     // >= 200, < 1000
		{999, TierCircle},     // < 1000
		{1001, TierPoint},     // >= 1000, < 10000
		{9999, TierPoint},     // < 10000
		{10001, TierCulled},   // >= 10000
		{50000, TierCulled},
	}

	for _, tt := range tests {
		tier := m.calcTier(tt.distance)
		if tier != tt.expected {
			t.Errorf("calcTier(%v) = %v, want %v", tt.distance, tier, tt.expected)
		}
	}
}

func TestAddRemove(t *testing.T) {
	m := NewManager(DefaultConfig())

	obj1 := NewObject("star1", Vector3{0, 0, 0}, 1.0, color.RGBA{255, 255, 255, 255})
	obj2 := NewObject("star2", Vector3{100, 0, 0}, 1.0, color.RGBA{255, 200, 200, 255})

	m.Add(obj1)
	m.Add(obj2)

	if m.ObjectCount() != 2 {
		t.Errorf("ObjectCount() = %d, want 2", m.ObjectCount())
	}

	m.Remove("star1")
	if m.ObjectCount() != 1 {
		t.Errorf("After Remove, ObjectCount() = %d, want 1", m.ObjectCount())
	}

	if m.GetObject("star1") != nil {
		t.Error("GetObject('star1') should return nil after removal")
	}
	if m.GetObject("star2") == nil {
		t.Error("GetObject('star2') should not return nil")
	}

	m.Clear()
	if m.ObjectCount() != 0 {
		t.Errorf("After Clear, ObjectCount() = %d, want 0", m.ObjectCount())
	}
}

func TestUpdate(t *testing.T) {
	m := NewManager(DefaultConfig())
	camera := NewSimpleCamera(800, 600)
	camera.Pos = Vector3{0, 0, 100}
	camera.LookAt = Vector3{0, 0, 0}

	// Object at origin - should be Full3D (distance ~100, but wait that's > 50)
	// Let's place objects at different distances from camera
	m.Add(NewObject("close", Vector3{0, 0, 80}, 1.0, color.RGBA{255, 0, 0, 255}))    // Distance 20 -> Full3D
	m.Add(NewObject("medium", Vector3{0, 0, -50}, 1.0, color.RGBA{0, 255, 0, 255}))  // Distance 150 -> Billboard
	m.Add(NewObject("far", Vector3{0, 0, -400}, 1.0, color.RGBA{0, 0, 255, 255}))    // Distance 500 -> Circle
	m.Add(NewObject("vfar", Vector3{0, 0, -5000}, 1.0, color.RGBA{255, 255, 0, 255})) // Distance 5100 -> Point

	m.Update(camera)

	stats := m.Stats()
	if stats.TotalObjects != 4 {
		t.Errorf("TotalObjects = %d, want 4", stats.TotalObjects)
	}

	// Check tier assignments
	close := m.GetObject("close")
	if close.CurrentTier != TierFull3D {
		t.Errorf("'close' tier = %v, want Full3D", close.CurrentTier)
	}

	medium := m.GetObject("medium")
	if medium.CurrentTier != TierBillboard {
		t.Errorf("'medium' tier = %v, want Billboard", medium.CurrentTier)
	}

	far := m.GetObject("far")
	if far.CurrentTier != TierCircle {
		t.Errorf("'far' tier = %v, want Circle", far.CurrentTier)
	}

	vfar := m.GetObject("vfar")
	if vfar.CurrentTier != TierPoint {
		t.Errorf("'vfar' tier = %v, want Point", vfar.CurrentTier)
	}
}

func TestMax3DObjects(t *testing.T) {
	config := DefaultConfig()
	config.Max3DObjects = 2
	m := NewManager(config)
	camera := NewSimpleCamera(800, 600)
	camera.Pos = Vector3{0, 0, 100}
	camera.LookAt = Vector3{0, 0, 0}

	// Add 5 objects all close enough for Full3D
	for i := 0; i < 5; i++ {
		z := 70 + float64(i)*5 // Distances: 30, 25, 20, 15, 10 from camera
		m.Add(NewObject("star"+string(rune('A'+i)), Vector3{0, 0, z}, 1.0, color.RGBA{255, 255, 255, 255}))
	}

	m.Update(camera)

	stats := m.Stats()
	if stats.Full3DCount != 2 {
		t.Errorf("Full3DCount = %d, want 2 (limited by Max3DObjects)", stats.Full3DCount)
	}

	// The extra 3 should be demoted to Billboard
	if stats.BillboardCount != 3 {
		t.Errorf("BillboardCount = %d, want 3 (demoted from Full3D)", stats.BillboardCount)
	}
}

func TestTierNames(t *testing.T) {
	tests := []struct {
		tier LODTier
		name string
	}{
		{TierFull3D, "Full3D"},
		{TierBillboard, "Billboard"},
		{TierCircle, "Circle"},
		{TierPoint, "Point"},
		{TierCulled, "Culled"},
	}

	for _, tt := range tests {
		if tt.tier.String() != tt.name {
			t.Errorf("LODTier(%d).String() = %s, want %s", tt.tier, tt.tier.String(), tt.name)
		}
	}
}
