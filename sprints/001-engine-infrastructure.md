# Sprint 001: Engine Infrastructure

**Status:** Planned
**Goal:** Prepare Go/Ebiten engine layer so it's ready when AILANG compiler ships
**Estimated Effort:** 3-4 sessions
**AILANG Dependency:** None (creates mock types to decouple)

## Context

AILANG compiler (`ailang compile --emit-go`) is still in development. This sprint builds all the Go/Ebiten infrastructure in parallel, using hand-written mock types that match the AILANG protocol.

**Key insight:** By creating `sim_gen/` mocks now, we can:
- Run `make run` and `make eval` immediately
- Build and test engine features
- Swap in real generated code later with zero changes

## Success Criteria

- [ ] `make run` launches a window showing colored tiles
- [ ] `make eval` produces `out/report.json` with passing scenarios
- [ ] Asset manager loads sprites from `assets/sprites/`
- [ ] Display manager supports F11 fullscreen toggle
- [ ] Config file persists settings between sessions

## Tasks

### Phase 1: Mock sim_gen Package (P0)

Create hand-written Go types matching [sim/protocol.ail](../sim/protocol.ail) and [sim/world.ail](../sim/world.ail).

| Task | File | Description |
|------|------|-------------|
| 1.1 | `sim_gen/types.go` | Core types: Coord, Tile, World, PlanetState, NPC |
| 1.2 | `sim_gen/protocol.go` | IO types: MouseState, KeyEvent, FrameInput, FrameOutput |
| 1.3 | `sim_gen/draw_cmd.go` | DrawCmd sum type (Rect, Sprite, Text variants) |
| 1.4 | `sim_gen/funcs.go` | InitWorld() and Step() functions |
| 1.5 | Verify | `make run` launches, `make eval` passes |

**Estimated:** 1 session

### Phase 2: Asset Management (P0)

Implement sprite/font loading from [asset-management.md](../design_docs/planned/v0_2_0/asset-management.md).

| Task | File | Description |
|------|------|-------------|
| 2.1 | `engine/assets/manager.go` | Manager struct, NewManager(), LoadManifests() |
| 2.2 | `engine/assets/sprites.go` | GetSprite(id), placeholder fallback |
| 2.3 | `engine/assets/fonts.go` | GetFont(name), TTF loading |
| 2.4 | `assets/sprites/manifest.json` | Initial sprite ID mappings |
| 2.5 | `assets/sprites/*.png` | Test sprites (4 biome tiles) |
| 2.6 | `engine/render/draw.go` | Use AssetManager instead of hardcoded colors |
| 2.7 | `cmd/game/main.go` | Initialize AssetManager at startup |

**Estimated:** 1-2 sessions

### Phase 3: Display Configuration (P1)

Implement display settings from [display-config.md](../design_docs/planned/v0_2_0/display-config.md).

| Task | File | Description |
|------|------|-------------|
| 3.1 | `engine/display/config.go` | Config struct, Load(), Save() |
| 3.2 | `engine/display/manager.go` | Manager, Layout(), ToggleFullscreen() |
| 3.3 | `cmd/game/main.go` | Use DisplayManager, handle F11 |
| 3.4 | Test | Verify fullscreen toggle, config persistence |

**Estimated:** 1 session

### Phase 4: Verification & Cleanup

| Task | Description |
|------|-------------|
| 4.1 | Run `make eval`, ensure all scenarios pass |
| 4.2 | Test game visually (`make run`) |
| 4.3 | Update design docs status (move to implemented/ if complete) |
| 4.4 | Document any issues for AILANG feedback |

**Estimated:** 0.5 session

## Technical Details

### Mock sim_gen Types

Based on [protocol.ail](../sim/protocol.ail), the Go types should be:

```go
// draw_cmd.go - Sum type using interface
type DrawCmd interface {
    isDrawCmd()
}

type DrawCmdRect struct {
    X, Y, W, H float64
    Color, Z   int
}
func (DrawCmdRect) isDrawCmd() {}

type DrawCmdSprite struct {
    ID   int
    X, Y float64
    Z    int
}
func (DrawCmdSprite) isDrawCmd() {}

type DrawCmdText struct {
    Text string
    X, Y float64
    Z    int
}
func (DrawCmdText) isDrawCmd() {}
```

### Asset Directory Structure

```
assets/
├── sprites/
│   ├── manifest.json
│   ├── tile_water.png      (16x16, blue)
│   ├── tile_forest.png     (16x16, green)
│   ├── tile_desert.png     (16x16, tan)
│   └── tile_mountain.png   (16x16, brown)
└── fonts/
    └── (optional for v0.2.0)
```

### Config File Location

```
./config.json   (game directory, checked first)
~/.config/stapledons_voyage/config.json  (fallback)
```

## Dependencies

| Package | Version | Purpose |
|---------|---------|---------|
| `ebiten/v2` | v2.6.0 | Already in go.mod |
| `ebiten/v2/text/v2` | - | Font rendering (Phase 2) |
| `image/png` | stdlib | PNG loading |

## Risks & Mitigations

| Risk | Mitigation |
|------|------------|
| Mock types diverge from AILANG output | Review `ailang prompt` before finalizing; types are simple |
| Asset loading blocks startup | Use goroutine for background loading (future) |
| Ebiten API changes | Pin to v2.6.0, test before upgrading |

## AILANG Feedback Checkpoint

After this sprint, report:
- [ ] Any AILANG documentation gaps discovered while creating mocks
- [ ] Suggestions for generated Go code structure
- [ ] Feature requests that would help engine integration

## Notes

- This sprint intentionally avoids AILANG compiler dependency
- When `ailang compile --emit-go` ships, delete `sim_gen/` and regenerate
- Mock types should be minimal - just enough to make engine work
- Focus on interface contracts, not game logic (that stays in AILANG)
