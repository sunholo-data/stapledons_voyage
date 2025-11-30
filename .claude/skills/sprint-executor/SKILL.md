---
name: Game Sprint Executor
description: Execute approved sprint plans for Stapledons Voyage with continuous testing, AILANG feedback, and progress tracking. Use when user says "execute sprint", "start implementation", or wants to begin an approved sprint plan.
---

# Game Sprint Executor

Execute an approved sprint plan with continuous AILANG testing, feedback reporting, and progress tracking.

## Quick Start

**Most common usage:**
```bash
# User says: "Execute the NPC movement sprint"
# This skill will:
# 1. Check AILANG modules compile
# 2. Create TodoWrite tasks for milestones
# 3. Implement each milestone with ailang check
# 4. Report AILANG issues as encountered
# 5. Pause after each milestone for review
```

## When to Use This Skill

Invoke this skill when:
- User says "execute sprint", "start sprint", "begin implementation"
- User has an approved sprint plan ready
- User wants guided game development with AILANG testing

## Core Principles

1. **Test-Driven**: Run `ailang check` after every change (or `go test` for mock-only sprints)
2. **Feedback-First**: Report AILANG issues immediately
3. **Pause Points**: Stop after each milestone for review
4. **Track Progress**: Use TodoWrite AND update sprint JSON
5. **Document Workarounds**: Record how you navigated limitations

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

## Mock-Only Sprints (No AILANG)

When AILANG compiler isn't ready, use mock sprints:

```bash
# Use -mock targets
make game-mock    # Build without AILANG
make run-mock     # Run without AILANG
make eval-mock    # Test without AILANG
go test ./...     # Run Go tests
```

For mock sprints, skip AILANG-specific steps and focus on Go implementation.

## Execution Flow

### Phase 1: Initialize Sprint

1. **Check Status**
   ```bash
   # For AILANG sprints:
   for f in sim/*.ail; do ailang check "$f"; done
   ailang agent inbox stapledons_voyage

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

3. **Developer Experience (DX) Report** (REQUIRED)

   After every sprint, send a DX feedback message reflecting on the overall experience:

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
   - AILANG issues encountered
   - Workarounds used
   - DX rating (1-5 stars)

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
1. Check inbox: `ailang agent inbox stapledons_voyage`
2. Verify fix: `ailang check sim/*.ail`
3. Run `mark_fixed.sh "<keyword>" "<version>"`
4. Remove workarounds from code where practical
5. Acknowledge: `ailang agent ack <msg-id>`

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

### Project Commands
- `make sim` - Compile AILANG to Go
- `make game` - Build game executable
- `make run` - Run game

### AILANG Commands
- `ailang check <file>` - Type-check
- `ailang run --entry <func> <file>` - Run with entry point
- `ailang prompt` - Syntax reference

### Feedback Commands
- `ailang agent inbox stapledons_voyage` - Check messages
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
