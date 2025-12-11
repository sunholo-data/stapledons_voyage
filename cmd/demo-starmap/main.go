// cmd/demo-starmap/main.go
// Demo for the starmap data model - displays the local stellar neighborhood.
//
// Usage:
//   go build -o bin/demo-starmap ./cmd/demo-starmap && bin/demo-starmap
//   go run ./cmd/demo-starmap --debug
//   go run ./cmd/demo-starmap --screenshot 30
//
// Controls:
//   Arrow keys: Pan view
//   +/-: Zoom in/out
//   R: Reset view to Sol
//   ESC: Exit
package main

import (
	"errors"
	"flag"
	"fmt"
	"image/color"
	"image/png"
	"log"
	"math"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"stapledons_voyage/engine/display"
	"stapledons_voyage/engine/handlers"
	"stapledons_voyage/engine/stardata"
	"stapledons_voyage/sim_gen"
)

var (
	debug      = flag.Bool("debug", false, "Show debug information")
	seed       = flag.Int64("seed", 42, "World seed for initialization")
	screenshot = flag.Int("screenshot", 0, "Capture screenshot at frame N and exit (0=disabled)")
	output     = flag.String("output", "out/starmap.png", "Screenshot output path")
)

// Spectral type colors (RGBA)
var spectralColors = map[string]color.RGBA{
	"O": {155, 176, 255, 255}, // Blue-white
	"B": {170, 191, 255, 255}, // Blue-white
	"A": {202, 215, 255, 255}, // White
	"F": {248, 247, 255, 255}, // Yellow-white
	"G": {255, 244, 214, 255}, // Yellow (like Sol)
	"K": {255, 210, 161, 255}, // Orange
	"M": {255, 204, 111, 255}, // Red-orange
}

type DemoGame struct {
	// Star data
	catalog   *stardata.Catalog
	octree    *stardata.Octree
	ailCatalog *sim_gen.StarCatalog

	// View state
	viewX, viewY float64  // Center of view in light-years
	zoom         float64  // Pixels per light-year

	// Screenshot
	frameCount  int
	captured    bool
	capturedImg *ebiten.Image
}

func NewDemoGame() *DemoGame {
	// Initialize effect handlers for AILANG
	sim_gen.Init(sim_gen.Handlers{
		Debug: sim_gen.NewDebugContext(),
		Rand:  handlers.NewSeededRandHandler(*seed),
		Clock: handlers.NewEbitenClockHandler(),
		AI:    handlers.NewStubAIHandler(),
	})

	// Initialize AILANG star catalog
	ailCatalog := sim_gen.InitLocalCatalog()

	// Build Go catalog from AILANG data for octree queries
	goCatalog := stardata.NewCatalog()
	for _, star := range ailCatalog.Stars {
		goCatalog.AddStar(stardata.Star{
			ID:           star.Id,
			Name:         star.Name,
			X:            star.Pos.X,
			Y:            star.Pos.Y,
			Z:            star.Pos.Z,
			SpectralType: spectralTypeString(star.Spectral),
			Luminosity:   star.Luminosity,
			HasHZPlanet:  star.HasHZPlanet,
		})
	}

	// Build octree for fast queries
	octree := stardata.BuildOctree(goCatalog)
	stats := octree.GetStats()

	fmt.Printf("Loaded %d stars from AILANG catalog\n", len(ailCatalog.Stars))
	fmt.Printf("Octree: %d nodes, %d leaves, depth %d\n", stats.TotalNodes, stats.LeafNodes, stats.MaxDepth)

	return &DemoGame{
		catalog:    goCatalog,
		octree:     octree,
		ailCatalog: ailCatalog,
		viewX:      0, // Start centered on Sol
		viewY:      0,
		zoom:       30, // 30 pixels per light-year (shows ~10 ly radius)
	}
}

func spectralTypeString(spec *sim_gen.SpectralType) string {
	switch spec.Kind {
	case sim_gen.SpectralTypeKindO:
		return "O"
	case sim_gen.SpectralTypeKindB:
		return "B"
	case sim_gen.SpectralTypeKindA:
		return "A"
	case sim_gen.SpectralTypeKindF:
		return "F"
	case sim_gen.SpectralTypeKindG:
		return "G"
	case sim_gen.SpectralTypeKindK:
		return "K"
	case sim_gen.SpectralTypeKindM:
		return "M"
	default:
		return "G"
	}
}

