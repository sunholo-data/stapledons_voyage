# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

**Related docs:**
- [README.md](README.md) - User-facing overview, quick start
- [docs/game-vision.md](docs/game-vision.md) - Full game design document
- [DEVELOPMENT.md](DEVELOPMENT.md) - Technical reference, data flow, types
- [design_docs/](design_docs/) - Feature design documentation
  - `planned/next/` - Features to implement next (need sprints)
  - `implemented/vX_Y_Z/` - Completed features by version
- [sprints/](sprints/) - Sprint plans tracking implementation
- [design_docs/reference/engine-capabilities.md](design_docs/reference/engine-capabilities.md) - Complete engine reference (DrawCmd, effects, shaders, physics)

## Project Overview

**Stapledon's Voyage** is a hard sci-fi philosophy simulator where you pilot a near-light-speed ship with 100 subjective years to explore the galaxy. Every journey triggers relativistic time dilation: civilizations rise and fall in the centuries that pass between your visits. At the end, the game fast-forwards to Year 1,000,000 and shows you what your choices did to the galaxy.

**This project is also the primary integration test for AILANG** - the game simulation logic will be written in AILANG, a new programming language. This repo is the "host" (Go/Ebiten engine) while AILANG is the "brain" (simulation logic).

## ⚠️ CRITICAL: AILANG-Only Game Logic

**STOP before writing ANY Go code for game features!**

This project tests AILANG as a game scripting language. The Go engine is a "dumb renderer" - it ONLY:
- Captures input → passes to AILANG
- Renders DrawCmd output from AILANG
- Loads assets and applies shaders

**ALL game logic MUST be in AILANG** (`sim/*.ail`):
- Tile layouts and positioning
- NPC/crew positions, movement, AI
- Player movement and interactions
- State management and game rules
- Draw command generation (what to render where)

**The Go engine should NEVER contain:**
- Game state definitions (use AILANG types)
- Movement or positioning logic
- AI or behavior logic
- Level layout or tile data
- Any "game design" decisions

**The Go engine CAN contain (purely visual, no gameplay impact):**
- Decorative particles (dust, sparks, debris) - visual only
- Screen transition animations (fade, wipe) - visual polish
- Shader effects (SR warp, bloom) - pure rendering
- UI layout math (positioning) - where things draw

**The key question: Does this affect gameplay outcomes?**
- YES → Must be AILANG (e.g., velocity affects time dilation)
- NO → Engine is OK (e.g., particle animation is decorative)

```
AILANG owns WHAT is happening (state, logic, decisions)
Engine owns HOW it looks (rendering, animation, polish)
```

**If you're about to write game logic in Go, STOP. Write it in AILANG instead.**

## ⚠️ CRITICAL: Handling AILANG Deficiencies

**When AILANG is missing a feature or has a bug:**

1. **DO NOT implement workarounds** - No Go workarounds, no "creative solutions"
2. **DO NOT try alternatives** - If std/math fails, don't try literal values or inline math
3. **DO report via messaging with GitHub issue:**
   ```bash
   ailang messages send user "Description of what's missing/broken..." \
     --type bug \   # or --type feature
     --github
   ```
4. **DO wait for AILANG team to fix it** - This is an integration test; waiting IS the correct action
5. **DO mark features as BLOCKED** - Add clear comments in code noting the blocker
6. **DO check messages for updates** - `ailang messages list --unread`

**Why wait instead of workaround?**
- This project tests AILANG - workarounds hide bugs from the AILANG team
- Go workarounds violate the architecture (all game logic in AILANG)
- The AILANG team responds quickly via the messaging system

**Current blockers:** None! All recent issues fixed.

**Recently fixed (2025-12-10):**
- #28: math codegen type assertions (FIXED)
- #27: math import missing (FIXED)
- #26: math codegen builtins (FIXED)
- #25: record list type inference (FIXED)
- #23/#24: record update nested record (FIXED)
- #22: std/option map() (FIXED)
- #21: std/math PI() (FIXED)
- #19: std/math module (FIXED)
- #18: Field access nested closures (non-blocking)

**This project exists to find these gaps.** Every deficiency reported improves AILANG.

The Go layer only needs to:
1. Call `InitX()` to get initial state from AILANG
2. Call `StepX(state, input)` each frame to get updated state
3. Call `RenderX(state)` to get DrawCmds
4. Render those DrawCmds to screen

## Build Commands

### With AILANG Compiler (when `ailc` is available)

```bash
make sim      # Compile AILANG → Go (prerequisite for all others)
make game     # Build executable to bin/game
make run      # Run game directly with go run
make eval     # Run benchmarks + scenarios, output to out/report.json
```

