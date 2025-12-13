# Voyage API Reference CLI

## Status
- Status: Planned
- Priority: P2 (Developer tooling)
- Estimated: 1 day

## Game Vision Alignment

Checked against [core-pillars.md](../../docs/vision/core-pillars.md):

| Pillar | Alignment | Notes |
|--------|-----------|-------|
| Time Dilation Consequence | N/A | Infrastructure/tooling |
| Civilization Simulation | N/A | Infrastructure/tooling |
| Philosophical Depth | N/A | Infrastructure/tooling |
| Ship & Crew Life | N/A | Infrastructure/tooling |
| Legacy Impact | N/A | Infrastructure/tooling |
| Hard Sci-Fi Authenticity | N/A | Infrastructure/tooling |

**This is a developer tooling feature** that improves development velocity by providing quick CLI access to engine API documentation. It has no direct gameplay impact.

## Problem Statement

When writing Go code that uses the engine (demos, entry points), finding the correct API signatures is time-consuming:

1. **API Discovery**: Methods like `scene3D.SetCameraLookAt` vs `scene3D.LookAt` require searching through source files
2. **Signature Lookup**: Parameter types and order (e.g., `shader.NewSRWarp(shaderMgr)` vs `shader.NewSRWarp()`) require reading source
3. **Package Exploration**: Understanding what's available in `engine/tetra`, `engine/lod`, `engine/shader` requires manual exploration

**Example pain point** (from AILANG Solar Demo sprint):
```
cmd/demo-ailang-solar/main.go:65:10: scene3D.SetCameraLookAt undefined
cmd/demo-ailang-solar/main.go:112:37: too many arguments in call to shader.NewManager
cmd/demo-ailang-solar/main.go:115:12: not enough arguments in call to shader.NewSRWarp
```

These errors required searching through demo-lod to find correct patterns.

## Proposed Solution

Add a `voyage api` CLI command that provides quick access to engine API documentation.

### Command Structure

```bash
# List all packages
voyage api

# List types and functions in a package
voyage api tetra
voyage api lod
voyage api shader

# Show details for a specific type
voyage api tetra.Scene
voyage api lod.Manager

# Show method signatures for a type
voyage api tetra.Scene --methods
voyage api shader.SRWarp --methods

# Search across all packages
voyage api --search "camera"
voyage api --search "warp"
```

### Example Output

```bash
$ voyage api tetra.Scene

tetra.Scene - 3D rendering scene using Tetra3D

Constructor:
  func NewScene(width, height int) *Scene

Methods:
  func (s *Scene) SetCameraPosition(x, y, z float64)
  func (s *Scene) LookAt(x, y, z float64)
  func (s *Scene) SetLightingEnabled(enabled bool)
  func (s *Scene) Render() *ebiten.Image
  func (s *Scene) AddNode(node tetra3d.INode)

See also: tetra.Planet, tetra.StarLight, tetra.AmbientLight
```

```bash
$ voyage api shader --methods

shader.Manager
  func NewManager() *Manager

shader.SRWarp
  func NewSRWarp(mgr *Manager) *SRWarp
  func (s *SRWarp) SetForwardVelocity(v float64)
  func (s *SRWarp) IsEnabled() bool
  func (s *SRWarp) Apply(dst, src *ebiten.Image) bool

shader.GRWarp
  func NewGRWarp(mgr *Manager) *GRWarp
  func (g *GRWarp) SetDemoMode(centerX, centerY, phi, rs float32)
  func (g *GRWarp) IsEnabled() bool
  func (g *GRWarp) Apply(dst, src *ebiten.Image) bool
```

## Implementation

### Option A: Go AST Parsing (Recommended)

Use Go's `go/ast` and `go/parser` packages to extract API information at runtime:

```go
// cmd/voyage/api.go
package main

import (
    "go/ast"
    "go/parser"
    "go/token"
)

func extractAPIInfo(packagePath string) (*PackageAPI, error) {
    fset := token.NewFileSet()
    pkgs, err := parser.ParseDir(fset, packagePath, nil, parser.ParseComments)
    // ... extract types, methods, comments
}
```

**Pros:**
- Always up-to-date with source
- Extracts doc comments automatically
- Can include unexported items with flag

**Cons:**
- Slightly slower (parses source each time)
- Requires source files to be present

### Option B: Pre-generated API Database

Generate a JSON/YAML file at build time with API info:

```bash
make api-docs  # Generates engine-api.json
```

**Pros:**
- Fast lookup
- Can work without source files

**Cons:**
- Can get out of sync
- Extra build step

### Recommendation

**Use Option A** for simplicity and accuracy. Performance is acceptable for developer tooling (< 1 second even for full scan).

### File Structure

```
cmd/
  voyage/
    main.go          # Entry point with subcommands
    api.go           # API lookup implementation
    api_formatter.go # Output formatting
```

### Integration with Existing CLI

If `cmd/voyage` doesn't exist yet, create it as the project's CLI tool:

```go
// cmd/voyage/main.go
package main

import (
    "fmt"
    "os"
)

func main() {
    if len(os.Args) < 2 {
        printUsage()
        return
    }

    switch os.Args[1] {
    case "api":
        handleAPI(os.Args[2:])
    case "help":
        printUsage()
    default:
        fmt.Printf("Unknown command: %s\n", os.Args[1])
    }
}
```

## Packages to Document

| Package | Key Types | Description |
|---------|-----------|-------------|
| `engine/tetra` | Scene, Planet, StarLight, AmbientLight | 3D rendering with Tetra3D |
| `engine/lod` | Manager, Object, SimpleCamera, PointRenderer, CircleRenderer, BillboardRenderer | Level-of-detail system |
| `engine/shader` | Manager, SRWarp, GRWarp | Relativity shader effects |
| `engine/assets` | SpriteAtlas, AudioManager, FontManager | Asset loading |
| `engine/render` | DrawFrame, CaptureInput | Render bridge |
| `sim_gen` | (generated) | AILANG-generated types |

## Success Criteria

- [ ] `voyage api` lists all engine packages
- [ ] `voyage api tetra` shows types in package
- [ ] `voyage api tetra.Scene` shows constructor and methods
- [ ] `voyage api --search "camera"` finds relevant APIs
- [ ] Output is formatted for easy reading in terminal
- [ ] Build with `go build -o bin/voyage ./cmd/voyage`

## Future Enhancements

1. **`--json` flag**: Machine-readable output for tooling
2. **`--markdown` flag**: Generate markdown for documentation
3. **Integration with Claude Code**: Add to CLAUDE.md as available tool
4. **Tab completion**: Shell completion for type names

## References

- Go AST package: https://pkg.go.dev/go/ast
- Go parser package: https://pkg.go.dev/go/parser
- Existing engine docs: [engine-capabilities.md](../reference/engine-capabilities.md)