func (g *DemoGame) Update() error {
	g.frameCount++

	// Pan with arrow keys
	panSpeed := 0.5 / g.zoom * 30 // Adjust for zoom level
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		g.viewX -= panSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		g.viewX += panSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
		g.viewY -= panSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
		g.viewY += panSpeed
	}

	// Zoom with +/-
	if ebiten.IsKeyPressed(ebiten.KeyEqual) || ebiten.IsKeyPressed(ebiten.KeyKPAdd) {
		g.zoom *= 1.02
		if g.zoom > 200 {
			g.zoom = 200
		}
	}
	if ebiten.IsKeyPressed(ebiten.KeyMinus) || ebiten.IsKeyPressed(ebiten.KeyKPSubtract) {
		g.zoom *= 0.98
		if g.zoom < 1 {
			g.zoom = 1
		}
	}

	// Reset view
	if ebiten.IsKeyPressed(ebiten.KeyR) {
		g.viewX = 0
		g.viewY = 0
		g.zoom = 30
	}

	// Exit
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		return errors.New("exit requested")
	}

	// Screenshot mode
	if *screenshot > 0 && g.captured {
		return errors.New("screenshot complete")
	}

	return nil
}

func (g *DemoGame) Draw(screen *ebiten.Image) {
	// Dark space background
	screen.Fill(color.RGBA{5, 5, 15, 255})

	screenW := float64(display.InternalWidth)
	screenH := float64(display.InternalHeight)
	centerX := screenW / 2
	centerY := screenH / 2

	// Calculate visible radius in light-years
	visibleRadius := math.Max(screenW, screenH) / g.zoom / 2 * 1.5

	// Query stars near view center
	stars := g.octree.Query(g.viewX, g.viewY, 0, visibleRadius)

	// Draw grid lines (every 5 light-years)
	g.drawGrid(screen, centerX, centerY)

	// Draw stars
	for _, star := range stars {
		// World to screen coordinates
		sx := centerX + (star.X-g.viewX)*g.zoom
		sy := centerY + (star.Y-g.viewY)*g.zoom

		// Skip if off screen
		if sx < -20 || sx > screenW+20 || sy < -20 || sy > screenH+20 {
			continue
		}

		// Star size based on luminosity
		radius := math.Log10(star.Luminosity+1)*2 + 2
		if radius < 2 {
			radius = 2
		}
		if radius > 10 {
			radius = 10
		}

		// Get color
		col := spectralColors[star.SpectralType]
		if star.SpectralType == "" {
			col = spectralColors["G"]
		}

		// Draw star glow
		g.drawStar(screen, sx, sy, radius, col)

		// Draw name for nearby bright stars
		dist := math.Sqrt(star.X*star.X + star.Y*star.Y + star.Z*star.Z)
		if dist < 15 && g.zoom > 15 {
			ebitenutil.DebugPrintAt(screen, star.Name, int(sx)+int(radius)+3, int(sy)-6)
		}
	}

	// Draw Sol marker (if visible)
	solSX := centerX + (0-g.viewX)*g.zoom
	solSY := centerY + (0-g.viewY)*g.zoom
	if solSX >= 0 && solSX < screenW && solSY >= 0 && solSY < screenH {
		// Draw crosshair for Sol
		crossCol := color.RGBA{255, 255, 100, 200}
		for i := -8.0; i <= 8.0; i++ {
			if i < -3 || i > 3 {
				screen.Set(int(solSX+i), int(solSY), crossCol)
				screen.Set(int(solSX), int(solSY+i), crossCol)
			}
		}
	}

	// Draw HUD
	g.drawHUD(screen, len(stars))

	// Capture screenshot
	if *screenshot > 0 && g.frameCount >= *screenshot && !g.captured {
		g.capturedImg = ebiten.NewImage(screen.Bounds().Dx(), screen.Bounds().Dy())
		g.capturedImg.DrawImage(screen, nil)
		g.captured = true
		if err := g.saveScreenshot(); err != nil {
			log.Printf("Failed to save screenshot: %v", err)
		} else {
			fmt.Printf("Screenshot saved to %s\n", *output)
		}
	}
}

