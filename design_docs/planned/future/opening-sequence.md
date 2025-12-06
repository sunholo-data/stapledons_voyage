# Opening Sequence

## Status
- **Status:** Planned
- **Sprint:** Vision Integration - Sprint 1
- **Priority:** P1 (First player experience)
- **Source:** [Interview: Black Hole Deep Dive](../../../docs/vision/interview-log.md#2025-12-02-session-black-hole-feature-deep-dive)

## Game Vision Alignment

| Pillar | Score | Notes |
|--------|-------|-------|
| Choices Are Final | ⚪ N/A | Setup, not choice |
| Game Doesn't Judge | ✅ Strong | Origin is ambiguous, not labeled |
| Time Has Emotional Weight | ⚪ N/A | Establishes context for future weight |
| Ship Is Home | ✅ Strong | First introduction to bubble ship |
| Grounded Strangeness | ✅ Strong | Cosmic mystery grounded in physics |
| We Are Not Built For This | ✅ Strong | Disorientation is the point |

## Feature Overview

Every playthrough begins with the player **emerging from a mysterious structure**. This reframes everything:

- You are not humanity's first traveler - you are a universe-immigrant
- You carry archives of a dead cosmos; Earth may or may not be real
- To every civ you meet, YOU are the impossible alien from beyond
- The end-game BH entry continues a cycle you're already part of

## The Big Revelation

The player doesn't know at first, but:

> **"You are not the first."**

This truth is discovered over multiple playthroughs:
- Smart players might figure it out by end of first run
- Most discover it through patterns across runs
- Archive confusion about spire readings provides clues
- Visual storytelling at start/end of game hints at recursion

## Opening Beat Structure

### Beat 1: Emergence
- Visual: Ship emerges from swirling cosmic structure
- Player perspective: Disorientation, unfamiliar stars
- Archive: "Systems nominal... location unknown"
- Duration: ~30 seconds cinematic

### Beat 2: Orientation
- Archive attempts to locate position
- Star patterns don't match any database (or do they?)
- Earth signal detected (or is it a signal from the Archive?)
- Player gets first choice: investigate signal or explore

### Beat 3: The Question
- If player investigates Earth signal:
  - Journey to origin point
  - Discover Earth exists (in this universe)
  - Learn of impending doom (rogue BH)
- If player explores first:
  - Encounter alien civ that sees YOU as the impossible one
  - Eventually learn about Earth later

## Ambiguity Levels

From [open-questions.md](../../../docs/vision/open-questions.md#should-the-black-hole-origin-be-explicit-or-implicit):

| Level | What Player Knows | When |
|-------|-------------------|------|
| **Fully Implicit** | Structure is strange, no explanation | First playthrough |
| **Semi-Explicit** | "You don't remember entering" | Mid-game hints |
| **Explicit-but-Mysterious** | Know it's a BH, not the implications | Late-game |
| **Different per Run** | First run mysterious, later runs acknowledge cycle | NG+ |

**Recommended:** Start mysterious. Hints accumulate. Archive upgrade path may reveal more. Never fully explicit - preserve mystery.

## Visual Design

### The Structure
- Should look like:
  - A black hole's accretion disk (scientific)
  - Something ancient and strange (mysterious)
  - The same structure you enter at end-game (cyclical)
- Color palette: Deep blues, impossible blacks, hints of light from within
- Not labeled or explained

### The Emergence
- Ship appears from the structure's center
- Transition from darkness to starfield
- Disorienting camera movement (player doesn't know which way is "forward")
- Stars resolve into unfamiliar patterns

### Earth (If Visited)
- Recognizable but somehow strange
- Player's first "home" feels foreign
- Subtle wrongness (are these really your people?)

## Archive Behavior

The Archive's confusion about the spire is a **clue mechanism**:

| Archive State | Behavior | Clue Value |
|---------------|----------|------------|
| **Fresh Start** | "Systems nominal" - doesn't notice strangeness | None |
| **First Anomaly** | "Calibration error" - dismisses spire readings | Low |
| **Accumulated Data** | "Inconsistent with known physics" | Medium |
| **Upgraded/Repaired** | "These readings predate the universe" | High |
| **Alien Perspective** | "What you call errors are facts" | Revelation |

## Design Decisions

From [design-decisions.md](../../../docs/vision/design-decisions.md):

| Decision | Relevance |
|----------|-----------|
| BH Entry = New Game+ | Opening is the other side of that entry |
| Earth Fate Always Shown | Even if player delays, Earth story resolves |
| Recursion Revelation | Opening contains subtle clues |

## Narrative Implications

### For the Player
- Immediate mystery: Where am I? What is this ship?
- Growing unease: Why does Archive have errors?
- Slow realization: I've done this before (maybe)

### For Aliens
- To every civilization you meet, you are impossible
- Your ship defies their physics
- Your archives contain data from "before"
- This makes you valuable AND threatening

### For Earth
- Earth might be "real" in this universe
- Or might be a reconstructed memory
- The game never confirms which
- This ambiguity is intentional

## Open Questions

1. **How much disorientation?** - Should player feel lost, or quickly oriented?
2. **Archive voice immediately?** - Or delay introduction?
3. **Choice timing?** - How soon does player make first real choice?
4. **Skip option?** - Should repeat players be able to skip opening?

## AILANG Types

```ailang
type OpeningPhase =
    | Emergence
    | Orientation
    | FirstChoice
    | InProgress

type ArchiveAwarenessLevel =
    | Oblivious
    | Suspicious
    | Confused
    | Enlightened

type OpeningState = {
    phase: OpeningPhase,
    archive_awareness: ArchiveAwarenessLevel,
    earth_signal_detected: bool,
    player_location_known: bool
}
```

## Engine Integration

### Cinematic System
- Pre-rendered or real-time emergence sequence
- Camera control for disorientation effect
- Transition to gameplay seamlessly

### Audio
- Ambient space sounds (unfamiliar)
- Ship systems powering up
- Archive voice (first introduction)
- Optional: faint echoes/whispers from the structure

### UI
- Minimal during emergence
- Gradually reveal HUD elements
- Archive introduces UI components naturally

## Testing Scenarios

1. **Fresh Start:** Player sees emergence, no prior context, feels mysterious
2. **NG+ Start:** Same emergence, but Archive has subtle differences
3. **Skip Opening:** Verify repeat players can bypass if option exists
4. **First Choice:** Both paths (Earth vs. explore) lead to viable gameplay

## Success Criteria

- [ ] Opening creates sense of mystery and wonder
- [ ] Player feels disoriented but not frustrated
- [ ] Archive introduction feels natural
- [ ] First choice feels meaningful
- [ ] Visual connection to end-game BH entry is subtle but present
- [ ] Multiple playthroughs reveal new details
