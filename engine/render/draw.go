package render

import (
	"image"
	"image/color"
	_ "image/jpeg"
	"log"
	"math"
	"os"
	"sort"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"stapledons_voyage/engine/assets"
	"stapledons_voyage/engine/camera"
	"stapledons_voyage/engine/depth"
	"stapledons_voyage/engine/tetra"
	"stapledons_voyage/sim_gen"
)

// Biome and structure colors for rendering (fallback when not using sprites)
var biomeColors = []color.RGBA{
	{0, 100, 200, 255},   // 0: Water (blue)
	{34, 139, 34, 255},   // 1: Forest (green)
	{210, 180, 140, 255}, // 2: Desert (tan)
	{139, 90, 43, 255},   // 3: Mountain (brown)
	{255, 255, 255, 180}, // 4: Selection highlight (white, semi-transparent)
	{139, 69, 19, 255},   // 5: House (saddle brown)
	{50, 205, 50, 255},   // 6: Farm (lime green)
	{128, 128, 128, 255}, // 7: Road (gray)
	{0, 0, 0, 255},       // 8: Reserved
	{0, 0, 0, 255},       // 9: Reserved
	{255, 0, 0, 255},     // 10: NPC 0 (red)
	{0, 255, 0, 255},     // 11: NPC 1 (green)
	{0, 0, 255, 255},     // 12: NPC 2 (blue)
	{255, 255, 0, 255},   // 13: NPC 3 (yellow)
	{255, 0, 255, 255},   // 14: NPC 4 (magenta)
}

// planet3DCache holds a reusable Tetra3D scene for rendering a single planet
type planet3DCache struct {
	scene   *tetra.Scene
	planet  *tetra.Planet
	ring    *tetra.Ring // Optional ring (for Saturn, Uranus)
	sun     *tetra.SunLight
	ambient *tetra.AmbientLight
}

// Renderer handles drawing FrameOutput to the screen.
type Renderer struct {
	assets         *assets.Manager
	anims          *AnimationManager
	lastTick       uint64        // Track simulation tick for animation updates
	galaxyBg       *ebiten.Image // Galaxy background image (loaded lazily)
	galaxyBgLoaded bool          // Whether we've attempted to load the background

	// Layer-aware rendering (parallax system)
	layers        *DepthLayerManager
	parallaxCam   *camera.ParallaxCamera
	layersEnabled bool // Whether to use layer-based rendering

	// Planet texture cache for TexturedPlanet DrawCmd
	planetTextures       map[string]*ebiten.Image
	planetTexturesLoaded bool

	// Tetra3D cached scenes for 3D planet rendering (one per planet for reuse)
	planet3DScenes map[string]*planet3DCache
}

// NewRenderer creates a renderer with the given asset manager.
func NewRenderer(assets *assets.Manager) *Renderer {
	r := &Renderer{
		assets: assets,
		anims:  NewAnimationManager(),
	}
	// Register animation definitions from asset manager
	if assets != nil {
		r.registerAnimations()
	}
	return r
}

// registerAnimations copies animation definitions from assets to the animation manager.
func (r *Renderer) registerAnimations() {
	// Get animation definitions from sprite manager
	// Check each sprite ID range that might have animations

	// World entities (100-105)
	for spriteID := 100; spriteID <= 105; spriteID++ {
		r.registerAnimationIfExists(spriteID)
	}

	// Bridge crew sprites (1200-1210)
	for spriteID := 1200; spriteID <= 1210; spriteID++ {
		r.registerAnimationIfExists(spriteID)
	}
}

// registerAnimationIfExists registers animation for a sprite if it has animations defined.
func (r *Renderer) registerAnimationIfExists(spriteID int) {
	animDef := r.assets.Sprites().GetAnimation(spriteID)
	if animDef != nil {
		r.anims.RegisterSprite(spriteID, &AnimationDef{
			Animations:  convertAnimations(animDef.Animations),
			FrameWidth:  animDef.FrameWidth,
			FrameHeight: animDef.FrameHeight,
		})
	}
}

// convertAnimations converts asset animation sequences to render animation sequences.
func convertAnimations(src map[string]assets.SpriteAnimSeq) map[string]AnimationSeq {
	dst := make(map[string]AnimationSeq)
	for name, seq := range src {
		dst[name] = AnimationSeq{
			StartFrame: seq.StartFrame,
			FrameCount: seq.FrameCount,
			FPS:        seq.FPS,
		}
	}
	return dst
}