func (g *DemoGame) drawGrid(screen *ebiten.Image, centerX, centerY float64) {
	gridCol := color.RGBA{30, 40, 60, 100}
	screenW := float64(display.InternalWidth)
	screenH := float64(display.InternalHeight)

	// Grid spacing (5 light-years)
	gridSpacing := 5.0
	if g.zoom < 5 {
		gridSpacing = 20.0
	} else if g.zoom < 15 {
		gridSpacing = 10.0
	}

	// Vertical lines
	startX := math.Floor((g.viewX-screenW/g.zoom/2)/gridSpacing) * gridSpacing
	for x := startX; x < g.viewX+screenW/g.zoom/2; x += gridSpacing {
		sx := centerX + (x-g.viewX)*g.zoom
		for y := 0.0; y < screenH; y += 2 {
			screen.Set(int(sx), int(y), gridCol)
		}
	}

	// Horizontal lines
	startY := math.Floor((g.viewY-screenH/g.zoom/2)/gridSpacing) * gridSpacing
	for y := startY; y < g.viewY+screenH/g.zoom/2; y += gridSpacing {
		sy := centerY + (y-g.viewY)*g.zoom
		for x := 0.0; x < screenW; x += 2 {
			screen.Set(int(x), int(sy), gridCol)
		}
	}
}

func (g *DemoGame) drawStar(screen *ebiten.Image, x, y, radius float64, col color.RGBA) {
	// Simple filled circle with glow
	r2 := radius * radius
	glowR2 := (radius * 1.5) * (radius * 1.5)

	for dy := -radius * 2; dy <= radius*2; dy++ {
		for dx := -radius * 2; dx <= radius*2; dx++ {
			d2 := dx*dx + dy*dy
			px, py := int(x+dx), int(y+dy)

			if d2 <= r2 {
				// Core
				screen.Set(px, py, col)
			} else if d2 <= glowR2 {
				// Glow (faded)
				alpha := uint8(float64(col.A) * (1 - d2/glowR2) * 0.5)
				glowCol := color.RGBA{col.R, col.G, col.B, alpha}
				screen.Set(px, py, glowCol)
			}
		}
	}
}

func (g *DemoGame) drawHUD(screen *ebiten.Image, visibleStars int) {
	// Background panel
	panelH := 100
	if *debug {
		panelH = 150
	}
	panelCol := color.RGBA{10, 15, 25, 200}
	for y := 0; y < panelH; y++ {
		for x := 0; x < 250; x++ {
			screen.Set(x, y, panelCol)
		}
	}

	y := 5
	ebitenutil.DebugPrintAt(screen, "STARMAP DEMO", 10, y)
	y += 16

	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("View: (%.1f, %.1f) ly", g.viewX, g.viewY), 10, y)
	y += 16

	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Zoom: %.1f px/ly", g.zoom), 10, y)
	y += 16

	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Visible stars: %d", visibleStars), 10, y)
	y += 16

	if *debug {
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Total in catalog: %d", g.catalog.Count()), 10, y)
		y += 16
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("FPS: %.1f", ebiten.ActualFPS()), 10, y)
		y += 16
	}

	// Controls help at bottom
	helpY := display.InternalHeight - 30
	ebitenutil.DebugPrintAt(screen, "Arrows: Pan | +/-: Zoom | R: Reset | ESC: Exit", 10, helpY)
}

func (g *DemoGame) saveScreenshot() error {
	f, err := os.Create(*output)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, g.capturedImg)
}

func (g *DemoGame) Layout(outsideWidth, outsideHeight int) (int, int) {
	return display.InternalWidth, display.InternalHeight
}

func main() {
	flag.Parse()

	fmt.Println("Starmap Demo")
	fmt.Println("============")
	fmt.Printf("Resolution: %dx%d\n", display.InternalWidth, display.InternalHeight)
	fmt.Printf("Seed: %d\n", *seed)
	fmt.Println()
	fmt.Println("Controls:")
	fmt.Println("  Arrow keys: Pan view")
	fmt.Println("  +/-: Zoom in/out")
	fmt.Println("  R: Reset view to Sol")
	fmt.Println("  ESC: Exit")
	fmt.Println()

	ebiten.SetWindowSize(display.InternalWidth*2, display.InternalHeight*2)
	ebiten.SetWindowTitle("Stapledon's Voyage - Starmap Demo")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	game := NewDemoGame()
	if err := ebiten.RunGame(game); err != nil && err.Error() != "exit requested" && err.Error() != "screenshot complete" {
		log.Fatal(err)
	}
}
