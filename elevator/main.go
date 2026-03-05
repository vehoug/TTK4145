package main

import (
	"elevator/assigner"
	"elevator/config"
	"elevator/distributor"
	"elevator/elevcontrol"
	"elevator/elevio"
	"elevator/lights"
	"elevator/network/bcast"
	"elevator/network/peers"
	"flag"
	"fmt"
	"strconv"
)

func main() {
	var port int
	var id int

	flag.IntVar(&port, "port", config.DefaultPort, "Default elevator port number")
	flag.IntVar(&id, "id", 0, "Default elevator ID")
	flag.Parse()

	elevio.Init(fmt.Sprintf("localhost:%d", port), config.NumFloors)

	newOrderCh := make(chan elevcontrol.Orders, config.Buffer)
	deliveredOrderCh := make(chan elevio.ButtonEvent, config.Buffer)
	newStateCh := make(chan elevcontrol.State, config.Buffer)

	syncedCommonStateCh := make(chan distributor.CommonState, config.Buffer)
	networkReceiveCh := make(chan distributor.CommonState, config.Buffer)
	networkTransmitCh := make(chan distributor.CommonState, config.Buffer)

	peerUpdateCh := make(chan peers.PeerUpdate, config.Buffer)
	peersTransmitCh := make(chan bool, config.Buffer)

	go peers.Receiver(config.PeersPortNumber, peerUpdateCh)
	go peers.Transmitter(config.PeersPortNumber, strconv.Itoa(id), peersTransmitCh)

	go bcast.Receiver(config.BcastPortNumber, networkReceiveCh)
	go bcast.Transmitter(config.BcastPortNumber, networkTransmitCh)

	go distributor.Distributor(
		networkReceiveCh,
		peerUpdateCh,
		syncedCommonStateCh,
		networkTransmitCh,
		deliveredOrderCh,
		newStateCh,
		id)

	go elevcontrol.ElevatorStateMachine(
		newOrderCh,
		newStateCh,
		deliveredOrderCh)

	fmt.Printf("Elevator Node [%d] initialized\n\tNumber of floors: %d\n\tNumber of elevators: %d\n",
		id, config.NumFloors, config.NumElevators)

	for commonState := range syncedCommonStateCh {

		newOrderCh <- assigner.CalculateOptimalOrders(commonState, id)
		lights.SetLights(commonState, id)
	}
}
