# DEPRECATED: demo-engine-interior

**Status:** Deprecated (2025-12-20)

## Why Deprecated

This demo implemented 3D first-person interior navigation with free movement (WASD). This approach has been **rejected** in favor of scene-based navigation.

**Reasons for rejection:**
- Free spatial movement not needed for interior gameplay
- Game focus is conversations and visual contrast, not spatial puzzles
- 3D navigation complexity doesn't serve core pillars
- Scene-based deck selection better serves conversation-focused gameplay

## Replacement

See design decision "[2025-12-20] Interior Ship Experience: Scene-Based Navigation" in [docs/vision/design-decisions.md](../../docs/vision/design-decisions.md)

New design doc: [design_docs/planned/scene-based-interior-navigation.md](../../design_docs/planned/scene-based-interior-navigation.md)

## Should This Demo Be Removed?

**Not yet.** Engine techniques (3D rendering, camera movement) may still be useful for:
- Future 3D exterior scenes
- Reference implementation for Tetra3D integration
- Camera system examples

**Keep for now** as a reference, but **do not extend** with new interior features.

---

**Related rejected docs:**
- [design_docs/rejected/bubble-ship-dome-system.md](../../design_docs/rejected/bubble-ship-dome-system.md)
