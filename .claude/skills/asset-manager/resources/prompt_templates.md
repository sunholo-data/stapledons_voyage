# AI Image Generation Prompt Templates

Tested prompts for generating consistent game assets.

## Core Style Phrases

**Always include these for consistency:**

```
pixel art style, retro 16-bit aesthetic, limited color palette,
crisp pixels, no anti-aliasing, clear dark outlines
```

**For isometric assets add:**
```
isometric view, 2:1 diamond ratio, top-down angled perspective
```

**For sci-fi theme add:**
```
hard science fiction, cosmic, deep space, NASA-inspired realism
```

## Isometric Tile Templates

### Basic Terrain Tile (CRITICAL REQUIREMENTS)
```
Create a 64x32 pixel isometric floor tile for [TERRAIN TYPE].

CRITICAL TESSELLATION REQUIREMENTS (ALL MUST BE MET):
1. EXACT dimensions: 64 pixels wide, 32 pixels tall
2. Diamond shape ONLY - NO rectangular background
3. PNG format with full transparency (alpha channel)
4. Content fills the EXACT isometric diamond shape
5. Diamond vertices at: top(32,0), right(63,16), bottom(32,31), left(0,16)
6. Corners at (0,0), (63,0), (0,31), (63,31) MUST be fully transparent (alpha=0)
7. Approximately 1024 opaque pixels (the diamond area is half of 64x32)

Style: Retro 16-bit pixel art, limited palette (8-12 colors),
crisp pixels, subtle top-left lighting.

WRONG (will not tessellate):
- Rectangle with diamond drawn inside it
- Any opaque pixels at the four corners
- Diamond that doesn't fill edge to edge

RIGHT:
- Pure diamond shape with transparent corners
- Content extends to diamond vertices
- Clean edges along the diamond boundary

[ADDITIONAL DETAILS]
```

**Examples:**
- `[TERRAIN TYPE]`: metallic ship floor, alien crystal formations, futuristic deck plating
- `[ADDITIONAL DETAILS]`: purple glow accents, panel lines, sci-fi aesthetic

### Water/Liquid Tile
```
Create a 64x32 pixel art isometric water tile.

Style: Retro pixel art, limited blue palette, subtle wave pattern.
Must show: Transparent/translucent water effect, light reflections.
Shape: Diamond isometric tile with transparent background.
Animation consideration: Design as single frame but suitable for
slight palette cycling animation.
```

### Vegetation Tile
```
Create a 64x32 pixel art isometric tile showing [PLANT TYPE].

Style: 16-bit pixel art, greens and earth tones, organic shapes.
Details: [SPECIFIC PLANTS], varied heights, natural clustering.
Shape: Isometric diamond tile, transparent background.
Lighting: Top-left light source creating subtle shadows.
```

## Entity Sprite Templates

### Humanoid Character (4-frame sheet)
```
Create a 128x48 pixel art sprite sheet of a [CHARACTER DESCRIPTION].

Layout: 4 frames side-by-side, each 32x48 pixels.
Frame 1: Standing idle pose
Frame 2-4: Walking animation cycle

Style: Retro 16-bit pixel art, limited palette, clear dark outlines.
Perspective: Isometric-compatible (slight 3/4 view from above).
Background: Transparent.

Character details: [SPECIFIC APPEARANCE]
```

**Example character descriptions:**
- "human ship crew member in blue jumpsuit"
- "four-armed alien merchant in flowing robes"
- "maintenance robot with tool arms"

### Alien Entity
```
Create a 128x48 pixel art sprite sheet of an alien creature.

Design: [ALIEN CONCEPT - body type, limbs, features]
Must convey: Non-human biology, otherworldly but readable silhouette.

Layout: 4 frames (32x48 each) for walk/movement cycle.
Style: Pixel art, limited colors, clear outlines.
Background: Transparent.
```

### Static Object/Item
```
Create a 32x48 pixel art sprite of [OBJECT].

Style: Isometric-compatible pixel art, clear outlines.
Details: [SPECIFIC FEATURES]
Background: Transparent.
Scale: Should fit on a 64x32 isometric tile.
```

## Star Sprite Templates

