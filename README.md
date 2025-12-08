# Stapledon's Voyage

*Travel as fast as you like. Live with the consequences.*

A hard sci-fi philosophy simulator where you pilot a near-light-speed ship with just 100 subjective years to explore the galaxy. Every journey triggers brutal time dilation: while you age slowly, entire civilizations rise, evolve, and die in the centuries that pass between your visits.

## Quick Start

```bash
# Run the game (development build)
make run-mock

# Run tests
go test ./...
```

## About

You discover the one "cheat" this universe allows: a drive that lets your ship cruise arbitrarily close to the speed of light. You don't get magic, FTL messaging, or time travel — just real relativistic time dilation.

**You have 100 subjective years aboard your ship.**

Every journey costs you time — and costs the universe centuries. The civilizations you visit age, change, forget you, or die while you travel. Your crew ages with you, the one constant in a universe that keeps changing out from under you.

At the end of your life, the simulation fast-forwards to Year 1,000,000 and shows you what your choices did to the galaxy. Not what was "good" or "bad" — just what happened. You decide what it means.

## What You'll Actually See

When you travel at near-light speeds, the universe doesn't just tick by faster — it looks fundamentally different. We implement the exact physics of special relativity:

![Relativistic Visual Effects](docs/images/sr_demo.gif)

**What's happening in this animation:**
1. **Acceleration** (0 → 0.5c): Stars ahead brighten and blueshift as you speed up
2. **Looking around at 0.5c**: The "relativistic ring" — aberration compresses all starlight into a bright halo
3. **Looking backward**: Near-total darkness — light from behind is aberrated forward and out of view
4. **Deceleration**: Stars return to normal as you slow down

This isn't artistic license — these are the exact formulas from Einstein's special relativity:
- **Aberration**: cos(θ) = (cos(θ') + β) / (1 + β·cos(θ'))
- **Doppler shift**: D = γ(1 + β·cos(θ))
- **Relativistic beaming**: Intensity scales as D³

At 0.9c looking forward, incoming light is amplified ~80× (whiteout). Looking backward, it's reduced to ~1% (near-darkness). This is what interstellar travel would actually look like.

### Near Black Holes: General Relativity

When you approach massive objects like black holes, spacetime itself warps. We implement gravitational lensing based on the Schwarzschild metric:

![Black Hole Journey](docs/images/gr_journey.gif)

**What's happening in this animation:**
1. **Approach**: As you get closer, the gravitational potential (φ) increases and spacetime curvature becomes visible
2. **Einstein ring**: Light from behind the black hole is bent around it, creating a bright ring at the photon sphere (r = 1.5 × Schwarzschild radius)
3. **Event horizon**: The central darkness where not even light can escape
4. **Retreat**: The lensing effect diminishes as you move away

The shader implements:
- **Gravitational lensing**: Light rays bend toward the mass, distorting the view of stars behind
- **Photon sphere glow**: At r = 1.5rs, photons orbit the black hole, creating a bright accretion ring
- **Schwarzschild radius (rs)**: The event horizon boundary — anything closer is lost forever
- **Gravitational potential (φ)**: Controls the intensity of spacetime curvature effects

This is what approaching a stellar-mass black hole would actually look like — beautiful, terrifying, and scientifically accurate.

### Coming Home: 3D Solar System Flyby

What does it feel like to decelerate from near-light speed back into our solar system? This demo combines 3D planet rendering with relativistic visual effects:

![Solar System Flyby](docs/images/solar-flyby.gif)

**What's happening in this animation:**
1. **Starting at 0.9c**: Blue-shifted starfield ahead, planets stretched by relativistic aberration
2. **Passing the outer planets**: Neptune, Uranus, Saturn (with rings), Jupiter flash by as we decelerate
3. **Doppler color shift**: Stars ahead are blue-white (approaching), behind would be red (receding)
4. **Decelerating to 0.3c**: The Lorentz factor drops from ~2.3 to ~1.05 as we slow down
5. **Arriving at Earth**: Coming home after a journey where centuries passed on Earth while we aged years

The demo uses:
- **Tetra3D**: Software 3D renderer for planet spheres with NASA texture maps
- **UV sphere mesh**: Proper equirectangular mapping for realistic planet surfaces
- **SR shader**: Real-time Doppler shift and aberration applied to the entire scene
- **Time dilation HUD**: Shows velocity (β), Lorentz factor (γ), and elapsed time

At 0.9c, the Lorentz factor γ = 1/√(1-0.81) ≈ 2.29. For every year you experience aboard the ship, 2.29 years pass on Earth. A 10-year subjective journey means coming home to find 23 years have passed.

## Design Pillars

Five non-negotiable constraints guide every feature:

| Pillar | What It Means |
|--------|---------------|
| **Choices Are Final** | No saves, no reloads. Live with consequences or start fresh. |
| **The Game Doesn't Judge** | Present facts, not morals. Players find their own meaning. |
| **Time Has Emotional Weight** | Loneliness, loss, and treasuring what remains — not just numbers. |
| **The Ship Is Home** | Crew provides human-scale grounding against cosmic-scale alienation. |
| **Grounded Strangeness** | Aliens are scientifically plausible, maximally diverse, and extensible. |

See [docs/vision/core-pillars.md](docs/vision/core-pillars.md) for full pillar definitions.

## Inspiration

Named after [Olaf Stapledon](https://en.wikipedia.org/wiki/Olaf_Stapledon), the science fiction author known for cosmic-scale narratives like *Star Maker* and *Last and First Men*. The game embodies his perspective: vast timescales, philosophical exploration, and the humbling realization of how small individual choices feel against deep time — yet how consequential they remain.

## Status

**Early Development** (v0.1.0) - Engine with relativistic visual effects is functional. Core gameplay (civilization simulation, crew management) is not yet implemented.

### Implemented

- 2D/3D rendering engine (Go/Ebiten + Tetra3D)
- **Special relativity shader** — aberration, Doppler shift, relativistic beaming
- **General relativity shader** — gravitational lensing, Einstein rings near black holes
- **3D planet rendering** — textured spheres with NASA maps, Saturn's rings
- Post-processing pipeline (bloom, vignette, CRT, chromatic aberration)
- Audio system, input handling, game loop
- Visual test framework with golden file comparison

### Roadmap

See [CHANGELOG.md](CHANGELOG.md) for release history and [design_docs/](design_docs/) for feature planning.

Next milestones:
- **v0.5.x** - Ship exploration, galaxy map, dialogue system
- **v0.6.x** - Journey system with time dilation, civilizations
- **v0.7.x+** - Exploration modes, endgame

See [docs/game-vision.md](docs/game-vision.md) for the complete game design.

## Technical

This project serves as the primary integration test for **AILANG**, a new programming language for game simulation.

| Layer | Purpose |
|-------|---------|
| **Engine** (`engine/`) | Input capture, rendering - Go/Ebiten |
| **Simulation** (`sim_gen/`) | Game logic - currently mock Go, will be AILANG |

See [CLAUDE.md](CLAUDE.md) and [DEVELOPMENT.md](DEVELOPMENT.md) for technical details.

## Releases

See [Releases](https://github.com/sunholo-data/stapledons_voyage/releases) for downloadable binaries and [CHANGELOG.md](CHANGELOG.md) for version history.

Current version: **v0.1.0**

## Requirements

- Go 1.21+
- Make

## License

MIT
