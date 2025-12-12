# Game Capabilities Reference (AILANG)

**Status**: Active (Updated 2025-12-12)
**Purpose**: Complete reference for game features implemented in AILANG (`sim/*.ail`)
**Audience**: Sprint executor, design docs, AI agents working on the game

---

## Quick Reference

| Module | File | Status | Demo |
|--------|------|--------|------|
| Protocol (DrawCmd, Input) | `protocol.ail` | Working | - |
| World State | `world.ail` | Working | `game` |
| Step Function | `step.ail` | Working | `game` |
| NPC AI | `npc_ai.ail` | Working | `game` |
| Celestial System | `celestial.ail` | Working | `demo-game-orbital` |
| Starmap Data | `starmap.ail` | Working | `demo-game-starmap` |
| Galaxy Model | `galaxy_model.ail` | Working | `demo-game-starmap` |
| Bridge View | `bridge.ail` | Working | `demo-game-bridge` |
| Depth/Parallax | `depth.ail` | Working | `demo-game-parallax` |
| Viewport System | `viewport.ail` | Working | `demo-game-bridge` |
| Ship Levels | `ship_levels.ail` | Working | - |
| Arrival Sequence | `arrival.ail` | Working | - |

---

## 1. Protocol Types (`sim/protocol.ail`)

Core types for engine communication.

### Input Types

```ailang
export type Coord = { x: int, y: int }
export type Camera = { x: float, y: float, zoom: float }

export type MouseState = {
    x: float, y: float,
    worldX: float, worldY: float,
    leftPressed: bool, rightPressed: bool
}

export type KeyEvent = { key: string, pressed: bool }
export type ClickKind = ClickLeft | ClickRight | ClickMiddle

export type PlayerAction =
    | ActionNone
    | ActionInspect
    | ActionBuild(StructureType)
    | ActionClear

export type FrameInput = {
    mouse: MouseState,
    keys: [KeyEvent],
    dt: float,
    action: PlayerAction
}
```

### Output Types (DrawCmd)

```ailang
export type DrawCmd =
    | Sprite(id: int, x: float, y: float, z: int)
    | Rect(x: float, y: float, w: float, h: float, color: int, z: int)
    | RectScreen(x: float, y: float, w: float, h: float, color: int, z: int)
    | Text(text: string, x: float, y: float, fontSize: int, color: int, z: int)
    | TextWrapped(text: string, x: float, y: float, maxWidth: int, fontSize: int, color: int, z: int)
    | Line(x1: float, y1: float, x2: float, y2: float, color: int, width: float, z: int)
    | Circle(x: float, y: float, radius: float, color: int, filled: bool, z: int)
    | CircleRGBA(x: float, y: float, radius: float, rgba: int, filled: bool, z: int)
    | RectRGBA(x: float, y: float, w: float, h: float, rgba: int, z: int)
    | IsoTile(tile: Coord, height: int, spriteId: int, layer: int, color: int)
    | IsoEntity(id: int, tile: Coord, offsetX: float, offsetY: float, height: int, spriteId: int, layer: int)
    | GalaxyBg(opacity: float, z: int, skyViewMode: bool, viewLon: float, viewLat: float, fov: float)
    | Star(x: float, y: float, spriteId: int, scale: float, alpha: float, z: int)
    | SpireBg(z: int)
    | Marker(x: float, y: float, w: float, h: float, rgba: int, parallaxLayer: int, z: int)
    | Ui(id: string, kind: UiKind, x: float, y: float, w: float, h: float, text: string, spriteId: int, z: int, color: int, value: float)
    | Viewport(id: string, shapeType: int, shapeParams: [float], contentType: int, contentParams: [float], effectType: int, effectParams: [float], layer: int, edgeBlend: float, opacity: float, screenX: float, screenY: float, z: int)
    | TexturedCircle(x: float, y: float, radius: float, spriteId: int, rotation: float, z: int)
    | OrbitPath(centerX: float, centerY: float, radiusX: float, radiusY: float, rgba: int, z: int)

export type UiKind =
    | UiPanel | UiButton | UiLabel | UiPortrait | UiSlider | UiProgressBar

export type FrameOutput = {
    commands: [DrawCmd],
    debugMessages: [string]
}
```

