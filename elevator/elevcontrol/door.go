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
	doorClosedCh  chan<- bool,
	doorOpenCh    <-chan bool,
	obstructionCh chan<- bool,
	doorTimerCh   <-chan time.Time,
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

			obstructionCh <- (obstruction && doorState != Closed)

		case <-doorOpenCh:
			switch doorState {
			case Closed:
				elevio.SetDoorOpenLamp(true)
				doorTimerCh = resetTimer(doorTimer, config.DoorOpenTime)
				doorState = OpenCountdown

			case OpenCountdown:
				doorTimerCh = resetTimer(doorTimer, config.DoorOpenTime)

			case Obstructed:
				doorTimerCh = resetTimer(doorTimer, config.DoorOpenTime)
				doorState = OpenCountdown

			default:
                fmt.Printf("[%v][ElevControl]: Invalid door state: Door opened while not closed.\n", time.Now().Format(time.TimeOnly))

			}

			obstructionCh <- (obstruction && doorState != Closed)

		case <-doorTimerCh:
			if doorState != OpenCountdown {
				fmt.Printf("[%v][ElevControl]: Invalid door state: Door timer expired while door not open.\n", time.Now().Format(time.TimeOnly))
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
