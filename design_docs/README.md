# Stapledons Voyage Design Documents

This directory contains design documentation for game features and systems.

## Directory Structure

```
design_docs/
├── planned/              # Features not yet implemented
│   ├── v0_1_0/           # Core architecture
│   ├── v0_2_0/           # Engine systems
│   ├── v0_3_0/           # Rendering & camera
│   ├── v0_4_0/           # Gameplay features
│   ├── v0_4_5/           # Engine extensions (prerequisites)
│   ├── v0_5_0/           # UI modes architecture
│   ├── v0_5_1/           # Ship exploration
│   ├── v0_5_2/           # Galaxy map
│   ├── v0_5_3/           # Dialogue system
│   ├── v0_6_0/           # Journey system
│   ├── v0_6_1/           # Civilization & trade
│   ├── v0_7_0/           # Exploration modes
│   ├── v0_8_0/           # Endgame
│   └── v0_9_0/           # Supporting UIs
├── implemented/          # Completed features (empty for now)
└── README.md             # This index
```

## Planned Features

### v0.1.0 - Core Architecture (First Milestone)

| Document | Description | Priority |
|----------|-------------|----------|
| [Architecture](planned/v0_1_0/architecture.md) | Three-layer design, data flow, build pipeline | P0 |
| [Engine Layer](planned/v0_1_0/engine-layer.md) | Go/Ebiten input capture and rendering | P0 |
| [Evaluation System](planned/v0_1_0/eval-system.md) | Benchmarks, scenarios, report generation | P1 |

### v0.2.0 - Core Engine Systems

| Document | Description | Priority |
|----------|-------------|----------|
| [Asset Management](planned/v0_2_0/asset-management.md) | Sprite, font, and sound loading | P0 |
| [Audio System](planned/v0_2_0/audio-system.md) | Sound effects and music playback | P1 |
| [Display Configuration](planned/v0_2_0/display-config.md) | Resolution, fullscreen, settings persistence | P1 |
| [Player Interaction](planned/v0_2_0/player-interaction.md) | Click-to-select, tile highlighting, input handling | P1 |

### v0.3.0 - Rendering & Camera

| Document | Description | Priority |
|----------|-------------|----------|
| [Camera & Viewport](planned/v0_3_0/camera-viewport.md) | Scrolling, zoom, viewport culling | P0 |
| [Tilemap Rendering](planned/v0_3_0/tilemap-rendering.md) | Batching, atlases, performance optimization | P1 |

### v0.4.0 - Gameplay Features

| Document | Description | Priority |
|----------|-------------|----------|
| [Player Actions](planned/v0_4_0/player-actions.md) | Inspect, build, clear actions on selected tiles | P1 |
| [NPC Movement](planned/v0_4_0/npc-movement.md) | NPC spawning, movement patterns, AI foundation | P1 |

### v0.4.5 - Engine Extensions (Prerequisites)

| Document | Description | Priority |
|----------|-------------|----------|
| [Engine Extensions](planned/v0_4_5/engine-extensions.md) | DrawCmdLine, TextWrapped, Circle, font sizes - required before UI modes | P0 |

### v0.5.0 - UI Modes Architecture

| Document | Description | Priority |
|----------|-------------|----------|
| [UI Modes](planned/v0_5_0/ui-modes.md) | WorldMode state machine, all 10+ UI surfaces, mode transitions | P0 |

### v0.5.1 - Ship Exploration

| Document | Description | Priority |
|----------|-------------|----------|
| [Ship Exploration](planned/v0_5_1/ship-exploration.md) | Multi-deck ship interior, player movement, crew interaction, room system | P0 |

### v0.5.2 - Galaxy Map

| Document | Description | Priority |
|----------|-------------|----------|
| [Galaxy Map](planned/v0_5_2/galaxy-map.md) | Pan/zoom starfield, civilization nodes, contact network, time dilation preview | P0 |

### v0.5.3 - Dialogue System

| Document | Description | Priority |
|----------|-------------|----------|
| [Dialogue System](planned/v0_5_3/dialogue-system.md) | Branching conversations, portrait system, choice effects, crew dialogues | P0 |

### v0.6.0 - Journey System

| Document | Description | Priority |
|----------|-------------|----------|
| [Journey System](planned/v0_6_0/journey-system.md) | Journey planning, time dilation calculator, crew projection, journey events, arrival sequence | P0 |
| [SR Rendering](planned/sr-rendering.md) | Special relativity visual effects: aberration, Doppler, beaming | P1 |
| [SR Rendering Go](planned/sr-rendering-go.md) | Go implementation: relativity package, Kage shader, renderer integration | P1 |

### v0.6.1 - Civilization & Trade

| Document | Description | Priority |
|----------|-------------|----------|
| [Civilization & Trade](planned/v0_6_1/civilization-trade.md) | Civilization detail screen, philosophy system, trade UI, impact preview | P1 |

### v0.7.0 - Exploration Modes

| Document | Description | Priority |
|----------|-------------|----------|
| [Exploration Modes](planned/v0_7_0/exploration-modes.md) | Planet surface mode, ruins/archaeology, artifacts, ancient logs | P1 |

### v0.8.0 - Endgame

| Document | Description | Priority |
|----------|-------------|----------|
| [Endgame Legacy](planned/v0_8_0/endgame-legacy.md) | Fast-forward to Year 1,000,000, network comparison, victory scoring, counterfactuals, epilogue | P0 |

### v0.9.0 - Supporting UIs

| Document | Description | Priority |
|----------|-------------|----------|
| [Supporting UIs](planned/v0_9_0/supporting-uis.md) | Logbook, crew sociogram, tech inventory, philosophy browser, time comparison | P2 |

## Creating New Design Documents

Use the design-doc-creator skill:

```bash
# In Claude Code, invoke the skill when asked to plan a feature
# The skill will create docs in design_docs/planned/
```

Or use the provided scripts:

```bash
# Create a new planned document
.claude/skills/design-doc-creator/scripts/create_planned_doc.sh <doc-name> [version]

# Move to implemented after completion
.claude/skills/design-doc-creator/scripts/move_to_implemented.sh <doc-name> <version>
```

## Document Template

See [.claude/skills/design-doc-creator/resources/design_doc_structure.md](../.claude/skills/design-doc-creator/resources/design_doc_structure.md) for the full template.

**Required sections for game design docs:**
1. Header (Status, Priority, Complexity, AILANG Workarounds)
2. Related Documents
3. Overview / Problem Statement
4. AILANG Implementation (types, functions)
5. Engine Integration (if applicable)
6. AILANG Constraints (limitations and workarounds)
7. Success Criteria (checklists)

## AILANG Integration Notes

This project is primarily an **AILANG integration test**. Design documents should:

- Document AILANG limitations encountered
- Describe workarounds used
- Track feedback sent to AILANG core
- Note version dependencies (e.g., "requires v0.5.1 for RNG")

See [CLAUDE.md](../CLAUDE.md) for current known limitations.
