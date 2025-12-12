package camera

import (
	"math"
	"testing"

	"stapledons_voyage/engine/depth"
)

func TestLayerParallaxFactors(t *testing.T) {
	// Verify expected parallax factors
	tests := []struct {
		layer    depth.Layer
		expected float64
	}{
		{depth.LayerDeepBackground, 0.1},
		{depth.LayerMidBackground, 0.3},
		{depth.LayerScene, 1.0},
		{depth.LayerForeground, 1.0},
	}

	for _, tc := range tests {
		got := tc.layer.Parallax()
		if got != tc.expected {
			t.Errorf("Layer %s: got parallax %v, want %v", tc.layer.Name(), got, tc.expected)
		}
	}
}

func TestParallaxCameraForLayer(t *testing.T) {
	cam := NewParallaxCamera(1280, 960)
	cam.SetPosition(100, 200)

	// Test DeepBackground layer (0.1x)
	x, y := cam.ForLayer(depth.LayerDeepBackground)
	if x != 10 || y != 20 {
		t.Errorf("DeepBackground: got (%v, %v), want (10, 20)", x, y)
	}

	// Test MidBackground layer (0.3x)
	x, y = cam.ForLayer(depth.LayerMidBackground)
	if x != 30 || y != 60 {
		t.Errorf("MidBackground: got (%v, %v), want (30, 60)", x, y)
	}

	// Test Scene layer (1.0x)
	x, y = cam.ForLayer(depth.LayerScene)
	if x != 100 || y != 200 {
		t.Errorf("Scene: got (%v, %v), want (100, 200)", x, y)
	}
}

func TestParallaxCameraTransformForLayer(t *testing.T) {
	cam := NewParallaxCamera(1280, 960)
	cam.SetPosition(100, 0)
	cam.SetZoom(1.0)

	// At zoom 1.0, camera at (100, 0)
	// Scene layer should center at screen center offset by camera position
	transform := cam.TransformForLayer(depth.LayerScene)

	// OffsetX = screenW/2 - camX*zoom = 640 - 100*1 = 540
	expectedOffsetX := 540.0
	if transform.OffsetX != expectedOffsetX {
		t.Errorf("Scene OffsetX: got %v, want %v", transform.OffsetX, expectedOffsetX)
	}

	// DeepBackground layer moves 0.1x, so effective camera is (10, 0)
	// OffsetX = 640 - 10*1 = 630
	bgTransform := cam.TransformForLayer(depth.LayerDeepBackground)
	expectedBgOffsetX := 630.0
	if bgTransform.OffsetX != expectedBgOffsetX {
		t.Errorf("DeepBackground OffsetX: got %v, want %v", bgTransform.OffsetX, expectedBgOffsetX)
	}
}

func TestParallaxCameraWorldToScreen(t *testing.T) {
	cam := NewParallaxCamera(1280, 960)
	cam.SetPosition(0, 0)
	cam.SetZoom(1.0)

	// At origin, world (0,0) should map to screen center (640, 480)
	sx, sy := cam.WorldToScreen(0, 0, depth.LayerScene)
	if sx != 640 || sy != 480 {
		t.Errorf("WorldToScreen(0,0): got (%v, %v), want (640, 480)", sx, sy)
	}

	// World (100, 100) should map to screen (740, 580)
	sx, sy = cam.WorldToScreen(100, 100, depth.LayerScene)
	if sx != 740 || sy != 580 {
		t.Errorf("WorldToScreen(100,100): got (%v, %v), want (740, 580)", sx, sy)
	}
}

func TestParallaxCameraScreenToWorld(t *testing.T) {
	cam := NewParallaxCamera(1280, 960)
	cam.SetPosition(0, 0)
	cam.SetZoom(1.0)

	// Screen center (640, 480) should map to world (0, 0)
	wx, wy := cam.ScreenToWorld(640, 480)
	if wx != 0 || wy != 0 {
		t.Errorf("ScreenToWorld(640,480): got (%v, %v), want (0, 0)", wx, wy)
	}
}

func TestParallaxCameraWithZoom(t *testing.T) {
	cam := NewParallaxCamera(1280, 960)
	cam.SetPosition(0, 0)
	cam.SetZoom(2.0) // 2x zoom

	// At 2x zoom, world (50, 50) should appear at screen (740, 580)
	// because 50*2 = 100 pixels from center
	sx, sy := cam.WorldToScreen(50, 50, depth.LayerScene)
	if sx != 740 || sy != 580 {
		t.Errorf("WorldToScreen at 2x zoom: got (%v, %v), want (740, 580)", sx, sy)
	}
}

func floatEquals(a, b, epsilon float64) bool {
	return math.Abs(a-b) < epsilon
}

func TestLayerNames(t *testing.T) {
	names := []struct {
		layer    depth.Layer
		expected string
	}{
		{depth.LayerDeepBackground, "DeepBackground"},
		{depth.LayerMidBackground, "MidBackground"},
		{depth.LayerScene, "Scene"},
		{depth.LayerForeground, "Foreground"},
	}

	for _, tc := range names {
		got := tc.layer.Name()
		if got != tc.expected {
			t.Errorf("Layer.Name(): got %q, want %q", got, tc.expected)
		}
	}
}
