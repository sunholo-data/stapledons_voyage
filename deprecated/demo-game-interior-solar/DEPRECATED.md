# DEPRECATED: demo-game-interior-solar

**Status:** Deprecated (2025-12-20)

## Why Deprecated

This demo combined 3D interior navigation with solar system views. Both the 3D navigation approach and the specific dual-coordinate implementation have been **rejected**.

**Reasons for rejection:**
- 3D player movement in ship-local meters is unnecessary (no spatial navigation needed)
- Dual coordinates still needed but for different purpose: deck scene rendering vs galactic positioning
- Scene-based approach achieves same goals (observation deck) without 3D complexity
- Interior-to-exterior marrying happens through 2D scene + windowed starmap compositing

## Replacement

See design decision "[2025-12-20] Interior Ship Experience: Scene-Based Navigation" in [docs/vision/design-decisions.md](../../docs/vision/design-decisions.md)

New design doc: [design_docs/planned/scene-based-interior-navigation.md](../../design_docs/planned/scene-based-interior-navigation.md)

**What to preserve:** The concept of "marrying internal player experience with accurate external physics" - now implemented through:
- Outward-facing deck scenes (Bridge, Observation Deck)
- Live starmap composited in window regions
- SR/GR effects visible through windows during cruise

## Status

This demo may still compile, but **should not be used** as a reference for future work.

Consider removing in future cleanup.

---

**Related rejected docs:**
- [design_docs/rejected/ailang-dome-demo-dual-coordinates.md](../../design_docs/rejected/ailang-dome-demo-dual-coordinates.md)
