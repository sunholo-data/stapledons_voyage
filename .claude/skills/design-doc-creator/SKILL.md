---
name: Game Design Doc Creator
description: Create design documents for Stapledons Voyage game features. Use when user asks to create a design doc, plan a feature, or document game mechanics. Handles both planned/ and implemented/ docs.
---

# Game Design Doc Creator

Create well-structured design documents for Stapledons Voyage game features.

## Quick Start

**Most common usage:**
```bash
# User says: "Create a design doc for NPC pathfinding"
# This skill will:
# 1. Ask for key details (priority, complexity)
# 2. Create design_docs/planned/npc-pathfinding.md
# 3. Fill template with proper structure
# 4. Guide you through customization
```

## When to Use This Skill

Invoke this skill when:
- User asks to "create a design doc" or "plan a feature"
- Before implementing new game mechanics
- Documenting world generation algorithms
- Planning NPC AI behaviors
- Designing rendering/visual features

## Available Scripts

### `scripts/create_planned_doc.sh <doc-name> [version]`
Create a new design document in `design_docs/planned/`.

### `scripts/move_to_implemented.sh <doc-name> <version>`
Move a design document from planned/ to implemented/ after completion.

## Workflow

### 1. Gather Requirements

Ask user:
- What game feature are you designing?
- What game systems does it affect? (world, NPCs, rendering, input)
- Estimated complexity? (simple, medium, complex)
- Any AILANG limitations to work around?

### 2. Game Vision Alignment

**Every feature should be scored against the game's core pillars:**

| Pillar | Question |
|--------|----------|
| **Time Dilation Consequence** | Does this reinforce irreversible time choices? |
| **Civilization Simulation** | Does it enhance galaxy-scale simulation? |
| **Philosophical Depth** | Does it add moral/philosophical decisions? |
| **Ship & Crew Life** | Does it connect to finite crew narrative? |
| **Legacy Impact** | Does it contribute to Year 1,000,000 report? |
| **Hard Sci-Fi Authenticity** | Does it maintain scientific plausibility? |

**Feature types:**
- **Gameplay features** should score positively on multiple pillars
- **Engine/Infrastructure** features can score N/A on most pillars (they're enabling tech)
- **No feature** should score negatively on any pillar (violates game vision)

**Reference:** [docs/game-vision.md](../../../docs/game-vision.md)

### 3. Consider AILANG Constraints

**Important for this project:** All game logic is written in AILANG. Consider:
- No mutable state - must use functional updates
- No loops - must use recursion (with depth limits!)
- Limited data structures - lists only, no arrays
- Known issues - check CLAUDE.md for current limitations

### 4. Design Doc Structure

**Game-specific sections:**
- **Game Vision Alignment**: Score against core pillars
- **Feature Overview**: What gameplay does this enable?
- **AILANG Implementation**: Types, functions, effects needed
- **Engine Integration**: How Go/Ebiten renders this
- **Performance**: Recursion depth, list operations needed
- **Testing**: How to verify the feature works

### 5. Example: NPC Movement Design Doc

```markdown
# NPC Movement System

## Status
- Status: Planned
- Priority: P1
- Estimated: 2 days

## Feature Overview
NPCs should move around the world grid, avoiding obstacles.

## AILANG Implementation

### Types (in sim/world.ail)
```ailang
type Direction = North | South | East | West
type MoveResult = Moved(Coord) | Blocked(string)
```

### Functions (in sim/npc_ai.ail)
```ailang
export pure func move(npc: NPC, dir: Direction, world: World) -> MoveResult
export pure func pathfind(npc: NPC, target: Coord, world: World) -> [Direction]
```

## Engine Integration
- Go code reads NPC positions from World state
- Renders sprites at grid positions * tile size

## Performance Concerns
- Pathfinding recursion: A* with depth limit
- List operations: O(n) for each tile lookup

## Success Criteria
- [ ] NPCs can move in 4 directions
- [ ] Movement blocked by obstacles
- [ ] No recursion overflow with 64x64 grid
```

## AILANG Feedback Integration

**If you encounter AILANG limitations while designing:**

1. Note the limitation in the design doc
2. Design a workaround
3. Report to AILANG core via `ailang-feedback` skill:
   ```bash
   ~/.claude/skills/ailang-feedback/scripts/send_feedback.sh feature \
     "Feature needed for <game feature>" \
     "Description of what would help" \
     "stapledons_voyage"
   ```

## Document Locations

```
design_docs/
├── planned/              # Future features
│   ├── feature.md
│   └── v0_1_0/           # Targeted for game v0.1.0
├── implemented/          # Completed features
│   └── v0_1_0/
└── README.md             # Feature index
```

## Best Practices

### 1. Start with AILANG Types
Define your data structures first - they drive the implementation.

### 2. Plan for Recursion Limits
Any operation on 64x64 grid (4096 tiles) needs careful design.

### 3. Test with `ailang check` Early
Type-check your planned code snippets before committing to the design.

### 4. Link to `ailang prompt`
Reference `ailang prompt` output when documenting AILANG syntax.

## Notes

- All game logic is AILANG, engine is Go/Ebiten
- Design docs should include AILANG code snippets
- Document workarounds for AILANG limitations
- Report feature requests to AILANG core
