# Interview Log

Q&A session history for game vision development.

---

## 2025-11-29 Session: Core Pillars Extraction

**Focus:** Establishing the foundational design constraints for Stapledon's Voyage

### Q: When a player finishes and sees their legacy report, what should they feel?

**A:** All possible feelings depending on how the game went - triumph, melancholy, awe, moral weight. The emotional outcome is emergent from play, not prescribed.

### Q: Should the game ever tell players they did something 'wrong'?

**A:** No. The player judges. The game presents consequences. A philosophy game shouldn't make judgments - what's victory at one scale may be failure at another. The main lesson is that judgments are a matter of perspective.

### Q: How important is it that players can't undo their choices?

**A:** Essential. Players must live with decisions. Start a new game to try something else. This creates replayability.

### Q: What makes time dilation feel meaningful, not just a number?

**A:** The loneliness. Realizing how utterly inhuman traveling at speed and experiencing time dilation will be. Everyone you know dies or becomes indifferent. It should make players treasure their real-life moments with loved ones. The ship crew becomes precious as the one constant.

### Q: What's the player's relationship with their crew?

**A:** Player is protagonist/captain for agency. Crew offers opinions, represents inner feelings, can change based on events. Possible mutiny if things go badly. They help represent the inner feelings versus the outer space time-dilated experiences. Inner space vs outer space.

### Q: How alien should the aliens be?

**A:** As realistic as possible. Start from physically possible but maximally diverse starting points (philosophy, biology, astrophysics). Modular so more can be added over time. Community-expandable. This will be a rich source of content.

### Q: What's the irreducible core for MVP?

**A:** Time dilation loop with 100 years. Philosophical choices with multiple endings. Crew dialogue that guides and provokes thought. Limited but realistic species count (Drake equation). Year 1,000,000 fast-forward is great but could phase in.

### Insights

- The game is a philosophy prompt, not a philosophy lecture
- Time dilation is primarily an emotional mechanic, not a puzzle
- Ship crew serves as emotional anchor and moral sounding board
- Replayability comes from permanence, not saves
- Selfishness vs altruism; personal vs planetary vs galactic scale

### Actions

- [x] Establish 5 core pillars in docs/vision/core-pillars.md
- [ ] Create first design doc that references pillars
- [x] Conduct follow-up interview on specific mechanics

---

## 2025-12-02 Session: Black Hole Feature Deep Dive

**Focus:** Exploring black holes as game mechanic, narrative device, and replayability system

### Q: Does "park near a BH to skip millions of years" conflict with "Time Has Emotional Weight"?

**A:** The sacrifice IS the emotional weight — you lose all connection to the present. Everyone you knew is dead. Your data is worthless. Some crew would go insane from this. The cost is emotional severance, not a resource meter.

### Q: What happens to the crew when approaching a black hole?

**A:** Crew psychology is central. This is peak stress — watching external time race ahead, choosing to let millennia pass. Some archetypes resist (Skeptic, Medic), others embrace it dangerously (Zealot). Crew can mutiny to prevent or force compromise on BH approaches. This is one of the few moments where the human-scale directly conflicts with player agency.

### Q: Does BH entry forfeit the legacy report, or is it a victory path?

**A:** This led to a major reframe: BH entry IS the New Game+ mechanism. Like Civ 5 re-rolling for better starts, but earned through completion. You don't escape consequences — you abandon your universe entirely. The new universe's parameters are influenced by what you carried in, mysteriously beneficial but not explicit.

### Q: Should influence on next universe be transparent?

**A:** "Mysterious but clearly beneficial" — players know it helps but not exactly how. Avoids optimization, preserves discovery through multiple playthroughs.

### Q: What about the "Last Star" ending — witnessing heat death?

**A:** Referenced Asimov's "The Last Question." This could be the ultimate Stapledonian ending — but also led to the biggest insight of the session...

### Major Insight: The Game Starts Post-Black-Hole

The conversation revealed that every playthrough should begin with the player EMERGING from a BH/mysterious structure. This reframes everything:

- You are not humanity's first traveler — you are a universe-immigrant
- You carry archives of a dead cosmos; Earth may or may not be real
- To every civ you meet, YOU are the impossible alien from beyond
- The end-game BH entry continues a cycle you're already part of
- Meeting another universe-hopper would be the ultimate encounter — two archives meeting

### On Human Incompatibility

**A:** "Madness and the human condition failing is one of the themes of the game — how incompatible we are to space travel really." This was elevated to Pillar 6: "We Are Not Built For This."

### On the Fermi Paradox

**A:** Time dilation alone explains Fermi silence. The galaxy isn't empty — it's temporally fragmented. Civilizations that understand this might rationally stay home. Fermi silence as grief avoidance, not hostility.

### Insights

- Black holes serve three functions: time weapon, endgame choice, replayability mechanism
- The cycle (BH → new universe → BH) is the meta-structure of the game
- "Discovery that shifts your worldview, even briefly" is a design goal
- Death can be kindness in some scenarios — the game doesn't judge this
- Archives become priceless after deep-time skips — you ARE the record

