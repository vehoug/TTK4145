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
	obstructionCh chan<- bool,
) {

	elevio.SetDoorOpenLamp(false)
	obstructionPollCh := make(chan bool)
	go elevio.PollObstructionSwitch(obstructionPollCh)

	obstruction := false
	doorState := Closed
	doorTimer := time.NewTimer(time.Hour)
	doorTimer.Stop()

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
				doorTimer.Reset(config.DoorOpenDuration)
				doorState = Closing

			case Closing:
				doorTimer.Reset(config.DoorOpenDuration)

			case Obstructed:
				doorTimer.Reset(config.DoorOpenDuration)
				doorState = Closing

			default:
				panic("Door state not implemented")

			}
		case <-doorTimer.C:
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
