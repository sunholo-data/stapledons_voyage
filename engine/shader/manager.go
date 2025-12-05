// Package shader provides GPU shader management and post-processing effects.
package shader

import (
	"embed"
	"fmt"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
)

//go:embed shaders/*.kage
var shaderFS embed.FS

// Manager handles shader compilation and caching.
type Manager struct {
	mu      sync.RWMutex
	shaders map[string]*ebiten.Shader
}

// NewManager creates a new shader manager.
func NewManager() *Manager {
	return &Manager{
		shaders: make(map[string]*ebiten.Shader),
	}
}

// Get returns a compiled shader, compiling on first access.
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

// Preload compiles all shaders at startup.
func (m *Manager) Preload() error {
	names := []string{
		"blur",
		"bloom_extract",
		"bloom_combine",
		"vignette",
		"crt",
		"aberration",
		"sr_warp",
		"gr_lensing",
		"gr_redshift",
	}

	for _, name := range names {
		if _, err := m.Get(name); err != nil {
			return err
		}
	}
	return nil
}

// Clear invalidates all cached shaders (for hot reload).
func (m *Manager) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shaders = make(map[string]*ebiten.Shader)
}

// ShaderNames returns the list of available shaders.
func (m *Manager) ShaderNames() []string {
	return []string{
		"blur",
		"bloom_extract",
		"bloom_combine",
		"vignette",
		"crt",
		"aberration",
		"sr_warp",
		"gr_lensing",
		"gr_redshift",
	}
}
