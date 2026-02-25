package elevcontrol

import (
	"elevator/config"
	"elevator/elevio"
	"fmt"
	"time"
)

type DoorState int

const (
	Closed DoorState = iota
	OpenCountdown
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
				doorTimer = time.NewTimer(config.DoorOpenDuration)
				doorState = OpenCountdown

			case OpenCountdown:
				doorTimer = time.NewTimer(config.DoorOpenDuration)

			case Obstructed:
				doorTimer = time.NewTimer(config.DoorOpenDuration)
				doorState = OpenCountdown

			default:
				panic("Door state not implemented")

			}
		case <-doorTimer.C:
			if doorState != OpenCountdown {
				fmt.Printf("Door timer fired in state=%d obstruction=%v\n", doorState, obstruction)
				panic("Door state not implemented")
			}

			if obstruction {
				doorState = Obstructed

			} else {
				elevio.SetDoorOpenLamp(false)
				doorClosedCh <- true
				doorState = Closed
			}
		}
	}
}
