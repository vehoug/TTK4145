package fsm

import (
	"elevator-go/elevator"
	"elevator-go/elevio"
	"elevator-go/requests"
	"elevator-go/timer"
)

var elev elevator.Elevator

func Init() {
	elev = elevator.ElevatorInit()
}

func OnRequestButtonPress(btnFloor int, btnType elevio.ButtonType) {

	switch elev.Behaviour {
	case elevator.EB_DoorOpen:
		if elev.Floor == btnFloor {
			timer.Start(3.0)
			return
		}
	}

	elev.Requests[btnFloor][btnType] = true
	elevio.SetButtonLamp(btnType, btnFloor, true)

	switch elev.Behaviour {
	case elevator.EB_Idle:
		elev.Dir = requests.ChooseDirection(elev)
		elevio.SetMotorDirection(elev.Dir)
		if elev.Dir == elevio.MD_Stop {
			elev.Behaviour = elevator.EB_DoorOpen
			elevio.SetDoorOpenLamp(true)
			timer.Start(3.0)
			elev.Requests[btnFloor][btnType] = false
			elevio.SetButtonLamp(btnType, btnFloor, false)
		} else {
			elev.Behaviour = elevator.EB_Moving
		}
	}
}

func OnFloorArrival(newFloor int) {
	elev.Floor = newFloor
	elevio.SetFloorIndicator(newFloor)

	switch elev.Behaviour {
	case elevator.EB_Moving:
		if requests.ShouldStop(elev) {
			elevio.SetMotorDirection(elevio.MD_Stop)
			elevio.SetDoorOpenLamp(true)

			elev.Requests[newFloor][elevio.BT_Cab] = false
			elevio.SetButtonLamp(elevio.BT_Cab, newFloor, false)

			switch elev.Dir {
			case elevio.MD_Up:
				elev.Requests[newFloor][elevio.BT_HallUp] = false
				elevio.SetButtonLamp(elevio.BT_HallUp, newFloor, false)
				if !requests.HasRequestsAbove(elev) {
					elev.Requests[newFloor][elevio.BT_HallDown] = false
					elevio.SetButtonLamp(elevio.BT_HallDown, newFloor, false)
				}
			case elevio.MD_Down:
				elev.Requests[newFloor][elevio.BT_HallDown] = false
				elevio.SetButtonLamp(elevio.BT_HallDown, newFloor, false)
				if !requests.HasRequestsBelow(elev) {
					elev.Requests[newFloor][elevio.BT_HallUp] = false
					elevio.SetButtonLamp(elevio.BT_HallUp, newFloor, false)
				}
			case elevio.MD_Stop:
				elev.Requests[newFloor][elevio.BT_HallUp] = false
				elevio.SetButtonLamp(elevio.BT_HallUp, newFloor, false)
				elev.Requests[newFloor][elevio.BT_HallDown] = false
				elevio.SetButtonLamp(elevio.BT_HallDown, newFloor, false)
			}

			elev.Behaviour = elevator.EB_DoorOpen
			timer.Start(3.0)
		}
	}
}

func OnDoorTimeout() {
	switch elev.Behaviour {
	case elevator.EB_DoorOpen:
		elevio.SetDoorOpenLamp(false)
		elev.Dir = requests.ChooseDirection(elev)
		elevio.SetMotorDirection(elev.Dir)

		if elev.Dir == elevio.MD_Stop {
			elev.Behaviour = elevator.EB_Idle
		} else {
			elev.Behaviour = elevator.EB_Moving
		}
	}
}
