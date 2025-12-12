# Voyage CLI Dev Tools

**Status**: Planned
**Target**: v0.4.0
**Priority**: P1 (High DX impact)
**Estimated**: 3-5 days total (incremental)
**Dependencies**: Existing `voyage` CLI infrastructure

## Game Vision Alignment

| Pillar | Relevance | Score | Notes |
|--------|-----------|-------|-------|
| Time Dilation Consequence | N/A | 0 | Dev tooling |
| Civilization Simulation | N/A | 0 | Dev tooling |
| Philosophical Depth | N/A | 0 | Dev tooling |
| Ship & Crew Life | N/A | 0 | Dev tooling |
| Legacy Impact | N/A | 0 | Dev tooling |
| Hard Sci-Fi Authenticity | N/A | 0 | Dev tooling |
| **Net Score** | | **0** | **Decision: Move forward (infrastructure)** |

**Feature type:** Infrastructure
- Dev tools don't affect gameplay but dramatically improve development velocity

## Problem Statement

**Current State:**
- Running demos requires remembering paths: `go run ./cmd/demo-game-bridge`
- No easy way to inspect game state during development
- Testing visual changes requires full game launch
- No tooling for inspecting AILANG-generated state
- Debugging relativistic timeline requires manual calculation

**Impact:**
- Developers (Claude + human) spend time on repetitive tasks
- Iteration cycles slower than necessary
- Hard to verify game state matches expectations

## Goals

**Primary Goal:** Provide CLI tools that accelerate game development iteration

**Success Metrics:**
- Common dev tasks reduced to single commands
- State inspection possible without launching game window
- Visual regression testing automated

## Tool Ranking

### Priority Matrix

Tools ranked by **DX Impact** (how much dev pain they solve) vs **Effort** (implementation time):

| Tool | DX Impact | Effort | Priority | Est. Time |
|------|-----------|--------|----------|-----------|
| `voyage watch` | 🟢 High | 🟢 Low | **P0** | 1h |
| `voyage screenshot` | 🟢 High | 🟢 Low | **P0** | 1h |
| `voyage manifest` | 🟢 High | 🟢 Low | **P0** | 30m |
| `voyage seed` | 🟢 High | 🟡 Med | **P1** | 2h |
| `voyage trace` | 🟢 High | 🟡 Med | **P1** | 2h |
| `voyage scene` | 🟡 Med | 🟢 Low | **P1** | 1h |
| `voyage timeline` | 🟡 Med | 🟡 Med | **P2** | 2h |
| `voyage galaxy` | 🟡 Med | 🟡 Med | **P2** | 3h |
| `voyage sprite` | 🟡 Med | 🟡 Med | **P2** | 2h |
| `voyage audio` | 🟡 Med | 🟡 Med | **P2** | 2h |
| `voyage profile` | 🟡 Med | 🟡 Med | **P2** | 2h |
| `voyage replay` | 🟢 High | 🔴 High | **P3** | 4h |
| `voyage save` | 🟡 Med | 🟡 Med | **P3** | 2h |
| `voyage diff` | 🟢 Low | 🟡 Med | **P3** | 1h |

## Tool Specifications

---

### Tier 1: Quick Wins (P0)

#### `voyage watch`
**DX Impact:** 🟢 High - Eliminates manual rebuild cycle
**Effort:** 🟢 Low - fsnotify + exec

```bash
voyage watch              # Watch sim/*.ail, auto-run make sim
voyage watch --run        # Also restart current demo after rebuild
voyage watch --test       # Run ailang test after each change
```

**Implementation:**
```go
// cmd/voyage/cmd_watch.go
// Use fsnotify to watch sim/ directory
// On change: run "make sim", optionally restart demo
```

**Files:** `cmd/voyage/cmd_watch.go` (~80 LOC)

---

#### `voyage screenshot`
**DX Impact:** 🟢 High - Visual testing without window
**Effort:** 🟢 Low - Already have `engine/screenshot/`

```bash
voyage screenshot bridge              # Capture demo-game-bridge frame 0
voyage screenshot bridge --frames 60  # Capture after 60 frames
voyage screenshot bridge -o out/      # Output to specific dir
voyage screenshot --all               # Screenshot all demos
```

**Implementation:**
- Leverage existing `engine/screenshot/` headless capture
- Add CLI wrapper

**Files:** `cmd/voyage/cmd_screenshot.go` (~60 LOC)

---

#### `voyage manifest`
**DX Impact:** 🟢 High - Catches missing assets early
**Effort:** 🟢 Low - Parse JSON, check files exist

```bash
voyage manifest                    # Validate all manifests
voyage manifest --verbose          # Show all assets found
voyage manifest sprites            # Check only sprites.json
```

