package elevator

import "elevator-go/elevio"

type ElevatorBehaviour int

const (
	EB_Idle ElevatorBehaviour = iota
	EB_DoorOpen
	EB_Moving
)

type Elevator struct {
	Floor     int
	Dir       elevio.MotorDirection
	Requests  [4][3]bool // [numFloors][numButtons]
	Behaviour ElevatorBehaviour
}

// Initial state of the elevator
func ElevatorInit() Elevator {
	return Elevator{
		Floor:     -1,
		Dir:       elevio.MD_Stop,
		Behaviour: EB_Idle,
	}
}
