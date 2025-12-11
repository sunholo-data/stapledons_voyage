# Galaxy Map Mode

**Version:** 0.5.2
**Status:** Planned
**Priority:** P0 (Core Navigation)
**Complexity:** High
**AILANG Workarounds:** Recursion depth (use viewport culling), lookup by ID (use Arrays + extern)
**Depends On:** v0.5.0 UI Modes Framework, v0.5.1 Ship Exploration (for transition)

## Related Documents

- [UI Modes Architecture](../v0_5_0/ui-modes.md) - Mode framework
- [Journey Planning](../v0_6_0/journey-planning.md) - Triggered from star selection
- [Civilization Detail](../v0_6_1/civilization-detail.md) - Opened from star click
- [Game Vision](../../../docs/game-vision.md) - Galaxy-scale gameplay

## Problem Statement

The galaxy map is the strategic heart of the game. Players need to:
- Visualize the entire galaxy and their position
- Understand civilization states at a glance
- Plan routes considering time dilation
- See the contact network they're building

**Current State:**
- No galaxy visualization
- No star/civilization data display
- No pan/zoom navigation
- No network graph rendering

**What's Needed:**
- Procedural starfield background
- Star nodes with civilization status indicators
- Contact network edge rendering
- Pan and zoom controls
- Time dilation preview
- Filtering and search

---

## Design Overview

### Map Philosophy

The galaxy map should evoke the **loneliness and vastness** of space travel:

- **Stars are sparse** - Most of space is empty
- **Connections are precious** - Network edges show hard-won contact
- **Time is visible** - See when you last visited, how much has changed
- **Choices have weight** - Previewing a journey shows what you'll miss

### Visual Layers

```
Layer 0: Deep space gradient (static)
Layer 1: Distant star particles (parallax slow)
Layer 2: Mid-ground stars (parallax medium)
Layer 3: Star nodes (main interactive layer)
Layer 4: Network edges (between visited stars)
Layer 5: UI overlays (tooltips, sidebars)
```

### Information Density

The map must balance **overview** with **detail**:

- Zoomed out: See whole galaxy, major hubs, network shape
- Zoomed in: See individual stars, read names, status details
- Selected: Full civilization panel on side

---

## Detailed Specification

### Galaxy Data Structure

```ailang
module sim/galaxy

type Galaxy = {
    stars: [Star],
    edges: [ContactEdge],
    bounds: Rect,
    currentPosition: StarID,
    yearVisited: [(StarID, int)]  -- when each star was visited
}

type Star = {
    id: StarID,
    name: string,
    x: float,
    y: float,
    starClass: StarClass,
    civilization: Maybe(CivilizationID),
    lastKnownState: CivState,
    lastVisitYear: Maybe(int),
    distanceFromPlayer: float  -- cached for display
}

type StarClass = ClassO | ClassB | ClassA | ClassF | ClassG | ClassK | ClassM

type CivState =
    | Unknown                      -- Never visited/contacted
    | Thriving(int)                -- Population tier
    | Declining(int)               -- Years of decline
    | Extinct(int)                 -- Year of extinction
    | Transcended(int)             -- Year of transcendence
    | PreContact                   -- Exists but not space-faring

type ContactEdge = {
    star1: StarID,
    star2: StarID,
    establishedYear: int,
    edgeType: EdgeType
}

type EdgeType =
    | PlayerVisited                -- You connected these
    | IndigenousContact            -- They contacted each other
    | LostConnection               -- Was connected, one is gone
```

### Galaxy Map State

```ailang
type GalaxyMapState = {
    cameraX: float,
    cameraY: float,
    zoomLevel: float,              -- 0.1 to 5.0
    hoveredStar: Maybe(StarID),
    selectedStar: Maybe(StarID),
    filterMode: MapFilter,
    showNetwork: bool,
    showLabels: bool,
    searchQuery: string,
    searchResults: [StarID],
    journeyPreview: Maybe(JourneyPreview),
    sidePanel: Maybe(SidePanelContent)
}

type MapFilter =
    | AllStars
    | VisitedOnly
    | UnvisitedOnly
    | HasCivilization
    | ExtinctCivs
    | ReachableInYears(int)        -- Stars you could reach in N subjective years

type SidePanelContent =
    | CivSummary(CivilizationID)
    | JourneyPlan(JourneyPreview)
    | StarInfo(StarID)

type JourneyPreview = {
    destination: StarID,
    distance: float,
    velocityOptions: [VelocityOption]
}

type VelocityOption = {
    velocity: float,               -- fraction of c
    subjectiveYears: float,
    objectiveYears: float,
    arrivalYear: int,
    crewDeaths: int,               -- estimated deaths during transit
    crewBirths: int                -- estimated births
}
```