// EnableLayers initializes and enables layer-based rendering with parallax support.
// Call this once after creating the renderer to enable the depth layer system.
func (r *Renderer) EnableLayers(screenW, screenH int) {
	r.layers = NewDepthLayerManager(screenW, screenH)
	r.parallaxCam = camera.NewParallaxCamera(screenW, screenH)
	r.layersEnabled = true
}

// ResizeLayers updates the layer buffers when screen size changes.
func (r *Renderer) ResizeLayers(screenW, screenH int) {
	if r.layers != nil {
		r.layers.Resize(screenW, screenH)
	}
	if r.parallaxCam != nil {
		r.parallaxCam.Resize(screenW, screenH)
	}
}

// getDepthLayer determines which depth layer a DrawCmd should render to.
// We support 20 layers (0-19) for flexible parallax composition:
//
//	Layer0 (0.00): Fixed at infinity (space/galaxy)
//	Layer1-5: Distant ship structures (0.05-0.25x)
//	Layer6-9: Mid-distance decks (0.30-0.60x)
//	Layer10-15: Near elements (0.70-0.95x)
//	Layer16-18: Scene layers (1.0x)
//	Layer19: UI (screen-fixed)
func getDepthLayer(cmd *sim_gen.DrawCmd) depth.Layer {
	switch cmd.Kind {
	// Marker: Use the parallaxLayer field to select layer (0-19)
	case sim_gen.DrawCmdKindMarker:
		layer := int(cmd.Marker.ParallaxLayer)
		if layer >= 0 && layer < int(depth.LayerCount) {
			return depth.Layer(layer)
		}
		return depth.Layer16 // Default to scene layer

	// Layer 0: Fixed at infinity (space, galaxy, stars - physically realistic)
	case sim_gen.DrawCmdKindGalaxyBg, sim_gen.DrawCmdKindSpaceBg, sim_gen.DrawCmdKindStar:
		return depth.Layer0

	// Layer 6: Mid-background (0.3x parallax) - spire, distant structures
	case sim_gen.DrawCmdKindSpireBg:
		return depth.Layer6

	// Layer 19: Foreground - UI elements (screen-fixed)
	case sim_gen.DrawCmdKindUi:
		return depth.Layer19

	// Layer 19: Screen-space elements
	case sim_gen.DrawCmdKindRectScreen, sim_gen.DrawCmdKindTextWrapped:
		return depth.Layer19

	// Layer 19: Text is typically UI/foreground
	case sim_gen.DrawCmdKindText:
		return depth.Layer19

	// Layer 16: Everything else is main scene layer
	default:
		return depth.Layer16
	}
}

