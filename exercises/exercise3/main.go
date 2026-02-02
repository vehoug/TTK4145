package main

import (
	"elevator-go/elevio"
	"elevator-go/fsm"
	"elevator-go/timer"
	"time"
)

func main() {
	elevio.Init("localhost:15657", 4)
	fsm.Init()

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)

	for {
		select {
		case btn := <-drv_buttons:
			fsm.OnRequestButtonPress(btn.Floor, btn.Button)

		case floor := <-drv_floors:
			fsm.OnFloorArrival(floor)

		default:
			if timer.IsTimeout() {
				fsm.OnDoorTimeout()
			}
			time.Sleep(10 * time.Millisecond)
		}
	}
}
