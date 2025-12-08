package view

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"stapledons_voyage/engine/tetra"
)

// PlanetLayer renders 3D planets using Tetra3D.
// It composites over the background and under UI.
type PlanetLayer struct {
	scene   *tetra.Scene
	planets []*tetra.Planet
	sun     *tetra.SunLight
	ambient *tetra.AmbientLight

	screenW int
	screenH int
}

// NewPlanetLayer creates a new planet rendering layer.
func NewPlanetLayer(screenW, screenH int) *PlanetLayer {
	pl := &PlanetLayer{
		screenW: screenW,
		screenH: screenH,
	}

	// Create 3D scene
	pl.scene = tetra.NewScene(screenW, screenH)

	// Add lighting
	pl.sun = tetra.NewSunLight()
	pl.sun.SetPosition(5, 3, 5) // Upper-right-front
	pl.sun.AddToScene(pl.scene)

	pl.ambient = tetra.NewAmbientLight(0.2, 0.2, 0.3, 0.5) // Dim blue ambient
	pl.ambient.AddToScene(pl.scene)

	// Set camera back to see planets
	pl.scene.SetCameraPosition(0, 0, 4)

	return pl
}

// AddPlanet adds a solid-color planet to the layer.
func (pl *PlanetLayer) AddPlanet(name string, radius float64, c color.RGBA) *tetra.Planet {
	planet := tetra.NewPlanet(name, radius, c)
	planet.AddToScene(pl.scene)
	pl.planets = append(pl.planets, planet)
	return planet
}

// AddTexturedPlanet adds a textured planet to the layer.
func (pl *PlanetLayer) AddTexturedPlanet(name string, radius float64, texture *ebiten.Image) *tetra.Planet {
	planet := tetra.NewTexturedPlanet(name, radius, texture)
	planet.AddToScene(pl.scene)
	pl.planets = append(pl.planets, planet)
	return planet
}

// AddExistingPlanet adds an already-created planet to the layer.
func (pl *PlanetLayer) AddExistingPlanet(planet *tetra.Planet) {
	planet.AddToScene(pl.scene)
	pl.planets = append(pl.planets, planet)
}

// RemovePlanet removes a planet by name.
func (pl *PlanetLayer) RemovePlanet(name string) {
	for i, p := range pl.planets {
		if p.Name() == name {
			p.RemoveFromScene()
			pl.planets = append(pl.planets[:i], pl.planets[i+1:]...)
			return
		}
	}
}

// ClearPlanets removes all planets.
func (pl *PlanetLayer) ClearPlanets() {
	for _, p := range pl.planets {
		p.RemoveFromScene()
	}
	pl.planets = nil
}

// Update updates planet animations.
func (pl *PlanetLayer) Update(dt float64) {
	for _, p := range pl.planets {
		p.Update(dt)
	}
}

// Draw renders the 3D planets to the screen.
func (pl *PlanetLayer) Draw(screen *ebiten.Image) {
	if pl.scene == nil || len(pl.planets) == 0 {
		return
	}

	// Render 3D scene
	img3d := pl.scene.Render()

	// Composite over existing content
	screen.DrawImage(img3d, nil)
}

// SetCameraPosition sets the camera position.
func (pl *PlanetLayer) SetCameraPosition(x, y, z float64) {
	if pl.scene != nil {
		pl.scene.SetCameraPosition(x, y, z)
	}
}

// SetSunPosition sets the sun light position.
func (pl *PlanetLayer) SetSunPosition(x, y, z float64) {
	if pl.sun != nil {
		pl.sun.SetPosition(x, y, z)
	}
}

// Scene returns the underlying Tetra3D scene.
func (pl *PlanetLayer) Scene() *tetra.Scene {
	return pl.scene
}

// Resize updates the layer for new screen dimensions.
func (pl *PlanetLayer) Resize(screenW, screenH int) {
	pl.screenW = screenW
	pl.screenH = screenH
	// Note: Scene would need to recreate buffers - for now, keep original size
}
