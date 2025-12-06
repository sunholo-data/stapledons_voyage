# Input Rebinding

**Version:** 0.5.0
**Status:** Planned
**Priority:** P2 (Polish)
**Complexity:** Low
**Dependencies:** Save/Load System (for persisting bindings)
**AILANG Impact:** None - Engine maps keys to logical actions, AILANG receives actions

## Problem Statement

**Current State:**
- Hardcoded key bindings in input capture
- No way for players to customize controls
- WASD/Arrow keys baked into engine code

**What's Needed:**
- Logical action names (MoveUp, ZoomIn, Select)
- Configurable key-to-action mapping
- Persist bindings to config file
- Rebinding UI (future - can be AILANG-driven)

**AILANG Interface:**
```
FrameInput{
    Actions: [ActionMoveUp, ActionZoomIn]  // logical actions, not raw keys
    RawKeys: [...]                         // still available for AILANG if needed
}
```

## Design

### Logical Actions

**Action enum:**
```go
type Action int

const (
    ActionNone Action = iota

    // Movement
    ActionMoveUp
    ActionMoveDown
    ActionMoveLeft
    ActionMoveRight

    // Camera
    ActionZoomIn
    ActionZoomOut
    ActionCameraUp
    ActionCameraDown
    ActionCameraLeft
    ActionCameraRight

    // Selection
    ActionSelect       // Primary click
    ActionSelectAlt    // Secondary click
    ActionCancel

    // UI
    ActionPause
    ActionMenu
    ActionConfirm

    // Modes
    ActionToggleMode   // M key currently

    // Debug
    ActionDebugOverlay
    ActionScreenshot
)
```

### Default Bindings

```go
var DefaultBindings = map[Action][]ebiten.Key{
    // Movement (WASD + Arrows)
    ActionMoveUp:    {ebiten.KeyW, ebiten.KeyArrowUp},
    ActionMoveDown:  {ebiten.KeyS, ebiten.KeyArrowDown},
    ActionMoveLeft:  {ebiten.KeyA, ebiten.KeyArrowLeft},
    ActionMoveRight: {ebiten.KeyD, ebiten.KeyArrowRight},

    // Camera (IJKL or numpad)
    ActionCameraUp:    {ebiten.KeyI, ebiten.KeyNumpad8},
    ActionCameraDown:  {ebiten.KeyK, ebiten.KeyNumpad2},
    ActionCameraLeft:  {ebiten.KeyJ, ebiten.KeyNumpad4},
    ActionCameraRight: {ebiten.KeyL, ebiten.KeyNumpad6},

    // Zoom
    ActionZoomIn:  {ebiten.KeyEqual, ebiten.KeyNumpadAdd},
    ActionZoomOut: {ebiten.KeyMinus, ebiten.KeyNumpadSubtract},

    // UI
    ActionSelect:    {}, // Mouse button, handled separately
    ActionSelectAlt: {},
    ActionCancel:    {ebiten.KeyEscape},
    ActionPause:     {ebiten.KeyP, ebiten.KeySpace},
    ActionMenu:      {ebiten.KeyTab},
    ActionConfirm:   {ebiten.KeyEnter, ebiten.KeyNumpadEnter},

    // Mode
    ActionToggleMode: {ebiten.KeyM},

    // Debug
    ActionDebugOverlay: {ebiten.KeyF3},
    ActionScreenshot:   {ebiten.KeyF12},
}
```

### Input Manager

```go
package input

type Manager struct {
    bindings     map[Action][]ebiten.Key
    mouseActions map[Action]ebiten.MouseButton
    activeKeys   map[ebiten.Key]bool
    justPressed  map[Action]bool  // true only on first frame
    held         map[Action]bool  // true while held
}

func NewManager() *Manager
func (m *Manager) LoadBindings(path string) error
func (m *Manager) SaveBindings(path string) error
func (m *Manager) SetBinding(action Action, keys []ebiten.Key)
func (m *Manager) GetBinding(action Action) []ebiten.Key

// Per-frame methods
func (m *Manager) Update()
func (m *Manager) IsPressed(action Action) bool    // true on first frame only
func (m *Manager) IsHeld(action Action) bool       // true while held
func (m *Manager) IsReleased(action Action) bool   // true when just released
func (m *Manager) GetActiveActions() []Action      // all currently active
```

