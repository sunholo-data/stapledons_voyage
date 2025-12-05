---
name: Game Post-Release Tasks
description: Run post-release tasks for Stapledons Voyage game versions. Sends AILANG feedback summary, updates documentation, and archives completed design docs. Use when user says "post-release tasks" or after completing a game milestone.
---

# Game Post-Release Tasks

Run post-release tasks after completing a game milestone or version.

## Quick Start

**Most common usage:**
```bash
# User says: "Run post-release tasks for v0.1.0"
# This skill will:
# 1. Verify all AILANG modules compile
# 2. Summarize AILANG feedback sent during development
# 3. Move design docs from planned/ to implemented/
# 4. Update project documentation
# 5. Create release summary
```

## When to Use This Skill

Invoke this skill when:
- User says "post-release tasks", "finalize release"
- After completing a game milestone
- Before tagging a version
- When wrapping up a sprint with a deliverable

## Post-Release Workflow

### 1. Verify Game State

```bash
# All AILANG modules compile
for f in sim/*.ail; do
  echo "=== $f ===" && ailang check "$f"
done

# Game builds
make game

# Quick test
make run
```

### 2. AILANG Feedback Summary

**Review all feedback sent during development:**

```bash
# Check inbox for any responses
ailang messages list --unread

# List all messages
ailang messages list

# Read a specific message
ailang messages read <msg-id>
```

**Create feedback summary:**
```markdown
## AILANG Feedback Summary (v0.1.0)

### Bugs Reported
- [x] Module imports not working (msg_xxx)
- [x] Recursion depth overflow (msg_xxx)
- [x] Record update syntax failure (msg_xxx)

### Features Requested
- [ ] RNG effect for random numbers
- [ ] Array type for O(1) access
- [ ] Numeric type conversion functions

### Documentation Gaps
- [ ] How to do random number generation
- [ ] Numeric conversion (int <-> float)

### Responses Received
- (List any responses from AILANG team)

### Still Blocking
- (List issues still preventing game features)
```

### 3. Update Design Docs

**Move completed docs:**
```bash
# Move from planned/ to implemented/
mv design_docs/planned/feature.md design_docs/implemented/v0_1_0/

# Update status in doc header
# Status: Planned → Implemented
```

**Update README.md:**
```markdown
## Implemented Features

### v0.1.0 (Date)
- Basic world rendering
- 2x2 tile grid
- Frame update loop
```

### 4. Update CLAUDE.md

**Update Known Limitations:**
- Remove fixed limitations
- Add newly discovered limitations
- Update workarounds

**Example update:**
```markdown
### Known Limitations (as of v0.1.0)

- **Module imports not working** - Use local type definitions
- **NEW: Large grids cause stack overflow** - Keep to 8x8 max for now
```

### 5. Create Release Notes

```markdown
# Stapledons Voyage v0.1.0

## Features
- Basic world rendering with tile grid
- Initial AILANG simulation structure

## AILANG Usage
- sim/protocol.ail - Core types (FrameInput, DrawCmd, etc.)
- sim/world.ail - World state types
- sim/step.ail - Game loop logic
- sim/npc_ai.ail - Placeholder for NPC AI

## Known Issues
- Large grids (>8x8) cause recursion overflow
- Module imports not working (types duplicated)

## AILANG Feedback
Reported 6 issues to AILANG core:
- 3 bugs
- 2 feature requests
- 1 documentation gap

## Next Steps
- Wait for AILANG responses
- Plan v0.2.0 features (NPC movement)
```

### 6. Git Operations

```bash
# Tag the release
git tag -a v0.1.0 -m "Initial game structure with AILANG"

# Push tag
git push origin v0.1.0
```

## Post-Release Checklist

```markdown
## v0.1.0 Post-Release Checklist

### Verification
- [ ] All sim/*.ail files compile with `ailang check`
- [ ] Game builds with `make game`
- [ ] Game runs with `make run`

### AILANG Feedback
- [ ] Reviewed all feedback sent during development
- [ ] Checked inbox for AILANG team responses
- [ ] Created feedback summary

### Documentation
- [ ] Design docs moved to implemented/
- [ ] CLAUDE.md updated with new limitations
- [ ] Release notes created

### Git
- [ ] All changes committed
- [ ] Version tagged
- [ ] Tag pushed to origin

### Next Sprint
- [ ] Identified blocking AILANG issues
- [ ] Planned workarounds for v0.2.0
```

## Available Scripts

### `scripts/verify_release.sh [version]`
Verify game is ready for release.

**Usage:**
```bash
.claude/skills/post-release/scripts/verify_release.sh v0.1.0
```

**What it checks:**
- All AILANG modules compile
- Entry functions run
- Game builds
- AILANG inbox status

## Best Practices

### 1. Always Check Inbox
Before release, check for AILANG team responses that might affect the release.

### 2. Document Workarounds
Future developers (or future Claude sessions) need to know what workarounds are in place.

### 3. Prioritize Feedback
Mark which AILANG issues are blocking next features.

### 4. Keep Feedback Summary
Track all issues reported - it's the project's contribution to AILANG.

## Notes

- This project is primarily an AILANG integration test
- Release notes should emphasize AILANG usage and issues
- Post-release is a good time to review feedback loop effectiveness
- Check if any reported issues have been fixed in newer AILANG versions
