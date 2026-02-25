package elevcontrol

import (
	"elevator/config"
	"elevator/elevio"
	"fmt"
)

func (orders Orders) orderInDirection(currentFloor int, dir Direction) bool {
	switch dir {
	case Up:
		for scanFloor := currentFloor + 1; scanFloor < config.NumFloors; scanFloor++ {
			for buttonIndex := range config.NumButtons {
				if orders[scanFloor][buttonIndex] {
					return true
				}
			}
		}
		return false
	case Down:
		for scanFloor := currentFloor - 1; scanFloor >= 0; scanFloor-- {
			for buttonIndex := range config.NumButtons {
				if orders[scanFloor][buttonIndex] {
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

func (orders Orders) reportCompletedOrder(floor int, dir Direction, orderDoneC chan<- elevio.ButtonEvent) {
	if orders[floor][elevio.BT_Cab] {
		orderDoneC <- elevio.ButtonEvent{Floor: floor, Button: elevio.BT_Cab}
	}
	if orders[floor][dir] {
		orderDoneC <- elevio.ButtonEvent{Floor: floor, Button: dir.directionToButton()}
	}
}
func (orders Orders) orderAtCurrentFloorOppositeDirection(floor int, dir Direction) bool {
	return orders[floor][dir.Opposite()]
}

func (orders Orders) orderOppositeDirection(floor int, dir Direction) bool {
	return orders.orderInDirection(floor, dir.Opposite())
}

func (orders Orders) orderAtCurrentFloorSameDirection(floor int, dir Direction) bool {
	return orders[floor][dir]
}

func (orders Orders) shouldStopForCabOrder(floor int, dir Direction) bool {
	if !orders[floor][elevio.BT_Cab] {
		return false
	}
	return orders.orderInDirection(floor, dir) || !orders.orderAtCurrentFloorOppositeDirection(floor, dir)
}
