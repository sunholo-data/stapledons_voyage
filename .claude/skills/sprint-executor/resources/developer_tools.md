# Stapledons Voyage Developer Tools

Quick reference for game development with AILANG.

## AILANG Commands

### Type Checking
```bash
# Check single file
ailang check sim/step.ail

# Check all modules
for f in sim/*.ail; do ailang check "$f"; done
```

### Running Code
```bash
# Run with entry function
ailang run --entry init_world sim/step.ail
ailang run --entry step sim/step.ail

# Run with capabilities (if needed)
ailang run --caps IO,FS file.ail
```

### REPL
```bash
# Start interactive REPL
ailang repl
```

### Syntax Reference
```bash
# Get AILANG teaching prompt
ailang prompt
```

## Agent Messaging

```bash
# Check inbox for AILANG team messages
ailang agent inbox stapledons_voyage

# Acknowledge message
ailang agent ack <msg-id>

# Send feedback
~/.claude/skills/ailang-feedback/scripts/send_feedback.sh <type> "<title>" "<desc>" "stapledons_voyage"
```

## Game Build Commands

```bash
# Compile AILANG → Go
make sim

# Build game executable
make game

# Run game directly
make run

# Run benchmarks
make eval

# Clean generated files
make clean
```

## Project Structure

```
sim/                  # AILANG game logic (edit these)
├── protocol.ail     # Core types (FrameInput, DrawCmd, etc.)
├── world.ail        # World state types
├── step.ail         # Main game logic
└── npc_ai.ail       # NPC AI logic

sim_gen/             # Generated Go (never edit)

engine/              # Go/Ebiten rendering
cmd/game/main.go     # Game entry point
```

## Common Workflows

### Adding a New Type
1. Define type in sim/world.ail or sim/protocol.ail
2. Run `ailang check sim/<file>.ail`
3. Use in step.ail or npc_ai.ail
4. Test with `ailang run`

### Implementing a Feature
1. Check `ailang prompt` for syntax
2. Check CLAUDE.md for limitations
3. Write AILANG code
4. Run `ailang check` after each change
5. Report issues via feedback skill

### Testing a Change
```bash
# 1. Type check
ailang check sim/step.ail

# 2. Run inline tests
ailang test sim/npc_ai.ail

# 3. Run entry function
ailang run --entry init_world sim/step.ail

# 4. Build and run game
make run
```

### Inline Tests (v0.4.7+)

Add inline tests for executable documentation:

```ailang
-- Syntax: tests [(input, expected), ...]
pure func isInBounds(x: int, y: int, w: int, h: int) -> bool
tests [
    ((0, 0, 10, 10), true),
    ((10, 5, 10, 10), false)
] {
    x >= 0 && x < w && y >= 0 && y < h
}
```

Run tests:
```bash
ailang test sim/npc_ai.ail
ailang test sim/            # all files in directory
```

**Limitation:** ADT constructors (North, South, etc.) crash the test harness. Only use primitive types (int, bool, string) in test inputs.

## Debugging

### AILANG Errors
- Read error message carefully
- Check `ailang prompt` for correct syntax
- Look for pattern matching issues
- Check recursion depth if runtime error

### Common Issues
| Issue | Cause | Fix |
|-------|-------|-----|
| Type inference fail | Nested field access | Break `a.b.c` into `let b = a.b; b.c` |
| Test harness panic | ADT in test input | Only use primitives in tests |
| Type mismatch | Wrong ADT constructor | Check pattern matching |
| Unknown effect | Invalid effect name | Use IO, FS, Net, Env |

## Feedback

When you hit an issue, report it:
```bash
~/.claude/skills/ailang-feedback/scripts/send_feedback.sh bug \
  "Issue title" \
  "Description of what happened" \
  "stapledons_voyage"
```
