# Bubble Constraint System

## Status
- **Status:** Planned
- **Sprint:** Vision Integration - Sprint 2
- **Priority:** P1 (Defines core game physics)
- **Source:** [Interview: Game Loop Origin](../../../docs/vision/interview-log.md#2025-12-06-session-game-loop-origin--bubble-constraint)

## Game Vision Alignment

| Pillar | Score | Notes |
|--------|-------|-------|
| Choices Are Final | ✅ Strong | Can't undo what you brought/didn't bring |
| Game Doesn't Judge | ⚪ N/A | Physics constraint, not moral |
| Time Has Emotional Weight | ⚪ N/A | Enables isolation |
| Ship Is Home | ✅ Strong | Defines the boundary of "home" |
| Grounded Strangeness | ✅ Strong | Hard sci-fi constraint |
| We Are Not Built For This | ✅ Strong | Permanent separation from universe |

## Feature Overview

The Higgs-bubble creates an **absolute boundary** between the ship and the universe:

> **Only information crosses the boundary. Mass cannot.**

This single constraint shapes the entire game:
- You are "memetic travelers" - carrying ideas, not cargo
- Alien tech is absorbed as blueprints, fabricated internally
- No physical rescue or supply is possible
- The bubble is self-contained or it dies

## What Crosses the Boundary

### ✅ CAN Cross (Inward)

| Type | Mechanism | Gameplay Impact |
|------|-----------|-----------------|
| **Light/EM** | Transparent to visible spectrum | See the universe |
| **Radio signals** | Low-energy EM passes | Communication with civs |
| **Data/blueprints** | Encoded in light | Proto-tech acquisition |
| **Trace hydrogen** | Sub-femtogram particles | Very slow mass gain |
| **Philosophical frameworks** | Ideas, not matter | Unlock new interpretations |

### ❌ CANNOT Cross (Inward)

| Type | Explanation | Gameplay Impact |
|------|-------------|-----------------|
| **Physical objects** | Higgs field blocks mass | No cargo, no gifts, no rescue |
| **People** | Mass cannot enter | Starting crew is all you have |
| **Alien artifacts** | Physical tech cannot enter | Must reverse-engineer from specs |
| **Resources** | No material resupply | Finite mass budget |

### ⬆️ CAN Cross (Outward)

| Type | Mechanism | Gameplay Impact |
|------|-----------|-----------------|
| **Light/signals** | Transparent both ways | Broadcast to civs |
| **Data transmission** | EM radiation | Share your archives |

### ❌ CANNOT Leave

| Type | Explanation | Gameplay Impact |
|------|-------------|-----------------|
| **Crew members** | Permanent containment | No EVA, no away missions |
| **Physical samples** | Mass trapped inside | Can't send probes |

## Proto-Tech Acquisition

When encountering alien technology:

1. **Receive specifications** - Blueprints, equations, principles cross as data
2. **Analyze with Archive** - AI helps interpret alien concepts
3. **Fabricate internally** - Use existing mass to build implementation
4. **Mass cost applied** - Each upgrade costs finite mass budget

```
Alien Civ → Data Transmission → Archive Analysis → Internal Fabrication → Working Tech
           (crosses boundary)                      (uses internal mass)
```

## Trace Hydrogen Absorption

The bubble can absorb extremely small mass from:
- Interstellar medium (ISM) - ~1 atom per cm³
- Stellar wind (near stars) - Higher density
- Nebulae - Dense regions

**Rate:** Roughly 1kg per year at typical ISM density (gameplay number, not hard physics)

**Impact:** Provides slight flexibility over long journeys but won't rescue poor planning.

## Design Decisions

From [design-decisions.md](../../../docs/vision/design-decisions.md):

| Decision | Summary |
|----------|---------|
| Proto-Tech via Information | Alien tech absorbed as blueprints, built internally |
| Finite Mass Budget | Competition between population and upgrades |
| Slow Mass Absorption | Trickle from ISM, not a solution |
| Radiation Shielding Automatic | Energy-dependent filtering, not player-managed |

## Boundary Physics

### Energy-Dependent Transparency

| Energy Level | Passes? | Examples |
|--------------|---------|----------|
| Visible light | ✅ Yes | Stars, planets visible |
| Infrared | ✅ Yes | Heat signatures detectable |
| Radio | ✅ Yes | Communication possible |
| UV | ✅ Mostly | Some filtering |
| X-ray | ❌ Filtered | Radiation protection |
| Gamma | ❌ Blocked | Cosmic ray shielding |

This explains why the ship is habitable - dangerous radiation is filtered automatically.

### Mass Threshold

The boundary has an effective "particle size" filter:
- **Photons:** Always pass (massless)
- **Neutrinos:** Pass (nearly massless, non-interacting)
- **Electrons:** Blocked (massive particles)
- **Atoms:** Blocked (except trace infiltration)
- **Molecules:** Blocked
- **Macroscopic objects:** Absolutely blocked

## Narrative Implications

### The Memetic Traveler Identity

You don't carry cargo - you carry:
- Ideas from civilizations
- Philosophies that reframe understanding
- Scientific principles that enable new technology
- Art, music, stories (digitized)
- Memories (in Archive and crew minds)

This makes every encounter about **exchange of meaning**, not trade of goods.

### Permanent Isolation

Once inside the bubble:
- You can never physically touch the universe again
- EVA is impossible
- If the ship breaks, no one can help
- You are truly alone, together

This serves Pillar 6: **We Are Not Built For This**

### First Contact Dynamics

When meeting aliens:
- They cannot board your ship
- You cannot board theirs
- All interaction is mediated by signals
- Trust must be built without physical presence

## Edge Cases

### Q: What about the spire?

The spire predates the bubble and may not obey the same rules. This is part of the mystery.

### Q: Can crew members leave and return?

No. The bubble is one-way for mass. Once inside, you stay inside.

### Q: What about births?

Babies are born inside the bubble, from mass already inside. Population growth uses internal mass.

### Q: What about death?

Bodies are recycled. Mass is conserved. This is both practical and thematically significant.

## AILANG Types

```ailang
type BoundaryTransfer =
    | LightSignal(string)           -- EM data
    | DataPacket(bytes)             -- Encoded information
    | TraceMass(float)              -- Femtogram-scale absorption
    | BlockedMass(string)           -- Rejected with reason

type AbsorptionSource =
    | InterstellarMedium
    | StellarWind(star_id: int)
    | Nebula(density: float)

type TransferResult = {
    success: bool,
    type: BoundaryTransfer,
    mass_delta: float,
    archive_data: Option(string)
}
```

## Engine Integration

### Visual Representation
- Subtle shimmer at bubble boundary
- Incoming signals show as light touches
- Blocked objects show rejection effect (for clarity)

### Audio
- Muffled external sounds (everything is mediated)
- Signal reception sounds
- Absorption hum (when near dense regions)

### UI
- Mass budget display (see mass-budget.md)
- Absorption rate indicator
- Signal log for received data

## Testing Scenarios

1. **Signal Reception:** Receive alien transmission, verify data crosses
2. **Mass Rejection:** Attempt to "receive" physical gift, verify blocked
3. **Trace Absorption:** Long journey in ISM, verify slow mass gain
4. **Proto-Tech Build:** Receive blueprints, fabricate tech, verify mass cost

## Success Criteria

- [ ] Boundary constraint is clear and consistent
- [ ] Proto-tech acquisition feels meaningful
- [ ] Mass absorption is slow but noticeable over decades
- [ ] Radiation filtering is automatic and reliable
- [ ] Player understands they are "memetic travelers"
- [ ] Isolation creates appropriate emotional weight
