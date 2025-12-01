# Test Manager Reference

## Creating New Test Scenarios

### Scenario JSON Format

```json
{
  "name": "scenario-name",
  "description": "What this scenario tests",
  "seed": 1234,
  "test_mode": true,
  "camera": {"x": 0, "y": 0, "zoom": 1.0},
  "events": [
    {"frame": 0, "capture": "initial.png"},
    {"frame": 1, "key": "W", "action": "down"},
    {"frame": 30, "key": "W", "action": "up"},
    {"frame": 30, "capture": "after-move.png"}
  ]
}
```

### Event Types

| Event Type | Fields | Description |
|------------|--------|-------------|
| Key down | `key`, `action: "down"` | Start holding a key |
| Key up | `key`, `action: "up"` | Release a key |
| Key press | `key`, `action: "press"` | Press and release (1 frame) |
| Click | `click: {x, y, button}` | Mouse click |
| Capture | `capture: "filename.png"` | Take screenshot |

### Key Names

Letters: `A`-`Z`
Arrows: `Up`, `Down`, `Left`, `Right` (or `ArrowUp`, etc.)

## Test Mode

The `--test-mode` flag strips all UI elements from screenshots:
- Camera info panel (top-left)
- Controls help panel (bottom)
- Any other DrawCmdUi elements

This ensures golden file comparisons only check game state, not UI changes.

## Golden File Best Practices

1. **Commit golden files to git** - They are the source of truth
2. **Use descriptive filenames** - `after-zoom-in.png` not `capture-2.png`
3. **One test per concern** - Separate scenarios for camera, NPCs, etc.
4. **Document changes** - Update scenario descriptions when updating golden files
5. **Review diffs carefully** - Only update if change is intentional

## Troubleshooting

### "No test output found"
Run tests first: `.claude/skills/test-manager/scripts/run_tests.sh`

### "No golden files for scenario"
Create golden files: `.claude/skills/test-manager/scripts/update_golden.sh <scenario>`

### Pixel differences in identical renders
- Check seed is deterministic
- Ensure test_mode is enabled
- Verify no floating point timing issues

### Tests pass locally but fail in CI
- Check platform-specific rendering differences
- Consider using perceptual diff instead of byte comparison
