package distributor

import (
	"elevator/config"
	"elevator/elevcontrol"
	"elevator/elevio"
	"elevator/network/peers"
	"reflect"
	"strconv"
)

type SyncStatus int

const (
	Synced SyncStatus = iota
	Pending
	Unavailable
)

type LocalState struct {
	State       elevcontrol.State
	CabRequests [config.NumFloors]bool
}

type CommonState struct {
	Version        uint64
	UpdaterID      int
	PeerSyncStatus [config.NumElevators]SyncStatus
	HallRequests   [config.NumFloors][config.NumDirections]bool
	LocalStates    [config.NumElevators]LocalState
}

func (commonState *CommonState) initCommonState(id int) {
	for elev := range commonState.PeerSyncStatus {
		commonState.PeerSyncStatus[elev] = Unavailable
	}
	commonState.PeerSyncStatus[id] = Synced

	commonState.LocalStates[id].State.IsActive = true
}

func (commonState *CommonState) addOrder(newOrder elevio.ButtonEvent, id int) {
	if newOrder.Button == elevio.BT_Cab {
		commonState.LocalStates[id].CabRequests[newOrder.Floor] = true
	} else {
		commonState.HallRequests[newOrder.Floor][newOrder.Button] = true
	}
}

func (commonState *CommonState) removeOrder(deliveredOrder elevio.ButtonEvent, id int) {
	if deliveredOrder.Button == elevio.BT_Cab {
		commonState.LocalStates[id].CabRequests[deliveredOrder.Floor] = false
	} else {
		commonState.HallRequests[deliveredOrder.Floor][deliveredOrder.Button] = false
	}
}

func (commonState *CommonState) updateState(newState elevcontrol.State, id int) {
	commonState.LocalStates[id] = LocalState{
		State:       newState,
		CabRequests: commonState.LocalStates[id].CabRequests,
	}
}

func (commonState *CommonState) mergeCommonStates(arrivedCommonState CommonState, id int) {
	for floor := range config.NumFloors {
		for direction := range config.NumDirections {
			commonState.HallRequests[floor][direction] =
				arrivedCommonState.HallRequests[floor][direction] || commonState.HallRequests[floor][direction]
		}
	}
	arrivedCommonState.LocalStates[id].CabRequests = commonState.LocalStates[id].CabRequests
}

func (commonState *CommonState) makeInactivePeersUnavailable(activePeers peers.PeerUpdate) {
	activeSet := make(map[int]bool, len(activePeers.Peers))
	for _, idStr := range activePeers.Peers {
		id, _ := strconv.Atoi(idStr)
		activeSet[id] = true
	}

	for elev := range commonState.PeerSyncStatus {
		if !activeSet[elev] {
			commonState.PeerSyncStatus[elev] = Unavailable
		}
	}
}

func (commonState *CommonState) makeOthersUnavailable(id int) {
	for elev := range commonState.PeerSyncStatus {
		if elev != id {
			commonState.PeerSyncStatus[elev] = Unavailable
		}
	}
}

func (commonState *CommonState) prepNewCommonState(id int) {
	commonState.Version++
	commonState.UpdaterID = id
	for elev := range commonState.PeerSyncStatus {
		if commonState.PeerSyncStatus[elev] == Synced {
			commonState.PeerSyncStatus[elev] = Pending
		}
	}
}

func (commonState CommonState) isNewerThan(otherCommonState CommonState) bool {
	if commonState.Version != otherCommonState.Version {
		return commonState.Version > otherCommonState.Version
	}
	return commonState.UpdaterID > otherCommonState.UpdaterID
}

func (commonState CommonState) isOlderThan(otherCommonState CommonState) bool {
	if commonState.Version != otherCommonState.Version {
		return commonState.Version < otherCommonState.Version
	}
	return commonState.UpdaterID < otherCommonState.UpdaterID
}

func (commonState CommonState) fullySynced(id int) bool {
	if commonState.PeerSyncStatus[id] == Unavailable {
		return false
	}
	for elev := range commonState.PeerSyncStatus {
		if commonState.PeerSyncStatus[elev] == Pending {
			return false
		}
	}
	return true
}

func (commonState CommonState) equals(arrivedCommonState CommonState) bool {
	commonState.PeerSyncStatus        = [config.NumElevators]SyncStatus{}
	arrivedCommonState.PeerSyncStatus = [config.NumElevators]SyncStatus{}
	return reflect.DeepEqual(commonState, arrivedCommonState)
}
