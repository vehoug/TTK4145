package elevcontrol

import (
	"elevator/config"
	"elevator/elevio"
	"fmt"
)

func (order Orders) orderInDirection(currentFloor int, dir Direction) bool {
	switch dir {
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
		fmt.Println("Invalid direction")
		return false
	}
}

func (order Orders) reportCompletedOrder(floor int, dir Direction, orderDoneC chan<- elevio.ButtonEvent) {
	if order[floor][elevio.BT_Cab] {
		orderDoneC <- elevio.ButtonEvent{Floor: floor, Button: elevio.BT_Cab}
	}
	if order[floor][dir] {
		orderDoneC <- elevio.ButtonEvent{Floor: floor, Button: dir.directionToButton()}
	}
}
func (order Orders) orderAtCurrentFloorOppositeDirection(floor int, dir Direction) bool {
	return order[floor][dir.Opposite()]
}

func (order Orders) orderOppositeDirection(floor int, dir Direction) bool {
	return order.orderInDirection(floor, dir.Opposite())
}

func (order Orders) orderAtCurrentFloorSameDirection(floor int, dir Direction) bool {
	return order[floor][dir]
}

func (order Orders) shouldStopForCabOrder(floor int, dir Direction) bool {
	if !order[floor][elevio.BT_Cab] {
		return false
	}
	return order.orderInDirection(floor, dir) || !order.orderAtCurrentFloorOppositeDirection(floor, dir)
}
