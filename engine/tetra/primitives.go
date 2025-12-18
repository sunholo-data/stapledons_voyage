// Package tetra provides 3D primitives for room construction.
package tetra

import (
	"fmt"
	"math"

	"github.com/solarlune/tetra3d"
)

// NewPlaneMesh creates a horizontal plane mesh (XZ plane, Y=0)
// width extends along X axis, depth extends along Z axis
// UVs are scaled so texture tiles once per meter (width=8 gives U range 0-8)
func NewPlaneMesh(name string, width, depth float32) *tetra3d.Mesh {
	return NewPlaneMeshUV(name, width, depth, width, depth)
}

// NewPlaneMeshUV creates a horizontal plane with explicit UV scale
// uScale and vScale control how many texture tiles fit across the surface
func NewPlaneMeshUV(name string, width, depth, uScale, vScale float32) *tetra3d.Mesh {
	mesh := tetra3d.NewMesh(name)

	halfW := width / 2
	halfD := depth / 2

	// Simple 4-vertex plane with correct winding for visibility from +Y
	mesh.AddVertices(
		tetra3d.VertexInfo{X: -halfW, Y: 0, Z: -halfD, U: 0, V: 0, NormalX: 0, NormalY: 1, NormalZ: 0},         // 0: back-left
		tetra3d.VertexInfo{X: halfW, Y: 0, Z: -halfD, U: uScale, V: 0, NormalX: 0, NormalY: 1, NormalZ: 0},     // 1: back-right
		tetra3d.VertexInfo{X: halfW, Y: 0, Z: halfD, U: uScale, V: vScale, NormalX: 0, NormalY: 1, NormalZ: 0}, // 2: front-right
		tetra3d.VertexInfo{X: -halfW, Y: 0, Z: halfD, U: 0, V: vScale, NormalX: 0, NormalY: 1, NormalZ: 0},     // 3: front-left
	)

	mat := tetra3d.NewMaterial("plane_mat")
	// Two triangles with CCW winding when viewed from +Y (looking down)
	// Triangle 1: back-left → front-left → front-right
	// Triangle 2: back-left → front-right → back-right
	mesh.AddMeshPart(mat, 0, 3, 2, 0, 2, 1)
	mesh.UpdateBounds()
	return mesh
}

// NewPlaneMeshSubdiv creates a subdivided horizontal plane
// divisionsX and divisionsZ control how many grid cells to create
func NewPlaneMeshSubdiv(name string, width, depth, uScale, vScale float32, divisionsX, divisionsZ int) *tetra3d.Mesh {
	mesh := tetra3d.NewMesh(name)

	halfW := width / 2
	halfD := depth / 2

	// Create grid of vertices
	numVertsX := divisionsX + 1
	numVertsZ := divisionsZ + 1

	for iz := 0; iz < numVertsZ; iz++ {
		for ix := 0; ix < numVertsX; ix++ {
			// Position: interpolate from -half to +half
			t_x := float32(ix) / float32(divisionsX)
			t_z := float32(iz) / float32(divisionsZ)
			x := -halfW + t_x*width
			z := -halfD + t_z*depth

			// UVs: interpolate from 0 to scale
			u := t_x * uScale
			v := t_z * vScale

			mesh.AddVertices(tetra3d.VertexInfo{
				X: x, Y: 0, Z: z,
				U: u, V: v,
				NormalX: 0, NormalY: 1, NormalZ: 0,
			})
		}
	}

	// Create triangles for each grid cell
	var indices []int
	for iz := 0; iz < divisionsZ; iz++ {
		for ix := 0; ix < divisionsX; ix++ {
			// Four corners of this grid cell
			bl := iz*numVertsX + ix           // bottom-left
			br := iz*numVertsX + ix + 1       // bottom-right
			tl := (iz+1)*numVertsX + ix       // top-left
			tr := (iz+1)*numVertsX + ix + 1   // top-right

			// Two triangles (CCW from +Y looking down)
			indices = append(indices, bl, tl, tr)
			indices = append(indices, bl, tr, br)
		}
	}

	mat := tetra3d.NewMaterial("plane_mat")
	mesh.AddMeshPart(mat, indices...)

	mesh.UpdateBounds()
	return mesh
}

// NewVerticalPlaneMesh creates a vertical plane mesh for walls
// width extends along X axis, height extends along Y axis, facing +Z direction
// UVs are scaled so texture tiles once per meter
func NewVerticalPlaneMesh(name string, width, height float32) *tetra3d.Mesh {
	return NewVerticalPlaneMeshUV(name, width, height, width, height)
}