// RenderFrame renders the FrameOutput to the Ebiten screen.
func (r *Renderer) RenderFrame(screen *ebiten.Image, out sim_gen.FrameOutput) {
	// Update animations (assume 60 FPS, ~16.67ms per frame)
	const dt = 1.0 / 60.0
	if r.anims != nil {
		r.anims.Update(dt)
	}

	// Get screen dimensions
	screenW, screenH := screen.Bounds().Dx(), screen.Bounds().Dy()

	// If layer rendering is enabled, use the layered path
	if r.layersEnabled && r.layers != nil {
		r.renderFrameLayered(screen, out, screenW, screenH)
		return
	}

	// Create camera transform (dereference pointer)
	cam := *out.Camera
	transform := camera.FromOutput(cam, screenW, screenH)

	// Calculate viewport for culling
	viewport := camera.CalculateViewport(cam, screenW, screenH)

	// Sort draw commands using isometric depth sorting
	// This handles both legacy Z-sorting and iso (layer, screenY) sorting
	sortables := make([]isoSortable, len(out.Draw))
	for i, cmd := range out.Draw {
		sortables[i] = isoSortable{
			cmd:     cmd,
			sortKey: getIsoSortKey(cmd, cam, screenW, screenH),
		}
	}
	sort.Slice(sortables, func(i, j int) bool {
		return sortables[i].sortKey < sortables[j].sortKey
	})

	// Render each command using Kind-based dispatch (discriminator struct pattern)
	for _, s := range sortables {
		cmd := s.cmd
		switch cmd.Kind {
		case sim_gen.DrawCmdKindRect:
			c := cmd.Rect
			// Cull if outside viewport
			if !viewport.ContainsRect(c.X, c.Y, c.W, c.H) {
				continue
			}
			// Transform to screen coordinates
			sx, sy := transform.WorldToScreen(c.X, c.Y)
			sw := c.W * transform.Scale
			sh := c.H * transform.Scale
			col := biomeColors[int(c.Color)%len(biomeColors)]
			ebitenutil.DrawRect(screen, sx, sy, sw, sh, col)

		case sim_gen.DrawCmdKindSprite:
			c := cmd.Sprite
			// Cull if outside viewport (assuming 16x16 sprite)
			if !viewport.ContainsRect(c.X, c.Y, 16, 16) {
				continue
			}
			r.drawSprite(screen, c, transform)

		case sim_gen.DrawCmdKindText:
			c := cmd.Text
			// Screen-space coordinates (no transform)
			r.drawText(screen, c, int(c.X), int(c.Y))

		case sim_gen.DrawCmdKindIsoTile:
			r.drawIsoTile(screen, cmd.IsoTile, cam, screenW, screenH)

		case sim_gen.DrawCmdKindIsoTileAlpha:
			r.drawIsoTileAlpha(screen, cmd.IsoTileAlpha, cam, screenW, screenH)

		case sim_gen.DrawCmdKindIsoEntity:
			r.drawIsoEntity(screen, cmd.IsoEntity, cam, screenW, screenH)

		case sim_gen.DrawCmdKindUi:
			r.drawUiElement(screen, cmd.Ui, screenW, screenH)

		case sim_gen.DrawCmdKindLine:
			r.drawLine(screen, cmd.Line)

		case sim_gen.DrawCmdKindTextWrapped:
			r.drawTextWrapped(screen, cmd.TextWrapped, screenW, screenH)

		case sim_gen.DrawCmdKindCircle:
			r.drawCircle(screen, cmd.Circle)

		case sim_gen.DrawCmdKindRectScreen:
			c := cmd.RectScreen
			// Screen-space rectangle (no camera transform)
			col := biomeColors[int(c.Color)%len(biomeColors)]
			ebitenutil.DrawRect(screen, c.X, c.Y, c.W, c.H, col)

		case sim_gen.DrawCmdKindGalaxyBg:
			c := cmd.GalaxyBg
			r.drawGalaxyBackground(screen, c.Opacity, screenW, screenH, c.SkyViewMode, c.ViewLon, c.ViewLat, c.Fov)

		case sim_gen.DrawCmdKindStar:
			r.drawStar(screen, cmd.Star)

		case sim_gen.DrawCmdKindSpireBg:
			r.drawSpireBg(screen, screenW, screenH)

		case sim_gen.DrawCmdKindRectRGBA:
			c := cmd.RectRGBA
			// Screen-space rectangle with packed RGBA color (0xRRGGBBAA format)
			col := unpackRGBA(c.Rgba)
			ebitenutil.DrawRect(screen, c.X, c.Y, c.W, c.H, col)

		case sim_gen.DrawCmdKindCircleRGBA:
			c := cmd.CircleRGBA
			// Screen-space circle with packed RGBA color
			col := unpackRGBA(c.Rgba)
			r.drawCircleRGBA(screen, c.X, c.Y, c.Radius, col, c.Filled)

		case sim_gen.DrawCmdKindTexturedPlanet:
			c := cmd.TexturedPlanet
			r.drawTexturedPlanet(screen, c.Name, c.X, c.Y, c.Radius, c.Rotation, c.HasRings, c.RingRgba)

		case sim_gen.DrawCmdKindMarker:
			// Marker in non-layered mode - render at screen position
			c := cmd.Marker
			col := unpackRGBA(c.Rgba)
			ebitenutil.DrawRect(screen, c.X, c.Y, c.W, c.H, col)
		}
	}

	// Render debug messages below UI panels (UI layer, not transformed)
	// Start at y=50 to avoid overlapping with camera panel
	for i, msg := range out.Debug {
		ebitenutil.DebugPrintAt(screen, msg, 10, 50+i*16)
	}
}

