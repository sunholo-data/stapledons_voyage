package view

import (
	"image/color"
	"log"
	"sort"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"stapledons_voyage/engine/assets"
	"stapledons_voyage/engine/display"
	"stapledons_voyage/engine/render"
	"stapledons_voyage/sim_gen"
)

// BridgeView is a view showing the ship's bridge interior in isometric style.
// The bridge features:
// - 16x12 isometric tile grid
// - Consoles at various stations (helm, comms, status, nav, science, captain)
// - Crew members at their stations
// - Player character that can move around
// - Dome viewport showing space (composite from SpaceView)
type BridgeView struct {
	state        *sim_gen.BridgeState
	renderer     *render.Renderer
	assets       *assets.Manager
	domeRenderer *DomeRenderer
	screenW      int
	screenH      int

	// State
	initialized bool
	frameCount  int64 // Frame counter for AILANG step function

	// UI elements
	uiPanels []*UIPanel

	// Cached static images (rendered once for performance)
	floorCache *ebiten.Image // Pre-rendered bridge floor, hull, spire
}

// NewBridgeView creates a new bridge view.
func NewBridgeView(assetMgr *assets.Manager) *BridgeView {
	return &BridgeView{
		screenW: display.InternalWidth,
		screenH: display.InternalHeight,
		assets:  assetMgr,
	}
}

// Type returns ViewBridge.
func (v *BridgeView) Type() ViewType {
	return ViewBridge
}

// Init initializes the view.
func (v *BridgeView) Init() error {
	if v.initialized {
		return nil
	}

	// Initialize bridge state from AILANG
	v.state = sim_gen.InitBridge()

	// Create renderer
	v.renderer = render.NewRenderer(v.assets)

	// Create dome renderer for observation viewport
	v.domeRenderer = NewDomeRenderer(DefaultDomeConfig())

	// Pre-render static bridge floor to cache (massive perf improvement)
	v.prerenderFloorCache()

	v.initialized = true
	return nil
}

// Enter is called when transitioning into this view.
func (v *BridgeView) Enter(from ViewType) {
	// Reset state if needed
	if v.state == nil {
		v.state = sim_gen.InitBridge()
	}
}

// Exit is called when transitioning out of this view.
func (v *BridgeView) Exit(to ViewType) {
	// Nothing to clean up
}

// Update updates the view state.
func (v *BridgeView) Update(dt float64) *ViewTransition {
	// Capture keyboard input for AILANG
	input := v.captureInput()

	// Process player input through AILANG
	if v.state != nil {
		v.state = sim_gen.ProcessBridgeInput(v.state, input)
	}

	// Step AILANG bridge state (crew movement, dome animation)
	if v.state != nil {
		v.state = sim_gen.StepBridge(v.state, v.frameCount)
		v.frameCount++
	}

	// Update dome renderer with camera position from AILANG state
	// AILANG owns the cruise animation (cameraZ, velocity)
	// Go renders the textured planets at that position
	if v.domeRenderer != nil {
		if v.state != nil && v.state.DomeState != nil {
			v.domeRenderer.SetCameraFromState(
				v.state.DomeState.CameraZ,
				v.state.DomeState.CruiseVelocity,
			)
		}
		v.domeRenderer.Update(dt)
	}

	return nil
}

