package requests

import (
	"elevator-go/elevator"
	"elevator-go/elevio"
)

// Check if there are any requests above the current floor
func HasRequestsAbove(e elevator.Elevator) bool {
	for f := e.Floor + 1; f < 4; f++ {
		for b := 0; b < 3; b++ {
			if e.Requests[f][b] {
				return true
			}
		}
	}
	return false
}

// Check if there are any requests below the current floor
func HasRequestsBelow(e elevator.Elevator) bool {
	for f := 0; f < e.Floor; f++ {
		for b := 0; b < 3; b++ {
			if e.Requests[f][b] {
				return true
			}
		}
	}
	return false
}

// Decides the next direction of travel
func ChooseDirection(e elevator.Elevator) elevio.MotorDirection {
	switch e.Dir {
	case elevio.MD_Up:
		if HasRequestsAbove(e) {
			return elevio.MD_Up
		}
		if HasRequestsBelow(e) {
			return elevio.MD_Down
		}
	case elevio.MD_Down, elevio.MD_Stop:
		if HasRequestsBelow(e) {
			return elevio.MD_Down
		}
		if HasRequestsAbove(e) {
			return elevio.MD_Up
		}
	}
	return elevio.MD_Stop
}

// Logic for deciding if the elevator should stop at the current floor
func ShouldStop(e elevator.Elevator) bool {
	switch e.Dir {
	case elevio.MD_Down:
		return e.Requests[e.Floor][elevio.BT_HallDown] || e.Requests[e.Floor][elevio.BT_Cab] || !HasRequestsBelow(e)
	case elevio.MD_Up:
		return e.Requests[e.Floor][elevio.BT_HallUp] || e.Requests[e.Floor][elevio.BT_Cab] || !HasRequestsAbove(e)
	default:
		return true
	}
}
