package main

import (
	"elevator/assigner"
	"elevator/config"
	"elevator/distributor"
	"elevator/elevator"
	"elevator/elevio"
	"elevator/lights"
	"elevator/network/bcast"
	"elevator/network/peers"
	"flag"
	"fmt"
	"strconv"
)

var Port int
var id int

func main() {
	port       := flag.Int("port", 15657, "Default port number")
	elevatorID := flag.Int("id", 0, "Default elevator ID")
	flag.Parse()

	id   = *elevatorID
	Port = *port

	elevio.Init("localhost:" + strconv.Itoa(Port), config.NumFloors)

	newOrderCh          := make(chan elevator.Orders, config.Buffer)
	deliverdOrderCh     := make(chan elevio.ButtonEvent, config.Buffer)
	newStateCh          := make(chan elevator.State, config.Buffer)
	syncedCommonStateCh := make(chan distributor.CommonState, config.Buffer)
	networkReceiveCh    := make(chan distributor.CommonState, config.Buffer)
	networkTransmitCh   := make(chan distributor.CommonState, config.Buffer)
	peerUpdateCh        := make(chan peers.PeerUpdate, config.Buffer)
	peersTransmitCh     := make(chan bool, config.Buffer)

	go peers.Receiver(config.PeersPortNumber, peerUpdateCh)
	go peers.Transmitter(config.PeersPortNumber, strconv.Itoa(id), peersTransmitCh)

	go bcast.Receiver(config.BcastPortNumber, networkReceiveCh)
	go bcast.Transmitter(config.BcastPortNumber, networkTransmitCh)

	fmt.Printf("Node [%d]: starting elevator control system\n", id)

	go distributor.Distributor(
		networkReceiveCh,
		peerUpdateCh,
		syncedCommonStateCh,
		networkTransmitCh,
		deliverdOrderCh,
		newStateCh,
		id)
	
	fmt.Println("Started distributor FSM")
	
	go elevator.Elevator(
		newOrderCh, 
		newStateCh, 
		deliverdOrderCh)
	
	fmt.Println("Started elevator FSM")
	
	for {
		select {
		case commonState := <-syncedCommonStateCh:
			fmt.Printf("Common state: %+v\n", commonState)
			newOrderCh <- assigner.CalculateOptimalOrders(commonState, id)
			lights.SetLights(commonState, id)
		
		default:
			continue
		}
	}
}