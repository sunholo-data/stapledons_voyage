// Package tetra provides 3D primitives for room construction.
package tetra

import (
	"math"

	"github.com/solarlune/tetra3d"
)

// NewPlaneMesh creates a horizontal plane mesh (XZ plane, Y=0)
// width extends along X axis, depth extends along Z axis
func NewPlaneMesh(name string, width, depth float32) *tetra3d.Mesh {
	mesh := tetra3d.NewMesh(name)

	halfW := width / 2
	halfD := depth / 2

	// Define 4 corners with UVs and normals pointing up (+Y)
	vertices := []tetra3d.VertexInfo{
		{X: -halfW, Y: 0, Z: -halfD, U: 0, V: 0, NormalX: 0, NormalY: 1, NormalZ: 0}, // Back-left
		{X: halfW, Y: 0, Z: -halfD, U: 1, V: 0, NormalX: 0, NormalY: 1, NormalZ: 0},  // Back-right
		{X: halfW, Y: 0, Z: halfD, U: 1, V: 1, NormalX: 0, NormalY: 1, NormalZ: 0},   // Front-right
		{X: -halfW, Y: 0, Z: halfD, U: 0, V: 1, NormalX: 0, NormalY: 1, NormalZ: 0},  // Front-left
	}

	mesh.AddVertices(vertices...)

	// Create triangles - counter-clockwise from +Y (looking down at floor)
	// Vertices: 0=back-left, 1=back-right, 2=front-right, 3=front-left
	// From +Y looking down: 0 is at top-left, 1 at top-right, 2 at bottom-right, 3 at bottom-left
	// CCW from that view: 0 → 3 → 2 → 1
	mat := tetra3d.NewMaterial("plane_mat")
	mesh.AddMeshPart(mat,
		0, 3, 2, // First triangle
		0, 2, 1, // Second triangle
	)

	mesh.UpdateBounds()
	return mesh
}

// NewVerticalPlaneMesh creates a vertical plane mesh for walls
// width extends along X axis, height extends along Y axis, facing +Z direction
func NewVerticalPlaneMesh(name string, width, height float32) *tetra3d.Mesh {
	mesh := tetra3d.NewMesh(name)

	halfW := width / 2
	halfH := height / 2

	// Define 4 corners with UVs and normals pointing forward (+Z)
	vertices := []tetra3d.VertexInfo{
		{X: -halfW, Y: -halfH, Z: 0, U: 0, V: 1, NormalX: 0, NormalY: 0, NormalZ: 1}, // Bottom-left
		{X: halfW, Y: -halfH, Z: 0, U: 1, V: 1, NormalX: 0, NormalY: 0, NormalZ: 1},  // Bottom-right
		{X: halfW, Y: halfH, Z: 0, U: 1, V: 0, NormalX: 0, NormalY: 0, NormalZ: 1},   // Top-right
		{X: -halfW, Y: halfH, Z: 0, U: 0, V: 0, NormalX: 0, NormalY: 0, NormalZ: 1},  // Top-left
	}

	mesh.AddVertices(vertices...)

	// Create triangles (counter-clockwise when viewed from +Z = normals point +Z)
	// Viewer at +Z looking at plane: 0(left-bottom), 1(right-bottom), 2(right-top), 3(left-top)
	// CCW from that view: 0 → 1 → 2 → 3
	mat := tetra3d.NewMaterial("wall_mat")
	mesh.AddMeshPart(mat,
		0, 1, 2, // First triangle
		0, 2, 3, // Second triangle
	)

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
	Floor   *tetra3d.Model
	Ceiling *tetra3d.Model
	Walls   []*tetra3d.Model // North, South, East, West
	Width   float32
	Depth   float32
	Height  float32
}

// NewRoom creates a simple room with the given dimensions
func NewRoom(width, depth, height float32) *Room {
	room := &Room{
		Width:  width,
		Depth:  depth,
		Height: height,
	}

	// Create floor
	floorMesh := NewPlaneMesh("floor", width, depth)
	room.Floor = tetra3d.NewModel("floor_model", floorMesh)
	room.Floor.SetLocalPosition(0, 0, 0)

	// Create ceiling
	ceilingMesh := NewPlaneMesh("ceiling", width, depth)
	room.Ceiling = tetra3d.NewModel("ceiling_model", ceilingMesh)
	room.Ceiling.SetLocalPosition(0, height, 0)
	// Flip ceiling to face down (rotate 180° around X axis)
	room.Ceiling.Rotate(1, 0, 0, float32(math.Pi))

	// Create walls - all should face INWARD toward the room center
	// Vertical plane mesh faces +Z by default

	// North wall at -Z edge, should face +Z (toward center)
	northMesh := NewVerticalPlaneMesh("north_wall", width, height)
	north := tetra3d.NewModel("north_wall_model", northMesh)
	north.SetLocalPosition(0, height/2, -depth/2)
	// Default +Z facing is correct (faces toward center)

	// South wall at +Z edge, should face -Z (toward center)
	southMesh := NewVerticalPlaneMesh("south_wall", width, height)
	south := tetra3d.NewModel("south_wall_model", southMesh)
	south.SetLocalPosition(0, height/2, depth/2)
	south.Rotate(0, 1, 0, float32(math.Pi)) // 180° to face inward

	// East wall at +X edge, should face -X (toward center)
	eastMesh := NewVerticalPlaneMesh("east_wall", depth, height)
	east := tetra3d.NewModel("east_wall_model", eastMesh)
	east.SetLocalPosition(width/2, height/2, 0)
	east.Rotate(0, 1, 0, float32(-math.Pi/2)) // -90° to face -X

	// West wall at -X edge, should face +X (toward center)
	westMesh := NewVerticalPlaneMesh("west_wall", depth, height)
	west := tetra3d.NewModel("west_wall_model", westMesh)
	west.SetLocalPosition(-width/2, height/2, 0)
	west.Rotate(0, 1, 0, float32(math.Pi/2)) // +90° to face +X

	room.Walls = []*tetra3d.Model{north, south, east, west}

	return room
}

// AddToScene adds all room components to a scene
func (r *Room) AddToScene(scene *Scene) {
	scene.Root().AddChildren(r.Floor)
	scene.Root().AddChildren(r.Ceiling)
	for _, wall := range r.Walls {
		scene.Root().AddChildren(wall)
	}
}

// SetFloorMaterial sets the material for the floor
func (r *Room) SetFloorMaterial(mat *tetra3d.Material) {
	if len(r.Floor.Mesh.MeshParts) > 0 {
		r.Floor.Mesh.MeshParts[0].Material = mat
	}
}

// SetCeilingMaterial sets the material for the ceiling
func (r *Room) SetCeilingMaterial(mat *tetra3d.Material) {
	if len(r.Ceiling.Mesh.MeshParts) > 0 {
		r.Ceiling.Mesh.MeshParts[0].Material = mat
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
