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
	worldIface := sim_gen.InitWorld(int64(42))
	world, ok := worldIface.(*sim_gen.World)
	if !ok {
		return Result{
			Name:    "init_world",
			Passed:  false,
			Message: "InitWorld did not return *World",
		}
	}

	if world.Tick != 0 {
		return Result{
			Name:    "init_world",
			Passed:  false,
			Message: "world tick should start at 0",
		}
	}

	// Note: mock uses 8x8, not 64x64
	if world.Planet.Width != 8 || world.Planet.Height != 8 {
		return Result{
			Name:    "init_world",
			Passed:  false,
			Message: "planet dimensions should be 8x8",
		}
	}

	return Result{
		Name:   "init_world",
		Passed: true,
		Ticks:  0,
	}
}

func runStepScenario() Result {
	// Initialize world - type assert to *World (M-DX16: struct types preserved)
	worldIface := sim_gen.InitWorld(int64(42))
	world, ok := worldIface.(*sim_gen.World)
	if !ok {
		return Result{
			Name:    "step_100_ticks",
			Passed:  false,
			Message: "InitWorld did not return *World",
		}
	}

	input := sim_gen.FrameInput{
		Keys:            []*sim_gen.KeyEvent{},
		ActionRequested: *sim_gen.NewPlayerActionActionNone(),
	}

	// Run 100 ticks - RecordUpdate preserves *World type through loop
	for i := 0; i < 100; i++ {
		result := sim_gen.Step(world, input)
		tuple, ok := result.([]interface{})
		if !ok || len(tuple) != 2 {
			return Result{
				Name:    "step_100_ticks",
				Passed:  false,
				Message: "Step did not return (World, FrameOutput) tuple",
			}
		}
		if w, ok := tuple[0].(*sim_gen.World); ok {
			world = w
		}
	}

	// World should still be typed after 100 steps
	worldTyped := world

	if worldTyped.Tick != 100 {
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
