---
name: Game Sprint Executor (AILANG)
description: Execute approved sprint plans for Stapledons Voyage using AILANG. ALL game logic must be AILANG - engine is rendering only. Use when user says "execute sprint", "start implementation", or wants to begin an approved sprint plan.
---

# Game Sprint Executor (AILANG-Only)

Execute an approved sprint plan with **AILANG as the primary implementation language**. All game logic, state machines, and gameplay code goes in `sim/*.ail` - the Go engine only renders DrawCmd output.

## MANDATORY: Sprint JSON Tracking

**Every sprint MUST have a tracking JSON file in `sprints/`:**

```bash
# Location: sprints/sprint-<sprint-id>.json
# Example: sprints/sprint-arrival-sequence.json
```

**Create immediately when starting sprint:**
```json
{
  "sprint_id": "<sprint-id>",
  "name": "Sprint Name",
  "status": "in_progress",
  "type": "ailang",
  "started": "YYYY-MM-DD",
  "design_doc": "design_docs/planned/next/<feature>.md",
  "phases": [
    {
      "id": "phase1",
      "name": "Phase Name",
      "status": "in_progress",
      "tasks": [
        {"id": "1.1", "name": "Task Name", "status": "pending", "notes": ""}
      ]
    }
  ],
  "ailang_issues": [],
  "dx_issues_discovered": [],
  "lessons_learned": [],
  "notes": []
}
```

**AILANG sprints MUST include these sections** (added to standard format):
- `ailang_issues` - Bugs/features reported to AILANG team
- `dx_issues_discovered` - DX friction discovered during sprint
- `lessons_learned` - What to do differently next time

**Update this JSON:**
- Mark milestones as `completed` immediately after finishing
- Add notes with workarounds used
- Document ALL DX issues as you discover them
- Add lessons learned in real-time, not at sprint end

## Pre-Sprint DX Checklist

Before writing any code, verify these items:

