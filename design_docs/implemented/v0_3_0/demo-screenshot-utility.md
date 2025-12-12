# Demo Screenshot Utility

## Status
- **Status:** Implemented
- **Priority:** P1 (Developer Experience)
- **Actual Effort:** 0.5 days
- **Category:** Engine/Infrastructure
- **Version:** v0.3.0

## Problem Statement

Each demo command needed to implement its own screenshot support by:
1. Adding flag parsing (`--screenshot`, `--output`)
2. Implementing save logic
3. Duplicating the same code across demos

The existing `engine/screenshot/screenshot.go` is tightly coupled to the game simulation and cannot be reused for arbitrary demos.

## Solution: Generic Demo Runner

Created a reusable `engine/demo` package that wraps any `ebiten.Game` with:
- Standard CLI flags (`--screenshot`, `--output`)
- Automatic screenshot capture at specified frame
- Clean exit after capture

## Implementation

### Files Created

| File | Purpose |
|------|---------|
| [engine/demo/run.go](../../../engine/demo/run.go) | Main `Run()` function, `captureWrapper`, screenshot logic |

### API

```go
package demo

// Config holds demo configuration.
type Config struct {
    Title  string // Window title
    Width  int    // Window width (0 = use display.InternalWidth)
    Height int    // Window height (0 = use display.InternalHeight)
}

// Run wraps any ebiten.Game with standard CLI flags and screenshot support.
func Run(game ebiten.Game, cfg Config) error
```

### CLI Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--screenshot N` | -1 (disabled) | Take screenshot at frame N and exit |
| `--output PATH` | `out/screenshots/<title>.png` | Screenshot output path |

### Usage

```go
func main() {
    game := NewDemoGame()
    demo.Run(game, demo.Config{Title: "Parallax Demo"})
}
```

### Running Demos

```bash
# Interactive mode
go run ./cmd/demo-game-parallax

# Take screenshot at frame 30
go run ./cmd/demo-game-parallax --screenshot 30 --output out/screenshots/parallax.png

# With custom camera position (demo-specific flags)
go run ./cmd/demo-game-parallax -camx 300 --screenshot 1 --output test.png
```

## How It Works

1. **Flag Parsing**: `demo.Run()` calls `flag.Parse()` to parse standard flags
2. **Game Wrapping**: If `--screenshot` is set, wraps the game with `captureWrapper`
3. **Frame Counting**: `captureWrapper.Update()` increments frame counter
4. **Screenshot Capture**: When target frame is reached, saves PNG and exits
5. **Directory Creation**: Automatically creates output directory if needed

## Integration

Works with existing screenshot script:
```bash
.claude/skills/sprint-executor/scripts/take_screenshot.sh -c demo-game-parallax -f 30
```

## Success Criteria - All Met

- [x] `engine/demo` package created with `Run()` function
- [x] `demo-game-parallax` migrated to use it
- [x] `take_screenshot.sh -c demo-game-parallax` works
- [x] Screenshot captured at correct frame
- [x] Output path respected
- [x] Clean exit after capture

## AILANG/Engine Boundary

This is purely **engine infrastructure** - it handles HOW demos are run and captured, not WHAT they simulate. AILANG code is not affected.

## References

- Implementation: [engine/demo/run.go](../../../engine/demo/run.go)
- Example usage: [cmd/demo-game-parallax/main.go](../../../cmd/demo-game-parallax/main.go)
- Screenshot script: [.claude/skills/sprint-executor/scripts/take_screenshot.sh](../../../.claude/skills/sprint-executor/scripts/take_screenshot.sh)