// NewVerticalPlaneMeshUV creates a vertical plane with explicit UV scale
// uScale controls horizontal tiling, vScale controls vertical tiling
func NewVerticalPlaneMeshUV(name string, width, height, uScale, vScale float32) *tetra3d.Mesh {
	mesh := tetra3d.NewMesh(name)

	halfW := width / 2
	halfH := height / 2

	// Four corners of plane (in XY plane at Z=0, facing +Z)
	// V is inverted for walls - 0 at top
	mesh.AddVertices(
		tetra3d.VertexInfo{X: -halfW, Y: -halfH, Z: 0, U: 0, V: vScale, NormalX: 0, NormalY: 0, NormalZ: 1},
		tetra3d.VertexInfo{X: halfW, Y: -halfH, Z: 0, U: uScale, V: vScale, NormalX: 0, NormalY: 0, NormalZ: 1},
		tetra3d.VertexInfo{X: halfW, Y: halfH, Z: 0, U: uScale, V: 0, NormalX: 0, NormalY: 0, NormalZ: 1},
		tetra3d.VertexInfo{X: -halfW, Y: halfH, Z: 0, U: 0, V: 0, NormalX: 0, NormalY: 0, NormalZ: 1},
	)

	mat := tetra3d.NewMaterial("wall_mat")
	// Two triangles: (0,1,2) and (0,2,3) - CCW from +Z
	mesh.AddMeshPart(mat, 0, 1, 2, 0, 2, 3)
	mesh.UpdateBounds()
	return mesh
}

// NewVerticalPlaneMeshSubdiv creates a subdivided vertical plane
// divisionsX and divisionsY control how many grid cells to create
func NewVerticalPlaneMeshSubdiv(name string, width, height, uScale, vScale float32, divisionsX, divisionsY int) *tetra3d.Mesh {
	mesh := tetra3d.NewMesh(name)

	halfW := width / 2
	halfH := height / 2

	// Create grid of vertices
	for iy := 0; iy <= divisionsY; iy++ {
		for ix := 0; ix <= divisionsX; ix++ {
			// Position: interpolate from -half to +half
			t_x := float32(ix) / float32(divisionsX)
			t_y := float32(iy) / float32(divisionsY)
			x := -halfW + t_x*width
			y := -halfH + t_y*height

			// UVs: interpolate (V is inverted for walls - 0 at top)
			u := t_x * uScale
			v := (1.0 - t_y) * vScale

			mesh.AddVertices(tetra3d.VertexInfo{
				X: x, Y: y, Z: 0,
				U: u, V: v,
				NormalX: 0, NormalY: 0, NormalZ: 1,
			})
		}
	}

	// Create triangles for each grid cell
	var indices []int
	rowSize := divisionsX + 1
	for iy := 0; iy < divisionsY; iy++ {
		for ix := 0; ix < divisionsX; ix++ {
			// Four corners of this grid cell
			bl := iy*rowSize + ix         // bottom-left
			br := iy*rowSize + ix + 1     // bottom-right
			tl := (iy+1)*rowSize + ix     // top-left
			tr := (iy+1)*rowSize + ix + 1 // top-right

			// Two triangles (CCW from +Z)
			indices = append(indices, bl, br, tr)
			indices = append(indices, bl, tr, tl)
		}
	}

	mat := tetra3d.NewMaterial("wall_mat")
	mesh.AddMeshPart(mat, indices...)

	mesh.UpdateBounds()
	return mesh
}

