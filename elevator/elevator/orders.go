package elevator

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
