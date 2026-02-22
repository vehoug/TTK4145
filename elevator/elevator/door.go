package elevator

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
			obstructionCh <- obstruction
			if obstruction {
				if doorState == OpenCountdown {
					doorTimer.Reset(config.DoorOpenDuration)
					doorState = Obstructed
				}
			} else {
				if doorState == Obstructed {
					doorTimer.Reset(config.DoorOpenDuration)
					doorState = OpenCountdown
				}
			}

		case <-doorOpenCh:
			switch doorState {
			case Closed:
				elevio.SetDoorOpenLamp(true)
				doorTimer.Reset(config.DoorOpenDuration)
				doorState = OpenCountdown

			case OpenCountdown:
				doorTimer.Reset(config.DoorOpenDuration)

			case Obstructed:
				doorTimer.Reset(config.DoorOpenDuration)
				doorState = OpenCountdown

			default:
				panic("Door state not implemented")

			}
		case <-doorTimer.C:
			if doorState != OpenCountdown || obstruction {
				fmt.Printf("Door timer fired in state=%d obstruction=%v\n", doorState, obstruction)
				continue
				//panic("Door state not implemented")
			} else {
				elevio.SetDoorOpenLamp(false)
				doorClosedCh <- true
				doorState = Closed
			}
		}
	}
}