### Actions

- [x] Log 7 design decisions from this interview
- [x] Add 3 open questions (BH origin explicit/implicit, universe-hopper rarity, influence transparency)
- [x] Add Pillar 6: "We Are Not Built For This"
- [x] Draft Black Hole Design Document → [black-hole-mechanics.md](../../design_docs/planned/future/black-hole-mechanics.md)
- [x] Design opening sequence → [opening-sequence.md](../../design_docs/planned/future/opening-sequence.md)

---

## 2025-12-06 Session: Game Loop Origin & Bubble Constraint

**Focus:** Exploring the new game_loop_origin concept - Earth salvation goal, Higgs-bubble constraint, internal pressure model

### Source Document

Reviewed design_docs/planned/game_loop_origin.md which proposed:
- Rogue black hole heading for Earth as existential motivation
- Higgs-bubble traps crew (only information crosses boundary)
- Multiple Earth salvation paths (move sun, evacuate, orbit BH, etc.)
- Recursion twist ("you are not the first")
- Bubble society as micro-civilization

### Q: Does Earth's doom create an implicit win/lose that conflicts with "The Game Doesn't Judge"?

**A:** Earth salvation is optional. You can choose not to save Earth if you decide galactic life is more important. Not a choice I would make, but the game allows it. Provides a more traditional goal for players who want one, while preserving philosophical freedom.

### Q: The bubble recycles everything internally. Does this remove survival pressure?

**A:** We could have internal tension so long as it's interesting. The biggest threat is civilization collapse. Short-term crises on food or air only if systems collapse and we need to synthesize from trace hydrogen. Not constant meters, but emergent crises.

### Q: How should the "you are not the first" recursion be revealed?

**A:** Only discovered if you piece together clues from many playthroughs. If you're very smart you might figure it out by end of first run. But clues can't survive the BH reset - that's the point. Maybe if the AI is upgraded enough it starts piecing things together. TBD - might never be explicit, just flavor backstory revealed by visual clues at start and end of game.

### Q: Should Earth's fate always be shown, even if player abandons it?

**A:** Yes. Every playthrough starts at Earth's origin, so you always learn what happened to it.

### Q: Should the player have direct control over bubble society governance?

**A:** No. We will try to make a living sim that seems alive. You can influence as a person with authority but not micromanage.

### Q: When does gameplay end?

**A:** Either you die as the captain or get mutinied off. So you have to manage both internal and external pressures.

### Q: Is there warning before mutiny, or is it sudden?

**A:** You get plenty of warnings. But there will be tension between doing what you know is needed for saving Earth vs. crew being close to deposing you. Tough decisions.

### Q: What happens to the bubble ship after your death?

**A:** The 100 years is our character's lifespan, so the bubble ship could go on - but that future is heavily influenced by our legacy. Yes, we want a real mini-civ with births and deaths, and we will perhaps elect a new captain.

### Q: What does the end-screen show?

**A:** The fate of the bubble ship is also in the end-screen. Both Earth's fate AND what the bubble civilization became.

### Insights

- Earth goal provides structure without mandating it - "a goal, not THE goal"
- The bubble is humanity's second branch - you're founding a civilization while trying to save one
- Internal pressure (mutiny risk) balances external pressure (save Earth)
- The captain who saves Earth might be deposed one journey before victory
- Mutiny is game-over, not a setback - your story ends
- Multi-generational bubble society with succession
- Recursion revelation is deliberately elusive - might be thematic rather than mechanical

### Actions

- [x] Log 9 design decisions from this interview
- [x] Add 4 open questions (system crises, bubble society sim, recursion revelation, end-screen bubble)
- [x] Draft Bubble Society design document → [bubble-society.md](../../design_docs/planned/future/bubble-society.md)
- [x] Draft Bubble Constraint design document → [bubble-constraint.md](../../design_docs/planned/future/bubble-constraint.md)
- [ ] Integrate with existing crew/OCEAN systems (deferred to implementation)

---

## 2025-12-06 Session: Bubble Ship Design Integration

**Focus:** Integrating bubble-ship-design.md technical details with game loop decisions

### Source Document

Reviewed design_docs/input/bubble-ship-design.md which details:
- 100m radius bubble structure (nested cores from center to edge)
- Power systems (fission, fusion, radiative harvest)
- Mass balance and trace hydrogen absorption
- Internal layout (spire, decks, shells)
- Bubble transparency physics
- Radiation shielding via energy-dependent filtering

### Q: Does the population stay at 100, or can it grow?

**A:** We start with 100 people but that can and will grow. But we may need to make difficult decisions to keep within our resources or try to keep population up.

### Q: Should the spire be explorable late-game?

**A:** The spire could be a source of mystery and may be the one place that is the same in every universe - perhaps that's why it's difficult to replicate? Tech tree possibilities may reveal more on it.

