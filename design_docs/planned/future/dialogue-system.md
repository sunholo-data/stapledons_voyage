# Dialogue System

**Version:** 0.5.3
**Status:** Planned
**Priority:** P0 (Core Interaction)
**Complexity:** High
**AILANG Workarounds:** Tree traversal recursion, state machine pattern
**Depends On:** v0.5.0 UI Modes Framework, v0.5.1 Ship Exploration

## Related Documents

- [UI Modes Architecture](../v0_5_0/ui-modes.md) - Mode framework
- [Ship Exploration](../v0_5_1/ship-exploration.md) - Triggers crew dialogue
- [Civilization Detail](../v0_6_1/civilization-detail.md) - First contact dialogues
- [Game Vision](../../../docs/game-vision.md) - Character interaction design

## Problem Statement

Meaningful character interaction is core to the ship-as-home experience. Players need to:
- Talk to crew members about their lives, relationships, concerns
- Navigate branching conversations with consequences
- Experience first contact with alien civilizations
- Make choices that affect relationships and outcomes

**Current State:**
- No dialogue rendering
- No conversation state machine
- No choice/consequence system
- No portrait display

**What's Needed:**
- Full-screen dialogue mode
- Portrait + text + choices layout
- Branching dialogue trees
- Relationship/mood impact tracking
- AI-assisted dialogue generation (future)

---

## Design Overview

### Dialogue Philosophy

Conversations should feel **meaningful and consequential**:

- **Every choice matters** - No throwaway responses
- **Characters remember** - Past conversations affect future ones
- **Relationships evolve** - Trust, romance, conflict develop over time
- **Time pressure** - Ship life continues during conversations

### Dialogue Types

| Type | Context | Features |
|------|---------|----------|
| Crew casual | Daily ship life | Relationship building, mood check |
| Crew critical | Events, decisions | Choices affect outcomes |
| First contact | New civilizations | Cultural exchange, risk |
| Negotiation | Trade, treaties | Multiple rounds, stakes |
| Crisis | Emergencies | Time-limited, high stakes |
| Memorial | Crew death | Emotional closure |

---

## Detailed Specification

### Dialogue State

```ailang
module sim/dialogue

type DialogueState = {
    conversationID: ConversationID,
    speaker: Speaker,
    currentNode: DialogueNodeID,
    history: [DialogueNodeID],          -- Nodes visited this conversation
    pendingEffects: [DialogueEffect],   -- Effects to apply on exit
    emotionalState: EmotionalState,
    timeSpent: int                       -- Ticks in dialogue
}

type Speaker =
    | CrewMember(CrewID, CrewData)
    | Alien(CivilizationID, AlienData)
    | ShipAI
    | Narrator

type CrewData = {
    name: string,
    portrait: SpriteID,
    mood: Mood,
    relationshipToPlayer: int,          -- -100 to 100
    traits: [Trait]
}

type AlienData = {
    species: string,
    portrait: SpriteID,
    culturalTraits: [CulturalTrait],
    trustLevel: int,
    communicationQuality: float         -- 0-1, affects understanding
}

type EmotionalState = {
    currentEmotion: Emotion,
    intensity: float,
    relationshipDelta: int              -- Change this conversation
}

type Emotion = Neutral | Happy | Sad | Angry | Fearful | Curious | Loving | Grieving
```

### Dialogue Tree Structure

