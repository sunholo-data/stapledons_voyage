# Shader System

**Version:** 0.4.5
**Status:** Planned
**Priority:** P1 (Visual Polish)
**Complexity:** High
**Package:** `engine/shader`
**AILANG Impact:** None - shaders are pure engine rendering

## Related Documents

- [SR Rendering Go](../sr-rendering-go.md) - Specific shader implementation for relativity
- [Particle System](../v0_5_0/particle-system.md) - May use shaders for particle effects
- [Screen Transitions](../v0_5_0/screen-transitions.md) - Transition effects

## Problem Statement

**Current State:**
- No shader support beyond basic Ebiten DrawImage
- Visual effects limited to sprite compositing
- No post-processing (blur, bloom, color grading)
- No custom sprite effects (outline, glow, distortion)

**What's Needed:**
- Kage shader compilation and management
- Post-processing pipeline (multi-pass)
- Per-sprite shader effects
- Full-screen effects (CRT scanlines, vignette, etc.)
- Shader hot-reload for development

**Design Principle:** Shaders are engine implementation details. AILANG requests visual effects by ID; engine decides whether to use shaders or not.

## Ebiten Kage Overview

Ebiten uses **Kage**, a custom shader language that compiles to various backends:

```kage
//kage:unit pixels
package main

var Time float  // Uniform

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
    // Sample source texture
    c := imageSrc0At(srcPos)

    // Apply effect
    c.rgb *= sin(Time) * 0.5 + 0.5

    return c
}
```

**Limitations:**
- No vertex shaders (Ebiten handles geometry)
- Limited to fragment-level effects
- Max 4 source images per draw call
- Float32 uniforms only

## Architecture

```
engine/shader/
├── manager.go       # Shader compilation and caching
├── postprocess.go   # Post-processing pipeline
├── effects/
│   ├── blur.go      # Gaussian blur
│   ├── bloom.go     # Glow/bloom effect
│   ├── crt.go       # CRT scanline effect
│   ├── vignette.go  # Screen edge darkening
│   ├── aberration.go # Chromatic aberration
│   └── outline.go   # Sprite outline
└── shaders/
    ├── blur.kage
    ├── bloom_extract.kage
    ├── bloom_combine.kage
    ├── crt.kage
    ├── vignette.kage
    ├── aberration.kage
    └── outline.kage
```

## Shader Manager (engine/shader/manager.go)

```go
package shader

import (
    "embed"
    "fmt"
    "sync"

    "github.com/hajimehoshi/ebiten/v2"
)

//go:embed shaders/*.kage
var shaderFS embed.FS

// Manager handles shader compilation and caching
type Manager struct {
    mu      sync.RWMutex
    shaders map[string]*ebiten.Shader
    debug   bool // Hot reload in debug mode
}

func NewManager() *Manager {
    return &Manager{
        shaders: make(map[string]*ebiten.Shader),
    }
}

// Get returns a compiled shader, compiling on first access
func (m *Manager) Get(name string) (*ebiten.Shader, error) {
    m.mu.RLock()
    if s, ok := m.shaders[name]; ok {
        m.mu.RUnlock()
        return s, nil
    }
    m.mu.RUnlock()

    // Compile shader
    m.mu.Lock()
    defer m.mu.Unlock()

    // Double-check after acquiring write lock
    if s, ok := m.shaders[name]; ok {
        return s, nil
    }

    src, err := shaderFS.ReadFile(fmt.Sprintf("shaders/%s.kage", name))
    if err != nil {
        return nil, fmt.Errorf("shader %s not found: %w", name, err)
    }

    shader, err := ebiten.NewShader(src)
    if err != nil {
        return nil, fmt.Errorf("failed to compile shader %s: %w", name, err)
    }

    m.shaders[name] = shader
    return shader, nil
}

// Preload compiles all shaders at startup
func (m *Manager) Preload() error {
    names := []string{
        "blur",
        "bloom_extract",
        "bloom_combine",
        "crt",
        "vignette",
        "aberration",
        "outline",
    }

    for _, name := range names {
        if _, err := m.Get(name); err != nil {
            return err
        }
    }
    return nil
}

// Clear invalidates all cached shaders (for hot reload)
func (m *Manager) Clear() {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.shaders = make(map[string]*ebiten.Shader)
}
```

## Post-Processing Pipeline (engine/shader/postprocess.go)

