package distributor

import (
	"elevator/config"
	"elevator/elevator"
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
	State       elevator.State
	CabRequests [config.NumFloors]bool
}

type CommonState struct {
	Version        uint64
	UpdaterID      int
	PeerSyncStatus [config.NumElevators]SyncStatus
	HallRequests   [config.NumFloors][2]bool
	LocalStates    [config.NumElevators]LocalState
}

func (commonState *CommonState) addOrder(newOrder elevio.ButtonEvent, id int) {
	if newOrder.Button == elevio.BT_Cab {
		commonState.LocalStates[id].CabRequests[newOrder.Floor] = true
	} else {
		commonState.HallRequests[newOrder.Floor][newOrder.Button] = true
	}
}

func (commonState *CommonState) addCabCall(newOrder elevio.ButtonEvent, id int) {
	if newOrder.Button == elevio.BT_Cab {
		commonState.LocalStates[id].CabRequests[newOrder.Floor] = true
	}
}

func (commonState *CommonState) removeOrder(deliveredOrder elevio.ButtonEvent, id int) {
	if deliveredOrder.Button == elevio.BT_Cab {
		commonState.LocalStates[id].CabRequests[deliveredOrder.Floor] = false
	} else {
		commonState.HallRequests[deliveredOrder.Floor][deliveredOrder.Button] = false
	}
}

func (commonState *CommonState) updateState(newState elevator.State, id int) {
	commonState.LocalStates[id] = LocalState{
		State:       newState,
		CabRequests: commonState.LocalStates[id].CabRequests,
	}
}

func (commonState *CommonState) fullySynced(id int) bool {
	if commonState.PeerSyncStatus[id] == Unavailable {
		return false
	}
	for index := range commonState.PeerSyncStatus {
		if commonState.PeerSyncStatus[index] == Pending {
			return false
		}
	}
	return true
}

func (commonState CommonState) equals(arrivedCommonState CommonState) bool {
	commonState.PeerSyncStatus = [config.NumElevators]SyncStatus{}
	arrivedCommonState.PeerSyncStatus = [config.NumElevators]SyncStatus{}
	return reflect.DeepEqual(commonState, arrivedCommonState)
}

func (commonState *CommonState) makeLostPeersUnavailable(peers peers.PeerUpdate) {
	for _, idStr := range peers.Lost {
		id, _ := strconv.Atoi(idStr)
		commonState.PeerSyncStatus[id] = Unavailable
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
	for node := range commonState.PeerSyncStatus {
		if commonState.PeerSyncStatus[node] == Synced {
			commonState.PeerSyncStatus[node] = Pending
		}
	}
}

func (commonState CommonState) isNewerThan(otherCommonState CommonState) bool {
	if commonState.Version != otherCommonState.Version {
		return commonState.Version > otherCommonState.Version
	}
	return commonState.UpdaterID > otherCommonState.UpdaterID
}

func (commonState *CommonState) applyTransaction(mutation func(), id int) {
	commonState.prepNewCommonState(id)
	mutation()
	commonState.PeerSyncStatus[id] = Synced
}

func (commonState CommonState) isOlderThan(otherCommonState CommonState) bool {
	return commonState.Version < otherCommonState.Version
}
