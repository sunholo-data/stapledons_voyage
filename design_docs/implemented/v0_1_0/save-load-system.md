# Save/Load System

**Version:** 0.5.0
**Status:** Implemented
**Priority:** P1 (Engine Foundation)
**Complexity:** Medium
**Dependencies:** None
**AILANG Impact:** None - Engine serializes World struct, AILANG defines the struct

## Problem Statement

**Current State:**
- Game state is lost on exit
- No save files
- No way to resume gameplay

**What's Needed:**
- Serialize World state to JSON
- Save to file on demand or autosave
- Load saved game on startup
- Handle save file versioning

**AILANG Interface:**
```
FrameInput{
    ActionRequested: ActionSave{Slot: 1}
    ActionRequested: ActionLoad{Slot: 1}
}
```

## Design

### Save File Format

**Location:** `~/.local/share/stapledons_voyage/saves/` (Linux/Mac) or `%APPDATA%/stapledons_voyage/saves/` (Windows)

**File Structure:**
```
saves/
├── slot_1.json
├── slot_2.json
├── slot_3.json
├── autosave.json
└── meta.json        # save slot metadata
```

**Save File Contents:**
```json
{
  "version": "0.5.0",
  "timestamp": "2025-12-01T15:30:00Z",
  "playtime_seconds": 3600,
  "world": {
    "tick": 12345,
    "planet": {...},
    "npcs": [...],
    "selection": {...},
    "camera": {...}
  }
}
```

**Metadata File:**
```json
{
  "slots": {
    "1": {"timestamp": "...", "playtime": 3600, "description": "Year 2350"},
    "2": null,
    "3": {"timestamp": "...", "playtime": 7200, "description": "Year 5000"}
  },
  "lastPlayed": 1
}
```

### Save Manager

```go
package save

type Manager struct {
    basePath string
    meta     *Metadata
}

type SaveFile struct {
    Version   string          `json:"version"`
    Timestamp time.Time       `json:"timestamp"`
    Playtime  int             `json:"playtime_seconds"`
    World     sim_gen.World   `json:"world"`
}

func NewManager() (*Manager, error)
func (m *Manager) Save(slot int, world sim_gen.World) error
func (m *Manager) Load(slot int) (*sim_gen.World, error)
func (m *Manager) Autosave(world sim_gen.World) error
func (m *Manager) ListSlots() []SlotInfo
func (m *Manager) DeleteSlot(slot int) error
```

### Serialization

**World struct must be JSON-serializable:**
```go
// sim_gen/types.go - all fields exported with json tags
type World struct {
    Tick      int         `json:"tick"`
    Planet    PlanetState `json:"planet"`
    NPCs      []NPC       `json:"npcs"`
    Selection Selection   `json:"selection"`
    Camera    Camera      `json:"camera"`
}
```

**Interface Types (Selection, etc.):**
```go
// Use discriminated union pattern for JSON
type SelectionJSON struct {
    Type string `json:"type"`
    X    int    `json:"x,omitempty"`
    Y    int    `json:"y,omitempty"`
}

func (s Selection) MarshalJSON() ([]byte, error)
func (s *Selection) UnmarshalJSON(data []byte) error
```

### Autosave

**Triggers:**
- Every N minutes (configurable, default 5)
- On mode transitions
- Before risky actions (journey commit)

**Implementation:**
```go
func (g *Game) Update() error {
    g.playtime += 1.0/60.0  // assuming 60fps

    if g.playtime - g.lastAutosave > g.autosaveInterval {
        g.saveManager.Autosave(g.world)
        g.lastAutosave = g.playtime
    }
    ...
}
```

### Version Migration

**When save version < game version:**
```go
func (m *Manager) migrate(save *SaveFile) error {
    switch save.Version {
    case "0.4.0":
        // Migrate 0.4.0 → 0.5.0
        save.World = migrate_0_4_to_0_5(save.World)
        save.Version = "0.5.0"
        fallthrough
    case "0.5.0":
        // Current version, no migration needed
        return nil
    default:
        return fmt.Errorf("unknown save version: %s", save.Version)
    }
}
```

## AILANG Integration

**New Action Types:**
```go
// sim_gen/types.go
type ActionSave struct {
    Slot int
}
func (ActionSave) isPlayerAction() {}

type ActionLoad struct {
    Slot int
}
func (ActionLoad) isPlayerAction() {}
```

**Game Loop Handling:**
```go
func (g *Game) Update() error {
    ...
    switch action := input.ActionRequested.(type) {
    case sim_gen.ActionSave:
        if err := g.saveManager.Save(action.Slot, g.world); err != nil {
            g.showError("Save failed: " + err.Error())
        } else {
            g.showMessage("Game saved")
        }
    case sim_gen.ActionLoad:
        world, err := g.saveManager.Load(action.Slot)
        if err != nil {
            g.showError("Load failed: " + err.Error())
        } else {
            g.world = *world
            g.showMessage("Game loaded")
        }
    }
    ...
}
```

## Implementation Plan

### Files to Create
| File | Purpose |
|------|---------|
| `engine/save/manager.go` | Save/load logic |
| `engine/save/migration.go` | Version migration |
| `engine/save/serialize.go` | JSON marshaling helpers |

### Files to Modify
| File | Change |
|------|--------|
| `sim_gen/types.go` | Add JSON tags, ActionSave/ActionLoad |
| `cmd/game/main.go` | Initialize SaveManager, handle save/load actions |

## Testing Strategy

### Unit Tests
```go
func TestSaveRoundTrip(t *testing.T)
func TestMigration(t *testing.T)
func TestCorruptedSave(t *testing.T)
func TestMissingSlot(t *testing.T)
```

### Integration Tests
```bash
# Save, exit, reload, verify state matches
make run-mock
# Press F5 to save
# Exit and restart
# Press F9 to load
```

## Success Criteria

- [ ] World state serializes to JSON
- [ ] Save files persist to disk
- [ ] Load restores exact game state
- [ ] Autosave triggers periodically
- [ ] Version migration works
- [ ] Corrupted saves handled gracefully
- [ ] Save slot metadata displays correctly

---

**Created:** 2025-12-01
