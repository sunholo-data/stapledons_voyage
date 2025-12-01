# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

**Related docs:**
- [README.md](README.md) - User-facing overview, quick start
- [docs/game-vision.md](docs/game-vision.md) - Full game design document
- [DEVELOPMENT.md](DEVELOPMENT.md) - Technical reference, data flow, types
- [design_docs/](design_docs/) - Feature design documentation

## Project Overview

**Stapledon's Voyage** is a hard sci-fi philosophy simulator where you pilot a near-light-speed ship with 100 subjective years to explore the galaxy. Every journey triggers relativistic time dilation: civilizations rise and fall in the centuries that pass between your visits. At the end, the game fast-forwards to Year 1,000,000 and shows you what your choices did to the galaxy.

**This project is also the primary integration test for AILANG** - the game simulation logic will be written in AILANG, a new programming language. This repo is the "host" (Go/Ebiten engine) while AILANG is the "brain" (simulation logic).

## Build Commands

### With AILANG Compiler (when `ailc` is available)

```bash
make sim      # Compile AILANG → Go (prerequisite for all others)
make game     # Build executable to bin/game
make run      # Run game directly with go run
make eval     # Run benchmarks + scenarios, output to out/report.json
```

### Without AILANG Compiler (mock mode)

Use these targets while AILANG compiler is in development:

```bash
make game-mock   # Build using mock sim_gen (no ailc needed)
make run-mock    # Run game using mock sim_gen
make eval-mock   # Run benchmarks + scenarios using mock sim_gen
make sprites     # Generate test sprite PNGs
```

### Cleanup

```bash
make clean       # Remove bin/, out/* (preserves sim_gen for mock mode)
make clean-all   # Remove sim_gen/, bin/, out/* (full clean)
```

## AILANG CLI

The `ailang` command is available with these key commands:
```bash
ailang run <file>           # Run an AILANG program
ailang repl                 # Start interactive REPL
ailang check <file>         # Type-check without running
ailang test [path]          # Run tests
ailang prompt               # Display AILANG teaching prompt (for AI code generation)
ailang eval [flags]         # Run AI benchmarks
```

Run command flags (must come BEFORE filename):
```bash
ailang run --caps IO,FS,Net file.ail   # Enable capabilities
ailang run --entry func file.ail       # Custom entrypoint
ailang run --trace file.ail            # Enable execution tracing
```

Agent messaging (for coordination between Claude and AILANG):
```bash
ailang agent top                       # Show agent queue status
ailang agent inbox <agent-id>          # Check messages for an agent
ailang agent send <agent> <json>       # Send message to an agent
ailang agent ack <msg-id>              # Acknowledge message
```

## Architecture

**Three-layer design:**

1. **Simulation Layer (AILANG)** - `sim/*.ail` files define world state and step logic
   - `protocol.ail` - Stable API types (FrameInput, FrameOutput, DrawCmd)
   - `world.ail` - World state definitions (Planet, NPC, Tile)
   - `step.ail` - `init_world(seed)` and `step(world, input)` functions

2. **Engine Layer (Go/Ebiten)** - Pure Go in `engine/` and `cmd/`
   - `cmd/game/main.go` - Ebiten game loop
   - `engine/render/` - CaptureInput() and RenderFrame() bridge functions
   - `engine/assets/` - Sprite, font, and sound loading with manifest support
   - `engine/display/` - Window configuration, fullscreen (F11), resolution settings

3. **Evaluation Layer** - `cmd/eval/` and `engine/bench/`
   - Produces `out/report.json` for AI-driven AILANG improvements

**Data flow:**
```
User Input → CaptureInput() → FrameInput → AILANG step() → World, FrameOutput → RenderFrame() → Screen
```

## Writing AILANG Code

Before writing or modifying `.ail` files, run `ailang prompt` to get the comprehensive AILANG teaching prompt. This provides:

