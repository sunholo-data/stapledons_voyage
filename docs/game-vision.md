# Stapledon's Voyage - Game Vision

## Elevator Pitch

Stapledon's Voyage is a hard sci-fi philosophy simulator where you pilot a near-light-speed ship with just 100 subjective years to explore the galaxy. Every journey you take triggers brutal time dilation: while you age slowly, entire civilizations rise, evolve, and die in the centuries that pass between your visits.

**Tagline:** *"Travel as fast as you like. Live with the consequences."*

## Short Description

In Stapledon's Voyage, you're the first traveler with near-light-speed drive in a realistically simulated galaxy. Plot irreversible journeys, watch millennia pass in the outside universe while only years pass for you, and decide which civilizations to uplift, connect, or leave alone. At the end of your 100-year career, the game fast-forwards to Year 1,000,000 and shows you what your choices did to the galaxy.

## Long Description

You discover the one "cheat" this universe allows: a drive that lets your ship cruise arbitrarily close to the speed of light. You don't get magic, FTL messaging, or time travel — just real relativistic time dilation.

**You have 100 subjective years aboard your ship.**

Every time you commit to a journey, you choose a destination and a cruise speed between 0.9c and 0.999999c. The faster you go, the less time passes for you… and the more centuries or millennia pass for everyone else. Civilizations you just met may be extinct, transcendent, or unrecognizable when you return.

### Galaxy Simulation

Behind the scenes, a galaxy-scale simulation is always running:

- Each civilization has population, energy access, innovation, cohesion, expansion drive, sustainability, contact openness, and existential risk
- Philosophies act like "meta-tech trees", changing how they develop, how stable they are across millennia, and how they respond to contact
- Extinctions, breakthroughs, colonization, and philosophical shifts all unfold while you're in transit

### Your Ship

Your ship is a self-contained story:

- A finite crew ages, forms relationships, has children, dies
- You install technologies in limited system slots, trade knowledge, archive cultures, and decide how much of humanity you export
- Each journey is locked in once you commit: you can't turn around without paying the time cost again

### Victory Conditions

At the end of your life, the simulation is accelerated to Year 1,000,000. The game then scores you against one chosen victory condition:

| Victory | Description |
|---------|-------------|
| **The Shepherd** | How many civilizations are still alive at Year 100 because of you |
| **The Gardener** | How much philosophical diversity you kept in play |
| **The Unifier** | How strongly connected the galactic contact network is |
| **The Witness** | How far, and how deep into cosmic time, you traveled |
| **The Founder** | Whether institutions, colonies, and lineages you started persist across deep time |
| **The Prometheus** | How widely near-light/FTL travel spread, and how much traces back to you |

Then you get a **legacy report**: a concrete breakdown of which civilizations you saved, which you doomed, which philosophies hybridized because you introduced them, and what the galaxy would have looked like if you'd stayed at home.

## Core Features

### Plan Irreversible Relativistic Journeys

Use a star-map and time-dilation calculator to choose destinations and speeds. See how many years you'll age, how many centuries will pass for the target civilization, and how many of your remaining 100 years you're burning.

### Watch Civilizations Evolve Over Deep Time

Each year of external time, the sim updates population, technology, stability, expansion, philosophy, and risks. Civilizations can colonize, change their worldview, go to war, hybridize, or quietly vanish.

### Meet Radically Different Minds

Engage with species that echolocate instead of seeing, hive minds with no concept of "I", fungal networks that experience weeks the way you experience seconds, and human-analogue cultures built on very different philosophical commitments (gift economies, absolute empiricism, sacred mortality, consensus-only politics, etc.).

### Trade Technology, Knowledge, and Ideas

Exchange drives, energy tech, biology, and philosophical frameworks. Each trade changes their trajectory: you can stabilize them, accelerate their expansion, or hand them tools they're not ready for.

### Shape — and Destabilize — a Contact Network

Every contact adds edges to a galactic graph. Civilizations you connect may cooperate, synthesize new philosophies, or destroy each other. The network structure at Year 100 determines whether you were a Shepherd, Gardener, Unifier, or arsonist.

### Live a Finite, Generational Ship Life

Your crew ages and dies across your 100 years; new generations are born having never seen Earth. On the ship, you manage morale, relationships, and which technologies to install. Off the ship, centuries of history tick past between each decision.

### End with a Concrete Legacy

After one last fast-forward to Year 1,000,000, the game tells you, in specific, system-grounded terms: who lived because of you, who died because of you, what new philosophies emerged, and whether your name, ship, or culture still exists in any meaningful sense.

## Inspiration

The game is named after [Olaf Stapledon](https://en.wikipedia.org/wiki/Olaf_Stapledon), the science fiction author known for his cosmic-scale narratives like *Star Maker* and *Last and First Men*. The game embodies his perspective: vast timescales, philosophical exploration, and the humbling realization of how small individual choices feel against the backdrop of deep time — yet how consequential they remain.

## Technical Implementation

The game is built using:

- **AILANG** - A custom programming language for simulation logic (in development)
- **Go/Ebiten** - 2D game engine for rendering and input
- **Mock simulation** - Hand-written Go for development while AILANG compiler is built

See [DEVELOPMENT.md](../DEVELOPMENT.md) for technical details.

## AI-Assisted Development

This project serves as a primary integration test for AILANG and embraces AI-assisted development. Key principles:

### Self-Testing & Autonomous Verification

The development environment should enable AI agents (like Claude Code) to:

1. **Verify visual output** - Capture screenshots programmatically to validate rendering
2. **Run automated scenarios** - Execute predefined input sequences and verify outcomes
3. **Detect regressions** - Compare current output against golden images/expected states
4. **Debug autonomously** - Investigate issues without requiring human screenshots

### Headless Testing Capabilities

The game supports headless rendering modes for automated testing:

- `--screenshot <frames> <output.png>` - Capture frame after N ticks
- `--scenario <name>` - Run predefined test scenario with simulated inputs
- Golden image comparison for visual regression testing

### Continuous Verification Loop

```
Code Change → Build → Headless Test → Screenshot → AI Analysis → Fix/Iterate
```

This enables rapid iteration where the AI can:
- Make a change
- Run the game headlessly
- View the screenshot output
- Verify correctness or identify issues
- Continue iterating without human intervention

### Benefits

- **Faster development** - AI can self-verify without waiting for human testing
- **Better coverage** - Automated scenarios catch edge cases
- **Documentation** - Screenshots serve as visual documentation of features
- **Regression prevention** - Golden images catch unintended visual changes
