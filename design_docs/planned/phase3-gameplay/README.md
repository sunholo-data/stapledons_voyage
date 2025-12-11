# Phase 3: Core Gameplay (The Heart of the Game)

**Priority:** P0
**Status:** Planned
**Depends On:** Phase 2 complete (galaxy map exists)

## Purpose

This phase implements the **unique mechanic** that defines Stapledon's Voyage:

> When you commit to a journey, time passes differently for you vs. the galaxy.
> Civilizations rise and fall. Crew members live and die. Choices are irreversible.

This is where the game becomes more than a tech demo.

## Design Docs

| Doc | Description | Has Sprint? | Priority |
|-----|-------------|-------------|----------|
| [journey-system.md](journey-system.md) | Time dilation, commit decisions, transit events | NO | P0 |

## The Journey Experience

```
1. PLANNING         → Select destination, choose velocity
2. COMMIT           → Irreversible decision point (Pillar 1!)
3. DEPARTURE        → Leave current star
4. TRANSIT          → Journey Events mode (crew lifecycle)
5. APPROACH         → See destination changes
6. ARRIVAL          → New galaxy map state
```

## Key Features

### Time Dilation Calculator
```
You travel 10 light-years at 0.99c:
- Your time:    1.4 years
- Galaxy time: 10.1 years
- Lorentz factor: 7.1x
```

### Crew Projection
- Who will die during transit?
- Who might be born?
- How will relationships change?

### The Commit Button
- Multi-step confirmation
- Crew voting/concerns
- "This cannot be undone"

### Journey Events
- Deaths and births
- Philosophical debates
- System failures
- Relationship milestones

## Success Criteria

- [ ] Time dilation calculation correct
- [ ] Velocity slider works
- [ ] Crew projection displays
- [ ] Commit flow with confirmations
- [ ] Events generate during transit
- [ ] Arrival shows galaxy changes

## Dependencies

- **Depends on:** Phase 2 (galaxy-map for destination selection)
- **Blocks:** Phase 4 (arrival cinematics)
