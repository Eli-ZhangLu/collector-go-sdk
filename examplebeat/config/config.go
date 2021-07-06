package config

import (
	"time"
)

// Config is the root of the examplebeat configuration hierarchy.
type Config struct {
	DataID   int32         `config:"dataid"`
	Interval time.Duration `config:"interval"`
	DataSize int32         `config:"datasize"`
}

// DefaultConfig
var DefaultConfig = Config{
	DataID:   0,
	Interval: time.Minute,
	DataSize: 0,
}
