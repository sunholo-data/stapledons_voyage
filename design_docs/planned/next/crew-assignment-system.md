# Crew Assignment System

## Status
- **Status:** Planned
- **Priority:** P1 (Core to Ship-as-Home pillar)
- **Estimated Effort:** 3-4 days
- **Target Version:** v0.4.0

---

## Game Vision Alignment

Checked against [core-pillars.md](../../../docs/vision/core-pillars.md):

| Pillar | Alignment | Rationale |
|--------|-----------|-----------|
| **Choices Are Final** | ✅ Supports | Crew assignments are permanent each journey; cannot be reassigned without time cost |
| **The Game Doesn't Judge** | ✅ Supports | Show assignment consequences (morale, efficiency) without moral framing |
| **Time Has Emotional Weight** | ✅ Supports | Crew specializations matter more on long journeys where they age and adapt |
| **The Ship Is Home** | ✅ Core | This feature directly realizes crew as meaningful characters with roles |
| **Grounded Strangeness** | N/A | Crew are humans, not alien; system applies universally |
| **We Are Not Built For This** | ✅ Supports | Crew in wrong roles suffer stress/morale penalties; forces player to respect limits |

**Pillar Analysis:**
- This feature is **foundational** to the game's core identity. Without crew assignment, crew are just NPCs with no agency.
- Per Pillar 4, crew provide "human-scale grounding" and have "opinions and reactions." Assignments formalize their roles and express their competencies.
- Per Pillar 6, assigning crew beyond their limits should have psychological costs, not just mechanical ones.

---

## Prior Design Decisions

Checked against [design-decisions.md](../../../docs/vision/design-decisions.md):

- **Single save file (no slots):** Assignments are per-journey; new playthrough resets assignments
- **Finite crew:** Crew ages over 100-year journey; assignments affect aging/stress
- **No silent crew:** Assignments unlock crew personality and reaction content
- **Ship as refuge:** Crew quarters and assignments reinforce intimacy of ship interior

---

## Feature Overview

The **Crew Assignment System** allows the player to assign ship crew members to specialized roles that affect:

1. **Ship Efficiency** - Some tasks run better with specialists (e.g., engineers, scientists, medics)
2. **Crew Morale** - Assignments matching crew strengths increase morale; mismatches decrease it
3. **Specialization Growth** - Crew develop expertise over time in their assigned roles
4. **Stress & Aging** - Extended assignments in high-stress roles accelerate aging and morale loss
5. **Dialogue & Personality** - Crew discuss their roles, express concerns, and reveal personalities through assignments

---

## Core Mechanics

### Crew Roles

Each crew member can be assigned to one of these roles:

| Role | Ship Function | Specialist Benefit | Stress Factor |
|------|---------------|-------------------|--------------|
| **Captain** | Command decisions | +20% decision speed | Low (command authority) |
| **Chief Engineer** | Engine efficiency | +25% fuel efficiency, faster repairs | High (constantly on alert) |
| **Chief Scientist** | Research & observation | +30% science discovery rate | Medium (mental fatigue) |
| **Chief Medical Officer** | Crew health management | +25% healing, reduced stress for others | High (emotional burden) |
| **Navigator** | Course plotting | +20% travel safety, less drift | Medium (cognitive strain) |
| **Security Chief** | Defense & monitoring | +15% threat response speed | High (vigilance stress) |
| **Botanist** | Life support & food | +20% resource efficiency, morale from fresh food | Medium |
| **Counselor** | Morale & psychology | +15% crew morale boost, stress recovery | High (absorbs crew problems) |
| **Generalist** | Flexible support | Can fill gaps; no bonus, no penalty | Low |

### Assignment Mechanics

#### Basic Rules
1. **One role per crew member** - Each assigned member covers one primary role
2. **Fallback to Generalist** - If a role is unassigned, it defaults to any unassigned Generalist
3. **Specialization Accrual** - Crew in a role for >10 game-weeks develop +5% bonus per specialization level (capped at 5 levels = 25%)
4. **Stress Accumulation** - High-stress roles accumulate 1-2 stress per week; low-stress accumulate 0.5-1
5. **Stress Recovery** - Crew recover 0.5 stress per week off-duty; Counselor can boost this to 1.5 per week

#### Crew Attributes

Each NPC gains crew-specific fields:

