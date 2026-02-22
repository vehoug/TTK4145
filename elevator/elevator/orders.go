package elevator

<<<<<<< Updated upstream
=======
import (
	"elevator/config"
	"elevator/elevio"
)

>>>>>>> Stashed changes
func (order Orders) OrderInDirection(floor int, dir Direction) bool {
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
		panic("Invalid direction")
	}
}
<<<<<<< Updated upstream
=======

func (order Orders) ReportCompletedOrder(floor int, dir Direction, orderDoneC chan<- elevio.ButtonEvent) {
	if order[floor][elevio.BT_Cab] {
		orderDoneC <- elevio.ButtonEvent{Floor: floor, Button: elevio.BT_Cab}
	}
	if order[floor][dir] {
		orderDoneC <- elevio.ButtonEvent{Floor: floor, Button: dir.directionToButton()}
	}
}
func (order Orders) OrderAtCurrentFloorOppositeDirection(floor int, dir Direction) bool {
	return order[floor][dir.Opposite()]
}
func (order Orders) OrderOppositeDirection(floor int, dir Direction) bool {
	return order.OrderInDirection(floor, dir.Opposite())
}
func (order Orders) OrderAtCurrentFloorSameDirection(floor int, dir Direction) bool {
	return order[floor][dir]
}

func (order Orders) ShouldStopForCabOrder(floor int, dir Direction) bool {
	if !order[floor][elevio.BT_Cab] {
		return false
	}
	return order.OrderInDirection(floor, dir) || !order.OrderAtCurrentFloorOppositeDirection(floor, dir)
}
>>>>>>> Stashed changes
