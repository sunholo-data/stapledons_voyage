# Interior Demo Iteration: AI Textures, Windows & LOD

**Status**: Planned
**Target**: v0.2.0+
**Priority**: P1 (Core Visual System)
**Estimated**: 8-12 days across phases
**Dependencies**: demo-game-interior (done), demo-lod (done), shader system (done)

## Game Vision Alignment

**Score this feature against Stapledon's Voyage core pillars:**

| Pillar | Relevance | Score | Notes |
|--------|-----------|-------|-------|
| Time Dilation Consequence | N/A | 0 | Infrastructure feature |
| Civilization Simulation | N/A | 0 | Infrastructure feature |
| Philosophical Depth | N/A | 0 | Infrastructure feature |
| Ship & Crew Life | + | +1 | Creates immersive ship interiors where crew live |
| Legacy Impact | N/A | 0 | Infrastructure feature |
| Hard Sci-Fi Authenticity | ++ | +2 | Real SR/GR effects visible through windows |
| **Net Score** | | **+3** | **Decision: Move forward** |

**Feature type:** Engine/Infrastructure
- Enables rich ship interiors with procedural AI-generated textures
- Window system with real relativistic physics visible from inside the ship

**Reference:** See [game-vision.md](../../../docs/game-vision.md)

## Problem Statement

The demo-game-interior works with basic textures, but several limitations prevent creating immersive, varied ship interiors:

**Current State:**
- Textures are 512x512 and simply tiled across surfaces
- All rooms use the same textures regardless of purpose (bridge vs engineering vs habitat)
- No windows - interiors feel isolated from space
- No LOD system for interiors - limits room complexity
- AI-generated textures require manual prompting and organization
- No caching system for lazy texture generation

**Impact:**
- Rooms feel generic and repetitive
- Missed opportunity: AI generation could create unique per-room textures
- Large complex spaces (corridors, multi-room decks) not possible
- Player can't see space from inside the ship (core to the game experience)

## Goals

**Primary Goal:** Create a system for AI-generated, dimension-aware textures with windows into space.

**Success Metrics:**
- Each room type has distinct, themed textures (bridge=blue/cyan, engineering=orange/industrial)
- Textures generated at exact surface dimensions (not just tiled 512x512)
- Windows render space view with SR/GR effects applied correctly
- Room LOD system handles 50+ rooms without performance issues
- Texture cache prevents redundant AI generation

## Solution Design

### Overview

Five interconnected subsystems working together:

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           INTERIOR SYSTEM OVERVIEW                          │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌──────────────────┐    ┌──────────────────┐    ┌──────────────────┐     │
│  │ Theme System     │    │ Texture Generator│    │ Texture Cache    │     │
│  │                  │───▶│                  │───▶│                  │     │
│  │ Room type →      │    │ Theme + Dims →   │    │ Hash → Texture   │     │
│  │   color palette  │    │   AI prompt      │    │ Disk + Memory    │     │
│  │   style keywords │    │   generation     │    │                  │     │
│  └──────────────────┘    └──────────────────┘    └──────────────────┘     │
│                                                                             │
│  ┌──────────────────┐    ┌──────────────────┐                              │
│  │ Window System    │    │ Room LOD         │                              │
│  │                  │    │                  │                              │
│  │ Space viewport   │    │ Near: Full 3D    │                              │
│  │ SR/GR shaders    │    │ Mid: Simplified  │                              │
│  │ Parallax layers  │    │ Far: Not loaded  │                              │
│  └──────────────────┘    └──────────────────┘                              │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Architecture

#### 1. Room Theme System

Define consistent visual themes per room type:

