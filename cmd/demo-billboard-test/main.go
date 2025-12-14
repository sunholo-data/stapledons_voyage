// Minimal billboard flickering test
// Creates overlapping billboards at fixed positions to isolate the issue
package main

import (
	"flag"
	"fmt"
	"image/color"
	_ "image/jpeg" // Register JPEG decoder
	"log"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"sort"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"stapledons_voyage/engine/lod"
)

const (
	screenWidth  = 1024
	screenHeight = 768
)

type Game struct {
	camera            *lod.SimpleCamera
	billboardRenderer *lod.BillboardRenderer
	circleRenderer    *lod.CircleRenderer
	pointRenderer     *lod.PointRenderer
	objects           []*lod.Object
	sprites           map[string]*ebiten.Image
	frame             int

	// Debug: track sort order changes
	lastSortOrder []string
	sortChanges   int

	// Deferred texture processing (can't call texture.At before game starts)
	pendingTextures map[string]*ebiten.Image
	useTextures     bool

	// LOD Manager (optional, for testing LOD-specific flickering)
	lodManager *lod.Manager
	useLOD     bool

	// Overlap filtering (optional, for testing filterOverlappingBillboards)
	filterOverlap   bool
	filteredCount   int
	filterChanges   int
	lastFilteredIDs map[string]bool

	// Multi-layer rendering (optional, for testing circle/point underneath)
	useMultiLayer bool

	// Individual layer controls
	showPoints  bool
	showCircles bool
}

func NewGame(objectCount int, spread float64, useTextures bool, useLOD bool, filterOverlap bool, useMultiLayer bool, showPoints bool, showCircles bool) *Game {
	camera := lod.NewSimpleCamera(screenWidth, screenHeight)
	camera.Pos = lod.Vector3{X: 0, Y: 0, Z: -50} // Looking at origin from Z=-50
	camera.LookAt = lod.Vector3{X: 0, Y: 0, Z: 0}
	camera.Fov = 60

	billboardRenderer := lod.NewBillboardRenderer()
	circleRenderer := lod.NewCircleRenderer()
	pointRenderer := lod.NewPointRenderer()

	// Create LOD manager if requested
	var lodManager *lod.Manager
	if useLOD {
		config := lod.DefaultConfig()
		// Force all objects to billboard tier for this test
		config.Full3DPixels = 1000000      // Very high = never 3D
		config.BillboardPixels = 0         // Always billboard if visible
		config.TransitionTime = 0          // No transitions
		lodManager = lod.NewManager(config)
	}

	// Load planet textures if requested
	var planetTextures []*ebiten.Image
	if useTextures {
		// Use texture files that actually exist in assets/planets/
		textureFiles := []string{
			"earth_daymap.jpg", "mars.jpg", "jupiter.jpg", "saturn.jpg",
			"mercury.jpg", "neptune.jpg", "moon.jpg", "sun.jpg",
		}
		fmt.Printf("Loading textures from assets/planets/...\n")
		for _, name := range textureFiles {
			path := filepath.Join("assets", "planets", name)
			if _, err := os.Stat(path); err != nil {
				fmt.Printf("  MISSING: %s (%v)\n", path, err)
				continue
			}
			img, _, err := ebitenutil.NewImageFromFile(path)
			if err != nil {
				fmt.Printf("  FAILED to load %s: %v\n", name, err)
				continue
			}
			planetTextures = append(planetTextures, img)
			fmt.Printf("  Loaded: %s (%dx%d)\n", name, img.Bounds().Dx(), img.Bounds().Dy())
		}
		fmt.Printf("Loaded %d textures\n", len(planetTextures))
	}

	// Create overlapping objects at similar distances
	objects := make([]*lod.Object, objectCount)
	sprites := make(map[string]*ebiten.Image)

	colors := []color.RGBA{
		{255, 0, 0, 255},     // Red
		{0, 255, 0, 255},     // Green
		{0, 0, 255, 255},     // Blue
		{255, 255, 0, 255},   // Yellow
		{255, 0, 255, 255},   // Magenta
		{0, 255, 255, 255},   // Cyan
		{255, 128, 0, 255},   // Orange
		{128, 0, 255, 255},   // Purple
	}

	// Use fixed seed for reproducibility
	rng := rand.New(rand.NewSource(42))

	for i := 0; i < objectCount; i++ {
		id := fmt.Sprintf("obj_%d", i)
		col := colors[i%len(colors)]

		var x, y, z float64
		if objectCount <= 10 {
			// Small object count: use grid pattern for easy debugging
			x = float64(i%3-1) * 2.0  // -2, 0, 2
			y = float64(i/3%3-1) * 2.0 // -2, 0, 2
			z = float64(i) * spread    // Slightly different depths
		} else {
			// Large object count: random 3D positions like in LOD demo
			x = (rng.Float64() - 0.5) * spread * 100
			y = (rng.Float64() - 0.5) * spread * 100
			z = (rng.Float64() - 0.5) * spread * 100
		}

		obj := lod.NewObject(id, lod.Vector3{X: x, Y: y, Z: z}, 3.0, col)
		obj.Visible = true
		obj.CurrentTier = lod.TierBillboard
		objects[i] = obj

		// For non-texture mode, create simple colored sprites now
		if !useTextures {
			sprites[id] = lod.CreateDefaultPlanetSprite(64, col)
		}
	}

	// Store pending textures for deferred processing (can't call texture.At before game starts)
	pendingTextures := make(map[string]*ebiten.Image)
	if useTextures && len(planetTextures) > 0 {
		for i, obj := range objects {
			tex := planetTextures[i%len(planetTextures)]
			pendingTextures[obj.ID] = tex
		}
	}

	// Register objects with LOD manager if enabled
	if useLOD && lodManager != nil {
		for _, obj := range objects {
			lodManager.Add(obj)
		}
	}

	return &Game{
		camera:            camera,
		billboardRenderer: billboardRenderer,
		circleRenderer:    circleRenderer,
		pointRenderer:     pointRenderer,
		objects:           objects,
		sprites:           sprites,
		lastSortOrder:     make([]string, 0),
		pendingTextures:   pendingTextures,
		useTextures:       useTextures,
		lodManager:        lodManager,
		useLOD:            useLOD,
		filterOverlap:     filterOverlap,
		lastFilteredIDs:   make(map[string]bool),
		useMultiLayer:     useMultiLayer,
		showPoints:        showPoints,
		showCircles:       showCircles,
	}
}

