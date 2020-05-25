package ringing

import "time"

// Config configuration for ringing
type Config struct {
	SendWait   time.Duration
	PingPeriod time.Duration
	PongWait   time.Duration
	ReadLimit  int64
	MsgBufSize int
	Topic      []string
}

func newConfig() *Config {
	t := []string{"orderbook"}
	return &Config{
		SendWait:   10 * time.Second,
		PongWait:   30 * time.Second,
		PingPeriod: (30 * time.Second * 9) / 10,
		ReadLimit:  512,
		MsgBufSize: 512,
		Topic:      t,
	}
}