**Implementation:**
- Load each manifest JSON
- Verify referenced files exist
- Report missing/orphaned assets

**Files:** `cmd/voyage/cmd_manifest.go` (~50 LOC)

---

### Tier 2: High Value (P1)

#### `voyage seed`
**DX Impact:** 🟢 High - Reproducible debugging
**Effort:** 🟡 Med - Need to expose seed through AILANG

```bash
voyage seed 42                     # Show what seed 42 generates
voyage seed 42 --demo bridge       # Run bridge with seed 42
voyage seed --find "npc_count>5"   # Find seeds matching criteria
```

**Implementation:**
- Set Rand handler seed
- Call InitWorld, dump resulting state
- Optionally launch demo with that seed

**Files:** `cmd/voyage/cmd_seed.go` (~100 LOC)

---

#### `voyage trace`
**DX Impact:** 🟢 High - Frame-by-frame debugging
**Effort:** 🟡 Med - Need state serialization

```bash
voyage trace bridge                # Run 1 frame, dump state
voyage trace bridge --frames 10    # Run 10 frames, dump each
voyage trace bridge --watch pos    # Show only position changes
voyage trace bridge --json         # Output as JSON for tooling
```

**Implementation:**
- Run demo in headless mode
- After each frame, serialize World state
- Output as structured text or JSON

**Files:** `cmd/voyage/cmd_trace.go` (~120 LOC)

---

#### `voyage scene`
**DX Impact:** 🟡 Med - Debug 3D hierarchy
**Effort:** 🟢 Low - Tetra3D has tree structure

```bash
voyage scene tetra                 # Dump scene graph for demo-engine-tetra
voyage scene tetra --depth 3       # Limit depth
voyage scene tetra --find "Planet" # Search for nodes by name
```

**Implementation:**
- Load Tetra3D scene
- Walk node tree recursively
- Print hierarchy with indentation

**Files:** `cmd/voyage/cmd_scene.go` (~80 LOC)

---

### Tier 3: Nice to Have (P2)

#### `voyage timeline`
**DX Impact:** 🟡 Med - Game-specific debugging
**Effort:** 🟡 Med - Need timeline state

```bash
voyage timeline                    # Show current timeline state
voyage timeline --visits           # Show planets visited + their current year
voyage timeline --simulate 0.9c 10ly  # Calculate time for trip
```

**Implementation:**
- Read World state
- Calculate/display relativistic time differences
- Show civilization states at different times

**Files:** `cmd/voyage/cmd_timeline.go` (~100 LOC)

---

#### `voyage galaxy`
**DX Impact:** 🟡 Med - Visualize galaxy state
**Effort:** 🟡 Med - ASCII rendering

```bash
voyage galaxy                      # ASCII galaxy map
voyage galaxy --zoom 2             # Zoom level
voyage galaxy --highlight visited  # Mark visited stars
voyage galaxy --json               # Export star positions
```

**Implementation:**
- Load galaxy state
- Render as ASCII art in terminal
- Show star positions, visited status

**Files:** `cmd/voyage/cmd_galaxy.go` (~150 LOC)

---

#### `voyage sprite`
**DX Impact:** 🟡 Med - Asset preview
**Effort:** 🟡 Med - Image→ASCII or sixel

```bash
voyage sprite crew_idle            # Preview sprite
voyage sprite crew_idle --frames   # Show all animation frames
voyage sprite crew_idle --save     # Export as PNG
voyage sprite --list               # List all sprites
```

**Implementation:**
- Load sprite from manifest
- Render as ASCII art (or sixel if terminal supports)
- Show animation frames

**Files:** `cmd/voyage/cmd_sprite.go` (~120 LOC)

---

#### `voyage audio`
**DX Impact:** 🟡 Med - Audio asset validation
**Effort:** 🟡 Med - Audio playback

```bash
voyage audio thruster              # Play sound
voyage audio thruster --info       # Show duration, format, size
voyage audio --list                # List all audio assets
```

**Implementation:**
- Load audio file
- Play through system audio (or show waveform info)

**Files:** `cmd/voyage/cmd_audio.go` (~80 LOC)

---

#### `voyage profile`
**DX Impact:** 🟡 Med - Performance debugging
**Effort:** 🟡 Med - Go profiling integration

```bash
voyage profile bridge              # Run with CPU profiling
voyage profile bridge --mem        # Memory profiling
voyage profile bridge --frames 300 # Profile for N frames
voyage profile bridge -o out/      # Output pprof files
```

**Implementation:**
- Wrap demo execution with pprof
- Output to out/ directory
- Optionally open in browser

