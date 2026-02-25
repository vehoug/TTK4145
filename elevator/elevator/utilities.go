package elevator

import (
	"elevator/config"
	"time"
)

type State struct {
	Obstructed       bool
	ActiveStatus     bool
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
		ActiveStatus:     true,
	}
}

func resetTimer(timer *time.Timer, timeout time.Duration) <-chan time.Time {
	timer = time.NewTimer(timeout)
	return timer.C
}

func stopTimer(timer *time.Timer) <-chan time.Time {
	timer.Stop()
	return nil
}

func (behaviour CurrentBehaviour) BehaviourToString() string {
	return map[CurrentBehaviour]string{Idle: "idle", DoorOpen: "doorOpen", Moving: "moving"}[behaviour]
}
