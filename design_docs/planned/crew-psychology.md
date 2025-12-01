# Crew Psychology System

**Status:** Planned
**Pillars served:** The Ship Is Home, The Game Doesn't Judge, Time Has Emotional Weight, Grounded Strangeness
**Dependencies:** AI dialogue system, end-screen UI
**Related decisions:** See [design-decisions.md](../../docs/vision/design-decisions.md) (2025-11-30 entries)

---

## Overview

Crew members have psychologically grounded personalities using the OCEAN (Big Five) model. Personality drives emergent behavior: dialogue tone, event reactions, relationship chemistry, crew votes, and civilization preferences. Personalities drift over time based on experiences, but this drift is hidden during gameplay and revealed in the end screen.

---

## OCEAN Model

Each crew member has five trait values from 0.0 to 1.0:

| Trait | Low (0.0) | High (1.0) |
|-------|-----------|------------|
| **O** (Openness) | Practical, conventional | Curious, imaginative |
| **C** (Conscientiousness) | Spontaneous, flexible | Organized, disciplined |
| **E** (Extraversion) | Reserved, solitary | Outgoing, energetic |
| **A** (Agreeableness) | Competitive, skeptical | Cooperative, trusting |
| **N** (Neuroticism) | Calm, secure | Anxious, sensitive |

---

## Archetype Baseline Profiles

| Archetype | O | C | E | A | N | MBTI Analogue | Behavioral Summary |
|-----------|---|---|---|---|---|---------------|-------------------|
| **Engineer** | 0.2 | 0.8 | 0.2 | 0.5 | 0.2 | ISTP | Practical, laconic, fiercely competent |
| **Scientist** | 0.8 | 0.8 | 0.2 | 0.5 | 0.5 | INTJ | Analytical, future-oriented, logic-confident |
| **Medic** | 0.5 | 0.5 | 0.5 | 0.8 | 0.5 | ISFJ | Warm, stabilizing, relationally attuned |
| **Diplomat** | 0.5 | 0.5 | 0.8 | 0.8 | 0.2 | ENFJ | Reads people, persuasive, consensus-builder |
| **Pilot** | 0.2 | 0.2 | 0.8 | 0.5 | 0.2 | ESTP | Decisive, thrill-driven, hates deliberation |
| **Quartermaster** | 0.2 | 0.9 | 0.5 | 0.3 | 0.2 | ESTJ | Craves structure, routine, clear rules |
| **Zealot** | 0.8 | 0.5 | 0.5 | 0.3 | 0.5 | INFJ | Passionate, morally intense, idealistic |
| **Dreamer** | 0.9 | 0.2 | 0.5 | 0.5 | 0.8 | INFP | Sensitive, poetic, sometimes overwhelmed |
| **Skeptic** | 0.5 | 0.8 | 0.2 | 0.3 | 0.8 | INTP/ISTJ | Doubts motives, challenges assumptions |
| **Fantasist** | 0.8 | 0.1 | 0.8 | 0.5 | 0.5 | ENFP/ENTP | Creative chaos agent, shakes up stale dynamics |
| **Analyst** | 0.5 | 0.8 | 0.2 | 0.5 | 0.2 | ISTJ/INTJ | Quiet pattern-seeker, sees what others miss |

**Player** archetype is "The Self" — OCEAN is emergent from gameplay choices, not declared.

---

## Psychological Needs and Offers

Each crew member has implicit psychological needs (what they seek from others) and offers (what they provide to relationships). These are **derived from OCEAN**, not manually defined.

### Derivation Formulas

**Needs** (what they seek):
```
validation   = N × (1 - C)     -- anxious + impulsive → seek reassurance
stability    = N × C           -- anxious + conscientious → want predictability
stimulation  = O × E           -- open + extraverted → crave novelty
autonomy     = (1-A) × (1-E)   -- disagreeable introverts → want space
connection   = E × A           -- extraverted + agreeable → seek bonds
```

**Offers** (what they provide):
```
validation   = A × E           -- agreeable extraverts give reassurance
stability    = C × (1 - N)     -- conscientious + calm → provide stability
stimulation  = O × E           -- open extraverts bring energy
autonomy     = (1-C) × (1-A)   -- low C + low A → leave others alone
connection   = E × A           -- same as needs (give what they seek)
```

### Example: Dreamer (O:0.9, C:0.2, E:0.5, A:0.5, N:0.8)