```ailang
type DialogueTree = {
    id: ConversationID,
    rootNode: DialogueNodeID,
    nodes: [DialogueNode],
    metadata: TreeMetadata
}

type TreeMetadata = {
    category: DialogueCategory,
    minRelationship: int,               -- Required relationship to trigger
    cooldown: int,                      -- Ticks before can repeat
    priority: int                       -- For event queue
}

type DialogueNode = {
    id: DialogueNodeID,
    content: NodeContent,
    choices: [DialogueChoice],
    autoAdvance: Maybe(DialogueNodeID), -- For narration, auto-continue
    conditions: [Condition]             -- Must pass to show node
}

type NodeContent =
    | Speech(string, Emotion)           -- Speaker says with emotion
    | Thought(string)                   -- Player internal monologue
    | Narration(string)                 -- Third-person description
    | Action(string)                    -- Physical action description
    | Silence(int)                      -- Pause for N frames

type DialogueChoice = {
    id: ChoiceID,
    text: string,
    tooltip: Maybe(string),             -- Hint at consequences
    targetNode: DialogueNodeID,
    effects: [DialogueEffect],
    conditions: [Condition],
    displayConditions: [Condition]      -- Whether to show at all
}

type DialogueEffect =
    | RelationshipChange(int)
    | MoodChange(Mood)
    | SetFlag(FlagID, bool)
    | GiveItem(ItemID)
    | TakeItem(ItemID)
    | TriggerEvent(EventID)
    | UnlockDialogue(ConversationID)
    | ModifyCrew(CrewID, CrewModifier)
    | LogEntry(string)

type Condition =
    | RelationshipAbove(int)
    | RelationshipBelow(int)
    | HasFlag(FlagID)
    | NotFlag(FlagID)
    | HasItem(ItemID)
    | CrewAlive(CrewID)
    | YearAfter(int)
    | YearBefore(int)
    | RandomChance(float)               -- Requires RNG (v0.5.1)
```

### Dialogue Processing

```ailang
-- Process dialogue input
pure func processDialogueInput(state: DialogueState, tree: DialogueTree, input: FrameInput) -> DialogueState {
    let currentNode = findNode(tree.nodes, state.currentNode);

    -- Handle auto-advance nodes
    match currentNode.autoAdvance {
        Some(nextID) => if input.clicked || state.timeSpent > 120 then
            advanceToNode(state, nextID)
        else
            { state | timeSpent: state.timeSpent + 1 },

        None => {
            -- Handle choice selection
            let availableChoices = filterChoices(currentNode.choices, state);
            match input.selectedChoice {
                None => { state | timeSpent: state.timeSpent + 1 },
                Some(choiceID) => {
                    let choice = findChoice(availableChoices, choiceID);
                    let withEffects = applyEffects(state, choice.effects);
                    advanceToNode(withEffects, choice.targetNode)
                }
            }
        }
    }
}

-- Move to new node
pure func advanceToNode(state: DialogueState, nodeID: DialogueNodeID) -> DialogueState {
    { state |
        currentNode: nodeID,
        history: nodeID :: state.history,
        timeSpent: 0 }
}

-- Filter choices by conditions
pure func filterChoices(choices: [DialogueChoice], state: DialogueState) -> [DialogueChoice] {
    filter(\c. allConditionsMet(c.conditions, state), choices)
}

-- Apply effects from choice
pure func applyEffects(state: DialogueState, effects: [DialogueEffect]) -> DialogueState {
    foldl(\s, e. applyEffect(s, e), state, effects)
}

pure func applyEffect(state: DialogueState, effect: DialogueEffect) -> DialogueState {
    match effect {
        RelationshipChange(delta) => {
            let newEmoState = { state.emotionalState |
                relationshipDelta: state.emotionalState.relationshipDelta + delta };
            { state | emotionalState: newEmoState }
        },
        MoodChange(mood) => {
            let newEmoState = { state.emotionalState | currentEmotion: mood };
            { state | emotionalState: newEmoState }
        },
        _ => {
            -- Queue effect for later application
            { state | pendingEffects: effect :: state.pendingEffects }
        }
    }
}
```

### Dialogue Exit and Effect Application

