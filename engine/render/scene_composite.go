package render

import (
	"fmt"
	"image"
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"stapledons_voyage/sim_gen"
)

// sceneWindowMasks caches window mask images for deck backgrounds
type sceneWindowMasks struct {
	bridgeMask      *ebiten.Image
	observationMask *ebiten.Image
	loaded          bool
}

// loadSceneWindowMasks loads window mask images for all decks
func (r *Renderer) loadSceneWindowMasks() *sceneWindowMasks {
	masks := &sceneWindowMasks{}

	// Load bridge window mask
	if bridgeMask, err := loadMaskImage("assets/decks/bridge/window_mask_large.png"); err == nil {
		masks.bridgeMask = bridgeMask
		log.Printf("Loaded bridge window mask")
	} else {
		log.Printf("Warning: Could not load bridge window mask: %v", err)
	}

	// Load observation deck window mask
	if obsMask, err := loadMaskImage("assets/decks/observation/window_mask_large.png"); err == nil {
		masks.observationMask = obsMask
		log.Printf("Loaded observation deck window mask")
	} else {
		log.Printf("Warning: Could not load observation window mask: %v", err)
	}

	masks.loaded = true
	return masks
}

// loadMaskImage loads a PNG mask image
func loadMaskImage(path string) (*ebiten.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("decode mask: %w", err)
	}

	return ebiten.NewImageFromImage(img), nil
}

// compositeSceneWindows renders scene-based interiors with window compositing
// This handles the ViewInterior case where we show:
// 1. 3D space (planets, stars) rendered to buffer
// 2. 2D deck background with transparent windows
// 3. Composited result showing space through windows
func (r *Renderer) compositeSceneWindows(screen *ebiten.Image, out *sim_gen.FrameOutput, masks *sceneWindowMasks, screenW, screenH int) {
	// Separate DrawCmds into:
	// - spaceCommands: 3D planets, stars, space background (render to buffer)
	// - deckBackground: 2D UI sprite for deck scene (render last, on top)
	// - hudCommands: Text/UI overlay (render after everything)
	var spaceCommands []*sim_gen.DrawCmd
	var deckBackground *sim_gen.DrawCmd
	var shipState *sim_gen.DrawCmd
	var hudCommands []*sim_gen.DrawCmd

	for _, cmd := range out.Draw {
		switch cmd.Kind {
		case sim_gen.DrawCmdKindShipState3D:
			shipState = cmd
		case sim_gen.DrawCmdKindSpaceBg, sim_gen.DrawCmdKindTexturedPlanet:
			spaceCommands = append(spaceCommands, cmd)
		case sim_gen.DrawCmdKindUi:
			// Deck background (sprite 9000-9001)
			if cmd.Ui.SpriteId >= 9000 && cmd.Ui.SpriteId < 9010 {
				deckBackground = cmd
			} else {
				// Other UI elements (render as HUD)
				hudCommands = append(hudCommands, cmd)
			}
		case sim_gen.DrawCmdKindText:
			// HUD text overlay
			hudCommands = append(hudCommands, cmd)
		}
	}

	// Determine which window mask to use based on deck sprite ID
	var windowMask *ebiten.Image
	if deckBackground != nil {
		switch deckBackground.Ui.SpriteId {
		case 9000: // Bridge
			windowMask = masks.bridgeMask
		case 9001: // Observation
			windowMask = masks.observationMask
		}
	}

	if windowMask == nil {
		// No mask available, render everything normally
		r.RenderFrame(screen, *out)
		return
	}

	// Step 1: Render deck background with same transform as mask
	// Get actual sprite dimensions (supports any aspect ratio: 16:9, 21:9, etc.)
	var bgW, bgH float64
	var sprite *ebiten.Image
	if deckBackground != nil && r.assets != nil {
		sprite = r.assets.GetSprite(int(deckBackground.Ui.SpriteId))
		if sprite != nil {
			bounds := sprite.Bounds()
			bgW = float64(bounds.Dx())
			bgH = float64(bounds.Dy())
		}
	}

	// If we couldn't get sprite dimensions, fall back to default 16:9
	if bgW == 0 || bgH == 0 {
		bgW, bgH = 1344.0, 768.0
	}

	// Calculate uniform scale and offset for both background and mask
	scaleX := float64(screenW) / bgW
	scaleY := float64(screenH) / bgH
	scale := min(scaleX, scaleY) // Maintain aspect ratio

	offsetX := (float64(screenW) - bgW*scale) / 2
	offsetY := (float64(screenH) - bgH*scale) / 2

	// Render deck background
	if sprite != nil {
		log.Printf("DEBUG: Rendering deck background sprite ID %d (%dx%d) at scale %.2f, offset (%.0f, %.0f)",
			deckBackground.Ui.SpriteId, int(bgW), int(bgH), scale, offsetX, offsetY)
		opBg := &ebiten.DrawImageOptions{}
		opBg.GeoM.Scale(scale, scale)
		opBg.GeoM.Translate(offsetX, offsetY)
		screen.DrawImage(sprite, opBg)
	} else {
		log.Printf("DEBUG: No sprite found for deck background ID %d", deckBackground.Ui.SpriteId)
	}

	// Step 2: Render 3D space to offscreen buffer
	spaceBuffer := ebiten.NewImage(screenW, screenH)
	spaceBuffer.Clear() // Start with transparent

	// Create temporary FrameOutput with just space commands
	spaceOut := &sim_gen.FrameOutput{
		Draw:       spaceCommands,
		Camera:     out.Camera,
		Relativity: out.Relativity,
		Lighting:   out.Lighting,
		Lod:        out.Lod,
	}

	// Add ShipState if present (needed for SR/GR effects)
	if shipState != nil {
		spaceOut.Draw = append([]*sim_gen.DrawCmd{shipState}, spaceOut.Draw...)
	}

	// Render space commands to buffer
	r.RenderFrame(spaceBuffer, *spaceOut)

	// Step 3: Create fullscreen mask using same scale/offset
	fullscreenMask := ebiten.NewImage(screenW, screenH)
	fullscreenMask.Clear()

	opMask := &ebiten.DrawImageOptions{}
	opMask.GeoM.Scale(scale, scale)
	opMask.GeoM.Translate(offsetX, offsetY)
	fullscreenMask.DrawImage(windowMask, opMask)

	// Step 4: Apply mask to space view (DestinationIn keeps only masked regions)
	maskedSpace := ebiten.NewImage(screenW, screenH)
	maskedSpace.Clear()
	maskedSpace.DrawImage(spaceBuffer, nil)

	opApply := &ebiten.DrawImageOptions{}
	opApply.CompositeMode = ebiten.CompositeModeDestinationIn
	maskedSpace.DrawImage(fullscreenMask, opApply)

	// Step 5: Composite masked space onto deck background (already drawn)
	screen.DrawImage(maskedSpace, nil)

	// Step 6: Render HUD overlay (text and UI elements on top of everything)
	for _, cmd := range hudCommands {
		switch cmd.Kind {
		case sim_gen.DrawCmdKindText:
			r.drawText(screen, cmd.Text, int(cmd.Text.X), int(cmd.Text.Y))
		case sim_gen.DrawCmdKindUi:
			r.drawUiElement(screen, cmd.Ui, screenW, screenH)
		}
	}
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
