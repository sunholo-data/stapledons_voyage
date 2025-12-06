//go:build !ci

package render

import (
	"math"
	"testing"

	"stapledons_voyage/sim_gen"
)

// tolerance for floating point comparisons
const epsilon = 0.001

func floatEqual(a, b float64) bool {
	return math.Abs(a-b) < epsilon
}

func TestTileToWorld_Origin(t *testing.T) {
	// Tile (0,0) at height 0 should be at world origin
	tile := sim_gen.Coord{X: 0, Y: 0}
	worldX, worldY := TileToWorld(tile, 0)

	if !floatEqual(worldX, 0) || !floatEqual(worldY, 0) {
		t.Errorf("TileToWorld(0,0,0) = (%f, %f), want (0, 0)", worldX, worldY)
	}
}

func TestTileToWorld_PositiveX(t *testing.T) {
	// Tile (1,0) should be to the right and down from origin
	tile := sim_gen.Coord{X: 1, Y: 0}
	worldX, worldY := TileToWorld(tile, 0)

	// worldX = (1-0) * 32 = 32
	// worldY = (1+0) * 16 = 16
	if !floatEqual(worldX, 32) || !floatEqual(worldY, 16) {
		t.Errorf("TileToWorld(1,0,0) = (%f, %f), want (32, 16)", worldX, worldY)
	}
}

func TestTileToWorld_PositiveY(t *testing.T) {
	// Tile (0,1) should be to the left and down from origin
	tile := sim_gen.Coord{X: 0, Y: 1}
	worldX, worldY := TileToWorld(tile, 0)

	// worldX = (0-1) * 32 = -32
	// worldY = (0+1) * 16 = 16
	if !floatEqual(worldX, -32) || !floatEqual(worldY, 16) {
		t.Errorf("TileToWorld(0,1,0) = (%f, %f), want (-32, 16)", worldX, worldY)
	}
}

func TestTileToWorld_Height(t *testing.T) {
	// Tile (0,0) at height 1 should be shifted up
	tile := sim_gen.Coord{X: 0, Y: 0}
	worldX, worldY := TileToWorld(tile, 1)

	// worldX = 0
	// worldY = 0 - 16 = -16
	if !floatEqual(worldX, 0) || !floatEqual(worldY, -16) {
		t.Errorf("TileToWorld(0,0,1) = (%f, %f), want (0, -16)", worldX, worldY)
	}
}

func TestWorldToTile_RoundTrip(t *testing.T) {
	// Test that WorldToTile is inverse of TileToWorld (at height 0)
	testCases := []sim_gen.Coord{
		{X: 0, Y: 0},
		{X: 1, Y: 0},
		{X: 0, Y: 1},
		{X: 1, Y: 1},
		{X: 5, Y: 3},
		{X: -2, Y: 4},
	}

	for _, tile := range testCases {
		worldX, worldY := TileToWorld(tile, 0)
		gotX, gotY := WorldToTile(worldX, worldY)

		if !floatEqual(gotX, float64(tile.X)) || !floatEqual(gotY, float64(tile.Y)) {
			t.Errorf("WorldToTile(TileToWorld(%d,%d)) = (%f, %f), want (%d, %d)",
				tile.X, tile.Y, gotX, gotY, tile.X, tile.Y)
		}
	}
}

func TestTileToScreen_WithCamera(t *testing.T) {
	// Camera at world origin, 1x zoom, 640x480 screen
	cam := sim_gen.Camera{X: 0, Y: 0, Zoom: 1.0}
	screenW, screenH := 640, 480

	// Tile (0,0) should be at screen center
	tile := sim_gen.Coord{X: 0, Y: 0}
	sx, sy := TileToScreen(tile, 0, cam, screenW, screenH)

	if !floatEqual(sx, 320) || !floatEqual(sy, 240) {
		t.Errorf("TileToScreen(0,0) with centered camera = (%f, %f), want (320, 240)", sx, sy)
	}
}

