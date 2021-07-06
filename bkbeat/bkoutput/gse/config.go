package gse

import "time"

type Config struct {
	// gse client config
	RetryTimes     uint          `config:"retrytimes"`
	RetryInterval  time.Duration `config:"retryinterval"`
	Nonblock       bool          `config:"nonblock"`
	EventBufferMax int32         `config:"eventbuffermax"`
	MsgQueueSize   uint32        `config:"mqsize"`
	Endpoint       string        `config:"endpoint"`
	WriteTimeout   time.Duration `config:"writetimeout"` // unit: second

	// monitor config
	MonitorID  int32 `config:"monitorid"`  // <= 0 : disable bk monitor tag
	ResourceID int32 `config:"resourceid"` // <= 0 : disable resource report
}

var (
	defaultConfig = Config{
		MonitorID: 295,
	}
)
