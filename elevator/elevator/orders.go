package elevator

import (
	"elevator/config"
	"elevator/elevio"
	"fmt"
)

func (order Orders) orderInDirection(floor int, dir Direction) bool {
	switch dir {
	case Up:
		for f := floor + 1; f < config.NumFloors; f++ {
			for b := 0; b < config.NumButtons; b++ {
				if order[f][b] {
					return true
				}
			}
		}
		return false
	case Down:
		for f := floor - 1; f >= 0; f-- {
			for b := 0; b < config.NumButtons; b++ {
				if order[f][b] {
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