### Legacy Mock Mode (deprecated)

Mock mode was used during early prototyping. **We are now ALL-IN on AILANG.**

```bash
# These still work but should rarely be needed:
make game-mock   # Build using mock sim_gen
make run-mock    # Run game using mock sim_gen
```

**For all new game features, write AILANG code in `sim/*.ail`.**

### Output Directories

**All binaries MUST go in `bin/`** - never write executables to the project root.

```bash
bin/game         # Main game executable
bin/demo-*       # Demo/test executables (e.g., bin/demo-bridge)
out/             # Reports, screenshots, test output
```

When creating new `cmd/` entrypoints, always build to `bin/`:
```bash
go build -o bin/demo-foo ./cmd/demo-foo
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

Messages (for coordination between Claude and AILANG):
```bash
ailang messages list                   # List all messages
ailang messages list --unread          # List unread messages only
ailang messages send <inbox> <msg>     # Send message to inbox
ailang messages ack <msg-id>           # Mark message as read
ailang messages read <msg-id>          # Show message content
ailang messages watch                  # Watch for new messages
```

## Architecture

### AILANG-First Development (CRITICAL)

**ALL game logic MUST be written in AILANG.** The Go engine is "dumb" - it only:
- Captures input → passes to AILANG
- Renders DrawCmd output from AILANG
- Loads assets and applies shaders

**If you're about to write game logic in Go, STOP. Write it in AILANG instead.**

| What | Where | Language |
|------|-------|----------|
| Game state, logic, AI | `sim/*.ail` | AILANG ✅ |
| Rendering, shaders, assets | `engine/*.go` | Go (rare edits) |
| Generated bridge code | `sim_gen/*.go` | Never edit ❌ |

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

## sim_gen Development

AILANG v0.5.0 Go codegen is now available. The workflow is:

```bash
# Generate Go code from AILANG
ailang compile --emit-go --package-name sim_gen --out sim_gen sim/*.ail

# Or use Makefile
make sim    # Compile AILANG → Go
make game   # Build with generated code
```

**Fallback:** If codegen fails, mock sim_gen (hand-written Go) is still available via `make game-mock`.

### Layer Boundaries

| Layer | Location | Contains | Edit? |
|-------|----------|----------|-------|
| **AILANG Source** | `sim/*.ail` | Game logic | ✅ Edit here |
| **Generated Code** | `sim_gen/*.go` | Auto-generated from AILANG | ❌ Never edit |
| **Engine** | `engine/*.go` | IO bridge only | ✅ Edit here |
| **Entry** | `cmd/*.go` | Wiring | ✅ Edit here |

### What Goes Where

**sim/*.ail (AILANG source):**
- World state types and step function
- Game logic (actions, building, NPC behavior)
- DrawCmd generation
- All simulation logic

**sim_gen/*.go (generated):**
- Auto-generated Go code from `ailang compile --emit-go`
- Never edit manually - changes will be overwritten

**engine/ (permanent):**
- Input capture (keyboard, mouse → FrameInput)
- Rendering (FrameOutput → pixels)
- Asset management (sprites, sounds)
- Display config (resolution, fullscreen)
- **NO GAME LOGIC** - just IO bridging

### Key Principle

The engine layer should be "dumb" - it captures input, passes it to sim_gen, and renders whatever sim_gen outputs. All game logic lives in `sim/*.ail`.

## Engine Feature Status (Updated 2025-12-04)

The Go/Ebiten engine layer is largely complete. Reference this when planning features.

### Working Features

| Component | Location | Capabilities |
|-----------|----------|--------------|
| Game loop | `cmd/game/main.go` | Ebiten Update/Draw cycle |
| Render bridge | `engine/render/draw.go` | All DrawCmd types: Rect, Sprite, Text, IsoTile, IsoEntity, Ui, Line, Circle, TextWrapped, GalaxyBg, Star |
| Camera | `engine/camera/` | WorldToScreen/ScreenToWorld transforms, viewport culling |
| Input | `engine/input/` | Mouse position, clicks, keyboard events |
| Sprites | `engine/assets/sprites.go` | Atlas loading, animation frames, manifest support |
| Audio | `engine/assets/audio.go` | OGG/WAV loading, PlaySound, volume control, manifest |
| Fonts | `engine/assets/fonts.go` | TTF loading, size scaling |
| UI | `engine/render/draw.go` | Panel, Button, Label, Portrait, Slider, ProgressBar |
| Display | `engine/display/` | Resolution config, F11 fullscreen toggle |
| Screenshot | `engine/screenshot/` | Headless capture for testing |
| Scenarios | `engine/scenario/` | Automated visual test runner |

### Effect Handlers (MUST BE INITIALIZED BY HOST)

**CRITICAL:** Effect handlers are NOT auto-initialized. The host (Go engine) MUST call `sim_gen.Init(handlers)` before any AILANG code that uses effects runs. Calling methods on uninitialized handlers will panic.

```go
// In cmd/game/main.go - BEFORE calling InitWorld or Step:
sim_gen.Init(sim_gen.Handlers{
    Debug: sim_gen.NewDebugContext(),
    Rand:  &DefaultRandHandler{},
    Clock: &EbitenClockHandler{},
    AI:    handlers.NewStubAIHandler(),
    // FS, Net, Env: only if needed
})
```

| Effect | Interface | Implementation | Status |
|--------|-----------|----------------|--------|
| Debug | `DebugHandler` | `sim_gen.NewDebugContext()` | ✅ Built-in |
| Rand | `RandHandler` | `engine/handlers/rand.go` | ✅ Working |
| Clock | `ClockHandler` | `engine/handlers/clock.go` | ✅ Working |
| AI | `AIHandler` | `engine/handlers/ai.go` | ✅ Stub exists |
| FS | `FSHandler` | Not needed (see design note) | - |
| Net | `NetHandler` | Not needed yet | - |
| Env | `EnvHandler` | Not needed yet | - |

**Runtime Fix (2025-12-04):** Fixed `ListLen`, `ListHead`, `ListTail`, `Length`, `Get`, `GetOpt` in `sim_gen/runtime.go` to handle typed slices (like `[]*NPC`, `[]*Direction`) using reflection, not just `[]interface{}`. Also fixed `makeOptionSome`/`makeOptionNone` to use typed `*Option` constructors. This was needed because AILANG generates typed slices but the original runtime only handled `[]interface{}`.

### Design Note: Single Save File (No Save Slots)

Per **Pillar 1 (Choices Are Final)**, players cannot maintain multiple save files:
- **Single save file** - overwrites on each save
- Player can save and load normally
- **No save slots** - can't keep backups to try different paths
- This prevents branching timelines while allowing normal session management

### Remaining Engine Work

| Gap | Priority | Effort | Blocks |
|-----|----------|--------|--------|
| ~~Clock handler~~ | ✅ Done | - | - |
| ~~AI handler stub~~ | ✅ Done | - | - |
| Save system | P2 | 2 days | Single-file save/load (no slots) |
| AI handler real impl | P3 | 2 days | NPC dialogue with LLM |

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
6. **Check inbox for responses** - `ailang messages list --unread`
7. **Iterate** - Incorporate feedback, fix issues, continue development

### Available in v0.5.0 (Current)

- **Go codegen** - `ailang compile --emit-go --package-name sim_gen --out sim_gen sim/*.ail`
- **Arrays** - `#[1, 2, 3]` with O(1) access via `std/array`
- **Rand effect** - `std/rand` with `rand_int(min, max)`, `rand_float(min, max)`, `rand_bool()`, `rand_seed(n)`
- **Debug effect** - `std/debug` with `Debug.log(msg)`, `Debug.check(cond, msg)`
- **AI effect** - `std/ai` with `ai_call(input)` for LLM integration
- **Game clock** - `std/game` with `delta_time()`, `frame_count()`, `total_time()`
- **Extern functions** - Call Go from AILANG for performance-critical code:
  ```ailang
  -- Declare in AILANG
  extern func octreeQuery(center: Vec3, radius: float) -> [Star]
  ```
  Compiler generates `extern_stubs.go` with panic stubs; implement in your own Go file.

### Syntax Notes

- **No tuple destructuring in let** - Use `match pair { (x, y) => ... }` instead of `let (x, y) = pair`

### Fixed in v0.5.0

- **Nested field access** - `npc.pos.x` now works through function params, match bindings, list patterns
- **Record update with nested values** - `{ npc | pos: { x: newX, y: newY } }` works correctly
- **ADT inline tests** - Tests like `tests [(North, 0), (South, 0)]` now resolve local constructors
- **Imported ADT type pollution** - Mixed ADT constructors (e.g., `PatternPatrol([Direction])` + `PatternRandomWalk(int)`) in same scope no longer cause type inference errors

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
- **Effects** - Rand, Debug, AI all available in v0.5.0
- **Extern functions** - Performance kernels implemented in Go

### Version Status

| Version | Status | Features |
|---------|--------|----------|
| v0.5.0 | **Current** | Go codegen, Arrays, Rand/Debug/AI effects, Extern functions |
| v0.5.1 | Planned | Additional CLI flags, performance improvements |
| v0.5.2 | Planned | ABI "stable preview" |

## Integration Testing & Feedback Loop

This repo serves as the primary integration test for AILANG. The development process is:

1. Write/modify AILANG code in `sim/*.ail`
2. Use `ailang` CLI to run, check, and evaluate
3. Use `ailang messages` to coordinate with AILANG team
4. Evaluation output (`out/report.json`) feeds back into AILANG improvements

If AILANG breaks, the game breaks - making this an effective stress test for the language.

## AILANG Feedback Workflow

**IMPORTANT:** When working on this project, proactively report any AILANG issues or improvement ideas using the `ailang-feedback` skill.

### Message Types

| Type | Use Case | Method |
|------|----------|--------|
| **Bugs** | Parser errors, codegen issues, crashes | `--type bug --github` |
| **Features** | Language features for game development | `--type feature --github` |
| **Coordination** | Acknowledgments, status updates, questions | Direct message (no `--github`) |
| **DX feedback** | Error messages, CLI friction, workflow issues | `--type bug --github` or direct |

### How to Report (with GitHub Integration)

**For bugs and features (creates GitHub issue):**
```bash
ailang messages send user "Description of the issue" \
  --title "Short title" \
  --from "stapledons_voyage" \
  --type bug \
  --github
```

Types: `bug`, `feature`, `general`

**For coordination/acknowledgments (no GitHub issue):**
```bash
ailang messages send user "Message content" \
  --title "Title" \
  --from "stapledons_voyage"
```

### GitHub Account Configuration

GitHub sync requires matching accounts. If you see account mismatch errors:

1. Check expected user in `~/.ailang/config.yaml`
2. Switch accounts: `gh auth switch --user <expected_user>`
3. Or update `expected_user` in config

To retry failed syncs, resend the message with `--github`:
```bash
ailang messages read <msg-id>  # Copy the content
ailang messages send user "<content>" --title "<title>" --from "stapledons_voyage" --type bug --github
```

### Checking for Responses

```bash
ailang messages list --unread           # Check for unread messages
ailang messages read <msg-id>           # Read message content
ailang messages ack <msg-id>            # Mark as read
```

### Import GitHub Issues

Pull issues from GitHub into messaging:
```bash
ailang messages import-github --repo owner/repo --labels bug
ailang messages import-github --dry-run  # Preview first
```

This feedback loop helps improve AILANG based on real-world usage in this integration test project.

## Feature Development Workflow

**IMPORTANT:** All game features MUST follow this workflow. Do not implement features without a sprint plan.

### Design Docs → Sprints → Implementation

```
design_docs/planned/next/  →  sprints/  →  code  →  design_docs/implemented/vX_Y_Z/
        (what)               (how+when)    (do)           (done)
```

### Step-by-Step Process

1. **Create Design Doc** (`design-doc-creator` skill)
   - Location: `design_docs/planned/next/<feature-name>.md`
   - Describes WHAT the feature does and WHY
   - Contains acceptance criteria

2. **Create Sprint Plan** (`sprint-planner` skill)
   - Location: `sprints/<feature-name>-sprint.md` or `sprints/plans/`
   - Sprint file MUST reference the design doc: `Design Doc: design_docs/planned/next/<name>.md`
   - Contains day-by-day breakdown with checkboxes
   - Lists files to create/modify
   - Estimates effort

3. **Execute Sprint** (`sprint-executor` skill)
   - Work through sprint tasks
   - Mark checkboxes as completed: `[x]`
   - Update sprint file with actual files created

4. **Move to Implemented**
   - When sprint is 100% complete
   - Move design doc: `git mv design_docs/planned/next/<name>.md design_docs/implemented/vX_Y_Z/`
   - Update design doc with implementation locations

### Audit Command

Check which design docs have sprints and their progress:

```bash
.claude/skills/game-architect/scripts/audit_design_docs.sh
```

This shows:
- Design docs WITH sprints (ready to implement)
- Design docs WITHOUT sprints (need planning first)
- Sprint completion percentages

### Why This Matters

- **No orphan code** - All changes traced to design docs
- **No orphan docs** - All planned features have sprint tracking
- **Progress visibility** - Sprint checkboxes show real progress
- **Easy auditing** - Can verify what's actually implemented vs planned

### Related Skills

| Skill | Purpose |
|-------|---------|
| `design-doc-creator` | Create new design docs in planned/next/ |
| `sprint-planner` | Create sprint plans for design docs |
| `sprint-executor` | Execute approved sprint plans |
| `game-architect` | Audit design docs, validate architecture |