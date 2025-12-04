package scenario

import (
	"errors"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"stapledons_voyage/engine/assets"
	"stapledons_voyage/engine/display"
	"stapledons_voyage/engine/render"
	"stapledons_voyage/sim_gen"
)

// VisualRunner executes a visual test scenario with screenshots.
type VisualRunner struct {
	scenario     *Scenario
	world        *sim_gen.World // Typed world (M-DX16: RecordUpdate preserves struct types)
	out          sim_gen.FrameOutput
	renderer     *render.Renderer
	currentFrame int
	maxFrame     int
	activeKeys   map[int]bool
	pressedKeys  map[int]bool // Keys pressed this frame (for "pressed" events)
	pendingClick *Click
	captures     map[int]string // frame -> filename
	outputDir    string
	capturedImgs map[string]*image.RGBA
	done         bool
	err          error

	// Frame timing for performance analysis
	frameStart  time.Time
	frameTimes  []time.Duration
	stepTimes   []time.Duration
	renderTimes []time.Duration
}

// NewVisualRunner creates a visual scenario runner.
func NewVisualRunner(s *Scenario, outputDir string) *VisualRunner {
	// Find max frame
	maxFrame := 0
	captures := make(map[int]string)
	for _, e := range s.Events {
		if e.Frame > maxFrame {
			maxFrame = e.Frame
		}
		if e.Capture != "" {
			captures[e.Frame] = e.Capture
		}
	}
	// Run one extra frame after last event to ensure final captures
	maxFrame++

	return &VisualRunner{
		scenario:     s,
		maxFrame:     maxFrame,
		activeKeys:   make(map[int]bool),
		pressedKeys:  make(map[int]bool),
		captures:     captures,
		outputDir:    outputDir,
		capturedImgs: make(map[string]*image.RGBA),
	}
}

func (r *VisualRunner) Update() error {
	if r.currentFrame > r.maxFrame {
		r.done = true
		return errors.New("scenario complete")
	}

	// Start frame timing
	r.frameStart = time.Now()

	// Clear pressed keys from previous frame
	r.pressedKeys = make(map[int]bool)

	// Process events for this frame
	for _, e := range r.scenario.Events {
		if e.Frame != r.currentFrame {
			continue
		}

		// Handle key events
		if e.Key != "" {
			code := KeyNameToCode(e.Key)
			if code >= 0 {
				switch e.Action {
				case "down":
					r.activeKeys[code] = true
				case "up":
					delete(r.activeKeys, code)
				case "press":
					// Press = down this frame + "pressed" event, up next frame
					r.activeKeys[code] = true
					r.pressedKeys[code] = true
				}
			}
		}

		// Handle click events
		if e.Click != nil {
			r.pendingClick = e.Click
		}
	}

	// Clear "press" keys after one frame
	for _, e := range r.scenario.Events {
		if e.Frame == r.currentFrame-1 && e.Action == "press" {
			code := KeyNameToCode(e.Key)
			delete(r.activeKeys, code)
		}
	}

	// Build input and step simulation
	input := BuildFrameInput(r.activeKeys, r.pressedKeys, r.pendingClick, r.scenario.TestMode)
	r.pendingClick = nil // Clear pending click

	// Step returns []interface{}{newWorld, output} - RecordUpdate preserves *World type
	stepStart := time.Now()
	result := sim_gen.Step(r.world, input)
	r.stepTimes = append(r.stepTimes, time.Since(stepStart))

	tuple, ok := result.([]interface{})
	if !ok || len(tuple) != 2 {
		r.err = fmt.Errorf("simulation error at frame %d: unexpected Step result", r.currentFrame)
		return r.err
	}
	if w, ok := tuple[0].(*sim_gen.World); ok {
		r.world = w
	}
	if out, ok := tuple[1].(*sim_gen.FrameOutput); ok {
		r.out = *out
	}

	r.currentFrame++
	return nil
}

