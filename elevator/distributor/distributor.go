package distributor

import (
	"elevator/config"
	"elevator/elevcontrol"
	"elevator/elevio"
	"elevator/network/peers"
	"fmt"
	"time"
)

type OperationType int

const (
	None OperationType = iota
	AddOrder
	RemoveOrder
	StateUpdate
)

type PendingOperation struct {
	Type  OperationType
	Order elevio.ButtonEvent
	State elevcontrol.State
}

func Distributor(
	networkReceiveCh <-chan CommonState,
	peerUpdateCh <-chan peers.PeerUpdate,
	syncedCommonStateCh chan<- CommonState,
	networkTransmitCh chan<- CommonState,
	deliveredOrderCh <-chan elevio.ButtonEvent,
	newStateCh <-chan elevcontrol.State,
	id int,
) {
	newOrderCh := make(chan elevio.ButtonEvent, config.Buffer)
	go elevio.PollButtons(newOrderCh)

	pendingQueue := make([]PendingOperation, 0)

	var commonState CommonState
	var deliveredOrder elevio.ButtonEvent
	var newOrder elevio.ButtonEvent
	var newState elevcontrol.State
	var peersStatus peers.PeerUpdate

	commonState.initCommonState(id)

	idle := true
	offline := false

	disconnectTimer := time.NewTimer(config.DisconnectTime)
	heartbeatTicker := time.NewTicker(config.HeartbeatTime)
	defer heartbeatTicker.Stop()

	for {
		select {
		case <-disconnectTimer.C:
			commonState.makeOthersUnavailable(id)
			offline = true
			fmt.Printf("Node [%d]: lost network connection\n", id)

		case peersStatus = <-peerUpdateCh:
			commonState.makeInactivePeersUnavailable(peersStatus)
			idle = false

		case <-heartbeatTicker.C:
			networkTransmitCh <- commonState

		case newOrder = <-newOrderCh:
			if offline {
				if commonState.LocalStates[id].State.IsActive {
					commonState.PeerSyncStatus[id] = Synced
					commonState.addOrder(newOrder, id)
					syncedCommonStateCh <- commonState
				}

			} else if idle {
				commonState.applyTransaction(func() {
					commonState.addOrder(newOrder, id)
				}, id)
				idle = false

			} else {
				pendingQueue = append(pendingQueue, PendingOperation{
					Type:  AddOrder,
					Order: newOrder,
				})
			}

		case deliveredOrder = <-deliveredOrderCh:
			if offline {
				commonState.PeerSyncStatus[id] = Synced
				commonState.removeOrder(deliveredOrder, id)
				syncedCommonStateCh <- commonState

			} else if idle {
				commonState.applyTransaction(func() {
					commonState.removeOrder(deliveredOrder, id)
				}, id)
				idle = false

			} else {
				pendingQueue = append(pendingQueue, PendingOperation{
					Type:  RemoveOrder,
					Order: deliveredOrder,
				})
			}

		case newState = <-newStateCh:
			if offline {
				if newState.IsActive && !newState.Obstructed {
					commonState.PeerSyncStatus[id] = Synced
					commonState.updateState(newState, id)
					syncedCommonStateCh <- commonState
				}

			} else if idle {
				commonState.applyTransaction(func() {
					commonState.updateState(newState, id)
				}, id)
				idle = false

			} else {
				pendingQueue = append(pendingQueue, PendingOperation{
					Type:  StateUpdate,
					State: newState,
				})
			}

		case arrivedCommonState := <-networkReceiveCh:
			if offline {
				allCabCallsDone  := commonState.LocalStates[id].CabRequests == [config.NumFloors]bool{}
				allHallCallsDone := commonState.HallRequests == [config.NumFloors][config.NumDirections]bool{} 

				if allCabCallsDone && allHallCallsDone {
					commonState = arrivedCommonState
					commonState.makeInactivePeersUnavailable(peersStatus)
					commonState.PeerSyncStatus[id] = Synced
					offline = false
					idle = false
					fmt.Printf("Node [%d]: reconnected to network\n", id)
				} else {
					commonState.PeerSyncStatus[id] = Unavailable
				}

			} else if idle {
				disconnectTimer = time.NewTimer(config.DisconnectTime)
				if arrivedCommonState.isNewerThan(commonState) {
					commonState = arrivedCommonState
					commonState.makeInactivePeersUnavailable(peersStatus)
					commonState.PeerSyncStatus[id] = Synced
					idle = false
				}

			} else {
				if arrivedCommonState.isOlderThan(commonState) {
					break
				}

				disconnectTimer = time.NewTimer(config.DisconnectTime)

				switch {
				case arrivedCommonState.isNewerThan(commonState):
					commonState = arrivedCommonState
					commonState.makeInactivePeersUnavailable(peersStatus)
					commonState.PeerSyncStatus[id] = Synced

				case arrivedCommonState.fullySynced(id):
					commonState = arrivedCommonState
					syncedCommonStateCh <- commonState

					if len(pendingQueue) > 0 {
						op := pendingQueue[0]
						pendingQueue = pendingQueue[1:]
						commonState.prepNewCommonState(id)

						switch op.Type {
						case AddOrder:
							commonState.addOrder(op.Order, id)
						case RemoveOrder:
							commonState.removeOrder(op.Order, id)
						case StateUpdate:
							commonState.updateState(op.State, id)
						}
						commonState.PeerSyncStatus[id] = Synced

					} else {
						idle = true
					}

				case commonState.equals(arrivedCommonState):
					commonState = arrivedCommonState
					commonState.makeInactivePeersUnavailable(peersStatus)
					commonState.PeerSyncStatus[id] = Synced
				}
			}
		}
	}
}
