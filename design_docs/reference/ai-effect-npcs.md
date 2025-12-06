# AI Effect for NPC & Civilization Decisions

**Status**: Planned
**Target**: v0.5.1 (tracks AILANG AI effect)
**Priority**: P2 - Medium
**Estimated**: 1 week
**Dependencies**: AILANG v0.5.1 with AI effect

## Game Vision Alignment

| Pillar | Relevance | Score | Notes |
|--------|-----------|-------|-------|
| Time Dilation Consequence | + | +1 | AI civs make decisions while you're in transit |
| Civilization Simulation | + | +1 | Rich, adaptive civ behavior |
| Philosophical Depth | + | +1 | Alien minds with truly alien reasoning |
| Ship & Crew Life | + | +1 | Crew decisions and dialogue |
| Legacy Impact | + | +1 | AI decisions shape long-term galaxy state |
| Hard Sci-Fi Authenticity | 0 | 0 | Abstracted as "alien cognition" |
| **Net Score** | | **+5** | **Decision: Move forward** |

**Feature type:** Gameplay (enables rich NPC/civ behavior without hand-coding every decision)

## Problem Statement

Civilizations and NPCs need to make decisions that feel:
1. **Intelligent** - Not random, considers context
2. **Alien** - Different civilizations think differently
3. **Emergent** - Surprising but explainable outcomes
4. **Consistent** - Same context should produce similar (not identical) behavior

Hand-coding every decision tree is infeasible for galaxy-scale simulation.

**Current State:**
- Mock uses simple rule-based decisions
- No AI integration

**Impact:**
- Enables truly alien civilizations
- Reduces hand-coding for NPC behavior
- Creates emergent narrative moments

## Goals

**Primary Goal:** Use AILANG's AI effect to enable context-aware decisions for civilizations and NPCs, with pluggable AI backends.

**Success Metrics:**
- Civ decisions consider full context (neighbors, history, philosophy)
- Different philosophies produce observably different behaviors
- AI calls are cacheable/deterministic when needed
- Fallback to rule-based when AI unavailable

## Solution Design

### Overview

AILANG provides a generic JSON-in/JSON-out AI effect:

```ailang
effect AI {
    decide(input: string) -> string
}
```

We wrap this with typed interfaces for our domain:

```ailang
-- Typed wrapper for civilization decisions
func civDecide(civ: Civ, context: CivContext) -> CivAction ! {AI} {
    let input = encodeCivContext(civ, context)
    let output = AI.decide(input)
    decodeCivAction(output)
}
```

### Decision Types

**Civilization Decisions:**
```ailang
type CivContext = {
    self: CivState,
    neighbors: [CivState],
    recentEvents: [Event],
    playerRelationship: Relationship,
    knownTech: [Tech],
    resourceState: ResourceState
}

type CivAction =
    | ExpandTo(Star)
    | SendEmissary(Civ)
    | DeclareWar(Civ)
    | OfferTrade(Civ, [Tech])
    | Isolate
    | SharePhilosophy(Civ)
    | RequestHelp(string)
    | Transcend
    | DoNothing
```

**Crew Decisions:**
```ailang
type CrewContext = {
    crewMember: CrewMember,
    shipState: ShipState,
    recentEvents: [Event],
    relationships: [Relationship],
    yearsRemaining: int
}

type CrewAction =
    | Express(Emotion, string)   -- Dialogue
    | RequestAction(string)      -- Ask captain to do something
    | FormRelationship(CrewMember)
    | Conflict(CrewMember)
    | Mutiny                     -- Extreme case
    | Accept                     -- Go along with current situation
```

### Philosophy-Driven Prompts

Each civilization philosophy has a different "personality" for AI decisions:

```ailang
func getPhilosophyPrompt(phil: Philosophy) -> string {
    match phil {
        Expansionist => "You believe growth is survival. Stagnation is death. " ++
                       "Always seek to expand, acquire, strengthen.",
        GiftEconomy => "You believe in reciprocity and generosity. " ++
                      "Resources shared freely return multiplied. " ++
                      "Hoarding is shameful.",
        DeathCelebrant => "You believe death gives life meaning. " ++
                         "Immortality is stagnation. Endings are sacred. " ++
                         "Do not fear collapse.",
        Isolationist => "You believe contact corrupts. Purity requires " ++
                       "separation. Other minds are dangerous.",
        -- ... etc
    }
}

func buildCivPrompt(civ: Civ, context: CivContext) -> string {
    let basePrompt = getPhilosophyPrompt(civ.philosophy)
    let contextJson = std/json.encode(context)

    "You are a civilization with this philosophy: " ++ basePrompt ++
    "\n\nCurrent situation:\n" ++ contextJson ++
    "\n\nWhat action do you take? Respond with JSON matching CivAction type."
}
```

### Go Handler Implementation

**Default stub (deterministic fallback):**
```go
func DefaultAIHandler(input string) string {
    // Parse input to determine context type
    var context map[string]interface{}
    json.Unmarshal([]byte(input), &context)

    // Simple rule-based fallback
    if civContext, ok := context["self"]; ok {
        return deterministicCivDecision(civContext)
    }
    if crewContext, ok := context["crewMember"]; ok {
        return deterministicCrewDecision(crewContext)
    }

    return `{"action": "DoNothing"}`
}
```

