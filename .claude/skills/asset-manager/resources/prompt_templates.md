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

### Basic Terrain Tile
```
Create a 64x32 pixel art isometric tile showing [TERRAIN TYPE].

Style: Retro 16-bit pixel art, limited palette (8-12 colors),
clear outlines, no anti-aliasing, crisp pixels.

Shape: Diamond/rhombus isometric tile, 2:1 width-to-height ratio.
Background: Transparent (alpha channel).
Lighting: Light source from top-left.

Specific details: [ADDITIONAL DETAILS]
```

**Examples:**
- `[TERRAIN TYPE]`: alien crystal formations
- `[ADDITIONAL DETAILS]`: purple and cyan crystals, bioluminescent glow

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
