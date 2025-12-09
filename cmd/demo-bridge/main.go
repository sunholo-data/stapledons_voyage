// cmd/demo-bridge/main.go
// Demo command for testing the bridge interior view.
// Usage:
//   go run ./cmd/demo-bridge                    # Basic bridge view demo
//   go run ./cmd/demo-bridge --debug            # Show debug info
//   go run ./cmd/demo-bridge --screenshot 30   # Capture frame 30 to out/bridge.png
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
	"stapledons_voyage/engine/assets"
	"stapledons_voyage/engine/display"
	"stapledons_voyage/engine/handlers"
	"stapledons_voyage/engine/view"
	"stapledons_voyage/sim_gen"
)

var (
	debug      = flag.Bool("debug", false, "Show debug information")
	seed       = flag.Int64("seed", 42, "World seed for initialization")
	screenshot = flag.Int("screenshot", 0, "Capture screenshot at frame N and exit (0=disabled)")
	output     = flag.String("output", "out/bridge.png", "Screenshot output path")
)

type DemoGame struct {
	bridgeView  *view.BridgeView
	assets      *assets.Manager
	frameCount  int
	captured    bool
	capturedImg *ebiten.Image
	time        float64

	// Camera for zoomed exploration
	cameraX float64 // Camera world position
	cameraY float64
	zoom    float64 // Zoom level (1.0 = default, 2.0 = 2x zoom)
}

func NewDemoGame() *DemoGame {
	// Initialize effect handlers
	sim_gen.Init(sim_gen.Handlers{
		Debug: sim_gen.NewDebugContext(),
		Rand:  handlers.NewSeededRandHandler(*seed),
		Clock: handlers.NewEbitenClockHandler(),
		AI:    handlers.NewStubAIHandler(),
	})

	// Create asset manager
	assetMgr, err := assets.NewManager("assets")
	if err != nil {
		log.Printf("Warning: failed to initialize assets: %v", err)
	}

	// Create bridge view
	bridgeView := view.NewBridgeView(assetMgr)
	if err := bridgeView.Init(); err != nil {
		log.Printf("Warning: failed to initialize bridge view: %v", err)
	}

	return &DemoGame{
		bridgeView: bridgeView,
		assets:     assetMgr,
	}
}

func (g *DemoGame) Update() error {
	dt := 1.0 / 60.0
	g.frameCount++

	// Update bridge view
	if g.bridgeView != nil {
		g.bridgeView.Update(dt)
	}

	// Screenshot mode: exit after capture
	if *screenshot > 0 && g.captured {
		return errors.New("screenshot complete")
	}

	return nil
}

func (g *DemoGame) Draw(screen *ebiten.Image) {
	// Draw bridge view
	if g.bridgeView != nil {
		g.bridgeView.Draw(screen)
	}

	// Draw HUD (skip for clean screenshots)
	if *screenshot == 0 || !g.captured {
		g.drawHUD(screen)
	}

	// Capture screenshot at target frame
	if *screenshot > 0 && g.frameCount >= *screenshot && !g.captured {
		g.capturedImg = ebiten.NewImage(screen.Bounds().Dx(), screen.Bounds().Dy())
		g.capturedImg.DrawImage(screen, nil)
		g.captured = true

		// Save to file
		if err := g.saveScreenshot(); err != nil {
			log.Printf("Failed to save screenshot: %v", err)
		} else {
			fmt.Printf("Screenshot saved to %s\n", *output)
		}
	}
}

func (g *DemoGame) saveScreenshot() error {
	f, err := os.Create(*output)
	if err != nil {
		return err
	}
	defer f.Close()

	return png.Encode(f, g.capturedImg)
}

