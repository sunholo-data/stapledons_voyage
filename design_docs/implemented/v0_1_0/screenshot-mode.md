# Screenshot Mode for Headless Testing

**Status**: Implemented
**Target**: v0.5.0 (Engine-side, independent of AILANG)
**Priority**: P1 - Enables AI self-testing
**Estimated**: 1 implementation session
**Dependencies**: None

## Game Vision Alignment

**Feature type:** Development Infrastructure

| Pillar | Relevance | Score | Notes |
|--------|-----------|-------|-------|
| Time Dilation Consequence | N/A | 0 | Development tooling |
| Civilization Simulation | N/A | 0 | Development tooling |
| Philosophical Depth | N/A | 0 | Development tooling |
| Ship & Crew Life | N/A | 0 | Development tooling |
| Legacy Impact | N/A | 0 | Development tooling |
| Hard Sci-Fi Authenticity | N/A | 0 | Development tooling |
| AI-Assisted Development | +++ | +3 | Core enabler for self-testing |
| **Net Score** | | **+3** | **Decision: Move forward** |

**Rationale:** This is essential infrastructure for AI-assisted development. Enables Claude Code to verify rendering changes without human intervention. No impact on gameplay pillars.

## Problem Statement

**Current State:**
- Game can only be tested by running it interactively
- AI agents cannot take screenshots or interact with GUI
- Visual bugs require human verification
- Development iteration is blocked on human availability

**Impact:**
- AI cannot self-verify rendering changes
- Longer feedback loops for visual features
- Regression testing is manual and error-prone

## Goals

**Primary Goal:** Add command-line flags to capture screenshots automatically for AI verification.

**Success Metrics:**
- `./bin/game --screenshot 60 out/test.png` produces valid PNG
- AI can read the PNG and verify rendering
- Screenshots are deterministic (same seed = same image)

## Solution Design

### Overview

Add CLI flags to the game executable:
- `--screenshot <frames> <output.png>` - Run N frames, save screenshot, exit
- `--seed <int>` - Set world seed for determinism
- `--camera <x,y,zoom>` - Set initial camera position

### Architecture

```
CLI Flags
    ↓
main.go parses flags
    ↓
If --screenshot:
    - Create headless game (no window)
    - Run N frames of simulation
    - Render final frame to off-screen image
    - Save as PNG
    - Exit
```

### Implementation Plan

**Phase 1: CLI Parsing**
- [ ] Add flag parsing in `cmd/game/main.go`
- [ ] Define `ScreenshotConfig` struct
- [ ] Parse `--screenshot`, `--seed`, `--camera` flags

**Phase 2: Headless Rendering**
- [ ] Create off-screen `ebiten.Image` for rendering
- [ ] Run game loop without window for N frames
- [ ] Render final `FrameOutput` to off-screen image

**Phase 3: PNG Export**
- [ ] Use `image/png` to encode off-screen image
- [ ] Write to specified output path
- [ ] Exit with status 0 on success

**Phase 4: Integration**
- [ ] Add `make screenshot` target
- [ ] Update documentation

### Files to Modify/Create

**Modified files:**
- `cmd/game/main.go` - Add flag parsing, headless mode (~80 LOC)

**New files:**
- `engine/screenshot/screenshot.go` - Headless rendering utilities (~60 LOC)

## Examples

### Example 1: Basic Screenshot

```bash
# Take screenshot after 60 frames
./bin/game --screenshot 60 out/test.png

# AI can then read it
# Read out/test.png -> [displays image]
```

### Example 2: Deterministic Screenshot

```bash
# Same seed = same output
./bin/game --screenshot 60 --seed 1234 out/a.png
./bin/game --screenshot 60 --seed 1234 out/b.png
# a.png and b.png should be identical
```

### Example 3: Camera Position

```bash
# Capture specific view
./bin/game --screenshot 60 --camera 100,100,2.0 out/zoomed.png
```

## Success Criteria

- [ ] `--screenshot 1 out/test.png` produces valid 640x480 PNG
- [ ] `--screenshot 60` shows NPCs in different positions than `--screenshot 1`
- [ ] `--seed 1234` produces deterministic output
- [ ] Exit code is 0 on success, non-zero on error
- [ ] No window appears in screenshot mode

## Testing Strategy

**Unit tests:**
- Flag parsing returns correct config
- Headless image creation works

**Integration tests:**
- Full screenshot capture produces valid PNG
- Determinism: same seed = same image (binary compare)

**AI verification:**
- Claude reads screenshot and identifies:
  - Isometric tile grid visible
  - UI panels present
  - Correct camera position

## Non-Goals

- **Interactive testing** - This is for automated capture only
- **Video recording** - Single frame capture, not video
- **GUI screenshot tool** - CLI only

## Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|-----------|
| Ebiten headless issues | High | Test on CI, fallback to virtual display |
| Non-deterministic rendering | Med | Control all random seeds |
| Large file sizes | Low | Use PNG compression |

## References

- [game-vision.md](../../../docs/game-vision.md) - AI-Assisted Development section
- Ebiten headless: https://ebitengine.org/en/documents/headless.html

---

**Document created**: 2025-12-01
**Last updated**: 2025-12-01