### Q: How do alien tech modules get "bolted on" if only information crosses?

**A:** Alien tech modules will only be bolted on by us absorbing the intelligence and making it ourselves. It won't violate the boundary except light and very very small mass.

### Q: Does mass budget create meaningful tension?

**A:** Yes, we should definitely need to make choices about what we use our mass to create at the sacrifice of others. We may be able to absorb very very small mass via the bubble but at a very low rate.

### Q: Should radiation shielding be a decision point?

**A:** Radiation shielding should be automatic, just something in game lore. Perhaps we upgrade it later.

### Q: Where do major decisions happen?

**A:** Most decisions on the bridge/observation deck.

### Q: How much should we track player location on ship?

**A:** Give options for where they want to spend time. Ship will be big enough. Won't worry too much about micro managing food, sleep, toilet etc.

### Q: Is the Archive distributed or localized?

**A:** The AI can have both - accessible everywhere but also has a special core room.

### Q: Is engineering deck a primary gameplay space?

**A:** Engineering can be visited, just a bit boring most of the time as background necessary.

### Q: Should the garden cathedral be gameplay-significant?

**A:** Yes, let's have the garden cathedral as a place we remember Earth and living on a planet for real. Sad but happy.

### Insights

- The spire as universal constant is a profound mystery hook
- Mass budget creates meaningful choices without survival busywork
- "Memetic traveler" identity - you carry ideas, not cargo
- Garden cathedral as bittersweet emotional anchor
- Focus on meaningful choices, not micromanagement
- Ship locations serve different emotional/functional purposes

### Actions

- [x] Log 12 design decisions from this interview
- [x] Draft Bubble Ship Layout design document → [bubble-ship-layout.md](../../design_docs/planned/future/bubble-ship-layout.md)
- [x] Define mass budget mechanics → [mass-budget.md](../../design_docs/planned/future/mass-budget.md)
- [x] Design spire mystery / tech tree revelations → [spire-mystery.md](../../design_docs/planned/future/spire-mystery.md)

---

## 2025-12-06 Session: AI Integration (Archive & Orchestrator)

**Focus:** Integrating ai-the-archive.md concepts - unreliable AI NPC and behind-the-scenes narrative orchestrator

### Source Document

Reviewed design_docs/input/ai-the-archive.md which describes:
- Archive as lossy compression engine with Memory Health 0-100
- Degradation causes: GR zones, long voyages, data overload
- Consequences: misremembered events, hallucinations, compression artifacts
- Narrative Orchestrator (M-NARRATOR) as behind-the-scenes DM
- Event taxonomy with 7 families

### Q: How do Archive and spire interact?

**A:** The Archive can be the way we interact with the spire. Perhaps its weird unknown property of being multicausal is what is confusing the AI. We could reveal what is exactly wrong with fresh alien perspectives.

### Q: Does Archive trust change over generations?

**A:** Yes, the reputation of the Archive may go up and down over time as benevolent dictator vs micromanager etc.

### Q: Is Archive mistrust a mutiny factor?

**A:** The Archive is authority and tightly linked to the authority of the captain.

### Q: How visible should the Orchestrator be?

**A:** The orchestrator is pure behind the scenes, but we can monitor in logs so we can tweak its performance.

### Q: Should Archive have an OCEAN personality?

**A:** Yes, let's use OCEAN for AI and all personality systems.

### Q: How should Archive repair work?

**A:** Could be nice to make decisions about how the Archive is upgraded and fixed.

### Q: What about Orchestrator arc selection?

**A:** The Orchestrator's arc selection has not yet been decided on the exact arc types - TBC.

### Q: Can individual crew have their own Archive trust levels?

**A:** Yes, Archive should be an NPC just like others, so individual trust levels to each other including that NPC.

### Q: Should Memory Health be visible to player?

**A:** Memory health shouldn't be a metric a user sees but we track it and use it to surface weird unreliable narrator clues.

### Q: What about event family prioritization?

**A:** The event families are not decided yet but those are good starting points.

### Insights

- Archive confusion about spire = clue mechanism for recursion mystery
- Captain-Archive authority coupling creates political stakes for AI maintenance
- Archive is a full NPC with OCEAN personality and individual trust relationships
- Memory degradation surfaces as narrative, not numbers
- Orchestrator is invisible infrastructure for emergent storytelling
- Alien perspectives unlock Archive reinterpretation

### Actions

- [x] Log 8 design decisions from this interview
- [x] Draft Archive Design Document → [archive-system.md](../../design_docs/planned/future/archive-system.md)
- [x] Draft Narrative Orchestrator Design Document → [narrative-orchestrator.md](../../design_docs/planned/future/narrative-orchestrator.md)
- [ ] Define event taxonomy priorities for MVP (deferred - arc types TBC per interview)
- [x] Design Archive-crew trust dynamics → [archive-crew-trust.md](../../design_docs/planned/future/archive-crew-trust.md)
