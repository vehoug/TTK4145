package elevator

import (
	"elevator/config"
	"elevator/elevio"
	"time"
)

func Elevator(
	newOrderCh <-chan Orders,
	newStateCh chan<- State,
	deliveredOrderCh chan<- elevio.ButtonEvent,
) {
	var (
		floorEnteredCh  = make(chan int, config.ChannelBufferSize)
		doorOpenCh      = make(chan bool, config.ChannelBufferSize)
		doorClosedCh    = make(chan bool, config.ChannelBufferSize)
		obstructionCh   = make(chan bool, config.ChannelBufferSize)
		motorInactiveCh <-chan time.Time
		orders          Orders
	)

	go Door(doorClosedCh, doorOpenCh, obstructionCh)
	go elevio.PollFloorSensor(floorEnteredCh)

	motorTimer := time.NewTimer(config.WatchdogTime)

	elevio.SetMotorDirection(elevio.MD_Down)
	motorInactiveCh = startMotorTimer(motorTimer, config.WatchdogTime)

	state := initState()

	for {
		select {
		case <-motorInactiveCh:
			if state.ActiveStatus && state.CurrentBehaviour == Moving {
				state.ActiveStatus = false
				newStateCh <- state
				motorInactiveCh = stopMotorTimer(motorTimer)
			}
		case obstructed := <-obstructionCh:
			if obstructed != state.Obstructed {
				state.Obstructed = obstructed
				newStateCh <- state
			}

		case <-doorClosedCh:
			switch state.CurrentBehaviour {
			case DoorOpen:
				switch {
				case orders.orderInDirection(state.CurrentFloor, state.Direction):
					elevio.SetMotorDirection(state.Direction.buttonToDirection())
					state.CurrentBehaviour = Moving
					motorInactiveCh = startMotorTimer(motorTimer, config.WatchdogTime)
					state.ActiveStatus = true
					newStateCh <- state

				case orders.orderAtCurrentFloorOppositeDirection(state.CurrentFloor, state.Direction):
					doorOpenCh <- true
					state.Direction = state.Direction.Opposite()
					orders.reportCompletedOrder(state.CurrentFloor, state.Direction, deliveredOrderCh)
					newStateCh <- state

				case orders.orderOppositeDirection(state.CurrentFloor, state.Direction):
					state.Direction = state.Direction.Opposite()
					state.CurrentBehaviour = Moving
					elevio.SetMotorDirection(state.Direction.buttonToDirection())
					motorInactiveCh = startMotorTimer(motorTimer, config.WatchdogTime)
					state.ActiveStatus = true
					newStateCh <- state

				default:
					state.CurrentBehaviour = Idle
					newStateCh <- state
				}
			default:
				panic("Invalid state: Door open with no orders")
			}

		case state.CurrentFloor = <-floorEnteredCh:
			state.ActiveStatus = true
			motorInactiveCh = stopMotorTimer(motorTimer)
			elevio.SetFloorIndicator(state.CurrentFloor)
			switch state.CurrentBehaviour {
			case Moving:
				switch {
				case orders.orderAtCurrentFloorSameDirection(state.CurrentFloor, state.Direction):
					elevio.SetMotorDirection(elevio.MD_Stop)
					doorOpenCh <- true
					orders.reportCompletedOrder(state.CurrentFloor, state.Direction, deliveredOrderCh)
					state.CurrentBehaviour = DoorOpen

				case orders.shouldStopForCabOrder(state.CurrentFloor, state.Direction):
					elevio.SetMotorDirection(elevio.MD_Stop)
					doorOpenCh <- true
					orders.reportCompletedOrder(state.CurrentFloor, state.Direction, deliveredOrderCh)
					state.CurrentBehaviour = DoorOpen

				case orders.orderInDirection(state.CurrentFloor, state.Direction):
					state.ActiveStatus = true
					motorInactiveCh = startMotorTimer(motorTimer, config.WatchdogTime)

				case orders.orderAtCurrentFloorOppositeDirection(state.CurrentFloor, state.Direction):
					elevio.SetMotorDirection(elevio.MD_Stop)
					doorOpenCh <- true
					state.Direction = state.Direction.Opposite()
					orders.reportCompletedOrder(state.CurrentFloor, state.Direction, deliveredOrderCh)
					state.CurrentBehaviour = DoorOpen

				case orders.orderOppositeDirection(state.CurrentFloor, state.Direction):
					state.Direction = state.Direction.Opposite()
					elevio.SetMotorDirection(state.Direction.buttonToDirection())
					state.ActiveStatus = true
					motorInactiveCh = startMotorTimer(motorTimer, config.WatchdogTime)

				default:
					state.CurrentBehaviour = Idle
					elevio.SetMotorDirection(elevio.MD_Stop)
				}
			default:
				panic("Invalid state: Floor entered while not moving")
			}
			newStateCh <- state

		case orders = <-newOrderCh:
			switch state.CurrentBehaviour {
			case Idle:
				switch {
				case orders.orderAtCurrentFloorSameDirection(state.CurrentFloor, state.Direction) || orders[state.CurrentFloor][elevio.BT_Cab]:
					doorOpenCh <- true
					orders.reportCompletedOrder(state.CurrentFloor, state.Direction, deliveredOrderCh)
					state.CurrentBehaviour = DoorOpen
					newStateCh <- state

				case orders.orderAtCurrentFloorOppositeDirection(state.CurrentFloor, state.Direction):
					doorOpenCh <- true
					state.Direction = state.Direction.Opposite()
					orders.reportCompletedOrder(state.CurrentFloor, state.Direction, deliveredOrderCh)
					state.CurrentBehaviour = DoorOpen
					newStateCh <- state

				case orders.orderInDirection(state.CurrentFloor, state.Direction):
					elevio.SetMotorDirection(state.Direction.buttonToDirection())
					state.CurrentBehaviour = Moving
					state.ActiveStatus = true
					motorInactiveCh = startMotorTimer(motorTimer, config.WatchdogTime)
					newStateCh <- state

				case orders.orderOppositeDirection(state.CurrentFloor, state.Direction):
					state.Direction = state.Direction.Opposite()
					elevio.SetMotorDirection(state.Direction.buttonToDirection())
					state.CurrentBehaviour = Moving
					state.ActiveStatus = true
					motorInactiveCh = startMotorTimer(motorTimer, config.WatchdogTime)
					newStateCh <- state
				}

			case DoorOpen:
				if orders.orderAtCurrentFloorSameDirection(state.CurrentFloor, state.Direction) || orders[state.CurrentFloor][elevio.BT_Cab] {
					doorOpenCh <- true
					orders.reportCompletedOrder(state.CurrentFloor, state.Direction, deliveredOrderCh)
				}
			case Moving:

			default:
			}
		}
	}
}
