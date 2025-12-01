// Package render provides isometric projection and rendering functions.
package render

import (
	"stapledons_voyage/sim_gen"
)

// =============================================================================
// Isometric Projection Constants
// =============================================================================

const (
	// TileWidth is the width of a tile in world units (before zoom)
	TileWidth = 64.0

	// TileHeight is the height of a tile in world units (isometric = half width)
	TileHeight = 32.0

	// HeightScale is how many world units per height level
	HeightScale = 16.0
)

// =============================================================================
// Isometric Projection Functions
// =============================================================================

// TileToWorld converts tile coordinates to world coordinates (before camera transform).
// This is the core isometric projection formula.
//
// The isometric projection maps a 2D tile grid to a diamond shape:
//
//	       (0,0)
//	      /    \
//	   (1,0)  (0,1)
//	     |      |
//	   (2,0)  (1,1)  (0,2)
//	        ...
func TileToWorld(tile sim_gen.Coord, height int) (worldX, worldY float64) {
	// Standard isometric projection formula
	worldX = float64(tile.X-tile.Y) * (TileWidth / 2)
	worldY = float64(tile.X+tile.Y) * (TileHeight / 2) - float64(height)*HeightScale
	return
}

// TileToWorldWithOffset converts tile + sub-tile offset to world coordinates.
// Offset values are in tile units (-0.5 to 0.5 for smooth movement).
func TileToWorldWithOffset(tile sim_gen.Coord, offsetX, offsetY float64, height int) (worldX, worldY float64) {
	// Apply offset as fractional tile position
	fx := float64(tile.X) + offsetX
	fy := float64(tile.Y) + offsetY

	worldX = (fx - fy) * (TileWidth / 2)
	worldY = (fx + fy) * (TileHeight / 2) - float64(height)*HeightScale
	return
}

// WorldToTile converts world coordinates back to tile coordinates.
// Returns fractional tile position for sub-tile accuracy.
// Note: height is not recoverable from 2D projection, assumes height=0.
func WorldToTile(worldX, worldY float64) (tileX, tileY float64) {
	// Inverse of TileToWorld (solving for tileX, tileY)
	// worldX = (tileX - tileY) * (TileWidth / 2)
	// worldY = (tileX + tileY) * (TileHeight / 2)
	//
	// Let a = worldX / (TileWidth / 2) = tileX - tileY
	// Let b = worldY / (TileHeight / 2) = tileX + tileY
	//
	// tileX = (a + b) / 2
	// tileY = (b - a) / 2

	a := worldX / (TileWidth / 2)
	b := worldY / (TileHeight / 2)

	tileX = (a + b) / 2
	tileY = (b - a) / 2
	return
}

// ScreenToTile converts screen coordinates to tile coordinates.
// Uses the camera transform to go screen → world → tile.
func ScreenToTile(screenX, screenY float64, cam sim_gen.Camera, screenW, screenH int) (tileX, tileY float64) {
	// First, convert screen to world using camera
	// Camera center is at screen center, so:
	// worldX = (screenX - screenW/2) / zoom + cam.X
	// worldY = (screenY - screenH/2) / zoom + cam.Y

	worldX := (screenX-float64(screenW)/2)/cam.Zoom + cam.X
	worldY := (screenY-float64(screenH)/2)/cam.Zoom + cam.Y

	// Then convert world to tile
	return WorldToTile(worldX, worldY)
}

// TileToScreen converts tile coordinates to screen coordinates.
// Uses camera transform to go tile → world → screen.
func TileToScreen(tile sim_gen.Coord, height int, cam sim_gen.Camera, screenW, screenH int) (screenX, screenY float64) {
	// First convert tile to world
	worldX, worldY := TileToWorld(tile, height)

	// Then convert world to screen using camera
	// screenX = (worldX - cam.X) * zoom + screenW/2
	// screenY = (worldY - cam.Y) * zoom + screenH/2

	screenX = (worldX-cam.X)*cam.Zoom + float64(screenW)/2
	screenY = (worldY-cam.Y)*cam.Zoom + float64(screenH)/2
	return
}

// TileToScreenWithOffset converts tile + offset to screen coordinates.
func TileToScreenWithOffset(tile sim_gen.Coord, offsetX, offsetY float64, height int, cam sim_gen.Camera, screenW, screenH int) (screenX, screenY float64) {
	// First convert tile+offset to world
	worldX, worldY := TileToWorldWithOffset(tile, offsetX, offsetY, height)

	// Then convert world to screen using camera
	screenX = (worldX-cam.X)*cam.Zoom + float64(screenW)/2
	screenY = (worldY-cam.Y)*cam.Zoom + float64(screenH)/2
	return
}

// =============================================================================
// Isometric Sorting
// =============================================================================

// IsoDepth calculates the depth value for sorting isometric objects.
// Higher depth values should be drawn later (on top).
// Formula: layer * 10000 + screenY (so layer dominates, then screenY for same layer)
func IsoDepth(layer int, screenY float64) float64 {
	return float64(layer)*10000 + screenY
}

// =============================================================================
// Tile Bounds
// =============================================================================

// TileBounds returns the screen-space bounding box for a tile.
// Returns (minX, minY, maxX, maxY) in screen coordinates.
func TileBounds(tile sim_gen.Coord, height int, cam sim_gen.Camera, screenW, screenH int) (minX, minY, maxX, maxY float64) {
	// Get center of tile
	cx, cy := TileToScreen(tile, height, cam, screenW, screenH)

	// Tile is a diamond shape, so bounds extend TileWidth/2 horizontally
	// and TileHeight/2 vertically from center
	halfW := (TileWidth / 2) * cam.Zoom
	halfH := (TileHeight / 2) * cam.Zoom

	minX = cx - halfW
	maxX = cx + halfW
	minY = cy - halfH
	maxY = cy + halfH
	return
}

// TileInView checks if a tile is visible on screen.
func TileInView(tile sim_gen.Coord, height int, cam sim_gen.Camera, screenW, screenH int) bool {
	minX, minY, maxX, maxY := TileBounds(tile, height, cam, screenW, screenH)

	// Check if bounds overlap with screen
	return maxX >= 0 && minX <= float64(screenW) &&
		maxY >= 0 && minY <= float64(screenH)
}