```ailang
-- sim/interior_themes.ail

-- Room themes define visual identity per room type
export type RoomTheme = {
    name: string,            -- "bridge", "engineering", "habitat"
    primaryColor: int,       -- Dominant color (RGBA packed)
    accentColor: int,        -- Accent lighting color
    styleKeywords: string,   -- "command center, clean, blue displays"
    floorStyle: string,      -- "metallic grating, LED strips"
    wallStyle: string,       -- "control panels, screens"
    ceilingStyle: string,    -- "recessed lighting, conduits"
    windowStyle: string      -- "viewport", "porthole", "dome", "none"
}

-- Predefined themes
export pure func themeBridge() -> RoomTheme {
    {
        name: "bridge",
        primaryColor: 0x1a3050FF,    -- Deep space blue
        accentColor: 0x00ccffFF,     -- Cyan accent
        styleKeywords: "command center, sleek, holographic displays, officer quarters",
        floorStyle: "brushed metal deck plating, subtle blue LED strip lighting in seams",
        wallStyle: "curved control panels, holographic tactical displays, status screens",
        ceilingStyle: "recessed white lighting panels, ventilation grilles, clean lines",
        windowStyle: "large viewport"
    }
}

export pure func themeEngineering() -> RoomTheme {
    {
        name: "engineering",
        primaryColor: 0x3d2817FF,    -- Industrial brown
        accentColor: 0xff6600FF,     -- Warning orange
        styleKeywords: "industrial, pipes, machinery, maintenance access",
        floorStyle: "heavy duty metal grating, yellow caution stripes, oil stains",
        wallStyle: "exposed pipes, conduits, access panels, warning labels",
        ceilingStyle: "cable trays, fluorescent lights, industrial ventilation",
        windowStyle: "small porthole"
    }
}

export pure func themeHabitat() -> RoomTheme {
    {
        name: "habitat",
        primaryColor: 0x4a4a4aFF,    -- Neutral gray
        accentColor: 0xfff0d0FF,     -- Warm white
        styleKeywords: "comfortable, residential, soft lighting, living quarters",
        floorStyle: "carpeted panels over deck plating, warm colors",
        wallStyle: "padded panels, personal storage, soft ambient lighting",
        ceilingStyle: "indirect warm lighting, acoustic panels, air circulation",
        windowStyle: "observation window"
    }
}

export pure func themeCorridor() -> RoomTheme {
    {
        name: "corridor",
        primaryColor: 0x2a2a30FF,    -- Dark gray
        accentColor: 0x80ff80FF,     -- Green directional
        styleKeywords: "transitional, functional, emergency lighting, deck markers",
        floorStyle: "deck plating with directional arrows, emergency lighting strips",
        wallStyle: "smooth panels, deck number markers, emergency equipment",
        ceilingStyle: "continuous lighting strip, fire suppression, emergency signs",
        windowStyle: "none"
    }
}
```

#### 2. Dimension-Aware Texture Generation

Generate textures at exact surface dimensions instead of tiling small textures:

```ailang
-- sim/texture_spec.ail

-- Texture specification for AI generation
export type TextureSpec = {
    surfaceType: string,     -- "floor", "wall", "ceiling"
    widthMeters: float,      -- Physical width in meters
    heightMeters: float,     -- Physical height in meters
    pixelsPerMeter: int,     -- Resolution (128, 256, 512 ppm)
    theme: RoomTheme,        -- Theme to apply
    roomId: string,          -- Unique room identifier for caching
    seed: int                -- Deterministic generation seed
}

-- Calculate pixel dimensions
export pure func specPixelWidth(spec: TextureSpec) -> int {
    floatToInt(spec.widthMeters * intToFloat(spec.pixelsPerMeter))
}

export pure func specPixelHeight(spec: TextureSpec) -> int {
    floatToInt(spec.heightMeters * intToFloat(spec.pixelsPerMeter))
}

-- Generate cache key for texture
export pure func specCacheKey(spec: TextureSpec) -> string {
    -- Combine room, surface, dimensions, theme for unique key
    spec.roomId ++ "_" ++ spec.surfaceType ++ "_" ++
    intToString(specPixelWidth(spec)) ++ "x" ++ intToString(specPixelHeight(spec)) ++
    "_" ++ spec.theme.name ++ "_seed" ++ intToString(spec.seed)
}
```

**Engine-side texture generation workflow:**

