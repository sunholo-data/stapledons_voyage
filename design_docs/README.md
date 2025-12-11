# Stapledon's Voyage Design Documents

This directory is the **source of truth** for feature planning and implementation status.

## Implementation Order

Features are organized into **phases** based on dependencies. Work through phases in order.

```
Phase 0 (Architecture)  ──►  Phase 1 (Data Models)  ──►  Phase 2 (Core Views)
         │                           │                          │
         │                           ▼                          ▼
         │                   Phase 3 (Gameplay)  ◄────────────────
         │                           │
         │                           ▼
         └───────────────►  Phase 4 (Polish)
```

## Quick Reference

| Phase | Status | Count | Location | Purpose |
|-------|--------|-------|----------|---------|
| **Phase 0** | **IN PROGRESS (50%)** | 2 | [planned/phase0-architecture/](planned/phase0-architecture/) | Fix AILANG-first architecture |
| Phase 1 | Planned | 3 | [planned/phase1-data-models/](planned/phase1-data-models/) | Galaxy, planet, ship data |
| Phase 2 | Planned | 3 | [planned/phase2-core-views/](planned/phase2-core-views/) | Navigation & exploration UIs |
| Phase 3 | Planned | 1 | [planned/phase3-gameplay/](planned/phase3-gameplay/) | Journey system (core mechanic) |
| Phase 4 | Planned | 7 | [planned/phase4-polish/](planned/phase4-polish/) | Cinematics & polish |
| Future | Backlog | 33 | [planned/future/](planned/future/) | Long-term features |
| Implemented | Done | 19 | [implemented/](implemented/) | Completed features |
| Reference | N/A | 13 | [reference/](reference/) | Architecture docs |

## Directory Structure

```
design_docs/
├── planned/
│   ├── phase0-architecture/   # MUST DO FIRST - blocks everything
│   ├── phase1-data-models/    # Galaxy, planet, ship structures
│   ├── phase2-core-views/     # Galaxy map, ship exploration, bridge
│   ├── phase3-gameplay/       # Journey system (the heart of the game)
│   ├── phase4-polish/         # Arrival cinematics, camera systems
│   └── future/                # Backlog (no phase assigned)
├── implemented/
│   ├── v0_1_0/                # Engine, shaders, audio, save system
│   └── v0_2_0/                # Tetra3D, 3D planets
├── reference/                 # Architecture & system docs
├── input/                     # Research notes, concepts
└── archive/                   # Superseded docs
```

---

## Phase 0: Architecture (BLOCKING)

**Status:** IN PROGRESS (50%)
**Sprint:** [bridge-dome-migration-sprint.md](../sprints/bridge-dome-migration-sprint.md)

This phase fixes the architectural debt where game logic exists in both Go and AILANG.

| Doc | Description | Sprint? |
|-----|-------------|---------|
| [view-layer-ailang-migration](planned/phase0-architecture/view-layer-ailang-migration.md) | Move view state to AILANG | YES (50%) |
| [view-types-cleanup](planned/phase0-architecture/view-types-cleanup.md) | Delete duplicate types | NO |

**Why first?** Until this is complete, new features get implemented in the wrong layer.

---

## Phase 1: Data Models

**Status:** Planned
**Depends on:** Phase 0 complete

| Doc | Description | Sprint? |
|-----|-------------|---------|
| [starmap-data-model](planned/phase1-data-models/starmap-data-model.md) | Galaxy structure, star systems | NO |
| [planet-data-migration](planned/phase1-data-models/planet-data-migration.md) | Planet properties, orbits | NO |
| [ship-structure](planned/phase1-data-models/ship-structure.md) | Deck layouts, rooms | NO |

**Why second?** These are the "nouns" - you can't build views without data.

---

## Phase 2: Core Views

**Status:** Planned
**Depends on:** Phase 1 complete

| Doc | Description | Sprint? |
|-----|-------------|---------|
| [galaxy-map](planned/phase2-core-views/galaxy-map.md) | Star navigation, pan/zoom | NO |
| [ship-exploration](planned/phase2-core-views/ship-exploration.md) | Deck traversal | NO |
| [02-bridge-interior](planned/phase2-core-views/02-bridge-interior.md) | Bridge layout, consoles | YES (0%) |