```go
package shader

import (
    "github.com/hajimehoshi/ebiten/v2"
)

// PostProcessPipeline manages multi-pass post-processing
type PostProcessPipeline struct {
    manager    *Manager
    stages     []PostProcessStage
    buffers    [2]*ebiten.Image // Ping-pong buffers
    screenW    int
    screenH    int
}

// PostProcessStage is a single effect in the pipeline
type PostProcessStage struct {
    Name     string
    Shader   string            // Shader name
    Enabled  bool
    Uniforms map[string]any
}

func NewPostProcessPipeline(manager *Manager) *PostProcessPipeline {
    return &PostProcessPipeline{
        manager: manager,
    }
}

// SetSize allocates/resizes render buffers
func (p *PostProcessPipeline) SetSize(w, h int) {
    if p.screenW == w && p.screenH == h {
        return
    }

    p.screenW = w
    p.screenH = h

    p.buffers[0] = ebiten.NewImage(w, h)
    p.buffers[1] = ebiten.NewImage(w, h)
}

// AddStage adds a post-processing effect
func (p *PostProcessPipeline) AddStage(name, shader string, uniforms map[string]any) {
    p.stages = append(p.stages, PostProcessStage{
        Name:     name,
        Shader:   shader,
        Enabled:  true,
        Uniforms: uniforms,
    })
}

// SetStageEnabled enables/disables a stage by name
func (p *PostProcessPipeline) SetStageEnabled(name string, enabled bool) {
    for i := range p.stages {
        if p.stages[i].Name == name {
            p.stages[i].Enabled = enabled
            return
        }
    }
}

// SetStageUniform updates a uniform for a stage
func (p *PostProcessPipeline) SetStageUniform(name, uniform string, value any) {
    for i := range p.stages {
        if p.stages[i].Name == name {
            p.stages[i].Uniforms[uniform] = value
            return
        }
    }
}

// Apply runs all enabled stages on the input image
func (p *PostProcessPipeline) Apply(screen *ebiten.Image, input *ebiten.Image) {
    if len(p.stages) == 0 {
        // No stages, direct copy
        screen.DrawImage(input, nil)
        return
    }

    // Count enabled stages
    enabledCount := 0
    for _, s := range p.stages {
        if s.Enabled {
            enabledCount++
        }
    }

    if enabledCount == 0 {
        screen.DrawImage(input, nil)
        return
    }

    // Ping-pong through buffers
    src := input
    dstIdx := 0

    stageNum := 0
    for _, stage := range p.stages {
        if !stage.Enabled {
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
        shader, err := p.manager.Get(stage.Shader)
        if err != nil {
            // Fallback: copy without effect
            dst.DrawImage(src, nil)
            continue
        }

        // Apply shader
        opts := &ebiten.DrawRectShaderOptions{}
        opts.Images[0] = src
        opts.Uniforms = stage.Uniforms

        dst.DrawRectShader(p.screenW, p.screenH, shader, opts)

        // Swap for next iteration
        if !isLast {
            src = p.buffers[dstIdx]
            dstIdx = 1 - dstIdx
        }
    }
}
```

## Effect Implementations

### Gaussian Blur (engine/shader/shaders/blur.kage)

```kage
//kage:unit pixels
package main

var Radius float  // Blur radius in pixels
var Direction vec2  // (1,0) for horizontal, (0,1) for vertical

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
    // 9-tap Gaussian blur
    weights := [9]float{
        0.0162, 0.0540, 0.1216, 0.1945, 0.2274,
        0.1945, 0.1216, 0.0540, 0.0162,
    }

    result := vec4(0)

    for i := 0; i < 9; i++ {
        offset := Direction * Radius * float(i-4)
        result += imageSrc0At(srcPos + offset) * weights[i]
    }

    return result
}
```

### Bloom Extract (engine/shader/shaders/bloom_extract.kage)

```kage
//kage:unit pixels
package main

var Threshold float  // Brightness threshold (0.8-1.0)

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
    c := imageSrc0At(srcPos)

    // Calculate luminance
    luminance := dot(c.rgb, vec3(0.299, 0.587, 0.114))

    // Extract bright areas
    if luminance > Threshold {
        return c
    }
    return vec4(0, 0, 0, c.a)
}
```

### Bloom Combine (engine/shader/shaders/bloom_combine.kage)

