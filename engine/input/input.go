// Package input provides input handling utilities for the game engine.
package input

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// IsMouseJustPressed returns true if the left mouse button was just pressed this frame.
// Uses Ebiten's inpututil to detect edge (not held).
func IsMouseJustPressed() bool {
	return inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)
}

// IsRightMouseJustPressed returns true if the right mouse button was just pressed.
func IsRightMouseJustPressed() bool {
	return inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight)
}

// GetMousePosition returns the current mouse position in screen coordinates.
func GetMousePosition() (x, y int) {
	return ebiten.CursorPosition()
}

// GetMousePositionFloat returns the mouse position as float64 for precision calculations.
func GetMousePositionFloat() (x, y float64) {
	ix, iy := ebiten.CursorPosition()
	return float64(ix), float64(iy)
}