```go
// engine/texgen/generator.go

type TextureGenerator struct {
    cache      *TextureCache
    aiHandler  AIHandler
    tempDir    string
}

// GenerateTexture creates or retrieves a texture for the given spec
func (g *TextureGenerator) GenerateTexture(spec TextureSpec) (*ebiten.Image, error) {
    cacheKey := spec.CacheKey()

    // 1. Check memory cache
    if tex := g.cache.GetFromMemory(cacheKey); tex != nil {
        return tex, nil
    }

    // 2. Check disk cache
    if tex := g.cache.GetFromDisk(cacheKey); tex != nil {
        g.cache.PutMemory(cacheKey, tex)
        return tex, nil
    }

    // 3. Generate with AI
    prompt := g.buildPrompt(spec)
    imgPath, err := g.aiHandler.GenerateImage(prompt, spec.PixelWidth(), spec.PixelHeight())
    if err != nil {
        return nil, err
    }

    // 4. Load and cache
    tex := loadImage(imgPath)
    g.cache.PutDisk(cacheKey, tex)
    g.cache.PutMemory(cacheKey, tex)

    return tex, nil
}

func (g *TextureGenerator) buildPrompt(spec TextureSpec) string {
    // Build dimension-aware prompt
    return fmt.Sprintf(
        "Create a %dx%d pixel seamless tileable texture for spaceship %s. "+
        "Style: %s. Surface type: %s. "+
        "Colors: %s primary, %s accent. "+
        "Details: %s. "+
        "Format: PNG, orthographic view, no perspective distortion. "+
        "Must tile seamlessly on all edges.",
        spec.PixelWidth(), spec.PixelHeight(),
        spec.Theme.Name,
        spec.Theme.StyleKeywords,
        spec.SurfaceType,
        colorToHex(spec.Theme.PrimaryColor),
        colorToHex(spec.Theme.AccentColor),
        g.getSurfaceStyle(spec),
    )
}
```

#### 3. Texture Cache System

Multi-level caching with lazy loading:

```go
// engine/texgen/cache.go

type TextureCache struct {
    memoryCache map[string]*ebiten.Image
    diskCache   string // Directory path
    maxMemory   int    // Max textures in memory
    lru         *list.List
}

func NewTextureCache(diskPath string, maxMemory int) *TextureCache {
    os.MkdirAll(diskPath, 0755)
    return &TextureCache{
        memoryCache: make(map[string]*ebiten.Image),
        diskCache:   diskPath,
        maxMemory:   maxMemory,
        lru:         list.New(),
    }
}

// Cache key format: assets/cache/textures/{room}_{surface}_{WxH}_{theme}_{seed}.png
func (c *TextureCache) diskPath(key string) string {
    return filepath.Join(c.diskCache, key+".png")
}

func (c *TextureCache) GetFromDisk(key string) *ebiten.Image {
    path := c.diskPath(key)
    if _, err := os.Stat(path); os.IsNotExist(err) {
        return nil
    }
    return loadImage(path)
}

func (c *TextureCache) PutDisk(key string, tex *ebiten.Image) {
    path := c.diskPath(key)
    saveImage(path, tex)
}

func (c *TextureCache) PutMemory(key string, tex *ebiten.Image) {
    // LRU eviction if at capacity
    if len(c.memoryCache) >= c.maxMemory {
        oldest := c.lru.Back()
        if oldest != nil {
            delete(c.memoryCache, oldest.Value.(string))
            c.lru.Remove(oldest)
        }
    }
    c.memoryCache[key] = tex
    c.lru.PushFront(key)
}
```

#### 4. Window System with Space Views

Windows render space through viewports with SR/GR effects:

```ailang
-- sim/window.ail

-- Window types for different room contexts
export type WindowType =
    | WindowViewport(float, float)   -- Large rectangular (width, height meters)
    | WindowPorthole(float)          -- Circular (diameter meters)
    | WindowDome(float, float)       -- Curved dome (width, height)
    | WindowNone

-- Window definition in a room
export type WindowDef = {
    id: string,
    windowType: WindowType,
    wallPosition: int,          -- 0=front, 1=right, 2=back, 3=left
    offsetX: float,             -- Horizontal offset from wall center
    offsetY: float,             -- Vertical offset from floor
    lookDirection: Vec3         -- Direction into space
}

-- Extended room with windows
export type RoomWithWindows = {
    room: RoomDef,
    windows: [WindowDef],
    props: [PropDef],
    characters: [CharacterDef]
}

-- DrawCmd for window rendering
-- Window3D(id, wallPos, x, y, width, height, windowType, z-layer)
-- The engine handles:
--   1. Clipping mask for window shape
--   2. Rendering space view to offscreen buffer
--   3. Applying SR/GR shaders to space content ONLY
--   4. Compositing window content into room
```

**Engine-side window rendering:**