---

## 2. World State (`sim/world.ail`)

Top-level game state.

```ailang
export type Tile = { biome: int }

export type PlanetState = {
    tiles: [[Tile]],
    width: int,
    height: int
}

export type Selection =
    | SelectNone
    | SelectTile(Coord)
    | SelectNPC(int)

export type ViewMode =
    | ViewPlanet
    | ViewGalaxy
    | ViewBridge
    | ViewStarmap

export type World = {
    tick: int,
    planet: PlanetState,
    npcs: [NPC],
    camera: Camera,
    selection: Selection,
    view: ViewMode,
    -- Additional state fields...
}
```

---

## 3. Step Function (`sim/step.ail`)

Main game loop entry points.

```ailang
-- Initialize world with seed for determinism
export func init_world(seed: int) -> World

-- Process one frame: input -> updated world + draw commands
export func step(world: World, input: FrameInput) -> (World, FrameOutput) ! {Rand}
```

**Usage from Go:**
```go
world := sim_gen.Init_world(42)  // Seed for determinism
world, output := sim_gen.Step(world, input)
render.DrawFrame(output.Commands)
```

---

## 4. NPC AI (`sim/npc_ai.ail`)

NPC movement and behavior.

```ailang
export type Direction = North | South | East | West

export type MovementPattern =
    | PatternStatic
    | PatternPatrol([Direction])
    | PatternRandomWalk(int)  -- range

export type NPC = {
    id: int,
    name: string,
    x: int, y: int,
    spriteId: int,
    pattern: MovementPattern,
    patrolIndex: int,
    moveTimer: float
}

-- Update single NPC (uses Rand effect)
export func updateNPC(npc: NPC, width: int, height: int) -> NPC ! {Rand}

-- Update all NPCs
export func updateAllNPCs(npcs: [NPC], width: int, height: int) -> [NPC] ! {Rand}
```

---

## 5. Celestial System (`sim/celestial.ail`)

Planet and star system simulation.

```ailang
export type PlanetType =
    | Rocky | GasGiant | IceGiant | Terrestrial | Ocean | Volcanic | Dwarf

export type SpectralClass = O | B | A | F | G | K | M

export type CelestialPlanet = {
    id: int,
    name: string,
    planetType: PlanetType,
    orbitRadius: float,      -- AU
    orbitPeriod: float,      -- Earth years
    orbitalAngle: float,     -- Current position (radians)
    radius: float,           -- Earth radii
    color: Color,
    hasAtmosphere: bool,
    hasRings: bool
}

export type StarSystem = {
    name: string,
    starType: StarType,
    position: SystemPos,
    planets: [CelestialPlanet]
}

-- Create Sol system with 8 planets
export pure func initSolSystem() -> StarSystem

-- Update orbital positions
export pure func stepSystem(system: StarSystem, dt: float) -> StarSystem

-- Rendering
export pure func renderPlanet2D(planet: CelestialPlanet, centerX: float, centerY: float, scale: float) -> DrawCmd
export pure func renderOrbitPath(planet: CelestialPlanet, centerX: float, centerY: float, scale: float) -> DrawCmd
export pure func renderPlanetTextured(planet: CelestialPlanet, centerX: float, centerY: float, scale: float) -> DrawCmd
```

**Demo:** `go run ./cmd/demo-game-orbital`

---

## 6. Starmap Data (`sim/starmap.ail`)

Local stellar neighborhood catalog.

