# Development Guide

Technical reference for Stapledons Voyage development. For AI/Claude Code guidance, see [CLAUDE.md](CLAUDE.md).

## Repository Structure

```
stapledons_voyage/
├── cmd/
│   ├── game/                 # Ebiten game runtime (main window loop)
│   │   └── main.go
│   └── eval/                 # Benchmark + scenario runner CLI
│       └── main.go
│
├── sim/                      # AILANG source code (future)
│   ├── world.ail
│   ├── protocol.ail
│   ├── step.ail
│   └── npc_ai.ail
│
├── sim_gen/                  # Simulation logic (mock Go or AILANG-generated)
│   ├── types.go              # World, NPC, Tile, Action types
│   ├── protocol.go           # FrameInput, FrameOutput, DrawCmd
│   └── funcs.go              # InitWorld, Step, NPC movement
│
├── engine/                   # Pure Go engine code
│   ├── render/               # Input capture and drawing
│   │   ├── assets.go
│   │   ├── draw.go
│   │   └── input.go
│   ├── camera/               # Camera transform utilities
│   ├── assets/               # Asset manager
│   ├── scenario/             # Scenario definitions and runner
│   └── bench/                # Benchmarks
│
├── assets/                   # Sprites, tilesheets, fonts
├── design_docs/              # Feature design documentation
├── sprints/                  # Sprint tracking JSON files
├── ailang_resources/         # AILANG consumer contracts
├── out/                      # Generated output (reports, screenshots)
│
├── Makefile
├── go.mod
├── README.md                 # User-facing documentation
├── CLAUDE.md                 # AI/Claude Code guidance
└── DEVELOPMENT.md            # This file
```

## Build Commands

```bash
# Using mock sim_gen (current development)
make run-mock      # Run game with mock simulation
make eval-mock     # Run benchmarks and scenarios
make game-mock     # Build executable to bin/game

# When AILANG compiler is available
make sim           # Compile AILANG → Go
make game          # Build with AILANG-generated code
make eval          # Run full evaluation
make run           # Run with AILANG code

# Maintenance
make clean         # Remove generated artifacts
go test ./...      # Run all tests
```

## Data Flow

```
┌─────────────────────────────────────────────────────────────────┐
│                         Game Loop                                │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│   ┌──────────┐    ┌──────────────┐    ┌───────────────┐         │
│   │  User    │───▶│ CaptureInput │───▶│  FrameInput   │         │
│   │  Input   │    │  (engine)    │    │  (protocol)   │         │
│   └──────────┘    └──────────────┘    └───────┬───────┘         │
│                                               │                  │
│                                               ▼                  │
│   ┌──────────┐    ┌──────────────┐    ┌───────────────┐         │
│   │  Screen  │◀───│ RenderFrame  │◀───│ FrameOutput   │         │
│   │          │    │  (engine)    │    │  (protocol)   │         │
│   └──────────┘    └──────────────┘    └───────┬───────┘         │
│                                               │                  │
│                                               ▼                  │
│                                       ┌───────────────┐         │
│                                       │    Step()     │         │
│                                       │  (sim_gen)    │         │
│                                       └───────────────┘         │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

## Key Types

### Protocol Types (sim_gen/protocol.go)

```go
type FrameInput struct {
    Mouse            MouseState
    Keys             []KeyEvent
    ClickedThisFrame bool
    WorldMouseX      float64
    WorldMouseY      float64
    ActionRequested  PlayerAction
}

type FrameOutput struct {
    Draw   []DrawCmd
    Sounds []int
    Debug  []string
    Camera Camera
}

type DrawCmd interface { isDrawCmd() }
type DrawCmdRect struct { X, Y, W, H float64; Color, Z int }
type DrawCmdSprite struct { ID int; X, Y float64; Z int }
type DrawCmdText struct { Text string; X, Y float64; Z int }
```

### World Types (sim_gen/types.go)

```go
type World struct {
    Tick      int
    Planet    PlanetState
    NPCs      []NPC
    Selection Selection
}

type Tile struct {
    Biome     int
    Structure Structure
}

type NPC struct {
    ID, X, Y, Sprite int
    Pattern     MovementPattern
    PatrolIndex int
    MoveCounter int
}
```

### Action Types

```go
type PlayerAction interface { isPlayerAction() }
type ActionInspect struct{}
type ActionBuild struct { StructureType StructureType }
type ActionClear struct{}
```

## Z-Index Layers

| Z | Content |
|---|---------|
| 0 | Terrain tiles |
| 1 | Structures (houses, farms) |
| 2 | NPCs |
| 3 | Selection highlight |
| 4+ | UI elements (future) |

## Color Indices

| Index | Color | Usage |
|-------|-------|-------|
| 0 | Blue | Water biome |
| 1 | Green | Forest biome |
| 2 | Tan | Desert biome |
| 3 | Brown | Mountain biome |
| 4 | Yellow (semi) | Selection highlight |
| 5 | Saddle brown | House structure |
| 6 | Lime green | Farm structure |
| 7 | Gray | Road structure |
| 10 | Red | NPC 0 |
| 11 | Green | NPC 1 |
| 12 | Blue | NPC 2 |

## AILANG Integration

When AILANG's Go codegen ships (`ailang compile --emit-go`):

1. Game logic moves to `sim/*.ail`
2. `make sim` generates `sim_gen/*.go`
3. Engine layer (`engine/`) stays unchanged
4. Mock code in `sim_gen/` is replaced

See [ailang_resources/consumer-contract-v0.5.md](ailang_resources/consumer-contract-v0.5.md) for the full contract.

## Testing

```bash
# Unit tests
go test ./sim_gen/...     # Simulation tests
go test ./engine/...      # Engine tests

# Full test suite
go test ./...

# Benchmarks
go test -bench=. -benchmem ./engine/bench

# Evaluation report
make eval-mock
cat out/report.json
```

## Sprint Tracking

Sprints are tracked in JSON files under `sprints/`:

```bash
ls sprints/*.json
# 001-engine-infrastructure.json
# 002-camera-viewport.json
# 003-player-interaction.json
# 004-player-actions.json
# 005-npc-movement.json
```

Use the sprint-executor skill to run sprints via Claude Code.

## Design Documents

Feature designs live in `design_docs/`:

- `planned/` - Features not yet implemented
- `implemented/` - Completed features

See [design_docs/README.md](design_docs/README.md) for the full index.

## CI/CD

GitHub Actions workflows in `.github/workflows/`:

### CI (ci.yml)

Runs on every push to `main` and on pull requests:
- Runs all tests
- Runs `go vet`
- Verifies the build compiles

### Release (release.yml)

Triggered by pushing a version tag:

```bash
# Create a release
git tag v0.1.0
git push origin v0.1.0
```

Builds binaries for:
- Linux (amd64)
- Windows (amd64)
- macOS Intel (amd64)
- macOS Apple Silicon (arm64)

Binaries are uploaded to GitHub Releases automatically.

## Creating a Release

1. Ensure all tests pass: `go test ./...`
2. Update version references if needed
3. Commit any final changes
4. Create and push a tag:
   ```bash
   git tag v0.1.0
   git push origin v0.1.0
   ```
5. GitHub Actions will build binaries and create the release
6. Edit the release notes on GitHub if desired
