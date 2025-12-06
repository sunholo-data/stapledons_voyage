# CLI Development Tools

**Version:** 0.1.0
**Status:** Implemented
**Priority:** P1 (High)
**Complexity:** Medium
**Package:** `cmd/`

## Related Documents

- [AI Handler System](ai-handler-system.md) - AI backend for `voyage ai`
- [Screenshot Mode](screenshot-mode.md) - Used by `grscreenshot`
- [Test Scenarios](test-scenarios.md) - Golden file testing

## Overview

Stapledon's Voyage includes a suite of CLI tools for development, testing, and content generation. These tools are essential for:
- Testing AI integrations without running the full game
- Inspecting simulation state
- Performance benchmarking
- Asset generation and validation
- Visual effect demonstrations

## CLI Tools Summary

| Command | Binary | Source | LOC | Purpose |
|---------|--------|--------|-----|---------|
| `voyage` | `bin/voyage` | `cmd/cli/main.go` | 991 | Main development CLI |
| `game` | `bin/game` | `cmd/game/main.go` | ~300 | Main game executable |
| `granimation` | `bin/granimation` | `cmd/granimation/main.go` | 138 | GR animation frames |
| `grscreenshot` | - | `cmd/grscreenshot/main.go` | 50 | GR effect screenshots |
| `gensprites` | - | `cmd/gensprites/main.go` | 380 | Placeholder sprite generation |
| `gensounds` | - | `cmd/gensounds/main.go` | 118 | Placeholder sound generation |
| `eval` | - | `cmd/eval/main.go` | ~200 | AILANG evaluation |

**Total CLI code:** ~2,000+ LOC

## Main CLI: voyage

The `voyage` CLI is the primary development tool.

### Building

```bash
go build -o bin/voyage ./cmd/cli
# Or via Makefile:
make cli
```

### Commands

```
voyage <command> [options]

Commands:
  ai       Test AI handlers (Claude, Gemini, multimodal)
  world    Inspect world state (NPCs, tiles, planets)
  bench    Run performance benchmarks (human-readable)
  perf     Run benchmarks with threshold checks (CI/JSON output)
  assets   Validate game assets
  sim      Run simulation stress tests
  help     Show help
```

---

## voyage ai

Test AI handlers with text, images, and TTS.

### Usage

```bash
voyage ai [options]

Options:
  -provider string   AI provider: claude, gemini, auto (default: auto-detect)
  -prompt string     Text prompt to send
  -system string     System prompt
  -image string      Path to image file to include (Gemini only)
  -generate-image    Generate an image from prompt (Gemini only)
  -tts               Generate speech from prompt (Gemini only)
  -voice string      TTS voice name (default: Kore)
  -edit-image        Edit a reference image with the prompt (Gemini only)
  -reference string  Path to reference image for editing
  -v                 Verbose output
  -list              List available providers and their status
  -list-voices       List available TTS voices
```

### Examples

```bash
# List available AI providers
voyage ai -list

# Test text generation
voyage ai -prompt "Hello, how are you?"
voyage ai -provider claude -prompt "Explain quantum entanglement"
voyage ai -provider gemini -prompt "What is the meaning of life?"

# Generate images
voyage ai -prompt "A spaceship approaching a black hole" -generate-image

# Edit images (iterative refinement)
voyage ai -prompt "Make the sky more purple" -edit-image -reference out/image.png
voyage ai -prompt "Add more stars" -edit-image  # Uses last generated image

# Text-to-speech
voyage ai -prompt "Welcome to Stapledon's Voyage" -tts
voyage ai -prompt "Warning: approaching event horizon" -tts -voice Charon

# Multimodal (image + text)
voyage ai -prompt "Describe this image" -image screenshot.png
```

### Environment Variables

| Variable | Description |
|----------|-------------|
| `ANTHROPIC_API_KEY` | Claude API key |
| `GOOGLE_CLOUD_PROJECT` | GCP project for Vertex AI (preferred) |
| `GOOGLE_API_KEY` | Gemini API key (fallback) |
| `AI_PROVIDER` | Default provider (claude, gemini) |

---

## voyage world

Inspect simulation world state.

### Usage

```bash
voyage world [options]

Options:
  -seed int      World seed for initialization (default: 42)
  -steps int     Run N steps before inspection
  -json          Output as JSON
  -summary       Show summary only
```

### Examples

```bash
# Inspect default world
voyage world

# Run 100 steps then inspect
voyage world -steps 100

# Different seed, JSON output
voyage world -seed 123 -json > world_state.json
```

---

## voyage bench

Human-readable performance benchmarks.

### Usage

```bash
voyage bench [options]

Options:
  -n int            Number of iterations (default: 1000)
  -warmup int       Warmup iterations (default: 100)
  -profile          Enable CPU profiling
  -profile-path     CPU profile output path (default: cpu.prof)
```

### Benchmarks

- **InitWorld** - Time to create a new world
- **Step** - Time per simulation step
- **Step100** - Time for 100 consecutive steps

---

## voyage perf

Performance benchmarks with threshold checks (for CI).

### Usage

```bash
voyage perf [options]

Options:
  -n int             Number of iterations (default: 1000)
  -warmup int        Warmup iterations (default: 100)
  -o string          Output JSON file path
  -fail              Exit code 1 if thresholds exceeded (default: true)
  -step-max          Max time for Step() (default: 5ms)
  -init-max          Max time for InitWorld() (default: 100ms)
  -step100-max       Max time for 100 steps (default: 500ms)
  -q                 Quiet mode (only output JSON)
```

