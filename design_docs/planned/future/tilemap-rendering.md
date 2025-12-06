# Tilemap Rendering Optimization

**Version:** 0.3.0
**Status:** Planned
**Priority:** P1 (Medium)
**Complexity:** Complex
**Package:** `engine/render`

## Related Documents

- [Camera and Viewport](camera-viewport.md) - Viewport culling
- [Asset Management](../v0_2_0/asset-management.md) - Tile sprite loading
- [Engine Layer Design](../../implemented/v0_1_0/engine-layer.md) - Current rendering

## Problem Statement

Rendering each tile as a separate DrawCmd is inefficient for large tilemaps. A 64x64 world = 4096 draw calls per frame.

**Current State:**
- Each tile generates a DrawCmd
- 4096+ draw calls for 64x64 world
- No batching or atlasing
- Performance degrades with world size

**What's Needed:**
- Tilemap batching (single draw call for visible tiles)
- Tile atlas support
- Only render visible tiles (with camera system)
- Maintain AILANG as source of truth

## Design Options

### Option A: AILANG Sends Tile Grid

AILANG sends compact tile data, Go engine renders efficiently.

```ailang
type TileLayer = {
    tiles: [int],      -- Flat array of tile IDs (row-major)
    width: int,
    height: int
}

type FrameOutput = {
    tile_layer: TileLayer,   -- Background tiles
    draw_cmds: [DrawCmd],    -- Entities, UI (on top of tiles)
    camera: Camera
}
```

**Pros:** Minimal data transfer, engine optimizes rendering
**Cons:** Requires array type (AILANG v0.5.0+)

### Option B: Tile Atlas in Go

AILANG sends DrawCmds as before, Go engine batches by sprite ID.

```go
// Group DrawCmds by sprite ID, render as batched quads
type SpriteBatch struct {
    spriteID int
    quads    []Quad  // Position, UV for each instance
}
```

**Pros:** No AILANG changes, works with current system
**Cons:** Less efficient, still processes many DrawCmds

### Option C: Pre-rendered Tile Chunks

Divide world into chunks (16x16), pre-render to textures.

```go
type Chunk struct {
    x, y    int
    texture *ebiten.Image
    dirty   bool  // Re-render if tiles changed
}
```

**Pros:** Very fast rendering (1 draw per chunk)
**Cons:** Memory usage, complexity with tile changes

**Decision:** Start with Option B (batching), upgrade to Option A when AILANG arrays available.

## Go Implementation (Option B)

### Sprite Batching

```go
package render

type SpriteBatch struct {
    image  *ebiten.Image
    quads  []BatchQuad
}

type BatchQuad struct {
    DstX, DstY, DstW, DstH float64
    SrcX, SrcY, SrcW, SrcH int  // In atlas
}

func BatchDrawCmds(cmds []sim_gen.DrawCmd, assets *assets.Manager) []SpriteBatch {
    batches := make(map[int]*SpriteBatch)

    for _, cmd := range cmds {
        if cmd.Kind != sim_gen.DrawCmdKindSprite {
            continue
        }
        id := cmd.Sprite.ID
        if batches[id] == nil {
            batches[id] = &SpriteBatch{image: assets.GetSprite(id)}
        }
        batches[id].quads = append(batches[id].quads, BatchQuad{
            DstX: cmd.Sprite.X, DstY: cmd.Sprite.Y,
            DstW: 16, DstH: 16,  // Tile size
        })
    }

    result := make([]SpriteBatch, 0, len(batches))
    for _, b := range batches {
        result = append(result, *b)
    }
    return result
}
```

### Batched Rendering

```go
func RenderBatch(screen *ebiten.Image, batch SpriteBatch, cam camera.Transform) {
    for _, q := range batch.quads {
        sx, sy := cam.WorldToScreen(q.DstX, q.DstY)

        opts := &ebiten.DrawImageOptions{}
        opts.GeoM.Scale(cam.Scale, cam.Scale)
        opts.GeoM.Translate(sx, sy)
        screen.DrawImage(batch.image, opts)
    }
}
```

### Tile Atlas

Combine multiple tile sprites into one texture:

```
┌────┬────┬────┬────┐
│ 0  │ 1  │ 2  │ 3  │  Each cell: 16x16
├────┼────┼────┼────┤
│ 4  │ 5  │ 6  │ 7  │  Atlas: 64x64 (4x4 tiles)
├────┼────┼────┼────┤
│ 8  │ 9  │ 10 │ 11 │
├────┼────┼────┼────┤
│ 12 │ 13 │ 14 │ 15 │
└────┴────┴────┴────┘
```

```go
type TileAtlas struct {
    image    *ebiten.Image
    tileSize int
    columns  int
}

func (a *TileAtlas) GetTileRect(id int) image.Rectangle {
    x := (id % a.columns) * a.tileSize
    y := (id / a.columns) * a.tileSize
    return image.Rect(x, y, x+a.tileSize, y+a.tileSize)
}
```

## Performance Targets

| World Size | Current (est.) | Target | Method |
|------------|----------------|--------|--------|
| 16x16 | 256 draws | 256 draws | No change needed |
| 32x32 | 1024 draws | ~100 draws | Batching |
| 64x64 | 4096 draws | ~200 draws | Batching + culling |
| 128x128 | 16384 draws | ~200 draws | Batching + culling |

**Key insight:** With camera culling, visible tiles ≈ (screen/tile)² regardless of world size.

## Implementation Plan

### Phase 1: Batching (v0.3.0)

| File | Change |
|------|--------|
| `engine/render/batch.go` | SpriteBatch implementation (new) |
| `engine/render/draw.go` | Use batching for tile sprites |

### Phase 2: Tile Atlas (v0.3.x)

| File | Change |
|------|--------|
| `engine/render/atlas.go` | TileAtlas implementation (new) |
| `assets/sprites/tiles.png` | Combined tile atlas |
| `assets/sprites/manifest.json` | Atlas metadata |

### Phase 3: TileLayer (v0.5.0+, requires AILANG arrays)

| File | Change |
|------|--------|
| `sim/protocol.ail` | Add TileLayer type |
| `sim/step.ail` | Emit TileLayer instead of tile DrawCmds |
| `engine/render/tilemap.go` | Render TileLayer efficiently |

## Testing Strategy

### Benchmarking

```go
func BenchmarkRenderTiles64x64(b *testing.B)
func BenchmarkRenderTilesBatched(b *testing.B)
```

### Visual Testing

```bash
make run
# Scroll around large world → should be smooth
```

### Performance Metrics

```bash
make eval
# Check draw call count in report
# Target: < 500 draws for 64x64 world with camera
```

## Success Criteria

### Phase 1
- [ ] Tile sprites batched by ID
- [ ] Draw calls reduced by 5-10x
- [ ] No visual difference from unbatched

### Phase 2
- [ ] Tile atlas loads from single image
- [ ] UV coordinates calculated correctly
- [ ] Memory usage reduced

### Phase 3 (Future)
- [ ] TileLayer type in AILANG
- [ ] Efficient tile grid rendering
- [ ] Draw calls < 10 for tile layer
