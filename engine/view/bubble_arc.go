package view

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
)

// BubbleArc renders the visible edge of the ship's protective dome.
// The dome curves UPWARD from the sides - we're on the bridge floor
// looking UP through the transparent dome at space.
// Features:
// - Curved translucent arc rising from left/right edges
// - Debris/particles flowing past OUTSIDE (showing speed/parallax)
type BubbleArc struct {
	screenW int
	screenH int

	// Arc geometry - curves UP from sides
	// Arc center is BELOW the screen (we see the top portion curving up)
	arcCenterY float64 // Y position of arc center (below screen)
	arcRadius  float64 // Radius of the arc

	// Debris particles (outside the dome, showing movement)
	debris []DebrisParticle
	rng    *rand.Rand

	// Animation state
	time           float64
	spawnAccum     float64 // Accumulator for particle spawning
	velocity       float64 // Ship velocity affects debris flow

	// Colors
	arcColor    color.RGBA
	arcGlow     color.RGBA
	debrisColor color.RGBA
}

// DebrisParticle represents space debris/dust deflecting around the ship.
// Particles come from ahead (top-center) and deflect outward to sides.
type DebrisParticle struct {
	X, Y       float64
	VX, VY     float64 // Deflection velocity (outward from center)
	Size       float64
	Brightness float64
	Streak     float64 // Length of motion blur streak
}

// NewBubbleArc creates a new bubble arc renderer.
// The arc originates from the disc edges and curves upward around the spire.
func NewBubbleArc(screenW, screenH int) *BubbleArc {
	// Disc parameters (must match bridge_view.go - LARGER disc)
	discCenterY := float64(screenH) * 0.65
	discRadiusX := float64(screenW) * 0.58

	ba := &BubbleArc{
		screenW:     screenW,
		screenH:     screenH,
		arcCenterY:  discCenterY + 250, // Arc center below disc
		arcRadius:   discRadiusX + 150, // Arc extends from disc edges upward
		rng:         rand.New(rand.NewSource(42)),
		debris:      make([]DebrisParticle, 0, 50),
		arcColor:    color.RGBA{60, 100, 160, 100},  // Translucent blue dome edge
		arcGlow:     color.RGBA{100, 150, 220, 60},  // Soft glow
		debrisColor: color.RGBA{200, 200, 220, 220}, // Brighter debris for visibility
	}

	// Spawn initial debris
	for i := 0; i < 15; i++ {
		ba.spawnDebris()
	}

	return ba
}

// spawnDebris creates a new debris particle that deflects around the ship.
// Particles spawn from ahead (upper-center) and fan UPWARD and outward
// to BOTH sides as they deflect off the bubble shield.
func (ba *BubbleArc) spawnDebris() {
	// Direction of travel is upper-center
	centerX := float64(ba.screenW) / 2
	travelPointY := float64(ba.screenH) * 0.25 // Upper area

	// Spawn position: near the center-top with randomization
	spawnRadius := 40.0 + ba.rng.Float64()*60.0
	spawnAngle := ba.rng.Float64() * 2 * 3.14159
	x := centerX + spawnRadius*math.Sin(spawnAngle)*0.8
	y := travelPointY + spawnRadius*math.Cos(spawnAngle)*0.4

	// Base velocity - particles deflect UPWARD and outward to BOTH sides
	speedMult := 100.0 + ba.velocity*400.0

	// Deflection direction based on which side of center
	// Particles fan out UP and to the side they're on
	dx := x - centerX

	// Outward velocity - proportional to distance from center
	// Particles on the left go left, particles on the right go right
	vx := speedMult * (dx / 150.0) * (0.8 + ba.rng.Float64()*0.4)

	// Upward velocity (negative Y = up)
	vy := -speedMult * (0.5 + ba.rng.Float64()*0.3)

	// Size and brightness
	size := 1.5 + ba.rng.Float64()*2.0
	brightness := 0.5 + ba.rng.Float64()*0.5

	// Streak based on velocity
	streak := math.Sqrt(vx*vx+vy*vy) * 0.02

	p := DebrisParticle{
		X:          x,
		Y:          y,
		VX:         vx,
		VY:         vy,
		Size:       size,
		Brightness: brightness,
		Streak:     streak,
	}

	ba.debris = append(ba.debris, p)
}

