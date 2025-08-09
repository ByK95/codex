package state

import (
	"math/rand"
	"time"
)

type ProbabilityMachine struct {
	current     State
	transitions map[State]map[State]float64
}

func NewProbabilityMachineFrom(sm *StateMachine, restartBias float64) *ProbabilityMachine {
	pm := &ProbabilityMachine{
		current:     sm.current,
		transitions: make(map[State]map[State]float64),
	}

	for from, toMap := range sm.transitions {
		total := uint(0)
		for _, count := range toMap {
			total += count
		}

		base := make(map[State]float64)
		for to, count := range toMap {
			base[to] = float64(count) / float64(total)
		}

		// Apply restart bias
		withBias := make(map[State]float64)
		for to, prob := range base {
			withBias[to] = prob * (1 - restartBias)
		}
		withBias[sm.current] += restartBias // bias to return

		pm.transitions[from] = withBias
	}

	return pm
}

func (pm *ProbabilityMachine) GetNextStates() map[State]float64 {
	return pm.transitions[pm.current]
}

func (pm *ProbabilityMachine) NextState() (State, bool) {
	nextStates := pm.transitions[pm.current]
	if len(nextStates) == 0 {
		return 0, false
	}

	r := rand.Float64() // random between 0 and 1
	sum := 0.0

	for state, prob := range nextStates {
		sum += prob
		if r <= sum {
			pm.current = state
			return state, true
		}
	}

	// // fallback (due to float precision)
	// for state := range nextStates {
	// 	pm.current = state
	// 	return state, true
	// }

	return 0, false
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
