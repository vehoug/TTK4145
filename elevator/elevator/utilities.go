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

func startMotorTimer(motorTimer *time.Timer, timeout time.Duration) <-chan time.Time {
	motorTimer.Reset(timeout)
	return motorTimer.C
}

func stopMotorTimer(motorTimer *time.Timer) <-chan time.Time {
	motorTimer.Stop()
	return nil
}

func (behaviour CurrentBehaviour) ToString() string {
	return map[CurrentBehaviour]string{Idle: "idle", DoorOpen: "doorOpen", Moving: "moving"}[behaviour]
}
