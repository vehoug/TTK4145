package lights

import (
	"elevator/config"
	"elevator/elevio"
)

func SetLights(commonState elevator.CommonState, ElevatorID int) {
	for floor := 0; floor < config.NumFloors; floor++ {
		for buttonType := 0; buttonType < 2; buttonType++ {
			if commonState.HallRequests[floor][buttonType] {
				elevio.SetButtonLamp(elevio.ButtonType(buttonType), floor, true)
			} else {
				elevio.SetButtonLamp(elevio.ButtonType(buttonType), floor, false)
			}
		}
	}
	for floor := 0; floor < config.NumFloors; floor++ {
		if commonState.States[ElevatorID].CabRequests[floor] {
			elevio.SetButtonLamp(elevio.BT_Cab, floor, true)
		} else {
			elevio.SetButtonLamp(elevio.BT_Cab, floor, false)
		}
	}
}
