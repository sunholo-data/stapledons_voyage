//go:build !ci

package view

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// mockView is a simple test view implementation.
type mockView struct {
	viewType    ViewType
	initialized bool
	entered     bool
	exited      bool
	updateCount int
	drawCount   int
}

func newMockView(vt ViewType) *mockView {
	return &mockView{viewType: vt}
}

func (m *mockView) Type() ViewType { return m.viewType }
func (m *mockView) Init() error    { m.initialized = true; return nil }
func (m *mockView) Enter(from ViewType) { m.entered = true }
func (m *mockView) Exit(to ViewType)    { m.exited = true }
func (m *mockView) Update(dt float64) *ViewTransition {
	m.updateCount++
	return nil
}
func (m *mockView) Draw(screen *ebiten.Image) { m.drawCount++ }
func (m *mockView) Layers() ViewLayers        { return ViewLayers{} }

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

func TestManagerRegisterAndSetCurrent(t *testing.T) {
	mgr := NewManager(1280, 960)

	// Create mock views
	space := newMockView(ViewSpace)
	bridge := newMockView(ViewBridge)

	// Register views
	if err := mgr.Register(space); err != nil {
		t.Fatalf("Register(space) failed: %v", err)
	}
	if err := mgr.Register(bridge); err != nil {
		t.Fatalf("Register(bridge) failed: %v", err)
	}

	// Set current view
	if err := mgr.SetCurrent(ViewSpace); err != nil {
		t.Fatalf("SetCurrent(ViewSpace) failed: %v", err)
	}

	// Verify initialization and enter
	if !space.initialized {
		t.Error("space.Init() was not called")
	}
	if !space.entered {
		t.Error("space.Enter() was not called")
	}

	// Verify current
	if mgr.Current() != space {
		t.Error("Current() should return space view")
	}

	// Switch to bridge
	if err := mgr.SetCurrent(ViewBridge); err != nil {
		t.Fatalf("SetCurrent(ViewBridge) failed: %v", err)
	}

	// Verify exit and enter
	if !space.exited {
		t.Error("space.Exit() was not called")
	}
	if !bridge.entered {
		t.Error("bridge.Enter() was not called")
	}
}

func TestManagerErrors(t *testing.T) {
	mgr := NewManager(1280, 960)

	// Register nil view should fail
	if err := mgr.Register(nil); err == nil {
		t.Error("Register(nil) should return error")
	}

	// SetCurrent for unregistered view should fail
	if err := mgr.SetCurrent(ViewSpace); err == nil {
		t.Error("SetCurrent for unregistered view should return error")
	}
}

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
