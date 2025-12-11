# Phase 0: Architecture (MUST DO FIRST)

**Priority:** P0 - BLOCKING
**Status:** In Progress (50%)
**Sprint:** [bridge-dome-migration-sprint.md](../../../sprints/bridge-dome-migration-sprint.md)

## Why This Phase is Critical

This phase fixes the **architectural debt** that blocks all other development. Currently, game logic exists in both Go (`engine/view/`) and AILANG (`sim/*.ail`), violating the AILANG-first mandate in CLAUDE.md.

Until this is complete:
- New features get implemented in the wrong layer
- Testing is harder (logic split between languages)
- AILANG codegen benefits are lost

## Design Docs

| Doc | Description | Has Sprint? |
|-----|-------------|-------------|
| [view-layer-ailang-migration.md](view-layer-ailang-migration.md) | Move view state from Go to AILANG | YES (50%) |
| [view-types-cleanup.md](view-types-cleanup.md) | Delete duplicate types in engine/view/ | NO |

## What Gets Fixed

| Go File | Problem | Action |
|---------|---------|--------|
| `engine/view/dome_renderer.go` | Has cruise state, planet configs | Keep rendering, move state to AILANG |
| `engine/view/bridge_view.go` | Has state, Update() logic | Remove state ownership |
| `engine/view/layer.go` | Duplicate Camera, Input types | DELETE |
| `engine/view/bubble_arc.go` | Has particle state | Move particles to AILANG |

## Success Criteria

- [ ] All view state defined in `sim/*.ail`
- [ ] `engine/view/` contains < 200 lines total (helpers only)
- [ ] No Go code has Update() methods that modify game state
- [ ] Game loop is pure: input → step → render → draw

## Dependencies

- **Depends on:** Nothing (this is the foundation)
- **Blocks:** Phase 1, Phase 2, Phase 3, Phase 4 (everything!)