// captureInput captures keyboard state and builds FrameInput for AILANG.
func (v *BridgeView) captureInput() *sim_gen.FrameInput {
	var keys []*sim_gen.KeyEvent

	// Check movement keys (WASD and arrows)
	movementKeys := []ebiten.Key{
		ebiten.KeyW, ebiten.KeyA, ebiten.KeyS, ebiten.KeyD,
		ebiten.KeyUp, ebiten.KeyDown, ebiten.KeyLeft, ebiten.KeyRight,
	}

	for _, key := range movementKeys {
		if inpututil.IsKeyJustPressed(key) {
			keys = append(keys, &sim_gen.KeyEvent{
				Key:  int64(key),
				Kind: "pressed",
			})
		}
	}

	// Check interaction keys (E for interact, Escape for cancel)
	if inpututil.IsKeyJustPressed(ebiten.KeyE) {
		keys = append(keys, &sim_gen.KeyEvent{
			Key:  int64(ebiten.KeyE),
			Kind: "pressed",
		})
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		keys = append(keys, &sim_gen.KeyEvent{
			Key:  int64(ebiten.KeyEscape),
			Kind: "pressed",
		})
	}

	return &sim_gen.FrameInput{
		Mouse:            &sim_gen.MouseState{X: 0, Y: 0, Buttons: []int64{}},
		Keys:             keys,
		ClickedThisFrame: false,
		WorldMouseX:      0,
		WorldMouseY:      0,
		TileMouseX:       0,
		TileMouseY:       0,
		ActionRequested:  &sim_gen.PlayerAction{Kind: sim_gen.PlayerActionKindActionNone},
		TestMode:         false,
	}
}

// Draw renders the view to the screen.
// AILANG controls the render order via Z values on DrawCmd variants.
// 3D commands (SpaceBg, Planets3D, BubbleArc) trigger domeRenderer calls.
func (v *BridgeView) Draw(screen *ebiten.Image) {
	if v.state == nil || v.renderer == nil {
		return
	}

	screenW := float64(screen.Bounds().Dx())
	screenH := float64(screen.Bounds().Dy())

	// Get draw commands from AILANG for bridge interior
	cmds := sim_gen.RenderBridge(v.state)

	// Update dome from AILANG state (velocity for SR effects)
	if v.domeRenderer != nil {
		v.domeRenderer.UpdateFromState(v.state.DomeView)
	}

	// Sort ALL commands by Z (AILANG controls render order)
	sort.Slice(cmds, func(i, j int) bool {
		return getCommandLayer(cmds[i]) < getCommandLayer(cmds[j])
	})

	// DEBUG: Print cmd count every 60 frames
	if v.frameCount%60 == 0 {
		log.Printf("AILANG: %d total cmds (sorted by Z)", len(cmds))
	}

	// Process commands in Z order - AILANG controls when 3D renders
	var batch []*sim_gen.DrawCmd
	for _, cmd := range cmds {
		switch cmd.Kind {
		case sim_gen.DrawCmdKindSpaceBg:
			// Flush any pending 2D commands first
			if len(batch) > 0 {
				v.renderBatch(screen, batch, screenW, screenH)
				batch = nil
			}
			// Render space background (starfield + galaxy)
			if v.domeRenderer != nil {
				v.domeRenderer.DrawBackground(screen)
			}

		case sim_gen.DrawCmdKindPlanets3D:
			// Flush any pending 2D commands first
			if len(batch) > 0 {
				v.renderBatch(screen, batch, screenW, screenH)
				batch = nil
			}
			// Render 3D textured planets (Tetra3D)
			if v.domeRenderer != nil {
				v.domeRenderer.DrawPlanets(screen)
			}

		case sim_gen.DrawCmdKindBubbleArc:
			// Flush any pending 2D commands first
			if len(batch) > 0 {
				v.renderBatch(screen, batch, screenW, screenH)
				batch = nil
			}
			// Render bubble arc edge effect
			if v.domeRenderer != nil {
				v.domeRenderer.DrawBubbleArc(screen)
			}

		default:
			// Batch 2D commands for efficient rendering
			batch = append(batch, cmd)
		}
	}

	// Flush any remaining 2D commands
	if len(batch) > 0 {
		v.renderBatch(screen, batch, screenW, screenH)
	}

	// UI panels on top of everything
	for _, panel := range v.uiPanels {
		if !panel.Visible {
			continue
		}
		bounds := ComputePanelBounds(panel, screenW, screenH)
		if panel.DrawFunc != nil {
			panel.DrawFunc(screen, bounds)
		}
	}
}

// renderBatch renders a batch of 2D draw commands.
func (v *BridgeView) renderBatch(screen *ebiten.Image, cmds []*sim_gen.DrawCmd, screenW, screenH float64) {
	out := sim_gen.FrameOutput{
		Draw:   cmds,
		Camera: &sim_gen.Camera{X: screenW / 2, Y: screenH / 2, Zoom: 1.0},
	}
	v.renderer.RenderFrame(screen, out)
}

