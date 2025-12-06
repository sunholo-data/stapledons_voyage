# Stapledon's Voyage Design Documents

This directory is the **source of truth** for feature planning and implementation status.

## Quick Reference

| Category | Count | Location | Purpose |
|----------|-------|----------|---------|
| Implemented | 6 | [implemented/](implemented/) | Completed features |
| Planned | 34 | [planned/](planned/) | Features in roadmap |
| Reference | 12 | [reference/](reference/) | Architecture & system docs |
| Input | 4 | [input/](input/) | Research & concept notes |

See [CHANGELOG.md](../CHANGELOG.md) for release history with dates.

## Directory Structure

```
design_docs/
├── implemented/          # Completed features
│   └── v0_X_Y/           # Organized by release version
├── planned/              # Features not yet implemented
│   └── v0_X_Y/           # Organized by target version
├── reference/            # Architecture, system design docs
│   └── *.md              # Not versioned, always current
├── input/                # Research notes, concept docs
│   └── *.md              # User-created reference material
└── README.md             # This index
```

## Validation

Run the validation script to check doc organization:

```bash
./scripts/validate_design_docs.sh
```

This checks for:
- Misplaced docs (status doesn't match location)
- Orphan docs (not in version folders)
- Missing metadata (Status, Version, Priority)

Use `--json` for machine-readable output.

---

## Implemented Features

### v0.1.0 - Initial Release (Current)

All features in this release. See [CHANGELOG.md](../CHANGELOG.md) for details.

| Feature | Description | Doc |
|---------|-------------|-----|
| **SR Effects** | Special relativity: aberration, Doppler, beaming | [sr-effects.md](implemented/v0_1_0/sr-effects.md) |
| **GR Effects** | Gravitational lensing near black holes | [sprint-relativistic-effects.md](implemented/v0_1_0/sprint-relativistic-effects.md) |
| **Shader Pipeline** | Bloom, vignette, CRT, chromatic aberration | [shader-system.md](implemented/v0_1_0/shader-system.md) |
| **Audio System** | OGG/WAV loading, PlaySound API | [audio-system.md](implemented/v0_1_0/audio-system.md) |
| **Screenshot Mode** | Headless capture for visual testing | [screenshot-mode.md](implemented/v0_1_0/screenshot-mode.md) |
| **Test Scenarios** | Golden file comparison for regressions | [test-scenarios.md](implemented/v0_1_0/test-scenarios.md) |

---

## Planned Features

### Near Term (v0.2.0 - v0.3.0)

| Version | Feature | Priority | Doc |
|---------|---------|----------|-----|
| v0.2.0 | Asset Management | P0 | [asset-management.md](planned/v0_2_0/asset-management.md) |
| v0.2.0 | Display Config | P1 | [display-config.md](planned/v0_2_0/display-config.md) |
| v0.2.0 | Player Interaction | P1 | [player-interaction.md](planned/v0_2_0/player-interaction.md) |
| v0.3.0 | Camera & Viewport | P0 | [camera-viewport.md](planned/v0_3_0/camera-viewport.md) |
| v0.3.0 | Tilemap Rendering | P1 | [tilemap-rendering.md](planned/v0_3_0/tilemap-rendering.md) |
| v0.3.0 | World Gen Settings | P1 | [world-gen-settings.md](planned/v0_3_0/world-gen-settings.md) |
| v0.3.0 | Starmap Data Model | P1 | [starmap-data-model.md](planned/v0_3_0/starmap-data-model.md) |

### Gameplay Foundation (v0.4.0 - v0.5.0)

| Version | Feature | Priority | Doc |
|---------|---------|----------|-----|
| v0.4.0 | Player Actions | P1 | [player-actions.md](planned/v0_4_0/player-actions.md) |
| v0.4.0 | NPC Movement | P1 | [npc-movement.md](planned/v0_4_0/npc-movement.md) |
| v0.4.0 | Planet State Transitions | P1 | [planet-state-transitions.md](planned/v0_4_0/planet-state-transitions.md) |
| v0.4.5 | Engine Extensions | P0 | [engine-extensions.md](planned/v0_4_5/engine-extensions.md) |
| v0.4.5 | UI Layout Engine | P1 | [ui-layout-engine.md](planned/v0_4_5/ui-layout-engine.md) |
| v0.5.0 | UI Modes | P0 | [ui-modes.md](planned/v0_5_0/ui-modes.md) |
| v0.5.0 | Animation System | P1 | [animation-system.md](planned/v0_5_0/animation-system.md) |
| v0.5.0 | Save/Load System | P1 | [save-load-system.md](planned/v0_5_0/save-load-system.md) |
| v0.5.0 | Isometric Engine | P1 | [isometric-engine.md](planned/v0_5_0/isometric-engine.md) |

### Core Game Systems (v0.5.1 - v0.6.1)

| Version | Feature | Priority | Doc |
|---------|---------|----------|-----|
| v0.5.1 | Ship Exploration | P0 | [ship-exploration.md](planned/v0_5_1/ship-exploration.md) |
| v0.5.2 | Galaxy Map | P0 | [galaxy-map.md](planned/v0_5_2/galaxy-map.md) |
| v0.5.3 | Dialogue System | P0 | [dialogue-system.md](planned/v0_5_3/dialogue-system.md) |
| v0.6.0 | Journey System | P0 | [journey-system.md](planned/v0_6_0/journey-system.md) |
| v0.6.0 | Black Holes | P1 | [black-holes.md](planned/v0_6_0/black-holes.md) |
| v0.6.0 | Crew Psychology | P1 | [crew-psychology.md](planned/v0_6_0/crew-psychology.md) |
| v0.6.0 | Journey Planning UI | P1 | [journey-planning-ui.md](planned/v0_6_0/journey-planning-ui.md) |
| v0.6.1 | Civilization & Trade | P1 | [civilization-trade.md](planned/v0_6_1/civilization-trade.md) |
| v0.6.1 | GR Visual Mechanics | P1 | [gr-visual-mechanics.md](planned/v0_6_1/gr-visual-mechanics.md) |

### Endgame & Polish (v0.7.0 - v0.9.0)

| Version | Feature | Priority | Doc |
|---------|---------|----------|-----|
| v0.7.0 | Exploration Modes | P1 | [exploration-modes.md](planned/v0_7_0/exploration-modes.md) |
| v0.8.0 | Endgame Legacy | P0 | [endgame-legacy.md](planned/v0_8_0/endgame-legacy.md) |
| v0.9.0 | Supporting UIs | P2 | [supporting-uis.md](planned/v0_9_0/supporting-uis.md) |

---

## Reference Documents

Architecture and system design documentation (not tied to versions):

| Document | Description |
|----------|-------------|
| [architecture.md](reference/architecture.md) | Overall system architecture |
| [engine-layer.md](reference/engine-layer.md) | Go/Ebiten engine design |
| [eval-system.md](reference/eval-system.md) | Benchmark and evaluation system |
| [ailang-integration.md](reference/ailang-integration.md) | AILANG language integration |
| [ailang-testing-matrix.md](reference/ailang-testing-matrix.md) | AILANG test coverage |
| [rng-determinism.md](reference/rng-determinism.md) | Reproducible randomness |
| [ai-effect-npcs.md](reference/ai-effect-npcs.md) | AI effect for NPC behavior |
| [debug-eval-system.md](reference/debug-eval-system.md) | Debug and evaluation tools |
| [performance-externs.md](reference/performance-externs.md) | Performance-critical externals |
| [engine-integration-gaps.md](reference/engine-integration-gaps.md) | Known integration gaps |
| [engine-discriminator-adaptation.md](reference/engine-discriminator-adaptation.md) | ADT discriminator handling |
| [sprint-engine-discriminator.md](reference/sprint-engine-discriminator.md) | Sprint for discriminator work |

---

## Input Documents

Research notes, concept documents, and reference material:

| Document | Description |
|----------|-------------|
| [game_loop_origin.md](input/game_loop_origin.md) | Original game loop concept |
| [bubble-ship-design.md](input/bubble-ship-design.md) | Ship design inspiration |
| [startmaps.md](input/startmaps.md) | Starmap data sources |
| [resources.md](input/resources.md) | External resources and references |

---

## Creating Design Documents

### Using the Skill

```bash
# In Claude Code, invoke the design-doc-creator skill
# The skill will create docs in design_docs/planned/
```

### Using Scripts

```bash
# Create a new planned document
.claude/skills/design-doc-creator/scripts/create_planned_doc.sh <doc-name> [version]

# Move to implemented after completion
.claude/skills/design-doc-creator/scripts/move_to_implemented.sh <doc-name> <version>

# Validate all docs
./scripts/validate_design_docs.sh
```

## Document Requirements

Every design doc should have these metadata fields at the top:

```markdown
**Version:** 0.X.Y
**Status:** Planned | Implemented
**Priority:** P0 | P1 | P2
**AILANG Workarounds:** Description of any language limitations
```

### Required Sections

1. **Header** - Status, Priority, Complexity, AILANG Workarounds
2. **Related Documents** - Links to dependencies
3. **Overview** - Problem statement
4. **AILANG Implementation** - Types, functions
5. **Engine Integration** - Go/Ebiten code if applicable
6. **Success Criteria** - Checklist for completion

See [design_doc_structure.md](../.claude/skills/design-doc-creator/resources/design_doc_structure.md) for the full template.

---

## AILANG Integration Notes

This project is the primary **integration test for AILANG**. Design docs should:

- Document AILANG limitations encountered
- Describe workarounds used
- Track feedback sent to AILANG core (via `ailang-feedback` skill)
- Note version dependencies (e.g., "requires v0.5.1 for RNG")

See [CLAUDE.md](../CLAUDE.md) for current known limitations.
