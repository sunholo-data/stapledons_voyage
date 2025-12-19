package render

import (
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"math"
	"math/rand"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/solarlune/tetra3d"
	"stapledons_voyage/engine/tetra"
	"stapledons_voyage/sim_gen"
)

// shipState3D holds the current ship navigation state for window rendering.
// Set by ShipState3D command, used by Window3D commands.
type shipState3D struct {
	shipX, shipY, shipZ       float64 // Position in light-years
	forwardX, forwardY, forwardZ float64 // Forward direction
	upX, upY, upZ             float64 // Up direction
	velocity                  float64 // Fraction of c
	grPhi                     float64 // Gravitational potential
}

// interiorScene holds the 3D scene for interior rendering.
// Lazily initialized when first 3D command is received.
type interiorScene struct {
	scene           *tetra.Scene
	room            *tetra.Room
	props           map[string]*tetra3d.Model
	billboard       map[string]*tetra3d.Model
	windows         map[string]*tetra3d.Model  // 3D window planes
	windowTextures  map[string]*ebiten.Image   // Per-window star textures
	textures        map[string]*ebiten.Image   // Cached loaded textures
	starfieldTex    *ebiten.Image              // Pre-rendered starfield for windows (fallback)
	spaceView       *SpaceView                 // Real star catalog renderer
	shipState       *shipState3D               // Current ship navigation state
	lastVelocity    float64                    // Track velocity for texture updates
	lastGrPhi       float64                    // Track grPhi for texture updates
	customWindowTex *ebiten.Image              // Override window texture (for 3D planet views)
	needsInit       bool
	lastFloorTex    string // Track texture changes to avoid reloading
	lastWallTex     string
	lastCeilTex     string
	lastUvScale     float32 // Track UV scale changes
}

// loadTexture loads a texture from a file path, caching the result.
func (s *interiorScene) loadTexture(path string) *ebiten.Image {
	if path == "" {
		return nil
	}

	// Check cache first
	if tex, ok := s.textures[path]; ok {
		return tex
	}

	// Load from file
	f, err := os.Open(path)
	if err != nil {
		log.Printf("Warning: could not load texture %s: %v", path, err)
		return nil
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		log.Printf("Warning: could not decode texture %s: %v", path, err)
		return nil
	}

	tex := ebiten.NewImageFromImage(img)
	s.textures[path] = tex
	log.Printf("Loaded interior texture: %s", path)
	return tex
}

// getOrCreateStarfield returns a cached starfield texture, generating it if needed.
// The texture is 512x512 with procedurally generated stars.
func (s *interiorScene) getOrCreateStarfield() *ebiten.Image {
	if s.starfieldTex != nil {
		return s.starfieldTex
	}

	const texSize = 512
	s.starfieldTex = ebiten.NewImage(texSize, texSize)

	// Fill with deep space black
	s.starfieldTex.Fill(color.RGBA{5, 8, 15, 255})

	// Generate stars with deterministic seed
	rng := rand.New(rand.NewSource(42))

	// Layer 1: Many dim small stars
	for i := 0; i < 400; i++ {
		x := rng.Intn(texSize)
		y := rng.Intn(texSize)
		brightness := uint8(80 + rng.Intn(80))
		s.starfieldTex.Set(x, y, color.RGBA{brightness, brightness, brightness + 20, 255})
	}

	// Layer 2: Medium stars
	for i := 0; i < 150; i++ {
		x := rng.Intn(texSize)
		y := rng.Intn(texSize)
		brightness := uint8(150 + rng.Intn(80))
		// Core pixel
		s.starfieldTex.Set(x, y, color.RGBA{brightness, brightness, brightness + 30, 255})
		// Slight glow
		if x > 0 {
			s.starfieldTex.Set(x-1, y, color.RGBA{brightness / 2, brightness / 2, brightness/2 + 15, 255})
		}
		if x < texSize-1 {
			s.starfieldTex.Set(x+1, y, color.RGBA{brightness / 2, brightness / 2, brightness/2 + 15, 255})
		}
	}

	// Layer 3: Few bright stars with glow
	for i := 0; i < 30; i++ {
		x := rng.Intn(texSize - 4) + 2
		y := rng.Intn(texSize - 4) + 2
		// Bright core
		s.starfieldTex.Set(x, y, color.RGBA{255, 255, 255, 255})
		// Glow ring
		glow := color.RGBA{180, 180, 200, 255}
		s.starfieldTex.Set(x-1, y, glow)
		s.starfieldTex.Set(x+1, y, glow)
		s.starfieldTex.Set(x, y-1, glow)
		s.starfieldTex.Set(x, y+1, glow)
		// Dimmer outer glow
		dimGlow := color.RGBA{100, 100, 120, 255}
		s.starfieldTex.Set(x-1, y-1, dimGlow)
		s.starfieldTex.Set(x+1, y-1, dimGlow)
		s.starfieldTex.Set(x-1, y+1, dimGlow)
		s.starfieldTex.Set(x+1, y+1, dimGlow)
	}

	log.Println("Generated starfield texture for windows")
	return s.starfieldTex
}

