// MANUAL WORKAROUND - Generated code failed, hand-written until codegen fixed
// TODO: Remove when ailang codegen for arrival.ail works
package sim_gen

// InitArrival creates the initial arrival state (starts inside black hole)
func InitArrival() *ArrivalState {
	return &ArrivalState{
		Phase:         *NewArrivalPhasePhaseBlackHole(),
		PhaseTime:     0.0,
		TotalTime:     0.0,
		Velocity:      0.5,       // Artistic license (real would be 0.99c)
		GrIntensity:   1.0,       // Maximum GR effects (inside black hole)
		CurrentPlanet: *NewCurrentPlanetNoPlanet(),
		ShipTimeYears: 47.3,
		GalaxyYear:    2157,
	}
}

// phaseTargetVelocity returns target velocity for phase
// NOTE: Artistic license applied - real physics would use 0.9c+ but that causes
// SR blueshift to wash out all visual content. These values maintain visual
// progression while keeping the SR effect visible but not overwhelming.
func phaseTargetVelocity(phase *ArrivalPhase) float64 {
	switch phase.Kind {
	case ArrivalPhaseKindPhaseBlackHole:
		return 0.5 // Ejected at high speed (artistic: real would be ~0.99c)
	case ArrivalPhaseKindPhaseEmergence:
		return 0.45
	case ArrivalPhaseKindPhaseStabilizing:
		return 0.35
	case ArrivalPhaseKindPhaseSaturn:
		return 0.3 // Visible blueshift but content still clear
	case ArrivalPhaseKindPhaseJupiter:
		return 0.2
	case ArrivalPhaseKindPhaseMars:
		return 0.1
	case ArrivalPhaseKindPhaseEarth, ArrivalPhaseKindPhaseComplete:
		return 0.0
	default:
		return 0.0
	}
}

// phasePlanet returns planet for phase
func phasePlanet(phase *ArrivalPhase) *CurrentPlanet {
	switch phase.Kind {
	case ArrivalPhaseKindPhaseSaturn:
		return NewCurrentPlanetSaturn()
	case ArrivalPhaseKindPhaseJupiter:
		return NewCurrentPlanetJupiter()
	case ArrivalPhaseKindPhaseMars:
		return NewCurrentPlanetMars()
	case ArrivalPhaseKindPhaseEarth:
		return NewCurrentPlanetEarth()
	default:
		return NewCurrentPlanetNoPlanet()
	}
}

// nextPhase returns the next phase
func nextPhase(phase *ArrivalPhase) *ArrivalPhase {
	switch phase.Kind {
	case ArrivalPhaseKindPhaseBlackHole:
		return NewArrivalPhasePhaseEmergence() // Exit black hole → tumbling
	case ArrivalPhaseKindPhaseEmergence:
		return NewArrivalPhasePhaseStabilizing()
	case ArrivalPhaseKindPhaseStabilizing:
		return NewArrivalPhasePhaseSaturn()
	case ArrivalPhaseKindPhaseSaturn:
		return NewArrivalPhasePhaseJupiter()
	case ArrivalPhaseKindPhaseJupiter:
		return NewArrivalPhasePhaseMars()
	case ArrivalPhaseKindPhaseMars:
		return NewArrivalPhasePhaseEarth()
	case ArrivalPhaseKindPhaseEarth, ArrivalPhaseKindPhaseComplete:
		return NewArrivalPhasePhaseComplete()
	default:
		return NewArrivalPhasePhaseComplete()
	}
}

// shouldTransition checks if phase should transition
func shouldTransition(phase *ArrivalPhase, phaseTime float64) bool {
	switch phase.Kind {
	case ArrivalPhaseKindPhaseBlackHole:
		return phaseTime > 8.0 // 8 seconds in black hole
	case ArrivalPhaseKindPhaseEmergence:
		return phaseTime > 5.0
	case ArrivalPhaseKindPhaseStabilizing:
		return phaseTime > 3.0
	case ArrivalPhaseKindPhaseSaturn:
		return phaseTime > 5.0
	case ArrivalPhaseKindPhaseJupiter:
		return phaseTime > 5.0
	case ArrivalPhaseKindPhaseMars:
		return phaseTime > 5.0
	case ArrivalPhaseKindPhaseEarth:
		return phaseTime > 5.0
	case ArrivalPhaseKindPhaseComplete:
		return false
	default:
		return false
	}
}

