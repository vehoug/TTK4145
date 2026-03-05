package elevcontrol

import "time"

func resetTimer(timer *time.Timer, timeout time.Duration) <-chan time.Time {
	timer = time.NewTimer(timeout)
	return timer.C
}

func stopTimer(timer *time.Timer) <-chan time.Time {
	timer.Stop()
	return nil
}