```ailang
export type CrewMember = {
    id: int,
    name: string,
    pos: Coord,
    pattern: MovementPattern,

    -- Assignment system (NEW)
    assignedRole: Role,           -- Current assignment
    specialization: int,          -- Level 0-5 (growth in current role)
    stress: float,                -- 0-100 (mental state)
    morale: float,                -- 0-100 (satisfaction)

    -- Lifetime tracking
    yearsAtRole: float,           -- Years in current assignment (ages crew)
    previousRoles: [Role],        -- Assignments in past journeys (affects initial morale)

    -- Personality
    preferredRole: Option[Role],  -- Role crew member excels at / prefers
    concerns: [string],           -- Current worries (hunger, fatigue, missing family)
}
```

#### Role Type (AILANG ADT)

```ailang
export type Role =
    | RoleCaptain
    | RoleChiefEngineer
    | RoleChiefScientist
    | RoleChiefMedicalOfficer
    | RoleNavigator
    | RoleSecurityChief
    | RoleBotanist
    | RoleCounselor
    | RoleGeneralist
```

---

## AILANG Implementation

### Types (in `sim/crew.ail` - NEW FILE)

```ailang
module sim/crew

import sim/protocol (Coord)
import sim/npc_ai (NPC, MovementPattern)
import std/option (Some, None, Option)

-- Role enum for crew assignments
export type Role =
    | RoleCaptain
    | RoleChiefEngineer
    | RoleChiefScientist
    | RoleChiefMedicalOfficer
    | RoleNavigator
    | RoleSecurityChief
    | RoleBotanist
    | RoleCounselor
    | RoleGeneralist

-- Crew member with assignment state
export type CrewMember = {
    id: int,
    name: string,
    pos: Coord,
    pattern: MovementPattern,

    -- Assignment state
    assignedRole: Role,
    specialization: int,      -- 0-5 levels
    stress: float,            -- 0-100
    morale: float,            -- 0-100

    -- Lifetime tracking
    yearsAtRole: float,
    previousRoles: [Role],

    -- Personality
    preferredRole: Option[Role],
    concerns: [string]
}

-- Assignment result (success or error)
export type AssignmentResult =
    | AssignmentSuccess(CrewMember)
    | AssignmentFailed(string)

-- Check if crew member is suitable for role (preferred role match)
export pure func isSuitableFor(crew: CrewMember, role: Role) -> bool tests [
    (CrewMember {..., preferredRole: Some(RoleChiefEngineer)}, RoleChiefEngineer, true),
    (CrewMember {..., preferredRole: Some(RoleChiefScientist)}, RoleChiefEngineer, false),
    (CrewMember {..., preferredRole: None()}, RoleGeneralist, true)
] {
    match crew.preferredRole {
        Some(pref) => match (pref, role) {
            (RoleCaptain, RoleCaptain) => true,
            (RoleChiefEngineer, RoleChiefEngineer) => true,
            (RoleChiefScientist, RoleChiefScientist) => true,
            (RoleChiefMedicalOfficer, RoleChiefMedicalOfficer) => true,
            (RoleNavigator, RoleNavigator) => true,
            (RoleSecurityChief, RoleSecurityChief) => true,
            (RoleBotanist, RoleBotanist) => true,
            (RoleCounselor, RoleCounselor) => true,
            _ => false
        },
        None() => true  -- No preference, suitable for any
    }
}

-- Get stress multiplier for a role
export pure func roleStressMultiplier(role: Role) -> float
tests [
    (RoleCaptain, 0.8),
    (RoleChiefEngineer, 1.5),
    (RoleChiefScientist, 1.0),
    (RoleChiefMedicalOfficer, 1.4),
    (RoleNavigator, 1.1),
    (RoleSecurityChief, 1.3),
    (RoleBotanist, 1.0),
    (RoleCounselor, 1.2),
    (RoleGeneralist, 0.9)
] {
    match role {
        RoleCaptain => 0.8,
        RoleChiefEngineer => 1.5,
        RoleChiefScientist => 1.0,
        RoleChiefMedicalOfficer => 1.4,
        RoleNavigator => 1.1,
        RoleSecurityChief => 1.3,
        RoleBotanist => 1.0,
        RoleCounselor => 1.2,
        RoleGeneralist => 0.9
    }
}

-- Get efficiency bonus for role (percent bonus)
export pure func roleEfficiencyBonus(role: Role, specialization: int) -> float {
    let baseBonus = match role {
        RoleCaptain => 20.0,
        RoleChiefEngineer => 25.0,
        RoleChiefScientist => 30.0,
        RoleChiefMedicalOfficer => 25.0,
        RoleNavigator => 20.0,
        RoleSecurityChief => 15.0,
        RoleBotanist => 20.0,
        RoleCounselor => 15.0,
        RoleGeneralist => 0.0
    };
    let specializationBonus = intToFloat(specialization) * 5.0;
    baseBonus + specializationBonus
}

-- Assign crew to role (validates and updates state)
export pure func assignCrew(crew: CrewMember, role: Role) -> AssignmentResult {
    let newCrew = {
        crew |
        assignedRole: role,
        specialization: 0,  -- Reset specialization for new role
        yearsAtRole: 0.0,
        previousRoles: crew.previousRoles :: [crew.assignedRole]  -- Track history
    };
    AssignmentSuccess(newCrew)
}

-- Update crew stress and morale (call each tick)
export pure func updateCrewState(crew: CrewMember, deltaTime: float, isCounseled: bool) -> CrewMember {
    let stressChange = roleStressMultiplier(crew.assignedRole) * 0.1 * deltaTime;
    let stressRecovery = if isCounseled then 0.15 else 0.05 in;
    let newStress = max(0.0, min(100.0, crew.stress + stressChange - stressRecovery));

    let moraleChange = if isSuitableFor(crew, crew.assignedRole) then
        0.05 * deltaTime  -- Morale increases for good fit
    else
        -0.1 * deltaTime;  -- Morale decreases for poor fit

    let newMorale = max(0.0, min(100.0, crew.morale + moraleChange));

    {
        crew |
        stress: newStress,
        morale: newMorale,
        yearsAtRole: crew.yearsAtRole + deltaTime
    }
}

-- Develop specialization in current role
export pure func developSpecialization(crew: CrewMember) -> CrewMember {
    if crew.yearsAtRole > intToFloat(crew.specialization + 1) * 10.0 then
        {crew | specialization: min(5, crew.specialization + 1)}
    else
        crew
}

-- Get morale modifier for crew UI/dialogue
export pure func getMoraleDescription(morale: float) -> string tests [
    (90.0, "Excellent"),
    (70.0, "Good"),
    (50.0, "Neutral"),
    (30.0, "Poor"),
    (10.0, "Critical")
] {
    if morale >= 85.0 then "Excellent"
    else if morale >= 70.0 then "Good"
    else if morale >= 50.0 then "Neutral"
    else if morale >= 30.0 then "Poor"
    else "Critical"
}

-- Get stress description for crew UI/dialogue
export pure func getStressDescription(stress: float) -> string tests [
    (10.0, "Calm"),
    (30.0, "Alert"),
    (50.0, "Tense"),
    (70.0, "Strained"),
    (90.0, "Critical")
] {
    if stress <= 20.0 then "Calm"
    else if stress <= 40.0 then "Alert"
    else if stress <= 60.0 then "Tense"
    else if stress <= 80.0 then "Strained"
    else "Critical"
}

-- Helper: max function (standard)
pure func max(a: float, b: float) -> float {
    if a > b then a else b
}

-- Helper: min function (standard)
pure func min(a: float, b: float) -> float {
    if a < b then a else b
}
```

