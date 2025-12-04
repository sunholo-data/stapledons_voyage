package shader

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// Effect represents a single post-processing effect.
type Effect struct {
	Name     string
	Shader   string
	Enabled  bool
	Uniforms map[string]any
}

// Pipeline manages multi-pass post-processing.
type Pipeline struct {
	manager  *Manager
	effects  []*Effect
	buffers  [2]*ebiten.Image
	screenW  int
	screenH  int
}

// NewPipeline creates a new post-processing pipeline.
func NewPipeline(manager *Manager) *Pipeline {
	return &Pipeline{
		manager: manager,
		effects: make([]*Effect, 0),
	}
}

// AddEffect adds an effect to the pipeline.
func (p *Pipeline) AddEffect(name, shader string, uniforms map[string]any) *Effect {
	effect := &Effect{
		Name:     name,
		Shader:   shader,
		Enabled:  true,
		Uniforms: uniforms,
	}
	p.effects = append(p.effects, effect)
	return effect
}

// GetEffect returns an effect by name.
func (p *Pipeline) GetEffect(name string) *Effect {
	for _, e := range p.effects {
		if e.Name == name {
			return e
		}
	}
	return nil
}

// SetEnabled enables or disables an effect by name.
func (p *Pipeline) SetEnabled(name string, enabled bool) {
	if e := p.GetEffect(name); e != nil {
		e.Enabled = enabled
	}
}

// IsEnabled returns whether an effect is enabled.
func (p *Pipeline) IsEnabled(name string) bool {
	if e := p.GetEffect(name); e != nil {
		return e.Enabled
	}
	return false
}

// ToggleEffect toggles an effect on/off.
func (p *Pipeline) ToggleEffect(name string) bool {
	if e := p.GetEffect(name); e != nil {
		e.Enabled = !e.Enabled
		return e.Enabled
	}
	return false
}

// SetUniform sets a uniform value for an effect.
func (p *Pipeline) SetUniform(name, uniform string, value any) {
	if e := p.GetEffect(name); e != nil {
		e.Uniforms[uniform] = value
	}
}

// SetSize allocates/resizes render buffers.
func (p *Pipeline) SetSize(w, h int) {
	if p.screenW == w && p.screenH == h {
		return
	}

	p.screenW = w
	p.screenH = h

	p.buffers[0] = ebiten.NewImage(w, h)
	p.buffers[1] = ebiten.NewImage(w, h)
}

// Apply runs all enabled effects on the input image.
func (p *Pipeline) Apply(screen *ebiten.Image, input *ebiten.Image) {

	// Count enabled effects
	enabledCount := 0
	for _, e := range p.effects {
		if e.Enabled {
			enabledCount++
		}
	}

	if enabledCount == 0 {
		// No effects, direct copy
		screen.DrawImage(input, nil)
		return
	}

	// Ping-pong through buffers
	src := input
	dstIdx := 0

	stageNum := 0
	for _, effect := range p.effects {
		if !effect.Enabled {
			continue
		}

		stageNum++
		isLast := stageNum == enabledCount

		// Determine destination
		var dst *ebiten.Image
		if isLast {
			dst = screen
		} else {
			dst = p.buffers[dstIdx]
			dst.Clear()
		}

		// Get shader
		shader, err := p.manager.Get(effect.Shader)
		if err != nil {
			// Fallback: copy without effect
			dst.DrawImage(src, nil)
			if !isLast {
				src = p.buffers[dstIdx]
				dstIdx = 1 - dstIdx
			}
			continue
		}

		// Apply shader
		opts := &ebiten.DrawRectShaderOptions{}
		opts.Images[0] = src
		opts.Uniforms = effect.Uniforms

		dst.DrawRectShader(p.screenW, p.screenH, shader, opts)

		// Swap for next iteration
		if !isLast {
			src = p.buffers[dstIdx]
			dstIdx = 1 - dstIdx
		}
	}
}

// EffectNames returns names of all effects.
func (p *Pipeline) EffectNames() []string {
	names := make([]string, len(p.effects))
	for i, e := range p.effects {
		names[i] = e.Name
	}
	return names
}

// EnabledEffects returns names of enabled effects.
func (p *Pipeline) EnabledEffects() []string {
	names := make([]string, 0)
	for _, e := range p.effects {
		if e.Enabled {
			names = append(names, e.Name)
		}
	}
	return names
}
