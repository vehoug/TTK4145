package elevcontrol

import "time"

func startMotorTimer(motorTimer *time.Timer, timeout time.Duration) <-chan time.Time {
	motorTimer.Reset(timeout)
	return motorTimer.C
}

func stopMotorTimer(motorTimer *time.Timer) <-chan time.Time {
	motorTimer.Stop()
	return nil
}
