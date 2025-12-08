package view

import (
	"fmt"
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

// Manager coordinates views and handles transitions between them.
type Manager struct {
	views       map[ViewType]View
	current     View
	next        View
	transition  *activeTransition
	screenW     int
	screenH     int
	initialized bool
}

// activeTransition holds the state of an in-progress transition.
type activeTransition struct {
	from     View
	to       View
	effect   TransitionEffect
	easing   EasingFunc // Easing function for smooth transitions
	duration float64
	elapsed  float64
	buffer   *ebiten.Image // Render buffer for transition effects
}

// NewManager creates a new view manager.
func NewManager(screenW, screenH int) *Manager {
	return &Manager{
		views:   make(map[ViewType]View),
		screenW: screenW,
		screenH: screenH,
	}
}

// Register registers a view with the manager.
// The view's Init() method will be called when it's first used.
func (m *Manager) Register(view View) error {
	if view == nil {
		return fmt.Errorf("cannot register nil view")
	}
	m.views[view.Type()] = view
	return nil
}

// SetCurrent sets the current view without a transition.
// The view must be registered first.
func (m *Manager) SetCurrent(viewType ViewType) error {
	view, ok := m.views[viewType]
	if !ok {
		return fmt.Errorf("view not registered: %s", viewType)
	}

	// Initialize if needed
	if !m.initialized {
		if err := view.Init(); err != nil {
			return fmt.Errorf("failed to init view %s: %w", viewType, err)
		}
	}

	// Handle exit/enter
	if m.current != nil {
		m.current.Exit(viewType)
	}

	oldType := ViewNone
	if m.current != nil {
		oldType = m.current.Type()
	}

	m.current = view
	m.current.Enter(oldType)
	m.initialized = true

	return nil
}

// Current returns the current active view.
func (m *Manager) Current() View {
	return m.current
}

// TransitionTo starts a transition to another view.
func (m *Manager) TransitionTo(trans *ViewTransition) error {
	if trans == nil {
		return fmt.Errorf("transition cannot be nil")
	}

	view, ok := m.views[trans.To]
	if !ok {
		return fmt.Errorf("view not registered: %s", trans.To)
	}

	// Initialize the target view if needed
	if err := view.Init(); err != nil {
		return fmt.Errorf("failed to init view %s: %w", trans.To, err)
	}

	// If no duration or no effect, just switch immediately
	if trans.Duration <= 0 || trans.Effect == TransitionNone {
		return m.SetCurrent(trans.To)
	}

	// Start the transition
	m.next = view
	m.transition = &activeTransition{
		from:     m.current,
		to:       view,
		effect:   trans.Effect,
		easing:   EaseInOutQuad, // Default easing
		duration: trans.Duration,
		elapsed:  0,
		buffer:   ebiten.NewImage(m.screenW, m.screenH),
	}

	return nil
}

// TransitionWithConfig starts a transition with full configuration.
func (m *Manager) TransitionWithConfig(to ViewType, config TransitionConfig) error {
	view, ok := m.views[to]
	if !ok {
		return fmt.Errorf("view not registered: %s", to)
	}

	// Initialize the target view if needed
	if err := view.Init(); err != nil {
		return fmt.Errorf("failed to init view %s: %w", to, err)
	}

	// If no duration or no effect, just switch immediately
	if config.Duration <= 0 || config.Effect == TransitionNone {
		return m.SetCurrent(to)
	}

	// Use default easing if none provided
	easing := config.Easing
	if easing == nil {
		easing = EaseInOutQuad
	}

	// Start the transition
	m.next = view
	m.transition = &activeTransition{
		from:     m.current,
		to:       view,
		effect:   config.Effect,
		easing:   easing,
		duration: config.Duration,
		elapsed:  0,
		buffer:   ebiten.NewImage(m.screenW, m.screenH),
	}

	return nil
}

// IsTransitioning returns true if a transition is in progress.
func (m *Manager) IsTransitioning() bool {
	return m.transition != nil
}

// Update updates the current view and any active transition.
// dt is delta time in seconds.
func (m *Manager) Update(dt float64) error {
	// Update transition
	if m.transition != nil {
		m.transition.elapsed += dt

		// Check if transition is complete
		if m.transition.elapsed >= m.transition.duration {
			m.completeTransition()
		}
	}

	// Update current view
	if m.current != nil && m.transition == nil {
		trans := m.current.Update(dt)
		if trans != nil {
			if err := m.TransitionTo(trans); err != nil {
				return err
			}
		}
	}

	// Update both views during crossfade
	if m.transition != nil && m.transition.effect == TransitionCrossfade {
		m.transition.from.Update(dt)
		m.transition.to.Update(dt)
	}

	return nil
}

// completeTransition finishes the current transition.
func (m *Manager) completeTransition() {
	if m.transition == nil {
		return
	}

	// Handle exit/enter
	fromType := ViewNone
	if m.transition.from != nil {
		fromType = m.transition.from.Type()
		m.transition.from.Exit(m.transition.to.Type())
	}

	m.current = m.transition.to
	m.current.Enter(fromType)

	// Clean up
	m.transition.buffer.Dispose()
	m.transition = nil
	m.next = nil
}

// Draw renders the current view or transition to the screen.
func (m *Manager) Draw(screen *ebiten.Image) {
	if m.transition != nil {
		m.drawTransition(screen)
		return
	}

	if m.current != nil {
		m.current.Draw(screen)
	}
}

// drawTransition renders a transition effect.
func (m *Manager) drawTransition(screen *ebiten.Image) {
	if m.transition == nil {
		return
	}

	// Calculate linear progress
	linearProgress := m.transition.elapsed / m.transition.duration
	if linearProgress > 1.0 {
		linearProgress = 1.0
	}

	// Apply easing function
	progress := linearProgress
	if m.transition.easing != nil {
		progress = m.transition.easing(linearProgress)
	}

	switch m.transition.effect {
	case TransitionFade:
		m.drawFadeTransition(screen, progress)
	case TransitionCrossfade:
		m.drawCrossfadeTransition(screen, progress)
	case TransitionWipe:
		m.drawWipeTransition(screen, progress)
	case TransitionZoom:
		m.drawZoomTransition(screen, progress)
	default:
		// Just draw the target view
		m.transition.to.Draw(screen)
	}
}

// drawFadeTransition fades to black, then fades in the new view.
func (m *Manager) drawFadeTransition(screen *ebiten.Image, progress float64) {
	if progress < 0.5 {
		// Fade out: draw old view with decreasing alpha
		m.transition.from.Draw(screen)
		alpha := progress * 2 // 0 to 1 over first half
		drawFadeOverlay(screen, alpha)
	} else {
		// Fade in: draw new view with increasing alpha
		m.transition.to.Draw(screen)
		alpha := (1.0 - progress) * 2 // 1 to 0 over second half
		drawFadeOverlay(screen, alpha)
	}
}

// drawCrossfadeTransition blends between two views.
func (m *Manager) drawCrossfadeTransition(screen *ebiten.Image, progress float64) {
	// Draw old view
	m.transition.from.Draw(screen)

	// Draw new view to buffer
	m.transition.buffer.Clear()
	m.transition.to.Draw(m.transition.buffer)

	// Blend new view on top with alpha
	op := &ebiten.DrawImageOptions{}
	op.ColorScale.ScaleAlpha(float32(progress))
	screen.DrawImage(m.transition.buffer, op)
}

// drawWipeTransition wipes from left to right.
func (m *Manager) drawWipeTransition(screen *ebiten.Image, progress float64) {
	// Draw old view
	m.transition.from.Draw(screen)

	// Draw new view to buffer
	m.transition.buffer.Clear()
	m.transition.to.Draw(m.transition.buffer)

	// Create wipe effect by using a sub-image
	wipeX := int(float64(m.screenW) * progress)
	if wipeX > 0 {
		subRect := m.transition.buffer.SubImage(
			image.Rect(0, 0, wipeX, m.screenH),
		).(*ebiten.Image)
		screen.DrawImage(subRect, nil)
	}
}

// drawZoomTransition zooms out, then zooms in.
func (m *Manager) drawZoomTransition(screen *ebiten.Image, progress float64) {
	op := &ebiten.DrawImageOptions{}

	if progress < 0.5 {
		// Zoom out: scale down from 1 to 0.5
		scale := 1.0 - progress // 1.0 to 0.5
		op.GeoM.Translate(-float64(m.screenW)/2, -float64(m.screenH)/2)
		op.GeoM.Scale(scale, scale)
		op.GeoM.Translate(float64(m.screenW)/2, float64(m.screenH)/2)
		op.ColorScale.ScaleAlpha(float32(1.0 - progress*2))

		m.transition.buffer.Clear()
		m.transition.from.Draw(m.transition.buffer)
		screen.DrawImage(m.transition.buffer, op)
	} else {
		// Zoom in: scale up from 0.5 to 1
		scale := progress // 0.5 to 1.0
		op.GeoM.Translate(-float64(m.screenW)/2, -float64(m.screenH)/2)
		op.GeoM.Scale(scale, scale)
		op.GeoM.Translate(float64(m.screenW)/2, float64(m.screenH)/2)
		op.ColorScale.ScaleAlpha(float32((progress - 0.5) * 2))

		m.transition.buffer.Clear()
		m.transition.to.Draw(m.transition.buffer)
		screen.DrawImage(m.transition.buffer, op)
	}
}

// drawFadeOverlay draws a black overlay with the given alpha.
func drawFadeOverlay(screen *ebiten.Image, alpha float64) {
	w, h := screen.Bounds().Dx(), screen.Bounds().Dy()
	overlay := ebiten.NewImage(w, h)
	overlay.Fill(color.RGBA{0, 0, 0, uint8(alpha * 255)})
	screen.DrawImage(overlay, nil)
	overlay.Dispose()
}

// Resize updates the manager's screen dimensions.
// Should be called when the window is resized.
func (m *Manager) Resize(screenW, screenH int) {
	m.screenW = screenW
	m.screenH = screenH
}

// GetView returns a registered view by type, or nil if not found.
func (m *Manager) GetView(viewType ViewType) View {
	return m.views[viewType]
}
