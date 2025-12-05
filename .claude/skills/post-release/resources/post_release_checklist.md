# Post-Release Checklist

For Stapledons Voyage game milestones/releases.

## Prerequisites

- [ ] All sim/*.ail files pass `ailang check`
- [ ] Game builds with `make game`
- [ ] Game runs with `make run`

## 1. AILANG Module Verification

```bash
# Verify all modules compile
for f in sim/*.ail; do
  echo "=== $f ===" && ailang check "$f"
done
```

- [ ] All modules compile
- [ ] No recursion overflow issues
- [ ] Entry functions work (`ailang run --entry init_world sim/step.ail`)

## 2. AILANG Feedback Summary

Check messages:
```bash
ailang messages list --unread
ailang messages read <msg-id>
```

Summarize issues reported during development:

- [ ] List bugs reported
- [ ] List features requested
- [ ] List documentation gaps found
- [ ] Note any responses received
- [ ] Identify issues still blocking features

## 3. Update CLAUDE.md

- [ ] Remove fixed limitations
- [ ] Add newly discovered limitations
- [ ] Update workarounds
- [ ] Note AILANG version tested against

## 4. Update Design Docs

```bash
# Move completed docs
mv design_docs/planned/<feature>.md design_docs/implemented/v0_X_X/
```

- [ ] Move completed docs to implemented/
- [ ] Update status: Planned → Implemented
- [ ] Note workarounds used
- [ ] Add implementation report section

## 5. Create Release Notes

Include:
- [ ] Features implemented
- [ ] AILANG modules used
- [ ] Known issues/limitations
- [ ] AILANG feedback summary
- [ ] Next steps

## 6. Git Operations

```bash
# Tag release
git tag -a v0.X.X -m "Description"
git push origin v0.X.X
```

- [ ] All changes committed
- [ ] Version tagged
- [ ] Tag pushed

## 7. Plan Next Sprint

- [ ] Check inbox for AILANG responses
- [ ] Identify blocking AILANG issues
- [ ] Plan workarounds for next features
- [ ] Create next sprint plan

## Final Verification

- [ ] Game runs without errors
- [ ] All AILANG issues documented
- [ ] Feedback sent to AILANG team
- [ ] CLAUDE.md up to date
- [ ] Design docs organized
- [ ] Release notes complete

## Notes

- This project is an AILANG integration test
- Release notes should emphasize AILANG usage
- Track AILANG issues - they're the project's contribution
- Check for AILANG updates that might fix issues
