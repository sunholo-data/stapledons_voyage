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

---

## Should the black hole origin be explicit or implicit?

**Why it matters:** The game starts with the player emerging from a BH/mysterious structure. The question is whether players KNOW this from the start or discover it through play. This affects the entire narrative framing and sense of mystery.

**Current thinking:** Leaning toward mystery — the structure is ambiguous at first, players piece together the truth through gameplay. Discovery that shifts your worldview is valuable. But some clarity needed so players understand the New Game+ mechanic after first completion.

**Options:**
1. **Fully implicit:** Players discover through environmental storytelling, late-game revelations
2. **Semi-explicit:** Start screen hints at strangeness ("You don't remember entering"), truth revealed mid-game
3. **Explicit but mysterious:** Players know it's a BH, but the implications (other universes, cycle) are discovered
4. **Different per playthrough:** First run is mysterious, subsequent runs acknowledge the cycle

**Needs:**
- Narrative design for the opening sequence
- Decision on whether "Earth" is confirmed real or ambiguous
- Playtesting to see what discovery moments feel best

---

## How rare should universe-hopper encounters be?

**Why it matters:** Meeting another traveler from a different dead universe is the "ultimate encounter." Rarity affects whether this feels legendary or routine.

**Current thinking:** Should feel legendary — not guaranteed per playthrough. But the player IS this to every civ they meet, even if they don't know their own origin.

**Options:**
1. **Once per playthrough, maybe:** Late-game, earned by deep exploration
2. **Never guaranteed:** Some players never find one, creating community legends
3. **Only after multiple BH cycles:** Meta-progression unlocks possibility
4. **Implied, never confirmed:** Hints that someone might be like you, never proof

**Needs:**
- Decision on how much meta-progression exists across BH cycles
- Whether this encounter has mechanical implications or is purely narrative
- What such an encounter would actually look like in gameplay

---

## How much does prior-universe play influence the next universe?

**Why it matters:** BH entry seeds a new universe with weighted parameters. The degree of influence affects whether this feels like earned progression or mostly random.

**Current thinking:** "Mysterious but clearly beneficial" — players know it helps, not exactly how. Avoids optimization, preserves discovery.

**Options:**
1. **Light touch:** Small weights, mostly random — nudging probability
2. **Heavy hand:** Clear rewards — "You archived 12 civs, +2 Anthropic Luck"
3. **Mysterious:** Influence exists but isn't explained — patterns emerge over many runs
4. **Thematic:** What you carried in (philosophies, archives) shapes what you find

**Needs:**
- Decision on meta-save system (what persists across BH cycles?)
- Whether to show any "inheritance" at new game start
- Balance testing across multiple playthroughs
