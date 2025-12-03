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


## [2025-12-02] Trace Metal Scarcity Model

**Decision:** Trace metals available in every solar system. Only becomes a threat after extended periods without visiting any system.

**Rationale:** Avoids constant micromanagement pressure. Creates consequence for isolation/long journeys rather than routine busywork.


## [2025-12-02] Earth Return: Optional but Compelling

**Decision:** Return to Earth is never mechanically required, but should be narratively very interesting.

**Rationale:** Serves Pillar 2 (no judgment) - game doesn't force you. Serves Pillar 3 (time weight) - what has Earth become after millennia?


## [2025-12-02] Crew Opinions: Universal System

**Decision:** Crew opinions apply to all major decisions, not just resource harvesting. This is a general system.

**Rationale:** Serves Pillar 4 (Ship Is Home) - crew are opinionated people, not silent resources. Consistency across all decision types.


## [2025-12-02] Alien Biosphere: Science Not Tropes

**Decision:** Alien biosphere risks follow real science (biochemical incompatibility, chirality, contamination ethics) rather than sci-fi tropes (instant plagues, adapting viruses). The tension is 'can we use it at all' and 'should we risk contaminating them', not 'will it kill us'.

**Rationale:** Serves Pillar 5 (Grounded Strangeness). Real biology is more interesting than Hollywood scenarios.


## [2025-12-02] Anthropic Luck Factor

**Decision:** Galaxy has tunable 'Anthropic Luck' (L) parameter. Higher L = denser civs in player's region. Justified as observer-selection bias: conditional on being a player who CAN meet civs, you're in the rare branch where they exist nearby. Default: L high enough for 5-15 civs within 500-1000 ly.

**Rationale:** Pillar 5 (Grounded Strangeness): Real physics with philosophical justification. Avoids both 'magic dense universe' and 'boring empty universe'. Player can adjust for hardcore vs narrative experience.


## [2025-12-02] Detection Model: Epistemic Gap

**Decision:** Players only see old light. Three detection tiers: (1) Astrometric survey - star type, planets, orbits; (2) Spectroscopic - biosignatures, atmosphere; (3) Technosignature search - radio/waste heat. All data labeled with 'last_light_year' showing how stale it is. Arrival reveals how wrong predictions were.

**Rationale:** Pillar 1 (Choices Final): Commit based on uncertain data. Pillar 3 (Time Weight): The gap between 'what you saw' and 'what you find' IS the time dilation made tangible.


## [2025-12-02] Gamma Cap: 10-20 Default

**Decision:** Player ship γ (Lorentz factor) capped around 10-20. At γ=20: 300 ly = 15 years subjective / 300 years external. This allows 4-6 major legs in 100 years, with 2-3 revisits to nearby civs. Sweet spot for meaningful sacrifice without making civs unreachable.

**Rationale:** Pillar 3 (Time Weight): Each journey costs centuries external. Not so high that everything is one-way, not so low that time feels cheap.


## [2025-12-02] Launch Order = Arrival Order

**Decision:** With Higgs bubble, any ship can achieve arbitrary γ instantly. Therefore arrival order depends ONLY on launch time in galaxy frame, not on technology level. Uplifted civs can reach destinations before you IF they launch earlier (while you detour or delay). They cannot overtake you mid-flight.

**Rationale:** Pillar 5 (Grounded): Clean physics rule with no paradoxes. Creates strategic tension: share tech and they might beat you to Earth.


## [2025-12-02] Player Role: Relativistic Perturbation Operator

**Decision:** Player is not an explorer or diplomat - they are a network perturbation operator. Actions: (1) Give civ technology (Higgs, fusion, etc); (2) Give civ maps of others; (3) Share philosophies between civs. Everything else is fallout. The sim is 'how does the contact graph evolve when I poke it?'

**Rationale:** Pillar 2 (No Judgment): You perturb, consequences cascade, game shows results without moral labels. Pillar 1 (Choices Final): Each poke is irreversible and propagates.


## [2025-12-02] Three Distance Regimes

**Decision:** Local bubble (≤500 ly): Many biospheres, maybe 1 civ, centuries pass per trip - 'serial saga' territory. Mid-range (1000-3000 ly): Good civ probability, millennia pass, 1-2 visits max - 'two bookends' territory. Deep-galaxy (≥5000 ly): Almost certainly find civs, but tens of thousands of years pass - 'one-shot mythic meeting' territory.

**Rationale:** Pillar 3 (Time Weight): Distance correlates with sacrifice. Pillar 1 (Choices Final): Deeper journeys are more irreversible.