// renderFrameLayered renders using the depth layer system with parallax support.
// Commands are grouped by depth layer and rendered to separate buffers,
// then composited back-to-front for proper parallax and transparency effects.
func (r *Renderer) renderFrameLayered(screen *ebiten.Image, out sim_gen.FrameOutput, screenW, screenH int) {
	// Update parallax camera from game camera
	cam := *out.Camera
	r.parallaxCam.SetPosition(cam.X, cam.Y)
	r.parallaxCam.SetZoom(cam.Zoom)

	// Clear all layer buffers
	r.layers.Clear()

	// Group commands by depth layer
	layerCmds := make([][]*sim_gen.DrawCmd, LayerCount)
	for i := range layerCmds {
		layerCmds[i] = make([]*sim_gen.DrawCmd, 0)
	}

	for _, cmd := range out.Draw {
		layer := getDepthLayer(cmd)
		layerCmds[layer] = append(layerCmds[layer], cmd)
	}

	// Render each layer (back to front)
	for layerIdx := 0; layerIdx < int(LayerCount); layerIdx++ {
		layer := depth.Layer(layerIdx)
		buffer := r.layers.GetBuffer(layer)
		cmds := layerCmds[layerIdx]

		if len(cmds) == 0 {
			continue
		}

		// Get layer-specific camera transform (with parallax)
		transform := r.parallaxCam.TransformForLayer(layer)
		viewport := camera.CalculateViewport(cam, screenW, screenH)

		// Sort commands within this layer
		sortables := make([]isoSortable, len(cmds))
		for i, cmd := range cmds {
			sortables[i] = isoSortable{
				cmd:     cmd,
				sortKey: getIsoSortKey(cmd, cam, screenW, screenH),
			}
		}
		sort.Slice(sortables, func(i, j int) bool {
			return sortables[i].sortKey < sortables[j].sortKey
		})

		// Render commands to this layer's buffer
		r.renderCommandsToBuffer(buffer, sortables, transform, viewport, cam, screenW, screenH)
	}

	// Composite all layers to screen (back to front)
	r.layers.Composite(screen)

	// Render debug messages on top (directly to screen)
	for i, msg := range out.Debug {
		ebitenutil.DebugPrintAt(screen, msg, 10, 50+i*16)
	}
}

// renderCommandsToBuffer renders a list of sorted commands to a target buffer.
func (r *Renderer) renderCommandsToBuffer(
	buffer *ebiten.Image,
	sortables []isoSortable,
	transform camera.Transform,
	viewport camera.Viewport,
	cam sim_gen.Camera,
	screenW, screenH int,
) {
	for _, s := range sortables {
		cmd := s.cmd
		switch cmd.Kind {
		case sim_gen.DrawCmdKindRect:
			c := cmd.Rect
			if !viewport.ContainsRect(c.X, c.Y, c.W, c.H) {
				continue
			}
			sx, sy := transform.WorldToScreen(c.X, c.Y)
			sw := c.W * transform.Scale
			sh := c.H * transform.Scale
			col := biomeColors[int(c.Color)%len(biomeColors)]
			ebitenutil.DrawRect(buffer, sx, sy, sw, sh, col)

		case sim_gen.DrawCmdKindSprite:
			c := cmd.Sprite
			if !viewport.ContainsRect(c.X, c.Y, 16, 16) {
				continue
			}
			r.drawSprite(buffer, c, transform)

		case sim_gen.DrawCmdKindText:
			c := cmd.Text
			r.drawText(buffer, c, int(c.X), int(c.Y))

		case sim_gen.DrawCmdKindIsoTile:
			r.drawIsoTile(buffer, cmd.IsoTile, cam, screenW, screenH)

		case sim_gen.DrawCmdKindIsoTileAlpha:
			r.drawIsoTileAlpha(buffer, cmd.IsoTileAlpha, cam, screenW, screenH)

		case sim_gen.DrawCmdKindIsoEntity:
			r.drawIsoEntity(buffer, cmd.IsoEntity, cam, screenW, screenH)

		case sim_gen.DrawCmdKindUi:
			r.drawUiElement(buffer, cmd.Ui, screenW, screenH)

		case sim_gen.DrawCmdKindLine:
			r.drawLine(buffer, cmd.Line)

		case sim_gen.DrawCmdKindTextWrapped:
			r.drawTextWrapped(buffer, cmd.TextWrapped, screenW, screenH)

		case sim_gen.DrawCmdKindCircle:
			r.drawCircle(buffer, cmd.Circle)

		case sim_gen.DrawCmdKindRectScreen:
			c := cmd.RectScreen
			col := biomeColors[int(c.Color)%len(biomeColors)]
			ebitenutil.DrawRect(buffer, c.X, c.Y, c.W, c.H, col)

		case sim_gen.DrawCmdKindGalaxyBg:
			c := cmd.GalaxyBg
			r.drawGalaxyBackgroundParallax(buffer, c.Opacity, screenW, screenH, c.SkyViewMode, c.ViewLon, c.ViewLat, c.Fov, transform)

		case sim_gen.DrawCmdKindStar:
			r.drawStar(buffer, cmd.Star)

		case sim_gen.DrawCmdKindSpireBg:
			r.drawSpireBgParallax(buffer, screenW, screenH, transform)

		case sim_gen.DrawCmdKindRectRGBA:
			c := cmd.RectRGBA
			col := unpackRGBA(c.Rgba)
			ebitenutil.DrawRect(buffer, c.X, c.Y, c.W, c.H, col)

		case sim_gen.DrawCmdKindCircleRGBA:
			c := cmd.CircleRGBA
			col := unpackRGBA(c.Rgba)
			r.drawCircleRGBA(buffer, c.X, c.Y, c.Radius, col, c.Filled)

		case sim_gen.DrawCmdKindMarker:
			// Marker: rectangle at screen position with parallax applied via layer system
			c := cmd.Marker
			col := unpackRGBA(c.Rgba)
			// Apply parallax offset from transform
			x := c.X + transform.OffsetX - float64(screenW)/2
			y := c.Y + transform.OffsetY - float64(screenH)/2
			ebitenutil.DrawRect(buffer, x, y, c.W, c.H, col)
		}
	}
}

