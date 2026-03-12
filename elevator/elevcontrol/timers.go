package elevcontrol

import "time"

func resetTimer(timer *time.Timer, timeOut time.Duration) <-chan time.Time {
	timer = time.NewTimer(timeOut)
	return timer.C
}

func stopTimer(timer *time.Timer) {
	timer.Stop()
}
