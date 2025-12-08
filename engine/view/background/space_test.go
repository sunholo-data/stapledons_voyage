//go:build !ci

package background

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestNewSpaceBackground(t *testing.T) {
	bg := NewSpaceBackground(1280, 960)

	if bg == nil {
		t.Fatal("NewSpaceBackground returned nil")
	}

	if bg.screenW != 1280 {
		t.Errorf("screenW = %d, want 1280", bg.screenW)
	}
	if bg.screenH != 960 {
		t.Errorf("screenH = %d, want 960", bg.screenH)
	}

	// Should have 3 default layers
	if len(bg.starLayers) != 3 {
		t.Errorf("len(starLayers) = %d, want 3", len(bg.starLayers))
	}
}

func TestSpaceBackgroundLayers(t *testing.T) {
	bg := NewSpaceBackground(1280, 960)

	// Verify layer parallax values
	expectedParallax := []float64{0.0, 0.3, 0.7}
	for i, layer := range bg.starLayers {
		if layer.config.Parallax != expectedParallax[i] {
			t.Errorf("layer[%d].Parallax = %v, want %v",
				i, layer.config.Parallax, expectedParallax[i])
		}
	}

	// Verify star counts
	expectedCounts := []int{500, 300, 100}
	for i, layer := range bg.starLayers {
		if layer.config.Count != expectedCounts[i] {
			t.Errorf("layer[%d].Count = %d, want %d",
				i, layer.config.Count, expectedCounts[i])
		}
	}
}

func TestSpaceBackgroundSetters(t *testing.T) {
	bg := NewSpaceBackground(1280, 960)

	bg.SetParallax(0.5)
	if bg.parallaxDepth != 0.5 {
		t.Errorf("parallaxDepth = %v, want 0.5", bg.parallaxDepth)
	}

	bg.SetVelocity(0.3)
	if bg.GetVelocity() != 0.3 {
		t.Errorf("velocity = %v, want 0.3", bg.GetVelocity())
	}

	bg.SetGRIntensity(0.8)
	if bg.GetGRIntensity() != 0.8 {
		t.Errorf("grIntensity = %v, want 0.8", bg.GetGRIntensity())
	}
}

func TestSpaceBackgroundDraw(t *testing.T) {
	bg := NewSpaceBackground(100, 100) // Small for testing
	screen := ebiten.NewImage(100, 100)
	camera := NewCameraOffset()

	// Should not panic
	bg.Draw(screen, camera)

	// Draw with nil camera
	bg.Draw(screen, nil)
}

func TestSpaceBackgroundResize(t *testing.T) {
	bg := NewSpaceBackground(100, 100)

	bg.Resize(200, 150)

	if bg.screenW != 200 {
		t.Errorf("after resize, screenW = %d, want 200", bg.screenW)
	}
	if bg.screenH != 150 {
		t.Errorf("after resize, screenH = %d, want 150", bg.screenH)
	}
}

func TestSpaceBackgroundCustomLayers(t *testing.T) {
	bg := NewSpaceBackground(100, 100)
	initialLayers := len(bg.starLayers)

	// Add custom layer
	custom := NewStarLayer(StarLayerConfig{
		Count:         50,
		MinBrightness: 0.5,
		MaxBrightness: 1.0,
		MinSize:       1.0,
		MaxSize:       2.0,
		Parallax:      0.5,
		Seed:          999,
	}, 100, 100)

	bg.AddStarLayer(custom)

	if len(bg.starLayers) != initialLayers+1 {
		t.Errorf("len(starLayers) = %d, want %d", len(bg.starLayers), initialLayers+1)
	}

	// Clear layers
	bg.ClearLayers()
	if len(bg.starLayers) != 0 {
		t.Errorf("after clear, len(starLayers) = %d, want 0", len(bg.starLayers))
	}
}

func TestStarLayerGeneration(t *testing.T) {
	layer := NewStarLayer(StarLayerConfig{
		Count:         100,
		MinBrightness: 0.3,
		MaxBrightness: 0.9,
		MinSize:       1.0,
		MaxSize:       3.0,
		Parallax:      0.5,
		Seed:          42,
	}, 800, 600)

	if len(layer.stars) != 100 {
		t.Errorf("len(stars) = %d, want 100", len(layer.stars))
	}

	// Verify all stars have valid values
	for i, star := range layer.stars {
		if star.Brightness < 0.3 || star.Brightness > 0.9 {
			t.Errorf("star[%d].Brightness = %v, out of range [0.3, 0.9]",
				i, star.Brightness)
		}
		if star.Size < 1.0 || star.Size > 3.0 {
			t.Errorf("star[%d].Size = %v, out of range [1.0, 3.0]",
				i, star.Size)
		}
	}
}

func TestStarLayerDeterministic(t *testing.T) {
	// Same seed should produce same stars
	layer1 := NewStarLayer(StarLayerConfig{
		Count: 10,
		Seed:  42,
	}, 100, 100)

	layer2 := NewStarLayer(StarLayerConfig{
		Count: 10,
		Seed:  42,
	}, 100, 100)

	for i := range layer1.stars {
		if layer1.stars[i].X != layer2.stars[i].X {
			t.Errorf("star[%d].X differs with same seed", i)
		}
		if layer1.stars[i].Y != layer2.stars[i].Y {
			t.Errorf("star[%d].Y differs with same seed", i)
		}
	}
}

func TestStarLayerRegenerate(t *testing.T) {
	layer := NewStarLayer(StarLayerConfig{
		Count: 50,
		Seed:  123,
	}, 100, 100)

	original := make([]Star, len(layer.stars))
	copy(original, layer.stars)

	// Regenerate with same dimensions should give same stars (same seed)
	layer.Regenerate(100, 100)

	for i := range layer.stars {
		if layer.stars[i].X != original[i].X ||
			layer.stars[i].Y != original[i].Y {
			t.Errorf("star[%d] changed after regenerate with same dimensions", i)
		}
	}
}

func TestStarLayerDraw(t *testing.T) {
	layer := NewStarLayer(StarLayerConfig{
		Count: 10,
		Seed:  42,
	}, 100, 100)

	screen := ebiten.NewImage(100, 100)

	// Should not panic
	layer.Draw(screen, 0, 0)
	layer.Draw(screen, 50, 50)   // With offset
	layer.Draw(screen, -50, -50) // Negative offset
}