// NewCubeMesh creates a cube mesh centered at origin
func NewCubeMesh(name string, size float32) *tetra3d.Mesh {
	mesh := tetra3d.NewMesh(name)

	half := size / 2

	// We need separate vertices for each face due to different normals
	// 6 faces * 4 vertices = 24 vertices total
	vertices := []tetra3d.VertexInfo{
		// Front face (+Z) - vertices 0-3
		{X: -half, Y: -half, Z: half, U: 0, V: 1, NormalX: 0, NormalY: 0, NormalZ: 1},
		{X: half, Y: -half, Z: half, U: 1, V: 1, NormalX: 0, NormalY: 0, NormalZ: 1},
		{X: half, Y: half, Z: half, U: 1, V: 0, NormalX: 0, NormalY: 0, NormalZ: 1},
		{X: -half, Y: half, Z: half, U: 0, V: 0, NormalX: 0, NormalY: 0, NormalZ: 1},

		// Back face (-Z) - vertices 4-7
		{X: half, Y: -half, Z: -half, U: 0, V: 1, NormalX: 0, NormalY: 0, NormalZ: -1},
		{X: -half, Y: -half, Z: -half, U: 1, V: 1, NormalX: 0, NormalY: 0, NormalZ: -1},
		{X: -half, Y: half, Z: -half, U: 1, V: 0, NormalX: 0, NormalY: 0, NormalZ: -1},
		{X: half, Y: half, Z: -half, U: 0, V: 0, NormalX: 0, NormalY: 0, NormalZ: -1},

		// Right face (+X) - vertices 8-11
		{X: half, Y: -half, Z: half, U: 0, V: 1, NormalX: 1, NormalY: 0, NormalZ: 0},
		{X: half, Y: -half, Z: -half, U: 1, V: 1, NormalX: 1, NormalY: 0, NormalZ: 0},
		{X: half, Y: half, Z: -half, U: 1, V: 0, NormalX: 1, NormalY: 0, NormalZ: 0},
		{X: half, Y: half, Z: half, U: 0, V: 0, NormalX: 1, NormalY: 0, NormalZ: 0},

		// Left face (-X) - vertices 12-15
		{X: -half, Y: -half, Z: -half, U: 0, V: 1, NormalX: -1, NormalY: 0, NormalZ: 0},
		{X: -half, Y: -half, Z: half, U: 1, V: 1, NormalX: -1, NormalY: 0, NormalZ: 0},
		{X: -half, Y: half, Z: half, U: 1, V: 0, NormalX: -1, NormalY: 0, NormalZ: 0},
		{X: -half, Y: half, Z: -half, U: 0, V: 0, NormalX: -1, NormalY: 0, NormalZ: 0},

		// Top face (+Y) - vertices 16-19
		{X: -half, Y: half, Z: half, U: 0, V: 1, NormalX: 0, NormalY: 1, NormalZ: 0},
		{X: half, Y: half, Z: half, U: 1, V: 1, NormalX: 0, NormalY: 1, NormalZ: 0},
		{X: half, Y: half, Z: -half, U: 1, V: 0, NormalX: 0, NormalY: 1, NormalZ: 0},
		{X: -half, Y: half, Z: -half, U: 0, V: 0, NormalX: 0, NormalY: 1, NormalZ: 0},

		// Bottom face (-Y) - vertices 20-23
		{X: -half, Y: -half, Z: -half, U: 0, V: 1, NormalX: 0, NormalY: -1, NormalZ: 0},
		{X: half, Y: -half, Z: -half, U: 1, V: 1, NormalX: 0, NormalY: -1, NormalZ: 0},
		{X: half, Y: -half, Z: half, U: 1, V: 0, NormalX: 0, NormalY: -1, NormalZ: 0},
		{X: -half, Y: -half, Z: half, U: 0, V: 0, NormalX: 0, NormalY: -1, NormalZ: 0},
	}

	mesh.AddVertices(vertices...)

	// Create material and add all face indices
	// Standard winding: (0,1,2), (0,2,3) for each quad face
	mat := tetra3d.NewMaterial("cube_mat")
	mesh.AddMeshPart(mat,
		// Front face (+Z)
		0, 1, 2, 0, 2, 3,
		// Back face (-Z)
		4, 5, 6, 4, 6, 7,
		// Right face (+X)
		8, 9, 10, 8, 10, 11,
		// Left face (-X)
		12, 13, 14, 12, 14, 15,
		// Top face (+Y)
		16, 17, 18, 16, 18, 19,
		// Bottom face (-Y)
		20, 21, 22, 20, 22, 23,
	)

	mesh.UpdateBounds()
	return mesh
}

// Room represents a simple rectangular room with walls, floor, and ceiling
type Room struct {
	Floor      []*tetra3d.Model // Tiled floor for better rendering
	Ceiling    []*tetra3d.Model // Tiled ceiling for better rendering
	Walls      []*tetra3d.Model // North, South, East, West
	Width      float32
	Depth      float32
	Height     float32
	FloorTiles int // Number of tiles per dimension
}

// NewRoom creates a simple room with the given dimensions (1 texture tile per meter)
func NewRoom(width, depth, height float32) *Room {
	return NewRoomUV(width, depth, height, 1.0)
}