### Integration with World State (in `sim/world.ail`)

```ailang
-- Add to World type:
export type World = {
    tick: int,
    planet: PlanetState,
    npcs: [NPC],
    selection: Selection,
    interior: InteriorState,
    viewMode: ViewMode,
    starCatalog: StarCatalog,
    currentSystem: StarSystem,
    shipLevels: ShipLevels,
    shipNavigation: ShipNavigation,

    -- Crew assignment system (NEW)
    crew: [CrewMember],           -- Crew roster with assignments
    counselorAssigned: bool       -- Is a Counselor role filled?
}
```

### Step Function Updates (in `sim/step.ail`)

```ailang
-- In step function, update crew each frame:
export pure func step(world: World, input: FrameInput) -> FrameOutput {
    -- ... existing logic ...

    -- Update crew state (NEW)
    let deltaTime = 1.0 / 60.0;  -- Assuming 60 FPS
    let counselorFilled = isCounselorAssigned(world.crew);
    let updatedCrew = updateAllCrew(world.crew, deltaTime, counselorFilled);
    let maturingCrew = developAllSpecializations(updatedCrew);

    let updatedWorld = {world | crew: maturingCrew};

    -- ... rest of step logic ...
}

-- Helper: check if Counselor role is filled
pure func isCounselorAssigned(crew: [CrewMember]) -> bool {
    matchList crew {
        [] => false,
        head :: tail => if isRoleCounselor(head.assignedRole) then
            true
        else
            isCounselorAssigned(tail)
    }
}

-- Helper: is a given role the Counselor role?
pure func isRoleCounselor(role: Role) -> bool {
    match role {
        RoleCounselor => true,
        _ => false
    }
}

-- Helper: update all crew members
pure func updateAllCrew(crew: [CrewMember], deltaTime: float, isCounseled: bool) -> [CrewMember] {
    match crew {
        [] => [],
        head :: tail => (updateCrewState(head, deltaTime, isCounseled)) :: updateAllCrew(tail, deltaTime, isCounseled)
    }
}

-- Helper: develop specializations for all crew
pure func developAllSpecializations(crew: [CrewMember]) -> [CrewMember] {
    match crew {
        [] => [],
        head :: tail => (developSpecialization(head)) :: developAllSpecializations(tail)
    }
}
```