**Why third?** These are the interfaces players interact with.

---

## Phase 3: Core Gameplay

**Status:** Planned
**Depends on:** Phase 2 complete (galaxy map)

| Doc | Description | Sprint? |
|-----|-------------|---------|
| [journey-system](planned/phase3-gameplay/journey-system.md) | Time dilation, commits | NO |

**Why here?** This is the **heart of Stapledon's Voyage** - the unique mechanic.

---

## Phase 4: Polish

**Status:** Planned
**Depends on:** Phase 3 complete

| Doc | Description | Sprint? |
|-----|-------------|---------|
| [arrival-sequence](planned/phase4-polish/arrival-sequence.md) | Planet approach | YES (40%) |
| [cinematic-arrival-system](planned/phase4-polish/cinematic-arrival-system.md) | Cinematic framework | NO |
| [tetra3d-planet-rendering](planned/phase4-polish/tetra3d-planet-rendering.md) | 3D planets | YES (0%) |
| [camera-*](planned/phase4-polish/) | Camera systems | NO |

**Why last?** Polish comes after core gameplay works.

---

## Implemented Features

### v0.2.0
| Feature | Doc |
|---------|-----|
| Tetra3D Integration | [02-tetra3d-integration](implemented/v0_2_0/02-tetra3d-integration.md) |
| 3D Sphere Planets | [03-3d-sphere-planets](implemented/v0_2_0/03-3d-sphere-planets.md) |
| Dome State Migration | [dome-state-migration](implemented/v0_2_0/dome-state-migration.md) |

### v0.1.0
| Feature | Doc |
|---------|-----|
| SR/GR Effects | [sr-effects](implemented/v0_1_0/sr-effects.md), [gr-effects](implemented/v0_1_0/gr-effects.md) |
| Shader Pipeline | [shader-system](implemented/v0_1_0/shader-system.md) |
| Audio System | [audio-system](implemented/v0_1_0/audio-system.md) |
| Save/Load | [save-load-system](implemented/v0_1_0/save-load-system.md) |
| Isometric Engine | [isometric-engine](implemented/v0_1_0/isometric-engine.md) |
| AI Handlers | [ai-handler-system](implemented/v0_1_0/ai-handler-system.md) |
| CLI Dev Tools | [cli-dev-tools](implemented/v0_1_0/cli-dev-tools.md) |

See [implemented/](implemented/) for all 19 completed feature docs.

---

## Reference Documents

| Document | Description |
|----------|-------------|
| [architecture](reference/architecture.md) | System architecture |
| [engine-layer](reference/engine-layer.md) | Go/Ebiten engine design |
| [ailang-integration](reference/ailang-integration.md) | AILANG integration patterns |
| [engine-capabilities](reference/engine-capabilities.md) | What the engine can do |
| [sprint-vision-integration](reference/sprint-vision-integration.md) | Vision-aligned sprint plan |

See [reference/](reference/) for all 13 reference documents.

---

## Future Backlog

33 features planned for future phases. Highlights:
- Crew psychology, dialogue system
- Black holes, civilization trade
- Endgame legacy, narrative orchestrator
- Archive AI system

See [planned/future/](planned/future/) for full backlog.

---

## Workflow

1. **Work through phases in order** (0 → 1 → 2 → 3 → 4)
2. **Create sprint** with `sprint-planner` skill before implementing
3. **Execute sprint** with `sprint-executor` skill
4. **Move to implemented/** when complete:
   ```bash
   git mv design_docs/planned/phaseX/feature.md design_docs/implemented/vX_Y_Z/
   ```

## Validation

```bash
# Audit design docs and sprints
.claude/skills/game-architect/scripts/audit_design_docs.sh

# Check architecture
.claude/skills/game-architect/scripts/validate_all.sh
```

## Creating Design Documents

Use the `design-doc-creator` skill:
```
invoke design-doc-creator skill
```

See [.claude/skills/design-doc-creator/](../.claude/skills/design-doc-creator/) for templates.
