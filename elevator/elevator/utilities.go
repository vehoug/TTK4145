package elevator

import (
	"elevator/config"
<<<<<<< Updated upstream
)

type State struct {
	Obstructed        bool
	ActiveStatus      bool
	CurrrentBehaviour CurrentBehaviour
	CurrentFloor      int
	Direction         Direction
=======
	"time"
)

type State struct {
	Obstructed       bool
	ActiveStatus     bool
	CurrentBehaviour CurrentBehaviour
	CurrentFloor     int
	Direction        Direction
>>>>>>> Stashed changes
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
<<<<<<< Updated upstream
=======

func startMotorTimer(motorTimer *time.Timer, timeout time.Duration) <-chan time.Time {
	motorTimer.Reset(timeout)
	return motorTimer.C
}

func stopMotorTimer(motorTimer *time.Timer) <-chan time.Time {
	motorTimer.Stop()
	return nil
}
>>>>>>> Stashed changes