---

## Engine Integration

### Rendering (in `engine/render/draw.go`)

The engine will render crew assignment UI:

1. **Crew Roster Panel**
   - List of crew with assigned roles
   - Color-coded by role (engineering = blue, science = green, etc.)
   - Stress/morale bars for each crew member

2. **Assignment Dialog**
   - When player clicks a crew member, show assignment options
   - Highlight "preferred role" in yellow
   - Show efficiency bonus and stress warnings

3. **Ship Status Display**
   - Overall crew morale (average)
   - Critical stress warnings (crew > 80 stress)
   - Unassigned roles highlight

### Go Types (in `engine/crew_types.go` - NEW FILE)

```go
package engine

// CrewRoleColor returns UI color for a role
func CrewRoleColor(roleIndex int) color.Color {
    switch roleIndex {
    case 0: // Captain - gold
        return color.RGBA{255, 215, 0, 255}
    case 1: // Chief Engineer - blue
        return color.RGBA{0, 119, 182, 255}
    case 2: // Chief Scientist - green
        return color.RGBA{34, 177, 76, 255}
    case 3: // Chief Medical Officer - red
        return color.RGBA{204, 41, 48, 255}
    case 4: // Navigator - purple
        return color.RGBA{128, 0, 128, 255}
    case 5: // Security Chief - orange
        return color.RGBA{255, 127, 39, 255}
    case 6: // Botanist - light green
        return color.RGBA{144, 238, 144, 255}
    case 7: // Counselor - pink
        return color.RGBA{255, 105, 180, 255}
    case 8: // Generalist - gray
        return color.RGBA{128, 128, 128, 255}
    default:
        return color.White
    }
}

// CrewRoleName returns display name for a role
func CrewRoleName(roleIndex int) string {
    switch roleIndex {
    case 0: return "Captain"
    case 1: return "Chief Engineer"
    case 2: return "Chief Scientist"
    case 3: return "Chief Medical Officer"
    case 4: return "Navigator"
    case 5: return "Security Chief"
    case 6: return "Botanist"
    case 7: return "Counselor"
    case 8: return "Generalist"
    default: return "Unknown"
    }
}

// MoraleColor returns color based on morale value
func MoraleColor(morale float64) color.Color {
    if morale >= 85 {
        return color.RGBA{34, 177, 76, 255} // Green - Excellent
    } else if morale >= 70 {
        return color.RGBA{144, 238, 144, 255} // Light green - Good
    } else if morale >= 50 {
        return color.RGBA{255, 215, 0, 255} // Gold - Neutral
    } else if morale >= 30 {
        return color.RGBA{255, 127, 39, 255} // Orange - Poor
    } else {
        return color.RGBA{204, 41, 48, 255} // Red - Critical
    }
}

// StressColor returns color based on stress value
func StressColor(stress float64) color.Color {
    if stress <= 20 {
        return color.RGBA{34, 177, 76, 255} // Green - Calm
    } else if stress <= 40 {
        return color.RGBA{144, 238, 144, 255} // Light green - Alert
    } else if stress <= 60 {
        return color.RGBA{255, 215, 0, 255} // Gold - Tense
    } else if stress <= 80 {
        return color.RGBA{255, 127, 39, 255} // Orange - Strained
    } else {
        return color.RGBA{204, 41, 48, 255} // Red - Critical
    }
}
```

### Input Handling (in `engine/input/keys.go`)