// getCommandLayer extracts the layer/Z value from a DrawCmd.
// Used for sorting all commands by Z (AILANG controls render order).
func getCommandLayer(cmd *sim_gen.DrawCmd) int64 {
	if cmd == nil {
		return 0
	}
	switch cmd.Kind {
	case sim_gen.DrawCmdKindRectRGBA:
		return cmd.RectRGBA.Z
	case sim_gen.DrawCmdKindCircleRGBA:
		return cmd.CircleRGBA.Z
	case sim_gen.DrawCmdKindIsoTile:
		return cmd.IsoTile.Layer
	case sim_gen.DrawCmdKindIsoEntity:
		return cmd.IsoEntity.Layer
	case sim_gen.DrawCmdKindRect:
		return cmd.Rect.Z
	case sim_gen.DrawCmdKindText:
		return cmd.Text.Z
	case sim_gen.DrawCmdKindLine:
		return cmd.Line.Z
	case sim_gen.DrawCmdKindCircle:
		return cmd.Circle.Z
	case sim_gen.DrawCmdKindSpaceBg:
		return cmd.SpaceBg.Z
	case sim_gen.DrawCmdKindPlanets3D:
		return cmd.Planets3D.Z
	case sim_gen.DrawCmdKindBubbleArc:
		return cmd.BubbleArc.Z
	case sim_gen.DrawCmdKindGalaxyBg:
		return cmd.GalaxyBg.Z
	case sim_gen.DrawCmdKindTexturedPlanet:
		return cmd.TexturedPlanet.Z
	default:
		return 0 // Default to floor layer
	}
}

// Layers returns the view's layer components.
func (v *BridgeView) Layers() ViewLayers {
	return ViewLayers{
		Background: nil,
		Content:    nil,
		UI:         nil,
	}
}

// GetState returns the current bridge state.
func (v *BridgeView) GetState() *sim_gen.BridgeState {
	return v.state
}

// SetState sets the bridge state directly (for testing).
func (v *BridgeView) SetState(state *sim_gen.BridgeState) {
	v.state = state
}

// AddUIPanel adds a UI panel to the view.
func (v *BridgeView) AddUIPanel(panel *UIPanel) {
	v.uiPanels = append(v.uiPanels, panel)
}

// RemoveUIPanel removes a panel by ID.
func (v *BridgeView) RemoveUIPanel(id string) {
	for i, p := range v.uiPanels {
		if p.ID == id {
			v.uiPanels = append(v.uiPanels[:i], v.uiPanels[i+1:]...)
			return
		}
	}
}

// Resize updates the view for new screen dimensions.
func (v *BridgeView) Resize(screenW, screenH int) {
	v.screenW = screenW
	v.screenH = screenH
}

// GetCruiseInfo returns velocity and progress from the dome renderer.
func (v *BridgeView) GetCruiseInfo() (velocity float64, progress float64) {
	if v.domeRenderer != nil {
		return v.domeRenderer.GetCruiseInfo()
	}
	return 0, 0
}

