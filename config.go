package ringing

import "time"

// Config configuration for ringing
type Config struct {
	SendWait     time.Duration
	PingPeriod   time.Duration
	PongWait     time.Duration
	ReadLimit    int64
	MsgBufSize   int
	Topics       []string
	TopicHost    string
	TopicChannel string
}

// NewConfig create a new config
func NewConfig(nsqAddr string) *Config {
	t := []string{"all"}
	return &Config{
		SendWait:     10 * time.Second,
		PongWait:     60 * time.Second,
		PingPeriod:   (60 * time.Second * 9) / 10,
		ReadLimit:    512,
		MsgBufSize:   512,
		Topics:       t,
		TopicHost:    nsqAddr,
		TopicChannel: "ch1",
	}
}
