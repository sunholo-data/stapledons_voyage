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

---

## How much should SR visual effects be exaggerated?

**Why it matters:** Real SR effects at γ=20 are dramatic, but may need adjustment for readability, aesthetic appeal, or to avoid disorientation that breaks gameplay.

**Current thinking:** Some exaggeration for readability is fine, but grounded in real physics. The question is where to draw the line between "accurate" and "playable."

**Options:**
1. **Pure physics:** D³ beaming, exact aberration angles — may be too extreme
2. **Clamped physics:** Apply formulas but clamp outputs to comfortable ranges
3. **Aesthetic mapping:** Map physical β → visual "wow factor" via custom curves
4. **Mode toggle:** "Hard mode" uses real physics, "cinematic mode" softens effects

**Sub-questions:**
- Should rear-view ever go completely black, or maintain some faint visibility?
- How aggressive should Doppler colour shifting be? (Full spectrum shift vs. tinting)
- Should beaming brightness be tied to actual display HDR capabilities?
- Motion blur: add as aesthetic overlay or derive from relativistic effects?

**Needs:**
- Prototyping to see what feels right
- Player testing for disorientation tolerance
- Decision on whether different camera modes (docked vs cruise) have different effects

---

## What visual mode for docked/orbital vs cruise?

**Why it matters:** SR effects only make sense during relativistic cruise. Need to define when effects activate and deactivate.

**Current thinking:** AILANG outputs a "camera mode" flag. Docked/orbital uses normal rendering; cruise enables SR shader pipeline.

**Options:**
1. **Binary switch:** Effects on above threshold γ, off below
2. **Smooth transition:** Effects fade in/out as γ increases/decreases
3. **Mode-based:** Docked=normal, cruise=SR, near-BH=GR overlay

**Needs:**
- Definition of γ threshold for effect activation
- Transition duration/smoothness
- Whether effects should be slightly visible even at low speeds (subtle educational element)

---

## How should GR lensing near black holes interact with SR effects?

**Why it matters:** At high speed near a black hole, both SR and GR effects apply. Need to decide if/how they combine or if one dominates.

**Current thinking:** GR lensing is a separate effect, applied additively or as a post-process. Near-BH is already a special state (potential mutiny, time-skip decisions).

**Options:**
1. **Separate effects:** SR pipeline → GR post-process
2. **Unified physics:** Single shader that handles both (complex)
3. **Mutually exclusive:** Near-BH disables cruise mode, uses GR-only rendering
4. **Aesthetic priority:** Let visual impact guide which dominates

**Needs:**
- Research on how SR+GR combine in practice
- Decision on BH approach being a separate "scene" or continuous with cruise
