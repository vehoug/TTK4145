package assigner

import (
	"elevator/config"
	"elevator/distributor"
	"elevator/elevator"
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
)

type HRAState struct {
	Behaviour   string                 `json:"behaviour"`
	Floor       int                    `json:"floor"`
	Direction   string                 `json:"direction"`
	CabRequests [config.NumFloors]bool `json:"cabRequests"`
}

type HRAInput struct {
	HallRequests [config.NumFloors][2]bool `json:"hallRequests"`
	States       map[string]HRAState       `json:"states"`
}

func CalculateOptimalOrders(commonState distributor.CommonState, id int) elevator.Orders {

	stateMap := make(map[string]HRAState)
	for nodeID, node := range commonState.LocalStates {
		if commonState.PeerSyncStatus[nodeID] == distributor.Unavailable || !node.State.ActiveStatus || node.State.Obstructed {
			continue
		} else {
			stateMap[strconv.Itoa(nodeID)] = HRAState{
				Behaviour:   node.State.CurrentBehaviour.BehaviourToString(),
				Floor:       node.State.CurrentFloor,
				Direction:   node.State.Direction.DirectionToString(),
				CabRequests: node.CabRequests,
			}
		}
	}

	if len(stateMap) == 0 {
		return elevator.Orders{}
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

	ret, err := exec.Command("assigner/executables/"+hraExecutable, "-i", "--includeCab", string(jsonBytes)).CombinedOutput()
	if err != nil {
		fmt.Println("exec.Command error: ", err)
		fmt.Println(string(ret))
		panic("exec.Command error")
	}

	output := new(map[string]elevator.Orders)
	err = json.Unmarshal(ret, &output)
	if err != nil {
		fmt.Println("json.Unmarshal error: ", err)
		panic("json.Unmarshal error")
	}

	return (*output)[strconv.Itoa(id)]
}
