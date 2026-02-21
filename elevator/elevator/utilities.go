package elevator

import (
	"elevator/config"
)

type State struct {
	Obstructed        bool
	ActiveStatus      bool
	CurrrentBehaviour CurrentBehaviour
	CurrentFloor      int
	Direction         Direction
}

type CurrentBehaviour int

const (
	Idle CurrentBehaviour = iota
	DoorOpen
	Moving
)

type Direction int

const (
	Up Direction = iota
	Down
)

type Orders [config.NumFloors][config.NumButtons]bool

func initState() State {
	return State{
		CurrentFloor:     -1,
		Direction:        Down,
		CurrentBehaviour: Moving,
		ActiveStatus:     true,
	}
}
