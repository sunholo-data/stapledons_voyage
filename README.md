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

**Early Development** - The game engine is functional but the core gameplay (relativistic travel, civilization simulation) is not yet implemented.

The current codebase provides:
- 2D rendering engine (Go/Ebiten)
- Input handling and game loop
- Foundation for AILANG integration

See [docs/game-vision.md](docs/game-vision.md) for the complete game design.

## Technical

This project serves as the primary integration test for **AILANG**, a new programming language for game simulation.

| Layer | Purpose |
|-------|---------|
| **Engine** (`engine/`) | Input capture, rendering - Go/Ebiten |
| **Simulation** (`sim_gen/`) | Game logic - currently mock Go, will be AILANG |

See [CLAUDE.md](CLAUDE.md) and [DEVELOPMENT.md](DEVELOPMENT.md) for technical details.

## Releases

See [Releases](https://github.com/sunholo/stapledons_voyage/releases) for downloadable binaries.

Current version: **v0.1.0**

## Requirements

- Go 1.21+
- Make

## License

MIT