// getBridgeSpriteColor returns a fallback color for bridge sprite IDs
func getBridgeSpriteColor(id int64) color.RGBA {
	switch {
	// Bridge tiles (1000-1099) - BRIGHT colors for visibility
	case id == 1000: // tileFloor
		return color.RGBA{80, 90, 110, 255} // Brighter floor
	case id == 1001: // tileFloorGlow
		return color.RGBA{100, 120, 160, 255} // Glowing floor
	case id == 1002: // tileConsoleBase
		return color.RGBA{70, 80, 100, 255} // Console base
	case id == 1003: // tileWalkway
		return color.RGBA{110, 120, 140, 255} // Bright walkway
	case id == 1004: // tileDomeEdge
		return color.RGBA{60, 100, 140, 230} // Blue-tinted dome edge
	case id == 1005: // tileWall
		return color.RGBA{50, 60, 80, 255} // Wall (darker)
	case id == 1006: // tileHatch
		return color.RGBA{120, 100, 80, 255} // Warm hatch
	case id == 1007: // tileCaptainArea
		return color.RGBA{100, 90, 120, 255} // Purple-tinted captain area
	// Console sprites (1100-1149)
	case id >= 1100 && id < 1150:
		return color.RGBA{80, 120, 180, 255} // Blue consoles
	// Crew sprites (1200-1249)
	case id == 1200: // pilot
		return color.RGBA{180, 80, 80, 255} // Red
	case id == 1201: // comms
		return color.RGBA{80, 180, 80, 255} // Green
	case id == 1202: // engineer
		return color.RGBA{180, 180, 80, 255} // Yellow
	case id == 1203: // scientist
		return color.RGBA{80, 180, 180, 255} // Cyan
	case id == 1204: // captain
		return color.RGBA{180, 80, 180, 255} // Magenta
	case id == 1205: // player
		return color.RGBA{255, 255, 255, 255} // White
	case id >= 1200 && id < 1250:
		return color.RGBA{200, 150, 100, 255} // Generic crew tan
	default:
		return color.RGBA{128, 128, 128, 255} // Gray fallback
	}
}

// drawSprite draws a sprite using the asset manager with camera transform.
func (r *Renderer) drawSprite(screen *ebiten.Image, c *sim_gen.DrawCmdSprite, transform camera.Transform) {
	sx, sy := transform.WorldToScreen(c.X, c.Y)
	sw := 16.0 * transform.Scale
	sh := 16.0 * transform.Scale

	if r.assets == nil {
		// Fallback: draw colored placeholder based on sprite ID
		col := getBridgeSpriteColor(c.Id)
		ebitenutil.DrawRect(screen, sx, sy, sw, sh, col)
		return
	}

	sprite := r.assets.GetSprite(int(c.Id))
	if sprite == nil {
		// No sprite loaded - draw colored placeholder
		col := getBridgeSpriteColor(c.Id)
		ebitenutil.DrawRect(screen, sx, sy, sw, sh, col)
		return
	}

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(transform.Scale, transform.Scale)
	op.GeoM.Translate(sx, sy)
	screen.DrawImage(sprite, op)
}