**Files:** `cmd/voyage/cmd_profile.go` (~100 LOC)

---

### Tier 4: Future (P3)

#### `voyage replay`
**DX Impact:** 🟢 High - Regression testing
**Effort:** 🔴 High - Input recording infrastructure

```bash
voyage replay record bridge        # Record input to file
voyage replay play recording.json  # Play back recording
voyage replay compare a.json b.json # Compare two runs
```

**Implementation:**
- Hook into input system
- Record FrameInput sequence to file
- Replay by feeding recorded inputs

**Files:** `cmd/voyage/cmd_replay.go` (~200 LOC), `engine/replay/` (~150 LOC)

---

#### `voyage save`
**DX Impact:** 🟡 Med - Save file debugging
**Effort:** 🟡 Med - Depends on save format

```bash
voyage save inspect                # Show save file contents
voyage save export --json          # Export as JSON
voyage save import save.json       # Import from JSON
```

**Implementation:**
- Parse save file format
- Display/convert contents

**Files:** `cmd/voyage/cmd_save.go` (~100 LOC)

---

#### `voyage diff`
**DX Impact:** 🟢 Low - Niche use case
**Effort:** 🟡 Med - Git + parsing

```bash
voyage diff                        # Show sim_gen changes since last make sim
voyage diff --types                # Show only type changes
voyage diff --funcs                # Show only function changes
```

**Implementation:**
- Run `git diff sim_gen/`
- Parse and summarize changes

**Files:** `cmd/voyage/cmd_diff.go` (~80 LOC)

---

## Implementation Plan

### Phase 1: Quick Wins (~3 hours)
- [ ] `voyage watch` - File watching + auto-rebuild
- [ ] `voyage screenshot` - Headless screenshot capture
- [ ] `voyage manifest` - Asset validation

### Phase 2: Core Tools (~6 hours)
- [ ] `voyage seed` - Reproducible world generation
- [ ] `voyage trace` - Frame-by-frame state dump
- [ ] `voyage scene` - Tetra3D scene inspection

### Phase 3: Visualization (~8 hours)
- [ ] `voyage timeline` - Relativistic timeline display
- [ ] `voyage galaxy` - ASCII galaxy map
- [ ] `voyage sprite` - Sprite preview
- [ ] `voyage audio` - Audio playback/info

### Phase 4: Advanced (~6 hours)
- [ ] `voyage profile` - Performance profiling
- [ ] `voyage replay` - Input recording/playback
- [ ] `voyage save` - Save file inspection
- [ ] `voyage diff` - Codegen diff summary

## Files to Create

**New files:**
- `cmd/voyage/cmd_watch.go` (~80 LOC)
- `cmd/voyage/cmd_screenshot.go` (~60 LOC)
- `cmd/voyage/cmd_manifest.go` (~50 LOC)
- `cmd/voyage/cmd_seed.go` (~100 LOC)
- `cmd/voyage/cmd_trace.go` (~120 LOC)
- `cmd/voyage/cmd_scene.go` (~80 LOC)
- `cmd/voyage/cmd_timeline.go` (~100 LOC)
- `cmd/voyage/cmd_galaxy.go` (~150 LOC)
- `cmd/voyage/cmd_sprite.go` (~120 LOC)
- `cmd/voyage/cmd_audio.go` (~80 LOC)
- `cmd/voyage/cmd_profile.go` (~100 LOC)
- `cmd/voyage/cmd_replay.go` (~200 LOC)
- `cmd/voyage/cmd_save.go` (~100 LOC)
- `cmd/voyage/cmd_diff.go` (~80 LOC)

**Modified files:**
- `cmd/voyage/main.go` - Add new command cases

## Success Criteria

- [ ] P0 tools implemented and working
- [ ] Each tool has `--help` documentation
- [ ] Tools work from any directory (find project root)
- [ ] All tools installed via `make install`

## Non-Goals

**Not in this feature:**
- GUI tools - CLI only for now
- IDE integration - Can be added later
- Remote debugging - Local only

## Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|-----------|
| fsnotify platform issues | Med | Test on macOS/Linux |
| Headless rendering quirks | Med | Use existing screenshot infrastructure |
| Terminal compatibility (sixel) | Low | Fall back to ASCII |

## References

- Existing CLI: `cmd/voyage/`
- Screenshot system: `engine/screenshot/`
- Scenario runner: `engine/scenario/`
- Asset manifests: `assets/sprites/manifest.json`, `assets/audio/manifest.json`

## Future Work

- IDE extension for VSCode
- Web-based state inspector
- Live reload in running game
- Automated golden file testing

---

**Document created**: 2025-12-12
**Last updated**: 2025-12-12
