# Audio System

**Version:** 0.2.0
**Status:** Planned
**Priority:** P1 (Medium)
**Complexity:** Medium
**Package:** `engine/audio`

## Related Documents

- [Asset Management](asset-management.md) - Sound file loading
- [Engine Layer Design](../../implemented/v0_1_0/engine-layer.md) - Integration context

## Problem Statement

The game has no audio. Sound effects and music enhance gameplay and provide feedback for player actions.

**Current State:**
- No audio support in v0.1.0
- FrameOutput has no sound data

**What's Needed:**
- Play sound effects triggered by game events
- Background music/ambient audio
- Volume control
- AILANG integration via FrameOutput

## AILANG Integration

### Protocol Extension (sim/protocol.ail)

```ailang
type SoundCmd =
    | PlaySound(int, float)      -- id, volume (0.0-1.0)
    | StopSound(int)             -- id
    | PlayMusic(int, float)      -- id, volume

type FrameOutput = {
    draw_cmds: [DrawCmd],
    sounds: [SoundCmd]           -- NEW: sound commands per frame
}
```

### Usage in Game Logic

```ailang
-- In step.ail
let sounds = if player_moved then
    [PlaySound(1, 1.0)]  -- footstep sound
else
    []
```

## Go Implementation

### Audio Manager

```go
package audio

type Manager struct {
    context    *audio.Context
    sfx        map[int]*audio.Player   // Sound effects (short, can overlap)
    music      *audio.Player           // Current music (one at a time)
    masterVol  float64
}

func NewManager(sampleRate int) (*Manager, error)
func (m *Manager) LoadSound(id int, path string) error
func (m *Manager) ProcessCommands(cmds []sim_gen.SoundCmd)
func (m *Manager) SetMasterVolume(vol float64)
```

### Sound Processing Per Frame

```go
func (m *Manager) ProcessCommands(cmds []sim_gen.SoundCmd) {
    for _, cmd := range cmds {
        switch cmd.Kind {
        case sim_gen.SoundCmdKindPlaySound:
            m.playSound(cmd.PlaySound.ID, cmd.PlaySound.Volume)
        case sim_gen.SoundCmdKindStopSound:
            m.stopSound(cmd.StopSound.ID)
        case sim_gen.SoundCmdKindPlayMusic:
            m.playMusic(cmd.PlayMusic.ID, cmd.PlayMusic.Volume)
        }
    }
}
```

### Game Loop Integration

```go
func (g *Game) Update() error {
    input := render.CaptureInput()
    w2, out, err := sim_gen.Step(g.world, input)
    g.world = w2
    g.out = out
    g.audio.ProcessCommands(out.Sounds)  // NEW
    return err
}
```

## Audio Design Constraints

### Ebiten Audio Limitations

| Constraint | Impact | Solution |
|------------|--------|----------|
| Single audio context | Must share across sounds | Create once at init |
| Sample rate fixed | All sounds same rate | Standardize on 44100Hz |
| Streaming for large files | Music needs streaming | Use `audio.NewInfiniteLoop` |

### Sound Categories

| Category | Behavior | Example |
|----------|----------|---------|
| SFX | Short, can overlap | footsteps, clicks |
| Music | Long, one at a time | background music |
| Ambient | Looping, layered | wind, water |

## File Formats

| Format | Use Case | Ebiten Support |
|--------|----------|----------------|
| WAV | Sound effects | Native |
| OGG | Music, ambient | Via vorbis package |
| MP3 | Music | Via mp3 package |

**Recommendation:** WAV for SFX (low latency), OGG for music (good compression).

## Implementation Plan

### Files to Create

| File | Purpose |
|------|---------|
| `engine/audio/manager.go` | Core audio manager |
| `engine/audio/sfx.go` | Sound effect handling |
| `engine/audio/music.go` | Background music |

### AILANG Changes

| File | Change |
|------|--------|
| `sim/protocol.ail` | Add SoundCmd type |
| `sim/protocol.ail` | Add sounds field to FrameOutput |

### Go Integration

| File | Change |
|------|--------|
| `cmd/game/main.go` | Initialize audio manager |
| `cmd/game/main.go` | Call ProcessCommands in Update |

## Testing Strategy

### Manual Testing

```bash
make run  # Listen for sounds on game events
```

### Automated Testing

```go
func TestPlaySound(t *testing.T)
func TestVolumeControl(t *testing.T)
func TestMusicSwitch(t *testing.T)
```

### Edge Cases

- [ ] Sound ID not loaded → log warning, skip
- [ ] Volume out of range → clamp to 0.0-1.0
- [ ] Rapid fire sounds → queue or skip duplicates
- [ ] Music already playing → crossfade or hard switch

## Success Criteria

### Core Functionality
- [ ] Sound effects play on command
- [ ] Background music loops correctly
- [ ] Volume control works (0.0-1.0)
- [ ] Multiple SFX can overlap

### AILANG Integration
- [ ] SoundCmd type compiles
- [ ] FrameOutput.Sounds populated by step()
- [ ] Engine processes sound commands each frame

### Performance
- [ ] No audio lag/delay perceptible
- [ ] Sound loading doesn't block gameplay
- [ ] Memory usage reasonable for loaded sounds
