# Open Questions

Unresolved design questions that need exploration.

---

<!-- Template for new questions:

## [Question]

**Why it matters:** [Impact on design]

**Current thinking:** [Where we're leaning]

**Needs:** [What would help decide]

-->

## How should player actions map to OCEAN traits?

**Why it matters:** Player OCEAN is emergent (per design decision), so we need rules for inferring personality from gameplay choices. This affects end-screen "who you became" reveal and potentially crew chemistry calculations.

**Current thinking:** Define broad categories, let specifics emerge from AI dialogue or be tuned per-conversation:

| Decision Category | Affects Traits | Examples |
|-------------------|---------------|----------|
| Journey risk choices | C, N | Speed selection, destination danger |
| Civilization interaction | A, O | Help vs. ignore, trade vs. hoard |
| Crew relationship | E, A | Listen to objections, social time |
| Exploration style | O, E | Seek strange aliens, revisit known |
| Resource management | C, A | Share freely, stockpile |

**Needs:**
- Playtesting to see which actions feel meaningful
- AI dialogue system design (may handle inference contextually)
- Decision on whether player OCEAN affects crew chemistry mid-game or only end-screen