```ailang
-- Exit dialogue and apply all effects to world
pure func exitDialogue(world: World, state: DialogueState) -> World {
    let worldWithEffects = foldl(\w, e. applyWorldEffect(w, e), world, state.pendingEffects);

    -- Update speaker relationship
    let worldWithRelationship = updateSpeakerRelationship(worldWithEffects, state);

    -- Log the conversation
    let logEntry = createDialogueLog(state);
    let worldWithLog = addLogEntry(worldWithRelationship, logEntry);

    -- Return to previous mode
    transitionTo(worldWithLog, state.returnMode)
}

pure func applyWorldEffect(world: World, effect: DialogueEffect) -> World {
    match effect {
        SetFlag(flagID, value) => setWorldFlag(world, flagID, value),
        GiveItem(itemID) => addItemToShip(world, itemID),
        TakeItem(itemID) => removeItemFromShip(world, itemID),
        TriggerEvent(eventID) => queueEvent(world, eventID),
        UnlockDialogue(convID) => unlockConversation(world, convID),
        ModifyCrew(crewID, modifier) => modifyCrewMember(world, crewID, modifier),
        LogEntry(text) => addLogEntry(world, { year: world.gameYear, text: text }),
        _ => world  -- Already handled in dialogue state
    }
}
```

### Rendering

```ailang
-- Generate draw commands for dialogue
pure func renderDialogue(world: World, state: DialogueState, tree: DialogueTree) -> [DrawCmd] {
    let node = findNode(tree.nodes, state.currentNode);

    -- Background dimming
    let bgCmd = Rect(0.0, 0.0, 1280.0, 720.0, 0, 40);  -- Semi-transparent black

    -- Portrait (left side)
    let portraitCmds = renderPortrait(state.speaker, state.emotionalState, 40);

    -- Text area (right side)
    let textCmds = renderDialogueText(node.content, state, 41);

    -- Choices (if available)
    let choiceCmds = renderChoices(node.choices, state, 42);

    -- Mood/relationship indicator
    let indicatorCmds = renderRelationshipIndicator(state.emotionalState, 43);

    concat(bgCmd :: portraitCmds, concat(textCmds, concat(choiceCmds, indicatorCmds)))
}

pure func renderPortrait(speaker: Speaker, emoState: EmotionalState, z: int) -> [DrawCmd] {
    let portrait = getSpeakerPortrait(speaker);
    let emotion = emoState.currentEmotion;
    let x = 50.0;
    let y = 100.0;
    let width = 400.0;
    let height = 500.0;

    [
        -- Portrait frame
        Panel(x - 10.0, y - 10.0, width + 20.0, height + 20.0, 1, 7, z),
        -- Portrait image
        Portrait(x, y, width, height, portrait, emotion, z + 1),
        -- Name plate
        Panel(x, y + height + 10.0, width, 40.0, 2, 7, z),
        Text(getSpeakerName(speaker), x + 10.0, y + height + 20.0, 8, z + 1)
    ]
}

pure func renderDialogueText(content: NodeContent, state: DialogueState, z: int) -> [DrawCmd] {
    let x = 500.0;
    let y = 100.0;
    let width = 730.0;
    let height = 300.0;

    let text = contentToText(content);
    let displayedText = if state.timeSpent < length(text) then
        substring(text, 0, state.timeSpent * 2)  -- Typewriter effect
    else
        text;

    [
        -- Text panel
        Panel(x, y, width, height, 1, 7, z),
        -- Dialogue text (wrapped)
        TextWrapped(displayedText, x + 20.0, y + 20.0, width - 40.0, 8, z + 1)
    ]
}

pure func renderChoices(choices: [DialogueChoice], state: DialogueState, z: int) -> [DrawCmd] {
    let availableChoices = filterChoices(choices, state);
    let x = 500.0;
    let startY = 420.0;
    let width = 730.0;
    let height = 50.0;
    let spacing = 60.0;

    mapWithIndex(\c, i. choiceToDrawCmd(c, x, startY + intToFloat(i) * spacing, width, height, state, z), availableChoices)
}

pure func choiceToDrawCmd(choice: DialogueChoice, x: float, y: float, w: float, h: float, state: DialogueState, z: int) -> DrawCmd {
    let isHovered = match state.hoveredChoice {
        Some(id) => id == choice.id,
        None => false
    };
    let bgColor = if isHovered then 3 else 2;
    Button(x, y, w, h, choice.text, bgColor, z)
}
```

---

## Portrait System

### Portrait Composition

Portraits can be:
1. **Static sprites** - Pre-made for each crew member
2. **Composed** - Base + expression overlay
3. **AI-generated** - Dynamic generation (future)

