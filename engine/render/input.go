package render

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"stapledons_voyage/sim_gen"
)

// CaptureInput converts Ebiten input state into FrameInput
func CaptureInput() sim_gen.FrameInput {
	x, y := ebiten.CursorPosition()

	var buttons []int
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		buttons = append(buttons, 0)
	}
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) {
		buttons = append(buttons, 1)
	}
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonMiddle) {
		buttons = append(buttons, 2)
	}

	var keys []sim_gen.KeyEvent
	for _, k := range inpututil.AppendPressedKeys(nil) {
		keys = append(keys, sim_gen.KeyEvent{
			Key:  int(k),
			Kind: "down",
		})
	}
	for _, k := range inpututil.AppendJustReleasedKeys(nil) {
		keys = append(keys, sim_gen.KeyEvent{
			Key:  int(k),
			Kind: "up",
		})
	}

	return sim_gen.FrameInput{
		Mouse: sim_gen.MouseState{
			X:       float64(x),
			Y:       float64(y),
			Buttons: buttons,
		},
		Keys: keys,
	}
}