### File Boundaries
- [ ] **sim_gen/*.go is GENERATED** - Never edit (except manual workarounds marked with PATCH)
- [ ] **engine/*.go is safe** - OK to edit for rendering
- [ ] **sim/*.ail is source** - Primary edit location

### For Visual Features
- [ ] **Screenshot testing ready** - Can run `go run ./cmd/game -screenshot N -output /tmp/test.png`
- [ ] **Test at multiple frames** - Don't declare rendering complete until tested
- [ ] **Use real assets from start** - No placeholder rectangles

### For Downloaded Assets
- [ ] **Verify with `file` command** - Catch HTML error pages masquerading as images
- [ ] **Check URLs still work** - NASA/external URLs may return 404

### DrawCmd Limitations
- [ ] **Color is INDEX not RGBA** - Use biomeColors[] indices or RectRGBA/CircleRGBA for custom colors
- [ ] **Renderer is isometric** - Use direct drawing for screen-space elements
- [ ] **Use RectRGBA/CircleRGBA for screen-space** - These bypass camera transforms and use packed 0xRRGGBBAA colors

### Screen Resolution
- [ ] **NEVER hardcode 640x480 or other fixed sizes** - Use `display.InternalWidth` and `display.InternalHeight`
- [ ] **Internal resolution is 1280x960** - All screen-space coordinates should be relative to this
- [ ] **Use percentages for positioning** - e.g., `screenW * 0.5` for center, not fixed `320`

### Physics/Shader Effects
- [ ] **Use artistic license on values** - Physically accurate may not be visually playable
- [ ] **Test SR blueshift >0.5c** - May wash out all content
- [ ] **Test GR at high intensity** - May render pure black

### Build & Test
- [ ] **Rebuild binary after Go changes** - `go build -o bin/game ./cmd/game` or use `go run`
- [ ] **Use `go run` for quick testing** - Compiles fresh each time
- [ ] **Use `make game` for release** - Ensures clean build

## CRITICAL: AILANG-First Architecture

**This project is ALL-IN on AILANG.** The architecture is:

| Layer | Language | Responsibility | Edit? |
|-------|----------|----------------|-------|
| `sim/*.ail` | AILANG | ALL game logic | ✅ Primary work |
| `sim_gen/*.go` | Generated | AILANG → Go output | ❌ Never edit |
| `engine/*.go` | Go | Rendering DrawCmd only | ⚠️ Rare, rendering only |

**If you're writing game logic in Go, STOP. Write it in AILANG.**

Mock-only mode was for early prototyping. We are past that phase.

## Quick Start

**Most common usage:**
```bash
# User says: "Execute the arrival sequence sprint"
# This skill will:
# 1. Check AILANG modules compile
# 2. Write new AILANG types and functions
# 3. Run ailang check after every change
# 4. Report AILANG issues immediately
# 5. Compile to Go and test in engine
```

## When to Use This Skill

Invoke this skill when:
- User says "execute sprint", "start sprint", "begin implementation"
- User has an approved sprint plan ready
- Implementing ANY game feature (it MUST be AILANG)

**Do NOT use this skill for:**
- Pure engine rendering changes (use dev-tools or direct edits)
- Asset pipeline work (use asset-manager skill)

## Core Principles

1. **AILANG-First**: ALL game logic in `sim/*.ail` - no exceptions
2. **Test-Driven**: Run `ailang check` after every change
3. **Feedback-First**: Report AILANG issues immediately
4. **Track Progress**: Use TodoWrite AND update sprint JSON
5. **Document Workarounds**: Record how you navigated limitations
6. **Physics Accuracy First**: Use exact formulas, document any approximations

## JSON Progress Tracking (IMPORTANT)

**Always update the sprint JSON file as you work.** This enables session continuity.

```bash
# Location: sprints/<sprint-id>.json

# Update task status (after completing each task)
.claude/skills/sprint-executor/scripts/update_progress.sh \
  sprints/<sprint-id>.json task <task_id> completed

# Update phase status (after completing all phase tasks)
.claude/skills/sprint-executor/scripts/update_progress.sh \
  sprints/<sprint-id>.json phase <phase_id> completed

# Update sprint status
.claude/skills/sprint-executor/scripts/update_progress.sh \
  sprints/<sprint-id>.json sprint in_progress

# Show current progress
.claude/skills/sprint-executor/scripts/update_progress.sh \
  sprints/<sprint-id>.json show
```

**Status values:** `pending` | `in_progress` | `completed` | `blocked`

### When to Update JSON
- Mark sprint as `in_progress` when starting
- Mark each task `completed` immediately after finishing it
- Mark phase `completed` after all its tasks are done
- Mark sprint `completed` at the end

## Engine-Only Changes (Rare)

Most work should be AILANG. Engine changes are only needed for:
- Adding new DrawCmd rendering (e.g., planet textures)
- Shader modifications
- Asset loading

For engine-only work:
```bash
go build ./...    # Verify Go compiles
make run          # Test rendering
```

**Remember:** Engine code should be "dumb" - it only renders what AILANG tells it via DrawCmd.

## Execution Flow

### Phase 1: Initialize Sprint

1. **Check Status**
   ```bash
   # For AILANG sprints:
   for f in sim/*.ail; do ailang check "$f"; done
   ailang messages list --unread

   # For mock-only sprints:
   go test ./...
   make game-mock
   ```

2. **Initialize JSON Progress**
   ```bash
   .claude/skills/sprint-executor/scripts/update_progress.sh \
     sprints/<sprint-id>.json sprint in_progress
   ```

3. **Create Todo List**
   - Use TodoWrite for all milestones
   - Mark first milestone as in_progress

### Phase 2: Execute Milestones

**For each milestone:**

1. **Pre-Implementation**
   - Mark milestone as in_progress in TodoWrite
   - Re-read `ailang prompt` for syntax reference

2. **Implement AILANG Code**
   ```bash
   # After each file edit:
   ailang check sim/<file>.ail

   # If errors, fix immediately
   # If stuck, report to AILANG team
   ```

3. **Test with ailang run**
   ```bash
   # Test entry function
   ailang run --entry <function> sim/step.ail
   ```

4. **Report Issues Encountered**
   ```bash
   # For bugs
   ~/.claude/skills/ailang-feedback/scripts/send_feedback.sh bug \
     "Issue title" "Description" "stapledons_voyage"

   # For missing features
   ~/.claude/skills/ailang-feedback/scripts/send_feedback.sh feature \
     "Feature needed" "Why it would help" "stapledons_voyage"
   ```

5. **Update Progress (IMPORTANT)**
   ```bash
   # After each task
   .claude/skills/sprint-executor/scripts/update_progress.sh \
     sprints/<sprint-id>.json task <task_id> completed

   # After completing all tasks in a phase
   .claude/skills/sprint-executor/scripts/update_progress.sh \
     sprints/<sprint-id>.json phase <phase_id> completed
   ```
   - Note workarounds used
   - Update CLAUDE.md if new limitations found

6. **Pause for Review**
   - Show progress: `.../update_progress.sh sprints/<id>.json show`
   - Ask if ready to continue

### Phase 3: Engine Integration (if needed)

1. **Compile AILANG to Go**
   ```bash
   make sim   # Generates sim_gen/*.go
   ```

2. **Test Game**
   ```bash
   make run   # Run with Ebiten
   ```

3. **Fix Integration Issues**
   - Check generated Go code
   - Update engine/ if needed

### Phase 4: Finalize Sprint

1. **Final Testing**
   ```bash
   # All AILANG modules
   for f in sim/*.ail; do ailang check "$f"; done

   # Game build
   make game
   ```

2. **AILANG Feedback Summary**
   - List all issues reported during sprint
   - Note any responses received

3. **Developer Experience (DX) Report** (AILANG sprints only)

   **Skip this for mock-only sprints** - the `ailang-feedback` skill is for AILANG language feedback, not general sprint completion.

   For sprints that involve AILANG code, send a DX feedback message reflecting on the overall experience:

   ```bash
   ~/.claude/skills/ailang-feedback/scripts/send_feedback.sh dx \
     "Sprint DX: <sprint-name>" \
     "<honest reflection on working with AILANG this sprint>" \
     "stapledons_voyage"
   ```

   **Include in your reflection:**
   - **Positives**: What worked well? What felt natural?
   - **Friction points**: Where did you get stuck? What was confusing?
   - **Productivity**: Could you express your intent easily?
   - **Error messages**: Were they helpful or cryptic?
   - **Documentation**: Did `ailang prompt` have what you needed?
   - **Overall sentiment**: Would you want to use AILANG again?

   **Be honest** - negative feedback is valuable. Examples:
   - "Pattern matching felt natural and expressive"
   - "Nested field access errors were frustrating to debug"
   - "The functional style made NPC updates clean"
   - "Had to fight the type system on record updates"

4. **Sprint Report**
   - Milestones completed
   - AILANG issues encountered (if any)
   - Workarounds used (if any)
   - DX rating (1-5 stars) - for AILANG sprints only

## Error Handling

### AILANG Compilation Fails
1. Show error output
2. Check `ailang prompt` for correct syntax
3. If unclear error, report to AILANG team
4. Don't proceed until `ailang check` passes

### Recursion Depth Exceeded
1. Reduce data size for testing
2. Consider iterative workaround (if possible)
3. Report as feature request (tail recursion optimization)

### Module Import Fails
1. Duplicate type definitions locally
2. Document the duplication
3. Report import issue to AILANG team

### Feature Not Available
1. Design workaround
2. Document the workaround
3. Report as feature request

## Common AILANG Workarounds

Quick reference for known issues and their solutions:

| Problem | Error Message | Workaround |
|---------|---------------|------------|
| Nested field access | "cannot unify open record with TVar2" | Break `a.b.c` into `let b = a.b; b.c` |
| Record update with derived value | "cannot unify open record" on `{b \| pos: newPos}` | Use explicit construction: `{ field1: b.field1, pos: newPos }` |
| Module-level `let` in function | "undefined variable" | Inline constant or pass as parameter (intentional design) |
| Tuple destructuring | Parse error on `let (x, y) = pair` | Use `match pair { (x, y) => ... }` |
| ADT in inline tests | Test harness crashes | Only use primitive types in test inputs |
| DrawCmd uses color index not RGBA | N/A (works but limited) | Use direct Ebiten drawing in engine/ for custom colors. Feature requested from AILANG. |
| Go codegen fails for large modules | "format error" in generated code | Manual workaround in sim_gen/<module>.go - EXCEPTION to no-edit rule |
| **Editing sim_gen/*.go** | Changes overwritten | **NEVER edit sim_gen files** (except manual workarounds). Request features via ailang-feedback. |

### Maintaining This List (IMPORTANT)

**This project's purpose is to surface and fix AILANG issues.** Keep this workarounds table current:

**Scripts for tracking workarounds:**

```bash
# Check inbox and verify workarounds still needed
.claude/skills/sprint-executor/scripts/check_workarounds.sh

# Add a new workaround (updates both SKILL.md and CLAUDE.md)
.claude/skills/sprint-executor/scripts/add_workaround.sh \
  "Problem name" "Error message" "Workaround description"

# Mark an issue as fixed (moves to Fixed section)
.claude/skills/sprint-executor/scripts/mark_fixed.sh \
  "problem keyword" "v0.5.0"
```

**When you discover a new issue:**
1. Run `add_workaround.sh` with problem, error, and workaround
2. Report via `ailang-feedback` with detailed repro steps

**When AILANG fixes an issue:**
1. Check inbox: `ailang messages list --unread`
2. Verify fix: `ailang check sim/*.ail`
3. Run `mark_fixed.sh "<keyword>" "<version>"`
4. Remove workarounds from code where practical
5. Acknowledge: `ailang messages ack <msg-id>`

**At sprint start:**
```bash
# Full status check
.claude/skills/sprint-executor/scripts/check_workarounds.sh
```

### Record Update Pattern

When updating nested records, this pattern works:
```ailang
-- WORKS: newPos comes from parameter or fresh construction
pure func moveBox(b: Box, newPos: Point) -> Box {
    {b | pos: newPos}
}

-- FAILS: newPos derived from b.pos
pure func moveBoxBad(b: Box) -> Box {
    let oldPos = b.pos;
    let newPos = { x: oldPos.x + 1, y: oldPos.y };
    {b | pos: newPos}  -- ERROR!
}

-- WORKAROUND: explicit construction
pure func moveBoxFixed(b: Box) -> Box {
    let oldPos = b.pos;
    let newPos = { x: oldPos.x + 1, y: oldPos.y };
    { pos: newPos, size: b.size }  -- Explicit fields
}
```

## Quality Checkpoints

After each milestone:

```bash
# 1. All AILANG modules compile
for f in sim/*.ail; do ailang check "$f"; done

# 2. Entry functions run
ailang run --entry init_world sim/step.ail

# 3. Game builds (if engine changes)
make game
```

## Resources

### Engine & Physics Reference

**IMPORTANT:** Before implementing features, review these reference documents:

| Document | Contents | When to Use |
|----------|----------|-------------|
| [engine-capabilities.md](../../../design_docs/reference/engine-capabilities.md) | All DrawCmd types, effect handlers, assets, shaders, physics | Any sprint involving rendering or engine features |
| [gr-effects.md](../../../design_docs/implemented/v0_1_0/gr-effects.md) | GR physics formulas, shader uniforms, danger levels | Black hole/neutron star features |
| [ai-handler-system.md](../../../design_docs/implemented/v0_1_0/ai-handler-system.md) | AI effect, multimodal APIs, provider config | NPC dialogue, AI-driven decisions |

**Key Engine Capabilities:**

| Category | What's Available |
|----------|------------------|
| **DrawCmd** | Sprite, Rect, Text, IsoTile, IsoEntity, GalaxyBg, Star, Ui, Line, Circle, TextWrapped |
| **Effects** | Debug, Rand, Clock, AI (with Claude/Gemini/stub backends) |
| **Assets** | Sprites (animated), Audio (OGG/WAV), Fonts (TTF with scaling) |
| **Shaders** | SR warp (Doppler, aberration), GR warp (lensing, redshift), bloom, vignette, CRT |
| **Physics** | Lorentz factor, time dilation, gravitational redshift, Schwarzschild radius |

### Project Commands
- `make sim` - Compile AILANG to Go
- `make game` - Build game executable
- `make run` - Run game

### AILANG Commands
- `ailang check <file>` - Type-check
- `ailang run --entry <func> <file>` - Run with entry point
- `ailang prompt` - Syntax reference

### Feedback Commands
- `ailang messages list --unread` - Check unread messages
- `ailang messages ack <msg-id>` - Acknowledge message
- `ailang messages send <inbox> <msg>` - Send message
- `~/.claude/skills/ailang-feedback/scripts/send_feedback.sh` - Report issues

## Milestone Checklist Template

```markdown
## Milestone: [Name]

### Implementation
- [ ] AILANG types defined
- [ ] AILANG functions implemented
- [ ] `ailang check` passes
- [ ] `ailang run` works

### Testing
- [ ] Edge cases handled
- [ ] Error conditions covered

### Feedback
- [ ] Issues reported: [list]
- [ ] Workarounds documented: [list]

### Status
- [ ] Complete and ready for next milestone
```

## Notes

- Every `ailang check` failure is a learning opportunity
- Report issues early, don't wait until sprint end
- Check inbox for AILANG team responses
- This project's purpose is testing AILANG, not just building a game
- Document everything - it helps AILANG improve