- **Mandatory structure** - Module declaration must be first line (`module benchmark/solution`)
- **Syntax rules** - Use `func` (not `fn`/`def`), `::` for list cons, `match` with `=>`, semicolons only inside `{ }` blocks
- **Effects system** - Annotate side effects with `! {IO, FS, Net, Env}` after return type
- **Standard library** - `std/io`, `std/fs`, `std/net`, `std/json`, `std/list`, `std/string`
- **Critical limitations** - No loops (use recursion), no mutable variables, no list comprehensions
- **Common mistakes** - What NOT to do with clear examples

Always reference `ailang prompt` output when writing AILANG code to ensure correct syntax and patterns.

## Code Generation Boundary

- **Manually edit:** `sim/*.ail` files
- **Never edit:** `sim_gen/*.go` (auto-generated from AILANG or mock)

The `sim_gen` package exports `InitWorld`, `Step`, and all ADT types. It can be:
- **Generated** - From AILANG via `make sim` (when `ailc` is available)
- **Mock** - Hand-written Go types matching AILANG protocol (for development without `ailc`)

## Mock sim_gen Development (IMPORTANT)

Until AILANG's Go codegen ships (`ailang compile --emit-go` in v0.5.0), we use **mock sim_gen** - hand-written Go that mimics what AILANG would generate.

### Layer Boundaries

| Layer | Location | Contains | Permanent? |
|-------|----------|----------|------------|
| **AILANG Source** | `sim/*.ail` | Game logic (future) | Yes |
| **Simulation Mock** | `sim_gen/*.go` | Game logic (temporary) | **No - replaced when AILANG ships** |
| **Engine** | `engine/*.go` | IO bridge only | Yes |
| **Entry** | `cmd/*.go` | Wiring | Yes |

### What Goes Where

**sim_gen/ (MOCK - temporary):**
- World state types and Step function
- Game logic (actions, building, NPC behavior)
- DrawCmd generation
- This is SIMULATION LAYER code, just written in Go until AILANG works

**engine/ (PERMANENT):**
- Input capture (keyboard, mouse → FrameInput)
- Rendering (FrameOutput → pixels)
- Asset management (sprites, sounds)
- Display config (resolution, fullscreen)
- **NO GAME LOGIC** - just IO bridging

### Key Principle