### Bindings File Format

**Location:** `~/.config/stapledons_voyage/bindings.json`

```json
{
    "version": "1",
    "bindings": {
        "MoveUp": ["W", "ArrowUp"],
        "MoveDown": ["S", "ArrowDown"],
        "ZoomIn": ["Equal", "NumpadAdd"],
        "Pause": ["P"]
    }
}
```

### Key Name Mapping

```go
var keyNames = map[string]ebiten.Key{
    "A": ebiten.KeyA,
    "B": ebiten.KeyB,
    // ... all letters
    "0": ebiten.Key0,
    // ... all numbers
    "ArrowUp": ebiten.KeyArrowUp,
    "ArrowDown": ebiten.KeyArrowDown,
    "ArrowLeft": ebiten.KeyArrowLeft,
    "ArrowRight": ebiten.KeyArrowRight,
    "Space": ebiten.KeySpace,
    "Enter": ebiten.KeyEnter,
    "Escape": ebiten.KeyEscape,
    "Tab": ebiten.KeyTab,
    "Shift": ebiten.KeyShift,
    "Control": ebiten.KeyControl,
    "Alt": ebiten.KeyAlt,
    // ... etc
}

var keyToName map[ebiten.Key]string  // reverse lookup
```

### Integration with FrameInput

**Updated FrameInput:**
```go
type FrameInput struct {
    // Existing
    Mouse            MouseState
    Keys             []KeyEvent
    ClickedThisFrame bool
    WorldMouseX      float64
    WorldMouseY      float64
    ActionRequested  PlayerAction
    TestMode         bool

    // New - logical actions
    Actions          []Action      // actions active this frame
    ActionsPressed   []Action      // actions just pressed
    ActionsReleased  []Action      // actions just released
}
```

**CaptureInput update:**
```go
func CaptureInput(inputMgr *input.Manager) sim_gen.FrameInput {
    inputMgr.Update()

    return sim_gen.FrameInput{
        // ... existing fields ...
        Actions:        inputMgr.GetActiveActions(),
        ActionsPressed: inputMgr.GetJustPressed(),
        ActionsReleased: inputMgr.GetJustReleased(),
    }
}
```

### Rebinding Flow (Future UI)

```
1. Player opens Settings → Controls
2. Clicks on action name (e.g., "Move Up")
3. Engine enters "listening" mode
4. Player presses desired key
5. Engine captures key, updates binding
6. Binding saved to file
```

**Rebind detection:**
```go
func (m *Manager) StartRebind(action Action)
func (m *Manager) IsRebinding() bool
func (m *Manager) CancelRebind()
func (m *Manager) GetRebindAction() Action

// In Update():
if m.rebinding {
    for key := ebiten.Key(0); key <= ebiten.KeyMax; key++ {
        if inpututil.IsKeyJustPressed(key) {
            m.SetBinding(m.rebindAction, []ebiten.Key{key})
            m.rebinding = false
            break
        }
    }
}
```

## Implementation Plan

### Files to Create
| File | Purpose |
|------|---------|
| `engine/input/manager.go` | Input manager with action mapping |
| `engine/input/bindings.go` | Load/save binding configuration |
| `engine/input/keys.go` | Key name ↔ ebiten.Key conversion |

### Files to Modify
| File | Change |
|------|--------|
| `sim_gen/protocol.go` | Add Actions to FrameInput |
| `engine/input.go` | Use InputManager instead of direct key checks |
| `cmd/game/main.go` | Initialize InputManager |

## Testing Strategy

### Manual Test
```bash
make run-mock
# 1. Verify default bindings work
# 2. Edit bindings.json manually
# 3. Restart, verify new bindings work
```

### Unit Tests
```go
func TestDefaultBindings(t *testing.T)
func TestLoadSaveBindings(t *testing.T)
func TestMultipleKeysPerAction(t *testing.T)
func TestKeyNameConversion(t *testing.T)
```

## Success Criteria

- [ ] Logical actions replace hardcoded keys
- [ ] Multiple keys can map to same action
- [ ] Bindings persist to config file
- [ ] Default bindings work out of box
- [ ] Invalid config handled gracefully
- [ ] Key names human-readable in config
- [ ] AILANG receives actions, not raw keys

---

**Created:** 2025-12-01