// prerenderFloorCache renders the static bridge floor to an offscreen image.
// This is called once at startup and avoids expensive per-frame pixel operations.
func (v *BridgeView) prerenderFloorCache() {
	w := v.screenW
	h := v.screenH

	// Create cache image
	v.floorCache = ebiten.NewImage(w, h)

	// Disc center position
	discCenterX := float64(w) / 2
	discCenterY := float64(h) * 0.65

	// Disc dimensions
	discRadiusX := float64(w) * 0.58
	discRadiusY := float64(h) * 0.25

	// Draw the disc as a filled isometric ellipse
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dx := (float64(x) - discCenterX) / discRadiusX
			dy := (float64(y) - discCenterY) / discRadiusY
			dist := dx*dx + dy*dy

			if dist <= 1.0 {
				depthFactor := (float64(y) - (discCenterY - discRadiusY)) / (2 * discRadiusY)
				r := uint8(30 + depthFactor*20)
				g := uint8(35 + depthFactor*20)
				b := uint8(45 + depthFactor*25)

				if dist > 0.92 {
					edgeFade := (dist - 0.92) / 0.08
					r = uint8(float64(r) + edgeFade*30)
					g = uint8(float64(g) + edgeFade*40)
					b = uint8(float64(b) + edgeFade*50)
				}

				v.floorCache.Set(x, y, color.RGBA{r, g, b, 255})
			}
		}
	}

	// Draw isometric grid pattern on disc
	gridColor := color.RGBA{50, 55, 70, 255}
	tileSize := 48.0

	for i := -20; i <= 20; i++ {
		for t := 0.0; t <= 1.0; t += 0.002 {
			lineX := discCenterX + float64(i)*tileSize*0.5 + (t-0.5)*discRadiusX*2*0.7
			lineY := discCenterY + (t-0.5)*discRadiusY*2*0.7

			dx := (lineX - discCenterX) / discRadiusX
			dy := (lineY - discCenterY) / discRadiusY
			if dx*dx+dy*dy <= 0.95 {
				v.floorCache.Set(int(lineX), int(lineY), gridColor)
			}
		}

		for t := 0.0; t <= 1.0; t += 0.002 {
			lineX := discCenterX + float64(i)*tileSize*0.5 + (t-0.5)*discRadiusX*2*0.7
			lineY := discCenterY - (t-0.5)*discRadiusY*2*0.7

			dx := (lineX - discCenterX) / discRadiusX
			dy := (lineY - discCenterY) / discRadiusY
			if dx*dx+dy*dy <= 0.95 {
				v.floorCache.Set(int(lineX), int(lineY), gridColor)
			}
		}
	}

	// Draw disc edge outline
	edgeColor := color.RGBA{70, 85, 110, 255}
	for angle := 0.0; angle < 360.0; angle += 0.5 {
		rad := angle * 3.14159 / 180.0
		x := discCenterX + discRadiusX*1.0*cosApprox(rad)
		y := discCenterY + discRadiusY*1.0*sinApprox(rad)
		v.floorCache.Set(int(x), int(y), edgeColor)
		v.floorCache.Set(int(x)+1, int(y), edgeColor)
		v.floorCache.Set(int(x), int(y)+1, edgeColor)
	}

	// Draw hull below disc
	v.prerenderHullBelow(discCenterX, discCenterY, discRadiusX, discRadiusY)

	// Draw central spire
	v.prerenderCentralSpire(discCenterX, discCenterY)
}

// prerenderHullBelow renders the hull to the floor cache.
func (v *BridgeView) prerenderHullBelow(centerX, centerY, radiusX, radiusY float64) {
	h := v.screenH

	hullColor := color.RGBA{25, 30, 40, 255}
	hullEdge := color.RGBA{40, 50, 65, 255}

	for y := int(centerY + radiusY*0.3); y < h; y++ {
		distFromCenter := float64(y) - centerY
		progress := distFromCenter / (float64(h) - centerY)
		hullWidth := radiusX * (1.0 - progress*0.3)

		for x := int(centerX - hullWidth); x <= int(centerX+hullWidth); x++ {
			dx := (float64(x) - centerX) / radiusX
			dy := (float64(y) - centerY) / radiusY
			inDisc := dx*dx+dy*dy <= 1.0

			if !inDisc {
				edgeDist := (float64(x) - (centerX - hullWidth)) / (hullWidth * 2)
				if edgeDist < 0.05 || edgeDist > 0.95 {
					v.floorCache.Set(x, y, hullEdge)
				} else {
					darken := progress * 0.3
					r := uint8(float64(hullColor.R) * (1.0 - darken))
					g := uint8(float64(hullColor.G) * (1.0 - darken))
					b := uint8(float64(hullColor.B) * (1.0 - darken))
					v.floorCache.Set(x, y, color.RGBA{r, g, b, 255})
				}
			}
		}
	}
}

