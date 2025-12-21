# DEPRECATED: demo-engine-dome

**Status:** Deprecated (2025-12-20)

## Why Deprecated

This demo implemented 3D dome rendering with dual-coordinate systems (player movement in ship-local meters vs galactic coordinates). This approach has been **rejected** in favor of scene-based 2D/2.5D navigation.

**Reasons for rejection:**
- 3D asset creation too complex for interior spaces
- Misaligned with game focus: space visualization and conversations, not spatial navigation
- AI-generated 2D images are faster to produce and more flexible
- Over-engineering for what the interior experience actually needs

## Replacement

See design decision "[2025-12-20] Interior Ship Experience: Scene-Based Navigation" in [docs/vision/design-decisions.md](../../docs/vision/design-decisions.md)

New design doc: [design_docs/planned/scene-based-interior-navigation.md](../../design_docs/planned/scene-based-interior-navigation.md)

## Should This Demo Be Removed?

**Not yet.** The engine-level techniques (dual coordinates, sky sphere rendering, window compositing) may be useful for:
- Outward-facing deck scenes (compositing live starmap in windows)
- Future features requiring coordinate system separation
- Reference implementation for rendering techniques

**Keep for now** as a reference, but **do not extend** with new features.

## Running This Demo

If you still need to run this for reference:

```bash
go run ./cmd/demo-engine-dome
```

Controls:
- WASD: Player movement (ship-local)
- V: Cycle ship velocity
- H: Cycle ship heading
- Mouse: Look around
- R: Reset

---

**Related rejected docs:**
- [design_docs/rejected/bubble-ship-dome-system.md](../../design_docs/rejected/bubble-ship-dome-system.md)
- [design_docs/rejected/bubble-ship-dome.md](../../design_docs/rejected/bubble-ship-dome.md)
