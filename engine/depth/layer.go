// Package depth defines depth layer types for parallax rendering.
// This package has no dependencies to avoid import cycles.
package depth

// Layer represents a rendering layer with associated parallax factor.
// Lower layers are further away (background), higher layers are closer (foreground).
// We support 20 layers (0-19) for flexible depth composition across ship structures.
type Layer int

const (
	Layer0  Layer = iota // 0.00 - Fixed at infinity (space/galaxy)
	Layer1               // 0.05 - Very distant stars
	Layer2               // 0.10 - Far spire segment
	Layer3               // 0.15 - Distant ship structure
	Layer4               // 0.20 - Opposite hull
	Layer5               // 0.25 - Far deck (5+ decks away)
	Layer6               // 0.30 - Mid-distance structure
	Layer7               // 0.40 - 4 decks away
	Layer8               // 0.50 - 3 decks away
	Layer9               // 0.60 - 2 decks away
	Layer10              // 0.70 - Adjacent deck
	Layer11              // 0.75 - Near background
	Layer12              // 0.80 - Same deck distant
	Layer13              // 0.85 - Same deck mid
	Layer14              // 0.90 - Same deck near
	Layer15              // 0.95 - Current deck background
	Layer16              // 1.00 - Main scene layer
	Layer17              // 1.00 - Scene overlay
	Layer18              // 1.00 - Foreground effects
	Layer19              // 1.00 - UI (screen-fixed)
	LayerCount
)

// Convenience aliases for common use cases
const (
	LayerDeepBackground = Layer0  // Galaxy, fixed at infinity
	LayerMidBackground  = Layer6  // Mid-distance, 0.3x
	LayerScene          = Layer16 // Main content, 1.0x
	LayerForeground     = Layer19 // UI, screen-fixed
)

// layerParallax defines the parallax factor for each layer.
// 0.0 = fixed, 1.0 = moves with camera.
// These are defaults and can be overridden via SetParallax.
var layerParallax = [LayerCount]float64{
	Layer0:  0.00, // Fixed at infinity (galaxy/space)
	Layer1:  0.05, // Very distant stars
	Layer2:  0.10, // Far spire segment
	Layer3:  0.15, // Distant ship structure
	Layer4:  0.20, // Opposite hull
	Layer5:  0.25, // Far deck (5+ decks away)
	Layer6:  0.30, // Mid-distance structure
	Layer7:  0.40, // 4 decks away
	Layer8:  0.50, // 3 decks away
	Layer9:  0.60, // 2 decks away
	Layer10: 0.70, // Adjacent deck
	Layer11: 0.75, // Near background
	Layer12: 0.80, // Same deck distant
	Layer13: 0.85, // Same deck mid
	Layer14: 0.90, // Same deck near
	Layer15: 0.95, // Current deck background
	Layer16: 1.00, // Main scene
	Layer17: 1.00, // Scene overlay
	Layer18: 1.00, // Foreground effects
	Layer19: 1.00, // UI (screen-fixed)
}

// layerNames for debugging
var layerNames = [LayerCount]string{
	"L0-Space",
	"L1-DistantStars",
	"L2-FarSpire",
	"L3-DistantShip",
	"L4-OppositeHull",
	"L5-FarDeck",
	"L6-MidDistance",
	"L7-Deck4Away",
	"L8-Deck3Away",
	"L9-Deck2Away",
	"L10-Adjacent",
	"L11-NearBg",
	"L12-SameDeckFar",
	"L13-SameDeckMid",
	"L14-SameDeckNear",
	"L15-DeckBackground",
	"L16-Scene",
	"L17-SceneOverlay",
	"L18-FgEffects",
	"L19-UI",
}

// Name returns a human-readable name for the layer.
func (l Layer) Name() string {
	if l >= 0 && l < LayerCount {
		return layerNames[l]
	}
	return "Unknown"
}

// Parallax returns the parallax factor for this layer.
func (l Layer) Parallax() float64 {
	if l >= 0 && l < LayerCount {
		return layerParallax[l]
	}
	return 1.0
}

// SetParallax allows configuring the parallax factor for a layer at runtime.
// Useful for tuning the feel of depth effects.
func SetParallax(layer Layer, factor float64) {
	if layer >= 0 && layer < LayerCount {
		layerParallax[layer] = factor
	}
}

// GetAllParallax returns a copy of all parallax factors for inspection.
func GetAllParallax() [LayerCount]float64 {
	return layerParallax
}
