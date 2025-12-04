# CLI Reference

Complete documentation for the Voyage CLI development tools.

## Building

```bash
make cli                    # Build to bin/voyage
go build -o bin/voyage ./cmd/cli  # Direct build
```

## Commands

### `voyage help`
Show usage information.

### `voyage world`
Inspect simulation world state.

| Flag | Default | Description |
|------|---------|-------------|
| `-seed` | 42 | World seed for initialization |
| `-steps` | 0 | Run N steps before inspection |
| `-json` | false | Output raw JSON |
| `-summary` | false | Show summary stats only |

**Examples:**
```bash
./bin/voyage world -summary           # Quick stats
./bin/voyage world -seed 123 -steps 50 -json  # Specific seed, 50 steps, JSON
```

### `voyage bench`
Run performance benchmarks (human-readable output).

| Flag | Default | Description |
|------|---------|-------------|
| `-n` | 1000 | Number of iterations |
| `-warmup` | 100 | Warmup iterations |
| `-profile` | false | Enable CPU profiling |
| `-profile-path` | cpu.prof | Profile output path |

**Benchmarks run:**
- `InitWorld`: Time to create new world
- `Step`: Time per simulation step
- `Step100`: Time for 100 consecutive steps

**Output includes:**
- Average time per operation
- Total time
- Operations count
- Memory stats (Alloc, TotalAlloc, NumGC)

### `voyage perf`
Run performance benchmarks with threshold checking (CI/JSON output).

| Flag | Default | Description |
|------|---------|-------------|
| `-n` | 1000 | Number of iterations |
| `-warmup` | 100 | Warmup iterations |
| `-o` | "" | Output JSON file path (default: stdout) |
| `-fail` | true | Exit with code 1 if thresholds exceeded |
| `-step-max` | 5ms | Max time for Step() |
| `-init-max` | 100ms | Max time for InitWorld() |
| `-step100-max` | 500ms | Max time for 100 steps |
| `-q` | false | Quiet mode (only output JSON) |

**Default Thresholds (for 60 FPS):**
- `Step`: 5ms (leaves 11ms for rendering)
- `InitWorld`: 100ms (one-time cost)
- `Step100`: 500ms (5ms average per step)

**Output includes:**
- PASS/FAIL status for each benchmark
- P95 latency vs threshold
- Full percentile distribution (min, avg, p50, p95, p99, max)
- JSON report with all metrics

**Examples:**
```bash
./bin/voyage perf                    # Run with defaults, check thresholds
./bin/voyage perf -o out/perf.json   # Output to file
./bin/voyage perf -fail=false        # Don't fail on threshold violations
./bin/voyage perf -step-max 10ms     # Custom Step threshold
./bin/voyage perf -n 5000 -q         # More iterations, quiet mode
```

**Exit Codes:**
- 0: All benchmarks passed thresholds
- 1: One or more benchmarks exceeded thresholds (if `-fail=true`)

### `voyage assets`
Validate game assets directory structure.

| Flag | Default | Description |
|------|---------|-------------|
| `-dir` | assets | Assets directory to validate |
| `-v` | false | Verbose (list files) |
| `-fix` | false | Create missing directories |

**Checks:**
- `sprites/` - PNG files
- `fonts/` - TTF/OTF files
- `sounds/` - WAV/OGG/MP3 files
- `generated/` - AI-generated assets
- `starmap/` - Starmap data
- `manifest.json` - Optional manifest

### `voyage sim`
Run simulation stress tests.

| Flag | Default | Description |
|------|---------|-------------|
| `-steps` | 10000 | Number of steps |
| `-seed` | 42 | World seed |
| `-check` | 1000 | Progress interval |
| `-validate` | false | Validate state each step |

**Output:**
- Progress updates with steps/sec
- Total time
- Steps per second (throughput)
- Error count
- Final world state summary

### `voyage ai`
Test AI handlers (Claude, Gemini).

| Flag | Default | Description |
|------|---------|-------------|
| `-provider` | auto | Provider: claude, gemini, auto |
| `-prompt` | "" | Text prompt to send |
| `-system` | "" | System prompt |
| `-image` | "" | Image file path (Gemini only) |
| `-generate-image` | false | Generate image (Gemini only) |
| `-tts` | false | Text-to-speech (Gemini only) |
| `-voice` | Kore | TTS voice name |
| `-v` | false | Verbose output |
| `-list` | false | List available providers |
| `-list-voices` | false | List TTS voices |

**Environment Variables:**
- `ANTHROPIC_API_KEY` - Claude API key
- `GOOGLE_CLOUD_PROJECT` - GCP project for Vertex AI
- `GOOGLE_API_KEY` - Gemini API key (fallback)
- `AI_PROVIDER` - Default provider

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Error (missing args, validation failed, etc.) |
| 2 | Panic (usually AILANG codegen issue) |

## Known Issues

### Debug Output Noise
AILANG-generated code prints "tick N" messages. Filter with:
```bash
./bin/voyage bench 2>&1 | grep -v "^tick"
```
Reported as AILANG bug msg_20251204_203751.

### Handler Initialization
The CLI initializes sim_gen handlers automatically. If you see nil pointer panics,
ensure `initSimGenHandlers()` is called in main().

## Adding New Commands

1. Edit `cmd/cli/main.go`
2. Add case to switch statement in `main()`
3. Implement `run<Command>Command(args []string)` function
4. Use `flag.NewFlagSet()` for command-specific flags
5. Update usage in `printUsage()`
6. Rebuild with `make cli`
7. Update this documentation

## Code Location

- Main CLI: `cmd/cli/main.go`
- Handlers: `engine/handlers/`
- Sim Gen: `sim_gen/`
- Makefile target: `cli`
