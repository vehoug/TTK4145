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
	doorTimerCh <-chan time.Time,
) {

	elevio.SetDoorOpenLamp(false)
	obstructionPollCh := make(chan bool)
	go elevio.PollObstructionSwitch(obstructionPollCh)

	obstruction := false
	doorState := Closed
	doorTimer := time.NewTimer(time.Hour)
	doorTimerCh = resetTimer(doorTimer, config.DoorOpenDuration)

	for {
		select {
		case obstruction = <-obstructionPollCh:
			obstructionCh <- obstruction
			if !obstruction && doorState == Obstructed {
				elevio.SetDoorOpenLamp(false)
				doorClosedCh <- true
				doorState = Closed
			}

		case <-doorOpenCh:
			switch doorState {
			case Closed:
				elevio.SetDoorOpenLamp(true)
				doorTimerCh = resetTimer(doorTimer, config.DoorOpenDuration)
				doorState = OpenCountdown

			case OpenCountdown:
				doorTimerCh = resetTimer(doorTimer, config.DoorOpenDuration)

			case Obstructed:
				doorTimerCh = resetTimer(doorTimer, config.DoorOpenDuration)
				doorState = OpenCountdown

			default:
				panic("Door state not implemented")

			}
		case <-doorTimerCh:
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
