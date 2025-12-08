# Sprint: View System Foundation

**Sprint ID**: view-system-v1
**Duration**: 4 days
**Design Doc**: design_docs/planned/next/01-view-system.md
**Priority**: P0 (Foundation)

## Goal

Implement the core view system that enables composable game views with three layers:
- Background layer (starfield, space effects)
- Content layer (3D planets, isometric tiles, entities)
- UI layer (HUD, panels, dialogue)

This is the foundation for all game views and enables the arrival sequence.

## Scope Assessment

### Engine Work (Primary - 90%)
- Create `engine/view/` package with interfaces
- Implement layer abstraction
- Build transition system
- Integrate with existing render pipeline

### AILANG Work (Minimal - 10%)
- Consider new DrawCmd variants for view control (if needed)
- No major AILANG changes - view system is engine-side compositing

### Dependencies
- **Requires**: Existing DrawCmd system, SR/GR shaders (both working)
- **Enables**: All game views, arrival sequence, Tetra3D integration

## Day-by-Day Breakdown

### Day 1: Core Interfaces

**Goal**: Define the view system architecture in Go

- [ ] Create `engine/view/view.go` - View interface, ViewType enum
- [ ] Create `engine/view/layer.go` - Layer interfaces (Background, Content, UI)
- [ ] Create `engine/view/manager.go` - ViewManager to coordinate views
- [ ] Unit tests for basic view lifecycle

**Files to create**:
```
engine/view/
├── view.go        # View interface, ViewType enum
├── layer.go       # Layer interfaces
├── manager.go     # ViewManager
└── view_test.go   # Unit tests
```

**AILANG check**: Run `ailang check sim/*.ail` to verify no regressions

### Day 2: Transition System

**Goal**: Implement smooth transitions between views

- [ ] Create `engine/view/transition.go` - Transition effects (fade, crossfade, wipe)
- [ ] Create `engine/view/easing.go` - Easing functions for smooth animations
- [ ] Add transition state machine to ViewManager
- [ ] Test transitions with placeholder views

**Key types**:
```go
type TransitionEffect int
const (
    TransitionNone TransitionEffect = iota
    TransitionFade
    TransitionCrossfade
    TransitionWipe
    TransitionZoom
)
```

### Day 3: Background Layer Implementation

**Goal**: Implement the space background with parallax

- [ ] Create `engine/view/background/space.go` - SpaceBackground with star layers
- [ ] Create `engine/view/background/stars.go` - Parallax star field rendering
- [ ] Integrate SR/GR shader support into background layer
- [ ] Test with existing GalaxyBg and Star DrawCmds

**Star layer configuration**:
| Layer | Stars | Parallax | Purpose |
|-------|-------|----------|---------|
| Far   | 500   | 0.0      | Fixed distant stars |
| Mid   | 300   | 0.3      | Slight motion |
| Near  | 100   | 0.7      | Foreground stars |

### Day 4: Integration & Demo

**Goal**: Wire into game and verify everything works

- [ ] Update `cmd/game/main.go` to use ViewManager
- [ ] Create `ViewSpace` implementation as first real view
- [ ] Add demo command: `./bin/game --demo-view-space`
- [ ] Verify SR/GR effects apply correctly to view layers
- [ ] Performance test: maintain 60fps during transitions
- [ ] Document any workarounds needed

**Demo command verification**:
```bash
./bin/game --demo-view-space          # Basic space view
./bin/game --demo-view-space --sr 0.3 # With SR effects
./bin/game --demo-view-transition     # Test fade between views
```

## Success Criteria

From design doc (01-view-system.md):
- [ ] Views compose background + content + UI layers
- [ ] Transitions between views are smooth
- [ ] Space background renders with parallax stars
- [ ] SR/GR effects apply to background layer
- [ ] UI panels can be added/removed dynamically
- [ ] 60fps maintained during transitions

## AILANG Feedback Checkpoint

### Pre-Sprint
- [x] Check inbox for AILANG team responses: No unread messages
- [x] Verify AILANG modules compile: protocol.ail compiles

### During Sprint
- Report any blockers to AILANG team immediately
- Document any workarounds needed for engine-AILANG boundary

### Post-Sprint
- Send summary of issues encountered (if any)
- Document lessons learned in sprint JSON

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Integration complexity with existing render | Medium | Medium | Keep interfaces simple, iterate |
| Performance impact of layer compositing | Low | Medium | Profile early, use render targets |
| AILANG DrawCmd changes needed | Low | Low | View system mostly engine-side |

## Complexity Factors

| Factor | Multiplier | Reason |
|--------|------------|--------|
| Engine-only (no AILANG changes) | 0.8x | Simpler, no cross-language boundary |
| Existing shader integration | 1.0x | Shaders already work, just wire in |
| New package structure | 1.1x | Need to design API carefully |
| **Total** | ~0.9x | Should be achievable in 4 days |

## Notes

- This is primarily Go/Ebiten engine work
- AILANG continues to generate DrawCmds as before
- The engine interprets DrawCmds and composites them through the view system
- No changes to sim_gen needed
