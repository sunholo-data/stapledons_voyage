# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**This project is primarily an integration test for AILANG** - the game itself is a rewarding side effect. We use the AILANG binary and messaging system to iteratively build the game while stress-testing AILANG's capabilities.

Stapledons Voyage is a game engine host that integrates with AILANG (a separate language/compiler repo). This repo is the "host" while AILANG is the "brain". The game compiles AILANG code to Go, links it with an Ebiten-based 2D engine, and runs benchmarks/scenarios.

## Build Commands

```bash
make sim      # Compile AILANG → Go (prerequisite for all others)
make game     # Build executable to bin/game
make run      # Run game directly with go run
make eval     # Run benchmarks + scenarios, output to out/report.json
make clean    # Remove generated artifacts (sim_gen/, bin/, out/*)
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
- **Never edit:** `sim_gen/*.go` (auto-generated from AILANG)

The generated Go package `sim_gen` exports `InitWorld`, `Step`, and all ADT types.

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

- **Module imports not working** - Use local type definitions as workaround
- **Recursion depth limits** - Avoid deep recursion; use literal lists for small data
- **RNG effect coming in v0.5.1** - `rand_float()` (0-1), `rand_int(max)` (0-max), `AILANG_SEED` for determinism
- **Array type coming in v0.5.0** - Currently only lists (O(n) access)
- **Record update may fail** - Construct new records explicitly instead of `{base | field: val}`

### Available in std/prelude (auto-imported)

- `intToFloat(n)` - Convert int to float (e.g., `intToFloat(42)` → `42.0`)
- `floatToInt(f)` - Convert float to int

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
| v0.4.9 | Bug fixes (in progress) |
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
   Types: `bug`, `feature`, `docs`, `compatibility`, `performance`

### Checking for Responses

```bash
ailang agent inbox stapledons_voyage    # Check for messages
ailang agent ack <msg-id>               # Acknowledge after reading
```

This feedback loop helps improve AILANG based on real-world usage in this integration test project.