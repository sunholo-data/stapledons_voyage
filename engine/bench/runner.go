package bench

import (
	"fmt"
	"sort"
	"time"

	"stapledons_voyage/sim_gen"
)

// Runner executes benchmarks and checks against thresholds.
type Runner struct {
	thresholds Thresholds
	iterations int
	warmup     int
}

// NewRunner creates a benchmark runner with the given settings.
func NewRunner(iterations, warmup int) *Runner {
	return &Runner{
		thresholds: DefaultThresholds(),
		iterations: iterations,
		warmup:     warmup,
	}
}

// SetThresholds allows overriding the default thresholds.
func (r *Runner) SetThresholds(t Thresholds) {
	r.thresholds = t
}

// RunAll executes all benchmarks and returns a performance report.
func (r *Runner) RunAll() PerfReport {
	report := PerfReport{
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
		Thresholds: r.thresholds,
		Results:    []BenchResult{},
		AllPassed:  true,
	}

	// Run each benchmark
	initResult := r.benchInitWorld()
	report.Results = append(report.Results, initResult)
	if !initResult.Passed {
		report.AllPassed = false
	}

	stepResult := r.benchStep()
	report.Results = append(report.Results, stepResult)
	if !stepResult.Passed {
		report.AllPassed = false
	}

	step100Result := r.benchStep100()
	report.Results = append(report.Results, step100Result)
	if !step100Result.Passed {
		report.AllPassed = false
	}

	return report
}

func (r *Runner) benchInitWorld() BenchResult {
	times := make([]time.Duration, r.iterations)

	// Warmup
	for i := 0; i < r.warmup; i++ {
		_ = sim_gen.InitWorld(int64(i))
	}

	// Benchmark
	for i := 0; i < r.iterations; i++ {
		start := time.Now()
		_ = sim_gen.InitWorld(int64(i))
		times[i] = time.Since(start)
	}

	return r.makeResult("InitWorld", times, r.thresholds.InitWorldMax)
}

func (r *Runner) benchStep() BenchResult {
	times := make([]time.Duration, r.iterations)
	world := sim_gen.InitWorld(int64(42))
	input := &sim_gen.FrameInput{}

	// Warmup
	for i := 0; i < r.warmup; i++ {
		result := sim_gen.Step(world, input)
		if tuple, ok := result.([]interface{}); ok && len(tuple) == 2 {
			if w, ok := tuple[0].(*sim_gen.World); ok {
				world = w
			}
		}
	}

	// Reset
	world = sim_gen.InitWorld(int64(42))

	// Benchmark
	for i := 0; i < r.iterations; i++ {
		start := time.Now()
		result := sim_gen.Step(world, input)
		times[i] = time.Since(start)
		if tuple, ok := result.([]interface{}); ok && len(tuple) == 2 {
			if w, ok := tuple[0].(*sim_gen.World); ok {
				world = w
			}
		}
	}

	return r.makeResult("Step", times, r.thresholds.StepMax)
}

func (r *Runner) benchStep100() BenchResult {
	// Fewer iterations since each does 100 steps
	iterations := r.iterations / 10
	if iterations < 10 {
		iterations = 10
	}
	times := make([]time.Duration, iterations)
	input := &sim_gen.FrameInput{}

	// Warmup
	for i := 0; i < r.warmup/10; i++ {
		world := sim_gen.InitWorld(int64(42))
		for j := 0; j < 100; j++ {
			result := sim_gen.Step(world, input)
			if tuple, ok := result.([]interface{}); ok && len(tuple) == 2 {
				if w, ok := tuple[0].(*sim_gen.World); ok {
					world = w
				}
			}
		}
	}

	// Benchmark
	for i := 0; i < iterations; i++ {
		world := sim_gen.InitWorld(int64(42))
		start := time.Now()
		for j := 0; j < 100; j++ {
			result := sim_gen.Step(world, input)
			if tuple, ok := result.([]interface{}); ok && len(tuple) == 2 {
				if w, ok := tuple[0].(*sim_gen.World); ok {
					world = w
				}
			}
		}
		times[i] = time.Since(start)
	}

	return r.makeResult("Step100", times, r.thresholds.Step100Max)
}

func (r *Runner) makeResult(name string, times []time.Duration, threshold time.Duration) BenchResult {
	sort.Slice(times, func(i, j int) bool {
		return times[i] < times[j]
	})

	var total time.Duration
	for _, t := range times {
		total += t
	}

	n := len(times)
	avg := total / time.Duration(n)
	p50 := times[n*50/100]
	p95 := times[n*95/100]
	p99 := times[n*99/100]

	passed := p95 <= threshold
	message := ""
	if !passed {
		message = fmt.Sprintf("P95 (%v) exceeds threshold (%v)", p95, threshold)
	}

	return BenchResult{
		Name:      name,
		Avg:       avg,
		Min:       times[0],
		Max:       times[n-1],
		P50:       p50,
		P95:       p95,
		P99:       p99,
		Total:     total,
		Ops:       n,
		Passed:    passed,
		Threshold: threshold,
		Message:   message,
	}
}

// ComputeFrameStats computes statistics from a slice of frame times.
func ComputeFrameStats(times []time.Duration, threshold time.Duration) *FrameStats {
	if len(times) == 0 {
		return nil
	}

	sorted := make([]time.Duration, len(times))
	copy(sorted, times)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i] < sorted[j]
	})

	var total time.Duration
	dropped := 0
	for _, t := range times {
		total += t
		if t > threshold {
			dropped++
		}
	}

	n := len(sorted)

	return &FrameStats{
		Count:   n,
		Avg:     total / time.Duration(n),
		Min:     sorted[0],
		Max:     sorted[n-1],
		P50:     sorted[n*50/100],
		P95:     sorted[n*95/100],
		P99:     sorted[n*99/100],
		Dropped: dropped,
	}
}