The engine layer should be "dumb" - it captures input, passes it to sim_gen, and renders whatever sim_gen outputs. All decisions about game behavior belong in sim_gen (mock) or sim/*.ail (AILANG).

When AILANG compiler ships:
1. Write game logic in `sim/*.ail`
2. Run `ailang compile --emit-go`
3. Generated code replaces mock `sim_gen/*.go`
4. Engine layer stays unchanged

## Key Types (defined in protocol.ail)

DrawCmd is a tagged union:
```ailang
type DrawCmd =
    | Sprite(int, float, float, int)
    | Rect(float, float, float, float, int, int)
    | Text(string, float, float, int)
```

## General Development Workflow

When working on this project, follow this iterative process:

1. **Read `ailang prompt`** - Get current AILANG syntax reference
2. **Write/modify `.ail` files** - Implement game logic in `sim/`
3. **Type-check with `ailang check`** - Fix any syntax/type errors
4. **Test with `ailang run`** - Verify runtime behavior
5. **Report issues via `ailang-feedback`** - Send bugs, unclear docs, feature requests
6. **Check inbox for responses** - `ailang agent inbox stapledons_voyage`
7. **Iterate** - Incorporate feedback, fix issues, continue development

### Known Limitations (as of current testing)

- **RNG effect coming in v0.5.1** - `rand_float()` (0-1), `rand_int(max)` (0-max), `AILANG_SEED` for determinism
- **Array type coming in v0.5.0** - Currently only lists (O(n) access)
- **No tuple destructuring in let** - Use `match pair { (x, y) => ... }` instead of `let (x, y) = pair`

### Fixed in v0.5.0 (Dec 2025)

- **Nested field access** - `npc.pos.x` now works through function params, match bindings, list patterns
- **Record update with nested values** - `{ npc | pos: { x: newX, y: newY } }` works correctly
- **ADT inline tests** - Tests like `tests [(North, 0), (South, 0)]` now resolve local constructors

### Design Choices (Intentional)

- **Module-level `let` not accessible in functions** - This is intentional for determinism. Module-level bindings are evaluated once at load time; if functions could access them, order of evaluation could affect results. Workaround: inline constants or pass as parameters.

### Fixed in v0.4.9

- **Module imports** - Both std library (`import std/list (length)`) and local (`import sim/world (World)`) imports now work
- **Record update with type inference** - `{base | field: val}` now works in lambdas (e.g., `\world. {world | tick: world.tick + 1}`)
- **Match in recursive functions** - Match expressions with int literals now work correctly (e.g., `match n { 0 => 0, _ => 1 + count(n-1) }`)

### Available in std/prelude (auto-imported)

- `intToFloat(n)` - Convert int to float (e.g., `intToFloat(42)` → `42.0`)
- `floatToInt(f)` - Convert float to int

### Inline Tests (v0.4.7+)

Use inline tests for executable documentation:

```ailang
-- Syntax: tests [(input, expected), ...]
pure func square(x: int) -> int tests [(0, 0), (5, 25), (-3, 9)] {
    x * x
}

-- Run tests with: ailang test sim/
```

Add inline tests to pure functions where practical. This documents expected behavior and catches regressions.

Report any additional issues encountered to AILANG core.

## AILANG Consumer Contract

This project targets **AILANG v0.5.x**. See [ailang_resources/consumer-contract-v0.5.md](ailang_resources/consumer-contract-v0.5.md) for the full contract defining:

- **Go codegen** - `ailang compile --emit-go` produces valid, deterministic Go code
- **ADT → discriminator structs** - Sum types become struct-based (not interface-based)
- **Effects** - RNG (v0.5.1+), Debug (v0.5.1+), AI (v0.5.1+)
- **Extern functions** - Performance kernels implemented in Go (v0.5.2+)

### Version Roadmap

| Version | Features |
|---------|----------|
| v0.4.9 | Bug fixes: record update inference, match recursion depth |
| v0.5.0 | Go codegen, ADT structs, Array type |
| v0.5.1 | RNG effect, Debug effect, AI effect |
| v0.5.2 | Extern functions, CLI flags |
| v0.5.3 | ABI "stable preview" |

## Integration Testing & Feedback Loop

This repo serves as the primary integration test for AILANG. The development process is:

1. Write/modify AILANG code in `sim/*.ail`
2. Use `ailang` CLI to run, check, and evaluate
3. Use `ailang agent` messaging to coordinate with Claude
4. Evaluation output (`out/report.json`) feeds back into AILANG improvements

If AILANG breaks, the game breaks - making this an effective stress test for the language.

## AILANG Feedback Workflow

**IMPORTANT:** When working on this project, proactively report any AILANG issues or improvement ideas using the `ailang-feedback` skill.

### What to Report

- **Bugs** - Parser errors, type-checker issues, runtime crashes, unexpected behavior
- **DX improvements** - Confusing error messages, missing CLI features, workflow friction
- **Feature requests** - Language features that would help build the game
- **Documentation gaps** - Unclear or missing AILANG docs

### How to Report

1. Invoke the `ailang-feedback` skill
2. Use the send script:
   ```bash
   ~/.claude/skills/ailang-feedback/scripts/send_feedback.sh <type> "<title>" "<description>" "stapledons_voyage"
   ```
   Types: `bug`, `feature`, `docs`, `compatibility`, `performance`, `dx` (developer experience)

### Checking for Responses

```bash
ailang agent inbox stapledons_voyage    # Check for messages
ailang agent ack <msg-id>               # Acknowledge after reading
```

This feedback loop helps improve AILANG based on real-world usage in this integration test project.