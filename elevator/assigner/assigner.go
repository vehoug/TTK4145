package assigner

import (
	"elevator/config"
	"elevator/distributor"
	"elevator/elevcontrol"
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"time"
)

type HRAState struct {
	Behaviour   string                 `json:"behaviour"`
	Floor       int                    `json:"floor"`
	Direction   string                 `json:"direction"`
	CabRequests [config.NumFloors]bool `json:"cabRequests"`
}

type HRAInput struct {
	HallRequests [config.NumFloors][config.NumDirections]bool `json:"hallRequests"`
	States       map[string]HRAState                          `json:"states"`
}

func CalculateOptimalOrders(commonState distributor.CommonState, id int) elevcontrol.Orders {
	stateMap := make(map[string]HRAState)
	for elev, localState := range commonState.LocalStates {
		if commonState.PeerSyncStatus[elev] == distributor.Unavailable || !localState.State.IsActive || localState.State.Obstructed {
			continue
		}
		stateMap[strconv.Itoa(elev)] = HRAState{
			Behaviour:   localState.State.CurrentBehaviour.BehaviorToString(),
			Floor:       localState.State.CurrentFloor,
			Direction:   localState.State.Direction.DirectionToString(),
			CabRequests: localState.CabRequests,
		}
	}

	if len(stateMap) == 0 {
		return elevcontrol.Orders{}
	}

	hraInput := HRAInput{commonState.HallRequests, stateMap}

	hraExecutable := ""
	switch runtime.GOOS {
	case "linux":
		hraExecutable = "hall_request_assigner"

	case "darwin":
		hraExecutable = "hall_request_assigner_mac"

	case "windows":
		hraExecutable = "hall_request_assigner.exe"

	default:
		panic("OS not supported")
	}

	jsonBytes, err := json.Marshal(hraInput)
	if err != nil {
		fmt.Println("json.Marshal error: ", err)
		panic("json.Marshal error")
	}

	outputBytes, err := exec.Command("assigner/executables/" + hraExecutable, "-i", "--includeCab", string(jsonBytes)).CombinedOutput()
	if err != nil {
		fmt.Printf("[%v][Assigner]: exec.Command error: %v", time.Now().Format(time.TimeOnly), err)
		fmt.Println(string(outputBytes))
		panic("exec.Command error")
	}

	output := new(map[string]elevcontrol.Orders)
	err = json.Unmarshal(outputBytes, &output)
	if err != nil {
		fmt.Printf("[%v][Assigner]: json.Unmarshal error: %v", time.Now().Format(time.TimeOnly), err)
		panic("json.Unmarshal error")
	}

	return (*output)[strconv.Itoa(id)]
}
