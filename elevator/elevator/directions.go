package elevator

func (d Direction) buttonToDirection() elevio.MotorDirection {
	return map[Direction]elevio.MotorDirection{Up: elevio.MD_Up, Down: elevio.MD_Down}[d]
}
func (d Direction) directionToButton() elevio.ButtonType {
	return map[Direction]elevio.ButtonType{Up: elevio.BT_HallUp, Down: elevio.BT_HallDown}[d]
}
func (d Direction) Opposite() Direction {
	return map[Direction]Direction{Up: Down, Down: Up}[d]
}