// getInteriorScene returns the cached interior scene, creating it if needed.
func (r *Renderer) getInteriorScene(screenW, screenH int) *interiorScene {
	if r.interior == nil {
		r.interior = &interiorScene{
			props:     make(map[string]*tetra3d.Model),
			billboard: make(map[string]*tetra3d.Model),
			windows:   make(map[string]*tetra3d.Model),
			textures:  make(map[string]*ebiten.Image),
			needsInit: true,
		}

		// Initialize SpaceView with real star catalog
		r.interior.spaceView = NewSpaceView()
		if err := r.interior.spaceView.Load("assets/data/starmap/stars.json"); err != nil {
			log.Printf("Warning: could not load star catalog: %v (using procedural fallback)", err)
		} else {
			log.Printf("Loaded %d stars for window rendering", r.interior.spaceView.GetStarCount())
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

	// Check if room needs to be created or parameters changed
	needsRebuild := scene.room == nil ||
		float32(c.Width) != scene.room.Width ||
		float32(c.Depth) != scene.room.Depth ||
		float32(c.Height) != scene.room.Height ||
		scene.lastFloorTex != c.FloorTex ||
		scene.lastWallTex != c.WallTex ||
		scene.lastCeilTex != c.CeilingTex ||
		scene.lastUvScale != float32(c.UvScale)

	if needsRebuild {
		// Create new room with UV scale for texture tiling
		scene.room = tetra.NewRoomUV(float32(c.Width), float32(c.Depth), float32(c.Height), float32(c.UvScale))

		// Floor material
		floorMat := tetra3d.NewMaterial("floor")
		if c.FloorTex != "" {
			if tex := scene.loadTexture(c.FloorTex); tex != nil {
				floorMat.Texture = tex
				floorMat.UseTexture = true // Enable texture rendering
				floorMat.Color = tetra3d.NewColor(1, 1, 1, 1) // Full brightness for texture
			} else {
				floorMat.Color = unpackRGBAToTetra(int(c.FloorColor))
			}
		} else {
			floorMat.Color = unpackRGBAToTetra(int(c.FloorColor))
		}
		floorMat.Shadeless = true

		// Wall material
		wallMat := tetra3d.NewMaterial("wall")
		if c.WallTex != "" {
			if tex := scene.loadTexture(c.WallTex); tex != nil {
				wallMat.Texture = tex
				wallMat.UseTexture = true
				wallMat.Color = tetra3d.NewColor(1, 1, 1, 1)
			} else {
				wallMat.Color = unpackRGBAToTetra(int(c.WallColor))
			}
		} else {
			wallMat.Color = unpackRGBAToTetra(int(c.WallColor))
		}
		wallMat.Shadeless = true

		// Ceiling material
		ceilingMat := tetra3d.NewMaterial("ceiling")
		if c.CeilingTex != "" {
			if tex := scene.loadTexture(c.CeilingTex); tex != nil {
				ceilingMat.Texture = tex
				ceilingMat.UseTexture = true
				ceilingMat.Color = tetra3d.NewColor(1, 1, 1, 1)
			} else {
				ceilingMat.Color = unpackRGBAToTetra(int(c.CeilingColor))
			}
		} else {
			ceilingMat.Color = unpackRGBAToTetra(int(c.CeilingColor))
		}
		ceilingMat.Shadeless = true

		scene.room.SetFloorMaterial(floorMat)
		scene.room.SetWallMaterial(wallMat)
		scene.room.SetCeilingMaterial(ceilingMat)

		// Add room to scene
		scene.room.AddToScene(scene.scene)
		scene.needsInit = false

		// Track current textures and UV scale
		scene.lastFloorTex = c.FloorTex
		scene.lastWallTex = c.WallTex
		scene.lastCeilTex = c.CeilingTex
		scene.lastUvScale = float32(c.UvScale)
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

// handleShipState3D processes a ShipState3D command, caching ship navigation for window rendering.
func (r *Renderer) handleShipState3D(c *sim_gen.DrawCmdShipState3D, screenW, screenH int) {
	scene := r.getInteriorScene(screenW, screenH)

	// Cache ship navigation state for subsequent Window3D commands
	scene.shipState = &shipState3D{
		shipX:    c.ShipX,
		shipY:    c.ShipY,
		shipZ:    c.ShipZ,
		forwardX: c.ForwardX,
		forwardY: c.ForwardY,
		forwardZ: c.ForwardZ,
		upX:      c.UpX,
		upY:      c.UpY,
		upZ:      c.UpZ,
		velocity: c.Velocity,
		grPhi:    c.GrPhi,
	}
}

// SetWindowTexture sets a custom texture to use for all windows.
// This overrides the default starfield/star catalog rendering.
// Set to nil to return to default behavior.
func (r *Renderer) SetWindowTexture(tex *ebiten.Image) {
	if r.interior == nil {
		return
	}
	r.interior.customWindowTex = tex
}

// getWindowStarfieldTexture returns a starfield texture for the given window.
// Uses real star data from SpaceView if available, with view based on ship state and window normal.
func (r *Renderer) getWindowStarfieldTexture(scene *interiorScene, windowNx, windowNy, windowNz, windowW, windowH float64) *ebiten.Image {
	// Check for custom override texture first (e.g., 3D planet render)
	if scene.customWindowTex != nil {
		return scene.customWindowTex
	}

	// Calculate texture dimensions from window aspect ratio
	// Base resolution: 512 pixels per meter (high quality)
	const pixelsPerMeter = 512
	texW := int(windowW * pixelsPerMeter)
	texH := int(windowH * pixelsPerMeter)
	// Clamp to reasonable range (256 min, 2048 max)
	if texW < 256 {
		texW = 256
	}
	if texH < 256 {
		texH = 256
	}
	if texW > 2048 {
		texW = 2048
	}
	if texH > 2048 {
		texH = 2048
	}

	// If SpaceView has real stars and we have ship state, render real star view
	if scene.spaceView != nil && scene.spaceView.IsLoaded() && scene.shipState != nil {
		ship := scene.shipState

		// Calculate view direction: window normal points INTO room, so view is OPPOSITE
		viewDirX := -windowNx
		viewDirY := -windowNy
		viewDirZ := -windowNz

		// Transform view direction by ship orientation
		// For now, use ship forward/up directly (ship faces +Z initially)
		// TODO: Apply proper quaternion rotation based on ship orientation

		params := ViewParams{
			ShipX:    ship.shipX,
			ShipY:    ship.shipY,
			ShipZ:    ship.shipZ,
			ViewDirX: viewDirX * ship.forwardZ + viewDirZ * ship.forwardX,
			ViewDirY: viewDirY,
			ViewDirZ: viewDirZ * ship.forwardZ - viewDirX * ship.forwardX,
			UpX:      ship.upX,
			UpY:      ship.upY,
			UpZ:      ship.upZ,
			FOV:      90,
			Velocity: ship.velocity,
			GrPhi:    ship.grPhi,
		}

		// Render with correct aspect ratio based on window dimensions
		return scene.spaceView.RenderView(params, texW, texH)
	}

	// Fallback to procedural starfield
	return scene.getOrCreateStarfield()
}

// handleWindow3D processes a Window3D draw command.
// Renders a window plane that shows space (starfield/planets) through it.
// Window textures are regenerated when ship velocity changes significantly.
func (r *Renderer) handleWindow3D(c *sim_gen.DrawCmdWindow3D, screenW, screenH int) {
	scene := r.getInteriorScene(screenW, screenH)

	// Track windows in scene (using a map similar to props)
	if scene.windows == nil {
		scene.windows = make(map[string]*tetra3d.Model)
	}
	if scene.windowTextures == nil {
		scene.windowTextures = make(map[string]*ebiten.Image)
	}

	// Check if velocity or grPhi changed significantly (requires texture regeneration)
	currentVelocity := 0.0
	currentGrPhi := 0.0
	if scene.shipState != nil {
		currentVelocity = scene.shipState.velocity
		currentGrPhi = scene.shipState.grPhi
	}
	velocityChanged := math.Abs(currentVelocity-scene.lastVelocity) > 0.05
	grPhiChanged := math.Abs(currentGrPhi-scene.lastGrPhi) > 0.05

	// Check if we have a custom texture override (always needs update when custom)
	hasCustomTex := scene.customWindowTex != nil

	// Get or create window model
	win, exists := scene.windows[c.Id]

	// Check if we need to regenerate texture (new window, SR/GR params changed, or custom texture)
	needsTexture := c.ShowStars && (!exists || velocityChanged || grPhiChanged || hasCustomTex)
	var starfieldTex *ebiten.Image
	if needsTexture {
		starfieldTex = r.getWindowStarfieldTexture(scene, c.Nx, c.Ny, c.Nz, c.Width, c.Height)
		scene.windowTextures[c.Id] = starfieldTex
		if velocityChanged {
			scene.lastVelocity = currentVelocity
		}
		if grPhiChanged {
			scene.lastGrPhi = currentGrPhi
		}
	}

	if !exists {
		// Create a VERTICAL plane for the window (XY plane, faces +Z by default)
		mesh := tetra.NewVerticalPlaneMeshUV(c.Id+"_win", float32(c.Width), float32(c.Height), 1, 1)
		win = tetra3d.NewModel(c.Id+"_win_model", mesh)

		// Window material - set texture during creation if available
		mat := tetra3d.NewMaterial(c.Id + "_win_mat")
		mat.Shadeless = true
		mat.BackfaceCulling = false
		mat.TransparencyMode = tetra3d.TransparencyModeOpaque

		if starfieldTex != nil {
			// Set texture at creation time for proper rendering
			mat.Texture = starfieldTex
			mat.UseTexture = true
			mat.Color = tetra3d.NewColor(1, 1, 1, 1)
		} else if !c.ShowStars {
			// Set solid color for non-star windows
			switch c.WindowType {
			case 0: // Viewport
				mat.Color = tetra3d.NewColor(0.05, 0.08, 0.15, 1.0)
			case 1: // Porthole
				mat.Color = tetra3d.NewColor(0.08, 0.1, 0.2, 1.0)
			case 2: // Dome
				mat.Color = tetra3d.NewColor(0.02, 0.05, 0.12, 1.0)
			case 3: // Strip
				mat.Color = tetra3d.NewColor(0.05, 0.08, 0.12, 1.0)
			default:
				mat.Color = tetra3d.NewColor(0.05, 0.05, 0.1, 1.0)
			}
		}

		if len(win.Mesh.MeshParts) > 0 {
			win.Mesh.MeshParts[0].Material = mat
		}

		scene.scene.Root().AddChildren(win)
		scene.windows[c.Id] = win
	} else if needsTexture && starfieldTex != nil {
		// Existing window needs texture update (velocity changed)
		// Create new material with new texture (Tetra3D doesn't update well otherwise)
		mat := tetra3d.NewMaterial(c.Id + "_win_mat_v2")
		mat.Shadeless = true
		mat.BackfaceCulling = false
		mat.TransparencyMode = tetra3d.TransparencyModeOpaque
		mat.Texture = starfieldTex
		mat.UseTexture = true
		mat.Color = tetra3d.NewColor(1, 1, 1, 1)

		if len(win.Mesh.MeshParts) > 0 {
			win.Mesh.MeshParts[0].Material = mat
		}
	}

	// Update position
	win.SetLocalPosition(float32(c.X), float32(c.Y), float32(c.Z))

	// Rotate window to face the correct direction
	if c.Nz > 0.5 { // North wall - window faces +Z (into room)
		// Default orientation is correct
	} else if c.Nz < -0.5 { // South wall - window faces -Z
		win.SetLocalRotation(tetra3d.NewMatrix4Rotate(0, 1, 0, math.Pi))
	} else if c.Nx < -0.5 { // East wall - window faces -X
		win.SetLocalRotation(tetra3d.NewMatrix4Rotate(0, 1, 0, -math.Pi/2))
	} else if c.Nx > 0.5 { // West wall - window faces +X
		win.SetLocalRotation(tetra3d.NewMatrix4Rotate(0, 1, 0, math.Pi/2))
	} else if c.Ny < -0.9 { // Ceiling window - faces down
		win.SetLocalRotation(tetra3d.NewMatrix4Rotate(1, 0, 0, math.Pi/2))
	} else if c.Ny > 0.9 { // Floor window - faces up
		win.SetLocalRotation(tetra3d.NewMatrix4Rotate(1, 0, 0, -math.Pi/2))
	}
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