## [2025-12-02] Anthropic Luck: World-Gen Only

**Decision:** Anthropic Luck (L) is set once during world generation, fixed for entire playthrough. Player chooses density at start, then lives with that universe.

**Rationale:** Pillar 1 (Choices Final): Even the universe itself is a choice you commit to.


## [2025-12-02] No Survey Budget: Auto-Discovery

**Decision:** No telescope time management or survey budget mechanic. Interesting destinations are automatically discovered/revealed. Remove friction that doesn't serve the core experience.

**Rationale:** Avoids tedious micromanagement. The interesting decisions are WHERE to go and WHAT to share, not HOW MUCH to scan.


## [2025-12-02] Earth Contact: Emergent Not Scripted

**Decision:** Civilizations reaching Earth before you is a natural emergent consequence of sharing Higgs tech, not a scripted event. If you give them the drive and maps, physics determines they CAN reach Earth. Consequences follow naturally from simulation.

**Rationale:** Pillar 2 (No Judgment): Game doesn't script drama, just simulates physics. Pillar 1 (Choices Final): You chose to share tech, consequences emerge.


## [2025-12-02] Wanderer Mythology: One Consequence Among Many

**Decision:** Civilizations forming religions/mythologies around the player is one possible emergent consequence, not special-cased. Could equally get: resentment, cargo cults, scientific study, dismissal as legend, active hunting, worship, etc. No predetermined 'canonical' reaction.

**Rationale:** Pillar 2 (No Judgment): All reactions are equally valid outcomes. Pillar 5 (Grounded Strangeness): Different civ philosophies produce different responses.


## [2025-12-02] Black Hole Entry = New Game+ Mechanism

**Decision:** Entering a black hole ends the current playthrough and seeds a new universe with weighted parameters influenced by prior play. This is THE replayability mechanism - like Civ 5 re-rolling for better starts, but earned through completion. Parameters affected include Anthropic Luck, γ cap, and other universal constants. Influence is mysterious but clearly beneficial - players know it helps but not exactly how.

**Rationale:** Serves Pillar 1 (Choices Final): You cannot undo your universe, only leave it. Serves Pillar 3 (Time Weight): The ultimate sacrifice - abandoning everyone and everything. References Asimov's 'The Last Question' - heat death followed by new creation.


## [2025-12-02] Game Starts Post-Black-Hole

**Decision:** Every playthrough begins with the player emerging from a black hole or mysterious structure (later revealed to be a BH). The player is already a universe-immigrant carrying archives of a dead cosmos. Earth may or may not be real - it exists in your archives but you have no proof. This reframes the entire game: you are not humanity's first traveler, you are a cycle-continuer.

**Rationale:** Serves Pillar 5 (Grounded Strangeness): Physics-plausible multiverse via Smolin's cosmological natural selection. Creates profound mystery and recontextualizes player's role. To every civilization you meet, YOU are the impossible alien from beyond.


## [2025-12-02] Human Incompatibility Theme

**Decision:** Elevate 'humans are not built for relativistic space travel' to a core thematic pillar. Madness, psychological breakdown, and the human condition failing under cosmic scales is intentional. Crew going insane near black holes, mutinies, OCEAN drift toward instability - these are features, not bugs. Victories are DESPITE human frailty, not because humans are special.

**Rationale:** Serves Pillar 3 (Time Weight): The emotional cost is partly that we cannot handle it. Serves Pillar 4 (Ship Is Home): Crew psychology grounds the cosmic horror in human-scale experience. This is honest about the real challenges of relativistic travel.


## [2025-12-02] Fermi Answer: Temporal Fragmentation

**Decision:** Time dilation provides a physics-grounded answer to the Fermi Paradox: even if intelligent life is common, relativistic travel makes contact temporally impossible. By the time you reach anyone, they're dead. By the time they reach you, you're dust. The galaxy isn't empty - it's temporally fragmented. This is discoverable in-game as a late realization. Civilizations that understand this might rationally choose to stay home - Fermi silence as grief avoidance, not hostility or indifference.

**Rationale:** Serves Pillar 5 (Grounded Strangeness): Real physics consequence, not hand-waved. Serves Pillar 2 (No Judgment): Neither optimistic nor pessimistic - just the physics playing out.


## [2025-12-02] Universe-Hopper Encounter Possibility

**Decision:** Rare possibility exists of meeting another traveler who emerged from a DIFFERENT dead universe. They carry archives of a cosmos with different physics, different histories. This is the ultimate encounter - two archives meeting, two universes' memories intersecting. Rarity undecided but should feel legendary. The player IS this to every civ they meet - they just may not know their own origin story.

