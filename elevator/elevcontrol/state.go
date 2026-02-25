package elevcontrol

import "elevator/config"

type State struct {
	Obstructed       bool
	Active           bool
	CurrentBehaviour CurrentBehaviour
	CurrentFloor     int
	Direction        Direction
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
		Active:           true,
	}
}

func (behaviour CurrentBehaviour) BehaviorToString() string {
	return map[CurrentBehaviour]string{
		Idle:     "idle",
		DoorOpen: "doorOpen",
		Moving:   "moving",
	}[behaviour]
}
