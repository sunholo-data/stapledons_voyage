package bench

import (
	"testing"

	"stapledons_voyage/sim_gen"
)

func BenchmarkInitWorld(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = sim_gen.InitWorld(i)
	}
}

func BenchmarkStep(b *testing.B) {
	world := sim_gen.InitWorld(42)
	input := sim_gen.FrameInput{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := sim_gen.Step(world, input)
		tuple, ok := result.([]interface{})
		if !ok || len(tuple) != 2 {
			b.Fatal("unexpected Step result")
		}
		world = tuple[0]
	}
}

func BenchmarkStep100(b *testing.B) {
	for i := 0; i < b.N; i++ {
		world := sim_gen.InitWorld(42)
		input := sim_gen.FrameInput{}

		for j := 0; j < 100; j++ {
			result := sim_gen.Step(world, input)
			tuple, ok := result.([]interface{})
			if !ok || len(tuple) != 2 {
				b.Fatal("unexpected Step result")
			}
			world = tuple[0]
		}
	}
}