```ailang
type PortraitData =
    | StaticPortrait(SpriteID)
    | ComposedPortrait(BaseID, ExpressionID)
    | GeneratedPortrait(GenerationParams)

type ExpressionID = ExpNeutral | ExpHappy | ExpSad | ExpAngry | ExpFearful | ExpCurious | ExpLoving

-- Map emotion to expression
pure func emotionToExpression(emotion: Emotion) -> ExpressionID {
    match emotion {
        Neutral => ExpNeutral,
        Happy => ExpHappy,
        Sad => ExpSad,
        Angry => ExpAngry,
        Fearful => ExpFearful,
        Curious => ExpCurious,
        Loving => ExpLoving,
        Grieving => ExpSad
    }
}
```

### Expression Animation

```ailang
type PortraitAnimation = {
    frames: [ExpressionID],
    currentFrame: int,
    frameTime: int,
    looping: bool
}

-- Idle animation: subtle movement
pure func idleAnimation() -> PortraitAnimation {
    { frames: [ExpNeutral, ExpNeutral, ExpCurious, ExpNeutral],
      currentFrame: 0,
      frameTime: 60,
      looping: true }
}

-- Transition animation: expression change
pure func transitionAnimation(from: ExpressionID, to: ExpressionID) -> PortraitAnimation {
    { frames: [from, ExpNeutral, to],
      currentFrame: 0,
      frameTime: 10,
      looping: false }
}
```

---

## Go/Engine Integration

### Dialogue Renderer

```go
// engine/render/dialogue.go

type DialogueRenderer struct {
    portraits  *PortraitAtlas
    fonts      *FontSet
    panelNine  *NineSlice
}

func (r *DialogueRenderer) Render(screen *ebiten.Image, state DialogueState, tree DialogueTree) {
    // Dim background
    r.drawDimmer(screen, 0.7)

    // Draw portrait panel
    r.drawPortrait(screen, state.Speaker, state.EmotionalState)

    // Draw text panel with typewriter effect
    r.drawTextPanel(screen, state, tree)

    // Draw choices if available
    node := tree.FindNode(state.CurrentNode)
    if len(node.Choices) > 0 && !node.AutoAdvance {
        r.drawChoices(screen, node.Choices, state)
    }

    // Draw relationship indicator
    r.drawRelationshipMeter(screen, state)
}

func (r *DialogueRenderer) drawTextPanel(screen *ebiten.Image, state DialogueState, tree DialogueTree) {
    node := tree.FindNode(state.CurrentNode)
    text := node.Content.Text()

    // Typewriter effect
    visibleChars := min(state.TimeSpent*2, len(text))
    displayText := text[:visibleChars]

    // Word wrap
    wrapped := r.fonts.WrapText(displayText, 700)

    // Draw panel background
    r.panelNine.Draw(screen, 500, 100, 730, 300)

    // Draw text
    r.fonts.DrawText(screen, wrapped, 520, 120, color.White)
}

func (r *DialogueRenderer) drawChoices(screen *ebiten.Image, choices []DialogueChoice, state DialogueState) {
    y := 420.0
    for i, choice := range choices {
        if !choice.IsAvailable(state) {
            continue
        }

        isHovered := state.HoveredChoice == choice.ID
        bgColor := color.RGBA{40, 40, 60, 255}
        if isHovered {
            bgColor = color.RGBA{60, 60, 100, 255}
        }

        r.drawChoiceButton(screen, choice, 500, y, 730, 50, bgColor)
        y += 60
    }
}
```

### Input Handler

