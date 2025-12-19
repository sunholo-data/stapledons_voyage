# Sprint: Interior Demo Iteration

**Design Doc:** [design_docs/planned/next/interior-demo-iteration.md](../design_docs/planned/next/interior-demo-iteration.md)
**Status:** In Progress
**Started:** 2025-12-19
**Estimated:** 8-10 days across 6 phases

## Goal

Iterate on demo-game-interior to add:
1. Room theme system for visual identity per room type
2. Dimension-aware AI texture generation
3. Texture caching (lazy generation)
4. Windows with space views (SR/GR effects)
5. Parallax depth through windows
6. Room LOD system for large ships

## Pre-Sprint Checklist

- [x] Design doc created and approved
- [x] AILANG inbox checked (no blockers)
- [x] Current sim/interior.ail compiles
- [ ] Demo-game-interior runs successfully

---

## Phase 1: Theme System (Day 1-2)

### Day 1: AILANG Theme Types

- [x] Create `sim/interior_themes.ail` with RoomTheme type
- [x] Define `themeBridge()` - blue/cyan command center
- [x] Define `themeEngineering()` - orange/industrial
- [x] Define `themeHabitat()` - warm residential
- [x] Define `themeCorridor()` - functional transit
- [x] Define `themeStorage()` - utilitarian
- [x] Define `themeMedical()` - clean/clinical
- [x] Run `ailang check sim/interior_themes.ail`

**Files created:**
- `sim/interior_themes.ail`

### Day 2: Integrate Themes into Interior

- [x] Update `sim/interior.ail` - add theme field to InteriorRoom
- [x] Create `makeThemedRoom()` function
- [x] Update `createBridgeInterior()` to include themeName
- [x] Create `createEngineeringInterior()` with `themeEngineering()`
- [x] Create `createHabitatInterior()`, `createCorridorInterior()`, `createStorageInterior()`, `createMedicalInterior()`
- [x] Create `createRoomByName()` for room selection
- [x] Create `initInteriorWithRoom()` for starting in specific room
- [x] Run `make sim` to regenerate Go code
- [x] Update demo-game-interior with `--room` flag
- [x] **TEST CHECKPOINT**: Run demo with different room types (user confirmed colors visible, fixed HUD overlap)

**Files modified:**
- `sim/interior.ail`
- `cmd/demo-game-interior/main.go`

**Test command:**
```bash
go run ./cmd/demo-game-interior --room bridge
go run ./cmd/demo-game-interior --room engineering
```

---

## Phase 2: Dimension-Aware Texture Spec (Day 3-4)

### Day 3: AILANG Texture Specification

- [x] Create `sim/texture_spec.ail` with TextureSpec type
- [x] Add `specPixelWidth()` and `specPixelHeight()` functions
- [x] Add `specCacheKey()` for deterministic cache keys
- [x] Add `buildRoomSpecs()` to generate specs from room + theme
- [x] Add `buildPrompt()` for AI prompt generation
- [x] Run `ailang check sim/texture_spec.ail`

**Files created:**
- `sim/texture_spec.ail` (~215 LOC)

### Day 4: Go Texture Generator Scaffold

- [x] Create `engine/texgen/generator.go` with TextureGenerator
- [x] Implement prompt building via AILANG `BuildPrompt()`
- [x] Add integration with existing AI handler
- [x] Add async generation support for room textures
- [ ] **TEST CHECKPOINT**: Generate single texture via prompt

**Files created:**
- `engine/texgen/generator.go` (~175 LOC)

**Test command:**
```bash
# Manual test: generate 1024x768 floor texture
bin/voyage ai -generate-image -prompt "Create a 1024x768 pixel seamless tileable texture for spaceship bridge floor..."
```

---

## Phase 3: Texture Cache System (Day 5)

### Day 5: Multi-Level Cache

- [x] Create `engine/texgen/cache.go` with TextureCache
- [x] Implement disk cache with hash-based paths (SHA256)
- [x] Implement memory LRU cache with configurable capacity
- [x] Add cache hit/miss logging
- [x] Integrate cache into TextureGenerator
- [ ] Update draw_interior.go to use cached textures
- [ ] **TEST CHECKPOINT**: Verify cache prevents re-generation

**Files created:**
- `engine/texgen/cache.go` (~165 LOC)

**Files modified:**
- `engine/texgen/generator.go`

**Test command:**
```bash
# Run twice, second should use cache (check logs)
go run ./cmd/demo-game-interior --room bridge
go run ./cmd/demo-game-interior --room bridge
```

