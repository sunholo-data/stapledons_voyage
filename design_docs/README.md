# Stapledon's Voyage Design Documents

This directory is the **source of truth** for feature planning and implementation status.

## Quick Reference

| Category | Count | Location | Purpose |
|----------|-------|----------|---------|
| Implemented | 13 | [implemented/](implemented/) | Completed features |
| Planned (Next) | 4 | [planned/next/](planned/next/) | Active development |
| Planned (Future) | 23 | [planned/future/](planned/future/) | Backlog |
| Reference | 12 | [reference/](reference/) | Architecture & system docs |
| Input | 4 | [input/](input/) | Research & concept notes |

See [CHANGELOG.md](../CHANGELOG.md) for release history with dates.

## Directory Structure

```
design_docs/
├── implemented/          # Completed features
│   └── v0_X_Y/           # Organized by release version
├── planned/              # Features not yet implemented
│   ├── next/             # Active development (v0.2.0 target)
│   └── future/           # Backlog (no version assigned)
├── reference/            # Architecture, system design docs
├── input/                # Research notes, concept docs
└── README.md             # This index
```

## Validation

Run the validation script to check doc organization:

```bash
./scripts/validate_design_docs.sh
```

---

## Implemented Features (v0.1.0)

| Feature | Description | Doc |
|---------|-------------|-----|
| **SR Effects** | Special relativity: aberration, Doppler, beaming | [sr-effects.md](implemented/v0_1_0/sr-effects.md) |
| **GR Effects** | Gravitational lensing near black holes | [sprint-relativistic-effects.md](implemented/v0_1_0/sprint-relativistic-effects.md) |
| **Shader Pipeline** | Bloom, vignette, CRT, chromatic aberration | [shader-system.md](implemented/v0_1_0/shader-system.md) |
| **Audio System** | OGG/WAV loading, PlaySound API | [audio-system.md](implemented/v0_1_0/audio-system.md) |
| **Screenshot Mode** | Headless capture for visual testing | [screenshot-mode.md](implemented/v0_1_0/screenshot-mode.md) |
| **Test Scenarios** | Golden file comparison for regressions | [test-scenarios.md](implemented/v0_1_0/test-scenarios.md) |
| **Save/Load** | Single-file save system (Pillar 1 compliant) | [save-load-system.md](implemented/v0_1_0/save-load-system.md) |
| **Isometric Engine** | Tile projection, sorting, view culling | [isometric-engine.md](implemented/v0_1_0/isometric-engine.md) |
| **Animation System** | Frame-based sprite animation | [animation-system.md](implemented/v0_1_0/animation-system.md) |
| **Camera/Viewport** | World transforms, viewport culling | [camera-viewport.md](implemented/v0_1_0/camera-viewport.md) |
| **Display Config** | Resolution, fullscreen, persistence | [display-config.md](implemented/v0_1_0/display-config.md) |
| **Asset Management** | Sprites, audio, fonts with manifests | [asset-management.md](implemented/v0_1_0/asset-management.md) |
| **Player Input** | Mouse handling, keyboard events | [player-interaction.md](implemented/v0_1_0/player-interaction.md) |

---

## Planned Features

### Next (v0.2.0 target)

| Feature | Description | Doc |
|---------|-------------|-----|
| **Starmap Data Model** | Galaxy structure, star systems | [starmap-data-model.md](planned/next/starmap-data-model.md) |
| **Galaxy Map** | Navigation UI, star selection | [galaxy-map.md](planned/next/galaxy-map.md) |
| **Ship Exploration** | Interior navigation, crew interaction | [ship-exploration.md](planned/next/ship-exploration.md) |
| **Journey System** | Time dilation, travel mechanics | [journey-system.md](planned/next/journey-system.md) |

### Future (Backlog)

See [planned/future/](planned/future/) for all backlog items including:
- Crew psychology, dialogue system, black holes
- UI systems, particle effects, screen transitions
- Civilization mechanics, exploration modes
- Endgame and legacy features

---

## Reference Documents

Architecture and system design documentation:

| Document | Description |
|----------|-------------|
| [architecture.md](reference/architecture.md) | Overall system architecture |
| [engine-layer.md](reference/engine-layer.md) | Go/Ebiten engine design |
| [ailang-integration.md](reference/ailang-integration.md) | AILANG language integration |

See [reference/](reference/) for all 12 reference documents.

---

## Input Documents

Research notes and concept documents (feed into planned docs):

| Document | Description |
|----------|-------------|
| [game_loop_origin.md](input/game_loop_origin.md) | Original game loop concept |
| [bubble-ship-design.md](input/bubble-ship-design.md) | Ship design inspiration |

---

## Workflow

1. **Input docs** capture ideas and research
2. When ready to implement, create a **planned doc** in `next/` or `future/`
3. When implemented, move to `implemented/vX_Y_Z/` and update status
4. Run `./scripts/validate_design_docs.sh` to verify organization

## Creating Design Documents

```bash
# Validate current state
./scripts/validate_design_docs.sh

# Move to implemented after completion
git mv design_docs/planned/next/feature.md design_docs/implemented/v0_2_0/
# Update Status: Planned → Implemented in the doc
```

See [.claude/skills/design-doc-creator/](../.claude/skills/design-doc-creator/) for the skill and templates.
