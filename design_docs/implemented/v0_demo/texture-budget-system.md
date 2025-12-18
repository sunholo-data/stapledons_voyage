# Texture Budget System

**Status:** Implemented
**Version:** v0.demo
**Location:** [engine/textures/budget.go](../../../engine/textures/budget.go)

## Problem

Runaway texture allocation can exhaust GPU memory and crash the system. This occurred in `demo-starmap-visuals` when creating 3,802 individual billboard textures (one per star) instead of caching by spectral type.

**Impact:** Complete system freeze requiring hard restart.

## Solution

A global texture budget system that:

1. **Tracks allocations** - Monitors all texture memory usage
2. **Enforces limits** - Rejects allocations that would exceed budget
3. **Auto-detects limits** - Uses ~25% of system memory (clamped 128MB-2GB)
4. **Provides warnings** - Logs at 80% usage before hitting limit
5. **Graceful degradation** - Returns nil + error instead of crashing

## API

### Safe Allocation (Recommended)

```go
import "stapledons_voyage/engine/textures"

// Create texture with budget check
img, err := textures.NewImageSafe(width, height)
if err != nil {
    // Handle gracefully - use fallback or skip
    log.Printf("Cannot allocate texture: %v", err)
    return nil
}

// Create from Go image
img, err := textures.NewImageFromImageSafe(srcImage)

// Check before allocating
if textures.CanAllocateImage(1024, 1024) {
    // Safe to proceed
}

// Release when done (optional but helps accuracy)
textures.ReleaseImage(img)
```

### Budget Configuration

```go
// Set custom limit (bytes)
textures.SetMaxBudget(512 * 1024 * 1024) // 512MB

// Get current stats
stats := textures.GetStats()
fmt.Printf("Using %.1f%% of budget (%s / %s)\n",
    stats.UsagePercent,
    formatBytes(stats.AllocatedBytes),
    formatBytes(stats.MaxBytes))
```

### Custom Budget Pools

```go
// Create isolated budget for specific subsystem
starfieldBudget := textures.NewBudget(256 * 1024 * 1024) // 256MB

// Set callbacks
starfieldBudget.SetWarningCallback(func(allocated, max int64) {
    log.Printf("Starfield textures at %.0f%% capacity",
        float64(allocated)/float64(max)*100)
})

starfieldBudget.SetLimitCallback(func(requested, allocated, max int64) {
    log.Printf("REJECTED: Cannot allocate %d bytes", requested)
})
```

## Integration Points

The LOD system's sprite creation functions now use the budget:

| Function | File | Behavior |
|----------|------|----------|
| `CreateDefaultPlanetSprite` | `engine/lod/billboard.go` | Returns nil if budget exceeded |
| `CreateDefaultStarSprite` | `engine/lod/billboard.go` | Returns nil if budget exceeded |
| `CreateBillboardFromTexture` | `engine/lod/billboard.go` | Returns nil if budget exceeded |
| `CreateSpriteAtlas` | `engine/lod/billboard.go` | Returns nil, nil if budget exceeded |

## Default Limits

The auto-detection algorithm:

1. Reads Go runtime memory stats (`runtime.MemStats.Sys`)
2. Checks `debug.SetMemoryLimit` if configured
3. Uses 25% of detected memory for texture budget
4. Clamps to range: **128MB minimum, 2GB maximum**

Example outputs:
- 8GB system → 500MB texture budget
- 16GB system → 1GB texture budget
- 32GB+ system → 2GB texture budget (capped)

## Best Practices

### DO: Cache by Category

```go
// Good: 7 sprites for 3,800 stars
spectralSprites := make(map[string]*ebiten.Image)
for spectralType, info := range spectralTypes {
    sprite, _ := textures.NewImageSafe(64, 64)
    // ... render sprite
    spectralSprites[spectralType] = sprite
}

// Map all stars to shared sprites
for _, star := range stars {
    starSprites[star.ID] = spectralSprites[star.SpectralType]
}
```

### DON'T: Create Per-Instance

```go
// Bad: 3,800 separate textures!
for _, star := range stars {
    sprite := ebiten.NewImage(64, 64) // Unbounded allocation
    starSprites[star.ID] = sprite
}
```

### Handle Nil Returns

```go
sprite := lod.CreateDefaultPlanetSprite(64, col)
if sprite == nil {
    // Budget exceeded - use fallback or skip rendering
    return
}
```

## Monitoring

The system logs automatically:
- `[textures] Budget initialized: 512.0 MB max`
- `[textures] Detected system memory ~8.0 GB, texture limit set to 512.0 MB`
- `[textures] WARNING: Budget 80% used (409.6 MB of 512.0 MB)`
- `[lod] WARNING: Failed to create planet sprite: texture budget exceeded...`

## Testing

```go
// Simulate low memory conditions
textures.SetMaxBudget(1024 * 1024) // 1MB only

// Attempt large allocation
img, err := textures.NewImageSafe(2048, 2048) // 16MB needed
if err != nil {
    // Expected: budget exceeded
}
```

## Future Improvements

1. **LRU eviction** - Auto-dispose least-recently-used textures
2. **Texture pooling** - Reuse textures of common sizes
3. **GPU memory query** - Use actual VRAM availability (platform-specific)
4. **Compression** - Support GPU-compressed formats (DXT, ETC)