```go
// engine/render/window_renderer.go

type WindowRenderer struct {
    spaceScene     *tetra.Scene      // LOD space scene
    srWarp         *shader.SRWarp    // Special relativity shader
    grWarp         *shader.GRWarp    // General relativity shader
    windowBuffers  map[string]*ebiten.Image
}

func (w *WindowRenderer) RenderWindow(
    screen *ebiten.Image,
    window WindowDef,
    roomTransform Matrix4,
    velocity float64,        // Ship velocity for SR
    nearestMass MassInfo,    // Nearest massive object for GR
) {
    // 1. Get or create offscreen buffer for this window
    buffer := w.getWindowBuffer(window)
    buffer.Clear()

    // 2. Render space scene to buffer
    //    - Camera direction from window.lookDirection
    //    - FOV based on window size and distance
    w.spaceScene.SetCameraDirection(window.LookDirection)
    spaceImg := w.spaceScene.Render()
    buffer.DrawImage(spaceImg, nil)

    // 3. Apply SR warp if moving at relativistic speed
    if velocity > 0.05 {
        w.srWarp.SetForwardVelocity(velocity)
        intermediate := ebiten.NewImage(buffer.Bounds().Dx(), buffer.Bounds().Dy())
        w.srWarp.Apply(intermediate, buffer)
        buffer = intermediate
    }

    // 4. Apply GR warp if near massive object
    if nearestMass.SchwarzschildRadius > 0.001 {
        w.grWarp.SetMassSource(nearestMass)
        intermediate := ebiten.NewImage(buffer.Bounds().Dx(), buffer.Bounds().Dy())
        w.grWarp.Apply(intermediate, buffer)
        buffer = intermediate
    }

    // 5. Apply window mask (viewport shape)
    masked := w.applyWindowMask(buffer, window.WindowType)

    // 6. Project and draw to screen at window position
    opts := w.calculateWindowTransform(window, roomTransform)
    screen.DrawImage(masked, opts)
}
```

#### 5. Parallax System for Window Depth

Create depth perception through windows:

```go
// engine/render/parallax.go

type ParallaxLayer struct {
    Name     string
    Image    *ebiten.Image
    Distance float64  // Relative distance (1.0 = reference, 0.5 = twice as close)
}

type ParallaxRenderer struct {
    layers []ParallaxLayer
}

// Layer examples:
// - Starfield background (distance: 10.0) - barely moves
// - Distant nebula (distance: 5.0) - slight movement
// - Nearby planet (distance: 0.5) - significant parallax
// - Orbital station (distance: 0.2) - strong parallax

func (p *ParallaxRenderer) Render(
    screen *ebiten.Image,
    cameraOffset Vec2,    // From player movement in room
    windowBounds Rect,
) {
    for _, layer := range p.layers {
        // Calculate parallax offset based on distance
        parallaxX := cameraOffset.X / layer.Distance
        parallaxY := cameraOffset.Y / layer.Distance

        opts := &ebiten.DrawImageOptions{}
        opts.GeoM.Translate(-parallaxX, -parallaxY)

        // Clip to window bounds
        subImg := layer.Image.SubImage(windowBounds.ToImageRect())
        screen.DrawImage(subImg.(*ebiten.Image), opts)
    }
}
```

#### 6. Room LOD System

Handle multiple rooms with level-of-detail:

```ailang
-- sim/room_lod.ail

-- LOD tiers for rooms
export type RoomLODTier =
    | RoomFull       -- Full 3D, all textures, all props
    | RoomSimplified -- Reduced props, lower-res textures
    | RoomMinimal    -- Just walls with flat colors
    | RoomUnloaded   -- Not in memory

-- Room with LOD state
export type RoomLODState = {
    room: InteriorRoom,
    tier: RoomLODTier,
    distanceFromPlayer: float,  -- In rooms/meters
    lastVisited: int            -- Frame number
}

-- Thresholds (in meters or rooms away)
pure func lodThresholdFull() -> float { 20.0 }       -- Current room + adjacent
pure func lodThresholdSimplified() -> float { 50.0 } -- 2-3 rooms away
pure func lodThresholdMinimal() -> float { 100.0 }   -- Far rooms
-- Beyond minimal = unloaded
```

### Implementation Plan