```kage
//kage:unit pixels
package main

var Intensity float  // Bloom intensity (0.5-2.0)

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
    original := imageSrc0At(srcPos)  // Original scene
    bloom := imageSrc1At(srcPos)     // Blurred bright areas

    return original + bloom * Intensity
}
```

### CRT Effect (engine/shader/shaders/crt.kage)

```kage
//kage:unit pixels
package main

var ScanlineIntensity float  // 0.0-1.0
var Curvature float          // 0.0 = flat, 0.1 = curved
var VignetteAmount float     // Edge darkening

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
    screenW, screenH := imageDstSize()

    // Normalize coordinates to -1..1
    uv := srcPos / vec2(screenW, screenH)
    uv = uv * 2.0 - 1.0

    // Apply barrel distortion (CRT curvature)
    r2 := dot(uv, uv)
    uv *= 1.0 + Curvature * r2

    // Back to 0..1
    uv = (uv + 1.0) / 2.0

    // Check bounds
    if uv.x < 0 || uv.x > 1 || uv.y < 0 || uv.y > 1 {
        return vec4(0, 0, 0, 1)
    }

    // Sample with distorted coords
    c := imageSrc0At(uv * vec2(screenW, screenH))

    // Scanlines
    scanline := sin(srcPos.y * 3.14159 * 2.0) * 0.5 + 0.5
    c.rgb *= 1.0 - ScanlineIntensity * (1.0 - scanline)

    // Vignette
    dist := length(uv - 0.5) * 2.0
    vignette := 1.0 - dist * dist * VignetteAmount
    c.rgb *= clamp(vignette, 0.0, 1.0)

    return c
}
```

### Vignette (engine/shader/shaders/vignette.kage)

```kage
//kage:unit pixels
package main

var Intensity float  // 0.0-1.0
var Softness float   // 0.2-0.8

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
    screenW, screenH := imageDstSize()

    c := imageSrc0At(srcPos)

    // Normalized coords centered at screen middle
    uv := srcPos / vec2(screenW, screenH)
    uv -= 0.5

    // Distance from center (ellipse to handle aspect ratio)
    aspect := screenW / screenH
    uv.x *= aspect
    dist := length(uv)

    // Smooth falloff
    vignette := smoothstep(Softness, Softness + 0.3, dist)
    c.rgb *= 1.0 - vignette * Intensity

    return c
}
```

### Chromatic Aberration (engine/shader/shaders/aberration.kage)

```kage
//kage:unit pixels
package main

var Amount float  // Offset amount (2-10 pixels)

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
    screenW, screenH := imageDstSize()

    // Direction from center
    center := vec2(screenW, screenH) / 2.0
    dir := normalize(srcPos - center)

    // Sample RGB channels with offset
    r := imageSrc0At(srcPos + dir * Amount).r
    g := imageSrc0At(srcPos).g
    b := imageSrc0At(srcPos - dir * Amount).b

    return vec4(r, g, b, 1.0)
}
```

### Sprite Outline (engine/shader/shaders/outline.kage)

```kage
//kage:unit pixels
package main

var OutlineColor vec4  // RGBA outline color
var OutlineWidth float // Outline width in pixels

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
    c := imageSrc0At(srcPos)

    // If pixel is transparent, check neighbors for outline
    if c.a < 0.1 {
        // Sample neighbors
        offsets := [8]vec2{
            vec2(-1, -1), vec2(0, -1), vec2(1, -1),
            vec2(-1,  0),              vec2(1,  0),
            vec2(-1,  1), vec2(0,  1), vec2(1,  1),
        }

        for i := 0; i < 8; i++ {
            neighbor := imageSrc0At(srcPos + offsets[i] * OutlineWidth)
            if neighbor.a > 0.5 {
                return OutlineColor
            }
        }
        return vec4(0)
    }

    return c
}
```

## Bloom Effect Implementation (engine/shader/effects/bloom.go)

