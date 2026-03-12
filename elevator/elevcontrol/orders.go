package elevcontrol

import (
	"elevator/config"
	"elevator/elevio"
    "time"
	"fmt"
	"time"
)

func (order Orders) orderInDirection(currentFloor int, direction Direction) bool {
	switch direction {
	case Up:
		for floor := currentFloor + 1; floor < config.NumFloors; floor++ {
			for button := range config.NumButtons {
				if order[floor][button] {
					return true
				}
			}
		}
		return false

	case Down:
		for floor := currentFloor - 1; floor >= 0; floor-- {
			for button := range config.NumButtons {
				if order[floor][button] {
					return true
				}
			}
		}
		return false

	default:
		fmt.Printf("[%v][ElevControl]: Invalid direction.\n", time.Now().Format(time.TimeOnly))
		return false
	}
}

func (order Orders) reportCompletedOrder(floor int, direction Direction, orderDoneC chan<- elevio.ButtonEvent) {
	if order[floor][elevio.BT_Cab] {
		orderDoneCh <- elevio.ButtonEvent{Floor: floor, Button: elevio.BT_Cab}
	}
	if order[floor][direction] {
		orderDoneC <- elevio.ButtonEvent{Floor: floor, Button: direction.directionToButton()}
	}
}

func (order Orders) orderAtCurrentFloorOppositeDirection(floor int, direction Direction) bool {
	return order[floor][direction.Opposite()]
}

func (order Orders) orderOppositeDirection(floor int, direction Direction) bool {
	return order.orderInDirection(floor, direction.Opposite())
}

func (order Orders) orderAtCurrentFloorSameDirection(floor int, direction Direction) bool {
	return order[floor][direction]
}

func (order Orders) shouldStopForCabOrder(floor int, direction Direction) bool {
	if !order[floor][elevio.BT_Cab] {
		return false
	}
	return order.orderInDirection(floor, direction) || !order.orderAtCurrentFloorOppositeDirection(floor, direction)
}
