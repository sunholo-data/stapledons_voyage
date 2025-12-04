# AILANG v0.5.x Testing Matrix

**Status**: Active
**Priority**: P0 - Critical for game development
**Target**: Ongoing (tracks AILANG versions)
**Purpose**: Define test coverage for AILANG features before using them in production game code

## Overview

This document defines what AILANG features need testing before relying on them for game development. Each feature has a test plan with success criteria.

## Current Status (v0.5.5)

| Feature | Status | Test Coverage | Notes |
|---------|--------|---------------|-------|
| Go codegen | Working | Full | Compiles and runs |
| ADT discriminator structs | Working | Full | All variants work |
| Records | Working | Full | Fields, nested, updates |
| Effect handlers (Rand) | Working | Full | Random NPC movement |
| Effect handlers (Debug) | Working | Full | Logging works |
| Arrays O(1) | Working | Full | PatternPatrol uses arrays |
| Lists | Working | Full | NPC list processing |
| Pattern matching | Working | Full | ADTs, literals, lists |
| Record update | Working | Full | `{ npc | field: val }` |
| Inline tests | Working | Full | `tests [(input, expected)]` |
| Multi-file compile | Working | Full | `sim/*.ail` |
| **AI effect** | **Untested** | **None** | Handler exists but not called |
| **Clock effect** | **Untested** | **None** | delta_time, frame_count |
| **Extern functions** | **Not available** | None | v0.5.2+ |
| **Release mode** | **Untested** | None | `-tags release` |
| **FS/Net effects** | **Untested** | None | Save/load, leaderboards |

---

## Test Plans

### 1. AI Effect Testing

**Priority:** P1 - Required for NPC dialogue and civ decisions

**Goal:** Verify AILANG can call AI handlers and receive responses.

#### Test: Basic AI Call

**AILANG code** (`sim/test_ai.ail`):
```ailang
module sim/test_ai

import std/ai (ai_call)

-- Simple AI wrapper that echoes input
export func testAiEcho(input: string) -> string ! {AI} {
    ai_call(input)
}

-- Structured AI call with JSON
export func testAiDecision(context: string) -> string ! {AI} {
    let prompt = "Given context: " ++ context ++ ". Reply with JSON: {\"action\": \"...\"}"
    ai_call(prompt)
}
```

**Go test handler** (`engine/ai/test_handler.go`):
```go
package ai

import "stapledons_voyage/sim_gen"

// TestHandler echoes input with "[AI]" prefix for verification
type TestHandler struct{}

func (h TestHandler) Call(input string) (string, error) {
    return "[AI RESPONSE] " + input, nil
}

// Register for tests
func RegisterTestHandler() {
    sim_gen.Init(sim_gen.Handlers{
        AI: TestHandler{},
        // ... other handlers
    })
}
```

**Test cases:**
- [ ] `testAiEcho("hello")` returns `"[AI RESPONSE] hello"`
- [ ] `testAiDecision` returns valid JSON structure
- [ ] Multiple AI calls in sequence work
- [ ] AI effect propagates through function calls
- [ ] Error from handler is handled gracefully

**Integration test** (`engine/ai/ai_test.go`):
```go
func TestAIEffect(t *testing.T) {
    RegisterTestHandler()

    // Call AILANG function that uses AI effect
    result := sim_gen.TestAiEcho("test input")

    if !strings.Contains(result, "[AI RESPONSE]") {
        t.Errorf("AI handler not called: %s", result)
    }
}
```

**Success criteria:**
- [ ] AI effect compiles in AILANG
- [ ] Generated Go code calls handler correctly
- [ ] Handler receives input string exactly
- [ ] Handler response returns to AILANG code
- [ ] Effect works inside step() function

---

### 2. Clock Effect Testing

**Priority:** P2 - Required for frame-rate-independent movement

**Goal:** Verify AILANG can access game clock for delta time and frame count.

#### Test: Clock Access

**AILANG code** (`sim/test_clock.ail`):
```ailang
module sim/test_clock

import std/game (delta_time, frame_count, total_time)

-- Get current delta time
export func testDeltaTime() -> float ! {Clock} {
    delta_time()
}

-- Get frame count
export func testFrameCount() -> int ! {Clock} {
    frame_count()
}

-- Use delta for smooth movement
export func smoothMove(pos: float, velocity: float) -> float ! {Clock} {
    let dt = delta_time()
    pos + velocity * dt
}
```

**Go test handler:**
```go
type TestClockHandler struct {
    dt    float64
    frame int64
    total float64
}

func (h *TestClockHandler) DeltaTime() float64  { return h.dt }
func (h *TestClockHandler) FrameCount() int64   { return h.frame }
func (h *TestClockHandler) TotalTime() float64  { return h.total }

func (h *TestClockHandler) SetFrame(dt float64, frame int64, total float64) {
    h.dt = dt
    h.frame = frame
    h.total = total
}
```