### Input Processing

```ailang
-- Process galaxy map input
pure func processGalaxyMapInput(state: GalaxyMapState, galaxy: Galaxy, input: FrameInput) -> GalaxyMapState {
    let afterZoom = processZoom(state, input);
    let afterPan = processPan(afterZoom, input);
    let afterHover = processStarHover(afterPan, galaxy, input);
    let afterClick = processStarClick(afterHover, galaxy, input);
    let afterKeys = processMapKeys(afterClick, input);
    afterKeys
}

-- Mouse wheel zoom
pure func processZoom(state: GalaxyMapState, input: FrameInput) -> GalaxyMapState {
    let newZoom = clamp(state.zoomLevel + input.scrollY * 0.1, 0.1, 5.0);
    { state | zoomLevel: newZoom }
}

-- Click-drag pan or keyboard pan
pure func processPan(state: GalaxyMapState, input: FrameInput) -> GalaxyMapState {
    let panSpeed = 10.0 / state.zoomLevel;
    let dx = intToFloat(input.moveX) * panSpeed;
    let dy = intToFloat(input.moveY) * panSpeed;
    if input.dragging then
        { state | cameraX: state.cameraX - input.dragDeltaX / state.zoomLevel,
                  cameraY: state.cameraY - input.dragDeltaY / state.zoomLevel }
    else
        { state | cameraX: state.cameraX + dx, cameraY: state.cameraY + dy }
}

-- Find star under mouse
pure func processStarHover(state: GalaxyMapState, galaxy: Galaxy, input: FrameInput) -> GalaxyMapState {
    let worldPos = screenToWorld(input.mouseX, input.mouseY, state);
    let hoveredStar = findNearestStar(galaxy.stars, worldPos, state.zoomLevel);
    { state | hoveredStar: hoveredStar }
}

-- Click to select star
pure func processStarClick(state: GalaxyMapState, galaxy: Galaxy, input: FrameInput) -> GalaxyMapState {
    if input.clicked then
        match state.hoveredStar {
            None => { state | selectedStar: None, sidePanel: None },
            Some(starID) => {
                let star = findStar(galaxy.stars, starID);
                let panel = match star.civilization {
                    None => StarInfo(starID),
                    Some(civID) => CivSummary(civID)
                };
                { state | selectedStar: Some(starID), sidePanel: Some(panel) }
            }
        }
    else if input.rightClicked then
        -- Right-click to plan journey
        match state.hoveredStar {
            None => state,
            Some(starID) => {
                let preview = calculateJourneyPreview(galaxy, starID);
                { state | journeyPreview: Some(preview), sidePanel: Some(JourneyPlan(preview)) }
            }
        }
    else
        state
}

-- Keyboard shortcuts
pure func processMapKeys(state: GalaxyMapState, input: FrameInput) -> GalaxyMapState {
    if input.keyN then { state | showNetwork: not(state.showNetwork) }
    else if input.keyL then { state | showLabels: not(state.showLabels) }
    else if input.keyF then { state | filterMode: cycleFilter(state.filterMode) }
    else if input.keyHome then centerOnPlayer(state)
    else state
}
```

### Time Dilation Calculator

```ailang
-- Calculate relativistic journey details
pure func calculateJourney(distance: float, velocity: float, currentYear: int) -> JourneyDetails {
    -- Lorentz factor: gamma = 1 / sqrt(1 - v^2/c^2)
    let gamma = 1.0 / sqrt(1.0 - velocity * velocity);

    -- Objective time (outside observer)
    let objectiveTime = distance / velocity;

    -- Subjective time (on ship)
    let subjectiveTime = objectiveTime / gamma;

    {
        distanceLightYears: distance,
        velocity: velocity,
        gamma: gamma,
        subjectiveYears: subjectiveTime,
        objectiveYears: objectiveTime,
        arrivalYear: currentYear + floatToInt(objectiveTime),
        departureYear: currentYear
    }
}

-- Generate velocity options for journey preview
pure func calculateJourneyPreview(galaxy: Galaxy, destID: StarID) -> JourneyPreview {
    let dest = findStar(galaxy.stars, destID);
    let current = findStar(galaxy.stars, galaxy.currentPosition);
    let distance = starDistance(current, dest);

    let velocities = [0.9, 0.95, 0.99, 0.999, 0.9999, 0.99999];
    let options = map(\v. velocityToOption(distance, v, galaxy.currentYear), velocities);

    { destination: destID, distance: distance, velocityOptions: options }
}
```

### Rendering Logic

