package view

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
)

// Manager coordinates views and handles transitions between them.
type Manager struct {
	views       map[ViewType]View
	current     View
	next        View
	transition  *Transition
	screenW     int
	screenH     int
	initialized bool
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
	config := TransitionConfig{
		Effect:   trans.Effect,
		Duration: trans.Duration,
		Easing:   EaseInOutQuad,
	}
	m.transition = NewTransitionState(m.current, view, config, m.screenW, m.screenH)

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

	// Start the transition
	m.next = view
	m.transition = NewTransitionState(m.current, view, config, m.screenW, m.screenH)

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
		if m.transition.Update(dt) {
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
	if m.transition != nil && m.transition.Effect() == TransitionCrossfade {
		m.transition.From().Update(dt)
		m.transition.To().Update(dt)
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
	if m.transition.From() != nil {
		fromType = m.transition.From().Type()
		m.transition.From().Exit(m.transition.To().Type())
	}

	m.current = m.transition.To()
	m.current.Enter(fromType)

	// Clean up
	m.transition.Dispose()
	m.transition = nil
	m.next = nil
}

// Draw renders the current view or transition to the screen.
func (m *Manager) Draw(screen *ebiten.Image) {
	if m.transition != nil {
		m.transition.Draw(screen)
		return
	}

	if m.current != nil {
		m.current.Draw(screen)
	}
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
