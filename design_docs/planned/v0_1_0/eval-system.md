# Evaluation System Design

**Version:** 0.1.0
**Status:** Planned
**Priority:** P1 (Medium)
**Complexity:** Medium
**Packages:** `engine/scenario`, `engine/bench`, `cmd/eval`

## Related Documents

- [Architecture Overview](architecture.md) - System context
- [Engine Layer Design](engine-layer.md) - Runtime integration

## Overview

The evaluation system measures simulation performance and correctness. It produces `out/report.json` for analysis, enabling data-driven improvements to both AILANG and game logic.

## Components

### Benchmarks (`engine/bench/`)

Go benchmarks measuring raw simulation performance.

**BenchmarkInitWorld:**
```go
func BenchmarkInitWorld(b *testing.B) {
    for i := 0; i < b.N; i++ {
        _ = sim_gen.InitWorld(int64(i))
    }
}
```
Measures world initialization time.

**BenchmarkStep:**
```go
func BenchmarkStep(b *testing.B) {
    world := sim_gen.InitWorld(42)
    input := sim_gen.FrameInput{}
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        world, _, _ = sim_gen.Step(world, input)
    }
}
```
Measures single-step throughput (amortized after init).

**BenchmarkStep100:**
```go
func BenchmarkStep100(b *testing.B) {
    for i := 0; i < b.N; i++ {
        world := sim_gen.InitWorld(42)
        for j := 0; j < 100; j++ {
            world, _, _ = sim_gen.Step(world, input)
        }
    }
}
```
Measures 100-tick burst performance (includes init).

### Scenarios (`engine/scenario/`)

Correctness tests that verify simulation behavior.

**runInitScenario:**
- Creates world with seed 42
- Verifies `tick == 0`
- Verifies planet dimensions (expected: 64x64)

**runStepScenario:**
- Runs 100 ticks from init
- Verifies `tick == 100` after completion
- Captures any errors from Step()

### Metrics Tracking (`metrics.go`)

```go
type Metrics struct {
    StartTime    time.Time
    EndTime      time.Time
    TickCount    int
    DrawCmdCount int
}
```

Tracks:
- Total duration
- Tick count
- Draw command count
- Average draw commands per tick

### Report Format (`runner.go`)

```go
type Report struct {
    Benchmarks map[string]BenchResult `json:"benchmarks"`
    Scenarios  []Result               `json:"scenarios"`
}

type BenchResult struct {
    NsPerOp   int64 `json:"ns_per_op"`
    AllocsOp  int64 `json:"allocs_per_op"`
    BytesOp   int64 `json:"bytes_per_op"`
}

type Result struct {
    Name    string `json:"name"`
    Passed  bool   `json:"passed"`
    Ticks   int    `json:"ticks"`
    Message string `json:"message,omitempty"`
}
```

## Output

**Location:** `out/report.json`

**Example:**
```json
{
  "benchmarks": {},
  "scenarios": [
    {
      "name": "init_world",
      "passed": true,
      "ticks": 0
    },
    {
      "name": "step_100_ticks",
      "passed": true,
      "ticks": 100
    }
  ]
}
```

## Build Targets

```bash
make eval
```

Runs:
1. `go test -bench=. -benchmem ./engine/bench > out/bench.txt`
2. `go run ./cmd/eval > out/report.json`

## Usage in Development

### Regression Testing
```bash
make eval
diff out/report.json out/report.json.baseline
```

### Performance Tracking
```bash
make eval
cat out/bench.txt | grep "ns/op"
```

### CI Integration
```bash
make eval
jq '.scenarios[] | select(.passed == false)' out/report.json
# Exit 1 if any scenarios failed
```

## Future Enhancements

### v0.2.0 (Planned)
- [ ] Benchmark result parsing into report.json
- [ ] Determinism verification (same seed = same output)
- [ ] Memory profile integration

### v0.3.0 (Planned)
- [ ] Scenario definitions in YAML/JSON
- [ ] Input replay from recorded sessions
- [ ] Visual diff for rendered output

## File Listing

```
engine/
├── bench/
│   └── bench_test.go     # Go benchmarks
└── scenario/
    ├── runner.go         # RunAll(), scenario execution
    └── metrics.go        # Performance tracking

cmd/eval/
└── main.go               # Report generation entrypoint

out/
├── bench.txt             # Raw benchmark output
└── report.json           # Structured evaluation report
```

## AILANG Integration Value

**Why this matters for AILANG testing:**

| Metric | What it Tests | AILANG Relevance |
|--------|---------------|------------------|
| InitWorld ns/op | World generation | List construction performance |
| Step ns/op | Per-tick logic | Pattern matching, recursion |
| Allocs/op | Memory behavior | ADT instantiation overhead |
| Scenario pass/fail | Correctness | Type system soundness |

**Feedback loop:**
```
eval output → analyze bottlenecks → report to AILANG core → iterate
```

## Success Criteria

### Benchmarks
- [ ] BenchmarkInitWorld runs without panic
- [ ] BenchmarkStep measures single-tick throughput
- [ ] BenchmarkStep100 measures burst performance
- [ ] Memory allocation tracked (allocs/op, bytes/op)

### Scenarios
- [ ] init_world scenario verifies world creation
- [ ] step_100_ticks scenario verifies simulation loop
- [ ] Scenario results captured in JSON format

### Output
- [ ] `out/report.json` generated on `make eval`
- [ ] Report structure matches expected schema
- [ ] Scenarios show pass/fail status
- [ ] Results parseable for CI/CD integration

### Integration Test Purpose
- [ ] Exercise AILANG codegen under real workload
- [ ] Detect performance regressions
- [ ] Validate simulation determinism