**Phase 1: Theme System** (~2 days)
- [ ] Create `sim/interior_themes.ail` with RoomTheme type
- [ ] Define themes: bridge, engineering, habitat, corridor, storage, medical
- [ ] Update RoomDef to include theme reference
- [ ] Update demo-game-interior to use themed room
- [ ] Test: Different room types render with distinct colors

**Phase 2: Dimension-Aware Texture Spec** (~2 days)
- [ ] Create `sim/texture_spec.ail` with TextureSpec type
- [ ] Add texture spec calculation based on room dimensions
- [ ] Create Go side texture generator scaffolding
- [ ] Build AI prompt generator with dimension parameters
- [ ] Test: Generate 8m x 6m floor texture (not 512x512 tiled)

**Phase 3: Texture Cache System** (~2 days)
- [ ] Create `engine/texgen/cache.go` with multi-level cache
- [ ] Implement disk cache with hash-based filenames
- [ ] Implement memory LRU cache with configurable size
- [ ] Add cache statistics and hit/miss logging
- [ ] Test: Second load uses cache, no AI regeneration

**Phase 4: Window System** (~3 days)
- [ ] Create `sim/window.ail` with WindowDef and WindowType
- [ ] Add Window3D DrawCmd to protocol
- [ ] Create `engine/render/window_renderer.go`
- [ ] Implement window mask shapes (viewport, porthole, dome)
- [ ] Integrate space scene rendering into windows
- [ ] Apply SR/GR shaders to window content only
- [ ] Test: See planets through bridge viewport with correct physics

**Phase 5: Parallax & Polish** (~2 days)
- [ ] Create `engine/render/parallax.go`
- [ ] Add parallax layers: starfield, nebula, planets, stations
- [ ] Integrate parallax with window renderer
- [ ] Add camera offset tracking from player movement
- [ ] Performance optimization pass
- [ ] Test: Parallax visible when moving in room near window

**Phase 6: Room LOD System** (~2 days)
- [ ] Create `sim/room_lod.ail` with LOD tiers
- [ ] Implement LOD tier calculation based on distance
- [ ] Add texture resolution scaling per LOD tier
- [ ] Add prop culling for simplified tier
- [ ] Test: 20+ room ship with 60 FPS

### Files to Modify/Create

**New AILANG files:**
- `sim/interior_themes.ail` - Theme definitions (~100 LOC)
- `sim/texture_spec.ail` - Texture specification types (~50 LOC)
- `sim/window.ail` - Window types and rendering (~80 LOC)
- `sim/room_lod.ail` - Room LOD system (~60 LOC)

**New Go files:**
- `engine/texgen/generator.go` - AI texture generation (~200 LOC)
- `engine/texgen/cache.go` - Multi-level texture cache (~150 LOC)
- `engine/render/window_renderer.go` - Window with space views (~300 LOC)
- `engine/render/parallax.go` - Parallax layer system (~100 LOC)

**Modified files:**
- `sim/interior.ail` - Add theme, windows to RoomDef (~50 LOC changes)
- `sim/protocol.ail` - Add Window3D DrawCmd (~20 LOC)
- `engine/render/draw_interior.go` - Handle Window3D, LOD (~100 LOC)
- `cmd/demo-game-interior/main.go` - Add window demo options (~50 LOC)
- `.claude/skills/asset-manager/SKILL.md` - Update with new workflows (~50 LOC)
- `.claude/skills/asset-manager/resources/prompt_templates.md` - Add dimension-aware templates (~100 LOC)

## Examples

### Example 1: Themed Bridge Room

**Before (generic):**
```ailang
let room = makeRoomTextured(8.0, 6.0, 3.0,
    "assets/textures/interior/bridge_floor.png",  -- Same as any room
    "assets/textures/interior/bridge_wall.png",
    "assets/textures/interior/bridge_ceiling.png"
);
```

**After (themed with exact dimensions):**
```ailang
let theme = themeBridge();
let room = makeThemedRoom(8.0, 6.0, 3.0, theme);
-- Textures auto-generated at 8m x 6m (floor), 8m x 3m (wall), etc.
-- Cached as: bridge_floor_1024x768_bridge_seed42.png
```

### Example 2: Window Rendering