**Needs:**
- validation: 0.8 × 0.8 = 0.64 (high)
- stability: 0.8 × 0.2 = 0.16 (low)
- stimulation: 0.9 × 0.5 = 0.45
- autonomy: 0.5 × 0.5 = 0.25
- connection: 0.5 × 0.5 = 0.25

**Offers:**
- validation: 0.5 × 0.5 = 0.25 (moderate)
- stability: 0.2 × 0.2 = 0.04 (very low)
- stimulation: 0.9 × 0.5 = 0.45
- autonomy: 0.8 × 0.5 = 0.40
- connection: 0.5 × 0.5 = 0.25

Dreamer desperately needs validation but offers little stability — creating asymmetric relationship dynamics.

---

## Directional Relationship Chemistry

Chemistry is **not symmetric**. A→B may differ from B→A.

### Chemistry Formula

```
chemistry(A → B) =
    0.7 × needMatch(A.needs, B.offers) +
    0.3 × similarity(A.ocean, B.ocean)

needMatch(needs, offers) =
    needs.validation × offers.validation +
    needs.stability × offers.stability +
    needs.stimulation × offers.stimulation +
    needs.autonomy × offers.autonomy +
    needs.connection × offers.connection

similarity(ocean1, ocean2) =
    1.0 - (|O1-O2| + |C1-C2| + |E1-E2| + |A1-A2| + |N1-N2|) / 5.0
```

### Example Asymmetric Relationships

| Pair | A→B | B→A | Dynamic |
|------|-----|-----|---------|
| Dreamer → Medic | 0.72 | 0.48 | Dreamer adores Medic; Medic finds Dreamer exhausting |
| Pilot → Zealot | 0.41 | 0.38 | Mutual tension, neither gives what the other needs |
| Scientist → Skeptic | 0.65 | 0.61 | Mutual respect, similar profiles |
| Fantasist → Quartermaster | 0.32 | 0.28 | Comedy or disaster — opposites |
| Diplomat → Anyone | 0.6+ | varies | Diplomat stabilizes most pairs |

---

## OCEAN Drift Over Time

Crew personalities change based on experiences, bounded by ±0.2 from baseline.

### Drift Triggers

| Event Type | O | C | E | A | N | Notes |
|------------|---|---|---|---|---|-------|
| Witnessed extinction | +0.05 | — | -0.02 | — | +0.08 | Cosmos cracks you open |
| Deep time dilation (>1000y) | +0.03 | -0.02 | -0.03 | — | +0.05 | Isolation weighs |
| Crew member death | — | — | -0.05 | +0.02 | +0.10 | Grief bonds or isolates |
| First alien contact | +0.08 | — | +0.03 | +0.03 | -0.02 | Wonder opens mind |
| Betrayal/conflict | — | +0.03 | -0.03 | -0.08 | +0.05 | Trust hardens |
| Successful cooperation | — | +0.02 | +0.02 | +0.05 | -0.03 | Trust softens |
| Philosophy contact (strange) | +0.10 | -0.02 | — | — | +0.03 | Worldview expands |
| Philosophy contact (familiar) | -0.02 | +0.02 | — | +0.02 | -0.02 | Comfort reinforces |

### Drift Bounds

Each trait is clamped to `baseline ± 0.2`. An Engineer (baseline O: 0.2) can drift to at most O: 0.4 — they might become more open, but never become a Dreamer.

### Drift Visibility

**During gameplay:** Hidden. Players experience drift through:
- Dialogue tone shifts (AI-generated, reflects current OCEAN)
- Changed reactions to similar events
- Shifted voting patterns on crew decisions

**End screen:** Full reveal showing:
- Before/after OCEAN radar charts
- Key drift moments ("After witnessing the Helix extinction, Engineer's Openness increased")
- How relationships evolved

---

## Behavioral Applications

### 1. Dialogue Tone (AI-generated)

OCEAN values inform dialogue generation prompts:

| Trait | Low → Speech Pattern | High → Speech Pattern |
|-------|---------------------|----------------------|
| O | Concrete, practical | Metaphorical, abstract |
| C | Casual, spontaneous | Structured, organized |
| E | Brief, reserved | Verbose, social |
| A | Blunt, confrontational | Diplomatic, supportive |
| N | Calm, assured | Anxious, hedging |

### 2. Event Reactions