**LLM handler (swappable):**
```go
func LLMHandler(client *anthropic.Client) func(string) string {
    return func(input string) string {
        resp, err := client.Messages.Create(context.Background(), anthropic.MessageCreateParams{
            Model:     "claude-3-haiku",
            MaxTokens: 200,
            Messages: []anthropic.MessageParam{
                {Role: "user", Content: input},
            },
        })
        if err != nil {
            return DefaultAIHandler(input) // Fallback
        }
        return resp.Content[0].Text
    }
}
```

**Registration:**
```go
func main() {
    // Choose AI handler based on config
    if config.UseLLM {
        sim_gen.RegisterAIHandler(LLMHandler(anthropicClient))
    } else {
        sim_gen.RegisterAIHandler(DefaultAIHandler)
    }
}
```

### Caching & Determinism

**Problem:** LLM calls are non-deterministic and slow.

**Solution:** Cache-by-hash for similar contexts:

```go
type AICache struct {
    cache map[string]string
    mu    sync.RWMutex
}

func (c *AICache) CachedHandler(underlying func(string) string) func(string) string {
    return func(input string) string {
        // Hash the input for cache key
        hash := sha256.Sum256([]byte(input))
        key := hex.EncodeToString(hash[:])

        // Check cache
        c.mu.RLock()
        if cached, ok := c.cache[key]; ok {
            c.mu.RUnlock()
            return cached
        }
        c.mu.RUnlock()

        // Call underlying
        result := underlying(input)

        // Store in cache
        c.mu.Lock()
        c.cache[key] = result
        c.mu.Unlock()

        return result
    }
}
```

**For deterministic replay:** Pre-populate cache with recorded decisions.

### Rate Limiting

AI calls should be rare and meaningful:

```ailang
func shouldConsultAI(civ: Civ, timeSinceLastDecision: int) -> bool {
    -- Only consult AI for major decisions, not every tick
    timeSinceLastDecision > 100 &&  -- At least 100 ticks between decisions
    civ.hasSignificantChange         -- Something important happened
}
```

### Implementation Plan

**Phase 1: Effect Integration** (~2 days)
- [ ] Define typed context/action types in AILANG
- [ ] Implement JSON encoding/decoding
- [ ] Create wrapper functions for each decision type

**Phase 2: Philosophy Prompts** (~2 days)
- [ ] Write prompts for each philosophy
- [ ] Test prompt quality with sample contexts
- [ ] Tune for consistent output format

**Phase 3: Go Handlers** (~2 days)
- [ ] Implement default deterministic handler
- [ ] Implement LLM handler wrapper
- [ ] Add caching layer
- [ ] Add rate limiting

**Phase 4: Testing** (~1 day)
- [ ] Verify fallback works when AI unavailable
- [ ] Test cache determinism
- [ ] Validate decision diversity across philosophies

### Files to Modify/Create

**AILANG source:**
- `sim/ai_decisions.ail` - Context/action types, wrapper functions (~200 LOC)
- `sim/philosophy_prompts.ail` - Per-philosophy prompt text (~150 LOC)

**Go source:**
- `engine/ai/handler.go` - AI handler implementations (~300 LOC)
- `engine/ai/cache.go` - Caching layer (~100 LOC)

## Examples

### Example 1: Civ Deciding on Contact

**Context:**
```json
{
    "self": {"name": "Kepler-442 Collective", "philosophy": "GiftEconomy", "techLevel": 5},
    "neighbors": [
        {"name": "HD-40307 Empire", "philosophy": "Expansionist", "techLevel": 7}
    ],
    "recentEvents": ["player_shared_higgs_tech"],
    "playerRelationship": "friendly"
}
```

**AI Response (Gift Economy):**
```json
{
    "action": "OfferTrade",
    "target": "HD-40307 Empire",
    "offer": ["bio_synthesis_tech"],
    "reasoning": "The player shared freely with us. We honor reciprocity by sharing with our neighbor, even if they follow a different path."
}
```

### Example 2: Crew Reacting to Long Journey

**Context:**
```json
{
    "crewMember": {"name": "Dr. Chen", "role": "Scientist", "age": 78, "yearsRemaining": 5},
    "shipState": {"inTransit": true, "destination": "HD-40307", "yearsRemaining": 15},
    "recentEvents": ["left_known_space", "birthday_80"]
}
```

**AI Response:**
```json
{
    "action": "Express",
    "emotion": "melancholy",
    "dialogue": "I won't see HD-40307. I've made peace with that. But I hope whoever reads my research notes will understand why we came."
}
```

## Success Criteria

- [ ] AI handler pluggable at runtime
- [ ] Default fallback produces reasonable decisions
- [ ] Philosophy affects decision style visibly
- [ ] Cache provides determinism for replay
- [ ] Rate limiting prevents spam

## Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|-----------|
| LLM output malformed | High | Strict JSON validation; fallback on parse error |
| LLM too slow | Med | Cache heavily; use fastest model (Haiku) |
| Decisions feel random | Med | Philosophy prompts; test for consistency |
| Cost of LLM calls | Low | Rate limiting; caching; fallback in dev |

## References

- [consumer-contract-v0.5.md](../../ailang_resources/consumer-contract-v0.5.md) - AI effect spec
- [planet-state-transitions.md](planet-state-transitions.md) - Philosophy types

## Future Work

- Fine-tuned models per philosophy
- Player-facing "why did they do that?" explanations
- Learning from player reactions
- Multi-turn negotiations via AI

---

**Document created**: 2025-12-02
**Last updated**: 2025-12-02
