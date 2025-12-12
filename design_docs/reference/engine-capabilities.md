# Engine Capabilities Reference

**Status**: Active (Updated 2025-12-12)
**Purpose**: Comprehensive reference for all Go/Ebiten engine capabilities
**Audience**: Sprint executor, design docs, AI agents working on the game

## Quick Reference

| Capability | Location | Status |
|------------|----------|--------|
| DrawCmd rendering | `engine/render/` | Working |
| Effect handlers | `engine/handlers/` | Working |
| Asset loading | `engine/assets/` | Working |
| Camera/viewport | `engine/camera/` | Working |
| **Parallax layers** | `engine/depth/` | **Working** |
| **Celestial system** | `sim/celestial.ail` | **Working** |
| SR/GR physics | `engine/relativity/` | Working |
| Shader effects | `engine/shader/` | Working |
| Input capture | `engine/input/` | Working |
| Display config | `engine/display/` | Working |
| Save system | `engine/save/` | Working |
| Screenshot/test | `engine/screenshot/` | Working |

---

## 1. DrawCmd Types (AILANG → Rendering)

All defined in `sim/protocol.ail`, rendered by `engine/render/draw.go`.

### Basic Commands

| Command | Signature | Purpose | Space |
|---------|-----------|---------|-------|
| `Sprite` | `(id, x, y, z)` | Draw sprite at world position | World |
| `Rect` | `(x, y, w, h, color, z)` | Solid rectangle | World |
| `RectScreen` | `(x, y, w, h, color, z)` | Rectangle in screen space | Screen |
| `Text` | `(text, x, y, fontSize, color, z)` | Draw text | Screen |
| `TextWrapped` | `(text, x, y, maxWidth, fontSize, color, z)` | Word-wrapped text | Screen |
| `Line` | `(x1, y1, x2, y2, color, width, z)` | Line with width | Screen |
| `Circle` | `(x, y, radius, color, filled, z)` | Circle (filled or outline) | Screen |

### Isometric Commands

| Command | Signature | Purpose |
|---------|-----------|---------|
| `IsoTile` | `(tile, height, spriteId, layer, color)` | Isometric ground tile |
| `IsoEntity` | `(id, tile, offsetX, offsetY, height, spriteId, layer)` | Entity with sub-tile positioning |

**Isometric Constants:**
- `TileWidth = 64.0` world units
- `TileHeight = 32.0` world units (width/2)
- `HeightScale = 16.0` units per height level

### Starmap Commands

| Command | Signature | Purpose |
|---------|-----------|---------|
| `GalaxyBg` | `(opacity, z, skyViewMode, viewLon, viewLat, fov)` | Galaxy background with scrolling |
| `Star` | `(x, y, spriteId, scale, alpha, z)` | Individual star with scale/alpha |
| `SpireBg` | `(z)` | Spire silhouette background (Layer 6, 0.3x parallax) |
| `CircleRGBA` | `(x, y, radius, rgba, filled, z)` | Circle with packed RGBA color |
| `RectRGBA` | `(x, y, w, h, rgba, z)` | Rectangle with packed RGBA color |

**RGBA Color Format:** `0xRRGGBBAA` (e.g., `0xFF0000FF` = opaque red)

### Celestial System (AILANG)

Planet and star system simulation defined in `sim/celestial.ail`:

| Function | Signature | Purpose |
|----------|-----------|---------|
| `initSolSystem` | `() -> StarSystem` | Create Sol with 8 planets |
| `stepSystem` | `(system, dt) -> StarSystem` | Update orbital positions |
| `renderSolarSystem` | `(system) -> [DrawCmd]` | Render planets as CircleRGBA |

**Planet Types:** `Rocky`, `GasGiant`, `IceGiant`, `Terrestrial`, `Ocean`, `Volcanic`, `Dwarf`

**Star Types:** Spectral classes `O`, `B`, `A`, `F`, `G`, `K`, `M`

### Parallax Layer Commands

| Command | Signature | Purpose |
|---------|-----------|---------|
| `Marker` | `(x, y, w, h, rgba, parallaxLayer, z)` | Rectangle on selectable parallax layer (0-19) |

**Marker Usage:** Place visual elements on any of the 20 depth layers. AILANG controls which layer via `parallaxLayer` field.