### Standard Star
```
Create a 16x16 pixel art star sprite.

Color: [SPECTRAL COLOR - see below]
Style: Soft glow effect, brightest at center, fading to edges.
Background: Transparent.
Shape: Roughly circular with subtle rays/twinkle.
```

**Spectral colors:**
- Class O/B (hot): Blue (#9bb0ff to #aabfff)
- Class A/F (medium): White to pale yellow (#cad7ff to #f8f7ff)
- Class G (Sun-like): Yellow (#fff4ea)
- Class K (cool): Orange (#ffd2a1)
- Class M (red dwarf): Red-orange (#ffcc6f)

### Giant Star
```
Create a 24x24 pixel art giant star sprite.

Type: [Red giant / Blue giant]
Style: Larger glow radius, more pronounced corona.
Color: [Appropriate for type]
Background: Transparent.
```

## Planet Sprite Templates

### Rocky Planet
```
Create a 256x256 pixel art planet sprite.

Type: Rocky terrestrial planet
Surface: [FEATURES - craters, mountains, canyons, deserts]
Atmosphere: [Thin/none visible, or colored haze]
Style: Pixel art but with more detail allowed at this size.
Lighting: Spherical shading, terminator line (day/night edge).
Background: Transparent.

Reference: Mars, Mercury, or Moon-like appearance.
```

### Gas Giant
```
Create a 256x256 pixel art gas giant planet.

Bands: Horizontal atmospheric bands in [COLORS]
Features: [Storm spots, swirls, specific patterns]
Style: Pixel art with smooth color gradients via dithering.
Lighting: Spherical, subtle shadow on one side.
Background: Transparent.

Reference: Jupiter or Saturn-like appearance.
```

### Habitable World
```
Create a 256x256 pixel art habitable planet.

Surface: Oceans (blue), continents (green/brown), ice caps (white).
Clouds: Scattered white cloud cover.
Atmosphere: Visible blue atmospheric rim/glow.
Style: Pixel art, Earth-like but can vary continents.
Background: Transparent.
```

## Portrait Templates

### Human Portrait
```
Create a 128x128 pixel art portrait of [CHARACTER].

View: Face-forward or 3/4 view.
Expression: [EXPRESSION - determined, wise, curious, etc.]
Attire: [CLOTHING/UNIFORM visible at shoulders]
Style: Retro pixel art, limited palette, clear features.
Background: Solid dark color (#1a1a2e) or transparent.

Specific features: [AGE, DISTINGUISHING MARKS, etc.]
```

### Alien Portrait
```
Create a 128x128 pixel art portrait of an alien.

Species concept: [ALIEN DESCRIPTION]
Must convey: Intelligence, personality, non-human biology.
Expression: [EMOTIONAL STATE]
Style: Pixel art, readable at 64x64 size.
Background: Transparent or solid color.
```

## Background Templates

### Star Field
```
Create a 1920x1080 deep space star field background.

Density: Varied - some dense clusters, some sparse regions.
Stars: Different sizes (mostly small dots, few larger).
Colors: Mostly white, occasional blue, yellow, red.
Depth: Sense of depth through size/brightness variation.
Style: Can be more detailed than sprites, but cohesive.

No nebulae or galaxies - pure star field.
```

### Nebula Background
```
Create a 1920x1080 space background with nebula.

Nebula: [COLOR] emission/reflection nebula.
Style: Soft, ethereal clouds of gas, not too busy.
Stars: Scattered through and around nebula.
Mood: [Mysterious / Vibrant / Ominous]

Should not distract from game UI in foreground.
```

### Galaxy View
```
Create a 1920x1080 background showing a spiral galaxy.

View: [Edge-on / Face-on / Angled]
Style: Realistic spiral structure, billions of stars implied.
Center: Bright galactic core.
Arms: Visible spiral arm structure.
Reference: Milky Way or Andromeda imagery from NASA.
```

## UI Element Templates

### Button
```
Create a pixel art UI button, [WIDTH]x[HEIGHT] pixels.

State: [Normal / Hover / Pressed]
Style: Sci-fi panel aesthetic, beveled edges.
Colors: Dark blue background (#16213e), cyan accent (#00fff5).
Text area: Leave center clear for text overlay.
```

### Panel
```
Create a pixel art UI panel frame, [WIDTH]x[HEIGHT] pixels.

Style: Sci-fi holographic/screen aesthetic.
Border: 2-3 pixel decorative frame.
Interior: Semi-transparent or solid dark.
Corners: Rounded or angular tech style.
```

## Modification Prompts

### Making Variants
```
Using the same style as [REFERENCE], create a variant that:
- Changes [SPECIFIC ELEMENT]
- Keeps [ELEMENTS TO PRESERVE]
- Adds [NEW ELEMENT]
```

### Style Correction
```
Adjust this image to be more pixel art styled:
- Reduce color count to [N] colors
- Remove anti-aliasing/blur
- Add clear dark outlines
- Make pixels crisp and defined
```

### Size Adjustment
```
Recreate this concept as a [NEW SIZE] pixel art image.
Simplify details to work at smaller scale.
Maintain readability and key identifying features.
```

## 3D Interior Texture Templates

### ⚠️ CRITICAL: Seamless Requirement

All floor and ceiling textures **MUST BE SEAMLESS/TILEABLE** because the engine uses a 4x4 grid of tiles (16 tiles visible). Non-seamless textures will show visible seams.

### Ship Interior Floor Texture
```
Create a 512x512 seamless tileable texture for a spaceship floor.

CRITICAL SEAMLESS REQUIREMENTS:
1. Left edge MUST match right edge perfectly
2. Top edge MUST match bottom edge perfectly
3. No obvious focal point that will repeat visibly
4. Pattern density should be uniform across texture

Style: [ROOM TYPE - bridge, engineering, habitat, etc.]
Surface: Metallic deck plating with subtle panel lines
Details: [ADDITIONAL FEATURES - grating, lights, wear marks]
Colors: Dark metallic grays, subtle cyan/blue accent lighting
Format: PNG with no alpha (opaque texture)

Test: When tiled 4x4, no seams should be visible.
```

### Ship Interior Wall Texture
```
Create a 512x512 texture for spaceship interior walls.

Style: [ROOM TYPE] - sci-fi wall panels
Features: Panel lines, subtle tech details, possible screen/light accents
Colors: Dark grays with [ACCENT COLOR] lighting/indicators
Format: PNG, opaque

Note: Walls use single planes so seamless is less critical,
but seamless is still recommended for consistency.
```

### Ship Interior Ceiling Texture
```
Create a 512x512 seamless tileable texture for a spaceship ceiling.

CRITICAL SEAMLESS REQUIREMENTS (SAME AS FLOOR):
1. All edges must tile perfectly
2. No obvious repeating focal points
3. Uniform pattern density

Style: [ROOM TYPE] ceiling
Features: Recessed lighting panels, ventilation grates, conduit covers
Colors: Dark gray base with [COLOR] light panels (emissive look)
Format: PNG, opaque

Should convey: Functional ship infrastructure above crew heads.
```

**Room Type Examples:**
- **Bridge**: Command/control, blue accent lights, clean surfaces
- **Engineering**: Industrial, orange/yellow warning lights, pipes/conduits
- **Habitat**: Softer, warmer lighting, residential feel
- **Culture**: Communal spaces, varied lighting, decorative elements
- **Core**: High-tech, red safety lighting, minimal aesthetic

## Scene Images with Transparent Windows

### ⚠️ CRITICAL: Gemini Cannot Generate True Alpha

Gemini image generation outputs RGB images only - it cannot produce RGBA with actual alpha transparency. When you ask for "transparent", it renders a checkerboard pattern as visible pixels.

**Solution:** Post-process AI-generated images using `generate-window-mask` tool to convert detected regions to true transparency.

### Recommended Workflow: Bright Windows + Keep-Top-N (Best Results)

The most reliable approach for window masking:

1. **Generate scene with bright/white windows** - AI understands "bright overexposed"
2. **Extract mask with `-keep-top=N`** - Keeps only the N largest bright regions (windows)
3. **Use mask for compositing** - Pixel-perfect alignment guaranteed

#### Step 1: Generate with Bright Windows

```
Spaceship bridge interior scene: [DESCRIPTION].
Windows showing pure bright white overexposed light (no space details visible).
Dark metallic interior with control consoles, crew chairs, detailed panels.
Art style: Detailed sci-fi illustration, warm orange/amber accent lighting.
The window areas will be extracted for compositing a dynamic starfield.
```

**Example prompt for bridge:**
```bash
go run ./cmd/voyage ai -generate-image -prompt "Spaceship bridge interior, three large panoramic windows showing pure bright white overexposed light, dark metallic interior with control consoles and pilot chairs, sci-fi cockpit with orange accent lighting, detailed panels and screens, cinematic composition"
```

#### Step 2: Extract Window Mask

Use `-mode=bright` with `-keep-top=N` to extract only the N largest bright regions (filtering out small lights, console LEDs, etc.):

```bash
# Extract top 3 largest bright regions as mask (for 3-window bridge)
go run ./cmd/generate-window-mask \
  -mode=bright \
  -threshold=180 \
  -keep-top=3 \
  -mask \
  input.png mask.png
```

#### Step 3: Generate Transparent Background

To get the bridge image with transparent windows (for overlay compositing):

```bash
# Same command without -mask flag creates transparent image
go run ./cmd/generate-window-mask \
  -mode=bright \
  -threshold=180 \
  -keep-top=3 \
  input.png transparent.png
```

#### Why This Works Better

- **AI can't preserve exact pixel positions** when given a mask as reference
- **Extracting from generated image** guarantees pixel-perfect alignment
- **keep-top=N** filters out small bright areas (lights, screens, reflections)
- **Result:** Clean mask with only the main window regions

### Alternative Methods

#### Method 1: Black Windows

Request pure black windows, then convert black pixels to transparent.

```
Sci-fi spaceship interior scene: [DESCRIPTION].
CRITICAL: All windows showing space must be PURE BLACK (#000000).
The window glass areas should be solid black with no stars, no reflections, no gradients.
Interior should have warm ambient lighting with detailed sci-fi aesthetic.
```

**Post-process:**
```bash
go run ./cmd/generate-window-mask -mode=black -keep-top=3 input.png output.png
```

**Pros:** Works when bright mode catches too many interior lights
**Cons:** Dark interior areas may be misidentified as windows

#### Method 2: Checkerboard (Frame-Only Images)

Request transparent areas, AI renders checkerboard pattern.

```
PNG with alpha transparency. Spaceship interior frame: [DESCRIPTION].
CRITICAL: Window openings must be COMPLETELY TRANSPARENT (alpha=0).
Only render the solid interior surfaces and frames.
```

**Post-process:**
```bash
go run ./cmd/generate-window-mask -mode=checker input.png output.png
```

#### Method 3: Magenta Chroma Key

Request magenta (#FF00FF) windows for easy detection.

```
[SCENE DESCRIPTION]
CRITICAL: All window/viewport areas must be filled with PURE MAGENTA (#FF00FF).
```

**Post-process:**
```bash
go run ./cmd/generate-window-mask -mode=magenta input.png output.png
```

### Tool Reference

```bash
# RECOMMENDED: Bright mode with keep-top filtering
go run ./cmd/generate-window-mask -mode=bright -threshold=180 -keep-top=3 -mask input.png mask.png
go run ./cmd/generate-window-mask -mode=bright -threshold=180 -keep-top=3 input.png transparent.png

# Other modes
go run ./cmd/generate-window-mask -mode=black input.png output.png
go run ./cmd/generate-window-mask -mode=checker input.png output.png
go run ./cmd/generate-window-mask -mode=magenta input.png output.png

# Adjust thresholds
go run ./cmd/generate-window-mask -mode=bright -threshold=200 ...  # Higher = stricter
go run ./cmd/generate-window-mask -mode=black -threshold=48 ...    # Higher = catches more

# Invert detection
go run ./cmd/generate-window-mask -mode=bright -invert input.png output.png
```

### Quick Pipeline Example

```bash
# 1. Generate bridge with bright windows
go run ./cmd/voyage ai -generate-image -prompt "Spaceship bridge, 3 windows showing bright white light, dark interior"

# 2. Extract mask (top 3 regions)
go run ./cmd/generate-window-mask -mode=bright -threshold=180 -keep-top=3 -mask \
  assets/generated/response_XXX.png \
  assets/decks/bridge/window_mask_large.png

# 3. Copy background
cp assets/generated/response_XXX.png assets/decks/bridge/background.png

# 4. Test
go run ./cmd/game
```