Add hotkeys for crew management:
- **C** - Open crew roster
- **1-9** - Quick-assign highlighted crew to role
- **Tab** - Cycle through unassigned crew
- **Enter** - Confirm assignment

---

## Performance Considerations

### Recursion Depth

The crew assignment system uses recursive list operations:

```ailang
-- updateAllCrew: O(n) crew members, recursion depth = n
-- Concern: With ~50 crew max, depth = 50 (acceptable)

-- developAllSpecializations: O(n), depth = 50 (acceptable)

-- isCounselorAssigned: O(n) worst case, depth = 50 (acceptable)
```

**Conclusion:** No performance issues with typical crew roster (50 members max).

### Memory Impact

New fields per crew member:
- `assignedRole: Role` - 1 byte (enum index)
- `specialization: int` - 4 bytes
- `stress: float` - 8 bytes
- `morale: float` - 8 bytes
- `yearsAtRole: float` - 8 bytes
- `previousRoles: [Role]` - pointer + dynamic list
- `preferredRole: Option[Role]` - 1-2 bytes + option wrapper
- `concerns: [string]` - pointer + dynamic list

**Per crew member:** ~50-70 bytes additional (acceptable).
**Total for 50 crew:** ~2.5-3.5 KB (negligible).

---

## Testing

### Unit Tests (in `sim/crew_test.ail` - NEW FILE)

```ailang
module sim/crew_test

import sim/crew (
    Role, CrewMember, isSuitableFor, roleStressMultiplier,
    roleEfficiencyBonus, assignCrew, updateCrewState,
    developSpecialization, getMoraleDescription, getStressDescription
)
import sim/protocol (Coord)
import sim/npc_ai (PatternStatic)
import std/option (Some, None)

-- Test: Crew suitable for preferred role
test crewtypeIsSuitableForPreferred() {
    let crew = CrewMember {
        id: 1,
        name: "Alice",
        pos: Coord(5, 5),
        pattern: PatternStatic(),
        assignedRole: RoleGeneralist,
        specialization: 0,
        stress: 50.0,
        morale: 70.0,
        yearsAtRole: 0.0,
        previousRoles: [],
        preferredRole: Some(RoleChiefEngineer),
        concerns: []
    };

    assert isSuitableFor(crew, RoleChiefEngineer) == true;
    assert isSuitableFor(crew, RoleCaptain) == false;
}

-- Test: Stress multipliers are correct
test roleStressMultipliers() {
    assert roleStressMultiplier(RoleCaptain) == 0.8;
    assert roleStressMultiplier(RoleChiefEngineer) == 1.5;
    assert roleStressMultiplier(RoleGeneralist) == 0.9;
}

-- Test: Efficiency bonus increases with specialization
test efficiencyBonusScaling() {
    assert roleEfficiencyBonus(RoleCaptain, 0) == 20.0;
    assert roleEfficiencyBonus(RoleCaptain, 1) == 25.0;
    assert roleEfficiencyBonus(RoleCaptain, 5) == 45.0;
}

-- Test: Morale descriptions
test moraleDescriptions() {
    assert getMoraleDescription(90.0) == "Excellent";
    assert getMoraleDescription(70.0) == "Good";
    assert getMoraleDescription(50.0) == "Neutral";
    assert getMoraleDescription(30.0) == "Poor";
    assert getMoraleDescription(10.0) == "Critical";
}

-- Test: Stress descriptions
test stressDescriptions() {
    assert getStressDescription(10.0) == "Calm";
    assert getStressDescription(30.0) == "Alert";
    assert getStressDescription(50.0) == "Tense";
    assert getStressDescription(70.0) == "Strained";
    assert getStressDescription(90.0) == "Critical";
}
```

### Integration Tests

1. **Crew Roster UI** - Verify assignment dialog opens/closes
2. **Role Assignment** - Assign crew to each role, verify state updates
3. **Morale Changes** - Assign crew to non-preferred roles, verify morale decreases
4. **Stress Accumulation** - Leave crew in high-stress role, verify stress increases
5. **Counselor Effect** - Verify Counselor role reduces stress for other crew
6. **Specialization Growth** - Assign crew to role for 10+ weeks, verify specialization increases

---

## Design Decisions

### Why Per-Role Specialization (Not Global Skills)

**Option A (Chosen):** Specialization is role-specific (0-5 levels per role)
- Crew in engineering for 10 weeks gains 1 level in engineering specialization
- If moved to science, they start at 0 science specialization
- Fits "Choices Are Final" - switching roles costs accumulated expertise