---

## Phase 4: Window System (Day 6-8)

### Day 6: AILANG Window Types

- [ ] Create `sim/window.ail` with WindowType ADT
- [ ] Add WindowDef record type
- [ ] Add RoomWithWindows type
- [ ] Update protocol.ail with Window3D DrawCmd
- [ ] Run `ailang check sim/window.ail`
- [ ] Run `make sim` to regenerate

**Files created:**
- `sim/window.ail`

**Files modified:**
- `sim/protocol.ail`

### Day 7: Engine Window Renderer

- [ ] Create `engine/render/window_renderer.go`
- [ ] Implement window mask shapes (viewport, porthole)
- [ ] Create offscreen buffer for space rendering
- [ ] Integrate existing LOD space scene
- [ ] Handle Window3D in draw.go switch

**Files created:**
- `engine/render/window_renderer.go`

**Files modified:**
- `engine/render/draw.go`

### Day 8: SR/GR Shader Integration

- [ ] Apply SR warp shader to window content only
- [ ] Apply GR warp shader to window content only
- [ ] Add velocity parameter to window rendering
- [ ] Add mass info for GR effects
- [ ] Update demo with `--velocity` and `--windows` flags
- [ ] **TEST CHECKPOINT**: See planets through window with effects

**Files modified:**
- `engine/render/window_renderer.go`
- `cmd/demo-game-interior/main.go`

**Test command:**
```bash
go run ./cmd/demo-game-interior --room bridge --windows --velocity 0.3
```

---

## Phase 5: Parallax & Polish (Day 9)

### Day 9: Parallax Layers

- [ ] Create `engine/render/parallax.go`
- [ ] Define parallax layers (starfield, nebula, planets)
- [ ] Integrate with window renderer
- [ ] Track camera offset from player movement
- [ ] Performance optimization pass
- [ ] **TEST CHECKPOINT**: Parallax visible when moving near window

**Files created:**
- `engine/render/parallax.go`

**Files modified:**
- `engine/render/window_renderer.go`

**Test command:**
```bash
go run ./cmd/demo-game-interior --room bridge --windows
# Move with WASD near window, observe parallax
```

---

## Phase 6: Room LOD System (Day 10)

### Day 10: LOD Tiers

- [ ] Create `sim/room_lod.ail` with RoomLODTier ADT
- [ ] Add distance-based tier calculation
- [ ] Implement texture resolution scaling per tier
- [ ] Add prop culling for simplified tier
- [ ] Create multi-room test scenario
- [ ] **FINAL TEST**: 20+ rooms at 60 FPS

**Files created:**
- `sim/room_lod.ail`

**Files modified:**
- `engine/render/draw_interior.go`
- `cmd/demo-game-interior/main.go`

**Test command:**
```bash
go run ./cmd/demo-game-interior --rooms 20 --lod
```

---

## Test Checkpoints Summary

| Phase | Checkpoint | Command |
|-------|------------|---------|
| 1 | Themed rooms render differently | `--room bridge` vs `--room engineering` |
| 2 | Dimension-aware texture generated | Manual voyage ai test |
| 3 | Cache prevents re-generation | Run twice, check logs |
| 4 | Windows show space with SR/GR | `--windows --velocity 0.3` |
| 5 | Parallax when moving | WASD near window |
| 6 | 60 FPS with 20 rooms | `--rooms 20 --lod` |

---

## AILANG Feedback Checkpoint

After sprint completion, report:
- [ ] Any AILANG bugs encountered
- [ ] Features that would have helped
- [ ] Documentation gaps
- [ ] Performance observations

---

## Files Summary

### New AILANG Files
- `sim/interior_themes.ail` (~100 LOC)
- `sim/texture_spec.ail` (~50 LOC)
- `sim/window.ail` (~80 LOC)
- `sim/room_lod.ail` (~60 LOC)

### New Go Files
- `engine/texgen/generator.go` (~200 LOC)
- `engine/texgen/cache.go` (~150 LOC)
- `engine/texgen/spec.go` (~50 LOC)
- `engine/render/window_renderer.go` (~300 LOC)
- `engine/render/parallax.go` (~100 LOC)

### Modified Files
- `sim/interior.ail` (~50 LOC changes)
- `sim/protocol.ail` (~20 LOC)
- `engine/render/draw_interior.go` (~100 LOC)
- `engine/render/draw.go` (~20 LOC)
- `cmd/demo-game-interior/main.go` (~100 LOC)

---

**Sprint created:** 2025-12-19
**Last updated:** 2025-12-19