func TestTileToScreen_WithCameraOffset(t *testing.T) {
	// Camera offset to the right in world space
	cam := sim_gen.Camera{X: 64, Y: 0, Zoom: 1.0}
	screenW, screenH := 640, 480

	// Tile (0,0) at world (0,0) should be left of center
	tile := sim_gen.Coord{X: 0, Y: 0}
	sx, sy := TileToScreen(tile, 0, cam, screenW, screenH)

	// screenX = (0 - 64) * 1 + 320 = 256
	if !floatEqual(sx, 256) || !floatEqual(sy, 240) {
		t.Errorf("TileToScreen with camera offset = (%f, %f), want (256, 240)", sx, sy)
	}
}

func TestTileToScreen_WithZoom(t *testing.T) {
	// Camera at origin with 2x zoom
	cam := sim_gen.Camera{X: 0, Y: 0, Zoom: 2.0}
	screenW, screenH := 640, 480

	// Tile (1,0) at world (32, 16)
	tile := sim_gen.Coord{X: 1, Y: 0}
	sx, sy := TileToScreen(tile, 0, cam, screenW, screenH)

	// screenX = (32 - 0) * 2 + 320 = 384
	// screenY = (16 - 0) * 2 + 240 = 272
	if !floatEqual(sx, 384) || !floatEqual(sy, 272) {
		t.Errorf("TileToScreen with 2x zoom = (%f, %f), want (384, 272)", sx, sy)
	}
}

func TestScreenToTile_RoundTrip(t *testing.T) {
	cam := sim_gen.Camera{X: 0, Y: 0, Zoom: 1.0}
	screenW, screenH := 640, 480

	testTiles := []sim_gen.Coord{
		{X: 0, Y: 0},
		{X: 2, Y: 1},
		{X: 3, Y: 5},
	}

	for _, tile := range testTiles {
		sx, sy := TileToScreen(tile, 0, cam, screenW, screenH)
		gotX, gotY := ScreenToTile(sx, sy, cam, screenW, screenH)

		if !floatEqual(gotX, float64(tile.X)) || !floatEqual(gotY, float64(tile.Y)) {
			t.Errorf("ScreenToTile(TileToScreen(%d,%d)) = (%f, %f), want (%d, %d)",
				tile.X, tile.Y, gotX, gotY, tile.X, tile.Y)
		}
	}
}

func TestIsoDepth(t *testing.T) {
	// Layer 0, screenY 100 should be less than layer 1, screenY 50
	depth0 := IsoDepth(0, 100)
	depth1 := IsoDepth(1, 50)

	if depth0 >= depth1 {
		t.Errorf("IsoDepth(0, 100)=%f should be less than IsoDepth(1, 50)=%f", depth0, depth1)
	}

	// Same layer, higher screenY should have higher depth
	depthA := IsoDepth(0, 100)
	depthB := IsoDepth(0, 200)

	if depthA >= depthB {
		t.Errorf("IsoDepth(0, 100)=%f should be less than IsoDepth(0, 200)=%f", depthA, depthB)
	}
}

func TestTileInView(t *testing.T) {
	cam := sim_gen.Camera{X: 0, Y: 0, Zoom: 1.0}
	screenW, screenH := 640, 480

	// Tile at origin should be in view
	if !TileInView(sim_gen.Coord{X: 0, Y: 0}, 0, cam, screenW, screenH) {
		t.Error("Tile (0,0) should be in view with centered camera")
	}

	// Tile very far away should not be in view
	if TileInView(sim_gen.Coord{X: 100, Y: 100}, 0, cam, screenW, screenH) {
		t.Error("Tile (100,100) should NOT be in view")
	}
}

