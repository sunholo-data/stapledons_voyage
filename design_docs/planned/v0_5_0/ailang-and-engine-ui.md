Here’s a focused design doc that answers “who owns isometric?” and “what do we need?” for AILANG vs the Go engine.

⸻

STAPLEDON’S VOYAGE – ISOMETRIC UI ARCHITECTURE

0. Design Decision (TL;DR)
	•	AILANG owns:
	•	All logical state: ship layout, tiles, entities, UI modes, camera intent.
	•	Everything in tile/world coordinates, not pixels.
	•	High-level draw descriptions: “draw this tile/entity/UI element at (tx, ty, h, layer)”.
	•	Engine (Go/Ebiten) owns:
	•	All projection & rendering math (isometric transform).
	•	Asset loading (spritesheets, fonts).
	•	Pixel coordinates, z-sorting, draw batching.
	•	Hit tests: screen → tile coords.

Rule: AILANG never thinks in pixels or trig; the engine never owns game rules.

⸻

1. Coordinate Systems

We standardise four spaces:
	1.	Tile space (AILANG)
	•	(tileX: int, tileY: int, height: int)
	•	Used for ship interior, planetary surfaces, pathfinding.
	2.	World/entity space (AILANG)
	•	Entity positions are either exact tile cells or subcell offsets:
EntityPos = { tile: Coord, offset_x: float, offset_y: float }.
	3.	Camera space (AILANG + engine)
	•	AILANG stores camera center in tile space.
	•	Engine uses this + viewport size to decide what’s visible.
	4.	Screen space (engine only)
	•	(px: float, py: float) in pixels.
	•	Computed via isometric projection + camera transform.

Isometric projection (engine side):

screenX = (tileX - tileY) * (tileWidth / 2)
screenY = (tileX + tileY) * (tileHeight / 2) - height * heightScale


⸻

2. Shared Protocol Types (AILANG → Engine)

We extend the existing FrameInput / FrameOutput idea with iso-aware semantics, but still logical.

2.1 UI Mode & Camera

AILANG defines which surface we’re on and which camera to use:

module game/ui

type UiMode =
  | ShipMode
  | PlanetMode
  | GalaxyMapMode
  | DialogueMode
  | LegacyMode

type Camera = {
    center_tile: Coord,   -- tile space
    zoom: float,          -- 1.0 = base, engine decides pixel scale
}

type ViewState = {
    mode: UiMode,
    camera: Camera,
}

ViewState lives inside World.

2.2 Iso Draw Commands (Logical)

We keep DrawCmd, but make “iso-ness” explicit via variants that carry tile coords:

type IsoTile = {
    tile: Coord,        -- {x:int, y:int}
    height: int,        -- tile height level
    sprite_id: SpriteId,
    layer: int,         -- coarse draw order hint (0=bg, 100=entities, 200=UI)
}

type IsoEntity = {
    id: EntityId,
    tile: Coord,
    offset_x: float,
    offset_y: float,
    height: int,
    sprite_id: SpriteId,
    layer: int,
}

type UiElement = {
    kind: UiKind,      -- Button, Panel, Label, Portrait etc.
    rect: UiRect,      -- logical units (0..1 for relative, or “UI grid” units)
    text: string,      -- if applicable
    sprite_id: SpriteId, -- for icons/portraits
    z: int,
}

type DrawCmd =
  | DrawIsoTile IsoTile
  | DrawIsoEntity IsoEntity
  | DrawUi UiElement
  | DrawRect { x: float, y: float, w: float, h: float, color: int, z: int }
  | DrawText { text: string, x: float, y: float, z: int }

No pixels here. Only tile coords, layers, and logical UI rects.

2.3 FrameOutput

type FrameOutput = {
    draw: List<DrawCmd>,
    sounds: List<SoundCmd>,
    debug: DebugOutput,
}

Engine takes this and does:
	1.	For DrawIsoTile / DrawIsoEntity:
	•	Convert tile coords → screen coords via iso projection & camera.
	2.	For DrawUi:
	•	Use UI layout system in screen space.
	3.	For Rect / Text:
	•	Either treat them as UI-space (normalized) or raw pixel-layer overlays.

⸻

3. Engine Responsibilities (Go/Ebiten)

3.1 Isometric Projection

Implement a renderer module, e.g. engine/render/iso.go, responsible for:
	•	tileToScreen(tile: Coord, height: int, cam: Camera) -> (px, py)
	•	screenToTile(px, py, cam) -> (tileX, tileY, heightApprox) for click handling.

Decision: projection parameters (tile width/height, height scale) are engine constants or config, not AILANG fields.

3.2 Layers and Sorting

Engine uses:
	•	layer (coarse order) from IsoTile / IsoEntity / UiElement.
	•	screenY as fine-grain depth within the layer (so entities closer to the “bottom” overlap correctly).

