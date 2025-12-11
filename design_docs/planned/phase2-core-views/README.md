# Phase 2: Core Views (Navigation & Exploration)

**Priority:** P0
**Status:** Planned
**Depends On:** Phase 1 complete (data models exist)

## Purpose

This phase implements the primary interfaces players use:
- Galaxy map for strategic navigation
- Ship exploration for crew interaction
- Bridge interior for decisions

These are the "verbs" that operate on Phase 1's "nouns".

## Design Docs

| Doc | Description | Has Sprint? | Priority |
|-----|-------------|-------------|----------|
| [galaxy-map.md](galaxy-map.md) | Star system navigation, pan/zoom, selection | NO | P0 |
| [ship-exploration.md](ship-exploration.md) | Deck traversal, room interaction | NO | P1 |
| [02-bridge-interior.md](02-bridge-interior.md) | Bridge layout, crew stations, consoles | YES (0%) | P1 |

## Key Features

### Galaxy Map
- Pan and zoom across galaxy
- Star selection and info display
- Civilization status indicators
- Network edges showing contacts
- Journey preview (right-click star)

### Ship Exploration
- Deck-to-deck transitions
- Room navigation
- Crew location tracking
- Interactable objects

### Bridge Interior
- Console interaction
- Crew dialogue triggers
- Galaxy map access point
- Decision making UI

## Success Criteria

- [ ] Galaxy renders with 100+ stars
- [ ] Pan and zoom work smoothly
- [ ] Star selection shows info panel
- [ ] Ship deck transitions work
- [ ] Bridge consoles are interactable

## Dependencies

- **Depends on:** Phase 1 (starmap-data-model, ship-structure)
- **Blocks:** Phase 3 (journey system needs galaxy map)
