# Stapledons Voyage Design Documents

This directory contains design documentation for game features and systems.

## Directory Structure

```
design_docs/
├── planned/              # Features not yet implemented
│   ├── v0_1_0/           # Core architecture (first milestone)
│   ├── v0_2_0/           # Engine systems
│   └── v0_3_0/           # Rendering & camera
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

### v0.3.0 - Rendering & Camera

| Document | Description | Priority |
|----------|-------------|----------|
| [Camera & Viewport](planned/v0_3_0/camera-viewport.md) | Scrolling, zoom, viewport culling | P0 |
| [Tilemap Rendering](planned/v0_3_0/tilemap-rendering.md) | Batching, atlases, performance optimization | P1 |

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