```ailang
-- Generate draw commands for galaxy map
pure func renderGalaxyMap(world: World, state: GalaxyMapState) -> [DrawCmd] {
    let starCmds = renderStars(world.galaxy.stars, state);
    let edgeCmds = if state.showNetwork then renderEdges(world.galaxy.edges, state) else [];
    let labelCmds = if state.showLabels then renderLabels(world.galaxy.stars, state) else [];
    let uiCmds = renderMapUI(state, world);
    concat(concat(concat(starCmds, edgeCmds), labelCmds), uiCmds)
}

-- Render star nodes
pure func renderStars(stars: [Star], state: GalaxyMapState) -> [DrawCmd] {
    let visibleStars = filterVisible(stars, state);
    map(\s. starToDrawCmd(s, state), visibleStars)
}

pure func starToDrawCmd(star: Star, state: GalaxyMapState) -> DrawCmd {
    let screenPos = worldToScreen(star.x, star.y, state);
    let size = starSize(star, state.zoomLevel);
    let color = civStateToColor(star.lastKnownState);
    let highlight = match state.selectedStar {
        Some(id) if id == star.id => true,
        _ => false
    };
    let hover = match state.hoveredStar {
        Some(id) if id == star.id => true,
        _ => false
    };
    StarNode(screenPos.x, screenPos.y, size, color, highlight, hover, 3)
}

-- Color coding for civilization states
pure func civStateToColor(state: CivState) -> int {
    match state {
        Unknown => 1,            -- Blue (mysterious)
        Thriving(_) => 2,        -- Green (alive)
        Declining(_) => 3,       -- Yellow (warning)
        Extinct(_) => 4,         -- Gray (dead)
        Transcended(_) => 5,     -- White (ascended)
        PreContact => 6          -- Dim blue (not ready)
    }
}

-- Render contact network edges
pure func renderEdges(edges: [ContactEdge], state: GalaxyMapState) -> [DrawCmd] {
    map(\e. edgeToDrawCmd(e, state), edges)
}

pure func edgeToDrawCmd(edge: ContactEdge, state: GalaxyMapState) -> DrawCmd {
    let pos1 = worldToScreen(edge.star1.x, edge.star1.y, state);
    let pos2 = worldToScreen(edge.star2.x, edge.star2.y, state);
    let color = edgeTypeToColor(edge.edgeType);
    Line(pos1.x, pos1.y, pos2.x, pos2.y, color, 2)  -- Z=2, below stars
}
```

---

## Visual Design

### Color Palette

| Element | Color | Hex |
|---------|-------|-----|
| Background | Deep space blue-black | #0a0a1a |
| Unknown star | Blue | #4488ff |
| Thriving civ | Green | #44ff88 |
| Declining civ | Yellow | #ffcc44 |
| Extinct civ | Gray | #666666 |
| Transcended | White glow | #ffffff |
| Player position | Gold | #ffdd00 |
| Network edge | Cyan | #00cccc |
| Lost connection | Red dashed | #cc4444 |

### Star Sizes (at zoom 1.0)

| Zoom Level | Base Star | With Civ | Selected |
|------------|-----------|----------|----------|
| 0.1 | 2px | 3px | 5px |
| 1.0 | 6px | 8px | 12px |
| 3.0 | 12px | 16px | 24px |

### UI Layout

```
┌────────────────────────────────────────────────────────────┐
│ [Filter: All ▼] [Show Network ✓] [Show Labels ✓]  [Search]│
├────────────────────────────────────────────────────────────┤
│                                                            │
│                                                            │
│                    GALAXY MAP AREA                         │
│                                                            │
│                        ★ Player                            │
│                                                            │
├─────────────────────────────────┬──────────────────────────┤
│ Year: 3,247                     │ SELECTED STAR            │
│ Subjective: 47.3 years          │ Tau Ceti                 │
│ Crew: 23 / Capacity: 50         │ ────────────────────     │
│                                 │ Civilization: Thrivine   │
│ [Return to Ship]                │ Last Visit: Year 2,891   │
│                                 │ [View Details] [Journey] │
└─────────────────────────────────┴──────────────────────────┘
```

---

## Go/Engine Integration

### Galaxy Renderer