### Thresholds (for 60 FPS)

| Benchmark | Threshold | Rationale |
|-----------|-----------|-----------|
| Step | 5ms | Leaves 11ms for rendering |
| InitWorld | 100ms | One-time cost |
| Step100 | 500ms | 5ms average per step |

---

## voyage assets

Validate game assets directory structure.

### Usage

```bash
voyage assets [options]

Options:
  -dir string    Assets directory to validate (default: assets)
  -v             Verbose output
  -fix           Create missing directories
```

### Expected Structure

```
assets/
├── sprites/     (.png files)
├── fonts/       (.ttf, .otf files)
├── sounds/      (.wav, .ogg, .mp3 files)
├── generated/   (AI-generated content)
├── starmap/     (star catalog data)
└── manifest.json (optional)
```

---

## voyage sim

Simulation stress tests.

### Usage

```bash
voyage sim [options]

Options:
  -steps int        Number of steps to simulate (default: 10000)
  -seed int         World seed (default: 42)
  -check int        Interval for progress output (default: 1000)
  -validate         Validate state after each step
```

---

## Asset Generation Tools

### gensprites

Generates placeholder PNG sprites for testing.

```bash
go run cmd/gensprites/main.go [output-dir]
# Default: assets/sprites/
```

**Generated sprites:**
- `iso_tiles/` - 64x32 isometric diamond tiles (water, forest, desert, mountain)
- `iso_entities/` - 128x48 animated sprite sheets (4 frames x 32x48)
- `stars/` - 16x16 star sprites by spectral type (O/B, A/F, G, K, M)

### gensounds

Generates placeholder WAV sound files.

```bash
go run cmd/gensounds/main.go [output-dir]
# Default: assets/sounds/
```

**Generated sounds:**
- `click.wav` - 800Hz, 50ms
- `build.wav` - 440Hz, 200ms
- `error.wav` - 220Hz, 300ms
- `select.wav` - 660Hz, 100ms

---

## Visual Effect Tools

### grscreenshot

Captures screenshots with various GR effect combinations.

```bash
go run cmd/grscreenshot/main.go
# Output: out/gr-screenshots/
```

**Generated screenshots:**
- `gr-none.png` - No effects
- `gr-subtle.png` - Subtle gravitational lensing
- `gr-strong.png` - Strong lensing
- `gr-extreme.png` - Extreme lensing
- `gr-with-bloom.png` - GR + bloom
- `gr-with-sr.png` - GR + SR aberration
- `gr-all-relativity.png` - All relativistic effects

### granimation

Generates frames for a black hole journey animation.

```bash
go run cmd/granimation/main.go
# or
./bin/granimation
# Output: out/gr-animation/frame_XXX.png
```

**Output:**
- 60 frames (2 seconds at 30fps)
- Simulates approach and retreat from black hole
- Phi varies 0 → 0.08 → 0

**Creating video:**
```bash
# GIF
ffmpeg -framerate 30 -i out/gr-animation/frame_%03d.png \
  -vf 'scale=640:-1:flags=lanczos,split[s0][s1];[s0]palettegen[p];[s1][p]paletteuse' \
  out/gr-journey.gif

# MP4
ffmpeg -framerate 30 -i out/gr-animation/frame_%03d.png \
  -c:v libx264 -pix_fmt yuv420p out/gr-journey.mp4
```

---

## Makefile Integration

```makefile
# Build CLI tools
make cli           # Build voyage CLI
make game          # Build main game
make sprites       # Generate test sprites
make sounds        # Generate test sounds

# Run with mock sim_gen (no AILANG compiler needed)
make run-mock      # Run game
make cli-mock      # Build CLI with mock

# Testing via CLI
make eval          # Run AILANG evaluation
make bench         # Run benchmarks
```

---

## Testing Features via CLI

The CLI should be able to test all major game features:

| Feature | CLI Command | Status |
|---------|-------------|--------|
| AI text generation | `voyage ai -prompt "..."` | Implemented |
| AI image generation | `voyage ai -generate-image -prompt "..."` | Implemented |
| AI image editing | `voyage ai -edit-image -prompt "..."` | Implemented |
| AI TTS | `voyage ai -tts -prompt "..."` | Implemented |
| World state | `voyage world` | Implemented |
| Performance | `voyage bench`, `voyage perf` | Implemented |
| Asset validation | `voyage assets` | Implemented |
| Stress testing | `voyage sim` | Implemented |
| GR screenshots | `grscreenshot` | Implemented |
| GR animation | `granimation` | Implemented |

### Planned CLI Features

- [ ] `voyage screenshot` - Capture game screenshots
- [ ] `voyage scenario` - Run test scenarios
- [ ] `voyage save` - Test save/load system
- [ ] `voyage starmap` - Inspect starmap data

---

## Success Criteria

### Implemented
- [x] Main voyage CLI with subcommands
- [x] AI testing (text, image gen, image edit, TTS)
- [x] World state inspection
- [x] Performance benchmarks with thresholds
- [x] Asset validation
- [x] Simulation stress tests
- [x] Sprite generation
- [x] Sound generation
- [x] GR effect screenshots
- [x] GR animation generation

### Planned
- [ ] Screenshot command
- [ ] Scenario runner command
- [ ] Save/load testing command
- [ ] Starmap inspection command

---

**Document created**: 2025-12-06
**Last updated**: 2025-12-06
