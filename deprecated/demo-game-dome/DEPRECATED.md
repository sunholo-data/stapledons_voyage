# DEPRECATED: demo-game-dome

**Status:** Deprecated (2025-12-20)

## Why Deprecated

This demo implemented AILANG-driven dome rendering with dual-coordinate systems. The underlying `sim/dome_demo.ail` file has been **removed** and this approach has been **rejected**.

**Reasons for rejection:**
- 3D player movement approach abandoned for scene-based navigation
- Interior experience doesn't need spatial exploration
- Scene-based selection (deck UI) better serves conversation-focused gameplay
- AILANG will generate DrawCmds for 2D deck scenes, not 3D navigation

## Replacement

See design decision "[2025-12-20] Interior Ship Experience: Scene-Based Navigation" in [docs/vision/design-decisions.md](../../docs/vision/design-decisions.md)

New design doc: [design_docs/planned/scene-based-interior-navigation.md](../../design_docs/planned/scene-based-interior-navigation.md)

## Status

**DO NOT USE.** This demo no longer compiles because `sim/dome_demo.ail` has been removed.

This directory is kept for historical reference only. It will be removed in a future cleanup.

---

**Related rejected docs:**
- [design_docs/rejected/ailang-dome-demo.md](../../design_docs/rejected/ailang-dome-demo.md)
- [design_docs/rejected/ailang-dome-demo-dual-coordinates.md](../../design_docs/rejected/ailang-dome-demo-dual-coordinates.md)