// Update updates the bubble arc animation.
func (ba *BubbleArc) Update(dt float64) {
	ba.time += dt

	// Update debris particles
	alive := ba.debris[:0]
	for i := range ba.debris {
		p := &ba.debris[i]

		// Move particle (both X and Y)
		p.X += p.VX * dt
		p.Y += p.VY * dt

		// Particles curve outward as they move UP and around the bubble
		// Accelerate outward from center as they rise
		centerX := float64(ba.screenW) / 2
		if p.X > centerX {
			p.VX += 30 * dt // Accelerate outward right
		} else {
			p.VX -= 30 * dt // Accelerate outward left
		}

		// Slight upward acceleration (being pushed up and over)
		p.VY -= 20 * dt

		// Update streak based on current velocity
		p.Streak = math.Sqrt(p.VX*p.VX+p.VY*p.VY) * 0.012

		// Keep if still on screen (particles move UP and out)
		onScreen := p.X > -100 && p.X < float64(ba.screenW)+100 &&
			p.Y > -100 && p.Y < float64(ba.screenH)*0.6
		if onScreen {
			alive = append(alive, *p)
		}
	}
	ba.debris = alive

	// Spawn new debris continuously - accumulate fractional spawns
	// Base rate of ~1 particle per second, more at higher velocity
	spawnRate := 1.0 + ba.velocity*5.0
	ba.spawnAccum += spawnRate * dt
	for ba.spawnAccum >= 1.0 {
		ba.spawnDebris()
		ba.spawnAccum -= 1.0
	}
}

// SetVelocity sets the ship velocity for debris flow effects.
func (ba *BubbleArc) SetVelocity(v float64) {
	ba.velocity = v
	// New particles will spawn with updated velocity
	// Existing particles keep their current trajectory
}

// Draw renders the bubble arc and debris.
func (ba *BubbleArc) Draw(screen *ebiten.Image) {
	// Draw debris first (behind the arc)
	ba.drawDebris(screen)

	// Draw arc edge (dome boundary curving up from sides)
	ba.drawArcEdge(screen)
}

// drawArcEdge draws an ephemeral, shimmering dome boundary.
// The arc is NOT solid - it's a subtle energy field effect.
// Shimmer originates from a horizontal line at center height, flows UP
// following the dome curve, and fades as it goes "behind" the viewer.
func (ba *BubbleArc) drawArcEdge(screen *ebiten.Image) {
	cx := float64(ba.screenW) / 2

	// Origin height - middle of the space area (where shimmer starts as horizontal line)
	originY := float64(ba.screenH) * 0.35
	// Top of visible area (where shimmer goes "behind" viewer)
	topY := float64(ba.screenH) * 0.05

	// Phase for animation - negative makes it flow upward
	flowPhase := -ba.time * 3.0
	pulsePhase := ba.time * 1.0

	// Draw ethereal arc - sparse points with shimmer
	steps := 400
	for i := 0; i <= steps; i++ {
		t := float64(i) / float64(steps)

		// Angle from -90° (left) to +90° (right)
		angle := (t*180 - 90) * math.Pi / 180

		x := cx + ba.arcRadius*math.Sin(angle)
		y := ba.arcCenterY - ba.arcRadius*math.Cos(angle)

		// Only draw if on screen (upper portion)
		if y < 0 || y > float64(ba.screenH)*0.55 || x < -10 || x > float64(ba.screenW)+10 {
			continue
		}

		// Flow parameter based on Y position:
		// 0 = at origin height (horizontal line in middle)
		// 1 = at top (going "behind" viewer)
		// Also includes points below origin that haven't reached it yet
		flowProgress := (originY - y) / (originY - topY)
		if flowProgress < 0 {
			flowProgress = 0 // Below origin - not yet in flow
		}
		if flowProgress > 1 {
			flowProgress = 1
		}

		// Horizontal distance from center (for spreading effect)
		horizDist := math.Abs(x-cx) / (float64(ba.screenW) * 0.5)

		// Shimmer waves that travel UP along the dome
		// Use flowProgress as the spatial component so waves flow upward
		wave1 := math.Sin(flowPhase + flowProgress*15 + horizDist*3)
		wave2 := math.Sin(flowPhase*1.3 + flowProgress*10 + horizDist*2)
		wave3 := math.Sin(pulsePhase + flowProgress*6)

		// Combine waves - only draw when waves align (sparse)
		combined := (wave1 + wave2 + wave3) / 3.0

		// Intensity fades as it goes "behind" viewer (higher flowProgress)
		// Also stronger at origin, fading both ways
		heightFade := 1.0 - flowProgress*0.7
		// Fade out at very bottom too (before origin)
		if y > originY {
			belowOrigin := (y - originY) / (float64(ba.screenH)*0.55 - originY)
			heightFade = 1.0 - belowOrigin*0.8
		}

		// Most of the arc is invisible - only show bright spots
		if combined > 0.25 {
			intensity := (combined - 0.25) / 0.75 * heightFade

			// Color shifts - cooler at origin, warmer/whiter as it rises
			warmth := flowProgress * 0.4

			r := uint8(80 + intensity*100 + warmth*60)
			g := uint8(120 + intensity*100 + warmth*40)
			b := uint8(220 + intensity*35 - warmth*30)
			a := uint8(intensity * 180)

			// Draw small glowing point
			screen.Set(int(x), int(y), color.RGBA{r, g, b, a})

			// Add glow around bright spots
			if intensity > 0.4 {
				glowA := uint8(intensity * 60)
				glowC := color.RGBA{r / 2, g / 2, b, glowA}
				screen.Set(int(x)-1, int(y), glowC)
				screen.Set(int(x)+1, int(y), glowC)
				screen.Set(int(x), int(y)-1, glowC)
				screen.Set(int(x), int(y)+1, glowC)
			}
		}
	}

	// Add occasional bright "energy nodes" along the arc
	nodeCount := 8
	for n := 0; n < nodeCount; n++ {
		nodeT := float64(n) / float64(nodeCount)
		nodePhase := math.Sin(flowPhase*2.0 + nodeT*6.28)

		if nodePhase > 0.7 {
			angle := (nodeT*180 - 90) * math.Pi / 180
			x := cx + ba.arcRadius*math.Sin(angle)
			y := ba.arcCenterY - ba.arcRadius*math.Cos(angle)

			if y >= 0 && y < float64(ba.screenH)*0.55 && x >= 0 && x < float64(ba.screenW) {
				brightness := (nodePhase - 0.7) / 0.3
				nodeColor := color.RGBA{
					uint8(150 + brightness*100),
					uint8(180 + brightness*75),
					255,
					uint8(brightness * 200),
				}

				// Draw small glowing node
				for dx := -2; dx <= 2; dx++ {
					for dy := -2; dy <= 2; dy++ {
						dist := dx*dx + dy*dy
						if dist <= 4 {
							fade := 1.0 - float64(dist)/5.0
							fadeColor := color.RGBA{
								nodeColor.R,
								nodeColor.G,
								nodeColor.B,
								uint8(float64(nodeColor.A) * fade),
							}
							px, py := int(x)+dx, int(y)+dy
							if px >= 0 && px < ba.screenW && py >= 0 && py < ba.screenH {
								screen.Set(px, py, fadeColor)
							}
						}
					}
				}
			}
		}
	}
}