```go
// engine/input/dialogue.go

func CaptureDialogueInput(state DialogueState, choices []DialogueChoice) FrameInput {
    var input FrameInput

    mx, my := ebiten.CursorPosition()
    input.MouseX, input.MouseY = float64(mx), float64(my)

    // Check which choice is hovered
    for i, choice := range choices {
        if !choice.IsAvailable(state) {
            continue
        }
        y := 420.0 + float64(i)*60.0
        if inRect(input.MouseX, input.MouseY, 500, y, 730, 50) {
            input.HoveredChoice = &choice.ID
            break
        }
    }

    // Click to select
    if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
        input.Clicked = true
        input.SelectedChoice = input.HoveredChoice
    }

    // Space/Enter to advance auto-text
    if inpututil.IsKeyJustPressed(ebiten.KeySpace) || inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
        input.Clicked = true
    }

    // Escape to exit (if allowed)
    if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
        input.RequestExit = true
    }

    return input
}
```

---

## Dialogue Content Examples

### Crew Casual Conversation

```yaml
id: crew_casual_001
speaker: crew_member
root: greeting

nodes:
  greeting:
    content:
      speech: "Captain. Got a moment?"
      emotion: neutral
    choices:
      - text: "Of course. What's on your mind?"
        target: concern_reveal
        effects: [relationship_change: 2]
      - text: "I'm busy. Later."
        target: dismissed
        effects: [relationship_change: -5]

  concern_reveal:
    content:
      speech: "It's just... we've been traveling for so long. I'm starting to forget what Earth smelled like."
      emotion: sad
    choices:
      - text: "I know. Sometimes I wonder if we made the right choice."
        target: shared_doubt
        effects: [relationship_change: 5, set_flag: shared_doubt_chen]
      - text: "Earth is gone for us. Best to focus on what's ahead."
        target: pragmatic_response
        effects: [relationship_change: 0]
      - text: "Tell me about your favorite memory of home."
        target: memory_share
        effects: [relationship_change: 3]

  shared_doubt:
    content:
      speech: "You too? I thought... I thought I was weak for feeling this way."
      emotion: relieved
    autoAdvance: bond_formed

  bond_formed:
    content:
      narration: "A quiet understanding passes between you. Some burdens are lighter shared."
    choices:
      - text: "[End conversation]"
        target: exit
        effects: [relationship_change: 10, unlock_dialogue: crew_casual_002]
```

### First Contact

```yaml
id: first_contact_tau_ceti
speaker: alien_tau_ceti
root: initial_transmission

nodes:
  initial_transmission:
    content:
      narration: "The signal resolves into a visual feed. A being of crystalline structures regards you with what might be curiosity."
    autoAdvance: greeting_attempt

  greeting_attempt:
    content:
      speech: "[GEOMETRIC PATTERNS] [RESONANCE QUERY]"
      emotion: curious
    choices:
      - text: "Send mathematical sequence: prime numbers"
        target: math_success
        effects: [set_flag: tau_ceti_math_first]
      - text: "Send visual: human waving"
        target: gesture_confusion
      - text: "Send audio: music sample"
        target: art_intrigue
        conditions: [has_item: music_archive]

  math_success:
    content:
      narration: "The crystalline being pulses with light. It responds with the next primes in the sequence, then an elaborate fractal."
    autoAdvance: communication_established

  communication_established:
    content:
      speech: "[PATTERN-MAKER] [QUERY: ORIGIN] [QUERY: PURPOSE]"
      emotion: curious
    choices:
      - text: "We are explorers from a distant star."
        target: explain_mission
        effects: [relationship_change: 5]
      - text: "We seek knowledge and connection."
        target: philosophical_interest
        effects: [relationship_change: 5]
      - text: "We come in peace, but we are also powerful."
        target: veiled_threat
        effects: [relationship_change: -10, set_flag: tau_ceti_threatened]
```

---

## Implementation Plan

### Phase 1: Basic Dialogue Display

| Task | File | Description |
|------|------|-------------|
| 1.1 | `sim_gen/dialogue.go` | DialogueState, DialogueNode types |
| 1.2 | `engine/render/dialogue.go` | Full-screen dialogue panel |
| 1.3 | `engine/render/dialogue.go` | Text display with wrapping |
| 1.4 | Test | See dialogue panel render |

### Phase 2: Choices and Navigation

