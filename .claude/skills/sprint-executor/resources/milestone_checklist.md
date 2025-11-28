# Milestone Execution Checklist

For Stapledons Voyage game development with AILANG.

## Pre-Implementation
- [ ] Mark milestone as `in_progress` in TodoWrite
- [ ] Review milestone goals and acceptance criteria
- [ ] Run `ailang prompt` to refresh syntax knowledge
- [ ] Check CLAUDE.md for relevant limitations

## AILANG Implementation
- [ ] Define/update types in sim/*.ail
- [ ] Implement functions
- [ ] Run `ailang check sim/<file>.ail` after each change
- [ ] Use pattern matching for ADT handling
- [ ] Keep recursion depth in mind

## AILANG Testing
- [ ] Run `ailang check` on all modules:
  ```bash
  for f in sim/*.ail; do ailang check "$f"; done
  ```
- [ ] Test entry functions:
  ```bash
  ailang run --entry init_world sim/step.ail
  ```
- [ ] Verify no recursion overflow

## Engine Integration (if needed)
- [ ] Run `make sim` to compile AILANG → Go
- [ ] Update engine/ Go code if needed
- [ ] Run `make game` to build
- [ ] Run `make run` to test

## AILANG Feedback
- [ ] Note any issues encountered:
  - Unclear error messages
  - Missing features
  - Documentation gaps
- [ ] Report issues:
  ```bash
  ~/.claude/skills/ailang-feedback/scripts/send_feedback.sh <type> "<title>" "<desc>" "stapledons_voyage"
  ```
- [ ] Document workarounds used

## Documentation
- [ ] Update CLAUDE.md if new limitations found
- [ ] Update sprint plan (mark milestone ✅)
- [ ] Note AILANG issues in sprint summary

## Pause for Breath
- [ ] Show what was completed
- [ ] Show AILANG issues reported
- [ ] Show sprint progress
- [ ] Ask user: "Ready to continue?"

## Milestone Complete
- [ ] All sim/*.ail pass `ailang check`
- [ ] Feature works in game (if applicable)
- [ ] Feedback sent to AILANG team
- [ ] Mark milestone as `completed` in TodoWrite
