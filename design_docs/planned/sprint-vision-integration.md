# Sprint Plan: Vision Integration

**Source:** Interview sessions 2025-12-06 (Game Loop Origin, Bubble Ship Design, AI Integration)

This sprint plan organizes design documents derived from vision interviews into implementable chunks.

---

## Sprint Overview

| Sprint | Theme | Docs | Dependencies |
|--------|-------|------|--------------|
| **Sprint 1** | Core Narrative | Black Hole, Opening Sequence | None |
| **Sprint 2** | Ship Foundation | Bubble Constraint, Ship Layout, Mass Budget | Sprint 1 |
| **Sprint 3** | Society Simulation | Bubble Society, Archive-Crew Trust | Sprint 2 |
| **Sprint 4** | AI Systems | Archive, Orchestrator, Spire Mystery | Sprint 3 |

---

## Sprint 1: Core Narrative Structure

**Goal:** Establish the meta-narrative framework (BH cycles, origin mystery)

| Doc | Purpose | Pillars |
|-----|---------|---------|
| [black-hole-mechanics.md](future/black-hole-mechanics.md) | Time weapon, endgame choice, New Game+ | 1, 3, 6 |
| [opening-sequence.md](future/opening-sequence.md) | Emergence from mysterious structure | 2, 5, 6 |

**Key Decisions:**
- BH entry = New Game+ (abandon universe, seed next)
- Player emerges FROM structure at start (mystery)
- "You are not the first" revelation

---

## Sprint 2: Bubble Ship Foundation

**Goal:** Define the physical constraints and spaces of the bubble ship

| Doc | Purpose | Pillars |
|-----|---------|---------|
| [bubble-constraint.md](future/bubble-constraint.md) | What crosses the boundary | 4, 5 |
| [bubble-ship-layout.md](future/bubble-ship-layout.md) | Physical spaces, emotional purposes | 3, 4 |
| [mass-budget.md](future/mass-budget.md) | Finite resources, meaningful choices | 1, 4 |

**Key Decisions:**
- Only information crosses (proto-tech via blueprints)
- Garden cathedral as emotional anchor
- Mass budget creates tension without busywork

---

## Sprint 3: Society Simulation

**Goal:** Make the bubble feel alive with emergent social dynamics

| Doc | Purpose | Pillars |
|-----|---------|---------|
| [bubble-society.md](future/bubble-society.md) | Living sim with generations, factions | 4, 6 |
| [archive-crew-trust.md](future/archive-crew-trust.md) | AI as NPC in social web | 4, 6 |

**Key Decisions:**
- Society is autonomous, player influences not controls
- Multi-generational with succession
- Archive has individual trust relationships like crew

---

## Sprint 4: AI Systems

**Goal:** Implement Archive NPC and invisible narrative orchestration

| Doc | Purpose | Pillars |
|-----|---------|---------|
| [archive-system.md](future/archive-system.md) | Personality, memory, repair mechanics | 4, 6 |
| [narrative-orchestrator.md](future/narrative-orchestrator.md) | Behind-the-scenes DM, arc types | 2, 3 |
| [spire-mystery.md](future/spire-mystery.md) | Tech tree reveals recursion clues | 5, 6 |

**Key Decisions:**
- Archive uses OCEAN, drifts with memory degradation
- Orchestrator is invisible to player
- Spire as universal constant = clue mechanism

---

## Implementation Order

```
Sprint 1 ─────────────────────────────────────────►
          Sprint 2 ───────────────────────────────►
                    Sprint 3 ─────────────────────►
                              Sprint 4 ───────────►
```

Sprints can overlap but each builds on previous foundations.

---

## Cross-References

- **Core Pillars:** [docs/vision/core-pillars.md](../../docs/vision/core-pillars.md)
- **Design Decisions:** [docs/vision/design-decisions.md](../../docs/vision/design-decisions.md)
- **Open Questions:** [docs/vision/open-questions.md](../../docs/vision/open-questions.md)
- **Interview Log:** [docs/vision/interview-log.md](../../docs/vision/interview-log.md)

## Input Documents

These design docs synthesize and formalize:
- [input/game_loop_origin.md](input/game_loop_origin.md)
- [input/bubble-ship-design.md](input/bubble-ship-design.md)
- [input/ai-the-archive.md](input/ai-the-archive.md)