```go
// engine/render/galaxy.go

type GalaxyRenderer struct {
    starfield    *ebiten.Image  // Pre-rendered background
    starSprites  *SpriteSheet   // Star node graphics
    edgeShader   *ebiten.Shader // For glowing edges (optional)
}

func (r *GalaxyRenderer) Render(screen *ebiten.Image, galaxy Galaxy, state GalaxyMapState) {
    // Draw starfield background with parallax
    r.drawStarfield(screen, state)

    // Draw network edges (if enabled)
    if state.ShowNetwork {
        r.drawEdges(screen, galaxy.Edges, state)
    }

    // Draw star nodes
    visibleStars := r.getVisibleStars(galaxy.Stars, state)
    for _, star := range visibleStars {
        r.drawStar(screen, star, state)
    }

    // Draw labels (if enabled and zoomed enough)
    if state.ShowLabels && state.ZoomLevel > 0.5 {
        for _, star := range visibleStars {
            r.drawLabel(screen, star, state)
        }
    }

    // Draw player position indicator
    r.drawPlayerMarker(screen, galaxy.CurrentPosition, state)

    // Draw UI panels
    r.drawMapUI(screen, state)
}

func (r *GalaxyRenderer) drawStar(screen *ebiten.Image, star Star, state GalaxyMapState) {
    screenX, screenY := worldToScreen(star.X, star.Y, state)

    // Skip if off-screen
    if screenX < -50 || screenX > screenWidth+50 || screenY < -50 || screenY > screenHeight+50 {
        return
    }

    size := r.starSize(star, state.ZoomLevel)
    color := r.civStateColor(star.LastKnownState)

    // Draw glow for selected/hovered
    if state.SelectedStar == star.ID || state.HoveredStar == star.ID {
        r.drawGlow(screen, screenX, screenY, size*2, color)
    }

    // Draw star node
    r.drawCircle(screen, screenX, screenY, size, color)

    // Draw civilization indicator ring
    if star.Civilization != nil {
        r.drawRing(screen, screenX, screenY, size+2, color)
    }
}
```

### Coordinate Transforms

```go
// engine/render/coords.go

func worldToScreen(worldX, worldY float64, state GalaxyMapState) (float64, float64) {
    // Apply camera offset
    relX := worldX - state.CameraX
    relY := worldY - state.CameraY

    // Apply zoom
    screenX := relX*state.ZoomLevel + screenWidth/2
    screenY := relY*state.ZoomLevel + screenHeight/2

    return screenX, screenY
}

func screenToWorld(screenX, screenY float64, state GalaxyMapState) (float64, float64) {
    // Inverse of worldToScreen
    relX := (screenX - screenWidth/2) / state.ZoomLevel
    relY := (screenY - screenHeight/2) / state.ZoomLevel

    worldX := relX + state.CameraX
    worldY := relY + state.CameraY

    return worldX, worldY
}
```

---

## Implementation Plan

### Phase 1: Basic Starfield

| Task | File | Description |
|------|------|-------------|
| 1.1 | `sim_gen/galaxy.go` | Galaxy, Star types |
| 1.2 | `sim_gen/galaxy.go` | Generate test galaxy (50 stars) |
| 1.3 | `engine/render/galaxy.go` | Background gradient |
| 1.4 | `engine/render/galaxy.go` | Star point rendering |
| 1.5 | Test | See stars on screen |

### Phase 2: Pan and Zoom

| Task | File | Description |
|------|------|-------------|
| 2.1 | `sim_gen/types.go` | GalaxyMapState type |
| 2.2 | `sim_gen/funcs.go` | Camera movement logic |
| 2.3 | `engine/input/galaxy.go` | Mouse drag, scroll wheel |
| 2.4 | `engine/render/galaxy.go` | Apply camera transforms |
| 2.5 | Test | Pan around, zoom in/out |

### Phase 3: Star Interaction

| Task | File | Description |
|------|------|-------------|
| 3.1 | `sim_gen/funcs.go` | Find star under cursor |
| 3.2 | `sim_gen/funcs.go` | Hover/select state |
| 3.3 | `engine/render/galaxy.go` | Highlight rendering |
| 3.4 | Test | Hover shows tooltip |

### Phase 4: Civilization Indicators

| Task | File | Description |
|------|------|-------------|
| 4.1 | `sim_gen/galaxy.go` | CivState type |
| 4.2 | `sim_gen/funcs.go` | Color coding |
| 4.3 | `engine/render/galaxy.go` | State-based colors |
| 4.4 | Test | See civ states visually |

### Phase 5: Network Edges

| Task | File | Description |
|------|------|-------------|
| 5.1 | `sim_gen/galaxy.go` | ContactEdge type |
| 5.2 | `sim_gen/funcs.go` | Edge generation from visits |
| 5.3 | `engine/render/galaxy.go` | Line rendering |
| 5.4 | Test | See contact network |

### Phase 6: Side Panel

| Task | File | Description |
|------|------|-------------|
| 6.1 | `engine/render/galaxy.go` | Panel layout |
| 6.2 | `sim_gen/funcs.go` | Panel content generation |
| 6.3 | `engine/render/ui.go` | Text rendering in panels |
| 6.4 | Test | Click star → see info |

