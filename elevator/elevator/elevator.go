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
		motorInactiveCh <-chan time.Time
		orders          Orders
	)

	go Door(doorClosedCh, doorOpenCh, obstructionCh)
	go elevio.PollFloorSensor(floorEnteredCh)

	elevio.SetMotorDirection(elevio.MD_Down)
	state := initState()

	motorTimer := time.NewTimer(config.WatchdogTime)
	motorInactiveCh = stopMotorTimer(motorTimer)
	for {
		select {
		case <-motorInactiveCh:
			if state.ActiveStatus {
				state.ActiveStatus = false
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
					elevio.SetMotorDirection(state.Direction.buttonToDirection())
					state.CurrentBehaviour = Moving
					motorInactiveCh = startMotorTimer(motorTimer, config.WatchdogTime)
					state.ActiveStatus = true

				case orders.OrderAtCurrentFloorOppositeDirection(state.CurrentFloor, state.Direction):
					doorOpenCh <- true
					state.Direction = state.Direction.Opposite()
					orders.ReportCompletedOrder(state.CurrentFloor, state.Direction, deliveredOrderCh)

				case orders.OrderOppositeDirection(state.CurrentFloor, state.Direction):
					state.Direction = state.Direction.Opposite()
					state.CurrentBehaviour = Moving
					elevio.SetMotorDirection(state.Direction.buttonToDirection())
					motorInactiveCh = startMotorTimer(motorTimer, config.WatchdogTime)
					state.ActiveStatus = true
				default:
					state.CurrentBehaviour = Idle
				}
			default:
				panic("Invalid state: Door open with no orders")
			}
			newStateCh <- state

		case state.CurrentFloor = <-floorEnteredCh:
			state.ActiveStatus = true
			motorInactiveCh = stopMotorTimer(motorTimer)
			elevio.SetFloorIndicator(state.CurrentFloor)
			switch state.CurrentBehaviour {
			case Moving:
				switch {
				case orders.OrderAtCurrentFloorSameDirection(state.CurrentFloor, state.Direction):
					elevio.SetMotorDirection(elevio.MD_Stop)
					doorOpenCh <- true
					orders.ReportCompletedOrder(state.CurrentFloor, state.Direction, deliveredOrderCh)
					state.CurrentBehaviour = DoorOpen

				case orders.ShouldStopForCabOrder(state.CurrentFloor, state.Direction):
					elevio.SetMotorDirection(elevio.MD_Stop)
					doorOpenCh <- true
					orders.ReportCompletedOrder(state.CurrentFloor, state.Direction, deliveredOrderCh)
					state.CurrentBehaviour = DoorOpen

				case orders.OrderInDirection(state.CurrentFloor, state.Direction):
					state.ActiveStatus = true
					motorInactiveCh = startMotorTimer(motorTimer, config.WatchdogTime)

				case orders.OrderAtCurrentFloorOppositeDirection(state.CurrentFloor, state.Direction):
					elevio.SetMotorDirection(elevio.MD_Stop)
					doorOpenCh <- true
					state.Direction = state.Direction.Opposite()
					orders.ReportCompletedOrder(state.CurrentFloor, state.Direction, deliveredOrderCh)
					state.CurrentBehaviour = DoorOpen

				case orders.OrderOppositeDirection(state.CurrentFloor, state.Direction):
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
				case orders.OrderAtCurrentFloorSameDirection(state.CurrentFloor, state.Direction) || orders[state.CurrentFloor][elevio.BT_Cab]:
					doorOpenCh <- true
					orders.ReportCompletedOrder(state.CurrentFloor, state.Direction, deliveredOrderCh)
					state.CurrentBehaviour = DoorOpen
					motorInactiveCh = stopMotorTimer(motorTimer)

				case orders.OrderAtCurrentFloorOppositeDirection(state.CurrentFloor, state.Direction):
					doorOpenCh <- true
					state.Direction = state.Direction.Opposite()
					orders.ReportCompletedOrder(state.CurrentFloor, state.Direction, deliveredOrderCh)
					state.CurrentBehaviour = DoorOpen
					motorInactiveCh = stopMotorTimer(motorTimer)

				case orders.OrderInDirection(state.CurrentFloor, state.Direction):
					elevio.SetMotorDirection(state.Direction.buttonToDirection())
					state.CurrentBehaviour = Moving
					state.ActiveStatus = true
					motorInactiveCh = startMotorTimer(motorTimer, config.WatchdogTime)

				case orders.OrderOppositeDirection(state.CurrentFloor, state.Direction):
					state.Direction = state.Direction.Opposite()
					elevio.SetMotorDirection(state.Direction.buttonToDirection())
					state.CurrentBehaviour = Moving
					state.ActiveStatus = true
					motorInactiveCh = startMotorTimer(motorTimer, config.WatchdogTime)
				}
			case DoorOpen:
				if orders.OrderAtCurrentFloorSameDirection(state.CurrentFloor, state.Direction) || orders[state.CurrentFloor][elevio.BT_Cab] {
					doorOpenCh <- true
					orders.ReportCompletedOrder(state.CurrentFloor, state.Direction, deliveredOrderCh)
				}
			default:
				panic("Invalid state: New order received while moving")
			}
			newStateCh <- state
		}
	}
}
