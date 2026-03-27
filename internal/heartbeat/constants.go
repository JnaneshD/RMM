package heartbeat

import "time"

const (
	PingInterval = 10 * time.Second
	PongWait     = 15 * time.Second
	WriteWait    = 10 * time.Second
)
