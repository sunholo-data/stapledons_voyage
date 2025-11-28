# Design Document Structure Guide

Complete reference for Stapledons Voyage game design documents.

## Template Overview

Game design docs should cover:
1. **Game feature** - What gameplay does this enable?
2. **AILANG implementation** - Types, functions, effects needed
3. **Engine integration** - How Go/Ebiten renders this
4. **AILANG constraints** - Known limitations and workarounds
5. **Testing** - How to verify it works

## Header Section

```markdown
# [Feature Name]

**Status**: Planned | Implemented
**Target**: v0.1.0
**Priority**: P0 (High) | P1 (Medium) | P2 (Low)
**Complexity**: Simple | Medium | Complex
**AILANG Workarounds**: [List any known limitations to navigate]
```

## Problem Statement

```markdown
## Problem Statement

[What game feature is missing? Why is it needed for gameplay?]

**Current State:**
- [What exists now]
- [What's missing]
- [Impact on gameplay]
```

## AILANG Implementation Section

```markdown
## AILANG Implementation

### Types (sim/world.ail or sim/protocol.ail)

```ailang
-- New types needed
type Direction = North | South | East | West
```

### Functions (sim/npc_ai.ail or sim/step.ail)

```ailang
-- New functions needed
export pure func move(npc: NPC, dir: Direction) -> NPC
```

### AILANG Constraints

**Known limitations affecting this feature:**
- [ ] Module imports not working - duplicate types locally
- [ ] Recursion depth - limit grid operations
- [ ] No RNG - use deterministic seed-based approach
- [ ] No Array - use list with O(n) access

**Workarounds planned:**
- [How you'll work around each limitation]
```

## Engine Integration Section

```markdown
## Engine Integration

### Data Flow
AILANG step() → World state → Go engine → Ebiten rendering

### Go Changes (engine/)
- `engine/render/render.go` - [Changes needed]

### DrawCmd Usage
```ailang
let sprite = Sprite(id, x, y, z)
```
```

## Testing Strategy

```markdown
## Testing Strategy

### AILANG Testing
```bash
ailang check sim/<file>.ail
ailang run --entry <func> sim/step.ail
```

### Runtime Testing
```bash
make run
```

### Edge Cases
- [ ] Case 1
- [ ] Case 2
```

## AILANG Feedback Section

```markdown
## AILANG Feedback

**Issues to report after implementation:**

| Type | Title | Description |
|------|-------|-------------|
| bug | [Title] | [What went wrong] |
| feature | [Title] | [What would have helped] |
| docs | [Title] | [What was unclear] |
```

## Success Criteria

```markdown
## Success Criteria

### AILANG
- [ ] All sim/*.ail files pass `ailang check`
- [ ] No recursion overflow
- [ ] Types documented

### Gameplay
- [ ] Feature visible in game
- [ ] Performance acceptable

### Feedback
- [ ] Issues documented
- [ ] Workarounds noted in CLAUDE.md
```

## Example: NPC Movement

```markdown
# NPC Movement System

**Status**: Planned
**Priority**: P1
**Complexity**: Medium
**AILANG Workarounds**: No RNG, duplicate types

## Problem Statement
NPCs are defined but never move.

## AILANG Implementation

### Types
```ailang
type Direction = North | South | East | West
```

### Functions
```ailang
export pure func move(npc: NPC, dir: Direction) -> NPC {
    let newX = match dir {
        East => npc.pos.x + 1,
        West => npc.pos.x - 1,
        _ => npc.pos.x
    };
    { id: npc.id, pos: { x: newX, y: npc.pos.y } }
}
```

### Constraints
- No RNG: Movement deterministic
- Duplicate types: Coord defined locally

## Success Criteria
- [ ] NPCs move each tick
- [ ] `ailang check` passes
- [ ] Feedback sent for issues
```

## Common Mistakes

### 1. Forgetting AILANG Constraints
❌ Design assuming features that don't exist
✅ Check CLAUDE.md limitations first

### 2. No Workaround Plan
❌ "Use random numbers"
✅ "Use tick-based values since no RNG"

### 3. Missing Feedback Loop
❌ Finish and move on
✅ Document issues, report to AILANG team

## File Organization

```
design_docs/
├── planned/
│   ├── npc-movement.md
│   └── v0_1_0/
├── implemented/
│   └── v0_1_0/
└── README.md
```
