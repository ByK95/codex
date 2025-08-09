package main

import (
	"example/enum"
	"example/state"
	"fmt"
)

func main() {
	e := enum.NewEnum()
	e.Add("Running")
	e.Add("Idle")
	e.Add("Jumping")
	e.Add("Attacking")
	e.Add("Patroling")

	s := state.NewStateMachine()
	s.AddBothDirections(0, 1)
	s.AddBothDirections(0, 2)
	s.AddBothDirections(0, 3)
	s.AddBothDirections(0, 4)
	s.Add(0,1)

	// s.Next(1)

	pm := state.NewProbabilityMachineFrom(s, 0.1)

	// for key, val := range pm.GetNextStates(){
	// 	fmt.Println(key, val)
	// }

	for i := 0; i < 20; i++ {
		state, _ := pm.NextState()
		key, _ :=e.GetValue(int(state))
		fmt.Println(i, state, key)
	}
	
}