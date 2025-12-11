//go:build !ci

package view

import (
	"testing"
)

func TestViewTypeString(t *testing.T) {
	tests := []struct {
		vt   ViewType
		want string
	}{
		{ViewNone, "None"},
		{ViewSpace, "Space"},
		{ViewBridge, "Bridge"},
		{ViewShip, "Ship"},
		{ViewGalaxyMap, "GalaxyMap"},
		{ViewPlanetSurface, "PlanetSurface"},
		{ViewArrival, "Arrival"},
	}

	for _, tt := range tests {
		if got := tt.vt.String(); got != tt.want {
			t.Errorf("ViewType(%d).String() = %q, want %q", tt.vt, got, tt.want)
		}
	}
}

func TestTransitionEffectString(t *testing.T) {
	tests := []struct {
		te   TransitionEffect
		want string
	}{
		{TransitionNone, "None"},
		{TransitionFade, "Fade"},
		{TransitionCrossfade, "Crossfade"},
		{TransitionWipe, "Wipe"},
		{TransitionZoom, "Zoom"},
	}

	for _, tt := range tests {
		if got := tt.te.String(); got != tt.want {
			t.Errorf("TransitionEffect(%d).String() = %q, want %q", tt.te, got, tt.want)
		}
	}
}

// NOTE: Manager tests removed - Manager was deleted as part of Phase 0 architecture cleanup

func TestComputePanelBounds(t *testing.T) {
	screenW, screenH := 1280.0, 960.0

	tests := []struct {
		name   string
		panel  *UIPanel
		wantX  float64
		wantY  float64
	}{
		{
			name:   "TopLeft",
			panel:  &UIPanel{X: 10, Y: 20, W: 100, H: 50, Anchor: AnchorTopLeft},
			wantX:  10,
			wantY:  20,
		},
		{
			name:   "Center",
			panel:  &UIPanel{X: 0, Y: 0, W: 200, H: 100, Anchor: AnchorCenter},
			wantX:  (screenW - 200) / 2,
			wantY:  (screenH - 100) / 2,
		},
		{
			name:   "BottomRight",
			panel:  &UIPanel{X: 10, Y: 20, W: 100, H: 50, Anchor: AnchorBottomRight},
			wantX:  screenW - 100 - 10,
			wantY:  screenH - 50 - 20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bounds := ComputePanelBounds(tt.panel, screenW, screenH)
			if bounds.X != tt.wantX {
				t.Errorf("X = %v, want %v", bounds.X, tt.wantX)
			}
			if bounds.Y != tt.wantY {
				t.Errorf("Y = %v, want %v", bounds.Y, tt.wantY)
			}
		})
	}
}

func TestRectContains(t *testing.T) {
	r := Rect{X: 10, Y: 20, W: 100, H: 50}

	// Inside
	if !r.Contains(50, 40) {
		t.Error("point (50, 40) should be inside")
	}

	// On edge (top-left)
	if !r.Contains(10, 20) {
		t.Error("point (10, 20) should be inside (inclusive)")
	}

	// Outside
	if r.Contains(5, 40) {
		t.Error("point (5, 40) should be outside")
	}
	if r.Contains(50, 75) {
		t.Error("point (50, 75) should be outside")
	}
}

func TestNewTransition(t *testing.T) {
	trans := NewTransition(ViewBridge, 0.5, TransitionFade)

	if trans.To != ViewBridge {
		t.Errorf("To = %v, want ViewBridge", trans.To)
	}
	if trans.Duration != 0.5 {
		t.Errorf("Duration = %v, want 0.5", trans.Duration)
	}
	if trans.Effect != TransitionFade {
		t.Errorf("Effect = %v, want TransitionFade", trans.Effect)
	}
}
