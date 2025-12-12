# Demo & CLI Reference

**Status**: Active (Updated 2025-12-12)
**Purpose**: Index of all demo binaries and CLI tools
**Location**: All demos build to `bin/` directory

---

## Naming Convention

| Prefix | Type | Description |
|--------|------|-------------|
| `demo-game-*` | AILANG-driven | Tests game features (uses `sim_gen`) |
| `demo-engine-*` | Pure Go | Tests engine features (rendering, shaders, 3D) |
| `gen-*` | Asset generator | Creates sprites, sounds, textures |
| `gr*` | GR effect tools | Black hole animations, screenshots |

---

## Quick Reference

### Game Demos (AILANG-driven)

| Demo | Feature | Run Command |
|------|---------|-------------|
| `demo-game-bridge` | Bridge view + dome viewport | `go run ./cmd/demo-game-bridge` |
| `demo-game-orbital` | Celestial system, planet orbits | `go run ./cmd/demo-game-orbital` |
| `demo-game-parallax` | 20-layer depth system | `go run ./cmd/demo-game-parallax` |
| `demo-game-starmap` | Local stellar neighborhood | `go run ./cmd/demo-game-starmap` |

### Engine Demos (Pure Go)

| Demo | Feature | Run Command |
|------|---------|-------------|
| `demo-engine-ai` | AI capabilities (text/image/TTS) | `go run ./cmd/demo-engine-ai` |
| `demo-engine-lookat` | Camera LookAt diagnostics | `go run ./cmd/demo-engine-lookat` |
| `demo-engine-tetra` | Tetra3D 3D rendering | `go run ./cmd/demo-engine-tetra` |
| `demo-engine-solar` | 3D planet compositing | `go run ./cmd/demo-engine-solar` |

### Asset Generators

| Tool | Purpose | Run Command |
|------|---------|-------------|
| `gen-ring-texture` | Saturn ring textures | `go run ./cmd/gen-ring-texture` |
| `gensounds` | Sine wave tones | `go run ./cmd/gensounds` |
| `gensprites` | Isometric tiles/entities | `go run ./cmd/gensprites` |

### GR Effect Tools

| Tool | Purpose | Run Command |
|------|---------|-------------|
| `granimation` | Black hole journey frames | `go run ./cmd/granimation` |
| `grscreenshot` | GR effect screenshots | `go run ./cmd/grscreenshot` |

### Main Binaries

| Binary | Purpose | Run Command |
|--------|---------|-------------|
| `game` | Full game | `make run` or `go run ./cmd/game` |
| `eval` | AI benchmark evaluation | `go run ./cmd/eval` |

---

## CLI Tool (`voyage`)

The main CLI provides development utilities.

```bash
go run ./cmd/cli <command> [options]
# Or after building:
voyage <command> [options]
```

### Subcommands

| Command | Uses AILANG | Description |
|---------|-------------|-------------|
| `ai` | No | Test AI handlers (Claude, Gemini, TTS) |
| `world` | Yes | Inspect world state (NPCs, tiles, planets) |
| `bench` | Yes | Performance benchmarks (human-readable) |
| `perf` | Yes | Benchmarks with threshold checks (CI/JSON) |
| `assets` | No | Validate game assets |
| `sim` | Yes | Simulation stress tests |

### AI Subcommand Examples

```bash
voyage ai -list                              # Show available providers
voyage ai -list-voices                       # Show 30 TTS voices
voyage ai -prompt "Hello!"                   # Auto-detect provider
voyage ai -provider gemini -prompt "Hi"      # Use Gemini
voyage ai -prompt "Draw a cat" -generate-image   # Generate image
voyage ai -prompt "Hello world" -tts         # Text to speech
voyage ai -prompt "Hello" -tts -voice Puck   # TTS with specific voice
```

---

## Screenshot Testing

All demos MUST support these flags for CI:

| Flag | Default | Description |
|------|---------|-------------|
| `--screenshot N` | 0 | Take screenshot after N frames |
| `--output PATH` | `out/screenshots/demo-NAME.png` | Screenshot output |

**Example:**
```bash
go run ./cmd/demo-game-orbital --screenshot 60 --output test.png
```

---

## Creating New Demos

Use the template:

```bash
cp -r cmd/demo-template cmd/demo-game-YOURNAME
mv cmd/demo-game-YOURNAME/TEMPLATE.go cmd/demo-game-YOURNAME/main.go
# Edit main.go: replace YOURNAME, DESCRIPTION
go build -o bin/demo-game-YOURNAME ./cmd/demo-game-YOURNAME
```

**Required features (DO NOT REMOVE):**
- `--screenshot N` flag
- `--output PATH` flag
- `takeScreenshot()` function
- Frame counter in HUD

---

## Feature Coverage

| Feature Area | Demo | Status |
|--------------|------|--------|
| AI (text/image/TTS) | `demo-engine-ai` | Working |
| Bridge/Viewport | `demo-game-bridge` | Working |
| Celestial/Orbital | `demo-game-orbital` | Working |
| Parallax/Depth | `demo-game-parallax` | Working |
| Starmap | `demo-game-starmap` | Working |
| 3D Planets | `demo-engine-tetra`, `demo-engine-solar` | Working |
| Galaxy Map | - | Needs demo |
| Journey/Time Dilation | - | Needs demo |
| Save System | - | Needs demo |
| NPC/Dialogue | - | Needs demo |
| Ship Levels | - | Needs demo |
| Arrival Sequence | - | Needs demo |

---

## Related Documents

- [engine-capabilities.md](engine-capabilities.md) - Go engine features
- [game-capabilities.md](game-capabilities.md) - AILANG game features
- [ai-capabilities.md](ai-capabilities.md) - AI features (voices, image gen)

---

**Document created**: 2025-12-12
**Last updated**: 2025-12-12
