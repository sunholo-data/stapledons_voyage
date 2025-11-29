package camera

import (
	"math"
	"testing"

	"stapledons_voyage/sim_gen"
)

func TestWorldToScreen(t *testing.T) {
	// Camera at center of 512x512 world, zoom 1.0
	cam := sim_gen.Camera{X: 256, Y: 256, Zoom: 1.0}
	transform := FromOutput(cam, 640, 480)

	// World center should map to screen center
	sx, sy := transform.WorldToScreen(256, 256)
	if sx != 320 || sy != 240 {
		t.Errorf("World center (256,256) should map to screen center (320,240), got (%v,%v)", sx, sy)
	}

	// World origin should be offset
	sx, sy = transform.WorldToScreen(0, 0)
	expectedX := 320.0 - 256.0 // 64
	expectedY := 240.0 - 256.0 // -16
	if sx != expectedX || sy != expectedY {
		t.Errorf("World origin (0,0) should map to (%v,%v), got (%v,%v)", expectedX, expectedY, sx, sy)
	}
}

func TestScreenToWorld(t *testing.T) {
	cam := sim_gen.Camera{X: 256, Y: 256, Zoom: 1.0}
	transform := FromOutput(cam, 640, 480)

	// Screen center should map to camera position
	wx, wy := transform.ScreenToWorld(320, 240)
	if wx != 256 || wy != 256 {
		t.Errorf("Screen center (320,240) should map to camera pos (256,256), got (%v,%v)", wx, wy)
	}
}

func TestRoundTrip(t *testing.T) {
	cam := sim_gen.Camera{X: 100, Y: 200, Zoom: 1.5}
	transform := FromOutput(cam, 800, 600)

	// Any world point should round-trip correctly
	origX, origY := 150.0, 250.0
	sx, sy := transform.WorldToScreen(origX, origY)
	wx, wy := transform.ScreenToWorld(sx, sy)

	if math.Abs(wx-origX) > 0.001 || math.Abs(wy-origY) > 0.001 {
		t.Errorf("Round trip failed: (%v,%v) -> (%v,%v) -> (%v,%v)", origX, origY, sx, sy, wx, wy)
	}
}

func TestViewportContains(t *testing.T) {
	cam := sim_gen.Camera{X: 256, Y: 256, Zoom: 1.0}
	vp := CalculateViewport(cam, 640, 480)

	// Camera center should be inside
	if !vp.Contains(256, 256, 0) {
		t.Error("Viewport should contain camera center")
	}

	// Far outside should not be inside
	if vp.Contains(1000, 1000, 0) {
		t.Error("Viewport should not contain far outside point")
	}
}

func TestViewportContainsRect(t *testing.T) {
	cam := sim_gen.Camera{X: 256, Y: 256, Zoom: 1.0}
	vp := CalculateViewport(cam, 640, 480)

	// Rect at camera center should be visible
	if !vp.ContainsRect(250, 250, 16, 16) {
		t.Error("Rect at center should be visible")
	}

	// Rect far outside should not be visible
	if vp.ContainsRect(1000, 1000, 16, 16) {
		t.Error("Rect far outside should not be visible")
	}
}