```ailang
-- Define a bridge with large forward viewport
let bridge = {
    room: makeThemedRoom(8.0, 6.0, 3.0, themeBridge()),
    windows: [
        {
            id: "main_viewport",
            windowType: WindowViewport(4.0, 2.5),  -- 4m x 2.5m viewport
            wallPosition: 0,  -- Front wall
            offsetX: 0.0,     -- Centered
            offsetY: 1.5,     -- 1.5m from floor (eye level)
            lookDirection: vec3(0.0, 0.0, -1.0)  -- Forward
        }
    ],
    props: [...],
    characters: [...]
};
```

### Example 3: Asset Manager Prompt for Exact Dimensions

```bash
# Generate exact-dimension floor texture for 8m x 6m bridge floor at 128 ppm
bin/voyage ai -generate-image \
  -prompt "Create a 1024x768 pixel seamless tileable texture for spaceship bridge floor. \
           Style: command center, sleek, holographic displays. \
           Colors: #1a3050 primary, #00ccff accent. \
           Surface: brushed metal deck plating, subtle blue LED strip lighting in seams. \
           Format: PNG, orthographic top-down view, no perspective distortion. \
           Must tile seamlessly on all edges." \
  -width 1024 \
  -height 768
```

## Success Criteria

- [ ] Bridge, engineering, habitat rooms each have distinct visual identity
- [ ] Textures are generated at surface dimensions (not just 512x512 tiled)
- [ ] Texture cache prevents redundant AI calls (verify with logs)
- [ ] Windows show space scene with correct SR effects at 0.3c velocity
- [ ] Windows show space scene with GR lensing near massive objects
- [ ] Parallax visible when player moves near window
- [ ] 20+ room ship maintains 60 FPS with LOD system
- [ ] Asset manager skill updated with new dimension-aware workflows
- [ ] Demo mode: `go run ./cmd/demo-game-interior --windows --velocity 0.3`

## Testing Strategy

**Unit tests:**
- TextureSpec.CacheKey() produces unique, deterministic keys
- TextureCache LRU eviction works correctly
- RoomLODTier calculation from distance

**Integration tests:**
- Full pipeline: Theme → TextureSpec → AI generation → Cache → Render
- Window rendering with SR/GR shader chain
- Multiple rooms with LOD transitions

**Manual testing:**
- Visual inspection of themed rooms (bridge should look different from engineering)
- Window content matches LOD demo space view
- Performance profiling with many rooms

**Demo commands:**
```bash
# Basic themed interior
go run ./cmd/demo-game-interior --room bridge

# With windows
go run ./cmd/demo-game-interior --room bridge --windows

# With relativistic effects
go run ./cmd/demo-game-interior --room bridge --windows --velocity 0.5

# Performance test
go run ./cmd/demo-game-interior --rooms 50 --lod
```

## Non-Goals

**Not in this feature:**
- Procedural room layout generation - Rooms are still manually defined
- Dynamic window breaking/damage - Windows are static viewports
- Real-time AI texture regeneration - Textures cached at room creation
- Multi-player window sync - Single-player only

## Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|-----------|
| AI generation latency | High | Aggressive caching, pre-generation at level load |
| AI texture quality inconsistency | Medium | Seed-based determinism, style keywords, rejection sampling |
| SR/GR shader performance through windows | Medium | Render space at lower resolution, cache static views |
| Memory pressure with many rooms | Medium | LOD tier eviction, texture streaming |
| Window clipping complexity | Medium | Use stencil buffer for masks |

## References

- [demo-game-interior main.go](../../cmd/demo-game-interior/main.go) - Current interior demo
- [demo-lod main.go](../../cmd/demo-lod/main.go) - LOD and SR/GR shader reference
- [sim/interior.ail](../../sim/interior.ail) - Current interior AILANG
- [engine/render/draw_interior.go](../../engine/render/draw_interior.go) - Current 3D room renderer
- [02-bridge-interior.md](phase2-core-views/02-bridge-interior.md) - Bridge design with observation dome
- [asset-manager SKILL.md](../../.claude/skills/asset-manager/SKILL.md) - Current asset generation workflow

## Future Work

After this feature:
- **Procedural corridor generation** - Connect rooms with auto-generated corridors
- **Dynamic lighting through windows** - Star/planet light affects room interior
- **Window interaction** - Zoom, target lock, information overlays
- **Environment hazards visible through windows** - Asteroid fields, debris
- **Time-of-day exterior lighting** - Orbital position affects light direction

---

**Document created**: 2025-12-19
**Last updated**: 2025-12-19
