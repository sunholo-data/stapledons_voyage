# Star Texture Sources

## Currently Used Textures

### Solar System Scope (CC BY 4.0 - Commercial OK)
- **Website:** https://www.solarsystemscope.com/textures/
- **License:** Creative Commons Attribution 4.0 International
- **Usage:** May use, adapt, and share for any purpose, even commercially
- **Attribution:** Credit required

Downloaded:
- `sun_2k.jpg` - 2048x1024, G-type star (our Sun)
- `sun_8k.jpg` - 8192x4096, G-type star (high-res)

### DeviantArt - FarGetaNik (CC BY-NC-SA 3.0 - Non-Commercial)
- **Website:** https://www.deviantart.com/fargetanik
- **License:** Creative Commons Attribution-NonCommercial-ShareAlike 3.0
- **Usage:** Non-commercial only, share-alike required
- **Note:** Great for prototyping, need commercial alternatives for release

Available (4096x2048 PNG):
- O, B, A spectral types: https://www.deviantart.com/fargetanik/art/Star-Texture-Map-4k-O-B-A-spectral-types-814254075
- F, G, K spectral types: https://www.deviantart.com/fargetanik/art/Star-Texture-Map-4k-F-G-K-Spectral-Types-814379574
- M spectral type: https://www.deviantart.com/fargetanik/art/Star-Texture-Map-4k-Spectral-Type-M-814462913

### Space Spheremaps (Free for Commercial Use)
- **Website:** https://www.spacespheremaps.com/
- **License:** Free for personal and commercial projects
- **Categories:** Local Star Spheremaps, HDRi Space Spheremaps

### NASA SVS (Public Domain)
- **Deep Star Maps 2020:** https://svs.gsfc.nasa.gov/4851/
- **Tycho Catalog Skymap:** https://svs.gsfc.nasa.gov/3572
- **License:** Public domain (US Government work)

## Spectral Type Color Reference

| Type | Temperature | Color | Example Stars |
|------|-------------|-------|---------------|
| O | 30,000-50,000K | Blue-white | Zeta Puppis |
| B | 10,000-30,000K | Blue-white | Rigel, Spica |
| A | 7,500-10,000K | White | Sirius, Vega |
| F | 6,000-7,500K | Yellow-white | Procyon |
| G | 5,200-6,000K | Yellow | Sun, Alpha Centauri A |
| K | 3,700-5,200K | Orange | Arcturus, Alpha Centauri B |
| M | 2,400-3,700K | Red | Proxima Centauri, Betelgeuse |

## TODO: Generate Color Variants

Since commercial-friendly spectral textures are limited, we can:
1. Use sun texture as base
2. Apply color tint in shader/material for each spectral type
3. Or procedurally generate star surfaces with noise + color gradient
