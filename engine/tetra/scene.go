// Package tetra provides Tetra3D integration for 3D rendering.
package tetra

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/solarlune/tetra3d"
)

// Scene wraps a Tetra3D scene for rendering 3D content.
type Scene struct {
	library *tetra3d.Library
	scene   *tetra3d.Scene
	camera  *tetra3d.Camera
	buffer  *ebiten.Image
	width   int
	height  int
}

// NewScene creates a new 3D scene with the given dimensions.
func NewScene(width, height int) *Scene {
	s := &Scene{
		library: tetra3d.NewLibrary(),
		width:   width,
		height:  height,
	}

	s.scene = s.library.AddScene("main")
	s.buffer = ebiten.NewImage(width, height)

	// Setup camera
	s.camera = tetra3d.NewCamera(width, height)
	s.camera.SetFieldOfView(60) // degrees
	s.camera.SetNear(0.1)
	s.camera.SetFar(1000)
	s.scene.Root.AddChildren(s.camera)

	// Position camera back from origin
	s.camera.SetLocalPosition(0, 0, 5)

	return s
}

// Render renders the scene and returns the result as an ebiten.Image.
func (s *Scene) Render() *ebiten.Image {
	s.buffer.Clear()
	// Use transparent clear so starfield shows through
	s.camera.ClearWithColor(tetra3d.NewColor(0, 0, 0, 0))
	s.camera.RenderScene(s.scene)

	// Draw camera's color texture to buffer
	opt := &ebiten.DrawImageOptions{}
	s.buffer.DrawImage(s.camera.ColorTexture(), opt)

	return s.buffer
}

// Root returns the scene root node for adding objects.
func (s *Scene) Root() *tetra3d.Node {
	return s.scene.Root
}

// Camera returns the scene camera.
func (s *Scene) Camera() *tetra3d.Camera {
	return s.camera
}

// SetCameraPosition sets the camera position.
func (s *Scene) SetCameraPosition(x, y, z float64) {
	s.camera.SetLocalPosition(float32(x), float32(y), float32(z))
}

// RotateCamera rotates the camera by angle radians around the given axis.
func (s *Scene) RotateCamera(axisX, axisY, axisZ, angle float64) {
	s.camera.Rotate(float32(axisX), float32(axisY), float32(axisZ), float32(angle))
}

// Width returns the render width.
func (s *Scene) Width() int {
	return s.width
}

// Height returns the render height.
func (s *Scene) Height() int {
	return s.height
}

// Resize updates the scene for new dimensions.
func (s *Scene) Resize(width, height int) {
	if s.width == width && s.height == height {
		return
	}

	s.width = width
	s.height = height

	// Dispose old buffer
	if s.buffer != nil {
		s.buffer.Dispose()
	}
	s.buffer = ebiten.NewImage(width, height)

	// Recreate camera with new dimensions
	oldPos := s.camera.LocalPosition()
	s.camera = tetra3d.NewCamera(width, height)
	s.camera.SetFieldOfView(60)
	s.camera.SetNear(0.1)
	s.camera.SetFar(1000)
	s.camera.SetLocalPosition(oldPos.X, oldPos.Y, oldPos.Z)
	s.scene.Root.AddChildren(s.camera)
}