### UI Commands

| Command | Signature | Purpose |
|---------|-----------|---------|
| `Ui` | `(id, kind, x, y, w, h, text, spriteId, z, color, value)` | UI element |

**UiKind Variants:**
- `UiPanel` - Background container
- `UiButton` - Clickable button with border
- `UiLabel` - Text label
- `UiPortrait` - Sprite display (scaled)
- `UiSlider` - Value slider (0.0-1.0)
- `UiProgressBar` - Progress indicator

**UI Coordinates:** Normalized (0.0-1.0), scaled to screen pixels.

### Color Palette

Colors are indexed 0-15 via `biomeColors`:
```
0: Water blue    4: Savanna      8: Mountain     12: NPC cyan
1: Desert tan    5: Forest       9: Snow         13: Player green
2: Grassland     6: Rainforest   10: Structure   14: Highlight
3: Tundra        7: Marsh        11: Road        15: UI background
```

---

## 2. Effect Handlers (AILANG → Go)

Defined in `sim_gen/handlers.go`, implemented in `engine/handlers/`.

**CRITICAL:** Call `sim_gen.Init(handlers)` BEFORE any AILANG code runs.

### Debug Effect

**Interface:** `DebugHandler`
**Implementation:** `sim_gen.NewDebugContext()` (built-in)

```go
type DebugHandler interface {
    Log(msg, location string)
    Assert(cond bool, msg, location string)
    SetTimestamp(t int64)
    Collect() DebugOutput  // Host-only
    Reset()                // Host-only
}
```

**AILANG Usage:**
```ailang
import std/debug (log, check)

func example() -> () ! {Debug} {
    Debug.log("message here");
    Debug.check(x > 0, "x must be positive")
}
```

### Rand Effect

**Interface:** `RandHandler`
**Implementations:** `engine/handlers/rand.go`
- `NewDefaultRandHandler()` - Time-seeded
- `NewSeededRandHandler(seed)` - Deterministic

```go
type RandHandler interface {
    RandInt(min, max int64) int64
    RandFloat(min, max float64) float64
    RandBool() bool
    SetSeed(seed int64)
}
```

**AILANG Usage:**
```ailang
import std/rand (rand_int, rand_float, rand_bool, rand_seed)

func example() -> int ! {Rand} {
    rand_seed(42);
    rand_int(1, 100)
}
```

### Clock Effect

**Interface:** `ClockHandler`
**Implementation:** `engine/handlers/clock.go` - `EbitenClockHandler`

```go
type ClockHandler interface {
    DeltaTime() float64   // Seconds since last frame
    TotalTime() float64   // Total game time
    FrameCount() int64    // Current frame number
}
```

**Host Responsibility:** Call `clockHandler.Update(dt)` each frame BEFORE `sim_gen.Step()`.

**AILANG Usage:**
```ailang
import std/game (delta_time, frame_count, total_time)

func smooth_move(pos: float, target: float) -> float ! {Clock} {
    let dt = delta_time();
    pos + (target - pos) * dt * 5.0
}
```

### AI Effect

**Interface:** `AIHandler`
**Implementations:** `engine/handlers/`
- `ai.go` - `StubAIHandler` (testing)
- `ai_claude.go` - Claude API
- `ai_gemini.go` - Gemini API (multimodal)
- `ai_factory.go` - Auto-detection

```go
type AIHandler interface {
    Call(input string) (string, error)
}
```

**Request Format (JSON):**
```json
{
  "messages": [
    {"type": "text", "text": "prompt here"},
    {"type": "image", "image_ref": "path/to/image.png"}
  ],
  "context": {"game_state": "..."},
  "system": "You are an alien civilization..."
}
```

**Response Format (JSON):**
```json
{
  "content": [{"type": "text", "text": "response"}],
  "error": ""
}
```

**Provider Detection (auto):**
1. `GOOGLE_CLOUD_PROJECT` → Vertex AI (Gemini)
2. `GOOGLE_API_KEY` → Gemini API
3. `ANTHROPIC_API_KEY` → Claude
4. Fallback → Stub

