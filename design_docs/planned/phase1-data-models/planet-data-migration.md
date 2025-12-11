# Planet Data Migration to AILANG

## Status
- Status: Planned
- Priority: P1 (Architecture)
- Complexity: Medium
- Part of: [view-layer-ailang-migration.md](view-layer-ailang-migration.md)
- Estimated: 1 day

## Problem Statement

The `DomeRenderer` in `engine/view/dome_renderer.go` has hardcoded planet data:

```go
// Current (WRONG) - Go owns game data
func (d *DomeRenderer) createSolarSystem() {
    planetConfigs := []struct {
        name     string
        color    color.RGBA
        radius   float64
        distance float64  // Orbital distance
    }{
        {"neptune", color.RGBA{80, 120, 200, 255}, 1.0, 15},
        {"saturn", color.RGBA{210, 190, 150, 255}, 1.8, 50},
        {"jupiter", color.RGBA{180, 140, 100, 255}, 2.2, 80},
        // ... more planets
    }
}
```

**Why this is wrong:**
- Planets are **game entities** - they have state (civilizations, resources)
- Planet positions affect gameplay (which systems to visit)
- Hardcoded data can't respond to game events (planet destroyed, terraformed)
- The dome should show planets from the **world state**, not static config

## Target Architecture

```
AILANG (sim/world.ail)         Engine (engine/view/planet_renderer.go)
├── StarSystem                 ├── Receives [Planet] from AILANG
│   ├── planets: [Planet]      ├── Renders 3D spheres at positions
│   └── position: GalacticCoord├── Applies textures/shaders
├── Planet                     └── No hardcoded planet data
│   ├── name: string
│   ├── orbit_distance: float
│   ├── radius: float
│   └── planet_type: PlanetType
```

## AILANG Implementation

### Types (sim/celestial.ail)

```ailang
module sim/celestial

import std/prelude

-- Planet classification
type PlanetType =
    | Rocky        -- Mercury, Mars-like
    | GasGiant     -- Jupiter, Saturn-like
    | IceGiant     -- Neptune, Uranus-like
    | Terrestrial  -- Earth-like
    | Ocean        -- Water world
    | Volcanic     -- Io-like
    | Dwarf        -- Pluto-like

-- Individual planet
type Planet = {
    id: PlanetID,
    name: string,
    planet_type: PlanetType,
    orbit_distance: float,    -- AU from star
    radius: float,            -- Earth radii
    orbital_period: float,    -- Years
    current_angle: float,     -- Current orbital position (radians)
    has_rings: bool,
    ring_color: Option(Color),
    atmosphere_color: Option(Color),
    civilization: Option(CivilizationID)
}

-- Star system containing planets
type StarSystem = {
    id: SystemID,
    name: string,
    star_type: StarType,
    position: GalacticCoord,
    planets: [Planet]
}

type StarType =
    | MainSequence(SpectralClass)  -- G2V (Sun), M0V (red dwarf)
    | Giant
    | WhiteDwarf
    | NeutronStar
    | BlackHole

type SpectralClass = O | B | A | F | G | K | M

type Color = { r: int, g: int, b: int, a: int }
```

### Planet Colors (derived from type)

```ailang
-- Get base color for planet type
export pure func planet_base_color(pt: PlanetType) -> Color {
    match pt {
        Rocky       => { r: 180, g: 140, b: 100, a: 255 },
        GasGiant    => { r: 200, g: 170, b: 130, a: 255 },
        IceGiant    => { r: 80, g: 120, b: 200, a: 255 },
        Terrestrial => { r: 80, g: 140, b: 200, a: 255 },
        Ocean       => { r: 40, g: 80, b: 180, a: 255 },
        Volcanic    => { r: 200, g: 80, b: 40, a: 255 },
        Dwarf       => { r: 150, g: 140, b: 130, a: 255 }
    }
}

-- Get atmosphere tint if present
export pure func atmosphere_tint(planet: Planet) -> Option(Color) {
    match planet.planet_type {
        GasGiant    => Some({ r: 220, g: 180, b: 140, a: 100 }),
        IceGiant    => Some({ r: 150, g: 200, b: 255, a: 100 }),
        Terrestrial => planet.atmosphere_color,
        _ => None
    }
}
```