```ailang
export type Vec3 = { x: float, y: float, z: float }

export type SpectralType = O | B | A | F | G | K | M

export type Star = {
    id: int,
    name: string,
    position: Vec3,          -- Light years from Sol
    spectralType: SpectralType,
    luminosity: float,
    hasHabitableZone: bool
}

export type StarCatalog = {
    stars: [Star],
    count: int
}

-- Catalog operations
export pure func emptyCatalog() -> StarCatalog
export pure func addStar(cat: StarCatalog, star: Star) -> StarCatalog
export pure func starCount(cat: StarCatalog) -> int
export pure func starsWithinRadius(cat: StarCatalog, center: Vec3, radius: float) -> [Star]
export pure func nearestStar(cat: StarCatalog, pos: Vec3) -> Star

-- Star creation
export pure func makeStar(id: int, name: string, x: float, y: float, z: float, spec: SpectralType, hasHZ: bool) -> Star
export pure func makeSol() -> Star
export pure func initLocalCatalog() -> StarCatalog  -- ~50 nearest stars

-- Visual properties
export pure func spectralColor(spec: SpectralType) -> int  -- RGBA packed
export pure func luminosityForSpectral(spec: SpectralType) -> float
```

**Demo:** `go run ./cmd/demo-game-starmap`

---

## 7. Galaxy Model (`sim/galaxy_model.ail`)

Procedural galaxy generation.

```ailang
-- Density function for stellar distribution
export pure func stellarDensity(pos: Vec3) -> float

-- Generate star from cell coordinates (deterministic)
export pure func generateStar(seed: int, cellX: int, cellY: int, cellZ: int, cellSize: float) -> Star

-- Blend weight for real/procedural stars
export pure func blendWeight(distanceFromSol: float) -> float

-- Should a cell have a star?
export pure func shouldGenerateStar(seed: int, cellX: int, cellY: int, cellZ: int, cellSize: float) -> bool
```

---

## 8. Bridge View (`sim/bridge.ail`)

Bridge interior with dome viewport and crew.

```ailang
export type BridgeStation =
    | StationHelm | StationNav | StationComms | StationScience | StationEngineering

export type CrewActivity =
    | ActivityIdle | ActivityWorking | ActivityWalking | ActivityTalking

export type CrewPosition = {
    x: float, y: float,
    station: BridgeStation,
    activity: CrewActivity
}

export type DomeViewState = {
    planetId: int,
    velocity: float,      -- For SR warp
    viewAngle: float
}

export type DomeState = {
    velocity: float,
    progress: float,
    targetPlanet: int,
    viewState: DomeViewState
}

export type BridgeState = {
    dome: DomeState,
    crew: [CrewPosition],
    consoles: [ConsoleState],
    selectedInteractable: Option[InteractableID]
}

-- Dome operations
export pure func initDomeState() -> DomeState
export pure func stepDome(state: DomeState, dt: float) -> DomeState
export pure func getDomeVelocity(state: DomeState) -> float
export pure func getDomeProgress(state: DomeState) -> float

-- Rendering
export pure func renderDome(state: DomeState) -> [DrawCmd]
export pure func renderDomePlanets(state: DomeState) -> [DrawCmd]
```

**Demo:** `go run ./cmd/demo-game-bridge`

---

## 9. Depth/Parallax Layers (`sim/depth.ail`)

20-layer depth system for parallax effects.

```ailang
export type DepthLayer =
    | DeepSpace       -- Layer 0, parallax 0.0 (static)
    | FarStars        -- Layer 1, parallax 0.05
    | NearStars       -- Layer 2, parallax 0.1
    | ...
    | UIOverlay       -- Layer 19, parallax 1.0 (moves with camera)

-- Get parallax factor (0.0 = static, 1.0 = moves with camera)
export pure func layerParallax(layer: DepthLayer) -> float

-- Get Z-base for draw ordering
export pure func layerZBase(layer: DepthLayer) -> int

-- Transparent/glass tiles
export type TransparentTile = {
    baseId: int,
    alpha: float,
    seeThroughLayer: DepthLayer,
    tintRgba: Option[int]
}

export pure func glassFloorTile(baseId: int) -> TransparentTile
export pure func domeEdgeTile(baseId: int) -> TransparentTile
export pure func frostedPanelTile(baseId: int) -> TransparentTile
```

**Demo:** `go run ./cmd/demo-game-parallax`

---

