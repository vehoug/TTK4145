package elevator

import (
	"elevator/config"
	"time"
)

func Elevator(
	newOrderCh <-chan Orders,
	newStateCh chan<- State,
	deliveredOrderCh chan<- elevio.ButtonEvent,
) {
	var (
		floorEnteredCh  = make(chan int, ChannelBufferSize)
		doorOpenCh      = make(chan bool, ChannelBufferSize)
		doorClosedCh    = make(chan bool, ChannelBufferSize)
		obstructionCh   = make(chan bool, ChannelBufferSize)
		motorInactiveCh = make(chan bool, ChannelBufferSize)
		orders          = make(Orders)
	)

	go Door(doorClosedCh, doorOpenCh, obstructionCh)
	go elevio.PollFloorSensor(floorEnteredCh)

	elevio.SetMotorDirection(elevio.MD_Down)
	state := initState()

	motorTimer := time.NewTimer(config.WatchdogTime)
	motorTimer.Stop()
	for {
		select {
		case <-motorTimer.C:
			if state.ActiveStatus {
				state.ActiveStatus = false
				newStateCh <- state
			}
		case <-motorInactiveCh:
			if !state.ActiveStatus {
				state.ActiveStatus = true
				newStateCh <- state
			}
		case obstructed := <-obstructionCh:
			if obstructed != state.Obstructed {
				state.Obstructed = obstructed
				newStateCh <- state
			}

		case <-doorOpenCh:
			switch state.CurrentBehaviour {
			case DoorOpen:
				switch {
				case orders.OrderInDirection(state.CurrentFloor, state.Direction):
				}

			}
		}
	}
}
