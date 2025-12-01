# Design Decisions

Log of design decisions with context and rationale.

---

<!-- Template for new decisions:

## [YYYY-MM-DD] [Decision Title]

**Context:** [Why this came up]

**Decision:** [What was decided]

**Rationale:** [How this serves the pillars]

**Alternatives rejected:** [What else was considered]

**Implications:** [What this means for other features]

-->

## [2025-11-29] Core Pillars Established

**Decision:** 5 core pillars define design constraints: Choices Are Final, The Game Doesn't Judge, Time Has Emotional Weight, The Ship Is Home, Grounded Strangeness

**Rationale:** These pillars emerged from vision interview exploring emotional goals, player agency, and game identity. Every future feature must serve at least one pillar.

---

## [2025-11-30] OCEAN Personality System for Crew

**Context:** Designing crew archetypes (Engineer, Scientist, Medic, Diplomat, Pilot, Quartermaster, Zealot, Dreamer, Skeptic, Fantasist, Analyst) needed a grounded psychological framework to drive emergent behavior.

**Decision:** Use OCEAN (Big Five) personality model with the following parameters:
- **O** (Openness): 0.0-1.0
- **C** (Conscientiousness): 0.0-1.0
- **E** (Extraversion): 0.0-1.0
- **A** (Agreeableness): 0.0-1.0
- **N** (Neuroticism): 0.0-1.0

Each archetype has baseline OCEAN values that inform dialogue tone, event reactions, relationship chemistry, crew votes, and civilization preferences.

**Rationale:** OCEAN is real psychology (Pillar 5: Grounded Strangeness), creates emergent rather than scripted behavior (Pillar 2: Game Doesn't Judge), and makes crew feel like real people (Pillar 4: Ship Is Home).

**Alternatives rejected:**
- MBTI alone (too prescriptive, pop-psychology feel)
- Custom trait system (less grounded, reinventing the wheel)
- No personality system (crew would feel generic)

---

## [2025-11-30] Player OCEAN Is Emergent

**Context:** Should player's personality profile be declared upfront or inferred from gameplay?

**Decision:** Player OCEAN is emergent from choices, not declared. The game tracks decisions and infers player's psychological profile over time.

**Rationale:** Serves "The Game Doesn't Judge" - player discovers who they are through play rather than selecting a label. Creates self-discovery moment in legacy report. Avoids players gaming their profile for optimal crew chemistry.

**Alternatives rejected:**
- Quiz at start (feels like a personality test, not a game)
- Direct selection (too on-the-nose, breaks immersion)

**Implications:** Need to define which decisions map to which OCEAN dimensions. Legacy report can reveal "who you became."

---

## [2025-11-30] OCEAN Values Drift Over Time

**Context:** Crew spend 100 subjective years together through traumatic experiences. Should personality be fixed?

**Decision:** OCEAN values drift based on experiences. Core archetype provides baseline, but trauma, relationships, time dilation exposure, and witnessed events shift values within bounds.

**Rationale:** Serves "Time Has Emotional Weight" - the cosmos changes you. Creates character arcs without scripting them. The Engineer who watched 10,000 years of civilizations die might have their Openness cracked open.

**Alternatives rejected:**
- Fixed values (feels static over 100-year journey)
- Unbounded drift (would lose archetype identity)

**Implications:** Need drift triggers (events that shift values), drift bounds (how far from baseline), and visual/dialogue indicators of change.

---

## [2025-11-30] Directional Relationship Chemistry

**Context:** Should relationship chemistry be symmetric (A↔B same) or directional (A→B ≠ B→A)?

**Decision:** Chemistry is directional. Each crew member's feelings toward another are calculated separately based on what they need vs. what the other offers.

**Rationale:** Real relationships are asymmetric. Dreamer might adore Medic while Medic finds Dreamer exhausting. Creates richer dynamics: unrequited bonds, one-sided tensions, mentor/protégé asymmetries.

**Implications:**
- Need "needs" and "offers" profiles per archetype
- Relationship UI might show both directions
- Crew conflicts can arise from asymmetric investment

---

## [2025-11-30] Needs/Offers Derived from OCEAN

**Context:** Should psychological needs/offers be manually defined per archetype, or derived algorithmically from OCEAN values?

**Decision:** Derive needs and offers from OCEAN values using formulas. Simpler to maintain and ensures consistency.

**Derivation rules:**
- `needs.validation` = N × (1 - C) — anxious + impulsive people seek reassurance
- `needs.stability` = N × C — anxious + conscientious people want predictability
- `needs.stimulation` = O × E — open + extraverted people crave novelty
- `needs.autonomy` = (1 - A) × (1 - E) — disagreeable introverts want space
- `needs.connection` = E × A — extraverted + agreeable people seek bonds

- `offers.validation` = A × E — agreeable extraverts give reassurance
- `offers.stability` = C × (1 - N) — conscientious + calm people provide stability
- `offers.stimulation` = O × E — open extraverts bring energy
- `offers.autonomy` = (1 - C) × (1 - A) — low C + low A leave others alone
- `offers.connection` = E × A — same as needs (they give what they seek)

**Rationale:** Single source of truth (OCEAN), no manual sync needed, mathematically grounded.

---

## [2025-11-30] OCEAN Drift Hidden Until End Screen

**Context:** How should players perceive crew personality changes over time?

**Decision:** Drift is invisible during gameplay. Players experience it through subtle dialogue shifts and behavior changes, but never see numbers. The end screen reveals "who your crew became" — showing before/after OCEAN profiles and key drift events.

**Rationale:**
- Serves "The Game Doesn't Judge" — no mid-game optimization of crew psychology
- Creates discovery moment at end — "I didn't realize the Engineer had changed so much"
- Dialogue shifts feel organic, not mechanical
- Avoids UI clutter during play

**Implications:**
- Dialogue generation must reflect current (not baseline) OCEAN
- End screen needs "crew evolution" section
- May want to show key moments that caused drift ("After witnessing the Helix extinction...")