// prerenderCentralSpire renders the spire to the floor cache.
func (v *BridgeView) prerenderCentralSpire(centerX, centerY float64) {
	h := v.screenH

	spireWidthBase := 12.0
	spireHeightUp := 140.0
	spireHeightDown := float64(h) - centerY

	baseColor := color.RGBA{45, 55, 70, 255}
	highlightColor := color.RGBA{80, 95, 120, 255}
	lowerColor := color.RGBA{35, 42, 55, 255}

	// Upper spire
	for y := int(centerY); y > int(centerY-spireHeightUp); y-- {
		progress := (centerY - float64(y)) / spireHeightUp
		width := spireWidthBase * (1.0 - progress*0.7)

		for dx := -width / 2; dx <= width/2; dx++ {
			x := int(centerX + dx)
			r := uint8(float64(baseColor.R) + progress*float64(highlightColor.R-baseColor.R))
			g := uint8(float64(baseColor.G) + progress*float64(highlightColor.G-baseColor.G))
			b := uint8(float64(baseColor.B) + progress*float64(highlightColor.B-baseColor.B))

			if dx < -width/2+2 || dx > width/2-2 {
				r = uint8(min(255, int(r)+20))
				g = uint8(min(255, int(g)+25))
				b = uint8(min(255, int(b)+30))
			}

			v.floorCache.Set(x, y, color.RGBA{r, g, b, 255})
		}
	}

	// Spire tip
	tipY := int(centerY - spireHeightUp)
	tipColor := color.RGBA{100, 150, 200, 255}
	for dy := 0; dy < 15; dy++ {
		width := 5 - dy/3
		for dx := -width; dx <= width; dx++ {
			v.floorCache.Set(int(centerX)+dx, tipY-dy, tipColor)
		}
	}

	// Glowing tip beacon
	glowColor := color.RGBA{150, 200, 255, 200}
	v.floorCache.Set(int(centerX), tipY-15, glowColor)
	v.floorCache.Set(int(centerX)-1, tipY-14, glowColor)
	v.floorCache.Set(int(centerX)+1, tipY-14, glowColor)
	v.floorCache.Set(int(centerX), tipY-14, glowColor)

	// Lower spire
	for y := int(centerY); y < int(centerY+spireHeightDown); y++ {
		progress := (float64(y) - centerY) / spireHeightDown
		width := spireWidthBase * (1.0 + progress*0.2 - progress*progress*0.5)

		for dx := -width / 2; dx <= width/2; dx++ {
			x := int(centerX + dx)
			darken := progress * 0.4
			r := uint8(float64(lowerColor.R) * (1.0 - darken))
			g := uint8(float64(lowerColor.G) * (1.0 - darken))
			b := uint8(float64(lowerColor.B) * (1.0 - darken))

			if dx < -width/2+2 || dx > width/2-2 {
				r = uint8(min(255, int(r)+15))
				g = uint8(min(255, int(g)+18))
				b = uint8(min(255, int(b)+22))
			}

			v.floorCache.Set(x, y, color.RGBA{r, g, b, 255})
		}

		if int(float64(y)-centerY)%40 == 0 && progress > 0.1 {
			ringGlow := color.RGBA{60, 80, 120, 150}
			for dx := -width/2 - 3; dx <= width/2+3; dx++ {
				v.floorCache.Set(int(centerX+dx), y, ringGlow)
			}
		}
	}
}

// drawBridgeFloor draws the bridge floor from the pre-rendered cache.
// This is much faster than drawing pixel-by-pixel every frame.
func (v *BridgeView) drawBridgeFloor(screen *ebiten.Image) {
	if v.floorCache != nil {
		screen.DrawImage(v.floorCache, nil)
	}
}

// Simple sin/cos approximations for prerendering
func sinApprox(x float64) float64 {
	// Normalize to 0-2π
	for x < 0 {
		x += 6.28318
	}
	for x > 6.28318 {
		x -= 6.28318
	}
	// Taylor series approximation
	if x > 3.14159 {
		x -= 3.14159
		return -(x - x*x*x/6 + x*x*x*x*x/120)
	}
	return x - x*x*x/6 + x*x*x*x*x/120
}

func cosApprox(x float64) float64 {
	return sinApprox(x + 1.5708)
}