### Phase 7: Journey Preview

| Task | File | Description |
|------|------|-------------|
| 7.1 | `sim_gen/galaxy.go` | JourneyPreview type |
| 7.2 | `sim_gen/funcs.go` | Time dilation calculation |
| 7.3 | `engine/render/galaxy.go` | Preview display |
| 7.4 | Test | Right-click shows journey options |

---

## Testing Strategy

### Manual Testing

```bash
make run-mock
# 1. Start game, navigate to galaxy map
# 2. See stars rendered
# 3. Scroll to zoom in/out
# 4. Drag to pan
# 5. Hover star → see highlight
# 6. Click star → see side panel
# 7. Right-click → see journey preview
# 8. Press N → toggle network
# 9. Press Home → center on player
```

### Automated Testing

```go
func TestWorldToScreenTransform(t *testing.T)
func TestScreenToWorldTransform(t *testing.T)
func TestStarHoverDetection(t *testing.T)
func TestZoomLimits(t *testing.T)
func TestTimeDilationCalculation(t *testing.T)
func TestJourneyPreviewGeneration(t *testing.T)
```

### Headless Scenarios

```go
scenario := Scenario{
    Name: "select_distant_star",
    Steps: []Step{
        {Input: KeyPress(KeyM), Frames: 1},     // Open galaxy map
        {Input: ScrollDown(10), Frames: 5},     // Zoom out
        {Input: ClickAt(400, 300), Frames: 1},  // Click a star
        {Assert: SidePanelVisible()},
    },
}
```

---

## AILANG Constraints

**Updated for v0.5.2** - Arrays and extern functions now available.

| Limitation | Impact | Workaround | Status |
|------------|--------|------------|--------|
| Recursion depth | Can't iterate 1000+ stars | Viewport culling, only process visible | Still needed |
| List O(n) lookup | Finding star by ID slow | Use `Array[Star]` with O(1) `get` | Improved with Arrays |
| Array O(n) update | Modifying star data expensive | Batch updates, rebuild once | Still relevant |
| No mutable state | Camera position updates | Functional state passing | By design |
| Float precision | Large galaxy coordinates | Normalize to manageable range | Still relevant |
| Complex queries | Octree/spatial queries slow in pure AILANG | Use `extern func` for Go impl | New option in v0.5.2+ |

### Optimization: Spatial Regions with Arrays

```ailang
import std/array as A

-- Use Arrays for O(1) star lookup by index
type GalaxyData = {
    stars: Array[Star],           -- O(1) access by StarID (if ID = index)
    regions: [GalaxyRegion]       -- Spatial partitioning
}

-- For complex spatial queries, delegate to Go via extern
extern func queryStarsInRect(stars: Array[Star], rect: Rect) -> [Star]
```

### When to Use extern func (v0.5.2+)

Use Go implementations via `extern func` for:
- Octree/spatial data structures
- Pathfinding algorithms (A*, Dijkstra)
- Large-scale iteration (1000+ items)
- Performance-critical loops

---

## Performance Considerations

### Star Count Limits

| Stars | Edges | Performance |
|-------|-------|-------------|
| 100 | 50 | Excellent |
| 500 | 200 | Good |
| 1000 | 500 | Needs culling |
| 5000+ | 2000+ | Requires LOD |

### Rendering Optimizations

1. **Viewport culling** - Only render visible stars
2. **LOD for zoom** - Fewer details when zoomed out
3. **Edge batching** - Draw all edges in one pass
4. **Label throttling** - Only show labels above zoom threshold

---

## Success Criteria

### Core Functionality
- [ ] Galaxy renders with 100+ stars
- [ ] Pan and zoom work smoothly
- [ ] Stars highlight on hover
- [ ] Click selects star

### Information Display
- [ ] Civ states color-coded
- [ ] Network edges visible
- [ ] Side panel shows star info
- [ ] Labels visible at high zoom

### Journey Integration
- [ ] Right-click opens journey preview
- [ ] Time dilation numbers correct
- [ ] Can transition to journey planning

### Performance
- [ ] 60 FPS with 500 stars
- [ ] Smooth zoom animation
- [ ] No lag on pan

---

## Future Extensions

| Feature | Description |
|---------|-------------|
| Search | Type to find star by name |
| Filters | Show only certain civ states |
| Path preview | Draw potential route |
| Timeline scrub | See galaxy at different years |
| Fog of war | Only show visited regions |
| Animated edges | Pulse to show activity |
| 3D tilt | Slight perspective (still 2D) |
