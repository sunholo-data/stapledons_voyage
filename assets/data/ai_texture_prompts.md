# AI Texture Generation Prompts

Generate these textures using AI image generators (DALL-E, Midjourney, Stable Diffusion).

**CRITICAL: All textures MUST be:**
- **Equirectangular projection (2:1 aspect ratio)** - e.g., 2048x1024, 1024x512
- **Seamless at left/right edges** (wraps around a sphere)
- **NO background** - just the surface texture
- **NO stars, space, or black areas**

## Missing Textures

### Priority 1: Major Objects

**Pluto** (tan-brown, heart-shaped nitrogen glacier):
```
Equirectangular planet texture map, 2048x1024 pixels, 2:1 aspect ratio.
Pluto's surface: tan-brown rocky terrain with light tan/beige heart-shaped 
nitrogen ice glacier (Sputnik Planitia). Rough cratered highlands in dark 
brown. No stars or background. Seamless edges. NASA New Horizons style.
```

**Io** (yellow sulfur, volcanic):
```
Equirectangular moon texture map, 2048x1024 pixels, 2:1 aspect ratio.
Io volcanic moon surface: bright yellow and orange sulfur deposits, 
black volcanic calderas, reddish lava flows, smooth volcanic plains.
No stars or background. Seamless edges. NASA Galileo mission style.
```

**Europa** (cracked ice, tan lines):
```
Equirectangular moon texture map, 2048x1024 pixels, 2:1 aspect ratio.
Europa ice moon surface: smooth pale tan/cream ice with dark reddish-brown
crisscrossing cracks and lineae. Slightly mottled texture. Very few craters.
No stars or background. Seamless edges. NASA Galileo mission style.
```

**Ganymede** (gray, grooved terrain):
```
Equirectangular moon texture map, 2048x1024 pixels, 2:1 aspect ratio.
Ganymede surface: mix of dark gray ancient cratered regions and lighter
gray grooved terrain with parallel ridges. Large impact basins. Brown
tinted areas. No stars or background. Seamless edges. NASA Galileo style.
```

**Callisto** (dark gray, heavily cratered):
```
Equirectangular moon texture map, 2048x1024 pixels, 2:1 aspect ratio.
Callisto surface: very dark gray/black heavily cratered terrain, bright
white ice deposits in crater floors, ancient surface. No smooth regions.
No stars or background. Seamless edges. NASA Galileo mission style.
```

**Titan** (orange haze, methane lakes):
```
Equirectangular moon texture map, 2048x1024 pixels, 2:1 aspect ratio.
Titan surface: orange/tan hazy atmosphere bands, dark hydrocarbon lake
regions near poles, lighter equatorial dunes. Murky, obscured surface.
No stars or background. Seamless edges. NASA Cassini mission style.
```

**Triton** (pinkish, cantaloupe terrain):
```
Equirectangular moon texture map, 2048x1024 pixels, 2:1 aspect ratio.
Triton surface: pinkish-white nitrogen ice, distinctive "cantaloupe terrain"
dimpled texture, dark streaks from geysers, south polar cap slightly darker.
No stars or background. Seamless edges. NASA Voyager 2 style.
```

### Priority 2: Saturn's Moons

**Enceladus** (brilliant white ice):
```
Equirectangular moon texture map, 2048x1024 pixels, 2:1 aspect ratio.
Enceladus surface: brilliant white pristine ice, tiger stripe cracks near
south pole in blue-white, smooth plains, few craters. Very bright albedo.
No stars or background. Seamless edges. NASA Cassini mission style.
```

**Rhea, Dione, Tethys** (gray/white icy):
```
Equirectangular moon texture map, 2048x1024 pixels, 2:1 aspect ratio.
Icy moon surface: bright white/gray ice, moderate cratering, subtle
wispy terrain features, smooth plains between craters. Clean ice surface.
No stars or background. Seamless edges.
```

**Mimas** (gray, huge crater):
```
Equirectangular moon texture map, 2048x1024 pixels, 2:1 aspect ratio.
Mimas surface: gray ice with one enormous crater (Herschel) dominating
one hemisphere, moderate cratering elsewhere. Death Star appearance.
No stars or background. Seamless edges. NASA Cassini style.
```

**Iapetus** (two-tone):
```
Equirectangular moon texture map, 2048x1024 pixels, 2:1 aspect ratio.
Iapetus surface: stark two-tone appearance - one hemisphere bright white
ice, opposite hemisphere very dark (almost black) material. Sharp boundary.
No stars or background. Seamless edges. NASA Cassini style.
```

### Priority 3: Other Objects

**Charon** (gray, chasms):
```
Equirectangular texture map, 2048x1024 pixels, 2:1 aspect ratio.
Charon surface: gray rock and ice, large canyon system (Serenity Chasma),
dark polar cap (Mordor Macula), cratered terrain. Pluto's companion.
No stars or background. Seamless edges. NASA New Horizons style.
```

**Vesta, Ceres, Pallas** (rocky asteroids):
```
Equirectangular asteroid texture map, 2048x1024 pixels, 2:1 aspect ratio.
Rocky asteroid surface: gray rock with craters of various sizes, some
bright spots (ice or salt deposits for Ceres), rough terrain, no atmosphere.
No stars or background. Seamless edges.
```

## Output Settings
- Size: 2048x1024 pixels (minimum 1024x512)
- Format: JPG or PNG
- No transparency needed (full surface coverage)
- Filename: lowercase, no spaces (e.g., `io.jpg`, `europa.jpg`)

## Save Location
`assets/planets/` in the project root