// RenderFrame is a convenience function for backwards compatibility.
// Prefer using Renderer.RenderFrame for access to sprites.
func RenderFrame(screen *ebiten.Image, out sim_gen.FrameOutput) {
	r := &Renderer{assets: nil}
	r.RenderFrame(screen, out)
}

// unpackRGBA converts a packed RGBA int64 (0xRRGGBBAA format) to color.RGBA
func unpackRGBA(rgba int64) color.RGBA {
	return color.RGBA{
		R: uint8((rgba >> 24) & 0xFF),
		G: uint8((rgba >> 16) & 0xFF),
		B: uint8((rgba >> 8) & 0xFF),
		A: uint8(rgba & 0xFF),
	}
}

// drawCircleRGBA draws a circle with an RGBA color
func (r *Renderer) drawCircleRGBA(screen *ebiten.Image, x, y, radius float64, col color.RGBA, filled bool) {
	if filled {
		// Draw filled circle using pixel-by-pixel approach
		for dy := -radius; dy <= radius; dy++ {
			for dx := -radius; dx <= radius; dx++ {
				if dx*dx+dy*dy <= radius*radius {
					screen.Set(int(x+dx), int(y+dy), col)
				}
			}
		}
	} else {
		// Draw circle outline using midpoint algorithm
		cx, cy := int(x), int(y)
		r := int(radius)
		px, py := 0, r
		d := 1 - r
		for px <= py {
			screen.Set(cx+px, cy+py, col)
			screen.Set(cx-px, cy+py, col)
			screen.Set(cx+px, cy-py, col)
			screen.Set(cx-px, cy-py, col)
			screen.Set(cx+py, cy+px, col)
			screen.Set(cx-py, cy+px, col)
			screen.Set(cx+py, cy-px, col)
			screen.Set(cx-py, cy-px, col)
			if d < 0 {
				d += 2*px + 3
			} else {
				d += 2*(px-py) + 5
				py--
			}
			px++
		}
	}
}

// planetTexturePaths maps planet names to texture file paths
var planetTexturePaths = map[string]string{
	"mercury": "assets/planets/mercury.jpg",
	"venus":   "assets/planets/venus_atmosphere.jpg",
	"earth":   "assets/planets/earth_daymap.jpg",
	"mars":    "assets/planets/mars.jpg",
	"jupiter": "assets/planets/jupiter.jpg",
	"saturn":  "assets/planets/saturn.jpg",
	"uranus":  "assets/planets/uranus.jpg",
	"neptune": "assets/planets/neptune.jpg",
	"sun":     "assets/planets/sun.jpg",
}

// loadPlanetTextures loads all planet textures into the cache
func (r *Renderer) loadPlanetTextures() {
	if r.planetTexturesLoaded {
		return
	}
	r.planetTextures = make(map[string]*ebiten.Image)
	loaded := 0
	for name, path := range planetTexturePaths {
		f, err := os.Open(path)
		if err != nil {
			log.Printf("Failed to open texture %s: %v", path, err)
			continue // Skip missing textures
		}
		img, _, err := image.Decode(f)
		f.Close()
		if err != nil {
			log.Printf("Failed to decode texture %s: %v", path, err)
			continue
		}
		r.planetTextures[name] = ebiten.NewImageFromImage(img)
		loaded++
	}
	log.Printf("Loaded %d/%d planet textures", loaded, len(planetTexturePaths))
	r.planetTexturesLoaded = true
}

