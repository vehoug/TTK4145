package distributor

import (
	"elevator/config"
	"elevator/elevator"
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
	State elevator.State
}

func Distributor(
	networkReceiveCh    <-chan CommonState,
	peerUpdateCh        <-chan peers.PeerUpdate,
	syncedCommonStateCh chan<- CommonState,
	networkTransmitCh   chan<- CommonState,
	deliveredOrderCh    <-chan elevio.ButtonEvent,
    newStateCh          <-chan elevator.State,
	id                  int,
) {
	newOrderCh := make(chan elevio.ButtonEvent, config.Buffer)
	go elevio.PollButtons(newOrderCh)

	pendingQueue := make([]PendingOperation, 0)

	var commonState    CommonState
	var deliveredOrder elevio.ButtonEvent
	var newOrder       elevio.ButtonEvent
	var newState       elevator.State
	var peersStatus    peers.PeerUpdate

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
			fmt.Printf("Node [%d]: lost network connection", id)
		
		case peersStatus = <-peerUpdateCh:
			commonState.makeOthersUnavailable(id)
			idle = false
		
		case <-heartbeatTicker.C:
			networkTransmitCh <- commonState
		
		default:
		}

		switch {
		case idle:
			select {
			case newOrder = <-newOrderCh:
				commonState.applyTransaction(func() {
					commonState.addOrder(newOrder, id)
				}, id)
				idle = false
			
			case deliveredOrder = <-deliveredOrderCh:
				commonState.applyTransaction(func() {
					commonState.removeOrder(deliveredOrder, id)
				}, id)
				idle = false
				
			case newState = <-newStateCh:
				commonState.applyTransaction(func() {
					commonState.updateState(newState, id)
				}, id)
				idle = false
			
			case arrivedCommonState := <-networkReceiveCh:
				disconnectTimer.Reset(config.DisconnectTime)
				if arrivedCommonState.isNewerThan(commonState) {
					commonState = arrivedCommonState
					commonState.makeLostPeersUnavailable(peersStatus)
					commonState.PeerSyncStatus[id] = Synced
					idle = false
				}
			
			default:
			}
		
		case !idle:
			select {
				case newOrder = <-newOrderCh:
					pendingQueue = append(pendingQueue, PendingOperation{
						Type:  AddOrder,
						Order: newOrder,
					})
				
				case deliveredOrder = <-deliveredOrderCh:
					pendingQueue = append(pendingQueue, PendingOperation{
						Type:  RemoveOrder,
						Order: deliveredOrder,
					})
				
				case newState = <-newStateCh:
					pendingQueue = append(pendingQueue, PendingOperation{
						Type:  StateUpdate,
						State: newState,
					})

				case arrivedCommonState := <-networkReceiveCh:
					if arrivedCommonState.isOlderThan(commonState) {
						break
					}
					disconnectTimer.Reset(config.DisconnectTime)

					switch {
					case arrivedCommonState.isNewerThan(commonState):
						commonState = arrivedCommonState
						commonState.makeLostPeersUnavailable(peersStatus)
						commonState.PeerSyncStatus[id] = Synced
					
					case arrivedCommonState.fullySynced(id):
						commonState = arrivedCommonState
						syncedCommonStateCh <- commonState

						if len(pendingQueue) > 0 {
							commonState.prepNewCommonState(id)
							
							for _, op := range pendingQueue {
								switch op.Type {
								case AddOrder:
									commonState.addOrder(op.Order, id)

								case RemoveOrder:
									commonState.removeOrder(op.Order, id)

								case StateUpdate:
									commonState.updateState(op.State, id)

								}
							}
							commonState.PeerSyncStatus[id] = Synced
							pendingQueue = pendingQueue[:0]

						} else {
							idle = true
						}
						
					case commonState.equals(arrivedCommonState):
						commonState = arrivedCommonState
						commonState.makeLostPeersUnavailable(peersStatus)
						commonState.PeerSyncStatus[id] = Synced
					
					default:
					}
				default:
			}

		case offline:
			select {
				case <-networkReceiveCh:
					if commonState.LocalStates[id].CabRequests == [config.NumFloors]bool{} {
						offline = false
						fmt.Printf("Node [%d]: reconnected to network", id)
					} else {
						commonState.PeerSyncStatus[id] = Unavailable
					}
				
				case newOrder = <-newOrderCh:
					if commonState.LocalStates[id].State.ActiveStatus {
						commonState.PeerSyncStatus[id] = Synced
						commonState.addOrder(newOrder, id)
						syncedCommonStateCh <- commonState
					}
				
				case deliveredOrder = <-deliveredOrderCh:
					commonState.PeerSyncStatus[id] = Synced
					commonState.removeOrder(deliveredOrder, id)
					syncedCommonStateCh <- commonState
				
				case newState = <-newStateCh:
					if newState.ActiveStatus && !newState.Obstructed {
						commonState.PeerSyncStatus[id] = Synced
						commonState.updateState(newState, id)
						syncedCommonStateCh <- commonState
					}
				default:
			}
		}
	}
}