// calcGRDecay calculates GR intensity decay during black hole phase
func calcGRDecay(phase *ArrivalPhase, phaseTime float64) float64 {
	if phase.Kind == ArrivalPhaseKindPhaseBlackHole {
		// GR fades from 1.0 to 0.0 over 8 seconds
		progress := phaseTime / 8.0
		if progress > 1.0 {
			return 0.0
		}
		return 1.0 - progress
	}
	return 0.0 // No GR effects after black hole
}

// IsArrivalComplete checks if arrival sequence is done
func IsArrivalComplete(state *ArrivalState) bool {
	return state.Phase.Kind == ArrivalPhaseKindPhaseComplete
}

// GetArrivalVelocity returns SR velocity for shader
func GetArrivalVelocity(state *ArrivalState) float64 {
	return state.Velocity
}

// GetGRIntensity returns GR effect intensity (0.0 = none, 1.0 = extreme)
func GetGRIntensity(state *ArrivalState) float64 {
	return state.GrIntensity
}

// GetArrivalPlanetName returns planet name for rendering
func GetArrivalPlanetName(state *ArrivalState) string {
	switch state.CurrentPlanet.Kind {
	case CurrentPlanetKindSaturn:
		return "saturn"
	case CurrentPlanetKindJupiter:
		return "jupiter"
	case CurrentPlanetKindMars:
		return "mars"
	case CurrentPlanetKindEarth:
		return "earth"
	default:
		return ""
	}
}

// GetArrivalPhaseName returns phase name for debugging
func GetArrivalPhaseName(state *ArrivalState) string {
	switch state.Phase.Kind {
	case ArrivalPhaseKindPhaseBlackHole:
		return "blackhole"
	case ArrivalPhaseKindPhaseEmergence:
		return "emergence"
	case ArrivalPhaseKindPhaseStabilizing:
		return "stabilizing"
	case ArrivalPhaseKindPhaseSaturn:
		return "saturn"
	case ArrivalPhaseKindPhaseJupiter:
		return "jupiter"
	case ArrivalPhaseKindPhaseMars:
		return "mars"
	case ArrivalPhaseKindPhaseEarth:
		return "earth"
	case ArrivalPhaseKindPhaseComplete:
		return "complete"
	default:
		return "unknown"
	}
}

// StepArrival updates arrival state for one frame
func StepArrival(state *ArrivalState, input *ArrivalInput) *ArrivalState {
	dt := input.Dt
	newPhaseTime := state.PhaseTime + dt
	newTotalTime := state.TotalTime + dt
	newGR := calcGRDecay(&state.Phase, newPhaseTime)

	if shouldTransition(&state.Phase, newPhaseTime) {
		nextPh := nextPhase(&state.Phase)
		nextVel := phaseTargetVelocity(nextPh)
		nextPlanet := phasePlanet(nextPh)
		return &ArrivalState{
			Phase:         *nextPh,
			PhaseTime:     0.0,
			TotalTime:     newTotalTime,
			Velocity:      nextVel,
			GrIntensity:   0.0, // GR effects end after black hole
			CurrentPlanet: *nextPlanet,
			ShipTimeYears: state.ShipTimeYears,
			GalaxyYear:    state.GalaxyYear,
		}
	}

	// No transition, just update times and GR
	return &ArrivalState{
		Phase:         state.Phase,
		PhaseTime:     newPhaseTime,
		TotalTime:     newTotalTime,
		Velocity:      state.Velocity,
		GrIntensity:   newGR,
		CurrentPlanet: state.CurrentPlanet,
		ShipTimeYears: state.ShipTimeYears,
		GalaxyYear:    state.GalaxyYear,
	}
}