**Test cases:**
- [ ] `delta_time()` returns handler value
- [ ] `frame_count()` returns handler value
- [ ] `total_time()` returns handler value
- [ ] `smoothMove` calculates correctly with various dt values
- [ ] Clock values update between frames

**Success criteria:**
- [ ] Clock effect compiles
- [ ] Handler values flow to AILANG
- [ ] Frame-independent movement works

---

### 3. Extern Function Testing (v0.5.2+)

**Priority:** P1 - Required for galaxy-scale performance

**Goal:** Verify AILANG can declare extern functions and call Go implementations.

**Note:** Extern functions are not yet available in AILANG v0.5.x. This test plan is for when they become available.

#### Test: Simple Extern

**AILANG declaration** (`sim/extern_test.ail`):
```ailang
module sim/extern_test

-- Declared but implemented in Go
extern fastSquare(x: int) -> int

-- Use the extern
export func testExtern(n: int) -> int {
    fastSquare(n)
}
```

**Go implementation** (`engine/extern/test_impl.go`):
```go
package extern

import "stapledons_voyage/sim_gen"

func init() {
    sim_gen.RegisterFastSquare(fastSquareImpl)
}

func fastSquareImpl(x int64) int64 {
    return x * x
}
```

**Test cases:**
- [ ] Extern declaration compiles
- [ ] Go registration works
- [ ] AILANG can call extern
- [ ] Return value correct
- [ ] Complex types work (records, ADTs, lists)

#### Test: Determinism

**AILANG code:**
```ailang
extern deterministicOp(seed: int, data: [int]) -> [int]

-- Call same extern multiple times
export func testDeterminism(seed: int, data: [int]) -> bool {
    let result1 = deterministicOp(seed, data)
    let result2 = deterministicOp(seed, data)
    result1 == result2  -- Should always be true
}
```

**Test cases:**
- [ ] Same input produces same output (100 iterations)
- [ ] Different seeds produce different outputs
- [ ] No map iteration order issues in Go impl

**Success criteria:**
- [ ] Extern syntax accepted by compiler
- [ ] Generated Go calls registered implementation
- [ ] Type conversion works both directions
- [ ] Determinism verified

---

### 4. Release Mode Testing

**Priority:** P3 - Required for production builds

**Goal:** Verify `-tags release` produces optimized, debug-stripped builds.

**Test cases:**
- [ ] `make game-release` produces smaller binary
- [ ] Debug.log calls are no-ops in release
- [ ] Performance measurably better
- [ ] Game still functions correctly

**Test procedure:**
```bash
# Build both modes
make game          # Debug build
make game-release  # Release build

# Compare sizes
ls -la bin/game bin/game-release

# Run both, verify functionality
./bin/game-release
```

---

### 5. FS/Net Effect Testing

**Priority:** P3 - Required for save/load and leaderboards

**Goal:** Verify AILANG can read/write files and make HTTP requests.

#### Test: File System

**AILANG code:**
```ailang
module sim/test_fs

import std/fs (read_file, write_file, exists)

export func testFileRoundtrip(path: string, content: string) -> bool ! {FS} {
    write_file(path, content)
    let read = read_file(path)
    read == content
}
```

**Test cases:**
- [ ] `write_file` creates file
- [ ] `read_file` returns content
- [ ] `exists` returns correct bool
- [ ] Errors handled gracefully

#### Test: Network

**AILANG code:**
```ailang
module sim/test_net

import std/net (http_get, http_post)

export func testHttpGet(url: string) -> string ! {Net} {
    http_get(url)
}
```

**Test cases:**
- [ ] HTTP GET returns response body
- [ ] HTTP POST sends body
- [ ] Network errors don't crash game
- [ ] Timeouts handled

---

## Testing Workflow

### For Each New AILANG Feature

1. **Create test AILANG file** in `sim/test_*.ail`
2. **Implement test Go handler** in `engine/*/test_*.go`
3. **Write Go unit test** that calls sim_gen functions
4. **Run tests**: `go test ./engine/...`
5. **Update this matrix** with results

### Before Using in Game Code

A feature should have:
- [ ] At least 3 test cases passing
- [ ] Edge cases documented
- [ ] Performance acceptable
- [ ] No known bugs

---

## Integration with Game

### Safe to Use Now
- Go codegen
- ADTs, Records, Lists
- Pattern matching
- Rand effect (NPC movement)
- Debug effect (logging)
- Arrays (patrol paths)

### Test Before Using
- AI effect (test handler works, but not with real LLM)
- Clock effect (needs frame-rate testing)

### Not Yet Available
- Extern functions (waiting for v0.5.2)
- Release mode optimization
- FS/Net effects (low priority for game)

---

## Related Documents

- [ai-effect-npcs.md](ai-effect-npcs.md) - Full AI effect design
- [performance-externs.md](performance-externs.md) - Full extern design
- [consumer-contract-v0.5.md](../../ailang_resources/consumer-contract-v0.5.md) - AILANG contract

---

**Document created**: 2025-12-04
**Last updated**: 2025-12-04
