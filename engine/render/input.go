package render

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"stapledons_voyage/engine/camera"
	"stapledons_voyage/sim_gen"
)

// CaptureInput converts Ebiten input state into FrameInput.
// Deprecated: Use CaptureInputWithCamera for click-to-select support.
func CaptureInput() sim_gen.FrameInput {
	return CaptureInputWithCamera(sim_gen.Camera{X: 0, Y: 0, Zoom: 1.0}, 640, 480)
}

// CaptureInputWithCamera converts Ebiten input state into FrameInput,
// using the provided camera to convert screen coords to world coords.
func CaptureInputWithCamera(cam sim_gen.Camera, screenW, screenH int) sim_gen.FrameInput {
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
	// Capture "just pressed" events (edge detection for mode switching, etc.)
	for _, k := range inpututil.AppendJustPressedKeys(nil) {
		keys = append(keys, sim_gen.KeyEvent{
			Key:  int(k),
			Kind: "pressed",
		})
	}
	// Capture held keys
	for _, k := range inpututil.AppendPressedKeys(nil) {
		keys = append(keys, sim_gen.KeyEvent{
			Key:  int(k),
			Kind: "down",
		})
	}
	// Capture released keys
	for _, k := range inpututil.AppendJustReleasedKeys(nil) {
		keys = append(keys, sim_gen.KeyEvent{
			Key:  int(k),
			Kind: "up",
		})
	}

	// Convert screen coords to world coords using camera
	transform := camera.FromOutput(cam, screenW, screenH)
	worldX, worldY := transform.ScreenToWorld(float64(x), float64(y))

	// Convert screen coords to tile coords using isometric projection
	tileXf, tileYf := ScreenToTile(float64(x), float64(y), cam, screenW, screenH)
	tileX, tileY := int(tileXf), int(tileYf)

	// Detect just-pressed (edge detection, not held)
	clicked := inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)

	// Detect action keys (I=inspect, B=build, X=clear)
	var action sim_gen.PlayerAction = sim_gen.ActionNone{}
	if inpututil.IsKeyJustPressed(ebiten.KeyI) {
		action = sim_gen.ActionInspect{}
	} else if inpututil.IsKeyJustPressed(ebiten.KeyB) {
		action = sim_gen.ActionBuild{StructureType: sim_gen.StructureHouse}
	} else if inpututil.IsKeyJustPressed(ebiten.KeyX) {
		action = sim_gen.ActionClear{}
	}

	return sim_gen.FrameInput{
		Mouse: sim_gen.MouseState{
			X:       float64(x),
			Y:       float64(y),
			Buttons: buttons,
		},
		Keys:             keys,
		ClickedThisFrame: clicked,
		WorldMouseX:      worldX,
		WorldMouseY:      worldY,
		TileMouseX:       tileX,
		TileMouseY:       tileY,
		ActionRequested:  action,
	}
}
