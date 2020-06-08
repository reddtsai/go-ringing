package ringing

import "time"

// Config configuration for ringing
type Config struct {
	SendWait   time.Duration
	PingPeriod time.Duration
	PongWait   time.Duration
	ReadLimit  int64
	MsgBufSize int
	NSQSetting *NSQSetting
}

// NSQSetting configuration for nsq
type NSQSetting struct {
	NSQAddr      string
	Topics       []string
	TopicChannel string
}

// NewConfig create a new config
func NewConfig(nsqSetting *NSQSetting) *Config {
	return &Config{
		SendWait:   10 * time.Second,
		PongWait:   60 * time.Second,
		PingPeriod: (60 * time.Second * 9) / 10,
		ReadLimit:  512,
		MsgBufSize: 512,
		NSQSetting: nsqSetting,
	}
}
