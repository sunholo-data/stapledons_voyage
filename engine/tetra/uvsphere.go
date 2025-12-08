package tetra

import (
	"math"

	"github.com/solarlune/tetra3d"
)

// NewUVSphereMesh creates a sphere mesh with proper UV coordinates for
// equirectangular texture mapping (2:1 aspect ratio textures).
//
// segments: horizontal divisions (longitude)
// rings: vertical divisions (latitude)
//
// Higher values = smoother sphere but more vertices.
// Recommended: segments=64, rings=32 for good quality (higher for less pixelation)
func NewUVSphereMesh(segments, rings int) *tetra3d.Mesh {
	if segments < 4 {
		segments = 4
	}
	if rings < 2 {
		rings = 2
	}

	mesh := tetra3d.NewMesh("UVSphere")

	// Generate vertices
	vertices := make([]tetra3d.VertexInfo, 0, (rings+1)*(segments+1))

	for y := 0; y <= rings; y++ {
		// v goes from 0 (top) to 1 (bottom)
		v := float64(y) / float64(rings)
		// phi is latitude from 0 (north pole) to PI (south pole)
		phi := v * math.Pi

		for x := 0; x <= segments; x++ {
			// u goes from 0 to 1 around the sphere
			u := float64(x) / float64(segments)
			// theta is longitude from 0 to 2*PI
			theta := u * 2 * math.Pi

			// Spherical to Cartesian conversion
			// Note: We use Y-up convention
			sinPhi := math.Sin(phi)
			cosPhi := math.Cos(phi)
			sinTheta := math.Sin(theta)
			cosTheta := math.Cos(theta)

			px := float32(sinPhi * cosTheta)
			py := float32(cosPhi)
			pz := float32(sinPhi * sinTheta)

			// UV coordinates for equirectangular mapping
			// u wraps around horizontally
			// v goes from top (0) to bottom (1)
			uv_u := float32(u)
			uv_v := float32(v)

			vertices = append(vertices, tetra3d.VertexInfo{
				X: px,
				Y: py,
				Z: pz,
				U: uv_u,
				V: uv_v,
				// Normals point outward (same as position for unit sphere)
				NormalX: px,
				NormalY: py,
				NormalZ: pz,
			})
		}
	}

	mesh.AddVertices(vertices...)

	// Generate indices for triangles
	indices := make([]int, 0, rings*segments*6)

	for y := 0; y < rings; y++ {
		for x := 0; x < segments; x++ {
			// Current quad's corner indices
			topLeft := y*(segments+1) + x
			topRight := topLeft + 1
			bottomLeft := (y+1)*(segments+1) + x
			bottomRight := bottomLeft + 1

			// First triangle (top-left, bottom-left, bottom-right)
			indices = append(indices, topLeft, bottomLeft, bottomRight)

			// Second triangle (top-left, bottom-right, top-right)
			indices = append(indices, topLeft, bottomRight, topRight)
		}
	}

	// Create material and mesh part
	mat := tetra3d.NewMaterial("UVSphere")
	mesh.AddMeshPart(mat, indices...)

	mesh.UpdateBounds()

	return mesh
}