// filterOverlappingBillboards removes billboards that overlap with each other.
// When billboards overlap, only the closest one is kept to prevent alpha blending artifacts.
func (g *Game) filterOverlappingBillboards(billboards []*lod.Object) []*lod.Object {
	if len(billboards) == 0 {
		return billboards
	}

	// Sort billboards by distance (closest first) with stable ordering
	sorted := make([]*lod.Object, len(billboards))
	copy(sorted, billboards)
	bucketSize := lod.DistanceBucketSize()
	sort.SliceStable(sorted, func(i, j int) bool {
		bucketI := int(sorted[i].Distance / bucketSize)
		bucketJ := int(sorted[j].Distance / bucketSize)
		if bucketI != bucketJ {
			return bucketI < bucketJ // Closer first
		}
		return sorted[i].ID < sorted[j].ID // Stable tiebreaker
	})

	// Track which billboards to keep (not overlapping with anything closer)
	keep := make([]bool, len(sorted))
	for i := range keep {
		keep[i] = true
	}

	// Check each billboard against closer billboards
	for i := 1; i < len(sorted); i++ {
		if !keep[i] {
			continue
		}
		bb := sorted[i]

		// Check against all closer billboards (indices 0 to i-1)
		for j := 0; j < i; j++ {
			if !keep[j] {
				continue
			}
			closer := sorted[j]

			dx := bb.ScreenX - closer.ScreenX
			dy := bb.ScreenY - closer.ScreenY
			dist := math.Sqrt(dx*dx + dy*dy)
			// Use 60% of combined radii - aggressive overlap detection
			minDist := (bb.ApparentRadius + closer.ApparentRadius) * 0.6

			if dist < minDist {
				// This billboard overlaps with a closer one - skip it
				keep[i] = false
				break
			}
		}
	}

	// Build result with only non-overlapping billboards
	filtered := make([]*lod.Object, 0, len(sorted))
	for i, bb := range sorted {
		if keep[i] {
			filtered = append(filtered, bb)
		}
	}
	return filtered
}

