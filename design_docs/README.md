# Stapledons Voyage Design Documents

This directory contains design documentation for game features and systems.

## Directory Structure

```
design_docs/
├── planned/              # Future features (not yet implemented)
│   └── v0_2_0/           # Targeted for game v0.2.0
├── implemented/          # Completed features
│   └── v0_1_0/           # Implemented in v0.1.0
└── README.md             # This index
```

## Implemented Features (v0.1.0)

| Document | Description | Status |
|----------|-------------|--------|
| [Architecture](implemented/v0_1_0/architecture.md) | Three-layer design, data flow, build pipeline | Complete |
| [Engine Layer](implemented/v0_1_0/engine-layer.md) | Go/Ebiten input capture and rendering | Complete |
| [Evaluation System](implemented/v0_1_0/eval-system.md) | Benchmarks, scenarios, report generation | Complete |

## Planned Features

*No planned features documented yet. Use the design-doc-creator skill to add new designs.*

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