### Solar System Generation

```ailang
-- Generate Sol system (our solar system)
export pure func init_sol_system() -> StarSystem {
    {
        id: SystemID(0),
        name: "Sol",
        star_type: MainSequence(G),
        position: { x: 0.0, y: 0.0, z: 0.0 },
        planets: [
            { id: PlanetID(0), name: "Mercury", planet_type: Rocky,
              orbit_distance: 0.39, radius: 0.38, orbital_period: 0.24,
              current_angle: 0.0, has_rings: false, ring_color: None,
              atmosphere_color: None, civilization: None },

            { id: PlanetID(1), name: "Venus", planet_type: Rocky,
              orbit_distance: 0.72, radius: 0.95, orbital_period: 0.62,
              current_angle: 1.2, has_rings: false, ring_color: None,
              atmosphere_color: Some({ r: 255, g: 200, b: 100, a: 150 }),
              civilization: None },

            { id: PlanetID(2), name: "Earth", planet_type: Terrestrial,
              orbit_distance: 1.0, radius: 1.0, orbital_period: 1.0,
              current_angle: 2.1, has_rings: false, ring_color: None,
              atmosphere_color: Some({ r: 100, g: 150, b: 255, a: 100 }),
              civilization: Some(CivilizationID(0)) },

            { id: PlanetID(3), name: "Mars", planet_type: Rocky,
              orbit_distance: 1.52, radius: 0.53, orbital_period: 1.88,
              current_angle: 3.5, has_rings: false, ring_color: None,
              atmosphere_color: Some({ r: 200, g: 150, b: 100, a: 50 }),
              civilization: None },

            { id: PlanetID(4), name: "Jupiter", planet_type: GasGiant,
              orbit_distance: 5.2, radius: 11.2, orbital_period: 11.86,
              current_angle: 0.8, has_rings: false, ring_color: None,
              atmosphere_color: None, civilization: None },

            { id: PlanetID(5), name: "Saturn", planet_type: GasGiant,
              orbit_distance: 9.5, radius: 9.4, orbital_period: 29.46,
              current_angle: 4.2, has_rings: true,
              ring_color: Some({ r: 210, g: 190, b: 150, a: 200 }),
              atmosphere_color: None, civilization: None },

            { id: PlanetID(6), name: "Uranus", planet_type: IceGiant,
              orbit_distance: 19.2, radius: 4.0, orbital_period: 84.01,
              current_angle: 5.1, has_rings: true,
              ring_color: Some({ r: 150, g: 180, b: 200, a: 100 }),
              atmosphere_color: None, civilization: None },

            { id: PlanetID(7), name: "Neptune", planet_type: IceGiant,
              orbit_distance: 30.1, radius: 3.9, orbital_period: 164.8,
              current_angle: 1.9, has_rings: false, ring_color: None,
              atmosphere_color: None, civilization: None }
        ]
    }
}

-- Update planet orbital positions over time
export pure func step_system(system: StarSystem, dt: float) -> StarSystem {
    let updated_planets = map(system.planets, \p. step_planet_orbit(p, dt));
    { system | planets: updated_planets }
}

pure func step_planet_orbit(planet: Planet, dt: float) -> Planet {
    -- Angular velocity = 2π / period (in years)
    let angular_vel = 6.28318 / planet.orbital_period;
    let new_angle = planet.current_angle + angular_vel * dt;
    { planet | current_angle: mod_float(new_angle, 6.28318) }
}
```

### Rendering

