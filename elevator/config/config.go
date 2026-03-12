package config

import (
	"time"
)

const (
	NumFloors         =  4
	NumElevators      =  3
	NumButtons        =  3
	NumDirections     =  2
    UnknownFloor      = -1

	PeersPortNumber   = 58735
	BcastPortNumber   = 58750
	IOBufferSize      = 1024
	ControlBufferSize = 16
	DefaultPort       = 15657

	DisconnectTime   =  1 * time.Second
	DoorOpenTime     =  3 * time.Second
	WatchdogTime     =  3 * time.Second
	HeartbeatTime    = 15 * time.Millisecond
)
