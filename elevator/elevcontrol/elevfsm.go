package elevcontrol

import (
	"elevator/config"
	"elevator/elevio"
	"fmt"
	"time"
)

func ElevatorStateMachine(
	newOrderCh <-chan Orders,
	newStateCh chan<- State,
	deliveredOrderCh chan<- elevio.ButtonEvent,
) {
	var (
		floorEnteredCh = make(chan int,  config.ControlBufferSize)
		doorOpenCh     = make(chan bool, config.ControlBufferSize)
		doorClosedCh   = make(chan bool, config.ControlBufferSize)
		obstructionCh  = make(chan bool, config.ControlBufferSize)
		orders         Orders
	)

	go Door(doorClosedCh, doorOpenCh, obstructionCh)
	go elevio.PollFloorSensor(floorEnteredCh)

	motorTimer := time.NewTimer(config.WatchdogTime)
	elevio.SetMotorDirection(elevio.MD_Down)
	state := initState()

	for {
		select {
		case <-motorTimer.C:
			if state.IsActive && state.CurrentBehaviour == Moving {
				state.IsActive = false
				fmt.Printf("[%v][ElevControl]: Motorpower lost", time.Now().Format(time.TimeOnly))
				newStateCh <- state
			}
			stopTimer(motorTimer)

		case obstructed := <-obstructionCh:
			if obstructed && state.CurrentBehaviour != DoorOpen {
				continue
			}
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
					resetTimer(motorTimer, config.WatchdogTime)
					state.IsActive = true
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
					resetTimer(motorTimer, config.WatchdogTime)
					state.IsActive = true
					newStateCh <- state

				default:
					state.CurrentBehaviour = Idle
					stopTimer(motorTimer)
					newStateCh <- state
				}
			default:
				fmt.Printf("[%v][ElevControl]: Invalid state: Door open with no orders", time.Now().Format(time.TimeOnly))
			}

		case state.CurrentFloor = <-floorEnteredCh:
			state.IsActive = true
			stopTimer(motorTimer)
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
					state.IsActive = true
					resetTimer(motorTimer, config.WatchdogTime)

				case orders.orderAtCurrentFloorOppositeDirection(state.CurrentFloor, state.Direction):
					elevio.SetMotorDirection(elevio.MD_Stop)
					doorOpenCh <- true
					state.Direction = state.Direction.Opposite()
					orders.reportCompletedOrder(state.CurrentFloor, state.Direction, deliveredOrderCh)
					state.CurrentBehaviour = DoorOpen

				case orders.orderOppositeDirection(state.CurrentFloor, state.Direction):
					state.Direction = state.Direction.Opposite()
					elevio.SetMotorDirection(state.Direction.buttonToDirection())
					state.IsActive = true
					resetTimer(motorTimer, config.WatchdogTime)

				default:
					state.CurrentBehaviour = Idle
					elevio.SetMotorDirection(elevio.MD_Stop)
				}
			default:
				fmt.Printf("[%v][ElevControl]: Invalid state: Floor entered while not moving", time.Now().Format(time.TimeOnly))
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
					stopTimer(motorTimer)
					newStateCh <- state

				case orders.orderAtCurrentFloorOppositeDirection(state.CurrentFloor, state.Direction):
					doorOpenCh <- true
					state.Direction = state.Direction.Opposite()
					orders.reportCompletedOrder(state.CurrentFloor, state.Direction, deliveredOrderCh)
					state.CurrentBehaviour = DoorOpen
					stopTimer(motorTimer)
					newStateCh <- state

				case orders.orderInDirection(state.CurrentFloor, state.Direction):
					elevio.SetMotorDirection(state.Direction.buttonToDirection())
					state.CurrentBehaviour = Moving
					state.IsActive = true
					resetTimer(motorTimer, config.WatchdogTime)
					newStateCh <- state

				case orders.orderOppositeDirection(state.CurrentFloor, state.Direction):
					state.Direction = state.Direction.Opposite()
					elevio.SetMotorDirection(state.Direction.buttonToDirection())
					state.CurrentBehaviour = Moving
					state.IsActive = true
					resetTimer(motorTimer, config.WatchdogTime)
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
