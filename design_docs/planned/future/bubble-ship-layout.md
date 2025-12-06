# Bubble Ship Layout

## Status
- **Status:** Planned
- **Sprint:** Vision Integration - Sprint 2
- **Priority:** P2 (Spaces serve gameplay, not critical path)
- **Source:** [Interview: Bubble Ship Design](../../../docs/vision/interview-log.md#2025-12-06-session-bubble-ship-design-integration)

## Game Vision Alignment

| Pillar | Score | Notes |
|--------|-------|-------|
| Choices Are Final | ⚪ N/A | Layout is fixed |
| Game Doesn't Judge | ⚪ N/A | Spaces are neutral |
| Time Has Emotional Weight | ✅ Strong | Garden cathedral for processing |
| Ship Is Home | ✅ Strong | Defines "home" spaces |
| Grounded Strangeness | ✅ Strong | Physics-consistent design |
| We Are Not Built For This | ✅ Strong | Spaces show human adaptation |

## Feature Overview

The bubble ship is a **100-meter radius sphere** with nested functional layers. Each space serves both practical and emotional purposes.

**Key Principle:** Focus on meaningful choices, not micromanagement. Ship is home, not a survival puzzle.

## Physical Structure

From [input/bubble-ship-design.md](../input/bubble-ship-design.md):

```
           OBSERVATION DECK (top)
                  │
            ┌─────┴─────┐
            │   BRIDGE  │ ◄── Decision Hub
            └─────┬─────┘
                  │
    ┌─────────────┼─────────────┐
    │      CREW QUARTERS        │ ◄── Living Space
    │   (Residential Ring)      │
    └─────────────┬─────────────┘
                  │
    ┌─────────────┼─────────────┐
    │     GARDEN CATHEDRAL      │ ◄── Emotional Anchor
    │   (Outer Shell Gardens)   │
    └─────────────┬─────────────┘
                  │
    ┌─────────────┼─────────────┐
    │       ARCHIVE CORE        │ ◄── AI Shrine
    │   (Data Systems Hub)      │
    └─────────────┬─────────────┘
                  │
    ┌─────────────┼─────────────┐
    │      ENGINEERING          │ ◄── Background Access
    │   (Life Support, Power)   │
    └─────────────┬─────────────┘
                  │
            ┌─────┴─────┐
            │   SPIRE   │ ◄── Mystery Zone
            │ (Higgs Gen)│
            └───────────┘
               (center)
```

## Location Details

### Observation Deck / Bridge

**Function:** Primary decision hub, command center

**Emotional Purpose:** Cosmic backdrop for major choices - starfields, SR/GR effects visible through transparent dome

**Gameplay:**
- Most journey decisions made here
- Crew briefings and discussions
- Viewing external universe
- Captain's authority most visible here

**Design Decision:** [Observation Deck as Decision Hub](../../../docs/vision/design-decisions.md)

---

### Crew Quarters (Residential Ring)

**Function:** Living spaces, private quarters, communal areas

**Emotional Purpose:** Normal human life amid cosmic voyage

**Gameplay:**
- Crew relationship building
- Faction meeting spots
- Personal conversations
- Where crew "live" between events

**Note:** No micromanagement of food, sleep, toilet. Focus on relationships, not survival meters.

---

### Garden Cathedral

**Function:** Hydroponic gardens, oxygen generation, food production

**Emotional Purpose:** "Sad but happy" - bittersweet remembrance of Earth

**Gameplay:**
- Where crew remember what it meant to live on a planet
- Cultural rituals develop here over generations
- Philosophical conversations and processing
- Bubble society's culture crystallizes here
- Memorial space for lost crew/Earth

**Design Decision:** [Garden Cathedral](../../../docs/vision/design-decisions.md)

**Visual:** Greenery against the starfield. Living things in dead space.

---

### Archive Core

**Function:** AI systems, data storage, computation center

**Emotional Purpose:** Shrine to knowledge, connection to the spire mystery

**Gameplay:**
- Special dialogues with Archive happen here
- Archive upgrades and repairs performed here
- Key revelations about spire/recursion
- Can visit via terminal anywhere, but core room has significance

**Design Decision:** [Archive: Distributed and Localized](../../../docs/vision/design-decisions.md)

---

### Engineering Deck

**Function:** Power generation, life support, fabrication

**Emotional Purpose:** The "guts" - necessary but unglamorous

**Gameplay:**
- Visit during crises or for upgrades
- Not a primary gameplay space
- "Necessary but boring most of the time"
- Where proto-tech is fabricated from mass budget

**Design Decision:** [Engineering Deck: Background Access](../../../docs/vision/design-decisions.md)

---

### The Spire (Higgs Generator)

**Function:** Creates and maintains the Higgs bubble

**Emotional Purpose:** Ultimate mystery - source of clues about recursion

**Gameplay:**
- Forbidden zone - crew cannot fully access
- Archive interfaces with it, produces "confused" readings
- Tech tree progression may reveal more
- May be constant across all universes

**Design Decision:** [The Spire as Universal Constant](../../../docs/vision/design-decisions.md)

## Player Location System

**Design Decision:** [Player Location Freedom](../../../docs/vision/design-decisions.md)

- Player chooses where to spend time
- No micromanagement required
- Ship is large enough that locations feel distinct
- Different events/conversations happen in different spaces

### Location Selection UI

```
┌─────────────────────────────────┐
│ Where would you like to go?     │
│                                 │
│ ◉ Observation Deck  [Current]   │
│ ○ Crew Quarters                 │
│ ○ Garden Cathedral              │
│ ○ Archive Core                  │
│ ○ Engineering                   │
│                                 │
│ [Go]                            │
└─────────────────────────────────┘
```

### Location-Specific Events

| Location | Event Types |
|----------|-------------|
| **Observation** | Journey decisions, cosmic events, official meetings |
| **Quarters** | Personal crew conversations, faction politics |
| **Garden** | Cultural events, memorials, philosophical talks |
| **Archive** | AI dialogues, spire readings, data analysis |
| **Engineering** | Crises, repairs, fabrication choices |

## Visual Design Guidelines

### Observation Deck
- Panoramic view of space
- Minimal interior - focus on exterior
- Captain's chair as focal point
- Holographic displays for navigation

### Crew Quarters
- Warm lighting, personal effects
- Evidence of lived-in space
- Cultural items accumulate over generations
- Mix of private and communal areas

### Garden Cathedral
- Green amid the metal
- Light filtered through canopy
- Water features if mass allows
- Memorial wall/space for Earth

### Archive Core
- Cool blue lighting
- Data visualization surfaces
- Central interface for Archive
- Subtle connection to spire below

### Engineering
- Industrial, functional
- Status displays for systems
- Fabrication area visible
- Less aesthetically designed

### Spire Access
- Limited visibility
- Strange geometry hints
- "Wrong" angles or perspectives
- Archive confusion visible in readings

## AILANG Types

```ailang
type ShipLocation =
    | ObservationDeck
    | CrewQuarters(room_id: int)
    | GardenCathedral
    | ArchiveCore
    | Engineering
    | SpireAccess  -- Limited, late-game

type LocationEvent =
    | CrewConversation(crew_id: int)
    | FactionMeeting(faction_id: int)
    | SystemCrisis(crisis: CrisisType)
    | JourneyDecision(decision: JourneyChoice)
    | ArchiveDialogue(topic: string)
    | Ritual(ritual_type: RitualType)

type PlayerLocation = {
    current: ShipLocation,
    time_here: float,  -- How long at this location
    events_available: [LocationEvent]
}
```

## Engine Integration

### Navigation
- Location selection UI
- Transition animations between spaces
- Distinct visual themes per location

### Events
- Location-aware event triggers
- Crew movement between locations
- Background activities per location

### Audio
- Ambient soundscapes per location
- Location-specific music themes
- Crew activity sounds

## Testing Scenarios

1. **Location Change:** Move between all locations, verify transitions
2. **Location Events:** Verify correct events trigger per location
3. **Long Stay:** Stay in one location, observe time passage effects
4. **Crisis Location:** Crisis in Engineering, verify player drawn there

## Success Criteria

- [ ] Each location feels distinct visually and emotionally
- [ ] Player understands what happens where
- [ ] No micromanagement of basic needs
- [ ] Garden cathedral creates intended emotional response
- [ ] Archive core feels significant for key dialogues
- [ ] Spire access feels mysterious and limited
