package timer

import "time"

var (
	timerEndTime time.Time
	timerActive  bool
)

func Start(duration float64) {
	timerEndTime = time.Now().Add(time.Duration(duration * float64(time.Second)))
	timerActive = true
}

func Stop() {
	timerActive = false
}

func IsTimeout() bool {
	if timerActive && time.Now().After(timerEndTime) {
		timerActive = false
		return true
	}
	return false
}