// drawDebris renders the debris particles deflecting around the ship.
func (ba *BubbleArc) drawDebris(screen *ebiten.Image) {
	for _, p := range ba.debris {
		// Only draw in the space area (above the disc)
		// Disc is at ~65% down, so draw above that
		if p.Y > float64(ba.screenH)*0.55 {
			continue
		}

		alpha := uint8(p.Brightness * 255)
		c := color.RGBA{ba.debrisColor.R, ba.debrisColor.G, ba.debrisColor.B, alpha}

		ix, iy := int(p.X), int(p.Y)

		// Draw streak (motion blur) in direction of motion
		if p.Streak > 2 && ba.velocity > 0.05 {
			// Normalize velocity for streak direction
			speed := math.Sqrt(p.VX*p.VX + p.VY*p.VY)
			if speed > 0 {
				dirX := -p.VX / speed // Opposite direction (trail behind)
				dirY := -p.VY / speed

				streakLen := int(p.Streak)
				for s := 0; s < streakLen; s++ {
					// Fade along streak
					t := float64(s) / float64(streakLen)
					streakAlpha := uint8(float64(alpha) * (1.0 - t))
					streakColor := color.RGBA{c.R, c.G, c.B, streakAlpha}

					sx := ix + int(dirX*float64(s))
					sy := iy + int(dirY*float64(s))
					if sx >= 0 && sx < ba.screenW && sy >= 0 && sy < ba.screenH {
						screen.Set(sx, sy, streakColor)
					}
				}
			}
		}

		// Draw core
		if p.Size <= 1.5 {
			if ix >= 0 && ix < ba.screenW && iy >= 0 && iy < ba.screenH {
				screen.Set(ix, iy, c)
			}
		} else {
			// Larger debris
			size := int(p.Size)
			for ddx := -size; ddx <= size; ddx++ {
				for ddy := -size; ddy <= size; ddy++ {
					if ddx*ddx+ddy*ddy <= size*size {
						px, py := ix+ddx, iy+ddy
						if px >= 0 && px < ba.screenW && py >= 0 && py < ba.screenH {
							screen.Set(px, py, c)
						}
					}
				}
			}
		}
	}
}

// Resize updates for new screen dimensions.
func (ba *BubbleArc) Resize(screenW, screenH int) {
	ba.screenW = screenW
	ba.screenH = screenH
	ba.arcCenterY = float64(screenH) + 400
	ba.arcRadius = float64(screenH) + 300
}