Sorting strategy:

sort by (layer, screenY, maybe entity.id)

AILANG just provides layer hints; the actual sort is engine-side.

3.3 Hit Testing (Mouse → Tile / UI)

When the user clicks:
	1.	Engine gets (mouseX, mouseY) in pixels.
	2.	If UI overlay captures click (buttons, panels), handle there.
	3.	Else, call screenToTile(...) to get (tileX, tileY); send to AILANG in FrameInput.

Example AILANG-facing input:

type ClickKind = Left | Right | Middle

type TileClick = {
    tile: Coord,
    click: ClickKind,
}

type UiClick = {
    element_id: UiElementId,
    click: ClickKind,
}

type FrameInput = {
    tile_clicks: List<TileClick>,
    ui_clicks: List<UiClick>,
    keys: List<KeyEvent>,
    mode: UiMode,
}

Engine resolves ids / tiles; AILANG only gets semantic events.

⸻

4. AILANG Responsibilities

4.1 World State & Modes

AILANG World tracks:
	•	current_view: ViewState
	•	ship_state: ShipState (rooms, crew positions → tile coords)
	•	planet_state: PlanetState (surface tiles)
	•	galaxy_state (nodes, edges, etc., drawn as 2D map, non-iso)
	•	dialogue_state (active speaker, choices, etc.)

4.2 Camera Logic

AILANG decides:
	•	Which mode is active (ShipMode, PlanetMode, etc.).
	•	Where camera center_tile should be (e.g. follow player avatar, center on clicked room, etc.).
	•	Zoom level changes (discrete steps: 0.75x / 1x / 1.5x).

Engine uses this camera but doesn’t change it silently; any camera movement is triggered via FrameInput (e.g. scroll wheel) and applied by AILANG in step.

4.3 Draw Command Emission (Logical Layout)

For each mode, AILANG builds draw:

ShipMode:
	•	For each visible ship tile: emit DrawIsoTile.
	•	For each crew member: emit DrawIsoEntity.
	•	For selection / highlights: extra DrawIsoEntity or overlay DrawRect.
	•	For HUD: DrawUi elements (minimap, crew list, etc.).

PlanetMode:
	•	Similar to ShipMode, but for planet tiles / entities.

GalaxyMapMode:
	•	Likely non-iso:
	•	DrawRect/DrawSprite for stars in a 2D plane.
	•	DrawUi for filters, overlays.

DialogueMode:
	•	Mostly DrawUi:
	•	Portrait, text box, choices, crew stats.

LegacyMode:
	•	Graphs / timelines done via DrawUi + DrawRect/DrawText.

AILANG never needs to know “how big is a tile in pixels”; it just emits logical elements.

⸻

5. Minimal Implementation Plan

Step 1 – Protocol & Types

In AILANG repo:
	•	module game/ui with:
	•	UiMode, Camera, ViewState
	•	IsoTile, IsoEntity, UiElement, DrawCmd
	•	Extend FrameInput / FrameOutput signature for step.

Step 2 – Engine Iso Renderer

In game repo:
	•	engine/render/iso.go:
	•	tileToScreen, screenToTile
	•	renderIsoTile, renderIsoEntity
	•	Sort pipeline (layer, screenY).
	•	engine/render/ui.go:
	•	Layout UiElement rects into pixels.
	•	Input hit testing (UI first, then iso).

Step 3 – MVP: Ship Interior Only
	•	Hard-code a small ship map in AILANG (e.g. 10×10).
	•	One controllable avatar + 2–3 standing crew.
	•	Arrow keys/WASD move avatar (tile coord changes).
	•	Camera follows avatar center.
	•	Engine renders isometric ship map & characters.

No planets, no dialogues, no galaxy map yet. Just verify:
	•	AILANG emits correct tile coords.
	•	Engine transforms to isometric and back.
	•	Click → tile → move/inspect works.

Step 4 – Add Modes

Once ship iso works:
	•	Add UiMode switching:
	•	Tab: ShipMode ↔ GalaxyMapMode.
	•	G: Galaxy map only, drawn via non-iso commands.
	•	Later, PlanetMode and DialogueMode reuse the same infrastructure.

⸻

6. Why Projection Belongs in the Engine
	•	Projection math is purely presentational; changing from iso → orthographic shouldn’t touch sim code.
	•	Engine may change resolution, DPI scaling, zoom gestures — AILANG should not.
	•	Asset decisions (tile size, art style) are renderer details.
	•	AILANG’s job is to be a pure state machine over World, not a view engine.

So the architecture is:

AILANG: “Tile (3,5), crew there, layer 100.”
Engine: “Great, that’s (512, 320) on screen, under this roof sprite, above that floor.”