```go
package shader

import "github.com/hajimehoshi/ebiten/v2"

// BloomEffect provides glow/bloom post-processing
type BloomEffect struct {
    manager     *Manager
    threshold   float32
    intensity   float32
    blurPasses  int
    blurBuffer1 *ebiten.Image
    blurBuffer2 *ebiten.Image
    screenW     int
    screenH     int
}

func NewBloomEffect(manager *Manager) *BloomEffect {
    return &BloomEffect{
        manager:    manager,
        threshold:  0.8,
        intensity:  1.0,
        blurPasses: 3,
    }
}

func (b *BloomEffect) SetThreshold(t float32) { b.threshold = t }
func (b *BloomEffect) SetIntensity(i float32) { b.intensity = i }
func (b *BloomEffect) SetBlurPasses(p int)    { b.blurPasses = p }

func (b *BloomEffect) SetSize(w, h int) {
    if b.screenW == w && b.screenH == h {
        return
    }

    b.screenW = w
    b.screenH = h

    // Half-resolution for blur (performance)
    halfW, halfH := w/2, h/2
    b.blurBuffer1 = ebiten.NewImage(halfW, halfH)
    b.blurBuffer2 = ebiten.NewImage(halfW, halfH)
}

func (b *BloomEffect) Apply(dst, src *ebiten.Image) error {
    // Step 1: Extract bright pixels
    extractShader, err := b.manager.Get("bloom_extract")
    if err != nil {
        return err
    }

    b.blurBuffer1.Clear()
    opts := &ebiten.DrawRectShaderOptions{}
    opts.Images[0] = src
    opts.Uniforms = map[string]any{"Threshold": b.threshold}
    b.blurBuffer1.DrawRectShader(b.screenW/2, b.screenH/2, extractShader, opts)

    // Step 2: Blur (multiple passes)
    blurShader, err := b.manager.Get("blur")
    if err != nil {
        return err
    }

    for i := 0; i < b.blurPasses; i++ {
        // Horizontal pass
        b.blurBuffer2.Clear()
        opts := &ebiten.DrawRectShaderOptions{}
        opts.Images[0] = b.blurBuffer1
        opts.Uniforms = map[string]any{
            "Radius":    float32(4),
            "Direction": [2]float32{1, 0},
        }
        b.blurBuffer2.DrawRectShader(b.screenW/2, b.screenH/2, blurShader, opts)

        // Vertical pass
        b.blurBuffer1.Clear()
        opts = &ebiten.DrawRectShaderOptions{}
        opts.Images[0] = b.blurBuffer2
        opts.Uniforms = map[string]any{
            "Radius":    float32(4),
            "Direction": [2]float32{0, 1},
        }
        b.blurBuffer1.DrawRectShader(b.screenW/2, b.screenH/2, blurShader, opts)
    }

    // Step 3: Combine original + bloom
    combineShader, err := b.manager.Get("bloom_combine")
    if err != nil {
        return err
    }

    // Scale blur buffer back to full size for combining
    opts = &ebiten.DrawRectShaderOptions{}
    opts.Images[0] = src
    opts.Images[1] = b.blurBuffer1
    opts.Uniforms = map[string]any{"Intensity": b.intensity}
    dst.DrawRectShader(b.screenW, b.screenH, combineShader, opts)

    return nil
}
```

## Game Loop Integration

```go
type Game struct {
    // ...existing fields...
    shaderManager *shader.Manager
    postProcess   *shader.PostProcessPipeline
    bloom         *shader.BloomEffect
    renderBuffer  *ebiten.Image
}

func NewGame() *Game {
    g := &Game{
        shaderManager: shader.NewManager(),
    }

    // Preload shaders at startup
    if err := g.shaderManager.Preload(); err != nil {
        log.Printf("Warning: shader preload failed: %v", err)
    }

    // Setup post-processing pipeline
    g.postProcess = shader.NewPostProcessPipeline(g.shaderManager)
    g.postProcess.AddStage("vignette", "vignette", map[string]any{
        "Intensity": float32(0.3),
        "Softness":  float32(0.5),
    })

    // Setup bloom
    g.bloom = shader.NewBloomEffect(g.shaderManager)
    g.bloom.SetThreshold(0.85)
    g.bloom.SetIntensity(0.8)

    return g
}

func (g *Game) Draw(screen *ebiten.Image) {
    // Ensure buffers are sized
    w, h := screen.Size()
    if g.renderBuffer == nil || g.renderBuffer.Bounds().Dx() != w {
        g.renderBuffer = ebiten.NewImage(w, h)
        g.postProcess.SetSize(w, h)
        g.bloom.SetSize(w, h)
    }

    // Render scene to buffer
    g.renderBuffer.Clear()
    render.RenderFrame(g.renderBuffer, g.out)

    // Apply bloom
    bloomOutput := ebiten.NewImage(w, h)
    g.bloom.Apply(bloomOutput, g.renderBuffer)

    // Apply post-processing pipeline
    g.postProcess.Apply(screen, bloomOutput)
}
```

