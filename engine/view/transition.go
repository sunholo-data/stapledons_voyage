package view

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

// Transition holds the state of an in-progress view transition.
type Transition struct {
	from     View
	to       View
	effect   TransitionEffect
	easing   EasingFunc
	duration float64
	elapsed  float64
	buffer   *ebiten.Image
	screenW  int
	screenH  int
}

// NewTransitionState creates a new transition state.
func NewTransitionState(from, to View, config TransitionConfig, screenW, screenH int) *Transition {
	easing := config.Easing
	if easing == nil {
		easing = EaseInOutQuad
	}
	return &Transition{
		from:     from,
		to:       to,
		effect:   config.Effect,
		easing:   easing,
		duration: config.Duration,
		elapsed:  0,
		buffer:   ebiten.NewImage(screenW, screenH),
		screenW:  screenW,
		screenH:  screenH,
	}
}

// Update advances the transition by dt seconds.
// Returns true if the transition is complete.
func (t *Transition) Update(dt float64) bool {
	t.elapsed += dt
	return t.elapsed >= t.duration
}

// Progress returns the eased progress of the transition (0 to 1).
func (t *Transition) Progress() float64 {
	linear := t.elapsed / t.duration
	if linear > 1.0 {
		linear = 1.0
	}
	if t.easing != nil {
		return t.easing(linear)
	}
	return linear
}

// From returns the source view.
func (t *Transition) From() View {
	return t.from
}

// To returns the target view.
func (t *Transition) To() View {
	return t.to
}

// Effect returns the transition effect type.
func (t *Transition) Effect() TransitionEffect {
	return t.effect
}

// Dispose cleans up the transition's resources.
func (t *Transition) Dispose() {
	if t.buffer != nil {
		t.buffer.Dispose()
		t.buffer = nil
	}
}

// Draw renders the transition effect to the screen.
func (t *Transition) Draw(screen *ebiten.Image) {
	progress := t.Progress()

	switch t.effect {
	case TransitionFade:
		t.drawFade(screen, progress)
	case TransitionCrossfade:
		t.drawCrossfade(screen, progress)
	case TransitionWipe:
		t.drawWipe(screen, progress)
	case TransitionZoom:
		t.drawZoom(screen, progress)
	default:
		// Just draw the target view
		t.to.Draw(screen)
	}
}

// drawFade fades to black, then fades in the new view.
func (t *Transition) drawFade(screen *ebiten.Image, progress float64) {
	if progress < 0.5 {
		// Fade out: draw old view with increasing black overlay
		t.from.Draw(screen)
		alpha := progress * 2 // 0 to 1 over first half
		drawFadeOverlay(screen, alpha)
	} else {
		// Fade in: draw new view with decreasing black overlay
		t.to.Draw(screen)
		alpha := (1.0 - progress) * 2 // 1 to 0 over second half
		drawFadeOverlay(screen, alpha)
	}
}

// drawCrossfade blends between two views.
func (t *Transition) drawCrossfade(screen *ebiten.Image, progress float64) {
	// Draw old view
	t.from.Draw(screen)

	// Draw new view to buffer
	t.buffer.Clear()
	t.to.Draw(t.buffer)

	// Blend new view on top with alpha
	op := &ebiten.DrawImageOptions{}
	op.ColorScale.ScaleAlpha(float32(progress))
	screen.DrawImage(t.buffer, op)
}

// drawWipe wipes from left to right.
func (t *Transition) drawWipe(screen *ebiten.Image, progress float64) {
	// Draw old view
	t.from.Draw(screen)

	// Draw new view to buffer
	t.buffer.Clear()
	t.to.Draw(t.buffer)

	// Create wipe effect using a sub-image
	wipeX := int(float64(t.screenW) * progress)
	if wipeX > 0 {
		subRect := t.buffer.SubImage(
			image.Rect(0, 0, wipeX, t.screenH),
		).(*ebiten.Image)
		screen.DrawImage(subRect, nil)
	}
}

// drawZoom zooms out, then zooms in.
func (t *Transition) drawZoom(screen *ebiten.Image, progress float64) {
	op := &ebiten.DrawImageOptions{}

	if progress < 0.5 {
		// Zoom out: scale down from 1 to 0.5
		scale := 1.0 - progress // 1.0 to 0.5
		op.GeoM.Translate(-float64(t.screenW)/2, -float64(t.screenH)/2)
		op.GeoM.Scale(scale, scale)
		op.GeoM.Translate(float64(t.screenW)/2, float64(t.screenH)/2)
		op.ColorScale.ScaleAlpha(float32(1.0 - progress*2))

		t.buffer.Clear()
		t.from.Draw(t.buffer)
		screen.DrawImage(t.buffer, op)
	} else {
		// Zoom in: scale up from 0.5 to 1
		scale := progress // 0.5 to 1.0
		op.GeoM.Translate(-float64(t.screenW)/2, -float64(t.screenH)/2)
		op.GeoM.Scale(scale, scale)
		op.GeoM.Translate(float64(t.screenW)/2, float64(t.screenH)/2)
		op.ColorScale.ScaleAlpha(float32((progress - 0.5) * 2))

		t.buffer.Clear()
		t.to.Draw(t.buffer)
		screen.DrawImage(t.buffer, op)
	}
}

// drawFadeOverlay draws a black overlay with the given alpha (0-1).
func drawFadeOverlay(screen *ebiten.Image, alpha float64) {
	w, h := screen.Bounds().Dx(), screen.Bounds().Dy()
	overlay := ebiten.NewImage(w, h)
	overlay.Fill(color.RGBA{0, 0, 0, uint8(alpha * 255)})
	screen.DrawImage(overlay, nil)
	overlay.Dispose()
}
