package elevator

import (
	"elevator/config"
	"elevator/elevio"
	"time"
)

type DoorState int

const (
	Closed DoorState = iota
	Closing
	Obstructed
)

func Door(
	doorClosedCh chan<- bool,
	doorOpenCh <-chan bool,
	obtructionCh chan<- bool,
) {

	elevio.SetDoorOpenLamp(false)
	obstructionPollCh := make(chan bool)
	go elevio.PollObstructionSwitch(obstructionPollCh)

	obstruction := false
	doorState := Closed
	timeCounter := time.NewTimer(time.Hour)
	timeCounter.Stop()

	for {
		select {
		case obstruction = <-obstructionPollCh:
			if !obstruction && doorState == Obstructed {
				elevio.SetDoorOpenLamp(false)
				doorClosedCh <- true
				doorState = Closed
			}
			if obstruction {
				obstructionCh <- true
			} else {
				obstructionCh <- false
			}

		case <-doorOpenCh:
			if obstruction {
				obstructionCh <- true
			}
			switch doorState {
			case Closed:
				elevio.SetDoorOpenLamp(true)
				timeCounter = time.NewTimer(config.DoorOpenDuration)
				doorState = Closing

			case Closing:
				timeCounter = time.NewTimer(config.DoorOpenDuration)

			case Obstructed:
				timeCounter = time.NewTimer(config.DoorOpenDuration)
				doorState = Closing

			default:
				panic("Door state not implemented")

			}
		case <-timeCounter.C:
			if doorState != Closing {
				panic("Door state not implemented")
			} else {
				elevio.SetDoorOpenLamp(false)
				doorClosedCh <- true
				doorState = Closed
			}
		}
	}
}
