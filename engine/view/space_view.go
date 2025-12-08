package view

import (
	"github.com/hajimehoshi/ebiten/v2"
	"stapledons_voyage/engine/display"
	"stapledons_voyage/engine/view/background"
)

// SpaceView is a view showing the exterior space with starfield background.
// This is used for arrival sequences, solar system views, and exterior shots.
type SpaceView struct {
	background *background.SpaceBackground
	camera     *Camera
	screenW    int
	screenH    int

	// State
	initialized bool
	velocity    float64 // For SR effects
	grIntensity float64 // For GR effects

	// UI elements
	uiPanels []*UIPanel
}

// NewSpaceView creates a new space view.
func NewSpaceView() *SpaceView {
	return &SpaceView{
		screenW: display.InternalWidth,
		screenH: display.InternalHeight,
		camera:  NewCamera(),
	}
}

// Type returns ViewSpace.
func (v *SpaceView) Type() ViewType {
	return ViewSpace
}

// Init initializes the view.
func (v *SpaceView) Init() error {
	if v.initialized {
		return nil
	}

	v.background = background.NewSpaceBackground(v.screenW, v.screenH)
	v.initialized = true
	return nil
}

// Enter is called when transitioning into this view.
func (v *SpaceView) Enter(from ViewType) {
	// Reset camera to origin
	v.camera.X = 0
	v.camera.Y = 0
	v.camera.Zoom = 1.0
}

// Exit is called when transitioning out of this view.
func (v *SpaceView) Exit(to ViewType) {
	// Nothing to clean up
}

// Update updates the view state.
func (v *SpaceView) Update(dt float64) *ViewTransition {
	// No automatic transitions
	return nil
}

// Draw renders the view to the screen.
func (v *SpaceView) Draw(screen *ebiten.Image) {
	if v.background != nil {
		// Convert view.Camera to background.CameraOffset
		camOffset := &background.CameraOffset{
			X:    v.camera.X,
			Y:    v.camera.Y,
			Zoom: v.camera.Zoom,
		}
		v.background.Draw(screen, camOffset)
	}

	// Draw UI panels
	screenW := float64(screen.Bounds().Dx())
	screenH := float64(screen.Bounds().Dy())

	for _, panel := range v.uiPanels {
		if !panel.Visible {
			continue
		}
		bounds := ComputePanelBounds(panel, screenW, screenH)
		if panel.DrawFunc != nil {
			panel.DrawFunc(screen, bounds)
		}
	}
}

// Layers returns the view's layer components.
// Note: SpaceView manages its background internally, so Background is nil here.
// Use GetBackground() for direct access to the SpaceBackground.
func (v *SpaceView) Layers() ViewLayers {
	return ViewLayers{
		Background: nil, // SpaceView manages background internally
		Content:    nil,
		UI:         nil,
	}
}

// GetBackground returns the SpaceBackground for direct access.
func (v *SpaceView) GetBackground() *background.SpaceBackground {
	return v.background
}

// SetVelocity sets the ship velocity for SR effects.
func (v *SpaceView) SetVelocity(velocity float64) {
	v.velocity = velocity
	if v.background != nil {
		v.background.SetVelocity(velocity)
	}
}

// SetGRIntensity sets the GR intensity for lensing effects.
func (v *SpaceView) SetGRIntensity(intensity float64) {
	v.grIntensity = intensity
	if v.background != nil {
		v.background.SetGRIntensity(intensity)
	}
}

// GetVelocity returns the current velocity.
func (v *SpaceView) GetVelocity() float64 {
	return v.velocity
}

// GetGRIntensity returns the current GR intensity.
func (v *SpaceView) GetGRIntensity() float64 {
	return v.grIntensity
}

// SetCamera updates the camera position.
func (v *SpaceView) SetCamera(x, y, zoom float64) {
	v.camera.X = x
	v.camera.Y = y
	v.camera.Zoom = zoom
}

// AddUIPanel adds a UI panel to the view.
func (v *SpaceView) AddUIPanel(panel *UIPanel) {
	v.uiPanels = append(v.uiPanels, panel)
}

// RemoveUIPanel removes a panel by ID.
func (v *SpaceView) RemoveUIPanel(id string) {
	for i, p := range v.uiPanels {
		if p.ID == id {
			v.uiPanels = append(v.uiPanels[:i], v.uiPanels[i+1:]...)
			return
		}
	}
}

// Resize updates the view for new screen dimensions.
func (v *SpaceView) Resize(screenW, screenH int) {
	v.screenW = screenW
	v.screenH = screenH
	if v.background != nil {
		v.background.Resize(screenW, screenH)
	}
}