**Gemini Multimodal Features:**
- Image generation: `"generate image"`, `"draw"`, `"imagen:"`
- Image editing: `"edit:"`, `"modify:"` + reference image
- Text-to-speech: `"speak:"`, `"say:"`, `"tts:"`
- Voices: Aoede, Charon, Fenrir, Kore, Puck, Zephyr

**AILANG Usage:**
```ailang
import std/ai (ai_call)

func ask_archive(question: string) -> string ! {AI} {
    let input = encodeJson({"messages": [{"type": "text", "text": question}]});
    let response = ai_call(input);
    decodeJson(response).content[0].text
}
```

### Optional Effects (Not Yet Needed)

| Effect | Interface | Purpose |
|--------|-----------|---------|
| FS | `FSHandler` | File read/write |
| Net | `NetHandler` | HTTP requests |
| Env | `EnvHandler` | Environment variables |

---

## 3. Asset Systems

All in `engine/assets/`, unified via `AssetManager`.

### Sprite Manager

**File:** `sprites.go`

```go
type SpriteManager interface {
    LoadManifest(path string) error
    Get(id int) *ebiten.Image
    Has(id int) bool
    GetAnimation(id int) *AnimationDef
    HasAnimation(id int) bool
}
```

**Manifest Format:** `assets/sprites/manifest.json`
```json
{
  "sprites": {
    "100": {
      "file": "player.png",
      "width": 32, "height": 48,
      "type": "entity",
      "animations": {
        "idle": {"startFrame": 0, "frameCount": 4, "fps": 6.0},
        "walk": {"startFrame": 4, "frameCount": 8, "fps": 12.0}
      },
      "frameWidth": 32, "frameHeight": 48
    }
  }
}
```

### Audio Manager

**File:** `audio.go`

```go
type AudioManager interface {
    LoadManifest(path string) error
    PlaySound(id int)
    PlaySoundWithVolume(id int, vol float64)
    StopSound(id int)
    PlayMusic(id int)
    StopMusic()
    SetMusicVolume(vol float64)
    IsMusicPlaying() bool
}
```

**Manifest Format:** `assets/sounds/manifest.json`
```json
{
  "sounds": {"100": {"file": "click.ogg", "volume": 1.0}},
  "bgm": {"50": {"file": "theme.ogg", "loop": true, "volume": 0.7}}
}
```

**Formats:** OGG Vorbis, WAV (44100 Hz)

### Font Manager

**File:** `fonts.go`

```go
type FontManager interface {
    LoadManifest(path string) error
    Get(name string) font.Face
    GetDefault() font.Face
    GetBySize(sizeIndex int) font.Face  // 0=small, 1=normal, 2=large, 3=title
    SetScale(screenHeight int)
}
```

**Standard Sizes (at 720p):**
- 0 (Small): 16pt
- 1 (Normal): 22pt
- 2 (Large): 28pt
- 3 (Title): 38pt

**Fallback:** Embedded Go Mono (monospace sci-fi aesthetic)

---

## 4. Camera & Display

### Camera Transform

**File:** `engine/camera/transform.go`

**AILANG Camera Type:**
```ailang
type Camera = { x: float, y: float, zoom: float }
```

**Go Transform:**
```go
type Transform struct {
    CenterX, CenterY float64  // World position
    Zoom             float64  // Scale factor
    ScreenW, ScreenH int      // Screen dimensions
}

func (t *Transform) WorldToScreen(wx, wy float64) (sx, sy float64)
func (t *Transform) ScreenToWorld(sx, sy float64) (wx, wy float64)
```

### Viewport Culling

**File:** `engine/camera/viewport.go`

```go
type Viewport struct {
    MinX, MinY, MaxX, MaxY float64  // World bounds
}

func (v *Viewport) Contains(x, y, margin float64) bool
func (v *Viewport) ContainsRect(x, y, w, h float64) bool
```

### Display Configuration

**File:** `engine/display/`

```go
type Config struct {
    Width, Height int
    Fullscreen    bool
    VSync         bool
    Scale         float64  // 0.5-4.0
}
```

**Key Methods:**
- `DefaultConfig()` → 1280×720, VSync on
- `ToggleFullscreen()` → F11 support
- `SetResolution(w, h)` → Resize window

**Internal Resolution:** 1280×960 (fixed game coords, Ebiten scales)

