package scenario

import (
	"stapledons_voyage/sim_gen"
)

// Result holds the outcome of a scenario run
type Result struct {
	Name    string `json:"name"`
	Passed  bool   `json:"passed"`
	Ticks   int    `json:"ticks"`
	Message string `json:"message,omitempty"`
}

// Report holds all benchmark and scenario results
type Report struct {
	Benchmarks map[string]BenchResult `json:"benchmarks"`
	Scenarios  []Result               `json:"scenarios"`
}

// BenchResult holds benchmark metrics
type BenchResult struct {
	NsPerOp   int64 `json:"ns_per_op"`
	AllocsOp  int64 `json:"allocs_per_op"`
	BytesOp   int64 `json:"bytes_per_op"`
}

// RunAll executes all defined scenarios and returns a report
func RunAll() Report {
	report := Report{
		Benchmarks: make(map[string]BenchResult),
		Scenarios:  []Result{},
	}

	// Run basic world initialization scenario
	report.Scenarios = append(report.Scenarios, runInitScenario())

	// Run step simulation scenario
	report.Scenarios = append(report.Scenarios, runStepScenario())

	return report
}

func runInitScenario() Result {
	world := sim_gen.InitWorld(42)

	if world.Tick != 0 {
		return Result{
			Name:    "init_world",
			Passed:  false,
			Message: "world tick should start at 0",
		}
	}

	if world.Planet.Width != 64 || world.Planet.Height != 64 {
		return Result{
			Name:    "init_world",
			Passed:  false,
			Message: "planet dimensions should be 64x64",
		}
	}

	return Result{
		Name:   "init_world",
		Passed: true,
		Ticks:  0,
	}
}

func runStepScenario() Result {
	world := sim_gen.InitWorld(42)
	input := sim_gen.FrameInput{}

	// Run 100 ticks
	var err error
	for i := 0; i < 100; i++ {
		world, _, err = sim_gen.Step(world, input)
		if err != nil {
			return Result{
				Name:    "step_100_ticks",
				Passed:  false,
				Message: err.Error(),
			}
		}
	}

	if world.Tick != 100 {
		return Result{
			Name:    "step_100_ticks",
			Passed:  false,
			Message: "tick count mismatch after 100 steps",
		}
	}

	return Result{
		Name:   "step_100_ticks",
		Passed: true,
		Ticks:  100,
	}
}