```ailang
-- Generate DrawCmds for planets in dome view
export pure func render_planets(system: StarSystem, view_scale: float) -> [DrawCmd] {
    flat_map(system.planets, \p. render_planet(p, view_scale))
}

pure func render_planet(planet: Planet, scale: float) -> [DrawCmd] {
    -- Convert orbital position to screen position
    let x = cos(planet.current_angle) * planet.orbit_distance * scale;
    let y = sin(planet.current_angle) * planet.orbit_distance * scale * 0.3;  -- Foreshorten for 3D effect
    let z = sin(planet.current_angle) * planet.orbit_distance;  -- Depth

    let base_color = planet_base_color(planet.planet_type);
    let screen_radius = planet.radius * scale * 2.0;

    -- Planet sphere
    let planet_cmd = Planet3D(planet.id, x, y, z, screen_radius, base_color);

    -- Rings if present
    let ring_cmds = match planet.has_rings {
        true => match planet.ring_color {
            Some(rc) => [PlanetRing(planet.id, x, y, z, screen_radius * 2.0, rc)],
            None => []
        },
        false => []
    };

    planet_cmd :: ring_cmds
}
```

## Engine Changes

### Before (dome_renderer.go)

```go
func (d *DomeRenderer) createSolarSystem() {
    // DELETE - hardcoded planet data
    planetConfigs := []struct{...}{
        {"neptune", color.RGBA{80, 120, 200, 255}, 1.0, 15},
        // ...
    }
}
```

### After (planet_renderer.go)

```go
type PlanetRenderer struct {
    // Only rendering resources
    sphereShader *shader.Planet3D
    ringShader   *shader.Ring
    textures     map[sim_gen.PlanetType]*ebiten.Image
}

// No createSolarSystem() - data comes from AILANG

func (r *PlanetRenderer) Render(screen *ebiten.Image, planets []sim_gen.Planet, viewScale float64) {
    for _, planet := range planets {
        // Position from AILANG
        x := math.Cos(planet.CurrentAngle) * planet.OrbitDistance * viewScale
        y := math.Sin(planet.CurrentAngle) * planet.OrbitDistance * viewScale * 0.3

        // Color from AILANG
        color := sim_gen.PlanetBaseColor(planet.PlanetType)

        r.renderSphere(screen, x, y, planet.Radius*viewScale, color)

        if planet.HasRings {
            r.renderRings(screen, x, y, planet.Radius*viewScale*2, planet.RingColor)
        }
    }
}
```

## Integration with World State

```ailang
-- In sim/world.ail
type World = {
    -- ... existing fields
    current_system: Option(StarSystem),  -- System we're in or approaching
    galaxy: Galaxy                        -- All known systems
}

-- When approaching a system, load its planets
pure func begin_approach(world: World, system_id: SystemID) -> World {
    let system = get_system(world.galaxy, system_id);
    { world | current_system: Some(system) }
}
```

## Migration Steps

### Phase 1: Add AILANG Types
- [ ] Create `sim/celestial.ail` with Planet, StarSystem types
- [ ] Add `planet_base_color()` function
- [ ] Add `init_sol_system()` for testing
- [ ] Run `make sim` to generate Go code

### Phase 2: Add to World
- [ ] Add `current_system: Option(StarSystem)` to World
- [ ] Add `step_system()` call in step function
- [ ] Pass planets to render functions

### Phase 3: Refactor Engine
- [ ] Remove `createSolarSystem()` from dome_renderer.go
- [ ] Update `PlanetRenderer.Render()` to take `[]sim_gen.Planet`
- [ ] Remove hardcoded color/size data

### Phase 4: Cleanup
- [ ] Verify planet rendering matches old visuals
- [ ] Add more star systems to test
- [ ] Update tests

## Success Criteria

- [ ] Planet types defined in AILANG
- [ ] Sol system generated by AILANG function
- [ ] Planet positions updated by `step_system()`
- [ ] No hardcoded planet data in Go
- [ ] Planets render from AILANG state
- [ ] Orbital animation works correctly

## Testing

```bash
# Generate and verify types
make sim
go build ./...

# Visual test
make run
# Verify planets orbit, colors match, rings render
```

## References

- [view-layer-ailang-migration.md](view-layer-ailang-migration.md) - Parent migration doc
- [dome-state-migration.md](dome-state-migration.md) - Related dome state work