| Task | File | Description |
|------|------|-------------|
| 2.1 | `sim_gen/dialogue.go` | DialogueChoice type |
| 2.2 | `sim_gen/funcs.go` | Node navigation logic |
| 2.3 | `engine/input/dialogue.go` | Choice hover/click |
| 2.4 | `engine/render/dialogue.go` | Choice button rendering |
| 2.5 | Test | Click through dialogue tree |

### Phase 3: Portrait System

| Task | File | Description |
|------|------|-------------|
| 3.1 | `engine/assets/portraits.go` | Portrait loading |
| 3.2 | `engine/render/dialogue.go` | Portrait rendering |
| 3.3 | `sim_gen/dialogue.go` | Emotion → expression mapping |
| 3.4 | Test | See portraits with expressions |

### Phase 4: Effects System

| Task | File | Description |
|------|------|-------------|
| 4.1 | `sim_gen/dialogue.go` | DialogueEffect type |
| 4.2 | `sim_gen/funcs.go` | Effect application |
| 4.3 | `sim_gen/funcs.go` | World state updates on exit |
| 4.4 | Test | Relationship changes persist |

### Phase 5: Conditions System

| Task | File | Description |
|------|------|-------------|
| 5.1 | `sim_gen/dialogue.go` | Condition type |
| 5.2 | `sim_gen/funcs.go` | Condition evaluation |
| 5.3 | `sim_gen/funcs.go` | Choice filtering by condition |
| 5.4 | Test | Conditional choices appear/hide |

### Phase 6: Typewriter Effect

| Task | File | Description |
|------|------|-------------|
| 6.1 | `sim_gen/funcs.go` | Text reveal timing |
| 6.2 | `engine/render/dialogue.go` | Partial text rendering |
| 6.3 | `engine/input/dialogue.go` | Click to complete |
| 6.4 | Test | Text reveals gradually |

### Phase 7: Content Integration

| Task | File | Description |
|------|------|-------------|
| 7.1 | `assets/dialogues/` | Sample dialogue trees |
| 7.2 | `sim_gen/dialogue.go` | Dialogue loading |
| 7.3 | `sim_gen/funcs.go` | Crew → dialogue mapping |
| 7.4 | Test | Talk to specific crew members |

---

## Testing Strategy

### Manual Testing

```bash
make run-mock
# 1. Enter ship exploration mode
# 2. Click on crew member
# 3. See dialogue panel appear
# 4. Read text typewriter effect
# 5. Hover over choices
# 6. Click choice → advances
# 7. See portrait emotion change
# 8. Exit dialogue → relationship updated
```

### Automated Testing

```go
func TestDialogueNodeNavigation(t *testing.T)
func TestChoiceFiltering(t *testing.T)
func TestEffectApplication(t *testing.T)
func TestConditionEvaluation(t *testing.T)
func TestTypewriterTiming(t *testing.T)
func TestDialogueExit(t *testing.T)
```

---

## AILANG Constraints

| Limitation | Impact | Workaround |
|------------|--------|------------|
| No mutable state | Can't update emotion in place | Immutable state passing |
| Recursion depth | Deep dialogue trees | Max 20 nodes per conversation |
| No RNG (until v0.5.1) | Random responses unavailable | Deterministic based on world state |
| String handling | Limited manipulation | Pre-compute text lengths |

---

## Success Criteria

### Core Functionality
- [ ] Dialogue mode renders
- [ ] Text displays with typewriter
- [ ] Choices are selectable
- [ ] Node navigation works

### Visual Quality
- [ ] Portrait displays correctly
- [ ] Expressions change with emotion
- [ ] Panel layout clean

### Effects System
- [ ] Relationship changes apply
- [ ] Flags set/check work
- [ ] Items can be given/taken

### Integration
- [ ] Ship → dialogue transition smooth
- [ ] Exit returns to ship
- [ ] Changes persist in world

---

## Future Extensions

| Feature | Description |
|---------|-------------|
| Voice acting | Audio for dialogue lines |
| AI dialogue | LLM-generated responses |
| Procedural portraits | AI-generated faces |
| Relationship web | Visual of all crew relationships |
| Conversation history | Review past dialogues |
| Translation UI | For alien first contacts |
