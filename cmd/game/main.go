package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"stapledons_voyage/engine/render"
	"stapledons_voyage/sim_gen"
)

type Game struct {
	world sim_gen.World
	out   sim_gen.FrameOutput
}

func (g *Game) Update() error {
	input := render.CaptureInput()
	w2, out, err := sim_gen.Step(g.world, input)
	if err != nil {
		return err
	}
	g.world = w2
	g.out = out
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	render.RenderFrame(screen, g.out)
}

func (g *Game) Layout(w, h int) (int, int) {
	return 640, 480
}

func main() {
	w := sim_gen.InitWorld(1234)
	game := &Game{world: w}

	ebiten.SetWindowTitle("Stapledons Voyage")
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
