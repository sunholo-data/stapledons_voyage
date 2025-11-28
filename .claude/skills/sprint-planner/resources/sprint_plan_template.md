# Sprint Plan: [Feature Name]

## Summary
[1-2 sentence goal describing what this sprint accomplishes]

**Duration:** X days
**Dependencies:** [List any blocking items or AILANG features needed]
**Risk Level:** Low/Medium/High
**AILANG Constraints:** [Known limitations to work around]

## Current Status

### AILANG Modules
```bash
# Run before planning
for f in sim/*.ail; do ailang check "$f"; done
```
- [ ] All modules compile
- [ ] Known limitations reviewed in CLAUDE.md

### Inbox Check
```bash
ailang agent inbox stapledons_voyage
```
- [ ] Messages checked
- [ ] Relevant responses incorporated

## Proposed Milestones

### Milestone 1: [Name]
**Goal:** [What this milestone achieves]
**Estimated:** X day(s)

**AILANG Tasks:**
- [ ] Define types in sim/*.ail
- [ ] Implement functions
- [ ] Run `ailang check`
- [ ] Test with `ailang run`

**Engine Tasks (if needed):**
- [ ] Update Go rendering
- [ ] Test with `make run`

**Acceptance Criteria:**
- [ ] `ailang check` passes for all modules
- [ ] Feature works in game
- [ ] AILANG issues documented

**AILANG Workarounds:**
- [Limitation]: [How we'll work around it]

### Milestone 2: [Name]
[Repeat structure]

## AILANG Feedback Plan

**During sprint, report:**
- Bugs encountered
- Features that would help
- Documentation gaps

**Command:**
```bash
~/.claude/skills/ailang-feedback/scripts/send_feedback.sh <type> "<title>" "<description>" "stapledons_voyage"
```

## Success Metrics

### AILANG
- [ ] All sim/*.ail pass `ailang check`
- [ ] No recursion overflow
- [ ] Workarounds documented

### Gameplay
- [ ] Feature visible/functional
- [ ] Performance acceptable

### Feedback Loop
- [ ] Issues reported to AILANG team
- [ ] CLAUDE.md updated with new limitations
- [ ] Inbox checked for responses

## Dependencies & Blockers

**AILANG blockers:**
- [Feature/fix needed from AILANG team]

**Game blockers:**
- [Engine work needed]

## Open Questions

- [Question requiring clarification]

## Notes

- [Assumptions or caveats]
- [Reference to `ailang prompt` for syntax]