---

## 5. Parallax Depth Layers

**File:** `engine/depth/layer.go`

The engine supports **20 depth layers (0-19)** for parallax rendering. Each layer has a configurable parallax factor that determines how fast it moves relative to the camera.

### Layer Definitions

| Layer | Parallax | Purpose |
|-------|----------|---------|
| L0 | 0.00 | Fixed at infinity (galaxy/space) |
| L1 | 0.05 | Very distant stars |
| L2 | 0.10 | Far spire segment |
| L3 | 0.15 | Distant ship structure |
| L4 | 0.20 | Opposite hull |
| L5 | 0.25 | Far deck (5+ decks away) |
| L6 | 0.30 | Mid-distance structure |
| L7 | 0.40 | 4 decks away |
| L8 | 0.50 | 3 decks away |
| L9 | 0.60 | 2 decks away |
| L10 | 0.70 | Adjacent deck |
| L11 | 0.75 | Near background |
| L12 | 0.80 | Same deck distant |
| L13 | 0.85 | Same deck mid |
| L14 | 0.90 | Same deck near |
| L15 | 0.95 | Current deck background |
| L16 | 1.00 | Main scene layer |
| L17 | 1.00 | Scene overlay |
| L18 | 1.00 | Foreground effects |
| L19 | 1.00 | UI (screen-fixed) |

### Convenience Aliases

```go
LayerDeepBackground = Layer0   // Galaxy, fixed at infinity
LayerMidBackground  = Layer6   // Mid-distance, 0.3x
LayerScene          = Layer16  // Main content, 1.0x
LayerForeground     = Layer19  // UI, screen-fixed
```

### Parallax Math

```
parallax_offset = camera_position × parallax_factor × zoom
```

