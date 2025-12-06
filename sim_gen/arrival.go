// MANUAL WORKAROUND - Generated code failed, hand-written until codegen fixed
// TODO: Remove when ailang codegen for arrival.ail works
package sim_gen

// InitArrival creates the initial arrival state
func InitArrival() *ArrivalState {
	return &ArrivalState{
		Phase:         *NewArrivalPhasePhaseEmergence(),
		PhaseTime:     0.0,
		TotalTime:     0.0,
		Velocity:      0.99,
		CurrentPlanet: *NewCurrentPlanetNoPlanet(),
		ShowHUD:       false,
		ShipTimeYears: 47.3,
		GalaxyYear:    2157,
	}
}

// phaseTargetVelocity returns target velocity for phase
func phaseTargetVelocity(phase *ArrivalPhase) float64 {
	switch phase.Kind {
	case ArrivalPhaseKindPhaseEmergence:
		return 0.99
	case ArrivalPhaseKindPhaseStabilizing:
		return 0.95
	case ArrivalPhaseKindPhaseSaturn:
		return 0.9
	case ArrivalPhaseKindPhaseJupiter:
		return 0.8
	case ArrivalPhaseKindPhaseMars:
		return 0.5
	case ArrivalPhaseKindPhaseEarth, ArrivalPhaseKindPhaseBridge, ArrivalPhaseKindPhaseComplete:
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
	case ArrivalPhaseKindPhaseEarth, ArrivalPhaseKindPhaseBridge:
		return NewCurrentPlanetEarth()
	default:
		return NewCurrentPlanetNoPlanet()
	}
}

// nextPhase returns the next phase
func nextPhase(phase *ArrivalPhase) *ArrivalPhase {
	switch phase.Kind {
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
	case ArrivalPhaseKindPhaseEarth:
		return NewArrivalPhasePhaseBridge()
	case ArrivalPhaseKindPhaseBridge, ArrivalPhaseKindPhaseComplete:
		return NewArrivalPhasePhaseComplete()
	default:
		return NewArrivalPhasePhaseComplete()
	}
}

// shouldTransition checks if phase should transition
func shouldTransition(phase *ArrivalPhase, phaseTime float64) bool {
	switch phase.Kind {
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
	case ArrivalPhaseKindPhaseBridge:
		return phaseTime > 3.0
	case ArrivalPhaseKindPhaseComplete:
		return false
	default:
		return false
	}
}

// IsArrivalComplete checks if arrival sequence is done
func IsArrivalComplete(state *ArrivalState) bool {
	return state.Phase.Kind == ArrivalPhaseKindPhaseComplete
}

// GetArrivalVelocity returns SR velocity for shader
func GetArrivalVelocity(state *ArrivalState) float64 {
	return state.Velocity
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
	case ArrivalPhaseKindPhaseBridge:
		return "bridge"
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

	if shouldTransition(&state.Phase, newPhaseTime) {
		nextPh := nextPhase(&state.Phase)
		nextVel := phaseTargetVelocity(nextPh)
		nextPlanet := phasePlanet(nextPh)
		return &ArrivalState{
			Phase:         *nextPh,
			PhaseTime:     0.0,
			TotalTime:     newTotalTime,
			Velocity:      nextVel,
			CurrentPlanet: *nextPlanet,
			ShowHUD:       true,
			ShipTimeYears: state.ShipTimeYears,
			GalaxyYear:    state.GalaxyYear,
		}
	}

	// No transition, just update times
	return &ArrivalState{
		Phase:         state.Phase,
		PhaseTime:     newPhaseTime,
		TotalTime:     newTotalTime,
		Velocity:      state.Velocity,
		CurrentPlanet: state.CurrentPlanet,
		ShowHUD:       state.ShowHUD,
		ShipTimeYears: state.ShipTimeYears,
		GalaxyYear:    state.GalaxyYear,
	}
}
