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
	networkReceiveCh    <-chan CommonState,
	peerUpdateCh        <-chan peers.PeerUpdate,
	syncedCommonStateCh chan<- CommonState,
	networkTransmitCh   chan<- CommonState,
	deliveredOrderCh    <-chan elevio.ButtonEvent,
	newStateCh          <-chan elevcontrol.State,
	id                  int,
) {
	newOrderCh := make(chan elevio.ButtonEvent, config.IOBufferSize)
	go elevio.PollButtons(newOrderCh)

	pendingQueue := make([]PendingOperation, 0)

	var commonState    CommonState
	var deliveredOrder elevio.ButtonEvent
	var newOrder       elevio.ButtonEvent
	var newState       elevcontrol.State
	var peersStatus    peers.PeerUpdate

	commonState.initCommonState(id)

	idle    := true
	offline := false

	disconnectTimer := time.NewTimer(config.DisconnectTime)
	heartbeatTicker := time.NewTicker(config.HeartbeatTime)
	defer heartbeatTicker.Stop()

	for {
		select {
		case <-disconnectTimer.C:
			commonState.makeOthersUnavailable(id)
			offline = true
			fmt.Printf("[%v][Distributor]: Network connection lost. Operating independently.\n", time.Now().Format(time.TimeOnly))

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
				commonState.startNewSyncRound(id)
				commonState.addOrder(newOrder, id)
				commonState.PeerSyncStatus[id] = Synced
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
				commonState.startNewSyncRound(id)
				commonState.removeOrder(deliveredOrder, id)
				commonState.PeerSyncStatus[id] = Synced
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
				commonState.startNewSyncRound(id)
				commonState.updateState(newState, id)
				commonState.PeerSyncStatus[id] = Synced
				idle = false

			} else {
				pendingQueue = append(pendingQueue, PendingOperation{
					Type:  StateUpdate,
					State: newState,
				})
			}

		case arrivedCommonState := <-networkReceiveCh:
            disconnectTimer = time.NewTimer(config.DisconnectTime)
			if offline {
				arrivedCommonState.mergeCommonStates(commonState, id)
				commonState = arrivedCommonState
				commonState.makeInactivePeersUnavailable(peersStatus)
				commonState.PeerSyncStatus[id] = Synced
				offline = false
				idle = false
				fmt.Printf("[%v][Distributor]: Network connection restored. Operating normally.\n", time.Now().Format(time.TimeOnly))

			} else if idle {
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

				switch {
				case arrivedCommonState.isNewerThan(commonState):
					commonState = arrivedCommonState
					commonState.makeInactivePeersUnavailable(peersStatus)
					commonState.PeerSyncStatus[id] = Synced

				case arrivedCommonState.fullySynced(id):
					commonState = arrivedCommonState
					syncedCommonStateCh <- commonState

					if len(pendingQueue) > 0 {
						operation := pendingQueue[0]
						pendingQueue = pendingQueue[1:]
						commonState.startNewSyncRound(id)

						switch operation.Type {
						case AddOrder:
							commonState.addOrder(operation.Order, id)

						case RemoveOrder:
							commonState.removeOrder(operation.Order, id)

						case StateUpdate:
							commonState.updateState(operation.State, id)
						}
						commonState.PeerSyncStatus[id] = Synced

					} else {
						idle = true
					}

				case commonState.equalsIgnoringSyncStatus(arrivedCommonState):
					commonState = arrivedCommonState
					commonState.makeInactivePeersUnavailable(peersStatus)
					commonState.PeerSyncStatus[id] = Synced
				}
			}
		}
	}
}