## Effect Catalog

| Effect | Use Case | Performance |
|--------|----------|-------------|
| **Blur** | UI backgrounds, depth of field | Medium |
| **Bloom** | Stars, engines, explosions | Medium-High |
| **CRT** | Retro aesthetic, ship monitors | Low |
| **Vignette** | Atmosphere, focus | Very Low |
| **Chromatic Aberration** | Damage, warp effects | Low |
| **Outline** | Selection, hover states | Low (per-sprite) |
| **SR Aberration** | Relativistic light bending | Medium |
| **SR Doppler** | Relativistic color shift | Medium |

## AILANG Integration

AILANG doesn't know about shaders. It requests effects by ID:

```ailang
type VisualEffect =
    | Bloom(float)           -- intensity
    | Vignette(float)        -- intensity
    | CRT(bool)              -- enabled
    | ChromaticAberration(float)

type FrameOutput = {
    draw: [DrawCmd],
    effects: [VisualEffect],  -- Active effects this frame
    -- ...
}
```

Engine maps these to shader stages:

```go
func (g *Game) applyEffectsFromAILANG(effects []sim_gen.VisualEffect) {
    for _, e := range effects {
        switch e.Kind {
        case sim_gen.VisualEffectKindBloom:
            g.bloom.SetIntensity(float32(e.Bloom))
        case sim_gen.VisualEffectKindVignette:
            g.postProcess.SetStageUniform("vignette", "Intensity", float32(e.Vignette))
        case sim_gen.VisualEffectKindCRT:
            g.postProcess.SetStageEnabled("crt", e.CRT)
        }
    }
}
```

## Implementation Plan

### Files to Create

| File | Purpose |
|------|---------|
| `engine/shader/manager.go` | Shader compilation and caching |
| `engine/shader/postprocess.go` | Multi-pass pipeline |
| `engine/shader/effects/bloom.go` | Bloom effect |
| `engine/shader/effects/blur.go` | Blur helper |
| `engine/shader/shaders/blur.kage` | Gaussian blur shader |
| `engine/shader/shaders/bloom_extract.kage` | Bloom threshold |
| `engine/shader/shaders/bloom_combine.kage` | Bloom additive blend |
| `engine/shader/shaders/crt.kage` | CRT scanlines |
| `engine/shader/shaders/vignette.kage` | Edge darkening |
| `engine/shader/shaders/aberration.kage` | Chromatic aberration |
| `engine/shader/shaders/outline.kage` | Sprite outline |

### Go Integration

| File | Change |
|------|--------|
| `cmd/game/main.go` | Add shader manager, post-process pipeline |
| `engine/render/draw.go` | Render to buffer, apply effects |

## Testing Strategy

### Visual Tests

```bash
make run-mock
# F5 = Toggle bloom
# F6 = Toggle CRT
# F7 = Toggle vignette
# F8 = Cycle chromatic aberration intensity
```

### Unit Tests

```go
func TestShaderCompilation(t *testing.T)
func TestPostProcessPipeline(t *testing.T)
func TestBloomThreshold(t *testing.T)
```

### Performance Tests

```go
func BenchmarkBloomFullHD(b *testing.B)
func BenchmarkPostProcessPipeline(b *testing.B)
```

## Success Criteria

### Core System
- [ ] All shaders compile without errors
- [ ] Shader hot-reload works in debug mode
- [ ] Post-process pipeline chains multiple effects

### Visual Effects
- [ ] Bloom creates visible glow on bright areas
- [ ] CRT effect shows scanlines and curvature
- [ ] Vignette darkens screen edges smoothly
- [ ] Chromatic aberration separates RGB channels

### Performance
- [ ] Full pipeline < 2ms at 1080p
- [ ] Bloom (3 passes) < 1ms at 1080p
- [ ] No stutter during effect enable/disable

### Integration
- [ ] AILANG VisualEffect commands map to shader settings
- [ ] Effects persist across frames correctly
- [ ] Clean degrade on shader compilation failure

---

**Created:** 2025-12-04
