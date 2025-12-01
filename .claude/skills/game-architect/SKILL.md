---
name: game-architect
description: Validate codebase architecture and organization before releases. Use when user asks to 'check architecture', 'validate structure', 'pre-release check', or before versioning.
---

# Game Architect

Validates Stapledon's Voyage codebase against the three-layer architecture and organization standards.

## Quick Start

```bash
# Full validation (all checks - run before releases)
.claude/skills/game-architect/scripts/validate_all.sh --full

# Quick validation (core checks only - faster)
.claude/skills/game-architect/scripts/validate_all.sh --quick

# Individual checks (core)
.claude/skills/game-architect/scripts/check_file_sizes.sh       # Files under 800 lines
.claude/skills/game-architect/scripts/check_layer_boundaries.sh # No game logic in engine/
.claude/skills/game-architect/scripts/check_structure.sh        # Files in correct locations
.claude/skills/game-architect/scripts/check_import_cycles.sh    # No circular imports

# Individual checks (extended)
.claude/skills/game-architect/scripts/check_complexity.sh       # Function size, nesting
.claude/skills/game-architect/scripts/check_ailang_sync.sh      # AILANG ↔ Go types match
.claude/skills/game-architect/scripts/check_api_stability.sh    # sim_gen API unchanged
.claude/skills/game-architect/scripts/check_coverage.sh         # Test coverage
.claude/skills/game-architect/scripts/check_dependencies.sh     # Package import graph
.claude/skills/game-architect/scripts/pre_release_check.sh      # Build, tests, TODOs
```

## When to Use This Skill

Invoke this skill when:
- Before tagging a version release
- After major refactoring
- User asks "check architecture" or "validate structure"
- Periodic codebase health checks
- After adding new files/directories

## Architecture Rules

### Three-Layer Separation

| Layer | Location | Purpose | Allowed Content |
|-------|----------|---------|-----------------|
| **Source** | `sim/*.ail` | Game logic (AILANG) | Types, pure functions |
| **Simulation** | `sim_gen/*.go` | Generated/mock Go | Game logic (temporary mock) |
| **Engine** | `engine/*.go` | IO bridging | Input capture, rendering, assets |
| **Entry** | `cmd/*.go` | Wiring | Main, game loop |

### Layer Boundaries (Critical)

**engine/ must NOT contain:**
- Game logic (NPC behavior, building, actions)
- World state manipulation beyond storing current World
- Decision-making code

**sim_gen/ rules:**
- Never manually edit when using AILANG compiler
- Currently mock (hand-written) until AILANG ships
- Contains ALL game logic temporarily

### File Size Limits

| Threshold | Action |
|-----------|--------|
| > 800 lines | **Error** - must split file |
| > 600 lines | **Warning** - consider splitting |
| > 100 lines/function | **Warning** - extract functions |

## Workflow

1. **Run full validation**: `./scripts/validate_all.sh`
2. **Review violations**: Check output for ✗ markers
3. **Fix issues**: Refactor code to correct layer
4. **Re-validate**: Ensure all checks pass
5. **Proceed with release**: Once clean

## Checks Performed

### Core Checks (blocking)

| Script | Purpose |
|--------|---------|
| `check_file_sizes.sh` | Max 800 lines/file, warn at 600 |
| `check_layer_boundaries.sh` | No game logic in engine/, no rendering in sim_gen/ |
| `check_structure.sh` | Files in correct directories |
| `check_import_cycles.sh` | No circular package dependencies |
| `pre_release_check.sh` | Build, tests, TODOs, debug code |

### Extended Checks (warnings)

| Script | Purpose |
|--------|---------|
| `check_complexity.sh` | Function size (<100 lines), nesting depth, param count |
| `check_ailang_sync.sh` | AILANG types match sim_gen Go types |
| `check_api_stability.sh` | sim_gen exports haven't changed unexpectedly |
| `check_coverage.sh` | Test coverage for critical paths |
| `check_dependencies.sh` | Package import graph, layer violations |

### API Stability

The API stability check maintains a baseline of sim_gen exports:

```bash
# Update baseline after intentional API changes
.claude/skills/game-architect/scripts/check_api_stability.sh --update
```

## Resources

- [Architecture Rules](resources/architecture_rules.md) - Detailed layer boundaries and examples