**Option B (Rejected):** Global skill tree (e.g., +1 repair, +1 observation, etc.)
- Problematic: Creates "best build" crew that dominates all roles
- Violates "We Are Not Built For This" - crew would become too specialized/machine-like
- Less thematic: real crew adapt over time, not accumulate universal powers

### Why Stress Accumulates Per Role

**Option A (Chosen):** Each role has different stress multiplier (engineer 1.5x, botanist 1.0x)
- Realistic: high-stakes roles (medicine, security) are more stressful
- Creates meaningful choices: efficient role vs. sustainable role
- Counselor role specifically reduces stress recovery - creates role interdependency

**Option B (Rejected):** Single stress rate for all roles
- Too simple, removes crew management depth
- Doesn't account for realistic variation in job stress

### Why Morale Affects Efficiency

**Chosen:** Low morale crew work less effectively; high morale crew work better
- Thematic: people do better work when happy
- Creates feedback loop: misassigned crew → low morale → low efficiency → more problems

**Not Implemented:** Morale as "mutiny risk"
- Possible in v0.5.0+ if crew relationships mature
- Currently too early; need crew personality system first

---

## AILANG Constraints & Workarounds

### Constraint: No Mutable State

**Issue:** Updating crew state each frame needs to accumulate stress, morale, etc.

**Workaround:** Functional updates in step function
```ailang
let updatedCrew = updateAllCrew(world.crew, deltaTime, counselorFilled);
let maturingCrew = developAllSpecializations(updatedCrew);
let finalWorld = {world | crew: maturingCrew};
```

**Why This Works:** AILANG's functional style forces explicit state threading, which is actually great for determinism.

### Constraint: No Loops (Only Recursion)

**Issue:** Updating N crew members requires recursion.

**Workaround:** List recursion with pattern matching
```ailang
pure func updateAllCrew(crew: [CrewMember], deltaTime: float, isCounseled: bool) -> [CrewMember] {
    match crew {
        [] => [],
        head :: tail => (updateCrewState(head, deltaTime, isCounseled)) :: updateAllCrew(tail, deltaTime, isCounseled)
    }
}
```

**Risk:** With 50 crew, recursion depth = 50 (very safe; typical stack limit is thousands).

### Constraint: Limited Comparison Operators

**Issue:** Role comparison for `isSuitableFor` requires comparing ADT variants.

**Workaround:** Pattern matching with exhaustive case analysis
```ailang
match (pref, role) {
    (RoleCaptain, RoleCaptain) => true,
    (RoleChiefEngineer, RoleChiefEngineer) => true,
    -- ... etc
    _ => false
}
```

**Why This Works:** Compiler ensures all cases handled; no runtime comparisons needed.

---

## Future Extensions (v0.5.0+)

1. **Crew Relationships** - Crew prefer working with crew they like
2. **Crew Dialogue** - Crew comment on assignments ("I hate engineering," "I thrive on research")
3. **Crew Conflicts** - Security Chief and Counselor may have tensions if morale is low
4. **Crew Families** - Families prefer same role assignments (work together, or separation anxiety)
5. **Mutiny System** - If morale goes critical for too long, crew can mutiny
6. **Generational Handoff** - Children of crew inherit some parent's skills
7. **Captain Stress** - Captain takes stress hit if crew morale is low
8. **Specialization Degradation** - Skills fade after 5+ years in different role

---

## Success Criteria

- [ ] `sim/crew.ail` compiles without errors
- [ ] All crew types defined with proper defaults
- [ ] Assignment function validates and tracks history
- [ ] Stress/morale update each frame correctly
- [ ] Specialization accrues at correct rate (1 level per 10 weeks)
- [ ] Counselor role reduces stress for other crew
- [ ] UI renders crew roster with color-coded roles
- [ ] Assignment dialog opens on crew click, shows preferred role
- [ ] Hotkeys (C, 1-9, Tab, Enter) work for crew management
- [ ] Integration tests pass (morale changes, stress accumulation, etc.)
- [ ] Performance acceptable with 50 crew (no FPS drops)

---

## Related Features

- [Interior Demo Iteration](interior-demo-iteration.md) - Shows crew in ship interior
- [AILANG Input Helpers](ailang-input-helpers.md) - Input system for crew UI
- Ship Relationships (planned v0.5.0) - Crew dialogue and personality

---

*Design document created 2026-01-15. Approved for sprint planning.*