**Rationale:** Serves Pillar 5 (Grounded Strangeness): Grounded in multiverse physics. Creates the 'ultimate archive' narrative. The most precious thing in existence: proof that other universes existed.


## [2025-12-02] BH Time-Skip Costs

**Decision:** Using a black hole for extreme time dilation (parking near horizon to skip millennia) has natural costs: all civilization relationships reset (they're dead), time-sensitive data becomes worthless, crew may suffer psychological damage or madness, potential mutiny if crew refuse. The cost is emotional severance from the present - everything you knew is gone. Some data survives: BH locations, galactic structure, physical constants, and YOUR ARCHIVES become priceless as the only record of extinct cultures.

**Rationale:** Serves Pillar 3 (Time Weight): Maximum sacrifice for maximum skip. Serves Pillar 4 (Ship Is Home): Crew psychology constrains player choice.


## [2025-12-02] BH Approach Can Trigger Mutiny

**Decision:** Crew can mutiny to prevent black hole approach if psychological stress is too high. Skeptic and Medic archetypes likely to resist; Zealot may be dangerously eager. If mutiny succeeds: crew seize ship, change direction, or force compromise (shallower orbit, less extreme skip). This is one of few moments where Pillar 4 (Ship Is Home) directly conflicts with player agency - the human-scale asserts itself against cosmic-scale decisions.

**Rationale:** Serves Pillar 4 (Ship Is Home): Crew are people with limits, not tools. Serves Human Incompatibility Theme: Sometimes the healthy response is to refuse.


## [2025-12-02] BH Crossing Causes Memory Loss

**Decision:** Crossing a black hole event horizon causes memory loss - hand-waved as 'causality reversal affecting the brain' or similar physics-adjacent explanation. This justifies the mysterious start: players emerge with archives but no memory of how they got there. Earth exists in records but cannot be confirmed as real. Creates narrative permission for the cycle mystery without breaking physics too badly.

**Rationale:** Serves mystery-discovery design goal. Provides in-universe justification for New Game+ memory reset. Keeps the 'where did I come from?' question open.


## [2025-12-03] Special Relativity Visual Effects: Hard SF Made Visible

**Context:** Relativistic travel is central to the game, but players need to SEE it, not just read numbers. The question was how to make SR tangible without resorting to fake "warp streaks."

**Decision:** Implement three physically accurate SR effects at high γ (10-20):

1. **Aberration (Headlight Effect):** Stars pile into a forward cone as speed increases. At high γ, almost everything is visible in a narrow tunnel ahead; behind is blackness. The universe "rushes toward" your velocity vector.

2. **Doppler Shift:** Light from ahead blue-shifts (stars become blue/white). Light from behind red-shifts (fades to infrared and disappears). Nebulae get blue tint ahead, red smear behind.

3. **Relativistic Beaming:** Intensity scales as I' ∝ D³ where D is Doppler factor. Forward directions become much brighter (blow out into "star wind" halo). Rear directions dim to near-black.

Additionally: External clocks (remote beacons, planet rotations) run "too fast" as you approach high speed. HUD shows "galaxy time" vs "ship time" advancing at different rates.

**Implementation boundary:**
- **AILANG outputs:** camera_pos, camera_vel (β vector), gamma, starfield in galaxy frame
- **Engine implements:** Aberration via direction transform, Doppler via D = γ(1 − β·n), beaming via I' ∝ D³

The sim says "how fast am I going"; the renderer bends the light.

**Core math (engine-side):**
- Direction transform: Split into parallel/perpendicular components to β
- Doppler factor: D = γ(1 − β·n) for each star direction n
- Beaming: brightness_factor = clamp(D³, min, max)
- Implementation: CPU transform for discrete stars, shader for background cubemap

**Rationale:** Serves Pillar 3 (Time Weight): Visual distortion makes time dilation visceral, not abstract. Serves Pillar 5 (Grounded Strangeness): Real physics that's actually visible. Serves Pillar 6 (Human Incompatibility): Disorienting visuals emphasize cosmic alienation—this isn't how humans evolved to see.

**Alternatives rejected:**
- Fake "warp streaks" (not physically accurate)
- Abstract representations (loses visceral impact)
- Ignoring SR effects (misses the core experience)

**Implications:** Need to define visual exaggeration curves (map physical β → visual "wow"). May want to add lens flare, motion blur as aesthetic overlays. GR lensing near black holes is a separate effect.

