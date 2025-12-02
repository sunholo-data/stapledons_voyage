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
- [ ] Draft Black Hole Design Document
- [ ] Design opening sequence (emergence from mysterious structure)
