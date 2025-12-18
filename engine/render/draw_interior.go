package render

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/solarlune/tetra3d"
	"stapledons_voyage/engine/tetra"
	"stapledons_voyage/sim_gen"
)

// interiorScene holds the 3D scene for interior rendering.
// Lazily initialized when first 3D command is received.
type interiorScene struct {
	scene     *tetra.Scene
	room      *tetra.Room
	props     map[string]*tetra3d.Model
	billboard map[string]*tetra3d.Model
	needsInit bool
}

// getInteriorScene returns the cached interior scene, creating it if needed.
func (r *Renderer) getInteriorScene(screenW, screenH int) *interiorScene {
	if r.interior == nil {
		r.interior = &interiorScene{
			props:     make(map[string]*tetra3d.Model),
			billboard: make(map[string]*tetra3d.Model),
			needsInit: true,
		}
	}
	if r.interior.scene == nil {
		r.interior.scene = tetra.NewScene(screenW, screenH)
		r.interior.scene.SetFieldOfView(70)
		r.interior.scene.SetNear(0.1)
		r.interior.scene.SetFar(100)
		r.interior.scene.SetLightingEnabled(false) // Shadeless for now
	}
	return r.interior
}

// handleCamera3D processes a Camera3D draw command.
func (r *Renderer) handleCamera3D(c *sim_gen.DrawCmdCamera3D, screenW, screenH int) {
	scene := r.getInteriorScene(screenW, screenH)

	// Set camera position
	scene.scene.SetCameraPosition(c.X, c.Y, c.Z)

	// Set field of view
	scene.scene.SetFieldOfView(c.Fov)

	// Calculate look direction from yaw and pitch
	lookX := c.X + math.Sin(c.Yaw)*math.Cos(c.Pitch)
	lookY := c.Y + math.Sin(c.Pitch)
	lookZ := c.Z - math.Cos(c.Yaw)*math.Cos(c.Pitch)
	scene.scene.LookAt(lookX, lookY, lookZ)
}

// handleRoom3D processes a Room3D draw command.
func (r *Renderer) handleRoom3D(c *sim_gen.DrawCmdRoom3D, screenW, screenH int) {
	scene := r.getInteriorScene(screenW, screenH)

	// Create or update room if dimensions changed
	if scene.room == nil ||
		float32(c.Width) != scene.room.Width ||
		float32(c.Depth) != scene.room.Depth ||
		float32(c.Height) != scene.room.Height {

		// Remove old room if exists
		if scene.room != nil {
			// Note: Tetra3D doesn't have RemoveChildren, we'd need to recreate the scene
			// For now, just update materials
		}

		// Create new room
		scene.room = tetra.NewRoom(float32(c.Width), float32(c.Depth), float32(c.Height))

		// Set materials
		floorMat := tetra3d.NewMaterial("floor")
		floorMat.Color = unpackRGBAToTetra(int(c.FloorColor))
		floorMat.Shadeless = true

		wallMat := tetra3d.NewMaterial("wall")
		wallMat.Color = unpackRGBAToTetra(int(c.WallColor))
		wallMat.Shadeless = true

		ceilingMat := tetra3d.NewMaterial("ceiling")
		ceilingMat.Color = unpackRGBAToTetra(int(c.CeilingColor))
		ceilingMat.Shadeless = true

		scene.room.SetFloorMaterial(floorMat)
		scene.room.SetWallMaterial(wallMat)
		scene.room.SetCeilingMaterial(ceilingMat)

		// Add room to scene
		scene.room.AddToScene(scene.scene)
		scene.needsInit = false
	}
}

// handleProp3D processes a Prop3D draw command.
func (r *Renderer) handleProp3D(c *sim_gen.DrawCmdProp3D, screenW, screenH int) {
	scene := r.getInteriorScene(screenW, screenH)

	// Get or create prop model
	prop, exists := scene.props[c.Id]
	if !exists {
		// Create cube mesh for prop
		mesh := tetra.NewCubeMesh(c.Id, 1.0)
		prop = tetra3d.NewModel(c.Id+"_model", mesh)

		// Set material
		mat := tetra3d.NewMaterial(c.Id + "_mat")
		mat.Color = unpackRGBAToTetra(int(c.Color))
		mat.Shadeless = true
		if len(prop.Mesh.MeshParts) > 0 {
			prop.Mesh.MeshParts[0].Material = mat
		}

		scene.scene.Root().AddChildren(prop)
		scene.props[c.Id] = prop
	}

	// Update position and scale
	prop.SetLocalPosition(float32(c.X), float32(c.Y), float32(c.Z))
	prop.SetLocalScale(float32(c.ScaleX), float32(c.ScaleY), float32(c.ScaleZ))
}

// handleBillboard3D processes a Billboard3D draw command.
// For now, renders as a colored cube at the position (sprite rendering TBD).
func (r *Renderer) handleBillboard3D(c *sim_gen.DrawCmdBillboard3D, screenW, screenH int) {
	scene := r.getInteriorScene(screenW, screenH)

	// Get or create billboard model
	bb, exists := scene.billboard[c.Id]
	if !exists {
		// Create thin cube as placeholder for sprite
		mesh := tetra.NewCubeMesh(c.Id+"_bb", 0.5)
		bb = tetra3d.NewModel(c.Id+"_bb_model", mesh)

		// Use a distinctive color for characters
		mat := tetra3d.NewMaterial(c.Id + "_bb_mat")
		mat.Color = tetra3d.NewColor(0.8, 0.6, 0.4, 1.0) // Tan/skin color
		mat.Shadeless = true
		if len(bb.Mesh.MeshParts) > 0 {
			bb.Mesh.MeshParts[0].Material = mat
		}

		scene.scene.Root().AddChildren(bb)
		scene.billboard[c.Id] = bb
	}

	// Update position (billboard at eye level)
	bb.SetLocalPosition(float32(c.X), float32(c.Y)+0.9, float32(c.Z))
	bb.SetLocalScale(float32(c.Scale)*0.4, float32(c.Scale)*1.8, float32(c.Scale)*0.2)
}

// renderInterior3D renders the 3D interior scene to the screen.
func (r *Renderer) renderInterior3D(screen *ebiten.Image, screenW, screenH int) {
	if r.interior == nil || r.interior.scene == nil {
		return
	}

	// Render 3D scene
	img3d := r.interior.scene.Render()
	screen.DrawImage(img3d, nil)
}

// unpackRGBAToTetra converts packed RGBA int to Tetra3D color.
func unpackRGBAToTetra(rgba int) tetra3d.Color {
	r := float32((rgba>>24)&0xFF) / 255.0
	g := float32((rgba>>16)&0xFF) / 255.0
	b := float32((rgba>>8)&0xFF) / 255.0
	a := float32(rgba&0xFF) / 255.0
	return tetra3d.NewColor(r, g, b, a)
}