func (g *Game) Update() error {
	g.frame++

	// Process pending textures on first frame (deferred because texture.At can't be called before game starts)
	if len(g.pendingTextures) > 0 {
		fmt.Printf("Processing %d pending textures...\n", len(g.pendingTextures))
		for id, tex := range g.pendingTextures {
			g.sprites[id] = lod.CreateBillboardFromTexture(tex, 128)
			fmt.Printf("  Created billboard sprite for %s\n", id)
		}
		g.pendingTextures = nil // Clear after processing
		fmt.Printf("Texture processing complete\n")
	}

	// Camera movement with WASD/arrow keys
	speed := 0.5
	if ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyUp) {
		g.camera.Pos.Z += speed
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyDown) {
		g.camera.Pos.Z -= speed
	}
	if ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyLeft) {
		g.camera.Pos.X -= speed
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyRight) {
		g.camera.Pos.X += speed
	}
	if ebiten.IsKeyPressed(ebiten.KeyQ) {
		g.camera.Pos.Y -= speed
	}
	if ebiten.IsKeyPressed(ebiten.KeyE) {
		g.camera.Pos.Y += speed
	}

	// Update distances from camera
	if g.useLOD && g.lodManager != nil {
		// Use LOD Manager (this is what demo-lod does)
		g.lodManager.Update(g.camera)
	} else {
		// Direct calculation (simpler, no LOD manager)
		camPos := g.camera.Position()
		for _, obj := range g.objects {
			obj.Distance = obj.Position.Distance(camPos)

			// Project to screen
			sx, sy, visible := g.camera.WorldToScreen(obj.Position)
			obj.ScreenX = sx
			obj.ScreenY = sy
			obj.Visible = visible

			// Calculate apparent radius
			if obj.Distance > 0.001 {
				obj.ApparentRadius = (obj.Radius / obj.Distance) * g.camera.FOVScale()
			}
		}
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{20, 20, 40, 255})

	// Get objects to render
	var objectsToRender []*lod.Object
	if g.useLOD && g.lodManager != nil {
		// Use the tier list from LOD Manager (this is what demo-lod does)
		objectsToRender = g.lodManager.GetTierBillboard()
	} else {
		objectsToRender = g.objects
	}

	// Multi-layer rendering: render points and circles underneath billboards
	// This mimics demo-lod's layer compositing
	if g.useMultiLayer {
		// Layer 1: Points (furthest - simulated with small circles)
		if g.showPoints {
			g.pointRenderer.RenderPoints(screen, g.objects)
		}
		// Layer 2: Circles (medium distance)
		if g.showCircles {
			g.circleRenderer.RenderCircles(screen, g.objects)
		}
	}

	// Apply overlap filtering if enabled
	if g.filterOverlap {
		beforeCount := len(objectsToRender)
		objectsToRender = g.filterOverlappingBillboards(objectsToRender)
		g.filteredCount = beforeCount - len(objectsToRender)

		// Track filter stability - which objects are being filtered
		currentFiltered := make(map[string]bool)
		for _, obj := range objectsToRender {
			currentFiltered[obj.ID] = true
		}
		if len(g.lastFilteredIDs) > 0 {
			changed := false
			for id := range currentFiltered {
				if !g.lastFilteredIDs[id] {
					changed = true
					break
				}
			}
			if !changed {
				for id := range g.lastFilteredIDs {
					if !currentFiltered[id] {
						changed = true
						break
					}
				}
			}
			if changed {
				g.filterChanges++
			}
		}
		g.lastFilteredIDs = currentFiltered
	}

	// Render billboards
	g.billboardRenderer.RenderBillboards(screen, objectsToRender, g.sprites)

	// Track sort order changes for debugging
	currentOrder := make([]string, len(g.objects))
	for i, obj := range g.objects {
		currentOrder[i] = obj.ID
	}

	if len(g.lastSortOrder) > 0 {
		changed := false
		for i := range currentOrder {
			if i < len(g.lastSortOrder) && currentOrder[i] != g.lastSortOrder[i] {
				changed = true
				break
			}
		}
		if changed {
			g.sortChanges++
		}
	}
	g.lastSortOrder = currentOrder

	// Debug overlay
	info := fmt.Sprintf("Billboard Flicker Test\n"+
		"Objects: %d\n"+
		"Frame: %d\n"+
		"Sort order changes: %d\n"+
		"Camera: (%.1f, %.1f, %.1f)\n",
		len(g.objects), g.frame, g.sortChanges,
		g.camera.Pos.X, g.camera.Pos.Y, g.camera.Pos.Z)

	if g.filterOverlap {
		info += fmt.Sprintf("Filtered: %d  Filter changes: %d\n", g.filteredCount, g.filterChanges)
	}
	info += "\nObject distances:\n"

	for i, obj := range g.objects {
		if i < 10 { // Show first 10
			bucket := int(obj.Distance / lod.DistanceBucketSize())
			info += fmt.Sprintf("  %s: dist=%.4f bucket=%d\n", obj.ID, obj.Distance, bucket)
		}
	}

	ebitenutil.DebugPrint(screen, info)
}

func (g *Game) Layout(w, h int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	objectCount := flag.Int("objects", 8, "Number of overlapping objects")
	spread := flag.Float64("spread", 0.001, "Z-spread between objects (smaller = more overlap)")
	useTextures := flag.Bool("textures", false, "Use planet textures instead of solid colors")
	useLOD := flag.Bool("use-lod", false, "Use LOD Manager (to test if it causes flickering)")
	filterOverlap := flag.Bool("filter-overlap", false, "Apply filterOverlappingBillboards (to test if it causes flickering)")
	useMultiLayer := flag.Bool("multi-layer", false, "Render circles/points underneath billboards (like demo-lod)")
	showPoints := flag.Bool("points", true, "Show points layer (requires --multi-layer)")
	showCircles := flag.Bool("circles", true, "Show circles layer (requires --multi-layer)")
	flag.Parse()

	fmt.Printf("Billboard Flicker Test: %d objects, spread=%.4f, textures=%v, use-lod=%v, filter-overlap=%v, multi-layer=%v\n",
		*objectCount, *spread, *useTextures, *useLOD, *filterOverlap, *useMultiLayer)
	if *useMultiLayer {
		fmt.Printf("  Points: %v, Circles: %v\n", *showPoints, *showCircles)
	}
	fmt.Println("If flickering occurs, 'Sort order changes' or 'Filter changes' will increment")

	game := NewGame(*objectCount, *spread, *useTextures, *useLOD, *filterOverlap, *useMultiLayer, *showPoints, *showCircles)

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Billboard Flicker Test")
	ebiten.SetVsyncEnabled(true)

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
