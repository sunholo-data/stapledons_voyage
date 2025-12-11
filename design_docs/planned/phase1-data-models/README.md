# Phase 1: Data Models (Foundation)

**Priority:** P0
**Status:** Planned
**Depends On:** Phase 0 complete

## Purpose

This phase defines the core data structures ("nouns") of the game:
- Galaxy structure (stars, systems, distances)
- Planet properties (orbital mechanics, characteristics)
- Ship layout (decks, rooms, transitions)

Without these, you can't build:
- Galaxy map (no star data)
- Journey planning (no distances)
- Ship exploration (no room layouts)

## Design Docs

| Doc | Description | Has Sprint? | Priority |
|-----|-------------|-------------|----------|
| [starmap-data-model.md](starmap-data-model.md) | Galaxy structure, star systems, civilizations | NO | P0 |
| [planet-data-migration.md](planet-data-migration.md) | Planet properties, orbital data | NO | P1 |
| [ship-structure.md](ship-structure.md) | Deck layouts, room definitions | NO | P1 |

## AILANG Types to Define

```ailang
-- sim/galaxy.ail
type Galaxy = { stars: [Star], edges: [ContactEdge], currentPosition: StarID }
type Star = { id: StarID, name: string, x: float, y: float, civilization: Option[CivID] }
type CivState = Unknown | Thriving(int) | Declining(int) | Extinct(int) | Transcended(int)

-- sim/ship.ail
type Ship = { decks: [Deck], currentDeck: DeckID }
type Deck = { id: DeckID, name: string, rooms: [Room] }
type Room = { id: RoomID, tiles: [Tile], connections: [Connection] }
```

## Success Criteria

- [ ] Galaxy data types defined in AILANG
- [ ] Star distances calculable
- [ ] Ship deck/room types defined
- [ ] Planet orbital mechanics types defined

## Dependencies

- **Depends on:** Phase 0 (architecture clean)
- **Blocks:** Phase 2 (galaxy map, ship exploration)
