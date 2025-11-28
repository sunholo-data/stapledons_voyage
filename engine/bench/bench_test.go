package bench

import (
	"testing"

	"stapledons_voyage/sim_gen"
)

func BenchmarkInitWorld(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = sim_gen.InitWorld(int64(i))
	}
}

func BenchmarkStep(b *testing.B) {
	world := sim_gen.InitWorld(42)
	input := sim_gen.FrameInput{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var err error
		world, _, err = sim_gen.Step(world, input)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkStep100(b *testing.B) {
	for i := 0; i < b.N; i++ {
		world := sim_gen.InitWorld(42)
		input := sim_gen.FrameInput{}

		for j := 0; j < 100; j++ {
			var err error
			world, _, err = sim_gen.Step(world, input)
			if err != nil {
				b.Fatal(err)
			}
		}
	}
}