// drawTexturedPlanet draws a planet as a 3D sphere using Tetra3D
func (r *Renderer) drawTexturedPlanet(screen *ebiten.Image, name string, x, y, radius, rotation float64, hasRings bool, ringRgba int64) {
	// Ensure textures are loaded
	r.loadPlanetTextures()

	// Normalize name to lowercase for texture lookup (AILANG uses "Mercury", map uses "mercury")
	normalizedName := strings.ToLower(name)

	// Get or create cached 3D scene for this planet (includes rings if hasRings=true)
	cache := r.getOrCreatePlanet3D(normalizedName, hasRings)
	if cache == nil {
		// Fallback to 2D circle
		col := planetFallbackColor(normalizedName)
		r.drawCircleRGBA(screen, x, y, radius, col, true)
		return
	}

	// Update planet rotation
	cache.planet.SetRotation(rotation)

	// Update ring rotation if present (synced with planet)
	if cache.ring != nil {
		cache.ring.Update(0.016) // Approximate 60fps delta
	}

	// Render the 3D scene
	rendered := cache.scene.Render()
	if rendered == nil {
		col := planetFallbackColor(normalizedName)
		r.drawCircleRGBA(screen, x, y, radius, col, true)
		return
	}

	// Composite the 3D render onto the screen at the specified position
	// The scene is 512x512, planet fills ~75% of it (diameter ~384px)
	// We need to scale so that 384px becomes radius*2
	sceneSize := 512.0
	planetVisualDiameter := 384.0 // Approximate diameter in scene pixels
	scale := (radius * 2) / planetVisualDiameter

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(scale, scale)
	// Center the scaled image at x, y
	scaledSize := sceneSize * scale
	op.GeoM.Translate(x-scaledSize/2, y-scaledSize/2)
	screen.DrawImage(rendered, op)
}

// getOrCreatePlanet3D returns a cached Tetra3D scene for the given planet, creating it if needed.
// If hasRings is true, a 3D ring is added to the scene for ringed planets like Saturn.
func (r *Renderer) getOrCreatePlanet3D(name string, hasRings bool) *planet3DCache {
	// Initialize cache map if needed
	if r.planet3DScenes == nil {
		r.planet3DScenes = make(map[string]*planet3DCache)
	}

	// Cache key includes ring status since scene structure differs
	cacheKey := name
	if hasRings {
		cacheKey = name + "_ringed"
	}

	// Return existing cache if available
	if cache, ok := r.planet3DScenes[cacheKey]; ok {
		return cache
	}

	// Create new cached scene for this planet
	// Use larger scene for better 3D quality (like demo-solar-system)
	sceneSize := 512
	scene := tetra.NewScene(sceneSize, sceneSize)

	// Get texture
	tex := r.planetTextures[name]

	// Create planet - 1.5 radius matches demo-solar-system
	var planet *tetra.Planet
	if tex != nil {
		planet = tetra.NewTexturedPlanet(name, 1.5, tex)
		log.Printf("Created 3D textured planet: %s", name)
	} else {
		col := planetFallbackColor(name)
		planet = tetra.NewPlanet(name, 1.5, col)
		log.Printf("Created 3D solid planet: %s (no texture)", name)
	}

	// Add planet to scene at origin
	planet.AddToScene(scene)
	planet.SetPosition(0, 0, 0)

	// Add 3D rings if this is a ringed planet
	var ring *tetra.Ring
	if hasRings {
		// Ring dimensions relative to planet radius (1.5)
		// Saturn's rings extend from ~1.1 to ~2.3 planet radii
		innerRadius := 1.5 * 1.2 // Just outside the planet surface
		outerRadius := 1.5 * 2.3 // Classic Saturn ring proportions
		ring = tetra.NewRing(name, innerRadius, outerRadius, nil)
		ring.AddToScene(scene)
		ring.SetPosition(0, 0, 0)
		// Tilt rings ~27 degrees (Saturn's axial tilt) for visual interest
		ring.SetTilt(0.47) // ~27 degrees in radians
		log.Printf("Added 3D rings to planet: %s", name)
	}

	// Add lighting - same as demo-solar-system
	sun := tetra.NewSunLight()
	sun.SetPosition(5, 3, 15)
	sun.AddToScene(scene)

	ambient := tetra.NewAmbientLight(0.5, 0.5, 0.6, 0.8)
	ambient.AddToScene(scene)

	// Position camera - z=4 with 1.5 radius sphere fills ~75% of viewport
	// NOTE: Don't call LookAt - default camera direction works correctly
	scene.SetCameraPosition(0, 0, 4)

	cache := &planet3DCache{
		scene:   scene,
		planet:  planet,
		ring:    ring,
		sun:     sun,
		ambient: ambient,
	}
	r.planet3DScenes[cacheKey] = cache

	return cache
}

// planetFallbackColor returns a fallback color for planets without textures
func planetFallbackColor(name string) color.RGBA {
	switch name {
	case "sun":
		return color.RGBA{255, 200, 50, 255}
	case "mercury":
		return color.RGBA{180, 140, 100, 255}
	case "venus":
		return color.RGBA{200, 180, 140, 255}
	case "earth":
		return color.RGBA{60, 120, 200, 255}
	case "mars":
		return color.RGBA{200, 100, 80, 255}
	case "jupiter":
		return color.RGBA{220, 180, 140, 255}
	case "saturn":
		return color.RGBA{210, 190, 150, 255}
	case "uranus":
		return color.RGBA{150, 200, 220, 255}
	case "neptune":
		return color.RGBA{80, 120, 200, 255}
	default:
		return color.RGBA{128, 128, 128, 255}
	}
}

