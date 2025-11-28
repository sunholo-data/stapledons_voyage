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

1. **Test-Driven**: Run `ailang check` after every change
2. **Feedback-First**: Report AILANG issues immediately
3. **Pause Points**: Stop after each milestone for review
4. **Track Progress**: Use TodoWrite for visibility
5. **Document Workarounds**: Record how you navigated limitations

## Execution Flow

### Phase 1: Initialize Sprint

1. **Check AILANG Status**
   ```bash
   # Verify all modules compile
   for f in sim/*.ail; do ailang check "$f"; done

   # Check for AILANG team messages
   ailang agent inbox stapledons_voyage
   ```

2. **Review Limitations**
   - Read CLAUDE.md "Known Limitations" section
   - Note which limitations affect this sprint

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

5. **Update Documentation**
   - Note workarounds used
   - Update CLAUDE.md if new limitations found
   - Mark milestone complete in sprint plan

6. **Pause for Review**
   - Show progress to user
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

3. **Sprint Report**
   - Milestones completed
   - AILANG issues encountered
   - Workarounds used
   - Time spent vs estimated

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