- **0.0x**: Fixed (doesn't move with camera) - stars at infinity
- **0.5x**: Moves half as fast as camera - distant objects
- **1.0x**: Moves with camera - scene layer

### Runtime Configuration

```go
// Change parallax factor at runtime (e.g., for zoom-dependent effects)
depth.SetParallax(depth.Layer5, 0.35)

// Get all current factors
factors := depth.GetAllParallax()
```

### AILANG Integration

DrawCmds are routed to layers by type OR by explicit `parallaxLayer` field:

| DrawCmd | Default Layer |
|---------|---------------|
| `GalaxyBg`, `SpaceBg`, `Star` | Layer0 (0.0x) |
| `SpireBg` | Layer6 (0.3x) |
| `IsoTile`, `IsoEntity`, `Sprite` | Layer16 (1.0x) |
| `Ui`, `Text`, `RectScreen` | Layer19 (UI) |
| `Marker(parallaxLayer=N)` | LayerN (selectable) |

### Enable Layer Rendering

```go
renderer.EnableLayers(screenW, screenH)  // Enable layer system
renderer.ResizeLayers(screenW, screenH)  // Handle resize
```

### Demo

```bash
go run ./cmd/demo-parallax                    # Interactive demo
go run ./cmd/demo-parallax -camx 400 --screenshot 5 --output test.png
```

---

## 6. Relativity Physics

### Special Relativity (SR)

**File:** `engine/relativity/transform.go`

**Core Functions:**
```go
// Lorentz factor: γ = 1/√(1-β²), clamped to 100
func Gamma(beta float64) float64

// Doppler factor: D = γ(1 - β·n), clamped [0.01, 100]
func DopplerFactor(beta Vec3, viewDir Vec3, gamma float64) float64

// Relativistic aberration
func AberrationAngle(beta, angle, gamma float64) float64
```

**Color Shifting:** `engine/relativity/color.go`
```go
// Blackbody temperature → RGB (1000K-40000K)
func TemperatureToRGB(kelvin float64) color.RGBA

// Shift color by Doppler factor
func DopplerShiftColor(baseRGB color.RGBA, dopplerFactor float64) color.RGBA

// Relativistic beaming brightness: D³
func BeamBrightness(dopplerFactor float64) float64

// Complete star processing
func ProcessStar(baseTemp float64, direction, velocity Vec3, gamma float64) (color.RGBA, float64)
```

**Spectral Classes:**
| Class | Temperature | Color |
|-------|-------------|-------|
| O | 30000K | Blue |
| B | 20000K | Blue-white |
| A | 10000K | White |
| F | 7500K | Yellow-white |
| G (Sun) | 5800K | Yellow |
| K | 4500K | Orange |
| M | 3000K | Red |

### General Relativity (GR)

**File:** `engine/relativity/gr_context.go`

**Massive Object Types:**
```go
const (
    BlackHole = iota
    NeutronStar
    WhiteDwarf
)
```

**GR Context:**
```go
type GRContext struct {
    Active           bool
    ObjectKind       MassiveObjectKind
    Distance         float64  // Ship → object
    Phi              float64  // Dimensionless potential: r_s/(2r)
    TimeDilation     float64  // dτ/dt = √(1 - r_s/r)
    RedshiftFactor   float64  // z = 1/√(1 - r_s/r)
    TidalSeverity    float64  // 0.0-1.0
    DangerLevel      GRDangerLevel
    CanHoverSafely   bool
    NearPhotonSphere bool
    Rs               float64  // Schwarzschild radius
}
```

**Danger Levels (based on Φ = r_s/2r):**
| Level | Φ Range | Visual Effects |
|-------|---------|----------------|
| None | Φ < 1e-4 | SR only |
| Subtle | 1e-4 ≤ Φ < 1e-3 | Light shading |
| Strong | 1e-3 ≤ Φ < 1e-2 | Visible lensing |
| Extreme | Φ ≥ 0.01 | Heavy distortion |

**Key Formulas:**
```
Schwarzschild radius: r_s = 2GM/c² ≈ 2.95 km per M_sun
Time dilation: dτ/dt = √(1 - r_s/r)
Gravitational redshift: z = 1/√(1 - r_s/r)
Photon sphere: r = 1.5 r_s (black holes only)
```

---

## 7. Shader Effects

**File:** `engine/shader/`

### Effects Manager

```go
type Effects struct {
    // Access methods
    Manager() *Manager
    Pipeline() *Pipeline
    Bloom() *Bloom
    SRWarp() *SRWarp
    GRWarp() *GRWarp
}
```

### Post-Processing Pipeline

**Built-in Effects:**
| Effect | Uniforms | Purpose |
|--------|----------|---------|
| `vignette` | Intensity, Softness | Edge darkening |
| `crt` | ScanlineIntensity, Curvature, VignetteAmount | Retro CRT look |
| `aberration` | Amount | RGB channel separation |

```go
pipeline.SetEnabled("vignette", true)
pipeline.SetUniform("vignette", "Intensity", 0.5)
```

### Bloom (Glow)

```go
bloom.SetThreshold(0.7)   // Brightness threshold (0.0-1.0)
bloom.SetIntensity(1.2)   // Glow intensity (0.0-2.0)
bloom.SetBlurPasses(3)    // Blur quality (1-5)
bloom.SetEnabled(true)
```

### SR Warp Effect

```go
srWarp.SetVelocity(0.0, 0.0, 0.9)  // 90% light speed forward
srWarp.SetForwardVelocity(0.9)     // Convenience
srWarp.SetFOV(1.57)                // ~90° FOV
srWarp.SetViewAngle(0.0)           // Looking forward
srWarp.SetEnabled(true)
```

**Applied Effects:**
- Aberration (direction warping)
- Doppler shift (color change)
- Relativistic beaming (brightness)

### GR Warp Effect

```go
grWarp.SetUniforms(relativity.GRShaderUniforms{...})
grWarp.SetDemoMode(0.5, 0.5, 0.05, 0.05)  // centerX, Y, rs, phi
grWarp.CycleDemoIntensity()  // Subtle → Strong → Extreme
grWarp.SetEnabled(true)
```

**Applied Effects:**
- Gravitational lensing
- Chromatic aberration
- Redshift coloring

**Keyboard Shortcuts:**
- F7: Toggle GR effects
- F8: Cycle GR intensity (demo mode)

---

## 8. Input System

**File:** `engine/input/`

### Input Capture

```go
func CaptureInputWithCamera(cam Transform, w, h int) FrameInput
```

**FrameInput Fields (from AILANG):**
```ailang
type FrameInput = {
    mouseX: float,           -- Screen position
    mouseY: float,
    worldMouseX: float,      -- World position (via camera)
    worldMouseY: float,
    tileMouseX: int,         -- Isometric tile
    tileMouseY: int,
    clickedThisFrame: bool,  -- Left button just pressed
    rightClickedThisFrame: bool,
    keys: [KeyEvent],
    actionRequested: PlayerAction
}
```

### Key Detection

```go
func IsKeyPressed(key ebiten.Key) bool
func IsKeyJustPressed(key ebiten.Key) bool
```

**PlayerAction (from keys I/B/X):**
```ailang
type PlayerAction =
    | ActionNone
    | ActionInspect
    | ActionBuild(StructureType)
    | ActionClear
```

---

## 9. Save System

**File:** `engine/save/save.go`

**Design:** Single save file (no slots) per Pillar 1 "Choices Are Final"

```go
type SaveManager interface {
    SaveGame(world interface{}) error
    LoadGame() (interface{}, error)
    HasSave() bool
    DeleteSave() error
}
```

**Save File:** `saves/game.json`
```json
{
  "version": "0.1.0",
  "timestamp": 1701864000,
  "playTime": 3600,
  "world": {...}
}
```

**Auto-save:** 5 minutes (configurable)

---

## 10. Screenshot & Testing

### Screenshot Capture

**File:** `engine/screenshot/screenshot.go`

```go
type Config struct {
    Frames     int      // Frames before capture
    OutputPath string   // PNG path
    Seed       int64    // World seed
    CameraX    float64  // Camera position
    CameraY    float64
    CameraZoom float64
    TestMode   bool     // Strip UI for golden files
    Effects    string   // Comma-separated effect names
    DemoScene  bool     // Shader demo mode
    Velocity   float64  // Ship velocity (0.0-0.99c)
    ViewAngle  float64  // View direction (radians)
}

func Capture(cfg Config) (*image.RGBA, error)
func CaptureToFile(cfg Config) error
```

### Scenario Testing

**File:** `engine/scenario/`

```json
{
  "name": "exploration-test",
  "seed": 1234,
  "events": [
    {"frame": 0, "key": "W", "action": "down"},
    {"frame": 30, "click": {"x": 640, "y": 360, "button": "left"}},
    {"frame": 60, "capture": "screenshot.png"}
  ]
}
```

---

## 11. Coordinate Systems

| System | Range | Transform | Use |
|--------|-------|-----------|-----|
| World | unbounded float | Camera zoom/pan | Game logic |
| Screen | 0 to width/height pixels | Direct | Mouse, UI |
| Tile | integer grid | Isometric projection | Grid navigation |
| Iso World | tile + offset + height | TileToWorld | 3D entities |
| Normalized UI | 0.0-1.0 | Scale to screen | Resolution-independent UI |

---

## 12. Initialization Sequence

**Order matters:**

```go
// 1. Asset Manager
assetMgr, _ := assets.NewManager("assets")

// 2. Display Manager
displayMgr := display.NewManager("config/display.json")

// 3. Effect Handlers
clockHandler := handlers.NewEbitenClockHandler()
randHandler := handlers.NewSeededRandHandler(seed)
aiHandler, _ := handlers.NewAIHandlerFromEnv(ctx)

// 4. CRITICAL: Initialize sim_gen BEFORE Step()
sim_gen.Init(sim_gen.Handlers{
    Debug: sim_gen.NewDebugContext(),
    Rand:  randHandler,
    Clock: clockHandler,
    AI:    aiHandler,
})

// 5. Renderer
renderer := render.NewRenderer(assetMgr)

// 6. Shader Effects
effects := shader.NewEffects()

// 7. Now safe to call sim_gen.InitWorld() and Step()
```

---

## 13. Z-Ordering & Depth

### Z-Value Ranges

| Range | Purpose |
|-------|---------|
| < 0 | Below terrain |
| 0-999 | Game world |
| 1000-9999 | UI overlay |
| 10000+ | Reserved |

### Isometric Sort Key

```
sortKey = layer × 10000 + screenY
```

- Layer dominates (higher = on top)
- Within layer: screenY (further down = behind in iso view)
- UI always renders last (layer 1000+)

---

**Document created**: 2025-12-06
**Last updated**: 2025-12-12