// drawTexturedCircle draws a texture cropped to a circle at the given position
func (r *Renderer) drawTexturedCircle(screen *ebiten.Image, tex *ebiten.Image, x, y, radius, rotation float64) {
	// Get texture dimensions
	texW := tex.Bounds().Dx()
	texH := tex.Bounds().Dy()

	// Scale texture to fit within the radius
	scale := (radius * 2) / float64(texW)
	if float64(texH)*scale > radius*2 {
		scale = (radius * 2) / float64(texH)
	}

	// Create draw options with rotation and positioning
	op := &ebiten.DrawImageOptions{}

	// Center the texture on origin for rotation
	op.GeoM.Translate(-float64(texW)/2, -float64(texH)/2)

	// Apply rotation
	op.GeoM.Rotate(rotation)

	// Scale
	op.GeoM.Scale(scale, scale)

	// Move to final position
	op.GeoM.Translate(x, y)

	// Draw the texture (for now, draw the full texture - circular masking would require shaders)
	// For orbital view, a square texture looks acceptable
	screen.DrawImage(tex, op)

	// Draw a circle border to make it look more planet-like
	r.drawCircleRGBA(screen, x, y, radius, color.RGBA{40, 40, 40, 100}, false)
}

// drawOrbitPath draws an orbital path circle
func (r *Renderer) drawOrbitPath(screen *ebiten.Image, centerX, centerY, radius float64, col color.RGBA) {
	// Draw circle outline using midpoint algorithm
	cx, cy := int(centerX), int(centerY)
	rad := int(radius)
	px, py := 0, rad
	d := 1 - rad
	for px <= py {
		screen.Set(cx+px, cy+py, col)
		screen.Set(cx-px, cy+py, col)
		screen.Set(cx+px, cy-py, col)
		screen.Set(cx-px, cy-py, col)
		screen.Set(cx+py, cy+px, col)
		screen.Set(cx-py, cy+px, col)
		screen.Set(cx+py, cy-px, col)
		screen.Set(cx-py, cy-px, col)
		if d < 0 {
			d += 2*px + 3
		} else {
			d += 2*(px-py) + 5
			py--
		}
		px++
	}
}

// Unused import placeholder to satisfy compiler during development
var _ = math.Pi

func getZ(cmd *sim_gen.DrawCmd) int {
	switch cmd.Kind {
	case sim_gen.DrawCmdKindRect:
		return int(cmd.Rect.Z)
	case sim_gen.DrawCmdKindSprite:
		return int(cmd.Sprite.Z)
	case sim_gen.DrawCmdKindText:
		return int(cmd.Text.Z)
	case sim_gen.DrawCmdKindLine:
		return int(cmd.Line.Z)
	case sim_gen.DrawCmdKindTextWrapped:
		return int(cmd.TextWrapped.Z)
	case sim_gen.DrawCmdKindCircle:
		return int(cmd.Circle.Z)
	case sim_gen.DrawCmdKindRectScreen:
		return int(cmd.RectScreen.Z)
	case sim_gen.DrawCmdKindGalaxyBg:
		return int(cmd.GalaxyBg.Z)
	case sim_gen.DrawCmdKindStar:
		return int(cmd.Star.Z)
	case sim_gen.DrawCmdKindIsoTile:
		return int(cmd.IsoTile.Layer)
	case sim_gen.DrawCmdKindIsoEntity:
		return int(cmd.IsoEntity.Layer)
	case sim_gen.DrawCmdKindUi:
		return int(cmd.Ui.Z) + 10000 // UI always on top
	case sim_gen.DrawCmdKindRectRGBA:
		return int(cmd.RectRGBA.Z)
	case sim_gen.DrawCmdKindCircleRGBA:
		return int(cmd.CircleRGBA.Z)
	case sim_gen.DrawCmdKindTexturedPlanet:
		return int(cmd.TexturedPlanet.Z)
	case sim_gen.DrawCmdKindMarker:
		return int(cmd.Marker.Z)
	}
	return 0
}