// TestClickWithinTileDiamond verifies that clicking within a tile diamond
// returns that same tile. Note: diamond vertices are shared between tiles,
// so we test at 50% offset from center (well inside the diamond).
func TestClickWithinTileDiamond(t *testing.T) {
	cam := sim_gen.Camera{X: 0, Y: 0, Zoom: 1.0}
	screenW, screenH := 640, 480

	// Test tile (5, 5)
	tile := sim_gen.Coord{X: 5, Y: 5}
	centerX, centerY := TileToScreen(tile, 0, cam, screenW, screenH)

	// The diamond extends TileWidth/2 horizontally and TileHeight/2 vertically
	halfW := TileWidth / 2 * cam.Zoom // 32
	halfH := TileHeight / 2 * cam.Zoom // 16

	// Test clicking at center
	gotX, gotY := ScreenToTile(centerX, centerY, cam, screenW, screenH)
	if int(math.Round(gotX)) != 5 || int(math.Round(gotY)) != 5 {
		t.Errorf("Click at tile center: got tile (%d,%d), want (5,5)", int(math.Round(gotX)), int(math.Round(gotY)))
	}

	// Test clicking 50% toward top of diamond (well inside the tile)
	gotX, gotY = ScreenToTile(centerX, centerY-halfH*0.5, cam, screenW, screenH)
	if int(math.Round(gotX)) != 5 || int(math.Round(gotY)) != 5 {
		t.Errorf("Click toward top of diamond: got tile (%d,%d), want (5,5)", int(math.Round(gotX)), int(math.Round(gotY)))
	}

	// Test clicking 50% toward bottom of diamond
	gotX, gotY = ScreenToTile(centerX, centerY+halfH*0.5, cam, screenW, screenH)
	if int(math.Round(gotX)) != 5 || int(math.Round(gotY)) != 5 {
		t.Errorf("Click toward bottom of diamond: got tile (%d,%d), want (5,5)", int(math.Round(gotX)), int(math.Round(gotY)))
	}

	// Test clicking 50% toward left of diamond
	gotX, gotY = ScreenToTile(centerX-halfW*0.5, centerY, cam, screenW, screenH)
	if int(math.Round(gotX)) != 5 || int(math.Round(gotY)) != 5 {
		t.Errorf("Click toward left of diamond: got tile (%d,%d), want (5,5)", int(math.Round(gotX)), int(math.Round(gotY)))
	}

	// Test clicking 50% toward right of diamond
	gotX, gotY = ScreenToTile(centerX+halfW*0.5, centerY, cam, screenW, screenH)
	if int(math.Round(gotX)) != 5 || int(math.Round(gotY)) != 5 {
		t.Errorf("Click toward right of diamond: got tile (%d,%d), want (5,5)", int(math.Round(gotX)), int(math.Round(gotY)))
	}
}

// TestClickWithCameraOffset verifies click detection works with camera offset
func TestClickWithCameraOffset(t *testing.T) {
	// Camera looking at world position (320, 160) which is tile (10, 0) + (0, 10) area
	cam := sim_gen.Camera{X: 320, Y: 160, Zoom: 1.0}
	screenW, screenH := 640, 480

	// Tile (10, 0) should be visible at screen center
	tile := sim_gen.Coord{X: 10, Y: 0}
	centerX, centerY := TileToScreen(tile, 0, cam, screenW, screenH)

	gotX, gotY := ScreenToTile(centerX, centerY, cam, screenW, screenH)
	if int(math.Round(gotX)) != 10 || int(math.Round(gotY)) != 0 {
		t.Errorf("Click with camera offset: got tile (%d,%d), want (10,0). Center=(%f,%f)",
			int(math.Round(gotX)), int(math.Round(gotY)), centerX, centerY)
	}
}

func TestTileToWorldWithOffset(t *testing.T) {
	tile := sim_gen.Coord{X: 0, Y: 0}

	// No offset should match TileToWorld
	wx1, wy1 := TileToWorld(tile, 0)
	wx2, wy2 := TileToWorldWithOffset(tile, 0, 0, 0)

	if !floatEqual(wx1, wx2) || !floatEqual(wy1, wy2) {
		t.Errorf("TileToWorldWithOffset(0,0,0,0,0) = (%f,%f), TileToWorld = (%f,%f)", wx2, wy2, wx1, wy1)
	}

	// Offset of 0.5 in X should move half a tile in world space
	wx3, wy3 := TileToWorldWithOffset(tile, 0.5, 0, 0)
	expectedX := 0.5 * (TileWidth / 2)   // 16
	expectedY := 0.5 * (TileHeight / 2)  // 8

	if !floatEqual(wx3, expectedX) || !floatEqual(wy3, expectedY) {
		t.Errorf("TileToWorldWithOffset(0,0,0.5,0,0) = (%f,%f), want (%f,%f)", wx3, wy3, expectedX, expectedY)
	}
}
