package sim_gen

// DrawCmd is the interface for all draw commands (tagged union in AILANG)
type DrawCmd interface {
	isDrawCmd()
}

// DrawCmdRect draws a solid color rectangle
// AILANG: Rect(float, float, float, float, int, int) -- x, y, w, h, color, z
type DrawCmdRect struct {
	X     float64
	Y     float64
	W     float64
	H     float64
	Color int
	Z     int
}

func (DrawCmdRect) isDrawCmd() {}

// DrawCmdSprite draws a sprite by ID
// AILANG: Sprite(int, float, float, int) -- id, x, y, z
type DrawCmdSprite struct {
	ID int
	X  float64
	Y  float64
	Z  int
}

func (DrawCmdSprite) isDrawCmd() {}

// DrawCmdText draws text at a position
// AILANG: Text(string, float, float, int) -- text, x, y, z
type DrawCmdText struct {
	Text string
	X    float64
	Y    float64
	Z    int
}

func (DrawCmdText) isDrawCmd() {}
