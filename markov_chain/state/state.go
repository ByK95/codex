package state

type State uint

type StateMachine struct {
	current     State
	transitions map[State]map[State]uint
}

func NewStateMachine() *StateMachine {
	return &StateMachine{
		current:     0,
		transitions: make(map[State]map[State]uint),
	}
}

func (sm *StateMachine) AddBothDirections(fromState, toState State) {
	sm.Add(fromState, toState)
	sm.Add(toState, fromState)
}

func (sm *StateMachine) Add(fromState, toState State) {
	if sm.transitions[fromState] == nil {
		sm.transitions[fromState] = make(map[State]uint)
	}

	if _, ok := sm.transitions[fromState][toState]; ok {
		sm.transitions[fromState][toState]++
	} else {
		sm.transitions[fromState][toState] = 1
	}
}

func (sm *StateMachine) GetNextStates() map[State]uint {
	return sm.transitions[sm.current]
}

func (sm *StateMachine) Next(toState State) bool {
	nextStates, ok := sm.transitions[sm.current]

	if !ok {
		return false
	}

	_, allowed := nextStates[toState]

	if allowed {
		sm.current = toState
		return true
	}

	return false
}