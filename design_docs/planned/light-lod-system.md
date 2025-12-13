# Light LOD System

## Overview

When rendering large-scale scenes with multiple star systems, the lighting system needs to adapt based on camera distance and visibility. This design doc outlines a Level of Detail (LOD) system for lights that switches between PointLight (close-up) and DirectionalLight (distant) to optimize performance while maintaining visual quality.

## Current State

The current lighting system uses:
- **PointLight** (via `StarLight`) at the star's position with high energy (8000+) to compensate for inverse square falloff
- **AmbientLight** for minimal scene fill
- Light properties are data-driven via `LODObject.Luminosity` and `LODObject.LightColor`

This works well for single-star systems but may need optimization for:
- Multiple visible stars
- Performance-sensitive scenarios
- Very large distances where PointLight falloff is computationally expensive

## Proposed LOD Tiers

### Tier 1: PointLight (Close Range)
- **When**: Star is close enough to show visible sphere (Full3D or Billboard LOD tier)
- **Light Type**: PointLight with inverse square falloff
- **Energy**: Object's `Luminosity` value
- **Benefits**: Physically accurate lighting, proper shadow terminator
- **Drawbacks**: Expensive falloff calculation, may clip at extreme distances

### Tier 2: DirectionalLight (Medium Range)
- **When**: Star is visible but distant (Circle or small Billboard tier)
- **Light Type**: DirectionalLight pointing from star direction
- **Energy**: Scaled based on approximate distance
- **Direction**: Camera-to-star vector (inverted)
- **Benefits**: Cheaper to compute, no falloff calculation
- **Drawbacks**: Less accurate terminator line, parallel rays

### Tier 3: Ambient Only (Far/Culled Range)
- **When**: Star is Point tier or culled
- **Light Type**: No direct light, only ambient contribution
- **Implementation**: Sum contributions from visible stars into ambient
- **Benefits**: Very cheap, handles thousands of stars
- **Drawbacks**: No directional information, just general brightness

## Dynamic Ambient Based on Star Proximity

Instead of static ambient light, calculate ambient dynamically:

```go
func CalculateDynamicAmbient(camera Position, stars []StarInfo) (r, g, b, energy float64) {
    // Base deep-space darkness
    r, g, b = 0.02, 0.02, 0.03
    energy = 0.1

    // Add contribution from each star based on distance
    for _, star := range stars {
        dist := camera.Distance(star.Position)
        if dist > star.MaxAmbientRange {
            continue
        }

        // Inverse square contribution capped at reasonable level
        contribution := math.Min(star.Luminosity / (dist*dist), 0.5)

        // Blend star color into ambient
        r += float64(star.LightColor.R)/255.0 * contribution * 0.1
        g += float64(star.LightColor.G)/255.0 * contribution * 0.1
        b += float64(star.LightColor.B)/255.0 * contribution * 0.1
        energy += contribution * 0.2
    }

    return r, g, b, math.Min(energy, 1.5) // Cap total ambient
}
```

## Implementation Plan

### Phase 1: Deferred (Current)
Keep current PointLight system - it works well for single-star scenarios.

### Phase 2: If Needed
1. Add `LightTier` field to track current light LOD
2. Implement DirectionalLight fallback
3. Add tier transition logic based on star's apparent size
4. Implement dynamic ambient calculation

### Phase 3: Multi-Star (Future)
1. Support multiple light sources per scene
2. Intelligent light budget (max N active lights)
3. Light source prioritization based on visual importance

## Trigger Conditions

Switch to DirectionalLight when:
- Star's apparent screen size < 20 pixels AND
- Star is not the primary navigation target AND
- Multiple stars are visible

Switch to Ambient-only when:
- Star's apparent screen size < 2 pixels OR
- Star is culled from view OR
- Light budget exceeded

## Open Questions

1. **Hysteresis**: Should light LOD have separate hysteresis from visual LOD?
2. **Transitions**: Smooth blend or instant switch?
3. **Shadow Maps**: Do we need different shadow map resolutions per tier?
4. **Multiple Stars**: How to handle binary/multiple star systems?

## Status

**Status**: DEFERRED - Current PointLight system is sufficient for single-star scenarios. Revisit when:
- Performance issues arise with many lights
- Multi-star systems are implemented
- Visual artifacts appear at extreme distances

## References

- [engine/lod/types.go](../../engine/lod/types.go) - LODObject.Luminosity, LightColor
- [engine/tetra/lighting.go](../../engine/tetra/lighting.go) - StarLight, AmbientLight implementations
- [cmd/demo-lod/main.go](../../cmd/demo-lod/main.go) - Current lighting setup