// NewRoomUV creates a room with configurable UV scale
// uvScale controls texture density: 1.0 = 1 tile/meter, 0.5 = 1 tile/2m (larger tiles)
// Floor and ceiling are subdivided into tiles to prevent texture stretching (tetra3d limitation)
func NewRoomUV(width, depth, height, uvScale float32) *Room {
	const tilesPerDim = 4 // 4x4 grid of tiles = 16 floor tiles, 16 ceiling tiles

	room := &Room{
		Width:      width,
		Depth:      depth,
		Height:     height,
		FloorTiles: tilesPerDim,
	}

	// Calculate UV dimensions based on scale
	wallUWidth := width * uvScale
	wallUDepth := depth * uvScale
	wallV := height * uvScale

	// Tile dimensions
	tileW := width / float32(tilesPerDim)
	tileD := depth / float32(tilesPerDim)
	tileU := tileW * uvScale
	tileV := tileD * uvScale

	// Create tiled floor (grid of small planes)
	halfW := width / 2
	halfD := depth / 2
	for iz := 0; iz < tilesPerDim; iz++ {
		for ix := 0; ix < tilesPerDim; ix++ {
			tileMesh := NewPlaneMeshUV(fmt.Sprintf("floor_%d_%d", ix, iz), tileW, tileD, tileU, tileV)
			tile := tetra3d.NewModel(fmt.Sprintf("floor_tile_%d_%d", ix, iz), tileMesh)
			// Position tile in grid (centered at room origin)
			tileX := -halfW + tileW/2 + float32(ix)*tileW
			tileZ := -halfD + tileD/2 + float32(iz)*tileD
			tile.SetLocalPosition(tileX, 0, tileZ)
			room.Floor = append(room.Floor, tile)
		}
	}

	// Create tiled ceiling (same grid, at ceiling height, rotated to face down)
	for iz := 0; iz < tilesPerDim; iz++ {
		for ix := 0; ix < tilesPerDim; ix++ {
			tileMesh := NewPlaneMeshUV(fmt.Sprintf("ceiling_%d_%d", ix, iz), tileW, tileD, tileU, tileV)
			tile := tetra3d.NewModel(fmt.Sprintf("ceiling_tile_%d_%d", ix, iz), tileMesh)
			tileX := -halfW + tileW/2 + float32(ix)*tileW
			tileZ := -halfD + tileD/2 + float32(iz)*tileD
			tile.SetLocalPosition(tileX, height, tileZ)
			tile.Rotate(1, 0, 0, float32(math.Pi)) // Face down
			room.Ceiling = append(room.Ceiling, tile)
		}
	}

	// Create walls - all should face INWARD toward the room center
	// Vertical plane mesh faces +Z by default

	// North wall at -Z edge, should face +Z (toward center)
	northMesh := NewVerticalPlaneMeshUV("north_wall", width, height, wallUWidth, wallV)
	north := tetra3d.NewModel("north_wall_model", northMesh)
	north.SetLocalPosition(0, height/2, -depth/2)
	// Default +Z facing is correct (faces toward center)

	// South wall at +Z edge, should face -Z (toward center)
	southMesh := NewVerticalPlaneMeshUV("south_wall", width, height, wallUWidth, wallV)
	south := tetra3d.NewModel("south_wall_model", southMesh)
	south.SetLocalPosition(0, height/2, depth/2)
	south.Rotate(0, 1, 0, float32(math.Pi)) // 180° to face inward

	// East wall at +X edge, should face -X (toward center)
	eastMesh := NewVerticalPlaneMeshUV("east_wall", depth, height, wallUDepth, wallV)
	east := tetra3d.NewModel("east_wall_model", eastMesh)
	east.SetLocalPosition(width/2, height/2, 0)
	east.Rotate(0, 1, 0, float32(-math.Pi/2)) // -90° to face -X

	// West wall at -X edge, should face +X (toward center)
	westMesh := NewVerticalPlaneMeshUV("west_wall", depth, height, wallUDepth, wallV)
	west := tetra3d.NewModel("west_wall_model", westMesh)
	west.SetLocalPosition(-width/2, height/2, 0)
	west.Rotate(0, 1, 0, float32(math.Pi/2)) // +90° to face +X

	room.Walls = []*tetra3d.Model{north, south, east, west}

	return room
}

// AddToScene adds all room components to a scene
func (r *Room) AddToScene(scene *Scene) {
	for _, tile := range r.Floor {
		scene.Root().AddChildren(tile)
	}
	for _, tile := range r.Ceiling {
		scene.Root().AddChildren(tile)
	}
	for _, wall := range r.Walls {
		scene.Root().AddChildren(wall)
	}
}

// SetFloorMaterial sets the material for all floor tiles
func (r *Room) SetFloorMaterial(mat *tetra3d.Material) {
	for _, tile := range r.Floor {
		if len(tile.Mesh.MeshParts) > 0 {
			tile.Mesh.MeshParts[0].Material = mat
		}
	}
}

// SetCeilingMaterial sets the material for all ceiling tiles
func (r *Room) SetCeilingMaterial(mat *tetra3d.Material) {
	for _, tile := range r.Ceiling {
		if len(tile.Mesh.MeshParts) > 0 {
			tile.Mesh.MeshParts[0].Material = mat
		}
	}
}

// SetWallMaterial sets the material for all walls
func (r *Room) SetWallMaterial(mat *tetra3d.Material) {
	for _, wall := range r.Walls {
		if len(wall.Mesh.MeshParts) > 0 {
			wall.Mesh.MeshParts[0].Material = mat
		}
	}
}