```
Example: Civilization extinction event

Engineer (Low O, High C): "Unfortunate. Nothing we could have done. Let's focus on the next system."
Dreamer (High O, High N): "They're gone... all of them... does any of this even matter?"
Zealot (High O, Low A): "Perhaps this is what they deserved. Their philosophy was corrupt."
Diplomat (High E, High A): "We should hold a memorial. The crew needs to process this together."
```

### 3. Crew Votes (Journey Planning)

Each archetype evaluates proposed journeys through their OCEAN lens:

| Evaluation | Formula | Meaning |
|------------|---------|---------|
| Risk tolerance | (1-N) × (1-C) | Low N + Low C = embrace risk |
| Exploration drive | O × E | High O + High E = want novelty |
| Moral weight | A × O | High A + High O = care about ethics |
| Stability needs | C × A | High C + High A = prefer safe choices |

### 4. Civilization Preferences

| Trait | Civilization Preference |
|-------|------------------------|
| High O | Fascinated by strange/alien minds |
| Low O | Prefer human-analogue cultures |
| High N | Stressed by chaotic/extinct civilizations |
| High A | Want peaceful, cooperative civilizations |
| High E | Excited by social, communicative species |

---

## Player OCEAN Inference

Player personality is emergent from choices. Broad category mappings:

| Decision Category | Affects | Examples |
|-------------------|---------|----------|
| Journey risk | C, N | Speed selection, destination danger |
| Civ interaction | A, O | Help vs ignore, trade vs hoard |
| Crew relationship | E, A | Listen to objections, social time |
| Exploration style | O, E | Seek strange aliens, revisit known |
| Resource management | C, A | Share freely, stockpile |

**Open question:** Exact mappings TBD via playtesting or AI-contextual inference. See [open-questions.md](../../docs/vision/open-questions.md).

---

## AILANG Type Definitions

```ailang
type OCEAN = {
    o: float,
    c: float,
    e: float,
    a: float,
    n: float
}

type PsychNeeds = {
    validation: float,
    stability: float,
    stimulation: float,
    autonomy: float,
    connection: float
}

type PsychOffers = {
    validation: float,
    stability: float,
    stimulation: float,
    autonomy: float,
    connection: float
}

type Archetype =
    | Engineer | Scientist | Medic | Diplomat | Pilot
    | Quartermaster | Zealot | Dreamer | Skeptic | Fantasist | Analyst

type CrewMember = {
    id: int,
    name: string,
    archetype: Archetype,
    baseline: OCEAN,
    current: OCEAN,
    age: int,
    yearsOnShip: int
}

type DriftEvent =
    | WitnessedExtinction(int)
    | DeepTimeDilation(float)
    | CrewDeath(int)
    | PhilosophyContact(Philosophy, bool)  -- (philosophy, is_strange)
    | RelationshipShift(int, float)
    | AlienContact(int)
    | Betrayal(int)
    | Cooperation(int)

-- Derived functions (computed, not stored)
pure func deriveNeeds(ocean: OCEAN) -> PsychNeeds
pure func deriveOffers(ocean: OCEAN) -> PsychOffers
pure func chemistry(a: CrewMember, b: CrewMember) -> float
pure func applyDrift(member: CrewMember, event: DriftEvent) -> CrewMember
pure func clampToBounds(baseline: OCEAN, current: OCEAN, bound: float) -> OCEAN
```

---

## End Screen: Crew Evolution

The legacy report includes a "Crew Evolution" section:

1. **Radar charts** — Before/after OCEAN for each surviving crew member
2. **Key moments** — "After the Helix extinction, Kira's Openness rose from 0.2 to 0.35"
3. **Relationship map** — Final chemistry scores, highlighting asymmetries
4. **The Captain** — Player's inferred OCEAN revealed: "You became..."
5. **Drift summary** — "Your crew grew more Open (+0.12 avg) but less Agreeable (-0.08 avg)"

---

## Implementation Priority

| Phase | Scope |
|-------|-------|
| **MVP** | Archetype baselines, static chemistry calculation |
| **v0.6** | Drift mechanics, AI dialogue integration |
| **v0.7** | End-screen evolution reveal |
| **v0.8** | Player OCEAN inference |

---

## References

- Big Five / OCEAN: Costa & McCrae (1992)
- MBTI correlations used only for dialogue flavor, not gameplay mechanics
- Design decisions: [design-decisions.md](../../docs/vision/design-decisions.md)