func (r *VisualRunner) Draw(screen *ebiten.Image) {
	renderStart := time.Now()
	r.renderer.RenderFrame(screen, r.out)
	r.renderTimes = append(r.renderTimes, time.Since(renderStart))

	// Record total frame time (Update + Draw)
	if !r.frameStart.IsZero() {
		r.frameTimes = append(r.frameTimes, time.Since(r.frameStart))
	}

	// Check if we should capture this frame
	// Note: currentFrame was already incremented in Update(), so check currentFrame-1
	frameJustProcessed := r.currentFrame - 1
	if frameJustProcessed >= 0 {
		if filename, ok := r.captures[frameJustProcessed]; ok {
			bounds := screen.Bounds()
			img := image.NewRGBA(bounds)
			screen.ReadPixels(img.Pix)
			r.capturedImgs[filename] = img
		}
	}
}

// FrameTimes returns collected frame timing data.
func (r *VisualRunner) FrameTimes() []time.Duration {
	return r.frameTimes
}

// StepTimes returns collected simulation step timing data.
func (r *VisualRunner) StepTimes() []time.Duration {
	return r.stepTimes
}

// RenderTimes returns collected render timing data.
func (r *VisualRunner) RenderTimes() []time.Duration {
	return r.renderTimes
}

func (r *VisualRunner) Layout(outsideWidth, outsideHeight int) (int, int) {
	return display.InternalWidth, display.InternalHeight
}

// RunVisualScenario executes a visual scenario and saves screenshots.
func RunVisualScenario(scenarioPath, outputDir string) error {
	return RunVisualScenarioWithOptions(scenarioPath, outputDir, false, false)
}

// RunVisualScenarioWithOptions executes a scenario with optional test mode override.
// If testModeOverride is true, testMode value overrides the scenario's JSON setting.
func RunVisualScenarioWithOptions(scenarioPath, outputDir string, testMode, testModeOverride bool) error {
	s, err := LoadScenario(scenarioPath)
	if err != nil {
		return err
	}

	// CLI flag overrides JSON setting if specified
	if testModeOverride {
		s.TestMode = testMode
	}

	fmt.Printf("Running scenario: %s\n", s.Name)
	fmt.Printf("Description: %s\n", s.Description)
	if s.TestMode {
		fmt.Println("Test mode: UI stripped for golden file comparison")
	}

	// Note: Star catalog loading removed - not available in AILANG codegen yet

	// Initialize asset manager
	assetMgr, _ := assets.NewManager("assets")
	renderer := render.NewRenderer(assetMgr)

	// Initialize world - type assert to *World (M-DX16: struct types preserved)
	worldIface := sim_gen.InitWorld(s.Seed)
	world, ok := worldIface.(*sim_gen.World)
	if !ok {
		return fmt.Errorf("InitWorld did not return *World")
	}

	runner := NewVisualRunner(s, outputDir)
	runner.world = world
	runner.renderer = renderer

	// Run with window
	ebiten.SetWindowSize(display.InternalWidth, display.InternalHeight)
	ebiten.SetWindowTitle("Scenario: " + s.Name)
	ebiten.SetScreenClearedEveryFrame(true)

	err = ebiten.RunGame(runner)
	if err != nil && err.Error() != "scenario complete" {
		return err
	}

	if runner.err != nil {
		return runner.err
	}

	// Save captured images
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	for filename, img := range runner.capturedImgs {
		path := filepath.Join(outputDir, filename)
		f, err := os.Create(path)
		if err != nil {
			return fmt.Errorf("failed to create %s: %w", path, err)
		}

		if err := png.Encode(f, img); err != nil {
			f.Close()
			return fmt.Errorf("failed to encode %s: %w", path, err)
		}
		f.Close()
		fmt.Printf("  Captured: %s\n", path)
	}

	fmt.Printf("Scenario complete: %d frames, %d captures\n", runner.currentFrame, len(runner.capturedImgs))
	return nil
}