func (g *DemoGame) drawHUD(screen *ebiten.Image) {
	y := 10.0
	lineHeight := 16.0

	// Title
	ebitenutil.DebugPrintAt(screen, "Bridge Interior Demo", 10, int(y))
	y += lineHeight

	// FPS
	fps := ebiten.ActualFPS()
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("FPS: %.1f", fps), 10, int(y))
	y += lineHeight

	// Frame count
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Frame: %d", g.frameCount), 10, int(y))
	y += lineHeight

	if *debug && g.bridgeView != nil {
		state := g.bridgeView.GetState()
		if state != nil {
			y += lineHeight
			ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Bridge: %dx%d tiles", state.Width, state.Height), 10, int(y))
			y += lineHeight
			ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Player: (%d, %d)", state.PlayerPos.X, state.PlayerPos.Y), 10, int(y))
			y += lineHeight
			ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Crew: %d", len(state.CrewPositions)), 10, int(y))
			y += lineHeight
			ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Consoles: %d", len(state.Consoles)), 10, int(y))
		}
	}

	// Draw velocity HUD in top-right corner
	g.drawVelocityHUD(screen)

	// Help at bottom
	y = float64(display.InternalHeight) - 30
	ebitenutil.DebugPrintAt(screen, "Bubble Ship - Cruising through Solar System", 10, int(y))
	y += lineHeight
	if !*debug {
		ebitenutil.DebugPrintAt(screen, "Run with --debug for more info", 10, int(y))
	}
}

// drawVelocityHUD draws velocity information in top-right corner.
func (g *DemoGame) drawVelocityHUD(screen *ebiten.Image) {
	if g.bridgeView == nil {
		return
	}

	velocity, progress := g.bridgeView.GetCruiseInfo()

	// Position in top-right
	hudX := display.InternalWidth - 220
	hudY := 10

	// Background panel
	panelColor := color.RGBA{20, 25, 35, 200}
	for py := hudY; py < hudY+90; py++ {
		for px := hudX; px < hudX+210; px++ {
			screen.Set(px, py, panelColor)
		}
	}

	// Border
	borderColor := color.RGBA{60, 80, 120, 255}
	for px := hudX; px < hudX+210; px++ {
		screen.Set(px, hudY, borderColor)
		screen.Set(px, hudY+89, borderColor)
	}
	for py := hudY; py < hudY+90; py++ {
		screen.Set(hudX, py, borderColor)
		screen.Set(hudX+209, py, borderColor)
	}

	// Velocity text
	ebitenutil.DebugPrintAt(screen, "VELOCITY", hudX+10, hudY+5)

	// Velocity value with c fraction
	velocityStr := fmt.Sprintf("%.0f%% c", velocity*100)
	ebitenutil.DebugPrintAt(screen, velocityStr, hudX+100, hudY+5)

	// Lorentz factor (time dilation)
	gamma := 1.0 / math.Sqrt(1-velocity*velocity+0.0001)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Time dilation: %.2fx", gamma), hudX+10, hudY+25)

	// Velocity bar
	barX := hudX + 10
	barY := hudY + 45
	barW := 190
	barH := 12

	// Bar background
	barBg := color.RGBA{30, 35, 45, 255}
	for py := barY; py < barY+barH; py++ {
		for px := barX; px < barX+barW; px++ {
			screen.Set(px, py, barBg)
		}
	}

	// Bar fill (blue-shifted color when moving fast)
	fillW := int(velocity * float64(barW))
	blueShift := uint8(velocity * 100)
	barFill := color.RGBA{100 - blueShift/2, 150, 200 + blueShift/3, 255}
	for py := barY + 1; py < barY+barH-1; py++ {
		for px := barX + 1; px < barX+1+fillW; px++ {
			screen.Set(px, py, barFill)
		}
	}

	// Progress text
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Cruise: %.0f%%", progress*100), hudX+10, hudY+62)
}

func (g *DemoGame) Layout(outsideWidth, outsideHeight int) (int, int) {
	return display.InternalWidth, display.InternalHeight
}

func main() {
	flag.Parse()

	// Print info
	fmt.Println("Bridge Interior Demo")
	fmt.Println("====================")
	fmt.Printf("Resolution: %dx%d\n", display.InternalWidth, display.InternalHeight)
	fmt.Printf("Debug: %v\n", *debug)
	fmt.Printf("Seed: %d\n", *seed)
	fmt.Println()

	// Set up window
	ebiten.SetWindowSize(display.InternalWidth*2, display.InternalHeight*2) // 2x scale for visibility
	ebiten.SetWindowTitle("Stapledon's Voyage - Bridge Interior Demo")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	// Run game
	game := NewDemoGame()
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
