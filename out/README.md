# out/ - Generated Output Directory

This directory contains all generated output from builds, tests, and evaluations.

**This entire directory is gitignored** - only the structure is tracked via `.gitkeep`.

## Directory Structure

```
out/
├── eval/           # Evaluation reports and benchmarks
├── generated/      # Final generated assets (GIFs, videos)
├── scenarios/      # Scenario runner output (temporary)
├── screenshots/    # Demo screenshots from sprint execution
└── test/           # Visual test output for golden file comparison
```

## Subdirectory Usage

### `eval/` - Evaluation Output
Used by: `make eval`, `dev-tools` skill

Contains benchmark reports and evaluation output:
- `report.json` - AI evaluation results from `make eval`
- `bench.txt` - Performance benchmark output
- `perf.json` - Detailed performance reports from `./bin/voyage perf`

Example:
```bash
make eval                           # Generates out/eval/report.json
./bin/voyage perf -o out/eval/perf.json
```

### `generated/` - Generated Assets
Used by: Video/GIF generation during development

Final rendered assets like:
- `*.gif` - Animated GIFs for documentation or demos
- `*.mp4` - Video captures

**Note:** Intermediate frame files should be cleaned up after video generation. Do not commit large frame directories.

### `scenarios/` - Scenario Runner
Used by: `engine/scenario/` runner, `test-manager` skill

Temporary output from scenario execution. Files here get moved to `test/` by the test-manager skill for golden file comparison.

### `screenshots/` - Demo Screenshots
Used by: `sprint-executor` skill

Screenshots captured during sprint execution for verification:
```bash
./bin/demo --screenshot 30 --output out/screenshots/initial.png
./bin/demo --screenshot 90 --output out/screenshots/final.png
```

### `test/` - Visual Test Output
Used by: `test-manager` skill

Contains test screenshots for golden file comparison:
```
test/
├── camera-pan/     # Camera panning test captures
├── camera-zoom/    # Camera zoom test captures
├── npc-movement/   # NPC movement test captures
└── <scenario>/     # Other scenario captures
    └── diff/       # Difference images when tests fail
```

Commands:
```bash
# Run visual tests
.claude/skills/test-manager/scripts/run_tests.sh

# Compare against golden files
.claude/skills/test-manager/scripts/compare_golden.sh

# Update golden files after verified changes
.claude/skills/test-manager/scripts/update_golden.sh
```

## Cleanup

This directory can grow large during development. Clean up with:

```bash
make clean      # Removes bin/, out/* (preserves out/.gitkeep)
make clean-all  # Full clean including sim_gen/
```

To manually clean specific directories:
```bash
rm -rf out/frames/       # Remove intermediate frame files
rm -rf out/screenshots/  # Clear old demo screenshots
```

## Adding New Output Categories

When adding a new type of output:

1. Create a subdirectory: `mkdir -p out/myoutput/`
2. Document it in this README
3. Update relevant skill documentation
4. Consider adding cleanup instructions
