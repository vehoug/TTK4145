package lights

import (
	"elevator/config"
	"elevator/distributor"
	"elevator/elevio"
)

func SetLights(commonState distributor.CommonState, id int) {
	for floor := range config.NumFloors {
		for buttonType := range config.NumDirections {
			activeHallRequest := commonState.HallRequests[floor][buttonType]
			elevio.SetButtonLamp(elevio.ButtonType(buttonType), floor, activeHallRequest)
		}
	}
	for floor := range config.NumFloors {
		activeCabRequest := commonState.LocalStates[id].CabRequests[floor]
		elevio.SetButtonLamp(elevio.BT_Cab, floor, activeCabRequest)
	}
}