## 10. Viewport System (`sim/viewport.ail`)

Shaped viewports for domes, portholes, windows.

```ailang
export type ViewportShape =
    | ShapeEllipse(centerX: float, centerY: float, radiusX: float, radiusY: float)
    | ShapeCircle(centerX: float, centerY: float, radius: float)
    | ShapeRect(x: float, y: float, width: float, height: float)
    | ShapeDome(centerX: float, centerY: float, width: float, height: float, archHeight: float)

export type ViewportContent =
    | ContentSpaceView(velocity: float, viewAngle: float)
    | ContentStarfield(density: float, scroll: bool)
    | ContentSolid(rgba: int)
    | ContentNone

export type ViewportEffect =
    | EffectNone
    | EffectSRWarp(velocity: float)
    | EffectGRLensing(mass: float, distance: float)
    | EffectTint(rgba: int, intensity: float)
    | EffectBlur(radius: float)

export type ViewportDef = {
    id: string,
    shape: ViewportShape,
    content: ViewportContent,
    effect: ViewportEffect,
    layer: int,
    edgeBlend: float,
    opacity: float
}

-- Factory functions
export pure func bridgeDome(centerX: float, centerY: float, width: float, height: float, velocity: float) -> ViewportDef
export pure func cabinWindow(x: float, y: float, w: float, h: float) -> ViewportDef
export pure func porthole(x: float, y: float, radius: float) -> ViewportDef

-- Convert to DrawCmd for engine
export pure func viewportToDrawCmd(vp: ViewportDef, z: int) -> DrawCmd
```

---

## 11. Ship Levels (`sim/ship_levels.ail`)

Multi-deck ship navigation.

```ailang
export type DeckType =
    | DeckBridge
    | DeckCrew
    | DeckCargo
    | DeckEngineering

export type DeckInfo = {
    name: string,
    description: string,
    spriteBase: int
}

export type TransitionState =
    | TransitionNone
    | TransitionUp(progress: float, target: DeckType)
    | TransitionDown(progress: float, target: DeckType)

export type ShipLevels = {
    currentDeck: DeckType,
    transition: TransitionState
}

-- Navigation
export pure func deck_above(deck: DeckType) -> DeckType
export pure func deck_below(deck: DeckType) -> DeckType
export pure func is_top_deck(deck: DeckType) -> bool
export pure func is_bottom_deck(deck: DeckType) -> bool

-- Transitions
export pure func init_ship_levels() -> ShipLevels
export pure func start_deck_transition(levels: ShipLevels, target: DeckType) -> ShipLevels
export pure func update_transition(levels: ShipLevels, dt: float) -> ShipLevels
export pure func is_transitioning(levels: ShipLevels) -> bool
```

---

## 12. Arrival Sequence (`sim/arrival.ail`)

Planet approach cinematics.

```ailang
export type ArrivalPhase =
    | PhaseApproach
    | PhaseDeceleration
    | PhaseOrbit
    | PhaseLanding
    | PhaseComplete

export type ArrivalState = {
    phase: ArrivalPhase,
    progress: float,
    velocity: float,
    targetPlanet: CurrentPlanet,
    grIntensity: float
}

export type ArrivalInput = {
    dt: float,
    skipRequested: bool
}

-- Arrival sequence
export pure func initArrival() -> ArrivalState
export pure func stepArrival(state: ArrivalState, input: ArrivalInput) -> ArrivalState
export pure func isArrivalComplete(state: ArrivalState) -> bool

-- Accessors for engine
export pure func getArrivalVelocity(state: ArrivalState) -> float
export pure func getArrivalPhaseName(state: ArrivalState) -> string
export pure func getGRIntensity(state: ArrivalState) -> float
```

---

## Related Documents

- [engine-capabilities.md](engine-capabilities.md) - Go engine features (what can be rendered)
- [ai-capabilities.md](ai-capabilities.md) - AI features (text, image, TTS)
- [demos.md](demos.md) - Demo index with run commands

---

**Document created**: 2025-12-12
**Last updated**: 2025-12-12
