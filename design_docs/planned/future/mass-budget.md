# Mass Budget System

## Status
- **Status:** Planned
- **Sprint:** Vision Integration - Sprint 2
- **Priority:** P2 (Resource system, not critical path)
- **Source:** [Interview: Bubble Ship Design](../../../docs/vision/interview-log.md#2025-12-06-session-bubble-ship-design-integration)

## Game Vision Alignment

| Pillar | Score | Notes |
|--------|-------|-------|
| Choices Are Final | ✅ Strong | Once mass is committed, it's committed |
| Game Doesn't Judge | ✅ Strong | No "right" allocation |
| Time Has Emotional Weight | ⚪ N/A | Resource system |
| Ship Is Home | ✅ Strong | Home has limits |
| Grounded Strangeness | ✅ Strong | Physics-based scarcity |
| We Are Not Built For This | ✅ Strong | Can't have everything |

## Feature Overview

The bubble contains **finite internal mass**. Everything inside competes for the same atoms:

- Population growth (bodies need mass)
- Proto-tech fabrication (upgrades need mass)
- Infrastructure maintenance (repairs need mass)
- Food/air/water cycling (stable, but damaged systems need mass)

**Key Insight:** You can't have everything. Choose what matters.

## Design Decisions

From [design-decisions.md](../../../docs/vision/design-decisions.md):

| Decision | Summary |
|----------|---------|
| Finite Mass Budget | Proto-tech and population compete |
| Slow Mass Absorption | Trickle from ISM, not a solution |
| Proto-Tech via Information | Tech costs mass to fabricate |

## Mass Categories

### Starting Mass (Approximate)

| Category | Mass | Notes |
|----------|------|-------|
| **Structure** | Fixed | Ship hull, spire, infrastructure |
| **Equipment** | 5,000 kg | Fabricators, life support, systems |
| **Crew** | 7,000 kg | ~100 people at ~70kg average |
| **Stores** | 3,000 kg | Food, water, raw materials buffer |
| **Available** | 5,000 kg | "Free" mass for player allocation |

**Total internal mass:** ~20,000 kg (not including structure)

*Note: These are gameplay numbers, not hard physics.*

### Mass Sinks

| Sink | Cost Range | Notes |
|------|------------|-------|
| **Population +1** | 70 kg | New person (birth or growth to adult) |
| **Minor Proto-tech** | 50-200 kg | Sensor upgrades, small systems |
| **Major Proto-tech** | 500-2000 kg | Engine improvements, weapons |
| **Emergency Repair** | 100-500 kg | Depending on damage |
| **Luxury Item** | 10-50 kg | Morale improvements |

### Mass Sources

| Source | Rate | Notes |
|--------|------|-------|
| **ISM Absorption** | ~1 kg/year | Typical interstellar medium |
| **Stellar Wind** | ~5 kg/year | Near active stars |
| **Nebula Transit** | ~20 kg/year | Dense regions, rare |
| **Recycling** | N/A | Deaths return mass to pool |

## Visibility to Player

From [open-questions.md](../../../docs/vision/open-questions.md#how-does-the-mass-budget-system-work):

**Options:**
1. **Hidden:** Player sees symptoms (can't build X), not numbers
2. **Abstract:** "Mass reserves: Comfortable / Tight / Critical"
3. **Visible:** Actual kg display, spreadsheet optimization risk

**Recommendation:** Abstract display with detail on demand. Avoid optimization gameplay.

```
┌─────────────────────────────────┐
│ MASS RESERVES                   │
│ ████████████░░░░ Comfortable    │
│                                 │
│ Available: ~2,300 kg            │
│ Absorption: +1.2 kg/year        │
│                                 │
│ [Details]                       │
└─────────────────────────────────┘
```

## Decision Framework

When player wants to spend mass:

```
┌─────────────────────────────────┐
│ FABRICATE: Advanced Sensors     │
│                                 │
│ Mass Cost: 450 kg               │
│ Current Available: 2,300 kg     │
│ After: 1,850 kg                 │
│                                 │
│ This will delay population      │
│ growth for approximately        │
│ 6 years at current rates.       │
│                                 │
│ [Confirm]  [Cancel]             │
└─────────────────────────────────┘
```

## Tradeoff Examples

### Population vs. Technology

| Choice | Consequence |
|--------|-------------|
| Allow unrestricted births | Less mass for proto-tech, slower upgrades |
| Limit population | More tech capability, social tension, fewer hands |
| Balanced approach | Slower everything, stable society |

### Emergency vs. Long-term

| Choice | Consequence |
|--------|-------------|
| Repair immediately | Solves crisis, depletes reserves |
| Defer repair | Risk escalation, save mass for better opportunity |
| Partial repair | Buys time, leaves vulnerability |

### Specialization vs. Flexibility

| Choice | Consequence |
|--------|-------------|
| Deep investment in one tech | Powerful capability, no backup |
| Broad shallow tech | Flexible but not exceptional |
| Save for unknowns | Prepared for opportunities, behind on everything |

## Recycling Mechanics

Mass is never destroyed, only transformed:

- **Deaths:** Bodies return mass to pool (respectfully handled)
- **Decommissioning:** Old tech can be dismantled for partial mass recovery
- **Food Cycle:** Eating doesn't consume mass, just transforms it

### Dismantling

| Original Cost | Recovery | Lost |
|---------------|----------|------|
| 100 kg | 70 kg | 30 kg (inefficiency) |
| 500 kg | 350 kg | 150 kg |

Dismantling is **not free** - some mass is lost to inefficiency. This prevents infinite reshuffling.

## Crisis Mechanics

Mass pressure can trigger crises:

| Condition | Crisis |
|-----------|--------|
| Available < 500 kg | "Low reserves" warning |
| Available < 100 kg | Rationing required, morale impact |
| Available = 0 | Emergency: must dismantle something |
| Population > sustainable | Food synthesis strain |

## AILANG Types

```ailang
type MassCategory =
    | Structure        -- Fixed, not player-accessible
    | Equipment        -- Ships systems
    | Population       -- Crew biomass
    | Stores           -- Buffer materials
    | Available        -- Free for allocation

type MassTransaction =
    | Fabricate(tech_id: int, cost: float)
    | Birth(person_id: int, mass: float)
    | Repair(system_id: int, cost: float)
    | Dismantle(item_id: int, recovery: float)
    | Absorb(source: AbsorptionSource, amount: float)
    | Recycle(person_id: int, mass: float)

type MassBudget = {
    categories: Map(MassCategory, float),
    available: float,
    absorption_rate: float,
    transactions: [MassTransaction]
}

type MassReserveLevel =
    | Abundant      -- > 3000 kg
    | Comfortable   -- 1500-3000 kg
    | Tight         -- 500-1500 kg
    | Critical      -- < 500 kg
    | Emergency     -- < 100 kg
```

## Integration with Other Systems

### Population System
- Births check mass availability
- Deaths return mass
- Population pressure affects available mass

### Proto-tech System
- Fabrication costs mass
- Blueprints don't cost mass (information)
- Building implementation costs mass

### Crisis System
- Low mass triggers crises
- Crises may require mass to resolve
- Cascading failures possible

### Faction System
- Resource allocation creates political tension
- "Who gets the mass?" becomes faction issue
- Hoarding vs. sharing creates conflict

## Engine Integration

### UI
- Mass budget panel (abstract or detailed)
- Transaction confirmation dialogs
- Reserve level indicators
- Absorption rate display

### Simulation
- Track mass across all categories
- Apply absorption over time
- Handle transactions atomically
- Trigger crisis checks

## Testing Scenarios

1. **Comfortable Start:** Begin game, observe mass reserves, make small purchases
2. **Population Pressure:** Allow unrestricted births, observe mass depletion
3. **Tech Investment:** Heavy tech spending, observe population consequences
4. **Crisis Resolution:** Trigger low-mass crisis, resolve by dismantling
5. **Long Journey:** Extended ISM travel, observe slow absorption accumulation

## Success Criteria

- [ ] Mass scarcity creates meaningful choices
- [ ] Players understand tradeoffs without spreadsheet optimization
- [ ] Population and tech compete naturally
- [ ] Crises are recoverable but costly
- [ ] Absorption feels like a trickle, not a solution
- [ ] Recycling prevents total loss but isn't free
